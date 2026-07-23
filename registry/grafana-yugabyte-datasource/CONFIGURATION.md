# Yugabyte configuration

How to configure the **Yugabyte** data source (`grafana-yugabyte-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/yugabyte/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)

## Connection

### Host URL

_**required** · string_

| | |
|---|---|
| Example | `localhost:5433` |

### Database

_**required** · string_

| | |
|---|---|
| Example | `yb_demo` |

## Authentication

### Username

_**required** · string_

| | |
|---|---|
| Example | `yugabyte` |

### Password

_🔒 secret (write-only) · optional · string_

| | |
|---|---|
| Example | `********` |

