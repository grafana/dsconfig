# Cloudflare configuration

How to configure the **Cloudflare** data source (`grafana-cloudflare-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-cloudflare-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Authentication](#authentication)

## Authentication

### API Key

_optional · select_

Cloudflare API Key. Provide relevant read-only permissions.

| | |
|---|---|
| Default | `bearer_token` |
| Allowed values | `bearer_token` (API Key) |

[Learn more](https://dash.cloudflare.com/profile/api-tokens)

### Token

_🔒 secret (write-only) · conditionally required · string_

Token for accessing the datasource API.

| | |
|---|---|
| Example | `Token value` |
| Shown when | **API Key** is **API Key** (`bearer_token`) |

[Learn more](https://dash.cloudflare.com/profile/api-tokens)

