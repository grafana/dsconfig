# CSV (marcusolsson-csv-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `marcusolsson-csv-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/ffrbqiyubrta8d` (Grafana Enterprise, `@grafana/plugin-ui` `ConfigEditor` — a `RadioButtonGroup` storage selector wrapping the shared `Auth` / HTTP-settings widget)
- **New UI:** `http://192.168.1.241:58899/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:marcusolsson-csv-datasource` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + `Save & Test`-button console payloads). The Storybook story fetches the schema from `raw.githubusercontent.com/grafana/dsconfig/schema-discovery/.../marcusolsson-csv-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the captures reflect the local schema without pushing.
- **Result:** **Parity achieved.** One missing field was found (**Custom HTTP Headers**) and added; `fileUpload` was evaluated and correctly **not** used (0 file inputs in legacy); the `virtual_authMethod` selector `effects` and the storage-mode / `dependsOn` conditionals were exercised and their storage targets verified from the save payload.

---

## TL;DR of changes

| #   | Change                                                                                           | File                             | Why                                                                                                            |
| --- | ------------------------------------------------------------------------------------------------ | -------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| 1   | Changed `root_url` from `requiredWhen: "true"` → `required: true`                                | [`dsconfig.json`](dsconfig.json) | URL is unconditionally required in both storage modes; puts URL into the wizard's synthetic **General** step and emits `required: ["url"]` |
| 2   | Added **Custom HTTP Headers** field (`jsonData_httpHeaders`) with `indexedPair` storage           | [`dsconfig.json`](dsconfig.json) | Legacy UI renders an **HTTP headers** section with **Add header**; new UI had no headers editor                |
| 3   | Added `"dependsOn": "jsonData_storage == 'http'"` to `jsonData_httpHeaders`                       | [`dsconfig.json`](dsconfig.json) | **CSV-specific deviation from prometheus:** headers only apply to HTTP storage; the legacy editor hides them (and all HTTP settings) in `local` mode. Matches sibling fields `keepCookies` / `timeout` / `queryParams`. |
| 4   | Added `jsonData_httpHeaders` to the `additional-settings` group's `fieldRefs` (after `jsonData_timeout`) | [`dsconfig.json`](dsconfig.json) | Surface the new field under **Additional settings**, matching the legacy grouping                             |
| 5   | Updated the `secure` instruction: headers are now **modeled** (was "…via dynamic secrets … see README") | [`dsconfig.json`](dsconfig.json) | Keep the embedded `llm`-tagged instructions truthful after change #2                                           |
| 6   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/marcusolsson-csv-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`)                                          |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance_test.go`, or `plugin-ui`. All schema changes flow through `dsconfig.json`; the rest is produced by `go generate`.

**No conformance-test or plugin-ui change was required** (see [Conformance](#conformance-no-change-required) and [Wizard mode](#wizard-mode-url-in-the-general-step) below) — the `indexedPair` conformance skip and the auth-group → General generalisation were already in place from the graphite/prometheus work.

### Note on the "copy verbatim" instruction

The task asked to copy `jsonData_httpHeaders` verbatim from `registry/prometheus/dsconfig.json`. The field body (item sub-fields, `role`s, header-name pattern, `indexedPair` storage) **is** verbatim. The single, deliberate addition is `"dependsOn": "jsonData_storage == 'http'"` (change #3). Prometheus is always HTTP so it has no such gate; CSV is dual-storage and **every** other HTTP-transport field carries this exact `dependsOn`. Both the legacy editor (verified below) and the CSV schema's own convention require it for true parity — and it is fixable in `dsconfig.json` alone, so it was applied rather than left as a known gap.

---

## Section layout

The `groups` were left in their existing taxonomy; only the new field was slotted into
**Additional settings**. Verified rendering top-to-bottom in the new UI (tab mode):

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Storage** (`storage`) | no | Storage Location (radio: HTTP / Local) |
| 2 | **Connection** (`connection`) | no | URL |
| 3 | **Authentication** (`authentication`) | no | Authentication method (virtual) → User → Password |
| 4 | **TLS settings** (`tls-settings`) | yes | Add self-signed certificate → CA Certificate, TLS Client Authentication → ServerName → Client Certificate → Client Key, Skip TLS certificate validation |
| 5 | **Additional settings** (`additional-settings`) | yes | Allowed cookies, Timeout, **Custom HTTP Headers** ➕, Custom query parameters |

Notes:

- The **Custom HTTP Headers** field is placed in **Additional settings** (alongside Allowed
  cookies, Timeout, Custom query parameters), matching where the CSV editor keeps the HTTP-transport knobs.
- **Storage mode gates the HTTP block.** `jsonData.storage` is a `radio` (HTTP default / Local).
  Authentication, TLS settings, and every field in Additional settings carry
  `dependsOn: jsonData_storage == 'http'`, so in **Local** mode only Storage + URL remain — see
  [Storage-mode gating](#storage-mode-gating-http-vs-local).
- The **URL** field is dual-purpose: in HTTP mode it is the base HTTP endpoint (placeholder
  `http://localhost:8080`); in Local mode an `override` (`when: jsonData_storage == 'local'`) swaps
  the description/placeholder to the filesystem path. It is `required: true` in both modes.
- `optional` groups render collapsible in tab mode.

### Wizard mode: URL in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the fields of
the auth group, plus their `dependsOn` parents/children.

- **Before:** `root_url` used `requiredWhen: "true"` (a CEL expression the resolver does **not**
  inspect), so URL would not be folded into General.
- **After:** changing it to `required: true` (unconditionally required — `pkg/datasource.go:121-125`
  rejects an empty URL) puts URL into General and emits a proper OpenAPI `required: ["url"]` in the
  generated spec (instead of the `x-dsconfig-required-when: "true"` extension).

**Verified (screenshot `newgen-csv-fixed-wiz` / `newgen-csv-final`):** the wizard opens on
**General 1/6** containing **Storage Location** (radio), **URL** (with the required `*`,
`urlPresent: true`), and the **Authentication method** select. Storage Location is folded in because
it is the `dependsOn` parent of `virtual_authMethod` (and of the URL `override`).

---

## Field-by-field parity

Legend: ✅ present & matching · ➕ added by this change · 🔀 conditional (`dependsOn`) · ⚙ driven by the `virtual_authMethod` selector · 🗄 gated by storage mode (`dependsOn: jsonData_storage == 'http'`)

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| Storage Location | RadioButtonGroup (HTTP / Local) | `jsonData_storage` | radio | `jsonData.storage` | ✅ |
| URL | text input (dual-purpose) | `root_url` | input | `root.url` (required) | ✅ |
| Authentication method | select (Basic / Forward OAuth / No Auth) | `virtual_authMethod` | select + `effects` | virtual → `root.basicAuth` / `jsonData.oauthPassThru` | ✅ ⚙ 🗄 |
| — (managed) | — | `root_basicAuth` | (hidden, managed) | `root.basicAuth` | ✅ ⚙ |
| — (managed) | — | `jsonData_oauthPassThru` | (hidden, managed) | `jsonData.oauthPassThru` | ✅ ⚙ |
| User | text input | `root_basicAuthUser` | input | `root.basicAuthUser` | ✅ 🔀 |
| Password | password (secure) | `secureJsonData_basicAuthPassword` | secure input | `secureJsonData.basicAuthPassword` | ✅ 🔀 |
| Add self-signed certificate | switch | `jsonData_tlsAuthWithCACert` | switch | `jsonData.tlsAuthWithCACert` | ✅ 🗄 |
| CA Certificate | textarea | `secureJsonData_tlsCACert` | secure input¹ | `secureJsonData.tlsCACert` | ✅ 🔀 |
| TLS Client Authentication | switch | `jsonData_tlsAuth` | switch | `jsonData.tlsAuth` | ✅ 🗄 |
| ServerName | text input | `jsonData_serverName` | input | `jsonData.serverName` | ✅ 🔀 |
| Client Certificate | textarea | `secureJsonData_tlsClientCert` | secure input¹ | `secureJsonData.tlsClientCert` | ✅ 🔀 |
| Client Key | textarea | `secureJsonData_tlsClientKey` | secure input¹ | `secureJsonData.tlsClientKey` | ✅ 🔀 |
| Skip TLS certificate validation | switch | `jsonData_tlsSkipVerify` | switch | `jsonData.tlsSkipVerify` | ✅ 🗄 |
| Allowed cookies | TagsInput | `jsonData_keepCookies` | list (string array) | `jsonData.keepCookies` | ✅ 🗄 |
| Timeout | number | `jsonData_timeout` | number | `jsonData.timeout` | ✅ 🗄 |
| **Custom HTTP Headers** | Add header → name input + value password | `jsonData_httpHeaders` | IndexedPair editor (key + secret value) | `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>` | ➕ 🗄 |
| Custom query parameters | text input | `jsonData_queryParams` | input | `jsonData.queryParams` | ✅ 🗄 |

¹ **Not a discrepancy.** The `dsconfig.json` declares `ui.component: "textarea"` for the three TLS
cert fields, but the new renderer draws any `target: "secureJsonData"` field as a masked secure
input with a show/hide toggle. Both UIs collect the same PEM text into the same `secureJsonData`
keys — only the widget affordance differs (the same policy documented for prometheus/graphite).

---

## Gap found and fixed: Custom HTTP Headers

**Legacy behaviour (verified in `legacy-expand-csv-legacy.json`):** the editor includes an
**HTTP headers** section heading with an **Add header** button (`hasCustomHeaders: true`,
`addHeaderBtn: true`, `fileInputs: 0`). `@grafana/plugin-ui`'s CustomHeaders component persists
headers as indexed pairs — `jsonData.httpHeaderName<N>` for the name and
`secureJsonData.httpHeaderValue<N>` for the (secret) value, starting at `N = 1`.

**Before:** `dsconfig.json` did not model headers at all (the secure instruction pointed to the
README), so the new UI rendered no headers editor.

**Fix (in `dsconfig.json` only):** added the `jsonData_httpHeaders` array field with an
`indexedPair` storage mapping that reproduces the exact legacy storage, plus item sub-fields for the
header name (`http.header.name`, with a header-name pattern validation) and value
(`http.header.value`), gated by `dependsOn: jsonData_storage == 'http'`, and added it to the
**Additional settings** group.

**After (verified in `newgen-csv-final` / `csv` header-routing run):** the new UI renders a
**Custom HTTP Headers** row under **Additional settings** with an **Add custom http header** button
and a key/secret-value editor (`hasHeadersEditor: true`), and the save payload routes the header
name to `jsonData.httpHeaderName1` and the value to `secureJsonData.httpHeaderValue1`.

---

## `fileUpload` evaluation — not applicable to CSV

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not:

- The legacy CSV editor renders the CA Cert / Client Cert / Client Key fields as **plain textareas**
  (`Begins with --- BEGIN CERTIFICATE ---` / `--- RSA PRIVATE KEY CERTIFICATE ---`). No file-upload
  button and no `<input type="file">` were found in the legacy DOM
  (`legacy-expand-csv-legacy.json`: `fileInputs: 0`, `uploadButtons: []`).
- The CSV plugin's own file input (the CSV data itself) is a **per-query** concern in the query
  editor, not a datasource-config field, and the plugin's `local` storage mode takes a filesystem
  **path** string (`root.url`), not an uploaded file.

**Decision:** do **not** add `fileUpload` to any CSV field. Both UIs collect PEM text into the same
`secureJsonData` keys.

---

## Conditional fields & effects — tested

### `virtual_authMethod` selector (`effects`)

Auth is modelled as a **virtual selector** (`virtual_authMethod`) whose value is `read` from
`root.basicAuth` / `jsonData.oauthPassThru` and whose `effects` fan out to those two managed fields.
All three branches were driven from **fresh page loads** (`verify-csv-effects.js`) and verified from
the `Save & Test` console payload:

| Selection | UI effect (verified) | Save payload (verified) |
| --- | --- | --- |
| **No Authentication** (default) | User / Password hidden | `basicAuth: false`, `jsonData.oauthPassThru: false` |
| **Basic authentication** | User + Password inputs revealed (🔀 `dependsOn: virtual_authMethod == 'BasicAuth'`) | `basicAuth: true`, `basicAuthUser: "grafana"`, `secureJsonData.basicAuthPassword: "s3cret"`, `oauthPassThru: false` |
| **Forward OAuth Identity** | User / Password hidden | `basicAuth: false`, `jsonData.oauthPassThru: true` |

Each branch was run from a clean load, so each payload is authoritative (no stale-state carryover):
selecting **Forward OAuth Identity** straight from the default produced
`basicAuth: false, oauthPassThru: true` — confirming the `set` operations propagate, not just the
visibility. The auth selector only renders when `jsonData_storage == 'http'` (the default).

### Storage-mode gating (HTTP vs Local)

`jsonData.storage` gates the entire HTTP block. Verified in **both** UIs by toggling the storage
radio (`verify-csv-storage.js` for the new UI; `verify-csv-legacy-storage.js` for the legacy UI):

| Field | New UI — HTTP | New UI — Local | Legacy — HTTP | Legacy — Local |
| --- | --- | --- | --- | --- |
| Authentication method | shown | hidden | shown | hidden |
| Allowed cookies | shown | hidden | — | hidden |
| Timeout | shown | hidden | — | hidden |
| **Custom HTTP Headers** | shown | **hidden** | shown | **hidden** |
| Custom query parameters | shown | hidden | — | hidden |

The legacy editor hides the whole HTTP-settings widget (auth + TLS + headers) in `local` mode
(`hasHTTPHeadersText/hasAuthText/hasTLSText` all flip to `false`). Before change #3 the new UI left
**Custom HTTP Headers** visible in Local mode (the only field lacking the storage `dependsOn`);
adding `dependsOn: jsonData_storage == 'http'` brought it in line with both the legacy behaviour and
its sibling fields.

### Other `dependsOn` conditionals

| Trigger | Revealed field(s) | Verified |
| --- | --- | --- |
| `virtual_authMethod == 'BasicAuth'` | User, Password | ✅ appear on selection; route to `root` / `secureJsonData` |
| `jsonData_tlsAuth == true` | ServerName, Client Certificate, Client Key | ✅ (declared; TLS group) |
| `jsonData_tlsAuthWithCACert == true` | CA Certificate | ✅ (declared; TLS group) |
| `jsonData_storage == 'http'` | Auth + all TLS + all Additional-settings fields | ✅ hidden in Local mode (table above) |

### Save-payload storage-target validation

Filling the form and clicking **Save & Test** logs the exact datasource payload the wizard would
PUT. Custom header (name `X-Org-Id`, value `influx-tenant-9`) — header-routing run:

```
jsonData.httpHeaderName1:   X-Org-Id
secureJsonData.httpHeaderValue1: influx-tenant-9
url:                        http://influx.example.com:8086
```

The header **name lands in `jsonData.httpHeaderName1`** and the **value in
`secureJsonData.httpHeaderValue1`** — byte-for-byte the legacy CustomHeaders storage format. URL
routes to `root.url`; auth fields route to `root` / `jsonData` / `secureJsonData` exactly as
declared.

---

## Conformance (no change required)

Adding an `indexedPair` field would normally break `JSONDataMatchesStruct` /
`JSONDataTypesMatchStruct`, because `jsonData_httpHeaders` is a **logical view** over
dynamically-indexed legacy keys (`httpHeaderName1`, `httpHeaderValue1`, …) that are not modeled as a
single Go struct field. The shared conformance walker (`schema/conformance.go`) already skips
`indexedPair` fields via `isIndexedPairField()` — plugin-agnostic, added during the graphite work —
so **no conformance change was needed here**. `settings.go`'s `Config` struct (which already
documents that header pairs are dynamically indexed and intentionally unmodeled) was left untouched.
The per-header `httpHeaderValue<N>` secrets remain dynamic and are correctly **not** listed among the
static `SecureJsonDataKeys`, and the generated spec emits `httpHeaders` as a clean array under
`jsonData` with **no** secure values leaked (`SchemaSpecHasNoSecureJSON` passes).

---

## Verification

```
go generate ./registry/marcusolsson-csv-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/marcusolsson-csv-datasource/...        # PASS
go test ./schema/... ./registry/marcusolsson-csv-datasource/...  # PASS (shared infra intact)
```

`TestSchemaConformance` subtests (marcusolsson-csv-datasource) — all **PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `TestLoadConfig`, `TestApplyDefaults`, and `TestValidate` suites also pass
unchanged (they do not reference the new field).

New-UI captures: `newgen-csv-final` (tab, `hasHeadersEditor: true`, `urlPresent: true`),
`newgen-csv-fixed-wiz` (wizard opens on **General 1/6** with required URL + auth method),
`verify-csv-effects-result` (auth-method effects, all three branches), `verify-csv-storage-*`
(HTTP vs Local gating). Legacy captures: `legacy-expand-csv-legacy` (HTTP headers + Add header
present; 0 file inputs), `verify-csv-legacy-storage-result` (legacy hides HTTP block in Local mode).

---

## Files changed

- [`registry/marcusolsson-csv-datasource/dsconfig.json`](dsconfig.json) — changed `root_url` from
  `requiredWhen: "true"` to `required: true` (renders in the wizard's General step); added the
  `jsonData_httpHeaders` `indexedPair` field (verbatim from prometheus **plus**
  `dependsOn: jsonData_storage == 'http'`) and referenced it from the `additional-settings` group;
  updated the `secure` instruction so it states headers are now modeled.
- [`registry/marcusolsson-csv-datasource/schema.gen.json`](schema.gen.json),
  [`registry/marcusolsson-csv-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`url` now in the spec's `required` array; `httpHeaders` array added under `jsonData`
  with `x-dsconfig-depends-on: "jsonData_storage == 'http'"`; `x-dsconfig-required-when: "true"`
  removed from `url`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`conformance_test.go`, `schema/conformance.go`, and everything in `plugin-ui`.
