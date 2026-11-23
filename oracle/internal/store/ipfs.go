package store

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const maxIPFSSBOMBytes = 64 << 20

var ipfsCID = regexp.MustCompile(`^b[a-z2-7]{58}$`)

type IPFSSBOMStore struct {
	*S3Store
	index    *SQLiteIndex
	endpoint *url.URL
	client   *http.Client
	apps     map[string]struct{}
}

func NewIPFSSBOMStore(primary *S3Store, index *SQLiteIndex, endpoint, publicApps string) (*IPFSSBOMStore, error) {
	server, err := url.Parse(endpoint)
	if err != nil || (server.Scheme != "http" && server.Scheme != "https") ||
		server.Hostname() == "" || server.User != nil || server.Path != "" || server.RawQuery != "" || server.Fragment != "" {
		return nil, errors.New("IPFS endpoint must be an HTTP origin")
	}
	if primary == nil || index == nil {
		return nil, errors.New("IPFS backup requires RustFS and SQLite")
	}
	apps := make(map[string]struct{})
	for _, app := range strings.Split(publicApps, ",") {
		if name := strings.TrimSpace(app); name != "" {
			apps[name] = struct{}{}
		}
	}
	if len(apps) == 0 {
		return nil, errors.New("IPFS backup requires an explicit public demo app allowlist")
	}
	return &IPFSSBOMStore{S3Store: primary, index: index, endpoint: server, apps: apps,
		client: &http.Client{
			Timeout:       5 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		}}, nil
}

func (s *IPFSSBOMStore) PinSBOM(ctx context.Context, app, key string, sbom []byte) error {
	if _, public := s.apps[app]; !public {
		return nil
	}
	if len(sbom) > maxIPFSSBOMBytes {
		return errors.New("SBOM exceeds the IPFS upload limit")
	}
	var body bytes.Buffer
	form := multipart.NewWriter(&body)
	file, err := form.CreateFormFile("file", "sbom.json")
	if err != nil {
		return err
	}
	if _, err := file.Write(sbom); err != nil {
		return err
	}
	if err := form.Close(); err != nil {
		return err
	}
	query := url.Values{"cid-version": {"1"}, "pin": {"true"}}
	response, err := s.call(ctx, "add", query, form.FormDataContentType(), &body)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	payload, err := readKuboResponse(response, 4096)
	if err != nil {
		return err
	}
	var result struct{ Hash string }
	if err := json.Unmarshal(payload, &result); err != nil {
		return err
	}
	if !ipfsCID.MatchString(result.Hash) {
		return errors.New("Kubo returned an invalid SBOM CID")
	}
	_, err = s.index.db.ExecContext(ctx, `
		INSERT INTO sbom_ipfs (storage_key, cid) VALUES (?, ?)
		ON CONFLICT(storage_key) DO UPDATE SET cid = excluded.cid`, key, result.Hash)
	return err
}

func (s *IPFSSBOMStore) GetSBOM(ctx context.Context, key string) ([]byte, error) {
	sbom, primaryErr := s.S3Store.GetSBOM(ctx, key)
	if primaryErr == nil {
		return sbom, nil
	}
	var cid string
	err := s.index.db.QueryRowContext(ctx, "SELECT cid FROM sbom_ipfs WHERE storage_key = ?", key).Scan(&cid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, primaryErr
	}
	if err != nil {
		return nil, errors.Join(primaryErr, err)
	}
	if !ipfsCID.MatchString(cid) {
		return nil, errors.Join(primaryErr, errors.New("stored SBOM CID is invalid"))
	}
	response, err := s.call(ctx, "cat", url.Values{"arg": {cid}}, "", nil)
	if err != nil {
		return nil, errors.Join(primaryErr, err)
	}
	defer response.Body.Close()
	sbom, err = readKuboResponse(response, maxIPFSSBOMBytes)
	if err != nil {
		return nil, errors.Join(primaryErr, err)
	}
	return sbom, nil
}

func readKuboResponse(response *http.Response, maximum int64) ([]byte, error) {
	data, err := io.ReadAll(io.LimitReader(response.Body, maximum+1))
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > maximum {
		return nil, errors.New("Kubo response exceeds the expected limit")
	}
	if message := response.Trailer.Get("X-Stream-Error"); message != "" {
		return nil, fmt.Errorf("Kubo stream error: %s", message)
	}
	return data, nil
}

func (s *IPFSSBOMStore) call(ctx context.Context, method string, query url.Values, contentType string, body io.Reader) (*http.Response, error) {
	target := *s.endpoint
	target.Path = "/api/v0/" + method
	target.RawQuery = query.Encode()
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, target.String(), body)
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	response, err := s.client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		return nil, fmt.Errorf("Kubo %s returned HTTP %d", method, response.StatusCode)
	}
	return response, nil
}
