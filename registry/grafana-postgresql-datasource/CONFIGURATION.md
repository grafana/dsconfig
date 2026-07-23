# PostgreSQL configuration

Configuration reference for the **PostgreSQL** data source (`grafana-postgresql-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/postgres/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `url` | string | root | yes | Host URL |
| `jsonData.database` | string | jsonData | yes | Database name |
| `user` | string | root | yes | Username |
| `secureJsonData.password` 🔒 | string | secureJsonData |  | Password |
| `jsonData.sslmode` | enum (disable, require, verify-ca, verify-full) | jsonData |  | This option determines whether or with what priority a secure TLS/SSL TCP/IP connection will be negotiated with the server. |
| `jsonData.tlsConfigurationMethod` | enum (file-path, file-content) | jsonData |  | This option determines how TLS/SSL certifications are configured. Selecting 'File system path' will allow you to configure certificates by specifying paths to existing certificates on the local file system where Grafana is running. Be sure that the file is readable by the user executing the Grafana process. Selecting 'Certificate content' will allow you to configure certificates by specifying its content. The content will be stored encrypted in Grafana's database. When connecting to the database the certificates will be written as files to Grafana's configured data path on the local file system. |
| `jsonData.sslRootCertFile` | string | jsonData |  | If the selected TLS/SSL mode requires a server root certificate, provide the path to the file here. |
| `jsonData.sslCertFile` | string | jsonData |  | To authenticate with an TLS/SSL client certificate, provide the path to the file here. Be sure that the file is readable by the user executing the grafana process. |
| `jsonData.sslKeyFile` | string | jsonData |  | To authenticate with a client TLS/SSL certificate, provide the path to the corresponding key file here. Be sure that the file is only readable by the user executing the grafana process. |
| `secureJsonData.tlsCACert` 🔒 | string (multiline) | secureJsonData |  | If the selected TLS/SSL mode requires a server root certificate, provide it here |
| `secureJsonData.tlsClientCert` 🔒 | string (multiline) | secureJsonData |  | To authenticate with an TLS/SSL client certificate, provide the client certificate here. |
| `secureJsonData.tlsClientKey` 🔒 | string (multiline) | secureJsonData |  | To authenticate with a client TLS/SSL certificate, provide the key here. |
| `jsonData.postgresVersion` | enum (900, 901, 902, 903, 904, 905, 906, 1000, 1100, 1200, 1300, 1400, 1500) | jsonData |  | This option controls what functions are available in the PostgreSQL query builder. |
| `jsonData.timeInterval` | string | jsonData |  | A lower limit for the auto group by time interval. Recommended to be set to write frequency, for example 1m if your data is written every minute. |
| `jsonData.timescaledb` | boolean | jsonData |  | TimescaleDB is a time-series database built as a PostgreSQL extension. If enabled, Grafana will use time_bucket in the $__timeGroup macro and display TimescaleDB specific aggregate functions in the query builder. |
| `jsonData.maxOpenConns` | number | jsonData |  | The maximum number of open connections to the database. If set to 0, there is no limit on the number of open connections. |
| `jsonData.connMaxLifetime` | number | jsonData |  | The maximum amount of time in seconds a connection may be reused. If set to 0, connections are reused forever. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: PostgreSQL
    type: grafana-postgresql-datasource
    access: proxy
    url: "localhost:5432"
    user: Username
    jsonData:
      database: Database
      postgresVersion: 903
      sslmode: require
      timescaledb: false
      tlsConfigurationMethod: file-path
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_postgresql_datasource" {
  type = "grafana-postgresql-datasource"
  name = "PostgreSQL"
  url = "localhost:5432"
  basic_auth_username = "Username"

  json_data_encoded = jsonencode({
    database = "Database"
    postgresVersion = 903
    sslmode = "require"
    timescaledb = false
    tlsConfigurationMethod = "file-path"
  })
}
```

