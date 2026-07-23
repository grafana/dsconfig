# Vercel configuration

How to configure the **Vercel** data source (`grafana-vercel-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-vercel-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)

## Connection

Provide information to connect to the data source

### Team ID

_optional · string_

The ID of the Vercel team to query.

| | |
|---|---|
| Example | `eg: team_1a2b3c4d5e6f7g8h9i0j1k2l` |

## Authentication

### Access Token

_optional · select_

Vercel Access Tokens are required to authenticate and use the Vercel API. Tokens are either scoped to your full account or a specific team. If a token is scoped to a team, you must also provide a team ID that matches the scope of the token.

| | |
|---|---|
| Default | `vercelApiKey` |
| Allowed values | `vercelApiKey` (Access Token) |

### Token

_🔒 secret (write-only) · conditionally required · string_

Token for accessing the datasource API.

| | |
|---|---|
| Example | `Token value` |
| Shown when | **Access Token** is **Access Token** (`vercelApiKey`) |

[Learn more](https://vercel.com/docs/rest-api#creating-an-access-token)

