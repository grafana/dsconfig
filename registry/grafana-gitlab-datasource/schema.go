package gitlabdatasource

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
// truth for the GitLab datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the GitLab datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the GitLab
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the GitLab
// datasource, covering the default configuration and each connection variant
// the config editor supports (GitLab SaaS vs a self-hosted instance). The
// GitLab datasource has a single authentication method (a personal access
// token), so the examples differ only by connection.
//
// Each example value is a full instance settings object with the root url, the
// plugin configuration under jsonData, and the write-only access token under
// secureJsonData (obviously-fake placeholder values — replace them with a real
// token).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: the GitLab SaaS API base URL and a page limit of 5. Only secureJsonData.accessToken (empty here) needs to be filled in to get a working datasource — the backend rejects an empty token (pkg/models/settings.go:48-50).",
					Value: map[string]any{
						"url": DefaultURL,
						"jsonData": map[string]any{
							"pageLimit": DefaultPageLimit,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessToken): "",
						},
					},
				},
			},
			"gitlabSaaS": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "GitLab SaaS (gitlab.com)",
					Description: "Authenticate against the hosted GitLab service. The root url is the default API base URL https://gitlab.com/api/v4; the personal access token is provided in secureJsonData.accessToken and sent as the PRIVATE-TOKEN header.",
					Value: map[string]any{
						"url": DefaultURL,
						"jsonData": map[string]any{
							"pageLimit": DefaultPageLimit,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessToken): "<gitlab-personal-access-token>",
						},
					},
				},
			},
			"selfHosted": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Self-hosted GitLab",
					Description: "Self-managed GitLab instance. The root url points at the instance base URL (must include an http/https scheme); go-gitlab appends api/v4/ automatically (go-gitlab gitlab.go:564-578), so no /api/v4 suffix is required here.",
					Value: map[string]any{
						"url": "https://gitlab.example.com",
						"jsonData": map[string]any{
							"pageLimit": DefaultPageLimit,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessToken): "<gitlab-personal-access-token>",
						},
					},
				},
			},
			"selfHostedApiV4CustomPageLimit": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Self-hosted GitLab with explicit /api/v4 and custom page limit",
					Description: "Self-managed GitLab with the API base URL given explicitly (already ending in /api/v4, so go-gitlab leaves it as-is) and a raised jsonData.pageLimit for queries that page through more results.",
					Value: map[string]any{
						"url": "https://gitlab.example.com/api/v4",
						"jsonData": map[string]any{
							"pageLimit": 10,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccessToken): "<gitlab-personal-access-token>",
						},
					},
				},
			},
		},
	}
}
