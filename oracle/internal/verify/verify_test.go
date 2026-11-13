package verify

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestComputeAndCompareHashes(t *testing.T) {
	data := []byte("hello software artifact world")
	_, hexHash := ComputeSHA256(data)

	if !CompareHashes(hexHash, hexHash) {
		t.Errorf("expected identical hashes to match")
	}
	if !CompareHashes("0x"+hexHash, hexHash) {
		t.Errorf("expected hash with 0x prefix to match")
	}
	if CompareHashes(hexHash, "badhash") {
		t.Errorf("expected bad hash to fail comparison")
	}
}

func TestValidateGrypeFreshness(t *testing.T) {
	now := time.Date(2026, 11, 13, 10, 0, 0, 0, time.UTC)
	maxAge := 24 * time.Hour

	// 1. Fresh scan (built 2 hours ago)
	freshJSON := `{"descriptor":{"db":{"built":"2026-11-13T08:00:00Z"}}}`
	builtTime, err := ValidateGrypeFreshness([]byte(freshJSON), maxAge, now)
	if err != nil {
		t.Fatalf("expected fresh scan to pass, got err: %v", err)
	}
	if builtTime.Hour() != 8 {
		t.Errorf("expected hour 8, got %d", builtTime.Hour())
	}

	// 2. Stale scan (built 3 days ago)
	staleJSON := `{"descriptor":{"db":{"built":"2026-11-10T08:00:00Z"}}}`
	_, err = ValidateGrypeFreshness([]byte(staleJSON), maxAge, now)
	if err == nil {
		t.Fatalf("expected stale scan to fail validation, got nil")
	}

	// 3. Missing DB timestamp
	missingJSON := `{"descriptor":{"db":{}}}`
	_, err = ValidateGrypeFreshness([]byte(missingJSON), maxAge, now)
	if err == nil {
		t.Fatalf("expected missing timestamp to fail, got nil")
	}
}

func TestVerifyPipeline(t *testing.T) {
	if !VerifyPipeline("valid-token", "my-app") {
		t.Errorf("expected valid token to pass")
	}
	if VerifyPipeline("", "my-app") {
		t.Errorf("expected empty token to fail")
	}
	if VerifyPipeline("unauthorized", "my-app") {
		t.Errorf("expected unauthorized token to fail")
	}
	if VerifyPipeline("valid-token", "../../etc/passwd") {
		t.Errorf("expected path traversal app name to fail")
	}
}

func TestVerdaccioClient(t *testing.T) {
	mockTarball := []byte("fake-tarball-bytes")

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/demo-lib":
			doc := map[string]interface{}{
				"versions": map[string]interface{}{
					"1.0.0": map[string]interface{}{
						"dist": map[string]interface{}{
							"tarball": "http://" + r.Host + "/demo-lib/-/demo-lib-1.0.0.tgz",
						},
					},
				},
			}
			json.NewEncoder(w).Encode(doc)
		case "/demo-lib/-/demo-lib-1.0.0.tgz":
			w.Write(mockTarball)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewVerdaccioClient(server.URL)

	// Fetch existing
	data, err := client.FetchArtifact(context.Background(), "demo-lib", "1.0.0")
	if err != nil {
		t.Fatalf("failed to fetch artifact: %v", err)
	}
	if string(data) != string(mockTarball) {
		t.Errorf("artifact content mismatch: %q vs %q", string(data), string(mockTarball))
	}

	// Fetch non-existing
	_, err = client.FetchArtifact(context.Background(), "non-existent", "1.0.0")
	if err == nil {
		t.Fatalf("expected 404 error for non-existent package")
	}
}
