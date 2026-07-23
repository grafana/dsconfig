# Google Cloud Monitoring configuration

Configuration reference for the **Google Cloud Monitoring** data source (`stackdriver`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/google-cloud-monitoring/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.authenticationType` | enum (jwt, gce, workloadIdentityFederation, forwardOAuthIdentity) | jsonData |  | Authentication type |
| `jsonData.defaultProject` | string | jsonData | conditional | Default project |
| `jsonData.clientEmail` | string | jsonData | conditional | Client email |
| `jsonData.tokenUri` | string | jsonData | conditional | Token URI |
| `jsonData.privateKeyPath` | string | jsonData |  | Paste private key or provide path to private key file |
| `secureJsonData.privateKey` 🔒 | string | secureJsonData | conditional | Paste private key or provide path to private key file |
| `jsonData.usingImpersonation` | boolean | jsonData |  | Read more about service account impersonation here: https://cloud.google.com/iam/docs/service-account-impersonation |
| `jsonData.serviceAccountToImpersonate` | string | jsonData |  | Service account to impersonate |
| `jsonData.workloadIdentityPoolProvider` | string | jsonData | conditional | Full resource name of the workload identity pool provider (e.g. projects/123/locations/global/workloadIdentityPools/my-pool/providers/my-provider) |
| `jsonData.wifServiceAccountEmail` | string | jsonData |  | Optional. If set, the federated identity impersonates this service account when calling Google APIs. |
| `jsonData.oauthPassThru` | boolean | jsonData |  | Automatically set to true when authenticationType is 'forwardOAuthIdentity' or 'workloadIdentityFederation'; otherwise false. Written by AuthConfig.tsx:73-74 as a side-effect of the auth-type radio, not by a direct UI toggle. |
| `jsonData.universeDomain` | string | jsonData |  | Optional Google Cloud universe domain (Trusted Partner Cloud / Trusted Cloud by S3NS). Empty string is treated as 'googleapis.com' by the backend (pkg/cloudmonitoring/httpclient.go:79-83). Only rendered in the editor when the Grafana instance has secureSocksDSProxyEnabled set (ConfigEditor.tsx:78) — provisioning can set it regardless of that flag. |
| `jsonData.gceDefaultProject` | string | jsonData |  | Frontend-managed cache of the GCE metadata server's default project ID, populated at runtime by ensureGCEDefaultProject (src/datasource.ts:186-191) via the /gceDefaultProject resource endpoint. The backend never reads this key — it always calls utils.GCEDefaultProject fresh for GCE auth (pkg/cloudmonitoring/cloudmonitoring.go:666-675). Do not populate this in provisioning payloads; leave it empty and let the frontend fill it in. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Google JWT File (`jwt`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Google Cloud Monitoring
    type: stackdriver
    access: proxy
    jsonData:
      authenticationType: jwt
      clientEmail: "<YOUR_CLIENT_EMAIL>"
      oauthPassThru: false
      tokenUri: "<YOUR_TOKEN_URI>"
      usingImpersonation: false
    secureJsonData:
      privateKey: "<YOUR_PRIVATE_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "stackdriver_jwt" {
  type = "stackdriver"
  name = "Google Cloud Monitoring"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authenticationType = "jwt"
    clientEmail = "<YOUR_CLIENT_EMAIL>"
    oauthPassThru = false
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
  - name: Google Cloud Monitoring
    type: stackdriver
    access: proxy
    jsonData:
      authenticationType: gce
      oauthPassThru: false
      usingImpersonation: false
```

**Terraform**

```hcl
resource "grafana_data_source" "stackdriver_gce" {
  type = "stackdriver"
  name = "Google Cloud Monitoring"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authenticationType = "gce"
    oauthPassThru = false
    usingImpersonation = false
  })
}
```

### Workload Identity Federation (`workloadIdentityFederation`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Google Cloud Monitoring
    type: stackdriver
    access: proxy
    jsonData:
      authenticationType: workloadIdentityFederation
      defaultProject: my-gcp-project
      oauthPassThru: false
      workloadIdentityPoolProvider: "projects/<number>/locations/global/workloadIdentityPools/<pool>/providers/<provider>"
```

**Terraform**

```hcl
resource "grafana_data_source" "stackdriver_workloadIdentityFederation" {
  type = "stackdriver"
  name = "Google Cloud Monitoring"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authenticationType = "workloadIdentityFederation"
    defaultProject = "my-gcp-project"
    oauthPassThru = false
    workloadIdentityPoolProvider = "projects/<number>/locations/global/workloadIdentityPools/<pool>/providers/<provider>"
  })
}
```

### Forward OAuth Identity (`forwardOAuthIdentity`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Google Cloud Monitoring
    type: stackdriver
    access: proxy
    jsonData:
      authenticationType: forwardOAuthIdentity
      defaultProject: my-gcp-project
      oauthPassThru: false
```

**Terraform**

```hcl
resource "grafana_data_source" "stackdriver_forwardOAuthIdentity" {
  type = "stackdriver"
  name = "Google Cloud Monitoring"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authenticationType = "forwardOAuthIdentity"
    defaultProject = "my-gcp-project"
    oauthPassThru = false
  })
}
```

