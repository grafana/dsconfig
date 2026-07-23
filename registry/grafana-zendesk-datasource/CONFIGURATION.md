# Zendesk configuration

Configuration reference for the **Zendesk** data source (`grafana-zendesk-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-zendesk-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.variables.subdomain` | string | jsonData | yes | If your Zendesk URL is "grafana.zendesk.com", the subdomain would be "grafana". |
| `jsonData.services.zendesk.auth.id` | enum (basic_auth) | jsonData |  | Identifier of the selected authentication method for the Tickets service. The Zendesk API server exposes a single method, `basic_auth`; the backend defaults to it when unset. |
| `jsonData.services.zendesk.auth.username` | string | jsonData | conditional | Email address used to login to Zendesk |
| `secureJsonData.zendesk.password` 🔒 | string | secureJsonData | conditional | API Token generated from Zendesk. Visit the [docs](https://support.zendesk.com/hc/en-us/articles/4408889192858-Managing-access-to-the-Zendesk-API) to learn how |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Basic Auth (`basic_auth`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Zendesk
    type: grafana-zendesk-datasource
    access: proxy
    jsonData:
      services:
        zendesk:
          auth:
            id: basic_auth
            username: Email
      variables:
        subdomain: "<YOUR_SUBDOMAIN>"
    secureJsonData:
      zendesk.password: "<YOUR_API_TOKEN>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_zendesk_datasource_basic_auth" {
  type = "grafana-zendesk-datasource"
  name = "Zendesk"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    services = {
      zendesk = {
        auth = {
          id = "basic_auth"
          username = "Email"
        }
      }
    }
    variables = {
      subdomain = "<YOUR_SUBDOMAIN>"
    }
  })

  secure_json_data_encoded = jsonencode({
    "zendesk.password" = "<YOUR_API_TOKEN>"
  })
}
```

