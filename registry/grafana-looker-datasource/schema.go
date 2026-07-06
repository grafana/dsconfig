package lookerdatasource

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
// truth for the Looker datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Looker datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Looker
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Looker
// datasource. The plugin supports a single authentication method
// (client_secret), so the set is the default configuration plus one realistic
// client-secret example. Each example value is a full instance settings object
// with the plugin configuration nested under jsonData and the write-only secret
// under secureJsonData (placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: client_secret authentication with an empty base URL, client ID, and client secret. jsonData.base_url, jsonData.client_id, and secureJsonData.client_secret must all be filled in to get a working datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"base_url":  "",
							"auth_type": string(AuthTypeClientSecret),
							"client_id": "",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyClientSecret): "",
						},
					},
				},
			},
			"clientSecret": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Client Secret (Looker API3 credentials)",
					Description: "Authenticate against a Looker instance with API3 credentials: jsonData.client_id plus secureJsonData.client_secret, connecting to jsonData.base_url. auth_type is 'client_secret' (the only supported method).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"base_url":  "https://your-instance.looker.app",
							"auth_type": string(AuthTypeClientSecret),
							"client_id": "<your-client-id>",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyClientSecret): "<your-client-secret>",
						},
					},
				},
			},
		},
	}
}
