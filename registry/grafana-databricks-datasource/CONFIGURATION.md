# Databricks configuration

Configuration reference for the **Databricks** data source (`grafana-databricks-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-databricks-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.host` | string | jsonData | yes | Host |
| `jsonData.httpPath` | string | jsonData | yes | Http Path |
| `jsonData.authType` | enum (Pat, OauthPT, OauthM2M, OauthOBO, AzureM2M) | jsonData |  | Authentication type of Databricks |
| `secureJsonData.token` 🔒 | string | secureJsonData | conditional | Token |
| `jsonData.azureCredentials` | any | jsonData | conditional | Discriminated-union credentials object written by the `@grafana/azure-sdk` `AzureCredentialsForm` when authType is 'OauthOBO'. Shape: `{ authType: 'clientsecret-obo', azureCloud, tenantId, clientId }` (the client secret is stored write-only in secureJsonData.azureClientSecret). Parsed by the backend via `azcredentials.FromDatasourceData` (pkg/models/settings.go:121-139). |
| `secureJsonData.azureClientSecret` 🔒 | string | secureJsonData | conditional | App Registration client secret for Azure On-Behalf-Of (authType 'OauthOBO'). Written write-only by `@grafana/azure-sdk`; check `secureJsonFields.azureClientSecret` on the read side. |
| `jsonData.clientId` | string | jsonData | conditional | Client ID |
| `secureJsonData.clientSecret` 🔒 | string | secureJsonData | conditional | Client Secret |
| `jsonData.tenantId` | string | jsonData | conditional | Directory (tenant) ID |
| `jsonData.azureCloud` | enum (AzureCloud, AzureChinaCloud, AzureUSGovernment) | jsonData |  | Azure Cloud |
| `jsonData.oauthPassThru` | boolean | jsonData |  | Set to true automatically when authType is 'OauthPT' (OAuth Passthrough) or 'OauthOBO' (Azure On-Behalf-Of) so Grafana forwards the caller's OAuth identity to Databricks. The backend hard-fails On-Behalf-Of auth if this is not true (pkg/models/settings.go:141-143, ErrInvalidOAuth). No editor UI — written as a side-effect of selecting the auth type. |
| `jsonData.retries` | string | jsonData |  | Retries |
| `jsonData.pause` | string | jsonData |  | Pause |
| `jsonData.timeout` | string | jsonData |  | Timeout |
| `jsonData.rows` | string | jsonData |  | Max Rows |
| `jsonData.retryTimeout` | string | jsonData |  | Retry Timeout |
| `jsonData.debug` | boolean | jsonData |  | Debug |
| `jsonData.enableUnitySupport` | boolean | jsonData |  | Enable Unity Catalog support for 3-level namespace (catalog.schema.table) |
| `jsonData.defaultQueryFormat` | enum (0, 1) | jsonData |  | Default Query Format |
| `jsonData.cloudFetch` | boolean | jsonData |  | Enables Databricks CloudFetch (parallel result download) in the SQL connector. Not exposed in the configuration editor; the backend force-sets it to true on every load unless the `disableCloudFetch` Grafana feature toggle is enabled (pkg/models/settings.go:161-168). |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Personal Access Token (`Pat`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Databricks
    type: grafana-databricks-datasource
    access: proxy
    jsonData:
      authType: Pat
      cloudFetch: true
      host: "https://your-databricks-instance.com"
      httpPath: /sql/protocolv1/o/0/1234567890
    secureJsonData:
      token: "<YOUR_TOKEN>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_databricks_datasource_Pat" {
  type = "grafana-databricks-datasource"
  name = "Databricks"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "Pat"
    cloudFetch = true
    host = "https://your-databricks-instance.com"
    httpPath = "/sql/protocolv1/o/0/1234567890"
  })

  secure_json_data_encoded = jsonencode({
    token = "<YOUR_TOKEN>"
  })
}
```

### OAuth Passthrough (`OauthPT`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Databricks
    type: grafana-databricks-datasource
    access: proxy
    jsonData:
      authType: OauthPT
      cloudFetch: true
      host: "https://your-databricks-instance.com"
      httpPath: /sql/protocolv1/o/0/1234567890
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_databricks_datasource_OauthPT" {
  type = "grafana-databricks-datasource"
  name = "Databricks"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "OauthPT"
    cloudFetch = true
    host = "https://your-databricks-instance.com"
    httpPath = "/sql/protocolv1/o/0/1234567890"
  })
}
```

### OAuth M2M (`OauthM2M`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Databricks
    type: grafana-databricks-datasource
    access: proxy
    jsonData:
      authType: OauthM2M
      clientId: XXXXXXXX-XXXXXXXX-XXXX-XXXXXXXXXXXX
      cloudFetch: true
      host: "https://your-databricks-instance.com"
      httpPath: /sql/protocolv1/o/0/1234567890
    secureJsonData:
      clientSecret: "<YOUR_CLIENT_SECRET>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_databricks_datasource_OauthM2M" {
  type = "grafana-databricks-datasource"
  name = "Databricks"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "OauthM2M"
    clientId = "XXXXXXXX-XXXXXXXX-XXXX-XXXXXXXXXXXX"
    cloudFetch = true
    host = "https://your-databricks-instance.com"
    httpPath = "/sql/protocolv1/o/0/1234567890"
  })

  secure_json_data_encoded = jsonencode({
    clientSecret = "<YOUR_CLIENT_SECRET>"
  })
}
```

### Azure (On-Behalf-Of) (`OauthOBO`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Databricks
    type: grafana-databricks-datasource
    access: proxy
    jsonData:
      authType: OauthOBO
      azureCredentials: "<YOUR_AZURECREDENTIALS>"
      cloudFetch: true
      host: "https://your-databricks-instance.com"
      httpPath: /sql/protocolv1/o/0/1234567890
    secureJsonData:
      azureClientSecret: "<YOUR_CLIENT_SECRET>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_databricks_datasource_OauthOBO" {
  type = "grafana-databricks-datasource"
  name = "Databricks"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "OauthOBO"
    azureCredentials = "<YOUR_AZURECREDENTIALS>"
    cloudFetch = true
    host = "https://your-databricks-instance.com"
    httpPath = "/sql/protocolv1/o/0/1234567890"
  })

  secure_json_data_encoded = jsonencode({
    azureClientSecret = "<YOUR_CLIENT_SECRET>"
  })
}
```

### Azure Entra ID M2M (`AzureM2M`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Databricks
    type: grafana-databricks-datasource
    access: proxy
    jsonData:
      authType: AzureM2M
      azureCloud: AzureCloud
      clientId: XXXXXXXX-XXXXXXXX-XXXX-XXXXXXXXXXXX
      cloudFetch: true
      host: "https://your-databricks-instance.com"
      httpPath: /sql/protocolv1/o/0/1234567890
      tenantId: XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX
    secureJsonData:
      clientSecret: "<YOUR_CLIENT_SECRET>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_databricks_datasource_AzureM2M" {
  type = "grafana-databricks-datasource"
  name = "Databricks"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "AzureM2M"
    azureCloud = "AzureCloud"
    clientId = "XXXXXXXX-XXXXXXXX-XXXX-XXXXXXXXXXXX"
    cloudFetch = true
    host = "https://your-databricks-instance.com"
    httpPath = "/sql/protocolv1/o/0/1234567890"
    tenantId = "XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"
  })

  secure_json_data_encoded = jsonencode({
    clientSecret = "<YOUR_CLIENT_SECRET>"
  })
}
```

