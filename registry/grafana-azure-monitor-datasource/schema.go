package azuremonitordatasource

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
// of truth for the Azure Monitor datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig
// schema for the Azure Monitor datasource.
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
// authentication type Azure Monitor supports, plus a legacy provisioning
// example. Every value is a full instance-settings object with the plugin
// configuration nested under `jsonData` and the relevant write-only
// secrets under `secureJsonData` (placeholder values — replace them with
// real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Default configuration",
					Description: "The defaults a new datasource starts with. Azure Monitor has no schema-side " +
						"credential default (the config editor chooses one at load time based on Grafana's " +
						"managedIdentityEnabled / workloadIdentityEnabled config), so a fresh datasource has " +
						"an empty `azureCredentials` object and no secrets. Users must complete authentication " +
						"before the datasource is usable.",
					Value: map[string]any{
						"jsonData": map[string]any{},
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
						"The secret is stored write-only in `secureJsonData.azureClientSecret`; the rest of " +
						"the credential lives in `jsonData.azureCredentials`.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"azureCredentials": map[string]any{
								"authType":   string(AuthTypeClientSecret),
								"azureCloud": "AzureCloud",
								"tenantId":   "00000000-0000-0000-0000-000000000000",
								"clientId":   "00000000-0000-0000-0000-000000000000",
							},
							"subscriptionId": "00000000-0000-0000-0000-000000000000",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAzureClientSecret): "changeme",
						},
					},
				},
			},
			"clientCertificatePEM": {
				ExampleProps: spec3.ExampleProps{
					Summary: "App Registration (Client Certificate — PEM)",
					Description: "Authenticate via an Entra ID app registration and a PEM client certificate. " +
						"Both `clientCertificate` (PEM body) and `privateKey` are stored write-only in " +
						"secureJsonData.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"azureCredentials": map[string]any{
								"authType":          string(AuthTypeClientCertificate),
								"azureCloud":        "AzureCloud",
								"tenantId":          "00000000-0000-0000-0000-000000000000",
								"clientId":          "00000000-0000-0000-0000-000000000000",
								"certificateFormat": string(CertificateFormatPEM),
							},
							"subscriptionId": "00000000-0000-0000-0000-000000000000",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyClientCertificate): "-----BEGIN CERTIFICATE-----\n...\n-----END CERTIFICATE-----",
							string(SecureJsonDataKeyPrivateKey):        "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----",
						},
					},
				},
			},
			"clientCertificatePFX": {
				ExampleProps: spec3.ExampleProps{
					Summary: "App Registration (Client Certificate — PFX)",
					Description: "Authenticate via an Entra ID app registration and a base64-encoded PFX " +
						"certificate bundle. `certificateFormat=pfx` requires `certificatePassword` and " +
						"omits `privateKey`.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"azureCredentials": map[string]any{
								"authType":          string(AuthTypeClientCertificate),
								"azureCloud":        "AzureCloud",
								"tenantId":          "00000000-0000-0000-0000-000000000000",
								"clientId":          "00000000-0000-0000-0000-000000000000",
								"certificateFormat": string(CertificateFormatPFX),
							},
							"subscriptionId": "00000000-0000-0000-0000-000000000000",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyClientCertificate):   "MIIKAwIBAzCCCcMGCSqGSIb3DQEHAaCCCbQ...",
							string(SecureJsonDataKeyCertificatePassword): "changeme",
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
						"jsonData": map[string]any{
							"azureCredentials": map[string]any{
								"authType": string(AuthTypeManagedIdentity),
							},
							"subscriptionId": "00000000-0000-0000-0000-000000000000",
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
						"jsonData": map[string]any{
							"azureCredentials": map[string]any{
								"authType": string(AuthTypeWorkloadIdentity),
							},
							"subscriptionId": "00000000-0000-0000-0000-000000000000",
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
					Description: "Forward the signed-in Grafana user's OAuth token to Azure. Only available " +
						"when Grafana's `azure.userIdentityEnabled` and the `azureMonitorEnableUserAuth` " +
						"feature toggle are both on. Enabling `serviceCredentialsEnabled` lets " +
						"user-context-less features (alerting, recorded queries, reporting) fall back to a " +
						"service credential — here a `clientsecret` fallback, whose secret shares the " +
						"`secureJsonData.azureClientSecret` key.",
					Value: map[string]any{
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
							"subscriptionId":      "00000000-0000-0000-0000-000000000000",
							"oauthPassThru":       true,
							"disableGrafanaCache": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAzureClientSecret): "changeme",
						},
					},
				},
			},
			"customizedCloud": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Customized Azure Cloud (provisioning-only)",
					Description: "Point the datasource at a non-standard sovereign Azure cloud by supplying the " +
						"per-route URLs in `jsonData.customizedRoutes`. Requires legacy " +
						"`jsonData.cloudName = customizedazuremonitor` (see `pkg/azuremonitor/routes.go:31-37`). " +
						"The config editor cannot express this shape today — it is provisioning-only.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"cloudName": string(LegacyCloudNameCustomized),
							"azureCredentials": map[string]any{
								"authType":   string(AuthTypeClientSecret),
								"azureCloud": "AzureCustomizedCloud",
								"tenantId":   "00000000-0000-0000-0000-000000000000",
								"clientId":   "00000000-0000-0000-0000-000000000000",
							},
							"customizedRoutes": map[string]any{
								"Azure Monitor": map[string]any{
									"URL":     "https://management.contoso.cloud",
									"Scopes":  []string{"https://management.contoso.cloud/.default"},
									"Headers": map[string]string{"x-ms-app": "Grafana"},
								},
								"Azure Log Analytics": map[string]any{
									"URL":    "https://api.loganalytics.contoso.cloud",
									"Scopes": []string{"https://api.loganalytics.contoso.cloud/.default"},
								},
								"Azure Resource Graph": map[string]any{
									"URL":    "https://management.contoso.cloud",
									"Scopes": []string{"https://management.contoso.cloud/.default"},
								},
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAzureClientSecret): "changeme",
						},
					},
				},
			},
			"legacyClientSecret": {
				ExampleProps: spec3.ExampleProps{
					Summary: "Legacy top-level Client Secret (pre-`azureCredentials`)",
					Description: "Datasources provisioned before the credentials migration stored auth type + " +
						"tenant/client at top level (`jsonData.{azureAuthType,cloudName,tenantId,clientId}`) " +
						"and the secret in `secureJsonData.clientSecret`. Both frontend and backend still " +
						"accept this shape via legacy fallbacks (`src/credentials.ts:33-64`, " +
						"`pkg/azuremonitor/azmoncredentials/builder.go:33-116`). Opening the editor and " +
						"saving migrates the datasource to the modern `jsonData.azureCredentials` shape.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"azureAuthType":  string(AuthTypeClientSecret),
							"cloudName":      string(LegacyCloudNamePublic),
							"tenantId":       "00000000-0000-0000-0000-000000000000",
							"clientId":       "00000000-0000-0000-0000-000000000000",
							"subscriptionId": "00000000-0000-0000-0000-000000000000",
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
