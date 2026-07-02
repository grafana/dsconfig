package health

// Error tags an underlying error with an explicit Code and optional sub-signals.
// A tagged error has the highest classification priority — it short-circuits the
// pipeline in Diagnose (RFC §6.4, step 1). Use it when the connection code
// already knows exactly what went wrong.
type Error struct {
	Code         Code
	ProviderCode string
	Field        string
	Err          error
}

func (e *Error) Error() string {
	if e.Err == nil {
		return string(e.Code)
	}
	return e.Err.Error()
}

func (e *Error) Unwrap() error { return e.Err }

// WithCode tags err with code.
func WithCode(code Code, err error) error {
	return &Error{Code: code, Err: err}
}

// Tag tags err with code plus an optional provider sub-code and offending field.
func Tag(err error, code Code, providerCode, field string) error {
	return &Error{Code: code, ProviderCode: providerCode, Field: field, Err: err}
}
