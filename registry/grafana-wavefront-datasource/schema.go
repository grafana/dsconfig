package wavefrontdatasource

import (
	_ "embed"

	"k8s.io/kube-openapi/pkg/spec3"

	"github.com/grafana/dsconfig/dsconfig"
	sdkschema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

//go:generate go test -generateArtifacts -run TestSchemaConformance ./...

// TargetAPIVersion is the API version this schema applies to.
const TargetAPIVersion = dsconfig.TargetAPIVersion

// editorDefaultAPIURL is the value the config editor pre-fills into the API
// URL input when jsonData.url is empty (src/selectors.ts:3, DEFAULT_API_URL;
// applied at src/components/ConfigEditor.tsx:16). It is a frontend convenience
// only — the backend does not default an empty url — so it is used purely to
// seed realistic example configurations here, not in the Go Config defaults.
const editorDefaultAPIURL = "https://try.wavefront.com"

// configSchemaJSON is the declarative dsconfig schema — the single source of
// truth for the Wavefront datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Wavefront datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Wavefront
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Wavefront
// datasource, covering the default configuration and each connection variant
// the config editor supports. Each example value is a full instance settings
// object with the plugin configuration nested under jsonData and the write-only
// token under secureJsonData (placeholder values — replace them with a real
// token).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: the editor's pre-filled API URL (https://try.wavefront.com), the default 30s request timeout, and an empty token placeholder. secureJsonData.token must be filled in for the datasource to load — pkg/models/settings.go:35-37 rejects an empty token with 'invalid credentials'.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"url":            editorDefaultAPIURL,
							"requestTimeout": DefaultRequestTimeout,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "",
						},
					},
				},
			},
			"apiToken": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "API token against the hosted Wavefront cluster",
					Description: "Standard configuration: the hosted Wavefront (VMware Aria Operations for Applications) cluster at https://try.wavefront.com authenticated with an API token. The token is sent as 'Authorization: Bearer <token>' on every request (pkg/datasource/datasource.go:45-47).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"url": editorDefaultAPIURL,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "<your-wavefront-api-token>",
						},
					},
				},
			},
			"selfManagedCluster": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Dedicated / self-managed cluster",
					Description: "A dedicated Wavefront cluster. jsonData.url points at the cluster base URL with no trailing slash — the backend trims one trailing '/' (pkg/datasource/datasource.go:42) and joins API paths onto it (pkg/datasource/handler_healthcheck.go:20).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"url": "https://mycluster.wavefront.com",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "<your-wavefront-api-token>",
						},
					},
				},
			},
			"customTimeout": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Custom request timeout",
					Description: "Overrides the per-request timeout via jsonData.requestTimeout (seconds). Any positive value is honored; a missing/null/non-positive value falls back to 30 (pkg/datasource/client.go:20-22).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"url":            editorDefaultAPIURL,
							"requestTimeout": 60,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "<your-wavefront-api-token>",
						},
					},
				},
			},
		},
	}
}
