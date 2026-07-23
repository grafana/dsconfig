# Adobe Analytics configuration

How to configure the **Adobe Analytics** data source (`grafana-adobeanalytics-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-adobeanalytics-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)

## Connection

Provide information to connect to the data source

### Global Company ID

_**required** · string_

Refer to [plugin documentation](http://grafana.com/docs/plugins/grafana-adobeanalytics-datasource/#connection) for more information on where to find your Global Company ID in Adobe portal.

## Authentication

### OAuth server to server authentication

_optional · select_

Authorization flow where application credentials are exchanged for an access token.

| | |
|---|---|
| Default | `oauth2_m2m` |
| Allowed values | `oauth2_m2m` (OAuth server to server authentication) |

### Client ID

_conditionally required · string_

| | |
|---|---|
| Shown when | **OAuth server to server authentication** is **OAuth server to server authentication** (`oauth2_m2m`) |

### Client Secret

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Shown when | **OAuth server to server authentication** is **OAuth server to server authentication** (`oauth2_m2m`) |

