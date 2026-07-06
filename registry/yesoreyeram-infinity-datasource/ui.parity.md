# Infinity — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `yesoreyeram-infinity-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (uid `ffrbd04goi7swf`; Infinity's own React `ConfigEditor` + shared `SecureFieldsEditor`)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:yesoreyeram-infinity-datasource` (Storybook, `ConfigEditor/DatasourceConfigWizard`)
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + `Save`-button console payload). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/yesoreyeram-infinity-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the captures reflect the local schema without pushing.
- **Result:** **Parity achieved for the general-HTTP surface.** Two missing editors were found (**Custom HTTP Headers** and **URL Query Params**) and added as `indexedPair` array fields; `fileUpload` was evaluated and correctly **not** used; `required: true` was evaluated and correctly **not** applied (Infinity's Base URL is intentionally optional); header + query-param storage routing was verified from the save payload. The OAuth2-only indexed-pair sets remain out of scope (documented as a follow-up below).

---

## TL;DR of changes

| #   | Change                                                                                                              | File                             | Why                                                                                                 |
| --- | ------------------------------------------------------------------------------------------------------------------- | -------------------------------- | --------------------------------------------------------------------------------------------------- |
| 1   | Added **Custom HTTP Headers** field (`jsonData_httpHeaders`) with `indexedPair` storage                             | [`dsconfig.json`](dsconfig.json) | Legacy renders an **Add Custom HTTP Header** editor; new UI had no headers editor                   |
| 2   | Added **URL Query Params** field (`jsonData_secureQuery`) with `indexedPair` storage                                | [`dsconfig.json`](dsconfig.json) | Legacy renders an **Add URL Query Param** editor; new UI had no query-params editor                 |
| 3   | Appended both field ids to the **`urls-headers-params`** group's `fieldRefs`                                        | [`dsconfig.json`](dsconfig.json) | Surface both editors in the existing "URL, Headers & Params" section                                |
| 4   | Updated the connection/indexed-pair `instructions` entry to note headers + URL query params are now modeled         | [`dsconfig.json`](dsconfig.json) | Keep the embedded LLM instructions truthful after changes 1–2 (oauth2 pairs marked still-unmodeled) |
| 5   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/yesoreyeram-infinity-datasource/...` | generated artifacts              | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`)                                |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `schema/conformance.go`, or `plugin-ui`. All changes flow through `dsconfig.json` with the rest produced by `go generate`. The shared conformance walker already skips `indexedPair` fields (`schema/conformance.go:isIndexedPairField`, landed with the graphite work), so no conformance change was required here.

---

## Section layout

The existing `groups` taxonomy was preserved; the two new editors were appended to the
existing **URL, Headers & Params** section (verified rendering top-to-bottom in the new UI):

| Order | Section (`id`)                                    | `optional` | Relevant fields                                                                                                                                             |
| ----- | ------------------------------------------------- | ---------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1     | **Main** (`main`)                                 | no         | (empty base group)                                                                                                                                          |
| 2     | **Authentication** (`authentication`)             | no         | Auth type + all method credentials, Allowed hosts                                                                                                           |
| 3     | **URL, Headers & Params** (`urls-headers-params`) | yes        | Base URL, Ignore status code check, Allow dangerous HTTP methods, Encode query params, Include cookies, **Custom HTTP Headers ➕**, **URL Query Params ➕** |
| 4     | **Network** (`network`)                           | yes        | Timeout, TLS toggles + certs, Proxy                                                                                                                         |
| 5     | **Security** (`security`)                         | yes        | Query security                                                                                                                                              |
| 6     | **Health check** (`health-check`)                 | yes        | Custom health check enable + URL                                                                                                                            |
| 7     | **Reference data** (`reference-data`)             | yes        | Reference data                                                                                                                                              |
| 8     | **Global queries** (`global-queries`)             | yes        | Global queries                                                                                                                                              |

---

## Field-by-field parity (indexed-pair surface)

Legend: ➕ added by this change · 🔀 dynamically-indexed (row editor)

| Legacy UI editor        | Control (legacy)                                         | New UI (schema id)     | Control (new)                           | Storage target                                                       | Status |
| ----------------------- | -------------------------------------------------------- | ---------------------- | --------------------------------------- | -------------------------------------------------------------------- | ------ |
| **Custom HTTP Headers** | **Add Custom HTTP Header** → name input + value password | `jsonData_httpHeaders` | IndexedPair editor (key + secret value) | `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>`   | ➕ 🔀  |
| **URL Query Params**    | **Add URL Query Param** → key input + value password     | `jsonData_secureQuery` | IndexedPair editor (key + secret value) | `jsonData.secureQueryName<N>` / `secureJsonData.secureQueryValue<N>` | ➕ 🔀  |

The remaining 54 fields (auth methods, TLS, proxy, timeouts, reference data, etc.) were
already modeled and are unchanged by this report.

---

## Gaps found and fixed

### 1. Custom HTTP Headers

**Legacy behaviour (verified, `legacy-expand-p3-infinity.json`):** the config editor exposes
an **Add Custom HTTP Header** button. Each row is a header-name input plus a **password**
value input. Infinity's shared `SecureFieldsEditor` (`src/components/config/SecureFieldsEditor.tsx:98-113`)
persists these as indexed pairs — `jsonData.httpHeaderName<N>` for the name and
`secureJsonData.httpHeaderValue<N>` for the (secret) value, starting at `N = 1`.

**Before (verified, `newgen-infinity-before.json`):** `dsconfig.json` did not model headers
(the embedded instructions explicitly called them "NOT modeled as first-class fields"), so the
new UI rendered no headers editor (`headersEditor: false`).

**Fix (in `dsconfig.json` only):** added the `jsonData_httpHeaders` array field with an
`indexedPair` storage mapping reproducing the exact legacy storage, plus item sub-fields for the
header name (`http.header.name`, with a header-name pattern validation) and value
(`http.header.value`). Placed in the **URL, Headers & Params** group.

### 2. URL Query Params

**Legacy behaviour (verified, `legacy-expand-p3-infinity.json`):** the config editor exposes an
**Add URL Query Param** button. Each row is a key input plus a **password** value input, persisted
by the same `SecureFieldsEditor` as indexed pairs — `jsonData.secureQueryName<N>` /
`secureJsonData.secureQueryValue<N>`, starting at `N = 1`.

**Before:** not modeled — no query-params editor rendered.

**Fix (in `dsconfig.json` only):** added the `jsonData_secureQuery` array field with an
`indexedPair` storage mapping, plus item sub-fields for the key (`http.query.name`) and value
(`http.query.value`). Placed in the **URL, Headers & Params** group.

**After (verified, `newgen-infinity-fixed.json`):** the new UI's **URL, Headers & Params** section
now renders both editors — buttons **Add custom http header** and **Add url query param**, and the
field labels **Custom HTTP Headers** and **URL Query Params**. `headersEditor: true`.

Both fields reuse the storage shape already handled by `settings.go`
(`aggregateSecretPairs("httpHeaderName", "httpHeaderValue")` and
`aggregateSecretPairs("secureQueryName", "secureQueryValue")`, `settings.go:303-304`) and the
existing `customHeadersAndQueryParams` settings example (`schema.go:302-319`), so no Go change was
needed.

---

## `fileUpload` evaluation — not applicable to Infinity

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not, for
Infinity:

- The legacy capture found **0 file inputs** and **0 upload buttons**
  (`legacy-expand-p3-infinity.json`: `"fileInputs": 0`, `"uploadButtons": []`). Infinity's TLS
  cert fields (CA Cert / Client Cert / Client Key) are plain textareas (`-----BEGIN CERTIFICATE-----`).
- The new UI's `fileUpload` component only activates for `ui.fileMapping` (multi-key JSON
  distribution, e.g. a GCP service-account file) — not single-PEM upload.

**Decision:** do **not** add `fileUpload` to any Infinity field.

## `required: true` evaluation — not applicable to Infinity

Infinity's **Base URL** (`root_url`) is intentionally **optional** — the query editor may supply
the full URL per query (frontend writes the `__IGNORE_URL__` sentinel when blank;
`settings.go:22,284-286`). There is no unconditional `requiredWhen: "true"` to convert to
`required: true`, so no required-field fix was made. All other required fields are conditional
(`requiredWhen` tied to the selected auth method / TLS toggle) and stay as-is.

---

## Storage-target routing — verified from the save payload

Driving one custom header (name `X-Tenant`, value `tenant-42`) and one URL query param
(key `trace`, value `trace-id-99`) in the new UI and clicking **Save & Test** logs the exact
datasource payload the wizard would PUT (`infinity-routing-result.json`):

```json
{
  "url": "https://api.example.com",
  "jsonData": {
    "auth_method": "none",
    "httpHeaderName1": "X-Tenant",
    "secureQueryName1": "trace"
  },
  "secureJsonData": {
    "httpHeaderValue1": "tenant-42",
    "secureQueryValue1": "trace-id-99"
  },
  "secureJsonFields": { "httpHeaderValue1": false, "secureQueryValue1": false }
}
```

- **Header** → name in `jsonData.httpHeaderName1`, value in `secureJsonData.httpHeaderValue1`.
- **URL query param** → name in `jsonData.secureQueryName1`, value in `secureJsonData.secureQueryValue1`.

Byte-for-byte the legacy Infinity `SecureFieldsEditor` storage format. Both value inputs render as
`type=password` (secure), matching the `indexedPair` value → `secureJsonData` mapping. The payload
carried no stray/dynamic secrets beyond the two rows entered.

---

## Follow-up (out of scope): OAuth2 indexed-pair sets

Infinity stores two further indexed-pair sets that are **OAuth2-only** and remain **not modeled**
as first-class fields (documented in the schema's `instructions` and left as an explicit
follow-up):

- **OAuth2 endpoint params** → `jsonData.oauth2EndPointParamsName<N>` / `secureJsonData.oauth2EndPointParamsValue<N>`
- **OAuth2 token request headers** → `jsonData.oauth2TokenHeadersName<N>` / `secureJsonData.oauth2TokenHeadersValue<N>`

These are niche (only meaningful for the `oauth2` auth method's token request) and were left out of
scope per the task. `settings.go` already aggregates them (`aggregateSecretPairs`, `settings.go:305-306`),
so modeling them later is purely a `dsconfig.json` addition (two more `indexedPair` fields in the
Authentication group), needing no Go change.

---

## Verification

```
go generate ./registry/yesoreyeram-infinity-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/yesoreyeram-infinity-datasource/...        # PASS
go test ./registry/... ./schema/...                           # entire suite PASS (80 pkg ok, 0 fail — no regressions)
```

Conformance subtests (infinity): `BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`,
`SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings` — all **PASS**. Plus the entry's own `TestLoadConfig`,
`TestApplyDefaults`, `TestValidate` suites — all **PASS**.

Why the parity checks still pass after adding two `indexedPair` fields:

- `JSONDataMatchesStruct` / `JSONDataTypesMatchStruct` skip `indexedPair` fields (they are logical
  views over dynamic keys, not single struct fields).
- `SecureValuesMatchLoadSettings` is unaffected: the array fields' `target` is `jsonData`, and their
  item value sub-fields are nested (not top-level `secureJsonData` fields), so the schema's static
  secure-key set stays at the 12 keys in `SecureJsonDataKeys`. The generated `settings.gen.json`
  `secureValues` list confirms it still contains only those 12 static secrets — no
  `httpHeaderValue<N>` / `secureQueryValue<N>` leaked in.

---

## Files changed

- [`registry/yesoreyeram-infinity-datasource/dsconfig.json`](dsconfig.json) — added `jsonData_httpHeaders` and `jsonData_secureQuery` `indexedPair` array fields; appended both ids to the `urls-headers-params` group's `fieldRefs`; updated the connection/indexed-pair `instructions` entry (headers + URL query params now modeled; OAuth2 pairs marked still-unmodeled follow-up).
- [`registry/yesoreyeram-infinity-datasource/schema.gen.json`](schema.gen.json), [`registry/yesoreyeram-infinity-datasource/settings.gen.json`](settings.gen.json) — regenerated by `go generate`.

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `schema/conformance.go`, `settings.examples.gen.json`, and `plugin-ui`.
