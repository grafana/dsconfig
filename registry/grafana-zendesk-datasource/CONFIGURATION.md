# Zendesk configuration

How to configure the **Zendesk** data source (`grafana-zendesk-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-zendesk-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)

## Connection

Provide information to connect to the data source

### Subdomain

_**required** · string_

If your Zendesk URL is "grafana.zendesk.com", the subdomain would be "grafana".

## Authentication

### Authentication method

_optional · select_

Identifier of the selected authentication method for the Tickets service. The Zendesk API server exposes a single method, `basic_auth`; the backend defaults to it when unset.

| | |
|---|---|
| Default | `basic_auth` |
| Allowed values | `basic_auth` (Basic Auth) |

### Email

_conditionally required · string_

Email address used to login to Zendesk.

| | |
|---|---|
| Example | `Email` |
| Shown when | **Authentication method** is **Basic Auth** (`basic_auth`) |

### API Token

_🔒 secret (write-only) · conditionally required · string_

API Token generated from Zendesk. Visit the [docs](https://support.zendesk.com/hc/en-us/articles/4408889192858-Managing-access-to-the-Zendesk-API) to learn how.

| | |
|---|---|
| Example | `API Token` |
| Shown when | **Authentication method** is **Basic Auth** (`basic_auth`) |

