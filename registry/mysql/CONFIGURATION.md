# MySQL configuration

Configuration reference for the **MySQL** data source (`mysql`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/mysql/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `url` | string | root | yes | Host URL |
| `jsonData.database` | string | jsonData |  | Database name |
| `user` | string | root | yes | Username |
| `secureJsonData.password` 🔒 | string | secureJsonData |  | Password |
| `jsonData.tlsAuth` | boolean | jsonData |  | Enables TLS authentication using client cert configured in secure json data. |
| `jsonData.tlsAuthWithCACert` | boolean | jsonData |  | Needed for verifying self-signed TLS Certs. |
| `jsonData.tlsSkipVerify` | boolean | jsonData |  | When enabled, skips verification of the MySQL server's TLS certificate chain and host name. |
| `jsonData.allowCleartextPasswords` | boolean | jsonData |  | Allows using the cleartext client side plugin if required by an account. |
| `secureJsonData.tlsClientCert` 🔒 | string (multiline) | secureJsonData |  | To authenticate with an TLS/SSL client certificate, provide the client certificate here. |
| `secureJsonData.tlsCACert` 🔒 | string (multiline) | secureJsonData |  | If the selected TLS/SSL mode requires a server root certificate, provide it here |
| `secureJsonData.tlsClientKey` 🔒 | string (multiline) | secureJsonData |  | To authenticate with a client TLS/SSL certificate, provide the key here. |
| `jsonData.timezone` | string | jsonData |  | Specify the timezone used in the database session, such as Europe/Berlin or +02:00. Required if the timezone of the database (or the host of the database) is set to something other than UTC. Set this to +00:00 so Grafana can handle times properly. Set the value used in the session with SET time_zone='...'. If you leave this field empty, the timezone will not be updated. You can find more information in the MySQL documentation. |
| `jsonData.timeInterval` | string | jsonData |  | A lower limit for the auto group by time interval. Recommended to be set to write frequency, for example 1m if your data is written every minute. |
| `jsonData.maxOpenConns` | number | jsonData |  | The maximum number of open connections to the database. If Max idle connections is greater than 0 and the Max open connections is less than Max idle connections, then Max idle connections will be reduced to match the Max open connections limit. If set to 0, there is no limit on the number of open connections. |
| `jsonData.maxIdleConnsAuto` | boolean | jsonData |  | If enabled, automatically set the number of Maximum idle connections to the same value as Max open connections. If the number of maximum open connections is not set it will be set to the default. |
| `jsonData.maxIdleConns` | number | jsonData |  | The maximum number of connections in the idle connection pool. If Max open connections is greater than 0 but less than the Max idle connections, then the Max idle connections will be reduced to match the Max open connections limit. If set to 0, no idle connections are retained. |
| `jsonData.connMaxLifetime` | number | jsonData |  | The maximum amount of time in seconds a connection may be reused. If set to 0, connections are reused forever. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: MySQL
    type: mysql
    access: proxy
    url: "localhost:3306"
    user: Username
    jsonData:
      allowCleartextPasswords: false
      maxIdleConnsAuto: true
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsSkipVerify: false
```

**Terraform**

```hcl
resource "grafana_data_source" "mysql" {
  type = "mysql"
  name = "MySQL"
  url = "localhost:3306"
  basic_auth_username = "Username"

  json_data_encoded = jsonencode({
    allowCleartextPasswords = false
    maxIdleConnsAuto = true
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsSkipVerify = false
  })
}
```

