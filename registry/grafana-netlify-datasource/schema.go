package netlifydatasource

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
// Netlify datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the dsconfig schema for the Netlify datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Netlify datasource.
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
					Description: "The defaults a new datasource starts with: the single service 'Netlify' preselects its only auth method, bearer_token. Fill in the personal access token (secureJsonData \"Netlify.token\", empty here) to get a working datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"services": map[string]any{
								"Netlify": map[string]any{
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
			"personalAccessToken": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Personal access token (bearer)",
					Description: "Authenticate against https://api.netlify.com/api/v1 with a Netlify personal access token provided as the write-only secret at secureJsonData \"Netlify.token\".",
					Value: map[string]any{
						"jsonData": map[string]any{
							"services": map[string]any{
								"Netlify": map[string]any{
									"auth": map[string]any{"id": string(AuthMethodBearerToken)},
								},
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "<netlify-personal-access-token>",
						},
					},
				},
			},
		},
	}
}
