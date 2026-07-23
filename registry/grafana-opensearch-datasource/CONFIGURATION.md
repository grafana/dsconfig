# OpenSearch configuration

How to configure the **OpenSearch** data source (`grafana-opensearch-datasource`) in Grafana.

For more information, see the [official documentation](https://github.com/grafana/opensearch-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [HTTP](#http)
- [Additional HTTP settings](#additional-http-settings) — _optional_
- [Auth](#auth)
- [TLS/SSL Auth Details](#tlsssl-auth-details) — _optional_
- [OpenSearch details](#opensearch-details)
- [Logs](#logs) — _optional_
- [Data links](#data-links) — _optional_

## HTTP

### URL

_**required** · string_

Specify a complete HTTP URL (for example http://your_server:8080).

| | |
|---|---|
| Example | `http://localhost:9200` |

### Access

_optional · select_

| | |
|---|---|
| Default | `proxy` |
| Allowed values | `proxy` (Server (default)), `direct` (Browser) |

## Additional HTTP settings

_This section is optional._

### Allowed cookies

_optional · list_

Grafana proxy deletes forwarded cookies by default. Specify cookies by name that should be forwarded to the data source.

| | |
|---|---|
| Shown when | **Access** is **Server (default)** (`proxy`) |

### Timeout

_optional · number_

HTTP request timeout in seconds.

| | |
|---|---|
| Example | `Timeout in seconds` |
| Shown when | **Access** is **Server (default)** (`proxy`) |

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

### With Credentials

_optional · toggle_

Whether credentials such as cookies or auth headers should be sent with cross-site requests.

| | |
|---|---|
| Default | `false` |

### SigV4 auth

_optional · toggle_

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
| Shown when | **Access** is **Server (default)** (`proxy`) |

## TLS/SSL Auth Details

_This section is optional._

### ServerName

_conditionally required · string_

| | |
|---|---|
| Example | `domain.example.com` |
| Shown when | **TLS Client Auth** is `true` |

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

### CA Cert

_🔒 secret (write-only) · conditionally required · multiline text_

| | |
|---|---|
| Example | `Begins with -----BEGIN CERTIFICATE-----` |
| Shown when | **With CA Cert** is `true` |

## OpenSearch details

### Index name

_**required** · string_

| | |
|---|---|
| Example | `es-index-name` |

### Pattern

_optional · select_

| | |
|---|---|
| Allowed values | `` (No pattern), `Hourly`, `Daily`, `Weekly`, `Monthly`, `Yearly` |

### Time field name

_**required** · string_

| | |
|---|---|
| Default | `@timestamp` |

### Serverless

_optional · toggle_

If this is a DataSource to query a serverless OpenSearch service.

| | |
|---|---|
| Default | `false` |

### Version

_**required** · string_

| | |
|---|---|
| Example | `version required` |
| Shown when | `jsonData_serverless != true` |

### Max concurrent Shard Requests

_optional · number_

| | |
|---|---|
| Shown when | `jsonData_serverless != true` |

### Min time interval

_optional · string_

A lower limit for the auto group by time interval. Recommended to be set to write frequency, for example 1m if your data is written every minute.

| | |
|---|---|
| Example | `10s` |
| Must match | `^\d+(ms|[Mwdhmsy])$` |

### PPL enabled

_optional · toggle_

Allow Piped Processing Language as an alternative query syntax in the OpenSearch query editor.

| | |
|---|---|
| Default | `true` |

## Logs

_This section is optional._

### Message field name

_optional · string_

| | |
|---|---|
| Example | `_source` |

### Level field name

_optional · string_

## Data links

_This section is optional._

### Data links

_optional · list_

Add links to existing fields. Links will be shown in log row details next to the field value.

Each item has the following fields:

#### Field

_**required** · string_

Can be exact field name or a regex pattern that will match on the field name.

#### Title

_optional · string_

#### URL

_optional · string_

| | |
|---|---|
| Example | `http://example.com/${__value.raw}` |

#### Internal link

_optional · string_

