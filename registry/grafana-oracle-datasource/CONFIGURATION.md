# Oracle Database configuration

How to configure the **Oracle Database** data source (`grafana-oracle-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-oracle-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)
- [Additional Settings](#additional-settings) — _optional_

## Connection

Provide information to connect to this datasource.

### Connection methods

_optional · select_

| | |
|---|---|
| Default | `tcp` |
| Allowed values | `tcp` (Host with TCP Port), `tns` (TNSNames Entry) |

### Host

_conditionally required · string_

Hostname or IP Address with TCP port number.

| | |
|---|---|
| Example | `Host` |
| Shown when | **Connection methods** is **Host with TCP Port** (`tcp`) |
| Required when | `jsonData_useTNSNamesBasedConnection != true` |

### Database

_conditionally required · string_

Name of database.

| | |
|---|---|
| Example | `Database` |
| Shown when | **Connection methods** is **Host with TCP Port** (`tcp`) |
| Required when | `jsonData_useTNSNamesBasedConnection != true` |

### TNSName

_conditionally required · string_

Use connection config specified in tnsnames.ora.

| | |
|---|---|
| Example | `server/DB` |
| Shown when | **Connection methods** is **TNSNames Entry** (`tns`) |
| Required when | **useTNSNamesBasedConnection** is `true` |

## Authentication

Provide information to grant access to this datasource.

### Oracle authentication

_optional · select_

| | |
|---|---|
| Default | `basic` |
| Allowed values | `basic` (Basic Authentication), `kerberos` (Kerberos Authentication) |
| Disabled when | **Connection methods** is **Host with TCP Port** (`tcp`) |

### User

_conditionally required · string_

An Oracle username with access to the specified database.

| | |
|---|---|
| Example | `User` |
| Shown when | **Oracle authentication** is **Basic Authentication** (`basic`) |
| Required when | `jsonData_useKerberosAuthentication != true` |

### Password

_🔒 secret (write-only) · conditionally required · string_

An Oracle password with access to the specified database.

| | |
|---|---|
| Example | `Password` |
| Shown when | **Oracle authentication** is **Basic Authentication** (`basic`) |
| Required when | `jsonData_useKerberosAuthentication != true` |

## Additional Settings

Additional settings are optional settings that can be configured for more control over your data source.

_This section is optional._

### Time zone

_optional · select_

Choose the default timezone of the Oracle server. Typically this is UTC.

| | |
|---|---|
| Default | `UTC` |

### Connection Pool size

_optional · number_

Choose the maximum number of connections in the connection pool. Defaults to 50. Takes precedence over environment variable 'GF_PLUGINS_ORACLE_DATASOURCE_POOLSIZE' if set.

| | |
|---|---|
| Default | `50` |

### Dataproxy Timeout

_optional · number_

Choose the maximum time in seconds to wait for a response. Defaults to 120 seconds. Takes precedence over environment variable 'GF_DATAPROXY_TIMEOUT' if set.

| | |
|---|---|
| Default | `120` |

### Prefetch Row Size

_optional · number_

Row Prefetching allow you to set the number of rows to prefetch into the client while a result set is being populated during a query. This feature reduces the number of round trips to the server.

### Row Limit

_optional · number_

Maximum number of rows returned per query. Defaults to 1,000,000.

| | |
|---|---|
| Default | `1e+06` |

## Other settings

### useTNSNamesBasedConnection

_optional · boolean_

Connection-method discriminator: false selects 'Host with TCP Port' (host + port + database), true selects 'TNSNames Entry' (a tnsnames.ora entry). Written by the config editor's Connection methods selector.

### useKerberosAuthentication

_optional · boolean_

Authentication discriminator: false selects Basic Authentication (username + password), true selects Kerberos Authentication. The config editor only exposes Kerberos when useTNSNamesBasedConnection is true, though the backend also accepts Kerberos with a host + port connection.

### use_dst

_optional · boolean_

Parsed by the backend into DBConnectionOptions.DSTEnabled (pkg/models/settings.go:23) but never read when building the connection string. Declared in the frontend OracleOptions type (src/types.ts:9) but not written by the config editor — use_dst is a per-query/annotation option (src/types.ts:30, src/datasource.ts:109), not a datasource-level setting.

