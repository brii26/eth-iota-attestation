package benchmark

type Dataset struct {
	SchemaVersion int               `json:"schemaVersion"`
	Version       string            `json:"version"`
	CreatedAt     string            `json:"createdAt"`
	Registry      string            `json:"registry,omitempty"`
	Tools         map[string]string `json:"tools"`
	Fixtures      []Fixture         `json:"fixtures"`
}

type Fixture struct {
	App          string `json:"app"`
	Version      string `json:"version"`
	SizeClass    string `json:"sizeClass"`
	Dependencies int    `json:"dependencies"`
	Artifact     File   `json:"artifact"`
	SBOM         File   `json:"sbom"`
	Scan         File   `json:"scan"`
}

type File struct {
	Path   string `json:"path"`
	Bytes  int64  `json:"bytes"`
	SHA256 string `json:"sha256"`
}
