# Netlify configuration

How to configure the **Netlify** data source (`grafana-netlify-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-netlify-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Authentication](#authentication)

## Authentication

### Personal access tokens

_optional · select_

Netlify REST API Key. Found here: https://app.netlify.com/user/applications#personal-access-tokens.

| | |
|---|---|
| Default | `bearer_token` |
| Allowed values | `bearer_token` (Personal access tokens) |

### Token

_🔒 secret (write-only) · conditionally required · string_

Token for accessing the datasource API.

| | |
|---|---|
| Example | `Token value` |
| Shown when | **Personal access tokens** is **Personal access tokens** (`bearer_token`) |

