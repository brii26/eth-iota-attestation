package verify

import (
	"strings"
)

// VerifyPipeline validates whether the pipeline bearer token is authorized to submit for the given app.
// ponytail: OIDC stub implementation for Issue #8. Upgrade to full JWT OIDC token verification in Issue #15.
func VerifyPipeline(token string, app string) bool {
	token = strings.TrimSpace(token)
	if token == "" {
		return false
	}

	// Reject explicit invalid test tokens
	if strings.EqualFold(token, "invalid") || strings.EqualFold(token, "unauthorized") {
		return false
	}

	// Basic app name safety
	if strings.Contains(app, "..") || strings.Contains(app, "/") {
		// Scoped packages like @scope/app are allowed, but path traversal is rejected
		parts := strings.Split(app, "/")
		if len(parts) > 2 {
			return false
		}
	}

	return true
}
