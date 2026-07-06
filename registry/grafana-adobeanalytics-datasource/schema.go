package adobeanalyticsdatasource

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
// Adobe Analytics datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the dsconfig schema for the Adobe Analytics datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Adobe Analytics datasource.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations: the default
// configuration and a populated OAuth2 client-credentials example.
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: the single service 'adobe_analytics' preselects its only auth method, oauth2_m2m. Fill in the global company id (jsonData.variables.global_company_id), client id, and client secret (secureJsonData \"adobe_analytics.clientSecret\", empty here) to get a working datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"services": map[string]any{
								"adobe_analytics": map[string]any{
									"auth": map[string]any{
										"id":       string(AuthMethodOAuth2M2M),
										"clientId": "",
									},
								},
							},
							"variables": map[string]any{"global_company_id": ""},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyClientSecret): "",
						},
					},
				},
			},
			"oauth2ClientCredentials": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "OAuth2 client credentials",
					Description: "Authenticate against https://analytics.adobe.io/api/{global_company_id} by exchanging the client id and client secret for an access token. The client secret is the write-only secret at secureJsonData \"adobe_analytics.clientSecret\".",
					Value: map[string]any{
						"jsonData": map[string]any{
							"services": map[string]any{
								"adobe_analytics": map[string]any{
									"auth": map[string]any{
										"id":       string(AuthMethodOAuth2M2M),
										"clientId": "fake_client_id",
									},
								},
							},
							"variables": map[string]any{"global_company_id": "examplecompany0"},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyClientSecret): "<adobe-client-secret>",
						},
					},
				},
			},
		},
	}
}
