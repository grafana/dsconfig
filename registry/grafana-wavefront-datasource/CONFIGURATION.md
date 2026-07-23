# Wavefront configuration

Configuration reference for the **Wavefront** data source (`grafana-wavefront-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-wavefront-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.url` | string | jsonData | yes | URL to Wavefront API |
| `secureJsonData.token` 🔒 | string | secureJsonData | yes | Wavefront token |
| `jsonData.requestTimeout` | number | jsonData |  | Request timeout in seconds. Defaults to 30 |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Wavefront
    type: grafana-wavefront-datasource
    access: proxy
    jsonData:
      requestTimeout: 30
      url: "https://try.wavefront.com"
    secureJsonData:
      token: "<YOUR_TOKEN>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_wavefront_datasource" {
  type = "grafana-wavefront-datasource"
  name = "Wavefront"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    requestTimeout = 30
    url = "https://try.wavefront.com"
  })

  secure_json_data_encoded = jsonencode({
    token = "<YOUR_TOKEN>"
  })
}
```

