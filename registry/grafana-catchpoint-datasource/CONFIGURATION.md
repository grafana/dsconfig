# Catchpoint configuration

How to configure the **Catchpoint** data source (`grafana-catchpoint-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-catchpoint-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Authentication](#authentication)

## Authentication

### API v2 Key

_optional · select_

Catchpoint REST API v2 Key.

| | |
|---|---|
| Default | `bearer_token` |
| Allowed values | `bearer_token` (API v2 Key) |

### Token

_🔒 secret (write-only) · conditionally required · string_

Token for accessing the datasource API.

| | |
|---|---|
| Example | `Token value` |
| Shown when | **API v2 Key** is **API v2 Key** (`bearer_token`) |

