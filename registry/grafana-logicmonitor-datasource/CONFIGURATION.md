# LogicMonitor Devices configuration

How to configure the **LogicMonitor Devices** data source (`grafana-logicmonitor-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-logicmonitor-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)

## Connection

Provide information to connect to the data source

### Account Name

_**required** · string_

Your LogicMonitor account name. Example: Use foo for the logic monitor URL https://foo.logicmonitor.com/`.

## Authentication

### API v3 Key

_optional · select_

Bearer token for LogicMonitor REST API v3.

| | |
|---|---|
| Default | `auth_bearer` |
| Allowed values | `auth_bearer` (API v3 Key) |

### Token

_🔒 secret (write-only) · conditionally required · string_

Token for accessing the datasource API.

| | |
|---|---|
| Example | `Token value` |
| Shown when | **API v3 Key** is **API v3 Key** (`auth_bearer`) |

