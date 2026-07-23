# CockroachDB configuration

How to configure the **CockroachDB** data source (`grafana-cockroachdb-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/grafana-cockroachdb-datasource/).

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
| Example | `localhost:26257` |

### Database

_**required** · string_

| | |
|---|---|
| Example | `defaultdb` |

## Authentication

### User

_**required** · string_

| | |
|---|---|
| Example | `User` |
| Shown when | `jsonData_authType != ''` |

### Password

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Example | `Password` |
| Shown when | `jsonData_authType == 'SQL Authentication' || jsonData_authType == 'TLS/SSL Authentication'` |

### Credential cache path

_conditionally required · string_

| | |
|---|---|
| Example | `/tmp/krb5cc_1000` |
| Shown when | **authType** is `Kerberos Authentication` |

### Kerberos server name

_optional · string_

The Kerberos service name to use (optional). Default is 'postgres'.

| | |
|---|---|
| Example | `postgres` |
| Shown when | **authType** is `Kerberos Authentication` |

### TLS/SSL Method

_optional · select_

This option determines how TLS/SSL certifications are configured. Selecting File system path will allow you to configure certificates by specifying paths to existing certificates on the local file system where Grafana is running. Be sure that the file is readable by the user executing the Grafana process. Selecting Certificate content will allow you to configure certificates by specifying its content. The content will be stored encrypted in Grafana's database. When connecting to the database the certificates will be written as files to Grafana's configured data path on the local file system.

| | |
|---|---|
| Default | `file-content` |
| Allowed values | `file-content`, `file-path` |
| Shown when | `jsonData_authType == 'TLS/SSL Authentication' && jsonData_sslmode != 'disable'` |

## TLS/SSL Auth Details

_This section is optional._

### TLS/SSL Root Certificate

_conditionally required · string_

If the selected TLS/SSL mode requires a server root certificate, provide the path to the file here.

| | |
|---|---|
| Example | `TLS/SSL root cert file` |
| Shown when | `jsonData_authType == 'TLS/SSL Authentication' && jsonData_sslmode != 'disable' && jsonData_tlsConfigurationMethod == 'file-path'` |

### TLS/SSL Client Certificate

_conditionally required · string_

To authenticate with an TLS/SSL client certificate, provide the path to the file here. Be sure that the file is readable by the user executing the grafana process.

| | |
|---|---|
| Example | `TLS/SSL client cert file` |
| Shown when | `jsonData_authType == 'TLS/SSL Authentication' && jsonData_sslmode != 'disable' && jsonData_tlsConfigurationMethod == 'file-path'` |

### TLS/SSL Client Key

_conditionally required · string_

To authenticate with a client TLS/SSL certificate, provide the path to the corresponding key file here. Be sure that the file is only readable by the user executing the grafana process.

| | |
|---|---|
| Example | `TLS/SSL client key file` |
| Shown when | `jsonData_authType == 'TLS/SSL Authentication' && jsonData_sslmode != 'disable' && jsonData_tlsConfigurationMethod == 'file-path'` |

### TLS/SSL Root Certificate

_🔒 secret (write-only) · conditionally required · multiline text_

If the selected TLS/SSL mode requires a server root certificate, provide it here.

| | |
|---|---|
| Example | `-----BEGIN CERTIFICATE-----` |
| Shown when | `jsonData_authType == 'TLS/SSL Authentication' && jsonData_sslmode != 'disable' && jsonData_tlsConfigurationMethod == 'file-content'` |

### TLS/SSL Client Certificate

_🔒 secret (write-only) · conditionally required · multiline text_

To authenticate with an TLS/SSL client certificate, provide the client certificate here.

| | |
|---|---|
| Example | `-----BEGIN CERTIFICATE-----` |
| Shown when | `jsonData_authType == 'TLS/SSL Authentication' && jsonData_sslmode != 'disable' && jsonData_tlsConfigurationMethod == 'file-content'` |

### TLS/SSL Client Key

_🔒 secret (write-only) · conditionally required · multiline text_

To authenticate with a client TLS/SSL certificate, provide the key here.

| | |
|---|---|
| Example | `-----BEGIN RSA PRIVATE KEY-----` |
| Shown when | `jsonData_authType == 'TLS/SSL Authentication' && jsonData_sslmode != 'disable' && jsonData_tlsConfigurationMethod == 'file-content'` |

## Additional settings

_This section is optional._

### Max open

_optional · number_

The maximum number of open connections to the database. If Max idle connections is greater than 0 and the Max open connections is less than Max idle connections, then Max idle connections will be reduced to match the Max open connections limit. If set to 0, there is no limit on the number of open connections.

| | |
|---|---|
| Default | `5` |

### Auto max idle

_optional · toggle_

If enabled, automatically set the number of Maximum idle connections to the same value as Max open connections. If the number of maximum open connections is not set it will be set to the default.

| | |
|---|---|
| Default | `false` |

### Max idle

_optional · number_

The maximum number of connections in the idle connection pool.If Max open connections is greater than 0 but less than the Max idle connections, then the Max idle connections will be reduced to match the Max open connections limit. If set to 0, no idle connections are retained.

| | |
|---|---|
| Default | `2` |

### Max lifetime

_optional · number_

The maximum amount of time in seconds a connection may be reused. If set to 0, connections are reused forever.

| | |
|---|---|
| Default | `300` |

### Query timeout

_optional · number_

The maximum amount of time in seconds to wait for a query to complete. Valid range: 5-600 seconds (10 minutes). If set to 0, the default of 30 seconds will be used.

| | |
|---|---|
| Default | `30` |

### krb5 config file path

_conditionally required · string_

The path to the configuration file for the [MIT krb5 package](https://web.mit.edu/kerberos/krb5-1.12/doc/admin/conf_files/krb5_conf.html). The default is `/etc/krb5.conf`.

| | |
|---|---|
| Default | `/etc/krb5.conf` |
| Shown when | **authType** is `Kerberos Authentication` |

### TLS/SSL Mode

_optional · select_

This option determines whether or with what priority a secure TLS/SSL TCP/IP connection will be negotiated with the server.

| | |
|---|---|
| Default | `require` |
| Allowed values | `disable`, `require`, `verify-ca`, `verify-full` |
| Shown when | **authType** is `TLS/SSL Authentication` |

