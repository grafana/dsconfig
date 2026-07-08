package databricksdatasource

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
// truth for the Databricks datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Databricks datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Databricks
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Databricks
// datasource, covering the default configuration and each authentication type
// the config editor supports. Each example value is a full instance settings
// object with the plugin configuration nested under jsonData and the relevant
// write-only secrets under secureJsonData (placeholder values — replace them
// with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: Personal Access Token authentication with CloudFetch enabled. host, httpPath, and secureJsonData.token (empty here) still need to be filled in to get a working datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":   string(AuthTypePat),
							"cloudFetch": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "",
						},
					},
				},
			},
			"personalAccessToken": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Personal Access Token",
					Description: "Authenticate with a Databricks Personal Access Token provided in secureJsonData.token. host is the workspace server hostname and httpPath is the SQL warehouse HTTP path.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType": string(AuthTypePat),
							"host":     "dbc-a1b2c3d4-e5f6.cloud.databricks.com",
							"httpPath": "/sql/1.0/warehouses/abc123def456",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "<your-databricks-personal-access-token>",
						},
					},
				},
			},
			"oauthPassthrough": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "OAuth Passthrough",
					Description: "Forward the signed-in Grafana user's OAuth token to Databricks. jsonData.oauthPassThru is set to true automatically; no stored secret is required.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AuthTypeOauthPT),
							"host":          "dbc-a1b2c3d4-e5f6.cloud.databricks.com",
							"httpPath":      "/sql/1.0/warehouses/abc123def456",
							"oauthPassThru": true,
						},
						"secureJsonData": map[string]any{},
					},
				},
			},
			"oauthM2M": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "OAuth M2M (service principal)",
					Description: "Databricks service-principal OAuth: jsonData.clientId plus secureJsonData.clientSecret.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType": string(AuthTypeOauthM2M),
							"host":     "dbc-a1b2c3d4-e5f6.cloud.databricks.com",
							"httpPath": "/sql/1.0/warehouses/abc123def456",
							"clientId": "11111111-1111-1111-1111-111111111111",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyClientSecret): "dose_1234567890abcdef1234567890abcdef",
						},
					},
				},
			},
			"azureM2M": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Azure Entra ID M2M",
					Description: "Azure Entra ID service principal: jsonData.tenantId, jsonData.clientId, jsonData.azureCloud plus secureJsonData.clientSecret.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":   string(AuthTypeAzureM2M),
							"host":       "adb-1234567890123456.7.azuredatabricks.net",
							"httpPath":   "/sql/1.0/warehouses/abc123def456",
							"tenantId":   "22222222-2222-2222-2222-222222222222",
							"clientId":   "11111111-1111-1111-1111-111111111111",
							"azureCloud": "AzureCloud",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyClientSecret): "Abc8Q~exampleAzureClientSecretValue000000000",
						},
					},
				},
			},
			"azureOBO": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Azure On-Behalf-Of",
					Description: "Azure On-Behalf-Of: credentials in the jsonData.azureCredentials object plus secureJsonData.azureClientSecret. jsonData.oauthPassThru must be true or the backend rejects the datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AuthTypeOauthOBO),
							"host":          "adb-1234567890123456.7.azuredatabricks.net",
							"httpPath":      "/sql/1.0/warehouses/abc123def456",
							"oauthPassThru": true,
							"azureCredentials": map[string]any{
								"authType":   "clientsecret-obo",
								"azureCloud": "AzureCloud",
								"tenantId":   "22222222-2222-2222-2222-222222222222",
								"clientId":   "11111111-1111-1111-1111-111111111111",
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAzureClientSecret): "Abc8Q~exampleAzureClientSecretValue000000000",
						},
					},
				},
			},
		},
	}
}
