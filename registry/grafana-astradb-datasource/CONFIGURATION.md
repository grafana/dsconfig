# AstraDB configuration

How to configure the **AstraDB** data source (`grafana-astradb-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-astradb-datasource/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)

## Connection

### URI

_conditionally required · string_

| | |
|---|---|
| Example | `$ASTRA_CLUSTER_ID-$ASTRA_REGION.apps.astra.datastax.com:443` |
| Shown when | **Authentication** is **Token** (`0`) |

### GRPC Endpoint

_conditionally required · string_

| | |
|---|---|
| Example | `localhost:8090` |
| Shown when | **Authentication** is **Credentials** (`1`) |

### Auth Endpoint

_conditionally required · string_

| | |
|---|---|
| Example | `localhost:8081` |
| Shown when | **Authentication** is **Credentials** (`1`) |

### Secure

_optional · toggle_

| | |
|---|---|
| Default | `false` |
| Shown when | **Authentication** is **Credentials** (`1`) |

## Authentication

### Authentication

_optional · radio_

| | |
|---|---|
| Default | `0` |
| Allowed values | `0` (Token), `1` (Credentials) |

### Token

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Example | `AstraCS:xxxxx` |
| Shown when | **Authentication** is **Token** (`0`) |

### User Name

_conditionally required · string_

| | |
|---|---|
| Example | `localhost:8090` |
| Shown when | **Authentication** is **Credentials** (`1`) |

### Password

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Example | `xxxxx` |
| Shown when | **Authentication** is **Credentials** (`1`) |

