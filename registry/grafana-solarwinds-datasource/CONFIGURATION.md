# Solarwinds configuration

How to configure the **Solarwinds** data source (`grafana-solarwinds-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-solarwinds-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)
- [TLS Settings](#tls-settings) — _optional_

## Connection

Provide information to connect to the data source

### URL

_**required** · string_

The URL of your SolarWinds instance, including `https://` and without trailing `/`.

## Authentication

### Basic Auth

_optional · select_

| | |
|---|---|
| Default | `basic_auth` |
| Allowed values | `basic_auth` (Basic Auth) |

### Username

_conditionally required · string_

SolarWinds Username.

| | |
|---|---|
| Example | `Username` |
| Shown when | **Basic Auth** is **Basic Auth** (`basic_auth`) |

### Password

_🔒 secret (write-only) · conditionally required · string_

SolarWinds Password.

| | |
|---|---|
| Example | `Password` |
| Shown when | **Basic Auth** is **Basic Auth** (`basic_auth`) |

## TLS Settings

_This section is optional._

### Add self-signed certificate

_optional · toggle_

### CA Certificate

_🔒 secret (write-only) · optional · multiline text_

Your self-signed certificate.

| | |
|---|---|
| Shown when | **Add self-signed certificate** is `true` |

### TLS Client Authentication

_optional · toggle_

Validate using TLS client authentication, in which the server authenticates the client.

### ServerName

_optional · string_

A Servername is used to verify the hostname on the returned certificate.

| | |
|---|---|
| Shown when | **TLS Client Authentication** is `true` |

### Client Certificate

_🔒 secret (write-only) · optional · multiline text_

The client certificate can be generated from a Certificate Authority or be self-signed.

| | |
|---|---|
| Shown when | **TLS Client Authentication** is `true` |

### Client Key

_🔒 secret (write-only) · optional · multiline text_

The client key can be generated from a Certificate Authority or be self-signed.

| | |
|---|---|
| Shown when | **TLS Client Authentication** is `true` |

### Skip TLS certificate validation

_optional · toggle_

