# Yugabyte configuration

Configuration reference for the **Yugabyte** data source (`grafana-yugabyte-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/yugabyte/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `url` | string | root | yes | Host URL |
| `jsonData.database` | string | jsonData | yes | Database |
| `user` | string | root | yes | Username |
| `secureJsonData.password` 🔒 | string | secureJsonData |  | Password |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Yugabyte
    type: grafana-yugabyte-datasource
    access: proxy
    url: "localhost:5433"
    user: yugabyte
    jsonData:
      database: yb_demo
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_yugabyte_datasource" {
  type = "grafana-yugabyte-datasource"
  name = "Yugabyte"
  url = "localhost:5433"
  basic_auth_username = "yugabyte"

  json_data_encoded = jsonencode({
    database = "yb_demo"
  })
}
```

