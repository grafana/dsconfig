package sumologicdatasource

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
// truth for the Sumo Logic datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Sumo Logic datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Sumo Logic
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Sumo Logic
// datasource, covering the default configuration and the access-key
// authentication method across connection (region) variants the config editor
// supports. Each example value is a full instance settings object: the plugin
// configuration nested under jsonData and the write-only access key under
// secureJsonData (obviously-fake placeholder values — replace them with real
// secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: access-key authentication against the US1 / Default region (https://api.sumologic.com/api/). Fill in jsonData.accessId and secureJsonData.accessKey (empty here) to get a working datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authMethod": string(AuthenticationMethodAccessKey),
							"apiUrl":     DefaultApiURL,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "",
						},
					},
				},
			},
			"accessKey": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Access key (US1 / Default region)",
					Description: "Access-key authentication against the US1 / Default Sumo Logic region. The access ID is the HTTP basic-auth username (jsonData.accessId) and the access key the password (secureJsonData.accessKey).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authMethod": string(AuthenticationMethodAccessKey),
							"apiUrl":     DefaultApiURL,
							"accessId":   "<your-access-id>",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "<your-access-key>",
						},
					},
				},
			},
			"accessKeyEU": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Access key (EU region)",
					Description: "Access-key authentication against the EU Sumo Logic region. jsonData.apiUrl selects the region; the region list is US1/Default, US2, EU, AU, CA, DE, IN, JP, FED, plus any custom value.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authMethod": string(AuthenticationMethodAccessKey),
							"apiUrl":     "https://api.eu.sumologic.com/api/",
							"accessId":   "<your-access-id>",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "<your-access-key>",
						},
					},
				},
			},
			"legacyNoAuthMethod": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy: no authMethod set",
					Description: "A datasource created without jsonData.authMethod: LoadSettings defaults it to 'accessKey' (pkg/models/settings.go:40-42). Do not assume a missing authMethod means the datasource is unconfigured.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"apiUrl":   DefaultApiURL,
							"accessId": "<your-access-id>",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessKey): "<your-access-key>",
						},
					},
				},
			},
		},
	}
}
