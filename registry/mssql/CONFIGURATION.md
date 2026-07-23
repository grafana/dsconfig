# Microsoft SQL Server configuration

Configuration reference for the **Microsoft SQL Server** data source (`mssql`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/mssql/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `url` | string | root | yes | Host |
| `jsonData.database` | string | jsonData | yes | Database |
| `jsonData.encrypt` | enum (disable, false, true) | jsonData |  | Determines whether or to which extent a secure SSL TCP/IP connection will be negotiated with the server. 'disable' - Data sent between client and server is not encrypted. 'false' - Data sent between client and server is not encrypted beyond the login packet. (default) 'true' - Data sent between client and server is encrypted. If you're using an older version of Microsoft SQL Server like 2008 and 2008R2 you may need to disable encryption to be able to connect. |
| `jsonData.tlsSkipVerify` | boolean | jsonData |  | Skip TLS Verify |
| `jsonData.sslRootCertFile` | string | jsonData |  | Path to file containing the public key certificate of the CA that signed the SQL Server certificate. Needed when the server certificate is self signed. |
| `jsonData.serverName` | string | jsonData |  | Hostname in server certificate |
| `jsonData.authenticationType` | enum (SQL Server Authentication, Windows Authentication, Windows AD: Username + password, Windows AD: Keytab, Windows AD: Credential cache, Windows AD: Credential cache file, Azure AD Authentication) | jsonData |  | 'SQL Server Authentication' Default mechanism (SQL login or Windows DOMAIN\User format). 'Windows Authentication' Integrated Security - SSO for users already logged on to Windows. 'Azure AD Authentication' Managed Service Identity or Client Secret. 'Windows AD: Username + password' Kerberos with username/password. 'Windows AD: Keytab' Kerberos with a keytab file. 'Windows AD: Credential cache' Kerberos with a credential cache path. 'Windows AD: Credential cache file' Kerberos with a credential-cache lookup file. |
| `user` | string | root | conditional | For 'Windows AD: Username + password' and 'Windows AD: Credential cache file' auth types, use the format user@EXAMPLE.COM. |
| `secureJsonData.password` 🔒 | string | secureJsonData | conditional | Password |
| `jsonData.keytabFilePath` | string | jsonData | conditional | Keytab file path |
| `jsonData.credentialCache` | string | jsonData | conditional | Credential cache path |
| `jsonData.credentialCacheLookupFile` | string | jsonData | conditional | Credential cache file path |
| `jsonData.azureCredentials` | any | jsonData |  | Object holding Azure AD auth credentials (authType: 'msi' \| 'workloadidentity' \| 'clientsecret' \| 'clientsecret-obo' \| 'ad-password' \| 'clientcertificate' \| 'currentuser', plus type-specific fields azureCloud / tenantId / clientId / userId). Written by the @grafana/azure-sdk AzureCredentialsForm; the secret component (clientSecret / password / privateKey / certificatePassword) is stored write-only in secureJsonData.azureClientSecret (and related keys) — read secureJsonFields.azureClientSecret to check whether it is configured. |
| `secureJsonData.azureClientSecret` 🔒 | string | secureJsonData |  | Azure AD client secret (for authType='clientsecret' or 'clientsecret-obo'). Written by @grafana/azure-sdk's AzureCredentialsForm. |
| `jsonData.configFilePath` | string | jsonData |  | The path to the configuration file for the MIT krb5 package. The default is /etc/krb5.conf. |
| `jsonData.UDPConnectionLimit` | number | jsonData |  | The default is 1 and means always use TCP and is optional. |
| `jsonData.enableDNSLookupKDC` | string | jsonData |  | Indicate whether DNS `SRV` records should be used to locate the KDCs and other servers for a realm. The default is 'true'. |
| `jsonData.timeInterval` | string | jsonData |  | A lower limit for the auto group by time interval. Recommended to be set to write frequency, for example 1m if your data is written every minute. |
| `jsonData.connectionTimeout` | number | jsonData |  | The number of seconds to wait before canceling the request when connecting to the database. The default is 0, meaning no timeout. |
| `jsonData.maxOpenConns` | number | jsonData |  | The maximum number of open connections to the database. If set to 0, there is no limit on the number of open connections. |
| `jsonData.maxIdleConnsAuto` | boolean | jsonData |  | If enabled, automatically set the number of Maximum idle connections to the same value as Max open connections. |
| `jsonData.maxIdleConns` | number | jsonData |  | The maximum number of connections in the idle connection pool. |
| `jsonData.connMaxLifetime` | number | jsonData |  | The maximum amount of time in seconds a connection may be reused. If set to 0, connections are reused forever. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### `SQL Server Authentication`

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Microsoft SQL Server
    type: mssql
    access: proxy
    url: "localhost:1433"
    user: user
    jsonData:
      authenticationType: SQL Server Authentication
      connectionTimeout: 0
      database: database name
      encrypt: "false"
      maxIdleConnsAuto: true
      tlsSkipVerify: false
    secureJsonData:
      password: "<YOUR_PASSWORD>"
```

**Terraform**

```hcl
resource "grafana_data_source" "mssql_SQL_Server_Authentication" {
  type = "mssql"
  name = "Microsoft SQL Server"
  url = "localhost:1433"
  basic_auth_username = "user"

  json_data_encoded = jsonencode({
    authenticationType = "SQL Server Authentication"
    connectionTimeout = 0
    database = "database name"
    encrypt = "false"
    maxIdleConnsAuto = true
    tlsSkipVerify = false
  })

  secure_json_data_encoded = jsonencode({
    password = "<YOUR_PASSWORD>"
  })
}
```

### `Windows Authentication`

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Microsoft SQL Server
    type: mssql
    access: proxy
    url: "localhost:1433"
    jsonData:
      authenticationType: Windows Authentication
      connectionTimeout: 0
      database: database name
      encrypt: "false"
      maxIdleConnsAuto: true
      tlsSkipVerify: false
```

**Terraform**

```hcl
resource "grafana_data_source" "mssql_Windows_Authentication" {
  type = "mssql"
  name = "Microsoft SQL Server"
  url = "localhost:1433"

  json_data_encoded = jsonencode({
    authenticationType = "Windows Authentication"
    connectionTimeout = 0
    database = "database name"
    encrypt = "false"
    maxIdleConnsAuto = true
    tlsSkipVerify = false
  })
}
```

### `Windows AD: Username + password`

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Microsoft SQL Server
    type: mssql
    access: proxy
    url: "localhost:1433"
    user: user
    jsonData:
      UDPConnectionLimit: 1
      authenticationType: "Windows AD: Username + password"
      configFilePath: /etc/krb5.conf
      connectionTimeout: 0
      database: database name
      encrypt: "false"
      maxIdleConnsAuto: true
      tlsSkipVerify: false
    secureJsonData:
      password: "<YOUR_PASSWORD>"
```

**Terraform**

```hcl
resource "grafana_data_source" "mssql_Windows_AD__Username___password" {
  type = "mssql"
  name = "Microsoft SQL Server"
  url = "localhost:1433"
  basic_auth_username = "user"

  json_data_encoded = jsonencode({
    UDPConnectionLimit = 1
    authenticationType = "Windows AD: Username + password"
    configFilePath = "/etc/krb5.conf"
    connectionTimeout = 0
    database = "database name"
    encrypt = "false"
    maxIdleConnsAuto = true
    tlsSkipVerify = false
  })

  secure_json_data_encoded = jsonencode({
    password = "<YOUR_PASSWORD>"
  })
}
```

### Windows AD: Keytab file (`Windows AD: Keytab`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Microsoft SQL Server
    type: mssql
    access: proxy
    url: "localhost:1433"
    user: user
    jsonData:
      UDPConnectionLimit: 1
      authenticationType: "Windows AD: Keytab"
      configFilePath: /etc/krb5.conf
      connectionTimeout: 0
      database: database name
      encrypt: "false"
      keytabFilePath: /home/grot/grot.keytab
      maxIdleConnsAuto: true
      tlsSkipVerify: false
```

**Terraform**

```hcl
resource "grafana_data_source" "mssql_Windows_AD__Keytab" {
  type = "mssql"
  name = "Microsoft SQL Server"
  url = "localhost:1433"
  basic_auth_username = "user"

  json_data_encoded = jsonencode({
    UDPConnectionLimit = 1
    authenticationType = "Windows AD: Keytab"
    configFilePath = "/etc/krb5.conf"
    connectionTimeout = 0
    database = "database name"
    encrypt = "false"
    keytabFilePath = "/home/grot/grot.keytab"
    maxIdleConnsAuto = true
    tlsSkipVerify = false
  })
}
```

### `Windows AD: Credential cache`

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Microsoft SQL Server
    type: mssql
    access: proxy
    url: "localhost:1433"
    jsonData:
      UDPConnectionLimit: 1
      authenticationType: "Windows AD: Credential cache"
      configFilePath: /etc/krb5.conf
      connectionTimeout: 0
      credentialCache: /tmp/krb5cc_1000
      database: database name
      encrypt: "false"
      maxIdleConnsAuto: true
      tlsSkipVerify: false
```

**Terraform**

```hcl
resource "grafana_data_source" "mssql_Windows_AD__Credential_cache" {
  type = "mssql"
  name = "Microsoft SQL Server"
  url = "localhost:1433"

  json_data_encoded = jsonencode({
    UDPConnectionLimit = 1
    authenticationType = "Windows AD: Credential cache"
    configFilePath = "/etc/krb5.conf"
    connectionTimeout = 0
    credentialCache = "/tmp/krb5cc_1000"
    database = "database name"
    encrypt = "false"
    maxIdleConnsAuto = true
    tlsSkipVerify = false
  })
}
```

### `Windows AD: Credential cache file`

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Microsoft SQL Server
    type: mssql
    access: proxy
    url: "localhost:1433"
    user: user
    jsonData:
      UDPConnectionLimit: 1
      authenticationType: "Windows AD: Credential cache file"
      configFilePath: /etc/krb5.conf
      connectionTimeout: 0
      credentialCacheLookupFile: /home/grot/cache.json
      database: database name
      encrypt: "false"
      maxIdleConnsAuto: true
      tlsSkipVerify: false
```

**Terraform**

```hcl
resource "grafana_data_source" "mssql_Windows_AD__Credential_cache_file" {
  type = "mssql"
  name = "Microsoft SQL Server"
  url = "localhost:1433"
  basic_auth_username = "user"

  json_data_encoded = jsonencode({
    UDPConnectionLimit = 1
    authenticationType = "Windows AD: Credential cache file"
    configFilePath = "/etc/krb5.conf"
    connectionTimeout = 0
    credentialCacheLookupFile = "/home/grot/cache.json"
    database = "database name"
    encrypt = "false"
    maxIdleConnsAuto = true
    tlsSkipVerify = false
  })
}
```

### `Azure AD Authentication`

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Microsoft SQL Server
    type: mssql
    access: proxy
    url: "localhost:1433"
    jsonData:
      authenticationType: Azure AD Authentication
      connectionTimeout: 0
      database: database name
      encrypt: "false"
      maxIdleConnsAuto: true
      tlsSkipVerify: false
```

**Terraform**

```hcl
resource "grafana_data_source" "mssql_Azure_AD_Authentication" {
  type = "mssql"
  name = "Microsoft SQL Server"
  url = "localhost:1433"

  json_data_encoded = jsonencode({
    authenticationType = "Azure AD Authentication"
    connectionTimeout = 0
    database = "database name"
    encrypt = "false"
    maxIdleConnsAuto = true
    tlsSkipVerify = false
  })
}
```

