# Grafana Mock — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-mock-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/ffrbqixmu1gjka` (the Mock plugin's `ConfigEditor` — plugin-owned CustomHealthCheck block + `@grafana/plugin-ui` `Auth` / `ConnectionSettings` / `CustomHeaders`)
- **New UI:** `http://192.168.1.241:58899/iframe.html?args=pluginType:grafana-mock-datasource&id=configeditor-datasourceconfigwizard--tab` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + `Save & Test`-button console payloads). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/grafana-mock-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the captures reflect the local schema without pushing.
- **Result:** **Parity achieved.** One missing field was found (**Custom HTTP Headers**) and added; `fileUpload` was evaluated and correctly **not** used (0 file inputs in legacy); no `required: true` fix was needed (the Mock legacy editor does not unconditionally require the URL, and no field used `requiredWhen: "true"`); the `virtual_authMethod` selector `effects` and the `dependsOn` conditionals were exercised and their storage targets verified from the save payload.

---

## TL;DR of changes

| #   | Change                                                                                           | File                             | Why                                                                                                            |
| --- | ------------------------------------------------------------------------------------------------ | -------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| 1   | Added **Custom HTTP Headers** field (`jsonData_httpHeaders`) with `indexedPair` storage           | [`dsconfig.json`](dsconfig.json) | Legacy UI renders a **Custom HTTP Headers** section with **Add header**; new UI had no headers editor         |
| 2   | Added `jsonData_httpHeaders` to the `connection` group's `fieldRefs` (after `root_url`)           | [`dsconfig.json`](dsconfig.json) | Surface the new field; Mock has no advanced-http/timeout group, so it slots under **Connection** (the HTTP block) |
| 3   | Added an `llm`-tagged instruction documenting the modeled headers field                           | [`dsconfig.json`](dsconfig.json) | Keep the embedded instructions complete/truthful (there was **no** prior "not modeled" instruction to update)  |
| 4   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-mock-datasource/...` | generated artifacts    | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`)                                          |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance.go`, or `plugin-ui`. All schema changes flow through `dsconfig.json`; the rest is produced by `go generate`.

**No conformance-test or plugin-ui change was required** (see [Conformance](#conformance-no-change-required) below). The shared `indexedPair` skip in `schema/conformance.go` and the auth-group generalisation in `plugin-ui` were already in place from earlier (graphite/prometheus) work.

> **Note — `required: true`:** Unlike prometheus (where `root_url` was changed from `requiredWhen: "true"` to `required: true`), **no such fix applies to Mock.** Mock's `root_url` has neither `required: true` nor `requiredWhen: "true"` — the Mock backend never dials the URL (`root.url is not enforced by the Mock backend`), so it is intentionally left unrequired. The task's step-1 pre-condition ("no unconditional `requiredWhen: "true"`") holds; nothing to change.

---

## Section layout

The `groups` were left in their existing taxonomy; only the new field was slotted into
**Connection**. Verified rendering top-to-bottom in the new UI (tab mode) sidebar/accordion:

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Connection** (`connection`) | no | URL, **Custom HTTP Headers** ➕ |
| 2 | **Authentication** (`authentication`) | no | Authentication method (virtual) → User → Password |
| 3 | **TLS settings** (`tls-settings`) | yes | Add self-signed certificate → CA Certificate, TLS Client Authentication → ServerName → Client Certificate → Client Key, Skip TLS certificate validation |
| 4 | **Custom HealthCheck** (`custom-health-check`) | no | Enable custom health check → Custom status → Custom message → Custom detail json → Skip backend |

Notes:

- The **Custom HTTP Headers** field is placed in **Connection**, alongside URL — Mock has no
  advanced-http/timeout group (as prometheus does), and the legacy editor keeps headers directly
  under the HTTP-transport block. Verified in `newgen-mock-tab.json`: sectionButtons include
  `"Add custom http header"` and fieldLabels include both `"URL"` and `"Custom HTTP Headers"`.
- The **Custom HealthCheck** group is plugin-owned (rendered by the Mock plugin's own `ConfigEditor`,
  not `@grafana/plugin-ui`) and was **already modeled** before this change; it was not touched.

### Wizard mode: General step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the
fields of the auth group, plus their `dependsOn` parents/children.

- Mock has **no** `required: true` field, so General is populated only from the **auth group**.
  **Verified (`newgen-mock-wiz.json`):** the wizard opens on **General 1/5** containing the
  **Authentication method** select (defaulting to "No Authentication"). The five steps are
  General + the four groups (Connection, Authentication, TLS settings, Custom HealthCheck).
- URL and **Custom HTTP Headers** are therefore reached on the **Connection** step (step 2/5),
  not on General — expected, because URL is not required. `hasHeadersEditor: false` on step 1 is
  correct; the headers editor lives on the Connection step (confirmed in tab mode).

The auth group folds into General because it uses the conventional `id: "authentication"`, which
the plugin-ui wizard already recognises. **No plugin-ui change was needed for Mock.**

---

## Field-by-field parity

Legend: ✅ present & matching · ➕ added by this change · 🔀 conditional (revealed by `dependsOn`) · ⚙ driven by the `virtual_authMethod` selector · 🧩 plugin-owned (Mock `ConfigEditor`, pre-existing)

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| URL | text input | `root_url` | input | `root.url` (not required) | ✅ |
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
| **Custom HTTP Headers** | Add header → name input + value password | `jsonData_httpHeaders` | IndexedPair editor (key + secret value) | `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>` | ➕ |
| Enable custom health check | switch | `jsonData_customHealthCheckEnabled` | switch | `jsonData.customHealthCheckEnabled` | ✅ 🧩 |
| Custom status | radio (OK/ERROR/UNKNOWN) | `jsonData_customHealthCheck_status` | radio | `jsonData.customHealthCheck.status` | ✅ 🧩 🔀 |
| Custom message | text input | `jsonData_customHealthCheck_message` | input | `jsonData.customHealthCheck.message` | ✅ 🧩 🔀 |
| Custom detail json | code editor | `jsonData_customHealthCheck_details` | code | `jsonData.customHealthCheck.details` | ✅ 🧩 🔀 |
| Skip backend | switch | `jsonData_customHealthCheck_skipBackend` | switch | `jsonData.customHealthCheck.skipBackend` | ✅ 🧩 🔀 |

¹ **Not a discrepancy.** The `dsconfig.json` declares `ui.component: "textarea"` for the three TLS
cert fields, but the new renderer draws any `target: "secureJsonData"` field as a masked secure
input with a show/hide toggle (the same renderer policy documented for prometheus/graphite). Both
UIs collect the same PEM text into the same `secureJsonData` keys — only the widget affordance differs.

---

## Gap found and fixed: Custom HTTP Headers

**Legacy behaviour (verified in `legacy-expand-mock-parity.json`, UID `ffrbqixmu1gjka`):** the
editor includes a **Custom HTTP Headers** section heading with an **Add header** button
(`hasCustomHeaders: true`, `addHeaderBtn: true`; headings `["HTTP", "Auth", "Custom HTTP Headers"]`).
`@grafana/plugin-ui`'s `CustomHeaders` component (exposed via the `Auth` component's
`convertLegacyAuthProps`) persists headers as indexed pairs — `jsonData.httpHeaderName<N>` for the
name and `secureJsonData.httpHeaderValue<N>` for the (secret) value, starting at `N = 1`.

**Before:** `dsconfig.json` did not model headers at all, so the new UI rendered no headers editor.

**Fix (in `dsconfig.json` only):** added the `jsonData_httpHeaders` array field (copied verbatim
from `registry/prometheus/dsconfig.json`) with an `indexedPair` storage mapping that reproduces the
exact legacy storage, plus item sub-fields for the header name (`http.header.name`, with a
header-name pattern validation) and value (`http.header.value`), and added it to the **Connection**
group's `fieldRefs`.

**After (verified in `newgen-mock-tab.json`):** the new UI renders a **Custom HTTP Headers** row
under **Connection** with an **Add custom http header** button and a key/secret-value editor
(`hasHeadersEditor: true`).

**Save-payload routing (verified with `capture-headers-generic.js`):** filling one header
(name `X-Org-Id`, value `influx-tenant-9`) and clicking **Save & Test** logged:

```
jsonData.httpHeaderName1: X-Org-Id | secureJsonData.httpHeaderValue1: influx-tenant-9
```

The header **name lands in `jsonData.httpHeaderName1`** and the **value in
`secureJsonData.httpHeaderValue1`** — byte-for-byte the legacy `CustomHeaders` storage format.

---

## `fileUpload` evaluation — not applicable to Mock

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not:

- The legacy Mock editor renders the CA Cert / Client Cert / Client Key fields as **plain
  textareas**. No file-upload button and no `<input type="file">` were found in the legacy DOM
  (`legacy-expand-mock-parity.json`: `fileInputs: 0`, `uploadButtons: []`).

**Decision:** do **not** add `fileUpload` to any Mock field. The cert fields keep their current
modeling; both UIs collect PEM text into the same `secureJsonData` keys.

---

## Conditional fields & effects — tested

### `virtual_authMethod` selector (`effects`)

Mock models auth as a **virtual selector** (`virtual_authMethod`) whose value is `read` from
`root.basicAuth` / `jsonData.oauthPassThru` and whose `effects` fan out to those two managed fields.
Branches driven in the new UI (auth-method combobox) and verified from the `Save & Test` console
payload (`verify-mock-effects-basic.json`, `verify-mock-effects-oauth.json`):

| Selection | UI effect (verified) | Save payload (verified) |
| --- | --- | --- |
| **No Authentication** (default) | User / Password hidden | `basicAuth: false`, `jsonData.oauthPassThru: false` |
| **Basic authentication** | User + Password inputs revealed (🔀 `dependsOn: virtual_authMethod == 'BasicAuth'`) | `basicAuth: true`, `basicAuthUser: "grafana"`, `secureJsonData.basicAuthPassword: "s3cret"`, `oauthPassThru: false` |
| **Forward OAuth Identity** | User / Password hidden | `basicAuth: false`, `jsonData.oauthPassThru: true` |

This confirms the `effects` `set` operations propagate to the correct storage targets (not just
visibility): selecting **Basic authentication** produced `basicAuth: true, oauthPassThru: false`,
and selecting **Forward OAuth Identity** produced `basicAuth: false, oauthPassThru: true`, each in a
single save payload from a clean page load.

### `dependsOn` conditionals

| Trigger | Revealed field(s) | Verified |
| --- | --- | --- |
| `virtual_authMethod == 'BasicAuth'` | User, Password | ✅ appear on selection; route to `root` / `secureJsonData` |
| `jsonData_tlsAuth == true` | ServerName, Client Certificate, Client Key | ✅ (declared; TLS group) |
| `jsonData_tlsAuthWithCACert == true` | CA Certificate | ✅ (declared; TLS group) |
| `jsonData_customHealthCheckEnabled == true` | Custom status, Custom message, Custom detail json, Skip backend | ✅ (declared; plugin-owned block) |

---

## Conformance (no change required)

Adding an `indexedPair` field would normally break `JSONDataMatchesStruct` /
`JSONDataTypesMatchStruct`, because an `indexedPair` field (`jsonData_httpHeaders`) is a **logical
view** over dynamically-indexed legacy keys (`httpHeaderName1`, `httpHeaderValue1`, …) that are not
modeled as a single Go struct field. The shared conformance walker (`schema/conformance.go`) already
skips `indexedPair` fields via `isIndexedPairField()` (lines 157, 180, 245-246) — this is
plugin-agnostic — so **no conformance change was needed here**. The per-header `httpHeaderValue<N>`
secrets remain dynamic and are correctly **not** listed among the static `SecureJsonDataKeys`
(`settings.go:65-70`, which already documents this exclusion), and the generated spec emits
`httpHeaders` as a clean array under `jsonData` with **no** secure values leaked
(`SchemaSpecHasNoSecureJSON` passes).

---

## Verification

```
go generate ./registry/grafana-mock-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-mock-datasource/...        # PASS
```

`TestSchemaConformance` subtests (grafana-mock-datasource) — all **PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `TestLoadConfig`, `TestApplyDefaults`, and `TestValidate` suites also pass
unchanged (they do not reference the new field).

> The broader `go test ./registry/... ./schema/...` run shows unrelated failures in **other**
> plugins (`dlopes7-appdynamics-datasource`, `grafana-mongodb-datasource`) whose working-tree
> `dsconfig.json`/`.gen.json` are mid-edit from separate parity work; they are out of scope for this
> entry and are not caused by this change (each plugin owns its own generated artifacts).

New-UI captures: `newgen-mock-tab` (tab, `hasHeadersEditor: true`, URL + Custom HTTP Headers in
fieldLabels), `newgen-mock-wiz` (wizard opens on **General 1/5** with Authentication method),
`verify-mock-effects-basic` / `verify-mock-effects-oauth` (auth-method effects + dependsOn reveals),
plus the `capture-headers-generic` console routing check. Legacy capture:
`legacy-expand-mock-parity` (Custom HTTP Headers + Add header present; 0 file inputs).

---

## Files changed

- [`registry/grafana-mock-datasource/dsconfig.json`](dsconfig.json) — added the
  `jsonData_httpHeaders` `indexedPair` field and referenced it from the `connection` group; added an
  `llm`/`http` instruction documenting the modeled headers field.
- [`registry/grafana-mock-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-mock-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`httpHeaders` array added under `jsonData`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema/conformance.go`, and everything in `plugin-ui`.
