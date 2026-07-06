package saphanadatasource

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
// truth for the SAP HANA datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// PEM placeholders used in the TLS examples. They are intentionally NOT valid
// certificates/keys — real secrets must never be committed.
const (
	examplePEMCert  = "-----BEGIN CERTIFICATE-----\n<redacted>\n-----END CERTIFICATE-----"
	examplePEMKey   = "-----BEGIN RSA PRIVATE KEY-----\n<redacted>\n-----END RSA PRIVATE KEY-----"
	examplePassword = "<your-password>"
)

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the SAP HANA datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the SAP HANA
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the SAP HANA
// datasource, covering the default configuration and each authentication method
// (basic / TLS client certificate), connection variant (host+port / tenant
// instance+database), and TLS mode (verify / skip-verify / custom CA / disabled)
// the config editor supports. Each example value is a full instance settings
// object with the plugin configuration nested under jsonData and the relevant
// write-only secrets under secureJsonData (obviously-fake placeholder values —
// replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: basic authentication, TLS enabled (tlsDisabled false), and a 30-second timeout. The user must still supply jsonData.server, a port (or tenant instance + database name), jsonData.username, and secureJsonData.password to get a working datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"server":      "",
							"port":        0,
							"username":    "",
							"tlsDisabled": false,
							"timeout":     DefaultTimeout,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "",
						},
					},
				},
			},
			"basicAuthPort": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic auth over host + port (SAP HANA Cloud)",
					Description: "The typical SAP HANA Cloud setup: a server address, port 443, a username, and a password. TLS is enabled by default.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"server":   "hana.example.com",
							"port":     443,
							"username": "GRAFANA_READER",
							"timeout":  DefaultTimeout,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): examplePassword,
						},
					},
				},
			},
			"basicAuthInstance": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic auth over a tenant instance (on-prem / multi-tenant)",
					Description: "Connect to a tenant database by instance number + database name instead of an explicit port. The backend derives the port as 3<instance>13 (e.g. instance '00' -> 30013). instance is stored as a string.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"server":       "hana.example.com",
							"instance":     "00",
							"databaseName": "HXE",
							"username":     "GRAFANA_READER",
							"timeout":      DefaultTimeout,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): examplePassword,
						},
					},
				},
			},
			"tlsClientAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "TLS client-certificate (X.509) authentication",
					Description: "Mutual TLS: jsonData.tlsAuth=true authenticates with an X.509 client certificate and key instead of a username/password. Both secureJsonData.tlsClientCert and secureJsonData.tlsClientKey are required.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"server":  "hana.example.com",
							"port":    443,
							"tlsAuth": true,
							"timeout": DefaultTimeout,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyTLSClientCert): examplePEMCert,
							string(SecureJsonDataKeyTLSClientKey):  examplePEMKey,
						},
					},
				},
			},
			"tlsWithCACert": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic auth verifying the server against a custom CA",
					Description: "jsonData.tlsAuthWithCACert=true verifies the SAP HANA server certificate against a custom / self-signed CA supplied in secureJsonData.tlsCACert, alongside basic authentication.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"server":            "hana.example.com",
							"port":              443,
							"username":          "GRAFANA_READER",
							"tlsAuthWithCACert": true,
							"timeout":           DefaultTimeout,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword):  examplePassword,
							string(SecureJsonDataKeyTLSCACert): examplePEMCert,
						},
					},
				},
			},
			"tlsSkipVerify": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic auth skipping TLS verification (development only)",
					Description: "jsonData.tlsSkipVerify=true accepts any server certificate. Use only for development against a server whose certificate cannot be verified.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"server":        "hana.example.com",
							"port":          443,
							"username":      "GRAFANA_READER",
							"tlsSkipVerify": true,
							"timeout":       DefaultTimeout,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): examplePassword,
						},
					},
				},
			},
			"tlsDisabled": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic auth over a plaintext (non-TLS) connection",
					Description: "jsonData.tlsDisabled=true turns TLS off entirely. Use only when the SAP HANA instance has no TLS configured. The editor clears the other TLS switches when TLS is disabled.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"server":      "hana.example.com",
							"port":        30015,
							"username":    "GRAFANA_READER",
							"tlsDisabled": true,
							"timeout":     DefaultTimeout,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): examplePassword,
						},
					},
				},
			},
		},
	}
}
