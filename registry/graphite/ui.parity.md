# Graphite — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `graphite`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (Grafana Enterprise 13.0.1, `@grafana/ui` `DataSourceHttpSettings`)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:graphite` (Storybook, `ConfigEditor/DatasourceConfigWizard`)
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + `Save`-button console payload). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/graphite/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the screenshots reflect the local schema without pushing.
- **Result:** **Parity achieved.** One missing field was found (**Custom HTTP Headers**) and added; `fileUpload` was evaluated and correctly **not** used; all conditional fields were exercised and their storage targets verified from the save payload.

---

## TL;DR of changes

| #   | Change                                                                                                                  | File                                                         | Why                                                                                                         |
| --- | ----------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------ | ----------------------------------------------------------------------------------------------------------- |
| 1   | Added **Custom HTTP Headers** field (`jsonData_httpHeaders`) with `indexedPair` storage | [`dsconfig.json`](dsconfig.json)                             | Legacy UI renders `CustomHeadersSettings`; new UI had no headers editor                                     |
| 2   | Updated the secure-values instruction to describe headers as modeled (was "not modeled here")                           | [`dsconfig.json`](dsconfig.json)                             | Keep the embedded LLM instructions truthful after change #1                                                 |
| 3   | Restructured `groups` into a standard taxonomy: **Connection · Authentication · Network & TLS · Graphite settings · Query settings · Advanced settings** | [`dsconfig.json`](dsconfig.json)                             | Clearer, convention-aligned section layout (Custom HTTP Headers, Timeout, Allowed cookies now live under Advanced settings) |
| 4   | Skip `indexedPair` fields in `JSONDataMatchesStruct` / `JSONDataTypesMatchStruct`                                       | [`../../schema/conformance.go`](../../schema/conformance.go) | An `indexedPair` field has no single backing Go struct field (see below). **User-approved** before editing. |
| 5   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./...`                                              | generated artifacts                                          | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`)                                        |

No changes were made to `settings.go`, `settings.ts`, or `README.md` — per the constraint that all changes flow through `dsconfig.json` with the rest produced by `go generate` (plus the approved conformance-test fix).

---

## Section layout

The `groups` are organised into a standard, convention-aligned taxonomy (verified rendering
top-to-bottom in the new UI sidebar and accordion):

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Connection** (`connection`) | no | URL |
| 2 | **Authentication** (`authentication`) | no | Basic auth, User, Password, With Credentials, Forward OAuth Identity |
| 3 | **Network & TLS** (`network-tls`) | yes | TLS Client Auth → ServerName → Client Cert → Client Key, With CA Cert → CA Cert, Skip TLS Verify |
| 4 | **Graphite settings** (`graphite-settings`, underlying database) | no | Version, Graphite backend type, Rollup indicator |
| 5 | **Query settings** (`query-settings`) | yes | Label mappings (`importConfiguration`) |
| 6 | **Advanced settings** (`advanced`) | yes | Timeout, Allowed cookies, Custom HTTP Headers |

Notes:

- **With Credentials** is placed under Authentication (it governs whether credentials are
  sent on cross-site requests; this matches Grafana's newer `Auth` component, which folds it
  in as `CrossSiteCredentials`).
- **Timeout / Allowed cookies / Custom HTTP Headers** live under Advanced settings, matching
  Grafana's "Advanced HTTP settings" grouping; TLS-specific fields stay in Network & TLS.
- **Network & TLS field order** pairs each toggle with the fields it reveals: `TLS Client Auth`
  → ServerName / Client Cert / Client Key (mTLS), then `With CA Cert` → CA Cert (server-cert
  verification), then the standalone `Skip TLS Verify`. So flipping a toggle reveals its
  dependent inputs directly beneath it.
- The authentication group uses **`id: "authentication"`** (the registry convention — 65 of 67
  entries use it). The plugin-ui wizard's required-fields resolver (`resolveRequiredFieldsGroup`)
  and Authorization-header helper originally keyed off only the short id `auth`, so auth fields
  did not fold into the wizard's **General** step for `authentication`-id schemas. That was fixed
  in `plugin-ui` (see [plugin-ui change](#plugin-ui-change) below) to recognise **both**
  `authentication` and `auth`, so graphite can keep the conventional id and still get auth in
  General.
- Section titles carry short `description` subtitles; `optional` groups are collapsible.

### Wizard mode: URL in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) every
field in the group whose id is `auth`, plus their `dependsOn` parents/children.

Two adjustments make this work for graphite:

1. `root_url` previously used `requiredWhen: "true"` (a CEL expression the resolver does **not**
   inspect), so no General step was created and the wizard opened on `Connection`. Changing it to
   `"required": true` (unconditionally required — the backend admission handler rejects an empty
   URL) puts URL into General and emits a proper OpenAPI `required: ["url"]` in the generated
   settings spec (instead of the `x-dsconfig-required-when` extension).
2. The authentication group keeps the conventional `id: "authentication"`, and `plugin-ui` was
   updated to treat both `authentication` and `auth` as the auth group, so the resolver folds the
   auth fields into General as well.

Result: the wizard opens on **General 1/7** containing URL, Basic auth, With Credentials, and
Forward OAuth Identity; toggling **Basic auth** reveals the dependent **User** / **Password**
inputs inline (verified). The now-redundant `Connection` and `Authentication` steps are
auto-skipped in the wizard flow because all their fields already appear in General. Tab mode is
unaffected — the synthetic `_required` group is filtered out there, so it still shows the six
sections in order.

### plugin-ui change

The auth-group recognition was generalised in the `grafana/plugin-ui` repo (branch `dsconfig`,
`src/components/ConfigEditor/DatasourceConfigWizard/`) so the wizard matches the full
`authentication` id used across the registry, not just the short `auth`:

- `config.ts` — added `AUTH_GROUP_IDS = ['authentication', 'auth']` and an `isAuthGroupId()`
  helper; `resolveRequiredFieldsGroup` now finds the auth group via `isAuthGroupId(g.id)`.
- `TabLayout.tsx` and `DatasourceConfigWizard.tsx` — the Authorization-header helper's group
  check now uses `isAuthGroupId(...)` instead of the literal `=== 'auth'`.

This is a separate repo/PR from the dsconfig schema change; `tsc`, `eslint`, and the existing
`DatasourceConfigWizard` Jest suite pass.

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
| Version                 | select (`0.9`/`1.0`/`1.1`)               | `jsonData_graphiteVersion`         | select                                  | `jsonData.graphiteVersion`                                         | ✅     |
| Graphite backend type   | select (Default/Metrictank)              | `jsonData_graphiteType`            | select                                  | `jsonData.graphiteType`                                            | ✅     |
| Rollup indicator        | switch                                   | `jsonData_rollupIndicatorEnabled`  | switch                                  | `jsonData.rollupIndicatorEnabled`                                  | ✅ 🔀  |
| Label mappings          | help drawer + mapping rows               | `jsonData_importConfiguration`     | complex-object note + help drawer²      | `jsonData.importConfiguration`                                     | ✅     |

¹ **Not a discrepancy.** The `dsconfig.json` declares `ui.component: "textarea"` for the three TLS cert fields, but the new renderer checks `target === "secureJsonData"` _before_ the `textarea` branch (`renderFieldInput.tsx:25`), so any secure field is drawn as a masked secure input with a show/hide toggle. Both UIs collect the same PEM text into the same `secureJsonData` keys; only the widget affordance differs (textarea vs. secure input). This is a renderer policy in `plugin-ui`, not a schema gap.

² The label-mappings storage is a 3-level nested object serialized from free-form strings by the plugin's own `parseLokiLabelMappings.ts`. The wizard shows it as a non-editable complex-field note plus the rich `help.markdown` drawer, consistent with how other entries model opaque nested objects. Not editable inline in either the array sense, but present and documented.

---

## Gap found and fixed: Custom HTTP Headers

**Legacy behaviour (verified in screenshot `legacy-tls`):** the Auth area includes a
**Custom HTTP Headers** section with an **Add header** button. Adding a header shows a
header-name text input (placeholder `X-Custom-Header`) and a header-value **password**
input (`aria-label="Value"`). `@grafana/ui`'s `CustomHeadersSettings` persists these as
indexed pairs — `jsonData.httpHeaderName<N>` for the name and
`secureJsonData.httpHeaderValue<N>` for the (secret) value, starting at `N = 1`.

**Before:** `dsconfig.json` did not model headers at all (the old README/instructions
explicitly excluded them), so the new UI rendered no headers section.

**Fix (in `dsconfig.json` only):** added the `jsonData_httpHeaders` array field with an
`indexedPair` storage mapping that reproduces the exact legacy storage, plus item
sub-fields for the header name (`http.header.name`, with a header-name pattern
validation) and value (`http.header.value`). It is placed in the **Advanced settings**
section (see [Section layout](#section-layout) below):

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

**After (verified in screenshot `new-fixed`):** the new UI renders a **Custom HTTP
Headers** section (sidebar entry + accordion) with an **Add custom http header** button
and a key/secret-value row editor.

---

## `fileUpload` evaluation — not applicable to graphite

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not,
for graphite:

- In Grafana 13.0.1, the legacy `TLSAuthSettings` / `CertificationKey` renders the CA
  Cert / Client Cert / Client Key fields as **plain textareas** (placeholders
  `Begins with -----BEGIN CERTIFICATE-----` / `-----BEGIN RSA PRIVATE KEY-----`). No
  file-upload button and no `<input type="file">` were found in the legacy DOM
  (screenshot `legacy-tls`: 0 upload buttons, 0 file inputs).
- The new UI's `fileUpload` component (`FileUploadField.tsx`) only activates when the
  field declares `ui.fileMapping` (multi-key JSON distribution, e.g. a GCP service-account
  file) and is hard-coded for that JSON-token use case — it does not model single-PEM
  upload.

**Decision:** do **not** add `fileUpload` to any graphite field. The cert fields keep
their current modeling; both UIs collect PEM text into the same `secureJsonData` keys.

---

## Conditional fields & effects — tested

All `dependsOn` conditionals were exercised in the new UI and confirmed to reveal/route
correctly (evidence from the `Save` console payload):

| Trigger                                 | Revealed field(s)                   | Verified                                                                                                                   |
| --------------------------------------- | ----------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `root_basicAuth == true`                | User, Password                      | ✅ appear on toggle; `basicAuthUser` → root, `basicAuthPassword` → secureJsonData                                          |
| `jsonData_tlsAuth == true`              | ServerName, Client Cert, Client Key | ✅ TLS/SSL Auth Details activates                                                                                          |
| `jsonData_tlsAuthWithCACert == true`    | CA Cert                             | ✅                                                                                                                         |
| `jsonData_graphiteType == 'metrictank'` | Rollup indicator                    | ✅ hidden by default; appears after selecting Metrictank; saves `graphiteType:"metrictank"`, `rollupIndicatorEnabled:true` |

**Effects:** graphite's schema contains **no** `effects` blocks. Its auth model is a set of
independent direct toggles (`basicAuth`, `withCredentials`, `oauthPassThru`), not a virtual
selector that fans out to multiple fields — so there is nothing for `effects` to model, and
none were added. (This mirrors the existing modeling notes in the entry's README.)

### Save-payload storage-target validation

Filling the form and clicking **Save & Test** logs the exact datasource payload the wizard
would PUT. Two representative captures:

Basic + TLS toggles on:

```json
{
  "url": "http://graphite.example.com:8080",
  "basicAuth": true,
  "basicAuthUser": "grafana",
  "withCredentials": false,
  "jsonData": {
    "tlsAuth": true,
    "tlsAuthWithCACert": true,
    "tlsSkipVerify": false,
    "oauthPassThru": false,
    "graphiteVersion": "1.1"
  },
  "secureJsonData": { "basicAuthPassword": "s3cret" }
}
```

One custom header added (name `X-Api-Token`, value `super-secret-token`):

```json
{
  "jsonData":      { "httpHeaderName1": "X-Api-Token", ... },
  "secureJsonData":{ "httpHeaderValue1": "super-secret-token" },
  "secureJsonFields": { "httpHeaderValue1": false }
}
```

The header **name lands in `jsonData.httpHeaderName1`** and the **value in
`secureJsonData.httpHeaderValue1`** — byte-for-byte the legacy `CustomHeadersSettings`
storage format. All other fields route to `root` / `jsonData` / `secureJsonData` exactly
as declared.

---

## Conformance-test change (approved before editing)

Adding an `indexedPair` field surfaced a real limitation in the shared conformance suite.
`JSONDataMatchesStruct` collects every `target: jsonData` field key and requires a matching
`json:"..."` tag on the Go settings struct. But an `indexedPair` field
(`jsonData_httpHeaders`) is a **logical view** over dynamically-indexed legacy keys
(`httpHeaderName1`, `httpHeaderValue1`, …). Those dynamic keys are intentionally **not**
modeled as a Go struct field — the SDK's `HTTPClientOptions` reads them by prefix — the
same way this entry's `SecureJsonDataKeys` already omits the per-header
`httpHeaderValue<N>` secrets. Since `settings.go` is hand-authored (not `go generate`d) and
out of scope for this task, the correct, generic fix is in the conformance walker.

**Fix (`schema/conformance.go`):** a new `isIndexedPairField()` helper; both
`JSONDataMatchesStruct` and `JSONDataTypesMatchStruct` now `continue` past `indexedPair`
fields. This is plugin-agnostic and unblocks any future entry that models indexed
name/value pairs.

Impact check: the full `./registry/...` + `./schema/...` suites pass unchanged, so no other
entry regressed. (`SchemaArtifactInSync` briefly failed before regeneration and passed after
`go generate`; the SDK converter already emits `httpHeaders` as a clean array under
`jsonData` and keeps `secureValues` limited to the four static secrets.)

---

## Verification

```
go generate ./registry/graphite/...      # regenerate schema.gen.json / settings.gen.json
go test ./registry/graphite/...          # 8/8 conformance subtests PASS
go test ./registry/... ./schema/...      # entire suite PASS (no regressions)
gofmt -l schema/ registry/graphite/      # clean
go vet ./schema/... ./registry/graphite/ # clean
```

Conformance subtests (graphite): `BaseFieldsResolved`, `SchemaRoundTrip`,
`SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`,
`JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings` — all
**PASS**.

---

## Files changed

- [`registry/graphite/dsconfig.json`](dsconfig.json) — added `jsonData_httpHeaders` field; restructured `groups` into the Connection / Authentication / Network & TLS / Graphite settings / Query settings / Advanced settings taxonomy; ordered the Network & TLS fields so each toggle precedes the inputs it reveals; changed `root_url` from `requiredWhen: "true"` to `required: true` (renders in the wizard's General step); updated the secure-values instruction.
- [`registry/graphite/schema.gen.json`](schema.gen.json), [`registry/graphite/settings.gen.json`](settings.gen.json) — regenerated by `go generate` (`url` now in the spec's `required` array).
- [`schema/conformance.go`](../../schema/conformance.go) — `isIndexedPairField()` helper; skip `indexedPair` fields in the jsonData↔struct parity checks (user-approved).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`.
