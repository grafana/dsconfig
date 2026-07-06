package tempodatasource

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
// truth for the Tempo datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Tempo datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Tempo
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Tempo
// datasource, covering the default configuration and each authentication
// method / TLS variant / Tempo-specific feature (streaming, service graph,
// trace-to-logs V2, trace-to-metrics, trace-to-profiles, node graph, span
// bar, TraceID query, TraceQL search) the config editor supports. Each
// example value is a full instance settings object with the plugin
// configuration nested under jsonData and the relevant write-only secrets
// under secureJsonData (placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: no authentication, no TLS overrides, no streaming, no trace-to-X mappings. Only root.url needs to be filled in to get a working datasource pointed at a Tempo server on localhost:3200.",
					Value: map[string]any{
						"url":      "http://localhost:3200",
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
					Description: "A public or network-isolated Tempo with no HTTP-level auth. Node graph is enabled for a richer trace view.",
					Value: map[string]any{
						"url": "http://tempo.example.com:3200",
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
					Description: "Authenticate with a Tempo that requires HTTP Basic. Both basicAuth and basicAuthUser live at the datasource root; only basicAuthPassword is secret. Basic auth is also injected as gRPC per-RPC credentials (pkg/tempo/grpc.go:178-184).",
					Value: map[string]any{
						"url":           "https://tempo.example.com",
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
					Description: "Forward the signed-in user's upstream OAuth identity to Tempo. The editor writes only jsonData.oauthPassThru; there is no accompanying secret.",
					Value: map[string]any{
						"url": "https://tempo.example.com",
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
					Description: "Tempo requiring mTLS. jsonData.tlsAuth=true triggers the SDK to build a client-authenticated transport using serverName plus the PEM-encoded client cert and key in secureJsonData; the same TLS config is picked up by the gRPC client (pkg/tempo/grpc.go:186-197).",
					Value: map[string]any{
						"url": "https://tempo.example.com",
						"jsonData": map[string]any{
							"tlsAuth":    true,
							"serverName": "tempo.example.com",
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
					Description: "Tempo behind a private CA. jsonData.tlsAuthWithCACert=true tells the SDK to verify the server certificate against the PEM-encoded CA in secureJsonData.tlsCACert.",
					Value: map[string]any{
						"url": "https://tempo.internal.corp",
						"jsonData": map[string]any{
							"tlsAuthWithCACert": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyTLSCACert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
						},
					},
				},
			},
			"streaming": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Streaming enabled (search + metrics)",
					Description: "TraceQL search and metrics streaming both enabled. Requires Tempo >= 2.7.0 (search alone needs 2.2.0). CheckHealth (pkg/tempo/tempo.go:150-208) probes the gRPC streaming endpoint when streamingEnabled.search is true.",
					Value: map[string]any{
						"url": "https://tempo.example.com",
						"jsonData": map[string]any{
							"streamingEnabled": map[string]any{
								"search":  true,
								"metrics": true,
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
					Summary:     "Full observability wiring (trace ↔ logs / metrics / profiles / service graph)",
					Description: "Tempo wired to Loki (logs), Prometheus (metrics + service graph), and Pyroscope (profiles). Uses the v2 tracesToLogs shape; nodeGraph, spanBar='Duration', and Tempo search are enabled.",
					Value: map[string]any{
						"url": "https://tempo.example.com",
						"jsonData": map[string]any{
							"nodeGraph": map[string]any{
								"enabled": true,
							},
							"spanBar": map[string]any{
								"type": string(SpanBarTypeDuration),
							},
							"serviceMap": map[string]any{
								"datasourceUid": "prometheus",
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
							"tracesToProfiles": map[string]any{
								"datasourceUid": "pyroscope",
								"profileTypeId": "process_cpu:cpu:nanoseconds:cpu:nanoseconds",
								"tags": []any{
									map[string]any{"key": "service.name"},
								},
								"customQuery": false,
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"traceQLSearchAndTraceID": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "TraceQL search tuning and TraceID time-range shifts",
					Description: "TraceQL search filters seeded, tag time-range extended to 3 hours, tag limit raised to 10000, and TraceID queries time-shifted so long-running traces fall inside the search window.",
					Value: map[string]any{
						"url": "https://tempo.example.com",
						"jsonData": map[string]any{
							"timeRangeForTags": TimeRangeForTagsLast3Hours,
							"tagLimit":         10000,
							"search": map[string]any{
								"hide": false,
								"filters": []any{
									map[string]any{
										"id":       "service-name",
										"tag":      "service.name",
										"operator": "=",
										"scope":    string(TraceqlSearchScopeResource),
									},
									map[string]any{
										"id":       "span-name",
										"tag":      "name",
										"operator": "=",
										"scope":    string(TraceqlSearchScopeSpan),
									},
								},
							},
							"traceQuery": map[string]any{
								"timeShiftEnabled":   true,
								"spanStartTimeShift": "30m",
								"spanEndTimeShift":   "30m",
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
						"url": "https://tempo.example.com",
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
