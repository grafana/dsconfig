package pagerdutydatasource

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
// truth for the PagerDuty datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the PagerDuty datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the PagerDuty
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the PagerDuty
// datasource. PagerDuty exposes a single authentication method (a REST API key)
// and a single, fixed connection (https://api.pagerduty.com), so there is one
// authentication example plus the default example keyed by the empty string.
// Each example value is a full instance settings object with the plugin
// configuration nested under jsonData and the write-only secret under
// secureJsonData (obviously-fake placeholder values — replace them with a real
// key).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: the only auth scheme (api_key) selected and an empty API key placeholder. secureJsonData['auth.api_key.apiKey'] must be filled in for the datasource to authenticate against PagerDuty.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"auth": map[string]any{
								"id": string(AuthSchemeIDAPIKey),
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIKey): "",
						},
					},
				},
			},
			"apiKey": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "PagerDuty REST API key",
					Description: "Authenticate against https://api.pagerduty.com with a PagerDuty REST API key. Enter the raw key; the backend sends it as 'Authorization: Token token=<key>'. A read-only key is recommended.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"auth": map[string]any{
								"id": string(AuthSchemeIDAPIKey),
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIKey): "<your-pagerduty-api-token>",
						},
					},
				},
			},
		},
	}
}
