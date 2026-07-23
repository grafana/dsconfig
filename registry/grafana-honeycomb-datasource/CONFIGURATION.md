# Honeycomb configuration

How to configure the **Honeycomb** data source (`grafana-honeycomb-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-honeycomb-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Access](#access)
- [Environment](#environment)
- [Advanced Settings](#advanced-settings) — _optional_

## Access

### Honeycomb API Key

_🔒 secret (write-only) · **required** · string_

| | |
|---|---|
| Example | `Honeycomb API Key` |

**Team API Key**

To create or retrieve new Team API Key, navigate to [Account & Profile](https://ui.honeycomb.io/account)

Ensure key has the following permissions:

- Manage Queries and Columns
- Run Queries

## Environment

### URL

_**required** · string_

Customize the api URL. By default this will be https://api.honeycomb.io.

| | |
|---|---|
| Default | `https://api.honeycomb.io` |
| Example | `https://api.honeycomb.io` |
| Must match | `^https://` |

### Team Name

_**required** · string_

Specify the team name. This will be useful in data links.

### Environment Name

_optional · string_

Optional. Specify the environment name. This will be useful in data links.

## Advanced Settings

_This section is optional._

### Time Window (days)

_optional · number_

Optional. The maximum time window, in days. Queries will only return data from the last N days, where N is the retention limit. Default is 7 days, since that is the maximum retention limit normally supported by the Honeycomb API. Honeycomb API docs: https://api-docs.honeycomb.io/api/query-data.

| | |
|---|---|
| Default | `7` |
| Example | `7` |

