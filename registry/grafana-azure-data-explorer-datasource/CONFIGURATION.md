# Azure Data Explorer Datasource configuration

Configuration reference for the **Azure Data Explorer Datasource** data source (`grafana-azure-data-explorer-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-azure-data-explorer-datasource/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.clusterUrl` | string | jsonData |  | The default cluster url for this data source. |
| `jsonData.azureCredentials` | any | jsonData |  | Choose the type of authentication to Azure services |
| `secureJsonData.azureClientSecret` 🔒 | string | secureJsonData |  | Client secret of the App Registration. Written write-only by `@grafana/azure-sdk`'s `AzureCredentialsForm`; check `secureJsonFields.azureClientSecret` on the read side. Used when `jsonData.azureCredentials.authType` is `clientsecret` or `clientsecret-obo`. |
| `secureJsonData.clientSecret` 🔒 | string | secureJsonData |  | Legacy client secret key. Preserved for backward compatibility with datasources provisioned before the credential migration to `azureClientSecret`. The backend reads it as a fallback in `pkg/azuredx/adxauth/adxcredentials/builder.go:57` (`getFromLegacy`) when no modern `jsonData.azureCredentials` is present. |
| `jsonData.oauthPassThru` | boolean | jsonData |  | Set to `true` by `@grafana/azure-sdk`'s `updateDatasourceCredentials` whenever the selected auth type is `clientsecret-obo` (On-Behalf-Of). The plugin backend enforces `oauthPassThru == true` for OBO in `pkg/azuredx/adxauth/adxcredentials/builder.go:89-98` (`ensureOnBehalfOfSupported`); creation fails otherwise. |
| `jsonData.queryTimeout` | string | jsonData |  | This value controls the client query timeout. |
| `jsonData.dynamicCaching` | boolean | jsonData |  | By enabling this feature Grafana will dynamically apply cache settings on a per query basis and the default cache max age will be ignored. For time series queries we will use the bin size to widen the time range but also as cache max age. |
| `jsonData.cacheMaxAge` | string | jsonData |  | By default the cache is disabled. If you want to enable the query caching please specify a max timespan for the cache to live. |
| `jsonData.dataConsistency` | enum (strongconsistency, weakconsistency) | jsonData |  | Query consistency controls how queries and updates are synchronized. Defaults to Strong. For more information see the Azure Data Explorer documentation. |
| `jsonData.defaultEditorMode` | enum (visual, raw) | jsonData |  | This setting dictates which mode the editor will open in. Defaults to Visual. |
| `jsonData.defaultDatabase` | enum | jsonData |  | Default database |
| `jsonData.useSchemaMapping` | boolean | jsonData |  | If enabled, allows tables, functions, and materialized views to be mapped to user friendly names. |
| `jsonData.schemaMappings` | list | jsonData |  | Schema mappings |
| `jsonData.schemaMappings[].type` | enum (function, table, materializedView) | jsonData |  |  |
| `jsonData.schemaMappings[].value` | string | jsonData |  |  |
| `jsonData.schemaMappings[].name` | string | jsonData |  |  |
| `jsonData.schemaMappings[].database` | string | jsonData |  |  |
| `jsonData.schemaMappings[].displayName` | string | jsonData |  |  |
| `jsonData.application` | string | jsonData |  | Application name to be displayed in ADX. |
| `jsonData.enableUserTracking` | boolean | jsonData |  | With this feature enabled, Grafana will pass the logged in user's username in the `x-ms-user-id` header and in the `x-ms-client-request-id` header when sending requests to ADX. Can be useful when tracking needs to be done in ADX. |
| `jsonData.keepCookies` | list | jsonData |  | Grafana proxy deletes forwarded cookies by default. Specify cookies by name that should be forwarded to the data source. |
| `secureJsonData.OpenAIAPIKey` 🔒 | string | secureJsonData |  | OpenAI API key used by the plugin's `askOpenAI` resource endpoint (`pkg/azuredx/resource_handler.go:55-89`) to power the AI query assistant. There is no config editor UI to set it — the ADX ConfigEditor only inspects `secureJsonFields.OpenAIAPIKey` to decide whether the Additional settings section starts open (`src/components/ConfigEditor/index.tsx:58`). Populate it through provisioning YAML or via the datasource API. |
| `jsonData.minimalCache` | number | jsonData |  | Declared on `AdxDataSourceOptions` (`src/types/index.ts:115`) but never written by the config editor and never read by the plugin backend (`pkg/azuredx/models/settings.go` reads only `cacheMaxAge`). Dead field; preserved so provisioned datasources with the property still parse. |
| `jsonData.azureCloud` | string | jsonData |  | Legacy Azure cloud discriminator written by the pre-2023 ADX config editor. Resolved by `pkg/azuredx/adxauth/adxcredentials/builder.go:107-121` `resolveLegacyCloudName`: `azuremonitor` → `AzureCloud`, `chinaazuremonitor` → `AzureChinaCloud`, `govazuremonitor` → `AzureUSGovernment`, empty → `AzureCloud`. The modern `AzureCredentialsForm` writes cloud inside `jsonData.azureCredentials.azureCloud` and clears this field. |
| `jsonData.onBehalfOf` | boolean | jsonData |  | Legacy On-Behalf-Of flag; when true and paired with legacy top-level `tenantId`/`clientId`/`secureJsonData.clientSecret`, the backend legacy fallback (`pkg/azuredx/adxauth/adxcredentials/builder.go:71-84`) constructs an `AzureClientSecretOboCredentials` instead of a plain `AzureClientSecretCredentials`. The modern form migrates this to `authType: clientsecret-obo`. |
| `jsonData.tenantId` | string | jsonData |  | Legacy top-level tenant ID (paired with legacy `clientId`/`secureJsonData.clientSecret`). Preserved so pre-migration datasources continue to authenticate; the current `AzureCredentialsForm` writes tenant inside `jsonData.azureCredentials.tenantId` and clears this field. |
| `jsonData.clientId` | string | jsonData |  | Legacy top-level client ID (paired with legacy `tenantId`/`secureJsonData.clientSecret`). Preserved so pre-migration datasources continue to authenticate; the current `AzureCredentialsForm` writes client inside `jsonData.azureCredentials.clientId` and clears this field. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Azure Data Explorer Datasource
    type: grafana-azure-data-explorer-datasource
    access: proxy
    jsonData:
      dataConsistency: strongconsistency
      defaultEditorMode: visual
      dynamicCaching: false
      enableUserTracking: false
      queryTimeout: "30s"
      useSchemaMapping: false
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_azure_data_explorer_datasource" {
  type = "grafana-azure-data-explorer-datasource"
  name = "Azure Data Explorer Datasource"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    dataConsistency = "strongconsistency"
    defaultEditorMode = "visual"
    dynamicCaching = false
    enableUserTracking = false
    queryTimeout = "30s"
    useSchemaMapping = false
  })
}
```

