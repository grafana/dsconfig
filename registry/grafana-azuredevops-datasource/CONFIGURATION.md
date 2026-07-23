# Azure DevOps configuration

Configuration reference for the **Azure DevOps** data source (`grafana-azuredevops-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-azuredevops-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.url` | string | jsonData | yes | Azure DevOps instance URL |
| `jsonData.authType` | enum (patToken) | jsonData |  |  |
| `secureJsonData.patToken` 🔒 | string | secureJsonData | yes | Azure DevOps personal access token |
| `jsonData.projectsLimit` | number | jsonData |  | Number of items to retrieve in projects list query |
| `jsonData.username` | string | jsonData |  | Username of the user that owns the Azure DevOps PAT. May be needed for some versions of Azure DevOps Server. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### `patToken`

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Azure DevOps
    type: grafana-azuredevops-datasource
    access: proxy
    jsonData:
      authType: patToken
      projectsLimit: 100
      url: "https://dev.azure.com/XXXX"
    secureJsonData:
      patToken: "<YOUR_PAT>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_azuredevops_datasource_patToken" {
  type = "grafana-azuredevops-datasource"
  name = "Azure DevOps"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "patToken"
    projectsLimit = 100
    url = "https://dev.azure.com/XXXX"
  })

  secure_json_data_encoded = jsonencode({
    patToken = "<YOUR_PAT>"
  })
}
```

