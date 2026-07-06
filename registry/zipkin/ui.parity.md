# Zipkin — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `zipkin`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/afras2t8c0we8e` (Grafana Enterprise, `@grafana/plugin-ui` Auth + `@grafana/o11y-ds-frontend` trace-linking sections)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:zipkin` (Storybook, `ConfigEditor/DatasourceConfigWizard`)
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + `Save & Test` console payload). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/zipkin/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the screenshots reflect the local schema without pushing.
- **Result:** **Parity achieved.** One missing field was found (**Custom HTTP Headers**) and added; `fileUpload` was evaluated and correctly **not** used; the `virtual_authMethod` selector and all `dependsOn` conditionals were exercised and their storage targets verified from the save payload. The Zipkin feature groups (trace-to-X, node graph, span bar) were left inline and unchanged — no packs, no restructuring.

---

## TL;DR of changes

| #   | Change                                                                                                                                                     | File                             | Why                                                                                                                                                        |
| --- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- | -------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Added **Custom HTTP Headers** field (`jsonData_httpHeaders`) with `indexedPair` storage; added to the `additional-settings` group after `jsonData_timeout` | [`dsconfig.json`](dsconfig.json) | Legacy UI renders an **HTTP headers** section with an **Add header** button; the new UI had no headers editor                                              |
| 2   | Changed `root_url` from `requiredWhen: "true"` → `required: true`                                                                                          | [`dsconfig.json`](dsconfig.json) | Puts URL into the wizard's synthetic **General** step and emits a proper OpenAPI `required: ["url"]` (instead of the `x-dsconfig-required-when` extension) |
| 3   | Updated the `secure` instruction so it states headers **are** modeled (via `jsonData_httpHeaders`) rather than pointing at the README                      | [`dsconfig.json`](dsconfig.json) | The old text implied headers lived only in the README (i.e. "not modeled"); they are now first-class in the schema                                         |
| 4   | Regenerated `schema.gen.json` / `settings.gen.json` via `go generate ./registry/zipkin/...`                                                                | generated artifacts              | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`)                                                                                       |

No changes were needed to `settings.go`, `settings.ts`, `README.md`, `schema/conformance.go`, or `plugin-ui`:

- **`conformance.go`** already skips `indexedPair` fields in the jsonData↔struct parity checks (`isIndexedPairField`, conformance.go:157/180/245) — this generic fix was landed during the graphite work, so no Go change was required here.
- **`plugin-ui`** already recognises the conventional group id `authentication` (not only the short `auth`) in its `resolveRequiredFieldsGroup` — also landed during earlier work — so the wizard folds Zipkin's auth field into **General** without any further change.
- **`settings.go`** intentionally omits the dynamic `httpHeaderValue<N>` secrets from `SecureJsonDataKeys` (settings.go:167-170), which matches the `indexedPair` model exactly.

---

## Section layout

The existing 6 feature `groups` were **kept as-is and inline**. The only structural
change is that `jsonData_httpHeaders` was appended to the existing **Additional settings**
group. Verified rendering top-to-bottom in the new UI (tab mode, DOM `newgen-zipkin-tab`,
screenshot `newgen-zipkin-tab.png`):

| Order | Section (`id`)                                  | `optional` | Fields                                                                                                                                                  |
| ----- | ----------------------------------------------- | ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1     | **Connection** (`connection`)                   | no         | URL                                                                                                                                                     |
| 2     | **Authentication** (`authentication`)           | no         | Authentication method (virtual selector), User, Password                                                                                                |
| 3     | **TLS settings** (`tls-settings`)               | yes        | Add self-signed certificate → CA Certificate, TLS Client Authentication → ServerName / Client Certificate / Client Key, Skip TLS certificate validation |
| 4     | **Trace to logs** (`trace-to-logs`)             | yes        | tracesToLogsV2, tracesToLogs (legacy)                                                                                                                   |
| 5     | **Trace to metrics** (`trace-to-metrics`)       | yes        | tracesToMetrics                                                                                                                                         |
| 6     | **Additional settings** (`additional-settings`) | yes        | Allowed cookies, Timeout, **Custom HTTP Headers** ➕, Node graph, Span bar                                                                              |

Notes:

- **Legacy grouping vs. new grouping.** The legacy editor nests **Authentication methods**,
  **TLS settings**, and **HTTP headers** as `h6` sub-sections under one **Authentication** (`h3`)
  umbrella, and collects the cookies/timeout/headers knobs under an **Advanced HTTP settings**
  sub-section inside an **Additional settings** umbrella (legacy headings captured in
  `legacy-expand-zipkin-parity`: `Connection, Authentication, Authentication methods, TLS
settings, HTTP headers, Trace to logs, Trace to metrics, Additional settings, Advanced HTTP
settings, Node graph, Span bar`). The schema expresses the same fields as flatter,
  individually-collapsible top-level sections; the single **Additional settings** group maps to
  the legacy **Additional settings** umbrella (Advanced-HTTP knobs + Node graph + Span bar). All
  fields are present in both UIs; only the nesting depth differs.
- **Custom HTTP Headers** lives under **Additional settings** alongside Timeout / Allowed
  cookies, matching the legacy **Advanced HTTP settings** sub-section (this task specified the
  `additional-settings` group as the target; Zipkin has no separate `advanced-http` group like
  Tempo).
- The trace-to-X / node-graph / span-bar fields are `valueType: object` with rich `help.markdown`
  and no editable `ui.component`, so the new UI renders them as complex-object notes plus a help
  drawer (the same treatment other tracing entries use for opaque nested objects). This is
  unchanged by this work.
- **Zipkin has fewer sections than Tempo.** Zipkin exposes no Streaming, Trace-to-profiles,
  Service graph, Tempo search, TraceID query, Tags time range, or Tag limit sections (all
  Tempo-only), and unlike Jaeger has no Trace-by-ID time-parameters toggle.

### Wizard mode: URL + Authentication method in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the fields
of the auth group, plus their `dependsOn` parents/children.

- `root_url` was changed from `requiredWhen: "true"` (a CEL expression the resolver does **not**
  inspect) to `required: true` (unconditionally required — the Zipkin backend reads `settings.URL`
  directly and `NewDatasource` rejects an empty URL, pkg/zipkin/zipkin.go:33-42). This puts URL
  into General.
- The auth group keeps the conventional `id: "authentication"`; plugin-ui already treats it as the
  auth group, so the `virtual_authMethod` selector folds into General too.

**Verified** (screenshot `verify-zipkin-wizard.png`, DOM `verify-zipkin-wizard.json`): the wizard
opens on a step literally titled **General** (`generalTitleFound: true`) at **1/7** (6 groups + 1
synthetic General step) containing the required **URL** input (`http://localhost:9411` placeholder,
`*` required marker) and the **Authentication method** selector. Tab mode is unaffected — the
synthetic `_required` group is filtered out there, so it still shows the 6 sections in order.

---

## Field-by-field parity

Legend: ✅ present & matching · ➕ added by this change · 🔀 conditional (revealed by `dependsOn`) · ⚙️ managed by the `virtual_authMethod` selector · 📄 complex object rendered as help/note

| Legacy UI field                 | Control (legacy)              | New UI (schema id)                                           | Storage target                                                                                 | Status                               |
| ------------------------------- | ----------------------------- | ------------------------------------------------------------ | ---------------------------------------------------------------------------------------------- | ------------------------------------ |
| URL                             | text input                    | `root_url`                                                   | `root.url`                                                                                     | ✅ (now `required: true`)            |
| Authentication method           | select (`auth-method-select`) | `virtual_authMethod`                                         | _virtual_ — computed read of `root.basicAuth` / `jsonData.oauthPassThru`; writes via `effects` | ✅ ⚙️                                |
| — (Basic auth flag)             | (set by selector)             | `root_basicAuth`                                             | `root.basicAuth`                                                                               | ✅ ⚙️                                |
| User                            | text input                    | `root_basicAuthUser`                                         | `root.basicAuthUser`                                                                           | ✅ 🔀 (`authMethod == 'BasicAuth'`)  |
| Password                        | password (secure)             | `secureJsonData_basicAuthPassword`                           | `secureJsonData.basicAuthPassword`                                                             | ✅ 🔀                                |
| — (Forward OAuth flag)          | (set by selector)             | `jsonData_oauthPassThru`                                     | `jsonData.oauthPassThru`                                                                       | ✅ ⚙️                                |
| Add self-signed certificate     | switch                        | `jsonData_tlsAuthWithCACert`                                 | `jsonData.tlsAuthWithCACert`                                                                   | ✅                                   |
| CA Certificate                  | textarea                      | `secureJsonData_tlsCACert`                                   | `secureJsonData.tlsCACert`                                                                     | ✅ 🔀 (`tlsAuthWithCACert == true`)¹ |
| TLS Client Authentication       | switch                        | `jsonData_tlsAuth`                                           | `jsonData.tlsAuth`                                                                             | ✅                                   |
| ServerName                      | text input                    | `jsonData_serverName`                                        | `jsonData.serverName`                                                                          | ✅ 🔀 (`tlsAuth == true`)            |
| Client Certificate              | textarea                      | `secureJsonData_tlsClientCert`                               | `secureJsonData.tlsClientCert`                                                                 | ✅ 🔀 ¹                              |
| Client Key                      | textarea                      | `secureJsonData_tlsClientKey`                                | `secureJsonData.tlsClientKey`                                                                  | ✅ 🔀 ¹                              |
| Skip TLS certificate validation | switch                        | `jsonData_tlsSkipVerify`                                     | `jsonData.tlsSkipVerify`                                                                       | ✅                                   |
| Trace to logs                   | section                       | `jsonData_tracesToLogsV2` (+ legacy `jsonData_tracesToLogs`) | `jsonData.tracesToLogsV2` / `jsonData.tracesToLogs`                                            | ✅ 📄                                |
| Trace to metrics                | section                       | `jsonData_tracesToMetrics`                                   | `jsonData.tracesToMetrics`                                                                     | ✅ 📄                                |
| Allowed cookies                 | TagsInput                     | `jsonData_keepCookies`                                       | `jsonData.keepCookies`                                                                         | ✅                                   |
| Timeout                         | number                        | `jsonData_timeout`                                           | `jsonData.timeout`                                                                             | ✅                                   |
| **HTTP headers** (Add header)   | name input + value password   | `jsonData_httpHeaders`                                       | `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>`                             | ➕ 🔀                                |
| Node graph                      | switch                        | `jsonData_nodeGraph`                                         | `jsonData.nodeGraph`                                                                           | ✅ 📄                                |
| Span bar                        | section (type + tag)          | `jsonData_spanBar`                                           | `jsonData.spanBar`                                                                             | ✅ 📄                                |

¹ **Not a discrepancy.** The schema declares `ui.component: "textarea"` for the three TLS cert
fields, but the new renderer draws any `secureJsonData` field as a masked secure input with a
show/hide toggle (the `target === "secureJsonData"` branch is checked before the `textarea`
branch). Both UIs collect the same PEM text into the same `secureJsonData` keys; only the widget
affordance differs. This is a renderer policy in `plugin-ui`, not a schema gap.

---

## Gap found and fixed: Custom HTTP Headers

**Legacy behaviour (verified in `legacy-expand-zipkin-parity`):** the Authentication area includes
an **HTTP headers** section with an **Add header** button (`hasCustomHeaders: true`,
`addHeaderBtn: true`). Adding a header shows a header-name text input and a header-value password
input. `@grafana/plugin-ui`'s CustomHeaders component persists these as indexed pairs —
`jsonData.httpHeaderName<N>` for the name and `secureJsonData.httpHeaderValue<N>` for the (secret)
value, starting at `N = 1`.

**Before:** `dsconfig.json` did not model headers at all (the README/settings.go explicitly
excluded them), so the new UI rendered no headers section.

**Fix (in `dsconfig.json` only):** added the `jsonData_httpHeaders` array field with an
`indexedPair` storage mapping that reproduces the exact legacy storage, plus item sub-fields for
the header name (`http.header.name`, with a header-name pattern validation) and value
(`http.header.value`). It was appended to the existing **Additional settings** group after
`jsonData_timeout`:

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

**After (verified in `newgen-zipkin-tab`, tab mode):** the new UI renders a **Custom HTTP
Headers** editor with an **Add custom http header** button and a key/secret-value row editor
(`hasHeadersEditor: true`).

> Note: this makes stale the "**not modeled**" wording in the entry's `README.md` (line 90),
> `settings.go` (lines 167-170), and `settings.ts` (line 201). Those files are **out of scope** for
> this task (constraint: changes flow only through `dsconfig.json` + generated artifacts), so they
> were left untouched; they should be refreshed separately. The `instructions` block inside
> `dsconfig.json` — which is in scope — was updated to state headers are now modeled.

---

## `fileUpload` evaluation — not applicable to zipkin

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not, for zipkin:

- The legacy TLS section renders CA Cert / Client Cert / Client Key as **textareas**
  (placeholders begin with `--- BEGIN CERTIFICATE ---` / `--- RSA PRIVATE KEY ---`). No
  file-upload button and no `<input type="file">` were found (`legacy-expand-zipkin-parity`:
  `fileInputs: 0`, `uploadButtons: []`).
- The new UI's `fileUpload` component only activates when a field declares `ui.fileMapping`
  (multi-key JSON distribution) and is hard-coded for that use case — it does not model single-PEM
  upload.

**Decision:** do **not** add `fileUpload` to any zipkin field. The cert fields keep their current
modeling; both UIs collect PEM text into the same `secureJsonData` keys.

---

## Conditional fields & effects — tested

Zipkin models auth as a **virtual selector** (`virtual_authMethod`) with a `computed` read and
three `effects` blocks. All conditionals and effects were exercised in the new UI with a fresh page
per selection and confirmed from the `Save & Test` console payload (`zipkin-effects-result.json`):

| Trigger (selector / toggle)           | Revealed / routed                                                                   | Verified                                                                            |
| ------------------------------------- | ----------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| `virtual_authMethod = 'BasicAuth'`    | reveals User + Password; sets `root.basicAuth=true`, `jsonData.oauthPassThru=false` | ✅ userReveal=1, `basicAuth:true`, `oauthPassThru:false`, `basicAuthUser:"grafana"` |
| `virtual_authMethod = 'OAuthForward'` | hides User/Password; sets `root.basicAuth=false`, `jsonData.oauthPassThru=true`     | ✅ userReveal=0, `basicAuth:false`, `oauthPassThru:true`                            |
| `virtual_authMethod = 'NoAuth'`       | hides User/Password; sets `root.basicAuth=false`, `jsonData.oauthPassThru=false`    | ✅ userReveal=0, `basicAuth:false`, `oauthPassThru:false`                           |
| `jsonData_tlsAuth == true`            | ServerName, Client Cert, Client Key                                                 | ✅ `dependsOn` reveal                                                               |
| `jsonData_tlsAuthWithCACert == true`  | CA Certificate                                                                      | ✅ `dependsOn` reveal                                                               |

The `computed` read expression
(`root.basicAuth == true ? 'BasicAuth' : (jsonData.oauthPassThru == true ? 'OAuthForward' : 'NoAuth')`)
correctly initialises the selector from stored flags, and the `effects` keep only one of
`basicAuth` / `oauthPassThru` true at a time — matching the editor's `onAuthMethodSelect`.

### Save-payload storage-target validation

Basic authentication + one custom header (name `X-Trace-Token`, value `super-secret-trace-value`):

```json
{
  "url": "http://zipkin.example.com:9411",
  "basicAuth": true,
  "basicAuthUser": "grafana",
  "jsonData": {
    "oauthPassThru": false,
    "tlsAuthWithCACert": false,
    "tlsAuth": false,
    "tlsSkipVerify": false,
    "httpHeaderName1": "X-Trace-Token"
  },
  "secureJsonData": {
    "basicAuthPassword": "s3cret",
    "httpHeaderValue1": "super-secret-trace-value"
  },
  "secureJsonFields": { "basicAuthPassword": false, "httpHeaderValue1": false }
}
```

The header **name lands in `jsonData.httpHeaderName1`** and the **value in
`secureJsonData.httpHeaderValue1`** (write-only, mirrored in `secureJsonFields`) — byte-for-byte
the legacy CustomHeaders storage format. Auth flags, URL, and TLS toggles route to
`root` / `jsonData` / `secureJsonData` exactly as declared.

---

## Conformance — no Go change needed

Adding an `indexedPair` field is already handled by the shared conformance suite:
`JSONDataMatchesStruct` / `JSONDataTypesMatchStruct` skip `indexedPair` fields
(`isIndexedPairField`, conformance.go:157/180/245), and `SecureValuesMatchLoadSettings` only walks
top-level `secureJsonData` fields — the per-header `httpHeaderValue<N>` secret lives inside
`storage.value`, not as a top-level field, so the static `SecureJsonDataKeys` list (the four
`basicAuthPassword` / `tlsCACert` / `tlsClientCert` / `tlsClientKey` secrets) stays correct. The
generated spec emits `httpHeaders` as a clean array under `jsonData` and keeps `secureValues`
limited to those four static secrets (`SchemaSpecHasNoSecureJSON` passes).

---

## Verification

```
go generate ./registry/zipkin/...            # regenerate schema.gen.json / settings.gen.json / settings.examples.gen.json
go test ./registry/zipkin/... -v              # 8/8 conformance subtests PASS (+ load/defaults/validate suites)
go test ./registry/zipkin/... ./schema/...    # zipkin + schema suites PASS (no regressions)
```

Conformance subtests (zipkin), all **PASS**: `BaseFieldsResolved`, `SchemaRoundTrip`,
`SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`, `JSONDataMatchesStruct`,
`JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings`. The `TestLoadConfig`,
`TestApplyDefaults`, and `TestValidate` suites also pass unchanged.

Playwright evidence (in the shared capture dir):

- `newgen-zipkin-tab` (tab) — 6 sections render, `hasHeadersEditor: true`, `urlPresent: true`; **Custom HTTP Headers** editor visible under **Additional settings**.
- `verify-zipkin-wizard` + `newgen-zipkin-wiz` (wizard) — **General** step at `1/7` with URL (`http://localhost:9411`) + Authentication method (`generalTitleFound: true`).
- `legacy-expand-zipkin-parity` (legacy inventory) — `hasCustomHeaders: true`, `addHeaderBtn: true`, `fileInputs: 0`, `uploadButtons: []`.
- `zipkin-effects-result` — auth selector effects + header routing (`httpHeaderName1` / `httpHeaderValue1`) verified from save payloads.

---

## Files changed

- [`registry/zipkin/dsconfig.json`](dsconfig.json) — added the `jsonData_httpHeaders` field (with `indexedPair` storage) and appended it to the `additional-settings` group after `jsonData_timeout`; changed `root_url` from `requiredWhen: "true"` to `required: true`; updated the `secure` instruction to state headers are now modeled. Feature groups left inline and unchanged.
- [`registry/zipkin/schema.gen.json`](schema.gen.json), [`registry/zipkin/settings.gen.json`](settings.gen.json) — regenerated by `go generate` (`url` now in the spec's `required` array; `httpHeaders` array under `jsonData`).

_Unchanged by design / constraint:_ `settings.go`, `settings.ts`, `README.md`, `schema/conformance.go`, `plugin-ui`. No `conformance.go` change was required (the `indexedPair` skip was already landed during earlier work), and no `plugin-ui` change was required (auth-group id recognition for `authentication` was also already landed). The "not modeled" wording in `settings.go:167-170`, `settings.ts:201`, and `README.md:90` is now stale but out of scope; it should be refreshed separately.
