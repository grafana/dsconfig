package influxdbdatasource

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
// truth for the InfluxDB datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the InfluxDB datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the InfluxDB
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the InfluxDB
// datasource, covering the default configuration and each query language +
// authentication variant the config editor supports. Each example value is a
// full instance settings object with the plugin configuration nested under
// jsonData and the relevant write-only secrets under secureJsonData
// (placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: no authentication, InfluxQL query language, HTTP GET, maxSeries=1000. dbName is empty and must be provided to connect to an InfluxDB 1.x instance.",
					Value: map[string]any{
						"url": "http://localhost:8086",
						"jsonData": map[string]any{
							"version":   string(DefaultInfluxVersion),
							"httpMode":  string(DefaultInfluxHTTPMode),
							"maxSeries": int32(DefaultMaxSeries),
							"dbName":    "mydb",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"influxqlBasicAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "InfluxQL with HTTP Basic auth (InfluxDB OSS 1.x / 2.x)",
					Description: "InfluxQL against an InfluxDB 1.x/2.x server that accepts HTTP Basic auth. root.basicAuth=true wires the SDK transport to attach basicAuthUser + secureJsonData.basicAuthPassword to outgoing requests. Backend consumes jsonData.dbName as the ?db= URL parameter (influxql.go:176).",
					Value: map[string]any{
						"url":           "https://influxdb.example.com:8086",
						"basicAuth":     true,
						"basicAuthUser": "grafana",
						"jsonData": map[string]any{
							"version":   string(InfluxVersionInfluxQL),
							"product":   string(InfluxProductOSS1x),
							"httpMode":  string(InfluxHTTPModePOST),
							"dbName":    "telegraf",
							"maxSeries": int32(DefaultMaxSeries),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "REPLACE_WITH_PASSWORD",
						},
					},
				},
			},
			"influxqlLegacyUserPassword": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "InfluxQL with legacy user/password (v1 editor)",
					Description: "Legacy shape written by the v1 InfluxQL config tab (InfluxInfluxQLConfig.tsx:87-107): root.user + secureJsonData.password (distinct from basicAuthUser/basicAuthPassword). The current backend does not automatically attach these to outgoing HTTP requests; provisioning that relies on 1.x database auth should enable root.basicAuth as well.",
					Value: map[string]any{
						"url":  "http://influxdb.example.com:8086",
						"user": "admin",
						"jsonData": map[string]any{
							"version":     string(InfluxVersionInfluxQL),
							"httpMode":    string(InfluxHTTPModeGET),
							"dbName":      "telegraf",
							"showTagTime": "12h",
							"maxSeries":   int32(DefaultMaxSeries),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "REPLACE_WITH_PASSWORD",
						},
					},
				},
			},
			"fluxToken": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Flux with token (InfluxDB 2.x)",
					Description: "Flux against InfluxDB Cloud (TSM) or OSS 2.x. Requires jsonData.organization + jsonData.defaultBucket + secureJsonData.token. httpMode is POST (the v1 editor forces this at ConfigEditor.tsx:65). Note: the v1 editor also force-sets root.basicAuth=true on Flux selection (ConfigEditor.tsx:63) — this example uses provisioning-clean settings with basicAuth omitted; Flux auth flows exclusively through the bearer token regardless.",
					Value: map[string]any{
						"url": "https://us-west-2-1.aws.cloud2.influxdata.com",
						"jsonData": map[string]any{
							"version":       string(InfluxVersionFlux),
							"product":       string(InfluxProductCloudTSM),
							"httpMode":      string(InfluxHTTPModePOST),
							"organization":  "grafana-org",
							"defaultBucket": "metrics",
							"timeInterval":  "10s",
							"maxSeries":     int32(DefaultMaxSeries),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "REPLACE_WITH_TOKEN",
						},
					},
				},
			},
			"sqlFlightSQL": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "SQL (FlightSQL, InfluxDB 3.x)",
					Description: "Native SQL against an InfluxDB 3.x server or InfluxDB Cloud Serverless / Dedicated. Requires jsonData.dbName + secureJsonData.token. The FlightSQL transport is separate from the HTTP client — jsonData.insecureGrpc disables TLS on that gRPC transport.",
					Value: map[string]any{
						"url": "https://us-east-1-1.aws.cloud2.influxdata.com",
						"jsonData": map[string]any{
							"version":      string(InfluxVersionSQL),
							"product":      string(InfluxProductCloudServerless),
							"dbName":       "metrics",
							"insecureGrpc": false,
							"maxSeries":    int32(DefaultMaxSeries),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "REPLACE_WITH_TOKEN",
						},
					},
				},
			},
			"tlsMutualAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "TLS mutual auth (mTLS)",
					Description: "InfluxDB behind mTLS. jsonData.tlsAuth=true triggers the SDK to build a client-authenticated transport using serverName plus the PEM-encoded client cert and key in secureJsonData. Applies to any query language.",
					Value: map[string]any{
						"url": "https://influxdb.example.com:8086",
						"jsonData": map[string]any{
							"version":    string(InfluxVersionInfluxQL),
							"httpMode":   string(InfluxHTTPModeGET),
							"dbName":     "telegraf",
							"maxSeries":  int32(DefaultMaxSeries),
							"tlsAuth":    true,
							"serverName": "influxdb.example.com",
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
					Description: "InfluxDB behind a private CA. jsonData.tlsAuthWithCACert=true tells the SDK to verify the server certificate against the PEM-encoded CA in secureJsonData.tlsCACert.",
					Value: map[string]any{
						"url": "https://influxdb.internal.corp",
						"jsonData": map[string]any{
							"version":           string(InfluxVersionInfluxQL),
							"httpMode":          string(InfluxHTTPModeGET),
							"dbName":            "telegraf",
							"maxSeries":         int32(DefaultMaxSeries),
							"tlsAuthWithCACert": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyTLSCACert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
						},
					},
				},
			},
			"oauthForward": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Forward OAuth Identity",
					Description: "Forward the signed-in user's upstream OAuth identity to InfluxDB. Independent of the query-language selection.",
					Value: map[string]any{
						"url": "https://influxdb.example.com",
						"jsonData": map[string]any{
							"version":       string(InfluxVersionInfluxQL),
							"httpMode":      string(InfluxHTTPModeGET),
							"dbName":        "telegraf",
							"maxSeries":     int32(DefaultMaxSeries),
							"oauthPassThru": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"legacyRootDatabase": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy root.database fallback",
					Description: "Pre-jsonData datasources stored the database name at the root level. The backend uses jsonData.dbName if set, else settings.Database (pkg/influxdb/influxdb.go:58-61). This example represents a legacy provisioning payload with no jsonData.dbName; ApplyDefaults copies root.database into DbName so callers see a consistent shape.",
					Value: map[string]any{
						"url":      "http://influxdb.example.com:8086",
						"database": "legacy_db",
						"jsonData": map[string]any{
							"version":   string(InfluxVersionInfluxQL),
							"httpMode":  string(InfluxHTTPModeGET),
							"maxSeries": int32(DefaultMaxSeries),
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
