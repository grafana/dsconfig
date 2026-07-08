# Splunk — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-splunk-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/bfrbqiyekwutca` (Grafana Enterprise, `@grafana/plugin-ui` `Auth` + Splunk `ConfigEditor`/`AdditionalSettingsEditor`)
- **New UI:** `http://192.168.1.241:58899/iframe.html?args=pluginType:grafana-splunk-datasource&id=configeditor-datasourceconfigwizard--tab` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + `Save & Test`-button console payloads). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/grafana-splunk-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the screenshots reflect the local schema without pushing.
- **Result:** **Parity achieved.** One missing field was found (**Custom HTTP Headers**) and added; `fileUpload` was evaluated and correctly **not** used (the legacy DOM has 0 file inputs); the `jsonData_authType` discriminator and every `dependsOn` conditional were exercised and their storage targets verified from the save payload.

---

## TL;DR of changes

| #   | Change                                                                                              | File                             | Why                                                                                                            |
| --- | ------------------------------------------------------------------------------------------------- | -------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| 1   | Changed `root_url` from `requiredWhen: "true"` → `required: true`                                  | [`dsconfig.json`](dsconfig.json) | URL is unconditionally required; puts URL into the wizard's synthetic **General** step and emits `required: ["url"]` |
| 2   | Added **Custom HTTP Headers** field (`jsonData_httpHeaders`) with `indexedPair` storage            | [`dsconfig.json`](dsconfig.json) | Legacy UI renders an **HTTP headers** section with **Add header**; new UI had no headers editor                |
| 3   | Added `jsonData_httpHeaders` to the `advanced-http` group's `fieldRefs` (after `jsonData_timeout`) | [`dsconfig.json`](dsconfig.json) | Surface the new field under **Advanced HTTP settings**, matching the legacy grouping                          |
| 4   | Updated the settings/legacy instruction: headers are now **modeled** (was silent on headers)       | [`dsconfig.json`](dsconfig.json) | Keep the embedded `llm`-tagged instructions truthful after change #2                                           |
| 5   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-splunk-datasource/...` | generated artifacts   | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`)                                          |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance.go`, or `plugin-ui`. All schema changes flow through `dsconfig.json`; the rest is produced by `go generate`.

**No conformance-test or plugin-ui change was required.** The shared conformance walker already skips `indexedPair` fields and the plugin-ui wizard already folds the conventional `authentication` group into the General step — both were in place from the graphite/prometheus work.

> Note: the entry `README.md` still carries a "Custom HTTP headers are not modeled" line (Modeling decisions). It is now stale, but `README.md` is **out of scope** for this change (schema-only edits), so it was left untouched — see [Out of scope](#out-of-scope).

---

## Section layout

The `groups` were left in their existing taxonomy; only the new field was slotted into
**Advanced HTTP settings**. Verified rendering top-to-bottom in the new UI sidebar/accordion:

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Connection** (`connection`) | no | URL |
| 2 | **Authentication** (`authentication`) | no | Authentication method (discriminator) → User → Password → Authentication token |
| 3 | **TLS settings** (`tls-settings`) | yes | Add self-signed certificate → CA Certificate, TLS Client Authentication → ServerName → Client Certificate → Client Key, Skip TLS certificate validation |
| 4 | **Advanced HTTP settings** (`advanced-http`) | yes | Allowed cookies, Timeout, **Custom HTTP Headers** ➕ |
| 5 | **Advanced options** (`advanced-options`) | yes | Limit number of results, Enable preview mode, Enable async queries → Min → Max, Auto cancel timeout, Timeout in seconds, Set maximum status buckets, Internal fields filtration → Internal field pattern, Set time stamp field, Set fields search mode, Set variables search mode, Set default earliest time |
| 6 | **Data links** (`data-links`) | yes | Data links (array of objects) |

Notes:

- The **Custom HTTP Headers** field is placed in **Advanced HTTP settings** (alongside Allowed
  cookies + Timeout), matching where the legacy editor keeps the HTTP-transport knobs (the legacy
  editor renders a separate **HTTP headers** heading between TLS and Advanced HTTP settings; the new
  UI groups it under Advanced HTTP settings).
- `jsonData.streamMode` is a `legacy`-tagged field the editor no longer renders (the backend migrates
  a `true` value into `previewMode`). It belongs to no group and is intentionally not surfaced in
  either UI. Parity preserved.
- `optional` groups render collapsed/collapsible in tab mode (verified: expanding **Advanced HTTP
  settings** was required before the headers editor appeared in the DOM).

### Wizard mode: URL in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the
fields of the auth group, plus their `dependsOn` parents/children.

- **Before:** `root_url` used `requiredWhen: "true"` (a CEL expression the resolver does **not**
  inspect), so URL was **not** in General.
- **After:** changing it to `required: true` (unconditionally required — the backend needs a URL to
  build its Splunk REST endpoints, `pkg/splunk/client.go:67-68`, and `Config.Validate` rejects an
  empty URL) puts URL into General and emits a proper OpenAPI `required: ["url"]` in the generated
  spec (instead of the `x-dsconfig-required-when: "true"` extension).

**Verified:** in both the tab and wizard stories the URL input (`placeholder="URL"`) is present and
the **Authentication method** select renders; the wizard body contains the **General** step
(`urlInputPresent: true`, `authMethodPresent: true`, `generalStepText: true` — `verify-splunk-url.json`).

The auth group folds into General because it uses the conventional `id: "authentication"`, which the
plugin-ui wizard already recognises. **No plugin-ui change was needed.**

---

## Field-by-field parity

Legend: ✅ present & matching · ➕ added by this change · 🔀 conditional (revealed by `dependsOn`) · ⚙ driven by the `jsonData_authType` discriminator · 🔒 backend/legacy-only (no editor UI in either)

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| URL | text input | `root_url` | input | `root.url` (required) | ✅ |
| Authentication method | select (Basic / Alternative / Forward OAuth) | `jsonData_authType` | select (discriminator) | `jsonData.authType` | ✅ ⚙ |
| — (managed) | — | `jsonData_oauthPassThru` | (hidden, managed) | `jsonData.oauthPassThru` | ✅ ⚙ |
| User | text input | `root_basicAuthUser` | input | `root.basicAuthUser` | ✅ 🔀 |
| Password | password (secure) | `secureJsonData_basicAuthPassword` | secure input | `secureJsonData.basicAuthPassword` | ✅ 🔀 |
| Authentication token | textarea (secure) | `secureJsonData_authToken` | secure input¹ | `secureJsonData.authToken` | ✅ 🔀 |
| Add self-signed certificate | switch | `jsonData_tlsAuthWithCACert` | switch | `jsonData.tlsAuthWithCACert` | ✅ |
| CA Certificate | textarea | `secureJsonData_tlsCACert` | secure input¹ | `secureJsonData.tlsCACert` | ✅ 🔀 |
| TLS Client Authentication | switch | `jsonData_tlsAuth` | switch | `jsonData.tlsAuth` | ✅ |
| ServerName | text input | `jsonData_serverName` | input | `jsonData.serverName` | ✅ 🔀 |
| Client Certificate | textarea | `secureJsonData_tlsClientCert` | secure input¹ | `secureJsonData.tlsClientCert` | ✅ 🔀 |
| Client Key | textarea | `secureJsonData_tlsClientKey` | secure input¹ | `secureJsonData.tlsClientKey` | ✅ 🔀 |
| Skip TLS certificate validation | switch | `jsonData_tlsSkipVerify` | switch | `jsonData.tlsSkipVerify` | ✅ |
| Allowed cookies | TagsInput | `jsonData_keepCookies` | list (string array) | `jsonData.keepCookies` | ✅ |
| Timeout | number | `jsonData_timeout` | number | `jsonData.timeout` | ✅ |
| **HTTP headers** | Add header → name input + value password | `jsonData_httpHeaders` | IndexedPair editor (key + secret value) | `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>` | ➕ |
| Limit number of results | number | `jsonData_maxResultCount` | input | `jsonData.maxResultCount` | ✅ |
| Enable preview mode | switch | `jsonData_previewMode` | switch | `jsonData.previewMode` | ✅ |
| Enable async queries | switch | `jsonData_pollSearchResult` | switch | `jsonData.pollSearchResult` | ✅ |
| Min (poll interval) | text input | `jsonData_minPollInterval` | input | `jsonData.minPollInterval` | ✅ 🔀 |
| Max (poll interval) | text input | `jsonData_maxPollInterval` | input | `jsonData.maxPollInterval` | ✅ 🔀 |
| Auto cancel timeout | text input | `jsonData_autoCancel` | input | `jsonData.autoCancel` | ✅ |
| Timeout in seconds | number | `jsonData_timeoutInSeconds` | input | `jsonData.timeoutInSeconds` | ✅ |
| Set maximum status buckets | text input | `jsonData_statusBuckets` | input | `jsonData.statusBuckets` | ✅ |
| Internal fields filtration | switch | `jsonData_internalFieldsFiltration` | switch | `jsonData.internalFieldsFiltration` | ✅ |
| Internal field pattern | text input | `jsonData_internalFieldPattern` | input | `jsonData.internalFieldPattern` | ✅ 🔀 |
| Set time stamp field | text input | `jsonData_tsField` | input | `jsonData.tsField` | ✅ |
| Set fields search mode | select (quick/full) | `jsonData_fieldSearchType` | select | `jsonData.fieldSearchType` | ✅ |
| Set variables search mode | select (fast/smart/verbose) | `jsonData_variableSearchLevel` | select | `jsonData.variableSearchLevel` | ✅ |
| Set default earliest time | text input | `jsonData_defaultEarliestTime` | input | `jsonData.defaultEarliestTime` | ✅ |
| Data links | repeated data-link rows | `jsonData_dataLinks` | array-of-objects editor | `jsonData.dataLinks` | ✅ |
| — (not in editor) | — | `jsonData_streamMode` | — | `jsonData.streamMode` | ✅ 🔒 |

¹ **Not a discrepancy.** The `dsconfig.json` declares `ui.component: "textarea"` for the token / TLS
cert fields, but the new renderer draws any `target: "secureJsonData"` field as a masked secure input
with a show/hide toggle (the same policy documented for prometheus/graphite). Both UIs collect the
same text into the same `secureJsonData` keys — only the widget affordance differs.

---

## Gap found and fixed: Custom HTTP Headers

**Legacy behaviour (verified in `legacy-expand-splunk-verify.json`):** the editor includes an
**HTTP headers** section heading with an **Add header** button (`hasCustomHeaders: true`,
`addHeaderBtn: true`). `@grafana/plugin-ui`'s CustomHeaders component persists headers as indexed
pairs — `jsonData.httpHeaderName<N>` for the name and `secureJsonData.httpHeaderValue<N>` for the
(secret) value, starting at `N = 1`.

**Before:** `dsconfig.json` did not model headers at all (the README's Modeling decisions explicitly
excluded them, following the pre-fix Prometheus entry), so the new UI rendered no headers editor
(`hasHeadersEditor: false`).

**Fix (in `dsconfig.json` only):** added the `jsonData_httpHeaders` array field with an `indexedPair`
storage mapping that reproduces the exact legacy storage, plus item sub-fields for the header name
(`http.header.name`, with a header-name pattern validation `^[A-Za-z][A-Za-z0-9-]*$`) and value
(`http.header.value`), and added it to the **Advanced HTTP settings** group's `fieldRefs`. Copied
verbatim from the prometheus entry.

**After (verified in `newgen-splunk-fixed.json`):** the new UI renders a **Custom HTTP Headers** row
under **Advanced HTTP settings** with an **Add custom http header** button and a key/secret-value
editor (`hasHeadersEditor: true`).

---

## `fileUpload` evaluation — not applicable to splunk

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not:

- The legacy Splunk editor renders the Auth token / CA Cert / Client Cert / Client Key fields as
  **plain textareas** (`Begins with --- BEGIN CERTIFICATE ---` / `--- RSA PRIVATE KEY CERTIFICATE ---`).
  No file-upload button and no `<input type="file">` were found in the legacy DOM
  (`legacy-expand-splunk-verify.json`: `fileInputs: 0`, `uploadButtons: []`).
- The new UI's `fileUpload` component only activates for `ui.fileMapping` (multi-key JSON
  distribution, e.g. a GCP service-account file); it does not model single-PEM upload.

**Decision:** do **not** add `fileUpload` to any splunk field.

---

## Conditional fields & the auth discriminator — tested

### `jsonData_authType` discriminator

Unlike prometheus (which uses a `virtual` selector with `effects`), splunk models auth as a **real
`jsonData` discriminator** (`jsonData_authType`, `role: auth.discriminator`) whose value is written
verbatim, paired with a managed `jsonData_oauthPassThru` boolean (set true only for `OAuthForward`).
The three methods drive these `dependsOn` reveals:

| Selection | `jsonData.authType` | Revealed field(s) |
| --- | --- | --- |
| **Basic authentication** (default) | `BasicAuth` (or empty) | User + Password (🔀 `dependsOn: authType == 'BasicAuth' \|\| authType == ''`) |
| **Alternative authentication** | `custom-splunk` | Authentication token (🔀 `dependsOn: authType == 'custom-splunk'`) |
| **Forward OAuth Identity** | `OAuthForward` | none (managed `jsonData.oauthPassThru = true`) |

Directly observed: on the default (BasicAuth) the **User** and **Password** inputs are present, and
the `Save & Test` save is **blocked** until both are filled — matching the
`requiredWhen: authType == 'BasicAuth' || authType == ''` contract (an empty attempt logged 0
payloads; filling URL + User + Password produced exactly 1 payload).

### `dependsOn` conditionals (declared, unchanged by this change)

| Trigger | Revealed field(s) |
| --- | --- |
| `authType == 'BasicAuth' \|\| authType == ''` | User, Password |
| `authType == 'custom-splunk'` | Authentication token |
| `jsonData_tlsAuth == true` | ServerName, Client Certificate, Client Key |
| `jsonData_tlsAuthWithCACert == true` | CA Certificate |
| `jsonData_pollSearchResult == true \|\| jsonData_previewMode == true` | Min, Max (poll interval) |
| `jsonData_internalFieldsFiltration == true` | Internal field pattern |

### Save-payload storage-target validation

Filling the form (URL + Basic-auth user/password) and adding a custom header (name `X-Api-Token`,
value `super-secret-token`), then clicking **Save & Test**, logs the exact datasource payload the
wizard would PUT — `verify-splunk-headers.json`:

```json
{
  "url": "https://splunk.example.com:8089",
  "basicAuthUser": "splunk_admin",
  "jsonData":       { "authType": "BasicAuth", "httpHeaderName1": "X-Api-Token", ... },
  "secureJsonData": { "basicAuthPassword": "s3cret", "httpHeaderValue1": "super-secret-token" },
  "secureJsonFields": { "httpHeaderValue1": false }
}
```

The header **name lands in `jsonData.httpHeaderName1`** and the **value in
`secureJsonData.httpHeaderValue1`** (with `secureJsonFields.httpHeaderValue1: false`) — byte-for-byte
the legacy CustomHeaders storage format. URL routes to `root.url`; the Basic-auth username routes to
`root.basicAuthUser` and password to `secureJsonData.basicAuthPassword`, exactly as declared.

---

## Conformance (no change required)

Adding an `indexedPair` field would normally break `JSONDataMatchesStruct` /
`JSONDataTypesMatchStruct`, because an `indexedPair` field (`jsonData_httpHeaders`) is a **logical
view** over dynamically-indexed legacy keys (`httpHeaderName1`, `httpHeaderValue1`, …) that are not
modeled as a single Go struct field. The shared conformance walker (`schema/conformance.go`) already
skips `indexedPair` fields via `isIndexedPairField()` (lines 157, 180, 245-247) — plugin-agnostic —
so **no conformance change was needed here**. The per-header `httpHeaderValue<N>` secrets remain
dynamic and are correctly **not** listed among the static `SecureJsonDataKeys` (`settings.go:105-111`);
`Config` needs no new struct field. The generated spec emits `httpHeaders` as a clean array under
`jsonData` with **no** secure values leaked (`SchemaSpecHasNoSecureJSON` passes).

---

## Verification

```
go generate ./registry/grafana-splunk-datasource/...   # regenerate schema.gen.json / settings.gen.json  → PASS
go test     ./registry/grafana-splunk-datasource/...   # PASS
```

`TestSchemaConformance` subtests (grafana-splunk-datasource) — all **PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `TestLoadConfig`, `TestApplyDefaults`, and `TestValidate` suites also pass
unchanged (they do not reference the new field).

Generated-spec deltas (`schema.gen.json` / `settings.gen.json`): `required: ["url"]` added to the
spec; `x-dsconfig-required-when: "true"` removed from `url`; `httpHeaders` array added under
`jsonData` (items `required: ["name"]`, name `pattern: ^[A-Za-z][A-Za-z0-9-]*$`).

New-UI captures: `newgen-splunk-fixed` (tab, `hasHeadersEditor: true`), `verify-splunk-url`
(tab + wizard: URL input present, auth method present, wizard General step),
`verify-splunk-headers` (header save-payload routing). Legacy capture:
`legacy-expand-splunk-verify` (HTTP headers + Add header present; 0 file inputs).

---

## Out of scope

- `README.md` — its "Modeling decisions" section still states custom HTTP headers are "not modeled".
  This is now stale after change #2, but `README.md` is out of scope for this schema-only change and
  was not edited. The embedded `dsconfig.json` `llm`/`settings`/`legacy` instruction **was** updated
  to state headers are now modeled.
- `settings.go`, `settings.ts`, `conformance.go`, `plugin-ui` — unchanged by design; the `indexedPair`
  field needs no struct field (walker skips it) and no plugin-ui/conformance change.

---

## Files changed

- [`registry/grafana-splunk-datasource/dsconfig.json`](dsconfig.json) — changed `root_url` from
  `requiredWhen: "true"` to `required: true` (renders in the wizard's General step); added the
  `jsonData_httpHeaders` `indexedPair` field and referenced it from the `advanced-http` group;
  updated the `settings`/`legacy` instruction so it states headers are now modeled.
- [`registry/grafana-splunk-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-splunk-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`url` now in the spec's `required` array; `httpHeaders` array added under `jsonData`;
  `x-dsconfig-required-when: "true"` removed from `url`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema/conformance.go`, and everything in `plugin-ui`.
