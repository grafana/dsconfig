# Azure Monitor configuration

Configuration reference for the **Azure Monitor** data source (`grafana-azure-monitor-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/azure-monitor/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.azureCredentials` | any | jsonData |  | Choose the type of authentication to Azure services |
| `jsonData.subscriptionId` | enum | jsonData |  | Default Subscription |
| `jsonData.basicLogsEnabled` | boolean | jsonData |  | Enabling this feature incurs Azure Monitor per-query costs on dashboard panels that query tables configured for Basic Logs. |
| `jsonData.timeout` | number | jsonData |  | HTTP request timeout in seconds |
| `jsonData.keepCookies` | list | jsonData |  | Grafana proxy deletes forwarded cookies by default. Specify cookies by name that should be forwarded to the data source. |
| `secureJsonData.azureClientSecret` 🔒 | string | secureJsonData |  | Client secret of the App Registration. Written write-only by `@grafana/azure-sdk`'s `AzureCredentialsForm`; check `secureJsonFields.azureClientSecret` on the read side. Used when `jsonData.azureCredentials.authType` is `clientsecret`, `clientsecret-obo`, or `currentuser` with a `clientsecret` fallback. |
| `secureJsonData.clientCertificate` 🔒 | string (multiline) | secureJsonData |  | Client certificate body (PEM text or base64-encoded PFX depending on `azureCredentials.certificateFormat`). Written by `@grafana/azure-sdk`. |
| `secureJsonData.privateKey` 🔒 | string (multiline) | secureJsonData |  | Private key paired with the PEM client certificate. Only used when `azureCredentials.certificateFormat == 'pem'`. Written by `@grafana/azure-sdk`. |
| `secureJsonData.certificatePassword` 🔒 | string | secureJsonData |  | Password protecting the PFX client certificate bundle. Only used when `azureCredentials.certificateFormat == 'pfx'`. Written by `@grafana/azure-sdk`. |
| `secureJsonData.clientSecret` 🔒 | string | secureJsonData |  | Legacy client secret key. Preserved for backward compatibility with datasources provisioned before the credential migration to `azureClientSecret`. The backend reads it as a fallback when `secureJsonData.azureClientSecret` is missing (grafana-azure-sdk-go v2 `azcredentials/builder.go` `getFromCredentialsObject` case `clientsecret` and `clientsecret-obo`, and the plugin's own `azmoncredentials/builder.go` `getFromLegacy`). |
| `secureJsonData.password` 🔒 | string | secureJsonData |  | Entra password used only when the backend `azcredentials` builder encounters `authType == 'ad-password'`. The Azure Monitor config editor does not expose this authentication type today; the field is only reachable through provisioning YAML. |
| `jsonData.appInsightsAppId` | string | jsonData |  | Deprecated. App ID of the Application Insights resource still parsed by the backend (`pkg/azuremonitor/types/types.go:31` `AzureMonitorSettings.AppInsightsAppId`) for provisioned datasources migrated from the old Application Insights integration. The config editor no longer renders this field (Application Insights was folded into Log Analytics — see `docs/deprecated-application-insights/`). |
| `secureJsonData.appInsightsApiKey` 🔒 | string | secureJsonData |  | Deprecated. Application Insights API key preserved on migrated datasources. Present in `src/types/types.ts:56` (`AzureMonitorDataSourceSecureJsonData.appInsightsApiKey`); no editor UI renders it, and the backend no longer authenticates against Application Insights directly. |
| `jsonData.logAnalyticsDefaultWorkspace` | string | jsonData |  | Deprecated. Default Log Analytics workspace, still parsed by the backend (`pkg/azuremonitor/types/types.go:30` `AzureMonitorSettings.LogAnalyticsDefaultWorkspace`) for legacy datasources. Marked `@deprecated` in `src/types/types.ts:45`. |
| `jsonData.azureLogAnalyticsSameAs` | any | jsonData |  | Deprecated. When set to `false`, the backend hard-fails Log Analytics queries with 'credentials for Log Analytics are no longer supported. Go to the data source configuration to update Azure Monitor credentials' (`pkg/azuremonitor/loganalytics/azure-log-analytics-datasource.go:415-432`). Accepts a bool or a string that `strconv.ParseBool` can parse. |
| `jsonData.logAnalyticsTenantId` | string | jsonData |  | Deprecated. Legacy Log Analytics tenant ID. Marked `@deprecated` in `src/types/types.ts:39`. Frontend-only (no backend reader). |
| `jsonData.logAnalyticsClientId` | string | jsonData |  | Deprecated. Legacy Log Analytics client ID. Marked `@deprecated` in `src/types/types.ts:41`. Frontend-only (no backend reader). |
| `jsonData.logAnalyticsSubscriptionId` | string | jsonData |  | Deprecated. Legacy Log Analytics subscription ID. Marked `@deprecated` in `src/types/types.ts:43`. Frontend-only (no backend reader). |
| `jsonData.azureAuthType` | string | jsonData |  | Legacy top-level auth-type field written by the pre-2023 Azure Monitor config editor before credentials were nested under `jsonData.azureCredentials`. Read by the backend legacy fallback in `pkg/azuremonitor/azmoncredentials/builder.go:34-56` (`getFromLegacy`) and by the frontend at `src/credentials.ts:40-57` when no `azureCredentials` object is present. `updateDatasourceCredentials` in `@grafana/azure-sdk` unsets this field whenever the editor saves. |
| `jsonData.cloudName` | string | jsonData |  | Legacy Azure cloud discriminator (values `azuremonitor` / `chinaazuremonitor` / `govazuremonitor` / `customizedazuremonitor`). Read by `pkg/azuremonitor/azmoncredentials/builder.go:126-142` `resolveLegacyCloudName` and by `resolveLegacyCloudName` in `@grafana/azure-sdk` (`src/clouds.ts:32-47`). Also gates `customizedRoutes` — `azureCloud == 'AzureCustomizedCloud'` (derived from `cloudName == 'customizedazuremonitor'`) is the only way the backend consults `jsonData.customizedRoutes`. |
| `jsonData.tenantId` | string | jsonData |  | Legacy top-level tenant ID (paired with legacy `clientId`/`clientSecret`). Preserved so pre-migration datasources continue to authenticate; the current `AzureCredentialsForm` writes tenant inside `jsonData.azureCredentials.tenantId` and clears this field. |
| `jsonData.clientId` | string | jsonData |  | Legacy top-level client ID (paired with legacy `tenantId`/`clientSecret`). Preserved so pre-migration datasources continue to authenticate; the current `AzureCredentialsForm` writes client inside `jsonData.azureCredentials.clientId` and clears this field. |
| `jsonData.oauthPassThru` | boolean | jsonData |  | Set to `true` by `@grafana/azure-sdk`'s `updateDatasourceCredentials` whenever the selected auth type is `clientsecret-obo` or `currentuser`. Consumed by Grafana's shared HTTP client to forward the caller's OAuth token. |
| `jsonData.disableGrafanaCache` | boolean | jsonData |  | Set to `true` by `@grafana/azure-sdk`'s `updateDatasourceCredentials` when `currentuser` authentication is selected (`grafana-azure-sdk-react/src/credentials/AzureCredentialsConfig.ts:424`) to prevent user-scoped results from leaking across users. No editor UI. |
| `jsonData.customizedRoutes` | map | jsonData |  | Map of route name → `{ URL, Scopes, Headers }` overriding the default Azure Monitor / Log Analytics / Resource Graph endpoints. Only consulted by the backend when the resolved cloud is `AzureCustomizedCloud` (`pkg/azuremonitor/routes.go:31-37,93-106`). Not rendered by the config editor — configured via provisioning YAML. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Azure Monitor
    type: grafana-azure-monitor-datasource
    access: proxy
    jsonData:
      basicLogsEnabled: false
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_azure_monitor_datasource" {
  type = "grafana-azure-monitor-datasource"
  name = "Azure Monitor"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    basicLogsEnabled = false
  })
}
```

