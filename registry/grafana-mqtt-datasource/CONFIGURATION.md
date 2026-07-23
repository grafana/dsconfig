# MQTT configuration

How to configure the **MQTT** data source (`grafana-mqtt-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-mqtt-datasource/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)
- [TLS Configuration](#tls-configuration) — _optional_

## Connection

### URI

_**required** · string_

| | |
|---|---|
| Example | `TCP (tcp://), TLS (tls://), or WebSocket (ws://)` |

### Client ID

_optional · string_

If not set, a random client ID is used.

## Authentication

### Username

_optional · string_

| | |
|---|---|
| Example | `Username` |

### Password

_🔒 secret (write-only) · optional · string_

| | |
|---|---|
| Example | `Password` |

### Use TLS Client Auth

_optional · toggle_

Enables TLS authentication using client cert configured in secure json data.

| | |
|---|---|
| Default | `false` |

### Skip TLS Verification

_optional · toggle_

When enabled, skips verification of the MQTT server's TLS certificate chain and host name.

| | |
|---|---|
| Default | `false` |

### With CA Cert

_optional · toggle_

Needed for verifying servers with self-signed TLS Certs.

| | |
|---|---|
| Default | `false` |

## TLS Configuration

_This section is optional._

### TLS CA Certificate

_🔒 secret (write-only) · optional · multiline text_

If a Certificate Authority certificate is required to verify the server's certificate, provide it here.

| | |
|---|---|
| Example | `-----BEGIN CERTIFICATE-----` |
| Shown when | **With CA Cert** is `true` |

### TLS Client Certificate

_🔒 secret (write-only) · optional · multiline text_

To authenticate with an TLS client certificate, provide the client certificate here.

| | |
|---|---|
| Example | `-----BEGIN CERTIFICATE-----` |
| Shown when | **Use TLS Client Auth** is `true` |

### TLS Client Key

_🔒 secret (write-only) · optional · multiline text_

To authenticate with a client TLS certificate, provide the private key here.

| | |
|---|---|
| Example | `-----BEGIN RSA PRIVATE KEY-----` |
| Shown when | **Use TLS Client Auth** is `true` |

