package jaegerdatasource

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
// truth for the Jaeger datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Jaeger datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Jaeger
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Jaeger
// datasource, covering the default configuration and each authentication
// method / TLS variant / Jaeger-specific feature the config editor supports.
// Each example value is a full instance settings object with the plugin
// configuration nested under jsonData and the relevant write-only secrets
// under secureJsonData (placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: no authentication, no TLS overrides, no trace-to-X mappings, trace-by-ID time parameters disabled. Only root.url needs to be filled in to get a working datasource pointed at a Jaeger query server on localhost:16686.",
					Value: map[string]any{
						"url":      "http://localhost:16686",
						"jsonData": map[string]any{},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"noAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "No authentication",
					Description: "A public or network-isolated Jaeger with no HTTP-level auth. Node graph is enabled for a richer trace view.",
					Value: map[string]any{
						"url": "http://jaeger.example.com:16686",
						"jsonData": map[string]any{
							"nodeGraph": map[string]any{
								"enabled": true,
							},
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
					Description: "Authenticate with a Jaeger that requires HTTP Basic. Both basicAuth and basicAuthUser live at the datasource root; only basicAuthPassword is secret. The SDK's HTTPClientOptions wires the credentials automatically (pkg/jaeger/jaeger.go:28).",
					Value: map[string]any{
						"url":           "https://jaeger.example.com",
						"basicAuth":     true,
						"basicAuthUser": "grafana",
						"jsonData":      map[string]any{},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "REPLACE_WITH_PASSWORD",
						},
					},
				},
			},
			"oauthForward": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Forward OAuth Identity",
					Description: "Forward the signed-in user's upstream OAuth identity to Jaeger. The editor writes only jsonData.oauthPassThru; there is no accompanying secret.",
					Value: map[string]any{
						"url": "https://jaeger.example.com",
						"jsonData": map[string]any{
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
					Description: "Jaeger requiring mTLS. jsonData.tlsAuth=true triggers the SDK to build a client-authenticated transport using serverName plus the PEM-encoded client cert and key in secureJsonData.",
					Value: map[string]any{
						"url": "https://jaeger.example.com",
						"jsonData": map[string]any{
							"tlsAuth":    true,
							"serverName": "jaeger.example.com",
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
					Description: "Jaeger behind a private CA. jsonData.tlsAuthWithCACert=true tells the SDK to verify the server certificate against the PEM-encoded CA in secureJsonData.tlsCACert.",
					Value: map[string]any{
						"url": "https://jaeger.internal.corp",
						"jsonData": map[string]any{
							"tlsAuthWithCACert": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyTLSCACert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
						},
					},
				},
			},
			"traceIdTimeParams": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Query trace by ID with time parameters",
					Description: "Enable the plugin-local traceIdTimeParams toggle so the backend appends start / end query parameters to GET /api/traces/{traceID}. Useful when Jaeger's storage backend requires a time hint to locate long-retained traces (pkg/jaeger/client.go:242-266).",
					Value: map[string]any{
						"url": "https://jaeger.example.com",
						"jsonData": map[string]any{
							"traceIdTimeParams": map[string]any{
								"enabled": true,
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"fullObservability": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Full observability wiring (trace ↔ logs / metrics)",
					Description: "Jaeger wired to Loki (logs) and Prometheus (metrics). Uses the v2 tracesToLogs shape; nodeGraph and spanBar='Duration' are enabled. Note that Jaeger, unlike Tempo, has no trace-to-profiles or service-graph section in the editor.",
					Value: map[string]any{
						"url": "https://jaeger.example.com",
						"jsonData": map[string]any{
							"nodeGraph": map[string]any{
								"enabled": true,
							},
							"spanBar": map[string]any{
								"type": string(SpanBarTypeDuration),
							},
							"tracesToLogsV2": map[string]any{
								"datasourceUid": "loki",
								"tags": []any{
									map[string]any{"key": "service.name", "value": "service_name"},
									map[string]any{"key": "cluster"},
								},
								"spanStartTimeShift": "-1m",
								"spanEndTimeShift":   "1m",
								"filterByTraceID":    true,
								"filterBySpanID":     false,
								"customQuery":        false,
							},
							"tracesToMetrics": map[string]any{
								"datasourceUid": "prometheus",
								"tags": []any{
									map[string]any{"key": "service.name", "value": "service"},
								},
								"queries": []any{
									map[string]any{
										"name":  "Error rate",
										"query": `sum(rate(traces_spanmetrics_calls_total{status_code="STATUS_CODE_ERROR", $__tags}[$__rate_interval]))`,
									},
									map[string]any{
										"name":  "P95 latency",
										"query": `histogram_quantile(0.95, sum(rate(traces_spanmetrics_latency_bucket{$__tags}[$__rate_interval])) by (le))`,
									},
								},
								"spanStartTimeShift": "-2m",
								"spanEndTimeShift":   "2m",
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"legacyTracesToLogs": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy tracesToLogs (v1) — migrated on read",
					Description: "A datasource stored before the v2 shape landed. `getTraceToLogsOptions` transforms tracesToLogs → tracesToLogsV2 at editor load; the next save wipes tracesToLogs entirely. Kept as an example so provisioning tooling knows the shape is still round-trippable.",
					Value: map[string]any{
						"url": "https://jaeger.example.com",
						"jsonData": map[string]any{
							"tracesToLogs": map[string]any{
								"datasourceUid":      "loki",
								"tags":               []any{"service.name", "cluster"},
								"mapTagNamesEnabled": false,
								"spanStartTimeShift": "-1m",
								"spanEndTimeShift":   "1m",
								"filterByTraceID":    true,
							},
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
