package prometheusdatasource

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
// truth for the Prometheus datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Prometheus datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Prometheus
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Prometheus
// datasource, covering the default configuration and each authentication
// method and TLS variant the config editor supports. Each example value is a
// full instance settings object with the plugin configuration nested under
// jsonData and the relevant write-only secrets under secureJsonData
// (placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: no authentication, POST as the HTTP method. Only root.url needs to be filled in to get a working datasource pointed at a Prometheus server on localhost:9090.",
					Value: map[string]any{
						"url": "http://localhost:9090",
						"jsonData": map[string]any{
							"httpMethod": string(HTTPMethodPOST),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"noAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "No authentication",
					Description: "A public or network-isolated Prometheus with no HTTP-level auth. Prometheus type and version are set explicitly so the query editor exposes flavour-specific query hints.",
					Value: map[string]any{
						"url": "http://prometheus.example.com:9090",
						"jsonData": map[string]any{
							"httpMethod":        string(HTTPMethodPOST),
							"prometheusType":    string(PromApplicationPrometheus),
							"prometheusVersion": "2.50.1",
							"timeInterval":      "15s",
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
					Description: "Authenticate with a Prometheus that requires HTTP Basic. Both basicAuth and basicAuthUser live at the datasource root; only basicAuthPassword is secret.",
					Value: map[string]any{
						"url":           "https://prometheus.example.com",
						"basicAuth":     true,
						"basicAuthUser": "grafana",
						"jsonData": map[string]any{
							"httpMethod": string(HTTPMethodPOST),
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
					Description: "Forward the signed-in user's upstream OAuth identity to Prometheus. The editor writes only jsonData.oauthPassThru; there is no accompanying secret.",
					Value: map[string]any{
						"url": "https://prometheus.example.com",
						"jsonData": map[string]any{
							"httpMethod":    string(HTTPMethodPOST),
							"oauthPassThru": true,
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
					Description: "Prometheus requiring mTLS. jsonData.tlsAuth=true triggers the SDK to build a client-authenticated transport using serverName plus the PEM-encoded client cert and key in secureJsonData.",
					Value: map[string]any{
						"url": "https://prometheus.example.com",
						"jsonData": map[string]any{
							"httpMethod": string(HTTPMethodPOST),
							"tlsAuth":    true,
							"serverName": "prometheus.example.com",
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
					Description: "Prometheus behind a private CA. jsonData.tlsAuthWithCACert=true tells the SDK to verify the server certificate against the PEM-encoded CA in secureJsonData.tlsCACert.",
					Value: map[string]any{
						"url": "https://prometheus.internal.corp",
						"jsonData": map[string]any{
							"httpMethod":        string(HTTPMethodPOST),
							"tlsAuthWithCACert": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyTLSCACert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
						},
					},
				},
			},
			"getHTTPMethod": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "GET HTTP method (legacy or restricted networks)",
					Description: "Older Prometheus (< 2.1) or environments that block POST. Everything else stays default. seriesEndpoint is enabled to prefer /api/v1/series (which supports POST) over /api/v1/label/*/values.",
					Value: map[string]any{
						"url": "http://legacy-prom.example.com:9090",
						"jsonData": map[string]any{
							"httpMethod":     string(HTTPMethodGET),
							"seriesEndpoint": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"mimirWithExemplars": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Mimir with exemplar drilldown",
					Description: "Grafana Mimir with exemplar trace-ID destinations wired to a Tempo data source. name is the label carrying the trace ID; datasourceUid takes precedence over url when set.",
					Value: map[string]any{
						"url":           "https://mimir.example.com/prometheus",
						"basicAuth":     true,
						"basicAuthUser": "grafana",
						"jsonData": map[string]any{
							"httpMethod":          string(HTTPMethodPOST),
							"prometheusType":      string(PromApplicationMimir),
							"prometheusVersion":   "2.9.1",
							"timeInterval":        "15s",
							"cacheLevel":          string(PrometheusCacheLevelMedium),
							"incrementalQuerying": true,
							"exemplarTraceIdDestinations": []any{
								map[string]any{
									"name":          "traceID",
									"datasourceUid": "tempo",
								},
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "REPLACE_WITH_PASSWORD",
						},
					},
				},
			},
		},
	}
}
