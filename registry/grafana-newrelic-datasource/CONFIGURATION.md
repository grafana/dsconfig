# New Relic configuration

Configuration reference for the **New Relic** data source (`grafana-newrelic-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-newrelic-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `secureJsonData.personalApiKey` 🔒 | string | secureJsonData | yes | Used for NRQL queries |
| `secureJsonData.accountId` 🔒 | string | secureJsonData | yes | Your New Relic Account ID |
| `jsonData.region` | enum (EU, US) | jsonData |  | Region hosting your service |
| `jsonData.timeoutInSeconds` | number | jsonData |  | Enter the timeout in seconds. Defaults to 300 |
| `jsonData.restBaseURL` | string | jsonData |  | Backend-only override for the New Relic REST API base URL. Used for internal testing and mocking; not exposed in the configuration editor. |
| `jsonData.infrastructureBaseURL` | string | jsonData |  | Backend-only override for the New Relic Infrastructure API base URL. Used for internal testing and mocking; not exposed in the configuration editor. |
| `jsonData.nerdGraphBaseURL` | string | jsonData |  | Backend-only override for the New Relic NerdGraph (GraphQL) API base URL. Used for internal testing and mocking; not exposed in the configuration editor. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: New Relic
    type: grafana-newrelic-datasource
    access: proxy
    jsonData:
      timeoutInSeconds: 300
    secureJsonData:
      accountId: "<YOUR_ACCOUNT_ID>"
      personalApiKey: "<YOUR_PERSONAL_API_KEY_USER_API_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_newrelic_datasource" {
  type = "grafana-newrelic-datasource"
  name = "New Relic"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    timeoutInSeconds = 300
  })

  secure_json_data_encoded = jsonencode({
    accountId = "<YOUR_ACCOUNT_ID>"
    personalApiKey = "<YOUR_PERSONAL_API_KEY_USER_API_KEY>"
  })
}
```

