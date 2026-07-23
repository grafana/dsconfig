# MySQL configuration

How to configure the **MySQL** data source (`mysql`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/mysql/).

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
| Example | `localhost:3306` |

### Database name

_optional · string_

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

### Use TLS Client Auth

_optional · toggle_

Enables TLS authentication using client cert configured in secure json data.

| | |
|---|---|
| Default | `false` |

### With CA Cert

_optional · toggle_

Needed for verifying self-signed TLS Certs.

| | |
|---|---|
| Default | `false` |

### Skip TLS Verification

_optional · toggle_

When enabled, skips verification of the MySQL server's TLS certificate chain and host name.

| | |
|---|---|
| Default | `false` |

### Allow Cleartext Passwords

_optional · toggle_

Allows using the cleartext client side plugin if required by an account.

| | |
|---|---|
| Default | `false` |

## TLS/SSL Auth Details

_This section is optional._

### TLS/SSL Client Certificate

_🔒 secret (write-only) · optional · multiline text_

To authenticate with an TLS/SSL client certificate, provide the client certificate here.

| | |
|---|---|
| Example | `-----BEGIN CERTIFICATE-----` |
| Shown when | **Use TLS Client Auth** is `true` |

### TLS/SSL Root Certificate

_🔒 secret (write-only) · optional · multiline text_

If the selected TLS/SSL mode requires a server root certificate, provide it here.

| | |
|---|---|
| Example | `-----BEGIN CERTIFICATE-----` |
| Shown when | **With CA Cert** is `true` |

### TLS/SSL Client Key

_🔒 secret (write-only) · optional · multiline text_

To authenticate with a client TLS/SSL certificate, provide the key here.

| | |
|---|---|
| Example | `-----BEGIN RSA PRIVATE KEY-----` |
| Shown when | **Use TLS Client Auth** is `true` |

## Additional settings

_This section is optional._

### Session timezone

_optional · string_

Specify the timezone used in the database session, such as Europe/Berlin or +02:00. Required if the timezone of the database (or the host of the database) is set to something other than UTC. Set this to +00:00 so Grafana can handle times properly. Set the value used in the session with SET time_zone='...'. If you leave this field empty, the timezone will not be updated. You can find more information in the MySQL documentation.

| | |
|---|---|
| Example | `Europe/Berlin or +02:00` |

### Min time interval

_optional · string_

A lower limit for the auto group by time interval. Recommended to be set to write frequency, for example 1m if your data is written every minute.

| | |
|---|---|
| Example | `1m` |

### Max open

_optional · number_

The maximum number of open connections to the database. If Max idle connections is greater than 0 and the Max open connections is less than Max idle connections, then Max idle connections will be reduced to match the Max open connections limit. If set to 0, there is no limit on the number of open connections.

### Auto max idle

_optional · toggle_

If enabled, automatically set the number of Maximum idle connections to the same value as Max open connections. If the number of maximum open connections is not set it will be set to the default.

| | |
|---|---|
| Default | `true` |

### Max idle

_optional · number_

The maximum number of connections in the idle connection pool. If Max open connections is greater than 0 but less than the Max idle connections, then the Max idle connections will be reduced to match the Max open connections limit. If set to 0, no idle connections are retained.

### Max lifetime

_optional · number_

The maximum amount of time in seconds a connection may be reused. If set to 0, connections are reused forever.

