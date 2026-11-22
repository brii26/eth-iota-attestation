package verify

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"testing"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v4"
)

func TestGitHubPipelineAuthorization(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	foreign, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	const repo = "brii26/eth-iota-attestation"
	const workflow = repo + "/.github/workflows/attest.yml@refs/heads/main"
	verifier := newPipelineVerifier("attestation-oracle", &oidc.StaticKeySet{PublicKeys: []crypto.PublicKey{&key.PublicKey}},
		map[string]PipelineRule{"@attestation/demo": {Repository: repo, Workflow: workflow}})
	for _, test := range []struct {
		name  string
		claim string
		value any
		key   *rsa.PrivateKey
		app   string
		ok    bool
	}{
		{name: "authorized", key: key, ok: true},
		{name: "wrong issuer", claim: "iss", value: "https://example.test", key: key},
		{name: "wrong audience", claim: "aud", value: "different-audience", key: key},
		{name: "expired", claim: "exp", value: time.Now().Add(-time.Minute).Unix(), key: key},
		{name: "incorrect signing key", key: foreign},
		{name: "wrong repository", claim: "repository", value: "example/untrusted", key: key},
		{name: "wrong workflow", claim: "workflow_ref", value: repo + "/.github/workflows/other.yml@refs/heads/main", key: key},
		{name: "pull request", claim: "event_name", value: "pull_request", key: key},
		{name: "unlisted app", key: key, app: "other-app"},
	} {
		t.Run(test.name, func(t *testing.T) {
			claims := jwt.MapClaims{
				"iss": GitHubIssuer, "aud": "attestation-oracle", "exp": time.Now().Add(time.Minute).Unix(),
				"repository": repo, "workflow_ref": workflow, "event_name": "push",
			}
			if test.claim != "" {
				claims[test.claim] = test.value
			}
			token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(test.key)
			if err != nil {
				t.Fatal(err)
			}
			app := test.app
			if app == "" {
				app = "@attestation/demo"
			}
			if got := verifier.Authorize(t.Context(), token, app); got != test.ok {
				t.Fatalf("authorization = %t, want %t", got, test.ok)
			}
		})
	}
}
