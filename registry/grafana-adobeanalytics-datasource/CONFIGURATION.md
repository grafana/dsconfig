# Adobe Analytics configuration

Configuration reference for the **Adobe Analytics** data source (`grafana-adobeanalytics-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-adobeanalytics-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.variables.global_company_id` | string | jsonData | yes | Refer to [plugin documentation](http://grafana.com/docs/plugins/grafana-adobeanalytics-datasource/#connection) for more information on where to find your Global Company ID in Adobe portal. |
| `jsonData.services.adobe_analytics.auth.id` | enum (oauth2_m2m) | jsonData |  | Authorization flow where application credentials are exchanged for an access token |
| `jsonData.services.adobe_analytics.auth.clientId` | string | jsonData | conditional | Client ID |
| `secureJsonData.adobe_analytics.clientSecret` 🔒 | string | secureJsonData | conditional | Client Secret |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### OAuth server to server authentication (`oauth2_m2m`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Adobe Analytics
    type: grafana-adobeanalytics-datasource
    access: proxy
    jsonData:
      services:
        adobe_analytics:
          auth:
            clientId: "<YOUR_CLIENT_ID>"
            id: oauth2_m2m
      variables:
        global_company_id: "<YOUR_GLOBAL_COMPANY_ID>"
    secureJsonData:
      adobe_analytics.clientSecret: "<YOUR_CLIENT_SECRET>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_adobeanalytics_datasource_oauth2_m2m" {
  type = "grafana-adobeanalytics-datasource"
  name = "Adobe Analytics"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    services = {
      adobe_analytics = {
        auth = {
          clientId = "<YOUR_CLIENT_ID>"
          id = "oauth2_m2m"
        }
      }
    }
    variables = {
      global_company_id = "<YOUR_GLOBAL_COMPANY_ID>"
    }
  })

  secure_json_data_encoded = jsonencode({
    "adobe_analytics.clientSecret" = "<YOUR_CLIENT_SECRET>"
  })
}
```

