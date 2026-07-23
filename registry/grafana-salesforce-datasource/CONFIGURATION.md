# Salesforce configuration

Configuration reference for the **Salesforce** data source (`grafana-salesforce-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-salesforce-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.authType` | enum (user, jwt) | jsonData |  | Authentication |
| `jsonData.user` | string | jsonData | conditional | User Name |
| `secureJsonData.password` 🔒 | string | secureJsonData | conditional | Password |
| `secureJsonData.securityToken` 🔒 | string | secureJsonData |  | Security Token |
| `secureJsonData.clientID` 🔒 | string | secureJsonData | conditional | Consumer Key |
| `secureJsonData.clientSecret` 🔒 | string | secureJsonData | conditional | Consumer Secret |
| `secureJsonData.cert` 🔒 | string (multiline) | secureJsonData | conditional | Certificate |
| `secureJsonData.privateKey` 🔒 | string (multiline) | secureJsonData | conditional | Private Key |
| `jsonData.tokenUrl` | enum (https://login.salesforce.com, https://test.salesforce.com) | jsonData |  | Environment |
| `jsonData.sandbox` | boolean | jsonData |  | Legacy boolean that selects the Salesforce login/token host when `tokenUrl` is empty (`true` → https://test.salesforce.com, `false` → https://login.salesforce.com). Deprecated in favor of `tokenUrl`; the config editor no longer writes it but reads it to derive the initial Environment selection, and the backend still honors it for backwards compatibility and provisioning. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Credentials (`user`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Salesforce
    type: grafana-salesforce-datasource
    access: proxy
    jsonData:
      authType: user
      sandbox: false
      tokenUrl: "https://login.salesforce.com"
      user: Salesforce User
    secureJsonData:
      clientID: "<YOUR_CONSUMER_KEY>"
      clientSecret: "<YOUR_CONSUMER_SECRET>"
      password: "<YOUR_PASSWORD>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_salesforce_datasource_user" {
  type = "grafana-salesforce-datasource"
  name = "Salesforce"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "user"
    sandbox = false
    tokenUrl = "https://login.salesforce.com"
    user = "Salesforce User"
  })

  secure_json_data_encoded = jsonencode({
    clientID = "<YOUR_CONSUMER_KEY>"
    clientSecret = "<YOUR_CONSUMER_SECRET>"
    password = "<YOUR_PASSWORD>"
  })
}
```

### JWT (`jwt`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Salesforce
    type: grafana-salesforce-datasource
    access: proxy
    jsonData:
      authType: jwt
      sandbox: false
      tokenUrl: "https://login.salesforce.com"
    secureJsonData:
      cert: "<YOUR_CERTIFICATE>"
      privateKey: "<YOUR_PRIVATE_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_salesforce_datasource_jwt" {
  type = "grafana-salesforce-datasource"
  name = "Salesforce"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "jwt"
    sandbox = false
    tokenUrl = "https://login.salesforce.com"
  })

  secure_json_data_encoded = jsonencode({
    cert = "<YOUR_CERTIFICATE>"
    privateKey = "<YOUR_PRIVATE_KEY>"
  })
}
```

