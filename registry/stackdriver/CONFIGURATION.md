# Google Cloud Monitoring configuration

How to configure the **Google Cloud Monitoring** data source (`stackdriver`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/google-cloud-monitoring/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Authentication](#authentication)
- [Additional settings](#additional-settings) — _optional_

## Authentication

### Authentication type

_optional · radio_

| | |
|---|---|
| Default | `jwt` |
| Allowed values | `jwt` (Google JWT File), `gce` (GCE Default Service Account), `workloadIdentityFederation` (Workload Identity Federation), `forwardOAuthIdentity` (Forward OAuth Identity) |

**How to configure Google Cloud Monitoring datasource?**

Don't know how to get a service account key file or create a service account? Read more [in the documentation](https://grafana.com/docs/grafana/latest/datasources/google-cloud-monitoring/google-authentication/).

### JWT token

_optional · file upload_

Upload or paste a Google service-account JWT key file (.json). Its project_id, client_email, token_uri, and private_key are distributed into the JWT fields below. Mirrors the legacy JWTConfig 'Upload or paste Google JWT token' control (ConfigEditor.tsx JWT Key Details).

| | |
|---|---|
| Shown when | **Authentication type** is **Google JWT File** (`jwt`) |

### Default project

_conditionally required · string_

| | |
|---|---|
| Example | `my-gcp-project` |
| Shown when | `jsonData_authenticationType == 'jwt' || jsonData_authenticationType == 'gce' || jsonData_authenticationType == 'workloadIdentityFederation' || jsonData_authenticationType == 'forwardOAuthIdentity'` |
| Required when | `jsonData_authenticationType == 'forwardOAuthIdentity' || jsonData_authenticationType == 'workloadIdentityFederation'` |

### Client email

_conditionally required · string_

| | |
|---|---|
| Shown when | **Authentication type** is **Google JWT File** (`jwt`) |
| Required when | `jsonData_authenticationType == 'jwt' && jsonData_privateKeyPath == ''` |

### Token URI

_conditionally required · string_

| | |
|---|---|
| Shown when | **Authentication type** is **Google JWT File** (`jwt`) |
| Required when | `jsonData_authenticationType == 'jwt' && jsonData_privateKeyPath == ''` |

### Private key path

_optional · string_

Paste private key or provide path to private key file.

| | |
|---|---|
| Example | `File location of your private key (e.g. /etc/secrets/gce.pem)` |
| Shown when | **Authentication type** is **Google JWT File** (`jwt`) |

### Private key

_🔒 secret (write-only) · conditionally required · string_

Paste private key or provide path to private key file.

| | |
|---|---|
| Example | `Enter Private key` |
| Shown when | **Authentication type** is **Google JWT File** (`jwt`) |
| Required when | `jsonData_authenticationType == 'jwt' && jsonData_privateKeyPath == ''` |

### Workload Identity Pool Provider

_conditionally required · string_

Full resource name of the workload identity pool provider (e.g. projects/123/locations/global/workloadIdentityPools/my-pool/providers/my-provider).

| | |
|---|---|
| Example | `projects/<number>/locations/global/workloadIdentityPools/<pool>/providers/<provider>` |
| Shown when | **Authentication type** is **Workload Identity Federation** (`workloadIdentityFederation`) |

### Service account email

_optional · string_

Optional. If set, the federated identity impersonates this service account when calling Google APIs.

| | |
|---|---|
| Example | `name@project.iam.gserviceaccount.com` |
| Shown when | **Authentication type** is **Workload Identity Federation** (`workloadIdentityFederation`) |

### Enable

_optional · toggle_

Read more about service account impersonation here: https://cloud.google.com/iam/docs/service-account-impersonation.

| | |
|---|---|
| Default | `false` |
| Shown when | `jsonData_authenticationType == 'jwt' || jsonData_authenticationType == 'gce'` |

### Service account to impersonate

_optional · string_

| | |
|---|---|
| Shown when | `(jsonData_authenticationType == 'jwt' || jsonData_authenticationType == 'gce') && jsonData_usingImpersonation == true` |

## Additional settings

_This section is optional._

### Universe Domain

_optional · string_

Optional Google Cloud universe domain (Trusted Partner Cloud / Trusted Cloud by S3NS). Empty string is treated as 'googleapis.com' by the backend (pkg/cloudmonitoring/httpclient.go:79-83). Only rendered in the editor when the Grafana instance has secureSocksDSProxyEnabled set (ConfigEditor.tsx:78) — provisioning can set it regardless of that flag.

| | |
|---|---|
| Example | `googleapis.com` |

## Other settings

### oauthPassThru

_optional · boolean_

Automatically set to true when authenticationType is 'forwardOAuthIdentity' or 'workloadIdentityFederation'; otherwise false. Written by AuthConfig.tsx:73-74 as a side-effect of the auth-type radio, not by a direct UI toggle.

| | |
|---|---|
| Default | `false` |

### gceDefaultProject

_optional · string_

Frontend-managed cache of the GCE metadata server's default project ID, populated at runtime by ensureGCEDefaultProject (src/datasource.ts:186-191) via the /gceDefaultProject resource endpoint. The backend never reads this key — it always calls utils.GCEDefaultProject fresh for GCE auth (pkg/cloudmonitoring/cloudmonitoring.go:666-675). Do not populate this in provisioning payloads; leave it empty and let the frontend fill it in.

