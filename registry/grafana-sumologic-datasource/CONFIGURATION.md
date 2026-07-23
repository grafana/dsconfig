# Sumo Logic configuration

Configuration reference for the **Sumo Logic** data source (`grafana-sumologic-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-sumologic-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.apiUrl` | string | jsonData | yes | SumoLogic API URL. [Read how to find your deployment.](https://help.sumologic.com/docs/api/getting-started/#which-endpoint-should-i-should-use) |
| `jsonData.timeout` | number | jsonData |  | Timeout in seconds for the data requests |
| `jsonData.interval` | number | jsonData |  | Interval in milliseconds for the log polling requests. Min value is 200. |
| `jsonData.authMethod` | enum (accessKey) | jsonData |  | Authentication method discriminator. The only supported value is 'accessKey' (HTTP basic auth with an access ID + access key). The configuration editor never writes this key — it renders a single fixed authentication method whose selector handler is a no-op — and the backend defaults it to 'accessKey' when empty and rejects any other value. |
| `jsonData.accessId` | string | jsonData | conditional | Sumo Logic Access Id |
| `secureJsonData.accessKey` 🔒 | string | secureJsonData | conditional | Sumo Logic Access Key |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### `accessKey`

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Sumo Logic
    type: grafana-sumologic-datasource
    access: proxy
    jsonData:
      accessId: Sumo Logic Access Id
      apiUrl: "https://api.sumologic.com/api/"
      authMethod: accessKey
      interval: 1000
      timeout: 30
    secureJsonData:
      accessKey: "<YOUR_ACCESSKEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_sumologic_datasource_accessKey" {
  type = "grafana-sumologic-datasource"
  name = "Sumo Logic"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    accessId = "Sumo Logic Access Id"
    apiUrl = "https://api.sumologic.com/api/"
    authMethod = "accessKey"
    interval = 1000
    timeout = 30
  })

  secure_json_data_encoded = jsonencode({
    accessKey = "<YOUR_ACCESSKEY>"
  })
}
```

