package odbcdatasource

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
// truth for the Sqlyze (ODBC) datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Sqlyze (ODBC) datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Sqlyze (ODBC)
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Sqlyze
// (ODBC) datasource, covering the default configuration and each connection
// variant (driver path, driver alias, and backend-only DSN connection string).
// Each example value is a full instance settings object with the plugin
// configuration nested under jsonData and the relevant write-only secret under
// secureJsonData (placeholder values — replace them with real secrets).
//
// Secret keys are dynamic (each equals a secure setting's Name); the examples
// use the conventional password key 'pwd'.
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: timeout '10' and no driver. jsonData.driver is required, so this default fails validation until a driver alias or driver path is supplied.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"timeout": "10",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPwd): "",
						},
					},
				},
			},
			"driverPathDB2": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Driver path (DB2) with a secure password",
					Description: "jsonData.driver is an absolute path to the ODBC driver shared library; connection parameters are supplied as driver settings. The 'pwd' setting is marked secure, so its value lives in secureJsonData.pwd and its settings[] entry carries only name+secure.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"driver":  "/opt/db2/clidriver/lib/libdb2.so.1",
							"timeout": "10",
							"settings": []any{
								map[string]any{"name": "host", "value": "127.0.0.1", "secure": false},
								map[string]any{"name": "port", "value": "50000", "secure": false},
								map[string]any{"name": "database", "value": "sample", "secure": false},
								map[string]any{"name": "uid", "value": "db2inst1", "secure": false},
								map[string]any{"name": "pwd", "secure": true},
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPwd): "<your-password>",
						},
					},
				},
			},
			"driverAlias": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Driver alias ({...} form)",
					Description: "jsonData.driver is a driver alias in braces (resolved from the host's odbcinst.ini). The connection string becomes 'Driver={MySQLDB};uid=grafana;pwd=<secret>;'.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"driver":  "{MySQLDB}",
							"timeout": "10",
							"settings": []any{
								map[string]any{"name": "uid", "value": "grafana", "secure": false},
								map[string]any{"name": "pwd", "secure": true},
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPwd): "<your-password>",
						},
					},
				},
			},
			"connectionStringDSN": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Backend-only DSN connection string",
					Description: "The backend-only jsonData.DSN field replaces the 'Driver=<driver>;' prefix with 'DSN=<DSN>;' (pkg/database/connect.go:76-78). jsonData.driver must still be non-empty — settings load rejects an empty driver before the connection string is built, so a DSN-only config cannot connect.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"driver":  "{TESTDB}",
							"DSN":     "TESTDB",
							"timeout": "10",
							"settings": []any{
								map[string]any{"name": "uid", "value": "db2inst1", "secure": false},
								map[string]any{"name": "pwd", "secure": true},
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPwd): "<your-password>",
						},
					},
				},
			},
		},
	}
}
