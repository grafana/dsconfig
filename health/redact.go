package health

import (
	"regexp"
	"strings"
)

// maxVerboseLen bounds the redacted verbose detail so a huge upstream body can
// never bloat logs or the response (RFC §11).
const maxVerboseLen = 2048

var (
	// credentials embedded in a URL: scheme://user:pass@host
	reURLCreds = regexp.MustCompile(`([a-zA-Z][a-zA-Z0-9+.-]*://)([^:/?#@\s]+):([^@/?#\s]+)@`)

	// key=value / key: value / "key":"value" for sensitive keys. Captures the
	// key + separator so it can be preserved while the value is masked.
	reSensitiveKV = regexp.MustCompile(`(?i)("?\b(?:password|passwd|pwd|secret|secret[_-]?key|access[_-]?key|api[_-]?key|apikey|token|authorization|auth|bearer|client[_-]?secret|private[_-]?key)\b"?\s*[:=]\s*)("?)([^"\s,;&]+)("?)`)

	// Authorization-style header values: "Bearer <token>", "Basic <token>".
	reBearer = regexp.MustCompile(`(?i)\b(bearer|basic)\s+[A-Za-z0-9._~+/=-]+`)
)

const redactedMask = "[REDACTED]"

// redactor masks secrets out of free-text before it is surfaced anywhere
// (verbose detail, logs, spans). It is best-effort defense-in-depth — the
// primary control is "classify, don't echo" (RFC §6.4a, §15). Callers must not
// rely on it to make an arbitrary upstream body safe.
type redactor struct{}

func (redactor) redact(s string) string {
	if s == "" {
		return s
	}
	s = reURLCreds.ReplaceAllString(s, "$1$2:"+redactedMask+"@")
	// Bearer/Basic must run before the key=value pass, otherwise the latter
	// masks the "Bearer" scheme word and leaves the token exposed.
	s = reBearer.ReplaceAllString(s, "$1 "+redactedMask)
	s = reSensitiveKV.ReplaceAllString(s, "$1$2"+redactedMask+"$4")
	return truncate(s, maxVerboseLen)
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "…(truncated)"
}

// redact masks secrets in s using the default redactor.
func redact(s string) string { return redactor{}.redact(s) }

// htmlTitle extracts the contents of the first <title> element, trimmed. It is
// used to build a safe, compact summary of an HTML body without dumping the page.
var reTitle = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)

func htmlTitle(body []byte) string {
	m := reTitle.FindSubmatch(body)
	if len(m) < 2 {
		return ""
	}
	return strings.TrimSpace(collapseSpaces(string(m[1])))
}

var reSpaces = regexp.MustCompile(`\s+`)

func collapseSpaces(s string) string { return reSpaces.ReplaceAllString(s, " ") }
