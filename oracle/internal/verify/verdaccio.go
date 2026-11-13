package verify

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// ErrArtifactNotFound is returned when Verdaccio does not have the requested app/version.
var ErrArtifactNotFound = errors.New("artifact not found in registry")

// VerdaccioClient fetches artifacts from the Verdaccio npm registry.
type VerdaccioClient struct {
	baseURL    string
	httpClient *http.Client
}

// NewVerdaccioClient constructs a client pointing to the Verdaccio registry URL.
func NewVerdaccioClient(baseURL string) *VerdaccioClient {
	if baseURL == "" {
		baseURL = "http://127.0.0.1:4873"
	}
	return &VerdaccioClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

type npmPackageDoc struct {
	Versions map[string]struct {
		Dist struct {
			Tarball string `json:"tarball"`
		} `json:"dist"`
	} `json:"versions"`
}

// FetchArtifact downloads the raw package tarball bytes for an app version.
func (c *VerdaccioClient) FetchArtifact(ctx context.Context, app, version string) ([]byte, error) {
	// Step 1: Query package metadata to resolve exact tarball URL
	pkgURL := fmt.Sprintf("%s/%s", c.baseURL, url.PathEscape(app))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, pkgURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create package metadata request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("connect to verdaccio: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrArtifactNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("verdaccio metadata returned status %d", resp.StatusCode)
	}

	var doc npmPackageDoc
	if err := json.NewDecoder(resp.Body).Decode(&doc); err != nil {
		return nil, fmt.Errorf("decode verdaccio package doc: %w", err)
	}

	verData, ok := doc.Versions[version]
	if !ok || verData.Dist.Tarball == "" {
		return nil, fmt.Errorf("%w: version %s not in package document", ErrArtifactNotFound, version)
	}

	tarballURL := verData.Dist.Tarball

	// Step 2: Download tarball bytes
	tbReq, err := http.NewRequestWithContext(ctx, http.MethodGet, tarballURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create tarball request: %w", err)
	}

	tbResp, err := c.httpClient.Do(tbReq)
	if err != nil {
		return nil, fmt.Errorf("download artifact from verdaccio: %w", err)
	}
	defer tbResp.Body.Close()

	if tbResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("verdaccio tarball download returned status %d", tbResp.StatusCode)
	}

	data, err := io.ReadAll(tbResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read artifact tarball: %w", err)
	}

	return data, nil
}
