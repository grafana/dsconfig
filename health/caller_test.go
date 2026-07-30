package health_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"go.opentelemetry.io/otel/trace"

	"github.com/grafana/dsconfig/health"
	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// These tests view the module strictly through its exported API — exactly what a
// plugin author sees — and focus on the two caller concerns: how context.Context
// propagates (trace, deadline, cancellation) and how wrapped/joined/tagged errors
// flow into classification.

// details mirrors the exported JSONDetails contract for assertions.
type details struct {
	ErrorCode     string              `json:"errorCode"`
	ProviderCode  string              `json:"providerCode"`
	ErrorSource   string              `json:"errorSource"`
	CorrelationID string              `json:"correlationId"`
	Remediation   *health.Remediation `json:"remediation"`
	Verbose       string              `json:"verbose"`
}

func decode(t *testing.T, res *backend.CheckHealthResult) details {
	t.Helper()
	var d details
	if err := json.Unmarshal(res.JSONDetails, &d); err != nil {
		t.Fatalf("decode JSONDetails: %v", err)
	}
	return d
}

// sampledTraceContext returns a context carrying a valid, sampled span — what a
// plugin receives when Grafana propagates a trace into CheckHealth (RFC §10.4).
func sampledTraceContext(t *testing.T, sampled bool) (context.Context, string) {
	t.Helper()
	tid, err := trace.TraceIDFromHex("4bf92f3577b34da6a3ce929d0e0e4736")
	if err != nil {
		t.Fatal(err)
	}
	sid, err := trace.SpanIDFromHex("00f067aa0ba902b7")
	if err != nil {
		t.Fatal(err)
	}
	cfg := trace.SpanContextConfig{TraceID: tid, SpanID: sid}
	if sampled {
		cfg.TraceFlags = trace.FlagsSampled
	}
	ctx := trace.ContextWithSpanContext(context.Background(), trace.NewSpanContext(cfg))
	return ctx, tid.String()
}

// modelDatasource is a minimal stand-in for a real plugin. ping models the
// connection attempt; CheckHealth assembles the health.Context from the request
// and threads the inbound ctx all the way through.
type modelDatasource struct {
	typ, name, network string
	ping               func(ctx context.Context) error
}

func (d modelDatasource) CheckHealth(ctx context.Context) *backend.CheckHealthResult {
	dims := health.Context{
		DatasourceType: d.typ,
		DatasourceName: d.name,
		NetworkPath:    d.network,
	}
	if err := d.ping(ctx); err != nil {
		return health.Result(ctx, err, health.WithContext(dims), health.WithVerbose(true))
	}
	return health.OK("")
}

// --- Context propagation -----------------------------------------------------

func TestCaller_TraceIDBecomesCorrelationID(t *testing.T) {
	ctx, traceID := sampledTraceContext(t, true)
	ds := modelDatasource{typ: "prometheus", name: "Prod", network: "direct",
		ping: func(context.Context) error { return errors.New("upstream 500") }}

	res := ds.CheckHealth(ctx)
	got := decode(t, res)
	if got.CorrelationID != traceID {
		t.Errorf("correlationId = %q, want trace id %q", got.CorrelationID, traceID)
	}
}

func TestCaller_UnsampledTraceFallsBackToGeneratedID(t *testing.T) {
	ctx, traceID := sampledTraceContext(t, false)
	ds := modelDatasource{name: "Prod",
		ping: func(context.Context) error { return errors.New("boom") }}

	got := decode(t, ds.CheckHealth(ctx))
	if got.CorrelationID == "" {
		t.Fatal("expected a generated correlation id")
	}
	if got.CorrelationID == traceID {
		t.Error("must not use an unsampled trace id as the correlation id")
	}
}

func TestCaller_DeadlinePropagates(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		<-release
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	defer close(release)

	// A plugin that propagates ctx into its outbound request and wraps failures.
	ds := modelDatasource{name: "Slow", ping: func(ctx context.Context) error {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
		if _, err := srv.Client().Do(req); err != nil {
			return fmt.Errorf("health probe: %w", err)
		}
		return nil
	}}

	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Millisecond)
	defer cancel()

	res := ds.CheckHealth(ctx)
	equalStr(t, "status-is-error", res.Status == backend.HealthStatusError, true)
	equalStr(t, "code", decode(t, res).ErrorCode, string(health.CodeConnectionTimeout))
}

func TestCaller_CancellationPropagatesAsBenign(t *testing.T) {
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		<-release
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()
	defer close(release)

	ds := modelDatasource{name: "Cancellable", ping: func(ctx context.Context) error {
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, srv.URL, nil)
		if _, err := srv.Client().Do(req); err != nil {
			return fmt.Errorf("health probe: %w", err)
		}
		return nil
	}}

	ctx, cancel := context.WithCancel(context.Background())
	go func() { time.Sleep(25 * time.Millisecond); cancel() }()

	res := ds.CheckHealth(ctx)
	equalStr(t, "status", res.Status == backend.HealthStatusUnknown, true)
}

// --- Error wrapping ----------------------------------------------------------

func TestCaller_TagSurvivesDeepWrapping(t *testing.T) {
	base := errors.New("policy: subject lacks scope")
	wrapped := fmt.Errorf("checkhealth: %w",
		fmt.Errorf("exec query: %w",
			health.Tag(base, health.CodePermissionDenied, "AccessDenied", "role")))

	d := health.Diagnose(wrapped, health.Context{})
	equalStr(t, "code", string(d.Code), string(health.CodePermissionDenied))
	equalStr(t, "providerCode", d.ProviderCode, "AccessDenied")
	equalStr(t, "path", string(d.Path), string(health.PathTag))
}

// typedSDKError stands in for a provider SDK's concrete error type.
type typedSDKError struct{ code, msg string }

func (e *typedSDKError) Error() string { return e.code + ": " + e.msg }

func TestCaller_FamilyClassifierMatchesWrappedTypedError(t *testing.T) {
	// A family library registers a classifier that uses errors.As — so it must
	// still match when the plugin wraps the SDK error with fmt.Errorf.
	health.RegisterClassifier(func(err error, _ health.Context) (health.Diagnosis, bool) {
		var se *typedSDKError
		if errors.As(err, &se) && se.code == "ThrottlingException" {
			return health.Diagnosis{Code: health.CodeRateLimited, ProviderCode: se.code}, true
		}
		return health.Diagnosis{}, false
	})

	wrapped := fmt.Errorf("list metrics: %w", &typedSDKError{code: "ThrottlingException", msg: "Rate exceeded"})
	d := health.Diagnose(wrapped, health.Context{})
	equalStr(t, "code", string(d.Code), string(health.CodeRateLimited))
	equalStr(t, "providerCode", d.ProviderCode, "ThrottlingException")
	equalStr(t, "path", string(d.Path), string(health.PathFamily))
}

func TestCaller_JoinedErrorsPickByPrecedence(t *testing.T) {
	// Several independent failures joined together — the most actionable single
	// code must win (config → TLS → auth → connection → permission, RFC §15).
	joined := errors.Join(
		health.Tag(errors.New("no read permission"), health.CodePermissionDenied, "", ""),
		health.Tag(errors.New("certificate expired"), health.CodeTLSError, "", ""),
		health.Tag(errors.New("auth failed"), health.CodeAuthenticationFailed, "", ""),
	)
	// Wrap the whole join too, to prove unwrapping still reaches every tag.
	d := health.Diagnose(fmt.Errorf("checkhealth: %w", joined), health.Context{})
	equalStr(t, "code", string(d.Code), string(health.CodeTLSError))
}

func TestCaller_WrappedSentinelClassifiesViaErrorsIs(t *testing.T) {
	// context.DeadlineExceeded wrapped by the plugin must still classify as a
	// timeout through Result (errors.Is traversal).
	res := health.Result(context.Background(),
		fmt.Errorf("ping: %w", context.DeadlineExceeded),
		health.WithContext(health.Context{DatasourceName: "DB"}))
	equalStr(t, "status", res.Status == backend.HealthStatusError, true)
	equalStr(t, "code", decode(t, res).ErrorCode, string(health.CodeConnectionTimeout))
}

func TestCaller_WrappedSecretNeverReachesMessage(t *testing.T) {
	// A realistic driver error carrying a DSN/secret, wrapped by the plugin.
	raw := fmt.Errorf("connect: %w",
		errors.New(`pq: auth failed for "svc" dsn=postgres://svc:p@ssw0rd@db:5432/app`))
	res := health.Result(context.Background(), raw,
		health.WithContext(health.Context{DatasourceName: "Reporting"}),
		health.WithVerbose(true))

	if strings.Contains(res.Message, "p@ssw0rd") || strings.Contains(res.Message, "postgres://") {
		t.Fatalf("secret/DSN leaked into message: %q", res.Message)
	}
	if v := decode(t, res).Verbose; strings.Contains(v, "p@ssw0rd") {
		t.Errorf("secret leaked into verbose: %q", v)
	}
}

// --- ResultForResponse from the caller, with a wrapped read error ------------

func TestCaller_ResultForResponseHandlesHTMLAndBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusBadGateway)
		_, _ = io.WriteString(w, "<html><title>502</title></html>")
	}))
	defer srv.Close()

	req, _ := http.NewRequestWithContext(context.Background(), http.MethodGet, srv.URL, nil)
	resp, err := srv.Client().Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))

	res := health.ResultForResponse(context.Background(), resp, body, nil,
		health.WithContext(health.Context{DatasourceName: "Gateway"}), health.WithVerbose(true))
	got := decode(t, res)
	equalStr(t, "code", got.ErrorCode, string(health.CodeUpstreamError))
	if !strings.Contains(got.Verbose, "httpStatus=502") {
		t.Errorf("verbose should fold in httpStatus, got %q", got.Verbose)
	}
}

// equalStr is a tiny assert helper local to the external test package.
func equalStr[T comparable](t *testing.T, name string, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("%s = %v, want %v", name, got, want)
	}
}
