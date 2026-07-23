# Looker configuration

How to configure the **Looker** data source (`grafana-looker-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-looker-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)

## Connection

### Looker URL

_**required** · string_

Looker base URL. Example: https://00001234-1234-1ab2-1234-a1b2c3d4.looker.app.

| | |
|---|---|
| Example | `https://xxxxx.looker.app` |

## Authentication

### Authentication type

_optional · radio_

Looker authentication type.

| | |
|---|---|
| Default | `client_secret` |
| Allowed values | `client_secret` (Client Secret) |

### Looker Client ID

_conditionally required · string_

API credentials Looker client id.

| | |
|---|---|
| Example | `Client ID` |
| Shown when | `jsonData_authType == 'client_secret' || jsonData_authType == ''` |
| Required when | **Authentication type** is **Client Secret** (`client_secret`) |

### Looker Client Secret

_🔒 secret (write-only) · conditionally required · string_

API credentials Looker client secret.

| | |
|---|---|
| Example | `Looker Client secret` |
| Shown when | `jsonData_authType == 'client_secret' || jsonData_authType == ''` |
| Required when | **Authentication type** is **Client Secret** (`client_secret`) |

