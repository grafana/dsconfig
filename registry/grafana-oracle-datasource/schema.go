package oracledatasource

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
// truth for the Oracle Database datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Oracle Database datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Oracle Database
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Oracle
// Database datasource, covering the default configuration and each combination
// of connection method (Host with TCP Port / TNSNames Entry) and authentication
// method (Basic / Kerberos) the config editor supports, plus the legacy
// TNSNames-in-URL storage shape. Each example value is a full instance settings
// object with the root url at the top level, plugin configuration nested under
// jsonData, and the write-only password under secureJsonData (placeholder values
// — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: Host with TCP Port connection, Basic authentication, UTC session time zone, and the backend connection-pool/timeout/row-limit defaults. The user must still supply url, jsonData.user, jsonData.database, and secureJsonData.password to get a working datasource.",
					Value: map[string]any{
						"url": "",
						"jsonData": map[string]any{
							"useTNSNamesBasedConnection": false,
							"useKerberosAuthentication":  false,
							"user":                       "",
							"database":                   "",
							"timezone_name":              DefaultTimezone,
							"connectionPoolSize":         DefaultConnectionPoolSize,
							"dataProxyTimeout":           DefaultDataProxyTimeout,
							"rowLimit":                   DefaultRowLimit,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "",
						},
					},
				},
			},
			"basicAuthTcp": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic auth over Host with TCP Port",
					Description: "The typical setup: host+port in the root url, a database name, and a password-authenticated Oracle user. useTNSNamesBasedConnection and useKerberosAuthentication default to false.",
					Value: map[string]any{
						"url": "oracle.example.com:1521",
						"jsonData": map[string]any{
							"database":      "ORCLPDB1",
							"user":          "grafana_reader",
							"timezone_name": DefaultTimezone,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"basicAuthTns": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic auth over a TNSNames entry",
					Description: "Connect using a tnsnames.ora entry (jsonData.tnsNamesEntry) with a password-authenticated user. No root url is needed. TNSNames connections are not supported in Grafana Cloud.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"useTNSNamesBasedConnection": true,
							"tnsNamesEntry":              "(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=oracle-db1)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=FREEPDB1)(SERVER=DEDICATED)))",
							"user":                       "grafana_reader",
							"timezone_name":              DefaultTimezone,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"kerberosTns": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Kerberos auth over a TNSNames entry",
					Description: "Kerberos authentication uses no username or password (they come from the Kerberos ticket / tnsnames.ora). The editor only exposes Kerberos when useTNSNamesBasedConnection is true. Neither TNSNames nor Kerberos is supported in Grafana Cloud.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"useTNSNamesBasedConnection": true,
							"useKerberosAuthentication":  true,
							"tnsNamesEntry":              "(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=oracle-db1)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=FREEPDB1)(SERVER=DEDICATED)))",
							"timezone_name":              DefaultTimezone,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "",
						},
					},
				},
			},
			"tunedSettings": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic auth with tuned Additional Settings",
					Description: "Override the optional performance knobs: a larger connection pool, a longer dataproxy timeout, driver row prefetching, a custom row limit, and a non-UTC display time zone.",
					Value: map[string]any{
						"url": "oracle.example.com:1521",
						"jsonData": map[string]any{
							"database":           "ORCLPDB1",
							"user":               "grafana_reader",
							"timezone_name":      "Europe/Berlin",
							"connectionPoolSize": 100,
							"dataProxyTimeout":   200,
							"prefetchRowsCount":  500,
							"rowLimit":           500000,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"legacyTnsInUrl": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy: TNSNames entry stored in the root url",
					Description: "v3.3.0 stored the TNSNames entry in the root url before v3.3.2 moved it to jsonData.tnsNamesEntry. When useTNSNamesBasedConnection is true and jsonData.tnsNamesEntry is empty, the backend falls back to the root url (pkg/models/settings.go:82-84).",
					Value: map[string]any{
						"url": "(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=oracle-db1)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=FREEPDB1)(SERVER=DEDICATED)))",
						"jsonData": map[string]any{
							"useTNSNamesBasedConnection": true,
							"user":                       "grafana_reader",
							"timezone_name":              DefaultTimezone,
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
