# Google Sheets (grafana-googlesheets-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-googlesheets-datasource` (product name: Google Sheets)
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/dfrbd04kwu0w0c` (Grafana Enterprise)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-googlesheets-datasource`
- **Method:** Playwright captured both UIs (screenshots + DOM + `Save` console payload); the local edited `dsconfig.json` was served to the new UI via `context.route(...)` interception of the remote schema fetch (`raw.githubusercontent.com/.../registry/grafana-googlesheets-datasource/dsconfig.json`).
- **Result:** **Parity achieved.** The one missing affordance — the **Google JWT service-account file upload** — was added as a `fileUpload` field and verified to distribute credentials into the JWT fields. The distribution targets the modern individual fields (`defaultProject` / `clientEmail` / `tokenUri` / `privateKey`) — **no `fileMapping` adjustment was needed** despite the plugin's legacy `secureJsonData.jwt` blob and `jsonData.authType` fields (see note below). No HTTP headers (cloud-auth datasource, not HTTP-proxy). No `required:true` fix needed (all required fields are conditionally required on the auth type).

---

## The gap: Google JWT file upload (`fileUpload`)

The legacy UI's **JWT Key Details** section (heading confirmed via Playwright:
`["Choosing an authentication type", "Generate a JWT file", "Authentication", "Authentication type", "JWT Key Details"]`, `fileInputs:1`) renders a Google JWT
(service-account) key uploader — "Drop the Google JWT file here / Click to browse files
(Accepted file type: .json)" — plus **Paste JWT Token** and **Fill In JWT Token manually**
toggles. Uploading/pasting the service-account JSON parses it and fills the JWT credential
fields. The schema modeled the individual JWT fields (defaultProject, clientEmail, tokenUri,
privateKey) but **not the upload control**, so the new UI could only be filled manually.

**Fix (in `dsconfig.json` only):** added a virtual `fileUpload` field with a `fileMapping` that
mirrors the legacy `JWTConfig` distribution (identical to the `stackdriver` fix).

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

- The upload control renders for JWT auth (the default auth type): FileDropzone + **Paste JWT Token** / **Fill In JWT Token manually** buttons (screenshot `googlesheets-fileupload.png`; DOM probe `{"hasDropzone":true,"pasteBtn":true,"manualBtn":true}`).
- Pasting a service-account JSON (`{project_id, client_email, token_uri, private_key, ...}`) distributed the values into the fields, and **Save & Test** produced:

```json
"jsonData":       { "authenticationType": "jwt", "defaultProject": "my-proj",
                    "clientEmail": "sa@my-proj.iam.gserviceaccount.com",
                    "tokenUri": "https://oauth2.googleapis.com/token" },
"secureJsonData": { "privateKey": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n" }
```

i.e. `project_id → jsonData.defaultProject`, `client_email → jsonData.clientEmail`,
`token_uri → jsonData.tokenUri`, `private_key → secureJsonData.privateKey` — matching the legacy
`JWTConfig` behavior. Evidence: `googlesheets-result.json`.

### Note on the legacy `jwt` / `authType` fields — no `fileMapping` change needed

The task flagged that Google Sheets _might_ persist the JWT differently: the schema also carries
a legacy write-only secret `secureJsonData.jwt` and a legacy `jsonData.authType`. The paste test
confirms the **new UI does not use those legacy keys** — the `FileUploadField` distributes the
parsed service-account values straight into the modern individual fields, and the save payload
contains **only** `jsonData.{defaultProject,clientEmail,tokenUri}` + `secureJsonData.privateKey`
(never `secureJsonData.jwt`, never `jsonData.authType`). This is exactly what the backend prefers
(`authType`/`jwt` are backward-compat only; the loader migrates `authType → authenticationType`
and decrypts but ignores `jwt`). The `fileMapping` therefore stays pointed at the modern fields —
**no adjustment required.**

---

## Field-by-field parity

| Legacy field                                          | new (schema id)                                                          | Target                        | Status                           |
| ----------------------------------------------------- | ------------------------------------------------------------------------ | ----------------------------- | -------------------------------- |
| Authentication type (API Key / Google JWT File / GCE) | `jsonData_authenticationType`                                            | `jsonData.authenticationType` | ✅ (radio)                       |
| **JWT file upload**                                   | `virtual_jwtUpload`                                                      | (virtual → distributes)       | ➕ added (`fileUpload`)          |
| API Key (secret)                                      | `secureJsonData_apiKey`                                                  | `secureJsonData.apiKey`       | ✅ 🔀 (key)                      |
| Default project                                       | `jsonData_defaultProject`                                                | `jsonData.defaultProject`     | ✅ 🔀 (jwt/gce)                  |
| Client email / Token URI / Private key path           | `jsonData_clientEmail` / `jsonData_tokenUri` / `jsonData_privateKeyPath` | `jsonData.*`                  | ✅ 🔀 (jwt)                      |
| Private key (secret)                                  | `secureJsonData_privateKey`                                              | `secureJsonData.privateKey`   | ✅ 🔀 (jwt)                      |
| Default Spreadsheet ID                                | `jsonData_defaultSheetID`                                                | `jsonData.defaultSheetID`     | ✅ (Settings)                    |
| _(legacy) auth type_                                  | `jsonData_authType`                                                      | `jsonData.authType`           | ✅ modeled, `legacy` tag (no UI) |
| _(legacy) JWT blob (secret)_                          | `secureJsonData_jwt`                                                     | `secureJsonData.jwt`          | ✅ modeled, `legacy` tag (no UI) |

🔀 = conditionally shown via `dependsOn` on the auth type.

---

## HTTP headers / required — not applicable

- **HTTP headers:** the legacy Google Sheets editor has **no** Custom HTTP Headers section (`hasCustomHeaders:false`, `addHeaderBtn:false`) — it is a cloud-auth datasource that uses the Google SDK, not the generic HTTP proxy. Not added; the new UI likewise shows no headers editor (`hasHeadersEditor:false`) and no URL field (`urlPresent:false`).
- **`required:true` fix:** googlesheets has **no** `requiredWhen:"true"` fields — every required field (apiKey / JWT credentials) is _conditionally_ required on the auth type, which is correct. No change.

## Conditional fields — tested

The `authenticationType` radio drives the visible credential set: `key` → API Key secret;
`jwt` → the upload control + JWT fields (defaultProject, clientEmail, tokenUri, privateKeyPath,
privateKey); `gce` → default project only. The `tab`-mode capture (`newgen-googlesheets-tab.json`)
confirms the rendered sections — `Authentication` and `Settings` (Optional) — and all JWT field
labels, plus the `Paste JWT Token` / `Fill In JWT Token manually` upload affordance.

---

## Verification

```
go generate ./registry/grafana-googlesheets-datasource/...
go test ./registry/grafana-googlesheets-datasource/...   # all subtests PASS
```

Test breakdown (all PASS): `TestSchemaConformance` (8/8 — BaseFieldsResolved, SchemaRoundTrip,
SchemaArtifactInSync, SchemaSpecHasNoSecureJSON, ConfigSchemaValid, JSONDataMatchesStruct,
JSONDataTypesMatchStruct, SecureValuesMatchLoadSettings), `TestLoadConfig` (16),
`TestApplyDefaults` (3), `TestValidate` (10).

The `fileUpload` field is `kind:virtual` (no storage target), so it is skipped by the
jsonData/secure conformance walkers and by the SDK spec converter — no artifact/struct impact
(`git status` shows only `dsconfig.json` changed; `schema.gen.json` / `settings.gen.json` are byte-identical after `go generate`).

## Files changed

- [`registry/grafana-googlesheets-datasource/dsconfig.json`](dsconfig.json) — added the `virtual_jwtUpload` `fileUpload` field (in the `authentication` group, right after the auth-type selector).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `schema.gen.json`, `settings.gen.json` (the virtual field yields no artifact diff), `conformance_test.go`, and `plugin-ui`.
