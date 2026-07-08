package mongodbdatasource

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
// truth for the MongoDB datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the MongoDB datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the MongoDB datasource.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the MongoDB
// datasource, covering the default configuration and each authentication method
// and connection variant. Each example value is a full instance settings object
// with root fields at the top level, the plugin configuration under jsonData,
// and the relevant write-only secrets under secureJsonData (obviously-fake
// placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: Credentials (basic auth). The user must still supply jsonData.connection, the root basicAuthUser, and secureJsonData.basicAuthPassword to get a working datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType": string(AuthTypeBasicAuth),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"credentials": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Credentials (basic auth)",
					Description: "Username/password authentication. The username is stored at the ROOT basicAuthUser field (not jsonData); the password is written to secureJsonData.basicAuthPassword. The backend injects them into the connection string as user:password@.",
					Value: map[string]any{
						"basicAuth":     true,
						"basicAuthUser": "grafana_reader",
						"jsonData": map[string]any{
							"authType":   string(AuthTypeBasicAuth),
							"connection": "mongodb://mongodb.example.com:27017/mydb",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "<your-password>",
						},
					},
				},
			},
			"noAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "No authentication",
					Description: "Connects without authentication. No secrets are required; secureJsonData is intentionally empty for this method.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":   string(AuthTypeNoAuth),
							"connection": "mongodb://mongodb.example.com:27017/mydb",
						},
						"secureJsonData": map[string]any{},
					},
				},
			},
			"kerberos": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Kerberos (GSSAPI)",
					Description: "Kerberos authentication. The connection string must include authMechanism=GSSAPI and jsonData.kerberosUser must be set for Kerberos to activate. Provide secureJsonData.kerberosPassword, or a keytab/ccache file path in jsonData. Requires a custom plugin build; not available on Grafana Cloud.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":     string(AuthTypeKerberos),
							"connection":   "mongodb://mongodb.example.com:27017/?authMechanism=GSSAPI",
							"kerberosUser": "hello@EXAMPLE.COM",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyKerberosPassword): "<kerberos-password>",
						},
					},
				},
			},
			"tlsClientAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "TLS client-certificate auth (provisioning)",
					Description: "TLS is configured through provisioning only. jsonData.tlsAuth supplies a client certificate + key; jsonData.tlsAuthWithCACert verifies the server against a custom CA. Populate the PEM material in secureJsonData; add secureJsonData.tlsCertificateKeyFilePassword if the client key is encrypted.",
					Value: map[string]any{
						"basicAuth":     true,
						"basicAuthUser": "grafana_reader",
						"jsonData": map[string]any{
							"authType":          string(AuthTypeBasicAuth),
							"connection":        "mongodb://mongodb.example.com:27017/?tls=true",
							"tlsAuth":           true,
							"tlsAuthWithCACert": true,
							"serverName":        "mongodb.example.com",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "<your-password>",
							string(SecureJsonDataKeyTLSCACert):         "-----BEGIN CERTIFICATE-----\n<redacted>\n-----END CERTIFICATE-----",
							string(SecureJsonDataKeyTLSClientCert):     "-----BEGIN CERTIFICATE-----\n<redacted>\n-----END CERTIFICATE-----",
							string(SecureJsonDataKeyTLSClientKey):      "-----BEGIN PRIVATE KEY-----\n<redacted>\n-----END PRIVATE KEY-----",
						},
					},
				},
			},
			"legacyCredentials": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy: username/password in jsonData",
					Description: "Datasources created before v1.9.0 stored the username in jsonData.user and the password in secureJsonData.password. The backend migrates them to the basic-auth username/password and enables basic auth. New configurations should use the root basicAuthUser + secureJsonData.basicAuthPassword instead.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"connection": "mongodb://mongodb.example.com:27017/mydb",
							"user":       "grafana_reader",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "<your-password>",
						},
					},
				},
			},
		},
	}
}
