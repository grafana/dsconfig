# Netlify configuration

Configuration reference for the **Netlify** data source (`grafana-netlify-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-netlify-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.services.Netlify.auth.id` | enum (bearer_token) | jsonData |  | Netlify REST API Key. Found here: https://app.netlify.com/user/applications#personal-access-tokens |
| `secureJsonData.Netlify.token` 🔒 | string | secureJsonData | conditional | Token for accessing the datasource API |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Personal access tokens (`bearer_token`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Netlify
    type: grafana-netlify-datasource
    access: proxy
    jsonData:
      services:
        Netlify:
          auth:
            id: bearer_token
    secureJsonData:
      Netlify.token: "<YOUR_TOKEN>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_netlify_datasource_bearer_token" {
  type = "grafana-netlify-datasource"
  name = "Netlify"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    services = {
      Netlify = {
        auth = {
          id = "bearer_token"
        }
      }
    }
  })

  secure_json_data_encoded = jsonencode({
    "Netlify.token" = "<YOUR_TOKEN>"
  })
}
```

