package health

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// simCheckHealth models a real HTTP data source's CheckHealth: it performs a GET
// and normalizes the outcome through the health module. This drives the end-to-end
// "simulated scenarios" against a live httptest server (RFC §6.4a/§6.4b).
func simCheckHealth(ctx context.Context, client *http.Client, url string, dims Context) *backend.CheckHealthResult {
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	resp, err := client.Do(req)
	if err != nil {
		return Result(ctx, err, WithContext(dims), WithVerbose(true))
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, maxBodyParse))
	return ResultForResponse(ctx, resp, body, nil, WithContext(dims), WithVerbose(true))
}

func detailsOf(t *testing.T, res *backend.CheckHealthResult) jsonDetails {
	t.Helper()
	return parseDetails(t, res)
}

// TestScenario_HTMLGateway: a reverse proxy / LB returns an HTML 502.
func TestScenario_HTMLGateway(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, "<html><title>502 Bad Gateway</title><body>nginx</body></html>")
	}))
	defer srv.Close()

	res := simCheckHealth(context.Background(), srv.Client(), srv.URL, Context{DatasourceName: "Metrics"})
	equal(t, "status", res.Status, backend.HealthStatusError)
	d := detailsOf(t, res)
	equal(t, "code", d.ErrorCode, string(CodeUpstreamError))
	equal(t, "httpStatus", d.HTTPStatus, http.StatusBadGateway)
	equal(t, "bodyKind", d.BodyKind, string(BodyHTML))
	if !strings.Contains(d.ContentType, "text/html") {
		t.Errorf("contentType should surface, got %q", d.ContentType)
	}
	if !strings.Contains(d.Verbose, "HTML response") {
		t.Errorf("verbose should summarize HTML, got %q", d.Verbose)
	}
}

// TestScenario_SSORedirect: a proxy silently returns a 200 HTML login page.
func TestScenario_SSORedirect(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, "<!doctype html><html><title>Sign in - Okta</title></html>")
	}))
	defer srv.Close()

	res := simCheckHealth(context.Background(), srv.Client(), srv.URL, Context{NetworkPath: "proxy"})
	d := detailsOf(t, res)
	equal(t, "code", d.ErrorCode, string(CodeAuthenticationFailed))
}

// TestScenario_WAFForbidden: a WAF blocks the request with an HTML 403.
func TestScenario_WAFForbidden(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusForbidden)
		_, _ = io.WriteString(w, "<html><title>Access Denied</title></html>")
	}))
	defer srv.Close()

	res := simCheckHealth(context.Background(), srv.Client(), srv.URL, Context{})
	d := detailsOf(t, res)
	equal(t, "code", d.ErrorCode, string(CodePermissionDenied))
}

// TestScenario_JSONErrorEnvelope: a Prometheus-style JSON error with a 400.
func TestScenario_JSONErrorEnvelope(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"status": "error", "errorType": "bad_data", "error": "invalid 'start' parameter",
		})
	}))
	defer srv.Close()

	res := simCheckHealth(context.Background(), srv.Client(), srv.URL, Context{})
	d := detailsOf(t, res)
	equal(t, "code", d.ErrorCode, string(CodeInvalidConfiguration))
	equal(t, "providerCode", d.ProviderCode, "bad_data")
	equal(t, "httpStatus", d.HTTPStatus, http.StatusBadRequest)
	if !strings.Contains(d.Verbose, "invalid 'start' parameter") {
		t.Errorf("verbose should carry extracted hint, got %q", d.Verbose)
	}
}

// TestScenario_Healthy: a 200 JSON body is reported OK.
func TestScenario_Healthy(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"status":"success"}`)
	}))
	defer srv.Close()

	res := simCheckHealth(context.Background(), srv.Client(), srv.URL, Context{DatasourceName: "Prom"})
	equal(t, "status", res.Status, backend.HealthStatusOk)
}

// TestScenario_ConnectionRefused: nothing listening on the address.
func TestScenario_ConnectionRefused(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := srv.URL
	client := srv.Client()
	srv.Close() // stop listening so the dial is refused

	res := simCheckHealth(context.Background(), client, url, Context{})
	equal(t, "status", res.Status, backend.HealthStatusError)
	d := detailsOf(t, res)
	equal(t, "code", d.ErrorCode, string(CodeHostUnreachable))
}

// TestScenario_Timeout: the server is slower than the client deadline.
func TestScenario_Timeout(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		<-release // block until the test releases us
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	defer close(release)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	res := simCheckHealth(ctx, srv.Client(), srv.URL, Context{})
	equal(t, "status", res.Status, backend.HealthStatusError)
	d := detailsOf(t, res)
	equal(t, "code", d.ErrorCode, string(CodeConnectionTimeout))
}

// TestScenario_Cancellation: the caller cancels mid-check → benign, not an error.
func TestScenario_Cancellation(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		<-release
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	defer close(release)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(30 * time.Millisecond)
		cancel()
	}()

	res := simCheckHealth(ctx, srv.Client(), srv.URL, Context{})
	equal(t, "status", res.Status, backend.HealthStatusUnknown)
}
