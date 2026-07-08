package elasticsearchdatasource

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
// truth for the Elasticsearch datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Elasticsearch datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Elasticsearch
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Elasticsearch
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
					Description: "The defaults a new datasource starts with: no authentication, timeField '@timestamp', maxConcurrentShardRequests 5, defaultQueryMode 'metrics'. Fill in root.url and a concrete jsonData.index to get a working datasource.",
					Value: map[string]any{
						"url": "http://localhost:9200",
						"jsonData": map[string]any{
							"index":                      "es-index-name",
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequests,
							"defaultQueryMode":           string(DefaultQueryMode),
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
					Description: "A network-isolated Elasticsearch with no HTTP-level auth. Index uses a daily logstash-style time pattern.",
					Value: map[string]any{
						"url": "http://elasticsearch.example.com:9200",
						"jsonData": map[string]any{
							"index":                      "[logstash-]YYYY.MM.DD",
							"interval":                   string(IntervalDaily),
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequests,
							"defaultQueryMode":           string(DefaultQueryMode),
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
					Description: "Authenticate with an Elasticsearch cluster that requires HTTP Basic. Both basicAuth and basicAuthUser live at the datasource root; only basicAuthPassword is secret.",
					Value: map[string]any{
						"url":           "https://elasticsearch.example.com:9200",
						"basicAuth":     true,
						"basicAuthUser": "grafana",
						"jsonData": map[string]any{
							"index":                      "grafana-logs",
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequests,
							"defaultQueryMode":           string(DefaultQueryMode),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "REPLACE_WITH_PASSWORD",
						},
					},
				},
			},
			"apiKey": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "API Key authentication",
					Description: "Elasticsearch API key auth: the backend sends `Authorization: ApiKey <value>` on every request (elasticsearch.go:125-131). Set jsonData.apiKeyAuth=true and provide the base64-encoded key in secureJsonData.apiKey.",
					Value: map[string]any{
						"url": "https://elasticsearch.example.com:9200",
						"jsonData": map[string]any{
							"apiKeyAuth":                 true,
							"index":                      "grafana-logs",
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequests,
							"defaultQueryMode":           string(QueryTypeLogs),
							"logMessageField":            "message",
							"logLevelField":              "level",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAPIKey): "REPLACE_WITH_BASE64_API_KEY",
						},
					},
				},
			},
			"sigV4": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "SigV4 authentication (AWS OpenSearch / managed Elasticsearch)",
					Description: "AWS SigV4-signed requests to an AWS-managed Elasticsearch or OpenSearch cluster. The sigV4* jsonData fields are contributed by @grafana/aws-sdk's SIGV4ConnectionConfig, and the backend forces the SigV4 service namespace to 'es' whenever the SDK builds a SigV4 transport (elasticsearch.go:120-123). Requires the Grafana instance to have sigV4AuthEnabled=true for the editor to expose the option.",
					Value: map[string]any{
						"url": "https://vpc-example.us-east-1.es.amazonaws.com",
						"jsonData": map[string]any{
							"sigV4Auth":                  true,
							"sigV4AuthType":              "default",
							"sigV4Region":                "us-east-1",
							"index":                      "grafana-logs",
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequests,
							"defaultQueryMode":           string(DefaultQueryMode),
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
					Description: "Forward the signed-in user's upstream OAuth identity to Elasticsearch. The editor writes only jsonData.oauthPassThru; there is no accompanying secret.",
					Value: map[string]any{
						"url": "https://elasticsearch.example.com:9200",
						"jsonData": map[string]any{
							"oauthPassThru":              true,
							"index":                      "grafana-logs",
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequests,
							"defaultQueryMode":           string(DefaultQueryMode),
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
					Description: "Elasticsearch requiring mTLS. jsonData.tlsAuth=true triggers the SDK to build a client-authenticated transport using serverName plus the PEM-encoded client cert and key in secureJsonData.",
					Value: map[string]any{
						"url": "https://elasticsearch.example.com:9200",
						"jsonData": map[string]any{
							"tlsAuth":                    true,
							"serverName":                 "elasticsearch.example.com",
							"index":                      "grafana-logs",
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequests,
							"defaultQueryMode":           string(DefaultQueryMode),
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
					Description: "Elasticsearch behind a private CA. jsonData.tlsAuthWithCACert=true tells the SDK to verify the server certificate against the PEM-encoded CA in secureJsonData.tlsCACert.",
					Value: map[string]any{
						"url": "https://elasticsearch.internal.corp:9200",
						"jsonData": map[string]any{
							"tlsAuthWithCACert":          true,
							"index":                      "grafana-logs",
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequests,
							"defaultQueryMode":           string(DefaultQueryMode),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyTLSCACert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
						},
					},
				},
			},
			"logsWithDataLinks": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Logs mode with data links",
					Description: "Log-oriented configuration with an explicit log message / log level field and two data link entries: one external URL and one internal Grafana link (datasourceUid) that treats jsonData.dataLinks[*].url as a query for the linked data source.",
					Value: map[string]any{
						"url":           "https://elasticsearch.example.com:9200",
						"basicAuth":     true,
						"basicAuthUser": "grafana",
						"jsonData": map[string]any{
							"index":                      "[app-logs-]YYYY.MM.DD",
							"interval":                   string(IntervalDaily),
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequests,
							"defaultQueryMode":           string(QueryTypeLogs),
							"logMessageField":            "message",
							"logLevelField":              "level",
							"includeFrozen":              false,
							"timeInterval":               "10s",
							"dataLinks": []any{
								map[string]any{
									"field":           "traceID",
									"url":             "${__value.raw}",
									"urlDisplayLabel": "View trace",
									"datasourceUid":   "tempo",
								},
								map[string]any{
									"field":           "requestID",
									"url":             "https://requests.example.com/${__value.raw}",
									"urlDisplayLabel": "Open request",
								},
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "REPLACE_WITH_PASSWORD",
						},
					},
				},
			},
			"legacyDatabaseFallback": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy: index stored in root.database",
					Description: "Datasources created before the index moved to jsonData.index stored it in the root-level `database` field instead. The backend's LoadConfig / NewDatasource promotes settings.Database to jsonData.index when the latter is empty (elasticsearch.go:164-170); the editor's indexChangeHandler always clears database on save (ElasticDetails.tsx:162), so this shape only appears on datasources that have never been re-saved through the editor.",
					Value: map[string]any{
						"url":      "http://elasticsearch.example.com:9200",
						"database": "grafana-legacy",
						"jsonData": map[string]any{
							"timeField":                  defaultTimeField,
							"maxConcurrentShardRequests": defaultMaxConcurrentShardRequests,
							"defaultQueryMode":           string(DefaultQueryMode),
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
