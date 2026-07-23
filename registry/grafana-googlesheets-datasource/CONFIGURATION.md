# Google Sheets configuration

Configuration reference for the **Google Sheets** data source (`grafana-googlesheets-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-googlesheets-datasource/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.authenticationType` | enum (key, jwt, gce) | jsonData |  | Authentication type |
| `jsonData.authType` | enum (key, jwt, gce) | jsonData |  | Legacy authentication type field. Older provisioning stored the auth type here; the backend copies its value into `authenticationType` on load. Prefer `authenticationType` for new configurations. |
| `secureJsonData.apiKey` 🔒 | string | secureJsonData | conditional | API Key |
| `jsonData.defaultProject` | string | jsonData |  | Default project |
| `jsonData.clientEmail` | string | jsonData | conditional | Client email |
| `jsonData.tokenUri` | string | jsonData | conditional | Token URI |
| `jsonData.privateKeyPath` | string | jsonData |  | Paste private key or provide path to private file |
| `secureJsonData.privateKey` 🔒 | string | secureJsonData | conditional | Paste private key or provide path to private file |
| `secureJsonData.jwt` 🔒 | string | secureJsonData |  | Legacy write-only secret used by older versions of the plugin. The backend still copies its decrypted value into memory but no runtime code path depends on it; new configurations should use `privateKey` instead. |
| `jsonData.defaultSheetID` | string | jsonData |  | Optional spreadsheet ID to use as default when creating new queries |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### API Key (`key`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Google Sheets
    type: grafana-googlesheets-datasource
    access: proxy
    jsonData:
      authenticationType: key
    secureJsonData:
      apiKey: "<YOUR_API_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_googlesheets_datasource_key" {
  type = "grafana-googlesheets-datasource"
  name = "Google Sheets"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authenticationType = "key"
  })

  secure_json_data_encoded = jsonencode({
    apiKey = "<YOUR_API_KEY>"
  })
}
```

### Google JWT File (`jwt`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Google Sheets
    type: grafana-googlesheets-datasource
    access: proxy
    jsonData:
      authenticationType: jwt
      clientEmail: "<YOUR_CLIENT_EMAIL>"
      tokenUri: "<YOUR_TOKEN_URI>"
    secureJsonData:
      privateKey: "<YOUR_PRIVATE_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_googlesheets_datasource_jwt" {
  type = "grafana-googlesheets-datasource"
  name = "Google Sheets"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authenticationType = "jwt"
    clientEmail = "<YOUR_CLIENT_EMAIL>"
    tokenUri = "<YOUR_TOKEN_URI>"
  })

  secure_json_data_encoded = jsonencode({
    privateKey = "<YOUR_PRIVATE_KEY>"
  })
}
```

### GCE Default Service Account (`gce`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Google Sheets
    type: grafana-googlesheets-datasource
    access: proxy
    jsonData:
      authenticationType: gce
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_googlesheets_datasource_gce" {
  type = "grafana-googlesheets-datasource"
  name = "Google Sheets"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authenticationType = "gce"
  })
}
```

