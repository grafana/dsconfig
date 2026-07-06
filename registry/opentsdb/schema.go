package opentsdbdatasource

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
// truth for the OpenTSDB datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the OpenTSDB datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the OpenTSDB
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the OpenTSDB
// datasource, covering the default configuration and each authentication /
// TLS / behaviour variant the config editor supports. Each example value is a
// full instance settings object with the plugin configuration nested under
// jsonData and the relevant write-only secrets under secureJsonData
// (placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: no authentication, no TLS overrides, tsdbVersion=1 (<=2.1), tsdbResolution=1 (second), lookupLimit=1000. Only root.url needs to be filled in to get a working datasource pointed at an OpenTSDB HTTP API on localhost:4242.",
					Value: map[string]any{
						"url": "http://localhost:4242",
						"jsonData": map[string]any{
							"tsdbVersion":    float32(DefaultOpenTsdbVersion),
							"tsdbResolution": int32(DefaultOpenTsdbResolution),
							"lookupLimit":    DefaultLookupLimit,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"noAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "No authentication (OpenTSDB 2.4)",
					Description: "A public or network-isolated OpenTSDB 2.4 server. tsdbVersion=4 unlocks array-response parsing (pkg/opentsdb/utils.go:138,247-254) and adds ?arrays=true to /api/query.",
					Value: map[string]any{
						"url": "http://opentsdb.example.com:4242",
						"jsonData": map[string]any{
							"tsdbVersion":    float32(OpenTsdbVersion24),
							"tsdbResolution": int32(OpenTsdbResolutionSecond),
							"lookupLimit":    int32(1000),
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
					Description: "Authenticate with an OpenTSDB that requires HTTP Basic. Both basicAuth and basicAuthUser live at the datasource root; only basicAuthPassword is secret.",
					Value: map[string]any{
						"url":           "https://opentsdb.example.com",
						"basicAuth":     true,
						"basicAuthUser": "grafana",
						"jsonData": map[string]any{
							"tsdbVersion":    float32(OpenTsdbVersionLTE21),
							"tsdbResolution": int32(OpenTsdbResolutionSecond),
							"lookupLimit":    int32(1000),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "REPLACE_WITH_PASSWORD",
						},
					},
				},
			},
			"oauthForward": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Forward OAuth Identity",
					Description: "Forward the signed-in user's upstream OAuth identity to OpenTSDB. The editor writes only jsonData.oauthPassThru; there is no accompanying secret.",
					Value: map[string]any{
						"url": "https://opentsdb.example.com",
						"jsonData": map[string]any{
							"tsdbVersion":    float32(OpenTsdbVersionLTE21),
							"tsdbResolution": int32(OpenTsdbResolutionSecond),
							"lookupLimit":    int32(1000),
							"oauthPassThru":  true,
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
					Description: "OpenTSDB requiring mTLS. jsonData.tlsAuth=true triggers the SDK to build a client-authenticated transport using serverName plus the PEM-encoded client cert and key in secureJsonData.",
					Value: map[string]any{
						"url": "https://opentsdb.example.com",
						"jsonData": map[string]any{
							"tsdbVersion":    float32(OpenTsdbVersionLTE21),
							"tsdbResolution": int32(OpenTsdbResolutionSecond),
							"lookupLimit":    int32(1000),
							"tlsAuth":        true,
							"serverName":     "opentsdb.example.com",
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
					Description: "OpenTSDB behind a private CA. jsonData.tlsAuthWithCACert=true tells the SDK to verify the server certificate against the PEM-encoded CA in secureJsonData.tlsCACert.",
					Value: map[string]any{
						"url": "https://opentsdb.internal.corp",
						"jsonData": map[string]any{
							"tsdbVersion":       float32(OpenTsdbVersionLTE21),
							"tsdbResolution":    int32(OpenTsdbResolutionSecond),
							"lookupLimit":       int32(1000),
							"tlsAuthWithCACert": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyTLSCACert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
						},
					},
				},
			},
			"millisecondResolution": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Millisecond resolution",
					Description: "OpenTSDB configured to serve millisecond-precision timestamps. tsdbResolution=2 causes the frontend to add msResolution=true to outgoing query bodies (src/datasource.ts:178-180) and the response parser to treat dps keys as millisecond epochs directly.",
					Value: map[string]any{
						"url": "http://opentsdb.example.com:4242",
						"jsonData": map[string]any{
							"tsdbVersion":    float32(OpenTsdbVersion23),
							"tsdbResolution": int32(OpenTsdbResolutionMillisecond),
							"lookupLimit":    int32(1000),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"largeLookupLimit": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Large tag-value autocomplete limit",
					Description: "Raise the /api/search/lookup response cap used for tag-value autocomplete. lookupLimit=10000 tells HandleKeyValueLookup (pkg/opentsdb/callresource.go:361) to request up to 10000 rows from OpenTSDB for each keyvalue lookup.",
					Value: map[string]any{
						"url": "http://opentsdb.example.com:4242",
						"jsonData": map[string]any{
							"tsdbVersion":    float32(OpenTsdbVersion24),
							"tsdbResolution": int32(OpenTsdbResolutionSecond),
							"lookupLimit":    int32(10000),
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
