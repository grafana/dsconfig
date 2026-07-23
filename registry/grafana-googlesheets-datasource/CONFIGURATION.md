# Google Sheets configuration

How to configure the **Google Sheets** data source (`grafana-googlesheets-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-googlesheets-datasource/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Authentication](#authentication)
- [Settings](#settings) — _optional_

## Authentication

### Authentication type

_optional · radio_

| | |
|---|---|
| Default | `jwt` |
| Allowed values | `key` (API Key), `jwt` (Google JWT File), `gce` (GCE Default Service Account) |

**Configure Google Sheets Authentication**

#### Choosing an authentication type

- **Google JWT File**: provides access to private spreadsheets and works in all environments where Grafana is running.
- **API Key**: simpler configuration, but requires spreadsheets to be public.
- **GCE Default Service Account**: automatically retrieves default credentials. Requires Grafana to be running on a Google Compute Engine virtual machine.

Select an Authentication type below and expand **Configure Google Sheets Authentication** for detailed guidance on configuration.

### JWT token

_optional · file upload_

Upload or paste a Google service-account JWT key file (.json). Its project_id, client_email, token_uri, and private_key are distributed into the JWT fields below.

| | |
|---|---|
| Shown when | **Authentication type** is **Google JWT File** (`jwt`) |

### API Key

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Example | `Enter API key` |
| Shown when | **Authentication type** is **API Key** (`key`) |

**Generate an API key**

1. Open the [Google Sheets](https://console.cloud.google.com/apis/library/sheets.googleapis.com?q=sheet) page in the API Library and enable access for your account.
2. Open the [Credentials page](https://console.developers.google.com/apis/credentials) in the Google API Console.
3. Click **Create Credentials** and then click **API key**.
4. Copy the key and paste it in the API Key field below. The file contents are encrypted and saved in the Grafana database.

### Default project

_optional · string_

| | |
|---|---|
| Shown when | `jsonData_authenticationType == 'jwt' || jsonData_authenticationType == 'gce'` |

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

Paste private key or provide path to private file.

| | |
|---|---|
| Example | `File location of your private key (e.g. /etc/secrets/gce.pem)` |
| Shown when | **Authentication type** is **Google JWT File** (`jwt`) |

### Private key

_🔒 secret (write-only) · conditionally required · string_

Paste private key or provide path to private file.

| | |
|---|---|
| Example | `Enter Private key` |
| Shown when | **Authentication type** is **Google JWT File** (`jwt`) |
| Required when | `jsonData_authenticationType == 'jwt' && jsonData_privateKeyPath == ''` |

## Settings

_This section is optional._

### Default Spreadsheet ID

_optional · string_

Optional spreadsheet ID to use as default when creating new queries.

| | |
|---|---|
| Example | `Select Spreadsheet ID` |

## Other settings

### authType

_optional · string_

Legacy authentication type field. Older provisioning stored the auth type here; the backend copies its value into `authenticationType` on load. Prefer `authenticationType` for new configurations.

| | |
|---|---|
| Allowed values | `key`, `jwt`, `gce` |

### jwt

_🔒 secret (write-only) · optional · string_

Legacy write-only secret used by older versions of the plugin. The backend still copies its decrypted value into memory but no runtime code path depends on it; new configurations should use `privateKey` instead.

