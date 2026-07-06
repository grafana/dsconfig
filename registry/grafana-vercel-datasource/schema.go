package verceldatasource

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
// Vercel datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the dsconfig schema for the Vercel datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Vercel datasource.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Vercel
// datasource: the default configuration and a fully populated bearer-token
// example. Each value is a full instance settings object with the plugin config
// nested under jsonData and the write-only secret under secureJsonData.
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: the single service 'vercel' preselects its only auth method, vercelApiKey. Fill in the access token (secureJsonData \"vercel.token\", empty here) to get a working datasource; set jsonData.variables.team_id only for team-scoped tokens.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"services": map[string]any{
								"vercel": map[string]any{
									"auth": map[string]any{
										"id": string(AuthMethodVercelAPIKey),
									},
								},
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "",
						},
					},
				},
			},
			"accessToken": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Access token (bearer)",
					Description: "Authenticate against https://api.vercel.com with a Vercel Access Token provided as the write-only secret at secureJsonData \"vercel.token\". team_id is set here for a team-scoped token.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"services": map[string]any{
								"vercel": map[string]any{
									"auth": map[string]any{
										"id": string(AuthMethodVercelAPIKey),
									},
								},
							},
							"variables": map[string]any{
								"team_id": "team_1a2b3c4d5e6f7g8h9i0j1k2l",
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "<vercel-access-token>",
						},
					},
				},
			},
		},
	}
}
