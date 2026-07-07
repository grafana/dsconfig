# MongoDB — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-mongodb-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/efrbqixougikgc` (Grafana Enterprise; MongoDB config editor with a `@grafana/plugin-ui`-style Connection / Authentication / TLS / **HTTP headers** / Additional Settings layout)
- **New UI:** `http://192.168.1.241:58899/iframe.html?args=pluginType:grafana-mongodb-datasource&id=configeditor-datasourceconfigwizard--tab` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + `Save & Test` console payloads). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/grafana-mongodb-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the screenshots reflect the local schema without pushing.
- **Result:** **Parity achieved for the in-scope fields.** Custom HTTP Headers **were applicable** (confirmed in the legacy editor) and were added; `requiredWhen: "true"` on the connection string was corrected to `required: true`; `fileUpload` was evaluated and correctly **not** used (0 file inputs in legacy). One **out-of-scope** discrepancy is documented below (legacy renders a **TLS settings** toggle section that the schema currently tags `backend-only`); it was **not** changed per the task's explicit scope.

---

## Headers applicable? **YES**

The legacy MongoDB editor genuinely renders a generic **Custom HTTP Headers** section — this is **not** a false positive:

- The capture collects section headings only from `<h1>`–`<h6>`/`<legend>` elements, and `HTTP headers` appears as a real section heading (not stray body text).
- There is an actual **Add header** button (`addHeaderBtn: true`), a dedicated subtitle ("Pass along additional context and metadata about the request/response"), and `hasCustomHeaders: true`.
- `fileInputs: 0`, `uploadButtons: []` → no `fileUpload`.

Legacy capture (`legacy-expand-mongodb-parity.json`, UID `efrbqixougikgc`):

```json
{
  "hasCustomHeaders": true,
  "addHeaderBtn": true,
  "uploadButtons": [],
  "fileInputs": 0,
  "headings": ["Connection", "Authentication", "Authentication methods", "TLS settings",
               "HTTP headers", "Additional Settings", "Query Syntax Validation",
               "TLS CA Key File Password", "Backend Response Rows Limit"],
  "relevantButtons": ["Add header"]
}
```

Screenshot `legacy-expand-mongodb-parity.png` shows the **HTTP headers** section (with **+ Add header**) sitting between **TLS settings** and **Additional Settings**.

---

## TL;DR of changes

| #   | Change                                                                                                     | File                             | Why                                                                                                                          |
| --- | -------------------------------------------------------------------------------------------------------- | -------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| 1   | Changed `jsonData_connection` from `requiredWhen: "true"` → `required: true`                              | [`dsconfig.json`](dsconfig.json) | Connection string is unconditionally required; puts it into the wizard's synthetic **General** step and emits OpenAPI `required: ["connection"]` |
| 2   | Added **Custom HTTP Headers** field (`jsonData_httpHeaders`) with `indexedPair` storage (verbatim from prometheus) | [`dsconfig.json`](dsconfig.json) | Legacy UI renders an **HTTP headers** section with **Add header**; the new UI had no headers editor                          |
| 3   | Added `jsonData_httpHeaders` to the `additional-settings` group's `fieldRefs` (after `jsonData_responseRowsLimit`) | [`dsconfig.json`](dsconfig.json) | Surface the new field under **Additional Settings** (per task instruction)                                                   |
| 4   | Appended a headers note to the secure-values `instructions` entry (headers are now **modeled**)            | [`dsconfig.json`](dsconfig.json) | Keep the embedded `llm`-tagged instructions truthful after change #2                                                         |
| 5   | Regenerated `schema.gen.json` / `settings.gen.json` via `go generate ./registry/grafana-mongodb-datasource/...` | generated artifacts              | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`)                                                         |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `schema/conformance.go`, or `plugin-ui`. All schema changes flow through `dsconfig.json`; the rest is produced by `go generate`.

**No conformance-test or plugin-ui change was required:**

- **`conformance.go`** already skips `indexedPair` fields in the jsonData↔struct parity checks (`isIndexedPairField`, conformance.go:157/180/245-246) — a generic fix landed during the graphite work — so no Go change was needed here.
- **`plugin-ui`** already recognises the conventional group id `authentication` in `resolveRequiredFieldsGroup`, so the wizard folds MongoDB's auth group into **General** without any change.
- **`settings.go`** already omits the dynamic `httpHeaderValue<N>` secrets from `SecureJsonDataKeys` (settings.go:71-79), which matches the `indexedPair` model exactly.

---

## Section layout

The existing 3 `groups` were kept as-is; only the new field was slotted into **Additional Settings**.
Verified rendering top-to-bottom in the new UI (tab mode, screenshot `newgen-mongodb-tab`):

| Order | Section (`id`)                              | `optional` | Fields (in display order)                                                                    |
| ----- | ------------------------------------------- | ---------- | -------------------------------------------------------------------------------------------- |
| 1     | **Connection** (`connection`)               | no         | Connection string \*                                                                         |
| 2     | **Authentication** (`authentication`)       | no         | Authentication method (radio) → User → Password → (Kerberos: User / Password / KeyTab / Global ccache / Ccache lookup) |
| 3     | **Additional Settings** (`additional-settings`) | yes    | Enable syntax validation, Password (TLS CA Key File Password), Rows to Return, **Custom HTTP Headers** ➕ |

Notes:

- **Legacy grouping vs. new grouping.** The legacy editor renders **HTTP headers** as its own section (between **TLS settings** and **Additional Settings**). The schema folds it into **Additional Settings** (per the task instruction to add it to the `additional-settings` group), so it renders as the last row of that section with an **Add custom http header** button. All fields are present; only the nesting/placement differs — the same treatment used by other entries (e.g. jaeger).
- The MongoDB auth field is a **discriminator radio** (`jsonData_authType`, `role: auth.discriminator`), not a virtual selector — so there are no `effects` to test. The three methods (No Authentication / Credentials / Kerberos) map directly to `jsonData.authType` values `NoAuth` / `BasicAuth` / `custom-Kerberos`.
- Many `jsonData`/`secureJsonData` fields are `backend-only` (serverName, tlsAuth, tlsAuthWithCACert, tlsSkipVerify, tlsCACert, tlsClientCert, tlsClientKey) or `legacy` (user, skipTLSValidation, credentials, password) and belong to no group — see the **out-of-scope TLS observation** below.

### Wizard mode: connection string in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the fields
of the auth group, plus their `dependsOn` parents/children.

- **Before:** `jsonData_connection` used `requiredWhen: "true"` (a CEL expression the resolver does **not** inspect), so the connection string was **not** in General.
- **After:** changing it to `required: true` (unconditionally required — the backend hard-fails on an empty connection string, `settings.go` `Validate` / `pkg/datasource/client.go:134-137`) puts it into General and emits a proper OpenAPI `required: ["connection"]` in the generated spec (replacing the `x-dsconfig-required-when: "true"` extension).

**Verified (screenshot `newgen-mongodb-wiz.png`, DOM `newgen-mongodb-wiz.json`):** the wizard opens on **General 1/4** containing **Connection string \*** (with the required asterisk, `mongodb+srv://…` placeholder) and the **Authentication method** radio (defaulting to "Credentials"), plus User / Password.

---

## Field-by-field parity

Legend: ✅ present & matching · ➕ added by this change · 🔀 conditional (revealed by `dependsOn`) · 🎛 auth discriminator · 🔒 backend-only (provisioning; no editor UI in the new UI) · ⚠️ legacy renders it but the schema tags it backend-only (out-of-scope gap)

| Legacy UI field                 | Control (legacy)                    | New UI (schema id)                              | Control (new)             | Storage target                                                    | Status  |
| ------------------------------- | ----------------------------------- | ----------------------------------------------- | ------------------------- | ---------------------------------------------------------------- | ------- |
| Connection string              | text input (`*`)                    | `jsonData_connection`                           | input                     | `jsonData.connection` (required)                                 | ✅ (now `required: true`) |
| Authentication method          | select (No Auth / Credentials / Kerberos) | `jsonData_authType`                       | radio                     | `jsonData.authType`                                              | ✅ 🎛   |
| User                           | text input                          | `root_basicAuthUser`                            | input                     | `root.basicAuthUser`                                            | ✅ 🔀 (`authType == 'BasicAuth'`) |
| Password                       | password (secure)                   | `secureJsonData_basicAuthPassword`              | secure input              | `secureJsonData.basicAuthPassword`                             | ✅ 🔀   |
| User (Kerberos)                | text input                          | `jsonData_kerberosUser`                         | input                     | `jsonData.kerberosUser`                                         | ✅ 🔀 (`authType == 'custom-Kerberos'`) |
| Password (Kerberos)            | password (secure)                   | `secureJsonData_kerberosPassword`               | secure input              | `secureJsonData.kerberosPassword`                             | ✅ 🔀   |
| KeyTab file path               | text input                          | `jsonData_keyTabFilePath`                       | input                     | `jsonData.keyTabFilePath`                                       | ✅ 🔀   |
| Global ccache file path        | text input                          | `jsonData_globalCcacheFilePath`                 | input                     | `jsonData.globalCcacheFilePath`                                | ✅ 🔀   |
| Ccache lookup file             | text input                          | `jsonData_ccacheLookupFile`                     | input                     | `jsonData.ccacheLookupFile`                                    | ✅ 🔀   |
| **HTTP headers** (Add header)  | name input + value password         | `jsonData_httpHeaders`                          | IndexedPair editor (key + secret value) | `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>` | ➕ 🔀   |
| Enable syntax validation       | switch                              | `jsonData_validate`                             | switch                    | `jsonData.validate`                                             | ✅      |
| TLS CA Key File Password       | password (secure)                   | `secureJsonData_tlsCertificateKeyFilePassword`  | secure input              | `secureJsonData.tlsCertificateKeyFilePassword`               | ✅      |
| Rows to Return                 | text input                          | `jsonData_responseRowsLimit`                    | input                     | `jsonData.responseRowsLimit`                                   | ✅      |
| TLS settings → Add self-signed certificate | checkbox                | `jsonData_tlsAuthWithCACert`                    | — (backend-only)          | `jsonData.tlsAuthWithCACert`                                   | ⚠️ (see below) |
| TLS settings → TLS Client Authentication   | checkbox                | `jsonData_tlsAuth`                              | — (backend-only)          | `jsonData.tlsAuth`                                             | ⚠️ (see below) |
| TLS settings → Skip TLS certificate validation | checkbox            | `jsonData_tlsSkipVerify`                        | — (backend-only)          | `jsonData.tlsSkipVerify`                                       | ⚠️ (see below) |
| — (revealed by TLS toggles)    | textarea / input                    | `secureJsonData_tlsCACert` / `secureJsonData_tlsClientCert` / `secureJsonData_tlsClientKey` / `jsonData_serverName` | — (backend-only) | `secureJsonData.*` / `jsonData.serverName` | 🔒 ⚠️ |
| — (not in editor)              | —                                   | `jsonData_user` / `jsonData_skipTLSValidation` / `jsonData_credentials` / `secureJsonData_password` | — | legacy migration fields | 🔒 (legacy) |

---

## Gap found and fixed: Custom HTTP Headers

**Legacy behaviour (verified in `legacy-expand-mongodb-parity`):** the editor includes an **HTTP
headers** section heading with an **Add header** button (`hasCustomHeaders: true`, `addHeaderBtn:
true`). The plugin-ui CustomHeaders component persists headers as indexed pairs —
`jsonData.httpHeaderName<N>` for the name and `secureJsonData.httpHeaderValue<N>` for the (secret)
value, starting at `N = 1`.

**Before:** `dsconfig.json` did not model headers at all, so the new UI rendered no headers editor.

**Fix (in `dsconfig.json` only):** added the `jsonData_httpHeaders` array field (copied verbatim from
`registry/prometheus/dsconfig.json`) with an `indexedPair` storage mapping that reproduces the exact
legacy storage, plus item sub-fields for the header name (`http.header.name`, with a header-name
pattern validation) and value (`http.header.value`), and referenced it from the **Additional
Settings** group.

**After (verified in `newgen-mongodb-tab`):** the new UI renders a **Custom HTTP Headers** row under
**Additional Settings** with an **Add custom http header** button and a key/secret-value editor
(`hasHeadersEditor: true`).

### Save-payload storage-target validation

Filling the connection string + one custom header (name `X-Api-Token`, value `super-secret-token`)
and clicking **Save & Test** logged the exact payload the wizard would PUT
(`mongodb-headers-route.json`):

```json
{
  "jsonData": {
    "connection": "mongodb://mongodb.example.com:27017/mydb",
    "authType": "BasicAuth",
    "httpHeaderName1": "X-Api-Token"
  },
  "secureJsonData": { "httpHeaderValue1": "super-secret-token" },
  "secureJsonFields": { "httpHeaderValue1": false, "basicAuthPassword": false, ... }
}
```

The header **name lands in `jsonData.httpHeaderName1`** and the **value in
`secureJsonData.httpHeaderValue1`** (with `secureJsonFields.httpHeaderValue1: false`) — byte-for-byte
the legacy CustomHeaders storage format. The connection string routes to `jsonData.connection`; no
secure value leaks into `jsonData`.

---

## `fileUpload` evaluation — not applicable to MongoDB

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not:

- No file-upload button and no `<input type="file">` were found in the legacy DOM
  (`legacy-expand-mongodb-parity.json`: `fileInputs: 0`, `uploadButtons: []`). The Kerberos keytab /
  ccache inputs are plain **text paths** (`/tmp/example.keytab`, `/tmp/krb5cc_1000`), and the TLS
  cert material is provisioning-only.
- The new UI's `fileUpload` component only activates for `ui.fileMapping` (multi-key JSON
  distribution); it does not model single-PEM or file-path upload.

**Decision:** do **not** add `fileUpload` to any MongoDB field.

---

## Out-of-scope observation: legacy "TLS settings" toggles (not changed)

**Not fixed — outside this task's scope** (which was limited to the `requiredWhen`→`required` fix and
Custom HTTP Headers).

The legacy editor renders a **TLS settings** section (screenshot `legacy-expand-mongodb-parity.png`)
with three checkboxes — **Add self-signed certificate** (`tlsAuthWithCACert`), **TLS Client
Authentication** (`tlsAuth`), **Skip TLS certificate validation** (`tlsSkipVerify`) — which, when
enabled, reveal the CA cert / client cert / client key / ServerName inputs. In the current schema all
of these fields are tagged `backend-only` with descriptions stating "Not exposed in the configuration
editor; set via datasource provisioning", so the **new UI does not render a TLS settings section**.

This is a genuine parity gap, but modeling it would require un-tagging 4+ fields, adding a `tls-settings`
group, and wiring the dependent secure cert fields with `dependsOn`/`requiredWhen` — a change beyond
the two edits this task authorized. It is **flagged here as a recommended follow-up** and left
untouched. (It is fixable via `dsconfig.json` alone; it just wasn't in scope.)

---

## Conformance — no Go change needed

Adding an `indexedPair` field is already handled by the shared conformance suite:
`JSONDataMatchesStruct` / `JSONDataTypesMatchStruct` skip `indexedPair` fields (`isIndexedPairField`,
conformance.go:157/180/245-246), and `SecureValuesMatchLoadSettings` only walks top-level
`secureJsonData` fields — the per-header `httpHeaderValue<N>` secret lives inside `storage.value`, not
as a top-level field, so the static `SecureJsonDataKeys` list (settings.go:71-79) stays correct. The
generated spec emits `httpHeaders` as a clean array under `jsonData` with no secure values leaked
(`SchemaSpecHasNoSecureJSON` passes).

---

## Verification

```
go generate ./registry/grafana-mongodb-datasource/...      # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-mongodb-datasource/...          # PASS
go test ./registry/grafana-mongodb-datasource/... ./schema/...  # PASS (no regressions)
```

`TestSchemaConformance` subtests (mongodb) — all **PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `TestLoadConfig`, `TestApplyDefaults`, `TestValidate`, `TestKerberosEnabled`,
`TestBasicAuthPassword`, and `TestSettingsExamplesShape` suites also pass unchanged (they do not
reference the new field).

Playwright evidence (in the shared capture dir):

- `legacy-expand-mongodb-parity` (legacy inventory, UID `efrbqixougikgc`) — `hasCustomHeaders: true`, `addHeaderBtn: true`, `fileInputs: 0`, `uploadButtons: []`; **HTTP headers** section visible between TLS settings and Additional Settings.
- `newgen-mongodb-tab` (tab) — Connection / Authentication / Additional Settings render; `hasHeadersEditor: true`; **Custom HTTP Headers** with **Add custom http header** in Additional Settings.
- `newgen-mongodb-wiz` (wizard) — opens on **General 1/4** with **Connection string \*** (required) + Authentication method.
- `mongodb-headers-route` — header save-payload routing verified (`httpHeaderName1` in jsonData, `httpHeaderValue1` in secureJsonData, connection in jsonData.connection).

---

## Files changed

- [`registry/grafana-mongodb-datasource/dsconfig.json`](dsconfig.json) — changed `jsonData_connection`
  from `requiredWhen: "true"` to `required: true`; added the `jsonData_httpHeaders` `indexedPair` field
  and referenced it from the `additional-settings` group; appended a "headers are now modeled" note to
  the secure-values `instructions` entry.
- [`registry/grafana-mongodb-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-mongodb-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`connection` now in the spec's `required` array; `x-dsconfig-required-when: "true"`
  removed from `connection`; `httpHeaders` array added under `jsonData`).

_Unchanged by design / constraint:_ `settings.go`, `settings.ts`, `README.md`,
`settings.examples.gen.json`, `schema/conformance.go`, and everything in `plugin-ui`. The legacy
**TLS settings** toggle section is a documented out-of-scope gap (see above), not addressed here.
