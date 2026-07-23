# Splunk Infrastructure Monitoring configuration

Configuration reference for the **Splunk Infrastructure Monitoring** data source (`grafana-splunk-monitoring-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-splunk-monitoring-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `secureJsonData.accessToken` 🔒 | string | secureJsonData | yes | Access Token |
| `jsonData.realmName` | string | jsonData | conditional | Realm Name |
| `jsonData.url_metrics_metadata` | string | jsonData |  | Optional Metrics MetaData URL. |
| `jsonData.url_signalflow` | string | jsonData |  | Optional SignalFlow URL |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Splunk Infrastructure Monitoring
    type: grafana-splunk-monitoring-datasource
    access: proxy
    jsonData:
      realmName: us1
    secureJsonData:
      accessToken: "<YOUR_ACCESS_TOKEN>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_splunk_monitoring_datasource" {
  type = "grafana-splunk-monitoring-datasource"
  name = "Splunk Infrastructure Monitoring"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    realmName = "us1"
  })

  secure_json_data_encoded = jsonencode({
    accessToken = "<YOUR_ACCESS_TOKEN>"
  })
}
```

