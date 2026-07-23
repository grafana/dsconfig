# CockroachDB configuration

Configuration reference for the **CockroachDB** data source (`grafana-cockroachdb-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/grafana-cockroachdb-datasource/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.url` | string | jsonData | yes | Host URL |
| `jsonData.database` | string | jsonData | yes | Database |
| `jsonData.authType` | enum (SQL Authentication, Kerberos Authentication, TLS/SSL Authentication) | jsonData |  |  |
| `jsonData.user` | string | jsonData | yes | User |
| `secureJsonData.password` 🔒 | string | secureJsonData | conditional | Password |
| `jsonData.credentialCache` | string | jsonData | conditional | Credential cache path |
| `jsonData.kerberosServerName` | string | jsonData |  | The Kerberos service name to use (optional). Default is 'postgres'. |
| `jsonData.tlsConfigurationMethod` | enum (file-content, file-path) | jsonData |  | This option determines how TLS/SSL certifications are configured. Selecting File system path will allow you to configure certificates by specifying paths to existing certificates on the local file system where Grafana is running. Be sure that the file is readable by the user executing the Grafana process. Selecting Certificate content will allow you to configure certificates by specifying its content. The content will be stored encrypted in Grafana's database. When connecting to the database the certificates will be written as files to Grafana's configured data path on the local file system. |
| `jsonData.sslRootCertFile` | string | jsonData | conditional | If the selected TLS/SSL mode requires a server root certificate, provide the path to the file here. |
| `jsonData.sslCertFile` | string | jsonData | conditional | To authenticate with an TLS/SSL client certificate, provide the path to the file here. Be sure that the file is readable by the user executing the grafana process. |
| `jsonData.sslKeyFile` | string | jsonData | conditional | To authenticate with a client TLS/SSL certificate, provide the path to the corresponding key file here. Be sure that the file is only readable by the user executing the grafana process. |
| `secureJsonData.tlsCACert` 🔒 | string (multiline) | secureJsonData | conditional | If the selected TLS/SSL mode requires a server root certificate, provide it here. |
| `secureJsonData.tlsClientCert` 🔒 | string (multiline) | secureJsonData | conditional | To authenticate with an TLS/SSL client certificate, provide the client certificate here. |
| `secureJsonData.tlsClientKey` 🔒 | string (multiline) | secureJsonData | conditional | To authenticate with a client TLS/SSL certificate, provide the key here. |
| `jsonData.maxOpenConns` | number | jsonData |  | The maximum number of open connections to the database. If Max idle connections is greater than 0 and the Max open connections is less than Max idle connections, then Max idle connections will be reduced to match the Max open connections limit. If set to 0, there is no limit on the number of open connections. |
| `jsonData.maxIdleConnsAuto` | boolean | jsonData |  | If enabled, automatically set the number of Maximum idle connections to the same value as Max open connections. If the number of maximum open connections is not set it will be set to the default. |
| `jsonData.maxIdleConns` | number | jsonData |  | The maximum number of connections in the idle connection pool.If Max open connections is greater than 0 but less than the Max idle connections, then the Max idle connections will be reduced to match the Max open connections limit. If set to 0, no idle connections are retained. |
| `jsonData.connMaxLifetime` | number | jsonData |  | The maximum amount of time in seconds a connection may be reused. If set to 0, connections are reused forever. |
| `jsonData.queryTimeout` | number | jsonData |  | The maximum amount of time in seconds to wait for a query to complete. Valid range: 5-600 seconds (10 minutes). If set to 0, the default of 30 seconds will be used. |
| `jsonData.configFilePath` | string | jsonData | conditional | The path to the configuration file for the [MIT krb5 package](https://web.mit.edu/kerberos/krb5-1.12/doc/admin/conf_files/krb5_conf.html). The default is `/etc/krb5.conf`. |
| `jsonData.sslmode` | enum (disable, require, verify-ca, verify-full) | jsonData |  | This option determines whether or with what priority a secure TLS/SSL TCP/IP connection will be negotiated with the server |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### `SQL Authentication`

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: CockroachDB
    type: grafana-cockroachdb-datasource
    access: proxy
    jsonData:
      authType: SQL Authentication
      connMaxLifetime: 300
      database: defaultdb
      maxIdleConns: 2
      maxIdleConnsAuto: false
      maxOpenConns: 5
      queryTimeout: 30
      sslCertFile: TLS/SSL client cert file
      sslKeyFile: TLS/SSL client key file
      sslRootCertFile: TLS/SSL root cert file
      tlsConfigurationMethod: file-content
      url: "localhost:26257"
      user: User
    secureJsonData:
      password: "<YOUR_PASSWORD>"
      tlsCACert: "<YOUR_TLS_SSL_ROOT_CERTIFICATE>"
      tlsClientCert: "<YOUR_TLS_SSL_CLIENT_CERTIFICATE>"
      tlsClientKey: "<YOUR_TLS_SSL_CLIENT_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_cockroachdb_datasource_SQL_Authentication" {
  type = "grafana-cockroachdb-datasource"
  name = "CockroachDB"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "SQL Authentication"
    connMaxLifetime = 300
    database = "defaultdb"
    maxIdleConns = 2
    maxIdleConnsAuto = false
    maxOpenConns = 5
    queryTimeout = 30
    sslCertFile = "TLS/SSL client cert file"
    sslKeyFile = "TLS/SSL client key file"
    sslRootCertFile = "TLS/SSL root cert file"
    tlsConfigurationMethod = "file-content"
    url = "localhost:26257"
    user = "User"
  })

  secure_json_data_encoded = jsonencode({
    password = "<YOUR_PASSWORD>"
    tlsCACert = "<YOUR_TLS_SSL_ROOT_CERTIFICATE>"
    tlsClientCert = "<YOUR_TLS_SSL_CLIENT_CERTIFICATE>"
    tlsClientKey = "<YOUR_TLS_SSL_CLIENT_KEY>"
  })
}
```

### `Kerberos Authentication`

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: CockroachDB
    type: grafana-cockroachdb-datasource
    access: proxy
    jsonData:
      authType: Kerberos Authentication
      configFilePath: /etc/krb5.conf
      connMaxLifetime: 300
      credentialCache: /tmp/krb5cc_1000
      database: defaultdb
      maxIdleConns: 2
      maxIdleConnsAuto: false
      maxOpenConns: 5
      queryTimeout: 30
      sslCertFile: TLS/SSL client cert file
      sslKeyFile: TLS/SSL client key file
      sslRootCertFile: TLS/SSL root cert file
      tlsConfigurationMethod: file-content
      url: "localhost:26257"
      user: User
    secureJsonData:
      tlsCACert: "<YOUR_TLS_SSL_ROOT_CERTIFICATE>"
      tlsClientCert: "<YOUR_TLS_SSL_CLIENT_CERTIFICATE>"
      tlsClientKey: "<YOUR_TLS_SSL_CLIENT_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_cockroachdb_datasource_Kerberos_Authentication" {
  type = "grafana-cockroachdb-datasource"
  name = "CockroachDB"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "Kerberos Authentication"
    configFilePath = "/etc/krb5.conf"
    connMaxLifetime = 300
    credentialCache = "/tmp/krb5cc_1000"
    database = "defaultdb"
    maxIdleConns = 2
    maxIdleConnsAuto = false
    maxOpenConns = 5
    queryTimeout = 30
    sslCertFile = "TLS/SSL client cert file"
    sslKeyFile = "TLS/SSL client key file"
    sslRootCertFile = "TLS/SSL root cert file"
    tlsConfigurationMethod = "file-content"
    url = "localhost:26257"
    user = "User"
  })

  secure_json_data_encoded = jsonencode({
    tlsCACert = "<YOUR_TLS_SSL_ROOT_CERTIFICATE>"
    tlsClientCert = "<YOUR_TLS_SSL_CLIENT_CERTIFICATE>"
    tlsClientKey = "<YOUR_TLS_SSL_CLIENT_KEY>"
  })
}
```

### `TLS/SSL Authentication`

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: CockroachDB
    type: grafana-cockroachdb-datasource
    access: proxy
    jsonData:
      authType: TLS/SSL Authentication
      connMaxLifetime: 300
      database: defaultdb
      maxIdleConns: 2
      maxIdleConnsAuto: false
      maxOpenConns: 5
      queryTimeout: 30
      sslCertFile: TLS/SSL client cert file
      sslKeyFile: TLS/SSL client key file
      sslRootCertFile: TLS/SSL root cert file
      sslmode: require
      tlsConfigurationMethod: file-content
      url: "localhost:26257"
      user: User
    secureJsonData:
      password: "<YOUR_PASSWORD>"
      tlsCACert: "<YOUR_TLS_SSL_ROOT_CERTIFICATE>"
      tlsClientCert: "<YOUR_TLS_SSL_CLIENT_CERTIFICATE>"
      tlsClientKey: "<YOUR_TLS_SSL_CLIENT_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_cockroachdb_datasource_TLS_SSL_Authentication" {
  type = "grafana-cockroachdb-datasource"
  name = "CockroachDB"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "TLS/SSL Authentication"
    connMaxLifetime = 300
    database = "defaultdb"
    maxIdleConns = 2
    maxIdleConnsAuto = false
    maxOpenConns = 5
    queryTimeout = 30
    sslCertFile = "TLS/SSL client cert file"
    sslKeyFile = "TLS/SSL client key file"
    sslRootCertFile = "TLS/SSL root cert file"
    sslmode = "require"
    tlsConfigurationMethod = "file-content"
    url = "localhost:26257"
    user = "User"
  })

  secure_json_data_encoded = jsonencode({
    password = "<YOUR_PASSWORD>"
    tlsCACert = "<YOUR_TLS_SSL_ROOT_CERTIFICATE>"
    tlsClientCert = "<YOUR_TLS_SSL_CLIENT_CERTIFICATE>"
    tlsClientKey = "<YOUR_TLS_SSL_CLIENT_KEY>"
  })
}
```

