package health

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"sort"
	"sync"
)

// Classifier maps a raw error to a Diagnosis. Family libraries register these via
// init(); each should be specific enough not to collide with others. Returning
// ok=false means "not mine" and the pipeline moves on.
type Classifier func(err error, ctx Context) (Diagnosis, bool)

type registeredClassifier struct {
	fn       Classifier
	priority int
	seq      int
}

var (
	registryMu        sync.RWMutex
	classifiers       []registeredClassifier
	classifierSeq     int
	classifiersSorted bool
)

// RegisterClassifier registers c at the default priority (0). Equivalent to
// RegisterClassifierWithPriority(0, c).
func RegisterClassifier(c Classifier) { RegisterClassifierWithPriority(0, c) }

// RegisterClassifierWithPriority registers c. Higher priority wins; ties break by
// registration order. Resolution never depends on init() ordering across
// packages (ADR-008).
func RegisterClassifierWithPriority(priority int, c Classifier) {
	if c == nil {
		return
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	classifiers = append(classifiers, registeredClassifier{fn: c, priority: priority, seq: classifierSeq})
	classifierSeq++
	classifiersSorted = false
}

func sortedClassifiers() []registeredClassifier {
	registryMu.Lock()
	defer registryMu.Unlock()
	if !classifiersSorted {
		sort.SliceStable(classifiers, func(i, j int) bool {
			if classifiers[i].priority != classifiers[j].priority {
				return classifiers[i].priority > classifiers[j].priority
			}
			return classifiers[i].seq < classifiers[j].seq
		})
		classifiersSorted = true
	}
	out := make([]registeredClassifier, len(classifiers))
	copy(out, classifiers)
	return out
}

// resetRegistry clears all registered classifiers and rules. Intended for tests.
func resetRegistry() {
	registryMu.Lock()
	classifiers = nil
	classifierSeq = 0
	classifiersSorted = false
	registryMu.Unlock()
	rulesMu.Lock()
	rules = nil
	rulesMu.Unlock()
}

// Diagnose runs the classification pipeline (RFC §6.4), first match wins:
//
//  1. explicit *Error tag,
//  2. registered family classifiers (by priority),
//  3. generic Go inspection (net / TLS / timeout),
//  4. CodeUnknown.
//
// It never returns an empty Code and never panics.
func Diagnose(err error, ctx Context) Diagnosis {
	if err == nil {
		return Diagnosis{Code: CodeUnknown, Path: PathUnknown, Context: ctx}
	}

	// 1. Explicit tag.
	var tagged *Error
	if errors.As(err, &tagged) {
		code := tagged.Code
		if code == "" {
			code = CodeUnknown
		}
		return Diagnosis{
			Code:         code,
			ProviderCode: tagged.ProviderCode,
			Field:        tagged.Field,
			Path:         PathTag,
			Context:      ctx,
		}
	}

	// 2. Registered family classifiers.
	for _, rc := range sortedClassifiers() {
		if d, ok := rc.fn(err, ctx); ok {
			if d.Code == "" {
				d.Code = CodeUnknown
			}
			d.Path = PathFamily
			d.Context = ctx
			return d
		}
	}

	// 3. Generic Go inspection.
	if d, ok := classifyGeneric(err); ok {
		d.Path = PathGeneric
		d.Context = ctx
		return d
	}

	// 4. Unknown.
	return Diagnosis{Code: CodeUnknown, Path: PathUnknown, Context: ctx}
}

// classifyGeneric inspects standard-library error types: context deadlines,
// net.DNSError / net.OpError, and crypto/x509 + crypto/tls failures. Cancellation
// (context.Canceled) is intentionally NOT handled here — it is benign and handled
// upstream in Result (RFC §6.4b).
func classifyGeneric(err error) (Diagnosis, bool) {
	// Timeouts. DeadlineExceeded → the CheckHealth deadline elapsed.
	if errors.Is(err, context.DeadlineExceeded) {
		return Diagnosis{Code: CodeConnectionTimeout, TimeoutKind: TimeoutDeadline}, true
	}

	// net.OpError carries the operation (dial/read/write) and may be a timeout.
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		switch {
		case opErr.Timeout() && opErr.Op == "dial":
			// A dial timeout is indistinguishable from a silently-dropped
			// connection — frame it as unreachable (RFC §6.4b).
			return Diagnosis{Code: CodeHostUnreachable, TimeoutKind: TimeoutDial}, true
		case opErr.Timeout():
			return Diagnosis{Code: CodeConnectionTimeout, TimeoutKind: TimeoutRead}, true
		default:
			// Connection refused / reset / no route.
			return Diagnosis{Code: CodeHostUnreachable}, true
		}
	}

	// DNS resolution failure.
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return Diagnosis{Code: CodeHostUnreachable}, true
	}

	// TLS / certificate failures.
	var unknownAuthority x509.UnknownAuthorityError
	if errors.As(err, &unknownAuthority) {
		return Diagnosis{Code: CodeTLSError, TLSKind: TLSUnknownAuthority}, true
	}
	var hostnameErr x509.HostnameError
	if errors.As(err, &hostnameErr) {
		return Diagnosis{Code: CodeTLSError, TLSKind: TLSHostnameMismatch}, true
	}
	var certInvalid x509.CertificateInvalidError
	if errors.As(err, &certInvalid) {
		kind := TLSKind("")
		if certInvalid.Reason == x509.Expired {
			kind = TLSExpired
		}
		return Diagnosis{Code: CodeTLSError, TLSKind: kind}, true
	}
	var recordHeader tls.RecordHeaderError
	if errors.As(err, &recordHeader) {
		return Diagnosis{Code: CodeTLSError}, true
	}

	// Generic net.Error timeout fallback (e.g. url.Error wrapping a timeout).
	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return Diagnosis{Code: CodeConnectionTimeout, TimeoutKind: TimeoutRead}, true
	}

	return Diagnosis{}, false
}

// isCancellation reports whether err is a benign cancellation (the user/Grafana
// aborted the check) rather than a real failure. DeadlineExceeded is excluded —
// that is a genuine timeout (RFC §6.4b).
func isCancellation(err error) bool {
	return err != nil && errors.Is(err, context.Canceled) && !errors.Is(err, context.DeadlineExceeded)
}
