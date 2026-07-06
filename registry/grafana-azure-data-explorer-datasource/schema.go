package azuredataexplorerdatasource

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
// of truth for the Azure Data Explorer datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig
// schema for the Azure Data Explorer datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the k8s-style SDK plugin schema bundle Grafana's
// datasource API server serves as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations covering the
// default (schema-defaults) configuration and one example per
// authentication type ADX supports, plus a legacy provisioning example.
// Every value is a full instance-settings object with the plugin
// configuration nested under `jsonData` and the relevant write-only
// secrets under `secureJsonData` (placeholder values — replace them with
// real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Default configuration",
					Description: "The defaults a new datasource starts with. ADX has no schema-side " +
						"credential default (the config editor chooses one at load time based on " +
						"Grafana's managedIdentityEnabled / workloadIdentityEnabled / " +
						"userIdentityEnabled config — see `AzureCredentialsConfig.ts:22-34`), so a " +
						"fresh datasource has no `azureCredentials` and empty secrets. `queryTimeout`, " +
						"`dataConsistency`, and `defaultEditorMode` carry their ApplyDefaults values.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"queryTimeout":      "30s",
							"dataConsistency":   string(DataConsistencyStrong),
							"defaultEditorMode": string(EditorModeVisual),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAzureClientSecret): "",
						},
					},
				},
			},
			"clientSecret": {
				ExampleProps: spec3.ExampleProps{
					Summary: "App Registration (Client Secret)",
					Description: "Authenticate via an Entra ID (Azure AD) app registration and client " +
						"secret. The secret is stored write-only in " +
						"`secureJsonData.azureClientSecret`; the rest of the credential lives in " +
						"`jsonData.azureCredentials`.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"azureCredentials": map[string]any{
								"authType":   string(AuthTypeClientSecret),
								"azureCloud": "AzureCloud",
								"tenantId":   "00000000-0000-0000-0000-000000000000",
								"clientId":   "00000000-0000-0000-0000-000000000000",
							},
							"clusterUrl":        "https://yourcluster.kusto.windows.net",
							"defaultDatabase":   "mydb",
							"queryTimeout":      "30s",
							"dataConsistency":   string(DataConsistencyStrong),
							"defaultEditorMode": string(EditorModeVisual),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAzureClientSecret): "changeme",
						},
					},
				},
			},
			"clientSecretOBO": {
				ExampleProps: spec3.ExampleProps{
					Summary: "App Registration (On-Behalf-Of)",
					Description: "OBO auth requires the `adxOnBehalfOf` feature toggle in the editor " +
						"AND `jsonData.oauthPassThru = true` — the plugin backend rejects OBO " +
						"instance creation without it (`adxcredentials/builder.go:89-98`). " +
						"`@grafana/azure-sdk` sets `oauthPassThru: true` automatically when the OBO " +
						"authType is selected.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"azureCredentials": map[string]any{
								"authType":   string(AuthTypeClientSecretOBO),
								"azureCloud": "AzureCloud",
								"tenantId":   "00000000-0000-0000-0000-000000000000",
								"clientId":   "00000000-0000-0000-0000-000000000000",
							},
							"oauthPassThru":     true,
							"clusterUrl":        "https://yourcluster.kusto.windows.net",
							"defaultDatabase":   "mydb",
							"queryTimeout":      "30s",
							"dataConsistency":   string(DataConsistencyStrong),
							"defaultEditorMode": string(EditorModeVisual),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAzureClientSecret): "changeme",
						},
					},
				},
			},
			"managedIdentity": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Managed Identity",
					Description: "Authenticate using the Grafana host's Azure Managed Identity. Only " +
						"available when Grafana's `azure.managedIdentityEnabled` is true. No secret " +
						"required.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"azureCredentials": map[string]any{
								"authType": string(AuthTypeManagedIdentity),
							},
							"clusterUrl":        "https://yourcluster.kusto.windows.net",
							"defaultDatabase":   "mydb",
							"queryTimeout":      "30s",
							"dataConsistency":   string(DataConsistencyStrong),
							"defaultEditorMode": string(EditorModeVisual),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAzureClientSecret): "",
						},
					},
				},
			},
			"workloadIdentity": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Workload Identity",
					Description: "Authenticate using an Entra ID Workload Identity federated to the " +
						"Kubernetes service account Grafana runs under. Only available when " +
						"Grafana's `azure.workloadIdentityEnabled` is true. No secret required.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"azureCredentials": map[string]any{
								"authType": string(AuthTypeWorkloadIdentity),
							},
							"clusterUrl":        "https://yourcluster.kusto.windows.net",
							"defaultDatabase":   "mydb",
							"queryTimeout":      "30s",
							"dataConsistency":   string(DataConsistencyStrong),
							"defaultEditorMode": string(EditorModeVisual),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAzureClientSecret): "",
						},
					},
				},
			},
			"currentUser": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Current User",
					Description: "Forward the signed-in Grafana user's identity to Azure Data " +
						"Explorer. Only available when Grafana's `azure.userIdentityEnabled` is " +
						"true. Alerting, recorded queries, and reporting run without a signed-in " +
						"user and may not work — see `doc/current-user-auth.md`. No secret required.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"azureCredentials": map[string]any{
								"authType": string(AuthTypeCurrentUser),
							},
							"clusterUrl":        "https://yourcluster.kusto.windows.net",
							"defaultDatabase":   "mydb",
							"queryTimeout":      "30s",
							"dataConsistency":   string(DataConsistencyStrong),
							"defaultEditorMode": string(EditorModeVisual),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAzureClientSecret): "",
						},
					},
				},
			},
			"openAI": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Client Secret + OpenAI API key (provisioning-only)",
					Description: "Populate `secureJsonData.OpenAIAPIKey` alongside a working auth " +
						"payload to enable the AI query assistant (`askOpenAI` resource endpoint, " +
						"`pkg/azuredx/resource_handler.go`). There is no config editor UI for this " +
						"key — it must be set via provisioning YAML or the datasource HTTP API.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"azureCredentials": map[string]any{
								"authType":   string(AuthTypeClientSecret),
								"azureCloud": "AzureCloud",
								"tenantId":   "00000000-0000-0000-0000-000000000000",
								"clientId":   "00000000-0000-0000-0000-000000000000",
							},
							"clusterUrl":        "https://yourcluster.kusto.windows.net",
							"defaultDatabase":   "mydb",
							"queryTimeout":      "30s",
							"dataConsistency":   string(DataConsistencyStrong),
							"defaultEditorMode": string(EditorModeVisual),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAzureClientSecret): "changeme",
							string(SecureJsonDataKeyOpenAIAPIKey):      "sk-XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX",
						},
					},
				},
			},
			"legacyClientSecret": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Legacy top-level Client Secret (pre-`azureCredentials`)",
					Description: "Datasources provisioned before the credentials migration stored " +
						"auth at top level (`jsonData.{azureCloud,tenantId,clientId,onBehalfOf}`) " +
						"and the secret in `secureJsonData.clientSecret`. Both frontend and " +
						"backend still accept this shape via legacy fallbacks " +
						"(`AzureCredentialsConfig.ts:46-67`, `adxcredentials/builder.go:40-87`). " +
						"Opening the editor and saving migrates the datasource to the modern " +
						"`jsonData.azureCredentials` shape.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"azureCloud":        string(LegacyCloudNamePublic),
							"tenantId":          "00000000-0000-0000-0000-000000000000",
							"clientId":          "00000000-0000-0000-0000-000000000000",
							"clusterUrl":        "https://yourcluster.kusto.windows.net",
							"defaultDatabase":   "mydb",
							"queryTimeout":      "30s",
							"dataConsistency":   string(DataConsistencyStrong),
							"defaultEditorMode": string(EditorModeVisual),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyClientSecret): "changeme",
						},
					},
				},
			},
		},
	}
}
