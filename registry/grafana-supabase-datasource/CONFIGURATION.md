# Supabase configuration

Configuration reference for the **Supabase** data source (`grafana-supabase-datasource`) in Grafana.

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.services.mgmt.auth.id` | enum (mgmt_bearer) | jsonData |  | Supabase personal token |
| `secureJsonData.mgmt.token` 🔒 | string | secureJsonData | conditional | Token for accessing the datasource API |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Supabase personal token (`mgmt_bearer`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Supabase
    type: grafana-supabase-datasource
    access: proxy
    jsonData:
      services:
        mgmt:
          auth:
            id: mgmt_bearer
    secureJsonData:
      mgmt.token: "<YOUR_TOKEN>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_supabase_datasource_mgmt_bearer" {
  type = "grafana-supabase-datasource"
  name = "Supabase"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    services = {
      mgmt = {
        auth = {
          id = "mgmt_bearer"
        }
      }
    }
  })

  secure_json_data_encoded = jsonencode({
    "mgmt.token" = "<YOUR_TOKEN>"
  })
}
```

