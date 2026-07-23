# Azure DevOps configuration

How to configure the **Azure DevOps** data source (`grafana-azuredevops-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-azuredevops-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Azure DevOps settings](#azure-devops-settings)
- [Optional Configuration](#optional-configuration) — _optional_

## Azure DevOps settings

### URL

_**required** · string_

Azure DevOps instance URL.

| | |
|---|---|
| Example | `https://dev.azure.com/XXXX` |

### PAT

_🔒 secret (write-only) · **required** · string_

Azure DevOps personal access token.

| | |
|---|---|
| Example | `Azure DevOps PAT` |

## Optional Configuration

_This section is optional._

### Projects limit

_optional · number_

Number of items to retrieve in projects list query.

| | |
|---|---|
| Default | `100` |
| Example | `100` |

### Username

_optional · string_

Username of the user that owns the Azure DevOps PAT. May be needed for some versions of Azure DevOps Server.

| | |
|---|---|
| Example | `ado` |

