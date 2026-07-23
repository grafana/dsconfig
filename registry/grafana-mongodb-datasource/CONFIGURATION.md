# MongoDB configuration

How to configure the **MongoDB** data source (`grafana-mongodb-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-mongodb-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)
- [Additional Settings](#additional-settings) — _optional_

## Connection

### Connection string

_**required** · string_

A connection string contains the parameters required to connect to MongoDB.

| | |
|---|---|
| Example | `mongodb+srv://cluster.host.net/dbname?retryWrites=true&w=majority` |

**Connection string**

A connection string contains the parameters required to connect to MongoDB. [View formatting details here.](https://www.mongodb.com/docs/manual/reference/connection-string/)

## Authentication

### Authentication method

_optional · radio_

Choose an authentication method to access the data source.

| | |
|---|---|
| Default | `BasicAuth` |
| Allowed values | `NoAuth` (No Authentication), `BasicAuth` (Credentials), `custom-Kerberos` (Kerberos) |

### User

_optional · string_

The username assigned to the MongoDB account.

| | |
|---|---|
| Example | `User` |
| Shown when | **Authentication method** is **Credentials** (`BasicAuth`) |

### Password

_🔒 secret (write-only) · optional · string_

The password assigned to the MongoDB account.

| | |
|---|---|
| Example | `Password` |
| Shown when | **Authentication method** is **Credentials** (`BasicAuth`) |

### User

_optional · string_

The client principal's username. Enabled when connection string includes query string authMethod=GSSAPI.

| | |
|---|---|
| Example | `hello@EXAMPLE.COM` |
| Shown when | **Authentication method** is **Kerberos** (`custom-Kerberos`) |

### Password

_🔒 secret (write-only) · optional · string_

The client principal password that will be used to authenticate. Optional if a keytab or cache file is present.

| | |
|---|---|
| Example | `Password` |
| Shown when | **Authentication method** is **Kerberos** (`custom-Kerberos`) |

### KeyTab file path

_optional · string_

Absolute file path KeyTab for keytab file. If present will ignore password. Enabled when connection string includes query string authMethod=GSSAPI.

| | |
|---|---|
| Example | `/tmp/example.keytab` |
| Shown when | **Authentication method** is **Kerberos** (`custom-Kerberos`) |

### Global ccache file path

_optional · string_

Absolute file path to global compiler cache (ccache) file. If present will ignore password. Enabled when connection string includes query string authMethod=GSSAPI.

| | |
|---|---|
| Example | `/tmp/krb5cc_1000` |
| Shown when | **Authentication method** is **Kerberos** (`custom-Kerberos`) |

### Ccache lookup file

_optional · string_

Absolute file path to  the JSON file that provides the Kerberos compiler cache (ccache) based on username principal and connection string. If present will ignore password. Enabled when connection string includes query string authMethod=GSSAPI.

| | |
|---|---|
| Example | `/tmp/krb5-ccache-lookup.json` |
| Shown when | **Authentication method** is **Kerberos** (`custom-Kerberos`) |

## Additional Settings

Additional settings are optional settings that can be configured for more control over your data source.

_This section is optional._

### Enable syntax validation

_optional · toggle_

Enable real time query syntax validation. MongoDB BSON syntax will be validated as you type and show contextual errors.

### Password

_🔒 secret (write-only) · optional · string_

Password.

| | |
|---|---|
| Example | `Password` |

### Rows to Return

_optional · string_

Increasing this too much may lead to performance issues for larger queries.

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

## Other settings

### Basic authentication enabled

_optional · boolean_

Standard Grafana basic-auth enabled flag. The backend initializes BasicAuthEnabled from this value and also forces it on when a username or password is present (pkg/models/settings.go). Set by datasource provisioning and by the config editor's on-load migration of legacy datasources, not by selecting the Credentials method in the current editor.

### serverName

_optional · string_

TLS server name used to verify the hostname on the server's certificate when tlsAuth is enabled. Not exposed in the configuration editor; set via datasource provisioning.

### tlsAuth

_optional · boolean_

Enables TLS client-certificate authentication; the backend supplies tlsClientCert and tlsClientKey to the server. Not exposed in the configuration editor; set via datasource provisioning.

### tlsAuthWithCACert

_optional · boolean_

Enables verification of the server's TLS certificate against a custom CA (tlsCACert). Not exposed in the configuration editor; set via datasource provisioning.

### tlsSkipVerify

_optional · boolean_

Skips verification of the server's TLS certificate chain and host name (applied by the backend when tlsAuthWithCACert is enabled). Not exposed in the configuration editor; set via datasource provisioning.

### tlsCACert

_🔒 secret (write-only) · optional · string_

CA certificate PEM used to verify the server's TLS certificate when tlsAuthWithCACert is enabled. Not exposed in the configuration editor; set via datasource provisioning.

### tlsClientCert

_🔒 secret (write-only) · optional · string_

Client certificate PEM used when tlsAuth is enabled. Not exposed in the configuration editor; set via datasource provisioning.

### tlsClientKey

_🔒 secret (write-only) · optional · string_

Client private key PEM used when tlsAuth is enabled. Not exposed in the configuration editor; set via datasource provisioning.

### user

_optional · string_

Legacy username field. Datasources created before v1.9.0 stored the MongoDB username here; the backend migrates it to the root basicAuthUser and enables basic auth (pkg/models/settings.go). New configurations use the root basicAuthUser field.

### skipTLSValidation

_optional · boolean_

Legacy flag; the backend copies it to tlsSkipVerify (InsecureSkipVerify) at load time (pkg/models/settings.go). New configurations use tlsSkipVerify.

### credentials

_optional · boolean_

Legacy frontend-only flag used by the config editor's on-load migration to detect pre-v1.9.0 basic-auth datasources. Never read by the backend.

### password

_🔒 secret (write-only) · optional · string_

Legacy secure password. Datasources created before v1.9.0 stored the MongoDB password here; the backend migrates it to the basic-auth password (pkg/models/settings.go). New configurations use basicAuthPassword.

