package githubdatasource

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
// truth for the GitHub datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the GitHub datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the GitHub
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the GitHub
// datasource, covering the default configuration and each authentication
// type and connection (license) variant the config editor supports. Each
// example value is a full instance settings object with the plugin
// configuration nested under jsonData and the relevant write-only secrets
// under secureJsonData (placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: personal access token authentication against GitHub.com (Free, Pro & Team). Only secureJsonData.accessToken (empty here) needs to be filled in to get a working datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"selectedAuthType": string(AuthTypePAT),
							"githubPlan":       string(LicenseTypeBasic),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessToken): "",
						},
					},
				},
			},
			"personalAccessToken": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Personal Access Token (GitHub.com)",
					Description: "Authenticate against GitHub.com (Free, Pro & Team) with a fine grained personal access token provided in secureJsonData.accessToken.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"selectedAuthType": string(AuthTypePAT),
							"githubPlan":       string(LicenseTypeBasic),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessToken): "github_pat_XXXXXXXXXXXXXXXXXXXXXX",
						},
					},
				},
			},
			"githubApp": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "GitHub App (GitHub.com)",
					Description: "Authenticate against GitHub.com as a GitHub App installation. appId and installationId may be JSON strings or numbers; secureJsonData.privateKey is the app's complete private key PEM including the BEGIN/END lines.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"selectedAuthType": string(AuthTypeGithubApp),
							"githubPlan":       string(LicenseTypeBasic),
							"appId":            "123456",
							"installationId":   "12345678",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPrivateKey): "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
						},
					},
				},
			},
			"enterpriseCloud": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Personal Access Token (GitHub Enterprise Cloud)",
					Description: "GitHub Enterprise Cloud uses the same API endpoints as GitHub.com, so no URL is configured; githubPlan only drives the config editor. The token is provided in secureJsonData.accessToken.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"selectedAuthType": string(AuthTypePAT),
							"githubPlan":       string(LicenseTypeEnterpriseCloud),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessToken): "github_pat_XXXXXXXXXXXXXXXXXXXXXX",
						},
					},
				},
			},
			"enterpriseServer": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Personal Access Token (GitHub Enterprise Server)",
					Description: "On-prem GitHub Enterprise Server: the backend derives <githubUrl>/api/v3 (REST) and <githubUrl>/api/graphql (GraphQL). The token is provided in secureJsonData.accessToken.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"selectedAuthType": string(AuthTypePAT),
							"githubPlan":       string(LicenseTypeEnterpriseServer),
							"githubUrl":        "https://github.example.com",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessToken): "github_pat_XXXXXXXXXXXXXXXXXXXXXX",
						},
					},
				},
			},
			"githubAppEnterpriseServer": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "GitHub App (GitHub Enterprise Server)",
					Description: "GitHub App installation on an on-prem GitHub Enterprise Server, with the app's private key PEM in secureJsonData.privateKey.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"selectedAuthType": string(AuthTypeGithubApp),
							"githubPlan":       string(LicenseTypeEnterpriseServer),
							"githubUrl":        "https://github.example.com",
							"appId":            "123456",
							"installationId":   "12345678",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPrivateKey): "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----",
						},
					},
				},
			},
			"legacyAccessTokenOnly": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy: access token without an auth type",
					Description: "Datasources created before selectedAuthType existed store only secureJsonData.accessToken; the backend defaults them to personal-access-token.",
					Value: map[string]any{
						"jsonData": map[string]any{},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessToken): "github_pat_XXXXXXXXXXXXXXXXXXXXXX",
						},
					},
				},
			},
		},
	}
}
