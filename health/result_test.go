package health

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/grafana/grafana-plugin-sdk-go/backend"
)

// fakeLogger records the last log call for assertions.
type fakeLogger struct {
	level string
	msg   string
	args  []any
}

func (f *fakeLogger) Error(msg string, args ...any) { f.level, f.msg, f.args = "error", msg, args }
func (f *fakeLogger) Warn(msg string, args ...any)  { f.level, f.msg, f.args = "warn", msg, args }
func (f *fakeLogger) Info(msg string, args ...any)  { f.level, f.msg, f.args = "info", msg, args }

func (f *fakeLogger) arg(key string) (any, bool) {
	for i := 0; i+1 < len(f.args); i += 2 {
		if k, ok := f.args[i].(string); ok && k == key {
			return f.args[i+1], true
		}
	}
	return nil, false
}

type fakeMetrics struct {
	observed bool
	diag     Diagnosis
	seconds  float64
}

func (m *fakeMetrics) Observe(d Diagnosis, secs float64) {
	m.observed, m.diag, m.seconds = true, d, secs
}

type fakeSpan struct {
	recorded bool
	verbose  string
	diag     Diagnosis
}

func (s *fakeSpan) Record(_ context.Context, d Diagnosis, v string) {
	s.recorded, s.diag, s.verbose = true, d, v
}

func parseDetails(t *testing.T, res *backend.CheckHealthResult) jsonDetails {
	t.Helper()
	var d jsonDetails
	if err := json.Unmarshal(res.JSONDetails, &d); err != nil {
		t.Fatalf("unmarshal JSONDetails: %v (%s)", err, res.JSONDetails)
	}
	return d
}

func TestResult_NilIsOK(t *testing.T) {
	res := Result(context.Background(), nil, WithContext(Context{DatasourceName: "Prod DB"}))
	equal(t, "status", res.Status, backend.HealthStatusOk)
	if !strings.Contains(res.Message, "Prod DB") {
		t.Errorf("message %q should include data source name", res.Message)
	}
}

func TestResult_CancellationIsBenign(t *testing.T) {
	log := &fakeLogger{}
	span := &fakeSpan{}
	err := context.Canceled
	res := Result(context.Background(), err, WithLogger(log), WithSpanRecorder(span))
	equal(t, "status", res.Status, backend.HealthStatusUnknown)
	equal(t, "no error log", log.level, "")
	equal(t, "span not marked", span.recorded, false)
}

func TestResult_ErrorShaping(t *testing.T) {
	t.Cleanup(resetRegistry)
	resetRegistry()

	// A downstream auth failure tagged by a (pretend) family classifier, with a
	// secret in the raw error that must never reach the user.
	raw := errors.New(`pq: password authentication failed for user "grafana" dsn=postgres://grafana:s3cr3t@db:5432`)
	err := Tag(raw, CodeAuthenticationFailed, "28P01", "user")

	res := Result(context.Background(), err,
		WithContext(Context{DatasourceName: "Prod DB", DatasourceType: "postgres"}),
		WithVerbose(true),
	)

	equal(t, "status", res.Status, backend.HealthStatusError)

	// Message is safe: no secret, no DSN, includes the data source name.
	if strings.Contains(res.Message, "s3cr3t") || strings.Contains(res.Message, "postgres://") {
		t.Fatalf("message leaked secret/DSN: %q", res.Message)
	}
	if !strings.HasPrefix(res.Message, "Prod DB:") {
		t.Errorf("message should be prefixed with DS name: %q", res.Message)
	}

	d := parseDetails(t, res)
	equal(t, "errorCode", d.ErrorCode, string(CodeAuthenticationFailed))
	equal(t, "providerCode", d.ProviderCode, "28P01")
	equal(t, "errorSource", d.ErrorSource, string(backend.ErrorSourceDownstream))
	if d.CorrelationID == "" {
		t.Error("expected a correlation id")
	}
	if d.Remediation == nil || d.Remediation.Summary == "" {
		t.Error("expected remediation summary")
	}
	// Verbose is present (opted in) but redacted.
	if d.Verbose == "" || strings.Contains(d.Verbose, "s3cr3t") {
		t.Errorf("verbose should be present and redacted: %q", d.Verbose)
	}
}

func TestResult_VerboseHiddenByDefault(t *testing.T) {
	res := Result(context.Background(), errors.New("boom"))
	d := parseDetails(t, res)
	equal(t, "verbose hidden", d.Verbose, "")
	if d.CorrelationID == "" {
		t.Error("correlation id should still be present as the bridge to detail")
	}
}

func TestResult_Sinks(t *testing.T) {
	t.Cleanup(resetRegistry)
	resetRegistry()

	log := &fakeLogger{}
	metrics := &fakeMetrics{}
	span := &fakeSpan{}

	// Downstream error → logged at warn (not error), metrics + span fire.
	err := Tag(errors.New("token=abc123 expired"), CodeAuthenticationFailed, "", "")
	Result(context.Background(), err,
		WithLogger(log), WithMetrics(metrics), WithSpanRecorder(span),
		WithDuration(1500*time.Millisecond),
	)

	equal(t, "log level", log.level, "warn")
	if code, _ := log.arg("errorCode"); code != string(CodeAuthenticationFailed) {
		t.Errorf("log errorCode = %v", code)
	}
	if v, _ := log.arg("verbose"); !strings.Contains(v.(string), "[REDACTED]") {
		t.Errorf("log verbose should be redacted: %v", v)
	}
	equal(t, "metrics observed", metrics.observed, true)
	equal(t, "metrics seconds", metrics.seconds, 1.5)
	equal(t, "span recorded", span.recorded, true)
	if strings.Contains(span.verbose, "abc123") {
		t.Errorf("span verbose leaked secret: %q", span.verbose)
	}
}

// TestResult_SignalsInVerbose checks the diagnostic sub-signals are folded into
// the verbose string (RFC §6.6) — and crucially that the JSONDetails schema does
// NOT grow new fields.
func TestResult_SignalsInVerbose(t *testing.T) {
	t.Cleanup(resetRegistry)
	resetRegistry()

	t.Run("tls kind in verbose", func(t *testing.T) {
		res := Result(context.Background(), x509.UnknownAuthorityError{}, WithVerbose(true))
		d := parseDetails(t, res)
		equal(t, "errorCode", d.ErrorCode, string(CodeTLSError))
		if !strings.Contains(d.Verbose, "tlsKind=unknown_authority") {
			t.Errorf("verbose should carry tlsKind, got %q", d.Verbose)
		}
	})

	t.Run("timeout kind in verbose", func(t *testing.T) {
		res := Result(context.Background(), context.DeadlineExceeded, WithVerbose(true))
		d := parseDetails(t, res)
		equal(t, "errorCode", d.ErrorCode, string(CodeConnectionTimeout))
		if !strings.Contains(d.Verbose, "timeoutKind=deadline") {
			t.Errorf("verbose should carry timeoutKind, got %q", d.Verbose)
		}
	})

	t.Run("offending field folded into remediation.fields", func(t *testing.T) {
		err := Tag(errors.New("missing url"), CodeInvalidConfiguration, "", "url")
		d := parseDetails(t, Result(context.Background(), err))
		if d.Remediation == nil || len(d.Remediation.Fields) != 1 || d.Remediation.Fields[0] != "url" {
			t.Errorf("expected remediation.fields=[url], got %+v", d.Remediation)
		}
	})

	t.Run("schema has no new top-level fields", func(t *testing.T) {
		res := Result(context.Background(), x509.UnknownAuthorityError{}, WithVerbose(true))
		raw := string(res.JSONDetails)
		for _, forbidden := range []string{`"httpStatus"`, `"tlsKind"`, `"timeoutKind"`, `"bodyKind"`, `"contentType"`} {
			if strings.Contains(raw, forbidden) {
				t.Errorf("JSONDetails must not add %s as a top-level field: %s", forbidden, raw)
			}
		}
	})
}

func TestResult_PluginErrorLogsAtError(t *testing.T) {
	t.Cleanup(resetRegistry)
	resetRegistry()

	log := &fakeLogger{}
	// Unknown → plugin source → error level (it's our bug).
	Result(context.Background(), errors.New("totally unexpected"), WithLogger(log))
	equal(t, "log level", log.level, "error")
	if path, _ := log.arg("classifierPath"); path != string(PathUnknown) {
		t.Errorf("classifierPath = %v, want unknown", path)
	}
}
