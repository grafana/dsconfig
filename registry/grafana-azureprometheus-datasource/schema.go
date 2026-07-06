package azureprometheusdatasource

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
// truth for the Azure Prometheus datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// for the Azure Prometheus datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the k8s-style SDK plugin schema bundle Grafana's
// datasource API server serves as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Azure
// Monitor Managed Service for Prometheus datasource, covering the default
// (schema-defaults) configuration plus one example per Azure authentication
// type the config editor supports and one legacy provisioning example.
//
// Each example value is a full instance-settings object with the plugin
// configuration nested under `jsonData` and the relevant write-only secrets
// under `secureJsonData` (placeholder values — replace them with real
// secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Default configuration",
					Description: "The defaults a new datasource starts with. Azure Prometheus has no schema-side " +
						"credential default (the config editor chooses one at load time based on Grafana's " +
						"`azure.managedIdentityEnabled` / `workloadIdentityEnabled` / `userIdentityEnabled` " +
						"config), so a fresh datasource has no `azureCredentials` and no secrets. `httpMethod` " +
						"defaults to POST (uppercased on load by both editor and backend). Users must complete " +
						"authentication before the datasource is usable.",
					Value: map[string]any{
						"url": "https://<workspace>.prometheus.monitor.azure.com",
						"jsonData": map[string]any{
							"httpMethod": string(HTTPMethodPOST),
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
					Description: "Authenticate via an Entra ID (Azure AD) app registration and client secret. " +
						"The secret is stored write-only in `secureJsonData.azureClientSecret`; the rest of the " +
						"credential lives in `jsonData.azureCredentials`.",
					Value: map[string]any{
						"url": "https://mimir.prometheus.monitor.azure.com",
						"jsonData": map[string]any{
							"azureCredentials": map[string]any{
								"authType":   string(AuthTypeClientSecret),
								"azureCloud": "AzureCloud",
								"tenantId":   "00000000-0000-0000-0000-000000000000",
								"clientId":   "00000000-0000-0000-0000-000000000000",
							},
							"httpMethod": string(HTTPMethodPOST),
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
					Description: "Authenticate using the Grafana host's Azure Managed Identity. Only available " +
						"when Grafana's `azure.managedIdentityEnabled` is true. No secret required.",
					Value: map[string]any{
						"url": "https://mimir.prometheus.monitor.azure.com",
						"jsonData": map[string]any{
							"azureCredentials": map[string]any{
								"authType": string(AuthTypeManagedIdentity),
							},
							"httpMethod": string(HTTPMethodPOST),
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
					Description: "Authenticate using an Entra ID Workload Identity federated to the Kubernetes " +
						"service account Grafana runs under. Only available when Grafana's " +
						"`azure.workloadIdentityEnabled` is true. No secret required.",
					Value: map[string]any{
						"url": "https://mimir.prometheus.monitor.azure.com",
						"jsonData": map[string]any{
							"azureCredentials": map[string]any{
								"authType": string(AuthTypeWorkloadIdentity),
							},
							"httpMethod": string(HTTPMethodPOST),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAzureClientSecret): "",
						},
					},
				},
			},
			"currentUser": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Current User (with Client Secret fallback)",
					Description: "Forward the signed-in Grafana user's OAuth identity to Azure. Only available " +
						"when Grafana's `azure.userIdentityEnabled` is true. Enabling `serviceCredentialsEnabled` " +
						"lets user-context-less features (alerting, recorded queries, reporting) fall back to a " +
						"service credential — here a `clientsecret` fallback whose secret shares the " +
						"`secureJsonData.azureClientSecret` key.",
					Value: map[string]any{
						"url": "https://mimir.prometheus.monitor.azure.com",
						"jsonData": map[string]any{
							"azureCredentials": map[string]any{
								"authType":                  string(AuthTypeCurrentUser),
								"serviceCredentialsEnabled": true,
								"serviceCredentials": map[string]any{
									"authType":   string(AuthTypeClientSecret),
									"azureCloud": "AzureCloud",
									"tenantId":   "00000000-0000-0000-0000-000000000000",
									"clientId":   "00000000-0000-0000-0000-000000000000",
								},
							},
							"httpMethod": string(HTTPMethodPOST),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAzureClientSecret): "changeme",
						},
					},
				},
			},
			"customEndpointResourceID": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Client Secret with custom endpoint resource ID (provisioning-only)",
					Description: "Override the OAuth scope audience the backend derives from the resolved Azure " +
						"cloud's `prometheusResourceId` property by supplying `jsonData.azureEndpointResourceId`. " +
						"Editor-hidden; provisioning-only. See `pkg/azureauth/azure.go:58-63`.",
					Value: map[string]any{
						"url": "https://mimir.prometheus.monitor.azure.com",
						"jsonData": map[string]any{
							"azureCredentials": map[string]any{
								"authType":   string(AuthTypeClientSecret),
								"azureCloud": "AzureCloud",
								"tenantId":   "00000000-0000-0000-0000-000000000000",
								"clientId":   "00000000-0000-0000-0000-000000000000",
							},
							"azureEndpointResourceId": "https://prometheus.monitor.azure.com",
							"httpMethod":              string(HTTPMethodPOST),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAzureClientSecret): "changeme",
						},
					},
				},
			},
			"legacyClientSecret": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Legacy secure key (pre-`azureClientSecret` migration)",
					Description: "Datasources provisioned before the credentials migration stored the client " +
						"secret in `secureJsonData.clientSecret` instead of the modern " +
						"`secureJsonData.azureClientSecret`. The backend still accepts this shape via the " +
						"shared `grafana-azure-sdk-go/v2/azcredentials/builder.go` fallback. Opening the editor " +
						"and saving migrates the datasource to the modern key.",
					Value: map[string]any{
						"url": "https://mimir.prometheus.monitor.azure.com",
						"jsonData": map[string]any{
							"azureCredentials": map[string]any{
								"authType":   string(AuthTypeClientSecret),
								"azureCloud": "AzureCloud",
								"tenantId":   "00000000-0000-0000-0000-000000000000",
								"clientId":   "00000000-0000-0000-0000-000000000000",
							},
							"httpMethod": string(HTTPMethodPOST),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyClientSecret): "changeme",
						},
					},
				},
			},
			"migratedFromPrometheus": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Migrated from vanilla Prometheus",
					Description: "Datasources migrated from the vanilla Prometheus plugin carry the sentinel " +
						"flag `jsonData['prometheus-type-migration'] = true`, which triggers the " +
						"'Data source migrated' banner at " +
						"`src/configuration/DataSourceHttpSettingsOverhaul.tsx:101-117`. The banner is purely " +
						"informational — nothing in the runtime depends on the flag.",
					Value: map[string]any{
						"url": "https://mimir.prometheus.monitor.azure.com",
						"jsonData": map[string]any{
							"azureCredentials": map[string]any{
								"authType":   string(AuthTypeClientSecret),
								"azureCloud": "AzureCloud",
								"tenantId":   "00000000-0000-0000-0000-000000000000",
								"clientId":   "00000000-0000-0000-0000-000000000000",
							},
							"httpMethod":                string(HTTPMethodPOST),
							"prometheus-type-migration": true,
							"prometheusType":            string(PromApplicationPrometheus),
							"prometheusVersion":         "2.50.1",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAzureClientSecret): "changeme",
						},
					},
				},
			},
		},
	}
}
