# Honeycomb configuration

Configuration reference for the **Honeycomb** data source (`grafana-honeycomb-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-honeycomb-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `secureJsonData.apiKey` 🔒 | string | secureJsonData | yes | Honeycomb API Key |
| `jsonData.hostname` | string | jsonData | yes | Customize the api URL. By default this will be https://api.honeycomb.io |
| `jsonData.team` | string | jsonData | yes | Specify the team name. This will be useful in data links |
| `jsonData.environment` | string | jsonData |  | Optional. Specify the environment name. This will be useful in data links |
| `jsonData.retentionLimit` | number | jsonData |  | Optional. The maximum time window, in days. Queries will only return data from the last N days, where N is the retention limit. Default is 7 days, since that is the maximum retention limit normally supported by the Honeycomb API. Honeycomb API docs: https://api-docs.honeycomb.io/api/query-data |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Honeycomb
    type: grafana-honeycomb-datasource
    access: proxy
    jsonData:
      hostname: "https://api.honeycomb.io"
      retentionLimit: 7
      team: "<YOUR_TEAM_NAME>"
    secureJsonData:
      apiKey: "<YOUR_HONEYCOMB_API_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_honeycomb_datasource" {
  type = "grafana-honeycomb-datasource"
  name = "Honeycomb"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    hostname = "https://api.honeycomb.io"
    retentionLimit = 7
    team = "<YOUR_TEAM_NAME>"
  })

  secure_json_data_encoded = jsonencode({
    apiKey = "<YOUR_HONEYCOMB_API_KEY>"
  })
}
```

