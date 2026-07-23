# Datadog configuration

Configuration reference for the **Datadog** data source (`grafana-datadog-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-datadog-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.pluginMode` | enum (default, hosted-metrics) | jsonData |  |  |
| `basicAuth` | boolean | root |  |  |
| `jsonData.url` | string | jsonData | yes | A URL to the Datadog API (e.g.: https://api.datadoghq.com) |
| `secureJsonData.apiKey` 🔒 | string | secureJsonData | conditional | An API key is unique to your organization. [Learn more](https://grafana.com/docs/plugins/grafana-datadog-datasource/latest/#get-an-api-key-and-application-key-from-datadog). |
| `secureJsonData.appKey` 🔒 | string | secureJsonData | conditional | An application key is used with the API key to give access to the Datadog API. By default, application keys have the permissions of the user who created them. [Learn more](https://grafana.com/docs/plugins/grafana-datadog-datasource/latest/#get-api-key-and-application-key-from-datadog). You can also customize the scope of the application key in the [Datadog docs](https://docs.datadoghq.com/api/latest/scopes/). |
| `basicAuthUser` | string | root | conditional | Your username is your Grafana Cloud Prometheus username. This can be found in the Prometheus details in your cloud portal. |
| `secureJsonData.basicAuthPassword` 🔒 | string | secureJsonData | conditional | Your password is your Grafana Cloud API Key with read permissions. This can be found in the Prometheus details in your cloud portal. |
| `jsonData.logApiRateLimits` | boolean | jsonData |  | Show Datadog API limits for each queried endpoint. To view the API rate limits, go to the **Query Inspector**, select **JSON**, and set **select source** to **DataFrame structure**. |
| `jsonData.rateLimitEnabled` | boolean | jsonData |  | Enable rate limit. Datadog query will stop once it reaches entered threshold. |
| `jsonData.rateLimitMetrics` | number | jsonData |  | Enter percentage of threshold. (If the API hit the % of rate limit, plugin will block subsequent requests till next reset) |
| `jsonData.disableDataLinks` | boolean | jsonData |  | Data links take users directly to the relevant location in the Datadog app when they interact with panels. |
| `jsonData.size` | number | jsonData |  | Set maximum number of items to retrieve in a single API request (default is 100). |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### `default`

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Datadog
    type: grafana-datadog-datasource
    access: proxy
    basicAuth: false
    jsonData:
      disableDataLinks: false
      logApiRateLimits: false
      pluginMode: default
      rateLimitEnabled: false
      rateLimitMetrics: 100
      size: 100
      url: "https://api.datadoghq.com"
    secureJsonData:
      apiKey: "<YOUR_API_KEY>"
      appKey: "<YOUR_APP_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_datadog_datasource_default" {
  type = "grafana-datadog-datasource"
  name = "Datadog"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    disableDataLinks = false
    logApiRateLimits = false
    pluginMode = "default"
    rateLimitEnabled = false
    rateLimitMetrics = 100
    size = 100
    url = "https://api.datadoghq.com"
  })

  secure_json_data_encoded = jsonencode({
    apiKey = "<YOUR_API_KEY>"
    appKey = "<YOUR_APP_KEY>"
  })
}
```

### `hosted-metrics`

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Datadog
    type: grafana-datadog-datasource
    access: proxy
    basicAuth: false
    basicAuthUser: User
    jsonData:
      disableDataLinks: false
      logApiRateLimits: false
      pluginMode: hosted-metrics
      rateLimitEnabled: false
      rateLimitMetrics: 100
      size: 100
      url: "https://api.datadoghq.com"
    secureJsonData:
      apiKey: "<YOUR_API_KEY>"
      appKey: "<YOUR_APP_KEY>"
      basicAuthPassword: "<YOUR_PASSWORD>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_datadog_datasource_hosted_metrics" {
  type = "grafana-datadog-datasource"
  name = "Datadog"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    disableDataLinks = false
    logApiRateLimits = false
    pluginMode = "hosted-metrics"
    rateLimitEnabled = false
    rateLimitMetrics = 100
    size = 100
    url = "https://api.datadoghq.com"
  })

  secure_json_data_encoded = jsonencode({
    apiKey = "<YOUR_API_KEY>"
    appKey = "<YOUR_APP_KEY>"
    basicAuthPassword = "<YOUR_PASSWORD>"
  })
}
```

