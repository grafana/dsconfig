# Drone configuration

Configuration reference for the **Drone** data source (`grafana-drone-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-drone-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.variables.url` | string | jsonData | yes | The URL of the Drone server, including `https://` and without trailing `/` |
| `jsonData.services.drone.auth.id` | enum (auth_bearer) | jsonData |  | You can find your API token under <YOUR_DRONE_URL>/account |
| `secureJsonData.drone.token` 🔒 | string | secureJsonData | conditional | Token for accessing the datasource API |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Drone API token (`auth_bearer`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Drone
    type: grafana-drone-datasource
    access: proxy
    jsonData:
      services:
        drone:
          auth:
            id: auth_bearer
      variables:
        url: "<YOUR_URL>"
    secureJsonData:
      drone.token: "<YOUR_TOKEN>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_drone_datasource_auth_bearer" {
  type = "grafana-drone-datasource"
  name = "Drone"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    services = {
      drone = {
        auth = {
          id = "auth_bearer"
        }
      }
    }
    variables = {
      url = "<YOUR_URL>"
    }
  })

  secure_json_data_encoded = jsonencode({
    "drone.token" = "<YOUR_TOKEN>"
  })
}
```

