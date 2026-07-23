# Google BigQuery configuration

How to configure the **Google BigQuery** data source (`grafana-bigquery-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-bigquery-datasource/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Authentication](#authentication)
- [Additional Settings](#additional-settings) — _optional_

## Authentication

### Authentication type

_optional · select_

| | |
|---|---|
| Default | `jwt` |
| Allowed values | `jwt` (Google JWT File), `gce` (GCE Default Service Account), `forwardOAuthIdentity` (Forward OAuth Identity), `workloadIdentityFederation` (Workload Identity Federation) |

**How to configure Google BigQuery datasource?**

##### Uploading Google Service Account key

Create a [Google Cloud Platform (GCP) Service Account](https://cloud.google.com/iam/docs/creating-managing-service-accounts) on the project you want to show data. The **BigQuery Data Viewer** role and the **Job User** role provide all the permissions that Grafana needs. The [BigQuery API](https://console.cloud.google.com/apis/library/bigquery.googleapis.com) has to be enabled on GCP for the data source to work.

##### Using GCE Default Service Account

When Grafana is running on a Google Compute Engine (GCE) virtual machine, it is possible for Grafana to automatically retrieve the default project id and authentication token from the metadata server. For this to work, you need to make sure that you have a service account that is setup as the default account for the virtual machine and that the service account has been given read access to the BigQuery API.

**Note that, Grafana data source integrates with a single GCP project. If you need to visualize data from multiple GCP projects, create one data source per GCP project.**

### JWT token

_optional · file upload_

Upload or paste a Google service-account JWT key file (.json). Its project_id, client_email, token_uri, and private_key are distributed into the JWT fields below.

| | |
|---|---|
| Example | `Paste Google JWT token here` |
| Shown when | **Authentication type** is **Google JWT File** (`jwt`) |

### Default project

_conditionally required · string_

| | |
|---|---|
| Shown when | **JWT token** is `manual` |
| Required when | `jsonData_authenticationType == 'jwt' && jsonData_privateKeyPath == ''` |

### Client email

_conditionally required · string_

| | |
|---|---|
| Shown when | **JWT token** is `manual` |
| Required when | `jsonData_authenticationType == 'jwt' && jsonData_privateKeyPath == ''` |

### Token URI

_conditionally required · string_

| | |
|---|---|
| Shown when | **JWT token** is `manual` |
| Required when | `jsonData_authenticationType == 'jwt' && jsonData_privateKeyPath == ''` |

### Private key path

_optional · string_

Paste private key or provide path to private file.

| | |
|---|---|
| Example | `File location of your private key (e.g. /etc/secrets/gce.pem)` |
| Shown when | **JWT token** is `manual` |

### Private key

_🔒 secret (write-only) · conditionally required · string_

Paste private key or provide path to private file.

| | |
|---|---|
| Example | `Enter Private key` |
| Shown when | **JWT token** is `manual` |
| Required when | `jsonData_authenticationType == 'jwt' && jsonData_privateKeyPath == ''` |

### Enable Service Account impersonation

_optional · toggle_

Enable service account impersonation. Read more about service account impersonation here: https://cloud.google.com/iam/docs/service-account-impersonation.

| | |
|---|---|
| Default | `false` |
| Shown when | `jsonData_authenticationType == 'jwt' || jsonData_authenticationType == 'gce'` |

### Service account to impersonate

_optional · string_

| | |
|---|---|
| Shown when | `(jsonData_authenticationType == 'jwt' || jsonData_authenticationType == 'gce') && jsonData_usingImpersonation == true` |

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

## Additional Settings

_This section is optional._

### Processing location

_optional · select_

Read more about processing location here: https://cloud.google.com/bigquery/docs/locations.

| | |
|---|---|
| Default | `""` |
| Example | `Automatic location selection` |
| Allowed values | `` (Automatic location selection), `US` (United States (US)), `EU` (European Union (EU)), `us-east5` (Columbus, Ohio (us-east5)), `us-south1` (Dallas (us-south1)), `us-central1` (Iowa (us-central1)), `us-west2` (Los Angeles (us-west2)), `us-west4` (Las Vegas (us-west4)), `northamerica-northeast1` (Montréal (northamerica-northeast1)), `us-east4` (Northern Virginia (us-east4)), `us-west1` (Oregon (us-west1)), `us-west3` (Salt Lake City (us-west3)), `southamerica-east1` (São Paulo (southamerica-east1)), `southamerica-west1` (Santiago (southamerica-west1)), `us-east1` (South Carolina (us-east1)), `northamerica-northeast2` (Toronto (northamerica-northeast2)), `europe-west1` (Belgium (europe-west1)), `europe-west10` (Berlin (europe-west10)), `europe-north1` (Finland (europe-north1)), `europe-west3` (Frankfurt (europe-west3)), `europe-west2` (London (europe-west2)), `europe-southwest1` (Madrid (europe-southwest1)), `europe-west8` (Milan (europe-west8)), `europe-west4` (Netherlands (europe-west4)), `europe-west9` (Paris (europe-west9)), `europe-west12` (Turin (europe-west12)), `europe-central2` (Warsaw (europe-central2)), `europe-west6` (Zürich (europe-west6)), `asia-south2` (Delhi (asia-south2)), `asia-east2` (Hong Kong (asia-east2)), `asia-southeast2` (Jakarta (asia-southeast2)), `australia-southeast2` (Melbourne (australia-southeast2)), `asia-south1` (Mumbai (asia-south1)), `asia-northeast2` (Osaka (asia-northeast2)), `asia-northeast3` (Seoul (asia-northeast3)), `asia-southeast1` (Singapore (asia-southeast1)), `australia-southeast1` (Sydney (australia-southeast1)), `asia-east1` (Taiwan (asia-east1)), `asia-northeast1` (Tokyo (asia-northeast1)), `me-central2` (Dammam (me-central2)), `me-central1` (Doha (me-central1)), `me-west1` (Tel Aviv (me-west1)) |

### Service endpoint

_optional · string_

Specifies the network address of an API service. Read more about service endpoint here: https://cloud.google.com/bigquery/docs/reference/rest#service-endpoint.

| | |
|---|---|
| Example | `Optional, example https://bigquery.googleapis.com/bigquery/v2/` |

### Max bytes billed

_optional · number_

Prevent queries that would process more than this amount of bytes. Read more about max bytes billed here: https://cloud.google.com/bigquery/docs/best-practices-costs.

| | |
|---|---|
| Example | `Optional, example 5242880` |

## Other settings

### oauthPassThru

_optional · boolean_

Automatically set to true when authenticationType is 'forwardOAuthIdentity' or 'workloadIdentityFederation'; otherwise false. Written by AuthConfig.tsx:73-74 as a side-effect of the auth-type radio, not by a direct UI toggle.

| | |
|---|---|
| Default | `false` |

### flatRateProject

_optional · string_

Defined in the backend Settings struct (pkg/bigquery/types/types.go:13) but not read by any current code path and not exposed in the configuration editor. Kept in the schema so provisioning payloads are validated instead of silently accepting unknown keys.

### queryPriority

_optional · string_

Defined in the backend Settings struct (pkg/bigquery/types/types.go:15) as the desired default query priority (INTERACTIVE / BATCH). Not read by any current code path and not exposed in the datasource configuration editor (a queryPriority also exists on individual queries at src/types.ts:107 but that is a separate storage location).

| | |
|---|---|
| Allowed values | `INTERACTIVE`, `BATCH` |

