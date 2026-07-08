package datadogdatasource

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
// truth for the Datadog datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Datadog datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Datadog
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Datadog
// datasource, covering the default configuration and each authentication mode
// and connection variant the config editor supports. Each example value is a
// full instance settings object: root basic-auth fields at the top level (only
// in hosted-metrics mode), the plugin configuration nested under jsonData, and
// the relevant write-only secrets under secureJsonData (obviously-fake
// placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: Default (direct API) mode against the US1 region (https://api.datadoghq.com). Fill in secureJsonData.apiKey and secureJsonData.appKey (empty here) to get a working datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"pluginMode": string(PluginModeDefault),
							"url":        DefaultDatadogAPIURL,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIKey): "",
							string(SecureJsonDataKeyAppKey): "",
						},
					},
				},
			},
			"directApiAppKey": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default mode (US1) with API key + App key",
					Description: "Direct connection to the Datadog US1 API. Authenticates with the API key (DD-API-KEY) and Application key (DD-APPLICATION-KEY) in secureJsonData.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"pluginMode": string(PluginModeDefault),
							"url":        DefaultDatadogAPIURL,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIKey): "<your-datadog-api-key>",
							string(SecureJsonDataKeyAppKey): "<your-application-key>",
						},
					},
				},
			},
			"directApiAppKeyEU": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default mode (EU region)",
					Description: "Direct connection to the Datadog EU site. jsonData.url selects the region; the region list is US1 (https://api.datadoghq.com), US3, US5, EU (https://api.datadoghq.eu), and US1-FED, plus any custom value.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"pluginMode": string(PluginModeDefault),
							"url":        "https://api.datadoghq.eu",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIKey): "<your-datadog-api-key>",
							string(SecureJsonDataKeyAppKey): "<your-application-key>",
						},
					},
				},
			},
			"hostedMetrics": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Hosted Datadog Metrics mode",
					Description: "Connection through a Grafana Cloud Datadog proxy. Uses datasource root basic auth: basicAuth=true, basicAuthUser is the Grafana Cloud Prometheus username, and secureJsonData.basicAuthPassword is a Grafana Cloud API key with read permissions. jsonData.url is the hosted-metrics proxy URL (must not be the default Datadog API URL).",
					Value: map[string]any{
						"basicAuth":     true,
						"basicAuthUser": "123456",
						"jsonData": map[string]any{
							"pluginMode": string(PluginModeHostedMetrics),
							"url":        "https://dd-prod-10-prod-us-central-0.grafana.net/datadog",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "<your-grafana-cloud-api-key>",
						},
					},
				},
			},
			"legacyHostedMetricsNoPluginMode": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy: hosted metrics without pluginMode",
					Description: "A datasource created before jsonData.pluginMode existed: root basicAuth is true but pluginMode is absent. getPluginMode (pkg/models/settings.go:84-92) and the editor treat this as hosted-metrics. Do not assume a missing pluginMode means Default mode.",
					Value: map[string]any{
						"basicAuth":     true,
						"basicAuthUser": "123456",
						"jsonData": map[string]any{
							"url": "https://dd-prod-10-prod-us-central-0.grafana.net/datadog",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "<your-grafana-cloud-api-key>",
						},
					},
				},
			},
		},
	}
}
