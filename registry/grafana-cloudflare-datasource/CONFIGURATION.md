# Cloudflare configuration

Configuration reference for the **Cloudflare** data source (`grafana-cloudflare-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-cloudflare-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.services.cloudflare.auth.id` | enum (bearer_token) | jsonData |  | Cloudflare API Key. Provide relevant read-only permissions |
| `secureJsonData.cloudflare.token` 🔒 | string | secureJsonData | conditional | Token for accessing the datasource API |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### API Key (`bearer_token`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Cloudflare
    type: grafana-cloudflare-datasource
    access: proxy
    jsonData:
      services:
        cloudflare:
          auth:
            id: bearer_token
    secureJsonData:
      cloudflare.token: "<YOUR_TOKEN>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_cloudflare_datasource_bearer_token" {
  type = "grafana-cloudflare-datasource"
  name = "Cloudflare"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    services = {
      cloudflare = {
        auth = {
          id = "bearer_token"
        }
      }
    }
  })

  secure_json_data_encoded = jsonencode({
    "cloudflare.token" = "<YOUR_TOKEN>"
  })
}
```

