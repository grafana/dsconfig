package graphitedatasource

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
// truth for the Graphite datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Graphite datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Graphite
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Graphite
// datasource, covering the default configuration and each authentication /
// TLS / backend-type variant the config editor supports. Each example value
// is a full instance settings object with the plugin configuration nested
// under jsonData and the relevant write-only secrets under secureJsonData
// (placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: no authentication, no TLS overrides, Graphite version '1.1' (written by the editor on load). Only root.url needs to be filled in to get a working datasource pointed at a Graphite server on localhost:8080.",
					Value: map[string]any{
						"url": "http://localhost:8080",
						"jsonData": map[string]any{
							"graphiteVersion": string(DefaultGraphiteVersion),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"noAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "No authentication (Graphite 1.1)",
					Description: "A public or network-isolated Graphite with no HTTP-level auth. Graphite version is set explicitly so the query editor exposes the 1.1 function library (tags + seriesByTag).",
					Value: map[string]any{
						"url": "http://graphite.example.com:8080",
						"jsonData": map[string]any{
							"graphiteVersion": string(GraphiteVersion11),
							"graphiteType":    string(GraphiteTypeDefault),
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
					Description: "Authenticate with a Graphite that requires HTTP Basic. Both basicAuth and basicAuthUser live at the datasource root; only basicAuthPassword is secret.",
					Value: map[string]any{
						"url":           "https://graphite.example.com",
						"basicAuth":     true,
						"basicAuthUser": "grafana",
						"jsonData": map[string]any{
							"graphiteVersion": string(GraphiteVersion11),
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
					Description: "Forward the signed-in user's upstream OAuth identity to Graphite. The editor writes only jsonData.oauthPassThru; there is no accompanying secret.",
					Value: map[string]any{
						"url": "https://graphite.example.com",
						"jsonData": map[string]any{
							"graphiteVersion": string(GraphiteVersion11),
							"oauthPassThru":   true,
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
					Description: "Graphite requiring mTLS. jsonData.tlsAuth=true triggers the SDK to build a client-authenticated transport using serverName plus the PEM-encoded client cert and key in secureJsonData.",
					Value: map[string]any{
						"url": "https://graphite.example.com",
						"jsonData": map[string]any{
							"graphiteVersion": string(GraphiteVersion11),
							"tlsAuth":         true,
							"serverName":      "graphite.example.com",
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
					Description: "Graphite behind a private CA. jsonData.tlsAuthWithCACert=true tells the SDK to verify the server certificate against the PEM-encoded CA in secureJsonData.tlsCACert.",
					Value: map[string]any{
						"url": "https://graphite.internal.corp",
						"jsonData": map[string]any{
							"graphiteVersion":   string(GraphiteVersion11),
							"tlsAuthWithCACert": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyTLSCACert): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
						},
					},
				},
			},
			"metrictank": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Metrictank backend with rollup indicator",
					Description: "A Metrictank server exposed via the Graphite render API. graphiteType='metrictank' unlocks Metrictank-specific features in the query editor; rollupIndicatorEnabled=true adds a badge to panels when Metrictank aggregates data.",
					Value: map[string]any{
						"url": "https://metrictank.example.com",
						"jsonData": map[string]any{
							"graphiteVersion":        string(GraphiteVersion11),
							"graphiteType":           string(GraphiteTypeMetrictank),
							"rollupIndicatorEnabled": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"withLabelMappings": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Graphite→Loki label mappings",
					Description: "Configures importConfiguration.loki.mappings so Explore's datasource-switch flow can convert 'servers.(cluster).(server).*' Graphite paths into Loki label selectors. The mapping is frontend-only: neither Graphite nor Loki reads it at query time.",
					Value: map[string]any{
						"url": "http://localhost:8080",
						"jsonData": map[string]any{
							"graphiteVersion": string(GraphiteVersion11),
							"importConfiguration": map[string]any{
								"loki": map[string]any{
									"mappings": []any{
										map[string]any{
											"matchers": []any{
												map[string]any{"value": "servers"},
												map[string]any{"value": "*", "labelName": "cluster"},
												map[string]any{"value": "*", "labelName": "server"},
												map[string]any{"value": "*"},
											},
										},
									},
								},
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "",
						},
					},
				},
			},
			"legacyDirectAccess": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy browser access mode",
					Description: "A pre-existing datasource stored with root.access='direct' (browser access mode). Graphite's editor renders a deprecation Alert (ConfigEditor.tsx:54-58) but preserves the mode on save. New datasources should not use this — the deprecation notice says it will be removed.",
					Value: map[string]any{
						"url":    "http://graphite.example.com:8080",
						"access": "direct",
						"jsonData": map[string]any{
							"graphiteVersion": string(GraphiteVersion10),
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
