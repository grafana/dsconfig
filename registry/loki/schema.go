package lokidatasource

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
// truth for the Loki datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Loki datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Loki
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Loki
// datasource, covering the default configuration and each authentication
// method / TLS variant / Loki-specific feature (Alerting, Maximum lines,
// Derived fields) the config editor supports. Each example value is a full
// instance settings object with the plugin configuration nested under
// jsonData and the relevant write-only secrets under secureJsonData
// (placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: no authentication, no TLS overrides, no derived fields. Only root.url needs to be filled in to get a working datasource pointed at a Loki server on localhost:3100.",
					Value: map[string]any{
						"url":      "http://localhost:3100",
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
					Description: "A public or network-isolated Loki with no HTTP-level auth. Maximum lines is bumped to 5000 for ad-hoc analysis and alerting UI management is enabled.",
					Value: map[string]any{
						"url": "http://loki.example.com:3100",
						"jsonData": map[string]any{
							"maxLines":     "5000",
							"manageAlerts": true,
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
					Description: "Authenticate with a Loki that requires HTTP Basic. Both basicAuth and basicAuthUser live at the datasource root; only basicAuthPassword is secret.",
					Value: map[string]any{
						"url":           "https://loki.example.com",
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
					Description: "Forward the signed-in user's upstream OAuth identity to Loki. The editor writes only jsonData.oauthPassThru; there is no accompanying secret.",
					Value: map[string]any{
						"url": "https://loki.example.com",
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
					Description: "Loki requiring mTLS. jsonData.tlsAuth=true triggers the SDK to build a client-authenticated transport using serverName plus the PEM-encoded client cert and key in secureJsonData.",
					Value: map[string]any{
						"url": "https://loki.example.com",
						"jsonData": map[string]any{
							"tlsAuth":    true,
							"serverName": "loki.example.com",
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
					Description: "Loki behind a private CA. jsonData.tlsAuthWithCACert=true tells the SDK to verify the server certificate against the PEM-encoded CA in secureJsonData.tlsCACert.",
					Value: map[string]any{
						"url": "https://loki.internal.corp",
						"jsonData": map[string]any{
							"tlsAuthWithCACert": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyTLSCACert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
						},
					},
				},
			},
			"withDerivedFields": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic auth with derived fields",
					Description: "Loki configured with a derived field that extracts a trace ID from the log line via regex and internally links to a Tempo data source. Two derived fields — one regex-based external link, one label-based internal link — demonstrate the shape.",
					Value: map[string]any{
						"url":           "https://loki.example.com",
						"basicAuth":     true,
						"basicAuthUser": "grafana",
						"jsonData": map[string]any{
							"maxLines": "1500",
							"derivedFields": []any{
								map[string]any{
									"name":          "TraceID",
									"matcherType":   string(DerivedFieldMatcherRegex),
									"matcherRegex":  "traceID=(\\w+)",
									"url":           "${__value.raw}",
									"datasourceUid": "tempo",
									"targetBlank":   false,
								},
								map[string]any{
									"name":            "requestId",
									"matcherType":     string(DerivedFieldMatcherRegex),
									"matcherRegex":    "req_id=(\\w+)",
									"url":             "https://requests.example.com/${__value.raw}",
									"urlDisplayLabel": "View request",
									"targetBlank":     true,
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
