# Dynatrace configuration

Configuration reference for the **Dynatrace** data source (`grafana-dynatrace-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-dynatrace-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.apiType` | enum (saas, managed, url) | jsonData |  | Dynatrace API Type |
| `jsonData.environmentId` | string | jsonData | yes | Get environment ID from your instance URL: [environmentId].live.dynatrace.com |
| `jsonData.domain` | string | jsonData | conditional | Domain |
| `secureJsonData.apiToken` 🔒 | string | secureJsonData | conditional | The API token for the Dynatrace API. This is required for Older api endpoints on Dynatrace like Metrics, Problems, Logs, etc. |
| `secureJsonData.platformToken` 🔒 | string | secureJsonData | conditional | The Platform token for the Dynatrace Platform API. This is required for Newer api endpoints on Dynatrace like Grail |
| `jsonData.httpClientTimeout` | number | jsonData |  | The timeout for the HTTP client in seconds. Default is 30 seconds. |
| `jsonData.tlsSkipVerify` | boolean | jsonData |  | Skip TLS Verify |
| `jsonData.tlsAuthWithCACert` | boolean | jsonData |  | Needed for verifying self-signed TLS Certificates |
| `secureJsonData.tlsCACert` 🔒 | string (multiline) | secureJsonData | conditional | TLS/SSL Certs are encrypted and stored in the Grafana database. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Dynatrace
    type: grafana-dynatrace-datasource
    access: proxy
    jsonData:
      apiType: saas
      domain: Your Domain
      environmentId: Your Environment ID
      httpClientTimeout: 30
      tlsAuthWithCACert: false
      tlsSkipVerify: false
    secureJsonData:
      apiToken: "<YOUR_DYNATRACE_API_TOKEN>"
      platformToken: "<YOUR_DYNATRACE_PLATFORM_TOKEN>"
      tlsCACert: "<YOUR_CA_CERT>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_dynatrace_datasource" {
  type = "grafana-dynatrace-datasource"
  name = "Dynatrace"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    apiType = "saas"
    domain = "Your Domain"
    environmentId = "Your Environment ID"
    httpClientTimeout = 30
    tlsAuthWithCACert = false
    tlsSkipVerify = false
  })

  secure_json_data_encoded = jsonencode({
    apiToken = "<YOUR_DYNATRACE_API_TOKEN>"
    platformToken = "<YOUR_DYNATRACE_PLATFORM_TOKEN>"
    tlsCACert = "<YOUR_CA_CERT>"
  })
}
```

