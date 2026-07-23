# ClickHouse configuration

How to configure the **ClickHouse** data source (`grafana-clickhouse-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-clickhouse-datasource/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Server](#server)
- [TLS / SSL Settings](#tls--ssl-settings)
- [Credentials](#credentials)
- [Configuration Mode](#configuration-mode)
- [Default DB and table](#default-db-and-table) — _optional_
- [Query settings](#query-settings) — _optional_
- [Additional settings](#additional-settings) — _optional_

## Server

### Server address

_**required** · string_

Enter server address.

| | |
|---|---|
| Example | `Server address` |

### Server port

_**required** · number_

ClickHouse server port.

| | |
|---|---|
| Default | `9000` |

ClickHouse supports two server protocols: Native TCP and HTTP. Both protocols can be secured with TLS. <br/>[Native TCP](https://clickhouse.com/docs/interfaces/tcp) is the default and recommended option.<br/>[HTTP](https://clickhouse.com/docs/interfaces/http) is for servers configured to accept HTTP connections.

### Protocol

_optional · radio_

Native or HTTP for server protocol.

| | |
|---|---|
| Default | `native` |
| Allowed values | `native` (Native), `http` (HTTP) |

### Secure Connection

_optional · toggle_

Toggle on if the connection is secure.

| | |
|---|---|
| Default | `false` |

### HTTP URL Path

_optional · string_

Additional URL path for HTTP requests.

| | |
|---|---|
| Example | `additional-path` |
| Shown when | **Protocol** is **HTTP** (`http`) |

### Custom HTTP Headers

_optional · list_

Add Custom HTTP headers when querying the database.

| | |
|---|---|
| Shown when | **Protocol** is **HTTP** (`http`) |

Each item has the following fields:

#### Header Name

_optional · string_

| | |
|---|---|
| Example | `X-Custom-Header` |

#### Header Value

_optional · string_

Empty when the header is stored securely; the encrypted value lives in secureJsonData under secureHttpHeaders.<Header Name>.

| | |
|---|---|
| Example | `Header Value` |

#### Secure

_optional · toggle_

### Forward Grafana HTTP Headers

_optional · toggle_

Forward Grafana HTTP Headers to datasource.

| | |
|---|---|
| Default | `false` |
| Shown when | **Protocol** is **HTTP** (`http`) |

## TLS / SSL Settings

### Skip TLS Verify

_optional · toggle_

Skip TLS Verify.

| | |
|---|---|
| Default | `false` |

### TLS Client Auth

_optional · toggle_

TLS Client Auth.

| | |
|---|---|
| Default | `false` |

### With CA Cert

_optional · toggle_

Needed for verifying self-signed TLS Certs.

| | |
|---|---|
| Default | `false` |

### CA Cert

_🔒 secret (write-only) · optional · multiline text_

| | |
|---|---|
| Example | `CA Cert. Begins with -----BEGIN CERTIFICATE-----` |
| Shown when | **With CA Cert** is `true` |

### Client Cert

_🔒 secret (write-only) · optional · multiline text_

| | |
|---|---|
| Example | `Client Cert. Begins with -----BEGIN CERTIFICATE-----` |
| Shown when | **TLS Client Auth** is `true` |

### Client Key

_🔒 secret (write-only) · optional · multiline text_

| | |
|---|---|
| Example | `Client Key. Begins with -----BEGIN RSA PRIVATE KEY-----` |
| Shown when | **TLS Client Auth** is `true` |

## Credentials

### Username

_optional · string_

We recommend configuring a read-only user.

| | |
|---|---|
| Example | `default` |

### Password

_🔒 secret (write-only) · optional · string_

ClickHouse password.

| | |
|---|---|
| Example | `password` |

## Configuration Mode

### Mode

_optional · radio_

Choose how this datasource is used. 'Single table' provides a focused, compact query editor for one table. 'All databases' gives full access to explore any database and table.

| | |
|---|---|
| Default | `classic` |
| Allowed values | `classic` (All databases), `single-table` (Single table) |

### Signal type

_optional · radio_

What kind of data does this table contain?.

| | |
|---|---|
| Default | `logs` |
| Allowed values | `logs` (Logs), `traces` (Traces) |
| Shown when | **Mode** is **Single table** (`single-table`) |

### Default log database

_optional · string_

the default database used by the logs query builder.

| | |
|---|---|
| Example | `default` |

### Default log table

_optional · string_

the default table used by the logs query builder.

| | |
|---|---|
| Default | `otel_logs` |

### Use OTel

_optional · toggle_

Enables Open Telemetry schema versioning.

### OTel version

_optional · select_

| | |
|---|---|
| Default | `latest` |

### Filter Time column

_optional · string_

A lower precision column for filtering logs by timestamp.

### Time column

_optional · string_

Column for the log timestamp, used for high precision sorting.

### Log Level column

_optional · string_

Column for the log level.

### Log Message column

_optional · string_

Column for log message.

### Auto-Select Columns

_optional · toggle_

When enabled, will always include context columns in log queries.

| | |
|---|---|
| Default | `true` |

### Context Columns

_optional · list_

Comma separated list of column names to use for identifying a log's source.

| | |
|---|---|
| Example | `Column name (enter key to add)` |

### Show "View logs" links

_optional · toggle_

Show "View logs" links on trace_id/traceid fields.

### Default trace database

_optional · string_

the default database used by the trace query builder.

| | |
|---|---|
| Example | `default` |

### Default trace table

_optional · string_

the default table used by the trace query builder.

| | |
|---|---|
| Default | `otel_traces` |

### Use OTel

_optional · toggle_

Enables Open Telemetry schema versioning.

### OTel version

_optional · select_

| | |
|---|---|
| Default | `latest` |

### Trace ID column

_optional · string_

Column for the trace ID.

### Span ID column

_optional · string_

Column for the span ID.

### Operation Name column

_optional · string_

Column for the operation name.

### Parent Span ID column

_optional · string_

Column for the parent span ID.

### Service Name column

_optional · string_

Column for the service name.

### Duration Time column

_optional · string_

Column for the duration time.

### Duration Unit

_optional · select_

Unit used by your Duration column. OTel stores nanoseconds; other schemas often use milliseconds or seconds.

| | |
|---|---|
| Default | `nanoseconds` |
| Allowed values | `nanoseconds` (Nanoseconds), `microseconds` (Microseconds), `milliseconds` (Milliseconds), `seconds` (Seconds) |

### Start Time column

_optional · string_

Column for the start time.

### Tags column

_optional · string_

Column for the trace tags.

### Service Tags column

_optional · string_

Column for the service tags.

### Kind column

_optional · string_

Column for the trace kind.

### Status Code column

_optional · string_

Column for the trace status code.

### Status Message column

_optional · string_

Column for the trace status message.

### State column

_optional · string_

Column for the trace state.

### Library Name column

_optional · string_

Column for the instrumentation library name.

### Library Version column

_optional · string_

Column for the instrumentation library version.

### Use Flatten Nested

_optional · toggle_

Enable if your traces table was created with flatten_nested=1.

| | |
|---|---|
| Default | `false` |

### Events prefix

_optional · string_

Prefix for the events column (Events.Timestamp, Events.Name, etc.).

### Links prefix

_optional · string_

Prefix for the trace references column (Links.TraceId, Links.TraceState, etc.).

### Show "View trace" links

_optional · toggle_

Show "View trace" links on trace_id/traceid fields.

### Trace timestamp table suffix

_optional · string_

Suffix appended to the traces table name to locate a companion index keyed by TraceId with Start/End columns. When such a table exists, trace ID lookups narrow the main query to a small time window instead of scanning the whole table. Leave blank to use the OTel default (_trace_id_ts).

| | |
|---|---|
| Example | `_trace_id_ts` |

## Default DB and table

_This section is optional._

### Default database

_optional · string_

the default database used by the query builder.

| | |
|---|---|
| Example | `default` |

### Default table

_optional · string_

the default table used by the query builder.

| | |
|---|---|
| Example | `table` |

## Query settings

_This section is optional._

### Dial Timeout (seconds)

_optional · string_

Timeout in seconds for connection.

| | |
|---|---|
| Default | `10` |
| Example | `10` |

### Query Timeout (seconds)

_optional · string_

Timeout in seconds for read queries.

| | |
|---|---|
| Default | `60` |
| Example | `60` |

### Connection Max Lifetime (minutes)

_optional · string_

Maximum lifetime of a connection in minutes.

| | |
|---|---|
| Default | `5` |
| Example | `5` |

### Max Idle Connections

_optional · string_

Maximum number of idle connections.

| | |
|---|---|
| Default | `25` |
| Example | `25` |

### Max Open Connections

_optional · string_

Maximum number of open connections.

| | |
|---|---|
| Default | `50` |
| Example | `50` |

### Validate SQL

_optional · toggle_

Validate SQL in the editor.

| | |
|---|---|
| Default | `false` |

### Suggest Map keys in filter editor

_optional · toggle_

When enabled, the filter editor probes Map(...) columns for distinct keys to populate the key-suggestion dropdown. On large tables with high-cardinality maps this probe can scan billions of rows. Disable to suppress the probe — operators can still type Map keys manually. Defaults to enabled.

| | |
|---|---|
| Default | `true` |

## Additional settings

_This section is optional._

### Column Alias Tables

_optional · list_

Provide alias tables with a (`alias` String, `select` String, `type` String) schema to use as a source for column selection.

Each item has the following fields:

#### Target Database

_optional · string_

| | |
|---|---|
| Example | `(optional)` |

#### Target Table

_optional · string_

#### Alias Database

_optional · string_

| | |
|---|---|
| Example | `(optional)` |

#### Alias Table

_optional · string_

### Custom Settings

_optional · list_

Additional ClickHouse settings sent with every query as SETTINGS name=value.

Each item has the following fields:

#### Setting

_optional · string_

| | |
|---|---|
| Example | `Setting` |

#### Value

_optional · string_

| | |
|---|---|
| Example | `Value` |

### Enable row limit

_optional · toggle_

Enable using the Grafana row limit setting to limit the number of rows returned from Clickhouse. Ensure the appropriate permissions are set for your user. Only supported for Grafana >= 11.0.0. Defaults to false.

| | |
|---|---|
| Default | `false` |

### Hide table name in ad hoc filters

_optional · toggle_

Show only column names in ad hoc filter keys instead of the full "table.column" format. This simplifies the filter interface when working with schemas that have many tables. Defaults to false.

| | |
|---|---|
| Default | `false` |

## Other settings

### Row Limit

_optional · number_

Row Limit.

### version

_optional · string_

Plugin version that last wrote this configuration. Stamped by the config editor's useConfigDefaults hook on every save; frontend-only.

### enableSchemaCache

_optional · boolean_

Gates the in-process cache that memoizes system.tables / system.columns / DISTINCT column-value lookups used by the query builder. Not exposed in the configuration editor; backend defaults it to true when unset.

| | |
|---|---|
| Default | `true` |

### schemaCacheTTLSeconds

_optional · number_

Controls how long schema-introspection results are considered fresh. Not exposed in the configuration editor; backend defaults it to 60 when unset or <=0.

| | |
|---|---|
| Default | `60` |

