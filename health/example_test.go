package health

import (
	"errors"
	"fmt"
)

// apiError stands in for a provider SDK's typed error (e.g. smithy APIError).
type apiError struct {
	Code    string
	Message string
}

func (e *apiError) Error() string { return e.Code + ": " + e.Message }

// Example_familyRegistration shows how a family library (grafana-aws-sdk, sqlds,
// …) plugs provider knowledge into the core without the core importing its SDK:
// it registers a Classifier that maps the SDK's typed error to a Diagnosis, and a
// Rule that supplies provider-specific remediation. See RFC ADR-002/003/004.
func Example_familyRegistration() {
	resetRegistry()

	// 1. A classifier mapping the SDK's typed error → Code + provider sub-code.
	RegisterClassifier(func(err error, _ Context) (Diagnosis, bool) {
		var apiErr *apiError
		if errors.As(err, &apiErr) && apiErr.Code == "ThrottlingException" {
			return Diagnosis{Code: CodeRateLimited, ProviderCode: apiErr.Code}, true
		}
		return Diagnosis{}, false
	})

	// 2. A remediation rule keyed on that provider sub-code (most specific wins).
	RegisterRules(Rule{
		When: Match{Code: CodeRateLimited, ProviderCode: "ThrottlingException"},
		Give: Remediation{Summary: "Reduce the query rate or request a service-limit increase."},
	})

	// At CheckHealth time the core classifies and resolves remediation.
	d := Diagnose(&apiError{Code: "ThrottlingException", Message: "Rate exceeded"}, Context{})
	fmt.Println("code:", d.Code)
	fmt.Println("providerCode:", d.ProviderCode)
	fmt.Println("remediation:", resolveRemediation(d).Summary)

	// Output:
	// code: RATE_LIMITED
	// providerCode: ThrottlingException
	// remediation: Reduce the query rate or request a service-limit increase.
}
