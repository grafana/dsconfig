# Graphite configuration

How to configure the **Graphite** data source (`graphite`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/graphite/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)
- [Network & TLS](#network--tls) — _optional_
- [Graphite settings](#graphite-settings)
- [Query settings](#query-settings) — _optional_
- [Advanced settings](#advanced-settings) — _optional_

## Connection

How to reach the Graphite server.

### URL

_**required** · string_

Specify a complete HTTP URL (for example http://your_server:8080).

| | |
|---|---|
| Example | `http://localhost:8080` |

## Authentication

How to authenticate requests to the Graphite server.

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

### Forward OAuth Identity

_optional · toggle_

Forward the user's upstream OAuth identity to the data source (Their access token gets passed along).

| | |
|---|---|
| Default | `false` |

## Network & TLS

TLS client authentication and certificate verification for the connection.

_This section is optional._

### TLS Client Auth

_optional · toggle_

| | |
|---|---|
| Default | `false` |

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

### With CA Cert

_optional · toggle_

Needed for verifying self-signed TLS Certs.

| | |
|---|---|
| Default | `false` |

### CA Cert

_🔒 secret (write-only) · conditionally required · multiline text_

| | |
|---|---|
| Example | `Begins with -----BEGIN CERTIFICATE-----` |
| Shown when | **With CA Cert** is `true` |

### Skip TLS Verify

_optional · toggle_

| | |
|---|---|
| Default | `false` |

## Graphite settings

Graphite backend type and version.

### Version

_optional · select_

This option controls what functions are available in the Graphite query editor.

| | |
|---|---|
| Default | `1.1` |
| Allowed values | `0.9` (0.9.x), `1.0` (1.0.x), `1.1` (1.1.x) |

### Graphite backend type

_optional · select_

There are different types of Graphite compatible backends. Here you can specify the type you are using. For Metrictank, this will enable specific features, like query processing meta data. Metrictank
        is a multi-tenant timeseries engine for Graphite and friends.

| | |
|---|---|
| Allowed values | `default` (Default), `metrictank` (Metrictank) |

### Rollup indicator

_optional · toggle_

Shows up as an info icon in panel headers when data is aggregated.

| | |
|---|---|
| Default | `false` |
| Shown when | **Graphite backend type** is **Metrictank** (`metrictank`) |

## Query settings

Query-time behaviour and cross-datasource query mappings.

_This section is optional._

### Label mappings

_optional · object_

Mappings are currently supported only between Graphite and Loki queries.

When you switch your data source from Graphite to Loki, your queries are mapped according to the mappings defined in the example below. To define a mapping, write the full path of the metric and replace nodes you want to map to label with the label name in parentheses. The value of the label is extracted from your Graphite query when you switch data sources.

All tags are automatically mapped to labels regardless of the mapping configuration. Graphite matching patterns (using `{}`) are converted to Loki's regular expressions matching patterns. When you use functions in your queries, the metrics, and tags are extracted to match them with defined mappings.

Example: for a mapping = `servers.(cluster).(server).*`:

| Graphite query | Mapped to Loki query |
| --- | --- |
| `alias(servers.west.001.cpu,1,2)` | `{cluster="west", server="001"}` |
| `alias(servers.*.{001,002}.*,1,2)` | `{server=~"(001|002)"}` |
| `interpolate(seriesByTag('foo=bar', 'server=002'), inf))` | `{foo="bar", server="002"}` |

Storage shape: `importConfiguration.loki.mappings` is an array of objects `{ matchers: Array<{ value: string, labelName?: string }> }`. The editor stringifies a full metric path such as `servers.(cluster).(server).*` into a list of matchers where segments wrapped in parentheses become `{value: '*', labelName: '<name>'}` and other segments become `{value: '<segment>'}`. See `src/configuration/parseLokiLabelMappings.ts` and `src/types.ts:56-71`.

## Advanced settings

Additional HTTP transport settings.

_This section is optional._

### Timeout

_optional · number_

HTTP request timeout in seconds.

| | |
|---|---|
| Example | `Timeout in seconds` |

### Allowed cookies

_optional · list_

Grafana proxy deletes forwarded cookies by default. Specify cookies by name that should be forwarded to the data source.

### Custom HTTP Headers

_optional · list_

Additional HTTP headers sent with every request. Header names are stored in jsonData (httpHeaderName<N>); header values are write-only in secureJsonData (httpHeaderValue<N>). Rendered by the @grafana/ui CustomHeadersSettings component (DataSourceHttpSettings.tsx:352) and forwarded to Graphite by the SDK's HTTPClientOptions.

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

