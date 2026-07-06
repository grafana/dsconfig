package jiradatasource

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
// truth for the Jira datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Jira datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Jira
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Jira
// datasource, covering the default configuration and each authentication method
// and connection variant the config editor supports. Each example value is a
// full instance settings object with the plugin configuration nested under
// jsonData and the relevant write-only secrets under secureJsonData
// (obviously-fake angle-bracket placeholder values — replace them with real
// secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: Basic Auth against Jira Cloud. jsonData.url and secureJsonData.token (empty here) must be filled in — pkg/models/settings.go:39-71 requires a URL and, for Basic Auth, a token.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authMethod": string(AuthMethodBasicAuth),
							"hosting":    string(HostingCloud),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "",
						},
					},
				},
			},
			"basicAuthCloud": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic Auth (Jira Cloud)",
					Description: "Atlassian Cloud with an account email + API token, sent as HTTP Basic auth. jsonData.url is the Atlassian site URL; secureJsonData.token is the API token created at https://id.atlassian.com/manage-profile/security/api-tokens.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authMethod": string(AuthMethodBasicAuth),
							"hosting":    string(HostingCloud),
							"url":        "https://mycompany.atlassian.net",
							"user":       "user@example.com",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "<your-jira-api-token>",
						},
					},
				},
			},
			"basicAuthServer": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic Auth (Jira Data Center / Jira Server)",
					Description: "Self-hosted Jira Data Center / Server (hosting='server' -> REST API v2, pkg/plugin.go:177-180) with an account email + token, sent as HTTP Basic auth.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authMethod": string(AuthMethodBasicAuth),
							"hosting":    string(HostingServer),
							"url":        "https://jira.example.com",
							"user":       "user@example.com",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "<your-jira-api-token>",
						},
					},
				},
			},
			"bearerTokenServer": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Bearer token / PAT (Jira Data Center, no user)",
					Description: "Jira Data Center personal access token: leave jsonData.user empty so the backend sends the token as a Bearer token instead of HTTP Basic auth (pkg/plugin.go:253-264).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authMethod": string(AuthMethodBasicAuth),
							"hosting":    string(HostingServer),
							"url":        "https://jira.example.com",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "<your-jira-personal-access-token>",
						},
					},
				},
			},
			"basicAuthScopedToken": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic Auth with scoped token (Jira Cloud)",
					Description: "A scoped Atlassian API token (scopedToken=true) routes requests through https://api.atlassian.com/ex/jira/<cloudId>, so jsonData.cloudId is required (pkg/plugin.go:221-226).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authMethod":  string(AuthMethodBasicAuth),
							"hosting":     string(HostingCloud),
							"url":         "https://mycompany.atlassian.net",
							"user":        "user@example.com",
							"scopedToken": true,
							"cloudId":     "<your-jira-cloud-id>",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "<your-scoped-api-token>",
						},
					},
				},
			},
			"oauth2": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "OAuth 2.0 — Service Account (Jira Cloud)",
					Description: "OAuth 2.0 client-credentials grant against https://auth.atlassian.com/oauth/token, targeting https://api.atlassian.com/ex/jira/<cloudId>. jsonData.oauthClientID, jsonData.cloudId and secureJsonData.oauthClientSecret are all required (pkg/models/settings.go:56-64). OAuth 2.0 is Jira Cloud only.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authMethod":    string(AuthMethodOAuth2),
							"hosting":       string(HostingCloud),
							"url":           "https://mycompany.atlassian.net",
							"oauthClientID": "<your-oauth-client-id>",
							"cloudId":       "<your-jira-cloud-id>",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyOAuthClientSecret): "<your-oauth-client-secret>",
						},
					},
				},
			},
			"legacyBasicAuthNoAuthMethod": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy: Basic Auth without an authMethod",
					Description: "Datasources created before jsonData.authMethod existed store only url + user + secureJsonData.token; the backend resolves the missing authMethod to basicAuth (pkg/models/auth_method.go:11-15). LoadConfig.ApplyDefaults fills it in on load.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"url":  "https://mycompany.atlassian.net",
							"user": "user@example.com",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyToken): "<your-jira-api-token>",
						},
					},
				},
			},
		},
	}
}
