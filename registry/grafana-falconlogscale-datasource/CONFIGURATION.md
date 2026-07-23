# Falcon LogScale configuration

How to configure the **Falcon LogScale** data source (`grafana-falconlogscale-datasource`) in Grafana.

For more information, see the [official documentation](https://github.com/grafana/falconlogscale-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)
- [Advanced settings](#advanced-settings) — _optional_
- [Additional settings](#additional-settings) — _optional_

## Connection

### URL

_**required** · string_

### Mode

_optional · select_

Select the data source mode. NGSIEM mode only supports OAuth2 client secret authentication.

| | |
|---|---|
| Default | `LogScale` |
| Allowed values | `LogScale`, `NGSIEM` |

## Authentication

### Authentication method

_optional · select_

| | |
|---|---|
| Default | `custom-token` |
| Allowed values | `custom-token` (LogScale Token Authentication), `custom-oauth-client-secret` (OAuth2 Client Credentials), `BasicAuth` (Basic authentication), `OAuthForward` (Forward OAuth Identity) |

### Token

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Example | `Token` |
| Shown when | **Authentication method** is **LogScale Token Authentication** (`custom-token`) |
| Required when | **authenticateWithToken** is `true` |

### Client ID

_conditionally required · string_

The OAuth2 client ID.

| | |
|---|---|
| Example | `Client ID` |
| Shown when | **Authentication method** is **OAuth2 Client Credentials** (`custom-oauth-client-secret`) |
| Required when | **oauth2** is `true` |

### Client Secret

_🔒 secret (write-only) · conditionally required · string_

The OAuth2 client secret.

| | |
|---|---|
| Example | `Client Secret` |
| Shown when | **Authentication method** is **OAuth2 Client Credentials** (`custom-oauth-client-secret`) |
| Required when | **oauth2** is `true` |

### User

_conditionally required · string_

| | |
|---|---|
| Example | `User` |
| Shown when | **Authentication method** is **Basic authentication** (`BasicAuth`) |
| Required when | **basicAuth** is `true` |

### Password

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Example | `Password` |
| Shown when | **Authentication method** is **Basic authentication** (`BasicAuth`) |
| Required when | **basicAuth** is `true` |

## Advanced settings

_This section is optional._

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

## Additional settings

Additional settings are optional settings that can be configured for more control over your data source. This includes the default repository or data links.

_This section is optional._

### Default Repository

_optional · select_

### Data links

_optional · list_

Add links to existing fields. Links will be shown in log row details next to the field value.

Each item has the following fields:

#### Field

_**required** · string_

Can be exact field name or a regex pattern that will match on the field name.

#### Label

_optional · string_

Use to provide a meaningful label to the data matched in the regex.

#### Regex

_**required** · string_

Use to parse and capture some part of the log message. You can use the captured groups in the template.

#### URL

_optional · string_

| | |
|---|---|
| Example | `http://example.com/${__value.raw}` |

#### Internal link data source

_optional · string_

UID of a Grafana data source. When set, the derived data link is treated as an internal link to that data source and the URL field is interpreted as a Query.

### Incremental querying (experimental)

_optional · toggle_

Results may be incomplete or incorrect in some cases. On auto-refresh, query new data and merge it with the cached result. This applies only to relative time ranges without aggregation functions.

| | |
|---|---|
| Default | `false` |

### Query overlap window

_optional · string_

Time window to re-fetch on each incremental query to catch late-arriving data (e.g. "10m", "30s", "1h"). Changes take effect after saving and reloading.

| | |
|---|---|
| Default | `10m` |
| Shown when | **Incremental querying (experimental)** is `true` |

## Other settings

### baseUrl

_optional · string_

Snapshot of the datasource URL written by the LogScale token authentication component (src/components/ConfigEditor/ConfigEditor.tsx:155). Never read by the backend, which always reads settings.URL (pkg/plugin/settings.go:38). Preserved for round-trip fidelity.

