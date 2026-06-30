// Package health normalizes data source CheckHealth failures into a consistent,
// safe, machine-classified and actionable result.
//
// Every failure is shaped into:
//
//   - a stable, machine-readable error Code (the closed public taxonomy),
//   - a clean, secret-free human Message,
//   - structured, context-aware Remediation, and
//   - a redacted verbose detail for support.
//
// The core depends only on github.com/grafana/grafana-plugin-sdk-go. Provider
// specific knowledge (AWS, Azure, SQL drivers, …) is injected by family
// libraries that register Classifiers and Rules via RegisterClassifier and
// RegisterRules; the core never imports those SDKs.
//
// The design is documented in the RFC at
// https://github.com/grafana/dsconfig/issues/110.
package health
