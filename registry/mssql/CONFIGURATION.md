# Microsoft SQL Server configuration

How to configure the **Microsoft SQL Server** data source (`mssql`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/mssql/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [TLS/SSL Auth](#tlsssl-auth)
- [Authentication](#authentication)
- [Additional settings](#additional-settings) — _optional_
- [Windows AD: Advanced Settings](#windows-ad-advanced-settings) — _optional_

## Connection

### Host

_**required** · string_

| | |
|---|---|
| Example | `localhost:1433` |

### Database

_**required** · string_

| | |
|---|---|
| Example | `database name` |

## TLS/SSL Auth

### Encrypt

_optional · select_

Determines whether or to which extent a secure SSL TCP/IP connection will be negotiated with the server. 'disable' - Data sent between client and server is not encrypted. 'false' - Data sent between client and server is not encrypted beyond the login packet. (default) 'true' - Data sent between client and server is encrypted. If you're using an older version of Microsoft SQL Server like 2008 and 2008R2 you may need to disable encryption to be able to connect.

| | |
|---|---|
| Default | `false` |
| Allowed values | `disable`, `false`, `true` |

### Skip TLS Verify

_optional · toggle_

| | |
|---|---|
| Default | `false` |
| Shown when | **Encrypt** is `true` |

### TLS/SSL Root Certificate

_optional · string_

Path to file containing the public key certificate of the CA that signed the SQL Server certificate. Needed when the server certificate is self signed.

| | |
|---|---|
| Example | `TLS/SSL root certificate file path` |
| Shown when | `jsonData_encrypt == 'true' && jsonData_tlsSkipVerify != true` |

### Hostname in server certificate

_optional · string_

| | |
|---|---|
| Example | `Common Name (CN) in server certificate` |
| Shown when | `jsonData_encrypt == 'true' && jsonData_tlsSkipVerify != true` |

## Authentication

### Authentication Type

_optional · select_

'SQL Server Authentication' Default mechanism (SQL login or Windows DOMAIN\User format). 'Windows Authentication' Integrated Security - SSO for users already logged on to Windows. 'Azure AD Authentication' Managed Service Identity or Client Secret. 'Windows AD: Username + password' Kerberos with username/password. 'Windows AD: Keytab' Kerberos with a keytab file. 'Windows AD: Credential cache' Kerberos with a credential cache path. 'Windows AD: Credential cache file' Kerberos with a credential-cache lookup file.

| | |
|---|---|
| Default | `SQL Server Authentication` |
| Allowed values | `SQL Server Authentication`, `Windows Authentication`, `Windows AD: Username + password`, `Windows AD: Keytab` (Windows AD: Keytab file), `Windows AD: Credential cache`, `Windows AD: Credential cache file`, `Azure AD Authentication` |

### Username

_conditionally required · string_

For 'Windows AD: Username + password' and 'Windows AD: Credential cache file' auth types, use the format user@EXAMPLE.COM.

| | |
|---|---|
| Example | `user` |
| Shown when | `jsonData_authenticationType == 'SQL Server Authentication' || jsonData_authenticationType == 'Windows AD: Username + password' || jsonData_authenticationType == 'Windows AD: Keytab' || jsonData_authenticationType == 'Windows AD: Credential cache file'` |

### Password

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Example | `Password` |
| Shown when | `jsonData_authenticationType == 'SQL Server Authentication' || jsonData_authenticationType == 'Windows AD: Username + password'` |

### Keytab file path

_conditionally required · string_

| | |
|---|---|
| Example | `/home/grot/grot.keytab` |
| Shown when | **Authentication Type** is **Windows AD: Keytab file** (`Windows AD: Keytab`) |

### Credential cache path

_conditionally required · string_

| | |
|---|---|
| Example | `/tmp/krb5cc_1000` |
| Shown when | **Authentication Type** is `Windows AD: Credential cache` |

### Credential cache file path

_conditionally required · string_

| | |
|---|---|
| Example | `/home/grot/cache.json` |
| Shown when | **Authentication Type** is `Windows AD: Credential cache file` |

### azureCredentials

_optional_

Object holding Azure AD auth credentials (authType: 'msi' | 'workloadidentity' | 'clientsecret' | 'clientsecret-obo' | 'ad-password' | 'clientcertificate' | 'currentuser', plus type-specific fields azureCloud / tenantId / clientId / userId). Written by the @grafana/azure-sdk AzureCredentialsForm; the secret component (clientSecret / password / privateKey / certificatePassword) is stored write-only in secureJsonData.azureClientSecret (and related keys) — read secureJsonFields.azureClientSecret to check whether it is configured.

| | |
|---|---|
| Shown when | **Authentication Type** is `Azure AD Authentication` |

### azureClientSecret

_🔒 secret (write-only) · optional · string_

Azure AD client secret (for authType='clientsecret' or 'clientsecret-obo'). Written by @grafana/azure-sdk's AzureCredentialsForm.

| | |
|---|---|
| Shown when | **Authentication Type** is `Azure AD Authentication` |

## Additional settings

_This section is optional._

### Max open

_optional · number_

The maximum number of open connections to the database. If set to 0, there is no limit on the number of open connections.

### Auto max idle

_optional · toggle_

If enabled, automatically set the number of Maximum idle connections to the same value as Max open connections.

| | |
|---|---|
| Default | `true` |

### Max idle

_optional · number_

The maximum number of connections in the idle connection pool.

### Max lifetime

_optional · number_

The maximum amount of time in seconds a connection may be reused. If set to 0, connections are reused forever.

### Min time interval

_optional · string_

A lower limit for the auto group by time interval. Recommended to be set to write frequency, for example 1m if your data is written every minute.

| | |
|---|---|
| Example | `1m` |

### Connection timeout

_optional · number_

The number of seconds to wait before canceling the request when connecting to the database. The default is 0, meaning no timeout.

| | |
|---|---|
| Default | `0` |

## Windows AD: Advanced Settings

_This section is optional._

### UDP Preference Limit

_optional · number_

The default is 1 and means always use TCP and is optional.

| | |
|---|---|
| Default | `1` |
| Example | `0` |
| Shown when | `jsonData_authenticationType == 'Windows AD: Username + password' || jsonData_authenticationType == 'Windows AD: Keytab' || jsonData_authenticationType == 'Windows AD: Credential cache' || jsonData_authenticationType == 'Windows AD: Credential cache file'` |

### DNS Lookup KDC

_optional · string_

Indicate whether DNS `SRV` records should be used to locate the KDCs and other servers for a realm. The default is 'true'.

| | |
|---|---|
| Example | `true` |
| Shown when | `jsonData_authenticationType == 'Windows AD: Username + password' || jsonData_authenticationType == 'Windows AD: Keytab' || jsonData_authenticationType == 'Windows AD: Credential cache' || jsonData_authenticationType == 'Windows AD: Credential cache file'` |

### krb5 config file path

_optional · string_

The path to the configuration file for the MIT krb5 package. The default is /etc/krb5.conf.

| | |
|---|---|
| Default | `/etc/krb5.conf` |
| Shown when | `jsonData_authenticationType == 'Windows AD: Username + password' || jsonData_authenticationType == 'Windows AD: Keytab' || jsonData_authenticationType == 'Windows AD: Credential cache' || jsonData_authenticationType == 'Windows AD: Credential cache file'` |

