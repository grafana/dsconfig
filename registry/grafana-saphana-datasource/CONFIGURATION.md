# SAP HANA® configuration

How to configure the **SAP HANA®** data source (`grafana-saphana-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-saphana-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [HTTP](#http)
- [Auth](#auth)
- [TLS / SSL Settings](#tls--ssl-settings)
- [Tenant Properties](#tenant-properties)
- [Additional Properties](#additional-properties)

## HTTP

### Server address

_**required** · string_

SAP HANA server address.

| | |
|---|---|
| Example | `Server address` |

### Server port

_conditionally required · number_

SAP HANA server port (optional if database name filled). Typically this will be 443 for SAP HANA Cloud. For on-prem/multi-tenanted instances, use the corresponding port number.

| | |
|---|---|
| Example | `Server port` |
| Required when | `jsonData_instance == '' || jsonData_databaseName == ''` |

## Auth

### Username

_conditionally required · string_

SAP HANA username.

| | |
|---|---|
| Example | `Username` |
| Required when | `jsonData_tlsAuth != true` |

### Password

_🔒 secret (write-only) · conditionally required · string_

SAP HANA password.

| | |
|---|---|
| Example | `Password` |
| Required when | `jsonData_tlsAuth != true` |

## TLS / SSL Settings

### TLS

_optional · toggle_

Enable TLS/SSL encryption for the connection to SAP HANA. Enabled by default. Disable only when your SAP HANA instance does not have TLS configured (plaintext connections).

| | |
|---|---|
| Default | `false` |

### Skip TLS Verify

_optional · toggle_

Skip TLS Verify.

| | |
|---|---|
| Default | `false` |
| Shown when | `jsonData_tlsDisabled != true` |

### TLS Client Auth

_optional · toggle_

TLS Client Auth.

| | |
|---|---|
| Default | `false` |
| Shown when | `jsonData_tlsDisabled != true` |

### Client Cert

_🔒 secret (write-only) · conditionally required · multiline text_

| | |
|---|---|
| Example | `Client Cert. Begins with -----BEGIN CERTIFICATE-----` |
| Shown when | `jsonData_tlsDisabled != true && jsonData_tlsAuth == true` |
| Required when | **TLS Client Auth** is `true` |

### Client Key

_🔒 secret (write-only) · conditionally required · multiline text_

| | |
|---|---|
| Example | `Client Key. Begins with -----BEGIN RSA PRIVATE KEY-----` |
| Shown when | `jsonData_tlsDisabled != true && jsonData_tlsAuth == true` |
| Required when | **TLS Client Auth** is `true` |

### With CA Cert

_optional · toggle_

Needed for verifying self-signed TLS Certs.

| | |
|---|---|
| Default | `false` |
| Shown when | `jsonData_tlsDisabled != true` |

### CA Cert

_🔒 secret (write-only) · optional · multiline text_

| | |
|---|---|
| Example | `CA Cert. Begins with -----BEGIN CERTIFICATE-----` |
| Shown when | `jsonData_tlsDisabled != true && jsonData_tlsAuthWithCACert == true` |

## Tenant Properties

### Tenant database name

_optional · string_

Tenant database name (optional). If database name is set as well as the instance number, port is not required.

| | |
|---|---|
| Example | `Tenant database name` |

### Tenant instance number

_optional · string_

SAP HANA tenant instance number (optional). If instance number is set, port is not required.

| | |
|---|---|
| Example | `Tenant instance number` |

## Additional Properties

### Default schema

_optional · string_

Default schema to be used. Can be empty.

| | |
|---|---|
| Example | `Default schema` |

### Timeout

_optional · string_

| | |
|---|---|
| Default | `30` |
| Example | `Timeout for connections` |

