package health

import (
	"net/http"
	"testing"
)

func mkResp(status int, contentType string) *http.Response {
	h := http.Header{}
	if contentType != "" {
		h.Set("Content-Type", contentType)
	}
	return &http.Response{StatusCode: status, Header: h}
}

// TestClassifyHTTPResponse_HTML covers the HTML interception table (RFC §6.4a).
func TestClassifyHTTPResponse_HTML(t *testing.T) {
	cases := []struct {
		name     string
		status   int
		ct       string
		body     string
		ctx      Context
		wantCode Code
	}{
		{
			name:     "502 gateway html",
			status:   http.StatusBadGateway,
			ct:       "text/html",
			body:     "<html><body>502 Bad Gateway</body></html>",
			wantCode: CodeUpstreamError,
		},
		{
			name:     "401 html",
			status:   http.StatusUnauthorized,
			ct:       "text/html",
			body:     "<html><title>Login</title></html>",
			wantCode: CodeAuthenticationFailed,
		},
		{
			name:     "403 waf block page",
			status:   http.StatusForbidden,
			ct:       "text/html",
			body:     "<html><title>Access Denied</title></html>",
			wantCode: CodePermissionDenied,
		},
		{
			name:     "200 html sso behind proxy",
			status:   http.StatusOK,
			ct:       "text/html",
			body:     "<html><title>Sign in - Okta</title></html>",
			ctx:      Context{NetworkPath: "proxy"},
			wantCode: CodeAuthenticationFailed,
		},
		{
			name:     "200 html wrong url direct",
			status:   http.StatusOK,
			ct:       "text/html",
			body:     "<!doctype html><title>Welcome to nginx</title>",
			ctx:      Context{NetworkPath: "direct"},
			wantCode: CodeInvalidConfiguration,
		},
		{
			name:     "html sniffed without content-type",
			status:   http.StatusOK,
			ct:       "",
			body:     "  <!DOCTYPE html><html></html>",
			wantCode: CodeInvalidConfiguration,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d, ok := ClassifyHTTPResponse(mkResp(tc.status, tc.ct), []byte(tc.body), tc.ctx)
			equal(t, "ok", ok, true)
			equal(t, "code", d.Code, tc.wantCode)
			equal(t, "bodyKind", d.BodyKind, BodyHTML)
			equal(t, "httpStatus", d.HTTPStatus, tc.status)
		})
	}
}

// TestClassifyHTTPResponse_JSONAndSuccess covers JSON error envelopes and the
// "do not opine on a 2xx" behaviour.
func TestClassifyHTTPResponse_JSON(t *testing.T) {
	t.Run("401 json envelope yields auth + provider code", func(t *testing.T) {
		body := `{"status":"error","errorType":"unauthorized","error":"token expired"}`
		d, ok := ClassifyHTTPResponse(mkResp(http.StatusUnauthorized, "application/json"), []byte(body), Context{})
		equal(t, "ok", ok, true)
		equal(t, "code", d.Code, CodeAuthenticationFailed)
		equal(t, "bodyKind", d.BodyKind, BodyJSON)
		equal(t, "providerCode", d.ProviderCode, "unauthorized")
	})

	t.Run("2xx json is success, not classified", func(t *testing.T) {
		_, ok := ClassifyHTTPResponse(mkResp(http.StatusOK, "application/json"), []byte(`{"ok":true}`), Context{})
		equal(t, "ok", ok, false)
	})

	t.Run("429 json maps to rate limited", func(t *testing.T) {
		d, ok := ClassifyHTTPResponse(mkResp(http.StatusTooManyRequests, "application/json"), []byte(`{"message":"slow down"}`), Context{})
		equal(t, "ok", ok, true)
		equal(t, "code", d.Code, CodeRateLimited)
	})
}

// TestExtractJSONError covers the inconsistent-envelope reader (RFC §6.4a).
func TestExtractJSONError(t *testing.T) {
	cases := []struct {
		name         string
		body         string
		wantMsg      string
		wantProvider string
		wantOK       bool
	}{
		{
			name:         "prometheus",
			body:         `{"status":"error","errorType":"bad_data","error":"invalid parameter"}`,
			wantMsg:      "invalid parameter",
			wantProvider: "bad_data",
			wantOK:       true,
		},
		{
			name:         "error as nested object (elastic-ish)",
			body:         `{"error":{"type":"index_not_found_exception","reason":"no such index"}}`,
			wantMsg:      "no such index",
			wantProvider: "index_not_found_exception",
			wantOK:       true,
		},
		{
			name:         "json:api errors array",
			body:         `{"errors":[{"status":"403","detail":"insufficient scope"}]}`,
			wantMsg:      "insufficient scope",
			wantProvider: "403",
			wantOK:       true,
		},
		{
			name:         "graphql extensions code",
			body:         `{"errors":[{"message":"unauthorized","extensions":{"code":"UNAUTHENTICATED"}}]}`,
			wantMsg:      "unauthorized",
			wantProvider: "UNAUTHENTICATED",
			wantOK:       true,
		},
		{
			name:         "oauth error_description",
			body:         `{"error":"invalid_grant","error_description":"AADSTS70008: expired"}`,
			wantMsg:      "AADSTS70008: expired",
			wantProvider: "",
			wantOK:       true,
		},
		{
			name:    "plain string error",
			body:    `{"error":"boom"}`,
			wantMsg: "boom",
			wantOK:  true,
		},
		{
			name:   "not json",
			body:   `<html></html>`,
			wantOK: false,
		},
		{
			name:   "empty",
			body:   ``,
			wantOK: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			msg, provider, ok := ExtractJSONError([]byte(tc.body))
			equal(t, "ok", ok, tc.wantOK)
			if tc.wantMsg != "" {
				equal(t, "msg", msg, tc.wantMsg)
			}
			if tc.wantProvider != "" {
				equal(t, "provider", provider, tc.wantProvider)
			}
		})
	}
}

// TestExtractJSONError_Defensive ensures oversized input degrades, never panics.
func TestExtractJSONError_Defensive(t *testing.T) {
	big := make([]byte, maxBodyParse+1)
	for i := range big {
		big[i] = 'a'
	}
	_, _, ok := ExtractJSONError(big)
	equal(t, "oversized ok", ok, false)
}

func TestSniffBodyKind(t *testing.T) {
	equal(t, "json ct", sniffBodyKind("application/json; charset=utf-8", nil), BodyJSON)
	equal(t, "json+ld", sniffBodyKind("application/ld+json", nil), BodyJSON)
	equal(t, "html ct", sniffBodyKind("text/html", nil), BodyHTML)
	equal(t, "html sniff", sniffBodyKind("", []byte("  <html>")), BodyHTML)
	equal(t, "json sniff", sniffBodyKind("", []byte(" {\"a\":1}")), BodyJSON)
	equal(t, "text", sniffBodyKind("", []byte("just text")), BodyText)
}
