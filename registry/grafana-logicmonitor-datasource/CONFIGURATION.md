# LogicMonitor Devices configuration

Configuration reference for the **LogicMonitor Devices** data source (`grafana-logicmonitor-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-logicmonitor-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.variables.account_name` | string | jsonData | yes | Your LogicMonitor account name. Example: Use foo for the logic monitor URL https://foo.logicmonitor.com/` |
| `jsonData.services.logicmonitor.auth.id` | enum (auth_bearer) | jsonData |  | Bearer token for LogicMonitor REST API v3. |
| `secureJsonData.logicmonitor.token` 🔒 | string | secureJsonData | conditional | Token for accessing the datasource API |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### API v3 Key (`auth_bearer`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: LogicMonitor Devices
    type: grafana-logicmonitor-datasource
    access: proxy
    jsonData:
      services:
        logicmonitor:
          auth:
            id: auth_bearer
      variables:
        account_name: "<YOUR_ACCOUNT_NAME>"
    secureJsonData:
      logicmonitor.token: "<YOUR_TOKEN>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_logicmonitor_datasource_auth_bearer" {
  type = "grafana-logicmonitor-datasource"
  name = "LogicMonitor Devices"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    services = {
      logicmonitor = {
        auth = {
          id = "auth_bearer"
        }
      }
    }
    variables = {
      account_name = "<YOUR_ACCOUNT_NAME>"
    }
  })

  secure_json_data_encoded = jsonencode({
    "logicmonitor.token" = "<YOUR_TOKEN>"
  })
}
```

