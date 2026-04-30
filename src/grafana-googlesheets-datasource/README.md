# Google Sheets Datasource

| Property | Value |
| --- | --- |
| **Plugin ID** | `grafana-googlesheets-datasource` |
| **Type** | Datasource |
| **Signed by** | Grafana Labs |
| **Version** | 2.5.0 |
| **Grafana** | ≥ 11.6.0 |
| **Backend** | Yes |
| **Alerting** | Yes |
| **Repository** | [grafana/google-sheets-datasource](https://github.com/grafana/google-sheets-datasource) |
| **Docs** | [grafana.com/docs/plugins/grafana-googlesheets-datasource](https://grafana.com/docs/plugins/grafana-googlesheets-datasource/) |

## Overview

The Google Sheets data source plugin for Grafana lets you visualize
your Google spreadsheets in Grafana.

---

## `jsonData` Fields

### Fields from `@grafana/google-sdk` (`DataSourceOptions`)

The plugin extends `DataSourceOptions` from `@grafana/google-sdk`
([src/types.ts](https://github.com/grafana/grafana-google-sdk-react/blob/main/src/types.ts)),
which provides the following fields:

#### `authenticationType`

| Property | Value |
| --- | --- |
| **Required** | Yes |
| **Type** | `"jwt" \| "key" \| "gce" \| "workloadIdentityFederation"` |
| **Default** | `"jwt"` (set by `getOptionsWithDefaults` in the frontend) |
| **Package** | `@grafana/google-sdk` → `DataSourceOptions.authenticationType` |

Which authentication mechanism the datasource uses to talk to Google APIs.

| Value | Description |
| --- | --- |
| `"jwt"` | Service account JWT / JSON key. Supports private spreadsheets. |
| `"key"` | API key. Spreadsheets must be publicly shared. **Sheets-specific extension** — not in the base SDK. |
| `"gce"` | Default GCE service account. Grafana must run on Google Compute Engine. |
| `"workloadIdentityFederation"` | Workload Identity Federation (WIF). |

**Backend behavior:**

- If empty, the backend errors with `"missing AuthenticationType setting"`.
  → [googleclient.go#L123-L131](https://github.com/grafana/google-sheets-datasource/blob/main/pkg/googlesheets/googleclient.go#L123-L131)

**UI hints:**

- Rendered as RadioButtonGroup with "Google JWT File", "API Key", and
  "GCE Default Service Account" options.
  → [ConfigEditor.tsx#L80-L109](https://github.com/grafana/google-sheets-datasource/blob/main/src/components/ConfigEditor.tsx#L80-L109)
- The `AuthConfig` component from `@grafana/google-sdk` handles
  JWT/GCE/WIF switching.
  → [AuthConfig.tsx](https://github.com/grafana/grafana-google-sdk-react/blob/main/src/components/AuthConfig.tsx)

---

#### `defaultProject`

| Property | Value |
| --- | --- |
| **Required** | Conditionally — for `"gce"` and `"jwt"` auth |
| **Type** | `string` |
| **Package** | `@grafana/google-sdk` → `DataSourceOptions.defaultProject` |

GCP project ID. In JWT mode, this is the `project_id` from the service
account JSON. In GCE mode, it identifies the GCE project.

**UI hints:**

- Shown as "Project ID" in the JWT form, "Default project" for GCE.
  → [JWTForm.tsx#L63-L96](https://github.com/grafana/grafana-google-sdk-react/blob/main/src/components/JWTForm.tsx#L63-L96)

---

#### `clientEmail`

| Property | Value |
| --- | --- |
| **Required** | Conditionally — for `"jwt"` auth with explicit fields |
| **Type** | `string` |
| **Package** | `@grafana/google-sdk` → `DataSourceOptions.clientEmail` |

Service account email address for JWT authentication.

**UI hints:**

- Shown as "Client email" in the JWT form.
  → [JWTForm.tsx#L63-L96](https://github.com/grafana/grafana-google-sdk-react/blob/main/src/components/JWTForm.tsx#L63-L96)

---

#### `tokenUri`

| Property | Value |
| --- | --- |
| **Required** | Conditionally — for `"jwt"` auth with explicit fields |
| **Type** | `string` |
| **Default** | `"https://oauth2.googleapis.com/token"` |
| **Package** | `@grafana/google-sdk` → `DataSourceOptions.tokenUri` |

OAuth2 token endpoint URI.

**UI hints:**

- Shown as "Token URI" in the JWT form.
  → [JWTForm.tsx#L63-L96](https://github.com/grafana/grafana-google-sdk-react/blob/main/src/components/JWTForm.tsx#L63-L96)

---

#### `privateKeyPath`

| Property | Value |
| --- | --- |
| **Required** | No |
| **Type** | `string` |
| **Package** | `@grafana/google-sdk` → `DataSourceOptions.privateKeyPath` |

Path to a local private key file on the Grafana server. Alternative to
providing the key via `secureJsonData.privateKey`.
**Not supported in hosted environments** (e.g. Grafana Cloud).

**UI hints:**

- Toggle between "Paste private key" and "Provide path to private key
  file" links in the JWT form.
  → [JWTForm.tsx#L31-L62](https://github.com/grafana/grafana-google-sdk-react/blob/main/src/components/JWTForm.tsx#L31-L62)

---

#### `usingImpersonation`

| Property | Value |
| --- | --- |
| **Required** | No |
| **Type** | `boolean` |
| **Package** | `@grafana/google-sdk` → `DataSourceOptions.usingImpersonation` |

Enable service account impersonation.

**UI hints:**

- Shown as a toggle switch under "Service account impersonation" fieldset.
  → [AuthConfig.tsx#L170-L197](https://github.com/grafana/grafana-google-sdk-react/blob/main/src/components/AuthConfig.tsx#L170-L197)

---

#### `serviceAccountToImpersonate`

| Property | Value |
| --- | --- |
| **Required** | Conditionally — when `usingImpersonation` is `true` |
| **Type** | `string` |
| **Package** | `@grafana/google-sdk` → `DataSourceOptions.serviceAccountToImpersonate` |

Email of the service account to impersonate.

---

#### `workloadIdentityPoolProvider`

| Property | Value |
| --- | --- |
| **Required** | Conditionally — for `"workloadIdentityFederation"` auth |
| **Type** | `string` |
| **Package** | `@grafana/google-sdk` → `DataSourceOptions.workloadIdentityPoolProvider` |

Full resource name of the workload identity pool provider.

**UI hints:**

- Shown in the WIF configuration editor.
  → [WIFConfigEditor.tsx](https://github.com/grafana/grafana-google-sdk-react/blob/main/src/components/WIFConfigEditor.tsx)

---

#### `wifServiceAccountEmail`

| Property | Value |
| --- | --- |
| **Required** | No |
| **Type** | `string` |
| **Package** | `@grafana/google-sdk` → `DataSourceOptions.wifServiceAccountEmail` |

Service account email for WIF impersonation.

---

### Plugin-specific `jsonData` fields

#### `authType`

| Property | Value |
| --- | --- |
| **Required** | No |
| **Type** | `"jwt" \| "key" \| "gce"` |
| **Deprecated** | Yes — legacy alias for `authenticationType` |

**Backend migration:**

- If `authType` is set and `authenticationType` is not, the backend
  copies `authType` → `authenticationType`.
  → [settings.go#L45-L52](https://github.com/grafana/google-sheets-datasource/blob/main/pkg/models/settings.go#L45-L52)

**UI normalization:**

- `authenticationType = authenticationType || authType`
  → [utils.ts#L4-L15](https://github.com/grafana/google-sheets-datasource/blob/main/src/utils.ts#L4-L15)

---

#### `defaultSheetID`

| Property | Value |
| --- | --- |
| **Required** | No |
| **Type** | `string` |

Default Spreadsheet ID pre-filled in new queries. Accepts a spreadsheet
ID or full Google Sheets URL.

**Source definition:**

- Frontend: `GoogleSheetsDataSourceOptions.defaultSheetID`
  → [src/types.ts#L16-L18](https://github.com/grafana/google-sheets-datasource/blob/main/src/types.ts#L16-L18)

**UI hints:**

- In JWT mode, the UI loads spreadsheet IDs from the backend and shows
  a selectable dropdown (SegmentAsync).
  → [ConfigEditor.tsx#L23-L49](https://github.com/grafana/google-sheets-datasource/blob/main/src/components/ConfigEditor.tsx#L23-L49)
- Users can select from the list, paste a URL, or manually enter an ID.
  → [configure.md#L128-L141](https://github.com/grafana/google-sheets-datasource/blob/main/docs/sources/configure.md#L128-L141)

---

## `secureJsonData` Fields

### Fields from `@grafana/google-sdk` (`DataSourceSecureJsonData`)

#### `privateKey`

| Property | Value |
| --- | --- |
| **Required** | When `authenticationType` is `"jwt"` (unless using `privateKeyPath`) |
| **Type** | `string` |
| **Package** | `@grafana/google-sdk` → `DataSourceSecureJsonData.privateKey` |

PEM-encoded private key for JWT / service account authentication.

---

### Plugin-specific `secureJsonData` fields

#### `apiKey`

| Property | Value |
| --- | --- |
| **Required** | When `authenticationType` is `"key"` |
| **Type** | `string` |

Google API key for public spreadsheet access.

**Source definition:**

- Frontend: `GoogleSheetsSecureJSONData.apiKey`
  → [src/types.ts#L11-L13](https://github.com/grafana/google-sheets-datasource/blob/main/src/types.ts#L11-L13)

**Backend behavior:**

- Required for API key auth; errors with `"missing API Key"` if empty.
  → [googleclient.go#L133-L144](https://github.com/grafana/google-sheets-datasource/blob/main/pkg/googlesheets/googleclient.go#L133-L144)

**UI hints:**

- Rendered as a SecretInput with placeholder "Enter API key". Only
  visible when `authenticationType` is `"key"`.
  → [ConfigEditor.tsx#L109-L139](https://github.com/grafana/google-sheets-datasource/blob/main/src/components/ConfigEditor.tsx#L109-L139)

---

#### `jwt`

| Property | Value |
| --- | --- |
| **Required** | No |
| **Type** | `string` |
| **Deprecated** | Yes — legacy field |

Legacy secure field for the full JWT JSON content. The backend still
reads this for backwards compatibility. When set, the backend extracts
`client_email`, `token_uri`, `project_id`, and `private_key` from
the JSON.

**Backend behavior:**

- Read by the backend and parsed via `google.JWTConfigFromJSON`.
  → [googleclient.go#L185-L212](https://github.com/grafana/google-sheets-datasource/blob/main/pkg/googlesheets/googleclient.go#L185-L212)
- Settings struct field: `JWT string`
  → [settings.go#L14](https://github.com/grafana/google-sheets-datasource/blob/main/pkg/models/settings.go#L14)

---

## Provisioning Examples

### JWT authentication (service account with explicit fields)

```yaml
apiVersion: 1
datasources:
  - name: Google Sheets (JWT)
    type: grafana-googlesheets-datasource
    jsonData:
      authenticationType: jwt
      clientEmail: grafana@my-project.iam.gserviceaccount.com
      tokenUri: https://oauth2.googleapis.com/token
      defaultProject: my-gcp-project
      defaultSheetID: 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms
    secureJsonData:
      privateKey: |
        -----BEGIN RSA PRIVATE KEY-----
        ...your private key here...
        -----END RSA PRIVATE KEY-----
```

### JWT authentication (full JSON key file via legacy field)

```yaml
apiVersion: 1
datasources:
  - name: Google Sheets (JWT legacy)
    type: grafana-googlesheets-datasource
    jsonData:
      authenticationType: jwt
    secureJsonData:
      jwt: |
        {
          "type": "service_account",
          "project_id": "my-project",
          "private_key_id": "abc123",
          "private_key": "-----BEGIN RSA PRIVATE KEY-----\n...\n-----END RSA PRIVATE KEY-----\n",
          "client_email": "grafana@my-project.iam.gserviceaccount.com",
          "client_id": "123456789",
          "auth_uri": "https://accounts.google.com/o/oauth2/auth",
          "token_uri": "https://oauth2.googleapis.com/token"
        }
```

### API key authentication (public spreadsheets only)

```yaml
apiVersion: 1
datasources:
  - name: Google Sheets (API Key)
    type: grafana-googlesheets-datasource
    jsonData:
      authenticationType: key
      defaultSheetID: 1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgVE2upms
    secureJsonData:
      apiKey: AIzaSyA1b2C3d4E5f6G7h8I9jKlMnOpQrStUvWx
```

### GCE authentication (Google Compute Engine)

```yaml
apiVersion: 1
datasources:
  - name: Google Sheets (GCE)
    type: grafana-googlesheets-datasource
    jsonData:
      authenticationType: gce
      defaultProject: my-gcp-project
```

### JWT with local private key file (self-hosted only)

```yaml
apiVersion: 1
datasources:
  - name: Google Sheets (JWT file path)
    type: grafana-googlesheets-datasource
    jsonData:
      authenticationType: jwt
      clientEmail: grafana@my-project.iam.gserviceaccount.com
      tokenUri: https://oauth2.googleapis.com/token
      defaultProject: my-gcp-project
      privateKeyPath: /etc/grafana/google-sheets-key.pem
```

### JWT with service account impersonation

```yaml
apiVersion: 1
datasources:
  - name: Google Sheets (JWT + impersonation)
    type: grafana-googlesheets-datasource
    jsonData:
      authenticationType: jwt
      clientEmail: grafana@my-project.iam.gserviceaccount.com
      tokenUri: https://oauth2.googleapis.com/token
      defaultProject: my-gcp-project
      usingImpersonation: true
      serviceAccountToImpersonate: target-sa@my-project.iam.gserviceaccount.com
    secureJsonData:
      privateKey: |
        -----BEGIN RSA PRIVATE KEY-----
        ...your private key here...
        -----END RSA PRIVATE KEY-----
```

---

## Source Code References

| Component | File | Repository |
| --- | --- | --- |
| Backend settings struct | [`pkg/models/settings.go`](https://github.com/grafana/google-sheets-datasource/blob/main/pkg/models/settings.go) | grafana/google-sheets-datasource |
| Backend auth client | [`pkg/googlesheets/googleclient.go`](https://github.com/grafana/google-sheets-datasource/blob/main/pkg/googlesheets/googleclient.go) | grafana/google-sheets-datasource |
| Frontend types | [`src/types.ts`](https://github.com/grafana/google-sheets-datasource/blob/main/src/types.ts) | grafana/google-sheets-datasource |
| Config editor UI | [`src/components/ConfigEditor.tsx`](https://github.com/grafana/google-sheets-datasource/blob/main/src/components/ConfigEditor.tsx) | grafana/google-sheets-datasource |
| Provisioning docs | [`docs/sources/configure.md`](https://github.com/grafana/google-sheets-datasource/blob/main/docs/sources/configure.md) | grafana/google-sheets-datasource |

## Packages Used

| Package | Language | Usage |
| --- | --- | --- |
| [`@grafana/google-sdk`](https://github.com/grafana/grafana-google-sdk-react) | TypeScript | `DataSourceOptions` base, `DataSourceSecureJsonData`, `GoogleAuthType` constants, `AuthConfig` component |
| [`@grafana/data`](https://github.com/grafana/grafana/tree/main/packages/grafana-data) | TypeScript | `DataSourceJsonData` — grandparent interface of `DataSourceOptions` |
| [`grafana-google-sdk-go`](https://github.com/grafana/grafana-google-sdk-go) | Go | Google authentication helpers (GCE, JWT token provider) |
| [`grafana-plugin-sdk-go`](https://github.com/grafana/grafana-plugin-sdk-go) | Go | `backend.DataSourceInstanceSettings` for loading settings |

---

*Auto-generated from [grafana.com/api/plugins/grafana-googlesheets-datasource](https://grafana.com/api/plugins/grafana-googlesheets-datasource) and source code analysis.*
