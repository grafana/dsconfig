package clickhousedatasource

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
// truth for the ClickHouse datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the ClickHouse datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the ClickHouse
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the
// ClickHouse datasource, covering the default (schema-defaults) configuration
// and one example per major protocol / TLS / OTel / single-table variant.
// Each example value is a full instance-settings object with the plugin
// configuration nested under jsonData and the relevant write-only secrets
// under secureJsonData (placeholder values — replace with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: native protocol, insecure, classic mode. The user must still supply host, port, username, and secureJsonData.password to get a working datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"protocol":               string(ProtocolNative),
							"secure":                 false,
							"configMode":             string(ConfigModeClassic),
							"enableMapKeysDiscovery": true,
							"dialTimeout":            "10",
							"queryTimeout":           "60",
							"connMaxLifetime":        "5",
							"maxIdleConns":           "25",
							"maxOpenConns":           "50",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "",
						},
					},
				},
			},
			"nativeInsecure": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Native TCP, insecure",
					Description: "Typical local development setup: ClickHouse native protocol on port 9000, plain TCP, username/password auth.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"host":                   "localhost",
							"port":                   9000,
							"protocol":               string(ProtocolNative),
							"secure":                 false,
							"username":               "default",
							"configMode":             string(ConfigModeClassic),
							"enableMapKeysDiscovery": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"nativeSecure": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Native TCP, TLS",
					Description: "ClickHouse Cloud-style setup: native protocol on port 9440 with TLS.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"host":                   "my-cluster.clickhouse.cloud",
							"port":                   9440,
							"protocol":               string(ProtocolNative),
							"secure":                 true,
							"username":               "default",
							"configMode":             string(ConfigModeClassic),
							"enableMapKeysDiscovery": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"httpSecureWithHeaders": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "HTTPS with custom headers",
					Description: "HTTP protocol on 8443 with TLS. Adds one plaintext HTTP header and one secure HTTP header — secure header values are stored under secureJsonData['secureHttpHeaders.<Header Name>'] and the value in jsonData.httpHeaders is left empty.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"host":                  "clickhouse.internal",
							"port":                  8443,
							"protocol":              string(ProtocolHTTP),
							"secure":                true,
							"path":                  "clickhouse",
							"username":              "default",
							"forwardGrafanaHeaders": true,
							"httpHeaders": []any{
								map[string]any{"name": "X-ClickHouse-User", "value": "grafana", "secure": false},
								map[string]any{"name": "X-Api-Key", "value": "", "secure": true},
							},
							"configMode":             string(ConfigModeClassic),
							"enableMapKeysDiscovery": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
							"secureHttpHeaders.X-Api-Key":     "abcd1234",
						},
					},
				},
			},
			"tlsClientAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "mTLS with a client certificate",
					Description: "Enables jsonData.tlsAuth so the driver supplies a client certificate + key to the server. Populate both secureJsonData.tlsClientCert and secureJsonData.tlsClientKey as full PEM blocks including BEGIN/END lines.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"host":                   "clickhouse.internal",
							"port":                   9440,
							"protocol":               string(ProtocolNative),
							"secure":                 true,
							"username":               "default",
							"tlsAuth":                true,
							"configMode":             string(ConfigModeClassic),
							"enableMapKeysDiscovery": true,
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
					Description: "Enables jsonData.tlsAuthWithCACert so the driver verifies the server's certificate against secureJsonData.tlsCACert. Independent of tlsAuth.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"host":                   "clickhouse.internal",
							"port":                   9440,
							"protocol":               string(ProtocolNative),
							"secure":                 true,
							"username":               "default",
							"tlsAuthWithCACert":      true,
							"configMode":             string(ConfigModeClassic),
							"enableMapKeysDiscovery": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword):  "changeme",
							string(SecureJsonDataKeyTLSCACert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----\n",
						},
					},
				},
			},
			"otelLogsSingleTable": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Single-table logs with OTel schema",
					Description: "configMode='single-table' + signalType='logs' pins the datasource to one table. Enabling logs.otelEnabled and picking an otelVersion makes the plugin fill every column-role field from the OTel column map at runtime.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"host":                   "clickhouse.internal",
							"port":                   9440,
							"protocol":               string(ProtocolNative),
							"secure":                 true,
							"username":               "default",
							"configMode":             string(ConfigModeSingleTable),
							"signalType":             string(SignalTypeLogs),
							"enableMapKeysDiscovery": true,
							"logs": map[string]any{
								"defaultDatabase":      "otel",
								"defaultTable":         "otel_logs",
								"otelEnabled":          true,
								"otelVersion":          "latest",
								"selectContextColumns": true,
								"showLogLinks":         true,
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"otelTracesSingleTable": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Single-table traces with OTel schema",
					Description: "configMode='single-table' + signalType='traces'. The plugin defaults traces.defaultTable to otel_traces and, with traces.otelEnabled=true, drives every column-role and duration unit from the OTel column map.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"host":                   "clickhouse.internal",
							"port":                   9440,
							"protocol":               string(ProtocolNative),
							"secure":                 true,
							"username":               "default",
							"configMode":             string(ConfigModeSingleTable),
							"signalType":             string(SignalTypeTraces),
							"enableMapKeysDiscovery": true,
							"traces": map[string]any{
								"defaultDatabase":           "otel",
								"defaultTable":              "otel_traces",
								"otelEnabled":               true,
								"otelVersion":               "latest",
								"durationUnit":              string(TraceDurationUnitNanoseconds),
								"showTraceLinks":            true,
								"traceTimestampTableSuffix": "_trace_id_ts",
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"legacyV3ServerField": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy: v3 `server` / `timeout` fields",
					Description: "Datasources created before ClickHouse plugin v4 stored the host under jsonData.server and the dial timeout under jsonData.timeout. The backend still maps them into settings.Host and settings.DialTimeout at load time (pkg/plugin/settings.go:86-92,159-168); the config editor migrates them on first render. New configurations should use host and dialTimeout.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"server":                 "legacy-clickhouse.internal",
							"port":                   9000,
							"protocol":               string(ProtocolNative),
							"timeout":                10,
							"username":               "default",
							"configMode":             string(ConfigModeClassic),
							"enableMapKeysDiscovery": true,
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
