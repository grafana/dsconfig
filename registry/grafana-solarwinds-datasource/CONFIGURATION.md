# Solarwinds configuration

Configuration reference for the **Solarwinds** data source (`grafana-solarwinds-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-solarwinds-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.variables.url` | string | jsonData | yes | The URL of your SolarWinds instance, including `https://` and without trailing `/` |
| `jsonData.services.solarwinds.auth.id` | enum (basic_auth) | jsonData |  | Basic Auth |
| `jsonData.services.solarwinds.auth.username` | string | jsonData | conditional | SolarWinds Username |
| `secureJsonData.solarwinds.password` 🔒 | string | secureJsonData | conditional | SolarWinds Password |
| `jsonData.services.solarwinds.auth.tls.selfSignedCert.enabled` | boolean | jsonData |  | Add self-signed certificate |
| `secureJsonData.solarwinds.tls.selfSignedCert` 🔒 | string (multiline) | secureJsonData |  | Your self-signed certificate |
| `jsonData.services.solarwinds.auth.tls.clientAuth.enabled` | boolean | jsonData |  | Validate using TLS client authentication, in which the server authenticates the client |
| `jsonData.services.solarwinds.auth.tls.clientAuth.serverName` | string | jsonData |  | A Servername is used to verify the hostname on the returned certificate |
| `secureJsonData.solarwinds.tls.clientCert` 🔒 | string (multiline) | secureJsonData |  | The client certificate can be generated from a Certificate Authority or be self-signed |
| `secureJsonData.solarwinds.tls.clientKey` 🔒 | string (multiline) | secureJsonData |  | The client key can be generated from a Certificate Authority or be self-signed |
| `jsonData.services.solarwinds.auth.tls.skipVerification` | boolean | jsonData |  | Skip TLS certificate validation |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Basic Auth (`basic_auth`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Solarwinds
    type: grafana-solarwinds-datasource
    access: proxy
    jsonData:
      services:
        solarwinds:
          auth:
            id: basic_auth
            username: Username
      variables:
        url: "<YOUR_URL>"
    secureJsonData:
      solarwinds.password: "<YOUR_PASSWORD>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_solarwinds_datasource_basic_auth" {
  type = "grafana-solarwinds-datasource"
  name = "Solarwinds"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    services = {
      solarwinds = {
        auth = {
          id = "basic_auth"
          username = "Username"
        }
      }
    }
    variables = {
      url = "<YOUR_URL>"
    }
  })

  secure_json_data_encoded = jsonencode({
    "solarwinds.password" = "<YOUR_PASSWORD>"
  })
}
```

