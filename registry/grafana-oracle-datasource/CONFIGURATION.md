# Oracle Database configuration

Configuration reference for the **Oracle Database** data source (`grafana-oracle-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-oracle-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.useTNSNamesBasedConnection` | boolean | jsonData |  | Connection-method discriminator: false selects 'Host with TCP Port' (host + port + database), true selects 'TNSNames Entry' (a tnsnames.ora entry). Written by the config editor's Connection methods selector. |
| `url` | string | root | conditional | Hostname or IP Address with TCP port number. |
| `jsonData.database` | string | jsonData | conditional | Name of database |
| `jsonData.tnsNamesEntry` | string | jsonData | conditional | Use connection config specified in tnsnames.ora |
| `jsonData.useKerberosAuthentication` | boolean | jsonData |  | Authentication discriminator: false selects Basic Authentication (username + password), true selects Kerberos Authentication. The config editor only exposes Kerberos when useTNSNamesBasedConnection is true, though the backend also accepts Kerberos with a host + port connection. |
| `jsonData.user` | string | jsonData | conditional | An Oracle username with access to the specified database. |
| `secureJsonData.password` 🔒 | string | secureJsonData | conditional | An Oracle password with access to the specified database. |
| `jsonData.timezone_name` | enum | jsonData |  | Choose the default timezone of the Oracle server. Typically this is UTC. |
| `jsonData.connectionPoolSize` | number | jsonData |  | Choose the maximum number of connections in the connection pool. Defaults to 50. Takes precedence over environment variable 'GF_PLUGINS_ORACLE_DATASOURCE_POOLSIZE' if set. |
| `jsonData.dataProxyTimeout` | number | jsonData |  | Choose the maximum time in seconds to wait for a response. Defaults to 120 seconds. Takes precedence over environment variable 'GF_DATAPROXY_TIMEOUT' if set. |
| `jsonData.prefetchRowsCount` | number | jsonData |  | Row Prefetching allow you to set the number of rows to prefetch into the client while a result set is being populated during a query. This feature reduces the number of round trips to the server. |
| `jsonData.rowLimit` | number | jsonData |  | Maximum number of rows returned per query. Defaults to 1,000,000. |
| `jsonData.use_dst` | boolean | jsonData |  | Parsed by the backend into DBConnectionOptions.DSTEnabled (pkg/models/settings.go:23) but never read when building the connection string. Declared in the frontend OracleOptions type (src/types.ts:9) but not written by the config editor — use_dst is a per-query/annotation option (src/types.ts:30, src/datasource.ts:109), not a datasource-level setting. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Oracle Database
    type: grafana-oracle-datasource
    access: proxy
    url: Host
    jsonData:
      connectionPoolSize: 50
      dataProxyTimeout: 120
      database: Database
      rowLimit: 1000000
      timezone_name: UTC
      tnsNamesEntry: server/DB
      user: User
    secureJsonData:
      password: "<YOUR_PASSWORD>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_oracle_datasource" {
  type = "grafana-oracle-datasource"
  name = "Oracle Database"
  url = "Host"

  json_data_encoded = jsonencode({
    connectionPoolSize = 50
    dataProxyTimeout = 120
    database = "Database"
    rowLimit = 1000000
    timezone_name = "UTC"
    tnsNamesEntry = "server/DB"
    user = "User"
  })

  secure_json_data_encoded = jsonencode({
    password = "<YOUR_PASSWORD>"
  })
}
```

