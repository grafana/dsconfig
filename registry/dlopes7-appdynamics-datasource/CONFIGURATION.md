# AppDynamics configuration

Configuration reference for the **AppDynamics** data source (`dlopes7-appdynamics-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/dlopes7-appdynamics-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `url` | string | root | yes | URL |
| `basicAuth` | boolean | root |  | Basic auth |
| `basicAuthUser` | string | root | conditional | User |
| `secureJsonData.basicAuthPassword` 🔒 | string | secureJsonData | conditional | Password |
| `jsonData.clientName` | string | jsonData | conditional | Client Name |
| `jsonData.clientDomain` | string | jsonData | conditional | Client Domain |
| `secureJsonData.clientSecret` 🔒 | string | secureJsonData | conditional | Authenticate to AppDynamics using an API key. This will override username/password (basic) authenticationLeave blank for username/password authentication. |
| `jsonData.httpHeaders` | list | jsonData |  | Additional HTTP headers sent with every request. Header names are stored in jsonData (httpHeaderName<N>); header values are write-only in secureJsonData (httpHeaderValue<N>). |
| `jsonData.httpHeaders[].name` | string | jsonData | yes | Header |
| `jsonData.httpHeaders[].value` | string | jsonData |  | Value |
| `jsonData.analyticsURL` | enum | jsonData |  | The Analytics API URL |
| `jsonData.globalAccountName` | string | jsonData |  | The global account name, as shown in the Controller UI License page. |
| `secureJsonData.analyticsAPIKey` 🔒 | string | secureJsonData |  | The Analytics API Key |
| `jsonData.tlsSkipVerify` | boolean | jsonData |  | Skip TLS Verify |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: AppDynamics
    type: dlopes7-appdynamics-datasource
    access: proxy
    basicAuth: false
    basicAuthUser: user
    url: "http://localhost:8086"
    jsonData:
      clientDomain: Client Domain
      clientName: Client Name
      tlsSkipVerify: false
    secureJsonData:
      basicAuthPassword: "<YOUR_PASSWORD>"
      clientSecret: "<YOUR_CLIENT_SECRET>"
```

**Terraform**

```hcl
resource "grafana_data_source" "dlopes7_appdynamics_datasource" {
  type = "dlopes7-appdynamics-datasource"
  name = "AppDynamics"
  url = "http://localhost:8086"

  json_data_encoded = jsonencode({
    clientDomain = "Client Domain"
    clientName = "Client Name"
    tlsSkipVerify = false
  })

  secure_json_data_encoded = jsonencode({
    basicAuthPassword = "<YOUR_PASSWORD>"
    clientSecret = "<YOUR_CLIENT_SECRET>"
  })
}
```

