package zendeskdatasource

import (
	_ "embed"

	"k8s.io/kube-openapi/pkg/spec3"

	"github.com/grafana/dsconfig/dsconfig"
	sdkschema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

//go:generate go test -generateArtifacts -run TestSchemaConformance ./...

// TargetAPIVersion is the API version this schema applies to.
const TargetAPIVersion = dsconfig.TargetAPIVersion

// configSchemaJSON is the dsconfig schema — the single source of
// truth for the Zendesk datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the dsconfig schema
// (single source of truth) for the Zendesk datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Zendesk
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Zendesk
// datasource: the default configuration and a fully populated basic-auth
// example. Each example value is a full instance settings object with the plugin
// configuration nested under jsonData (the service-keyed shape) and
// the write-only secret under secureJsonData (placeholder value — replace it with
// a real Zendesk API token).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: the single service 'zendesk' preselects its only auth method, basic_auth. Fill in the login email (jsonData.services.zendesk.auth.username), the subdomain (jsonData.variables.subdomain), and the API token (secureJsonData \"zendesk.password\", empty here) to get a working datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"services": map[string]any{
								"zendesk": map[string]any{
									"auth": map[string]any{
										"id":       string(AuthMethodBasic),
										"username": "",
									},
								},
							},
							"variables": map[string]any{
								"subdomain": "",
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "",
						},
					},
				},
			},
			"basicAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic auth (email + API token)",
					Description: "Authenticate against https://{subdomain}.zendesk.com/api/v2/ with the account login email and a Zendesk API token. The token is the write-only secret at secureJsonData \"zendesk.password\".",
					Value: map[string]any{
						"jsonData": map[string]any{
							"services": map[string]any{
								"zendesk": map[string]any{
									"auth": map[string]any{
										"id":       string(AuthMethodBasic),
										"username": "agent@example.com",
									},
								},
							},
							"variables": map[string]any{
								"subdomain": "mycompany",
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "<zendesk-api-token>",
						},
					},
				},
			},
		},
	}
}
