package hellodatasource

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
// Hello datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the dsconfig schema for the Hello datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Hello datasource.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations. Both services use
// no authentication, so the examples carry only jsonData (no secureJsonData).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "Both services ('httpbin', 'postman_echo') use the 'none' auth method and fixed server URLs, so no configuration is required — an empty datasource works. The auth.id discriminators default to 'none'.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"services": map[string]any{
								"httpbin": map[string]any{
									"auth": map[string]any{"id": string(AuthMethodNone)},
								},
								"postman_echo": map[string]any{
									"auth": map[string]any{"id": string(AuthMethodNone)},
								},
							},
						},
					},
				},
			},
		},
	}
}
