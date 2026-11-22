package main

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/brii26/eth-iota-attestation/oracle/internal/benchmark"
)

type npmManifest struct {
	Name                string            `json:"name"`
	Version             string            `json:"version"`
	Main                string            `json:"main"`
	Dependencies        map[string]string `json:"dependencies,omitempty"`
	BundledDependencies []string          `json:"bundledDependencies,omitempty"`
}

var sizeClasses = []struct {
	name         string
	dependencies int
}{{"small", 2}, {"medium", 24}, {"large", 96}}

func main() {
	output := flag.String("output", "", "directory in which to create a new dataset run")
	registry := flag.String("registry", "", "optional Verdaccio URL for publication")
	userConfig := flag.String("npm-userconfig", "", "optional npm config file for registry authentication")
	workers := flag.Int("scan-workers", 1, "maximum simultaneous Grype scans")
	flag.Parse()
	if *output == "" {
		log.Fatal("-output is required")
	}
	if *workers < 1 || *workers > 10 {
		log.Fatal("-scan-workers must be between 1 and 10")
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx, *output, *registry, *userConfig, *workers); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context, output, registry, userConfig string, workers int) error {
	if registry != "" {
		endpoint, err := url.Parse(registry)
		if err != nil || (endpoint.Scheme != "http" && endpoint.Scheme != "https") ||
			endpoint.Host == "" || endpoint.User != nil || endpoint.RawQuery != "" || endpoint.Fragment != "" {
			return errors.New("registry must be an HTTP(S) URL without embedded credentials")
		}
	}
	if userConfig != "" {
		var err error
		userConfig, err = filepath.Abs(userConfig)
		if err != nil {
			return err
		}
	}
	output, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(output, 0o755); err != nil {
		return err
	}
	directory, err := os.MkdirTemp(output, "run-")
	if err != nil {
		return err
	}
	now := time.Now().UTC()
	dataset := benchmark.Dataset{
		SchemaVersion: 1, Version: fmt.Sprintf("1.0.0-bench.%d.%s", now.Unix(), hex.EncodeToString(randBytes())),
		CreatedAt: now.Format(time.RFC3339), Registry: registry, Tools: make(map[string]string),
	}
	for _, tool := range []string{"node", "npm", "syft", "grype"} {
		version, err := command(ctx, directory, tool, "--version")
		if err != nil {
			return err
		}
		dataset.Tools[tool] = strings.TrimSpace(string(version))
	}
	for index := 0; index < 10; index++ {
		size := sizeClasses[index%len(sizeClasses)]
		name := fmt.Sprintf("bench-%02d", index+1)
		fixture, err := generate(ctx, directory, name, dataset.Version, size.name, size.dependencies)
		if err != nil {
			return fmt.Errorf("generate %s: %w", name, err)
		}
		dataset.Fixtures = append(dataset.Fixtures, fixture)
	}
	if err := scanFixtures(ctx, directory, dataset.Fixtures, workers); err != nil {
		return err
	}
	if registry != "" {
		for _, fixture := range dataset.Fixtures {
			args := []string{"publish", filepath.Join(directory, fixture.Artifact.Path), "--registry", registry, "--tag", "bench", "--ignore-scripts"}
			if userConfig != "" {
				args = append(args, "--userconfig", userConfig)
			}
			if _, err := command(ctx, directory, "npm", args...); err != nil {
				return fmt.Errorf("publish %s: %w", fixture.App, err)
			}
		}
	}
	manifest := filepath.Join(directory, "manifest.json")
	if err := writeJSON(manifest, dataset); err != nil {
		return err
	}
	fmt.Println(manifest)
	return nil
}

func randBytes() []byte {
	value := make([]byte, 6)
	rand.Read(value)
	return value
}

func generate(ctx context.Context, output, name, version, class string, count int) (benchmark.Fixture, error) {
	app := "@attestation/" + name
	work, err := os.MkdirTemp("", "fixture-")
	if err != nil {
		return benchmark.Fixture{}, err
	}
	defer os.RemoveAll(work)
	manifest := npmManifest{Name: app, Version: version, Main: "index.js", Dependencies: make(map[string]string)}
	for index := 1; index <= count; index++ {
		dependency := fmt.Sprintf("attestation-fixture-lib-%03d", index)
		manifest.Dependencies[dependency] = "1.0.0"
		manifest.BundledDependencies = append(manifest.BundledDependencies, dependency)
		path := filepath.Join(work, "node_modules", dependency)
		if err := writePackage(path, npmManifest{Name: dependency, Version: "1.0.0", Main: "index.js"}); err != nil {
			return benchmark.Fixture{}, err
		}
	}
	if err := writePackage(work, manifest); err != nil {
		return benchmark.Fixture{}, err
	}
	packed, err := command(ctx, work, "npm", "pack", "--ignore-scripts", "--json", "--pack-destination", output)
	if err != nil {
		return benchmark.Fixture{}, err
	}
	var packs []struct {
		Filename string `json:"filename"`
	}
	if err := json.Unmarshal(packed, &packs); err != nil || len(packs) != 1 || packs[0].Filename == "" {
		return benchmark.Fixture{}, errors.New("npm pack did not report one archive")
	}
	archive := packs[0].Filename
	sbom, scan := name+".sbom.json", name+".scan.json"
	scanInput := filepath.Join(work, "unpacked")
	if err := os.Mkdir(scanInput, 0o755); err != nil {
		return benchmark.Fixture{}, err
	}
	if _, err := command(ctx, work, "tar", "-xzf", filepath.Join(output, archive), "-C", scanInput); err != nil {
		return benchmark.Fixture{}, err
	}
	if _, err := command(ctx, output, "syft", "dir:"+filepath.Join(scanInput, "package"), "--select-catalogers", "+javascript-package-cataloger",
		"--source-name", app, "--source-version", version, "-o", "cyclonedx-json="+sbom); err != nil {
		return benchmark.Fixture{}, err
	}
	fixture := benchmark.Fixture{App: app, Version: version, SizeClass: class, Dependencies: count, Scan: benchmark.File{Path: scan}}
	for _, item := range []struct {
		path string
		out  *benchmark.File
	}{{archive, &fixture.Artifact}, {sbom, &fixture.SBOM}} {
		file, err := inspectFile(output, item.path)
		if err != nil {
			return benchmark.Fixture{}, err
		}
		*item.out = file
	}
	return fixture, nil
}

func scanFixtures(ctx context.Context, directory string, fixtures []benchmark.Fixture, workers int) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	jobs := make(chan int, len(fixtures))
	for index := range fixtures {
		jobs <- index
	}
	close(jobs)
	var work sync.WaitGroup
	var first sync.Once
	var failure error
	for range workers {
		work.Add(1)
		go func() {
			defer work.Done()
			for index := range jobs {
				if ctx.Err() != nil {
					return
				}
				fixture := &fixtures[index]
				scan, err := scanFixture(ctx, directory, *fixture)
				if err != nil {
					first.Do(func() {
						failure = fmt.Errorf("scan %s: %w", fixture.App, err)
						cancel()
					})
					return
				}
				fixture.Scan = scan
			}
		}()
	}
	work.Wait()
	return errors.Join(failure, ctx.Err())
}

func scanFixture(ctx context.Context, directory string, fixture benchmark.Fixture) (benchmark.File, error) {
	report, err := command(ctx, directory, "grype", "sbom:"+fixture.SBOM.Path, "-o", "json")
	if err != nil {
		return benchmark.File{}, err
	}
	if !json.Valid(report) {
		return benchmark.File{}, errors.New("Grype did not produce valid JSON")
	}
	if err := os.WriteFile(filepath.Join(directory, fixture.Scan.Path), report, 0o644); err != nil {
		return benchmark.File{}, err
	}
	return inspectFile(directory, fixture.Scan.Path)
}

func inspectFile(directory, name string) (benchmark.File, error) {
	file, err := os.Open(filepath.Join(directory, name))
	if err != nil {
		return benchmark.File{}, err
	}
	defer file.Close()
	hash := sha256.New()
	size, err := io.Copy(hash, file)
	return benchmark.File{Path: name, Bytes: size, SHA256: hex.EncodeToString(hash.Sum(nil))}, err
}

func writePackage(path string, manifest npmManifest) error {
	if err := os.MkdirAll(path, 0o755); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(path, "package.json"), manifest); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(path, "index.js"), []byte("export const name = "+fmt.Sprintf("%q", manifest.Name)+";\n"), 0o644)
}

func writeJSON(path string, value any) error {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}

func command(ctx context.Context, dir, program string, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, program, args...)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	data, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", program, err)
	}
	return data, nil
}
