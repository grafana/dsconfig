# Jenkins configuration

Configuration reference for the **Jenkins** data source (`grafana-jenkins-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-jenkins-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.url` | string | jsonData | yes | Jenkins URL, e.g. https://jenkins.example.com |
| `jsonData.username` | string | jsonData |  | The username to use for authentication |
| `secureJsonData.password` 🔒 | string | secureJsonData |  | The password to use for authentication |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Jenkins
    type: grafana-jenkins-datasource
    access: proxy
    jsonData:
      url: "Jenkins URL, e.g. https://jenkins.example.com"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_jenkins_datasource" {
  type = "grafana-jenkins-datasource"
  name = "Jenkins"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    url = "Jenkins URL, e.g. https://jenkins.example.com"
  })
}
```

