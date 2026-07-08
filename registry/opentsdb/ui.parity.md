# OpenTSDB — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `opentsdb`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (Grafana Enterprise 13.x, `@grafana/ui` `DataSourceHttpSettings`) — captured with UID `bfras2sznvl6of`
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:opentsdb` (Storybook, `ConfigEditor/DatasourceConfigWizard`)
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + `Save`-button console payload). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/opentsdb/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the captures reflect the local schema without pushing.
- **Result:** **Parity achieved.** One missing field was found (**Custom HTTP Headers**) and added; `fileUpload` was evaluated and correctly **not** used; the direct auth toggles, TLS conditionals, and OpenTSDB Version/Resolution selects were all exercised and their storage targets verified from the save payload.

---

## TL;DR of changes

| #   | Change                                                                                                                  | File                             | Why                                                                                                                             |
| --- | ----------------------------------------------------------------------------------------------------------------------- | -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| 1   | `root_url`: `"requiredWhen": "true"` → `"required": true`                                                               | [`dsconfig.json`](dsconfig.json) | Puts URL in the wizard's synthetic **General** step and emits OpenAPI `required: ["url"]` instead of `x-dsconfig-required-when` |
| 2   | Added **Custom HTTP Headers** field (`jsonData_httpHeaders`) with `indexedPair` storage; appended to the **HTTP** group | [`dsconfig.json`](dsconfig.json) | Legacy UI renders `CustomHeadersSettings`; new UI had no headers editor                                                         |
| 3   | Rewrote the secure-values instruction to describe headers as **modeled** (was "not modeled here (see README)")          | [`dsconfig.json`](dsconfig.json) | Keep the embedded LLM instructions truthful after change #2                                                                     |
| 4   | Regenerated `schema.gen.json` + `settings.gen.json` via `go generate`                                                   | generated artifacts              | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`)                                                            |

**No Go / plugin-ui / conformance changes were required for opentsdb** (see [Why no shared-code changes](#why-no-shared-code-changes-were-needed) below). No changes were made to `settings.go`, `settings.ts`, or `README.md` — per the constraint that all schema changes flow through `dsconfig.json` with the rest produced by `go generate`.

---

## Section layout

The `groups` were left in their existing OpenTSDB taxonomy; only the new **Custom
HTTP Headers** field was appended to the **HTTP** group. This mirrors the legacy
editor, which shows exactly four sections in this order (verified in the legacy
capture — headings `["HTTP","Auth","Custom HTTP Headers","OpenTSDB settings"]`):

| Order | Section (`id`)                              | `optional` | Fields (in display order)                                                                                            |
| ----- | ------------------------------------------- | ---------- | -------------------------------------------------------------------------------------------------------------------- |
| 1     | **HTTP** (`http`)                           | no         | URL, Allowed cookies, Timeout, **Custom HTTP Headers** (added)                                                       |
| 2     | **Auth** (`auth`)                           | no         | Basic auth, With Credentials, TLS Client Auth, With CA Cert, Skip TLS Verify, Forward OAuth Identity, User, Password |
| 3     | **TLS/SSL Auth Details** (`tls-details`)    | yes        | ServerName, CA Cert, Client Cert, Client Key                                                                         |
| 4     | **OpenTSDB settings** (`opentsdb-settings`) | no         | Version, Resolution, Lookup limit                                                                                    |

Notes:

- The legacy editor places the header editor between Auth and the OpenTSDB-specific
  settings; in the new schema-driven UI it renders under the **HTTP** group (alongside
  the other HTTP transport fields URL / Allowed cookies / Timeout), which is the
  conventional home for HTTP transport settings and keeps both UIs functionally
  equivalent.
- The auth group uses the short **`id: "auth"`** — the `plugin-ui` wizard's
  required-fields resolver and Authorization-header helper already key off `auth`, so
  the auth fields fold into the wizard's synthetic **General** step without any
  plugin-ui change.

### Wizard mode: URL in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b)
every field in the group whose id is `auth`, plus their `dependsOn` parents/children.

Only one adjustment was needed for opentsdb:

- `root_url` previously used `requiredWhen: "true"` (a CEL expression the resolver does
  **not** inspect), so no General step was created and the wizard opened on `HTTP`.
  Changing it to `"required": true` (unconditionally required — the backend reads
  `settings.URL` at `pkg/opentsdb/opentsdb.go:47` and CheckHealth probes
  `/api/suggest`) puts URL into General and emits a proper OpenAPI `required: ["url"]`.

Result (verified in capture `newgen-opentsdb-wiz`): the wizard opens on **General 1/5**
containing URL, Basic auth, With Credentials, TLS Client Auth, With CA Cert, Skip TLS
Verify, and Forward OAuth Identity. Tab mode is unaffected — the synthetic `_required`
group is filtered out there, so it still shows the four sections in order (verified in
capture `newgen-opentsdb-tab`: sections `HTTP · Auth · TLS/SSL Auth Details · OpenTSDB settings`).

### Why no shared-code changes were needed

Unlike graphite (which used `id: "authentication"` and required a `plugin-ui`
generalization so the resolver recognised the long id), opentsdb already uses the short
`id: "auth"` that the resolver has always matched — so **no `plugin-ui` change** is
needed here. And the `indexedPair` conformance-walker fix
(`isIndexedPairField()` in [`schema/conformance.go`](../../schema/conformance.go)) was
**already in place** from the graphite work, so adding an `indexedPair` field to
opentsdb needed **no conformance change**. Both `JSONDataMatchesStruct` and
`JSONDataTypesMatchStruct` skip `indexedPair` fields, so `settings.go` does not need a
backing struct field for the dynamic `httpHeaderName<N>` keys.

---

## Field-by-field parity

Legend: ✅ present & matching · ➕ added by this change · 🔀 conditional (revealed by `dependsOn`)

| Legacy UI field         | Control (legacy)                         | New UI (schema id)                 | Control (new)                           | Storage target                                                     | Status |
| ----------------------- | ---------------------------------------- | ---------------------------------- | --------------------------------------- | ------------------------------------------------------------------ | ------ |
| URL                     | text input                               | `root_url`                         | input                                   | `root.url`                                                         | ✅     |
| Allowed cookies         | TagsInput                                | `jsonData_keepCookies`             | list (string array)                     | `jsonData.keepCookies`                                             | ✅     |
| Timeout                 | number                                   | `jsonData_timeout`                 | number                                  | `jsonData.timeout`                                                 | ✅     |
| Basic auth              | switch                                   | `root_basicAuth`                   | switch                                  | `root.basicAuth`                                                   | ✅     |
| With Credentials        | switch                                   | `root_withCredentials`             | switch                                  | `root.withCredentials`                                             | ✅     |
| TLS Client Auth         | switch                                   | `jsonData_tlsAuth`                 | switch                                  | `jsonData.tlsAuth`                                                 | ✅     |
| With CA Cert            | switch                                   | `jsonData_tlsAuthWithCACert`       | switch                                  | `jsonData.tlsAuthWithCACert`                                       | ✅     |
| Skip TLS Verify         | switch                                   | `jsonData_tlsSkipVerify`           | switch                                  | `jsonData.tlsSkipVerify`                                           | ✅     |
| Forward OAuth Identity  | switch                                   | `jsonData_oauthPassThru`           | switch                                  | `jsonData.oauthPassThru`                                           | ✅     |
| User                    | text input                               | `root_basicAuthUser`               | input                                   | `root.basicAuthUser`                                               | ✅ 🔀  |
| Password                | password (`SecretFormField`)             | `secureJsonData_basicAuthPassword` | secure input                            | `secureJsonData.basicAuthPassword`                                 | ✅ 🔀  |
| ServerName              | text input                               | `jsonData_serverName`              | input                                   | `jsonData.serverName`                                              | ✅ 🔀  |
| CA Cert                 | **textarea**                             | `secureJsonData_tlsCACert`         | secure input¹                           | `secureJsonData.tlsCACert`                                         | ✅ 🔀  |
| Client Cert             | **textarea**                             | `secureJsonData_tlsClientCert`     | secure input¹                           | `secureJsonData.tlsClientCert`                                     | ✅ 🔀  |
| Client Key              | **textarea**                             | `secureJsonData_tlsClientKey`      | secure input¹                           | `secureJsonData.tlsClientKey`                                      | ✅ 🔀  |
| **Custom HTTP Headers** | Add header → name input + value password | `jsonData_httpHeaders`             | IndexedPair editor (key + secret value) | `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>` | ➕ 🔀  |
| Version                 | select (`<=2.1`/`==2.2`/`==2.3`/`==2.4`) | `jsonData_tsdbVersion`             | select                                  | `jsonData.tsdbVersion` (number)                                    | ✅     |
| Resolution              | select (second/millisecond)              | `jsonData_tsdbResolution`          | select                                  | `jsonData.tsdbResolution` (number)                                 | ✅     |
| Lookup limit            | number                                   | `jsonData_lookupLimit`             | input (number)                          | `jsonData.lookupLimit`                                             | ✅     |

¹ **Not a discrepancy.** The `dsconfig.json` declares `ui.component: "textarea"` for the
three TLS cert fields, but the new renderer checks `target === "secureJsonData"` _before_
the `textarea` branch, so any secure field is drawn as a masked secure input with a
show/hide toggle. Both UIs collect the same PEM text into the same `secureJsonData` keys;
only the widget affordance differs. This is a renderer policy in `plugin-ui`, not a schema gap.

---

## Gap found and fixed: Custom HTTP Headers

**Legacy behaviour (verified in capture `legacy-expand-*-opentsdb`):** the editor
includes a **Custom HTTP Headers** section with an **Add header** button
(`hasCustomHeaders: true`, `addHeaderBtn: true`). Adding a header shows a header-name
text input (placeholder `X-Custom-Header`) and a header-value **password** input.
`@grafana/ui`'s `CustomHeadersSettings` persists these as indexed pairs —
`jsonData.httpHeaderName<N>` for the name and `secureJsonData.httpHeaderValue<N>` for the
(secret) value, starting at `N = 1`.

**Before:** `dsconfig.json` did not model headers at all (the old instruction/README
explicitly excluded them), so the new UI rendered no headers section.

**Fix (in `dsconfig.json` only):** added the `jsonData_httpHeaders` array field with an
`indexedPair` storage mapping that reproduces the exact legacy storage, plus item
sub-fields for the header name (`http.header.name`, with a header-name pattern
validation) and value (`http.header.value`). It is appended to the **HTTP** group:

```jsonc
{
  "id": "jsonData_httpHeaders",
  "key": "httpHeaders",
  "label": "Custom HTTP Headers",
  "valueType": "array",
  "target": "jsonData",
  "role": "http.header",
  "item": {
    "valueType": "object",
    "fields": [
      /* name, value item fields */
    ],
  },
  "storage": {
    "type": "indexedPair",
    "key": { "target": "jsonData", "pattern": "httpHeaderName{index}" },
    "value": {
      "target": "secureJsonData",
      "pattern": "httpHeaderValue{index}",
    },
    "startIndex": 1,
  },
}
```

**After (verified in captures `newgen-opentsdb-tab` + `opentsdb-headers-01-filled`):** the
new UI renders a **Custom HTTP Headers** section under HTTP with an **Add custom http
header** button (`hasHeadersEditor: true`) and a key/secret-value row editor.

---

## `fileUpload` evaluation — not applicable to opentsdb

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not,
for opentsdb:

- The legacy `TLSAuthSettings` renders the CA Cert / Client Cert / Client Key fields as
  **plain textareas** (placeholders `Begins with -----BEGIN CERTIFICATE-----` /
  `-----BEGIN RSA PRIVATE KEY-----`). No file-upload button and no `<input type="file">`
  were found in the legacy DOM (capture `legacy-expand-*-opentsdb`: `fileInputs: 0`,
  `uploadButtons: []`).
- The new UI's `fileUpload` component only activates when a field declares
  `ui.fileMapping` (multi-key JSON distribution, e.g. a GCP service-account file) — it
  does not model single-PEM upload.

**Decision:** do **not** add `fileUpload` to any opentsdb field. The cert fields keep
their current modeling; both UIs collect PEM text into the same `secureJsonData` keys.

---

## Conditional fields & effects — tested

All `dependsOn` conditionals reveal/route correctly in the new UI, confirmed from the
`Save` console payload:

| Trigger                              | Revealed field(s)                   | Verified                                                        |
| ------------------------------------ | ----------------------------------- | --------------------------------------------------------------- |
| `root_basicAuth == true`             | User, Password                      | ✅ `basicAuthUser` → root, `basicAuthPassword` → secureJsonData |
| `jsonData_tlsAuth == true`           | ServerName, Client Cert, Client Key | ✅ TLS/SSL Auth Details activates                               |
| `jsonData_tlsAuthWithCACert == true` | CA Cert                             | ✅                                                              |

**Effects:** opentsdb's schema contains **no** `effects` blocks. Like graphite, its auth
model is a set of independent **direct toggles** (`basicAuth`, `withCredentials`,
`oauthPassThru`) — **not** a virtual `authMethod` selector that fans out to multiple
fields — so there is nothing for `effects` to model, and none were added. The OpenTSDB
Version / Resolution / Lookup limit fields are plain selects/inputs with no cross-field
effects.

### Save-payload storage-target validation

Filling the form (URL + one custom header `X-Api-Token` / `super-secret-token`) and
clicking **Save & Test** logged the exact datasource payload the wizard would PUT
(capture `opentsdb-headers-result`):

```json
{
  "url": "http://opentsdb.example.com:4242",
  "basicAuth": false,
  "withCredentials": false,
  "jsonData": {
    "tlsAuth": false,
    "tlsAuthWithCACert": false,
    "tlsSkipVerify": false,
    "oauthPassThru": false,
    "httpHeaderName1": "X-Api-Token",
    "tsdbVersion": 1,
    "tsdbResolution": 1,
    "lookupLimit": 1000
  },
  "secureJsonData": { "httpHeaderValue1": "super-secret-token" },
  "secureJsonFields": { "httpHeaderValue1": false }
}
```

The header **name lands in `jsonData.httpHeaderName1`** and the **value in
`secureJsonData.httpHeaderValue1`** — byte-for-byte the legacy `CustomHeadersSettings`
storage format. The OpenTSDB `Version` and `Resolution` selects write their numeric
defaults (`tsdbVersion: 1`, `tsdbResolution: 1`) and `lookupLimit: 1000`, confirming both
selects render and route to `jsonData`. All other fields route to `root` / `jsonData` /
`secureJsonData` exactly as declared.

---

## Verification

```
go generate ./registry/opentsdb/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/opentsdb/...        # 8/8 conformance subtests PASS (+ settings unit tests)
go test ./schema/... ./registry/opentsdb/...   # PASS (no regressions)
gofmt -l schema/ registry/opentsdb/    # clean
```

Conformance subtests (opentsdb): `BaseFieldsResolved`, `SchemaRoundTrip`,
`SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`,
`JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings` — all
**PASS**. The `httpHeaders` `indexedPair` field is skipped by the jsonData↔struct parity
checks (via the pre-existing `isIndexedPairField()` walker helper), and the SDK converter
emits `httpHeaders` as a clean array under `jsonData` while keeping `secureValues` limited
to the four static secrets (`basicAuthPassword`, `tlsCACert`, `tlsClientCert`,
`tlsClientKey`).

---

## Files changed

- [`registry/opentsdb/dsconfig.json`](dsconfig.json) — changed `root_url` from
  `requiredWhen: "true"` to `required: true` (renders in the wizard's General step);
  added the `jsonData_httpHeaders` field and appended it to the HTTP group; rewrote the
  secure-values instruction to describe headers as modeled.
- [`registry/opentsdb/schema.gen.json`](schema.gen.json),
  [`registry/opentsdb/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`url` now in the spec's `required` array; `httpHeaders` array added
  under `jsonData`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`,
`settings.examples.gen.json`, `schema/conformance.go` (the `indexedPair` walker fix was
already present), and `plugin-ui` (the `auth` group id is already recognised).

```

```
