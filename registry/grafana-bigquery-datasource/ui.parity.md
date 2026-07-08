# Google BigQuery (grafana-bigquery-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-bigquery-datasource` (product name: Google BigQuery)
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (legacy UID `cfrbd04mur6kga`)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-bigquery-datasource`
- **Method:** Playwright captured both UIs (screenshots + DOM + `Save` console payload); the local edited `dsconfig.json` was served to the new UI via `context.route(...)` interception of the remote schema fetch (`capture-bigquery.js`, adapted from `capture-stackdriver.js`).
- **Result:** **Parity achieved.** The one missing affordance — the **Google JWT service-account file upload** (legacy "JWT Key Details") — was added as a `fileUpload` field and verified to distribute credentials into the JWT fields. No HTTP headers (cloud-auth datasource, not HTTP-proxy). No `required:true` fix needed (all required fields are conditionally required on the auth type).

---

## The gap: Google JWT file upload (`fileUpload`)

The legacy UI's **JWT Key Details** section renders a Google JWT (service-account) key
uploader with a single `<input type="file">` (`fileInputs:1`, accept `.json`), plus
paste / manual-fill toggles. Uploading/pasting the service-account JSON parses it and fills
the JWT credential fields. The schema modeled the individual JWT fields (defaultProject,
clientEmail, tokenUri, privateKey) but **not the upload control**, so the new UI could only
be filled manually.

**Fix (in `dsconfig.json` only):** added a virtual `fileUpload` field with a `fileMapping` that
mirrors the legacy `JWTConfig` distribution — identical shape to the `stackdriver`
`virtual_jwtUpload` field (both are Google service-account uploaders). The `plugin-ui`
`FileUploadField` component (labels literally "…JWT Token") is purpose-built for this control.

```jsonc
{
  "id": "virtual_jwtUpload",
  "key": "jwtUpload",
  "label": "JWT token",
  "valueType": "string",
  "kind": "virtual",
  "dependsOn": "jsonData_authenticationType == 'jwt'",
  "ui": {
    "component": "fileUpload",
    "accept": [".json"],
    "fileMapping": {
      "project_id": "jsonData.defaultProject",
      "client_email": "jsonData.clientEmail",
      "token_uri": "jsonData.tokenUri",
      "private_key": "secureJsonData.privateKey",
    },
  },
}
```

Placed in the `authentication` group right after the auth-type selector; shown only when
`authenticationType == 'jwt'` (matching the legacy "Google JWT File" flow).

### Verified end-to-end (new UI)

- The upload control renders for JWT auth: FileDropzone + **Paste JWT Token** / **Fill In JWT Token manually** buttons (screenshot `bigquery-fileupload.png`).
  - `fileUpload renders: {"hasDropzone":true,"pasteBtn":true,"buttons":["Paste JWT Token","Fill In JWT Token manually"]}`
- Pasting a service-account JSON (`{project_id, client_email, token_uri, private_key, …}`) distributed the values into the fields (screenshot `bigquery-fileupload-distributed.png`):
  - `distributed to fields: {"defaultProject":"my-bq-proj","clientEmail":"sa@my-bq-proj.iam.gserviceaccount.com","tokenUri":"https://oauth2.googleapis.com/token"}`
- **Save & Test** produced:

```json
"jsonData":       { "authenticationType": "jwt", "defaultProject": "my-bq-proj",
                    "clientEmail": "sa@my-bq-proj.iam.gserviceaccount.com",
                    "tokenUri": "https://oauth2.googleapis.com/token",
                    "usingImpersonation": false, "oauthPassThru": false },
"secureJsonData": { "privateKey": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n" }
```

i.e. `project_id → jsonData.defaultProject`, `client_email → jsonData.clientEmail`,
`token_uri → jsonData.tokenUri`, `private_key → secureJsonData.privateKey` — matching the legacy
`JWTConfig` behavior. The distribution targets **match what persists** in the save payload, so no
`fileMapping` adjustment was needed.

---

## Field-by-field parity

| Legacy field (section)                                                          | new (schema id)                                                                        | Target                        | Status                  |
| ------------------------------------------------------------------------------- | -------------------------------------------------------------------------------------- | ----------------------------- | ----------------------- |
| Authentication type (JWT / GCE / Forward OAuth / WIF)                           | `jsonData_authenticationType`                                                          | `jsonData.authenticationType` | ✅ (radio)              |
| **JWT file upload** (JWT Key Details)                                           | `virtual_jwtUpload`                                                                    | (virtual → distributes)       | ➕ added (`fileUpload`) |
| Default project                                                                 | `jsonData_defaultProject`                                                              | `jsonData.defaultProject`     | ✅ 🔀                   |
| Client email / Token URI / Private key path                                     | `jsonData_clientEmail` / `jsonData_tokenUri` / `jsonData_privateKeyPath`               | `jsonData.*`                  | ✅ 🔀 (jwt)             |
| Private key (secret)                                                            | `secureJsonData_privateKey`                                                            | `secureJsonData.privateKey`   | ✅ 🔀 (jwt)             |
| Service account impersonation (Enable + SA)                                     | `jsonData_usingImpersonation` / `jsonData_serviceAccountToImpersonate`                 | `jsonData.*`                  | ✅ 🔀                   |
| WIF pool provider / SA email                                                    | `jsonData_workloadIdentityPoolProvider` / `jsonData_wifServiceAccountEmail`            | `jsonData.*`                  | ✅ 🔀 (WIF)             |
| Processing location / Service endpoint / Max bytes billed (Additional Settings) | `jsonData_processingLocation` / `jsonData_serviceEndpoint` / `jsonData_MaxBytesBilled` | `jsonData.*`                  | ✅                      |

`jsonData_oauthPassThru` (editor-managed side-effect of the auth radio) and the backend-only
unused keys `jsonData_flatRateProject` / `jsonData_queryPriority` are modeled but not surfaced —
unchanged, no legacy UI counterpart.

---

## HTTP headers / required — not applicable

- **HTTP headers:** the legacy Google BigQuery editor has **no** Custom HTTP Headers section (`hasCustomHeaders:false`, `addHeaderBtn:false`) — it is a cloud-auth datasource that uses the Google SDK, not the generic HTTP proxy. New UI confirms `headersEditor:false`. Not added.
- **`required:true` fix:** grafana-bigquery-datasource has **no** `requiredWhen:"true"` fields — every required field (JWT credential set, WIF pool provider) is _conditionally_ required on the auth type, which is correct. No change.

## Conditional fields — tested

The `authenticationType` radio drives the visible credential set: JWT → JWT fields + the upload
control; WIF → pool-provider/SA-email; GCE → impersonation; Forward OAuth → no extra fields.
All conditionals render per the schema `dependsOn` (new-UI `fieldLabels` snapshot in
`newgen-bigquery-parity.json`: JWT token, Default project, Client email, Token URI, Private key
path, Private key, Enable, + Additional Settings group).

---

## Verification

```
go generate ./registry/grafana-bigquery-datasource/...
go test ./registry/grafana-bigquery-datasource/...   # PASS
```

`TestSchemaConformance` 8/8 subtests PASS (BaseFieldsResolved, SchemaRoundTrip,
SchemaArtifactInSync, SchemaSpecHasNoSecureJSON, ConfigSchemaValid, JSONDataMatchesStruct,
JSONDataTypesMatchStruct, SecureValuesMatchLoadSettings); `TestLoadConfig`, `TestApplyDefaults`,
and `TestValidate` all PASS. The `fileUpload` field is `kind:virtual` (no storage target), so it
is skipped by the jsonData/secure conformance walkers and by the SDK spec converter —
`SchemaArtifactInSync` confirms the `.gen.json` artifacts are unchanged (no regeneration needed).

## Files changed

- [`registry/grafana-bigquery-datasource/dsconfig.json`](dsconfig.json) — added the `virtual_jwtUpload` `fileUpload` field (in the `authentication` group, right after `jsonData_authenticationType`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `schema.gen.json`, `settings.gen.json` (virtual field has no artifact/struct impact).
