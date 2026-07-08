# Falcon LogScale — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-falconlogscale-datasource` (CrowdStrike Falcon LogScale / NGSIEM)
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/cfrbqix509jpcc` (Grafana Enterprise; the plugin's own `ConfigEditor` + `@grafana/plugin-ui` `Auth` via `convertLegacyAuthProps`)
- **New UI:** `http://192.168.1.241:58899/iframe.html?args=pluginType:grafana-falconlogscale-datasource&id=configeditor-datasourceconfigwizard--tab` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + `Save & Test`-button console payloads). The Storybook story fetches the schema from `raw.githubusercontent.com/.../schema-discovery/registry/grafana-falconlogscale-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the captures reflect the local schema without pushing.
- **Result:** **Parity achieved for the modeled surface.** One missing field was found (**Custom HTTP Headers**) and added; `root_url` was promoted to unconditionally required so it enters the wizard **General** step; the four-way `virtual_authMethod` selector `effects` and every `dependsOn` conditional were exercised and their storage targets verified from the save payload. `fileUpload` was evaluated and correctly **not** used. **One legacy section (TLS settings) is not modeled and cannot be added via `dsconfig.json` alone** — see [Known gap](#known-gap-tls-settings-not-fixable-via-dsconfigjson-alone).

---

## TL;DR of changes

| #   | Change                                                                                           | File                             | Why                                                                                                            |
| --- | ------------------------------------------------------------------------------------------------ | -------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| 1   | Changed `root_url` from `requiredWhen: "true"` → `required: true`                                | [`dsconfig.json`](dsconfig.json) | URL is unconditionally required (backend returns "URL can not be blank", `pkg/plugin/settings.go:41`); puts URL into the wizard's synthetic **General** step and emits `required: ["url"]` |
| 2   | Added **Custom HTTP Headers** field (`jsonData_httpHeaders`) with `indexedPair` storage (verbatim from prometheus) | [`dsconfig.json`](dsconfig.json) | Legacy UI renders an **HTTP headers** section with **Add header**; new UI had no headers editor                |
| 3   | Added `jsonData_httpHeaders` to the `advanced-http` group's `fieldRefs` (after `jsonData_timeout`) | [`dsconfig.json`](dsconfig.json) | Surface the new field under **Advanced settings**, matching the legacy grouping                               |
| 4   | Added a truthful `llm`/`settings` instruction documenting the modeled headers                     | [`dsconfig.json`](dsconfig.json) | No pre-existing "not modeled" header instruction existed; keep the embedded instructions complete after change #2 |
| 5   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-falconlogscale-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`)                                          |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance_test.go`, `settings.examples.gen.json`, or `plugin-ui`. All schema changes flow through `dsconfig.json`; the rest is produced by `go generate`.

**No conformance-test or plugin-ui change was required** for the changes made — the shared conformance walker already skips `indexedPair` fields (`schema/conformance.go:157,180,245-247`) and the plugin-ui wizard already folds the `authentication` group + `required` fields into the **General** step. Both were already in place from earlier (graphite/prometheus) work.

---

## Section layout

The `groups` were left in their existing taxonomy; only the new field was slotted into
**Advanced settings**. Verified rendering top-to-bottom in the new UI (tab mode, capture
`newgen-falcon-fixed`):

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Connection** (`connection`) | no | URL, Mode |
| 2 | **Authentication** (`authentication`) | no | Authentication method (virtual) → Token / Client ID + Client Secret / User + Password (revealed per method) |
| 3 | **Advanced settings** (`advanced-http`) | yes | Allowed cookies, Timeout, **Custom HTTP Headers** ➕ |
| 4 | **Additional settings** (`additional-settings`) | yes | Default Repository, Data links, Incremental querying → Query overlap window |

Notes:

- The **Custom HTTP Headers** field is placed in **Advanced settings** (alongside Allowed
  cookies + Timeout), matching where the legacy editor keeps the HTTP-transport knobs.
- Four `jsonData`/`root` fields (`jsonData_authenticateWithToken`, `jsonData_oauth2`,
  `jsonData_oauthPassThru`, `root_basicAuth`) are **managed** by the `virtual_authMethod`
  selector (`managed-by:virtual_authMethod`) and belong to no group — they render as hidden
  managed fields, not standalone controls.
- `jsonData_baseUrl` is **frontend-only** (a round-trip snapshot of `root.url` written by the
  LogScale token component, `ConfigEditor.tsx:155`, never read by the backend) and belongs to
  no group. Parity preserved (neither UI shows it as a control).
- `optional` groups (**Advanced settings**, **Additional settings**) render collapsed/collapsible
  in tab mode.

### Wizard mode: URL in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the
fields of the auth group, plus their `dependsOn` parents/children.

- **Before:** `root_url` used `requiredWhen: "true"` (a CEL expression the resolver does **not**
  inspect), so URL was **not** in General. Verified (`inspect-falcon-wizard-before`): the
  General **1/5** step contained only **Authentication method** → **Token\*** (folded from the
  auth group); `urlInputs: 0`.
- **After:** changing it to `required: true` (unconditionally required) puts URL into General and
  emits a proper OpenAPI `required: ["url"]` in the generated spec (instead of the
  `x-dsconfig-required-when: "true"` extension).

**Verified (`inspect-falcon-wizard-after`, screenshot `verify-falcon-wizard`):** the wizard opens
on **General 1/5** containing **URL\*** (with the required asterisk) followed by the
**Authentication method** select (defaulting to "LogScale Token Authentication") and its
**Token\*** field; `urlInputs: 1`, `authMethodPresent: true`.

The auth group folds into General because it uses the conventional `id: "authentication"`, which
the plugin-ui wizard already recognises. **No plugin-ui change was needed.**

---

## Field-by-field parity

Legend: ✅ present & matching · ➕ added by this change · 🔀 conditional (revealed by `dependsOn`) · ⚙ driven by the `virtual_authMethod` selector · 🧩 frontend-only round-trip · ❌ legacy-only, not modeled (see [Known gap](#known-gap-tls-settings-not-fixable-via-dsconfigjson-alone))

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| URL \* | text input | `root_url` | input | `root.url` (required) | ✅ |
| Mode | select (LogScale / NGSIEM) | `jsonData_mode` | select | `jsonData.mode` | ✅ |
| Authentication method | select (4 methods) | `virtual_authMethod` | select + `effects` | virtual → 4 managed flags | ✅ ⚙ |
| — (managed) | — | `jsonData_authenticateWithToken` | (hidden, managed) | `jsonData.authenticateWithToken` | ✅ ⚙ |
| — (managed) | — | `jsonData_oauth2` | (hidden, managed) | `jsonData.oauth2` | ✅ ⚙ |
| — (managed) | — | `jsonData_oauthPassThru` | (hidden, managed) | `jsonData.oauthPassThru` | ✅ ⚙ |
| — (managed) | — | `root_basicAuth` | (hidden, managed) | `root.basicAuth` | ✅ ⚙ |
| Token | password (secure) | `secureJsonData_accessToken` | secure input | `secureJsonData.accessToken` | ✅ 🔀 (custom-token) |
| Client ID | text input | `jsonData_oauth2ClientId` | input | `jsonData.oauth2ClientId` | ✅ 🔀 (oauth) |
| Client Secret | password (secure) | `secureJsonData_oauth2ClientSecret` | secure input | `secureJsonData.oauth2ClientSecret` | ✅ 🔀 (oauth) |
| User | text input | `root_basicAuthUser` | input | `root.basicAuthUser` | ✅ 🔀 (basic) |
| Password | password (secure) | `secureJsonData_basicAuthPassword` | secure input | `secureJsonData.basicAuthPassword` | ✅ 🔀 (basic) |
| Add self-signed certificate | switch | — | — | `jsonData.tlsAuthWithCACert` | ❌ |
| TLS Client Authentication | switch | — | — | `jsonData.tlsAuth` (+ serverName / client cert / client key) | ❌ |
| Skip TLS certificate validation | switch | — | — | `jsonData.tlsSkipVerify` | ❌ |
| **HTTP headers** (Add header) | name input + value password | `jsonData_httpHeaders` | IndexedPair editor (key + secret value) | `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>` | ➕ |
| Allowed cookies | TagsInput | `jsonData_keepCookies` | list (string array) | `jsonData.keepCookies` | ✅ |
| Timeout | number | `jsonData_timeout` | number | `jsonData.timeout` | ✅ |
| Default Repository | select (allowCustom) | `jsonData_defaultRepository` | select | `jsonData.defaultRepository` | ✅ |
| Data links | repeated rows | `jsonData_dataLinks` | array-of-objects editor | `jsonData.dataLinks` | ✅ |
| Incremental querying (experimental) | switch | `jsonData_incrementalQuerying` | switch | `jsonData.incrementalQuerying` | ✅ |
| Query overlap window | text input | `jsonData_incrementalQueryOverlapWindow` | input | `jsonData.incrementalQueryOverlapWindow` | ✅ 🔀 |
| — (frontend snapshot) | — | `jsonData_baseUrl` | — | `jsonData.baseUrl` | ✅ 🧩 |

"Name" and "Default" in the legacy DOM are Grafana's standard datasource chrome (instance name +
"set as default" toggle), not plugin configuration — present in every datasource editor and
intentionally not part of the schema.

Secure fields (`accessToken`, `oauth2ClientSecret`, `basicAuthPassword`) declare
`ui.component: "input"` but the new renderer draws any `target: "secureJsonData"` field as a
masked secure input with a show/hide toggle. Both UIs collect the same values into the same
`secureJsonData` keys — only the widget affordance differs.

---

## Gap found and fixed: Custom HTTP Headers

**Legacy behaviour (verified in `legacy-expand-ent-grafana-falconlogscale-datasource.json` and
`dump-legacy-fields-falcon.json`):** the editor includes an **HTTP headers** section heading with
an **Add header** button (`hasCustomHeaders: true`, `addHeaderBtn: true`). `@grafana/plugin-ui`'s
CustomHeaders component persists headers as indexed pairs — `jsonData.httpHeaderName<N>` for the
name and `secureJsonData.httpHeaderValue<N>` for the (secret) value, starting at `N = 1`.

**Before:** `dsconfig.json` did not model headers at all, so the new UI rendered no headers editor
(`hasHeadersEditor: false` in `newgen-falcon-before.json`).

**Fix (in `dsconfig.json` only):** added the `jsonData_httpHeaders` array field with an
`indexedPair` storage mapping (copied verbatim from `registry/prometheus/dsconfig.json`) that
reproduces the exact legacy storage, plus item sub-fields for the header name
(`http.header.name`, with a header-name pattern validation) and value (`http.header.value`), and
added it to the **Advanced settings** group's `fieldRefs` after `jsonData_timeout`.

**After (verified in `newgen-falcon-fixed.json` / screenshot `verify-falcon-tab`):** the new UI
renders a **Custom HTTP Headers** row under **Advanced settings** with an **Add custom http
header** button and a key/secret-value editor (`hasHeadersEditor: true`).

**Save-payload routing (verified `verify-falcon-result.json`):** a header (name `X-Falcon-Token`,
value `falcon-secret-123`) routed to:

```json
{
  "url": "https://cloud.humio.com",
  "jsonData":       { "httpHeaderName1": "X-Falcon-Token", "authenticateWithToken": true },
  "secureJsonData": { "httpHeaderValue1": "falcon-secret-123", "accessToken": "logscale-tok" }
}
```

The header **name lands in `jsonData.httpHeaderName1`** and the **value in
`secureJsonData.httpHeaderValue1`** — byte-for-byte the legacy CustomHeaders storage format.

---

## `fileUpload` evaluation — not applicable to falconlogscale

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not:

- The legacy editor has **no** file inputs and **no** upload buttons
  (`legacy-expand-ent-grafana-falconlogscale-datasource.json`: `fileInputs: 0`,
  `uploadButtons: []`). No packs.
- The new UI's `fileUpload` component only activates for `ui.fileMapping` (multi-key JSON
  distribution, e.g. a GCP service-account file); falcon has no such input.

**Decision:** do **not** add `fileUpload` to any falconlogscale field.

---

## Conditional fields & effects — tested

### `virtual_authMethod` selector (`effects`) — four-way

Falcon models auth as a **virtual selector** (`virtual_authMethod`) with **four** methods. Its
value is `read` from a computed expression over `jsonData.oauth2` / `jsonData.authenticateWithToken`
/ `root.basicAuth` / `jsonData.oauthPassThru`, and its `effects` fan out to those four managed
fields, enforcing mutual exclusivity (exactly the invariant `settings.go` `Validate()` checks:
"only one authentication method may be enabled at a time"). All four branches were driven fresh
from a clean page load and verified from the `Save & Test` console payload
(`verify-falcon-result.json`):

| Selection (label) | `authenticateWithToken` | `oauth2` | `oauthPassThru` | `root.basicAuth` | secure key written |
| --- | --- | --- | --- | --- | --- |
| **LogScale Token Authentication** (`custom-token`, default) | **true** | false | false | false | `accessToken` |
| **OAuth2 Client Credentials** (`custom-oauth-client-secret`) | false | **true** | false | false | `oauth2ClientSecret` (+ `jsonData.oauth2ClientId`) |
| **Basic authentication** (`BasicAuth`) | false | false | false | **true** | `basicAuthPassword` (+ `root.basicAuthUser`) |
| **Forward OAuth Identity** (`OAuthForward`) | false | false | **true** | false | — (no secret) |

Every branch produced exactly one `true` flag with the other three `false` — the `set` operations
propagate, not just the visibility. The dependent credential fields appeared on selection
(🔀 `dependsOn: virtual_authMethod == '<method>'`) and routed to the correct
`root` / `jsonData` / `secureJsonData` targets.

### `dependsOn` conditionals

| Trigger | Revealed field(s) | Verified |
| --- | --- | --- |
| `virtual_authMethod == 'custom-token'` | Token → `secureJsonData.accessToken` | ✅ |
| `virtual_authMethod == 'custom-oauth-client-secret'` | Client ID, Client Secret | ✅ route to `jsonData.oauth2ClientId` / `secureJsonData.oauth2ClientSecret` |
| `virtual_authMethod == 'BasicAuth'` | User, Password | ✅ route to `root.basicAuthUser` / `secureJsonData.basicAuthPassword` |
| `jsonData_incrementalQuerying == true` | Query overlap window | ✅ (declared; in Additional settings) |

---

## Known gap: TLS settings (not fixable via `dsconfig.json` alone)

**Legacy behaviour (verified `dump-legacy-fields-falcon.json`):** the legacy editor renders a
**TLS settings** section — supplied by the shared `@grafana/plugin-ui` HTTP/auth component via
`convertLegacyAuthProps` (`ConfigEditor.tsx:171-174`) — with three switches:

- **Add self-signed certificate** (`jsonData.tlsAuthWithCACert` + `secureJsonData.tlsCACert`)
- **TLS Client Authentication** (`jsonData.tlsAuth` + `jsonData.serverName` +
  `secureJsonData.tlsClientCert` + `secureJsonData.tlsClientKey`)
- **Skip TLS certificate validation** (`jsonData.tlsSkipVerify`)

These are **not modeled** in `dsconfig.json`, so the new UI does not render them.

**Why it was not fixed here:** falcon's Go settings model (`settings.go` `Config`) declares **no
TLS fields**. The shared conformance suite's `JSONDataMatchesStruct` / `JSONDataTypesMatchStruct`
require every `jsonData` field in the schema to have a matching struct field (unless it is an
`indexedPair` view or a `json:"-"` root field). Adding `tlsAuth`/`tlsCACert`/… to `dsconfig.json`
would therefore **fail conformance** unless matching fields were added to `settings.go` — which is
**out of scope** (`settings.go` must not be edited). Per the task's stop-and-report rule, this is
reported rather than changed.

**Impact:** at the transport layer TLS still functions when provisioned (the SDK's
`HTTPClientOptions` reads these keys generically), but a user configuring via the **new UI** cannot
toggle TLS. Closing this gap requires a coordinated `settings.go` + `dsconfig.json` change (and
regeneration) and should be handled as a separate, non-`dsconfig.json`-only task.

---

## Conformance (no change required)

Adding an `indexedPair` field would normally break `JSONDataMatchesStruct` /
`JSONDataTypesMatchStruct`, because `jsonData_httpHeaders` is a **logical view** over
dynamically-indexed legacy keys (`httpHeaderName1`, `httpHeaderValue1`, …) not modeled as a single
Go struct field. The shared conformance walker (`schema/conformance.go`) already skips
`indexedPair` fields via `isIndexedPairField()` (lines 157, 180, 245-247) — plugin-agnostic — so
**no conformance change was needed**. The per-header `httpHeaderValue<N>` secrets remain dynamic
and are correctly **not** listed among the static `SecureJsonDataKeys` (`settings.go:74-78`), and
the generated spec emits `httpHeaders` as a clean array under `jsonData` with **no** secure values
leaked (`SchemaSpecHasNoSecureJSON` passes).

---

## Verification

```
go generate ./registry/grafana-falconlogscale-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-falconlogscale-datasource/...       # PASS
```

`TestSchemaConformance` subtests — all **PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `TestLoadConfig`, `TestApplyDefaults`, and `TestValidate` suites also pass
unchanged (they do not reference the new field).

Generated-spec deltas: `required: ["url"]` added at the spec level; `x-dsconfig-required-when:
"true"` removed from `url`; `httpHeaders` array added under `jsonData` (item `name` with pattern +
`required: ["name"]`); no `httpHeaderValue` secret leaked into the spec.

New-UI captures: `newgen-falcon-before` (tab, `hasHeadersEditor: false`) →
`newgen-falcon-fixed` (tab, `hasHeadersEditor: true`, all 4 sections);
`inspect-falcon-wizard-before` (General 1/5: no URL) → `inspect-falcon-wizard-after` /
`verify-falcon-wizard` (General 1/5 with **URL\*** + auth method); `verify-falcon-result`
(4 auth-method effects + header save-payload routing). Legacy captures:
`legacy-expand-ent-grafana-falconlogscale-datasource` (HTTP headers + Add header present; 0 file
inputs) and `dump-legacy-fields-falcon` (full label/switch list, including the unmodeled TLS
switches).

---

## Files changed

- [`registry/grafana-falconlogscale-datasource/dsconfig.json`](dsconfig.json) — changed
  `root_url` from `requiredWhen: "true"` to `required: true`; added the `jsonData_httpHeaders`
  `indexedPair` field and referenced it from the `advanced-http` group; added an `llm`/`settings`
  instruction documenting the now-modeled headers.
- `registry/grafana-falconlogscale-datasource/schema.gen.json`,
  `registry/grafana-falconlogscale-datasource/settings.gen.json` — regenerated by `go generate`.

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`conformance_test.go`, `schema/conformance.go`, and everything in `plugin-ui`.

_Reported, not changed (out of scope):_ the legacy **TLS settings** section — modeling it requires
`settings.go` struct fields, not a `dsconfig.json`-only change.
```
