package csvdatasource

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
// truth for the Grafana CSV datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Grafana CSV datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Grafana
// CSV datasource: the settings (configuration) spec derived from
// dsconfig.json, the secure values, and example configurations, stamped
// with TargetAPIVersion. Grafana's datasource API server serves this bundle
// as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Grafana
// CSV datasource, covering the default configuration and each storage /
// authentication / TLS variant the config editor supports. Each example
// value is a full instance settings object with the plugin configuration
// nested under jsonData and the relevant write-only secrets under
// secureJsonData (placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: HTTP storage, no authentication, no TLS overrides. Only root.url needs to be filled in to get a working datasource pointed at a CSV endpoint on localhost:8080. jsonData.storage defaults to \"http\" for backwards compatibility (src/utils.ts:9, pkg/settings.go:22-24).",
					Value: map[string]any{
						"url": "http://localhost:8080",
						"jsonData": map[string]any{
							"storage": string(DefaultStorageMode),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"httpNoAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "HTTP storage, no authentication",
					Description: "A public or network-isolated CSV endpoint. jsonData.queryParams (\"limit=100\") is appended to every outgoing HTTP request — its values override any collision with per-query params (pkg/http_storage.go:102-109).",
					Value: map[string]any{
						"url": "http://csv.example.com/data.csv",
						"jsonData": map[string]any{
							"storage":     string(StorageModeHTTP),
							"queryParams": "limit=100",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"httpBasicAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "HTTP storage, Basic authentication",
					Description: "Authenticate against a CSV endpoint that requires HTTP Basic. Both basicAuth and basicAuthUser live at the datasource root; only basicAuthPassword is secret.",
					Value: map[string]any{
						"url":           "https://csv.example.com/reports.csv",
						"basicAuth":     true,
						"basicAuthUser": "grafana",
						"jsonData": map[string]any{
							"storage": string(StorageModeHTTP),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "REPLACE_WITH_PASSWORD",
						},
					},
				},
			},
			"httpOAuthForward": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "HTTP storage, Forward OAuth Identity",
					Description: "Forward the signed-in user's upstream OAuth identity to the CSV endpoint. The editor writes only jsonData.oauthPassThru; there is no accompanying secret.",
					Value: map[string]any{
						"url": "https://csv.example.com/reports.csv",
						"jsonData": map[string]any{
							"storage":       string(StorageModeHTTP),
							"oauthPassThru": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"httpTLSMutualAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "HTTP storage, TLS mutual auth (mTLS)",
					Description: "CSV endpoint requiring mTLS. jsonData.tlsAuth=true triggers the SDK to build a client-authenticated transport using serverName plus the PEM-encoded client cert and key in secureJsonData.",
					Value: map[string]any{
						"url": "https://csv.example.com/reports.csv",
						"jsonData": map[string]any{
							"storage":    string(StorageModeHTTP),
							"tlsAuth":    true,
							"serverName": "csv.example.com",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyTLSClientCert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
							string(SecureJsonDataKeyTLSClientKey):  "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
						},
					},
				},
			},
			"httpTLSSelfSignedCA": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "HTTP storage, self-signed CA verification",
					Description: "CSV endpoint behind a private CA. jsonData.tlsAuthWithCACert=true tells the SDK to verify the server certificate against the PEM-encoded CA in secureJsonData.tlsCACert.",
					Value: map[string]any{
						"url": "https://csv.internal.corp/reports.csv",
						"jsonData": map[string]any{
							"storage":           string(StorageModeHTTP),
							"tlsAuthWithCACert": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyTLSCACert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
						},
					},
				},
			},
			"httpAdvanced": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "HTTP storage, timeout + cookies + query params",
					Description: "HTTP CSV endpoint with a 30-second request timeout, an explicit list of forwarded cookies, and admin-configured query parameters merged into every request.",
					Value: map[string]any{
						"url": "https://csv.example.com/reports.csv",
						"jsonData": map[string]any{
							"storage":     string(StorageModeHTTP),
							"timeout":     30,
							"keepCookies": []any{"session_id"},
							"queryParams": "format=csv&limit=1000",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"localFile": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Local filesystem storage",
					Description: "Read CSV data from the plugin host's local filesystem. root.url carries the base filesystem path (pkg/local_storage.go:33-45); per-query Path values are joined to it. REQUIRES the plugin process to have been started with GF_PLUGIN_ALLOW_LOCAL_MODE=true, otherwise every query returns \"local mode has been disabled by your administrator\" (pkg/datasource.go:44,158-160).",
					Value: map[string]any{
						"url": "/var/lib/csv-data",
						"jsonData": map[string]any{
							"storage": string(StorageModeLocal),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"legacyEmptyStorage": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy storage (pre-storage-selector datasource)",
					Description: "A pre-existing datasource that was provisioned before jsonData.storage existed. Both the frontend (src/utils.ts:4-10) and the backend (pkg/settings.go:22-24) treat a missing storage value as \"http\" for backwards compatibility; LoadConfig.ApplyDefaults fills it in on load.",
					Value: map[string]any{
						"url":      "http://csv.legacy.example.com/data.csv",
						"jsonData": map[string]any{},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
		},
	}
}
