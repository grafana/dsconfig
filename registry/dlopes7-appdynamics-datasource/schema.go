package appdynamicsdatasource

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
// truth for the AppDynamics datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the AppDynamics datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the AppDynamics
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the AppDynamics
// datasource, covering the default configuration and each Controller
// authentication method (basic auth, API Client) with and without the optional
// Analytics (Events) API configured. Each example value is a full instance
// settings object: the Controller URL and basic-auth fields at the root, the
// plugin configuration nested under jsonData, and the relevant write-only
// secrets under secureJsonData (obviously-fake placeholder values — replace
// them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The empty defaults a new datasource starts with: no Controller URL, no auth method, and no Analytics. A working datasource needs root.url plus one Controller auth method (basic auth or API Client), so this example intentionally fails validation until those are filled in.",
					Value: map[string]any{
						"jsonData": map[string]any{},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyClientSecret): "",
						},
					},
				},
			},
			"apiClient": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "API Client (OAuth) authentication",
					Description: "Authenticate the Controller (Metrics) API with an API Client: jsonData.clientName + jsonData.clientDomain form the OAuth2 client_id (clientName@clientDomain) and secureJsonData.clientSecret is the client secret. clientSecret takes precedence over basic auth.",
					Value: map[string]any{
						"url": "https://controller.example.com",
						"jsonData": map[string]any{
							"clientName":   "<your-api-client-name>",
							"clientDomain": "<your-account-name>",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyClientSecret): "<your-api-client-secret>",
						},
					},
				},
			},
			"basicAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic authentication",
					Description: "Authenticate the Controller (Metrics) API with basic auth: root.basicAuth=true, root.basicAuthUser (typically username@account) and secureJsonData.basicAuthPassword. Used when no clientSecret is set.",
					Value: map[string]any{
						"url":           "https://controller.example.com",
						"basicAuth":     true,
						"basicAuthUser": "<your-username>",
						"jsonData":      map[string]any{},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "<your-password>",
						},
					},
				},
			},
			"apiClientWithAnalytics": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "API Client + Analytics (Events) API",
					Description: "API Client auth for the Controller plus the optional Analytics API. Analytics requires all three of jsonData.analyticsURL, jsonData.globalAccountName and secureJsonData.analyticsAPIKey; analyticsURL is a separate Events API endpoint.",
					Value: map[string]any{
						"url": "https://controller.example.com",
						"jsonData": map[string]any{
							"clientName":        "<your-api-client-name>",
							"clientDomain":      "<your-account-name>",
							"analyticsURL":      "https://analytics.api.appdynamics.com",
							"globalAccountName": "<your-global-account-name>",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyClientSecret):    "<your-api-client-secret>",
							string(SecureJsonDataKeyAnalyticsAPIKey): "<your-analytics-api-key>",
						},
					},
				},
			},
			"basicAuthWithAnalytics": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic auth + Analytics (Events) API",
					Description: "Basic auth for the Controller plus the optional Analytics API. Analytics is configured independently of the Controller credentials via analyticsURL + globalAccountName + analyticsAPIKey.",
					Value: map[string]any{
						"url":           "https://controller.example.com",
						"basicAuth":     true,
						"basicAuthUser": "<your-username>",
						"jsonData": map[string]any{
							"analyticsURL":      "https://analytics.api.appdynamics.com",
							"globalAccountName": "<your-global-account-name>",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "<your-password>",
							string(SecureJsonDataKeyAnalyticsAPIKey):   "<your-analytics-api-key>",
						},
					},
				},
			},
		},
	}
}
