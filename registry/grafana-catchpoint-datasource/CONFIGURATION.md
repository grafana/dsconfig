# Catchpoint configuration

Configuration reference for the **Catchpoint** data source (`grafana-catchpoint-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-catchpoint-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.services.catchpoint.auth.id` | enum (bearer_token) | jsonData |  | Catchpoint REST API v2 Key. |
| `secureJsonData.catchpoint.token` 🔒 | string | secureJsonData | conditional | Token for accessing the datasource API |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### API v2 Key (`bearer_token`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Catchpoint
    type: grafana-catchpoint-datasource
    access: proxy
    jsonData:
      services:
        catchpoint:
          auth:
            id: bearer_token
    secureJsonData:
      catchpoint.token: "<YOUR_TOKEN>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_catchpoint_datasource_bearer_token" {
  type = "grafana-catchpoint-datasource"
  name = "Catchpoint"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    services = {
      catchpoint = {
        auth = {
          id = "bearer_token"
        }
      }
    }
  })

  secure_json_data_encoded = jsonencode({
    "catchpoint.token" = "<YOUR_TOKEN>"
  })
}
```

