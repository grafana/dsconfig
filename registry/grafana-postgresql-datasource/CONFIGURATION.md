# PostgreSQL configuration

How to configure the **PostgreSQL** data source (`grafana-postgresql-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/postgres/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)
- [TLS/SSL Auth Details](#tlsssl-auth-details) — _optional_
- [Additional settings](#additional-settings) — _optional_

## Connection

### Host URL

_**required** · string_

| | |
|---|---|
| Example | `localhost:5432` |

### Database name

_**required** · string_

| | |
|---|---|
| Example | `Database` |

## Authentication

### Username

_**required** · string_

| | |
|---|---|
| Example | `Username` |

### Password

_🔒 secret (write-only) · optional · string_

| | |
|---|---|
| Example | `Password` |

### TLS/SSL Mode

_optional · select_

This option determines whether or with what priority a secure TLS/SSL TCP/IP connection will be negotiated with the server.

| | |
|---|---|
| Default | `require` |
| Allowed values | `disable`, `require`, `verify-ca`, `verify-full` |

### TLS/SSL Method

_optional · select_

This option determines how TLS/SSL certifications are configured. Selecting 'File system path' will allow you to configure certificates by specifying paths to existing certificates on the local file system where Grafana is running. Be sure that the file is readable by the user executing the Grafana process. Selecting 'Certificate content' will allow you to configure certificates by specifying its content. The content will be stored encrypted in Grafana's database. When connecting to the database the certificates will be written as files to Grafana's configured data path on the local file system.

| | |
|---|---|
| Default | `file-path` |
| Allowed values | `file-path` (File system path), `file-content` (Certificate content) |
| Shown when | `jsonData_sslmode != 'disable'` |

## TLS/SSL Auth Details

_This section is optional._

### TLS/SSL Root Certificate

_optional · string_

If the selected TLS/SSL mode requires a server root certificate, provide the path to the file here.

| | |
|---|---|
| Example | `TLS/SSL root cert file` |
| Shown when | `jsonData_sslmode != 'disable' && jsonData_tlsConfigurationMethod == 'file-path' && (jsonData_sslmode == 'verify-ca' || jsonData_sslmode == 'verify-full')` |

### TLS/SSL Client Certificate

_optional · string_

To authenticate with an TLS/SSL client certificate, provide the path to the file here. Be sure that the file is readable by the user executing the grafana process.

| | |
|---|---|
| Example | `TLS/SSL client cert file` |
| Shown when | `jsonData_sslmode != 'disable' && jsonData_tlsConfigurationMethod == 'file-path'` |

### TLS/SSL Client Key

_optional · string_

To authenticate with a client TLS/SSL certificate, provide the path to the corresponding key file here. Be sure that the file is only readable by the user executing the grafana process.

| | |
|---|---|
| Example | `TLS/SSL client key file` |
| Shown when | `jsonData_sslmode != 'disable' && jsonData_tlsConfigurationMethod == 'file-path'` |

### TLS/SSL Root Certificate

_🔒 secret (write-only) · optional · multiline text_

If the selected TLS/SSL mode requires a server root certificate, provide it here.

| | |
|---|---|
| Example | `-----BEGIN CERTIFICATE-----` |
| Shown when | `jsonData_sslmode != 'disable' && jsonData_tlsConfigurationMethod == 'file-content' && (jsonData_sslmode == 'verify-ca' || jsonData_sslmode == 'verify-full')` |

### TLS/SSL Client Certificate

_🔒 secret (write-only) · optional · multiline text_

To authenticate with an TLS/SSL client certificate, provide the client certificate here.

| | |
|---|---|
| Example | `-----BEGIN CERTIFICATE-----` |
| Shown when | `jsonData_sslmode != 'disable' && jsonData_tlsConfigurationMethod == 'file-content'` |

### TLS/SSL Client Key

_🔒 secret (write-only) · optional · multiline text_

To authenticate with a client TLS/SSL certificate, provide the key here.

| | |
|---|---|
| Example | `-----BEGIN RSA PRIVATE KEY-----` |
| Shown when | `jsonData_sslmode != 'disable' && jsonData_tlsConfigurationMethod == 'file-content'` |

## Additional settings

_This section is optional._

### Version

_optional · select_

This option controls what functions are available in the PostgreSQL query builder.

| | |
|---|---|
| Default | `903` |
| Allowed values | `900` (9.0), `901` (9.1), `902` (9.2), `903` (9.3), `904` (9.4), `905` (9.5), `906` (9.6), `1000` (10), `1100` (11), `1200` (12), `1300` (13), `1400` (14), `1500` (15) |

### Min time interval

_optional · string_

A lower limit for the auto group by time interval. Recommended to be set to write frequency, for example 1m if your data is written every minute.

| | |
|---|---|
| Example | `1m` |

### TimescaleDB

_optional · toggle_

TimescaleDB is a time-series database built as a PostgreSQL extension. If enabled, Grafana will use time_bucket in the $__timeGroup macro and display TimescaleDB specific aggregate functions in the query builder.

| | |
|---|---|
| Default | `false` |

### Max open

_optional · number_

The maximum number of open connections to the database. If set to 0, there is no limit on the number of open connections.

### Max lifetime

_optional · number_

The maximum amount of time in seconds a connection may be reused. If set to 0, connections are reused forever.

