# Databricks configuration

How to configure the **Databricks** data source (`grafana-databricks-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-databricks-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)
- [Additional settings](#additional-settings)

## Connection

### Host

_**required** · string_

| | |
|---|---|
| Example | `https://your-databricks-instance.com` |

### Http Path

_**required** · string_

| | |
|---|---|
| Example | `/sql/protocolv1/o/0/1234567890` |

## Authentication

### Authentication Type

_optional · select_

Authentication type of Databricks.

| | |
|---|---|
| Default | `Pat` |
| Allowed values | `Pat` (Personal Access Token), `OauthPT` (OAuth Passthrough), `OauthM2M` (OAuth M2M), `OauthOBO` (Azure (On-Behalf-Of)), `AzureM2M` (Azure Entra ID M2M) |

### Token

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Example | `XXXXXXXX` |
| Shown when | `jsonData_authType == 'Pat' || jsonData_authType == ''` |

### azureCredentials

_conditionally required_

Discriminated-union credentials object written by the `@grafana/azure-sdk` `AzureCredentialsForm` when authType is 'OauthOBO'. Shape: `{ authType: 'clientsecret-obo', azureCloud, tenantId, clientId }` (the client secret is stored write-only in secureJsonData.azureClientSecret). Parsed by the backend via `azcredentials.FromDatasourceData` (pkg/models/settings.go:121-139).

| | |
|---|---|
| Shown when | **Authentication Type** is **Azure (On-Behalf-Of)** (`OauthOBO`) |

### Client Secret

_🔒 secret (write-only) · conditionally required · string_

App Registration client secret for Azure On-Behalf-Of (authType 'OauthOBO'). Written write-only by `@grafana/azure-sdk`; check `secureJsonFields.azureClientSecret` on the read side.

| | |
|---|---|
| Example | `Client Secret` |
| Shown when | **Authentication Type** is **Azure (On-Behalf-Of)** (`OauthOBO`) |

### Client ID

_conditionally required · string_

| | |
|---|---|
| Example | `XXXXXXXX-XXXXXXXX-XXXX-XXXXXXXXXXXX` |
| Shown when | `jsonData_authType == 'OauthM2M' || jsonData_authType == 'AzureM2M'` |

### Client Secret

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Example | `XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX` |
| Shown when | `jsonData_authType == 'OauthM2M' || jsonData_authType == 'AzureM2M'` |

### Directory (tenant) ID

_conditionally required · string_

| | |
|---|---|
| Example | `XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX` |
| Shown when | **Authentication Type** is **Azure Entra ID M2M** (`AzureM2M`) |

### Azure Cloud

_optional · select_

| | |
|---|---|
| Default | `AzureCloud` |
| Allowed values | `AzureCloud` (Azure), `AzureChinaCloud` (Azure China), `AzureUSGovernment` (Azure US Government) |
| Shown when | **Authentication Type** is **Azure Entra ID M2M** (`AzureM2M`) |

## Additional settings

### Retries

_optional · string_

| | |
|---|---|
| Example | `5` |

### Pause

_optional · string_

| | |
|---|---|
| Example | `0` |

### Timeout

_optional · string_

| | |
|---|---|
| Example | `60` |

### Max Rows

_optional · string_

| | |
|---|---|
| Example | `10000` |

### Retry Timeout

_optional · string_

| | |
|---|---|
| Example | `40` |

### Debug

_optional · toggle_

### Unity Catalog Support

_optional · toggle_

Enable Unity Catalog support for 3-level namespace (catalog.schema.table).

### Default Query Format

_optional · select_

| | |
|---|---|
| Allowed values | `0` (Timeseries), `1` (Table) |

## Other settings

### oauthPassThru

_optional · boolean_

Set to true automatically when authType is 'OauthPT' (OAuth Passthrough) or 'OauthOBO' (Azure On-Behalf-Of) so Grafana forwards the caller's OAuth identity to Databricks. The backend hard-fails On-Behalf-Of auth if this is not true (pkg/models/settings.go:141-143, ErrInvalidOAuth). No editor UI — written as a side-effect of selecting the auth type.

### cloudFetch

_optional · boolean_

Enables Databricks CloudFetch (parallel result download) in the SQL connector. Not exposed in the configuration editor; the backend force-sets it to true on every load unless the `disableCloudFetch` Grafana feature toggle is enabled (pkg/models/settings.go:161-168).

| | |
|---|---|
| Default | `true` |

