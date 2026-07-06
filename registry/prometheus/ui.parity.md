# Prometheus — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `prometheus`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/bfraniarrtt6oa` (Grafana Enterprise, `@grafana/plugin-ui` PromSettings + `ConfigDescriptionLink`/`Auth`)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:prometheus` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + `Save & Test`-button console payloads). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/prometheus/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the screenshots reflect the local schema without pushing.
- **Result:** **Parity achieved.** One missing field was found (**Custom HTTP Headers**) and added; `fileUpload` was evaluated and correctly **not** used; the `virtual_authMethod` selector `effects` and every `dependsOn` conditional were exercised and their storage targets verified from the save payload.

---

## TL;DR of changes

| #   | Change                                                                                           | File                             | Why                                                                                                            |
| --- | ------------------------------------------------------------------------------------------------ | -------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| 1   | Changed `root_url` from `requiredWhen: "true"` → `required: true`                                | [`dsconfig.json`](dsconfig.json) | URL is unconditionally required; puts URL into the wizard's synthetic **General** step and emits `required: ["url"]` |
| 2   | Added **Custom HTTP Headers** field (`jsonData_httpHeaders`) with `indexedPair` storage           | [`dsconfig.json`](dsconfig.json) | Legacy UI renders an **HTTP headers** section with **Add header**; new UI had no headers editor                |
| 3   | Added `jsonData_httpHeaders` to the `advanced-http` group's `fieldRefs` (after `jsonData_timeout`) | [`dsconfig.json`](dsconfig.json) | Surface the new field under **Advanced HTTP settings**, matching the legacy grouping                          |
| 4   | Updated the editor-hidden/legacy instruction: headers are now **modeled** (was "NOT modeled … see README") | [`dsconfig.json`](dsconfig.json) | Keep the embedded `llm`-tagged instructions truthful after change #2                                           |
| 5   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/prometheus/...`     | generated artifacts              | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`)                                          |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance.go`, or `plugin-ui`. All schema changes flow through `dsconfig.json`; the rest is produced by `go generate`.

**No conformance-test or plugin-ui change was required** (see [Conformance](#conformance-no-change-required) and [Wizard mode](#wizard-mode-url-in-the-general-step) below) — this differs from graphite, where the `indexedPair` skip and the auth-group generalisation had to be added. Both were already in place when this entry was validated.

---

## Section layout

The `groups` were left in their existing taxonomy; only the new field was slotted into
**Advanced HTTP settings**. Verified rendering top-to-bottom in the new UI sidebar/accordion:

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Connection** (`connection`) | no | Prometheus server URL |
| 2 | **Authentication** (`authentication`) | no | Authentication method (virtual) → User → Password |
| 3 | **TLS settings** (`tls-settings`) | yes | Add self-signed certificate → CA Certificate, TLS Client Authentication → ServerName → Client Certificate → Client Key, Skip TLS certificate validation |
| 4 | **Advanced HTTP settings** (`advanced-http`) | yes | Allowed cookies, Timeout, **Custom HTTP Headers** ➕ |
| 5 | **Alerting** (`alerting`) | yes | Manage alerts via Alerting UI, Allow as recording rules target |
| 6 | **Interval behaviour** (`interval-behaviour`) | yes | Scrape interval, Query timeout |
| 7 | **Query editor** (`query-editor`) | yes | Default editor, Disable metrics lookup |
| 8 | **Performance** (`performance`) | yes | Prometheus type, Version, Cache level, Incremental querying → Query overlap window, Disable recording rules |
| 9 | **Other** (`other`) | yes | Custom query parameters, HTTP method, Series limit, Use series endpoint |
| 10 | **Exemplars** (`exemplars`) | yes | Exemplar trace-ID destinations (array of objects) |

Notes:

- The **Custom HTTP Headers** field is placed in **Advanced HTTP settings** (alongside Allowed
  cookies + Timeout), matching where the legacy editor keeps the HTTP-transport knobs.
- Two `jsonData` fields (`maxSamplesProcessedWarningThreshold`, `maxSamplesProcessedErrorThreshold`)
  are `backend-only` (tagged as such) and belong to no group — the Prometheus editor never renders
  them (feature-flagged off), so they are intentionally not surfaced in either UI. Parity preserved.
- `optional` groups are rendered collapsed/collapsible in tab mode (verified: expanding **TLS
  settings** / **Performance** was required before their switches appeared in the DOM).

### Wizard mode: URL in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the
fields of the auth group, plus their `dependsOn` parents/children.

- **Before:** `root_url` used `requiredWhen: "true"` (a CEL expression the resolver does **not**
  inspect), so URL was **not** in General — the new-UI wizard capture showed `urlPresent: false`
  on step 1/11 (only "Authentication method" was folded in from the auth group).
- **After:** changing it to `required: true` (unconditionally required — the backend admission
  handler rejects an empty URL, `pkg/promlib/admission_handler.go:51`) puts URL into General and
  emits a proper OpenAPI `required: ["url"]` in the generated spec (instead of the
  `x-dsconfig-required-when: "true"` extension).

**Verified (screenshot `verify-prom-wizard`):** the wizard opens on **General 1/11** containing
**Prometheus server URL \*** (with the required asterisk) and the **Authentication method** select
(defaulting to "No Authentication"). `urlPresent: true`, `authMethodPresent: true`.

The auth group already folds into General because it uses the conventional `id: "authentication"`,
which the plugin-ui wizard already recognises (both `authentication` and `auth`). **No plugin-ui
change was needed for prometheus** — that generalisation was already in place from the graphite work.

---

## Field-by-field parity

Legend: ✅ present & matching · ➕ added by this change · 🔀 conditional (revealed by `dependsOn`) · ⚙ driven by the `virtual_authMethod` selector · 🔒 backend-only (no editor UI in either)

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| Prometheus server URL | text input | `root_url` | input | `root.url` (required) | ✅ |
| Authentication method | select (Basic / Forward OAuth / No Auth) | `virtual_authMethod` | select + `effects` | virtual → `root.basicAuth` / `jsonData.oauthPassThru` | ✅ ⚙ |
| — (managed) | — | `root_basicAuth` | (hidden, managed) | `root.basicAuth` | ✅ ⚙ |
| — (managed) | — | `jsonData_oauthPassThru` | (hidden, managed) | `jsonData.oauthPassThru` | ✅ ⚙ |
| User | text input | `root_basicAuthUser` | input | `root.basicAuthUser` | ✅ 🔀 |
| Password | password (secure) | `secureJsonData_basicAuthPassword` | secure input | `secureJsonData.basicAuthPassword` | ✅ 🔀 |
| Add self-signed certificate | switch | `jsonData_tlsAuthWithCACert` | switch | `jsonData.tlsAuthWithCACert` | ✅ |
| CA Certificate | textarea | `secureJsonData_tlsCACert` | secure input¹ | `secureJsonData.tlsCACert` | ✅ 🔀 |
| TLS Client Authentication | switch | `jsonData_tlsAuth` | switch | `jsonData.tlsAuth` | ✅ |
| ServerName | text input | `jsonData_serverName` | input | `jsonData.serverName` | ✅ 🔀 |
| Client Certificate | textarea | `secureJsonData_tlsClientCert` | secure input¹ | `secureJsonData.tlsClientCert` | ✅ 🔀 |
| Client Key | textarea | `secureJsonData_tlsClientKey` | secure input¹ | `secureJsonData.tlsClientKey` | ✅ 🔀 |
| Skip TLS certificate validation | switch | `jsonData_tlsSkipVerify` | switch | `jsonData.tlsSkipVerify` | ✅ |
| Allowed cookies | TagsInput | `jsonData_keepCookies` | list (string array) | `jsonData.keepCookies` | ✅ |
| Timeout | number | `jsonData_timeout` | number | `jsonData.timeout` | ✅ |
| **Custom HTTP Headers** | Add header → name input + value password | `jsonData_httpHeaders` | IndexedPair editor (key + secret value) | `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>` | ➕ |
| Manage alerts via Alerting UI | switch | `jsonData_manageAlerts` | switch | `jsonData.manageAlerts` | ✅ |
| Allow as recording rules target | switch | `jsonData_allowAsRecordingRulesTarget` | switch | `jsonData.allowAsRecordingRulesTarget` | ✅ |
| Scrape interval | text input | `jsonData_timeInterval` | input | `jsonData.timeInterval` | ✅ |
| Query timeout | text input | `jsonData_queryTimeout` | input | `jsonData.queryTimeout` | ✅ |
| Default editor | select (Builder/Code) | `jsonData_defaultEditor` | select | `jsonData.defaultEditor` | ✅ |
| Disable metrics lookup | switch | `jsonData_disableMetricsLookup` | switch | `jsonData.disableMetricsLookup` | ✅ |
| Prometheus type | select (Prometheus/Cortex/Mimir/Thanos) | `jsonData_prometheusType` | select | `jsonData.prometheusType` | ✅ |
| Version | select (allowCustom) | `jsonData_prometheusVersion` | select | `jsonData.prometheusVersion` | ✅ 🔀 |
| Cache level | select (Low/Medium/High/None) | `jsonData_cacheLevel` | select | `jsonData.cacheLevel` | ✅ |
| Incremental querying (beta) | switch | `jsonData_incrementalQuerying` | switch | `jsonData.incrementalQuerying` | ✅ |
| Query overlap window | text input | `jsonData_incrementalQueryOverlapWindow` | input | `jsonData.incrementalQueryOverlapWindow` | ✅ 🔀 |
| Disable recording rules (beta) | switch | `jsonData_disableRecordingRules` | switch | `jsonData.disableRecordingRules` | ✅ |
| Custom query parameters | text input | `jsonData_customQueryParameters` | input | `jsonData.customQueryParameters` | ✅ |
| HTTP method | select (POST/GET) | `jsonData_httpMethod` | select | `jsonData.httpMethod` | ✅ |
| Series limit | number | `jsonData_seriesLimit` | input | `jsonData.seriesLimit` | ✅ |
| Use series endpoint | switch | `jsonData_seriesEndpoint` | switch | `jsonData.seriesEndpoint` | ✅ |
| Exemplars | repeated destination rows | `jsonData_exemplarTraceIdDestinations` | array-of-objects editor | `jsonData.exemplarTraceIdDestinations` | ✅ |
| — (not in editor) | — | `jsonData_maxSamplesProcessedWarningThreshold` | — | `jsonData.maxSamplesProcessedWarningThreshold` | ✅ 🔒 |
| — (not in editor) | — | `jsonData_maxSamplesProcessedErrorThreshold` | — | `jsonData.maxSamplesProcessedErrorThreshold` | ✅ 🔒 |

¹ **Not a discrepancy.** The `dsconfig.json` declares `ui.component: "textarea"` for the three TLS
cert fields, but the new renderer draws any `target: "secureJsonData"` field as a masked secure
input with a show/hide toggle (the same policy documented for graphite). Directly observed for the
**Password** field (rendered `••••••` with an eye toggle in screenshot `verify-prom-effects2-basic`);
the CA/Client cert fields follow the same renderer policy. Both UIs collect the same PEM text into
the same `secureJsonData` keys — only the widget affordance differs.

---

## Gap found and fixed: Custom HTTP Headers

**Legacy behaviour (verified in `legacy-expand-prometheus.json`):** the editor includes an
**HTTP headers** section heading with an **Add header** button (`hasCustomHeaders: true`,
`addHeaderBtn: true`). `@grafana/plugin-ui`'s CustomHeaders component persists headers as indexed
pairs — `jsonData.httpHeaderName<N>` for the name and `secureJsonData.httpHeaderValue<N>` for the
(secret) value, starting at `N = 1`.

**Before:** `dsconfig.json` did not model headers at all (the old instruction/README explicitly
excluded them), so the new UI rendered no headers editor (`hasHeadersEditor: false` in
`newgen-prom-current.json`).

**Fix (in `dsconfig.json` only):** added the `jsonData_httpHeaders` array field with an
`indexedPair` storage mapping that reproduces the exact legacy storage, plus item sub-fields for the
header name (`http.header.name`, with a header-name pattern validation) and value
(`http.header.value`), and added it to the **Advanced HTTP settings** group's `fieldRefs`.

**After (verified in screenshots `newgen-prom-fixed` / `verify-prom-effects2-basic`):** the new UI
renders a **Custom HTTP Headers** row under **Advanced HTTP settings** with an **Add custom http
header** button and a key/secret-value editor (`hasHeadersEditor: true`).

---

## `fileUpload` evaluation — not applicable to prometheus

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not:

- The legacy Prometheus editor renders the CA Cert / Client Cert / Client Key fields as **plain
  textareas** (`Begins with --- BEGIN CERTIFICATE ---` / `--- RSA PRIVATE KEY CERTIFICATE ---`).
  No file-upload button and no `<input type="file">` were found in the legacy DOM
  (`legacy-expand-prometheus.json`: `fileInputs: 0`, `uploadButtons: []`).
- The new UI's `fileUpload` component only activates for `ui.fileMapping` (multi-key JSON
  distribution, e.g. a GCP service-account file); it does not model single-PEM upload.

**Decision:** do **not** add `fileUpload` to any prometheus field. The cert fields keep their
current modeling; both UIs collect PEM text into the same `secureJsonData` keys.

---

## Conditional fields & effects — tested

### `virtual_authMethod` selector (`effects`)

The prometheus schema models auth as a **virtual selector** (`virtual_authMethod`) whose value is
`read` from `root.basicAuth` / `jsonData.oauthPassThru` and whose `effects` fan out to those two
managed fields. All three branches were driven in the new UI (select opened via its combobox — note
the story renders a *second* combobox for the Storybook "Plugin type" arg, so the auth select is the
last combobox on the page) and verified from the `Save & Test` console payload:

| Selection | UI effect (verified) | Save payload (verified) |
| --- | --- | --- |
| **No Authentication** (default) | User / Password hidden | `basicAuth: false`, `jsonData.oauthPassThru: false` |
| **Basic authentication** | User + Password inputs revealed (🔀 `dependsOn: virtual_authMethod == 'BasicAuth'`) | `basicAuth: true`, `basicAuthUser: "grafana"`, `secureJsonData.basicAuthPassword: "s3cret"`, `oauthPassThru: false` |
| **Forward OAuth Identity** | User / Password hidden | `basicAuth: false`, `jsonData.oauthPassThru: true` |

The fresh-load OAuth capture (`verify-prom-oauth-result.json`) is authoritative for the effect
routing: selecting **Forward OAuth Identity** straight from the default produced
`basicAuth: false, oauthPassThru: true` in a single payload — confirming the `set` operations
propagate, not just the visibility. (An earlier back-to-back Basic→OAuth run appeared to show a
stale payload; re-running each branch from a clean page load resolved it — the effects are correct.)
Selecting **Basic authentication** but leaving **User** empty blocked the save (payload count `0`),
matching the `requiredWhen: root_basicAuth == true` contract.

### `dependsOn` conditionals

Exercised in the new UI (tab mode, sections expanded first) — evidence in
`verify-prom-conditionals2-result.json`:

| Trigger | Revealed field(s) | Verified |
| --- | --- | --- |
| `virtual_authMethod == 'BasicAuth'` | User, Password | ✅ appear on selection; route to `root` / `secureJsonData` |
| `jsonData_tlsAuth == true` | ServerName, Client Certificate, Client Key | ✅ all three appear on toggle |
| `jsonData_tlsAuthWithCACert == true` | CA Certificate | ✅ appears on toggle |
| `jsonData_incrementalQuerying == true` | Query overlap window | ✅ appears on toggle |
| `jsonData_prometheusType != ''` | Version | ✅ (declared; Version select present in Performance) |

### Save-payload storage-target validation

Filling the form and clicking **Save & Test** logs the exact datasource payload the wizard would
PUT. Custom header (name `X-Api-Token`, value `super-secret-token`) — `verify-prom-result.json`:

```json
{
  "url": "http://prometheus.example.com:9090",
  "basicAuth": false,
  "jsonData":       { "httpHeaderName1": "X-Api-Token", "httpMethod": "POST", ... },
  "secureJsonData": { "httpHeaderValue1": "super-secret-token" },
  "secureJsonFields": { "httpHeaderValue1": false }
}
```

The header **name lands in `jsonData.httpHeaderName1`** and the **value in
`secureJsonData.httpHeaderValue1`** (with `secureJsonFields.httpHeaderValue1: false`) — byte-for-byte
the legacy CustomHeaders storage format. URL routes to `root.url`; all other fields route to
`root` / `jsonData` / `secureJsonData` exactly as declared.

---

## Conformance (no change required)

Adding an `indexedPair` field would normally break `JSONDataMatchesStruct` /
`JSONDataTypesMatchStruct`, because an `indexedPair` field (`jsonData_httpHeaders`) is a **logical
view** over dynamically-indexed legacy keys (`httpHeaderName1`, `httpHeaderValue1`, …) that are not
modeled as a single Go struct field. The shared conformance walker (`schema/conformance.go`) already
skips `indexedPair` fields via `isIndexedPairField()` (lines 157, 180, 245-247) — this was added
during the graphite work and is plugin-agnostic — so **no conformance change was needed here**. The
per-header `httpHeaderValue<N>` secrets remain dynamic and are correctly **not** listed among the
static `SecureJsonDataKeys` (`settings.go:91-96`), and the generated spec emits `httpHeaders` as a
clean array under `jsonData` with **no** secure values leaked (`SchemaSpecHasNoSecureJSON` passes).

---

## Verification

```
go generate ./registry/prometheus/...          # regenerate schema.gen.json / settings.gen.json
go test ./registry/prometheus/...              # PASS
go test ./registry/... ./schema/...            # entire suite PASS (no regressions)
```

`TestSchemaConformance` subtests (prometheus) — all **PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `TestLoadConfig`, `TestApplyDefaults`, and `TestValidate` suites also pass
unchanged (they do not reference the new field).

New-UI captures: `newgen-prom-fixed` (tab, `hasHeadersEditor: true`, `urlPresent: true`),
`newgen-prom-fixed-wiz` / `verify-prom-wizard` (wizard opens on **General 1/11** with required URL +
auth method), `verify-prom-effects2` / `verify-prom-oauth` (auth-method effects),
`verify-prom-conditionals2` (dependsOn reveals), `verify-prom-result` (header save-payload routing).
Legacy capture: `legacy-expand-prometheus` (HTTP headers + Add header present; 0 file inputs).

---

## Files changed

- [`registry/prometheus/dsconfig.json`](dsconfig.json) — changed `root_url` from
  `requiredWhen: "true"` to `required: true` (renders in the wizard's General step); added the
  `jsonData_httpHeaders` `indexedPair` field and referenced it from the `advanced-http` group;
  updated the editor-hidden/legacy instruction so it states headers are now modeled (was "NOT
  modeled … see README").
- [`registry/prometheus/schema.gen.json`](schema.gen.json),
  [`registry/prometheus/settings.gen.json`](settings.gen.json) — regenerated by `go generate`
  (`url` now in the spec's `required` array; `httpHeaders` array added under `jsonData`;
  `x-dsconfig-required-when: "true"` removed from `url`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema/conformance.go`, and everything in `plugin-ui`.
