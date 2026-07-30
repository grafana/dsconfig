package health

// TLSKind is a TLS/certificate failure sub-kind (RFC §6.3).
type TLSKind string

const (
	TLSUnknownAuthority TLSKind = "unknown_authority"
	TLSHostnameMismatch TLSKind = "hostname_mismatch"
	TLSExpired          TLSKind = "expired"
	TLSClientCert       TLSKind = "client_cert"
)

// BodyKind captures the shape of an upstream response body (RFC §6.4a).
type BodyKind string

const (
	BodyJSON BodyKind = "json"
	BodyHTML BodyKind = "html"
	BodyText BodyKind = "text"
)

// TimeoutKind distinguishes dial vs read vs the CheckHealth deadline (RFC §6.4b).
type TimeoutKind string

const (
	TimeoutDial     TimeoutKind = "dial"     // connect timed out (often firewall/wrong port)
	TimeoutRead     TimeoutKind = "read"     // reachable but slow
	TimeoutDeadline TimeoutKind = "deadline" // CheckHealth ctx deadline elapsed
)

// ClassifierPath records which stage of the pipeline produced a Diagnosis. It is
// used by the coverage metric (RFC §10.3) — a rising "generic"/"unknown" rate is
// the backlog signal and the regression alarm.
type ClassifierPath string

const (
	PathTag     ClassifierPath = "tag"
	PathFamily  ClassifierPath = "family"
	PathGeneric ClassifierPath = "generic"
	PathUnknown ClassifierPath = "unknown"
)

// Context carries environment dimensions injected at the call site. Every field
// is optional; classification and remediation degrade gracefully when absent.
type Context struct {
	DatasourceType string
	DatasourceName string
	AuthType       string
	Deployment     string // cloud | enterprise | oss
	NetworkPath    string // direct | pdc | proxy
	DocsBaseURL    string
	Vars           map[string]string
}

// Diagnosis is the structured classification result.
type Diagnosis struct {
	Code         Code
	ProviderCode string
	HTTPStatus   int
	TLSKind      TLSKind
	TimeoutKind  TimeoutKind
	BodyKind     BodyKind
	ContentType  string // upstream response Content-Type, when known
	Field        string // offending config field, when known
	Path         ClassifierPath
	Context      Context
}
