# Hello World configuration

Configuration reference for the **Hello World** data source (`grafana-helloworld-datasource`) in Grafana.

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `secureJsonData.apiKey` 🔒 | string | secureJsonData |  | Placeholder secret. The Hello World datasource reads no configuration: its config editor renders static text and its backend ignores instance settings. This key exists only because a dsconfig entry must declare at least one field and the shared conformance suite requires at least one secureJsonData key. The plugin never reads it. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Hello World
    type: grafana-helloworld-datasource
    access: proxy
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_helloworld_datasource" {
  type = "grafana-helloworld-datasource"
  name = "Hello World"
  url = "https://example.com"
}
```

