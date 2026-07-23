# Azure Cosmos DB configuration

Configuration reference for the **Azure Cosmos DB** data source (`grafana-azurecosmosdb-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-azurecosmosdb-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.accountEndpoint` | string | jsonData | yes | Account Endpoint |
| `secureJsonData.accountKey` 🔒 | string | secureJsonData | yes | Account Key |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Azure Cosmos DB
    type: grafana-azurecosmosdb-datasource
    access: proxy
    jsonData:
      accountEndpoint: "<YOUR_ACCOUNT_ENDPOINT>"
    secureJsonData:
      accountKey: "<YOUR_ACCOUNT_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_azurecosmosdb_datasource" {
  type = "grafana-azurecosmosdb-datasource"
  name = "Azure Cosmos DB"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    accountEndpoint = "<YOUR_ACCOUNT_ENDPOINT>"
  })

  secure_json_data_encoded = jsonencode({
    accountKey = "<YOUR_ACCOUNT_KEY>"
  })
}
```

