package newrelicdatasource

import (
	_ "embed"

	"k8s.io/kube-openapi/pkg/spec3"

	"github.com/grafana/dsconfig/dsconfig"
	sdkschema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

//go:generate go test -generateArtifacts -run TestSchemaConformance ./...

// TargetAPIVersion is the API version this schema applies to.
const TargetAPIVersion = dsconfig.TargetAPIVersion

// configSchemaJSON is the declarative dsconfig schema — the single source of
// truth for the New Relic datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the New Relic datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the New Relic
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the New Relic
// datasource, covering the default configuration, each region variant, and the
// legacy account-ID storage shape. Each example value is a full instance
// settings object with the plugin configuration nested under jsonData and the
// write-only secrets under secureJsonData (obviously-fake placeholder values —
// replace them with real secrets; the numeric accountId is just a New Relic
// account number, not a credential).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: the 300s HTTP timeout and no region selected (the New Relic client falls back to US). Both secrets are empty here — secureJsonData.personalApiKey and secureJsonData.accountId must be filled in for the datasource to load (pkg/datasource/handler_checkhealth.go:139-145).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"timeoutInSeconds": DefaultTimeoutInSeconds,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPersonalAPIKey): "",
							string(SecureJsonDataKeyAccountID):      "",
						},
					},
				},
			},
			"usRegion": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "US region",
					Description: "Standard configuration against the New Relic US data center. timeoutInSeconds is omitted, so the backend applies the 300s default (pkg/models/settings.go:38-40). The API key authenticates requests and the numeric account ID scopes NRQL queries.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"region": string(RegionUS),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPersonalAPIKey): "<your-newrelic-api-key>",
							string(SecureJsonDataKeyAccountID):      "1234567",
						},
					},
				},
			},
			"euRegion": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "EU region with a custom timeout",
					Description: "Connection to the New Relic EU data center with a 600s HTTP timeout. region must be 'US' or 'EU' (src/types.ts:4,187-190); it is passed to the New Relic client's ConfigRegion (pkg/datasource/newrelic_client.go:47-49).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"region":           string(RegionEU),
							"timeoutInSeconds": 600,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPersonalAPIKey): "<your-newrelic-api-key>",
							string(SecureJsonDataKeyAccountID):      "1234567",
						},
					},
				},
			},
			"legacyAccountIdInJsonData": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy: account ID stored in jsonData",
					Description: "A pre-migration datasource that stored the account ID in plaintext under jsonData.accountId. The backend reads only secureJsonData.accountId (pkg/models/settings.go:36), so this datasource fails the account-ID check until the config editor migrates it (src/components/ConfigEditor.tsx:20-40) — open and re-save the config page. personalApiKey is already in secureJsonData.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"region":    string(RegionUS),
							"accountId": "1234567",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPersonalAPIKey): "<your-newrelic-api-key>",
						},
					},
				},
			},
		},
	}
}
