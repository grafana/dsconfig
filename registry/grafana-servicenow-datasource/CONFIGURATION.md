# ServiceNow configuration

Configuration reference for the **ServiceNow** data source (`grafana-servicenow-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-servicenow-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `url` | string | root | yes | Your access method is Server, this means the URL needs to be accessible from the grafana backend/server. |
| `jsonData.authMethod` | enum (basicAuth, serviceNowOAuth) | jsonData |  | Type of authentication to use. Defaults to basic auth |
| `basicAuthUser` | string | root | yes | Username of the ServiceNow account |
| `secureJsonData.basicAuthPassword` 🔒 | string | secureJsonData | yes | Password for the ServiceNow account |
| `jsonData.oauthClientID` | string | jsonData | conditional | Client ID for OAuth |
| `secureJsonData.oauthClientSecret` 🔒 | string | secureJsonData | conditional | Client Secret for OAuth |
| `jsonData.useSysTables` | boolean | jsonData |  | Query sys tables for schema/meta lookups (requires elevated permissions) |
| `jsonData.queryTimeoutSeconds` | number | jsonData |  | Maximum time in seconds for queries to complete. Increase this for slow ServiceNow instances or large tables. Default is 30 seconds. |
| `jsonData.httpHeaders` | list | jsonData |  | Additional HTTP headers sent with every request. Header names are stored in jsonData (httpHeaderName<N>); header values are write-only in secureJsonData (httpHeaderValue<N>). |
| `jsonData.httpHeaders[].name` | string | jsonData | yes | Header |
| `jsonData.httpHeaders[].value` | string | jsonData |  | Value |
| `jsonData.oauthEnabled` | boolean | jsonData |  | Deprecated legacy boolean that predates `authMethod`. Older plugin versions stored `oauthEnabled: true` to select ServiceNow OAuth. Not written by the current config editor, but still read for backwards compatibility by both the editor (initial auth-method derivation) and the backend (`GetAuthMethod`): when `authMethod` is empty, `oauthEnabled: true` selects `serviceNowOAuth`, otherwise `basicAuth`. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Basic auth (`basicAuth`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: ServiceNow
    type: grafana-servicenow-datasource
    access: proxy
    basicAuthUser: ServiceNow username
    url: "https://<YOUR INSTANCE ID>.service-now.com"
    jsonData:
      authMethod: basicAuth
      queryTimeoutSeconds: 30
      useSysTables: false
    secureJsonData:
      basicAuthPassword: "<YOUR_PASSWORD>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_servicenow_datasource_basicAuth" {
  type = "grafana-servicenow-datasource"
  name = "ServiceNow"
  url = "https://<YOUR INSTANCE ID>.service-now.com"

  json_data_encoded = jsonencode({
    authMethod = "basicAuth"
    queryTimeoutSeconds = 30
    useSysTables = false
  })

  secure_json_data_encoded = jsonencode({
    basicAuthPassword = "<YOUR_PASSWORD>"
  })
}
```

### ServiceNow OAuth (`serviceNowOAuth`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: ServiceNow
    type: grafana-servicenow-datasource
    access: proxy
    basicAuthUser: ServiceNow username
    url: "https://<YOUR INSTANCE ID>.service-now.com"
    jsonData:
      authMethod: serviceNowOAuth
      oauthClientID: OAuth Client ID
      queryTimeoutSeconds: 30
      useSysTables: false
    secureJsonData:
      basicAuthPassword: "<YOUR_PASSWORD>"
      oauthClientSecret: "<YOUR_CLIENT_SECRET>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_servicenow_datasource_serviceNowOAuth" {
  type = "grafana-servicenow-datasource"
  name = "ServiceNow"
  url = "https://<YOUR INSTANCE ID>.service-now.com"

  json_data_encoded = jsonencode({
    authMethod = "serviceNowOAuth"
    oauthClientID = "OAuth Client ID"
    queryTimeoutSeconds = 30
    useSysTables = false
  })

  secure_json_data_encoded = jsonencode({
    basicAuthPassword = "<YOUR_PASSWORD>"
    oauthClientSecret = "<YOUR_CLIENT_SECRET>"
  })
}
```

