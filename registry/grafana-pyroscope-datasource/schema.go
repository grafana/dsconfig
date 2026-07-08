package pyroscopedatasource

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
// truth for the Grafana Pyroscope datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Grafana Pyroscope datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Grafana
// Pyroscope datasource: the settings (configuration) spec derived from
// dsconfig.json, the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Grafana
// Pyroscope datasource, covering the default configuration and each
// authentication method / TLS variant / Pyroscope-specific feature
// (`minStep`) the config editor supports. Each example value is a full
// instance settings object with the plugin configuration nested under
// jsonData and the relevant write-only secrets under secureJsonData
// (placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: no authentication, no TLS overrides, no minStep. Only root.url needs to be filled in to get a working datasource pointed at a Pyroscope server on localhost:4040.",
					Value: map[string]any{
						"url":      "http://localhost:4040",
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
					Description: "A public or network-isolated Pyroscope with no HTTP-level auth. A 15-second minStep is set explicitly to match the Pyroscope scrape interval.",
					Value: map[string]any{
						"url": "http://pyroscope.example.com:4040",
						"jsonData": map[string]any{
							"minStep": "15s",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"basicAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic authentication",
					Description: "Authenticate with a Pyroscope that requires HTTP Basic. Both basicAuth and basicAuthUser live at the datasource root; only basicAuthPassword is secret.",
					Value: map[string]any{
						"url":           "https://pyroscope.example.com",
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
					Description: "Forward the signed-in user's upstream OAuth identity to Pyroscope. The editor writes only jsonData.oauthPassThru; there is no accompanying secret.",
					Value: map[string]any{
						"url": "https://pyroscope.example.com",
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
					Description: "Pyroscope requiring mTLS. jsonData.tlsAuth=true triggers the SDK to build a client-authenticated transport using serverName plus the PEM-encoded client cert and key in secureJsonData.",
					Value: map[string]any{
						"url": "https://pyroscope.example.com",
						"jsonData": map[string]any{
							"tlsAuth":    true,
							"serverName": "pyroscope.example.com",
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
					Description: "Pyroscope behind a private CA. jsonData.tlsAuthWithCACert=true tells the SDK to verify the server certificate against the PEM-encoded CA in secureJsonData.tlsCACert.",
					Value: map[string]any{
						"url": "https://pyroscope.internal.corp",
						"jsonData": map[string]any{
							"tlsAuthWithCACert": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyTLSCACert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
						},
					},
				},
			},
			"withMinStep": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic auth with a coarser minStep",
					Description: "Pyroscope with Basic auth and a 1-minute minStep for coarser metric aggregation over long time ranges. The effective per-query step is max(query.Interval, minStep).",
					Value: map[string]any{
						"url":           "https://pyroscope.example.com",
						"basicAuth":     true,
						"basicAuthUser": "grafana",
						"jsonData": map[string]any{
							"minStep": "1m",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "REPLACE_WITH_PASSWORD",
						},
					},
				},
			},
		},
	}
}
