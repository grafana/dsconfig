package sentrydatasource

import (
	_ "embed"

	"k8s.io/kube-openapi/pkg/spec3"

	"github.com/grafana/dsconfig/dsconfig"
	sdkschema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

//go:generate go test -generateArtifacts -run TestSchemaConformance ./...

// TargetAPIVersion is the API version this schema applies to.
const TargetAPIVersion = dsconfig.TargetAPIVersion

// configSchemaJSON is the declarative dsconfig schema — the single source
// of truth for the Sentry datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig
// schema (single source of truth) for the Sentry datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Sentry
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Sentry
// datasource, covering the default configuration and each connection
// variant the config editor supports. Each example value is a full
// instance settings object with the plugin configuration nested under
// jsonData and the relevant write-only secrets under secureJsonData
// (placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: pointed at https://sentry.io with an empty org slug and an empty auth token placeholder. Both jsonData.orgSlug and secureJsonData.authToken must be filled in for the datasource to load — pkg/plugin/settings.go:41-50 rejects either being empty.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"url": DefaultSentryURL,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAuthToken): "",
						},
					},
				},
			},
			"sentrySaaS": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Sentry SaaS (sentry.io)",
					Description: "Standard configuration against the hosted Sentry service at https://sentry.io. jsonData.orgSlug is the org slug — the last segment of https://sentry.io/organizations/{organization_slug}/.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"url":     DefaultSentryURL,
							"orgSlug": "example-org",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAuthToken): "sntrys_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
						},
					},
				},
			},
			"selfHosted": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Self-hosted Sentry",
					Description: "On-prem Sentry instance. jsonData.url points at the deployment's base URL (no trailing slash — the client string-concatenates request paths onto it at pkg/sentry/sentry.go:54-56).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"url":     "https://sentry.example.com",
							"orgSlug": "example-org",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAuthToken): "sntrys_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
						},
					},
				},
			},
			"selfHostedTLSSkipVerify": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Self-hosted Sentry with a self-signed certificate",
					Description: "On-prem Sentry with a self-signed / private-CA certificate. jsonData.tlsSkipVerify=true disables server certificate verification on the SDK HTTP client (pkg/plugin/plugin.go:57-63). Only use with trusted infrastructure.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"url":           "https://sentry.internal.corp",
							"orgSlug":       "example-org",
							"tlsSkipVerify": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAuthToken): "sntrys_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
						},
					},
				},
			},
			"legacyMissingURL": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy datasource with no URL set",
					Description: "A pre-existing datasource that was provisioned without jsonData.url. Both the editor initial state (src/editors/SentryConfigEditor.tsx:19) and the backend (pkg/plugin/settings.go:37-40) treat an empty URL as https://sentry.io; LoadConfig.ApplyDefaults fills it in on load.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"orgSlug": "example-org",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAuthToken): "sntrys_XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
						},
					},
				},
			},
		},
	}
}
