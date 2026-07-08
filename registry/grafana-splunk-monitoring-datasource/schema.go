package splunkmonitoringdatasource

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
// truth for the Splunk Infrastructure Monitoring datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Splunk Infrastructure Monitoring datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Splunk
// Infrastructure Monitoring datasource: the settings (configuration) spec
// derived from dsconfig.json, the secure values, and example configurations,
// stamped with TargetAPIVersion. Grafana's datasource API server serves this
// bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Splunk
// Infrastructure Monitoring datasource, covering the default configuration and
// each connection variant the config editor supports (realm-based, custom URL
// overrides with a realm, and custom URLs without a realm). Each example value
// is a full instance settings object with the plugin configuration nested under
// jsonData and the write-only access token under secureJsonData (placeholder
// values — replace them with a real access token).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The empty state a new datasource starts with. The plugin defines no config defaults, so both jsonData.realmName and secureJsonData.accessToken must be filled in for the datasource to load — an empty access token is rejected (pkg/models/settings.go:27-30) and an empty realm with no custom URLs yields the broken host https://api..signalfx.com.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"realmName": "",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessToken): "",
						},
					},
				},
			},
			"realm": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Realm-based (default URLs)",
					Description: "Standard configuration. The realm (e.g. us1) drives both derived endpoints: https://api.us1.signalfx.com (metrics-metadata) and https://stream.us1.signalfx.com (SignalFlow). The access token is provided in secureJsonData.accessToken and sent as the X-SF-TOKEN header.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"realmName": "us1",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessToken): "<your-access-token>",
						},
					},
				},
			},
			"customUrls": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Realm plus custom URL overrides",
					Description: "A realm is set, but the two derived endpoints are overridden with custom SignalFlow domains via jsonData.url_metrics_metadata and jsonData.url_signalflow. When set, these override the realm-derived defaults (pkg/client/rest.go:342-352).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"realmName":            "us1",
							"url_metrics_metadata": "https://api.custom.example.com",
							"url_signalflow":       "https://stream.custom.example.com",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessToken): "<your-access-token>",
						},
					},
				},
			},
			"customUrlsWithoutRealm": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Custom URLs without a realm",
					Description: "Fully custom SignalFlow domains with no realm. Valid because both jsonData.url_metrics_metadata and jsonData.url_signalflow are set, so the backend never falls through to the realm-derived defaults (pkg/client/rest.go:342-352).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"url_metrics_metadata": "https://api.custom.example.com",
							"url_signalflow":       "https://stream.custom.example.com",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessToken): "<your-access-token>",
						},
					},
				},
			},
		},
	}
}
