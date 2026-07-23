# OpenTSDB configuration

How to configure the **OpenTSDB** data source (`opentsdb`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/opentsdb/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [HTTP](#http)
- [Auth](#auth)
- [TLS/SSL Auth Details](#tlsssl-auth-details) — _optional_
- [OpenTSDB settings](#opentsdb-settings)

## HTTP

### URL

_**required** · string_

Specify a complete HTTP URL (for example http://your_server:8080).

| | |
|---|---|
| Example | `http://localhost:4242` |

### Allowed cookies

_optional · list_

Grafana proxy deletes forwarded cookies by default. Specify cookies by name that should be forwarded to the data source.

### Timeout

_optional · number_

HTTP request timeout in seconds.

| | |
|---|---|
| Example | `Timeout in seconds` |

### Custom HTTP Headers

_optional · list_

Additional HTTP headers sent with every request. Header names are stored in jsonData (httpHeaderName<N>); header values are write-only in secureJsonData (httpHeaderValue<N>).

Each item has the following fields:

#### Header

_**required** · string_

| | |
|---|---|
| Example | `X-Custom-Header` |
| Must match | `^[A-Za-z][A-Za-z0-9-]*$` |

#### Value

_optional · string_

| | |
|---|---|
| Example | `Header Value` |

## Auth

### Basic auth

_optional · toggle_

| | |
|---|---|
| Default | `false` |

### With Credentials

_optional · toggle_

Whether credentials such as cookies or auth headers should be sent with cross-site requests.

| | |
|---|---|
| Default | `false` |

### TLS Client Auth

_optional · toggle_

| | |
|---|---|
| Default | `false` |

### With CA Cert

_optional · toggle_

Needed for verifying self-signed TLS Certs.

| | |
|---|---|
| Default | `false` |

### Skip TLS Verify

_optional · toggle_

| | |
|---|---|
| Default | `false` |

### Forward OAuth Identity

_optional · toggle_

Forward the user's upstream OAuth identity to the data source (Their access token gets passed along).

| | |
|---|---|
| Default | `false` |

### User

_conditionally required · string_

| | |
|---|---|
| Example | `user` |
| Shown when | **Basic auth** is `true` |

### Password

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Example | `Password` |
| Shown when | **Basic auth** is `true` |

## TLS/SSL Auth Details

_This section is optional._

### ServerName

_conditionally required · string_

| | |
|---|---|
| Example | `domain.example.com` |
| Shown when | **TLS Client Auth** is `true` |

### CA Cert

_🔒 secret (write-only) · conditionally required · multiline text_

| | |
|---|---|
| Example | `Begins with -----BEGIN CERTIFICATE-----` |
| Shown when | **With CA Cert** is `true` |

### Client Cert

_🔒 secret (write-only) · conditionally required · multiline text_

| | |
|---|---|
| Example | `Begins with -----BEGIN CERTIFICATE-----` |
| Shown when | **TLS Client Auth** is `true` |

### Client Key

_🔒 secret (write-only) · conditionally required · multiline text_

| | |
|---|---|
| Example | `Begins with -----BEGIN RSA PRIVATE KEY-----` |
| Shown when | **TLS Client Auth** is `true` |

## OpenTSDB settings

### Version

_optional · select_

| | |
|---|---|
| Default | `1` |
| Allowed values | `1` (<=2.1), `2` (==2.2), `3` (==2.3), `4` (==2.4) |

### Resolution

_optional · select_

| | |
|---|---|
| Default | `1` |
| Allowed values | `1` (second), `2` (millisecond) |

### Lookup limit

_optional · number_

| | |
|---|---|
| Default | `1000` |
| Range | at least 0 |

