package postgresqldatasource

import (
	_ "embed"

	"k8s.io/kube-openapi/pkg/spec3"

	"github.com/grafana/dsconfig/dsconfig"
	sdkschema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

//go:generate go test -generateArtifacts -run TestSchemaConformance ./...

const TargetAPIVersion = dsconfig.TargetAPIVersion

//go:embed dsconfig.json
var configSchemaJSON []byte

func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: sslmode='require', tlsConfigurationMethod='file-path'. The user must still supply url, user, jsonData.database, and secureJsonData.password.",
					Value: map[string]any{
						"url":  "",
						"user": "",
						"jsonData": map[string]any{
							"sslmode":                string(TLSModeRequire),
							"tlsConfigurationMethod": string(TLSMethodFilePath),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "",
						},
					},
				},
			},
			"basicAuthNoTLS": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Username / password, TLS disabled",
					Description: "Local dev-style setup with sslmode='disable'. No certificate configuration required.",
					Value: map[string]any{
						"url":  "localhost:5432",
						"user": "grafana_reader",
						"jsonData": map[string]any{
							"database": "metrics",
							"sslmode":  string(TLSModeDisable),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"tlsRequireFilePath": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "sslmode=require with file-path certs",
					Description: "The typical production TLS setup: sslmode='require' asks for encryption but does not verify identity; certificate file paths are read from the Grafana host's local filesystem.",
					Value: map[string]any{
						"url":  "postgres.internal:5432",
						"user": "grafana_reader",
						"jsonData": map[string]any{
							"database":               "metrics",
							"sslmode":                string(TLSModeRequire),
							"tlsConfigurationMethod": string(TLSMethodFilePath),
							"sslCertFile":            "/etc/secrets/postgres/client.crt",
							"sslKeyFile":             "/etc/secrets/postgres/client.key",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"tlsVerifyFullFilePath": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "sslmode=verify-full with file-path certs",
					Description: "Strictest TLS mode — verifies the server certificate is signed by the supplied root CA AND that the server hostname matches the certificate.",
					Value: map[string]any{
						"url":  "postgres.internal:5432",
						"user": "grafana_reader",
						"jsonData": map[string]any{
							"database":               "metrics",
							"sslmode":                string(TLSModeVerifyFull),
							"tlsConfigurationMethod": string(TLSMethodFilePath),
							"sslRootCertFile":        "/etc/secrets/postgres/ca.crt",
							"sslCertFile":            "/etc/secrets/postgres/client.crt",
							"sslKeyFile":             "/etc/secrets/postgres/client.key",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"tlsVerifyCAInlineContent": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "sslmode=verify-ca with inline PEM content",
					Description: "TLS credentials stored as encrypted content in secureJsonData; the backend writes them to disk at connection time.",
					Value: map[string]any{
						"url":  "postgres.internal:5432",
						"user": "grafana_reader",
						"jsonData": map[string]any{
							"database":               "metrics",
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
			"timescaledb": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "TimescaleDB extension enabled",
					Description: "jsonData.timescaledb=true tells the query builder to use time_bucket in $__timeGroup and show TimescaleDB-specific aggregate functions.",
					Value: map[string]any{
						"url":  "timescale.internal:5432",
						"user": "grafana_reader",
						"jsonData": map[string]any{
							"database":        "metrics",
							"sslmode":         string(TLSModeRequire),
							"timescaledb":     true,
							"postgresVersion": 1500,
							"timeInterval":    "1m",
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
					Description: "Older datasources stored the database name at the root level. useMigrateDatabaseFields migrates it; the backend (pkg/postgresql/postgres.go:101-104) falls back to root.database when jsonData.database is empty. New configurations should write jsonData.database.",
					Value: map[string]any{
						"url":      "localhost:5432",
						"user":     "grafana_reader",
						"database": "metrics",
						"jsonData": map[string]any{
							"sslmode": string(TLSModeRequire),
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
