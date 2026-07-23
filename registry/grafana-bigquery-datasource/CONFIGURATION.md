# Google BigQuery configuration

Configuration reference for the **Google BigQuery** data source (`grafana-bigquery-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-bigquery-datasource/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.authenticationType` | enum (jwt, gce, forwardOAuthIdentity, workloadIdentityFederation) | jsonData |  | Authentication type |
| `jsonData.defaultProject` | string | jsonData | conditional | Default project |
| `jsonData.clientEmail` | string | jsonData | conditional | Client email |
| `jsonData.tokenUri` | string | jsonData | conditional | Token URI |
| `jsonData.privateKeyPath` | string | jsonData |  | Paste private key or provide path to private file |
| `secureJsonData.privateKey` 🔒 | string | secureJsonData | conditional | Paste private key or provide path to private file |
| `jsonData.usingImpersonation` | boolean | jsonData |  | Enable service account impersonation. Read more about service account impersonation here: https://cloud.google.com/iam/docs/service-account-impersonation |
| `jsonData.serviceAccountToImpersonate` | string | jsonData |  | Service account to impersonate |
| `jsonData.workloadIdentityPoolProvider` | string | jsonData | conditional | Full resource name of the workload identity pool provider (e.g. projects/123/locations/global/workloadIdentityPools/my-pool/providers/my-provider) |
| `jsonData.wifServiceAccountEmail` | string | jsonData |  | Optional. If set, the federated identity impersonates this service account when calling Google APIs. |
| `jsonData.oauthPassThru` | boolean | jsonData |  | Automatically set to true when authenticationType is 'forwardOAuthIdentity' or 'workloadIdentityFederation'; otherwise false. Written by AuthConfig.tsx:73-74 as a side-effect of the auth-type radio, not by a direct UI toggle. |
| `jsonData.processingLocation` | enum (, US, EU, us-east5, us-south1, us-central1, us-west2, us-west4, northamerica-northeast1, us-east4, us-west1, us-west3, southamerica-east1, southamerica-west1, us-east1, northamerica-northeast2, europe-west1, europe-west10, europe-north1, europe-west3, europe-west2, europe-southwest1, europe-west8, europe-west4, europe-west9, europe-west12, europe-central2, europe-west6, asia-south2, asia-east2, asia-southeast2, australia-southeast2, asia-south1, asia-northeast2, asia-northeast3, asia-southeast1, australia-southeast1, asia-east1, asia-northeast1, me-central2, me-central1, me-west1) | jsonData |  | Read more about processing location here: https://cloud.google.com/bigquery/docs/locations |
| `jsonData.serviceEndpoint` | string | jsonData |  | Specifies the network address of an API service. Read more about service endpoint here: https://cloud.google.com/bigquery/docs/reference/rest#service-endpoint |
| `jsonData.MaxBytesBilled` | number | jsonData |  | Prevent queries that would process more than this amount of bytes. Read more about max bytes billed here: https://cloud.google.com/bigquery/docs/best-practices-costs |
| `jsonData.flatRateProject` | string | jsonData |  | Defined in the backend Settings struct (pkg/bigquery/types/types.go:13) but not read by any current code path and not exposed in the configuration editor. Kept in the schema so provisioning payloads are validated instead of silently accepting unknown keys. |
| `jsonData.queryPriority` | enum (INTERACTIVE, BATCH) | jsonData |  | Defined in the backend Settings struct (pkg/bigquery/types/types.go:15) as the desired default query priority (INTERACTIVE / BATCH). Not read by any current code path and not exposed in the datasource configuration editor (a queryPriority also exists on individual queries at src/types.ts:107 but that is a separate storage location). |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Google JWT File (`jwt`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Google BigQuery
    type: grafana-bigquery-datasource
    access: proxy
    jsonData:
      authenticationType: jwt
      clientEmail: "<YOUR_CLIENT_EMAIL>"
      defaultProject: "<YOUR_DEFAULT_PROJECT>"
      oauthPassThru: false
      processingLocation: ""
      tokenUri: "<YOUR_TOKEN_URI>"
      usingImpersonation: false
    secureJsonData:
      privateKey: "<YOUR_PRIVATE_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_bigquery_datasource_jwt" {
  type = "grafana-bigquery-datasource"
  name = "Google BigQuery"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authenticationType = "jwt"
    clientEmail = "<YOUR_CLIENT_EMAIL>"
    defaultProject = "<YOUR_DEFAULT_PROJECT>"
    oauthPassThru = false
    processingLocation = ""
    tokenUri = "<YOUR_TOKEN_URI>"
    usingImpersonation = false
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
  - name: Google BigQuery
    type: grafana-bigquery-datasource
    access: proxy
    jsonData:
      authenticationType: gce
      clientEmail: "<YOUR_CLIENT_EMAIL>"
      defaultProject: "<YOUR_DEFAULT_PROJECT>"
      oauthPassThru: false
      processingLocation: ""
      tokenUri: "<YOUR_TOKEN_URI>"
      usingImpersonation: false
    secureJsonData:
      privateKey: "<YOUR_PRIVATE_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_bigquery_datasource_gce" {
  type = "grafana-bigquery-datasource"
  name = "Google BigQuery"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authenticationType = "gce"
    clientEmail = "<YOUR_CLIENT_EMAIL>"
    defaultProject = "<YOUR_DEFAULT_PROJECT>"
    oauthPassThru = false
    processingLocation = ""
    tokenUri = "<YOUR_TOKEN_URI>"
    usingImpersonation = false
  })

  secure_json_data_encoded = jsonencode({
    privateKey = "<YOUR_PRIVATE_KEY>"
  })
}
```

### Forward OAuth Identity (`forwardOAuthIdentity`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Google BigQuery
    type: grafana-bigquery-datasource
    access: proxy
    jsonData:
      authenticationType: forwardOAuthIdentity
      clientEmail: "<YOUR_CLIENT_EMAIL>"
      defaultProject: "<YOUR_DEFAULT_PROJECT>"
      oauthPassThru: false
      processingLocation: ""
      tokenUri: "<YOUR_TOKEN_URI>"
    secureJsonData:
      privateKey: "<YOUR_PRIVATE_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_bigquery_datasource_forwardOAuthIdentity" {
  type = "grafana-bigquery-datasource"
  name = "Google BigQuery"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authenticationType = "forwardOAuthIdentity"
    clientEmail = "<YOUR_CLIENT_EMAIL>"
    defaultProject = "<YOUR_DEFAULT_PROJECT>"
    oauthPassThru = false
    processingLocation = ""
    tokenUri = "<YOUR_TOKEN_URI>"
  })

  secure_json_data_encoded = jsonencode({
    privateKey = "<YOUR_PRIVATE_KEY>"
  })
}
```

### Workload Identity Federation (`workloadIdentityFederation`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Google BigQuery
    type: grafana-bigquery-datasource
    access: proxy
    jsonData:
      authenticationType: workloadIdentityFederation
      clientEmail: "<YOUR_CLIENT_EMAIL>"
      defaultProject: "<YOUR_DEFAULT_PROJECT>"
      oauthPassThru: false
      processingLocation: ""
      tokenUri: "<YOUR_TOKEN_URI>"
      workloadIdentityPoolProvider: "projects/<number>/locations/global/workloadIdentityPools/<pool>/providers/<provider>"
    secureJsonData:
      privateKey: "<YOUR_PRIVATE_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_bigquery_datasource_workloadIdentityFederation" {
  type = "grafana-bigquery-datasource"
  name = "Google BigQuery"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authenticationType = "workloadIdentityFederation"
    clientEmail = "<YOUR_CLIENT_EMAIL>"
    defaultProject = "<YOUR_DEFAULT_PROJECT>"
    oauthPassThru = false
    processingLocation = ""
    tokenUri = "<YOUR_TOKEN_URI>"
    workloadIdentityPoolProvider = "projects/<number>/locations/global/workloadIdentityPools/<pool>/providers/<provider>"
  })

  secure_json_data_encoded = jsonencode({
    privateKey = "<YOUR_PRIVATE_KEY>"
  })
}
```

