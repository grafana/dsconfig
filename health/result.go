package health

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
	"github.com/grafana/grafana-plugin-sdk-go/backend/tracing"
)

// Logger is the minimal structured-logging surface the health module needs.
// backend.Logger from grafana-plugin-sdk-go satisfies it, so a plugin can pass
// its logger directly.
type Logger interface {
	Error(msg string, args ...any)
	Warn(msg string, args ...any)
	Info(msg string, args ...any)
}

// Metrics receives one observation per shaped result, for the bounded-cardinality
// counters/histogram in RFC §10.3. Implementations must keep label cardinality
// bounded (no ProviderCode as a metric label).
type Metrics interface {
	Observe(d Diagnosis, durationSeconds float64)
}

// SpanRecorder records the (already redacted) classification onto the active
// trace span. It is injected so the core stays free of a direct OpenTelemetry
// dependency (ADR-001); an otel-backed adapter lives in the calling plugin.
type SpanRecorder interface {
	Record(ctx context.Context, d Diagnosis, redactedVerbose string)
}

type options struct {
	dims           Context
	logger         Logger
	metrics        Metrics
	span           SpanRecorder
	includeVerbose bool
	duration       time.Duration
}

// Option configures Result.
type Option func(*options)

// WithContext supplies the environment dimensions used for classification and
// remediation.
func WithContext(c Context) Option { return func(o *options) { o.dims = c } }

// WithLogger injects the structured log sink (RFC §10.2).
func WithLogger(l Logger) Option { return func(o *options) { o.logger = l } }

// WithMetrics injects the metrics sink (RFC §10.3).
func WithMetrics(m Metrics) Option { return func(o *options) { o.metrics = m } }

// WithSpanRecorder injects the trace-span sink (RFC §10.4).
func WithSpanRecorder(s SpanRecorder) Option { return func(o *options) { o.span = s } }

// WithVerbose includes the redacted verbose detail in JSONDetails. Off by
// default — the correlation ID + logs are the bridge to detail (RFC §14).
func WithVerbose(include bool) Option { return func(o *options) { o.includeVerbose = include } }

// WithDuration records how long the health check took, for the latency histogram
// and the log line.
func WithDuration(d time.Duration) Option { return func(o *options) { o.duration = d } }

// jsonDetails is the stable, additively-versioned JSONDetails contract (RFC §6.6).
// It is self-describing: errorCode is the primary machine-readable reference;
// providerCode/httpStatus carry the upstream's own identifiers; and the diagnostic
// sub-signals (tlsKind/timeoutKind/bodyKind/contentType) let the UI render a full
// description and highlight the offending fields without consulting logs.
type jsonDetails struct {
	ErrorCode     string       `json:"errorCode"`
	ProviderCode  string       `json:"providerCode,omitempty"`
	HTTPStatus    int          `json:"httpStatus,omitempty"`
	TLSKind       string       `json:"tlsKind,omitempty"`
	TimeoutKind   string       `json:"timeoutKind,omitempty"`
	BodyKind      string       `json:"bodyKind,omitempty"`
	ContentType   string       `json:"contentType,omitempty"`
	ErrorSource   string       `json:"errorSource"`
	CorrelationID string       `json:"correlationId,omitempty"`
	Remediation   *Remediation `json:"remediation,omitempty"`
	Verbose       string       `json:"verbose,omitempty"`
}

// OK is the success constructor.
func OK(message string) *backend.CheckHealthResult {
	if message == "" {
		message = "Data source is working"
	}
	return &backend.CheckHealthResult{Status: backend.HealthStatusOk, Message: message}
}

func apply(opts []Option) options {
	o := options{}
	for _, fn := range opts {
		fn(&o)
	}
	return o
}

// Result classifies err once and renders it three ways (RFC ADR-009): a safe UI
// payload (returned), a structured masked log line, metrics, and a redacted span
// record — the latter three only when their sinks are injected. It never returns
// nil and never panics.
//
//   - err == nil           → OK.
//   - benign cancellation  → HealthStatusUnknown, no error surfaced (RFC §6.4b).
//   - anything else        → HealthStatusError with normalized Message + JSONDetails.
func Result(ctx context.Context, err error, opts ...Option) *backend.CheckHealthResult {
	o := apply(opts)
	if err == nil {
		return OK(okMessage(o.dims.DatasourceName))
	}
	if isCancellation(err) {
		return canceledResult()
	}
	return shape(ctx, Diagnose(err, o.dims), redact(errString(err)), o)
}

// ResultForResponse is the response-aware entry point for HTTP families (RFC
// §6.4a). It inspects the response (status + Content-Type + body) so HTML/odd
// answers classify correctly, preserving sub-signals for remediation. A 2xx with
// a normal body and no rawErr yields OK.
func ResultForResponse(ctx context.Context, resp *http.Response, body []byte, rawErr error, opts ...Option) *backend.CheckHealthResult {
	o := apply(opts)
	if rawErr != nil && isCancellation(rawErr) {
		return canceledResult()
	}
	if d, ok := ClassifyHTTPResponse(resp, body, o.dims); ok {
		return shape(ctx, d, bodySummary(d, resp, body), o)
	}
	if rawErr != nil {
		return shape(ctx, Diagnose(rawErr, o.dims), redact(errString(rawErr)), o)
	}
	return OK(okMessage(o.dims.DatasourceName))
}

func canceledResult() *backend.CheckHealthResult {
	return &backend.CheckHealthResult{
		Status:  backend.HealthStatusUnknown,
		Message: "The health check was canceled.",
	}
}

// shape renders a Diagnosis into the four surfaces. verbose is the already-safe
// (redacted/summarized) detail.
func shape(ctx context.Context, d Diagnosis, verbose string, o options) *backend.CheckHealthResult {
	entry := entryFor(d.Code)
	source := entry.Source
	rem := applyDocsBase(resolveRemediation(d), o.dims.DocsBaseURL)
	// Make the offending field self-describing: if the diagnosis pinpointed a
	// config field and the remediation didn't already name fields to highlight,
	// surface it so the UI can flag it.
	if len(rem.Fields) == 0 && d.Field != "" {
		rem.Fields = []string{d.Field}
	}
	corrID := correlationID(ctx)

	details := jsonDetails{
		ErrorCode:     string(d.Code),
		ProviderCode:  d.ProviderCode,
		HTTPStatus:    d.HTTPStatus,
		TLSKind:       string(d.TLSKind),
		TimeoutKind:   string(d.TimeoutKind),
		BodyKind:      string(d.BodyKind),
		ContentType:   d.ContentType,
		ErrorSource:   string(source),
		CorrelationID: corrID,
		Remediation:   &rem,
	}
	if o.includeVerbose {
		details.Verbose = verbose
	}
	raw, _ := json.Marshal(details)

	// Fan out to the optional sinks (RFC §10).
	if o.logger != nil {
		emitLog(o.logger, d, source, corrID, verbose, o.duration)
	}
	if o.metrics != nil {
		o.metrics.Observe(d, o.duration.Seconds())
	}
	if o.span != nil {
		o.span.Record(ctx, d, verbose)
	}

	return &backend.CheckHealthResult{
		Status:      backend.HealthStatusError,
		Message:     buildMessage(o.dims.DatasourceName, entry.Headline, rem.Summary),
		JSONDetails: raw,
	}
}

// bodySummary builds a safe, compact verbose detail from a response body without
// dumping it (RFC §11). HTML → "HTML response (ct, N bytes, status): <title>";
// JSON → the redacted extracted message; otherwise the status line.
func bodySummary(d Diagnosis, resp *http.Response, body []byte) string {
	status := 0
	if resp != nil {
		status = resp.StatusCode
	}
	switch d.BodyKind {
	case BodyHTML:
		title := htmlTitle(body)
		s := fmt.Sprintf("HTML response (%s, %d bytes, %d)", d.ContentType, len(body), status)
		if title != "" {
			s += ": " + redact(title)
		}
		return s
	case BodyJSON:
		if msg, _, ok := ExtractJSONError(body); ok && msg != "" {
			return redact(msg)
		}
		return fmt.Sprintf("JSON error response (%d)", status)
	default:
		return fmt.Sprintf("response status %d", status)
	}
}

// buildMessage assembles the safe UI message from catalog/rule copy and the data
// source name only — never the raw error (RFC §6.6).
func buildMessage(dsName, headline, summary string) string {
	parts := make([]string, 0, 2)
	if headline != "" {
		parts = append(parts, headline)
	}
	if summary != "" {
		parts = append(parts, summary)
	}
	body := strings.Join(parts, " ")
	if dsName != "" {
		return dsName + ": " + body
	}
	return body
}

func okMessage(dsName string) string {
	if dsName != "" {
		return dsName + ": data source is working"
	}
	return "Data source is working"
}

// emitLog writes a single structured line at the boundary, with severity keyed on
// the error source to avoid alert fatigue (RFC §10.2).
func emitLog(l Logger, d Diagnosis, source backend.ErrorSource, corrID, verbose string, dur time.Duration) {
	args := []any{
		"errorCode", string(d.Code),
		"errorSource", string(source),
		"classifierPath", string(d.Path),
		"correlationId", corrID,
	}
	if d.ProviderCode != "" {
		args = append(args, "providerCode", d.ProviderCode)
	}
	if d.HTTPStatus != 0 {
		args = append(args, "httpStatus", d.HTTPStatus)
	}
	if d.TimeoutKind != "" {
		args = append(args, "timeoutKind", string(d.TimeoutKind))
	}
	if d.TLSKind != "" {
		args = append(args, "tlsKind", string(d.TLSKind))
	}
	if d.BodyKind != "" {
		args = append(args, "bodyKind", string(d.BodyKind))
	}
	if d.Context.DatasourceType != "" {
		args = append(args, "datasourceType", d.Context.DatasourceType)
	}
	if dur > 0 {
		args = append(args, "duration", dur.String())
	}
	if verbose != "" {
		args = append(args, "verbose", verbose)
	}

	msg := "data source health check failed"
	switch source {
	case backend.ErrorSourceDownstream:
		l.Warn(msg, args...)
	default:
		// plugin / unknown / classification-failed are our bugs.
		l.Error(msg, args...)
	}
}

// applyDocsBase resolves a relative DocsURL against the context's docs base URL.
func applyDocsBase(rem Remediation, base string) Remediation {
	if base == "" || rem.DocsURL == "" {
		return rem
	}
	if strings.Contains(rem.DocsURL, "://") {
		return rem
	}
	rem.DocsURL = strings.TrimRight(base, "/") + "/" + strings.TrimLeft(rem.DocsURL, "/")
	return rem
}

// correlationID returns the trace ID when a sampled trace is present, else a
// freshly generated short reference (RFC §6.6 / §10.4).
func correlationID(ctx context.Context) string {
	if id := tracing.TraceIDFromContext(ctx, true); id != "" {
		return id
	}
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return ""
	}
	return hex.EncodeToString(b[:])
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
