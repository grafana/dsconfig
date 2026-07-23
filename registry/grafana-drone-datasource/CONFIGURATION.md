# Drone configuration

How to configure the **Drone** data source (`grafana-drone-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-drone-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)

## Connection

Provide information to connect to the data source

### URL

_**required** · string_

The URL of the Drone server, including `https://` and without trailing `/`.

## Authentication

### Drone API token

_optional · select_

You can find your API token under <YOUR_DRONE_URL>/account.

| | |
|---|---|
| Default | `auth_bearer` |
| Allowed values | `auth_bearer` (Drone API token) |

### Token

_🔒 secret (write-only) · conditionally required · string_

Token for accessing the datasource API.

| | |
|---|---|
| Example | `Token value` |
| Shown when | **Drone API token** is **Drone API token** (`auth_bearer`) |

