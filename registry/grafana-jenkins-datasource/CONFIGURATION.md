# Jenkins configuration

How to configure the **Jenkins** data source (`grafana-jenkins-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-jenkins-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)

## Connection

### URL

_**required** · string_

Jenkins URL, e.g. https://jenkins.example.com.

| | |
|---|---|
| Example | `Jenkins URL, e.g. https://jenkins.example.com` |

## Authentication

### User

_optional · string_

The username to use for authentication.

| | |
|---|---|
| Example | `Username` |

### Password

_🔒 secret (write-only) · optional · string_

The password to use for authentication.

| | |
|---|---|
| Example | `Password` |

