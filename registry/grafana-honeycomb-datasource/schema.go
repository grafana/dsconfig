package honeycombdatasource

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
// truth for the Honeycomb datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Honeycomb datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Honeycomb
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Honeycomb
// datasource, covering the default configuration and the connection variants
// the config editor supports. Each example value is a full instance settings
// object: the plugin configuration nested under jsonData and the write-only
// apiKey secret under secureJsonData (obviously-fake placeholder values —
// replace them with a real API key).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: the Honeycomb US API (https://api.honeycomb.io) with a 7-day query time window. Fill in jsonData.team and secureJsonData.apiKey (both empty here) to get a working datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"hostname":       DefaultHoneycombAPIURL,
							"team":           "",
							"retentionLimit": DefaultRetentionLimitDays,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIKey): "",
						},
					},
				},
			},
			"usApi": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "US API with API key",
					Description: "Minimal working configuration against the Honeycomb US API. Authenticates with the Team API key (sent as the X-Honeycomb-Team header) in secureJsonData.apiKey.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"hostname": DefaultHoneycombAPIURL,
							"team":     "<your-honeycomb-team-slug>",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIKey): "<your-honeycomb-api-key>",
						},
					},
				},
			},
			"euApi": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "EU region API",
					Description: "Connection to a non-default (EU) Honeycomb API host. jsonData.hostname selects the region; it must be an https URL and defaults to the US host https://api.honeycomb.io when omitted.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"hostname": "https://api.eu1.honeycomb.io",
							"team":     "<your-honeycomb-team-slug>",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIKey): "<your-honeycomb-api-key>",
						},
					},
				},
			},
			"withEnvironment": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "With environment name",
					Description: "Sets the optional environment name, which (together with the team) is used to build the 'Open in Honeycomb' data-link URLs (.../<team>/environments/<environment>/...). It does not affect API authentication.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"hostname":    DefaultHoneycombAPIURL,
							"team":        "<your-honeycomb-team-slug>",
							"environment": "<your-honeycomb-environment>",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIKey): "<your-honeycomb-api-key>",
						},
					},
				},
			},
			"extendedRetention": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Extended query time window",
					Description: "Raises jsonData.retentionLimit above the default 7 days. Only set this higher than 7 if your Honeycomb plan retains more than 7 days of data; otherwise queries are clipped and a 'Partial results' warning is attached.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"hostname":       DefaultHoneycombAPIURL,
							"team":           "<your-honeycomb-team-slug>",
							"retentionLimit": 30,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIKey): "<your-honeycomb-api-key>",
						},
					},
				},
			},
		},
	}
}
