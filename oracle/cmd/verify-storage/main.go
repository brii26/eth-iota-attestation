// Command verify-storage checks the local RustFS and Verdaccio services and
// prints a JSON report. It exits with status 1 when any check fails.
package main

import (
	"bytes"
	"context"
	"crypto/sha1"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type objectResult struct {
	Verified bool   `json:"verified"`
	SHA256   string `json:"sha256"`
	Bytes    int    `json:"bytes"`
	Listed   bool   `json:"listed"`
}

type s3Result struct {
	Buckets map[string]string       `json:"buckets"`
	Objects map[string]objectResult `json:"objects"`
}

type rejection struct {
	Status   int  `json:"status"`
	Asserted bool `json:"asserted"`
}

type verdaccioResult struct {
	FirstPublish struct {
		Status  int  `json:"status"`
		Success bool `json:"success"`
	} `json:"firstPublish"`
	DuplicatePublishRejection rejection `json:"duplicatePublishRejection"`
	UnpublishRejection        rejection `json:"unpublishRejection"`
}

type report struct {
	Timestamp string `json:"timestamp"`
	Status    string `json:"status"`
	Tests     struct {
		S3        *s3Result        `json:"s3,omitempty"`
		Verdaccio *verdaccioResult `json:"verdaccio,omitempty"`
	} `json:"tests"`
	Error string `json:"error,omitempty"`
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func sha256Hex(b []byte) string {
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func verifyS3(ctx context.Context) (*s3Result, error) {
	endpoint, err := url.Parse(env("RUSTFS_ENDPOINT", "http://127.0.0.1:9000"))
	if err != nil {
		return nil, err
	}
	client, err := minio.New(endpoint.Host, &minio.Options{
		Creds: credentials.NewStaticV4(
			env("RUSTFS_ACCESS_KEY", "rustfsadmin"),
			env("RUSTFS_SECRET_KEY", "rustfsdevpassword123"), ""),
		Secure:       endpoint.Scheme == "https",
		Region:       env("AWS_REGION", "us-east-1"),
		BucketLookup: minio.BucketLookupPath,
	})
	if err != nil {
		return nil, err
	}

	res := &s3Result{Buckets: map[string]string{}, Objects: map[string]objectResult{}}
	for _, bucket := range []string{"sboms", "scans"} {
		exists, err := client.BucketExists(ctx, bucket)
		if err != nil || !exists {
			if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
				return nil, fmt.Errorf("create bucket %s: %w", bucket, err)
			}
		}
		res.Buckets[bucket] = "exists_or_created"
	}

	files := []struct {
		bucket, key string
		content     []byte
	}{
		{"sboms", "sample-sbom.json", []byte(`{"name":"app","version":"1.0.0","dependencies":{}}`)},
		{"scans", "sample-scan.json", []byte(`{"scanId":"12345","vulnerabilities":[]}`)},
	}
	for _, f := range files {
		name := f.bucket + "/" + f.key
		_, err := client.PutObject(ctx, f.bucket, f.key, bytes.NewReader(f.content),
			int64(len(f.content)), minio.PutObjectOptions{ContentType: "application/json"})
		if err != nil {
			return nil, fmt.Errorf("put %s: %w", name, err)
		}

		obj, err := client.GetObject(ctx, f.bucket, f.key, minio.GetObjectOptions{})
		if err != nil {
			return nil, fmt.Errorf("get %s: %w", name, err)
		}
		got, err := io.ReadAll(obj)
		obj.Close()
		if err != nil {
			return nil, fmt.Errorf("read %s: %w", name, err)
		}
		if sha256Hex(got) != sha256Hex(f.content) || !bytes.Equal(got, f.content) {
			return nil, fmt.Errorf("data corruption detected for %s", name)
		}

		listed := false
		for o := range client.ListObjects(ctx, f.bucket, minio.ListObjectsOptions{}) {
			if o.Err != nil {
				return nil, fmt.Errorf("list %s: %w", f.bucket, o.Err)
			}
			if o.Key == f.key {
				listed = true
			}
		}
		if !listed {
			return nil, fmt.Errorf("key %s not found in listing of bucket %s", f.key, f.bucket)
		}

		res.Objects[name] = objectResult{true, sha256Hex(got), len(got), true}
	}
	return res, nil
}

func do(ctx context.Context, method, target, token string, body any) (int, []byte, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return 0, nil, err
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequestWithContext(ctx, method, target, r)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	return resp.StatusCode, b, err
}

func verifyVerdaccio(ctx context.Context) (*verdaccioResult, error) {
	base := env("VERDACCIO_URL", "http://127.0.0.1:4873")
	stamp := time.Now().UnixMilli()
	username := fmt.Sprintf("testuser_%d", stamp)
	pkg := fmt.Sprintf("@test/artifact-%d", stamp)
	version := "1.0.0"

	status, body, err := do(ctx, http.MethodPut, base+"/-/user/org.couchdb.user:"+username, "", map[string]any{
		"name": username, "password": "Password123!", "email": username + "@example.com",
		"type": "user", "roles": []string{}, "date": time.Now().UTC().Format(time.RFC3339),
	})
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK && status != http.StatusCreated {
		return nil, fmt.Errorf("failed to register Verdaccio user: %d %s", status, body)
	}
	var auth struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(body, &auth); err != nil || auth.Token == "" {
		return nil, fmt.Errorf("no token in Verdaccio response: %s", body)
	}

	tarball := []byte("mock package content")
	sum := sha1.Sum(tarball)
	filename := fmt.Sprintf("%s-%s.tgz", pkg[strings.Index(pkg, "/")+1:], version)
	payload := map[string]any{
		"_id": pkg, "name": pkg,
		"dist-tags": map[string]string{"latest": version},
		"versions": map[string]any{version: map[string]any{
			"name": pkg, "version": version,
			"dist": map[string]string{
				"shasum":  hex.EncodeToString(sum[:]),
				"tarball": fmt.Sprintf("%s/%s/-/%s", base, pkg, filename),
			},
		}},
		"_attachments": map[string]any{filename: map[string]any{
			"content_type": "application/octet-stream",
			"data":         base64.StdEncoding.EncodeToString(tarball),
			"length":       len(tarball),
		}},
	}

	pkgURL := base + "/" + url.PathEscape(pkg)
	res := &verdaccioResult{}

	status, body, err = do(ctx, http.MethodPut, pkgURL, auth.Token, payload)
	if err != nil {
		return nil, err
	}
	if status < 200 || status > 299 {
		return nil, fmt.Errorf("first publish failed unexpectedly: %d %s", status, body)
	}
	res.FirstPublish.Status, res.FirstPublish.Success = status, true

	status, _, err = do(ctx, http.MethodPut, pkgURL, auth.Token, payload)
	if err != nil {
		return nil, err
	}
	if status != http.StatusConflict && status != http.StatusForbidden {
		return nil, fmt.Errorf("duplicate publish expected 409 or 403, got %d", status)
	}
	res.DuplicatePublishRejection = rejection{status, true}

	status, _, err = do(ctx, http.MethodDelete, pkgURL+"/-rev/1-rev", auth.Token, nil)
	if err != nil {
		return nil, err
	}
	if status >= 200 && status <= 299 {
		return nil, fmt.Errorf("unpublish expected rejection, got %d", status)
	}
	res.UnpublishRejection = rejection{status, true}
	return res, nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	out := report{Timestamp: time.Now().UTC().Format(time.RFC3339), Status: "PASSED"}
	var err error
	if out.Tests.S3, err = verifyS3(ctx); err == nil {
		out.Tests.Verdaccio, err = verifyVerdaccio(ctx)
	}
	if err != nil {
		out.Status, out.Error = "FAILED", err.Error()
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	_ = enc.Encode(out)
	if err != nil {
		os.Exit(1)
	}
}
