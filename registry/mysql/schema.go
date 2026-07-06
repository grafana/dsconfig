package mysqldatasource

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
// truth for the MySQL datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the MySQL datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the MySQL datasource.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the MySQL datasource.
// Each example value is a full instance settings object with root fields at the top
// level, jsonData nested, and the relevant write-only secrets under secureJsonData
// (placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: no TLS, no cleartext passwords, connection-pool auto-idle enabled. The user must still supply url, user, jsonData.database, and secureJsonData.password to get a working datasource.",
					Value: map[string]any{
						"url":  "",
						"user": "",
						"jsonData": map[string]any{
							"maxIdleConnsAuto": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "",
						},
					},
				},
			},
			"basicAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Username / password",
					Description: "The typical local MySQL setup: host+port, database, and a password-authenticated user. No TLS.",
					Value: map[string]any{
						"url":  "localhost:3306",
						"user": "grafana_reader",
						"jsonData": map[string]any{
							"database":         "metrics",
							"maxIdleConnsAuto": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"tlsClientAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "mTLS client-certificate auth",
					Description: "Enables jsonData.tlsAuth so the backend supplies a client certificate + key to the server. Populate both secureJsonData.tlsClientCert and secureJsonData.tlsClientKey as full PEM blocks including BEGIN/END lines.",
					Value: map[string]any{
						"url":  "mysql.internal:3306",
						"user": "grafana_reader",
						"jsonData": map[string]any{
							"database":         "metrics",
							"tlsAuth":          true,
							"maxIdleConnsAuto": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword):      "changeme",
							string(SecureJsonDataKeyTLSClientCert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----\n",
							string(SecureJsonDataKeyTLSClientKey):  "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----\n",
						},
					},
				},
			},
			"tlsWithCACert": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "TLS with a custom CA",
					Description: "Enables jsonData.tlsAuthWithCACert so the backend verifies the server's certificate against secureJsonData.tlsCACert.",
					Value: map[string]any{
						"url":  "mysql.internal:3306",
						"user": "grafana_reader",
						"jsonData": map[string]any{
							"database":          "metrics",
							"tlsAuthWithCACert": true,
							"maxIdleConnsAuto":  true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword):  "changeme",
							string(SecureJsonDataKeyTLSCACert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----\n",
						},
					},
				},
			},
			"cleartextPasswords": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "MySQL cleartext client-side plugin",
					Description: "Enables jsonData.allowCleartextPasswords, required by some MySQL user accounts. The backend appends allowCleartextPasswords=true to the DSN at pkg/mysql/mysql.go:106-108.",
					Value: map[string]any{
						"url":  "mysql.internal:3306",
						"user": "grafana_reader",
						"jsonData": map[string]any{
							"database":                "metrics",
							"allowCleartextPasswords": true,
							"maxIdleConnsAuto":        true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"tunedConnectionPool": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Explicit connection-pool tuning",
					Description: "Override the Grafana-wide defaults for pool size and lifetime. When maxIdleConnsAuto is false, maxIdleConns must be set explicitly.",
					Value: map[string]any{
						"url":  "mysql.internal:3306",
						"user": "grafana_reader",
						"jsonData": map[string]any{
							"database":         "metrics",
							"maxOpenConns":     50,
							"maxIdleConns":     10,
							"maxIdleConnsAuto": false,
							"connMaxLifetime":  3600,
							"timeInterval":     "1m",
							"timezone":         "+00:00",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"legacyRootDatabase": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy: database stored at root level",
					Description: "Older datasources stored the database name at the root level rather than in jsonData. The backend still falls back to root.database when jsonData.database is empty (pkg/mysql/mysql.go:58-61); useMigrateDatabaseFields (@grafana/sql) migrates it on first render. New configurations should write jsonData.database.",
					Value: map[string]any{
						"url":      "localhost:3306",
						"user":     "grafana_reader",
						"database": "metrics",
						"jsonData": map[string]any{
							"maxIdleConnsAuto": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
		},
	}
}
