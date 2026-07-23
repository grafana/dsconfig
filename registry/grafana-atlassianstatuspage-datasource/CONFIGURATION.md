# Atlassian Statuspage configuration

Configuration reference for the **Atlassian Statuspage** data source (`grafana-atlassianstatuspage-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-atlassianstatuspage-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.variables.url` | string | jsonData | yes | The URL of the Atlassian Statuspage, including `https://` and without trailing `/` |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Atlassian Statuspage
    type: grafana-atlassianstatuspage-datasource
    access: proxy
    jsonData:
      variables:
        url: "<YOUR_URL>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_atlassianstatuspage_datasource" {
  type = "grafana-atlassianstatuspage-datasource"
  name = "Atlassian Statuspage"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    variables = {
      url = "<YOUR_URL>"
    }
  })
}
```

