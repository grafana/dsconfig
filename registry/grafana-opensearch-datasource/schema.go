package opensearchdatasource

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
// truth for the OpenSearch datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the OpenSearch datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the OpenSearch
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the OpenSearch
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
					Description: "The defaults a new OpenSearch datasource ends up with after `coerceOptions` runs: timeField '@timestamp', maxConcurrentShardRequests 5 (OpenSearch flavor), pplEnabled true, and a placeholder index. Fill in root.url + a concrete jsonData.database + jsonData.flavor + jsonData.version to get a working datasource.",
					Value: map[string]any{
						"url":    "http://localhost:9200",
						"access": string(AccessModeProxy),
						"jsonData": map[string]any{
							"flavor":                     string(FlavorOpenSearch),
							"version":                    "1.0.0",
							"database":                   "es-index-name",
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequestsOpenSearch,
							"pplEnabled":                 true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"noAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "OpenSearch 1.x, no authentication",
					Description: "A network-isolated OpenSearch cluster with no HTTP-level auth. Index uses a daily logstash-style time pattern.",
					Value: map[string]any{
						"url":    "http://opensearch.example.com:9200",
						"access": string(AccessModeProxy),
						"jsonData": map[string]any{
							"flavor":                     string(FlavorOpenSearch),
							"version":                    "2.11.0",
							"database":                   "[logstash-]YYYY.MM.DD",
							"interval":                   string(IntervalDaily),
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequestsOpenSearch,
							"pplEnabled":                 true,
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
					Description: "Authenticate with an OpenSearch cluster that requires HTTP Basic. `basicAuth` and `basicAuthUser` live at the datasource root; only `basicAuthPassword` is secret.",
					Value: map[string]any{
						"url":           "https://opensearch.example.com:9200",
						"access":        string(AccessModeProxy),
						"basicAuth":     true,
						"basicAuthUser": "grafana",
						"jsonData": map[string]any{
							"flavor":                     string(FlavorOpenSearch),
							"version":                    "2.11.0",
							"database":                   "grafana-logs",
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequestsOpenSearch,
							"pplEnabled":                 true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "REPLACE_WITH_PASSWORD",
						},
					},
				},
			},
			"sigV4Managed": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "SigV4 authentication (AWS Managed OpenSearch)",
					Description: "AWS SigV4-signed requests to an AWS-managed OpenSearch domain. The sigV4* jsonData fields are contributed by @grafana/aws-sdk's SIGV4ConnectionConfig; the backend forces `httpCliOpts.SigV4.Service='es'` when serverless is false (client.go:49-53). Requires the Grafana instance to have `sigV4AuthEnabled=true` for the editor to expose the toggle.",
					Value: map[string]any{
						"url":    "https://vpc-example.us-east-1.es.amazonaws.com",
						"access": string(AccessModeProxy),
						"jsonData": map[string]any{
							"flavor":                     string(FlavorOpenSearch),
							"version":                    "2.11.0",
							"sigV4Auth":                  true,
							"sigV4AuthType":              "default",
							"sigV4Region":                "us-east-1",
							"database":                   "grafana-logs",
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequestsOpenSearch,
							"pplEnabled":                 true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"sigV4Serverless": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "SigV4 authentication (AWS OpenSearch Serverless)",
					Description: "AWS OpenSearch Serverless (`aoss`). Setting `jsonData.serverless=true` hard-codes flavor='opensearch', version='1.0.0', maxConcurrentShardRequests=5, pplEnabled=true, swaps the SigV4 service namespace from 'es' to 'aoss' (client.go:49-53), and adds the required `x-amz-content-sha256` header on every non-GET request (client.go:300-302).",
					Value: map[string]any{
						"url":    "https://<collection>.us-east-1.aoss.amazonaws.com",
						"access": string(AccessModeProxy),
						"jsonData": map[string]any{
							"serverless":                 true,
							"flavor":                     string(FlavorOpenSearch),
							"version":                    "1.0.0",
							"sigV4Auth":                  true,
							"sigV4AuthType":              "default",
							"sigV4Region":                "us-east-1",
							"database":                   "grafana-logs",
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequestsOpenSearch,
							"pplEnabled":                 true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"oauthForward": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Forward OAuth Identity",
					Description: "Forward the signed-in user's upstream OAuth identity to OpenSearch. The editor writes `jsonData.oauthPassThru=true` and the backend enables `ForwardHTTPHeaders` on the HTTP client (client.go:45-47). Only available when access='proxy'.",
					Value: map[string]any{
						"url":    "https://opensearch.example.com:9200",
						"access": string(AccessModeProxy),
						"jsonData": map[string]any{
							"oauthPassThru":              true,
							"flavor":                     string(FlavorOpenSearch),
							"version":                    "2.11.0",
							"database":                   "grafana-logs",
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequestsOpenSearch,
							"pplEnabled":                 true,
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
					Description: "OpenSearch cluster requiring mTLS. jsonData.tlsAuth=true triggers @grafana/ui's TLSAuthSettings to expose serverName + client cert + client key inputs; the SDK's HTTPClientOptions builds a client-authenticated transport.",
					Value: map[string]any{
						"url":    "https://opensearch.example.com:9200",
						"access": string(AccessModeProxy),
						"jsonData": map[string]any{
							"flavor":                     string(FlavorOpenSearch),
							"version":                    "2.11.0",
							"tlsAuth":                    true,
							"serverName":                 "opensearch.example.com",
							"database":                   "grafana-logs",
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequestsOpenSearch,
							"pplEnabled":                 true,
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
					Description: "OpenSearch behind a private CA. jsonData.tlsAuthWithCACert=true tells the SDK to verify the server certificate against the PEM-encoded CA in secureJsonData.tlsCACert.",
					Value: map[string]any{
						"url":    "https://opensearch.internal.corp:9200",
						"access": string(AccessModeProxy),
						"jsonData": map[string]any{
							"flavor":                     string(FlavorOpenSearch),
							"version":                    "2.11.0",
							"tlsAuthWithCACert":          true,
							"database":                   "grafana-logs",
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequestsOpenSearch,
							"pplEnabled":                 true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyTLSCACert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
						},
					},
				},
			},
			"elasticsearchLegacy": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Elasticsearch <7.0.0 (legacy flavor)",
					Description: "OpenSearch's data source also supports Elasticsearch clusters running the Open Distro plugins. Setting flavor='elasticsearch' switches the PPL endpoint from '_plugins/_ppl' to '_opendistro/_ppl' (client.go:557-560) and, for version <7.0.0, sets the default maxConcurrentShardRequests to 256 (client.go:411-416).",
					Value: map[string]any{
						"url":    "https://es.internal.corp:9200",
						"access": string(AccessModeProxy),
						"jsonData": map[string]any{
							"flavor":                     string(FlavorElasticsearch),
							"version":                    "6.8.0",
							"database":                   "logstash-*",
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequestsESLegacy,
							"pplEnabled":                 true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"logsWithDataLinks": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Logs configuration with data links",
					Description: "Log-oriented configuration with an explicit log message / log level field and two data link entries: one external URL and one internal Grafana link (datasourceUid) that treats jsonData.dataLinks[*].url as a query for the linked data source.",
					Value: map[string]any{
						"url":           "https://opensearch.example.com:9200",
						"access":        string(AccessModeProxy),
						"basicAuth":     true,
						"basicAuthUser": "grafana",
						"jsonData": map[string]any{
							"flavor":                     string(FlavorOpenSearch),
							"version":                    "2.11.0",
							"database":                   "[app-logs-]YYYY.MM.DD",
							"interval":                   string(IntervalDaily),
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequestsOpenSearch,
							"timeInterval":               "10s",
							"pplEnabled":                 true,
							"logMessageField":            "message",
							"logLevelField":              "level",
							"dataLinks": []any{
								map[string]any{
									"field":         "traceID",
									"title":         "View trace",
									"url":           "${__value.raw}",
									"datasourceUid": "tempo",
								},
								map[string]any{
									"field": "requestID",
									"title": "Open request",
									"url":   "https://requests.example.com/${__value.raw}",
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

// AccessModeProxy is a stringly-typed alias for the SDK root `access` field's
// default value ("proxy"). Exported so examples read cleanly without pulling
// magic strings into the SettingsExamples map.
const AccessModeProxy = "proxy"

// AccessModeDirect is the browser-mode alternative to AccessModeProxy.
const AccessModeDirect = "direct"
