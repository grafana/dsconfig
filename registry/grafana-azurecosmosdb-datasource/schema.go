package azurecosmosdbdatasource

import (
	_ "embed"

	"k8s.io/kube-openapi/pkg/spec3"

	"github.com/grafana/dsconfig/dsconfig"
	sdkschema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

//go:generate go test -generateArtifacts -run TestSchemaConformance ./...

// TargetAPIVersion is the API version this schema applies to.
const TargetAPIVersion = dsconfig.TargetAPIVersion

// configSchemaJSON is the declarative dsconfig schema — the single source
// of truth for the Azure Cosmos DB datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig
// schema (single source of truth) for the Azure Cosmos DB datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Azure
// Cosmos DB datasource: the settings (configuration) spec derived from
// dsconfig.json, the secure values, and example configurations, stamped
// with TargetAPIVersion. Grafana's datasource API server serves this
// bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the
// Azure Cosmos DB datasource. The plugin supports exactly one auth
// method (Cosmos DB account master key), so the examples cover only the
// default (empty-secret) shape plus a fully populated realistic
// configuration. Each example value is a full instance-settings object
// with the plugin configuration nested under jsonData and the write-only
// account key under secureJsonData (placeholder values — replace them
// with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: empty accountEndpoint and an empty accountKey placeholder. Both jsonData.accountEndpoint and secureJsonData.accountKey must be filled in for the datasource to load — pkg/plugin/settings.go:15-23 rejects either being empty.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"accountEndpoint": "",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccountKey): "",
						},
					},
				},
			},
			"accountKey": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Account key authentication",
					Description: "Standard configuration against an Azure Cosmos DB account. jsonData.accountEndpoint is the account URI (as shown on the account's Keys blade in the Azure portal, typically https://<account>.documents.azure.com:443/) and secureJsonData.accountKey is the primary or secondary master key. The backend wraps the key with azcosmos.NewKeyCredential and calls azcosmos.NewClientWithKey at pkg/cosmos/client.go:24-52.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"accountEndpoint": "https://my-account.documents.azure.com:443/",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAccountKey): "XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX==",
						},
					},
				},
			},
		},
	}
}
