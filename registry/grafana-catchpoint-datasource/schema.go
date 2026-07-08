package catchpointdatasource

import (
	_ "embed"

	"k8s.io/kube-openapi/pkg/spec3"

	"github.com/grafana/dsconfig/dsconfig"
	sdkschema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

//go:generate go test -generateArtifacts -run TestSchemaConformance ./...

// TargetAPIVersion is the API version this schema applies to.
const TargetAPIVersion = dsconfig.TargetAPIVersion

// configSchemaJSON is the dsconfig schema — the single source of truth for the
// Catchpoint datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the dsconfig schema for the Catchpoint datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Catchpoint datasource.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations: the default
// configuration and a populated bearer-token example.
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: the single service 'catchpoint' preselects its only auth method, bearer_token. Fill in the API v2 key (secureJsonData \"catchpoint.token\", empty here) to get a working datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"services": map[string]any{
								"catchpoint": map[string]any{
									"auth": map[string]any{"id": string(AuthMethodBearerToken)},
								},
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "",
						},
					},
				},
			},
			"apiKey": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "API v2 key (bearer)",
					Description: "Authenticate against https://io.catchpoint.com/api/v2 with a Catchpoint REST API v2 Key provided as the write-only secret at secureJsonData \"catchpoint.token\".",
					Value: map[string]any{
						"jsonData": map[string]any{
							"services": map[string]any{
								"catchpoint": map[string]any{
									"auth": map[string]any{"id": string(AuthMethodBearerToken)},
								},
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "<catchpoint-api-v2-key>",
						},
					},
				},
			},
		},
	}
}
