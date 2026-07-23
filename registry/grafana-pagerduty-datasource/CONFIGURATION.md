# PagerDuty configuration

Configuration reference for the **PagerDuty** data source (`grafana-pagerduty-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-pagerduty-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.auth.id` | enum (api_key) | jsonData |  |  |
| `secureJsonData.auth.api_key.apiKey` 🔒 | string | secureJsonData | conditional | PagerDuty REST API Key (prefer generating read-only key) |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### `api_key`

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: PagerDuty
    type: grafana-pagerduty-datasource
    access: proxy
    jsonData:
      auth:
        id: api_key
    secureJsonData:
      auth.api_key.apiKey: "<YOUR_API_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_pagerduty_datasource_api_key" {
  type = "grafana-pagerduty-datasource"
  name = "PagerDuty"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    auth = {
      id = "api_key"
    }
  })

  secure_json_data_encoded = jsonencode({
    "auth.api_key.apiKey" = "<YOUR_API_KEY>"
  })
}
```

