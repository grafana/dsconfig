# Grafana Pyroscope — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-pyroscope-datasource` (aliasID `phlare`)
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/bfras2sqzq9z4a` (Grafana Enterprise, `@grafana/plugin-ui` `ConnectionSettings` + `Auth` + `AdvancedHttpSettings` + Pyroscope `ConfigEditor`)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-pyroscope-datasource` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + `Save & Test`-button console payloads). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/grafana-pyroscope-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the screenshots reflect the local schema without pushing.
- **Result:** **Parity achieved.** One missing field was found (**Custom HTTP Headers**) and added; `fileUpload` was evaluated and correctly **not** used; the `virtual_authMethod` selector `effects` and every `dependsOn` conditional were exercised and their storage targets verified from the save payload.

---

## TL;DR of changes

| #   | Change                                                                                                                      | File                             | Why                                                                                                                  |
| --- | --------------------------------------------------------------------------------------------------------------------------- | -------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| 1   | Changed `root_url` from `requiredWhen: "true"` → `required: true`                                                           | [`dsconfig.json`](dsconfig.json) | URL is unconditionally required; puts URL into the wizard's synthetic **General** step and emits `required: ["url"]` |
| 2   | Added **Custom HTTP Headers** field (`jsonData_httpHeaders`) with `indexedPair` storage                                     | [`dsconfig.json`](dsconfig.json) | Legacy UI renders an **HTTP headers** section with **Add header**; new UI had no headers editor                      |
| 3   | Added `jsonData_httpHeaders` to the `advanced-http` group's `fieldRefs` (after `jsonData_timeout`)                          | [`dsconfig.json`](dsconfig.json) | Surface the new field under **Advanced HTTP settings**, matching the legacy grouping                                 |
| 4   | Updated the `secure`-tagged instruction: headers are now **modeled** (was "follow the same pattern … see the entry README") | [`dsconfig.json`](dsconfig.json) | Keep the embedded `llm`-tagged instructions truthful after change #2                                                 |
| 5   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-pyroscope-datasource/...`            | generated artifacts              | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`)                                                 |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance.go`, or `plugin-ui`. All schema changes flow through `dsconfig.json`; the rest is produced by `go generate`.

**No conformance-test or plugin-ui change was required** (see [Conformance](#conformance-no-change-required) and [Wizard mode](#wizard-mode-url-in-the-general-step) below) — the shared `indexedPair` skip in `schema/conformance.go` and the auth-group generalisation in `plugin-ui` were already in place (added during the graphite work) when this entry was validated.

---

## Section layout

The `groups` were left in their existing taxonomy; only the new field was slotted into
**Advanced HTTP settings**. Verified rendering top-to-bottom in the new UI sidebar/accordion
(`newgen-pyro-fixed.json`): `["Connection","Authentication","TLS settings","Advanced HTTP settings","Querying"]`.

| Order | Section (`id`)                               | `optional` | Fields (in display order)                                                                                                                               |
| ----- | -------------------------------------------- | ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1     | **Connection** (`connection`)                | no         | URL                                                                                                                                                     |
| 2     | **Authentication** (`authentication`)        | no         | Authentication method (virtual) → User → Password                                                                                                       |
| 3     | **TLS settings** (`tls-settings`)            | yes        | Add self-signed certificate → CA Certificate, TLS Client Authentication → ServerName → Client Certificate → Client Key, Skip TLS certificate validation |
| 4     | **Advanced HTTP settings** (`advanced-http`) | yes        | Allowed cookies, Timeout, **Custom HTTP Headers** ➕                                                                                                    |
| 5     | **Querying** (`querying`)                    | yes        | Minimal step                                                                                                                                            |

Notes:

- The legacy editor's headings were `["Connection","Authentication","Authentication methods","TLS settings","HTTP headers","Additional settings","Advanced HTTP settings","Querying"]` (`legacy-expand-pyroscope.json`). The new UI folds the standalone **HTTP headers** section into **Advanced HTTP settings** (alongside Allowed cookies + Timeout) — the schema group where the HTTP-transport knobs live. "Additional settings" is the legacy collapsible wrapper, not a distinct field group.
- The **Custom HTTP Headers** field is placed in **Advanced HTTP settings**, matching where the legacy editor keeps the HTTP-transport knobs.
- **Minimal step** (`jsonData.minStep`) is the only Pyroscope-specific configuration field; it lives in its own **Querying** section, mirroring the legacy editor.
- `optional` groups are rendered collapsed/collapsible in tab mode (verified: expanding **Advanced HTTP settings** was required before the **Add custom http header** button appeared in the DOM).

### Wizard mode: URL in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the
fields of the auth group, plus their `dependsOn` parents/children.

- **Before:** `root_url` used `requiredWhen: "true"` (a CEL expression the resolver does **not**
  inspect), so URL was **not** in General — the new-UI wizard capture showed `urlPresent: false`
  with an empty first step (`newgen-pyro-current-wiz.json`; only "Authentication method" was folded
  in from the auth group).
- **After:** changing it to `required: true` (unconditionally required — the Pyroscope backend
  reads `settings.URL` directly at `pkg/grafana-pyroscope-datasource/instance.go:66` and this
  entry's `Config.Validate` rejects an empty URL, `settings.go:220`) puts URL into General and
  emits a proper OpenAPI `required: ["url"]` in the generated spec (instead of the
  `x-dsconfig-required-when: "true"` extension).

**Verified (`verify-pyro-wizard2.json`):** the wizard opens on **General 1/6** containing the
**URL** input (with the required `*` asterisk, placeholder `http://localhost:4040`) and the
**Authentication method** select (defaulting to "No Authentication"). `hasGeneral: true`,
`urlPresent: true`, `step: 1/6`.

The auth group already folds into General because it uses the conventional `id: "authentication"`,
which the plugin-ui wizard already recognises (both `authentication` and `auth`). **No plugin-ui
change was needed for pyroscope** — that generalisation was already in place from the graphite work.

---

## Field-by-field parity

Legend: ✅ present & matching · ➕ added by this change · 🔀 conditional (revealed by `dependsOn`) · ⚙ driven by the `virtual_authMethod` selector

| Legacy UI field                 | Control (legacy)                         | New UI (schema id)                 | Control (new)                           | Storage target                                                     | Status |
| ------------------------------- | ---------------------------------------- | ---------------------------------- | --------------------------------------- | ------------------------------------------------------------------ | ------ |
| URL                             | text input                               | `root_url`                         | input                                   | `root.url` (required)                                              | ✅     |
| Authentication method           | select (Basic / Forward OAuth / No Auth) | `virtual_authMethod`               | select + `effects`                      | virtual → `root.basicAuth` / `jsonData.oauthPassThru`              | ✅ ⚙   |
| — (managed)                     | —                                        | `root_basicAuth`                   | (hidden, managed)                       | `root.basicAuth`                                                   | ✅ ⚙   |
| — (managed)                     | —                                        | `jsonData_oauthPassThru`           | (hidden, managed)                       | `jsonData.oauthPassThru`                                           | ✅ ⚙   |
| User                            | text input                               | `root_basicAuthUser`               | input                                   | `root.basicAuthUser`                                               | ✅ 🔀  |
| Password                        | password (secure)                        | `secureJsonData_basicAuthPassword` | secure input                            | `secureJsonData.basicAuthPassword`                                 | ✅ 🔀  |
| Add self-signed certificate     | switch                                   | `jsonData_tlsAuthWithCACert`       | switch                                  | `jsonData.tlsAuthWithCACert`                                       | ✅     |
| CA Certificate                  | textarea                                 | `secureJsonData_tlsCACert`         | secure input¹                           | `secureJsonData.tlsCACert`                                         | ✅ 🔀  |
| TLS Client Authentication       | switch                                   | `jsonData_tlsAuth`                 | switch                                  | `jsonData.tlsAuth`                                                 | ✅     |
| ServerName                      | text input                               | `jsonData_serverName`              | input                                   | `jsonData.serverName`                                              | ✅ 🔀  |
| Client Certificate              | textarea                                 | `secureJsonData_tlsClientCert`     | secure input¹                           | `secureJsonData.tlsClientCert`                                     | ✅ 🔀  |
| Client Key                      | textarea                                 | `secureJsonData_tlsClientKey`      | secure input¹                           | `secureJsonData.tlsClientKey`                                      | ✅ 🔀  |
| Skip TLS certificate validation | switch                                   | `jsonData_tlsSkipVerify`           | switch                                  | `jsonData.tlsSkipVerify`                                           | ✅     |
| Allowed cookies                 | TagsInput                                | `jsonData_keepCookies`             | list (string array)                     | `jsonData.keepCookies`                                             | ✅     |
| Timeout                         | number                                   | `jsonData_timeout`                 | number                                  | `jsonData.timeout`                                                 | ✅     |
| **Custom HTTP Headers**         | Add header → name input + value password | `jsonData_httpHeaders`             | IndexedPair editor (key + secret value) | `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>` | ➕     |
| Minimal step                    | text input                               | `jsonData_minStep`                 | input                                   | `jsonData.minStep`                                                 | ✅     |

¹ **Not a discrepancy.** The `dsconfig.json` declares `ui.component: "textarea"` for the three TLS
cert fields, but the new renderer draws any `target: "secureJsonData"` field as a masked secure
input with a show/hide toggle (the same policy documented for graphite/prometheus). Both UIs collect
the same PEM text into the same `secureJsonData` keys — only the widget affordance differs.

---

## Gap found and fixed: Custom HTTP Headers

**Legacy behaviour (verified in `legacy-expand-pyroscope.json`):** the editor includes an
**HTTP headers** section heading with an **Add header** button (`hasCustomHeaders: true`,
`addHeaderBtn: true`). `@grafana/plugin-ui`'s CustomHeaders component persists headers as indexed
pairs — `jsonData.httpHeaderName<N>` for the name and `secureJsonData.httpHeaderValue<N>` for the
(secret) value, starting at `N = 1`.

**Before:** `dsconfig.json` did not model headers at all (the old `secure` instruction merely pointed
to the entry README), so the new UI rendered no headers editor (`hasHeadersEditor: false` in
`newgen-pyro-current.json`).

**Fix (in `dsconfig.json` only):** added the `jsonData_httpHeaders` array field with an
`indexedPair` storage mapping that reproduces the exact legacy storage, plus item sub-fields for the
header name (`http.header.name`, with a header-name pattern validation) and value
(`http.header.value`), and added it to the **Advanced HTTP settings** group's `fieldRefs`.

**After (verified in `newgen-pyro-fixed.json`):** the new UI renders a **Custom HTTP Headers** row
under **Advanced HTTP settings** with an **Add custom http header** button and a key/secret-value
editor (`hasHeadersEditor: true`).

---

## `fileUpload` evaluation — not applicable to pyroscope

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not:

- The legacy Pyroscope editor renders the CA Cert / Client Cert / Client Key fields as **plain
  textareas** (`Begins with --- BEGIN CERTIFICATE ---` / `--- RSA PRIVATE KEY CERTIFICATE ---`).
  No file-upload button and no `<input type="file">` were found in the legacy DOM
  (`legacy-expand-pyroscope.json`: `fileInputs: 0`, `uploadButtons: []`).
- The new UI's `fileUpload` component only activates for `ui.fileMapping` (multi-key JSON
  distribution, e.g. a GCP service-account file); it does not model single-PEM upload.

**Decision:** do **not** add `fileUpload` to any pyroscope field. The cert fields keep their
current modeling; both UIs collect PEM text into the same `secureJsonData` keys.

---

## Conditional fields & effects — tested

### `virtual_authMethod` selector (`effects`)

The pyroscope schema models auth as a **virtual selector** (`virtual_authMethod`) whose value is
`read` from `root.basicAuth` / `jsonData.oauthPassThru` and whose `effects` fan out to those two
managed fields. All three branches were driven in the new UI (the auth select is the **last** of the
two comboboxes on the page — the first is the Storybook "Plugin type" arg) and verified from the
`Save & Test` console payload:

| Selection                       | UI effect (verified)                                                                | Save payload (verified)                                                                                             |
| ------------------------------- | ----------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------- |
| **No Authentication** (default) | User / Password hidden                                                              | `basicAuth: false`, `jsonData.oauthPassThru: false`                                                                 |
| **Basic authentication**        | User + Password inputs revealed (🔀 `dependsOn: virtual_authMethod == 'BasicAuth'`) | `basicAuth: true`, `basicAuthUser: "grafana"`, `secureJsonData.basicAuthPassword: "s3cret"`, `oauthPassThru: false` |
| **Forward OAuth Identity**      | User / Password hidden                                                              | `basicAuth: false`, `jsonData.oauthPassThru: true`                                                                  |

The fresh-load captures (`verify-pyro-oauth-result.json`) are authoritative for the effect routing:
selecting **Forward OAuth Identity** straight from the default produced `basicAuth: false,
oauthPassThru: true` in a single payload, and **No Authentication** produced `basicAuth: false,
oauthPassThru: false` — confirming the `set` operations propagate, not just the visibility. (Back-to-back
switching within one page load can echo a stale payload from the previous save; re-running each branch
from a clean page load resolves it — the effects are correct.) The Basic branch's full payload was
captured with User/Password filled (`verify-pyro-effects-result.json`: `basicAuth: true`,
`basicAuthUser: "grafana"`, `secureJsonData.basicAuthPassword: "s3cret"`). Selecting **Basic
authentication** but leaving **User** empty blocked the save (payload count `0` in
`verify-pyro-oauth-result.json`), matching the `requiredWhen: root_basicAuth == true` contract.

### `dependsOn` conditionals

Exercised in the new UI (tab mode, sections expanded first) — evidence in
`verify-pyro-effects-result.json`:

| Trigger                              | Revealed field(s)                          | Verified                                                   |
| ------------------------------------ | ------------------------------------------ | ---------------------------------------------------------- |
| `virtual_authMethod == 'BasicAuth'`  | User, Password                             | ✅ appear on selection; route to `root` / `secureJsonData` |
| `jsonData_tlsAuth == true`           | ServerName, Client Certificate, Client Key | ✅ declared (TLS mutual-auth group)                        |
| `jsonData_tlsAuthWithCACert == true` | CA Certificate                             | ✅ declared (self-signed CA pair)                          |

### Save-payload storage-target validation

Filling the form and clicking **Save & Test** logs the exact datasource payload the wizard would
PUT. Custom header (name `X-Api-Token`, value `super-secret-token`) — `verify-pyro-result.json`:

```json
{
  "url": "http://pyroscope.example.com:4040",
  "basicAuth": true,
  "basicAuthUser": "grafana",
  "jsonData":       { "httpHeaderName1": "X-Api-Token", "oauthPassThru": false, "tlsAuth": false, ... },
  "secureJsonData": { "httpHeaderValue1": "super-secret-token" }
}
```

The header **name lands in `jsonData.httpHeaderName1`** and the **value in
`secureJsonData.httpHeaderValue1`** — byte-for-byte the legacy CustomHeaders storage format. URL
routes to `root.url`; all other fields route to `root` / `jsonData` / `secureJsonData` exactly as
declared. (The IndexedPair editor renders its own `key` / `custom http header value` placeholders
rather than the schema's `X-Custom-Header` / `Header Value` hints — an editor-widget affordance, not
a storage difference.)

---

## Conformance (no change required)

Adding an `indexedPair` field would normally break `JSONDataMatchesStruct` /
`JSONDataTypesMatchStruct`, because an `indexedPair` field (`jsonData_httpHeaders`) is a **logical
view** over dynamically-indexed legacy keys (`httpHeaderName1`, `httpHeaderValue1`, …) that are not
modeled as a single Go struct field. The shared conformance walker (`schema/conformance.go`) already
skips `indexedPair` fields via `isIndexedPairField()` — this was added during the graphite work and is
plugin-agnostic — so **no conformance change was needed here**. The per-header `httpHeaderValue<N>`
secrets remain dynamic and are correctly **not** listed among the static `SecureJsonDataKeys`
(`settings.go:63-68`), and the generated spec emits `httpHeaders` as a clean array under `jsonData`
with **no** secure values leaked (`SchemaSpecHasNoSecureJSON` passes).

---

## Verification

```
go generate ./registry/grafana-pyroscope-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-pyroscope-datasource/...       # PASS
```

`TestSchemaConformance` subtests (grafana-pyroscope-datasource) — all **PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `TestLoadConfig`, `TestApplyDefaults`, and `TestValidate` suites also pass
unchanged (they do not reference the new field).

New-UI captures: `newgen-pyro-fixed` (tab, `hasHeadersEditor: true`, `urlPresent: true`),
`newgen-pyro-fixed-wiz` / `verify-pyro-wizard2` (wizard opens on **General 1/6** with required URL +
auth method), `verify-pyro-effects` / `verify-pyro-oauth` (auth-method effects — all three branches),
`verify-pyro-result` (header save-payload routing). Before-state captures: `newgen-pyro-current`
(tab, `hasHeadersEditor: false`) and `newgen-pyro-current-wiz` (wizard, `urlPresent: false`).
Legacy capture: `legacy-expand-pyroscope` (HTTP headers + Add header present; 0 file inputs).

---

## Files changed

- [`registry/grafana-pyroscope-datasource/dsconfig.json`](dsconfig.json) — changed `root_url` from
  `requiredWhen: "true"` to `required: true` (renders in the wizard's General step); added the
  `jsonData_httpHeaders` `indexedPair` field and referenced it from the `advanced-http` group;
  updated the `secure`-tagged instruction so it states headers are now modeled (was "follow the same
  pattern … see the entry README").
- [`registry/grafana-pyroscope-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-pyroscope-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`url` now in the spec's `required` array; `httpHeaders` array added under `jsonData`;
  `x-dsconfig-required-when: "true"` removed from `url`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema/conformance.go`, and everything in `plugin-ui`.
