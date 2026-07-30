package health

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
)

// maxBodyParse bounds how much of an upstream body is ever parsed/sniffed. The
// body is untrusted; oversized input degrades to a fallback code rather than
// being fully read (RFC §6.4a, §11).
const maxBodyParse = 64 * 1024

// sniffBodyKind inspects Content-Type and the leading bytes of a body to decide
// whether it is JSON, HTML or plain text.
func sniffBodyKind(contentType string, body []byte) BodyKind {
	ct := strings.ToLower(contentType)
	switch {
	case strings.Contains(ct, "application/json"), strings.Contains(ct, "+json"):
		return BodyJSON
	case strings.Contains(ct, "text/html"), strings.Contains(ct, "application/xhtml"):
		return BodyHTML
	}
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) > maxBodyParse {
		trimmed = trimmed[:maxBodyParse]
	}
	lower := bytes.ToLower(trimmed)
	switch {
	case bytes.HasPrefix(lower, []byte("<!doctype")), bytes.HasPrefix(lower, []byte("<html")),
		bytes.HasPrefix(lower, []byte("<?xml")), bytes.HasPrefix(lower, []byte("<")):
		return BodyHTML
	case bytes.HasPrefix(trimmed, []byte("{")), bytes.HasPrefix(trimmed, []byte("[")):
		return BodyJSON
	default:
		return BodyText
	}
}

// ClassifyHTTPResponse inspects an HTTP response (status + Content-Type + body)
// and produces a Diagnosis for the "server answered, but not as expected" cases
// (RFC §6.4a). HTTP families call this before unmarshalling so the status and
// headers are available. It returns ok=false when the response looks like a
// normal success or a body it should not opine on.
func ClassifyHTTPResponse(resp *http.Response, body []byte, ctx Context) (Diagnosis, bool) {
	if resp == nil {
		return Diagnosis{}, false
	}
	contentType := resp.Header.Get("Content-Type")
	kind := sniffBodyKind(contentType, body)

	d := Diagnosis{
		HTTPStatus:  resp.StatusCode,
		ContentType: contentType,
		BodyKind:    kind,
		Context:     ctx,
	}

	switch kind {
	case BodyHTML:
		d.Code = htmlCodeFor(resp.StatusCode, ctx)
		return d, true
	case BodyJSON:
		// A 2xx JSON body is a normal success — not ours to classify.
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return Diagnosis{}, false
		}
		_, providerCode, _ := ExtractJSONError(body)
		d.ProviderCode = providerCode
		d.Code = codeFromHTTPStatus(resp.StatusCode)
		return d, true
	default:
		if resp.StatusCode >= 200 && resp.StatusCode < 300 {
			return Diagnosis{}, false
		}
		d.Code = codeFromHTTPStatus(resp.StatusCode)
		return d, true
	}
}

// htmlCodeFor maps an HTML response to a Code by status + context (RFC §6.4a).
func htmlCodeFor(status int, ctx Context) Code {
	switch {
	case status == http.StatusBadGateway, status == http.StatusServiceUnavailable, status == http.StatusGatewayTimeout:
		return CodeUpstreamError
	case status == http.StatusUnauthorized:
		return CodeAuthenticationFailed
	case status == http.StatusForbidden:
		return CodePermissionDenied
	case status >= 200 && status < 300:
		// 200 OK with HTML: a login/SSO page behind a proxy, or the URL points
		// at a web root instead of the API.
		if np := strings.ToLower(ctx.NetworkPath); np == "pdc" || np == "proxy" {
			return CodeAuthenticationFailed
		}
		return CodeInvalidConfiguration
	default:
		return CodeUnexpectedResponse
	}
}

// codeFromHTTPStatus maps a status code to a Code for non-HTML error responses.
func codeFromHTTPStatus(status int) Code {
	switch {
	case status == http.StatusUnauthorized:
		return CodeAuthenticationFailed
	case status == http.StatusForbidden:
		return CodePermissionDenied
	case status == http.StatusNotFound:
		return CodeNotFound
	case status == http.StatusTooManyRequests:
		return CodeRateLimited
	case status == http.StatusBadGateway, status == http.StatusServiceUnavailable, status == http.StatusGatewayTimeout:
		return CodeUpstreamError
	case status >= 500:
		return CodeUpstreamError
	case status >= 400:
		return CodeInvalidConfiguration
	default:
		return CodeUnexpectedResponse
	}
}

// jsonErrorKeys are the common free-text message keys, checked case-insensitively.
var jsonErrorKeys = []string{"error_description", "message", "detail", "reason", "error", "title"}

// jsonCodeKeys are the common machine-code keys.
var jsonCodeKeys = []string{"errorType", "error_type", "code", "status", "type"}

// ExtractJSONError is a best-effort, never-authoritative reader for arbitrary
// JSON error envelopes (RFC §6.4a). It tolerates error-as-string|object|array,
// numeric-or-string codes, nesting and missing fields; parsing is size-bounded
// and it never panics. It returns ok=false when nothing useful parses.
//
// The extracted message is a hint for the redacted verbose detail only — it must
// never become the surfaced Message ("classify, don't echo").
func ExtractJSONError(body []byte) (msg string, providerCode string, ok bool) {
	if len(body) == 0 || len(body) > maxBodyParse {
		return "", "", false
	}
	var root any
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.UseNumber()
	if err := dec.Decode(&root); err != nil {
		return "", "", false
	}

	switch v := root.(type) {
	case map[string]any:
		msg, providerCode = extractFromObject(v)
	case []any:
		if len(v) > 0 {
			if obj, isObj := v[0].(map[string]any); isObj {
				msg, providerCode = extractFromObject(obj)
			}
		}
	}
	return msg, providerCode, msg != "" || providerCode != ""
}

func extractFromObject(obj map[string]any) (msg string, providerCode string) {
	// A nested error object (e.g. Elasticsearch error.{type,reason}) is
	// authoritative — recurse and use its result.
	if e, present := lookupCI(obj, "error"); present {
		if ev, isObj := e.(map[string]any); isObj {
			if m, pc := extractFromObject(ev); m != "" || pc != "" {
				return m, pc
			}
		}
	}
	// errors: [ { ... } ]  (JSON:API / GraphQL): use the first entry.
	if arr, present := lookupCI(obj, "errors"); present {
		if list, isArr := arr.([]any); isArr && len(list) > 0 {
			if first, isObj := list[0].(map[string]any); isObj {
				msg, providerCode = extractFromObject(first)
			}
		}
	}
	// root_cause[0].reason (Elasticsearch).
	if rc, present := lookupCI(obj, "root_cause"); present && msg == "" {
		if list, isArr := rc.([]any); isArr && len(list) > 0 {
			if first, isObj := list[0].(map[string]any); isObj {
				if m, _ := extractFromObject(first); m != "" {
					msg = m
				}
			}
		}
	}
	// extensions.code (GraphQL).
	if ext, present := lookupCI(obj, "extensions"); present && providerCode == "" {
		if extObj, isObj := ext.(map[string]any); isObj {
			if c := firstStringValue(extObj, []string{"code"}); c != "" {
				providerCode = c
			}
		}
	}

	// Scalar fallbacks. jsonErrorKeys is ordered so the detailed human message
	// (error_description) wins over a bare code-like `error` string.
	if msg == "" {
		msg = firstStringValue(obj, jsonErrorKeys)
	}
	if providerCode == "" {
		providerCode = firstStringValue(obj, jsonCodeKeys)
	}
	return msg, providerCode
}

// lookupCI does a case-insensitive key lookup.
func lookupCI(obj map[string]any, key string) (any, bool) {
	if v, ok := obj[key]; ok {
		return v, true
	}
	for k, v := range obj {
		if strings.EqualFold(k, key) {
			return v, true
		}
	}
	return nil, false
}

// firstStringValue returns the first key whose value renders as a non-empty
// scalar string (string or json.Number).
func firstStringValue(obj map[string]any, keys []string) string {
	for _, k := range keys {
		if v, ok := lookupCI(obj, k); ok {
			switch s := v.(type) {
			case string:
				if strings.TrimSpace(s) != "" {
					return s
				}
			case json.Number:
				return s.String()
			}
		}
	}
	return ""
}
