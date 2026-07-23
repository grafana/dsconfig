# Azure Monitor configuration

How to configure the **Azure Monitor** data source (`grafana-azure-monitor-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/azure-monitor/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Authentication](#authentication)
- [Azure Monitor](#azure-monitor)
- [Additional settings](#additional-settings) — _optional_
- [Deprecated Application Insights / Log Analytics fields](#deprecated-application-insights--log-analytics-fields) — _optional_
- [Legacy top-level credentials](#legacy-top-level-credentials) — _optional_
- [Customized cloud](#customized-cloud) — _optional_

## Authentication

### Authentication type

_optional_

Choose the type of authentication to Azure services.

Discriminated-union object written by the `@grafana/azure-sdk` `AzureCredentialsForm`. Shape depends on `authType`:

- `msi` — `{ authType: 'msi' }`
- `workloadidentity` — `{ authType: 'workloadidentity' }`
- `clientsecret` — `{ authType: 'clientsecret', azureCloud, tenantId, clientId }` (secret in `secureJsonData.azureClientSecret`)
- `clientcertificate` — `{ authType: 'clientcertificate', azureCloud, tenantId, clientId, certificateFormat: 'pem' | 'pfx' }` (cert/key/password in `secureJsonData.clientCertificate` / `secureJsonData.privateKey` / `secureJsonData.certificatePassword`)
- `currentuser` — `{ authType: 'currentuser', serviceCredentialsEnabled?: boolean, serviceCredentials?: {authType: 'msi' | 'workloadidentity' | 'clientsecret', ...} }`

See `src/credentials/AzureCredentials.ts` in `grafana/grafana-azure-sdk-react` and the backend builder in `github.com/grafana/grafana-azure-sdk-go/v2/azcredentials/builder.go`.

### Client Secret

_🔒 secret (write-only) · optional · string_

Client secret of the App Registration. Written write-only by `@grafana/azure-sdk`'s `AzureCredentialsForm`; check `secureJsonFields.azureClientSecret` on the read side. Used when `jsonData.azureCredentials.authType` is `clientsecret`, `clientsecret-obo`, or `currentuser` with a `clientsecret` fallback.

| | |
|---|---|
| Example | `XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX` |

### Client Certificate

_🔒 secret (write-only) · optional · multiline text_

Client certificate body (PEM text or base64-encoded PFX depending on `azureCredentials.certificateFormat`). Written by `@grafana/azure-sdk`.

| | |
|---|---|
| Example | `-----BEGIN CERTIFICATE-----` |

### Private Key

_🔒 secret (write-only) · optional · multiline text_

Private key paired with the PEM client certificate. Only used when `azureCredentials.certificateFormat == 'pem'`. Written by `@grafana/azure-sdk`.

| | |
|---|---|
| Example | `-----BEGIN PRIVATE KEY-----` |

### Certificate Password

_🔒 secret (write-only) · optional · string_

Password protecting the PFX client certificate bundle. Only used when `azureCredentials.certificateFormat == 'pfx'`. Written by `@grafana/azure-sdk`.

| | |
|---|---|
| Example | `XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX` |

### password

_🔒 secret (write-only) · optional · string_

Entra password used only when the backend `azcredentials` builder encounters `authType == 'ad-password'`. The Azure Monitor config editor does not expose this authentication type today; the field is only reachable through provisioning YAML.

### clientSecret

_🔒 secret (write-only) · optional · string_

Legacy client secret key. Preserved for backward compatibility with datasources provisioned before the credential migration to `azureClientSecret`. The backend reads it as a fallback when `secureJsonData.azureClientSecret` is missing (grafana-azure-sdk-go v2 `azcredentials/builder.go` `getFromCredentialsObject` case `clientsecret` and `clientsecret-obo`, and the plugin's own `azmoncredentials/builder.go` `getFromLegacy`).

### oauthPassThru

_optional · boolean_

Set to `true` by `@grafana/azure-sdk`'s `updateDatasourceCredentials` whenever the selected auth type is `clientsecret-obo` or `currentuser`. Consumed by Grafana's shared HTTP client to forward the caller's OAuth token.

### disableGrafanaCache

_optional · boolean_

Set to `true` by `@grafana/azure-sdk`'s `updateDatasourceCredentials` when `currentuser` authentication is selected (`grafana-azure-sdk-react/src/credentials/AzureCredentialsConfig.ts:424`) to prevent user-scoped results from leaking across users. No editor UI.

## Azure Monitor

### Default Subscription

_optional · select_

### Enable Basic Logs

_optional · toggle_

Enabling this feature incurs Azure Monitor per-query costs on dashboard panels that query tables configured for Basic Logs.

| | |
|---|---|
| Default | `false` |

## Additional settings

Additional settings are optional settings that can be configured for more control over your data source. This includes Secure Socks Proxy, request timeout, and forwarded cookies.

_This section is optional._

### Timeout

_optional · number_

HTTP request timeout in seconds.

| | |
|---|---|
| Example | `Timeout in seconds` |

### Allowed cookies

_optional · list_

Grafana proxy deletes forwarded cookies by default. Specify cookies by name that should be forwarded to the data source.

| | |
|---|---|
| Example | `New cookie (hit enter to add)` |

## Deprecated Application Insights / Log Analytics fields

_This section is optional._

### appInsightsAppId

_optional · string_

Deprecated. App ID of the Application Insights resource still parsed by the backend (`pkg/azuremonitor/types/types.go:31` `AzureMonitorSettings.AppInsightsAppId`) for provisioned datasources migrated from the old Application Insights integration. The config editor no longer renders this field (Application Insights was folded into Log Analytics — see `docs/deprecated-application-insights/`).

### appInsightsApiKey

_🔒 secret (write-only) · optional · string_

Deprecated. Application Insights API key preserved on migrated datasources. Present in `src/types/types.ts:56` (`AzureMonitorDataSourceSecureJsonData.appInsightsApiKey`); no editor UI renders it, and the backend no longer authenticates against Application Insights directly.

### logAnalyticsDefaultWorkspace

_optional · string_

Deprecated. Default Log Analytics workspace, still parsed by the backend (`pkg/azuremonitor/types/types.go:30` `AzureMonitorSettings.LogAnalyticsDefaultWorkspace`) for legacy datasources. Marked `@deprecated` in `src/types/types.ts:45`.

### azureLogAnalyticsSameAs

_optional_

Deprecated. When set to `false`, the backend hard-fails Log Analytics queries with 'credentials for Log Analytics are no longer supported. Go to the data source configuration to update Azure Monitor credentials' (`pkg/azuremonitor/loganalytics/azure-log-analytics-datasource.go:415-432`). Accepts a bool or a string that `strconv.ParseBool` can parse.

### logAnalyticsTenantId

_optional · string_

Deprecated. Legacy Log Analytics tenant ID. Marked `@deprecated` in `src/types/types.ts:39`. Frontend-only (no backend reader).

### logAnalyticsClientId

_optional · string_

Deprecated. Legacy Log Analytics client ID. Marked `@deprecated` in `src/types/types.ts:41`. Frontend-only (no backend reader).

### logAnalyticsSubscriptionId

_optional · string_

Deprecated. Legacy Log Analytics subscription ID. Marked `@deprecated` in `src/types/types.ts:43`. Frontend-only (no backend reader).

## Legacy top-level credentials

_This section is optional._

### azureAuthType

_optional · string_

Legacy top-level auth-type field written by the pre-2023 Azure Monitor config editor before credentials were nested under `jsonData.azureCredentials`. Read by the backend legacy fallback in `pkg/azuremonitor/azmoncredentials/builder.go:34-56` (`getFromLegacy`) and by the frontend at `src/credentials.ts:40-57` when no `azureCredentials` object is present. `updateDatasourceCredentials` in `@grafana/azure-sdk` unsets this field whenever the editor saves.

### cloudName

_optional · string_

Legacy Azure cloud discriminator (values `azuremonitor` / `chinaazuremonitor` / `govazuremonitor` / `customizedazuremonitor`). Read by `pkg/azuremonitor/azmoncredentials/builder.go:126-142` `resolveLegacyCloudName` and by `resolveLegacyCloudName` in `@grafana/azure-sdk` (`src/clouds.ts:32-47`). Also gates `customizedRoutes` — `azureCloud == 'AzureCustomizedCloud'` (derived from `cloudName == 'customizedazuremonitor'`) is the only way the backend consults `jsonData.customizedRoutes`.

### tenantId

_optional · string_

Legacy top-level tenant ID (paired with legacy `clientId`/`clientSecret`). Preserved so pre-migration datasources continue to authenticate; the current `AzureCredentialsForm` writes tenant inside `jsonData.azureCredentials.tenantId` and clears this field.

### clientId

_optional · string_

Legacy top-level client ID (paired with legacy `tenantId`/`clientSecret`). Preserved so pre-migration datasources continue to authenticate; the current `AzureCredentialsForm` writes client inside `jsonData.azureCredentials.clientId` and clears this field.

## Customized cloud

_This section is optional._

### customizedRoutes

_optional_

Map of route name → `{ URL, Scopes, Headers }` overriding the default Azure Monitor / Log Analytics / Resource Graph endpoints. Only consulted by the backend when the resolved cloud is `AzureCustomizedCloud` (`pkg/azuremonitor/routes.go:31-37,93-106`). Not rendered by the config editor — configured via provisioning YAML.

