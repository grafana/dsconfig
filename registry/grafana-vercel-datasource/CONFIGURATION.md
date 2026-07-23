# Vercel configuration

Configuration reference for the **Vercel** data source (`grafana-vercel-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-vercel-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.variables.team_id` | string | jsonData |  | The ID of the Vercel team to query. |
| `jsonData.services.vercel.auth.id` | enum (vercelApiKey) | jsonData |  | Vercel Access Tokens are required to authenticate and use the Vercel API. Tokens are either scoped to your full account or a specific team. If a token is scoped to a team, you must also provide a team ID that matches the scope of the token. |
| `secureJsonData.vercel.token` 🔒 | string | secureJsonData | conditional | Token for accessing the datasource API |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Access Token (`vercelApiKey`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Vercel
    type: grafana-vercel-datasource
    access: proxy
    jsonData:
      services:
        vercel:
          auth:
            id: vercelApiKey
    secureJsonData:
      vercel.token: "<YOUR_TOKEN>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_vercel_datasource_vercelApiKey" {
  type = "grafana-vercel-datasource"
  name = "Vercel"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    services = {
      vercel = {
        auth = {
          id = "vercelApiKey"
        }
      }
    }
  })

  secure_json_data_encoded = jsonencode({
    "vercel.token" = "<YOUR_TOKEN>"
  })
}
```

