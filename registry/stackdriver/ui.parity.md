# Google Cloud Monitoring (stackdriver) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `stackdriver` (product name: Google Cloud Monitoring)
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (Grafana Enterprise 13.0.1)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:stackdriver`
- **Method:** Playwright captured both UIs (screenshots + DOM + `Save` console payload); the local edited `dsconfig.json` was served to the new UI via `context.route(...)` interception of the remote schema fetch.
- **Result:** **Parity achieved.** The one missing affordance — the **Google JWT service-account file upload** — was added as a `fileUpload` field and verified to distribute credentials into the JWT fields. No HTTP headers (cloud-auth datasource, not HTTP-proxy). No `required:true` fix needed (all required fields are conditionally required on the auth type).

---

## The gap: Google JWT file upload (`fileUpload`)

The legacy UI's **JWT Key Details** section renders a Google JWT (service-account) key
uploader: "Drop the Google JWT file here / Click to browse files (Accepted file type: .json)",
plus **Paste JWT Token** and **Fill In JWT Token manually** toggles. Uploading/pasting the
service-account JSON parses it and fills the JWT credential fields. The schema modeled the
individual JWT fields (defaultProject, clientEmail, tokenUri, privateKey) but **not the upload
control**, so the new UI could only be filled manually.

**Fix (in `dsconfig.json` only):** added a virtual `fileUpload` field with a `fileMapping` that
mirrors the legacy `JWTConfig` distribution. This is the **first use of `fileUpload` in the
registry** — the `plugin-ui` `FileUploadField` component (whose labels are literally "…JWT
Token") is purpose-built for exactly this control.

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

- The upload control renders for JWT auth: FileDropzone + **Paste JWT Token** / **Fill In JWT Token manually** buttons (screenshot `stackdriver-fileupload.png`).
- Pasting a service-account JSON (`{project_id, client_email, token_uri, private_key, ...}`) distributed the values into the fields, and **Save & Test** produced:

```json
"jsonData":       { "authenticationType": "jwt", "defaultProject": "my-proj",
                    "clientEmail": "sa@my-proj.iam.gserviceaccount.com",
                    "tokenUri": "https://oauth2.googleapis.com/token" },
"secureJsonData": { "privateKey": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n" }
```

i.e. `project_id → jsonData.defaultProject`, `client_email → jsonData.clientEmail`,
`token_uri → jsonData.tokenUri`, `private_key → secureJsonData.privateKey` — matching the legacy
`JWTConfig` behavior.

---

## Field-by-field parity

| Legacy field                                          | new (schema id)                                                             | Target                        | Status                  |
| ----------------------------------------------------- | --------------------------------------------------------------------------- | ----------------------------- | ----------------------- |
| Authentication type (JWT / GCE / WIF / Forward OAuth) | `jsonData_authenticationType`                                               | `jsonData.authenticationType` | ✅ (radio)              |
| **JWT file upload**                                   | `virtual_jwtUpload`                                                         | (virtual → distributes)       | ➕ added (`fileUpload`) |
| Default project                                       | `jsonData_defaultProject`                                                   | `jsonData.defaultProject`     | ✅ 🔀                   |
| Client email / Token URI / Private key path           | `jsonData_clientEmail` / `jsonData_tokenUri` / `jsonData_privateKeyPath`    | `jsonData.*`                  | ✅ 🔀 (jwt)             |
| Private key (secret)                                  | `secureJsonData_privateKey`                                                 | `secureJsonData.privateKey`   | ✅ 🔀 (jwt)             |
| Service-account impersonation (Enable + SA)           | `jsonData_usingImpersonation` / `jsonData_serviceAccountToImpersonate`      | `jsonData.*`                  | ✅ 🔀                   |
| WIF pool provider / SA email                          | `jsonData_workloadIdentityPoolProvider` / `jsonData_wifServiceAccountEmail` | `jsonData.*`                  | ✅ 🔀 (WIF)             |
| Universe domain                                       | `jsonData_universeDomain`                                                   | `jsonData.universeDomain`     | ✅                      |

---

## HTTP headers / required — not applicable

- **HTTP headers:** the legacy Google Cloud Monitoring editor has **no** Custom HTTP Headers section (it is a cloud-auth datasource that uses the Google SDK, not the generic HTTP proxy). Not added.
- **`required:true` fix:** stackdriver has **no** `requiredWhen:"true"` fields — every required field (tenant/WIF provider/etc.) is _conditionally_ required on the auth type, which is correct. No change.

## Conditional fields — tested

The `authenticationType` radio drives the visible credential set: JWT → JWT fields + the upload
control; WIF → pool-provider/SA-email; GCE → impersonation; Forward OAuth → default project only.
All conditionals render per the schema `dependsOn`.

---

## Verification

```
go generate ./registry/stackdriver/...
go test ./registry/stackdriver/...     # 8/8 conformance subtests PASS
```

The `fileUpload` field is `kind:virtual` (no storage target), so it is skipped by the
jsonData/secure conformance walkers and by the SDK spec converter — no artifact/struct impact.
Full `./registry/... ./schema/...` suite passes.

## Files changed

- [`registry/stackdriver/dsconfig.json`](dsconfig.json) — added the `virtual_jwtUpload` `fileUpload` field (in the `authentication` group).
- [`registry/stackdriver/schema.gen.json`](schema.gen.json), [`registry/stackdriver/settings.gen.json`](settings.gen.json) — regenerated by `go generate` (no change from the virtual field, refreshed for consistency).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`.
