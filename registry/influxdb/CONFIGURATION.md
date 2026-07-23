# InfluxDB configuration

How to configure the **InfluxDB** data source (`influxdb`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/influxdb/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [HTTP](#http)
- [Auth](#auth)
- [TLS/SSL Auth Details](#tlsssl-auth-details) — _optional_
- [Query language](#query-language)
- [InfluxDB Details (InfluxQL)](#influxdb-details-influxql) — _optional_
- [InfluxDB Details (Flux)](#influxdb-details-flux) — _optional_
- [InfluxDB Details (SQL)](#influxdb-details-sql) — _optional_
- [Other settings](#other-settings)

## HTTP

### URL

_**required** · string_

| | |
|---|---|
| Example | `http://localhost:8086` |

### Allowed cookies

_optional · list_

Grafana proxy deletes forwarded cookies by default. Specify cookies by name that should be forwarded to the data source.

| | |
|---|---|
| Example | `New cookie (hit enter to add)` |

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

## Query language

### Query language

_optional · select_

| | |
|---|---|
| Default | `InfluxQL` |
| Allowed values | `InfluxQL`, `SQL`, `Flux` |

### Product

_optional · select_

Use InfluxDB detection to identify the product.

| | |
|---|---|
| Allowed values | `InfluxDB Cloud Dedicated`, `InfluxDB Cloud Serverless`, `InfluxDB Clustered`, `InfluxDB Enterprise 1.x`, `InfluxDB Enterprise 3.x`, `InfluxDB Cloud (TSM)`, `InfluxDB Cloud 1`, `InfluxDB OSS 1.x`, `InfluxDB OSS 2.x`, `InfluxDB OSS 3.x` |

## InfluxDB Details (InfluxQL)

_This section is optional._

### Database

_conditionally required · string_

| | |
|---|---|
| Example | `mydb` |
| Shown when | `jsonData_version == 'InfluxQL' || jsonData_version == 'SQL'` |

### User

_optional · string_

Legacy root-level user field written by the v1 InfluxQL editor. Distinct from root.basicAuthUser — the v1 editor writes options.user (SDK root User field) while the v2 editor writes options.basicAuthUser (SDK root BasicAuthUser field). Not consumed by the current backend or the SDK HTTPClientOptions auth handler; effectively a display-only echo unless the operator also enables root.basicAuth so the SDK reads basicAuthUser instead.

| | |
|---|---|
| Example | `myuser` |
| Shown when | **Query language** is `InfluxQL` |

### Password

_🔒 secret (write-only) · optional · string_

Legacy secure password paired with root.user for the v1 InfluxQL editor's User + Password inputs. Distinct from secureJsonData.basicAuthPassword. Not consumed by the current backend or SDK HTTP auth path.

| | |
|---|---|
| Example | `Password` |
| Shown when | **Query language** is `InfluxQL` |

### HTTP Method

_optional · select_

You can use either GET or POST HTTP method to query your InfluxDB database. The POST method allows you to perform heavy requests (with a lots of WHERE clause) while the GET method will restrict you and return an error if the query is too large.

| | |
|---|---|
| Default | `GET` |
| Allowed values | `GET`, `POST` |
| Shown when | **Query language** is `InfluxQL` |

### Min time interval

_optional · string_

A lower limit for the auto group by time interval. Recommended to be set to write frequency, for example 1m if your data is written every minute.

| | |
|---|---|
| Example | `10s` |
| Shown when | `jsonData_version == 'InfluxQL' || jsonData_version == 'Flux'` |

### Autocomplete range

_optional · string_

This time range is used in the query editor's autocomplete to reduce the execution time of tag filter queries.

| | |
|---|---|
| Example | `12h` |
| Shown when | **Query language** is `InfluxQL` |

## InfluxDB Details (Flux)

_This section is optional._

### Organization

_conditionally required · string_

| | |
|---|---|
| Example | `myorg` |
| Shown when | **Query language** is `Flux` |

### Token

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Example | `Token` |
| Shown when | `jsonData_version == 'Flux' || jsonData_version == 'SQL'` |

### Default Bucket

_conditionally required · string_

| | |
|---|---|
| Example | `default bucket` |
| Shown when | **Query language** is `Flux` |

## InfluxDB Details (SQL)

_This section is optional._

### Insecure Connection

_optional · toggle_

Disable TLS for the FlightSQL gRPC connection used by the SQL query path.

| | |
|---|---|
| Default | `false` |
| Shown when | **Query language** is `SQL` |

## Other settings

### Max series

_optional · number_

Limit the number of series/tables that Grafana will process. Lower this number to prevent abuse, and increase it if you have lots of small time series and not all are shown. Defaults to 1000.

| | |
|---|---|
| Default | `1000` |
| Example | `1000` |

### pdcInjected

_optional · boolean_

Backend-controlled indicator that a Private Datasource Connect (PDC) proxy has been injected into this datasource's HTTP transport. Not editor-writable; read by the v2 LeftSideBar to render PDC-specific section headers (LeftSideBar.tsx:12).

| | |
|---|---|
| Default | `false` |

