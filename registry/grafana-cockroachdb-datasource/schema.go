package cockroachdbdatasource

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
// truth for the CockroachDB datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the CockroachDB datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the CockroachDB
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the CockroachDB
// datasource: the empty default keyed by "" plus one example per authentication
// method (SQL, Kerberos) and TLS variant (file-path, file-content). Every value
// is a full instance-settings object with jsonData and, where applicable, the
// secureJsonData placeholders.
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with (sslmode='require', tlsConfigurationMethod='file-content', pool defaults 5/2/300/30). The user must still choose authType and supply jsonData.url, jsonData.user, jsonData.database and (for non-Kerberos auth) secureJsonData.password.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"url":                    "",
							"user":                   "",
							"database":               "",
							"sslmode":                string(TLSModeRequire),
							"tlsConfigurationMethod": string(TLSMethodFileContent),
							"maxOpenConns":           DefaultMaxOpenConns,
							"maxIdleConns":           DefaultMaxIdleConns,
							"connMaxLifetime":        DefaultConnMaxLifetime,
							"queryTimeout":           DefaultQueryTimeout,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "",
						},
					},
				},
			},
			"sqlAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "SQL Authentication (username / password)",
					Description: "The most common setup: username + password over the PostgreSQL wire protocol. All connection fields live in jsonData; the password is the only secret.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType": string(AuthTypeSQL),
							"url":      "localhost:26257",
							"database": "defaultdb",
							"user":     "grafana_reader",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"kerberosAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Kerberos Authentication",
					Description: "Kerberos auth needs NO password. credentialCache is required; configFilePath defaults to /etc/krb5.conf; kerberosServerName (krbsrvname) is optional and defaults to 'postgres'. The backend always negotiates sslmode=require with authenticator=krb5.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":           string(AuthTypeKerberos),
							"url":                "crdb.internal:26257",
							"database":           "defaultdb",
							"user":               "grafana_reader",
							"credentialCache":    "/tmp/krb5cc_1000",
							"configFilePath":     "/etc/krb5.conf",
							"kerberosServerName": "postgres",
						},
						"secureJsonData": map[string]any{},
					},
				},
			},
			"tlsVerifyFullFilePath": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "TLS/SSL Authentication, sslmode=verify-full, file-path certs",
					Description: "TLS auth with certificate paths read from the Grafana host's local filesystem. All three paths (root/cert/key) are required. A password is still required.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":               string(AuthTypeTLS),
							"url":                    "crdb.internal:26257",
							"database":               "defaultdb",
							"user":                   "grafana_reader",
							"sslmode":                string(TLSModeVerifyFull),
							"tlsConfigurationMethod": string(TLSMethodFilePath),
							"sslRootCertFile":        "/etc/secrets/cockroach/ca.crt",
							"sslCertFile":            "/etc/secrets/cockroach/client.crt",
							"sslKeyFile":             "/etc/secrets/cockroach/client.key",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"tlsVerifyCAFileContent": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "TLS/SSL Authentication, sslmode=verify-ca, inline PEM content",
					Description: "TLS auth with certificates stored as encrypted content in secureJsonData; the backend writes them to Grafana's data path as files before connecting. All three inline PEMs are required. This is the editor's default TLS method (file-content).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":               string(AuthTypeTLS),
							"url":                    "crdb.internal:26257",
							"database":               "defaultdb",
							"user":                   "grafana_reader",
							"sslmode":                string(TLSModeVerifyCA),
							"tlsConfigurationMethod": string(TLSMethodFileContent),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword):      "changeme",
							string(SecureJsonDataKeyTLSCACert):     "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----\n",
							string(SecureJsonDataKeyTLSClientCert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----\n",
							string(SecureJsonDataKeyTLSClientKey):  "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----\n",
						},
					},
				},
			},
		},
	}
}
