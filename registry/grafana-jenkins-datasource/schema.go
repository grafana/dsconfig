package jenkinsdatasource

import (
	_ "embed"

	"k8s.io/kube-openapi/pkg/spec3"

	"github.com/grafana/dsconfig/dsconfig"
	sdkschema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

//go:generate go test -generateArtifacts -run TestSchemaConformance ./...

// TargetAPIVersion is the API version this schema applies to.
const TargetAPIVersion = dsconfig.TargetAPIVersion

// configSchemaJSON is the declarative dsconfig schema — the single
// source of truth for the Jenkins datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig
// schema (single source of truth) for the Jenkins datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the
// Jenkins datasource: the settings (configuration) spec derived from
// dsconfig.json, the secure values, and example configurations,
// stamped with TargetAPIVersion. Grafana's datasource API server
// serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the
// Jenkins datasource, covering the default configuration and each
// authentication variant the config editor supports (anonymous and
// HTTP Basic auth). Each example value is a full instance settings
// object with the plugin configuration nested under jsonData and the
// relevant write-only secrets under secureJsonData (placeholder
// values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: an empty URL and an empty password placeholder. jsonData.url must be filled in for the datasource to load — pkg/plugin/settings.go:23-25 rejects an empty URL with DownstreamError(\"URL is missing\"). Username and password are optional.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"url": "",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "",
						},
					},
				},
			},
			"anonymous": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Anonymous access",
					Description: "Configuration for a Jenkins instance that permits anonymous read access. jsonData.username is omitted, so pkg/plugin/datasource.go:66-71 skips the BasicAuth wiring and the client issues unauthenticated requests. secureJsonData.password is present but empty because the field is declared in the plugin's secure JSON schema.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"url": "https://jenkins.example.com",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "",
						},
					},
				},
			},
			"basicAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "HTTP Basic auth",
					Description: "Standard configuration: jsonData.username plus secureJsonData.password. Because jsonData.username is non-empty, pkg/plugin/datasource.go:66-71 configures httpclient.BasicAuthOptions and every outgoing request sets the Authorization: Basic header (pkg/jenkins/client.go:498-500).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"url":      "https://jenkins.example.com",
							"username": "grafana",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "hunter2",
						},
					},
				},
			},
			"apiToken": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "HTTP Basic auth with a Jenkins API token",
					Description: "Same shape as HTTP Basic auth, but the value in secureJsonData.password is a Jenkins API token (recommended over a user password). Jenkins accepts an API token in the Basic-auth password position, so no discriminator is needed — the plugin treats password and API token identically.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"url":      "https://jenkins.example.com",
							"username": "grafana",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "11a1b2c3d4e5f60718293a4b5c6d7e8f90",
						},
					},
				},
			},
			"legacyUsernameOnly": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Legacy: username without a password",
					Description: "A legacy configuration that has jsonData.username set but never had secureJsonData.password provided. Because a username is set, pkg/plugin/datasource.go:66-71 still wires BasicAuth — with an empty password — and Jenkins decides whether the resulting empty-password request is authorised.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"url":      "https://jenkins.example.com",
							"username": "grafana",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "",
						},
					},
				},
			},
		},
	}
}
