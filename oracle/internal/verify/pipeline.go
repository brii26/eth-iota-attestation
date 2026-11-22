package verify

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
)

const GitHubIssuer = "https://token.actions.githubusercontent.com"

var packageName = regexp.MustCompile(`^(?:@[a-z0-9][a-z0-9._-]*/)?[a-z0-9][a-z0-9._-]*$`)
var githubRepository = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9-]*/[a-zA-Z0-9_.-]+$`)

type PipelineVerifier struct {
	token *oidc.IDTokenVerifier
	rules map[string]PipelineRule
}

type PipelineRule struct {
	Repository string `json:"repository"`
	Workflow   string `json:"workflowRef"`
}

func NewPipelineVerifier(ctx context.Context, audience, allowlist string) (*PipelineVerifier, error) {
	if audience == "" || allowlist == "" {
		return nil, errors.New("pipeline audience and app allowlist are required")
	}
	var rules map[string]PipelineRule
	if err := json.Unmarshal([]byte(allowlist), &rules); err != nil || len(rules) == 0 {
		return nil, errors.New("pipeline allowlist must map package names to repository and workflowRef")
	}
	for app, rule := range rules {
		if !packageName.MatchString(app) || !githubRepository.MatchString(rule.Repository) ||
			!strings.HasPrefix(rule.Workflow, rule.Repository+"/.github/workflows/") ||
			!strings.Contains(rule.Workflow, "@refs/heads/") {
			return nil, errors.New("pipeline allowlist contains an invalid package, repository or branch workflow")
		}
	}
	ctx = oidc.ClientContext(ctx, &http.Client{Timeout: 10 * time.Second})
	keys := oidc.NewRemoteKeySet(ctx, GitHubIssuer+"/.well-known/jwks")
	return newPipelineVerifier(audience, keys, rules), nil
}

func newPipelineVerifier(audience string, keys oidc.KeySet, rules map[string]PipelineRule) *PipelineVerifier {
	return &PipelineVerifier{
		token: oidc.NewVerifier(GitHubIssuer, keys, &oidc.Config{
			ClientID: audience, SupportedSigningAlgs: []string{oidc.RS256},
		}),
		rules: rules,
	}
}

func (p *PipelineVerifier) Authorize(ctx context.Context, token, app string) bool {
	rule, allowed := p.rules[app]
	if !allowed || token == "" {
		return false
	}
	verified, err := p.token.Verify(ctx, token)
	if err != nil {
		return false
	}
	var claims struct {
		Repository string `json:"repository"`
		Workflow   string `json:"workflow_ref"`
		Event      string `json:"event_name"`
	}
	if err := verified.Claims(&claims); err != nil {
		return false
	}
	return claims.Repository == rule.Repository && claims.Workflow == rule.Workflow &&
		(claims.Event == "push" || claims.Event == "workflow_dispatch")
}
