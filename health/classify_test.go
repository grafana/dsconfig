package health

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/url"
	"testing"
)

// TestDiagnose_Generic exercises the generic Go-inspection stage against
// realistic standard-library error values (RFC §6.4 step 3, §6.4b).
func TestDiagnose_Generic(t *testing.T) {
	t.Cleanup(resetRegistry)
	resetRegistry()

	cases := []struct {
		name        string
		err         error
		wantCode    Code
		wantTimeout TimeoutKind
		wantTLS     TLSKind
		wantPath    ClassifierPath
	}{
		{
			name:        "checkhealth deadline exceeded",
			err:         fmt.Errorf("query: %w", context.DeadlineExceeded),
			wantCode:    CodeConnectionTimeout,
			wantTimeout: TimeoutDeadline,
			wantPath:    PathGeneric,
		},
		{
			name:        "dial timeout reads as unreachable",
			err:         &net.OpError{Op: "dial", Net: "tcp", Err: timeoutErr{msg: "i/o timeout"}},
			wantCode:    CodeHostUnreachable,
			wantTimeout: TimeoutDial,
			wantPath:    PathGeneric,
		},
		{
			name:        "read timeout",
			err:         &net.OpError{Op: "read", Net: "tcp", Err: timeoutErr{msg: "i/o timeout"}},
			wantCode:    CodeConnectionTimeout,
			wantTimeout: TimeoutRead,
			wantPath:    PathGeneric,
		},
		{
			name:     "connection refused",
			err:      &net.OpError{Op: "dial", Net: "tcp", Err: errors.New("connect: connection refused")},
			wantCode: CodeHostUnreachable,
			wantPath: PathGeneric,
		},
		{
			name:     "dns no such host",
			err:      &net.DNSError{Err: "no such host", Name: "db.invalid", IsNotFound: true},
			wantCode: CodeHostUnreachable,
			wantPath: PathGeneric,
		},
		{
			name:     "tls unknown authority",
			err:      x509.UnknownAuthorityError{},
			wantCode: CodeTLSError,
			wantTLS:  TLSUnknownAuthority,
			wantPath: PathGeneric,
		},
		{
			name:     "tls hostname mismatch",
			err:      x509.HostnameError{Host: "wrong.example.com"},
			wantCode: CodeTLSError,
			wantTLS:  TLSHostnameMismatch,
			wantPath: PathGeneric,
		},
		{
			name:     "tls expired",
			err:      x509.CertificateInvalidError{Reason: x509.Expired, Detail: "expired"},
			wantCode: CodeTLSError,
			wantTLS:  TLSExpired,
			wantPath: PathGeneric,
		},
		{
			name:     "tls record header",
			err:      tls.RecordHeaderError{Msg: "first record does not look like a TLS handshake"},
			wantCode: CodeTLSError,
			wantPath: PathGeneric,
		},
		{
			name:        "url error wrapping timeout",
			err:         &url.Error{Op: "Get", URL: "http://x", Err: timeoutErr{msg: "timeout"}},
			wantCode:    CodeConnectionTimeout,
			wantTimeout: TimeoutRead,
			wantPath:    PathGeneric,
		},
		{
			name:     "unrecognized error",
			err:      errors.New("something weird"),
			wantCode: CodeUnknown,
			wantPath: PathUnknown,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			d := Diagnose(tc.err, Context{})
			equal(t, "code", d.Code, tc.wantCode)
			equal(t, "timeoutKind", d.TimeoutKind, tc.wantTimeout)
			equal(t, "tlsKind", d.TLSKind, tc.wantTLS)
			equal(t, "path", d.Path, tc.wantPath)
		})
	}
}

// TestDiagnose_ExplicitTag verifies a tagged error short-circuits the pipeline.
func TestDiagnose_ExplicitTag(t *testing.T) {
	t.Cleanup(resetRegistry)
	resetRegistry()

	err := Tag(errors.New("boom"), CodePermissionDenied, "AccessDenied", "role")
	d := Diagnose(err, Context{})
	equal(t, "code", d.Code, CodePermissionDenied)
	equal(t, "providerCode", d.ProviderCode, "AccessDenied")
	equal(t, "field", d.Field, "role")
	equal(t, "path", d.Path, PathTag)

	// A tag with no code falls back to UNKNOWN, never empty.
	d2 := Diagnose(&Error{Err: errors.New("x")}, Context{})
	equal(t, "empty-tag code", d2.Code, CodeUnknown)
}

// TestDiagnose_FamilyClassifierPriority verifies higher priority wins and that
// resolution does not depend on registration order (ADR-008).
func TestDiagnose_FamilyClassifierPriority(t *testing.T) {
	t.Cleanup(resetRegistry)
	resetRegistry()

	sentinel := errors.New("provider boom")

	// Registered first, low priority — would match but must lose.
	RegisterClassifierWithPriority(1, func(err error, _ Context) (Diagnosis, bool) {
		if errors.Is(err, sentinel) {
			return Diagnosis{Code: CodeUpstreamError}, true
		}
		return Diagnosis{}, false
	})
	// Registered second, high priority — must win.
	RegisterClassifierWithPriority(10, func(err error, _ Context) (Diagnosis, bool) {
		if errors.Is(err, sentinel) {
			return Diagnosis{Code: CodeRateLimited, ProviderCode: "Throttling"}, true
		}
		return Diagnosis{}, false
	})

	d := Diagnose(fmt.Errorf("wrap: %w", sentinel), Context{})
	equal(t, "code", d.Code, CodeRateLimited)
	equal(t, "providerCode", d.ProviderCode, "Throttling")
	equal(t, "path", d.Path, PathFamily)
}

// TestDiagnose_NilError is defensive: never panics, never empty code.
func TestDiagnose_NilError(t *testing.T) {
	d := Diagnose(nil, Context{})
	equal(t, "code", d.Code, CodeUnknown)
}

// TestIsCancellation separates benign cancellation from a real deadline timeout.
func TestIsCancellation(t *testing.T) {
	equal(t, "canceled", isCancellation(fmt.Errorf("x: %w", context.Canceled)), true)
	equal(t, "deadline", isCancellation(context.DeadlineExceeded), false)
	equal(t, "other", isCancellation(errors.New("x")), false)
	equal(t, "nil", isCancellation(nil), false)
}
