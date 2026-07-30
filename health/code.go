package health

import "github.com/grafana/grafana-plugin-sdk-go/backend"

// Code is the stable, coarse, public error vocabulary. Codes are additive and
// never repurposed (see RFC §6.2 / ADR-006).
type Code string

const (
	CodeInvalidConfiguration Code = "INVALID_CONFIGURATION"
	CodeAuthenticationFailed Code = "AUTHENTICATION_FAILED"
	CodePermissionDenied     Code = "PERMISSION_DENIED"
	CodeHostUnreachable      Code = "HOST_UNREACHABLE"
	CodeConnectionTimeout    Code = "CONNECTION_TIMEOUT"
	CodeTLSError             Code = "TLS_ERROR"
	CodeNotFound             Code = "NOT_FOUND"
	CodeRateLimited          Code = "RATE_LIMITED"
	CodeQuotaExceeded        Code = "QUOTA_EXCEEDED"
	CodeUpstreamError        Code = "UPSTREAM_ERROR"
	CodeUnsupportedVersion   Code = "UNSUPPORTED_VERSION"
	CodeQueryValidation      Code = "QUERY_VALIDATION_FAILED"
	// CodeUnexpectedResponse is the fallback for "the server answered, but not
	// in the format we expected" — an HTML error/login page where JSON was
	// expected, or a JSON error body we could not interpret (see RFC §6.4a).
	CodeUnexpectedResponse Code = "UNEXPECTED_RESPONSE"
	CodeUnknown            Code = "UNKNOWN"
)

// catalogEntry holds the canonical copy and default attribution for a Code.
type catalogEntry struct {
	// Headline is the one-line, user-facing summary of the category.
	Headline string
	// Fallback is the generic remediation used when no Rule matches.
	Fallback Remediation
	// Source is the default ErrorSource for the category.
	Source backend.ErrorSource
}

// catalog maps every Code to its canonical headline, fallback remediation and
// default error source. It is the single source of truth for backend copy; the
// frontend localizes off the Code (ADR-006).
var catalog = map[Code]catalogEntry{
	CodeInvalidConfiguration: {
		Headline: "The data source configuration is incomplete or invalid.",
		Fallback: Remediation{Summary: "Check the highlighted settings and save again."},
		Source:   backend.ErrorSourcePlugin,
	},
	CodeAuthenticationFailed: {
		Headline: "Authentication failed.",
		Fallback: Remediation{Summary: "Verify the credentials in the data source settings."},
		Source:   backend.ErrorSourceDownstream,
	},
	CodePermissionDenied: {
		Headline: "Access was denied.",
		Fallback: Remediation{Summary: "Ensure the account has permission for this resource."},
		Source:   backend.ErrorSourceDownstream,
	},
	CodeHostUnreachable: {
		Headline: "Could not reach the server.",
		Fallback: Remediation{Summary: "Check the URL/host and network connectivity."},
		Source:   backend.ErrorSourceDownstream,
	},
	CodeConnectionTimeout: {
		Headline: "The connection timed out.",
		Fallback: Remediation{Summary: "The server may be slow or a firewall may be dropping the connection — check the network/firewall and that the service is running."},
		Source:   backend.ErrorSourceDownstream,
	},
	CodeTLSError: {
		Headline: "A TLS/certificate error occurred.",
		Fallback: Remediation{Summary: "Verify the TLS settings and CA certificate."},
		Source:   backend.ErrorSourceDownstream,
	},
	CodeNotFound: {
		Headline: "The target resource was not found.",
		Fallback: Remediation{Summary: "Check the database/endpoint name."},
		Source:   backend.ErrorSourceDownstream,
	},
	CodeRateLimited: {
		Headline: "The service is rate-limiting requests.",
		Fallback: Remediation{Summary: "Retry later or reduce request volume."},
		Source:   backend.ErrorSourceDownstream,
	},
	CodeQuotaExceeded: {
		Headline: "A service quota was exceeded.",
		Fallback: Remediation{Summary: "Request a quota increase or reduce usage."},
		Source:   backend.ErrorSourceDownstream,
	},
	CodeUpstreamError: {
		Headline: "The service reported an internal error.",
		Fallback: Remediation{Summary: "Retry later or check the service status."},
		Source:   backend.ErrorSourceDownstream,
	},
	CodeUnsupportedVersion: {
		Headline: "The server version is not supported.",
		Fallback: Remediation{Summary: "Upgrade the server to a supported version."},
		Source:   backend.ErrorSourceDownstream,
	},
	CodeQueryValidation: {
		Headline: "The health-check query failed.",
		Fallback: Remediation{Summary: "Check that the account can run the test query."},
		Source:   backend.ErrorSourceDownstream,
	},
	CodeUnexpectedResponse: {
		Headline: "The server returned an unexpected response.",
		Fallback: Remediation{Summary: "Check that the URL points at the data source API and not a proxy, login page or web server."},
		Source:   backend.ErrorSourceDownstream,
	},
	CodeUnknown: {
		Headline: "An unexpected error occurred.",
		Fallback: Remediation{Summary: "See the data source logs for details."},
		Source:   backend.ErrorSourcePlugin,
	},
}

// entryFor returns the catalog entry for code, falling back to CodeUnknown for
// any code that is somehow not registered (defensive — never panics).
func entryFor(code Code) catalogEntry {
	if e, ok := catalog[code]; ok {
		return e
	}
	return catalog[CodeUnknown]
}
