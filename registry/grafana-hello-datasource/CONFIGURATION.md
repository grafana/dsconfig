# Hello configuration

Configuration reference for the **Hello** data source (`grafana-hello-datasource`) in Grafana.

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.services.httpbin.auth.id` | enum (none) | jsonData |  | No Auth |
| `jsonData.services.postman_echo.auth.id` | enum (none) | jsonData |  | No Auth |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Hello
    type: grafana-hello-datasource
    access: proxy
    jsonData:
      services:
        httpbin:
          auth:
            id: none
        postman_echo:
          auth:
            id: none
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_hello_datasource" {
  type = "grafana-hello-datasource"
  name = "Hello"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    services = {
      httpbin = {
        auth = {
          id = "none"
        }
      }
      postman_echo = {
        auth = {
          id = "none"
        }
      }
    }
  })
}
```

