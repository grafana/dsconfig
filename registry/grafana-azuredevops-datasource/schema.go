package azuredevopsdatasource

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
// truth for the Azure DevOps datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Azure DevOps datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Azure DevOps
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Azure DevOps
// datasource, covering the default configuration and each connection variant
// the config editor supports (Azure DevOps Services with a PAT, and Azure DevOps
// Server with a Basic-auth username). Each example value is a full instance
// settings object with the plugin configuration nested under jsonData and the
// write-only PAT under secureJsonData (placeholder value — replace it with a
// real token).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: patToken authentication with a projects limit of 100. jsonData.url and secureJsonData.patToken (empty here) must both be filled in to get a working datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AuthTypePAT),
							"projectsLimit": DefaultProjectsLimit,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPATToken): "",
						},
					},
				},
			},
			"patToken": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Personal Access Token (Azure DevOps Services)",
					Description: "Authenticate against Azure DevOps Services (dev.azure.com) with a personal access token in secureJsonData.patToken. The URL is the organization URL; the PAT is sent as HTTP Basic auth (empty username + PAT).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"url":           "https://dev.azure.com/your-organization",
							"authType":      string(AuthTypePAT),
							"projectsLimit": DefaultProjectsLimit,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPATToken): "<your-azure-devops-pat>",
						},
					},
				},
			},
			"azureDevOpsServer": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Personal Access Token with username (Azure DevOps Server)",
					Description: "On-prem Azure DevOps Server: setting jsonData.username makes the backend send an explicit HTTP Basic header CreateBasicAuthHeaderValue(username, patToken) with a normalized (lowercased, trailing-slash-trimmed) collection URL. Needed for some Azure DevOps Server versions.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"url":           "https://azuredevops.example.com/DefaultCollection",
							"authType":      string(AuthTypePAT),
							"username":      "ado",
							"projectsLimit": DefaultProjectsLimit,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPATToken): "<your-azure-devops-pat>",
						},
					},
				},
			},
		},
	}
}
