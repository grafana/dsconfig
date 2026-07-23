# Azure Data Explorer Datasource configuration

How to configure the **Azure Data Explorer Datasource** data source (`grafana-azure-data-explorer-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-azure-data-explorer-datasource/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)
- [Query Optimizations](#query-optimizations) — _optional_
- [Database schema settings](#database-schema-settings) — _optional_
- [Application](#application) — _optional_
- [Tracking](#tracking) — _optional_
- [Additional settings](#additional-settings) — _optional_
- [OpenAI (provisioning-only)](#openai-provisioning-only) — _optional_
- [Legacy top-level credentials](#legacy-top-level-credentials) — _optional_
- [Deprecated / dead fields](#deprecated--dead-fields) — _optional_

## Connection

### Default cluster URL (Optional)

_optional · string_

The default cluster url for this data source.

| | |
|---|---|
| Example | `https://yourcluster.kusto.windows.net` |

## Authentication

### Authentication Method

_optional_

Choose the type of authentication to Azure services.

Discriminated-union object written by the `@grafana/azure-sdk` `AzureCredentialsForm` in `src/components/ConfigEditor/AzureCredentialsForm.tsx`. Shape depends on `authType`:

- `msi` — `{ authType: 'msi' }` (only rendered when Grafana has `azure.managedIdentityEnabled`).
- `workloadidentity` — `{ authType: 'workloadidentity' }` (only when `azure.workloadIdentityEnabled`).
- `currentuser` — `{ authType: 'currentuser' }` (only when `azure.userIdentityEnabled`).
- `clientsecret` — `{ authType: 'clientsecret', azureCloud, tenantId, clientId }` (secret in `secureJsonData.azureClientSecret`). Always available.
- `clientsecret-obo` — `{ authType: 'clientsecret-obo', azureCloud, tenantId, clientId }` (secret in `secureJsonData.azureClientSecret`; requires the `adxOnBehalfOf` feature toggle AND `jsonData.oauthPassThru = true`).

See `src/credentials/AzureCredentials.ts` in `grafana/grafana-azure-sdk-react@0.1.0` and the backend builder in `pkg/azuredx/adxauth/adxcredentials/builder.go`. Note that Azure Data Explorer's editor does **not** expose the `clientcertificate` or `ad-password` auth types offered by the shared SDK.

### Client Secret

_🔒 secret (write-only) · optional · string_

Client secret of the App Registration. Written write-only by `@grafana/azure-sdk`'s `AzureCredentialsForm`; check `secureJsonFields.azureClientSecret` on the read side. Used when `jsonData.azureCredentials.authType` is `clientsecret` or `clientsecret-obo`.

| | |
|---|---|
| Example | `XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX` |

### clientSecret

_🔒 secret (write-only) · optional · string_

Legacy client secret key. Preserved for backward compatibility with datasources provisioned before the credential migration to `azureClientSecret`. The backend reads it as a fallback in `pkg/azuredx/adxauth/adxcredentials/builder.go:57` (`getFromLegacy`) when no modern `jsonData.azureCredentials` is present.

### oauthPassThru

_optional · boolean_

Set to `true` by `@grafana/azure-sdk`'s `updateDatasourceCredentials` whenever the selected auth type is `clientsecret-obo` (On-Behalf-Of). The plugin backend enforces `oauthPassThru == true` for OBO in `pkg/azuredx/adxauth/adxcredentials/builder.go:89-98` (`ensureOnBehalfOfSupported`); creation fails otherwise.

## Query Optimizations

Various settings for controlling query behavior.

_This section is optional._

### Query timeout

_optional · string_

This value controls the client query timeout.

| | |
|---|---|
| Default | `30s` |
| Example | `30s` |

### Use dynamic caching

_optional · toggle_

By enabling this feature Grafana will dynamically apply cache settings on a per query basis and the default cache max age will be ignored. For time series queries we will use the bin size to widen the time range but also as cache max age.

| | |
|---|---|
| Default | `false` |

### Cache max age

_optional · string_

By default the cache is disabled. If you want to enable the query caching please specify a max timespan for the cache to live.

| | |
|---|---|
| Example | `0m` |

### Data consistency

_optional · select_

Query consistency controls how queries and updates are synchronized. Defaults to Strong. For more information see the Azure Data Explorer documentation.

| | |
|---|---|
| Default | `strongconsistency` |
| Allowed values | `strongconsistency` (Strong), `weakconsistency` (Weak) |

### Default editor mode

_optional · select_

This setting dictates which mode the editor will open in. Defaults to Visual.

| | |
|---|---|
| Default | `visual` |
| Allowed values | `visual` (Visual), `raw` (Raw) |

## Database schema settings

Configuration for the database schema including the default database.

_This section is optional._

### Default database

_optional · select_

### Use managed schema

_optional · toggle_

If enabled, allows tables, functions, and materialized views to be mapped to user friendly names.

| | |
|---|---|
| Default | `false` |

### Schema mappings

_optional · list_

| | |
|---|---|
| Shown when | **Use managed schema** is `true` |

Each item has the following fields:

## Application

_This section is optional._

### Application name (Optional)

_optional · string_

Application name to be displayed in ADX.

| | |
|---|---|
| Example | `Grafana-ADX` |

## Tracking

_This section is optional._

### Send username header to host

_optional · toggle_

With this feature enabled, Grafana will pass the logged in user's username in the `x-ms-user-id` header and in the `x-ms-client-request-id` header when sending requests to ADX. Can be useful when tracking needs to be done in ADX.

| | |
|---|---|
| Default | `false` |

## Additional settings

Additional settings are optional settings that can be configured for more control over your data source. This includes query optimizations, schema settings, tracking configuration, OpenAI configuration, request timeout, and forwarded cookies.

_This section is optional._

### Allowed cookies

_optional · list_

Grafana proxy deletes forwarded cookies by default. Specify cookies by name that should be forwarded to the data source.

| | |
|---|---|
| Example | `New cookie (hit enter to add)` |

## OpenAI (provisioning-only)

_This section is optional._

### OpenAIAPIKey

_🔒 secret (write-only) · optional · string_

OpenAI API key used by the plugin's `askOpenAI` resource endpoint (`pkg/azuredx/resource_handler.go:55-89`) to power the AI query assistant. There is no config editor UI to set it — the ADX ConfigEditor only inspects `secureJsonFields.OpenAIAPIKey` to decide whether the Additional settings section starts open (`src/components/ConfigEditor/index.tsx:58`). Populate it through provisioning YAML or via the datasource API.

## Legacy top-level credentials

_This section is optional._

### azureCloud

_optional · string_

Legacy Azure cloud discriminator written by the pre-2023 ADX config editor. Resolved by `pkg/azuredx/adxauth/adxcredentials/builder.go:107-121` `resolveLegacyCloudName`: `azuremonitor` → `AzureCloud`, `chinaazuremonitor` → `AzureChinaCloud`, `govazuremonitor` → `AzureUSGovernment`, empty → `AzureCloud`. The modern `AzureCredentialsForm` writes cloud inside `jsonData.azureCredentials.azureCloud` and clears this field.

### onBehalfOf

_optional · boolean_

Legacy On-Behalf-Of flag; when true and paired with legacy top-level `tenantId`/`clientId`/`secureJsonData.clientSecret`, the backend legacy fallback (`pkg/azuredx/adxauth/adxcredentials/builder.go:71-84`) constructs an `AzureClientSecretOboCredentials` instead of a plain `AzureClientSecretCredentials`. The modern form migrates this to `authType: clientsecret-obo`.

### tenantId

_optional · string_

Legacy top-level tenant ID (paired with legacy `clientId`/`secureJsonData.clientSecret`). Preserved so pre-migration datasources continue to authenticate; the current `AzureCredentialsForm` writes tenant inside `jsonData.azureCredentials.tenantId` and clears this field.

### clientId

_optional · string_

Legacy top-level client ID (paired with legacy `tenantId`/`secureJsonData.clientSecret`). Preserved so pre-migration datasources continue to authenticate; the current `AzureCredentialsForm` writes client inside `jsonData.azureCredentials.clientId` and clears this field.

## Deprecated / dead fields

_This section is optional._

### minimalCache

_optional · number_

Declared on `AdxDataSourceOptions` (`src/types/index.ts:115`) but never written by the config editor and never read by the plugin backend (`pkg/azuredx/models/settings.go` reads only `cacheMaxAge`). Dead field; preserved so provisioned datasources with the property still parse.

