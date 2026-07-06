package splunkdatasource

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
// truth for the Splunk datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Splunk datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Splunk
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// PEM placeholders used in the TLS example below. They are deliberately obvious,
// non-secret placeholders (never real key material).
const (
	examplePEMCertificate = "-----BEGIN CERTIFICATE-----\n<redacted>\n-----END CERTIFICATE-----"
	examplePEMPrivateKey  = "-----BEGIN RSA PRIVATE KEY-----\n<redacted>\n-----END RSA PRIVATE KEY-----"
)

// SettingsExamples returns k8s-style example configurations for the Splunk
// datasource, covering the default configuration and each authentication method
// plus TLS and advanced-option variants. Each example value is a full instance
// settings object with the plugin configuration nested under jsonData and the
// relevant write-only secrets under secureJsonData (placeholder values — replace
// them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: Basic authentication against the Splunk management/REST API. root.url, root.basicAuthUser and secureJsonData.basicAuthPassword (empty here) must be filled in to get a working datasource.",
					Value: map[string]any{
						"url": "https://splunk.example.com:8089",
						"jsonData": map[string]any{
							"authType": string(AuthTypeBasicAuth),
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
					Description: "Authenticate with a Splunk username and password. The username is stored at root.basicAuthUser and the password in secureJsonData.basicAuthPassword; the backend derives BasicAuthEnabled from authType (pkg/models/settings.go:95), so root.basicAuth is not set.",
					Value: map[string]any{
						"url":           "https://splunk.example.com:8089",
						"basicAuthUser": "splunk_admin",
						"jsonData": map[string]any{
							"authType": string(AuthTypeBasicAuth),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "<your-splunk-password>",
						},
					},
				},
			},
			"alternativeToken": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Alternative authentication (Splunk auth token)",
					Description: "Authenticate with a Splunk authentication token instead of a username/password. The token is stored in secureJsonData.authToken and sent as 'Authorization: Bearer <token>' (pkg/splunk/client.go:229-230).",
					Value: map[string]any{
						"url": "https://splunk.example.com:8089",
						"jsonData": map[string]any{
							"authType": string(AuthTypeAlternativeToken),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAuthToken): "<splunk-authentication-token>",
						},
					},
				},
			},
			"oauthForward": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Forward OAuth Identity",
					Description: "Forward the signed-in Grafana user's OAuth identity to Splunk. Sets jsonData.authType='OAuthForward' and jsonData.oauthPassThru=true; there is no secret to configure. Only offered in the editor when the splunkEnableOAuthForwarding feature toggle is enabled.",
					Value: map[string]any{
						"url": "https://splunk.example.com:8089",
						"jsonData": map[string]any{
							"authType":      string(AuthTypeOAuthForward),
							"oauthPassThru": true,
						},
					},
				},
			},
			"basicAuthWithTLS": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic authentication with TLS client auth and custom CA",
					Description: "Basic auth combined with mutual TLS (jsonData.tlsAuth + serverName + client cert/key) and custom-CA verification (jsonData.tlsAuthWithCACert + CA cert). TLS values are consumed by the SDK via config.HTTPClientOptions (pkg/models/settings.go:125).",
					Value: map[string]any{
						"url":           "https://splunk.example.com:8089",
						"basicAuthUser": "splunk_admin",
						"jsonData": map[string]any{
							"authType":          string(AuthTypeBasicAuth),
							"tlsAuth":           true,
							"serverName":        "splunk.example.com",
							"tlsAuthWithCACert": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "<your-splunk-password>",
							string(SecureJsonDataKeyTLSClientCert):     examplePEMCertificate,
							string(SecureJsonDataKeyTLSClientKey):      examplePEMPrivateKey,
							string(SecureJsonDataKeyTLSCACert):         examplePEMCertificate,
						},
					},
				},
			},
			"tokenWithAdvancedOptions": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Alternative token auth with advanced options and data links",
					Description: "A fully-populated instance: token auth, preview + async queries with a poll interval, a plugin request timeout, internal-field filtering, a custom result limit, forwarded cookies, and two data links (one external URL template, one internal link to a Tempo datasource).",
					Value: map[string]any{
						"url": "https://splunk.example.com:8089",
						"jsonData": map[string]any{
							"authType":                 string(AuthTypeAlternativeToken),
							"previewMode":              true,
							"pollSearchResult":         true,
							"minPollInterval":          "500",
							"maxPollInterval":          "3000",
							"autoCancel":               "30",
							"timeoutInSeconds":         60,
							"statusBuckets":            "300",
							"internalFieldsFiltration": true,
							"internalFieldPattern":     "^_.+",
							"tsField":                  "_time",
							"fieldSearchType":          string(FieldSearchTypeFull),
							"variableSearchLevel":      string(VariableSearchLevelSmart),
							"defaultEarliestTime":      "-1hr",
							"maxResultCount":           5000,
							"keepCookies":              []any{"session_id"},
							"timeout":                  30,
							"dataLinks": []any{
								map[string]any{
									"field":        "trace_id",
									"label":        "View trace",
									"matcherRegex": "/trace_id=(\\w+)/",
									"url":          "https://traces.example.com/${__value.raw}",
								},
								map[string]any{
									"field":         "trace_id",
									"label":         "Trace",
									"matcherRegex":  "/trace_id=(\\w+)/",
									"url":           "${__value.raw}",
									"datasourceUid": "tempo",
								},
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAuthToken): "<splunk-authentication-token>",
						},
					},
				},
			},
		},
	}
}
