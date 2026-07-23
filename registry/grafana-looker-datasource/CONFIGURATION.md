# Looker configuration

Configuration reference for the **Looker** data source (`grafana-looker-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-looker-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.base_url` | string | jsonData | yes | Looker base URL. Example: https://00001234-1234-1ab2-1234-a1b2c3d4.looker.app |
| `jsonData.auth_type` | enum (client_secret) | jsonData |  | Looker authentication type |
| `jsonData.client_id` | string | jsonData | conditional | API credentials Looker client id |
| `secureJsonData.client_secret` 🔒 | string | secureJsonData | conditional | API credentials Looker client secret |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Client Secret (`client_secret`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Looker
    type: grafana-looker-datasource
    access: proxy
    jsonData:
      auth_type: client_secret
      base_url: "https://xxxxx.looker.app"
      client_id: Client ID
    secureJsonData:
      client_secret: "<YOUR_LOOKER_CLIENT_SECRET>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_looker_datasource_client_secret" {
  type = "grafana-looker-datasource"
  name = "Looker"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    auth_type = "client_secret"
    base_url = "https://xxxxx.looker.app"
    client_id = "Client ID"
  })

  secure_json_data_encoded = jsonencode({
    client_secret = "<YOUR_LOOKER_CLIENT_SECRET>"
  })
}
```

