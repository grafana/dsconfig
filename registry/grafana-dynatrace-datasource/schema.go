package dynatracedatasource

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
// truth for the Dynatrace datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Dynatrace datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Dynatrace
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Dynatrace
// datasource, covering the default configuration plus each connection type
// (saas/managed/url) and token combination the config editor supports. Each
// example value is a full instance settings object: the plugin configuration
// nested under jsonData and the relevant write-only secrets under
// secureJsonData (obviously-fake angle-bracket placeholder values — replace
// them with real secrets; the default example keyed by "" carries an empty
// apiToken to show what must be filled in).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: SaaS connection type. Fill in jsonData.environmentId and at least one of secureJsonData.apiToken (classic API) or secureJsonData.platformToken (Grail) — apiToken is empty here — to get a working datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"apiType": string(APITypeSaaS),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIToken): "",
						},
					},
				},
			},
			"saasApiToken": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "SaaS with API token (classic API)",
					Description: "SaaS environment authenticating the classic API endpoints (Metrics, Problems, Logs, USQL, audit logs) with a Dynatrace API token sent as 'Api-Token <token>'. The base URL becomes https://<environmentId>.live.dynatrace.com/api/...",
					Value: map[string]any{
						"jsonData": map[string]any{
							"apiType":       string(APITypeSaaS),
							"environmentId": "abc12345",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIToken): "<your-dynatrace-api-token>",
						},
					},
				},
			},
			"saasPlatformToken": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "SaaS with platform token (Grail)",
					Description: "SaaS environment authenticating the Grail platform API with a Dynatrace platform token sent as 'Bearer <token>'. For SaaS Grail the backend switches the host to https://<environmentId>.apps.dynatrace.com automatically.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"apiType":       string(APITypeSaaS),
							"environmentId": "abc12345",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPlatformToken): "<your-dynatrace-platform-token>",
						},
					},
				},
			},
			"saasBothTokens": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "SaaS with both tokens (classic + Grail)",
					Description: "SaaS environment with both an API token (classic API endpoints) and a platform token (Grail). Set both to query classic and Grail endpoints from the same datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"apiType":       string(APITypeSaaS),
							"environmentId": "abc12345",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIToken):      "<your-dynatrace-api-token>",
							string(SecureJsonDataKeyPlatformToken): "<your-dynatrace-platform-token>",
						},
					},
				},
			},
			"managed": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Managed Cluster",
					Description: "Dynatrace Managed cluster. jsonData.domain is the cluster host and is required in this mode; the base URL becomes https://<domain>/e/<environmentId>/api/...",
					Value: map[string]any{
						"jsonData": map[string]any{
							"apiType":       string(APITypeManaged),
							"environmentId": "abc12345",
							"domain":        "dynatrace.example.com",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIToken): "<your-dynatrace-api-token>",
						},
					},
				},
			},
			"rawUrl": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Raw URL",
					Description: "Raw URL connection type: jsonData.environmentId holds the full base URL (the editor relabels the field to 'URL'), and the backend appends /api/... The URL must include the scheme.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"apiType":       string(APITypeURL),
							"environmentId": "https://abc12345.live.dynatrace.com",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIToken): "<your-dynatrace-api-token>",
						},
					},
				},
			},
			"tlsCACert": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "SaaS with custom CA certificate",
					Description: "SaaS environment with a custom CA certificate for TLS verification: jsonData.tlsAuthWithCACert=true enables the secureJsonData.tlsCACert PEM, which the backend validates as a parseable certificate.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"apiType":           string(APITypeSaaS),
							"environmentId":     "abc12345",
							"tlsAuthWithCACert": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIToken):  "<your-dynatrace-api-token>",
							string(SecureJsonDataKeyTLSCACert): "-----BEGIN CERTIFICATE-----\n<your-ca-certificate-pem>\n-----END CERTIFICATE-----",
						},
					},
				},
			},
		},
	}
}
