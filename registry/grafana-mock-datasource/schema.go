package mockdatasource

import (
	_ "embed"

	"k8s.io/kube-openapi/pkg/spec3"

	"github.com/grafana/dsconfig/dsconfig"
	sdkschema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

//go:generate go test -generateArtifacts -run TestSchemaConformance ./...

// TargetAPIVersion is the API version this schema applies to.
const TargetAPIVersion = dsconfig.TargetAPIVersion

// configSchemaJSON is the declarative dsconfig schema — the single
// source of truth for the Grafana Mock datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig
// schema (single source of truth) for the Grafana Mock datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the
// Grafana Mock datasource: the settings (configuration) spec derived
// from dsconfig.json, the secure values, and example configurations,
// stamped with TargetAPIVersion. Grafana's datasource API server
// serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the
// Grafana Mock datasource, covering the default configuration, each
// authentication method / TLS variant the config editor supports, and
// each shape of the plugin-owned CustomHealthCheck override. Each
// example value is a full instance settings object with the plugin
// configuration nested under jsonData and the relevant write-only
// secrets under secureJsonData (placeholder values — replace them with
// real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: no authentication, no TLS overrides, no custom health check. root.url can be blank — the Mock plugin never dials it.",
					Value: map[string]any{
						"url":      "",
						"jsonData": map[string]any{},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"noAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "No authentication",
					Description: "A Mock datasource with a plausible URL and no HTTP-level auth. The Mock backend does not dial the URL; it is only relevant when downstream tools proxy through the SDK's HTTP client.",
					Value: map[string]any{
						"url":      "http://mock.example.com",
						"jsonData": map[string]any{},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"basicAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic authentication",
					Description: "Authenticate the SDK HTTP client with HTTP Basic. Both basicAuth and basicAuthUser live at the datasource root; only basicAuthPassword is secret.",
					Value: map[string]any{
						"url":           "https://mock.example.com",
						"basicAuth":     true,
						"basicAuthUser": "grafana",
						"jsonData":      map[string]any{},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "REPLACE_WITH_PASSWORD",
						},
					},
				},
			},
			"oauthForward": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Forward OAuth Identity",
					Description: "Forward the signed-in user's upstream OAuth identity through the SDK HTTP client. The editor writes only jsonData.oauthPassThru; there is no accompanying secret.",
					Value: map[string]any{
						"url": "https://mock.example.com",
						"jsonData": map[string]any{
							"oauthPassThru": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"tlsMutualAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "TLS mutual auth (mTLS)",
					Description: "SDK HTTP client using mTLS. jsonData.tlsAuth=true triggers the SDK to build a client-authenticated transport using serverName plus the PEM-encoded client cert and key in secureJsonData.",
					Value: map[string]any{
						"url": "https://mock.example.com",
						"jsonData": map[string]any{
							"tlsAuth":    true,
							"serverName": "mock.example.com",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyTLSClientCert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
							string(SecureJsonDataKeyTLSClientKey):  "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
						},
					},
				},
			},
			"tlsSelfSignedCA": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Self-signed CA verification",
					Description: "SDK HTTP client verifying the upstream against a private CA. jsonData.tlsAuthWithCACert=true tells the SDK to verify the server certificate against the PEM-encoded CA in secureJsonData.tlsCACert.",
					Value: map[string]any{
						"url": "https://mock.internal.corp",
						"jsonData": map[string]any{
							"tlsAuthWithCACert": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyTLSCACert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
						},
					},
				},
			},
			"customHealthCheckError": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Custom health check returning ERROR",
					Description: "Use the CustomHealthCheck override to force a failing CheckHealth response for testing. status=2 maps to backend.HealthStatusError; details is passed through as jsonDetails on the CheckHealth response.",
					Value: map[string]any{
						"url": "http://mock.example.com",
						"jsonData": map[string]any{
							"customHealthCheckEnabled": true,
							"customHealthCheck": map[string]any{
								"status":      2,
								"message":     "upstream is unreachable",
								"details":     "{\"verboseMessage\":\"simulated failure via customHealthCheck\"}",
								"skipBackend": false,
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"customHealthCheckSkipBackend": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Custom health check with skipBackend=true (frontend-only)",
					Description: "With skipBackend=true, the frontend testDatasource() short-circuits and never calls the backend (src/datasource.ts:32-58). The response is synthesised entirely from jsonData.customHealthCheck. status=1 maps to a success response on the UI side.",
					Value: map[string]any{
						"url": "http://mock.example.com",
						"jsonData": map[string]any{
							"customHealthCheckEnabled": true,
							"customHealthCheck": map[string]any{
								"status":      1,
								"message":     "frontend-only OK",
								"details":     "{\"verboseMessage\":\"backend was never called\"}",
								"skipBackend": true,
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
		},
	}
}
