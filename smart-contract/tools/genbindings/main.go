// Command genbindings generates Go bindings for AttestationRegistry from the
// Hardhat build artifact. Run `npm run build` in smart-contract first.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi/abigen"
)

func main() {
	artifact := flag.String("artifact", "../artifacts/contracts/AttestationRegistry.sol/AttestationRegistry.json", "Hardhat artifact path")
	out := flag.String("out", "registry.go", "output Go file")
	pkg := flag.String("pkg", "registry", "Go package name")
	flag.Parse()

	if err := run(*artifact, *out, *pkg); err != nil {
		fmt.Fprintln(os.Stderr, "genbindings:", err)
		os.Exit(1)
	}
}

func run(artifactPath, outPath, pkg string) error {
	raw, err := os.ReadFile(artifactPath)
	if err != nil {
		return fmt.Errorf("read artifact (run npm run build first): %w", err)
	}
	var artifact struct {
		Name     string          `json:"contractName"`
		ABI      json.RawMessage `json:"abi"`
		Bytecode string          `json:"bytecode"`
	}
	if err := json.Unmarshal(raw, &artifact); err != nil {
		return fmt.Errorf("parse artifact: %w", err)
	}
	if artifact.Name == "" || len(artifact.ABI) == 0 || !strings.HasPrefix(artifact.Bytecode, "0x") {
		return fmt.Errorf("artifact %s has no contract name, ABI, or bytecode", artifactPath)
	}
	code, err := abigen.Bind(
		[]string{artifact.Name}, []string{string(artifact.ABI)}, []string{artifact.Bytecode},
		nil, pkg, nil, nil,
	)
	if err != nil {
		return fmt.Errorf("generate bindings: %w", err)
	}
	return os.WriteFile(outPath, []byte(code), 0o644)
}
