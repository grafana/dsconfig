package servicenowdatasource

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
// truth for the ServiceNow datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the ServiceNow datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the ServiceNow
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the ServiceNow
// datasource, covering the default configuration and each authentication method
// (including the legacy oauthEnabled shape). Each example value is a full
// instance settings object with the root fields (url, basicAuthUser), the plugin
// configuration under jsonData, and the relevant write-only secrets under
// secureJsonData (obviously-fake placeholder values — replace them with real
// secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: HTTP Basic authentication with a 30s query timeout. Fill in root.url, root.basicAuthUser and secureJsonData.basicAuthPassword to get a working datasource.",
					Value: map[string]any{
						"url":           "https://acme.service-now.com",
						"basicAuthUser": "",
						"jsonData": map[string]any{
							"authMethod":          string(AuthMethodBasicAuth),
							"queryTimeoutSeconds": DefaultQueryTimeoutSeconds,
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
					Description: "Authenticate with a ServiceNow username and password. The instance URL and username live at the datasource root; only the password is secret (secureJsonData.basicAuthPassword).",
					Value: map[string]any{
						"url":           "https://acme.service-now.com",
						"basicAuthUser": "grafana_reader",
						"jsonData": map[string]any{
							"authMethod":          string(AuthMethodBasicAuth),
							"useSysTables":        true,
							"queryTimeoutSeconds": DefaultQueryTimeoutSeconds,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "<your-servicenow-password>",
						},
					},
				},
			},
			"serviceNowOAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "ServiceNow OAuth",
					Description: "ServiceNow OAuth is an OAuth2 resource-owner password grant, so it needs the account username/password AND the OAuth application's client id/secret. jsonData.oauthClientID is not secret; secureJsonData.oauthClientSecret is.",
					Value: map[string]any{
						"url":           "https://acme.service-now.com",
						"basicAuthUser": "grafana_reader",
						"jsonData": map[string]any{
							"authMethod":          string(AuthMethodServiceNowOAuth),
							"oauthClientID":       "<your-oauth-client-id>",
							"queryTimeoutSeconds": DefaultQueryTimeoutSeconds,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "<your-servicenow-password>",
							string(SecureJsonDataKeyOAuthClientSecret): "<your-oauth-client-secret>",
						},
					},
				},
			},
			"legacyOAuthEnabled": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy: oauthEnabled instead of authMethod",
					Description: "Datasources created before authMethod existed store jsonData.oauthEnabled=true (deprecated) with no authMethod; GetAuthMethod resolves them to serviceNowOAuth. The credential shape is otherwise identical to the ServiceNow OAuth example.",
					Value: map[string]any{
						"url":           "https://acme.service-now.com",
						"basicAuthUser": "grafana_reader",
						"jsonData": map[string]any{
							"oauthEnabled":  true,
							"oauthClientID": "<your-oauth-client-id>",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyBasicAuthPassword): "<your-servicenow-password>",
							string(SecureJsonDataKeyOAuthClientSecret): "<your-oauth-client-secret>",
						},
					},
				},
			},
		},
	}
}
