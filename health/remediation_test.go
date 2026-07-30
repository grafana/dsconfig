package health

import "testing"

// TestResolveRemediation_Specificity verifies the most-constrained matching rule
// wins, with graceful fallback to the catalog (RFC §6.5 / ADR-004).
func TestResolveRemediation_Specificity(t *testing.T) {
	t.Cleanup(resetRegistry)
	resetRegistry()

	RegisterRules(
		// Least specific: any auth failure.
		Rule{When: Match{Code: CodeAuthenticationFailed}, Give: Remediation{Summary: "generic auth"}},
		// More specific: auth failure for a particular provider code.
		Rule{When: Match{Code: CodeAuthenticationFailed, ProviderCode: "AADSTS70008"}, Give: Remediation{Summary: "token expired - reauthenticate"}},
		// Different auth type dimension.
		Rule{When: Match{Code: CodeAuthenticationFailed, AuthType: "oauth"}, Give: Remediation{Summary: "check oauth client secret"}},
	)

	t.Run("provider code beats generic", func(t *testing.T) {
		d := Diagnosis{Code: CodeAuthenticationFailed, ProviderCode: "AADSTS70008"}
		equal(t, "summary", resolveRemediation(d).Summary, "token expired - reauthenticate")
	})

	t.Run("auth type rule when no provider match", func(t *testing.T) {
		d := Diagnosis{Code: CodeAuthenticationFailed, Context: Context{AuthType: "oauth"}}
		equal(t, "summary", resolveRemediation(d).Summary, "check oauth client secret")
	})

	t.Run("generic rule when only code matches", func(t *testing.T) {
		d := Diagnosis{Code: CodeAuthenticationFailed}
		equal(t, "summary", resolveRemediation(d).Summary, "generic auth")
	})

	t.Run("falls back to catalog when no rule matches", func(t *testing.T) {
		d := Diagnosis{Code: CodeTLSError}
		equal(t, "summary", resolveRemediation(d).Summary, entryFor(CodeTLSError).Fallback.Summary)
	})
}

// TestResolveRemediation_BodyKindMatch shows rules can key off response shape.
func TestResolveRemediation_BodyKindMatch(t *testing.T) {
	t.Cleanup(resetRegistry)
	resetRegistry()

	RegisterRules(Rule{
		When: Match{Code: CodeAuthenticationFailed, BodyKind: BodyHTML},
		Give: Remediation{Summary: "an auth proxy returned a login page"},
	})

	d := Diagnosis{Code: CodeAuthenticationFailed, BodyKind: BodyHTML}
	equal(t, "summary", resolveRemediation(d).Summary, "an auth proxy returned a login page")
}

func TestApplyDocsBase(t *testing.T) {
	equal(t, "relative joined",
		applyDocsBase(Remediation{DocsURL: "datasources/postgres"}, "https://grafana.com/docs/").DocsURL,
		"https://grafana.com/docs/datasources/postgres")
	equal(t, "absolute untouched",
		applyDocsBase(Remediation{DocsURL: "https://x/y"}, "https://grafana.com/docs").DocsURL,
		"https://x/y")
	equal(t, "no base",
		applyDocsBase(Remediation{DocsURL: "rel"}, "").DocsURL,
		"rel")
}
