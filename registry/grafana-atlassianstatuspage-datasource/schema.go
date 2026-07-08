package atlassianstatuspagedatasource

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
// Atlassian Statuspage datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the dsconfig schema for the Atlassian Statuspage datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Atlassian Statuspage datasource.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations. This datasource has
// no authentication, so the examples carry only jsonData (no secureJsonData).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "A new datasource only needs the Statuspage URL (jsonData.variables.url, empty here). No authentication is required.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"variables": map[string]any{"url": ""},
						},
					},
				},
			},
			"configured": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Configured",
					Description: "Query the public Statuspage API at {url}/api/v2. Set jsonData.variables.url to the Statuspage URL (e.g. https://www.githubstatus.com).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"variables": map[string]any{"url": "https://www.githubstatus.com"},
						},
					},
				},
			},
		},
	}
}
