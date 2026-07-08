# ServiceNow — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-servicenow-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/bfrbqiy8onhfkb` (Grafana Enterprise; the ServiceNow plugin's own `ConfigEditor.tsx`, which renders the **ServiceNow Instance Settings** block over `@grafana/ui` `DataSourceHttpSettings` — root `url` + Basic-auth `basicAuthUser`/`basicAuthPassword` — plus the `authMethod` radio, the OAuth fields, and a `CustomHeadersSettings` **Custom HTTP Headers** section)
- **New UI:** `http://192.168.1.241:58899/iframe.html?args=pluginType:grafana-servicenow-datasource&id=configeditor-datasourceconfigwizard--tab` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + `Save & Test`-button console payloads). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/grafana-servicenow-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the screenshots reflect the local schema without pushing.
- **Result:** **Parity achieved.** One missing field was found (**Custom HTTP Headers**) and added; `fileUpload` was evaluated and correctly **not** used (no cert fields / file inputs in the legacy editor); the unconditionally-required root fields were corrected to `required: true`, and the `authMethod` discriminator's `dependsOn` conditionals were exercised and their storage targets verified from the save payload.

---

## TL;DR of changes

| #   | Change | File | Why |
| --- | --- | --- | --- |
| 1   | Changed `root_url`, `root_basicAuthUser`, and `secureJsonData_basicAuthPassword` from `requiredWhen: "true"` → `required: true` | [`dsconfig.json`](dsconfig.json) | All three are **unconditionally** required (the plugin's `IsValid`/`Validate` reject an empty URL, username, or password; ServiceNow always uses Basic-auth credentials, even for OAuth). Puts all three into the wizard's synthetic **General** step and emits `required: ["url","basicAuthUser"]` + secure `required: true` in the generated spec. |
| 2   | Added **Custom HTTP Headers** field (`jsonData_httpHeaders`) with `indexedPair` storage | [`dsconfig.json`](dsconfig.json) | Legacy editor renders a **Custom HTTP Headers** section with **Add header**; new UI had no headers editor |
| 3   | Added `jsonData_httpHeaders` to the `additional-settings` group's `fieldRefs` (after `jsonData_queryTimeoutSeconds`) | [`dsconfig.json`](dsconfig.json) | Surface the new field under **Additional settings** |
| 4   | Updated the `connection` instruction: headers are now **modeled** (was "supported by the editor but … not modeled as first-class fields") | [`dsconfig.json`](dsconfig.json) | Keep the embedded `llm`-tagged instructions truthful after change #2 |
| 5   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-servicenow-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`) |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance.go`, or `plugin-ui`. All schema changes flow through `dsconfig.json`; the rest is produced by `go generate`.

**No conformance-test or plugin-ui change was required** (see [Conformance](#conformance-no-change-required) and [Wizard mode](#wizard-mode-url--credentials-in-the-general-step) below) — the `indexedPair` conformance skip and the auth-group generalisation were already in place from the graphite work.

---

## Section layout

The `groups` were left in their existing taxonomy; the new field was slotted into
**Additional settings**. Verified rendering top-to-bottom in the new UI sidebar/accordion
(`newgen-servicenow-tab.json`):

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Connection** (`connection`) | no | URL |
| 2 | **Authentication** (`authentication`) | no | Authentication Type (radio) → Username → Password → Client ID → Client Secret |
| 3 | **Additional settings** (`additional-settings`) | yes | Use Sys Tables?, Query Timeout, **Custom HTTP Headers** ➕ |

Notes:

- The **Custom HTTP Headers** field is placed in **Additional settings** (alongside Use Sys
  Tables? + Query Timeout), consistent with where the other non-connection/non-auth knobs live.
- `jsonData_oauthEnabled` is a **legacy backend-only** field (tagged `legacy`, no `ui` block) that
  belongs to no group — it is a deprecated predecessor of `authMethod` read only for backwards
  compatibility, and neither UI renders it. Parity preserved.
- The `additional-settings` group is `optional` and rendered collapsed/collapsible in tab mode
  (expanding it surfaces the **Add custom http header** button in the DOM).

### Wizard mode: URL + credentials in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the
fields of the auth group, plus their `dependsOn` parents/children.

- **Before:** `root_url`, `root_basicAuthUser`, and `secureJsonData_basicAuthPassword` used
  `requiredWhen: "true"` (a CEL expression the resolver does **not** inspect), so none of them were
  folded into General.
- **After:** changing all three to `required: true` (they are unconditionally required — the plugin
  rejects an empty URL/username/password) puts URL, Username, and Password into General and emits a
  proper OpenAPI `required: ["url","basicAuthUser"]` + secure `required: true` in the generated spec
  (instead of the `x-dsconfig-required-when: "true"` extension).

**Verified (`verify-servicenow-wizard.json`):** the wizard opens on **General** (step **1/4**)
with `hasUrl: true`, `hasUser: true`, `hasPassword: true`, `hasAuthType: true` — i.e. the
**URL** (`https://<YOUR INSTANCE ID>.service-now.com`), **Username** (`ServiceNow username`),
**Password** (masked, `Password for the ServiceNow account`), and the **Authentication Type**
radio all render on the first step. The auth group folds into General because it uses the
conventional `id: "authentication"`, which the plugin-ui wizard already recognises (both
`authentication` and `auth`) — **no plugin-ui change was needed**.

---

## Field-by-field parity

Legend: ✅ present & matching · ➕ added by this change · 🔀 conditional (revealed by `dependsOn`) · 🎛 auth discriminator · 🔒 backend-only (no editor UI in either)

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| URL | text input (`DataSourceHttpSettings`) | `root_url` | input | `root.url` (required) | ✅ |
| Authentication Type | radio (Basic auth / ServiceNow OAuth) | `jsonData_authMethod` | radio | `jsonData.authMethod` | ✅ 🎛 |
| Username | text input | `root_basicAuthUser` | input | `root.basicAuthUser` (required) | ✅ |
| Password | password (secure) | `secureJsonData_basicAuthPassword` | secure input | `secureJsonData.basicAuthPassword` (required) | ✅ |
| Client ID | text input | `jsonData_oauthClientID` | input | `jsonData.oauthClientID` | ✅ 🔀 |
| Client Secret | password (secure) | `secureJsonData_oauthClientSecret` | secure input | `secureJsonData.oauthClientSecret` | ✅ 🔀 |
| Use Sys Tables? | switch | `jsonData_useSysTables` | switch | `jsonData.useSysTables` | ✅ |
| Query Timeout | number | `jsonData_queryTimeoutSeconds` | number | `jsonData.queryTimeoutSeconds` | ✅ |
| **Custom HTTP Headers** | Add header → name input + value password | `jsonData_httpHeaders` | IndexedPair editor (key + secret value) | `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>` | ➕ 🔀 |
| — (legacy, deprecated) | — | `jsonData_oauthEnabled` | — | `jsonData.oauthEnabled` | ✅ 🔒 |

Notes:

- **Username / Password are unconditionally required** (not `dependsOn`-gated). ServiceNow always
  authenticates with the account username/password: Basic auth uses them directly, and ServiceNow
  OAuth (an OAuth2 resource-owner password grant) uses the same username/password **plus** the OAuth
  Client ID/Secret. Verified: they render for both auth methods (`verify-servicenow-oauth-cond.json`).
- **Client ID / Client Secret** are conditional on `jsonData_authMethod == 'serviceNowOAuth'` — they
  keep their genuine `requiredWhen: "jsonData_authMethod == 'serviceNowOAuth'"` (correctly **not**
  converted to `required: true`).
- The **Password / Client Secret** secure fields render as masked secure inputs with a show/hide
  toggle (the renderer draws any `target: "secureJsonData"` field that way) — same widget policy
  documented for graphite/prometheus.

---

## `fileUpload` evaluation — not applicable to ServiceNow

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not:

- The legacy ServiceNow editor exposes only URL + Basic-auth credentials + OAuth fields + custom
  headers. It has **no** TLS cert / key fields and **no** file pickers — the legacy DOM capture
  (`legacy-expand-ent-grafana-servicenow-datasource.json` / `legacy-expand-servicenow-verify.json`)
  reports `fileInputs: 0`, `uploadButtons: []`.
- The new UI's `fileUpload` component only activates for `ui.fileMapping` (multi-key JSON
  distribution, e.g. a GCP service-account file); nothing in ServiceNow needs it.

**Decision:** do **not** add `fileUpload` to any ServiceNow field.

---

## Gap found and fixed: Custom HTTP Headers

**Legacy behaviour (verified):** the editor includes a **Custom HTTP Headers** section heading with
an **Add header** button (`hasCustomHeaders: true`, `addHeaderBtn: true`; headings observed:
`["ServiceNow Instance Settings", "Custom HTTP Headers"]`). `@grafana/plugin-ui`'s
`CustomHeadersSettings` (`ConfigEditor.tsx:256`) persists headers as indexed pairs —
`jsonData.httpHeaderName<N>` for the name and `secureJsonData.httpHeaderValue<N>` for the (secret)
value, starting at `N = 1`.

**Before:** `dsconfig.json` did not model headers at all (the old `connection` instruction / README
explicitly excluded them), so the new UI rendered no headers editor.

**Fix (in `dsconfig.json` only):** added the `jsonData_httpHeaders` array field (copied verbatim
from the prometheus entry) with an `indexedPair` storage mapping that reproduces the exact legacy
storage, plus item sub-fields for the header name (`http.header.name`, with a header-name pattern
validation) and value (`http.header.value`), and added it to the **Additional settings** group's
`fieldRefs`.

**After (verified in `newgen-servicenow-tab.json`):** the new UI renders a **Custom HTTP Headers**
row under **Additional settings** with an **Add custom http header** button and a key/secret-value
editor (`hasHeadersEditor: true`).

---

## Conditional fields & auth discriminator — tested

ServiceNow models auth as a **discriminator radio** (`jsonData_authMethod`, default `basicAuth`),
not a virtual selector with `effects` and not a set of direct toggles. There are **no** `effects`
blocks. The `dependsOn` conditionals were exercised in the new UI (tab mode) —
evidence in `verify-servicenow-oauth-cond.json`:

| Auth Type selection | Revealed field(s) | Verified |
| --- | --- | --- |
| **Basic auth** (default) | Username, Password (always) | ✅ `user: true, password: true`; Client ID / Secret hidden |
| **ServiceNow OAuth** | Username, Password (still) **+** Client ID, Client Secret | ✅ `clientId: true, clientSecret: true` appear; user/password remain |

This confirms the correct model: Username/Password are always present (unconditionally required),
and the OAuth application credentials are added only for `serviceNowOAuth`.

### Save-payload storage-target validation

Filling **all** required fields (URL + Username + Password) plus one custom header (name `X-Org-Id`,
value `snow-tenant-7`) and clicking **Save & Test** logs the exact datasource payload the wizard
would PUT (`verify-servicenow-headers-routing.json`):

```json
{
  "url": "https://dev12345.service-now.com",
  "basicAuthUser": "snow-admin",
  "jsonData": {
    "authMethod": "basicAuth",
    "useSysTables": false,
    "queryTimeoutSeconds": 30,
    "httpHeaderName1": "X-Org-Id"
  },
  "secureJsonData": {
    "basicAuthPassword": "snow-pass",
    "httpHeaderValue1": "snow-tenant-7"
  },
  "secureJsonFields": { "basicAuthPassword": false, "httpHeaderValue1": false }
}
```

The header **name lands in `jsonData.httpHeaderName1`** and the **value in
`secureJsonData.httpHeaderValue1`** (with `secureJsonFields.httpHeaderValue1: false`) — byte-for-byte
the legacy `CustomHeadersSettings` storage format. **URL → `root.url`**, **Username →
`root.basicAuthUser`**, **Password → `secureJsonData.basicAuthPassword`**; all other fields route to
`jsonData` exactly as declared.

**Required-field enforcement (verified):** an earlier attempt that filled only the URL + a header
(leaving Username/Password empty) produced **no** save payload (`payloadCount: 0`) — the wizard
blocks Save & Test until the three `required: true` fields are set, exactly the intended contract.

---

## Conformance (no change required)

Adding an `indexedPair` field would normally break `JSONDataMatchesStruct` /
`JSONDataTypesMatchStruct`, because an `indexedPair` field (`jsonData_httpHeaders`) is a **logical
view** over dynamically-indexed legacy keys (`httpHeaderName1`, `httpHeaderValue1`, …) that are not
modeled as a single Go struct field. The shared conformance walker (`schema/conformance.go`) already
skips `indexedPair` fields via `isIndexedPairField()` (lines 157, 180, 245-247) — added during the
graphite work and plugin-agnostic — so **no conformance change was needed here**. The per-header
`httpHeaderValue<N>` secrets remain dynamic and are correctly **not** listed among the static
`SecureJsonDataKeys` (`settings.go:78-81`, which are `basicAuthPassword` + `oauthClientSecret`), and
the generated spec emits `httpHeaders` as a clean array under `jsonData` with **no** secure values
leaked (`SchemaSpecHasNoSecureJSON` and `SecureValuesMatchLoadSettings` pass; `secureValues` stays
limited to the two static secrets).

---

## Verification

```
go generate ./registry/grafana-servicenow-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-servicenow-datasource/...       # PASS
```

`TestSchemaConformance` subtests (ServiceNow) — all **PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `TestLoadConfig`, `TestApplyDefaults`, and `TestValidate` suites also pass
unchanged (they do not reference the new field).

New-UI captures: `newgen-servicenow-tab` (tab, `hasHeadersEditor: true`, `urlPresent: true`,
Additional settings → Custom HTTP Headers), `verify-servicenow-wizard` (wizard opens on **General
1/4** with required URL + Username + Password + Authentication Type), `verify-servicenow-oauth-cond`
(auth-discriminator `dependsOn` reveals Client ID/Secret), `verify-servicenow-headers-routing`
(header save-payload routing + required-field enforcement). Legacy capture:
`legacy-expand-ent-grafana-servicenow-datasource` / `legacy-expand-servicenow-verify` (Custom HTTP
Headers + Add header present; 0 file inputs; 0 upload buttons).

---

## Files changed

- [`registry/grafana-servicenow-datasource/dsconfig.json`](dsconfig.json) — changed `root_url`,
  `root_basicAuthUser`, and `secureJsonData_basicAuthPassword` from `requiredWhen: "true"` to
  `required: true` (they render in the wizard's General step); added the `jsonData_httpHeaders`
  `indexedPair` field and referenced it from the `additional-settings` group; updated the
  `connection` instruction so it states headers are now modeled. The OAuth fields
  (`jsonData_oauthClientID`, `secureJsonData_oauthClientSecret`) keep their genuine conditional
  `requiredWhen: "jsonData_authMethod == 'serviceNowOAuth'"`.
- [`registry/grafana-servicenow-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-servicenow-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`url` + `basicAuthUser` now in the spec's `required` array; `basicAuthPassword`
  secure value now `required: true`; `httpHeaders` array added under `jsonData`;
  `x-dsconfig-required-when: "true"` removed from `url` / `basicAuthUser`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema/conformance.go`, and everything in `plugin-ui`.

> **Out-of-scope note:** `README.md` (and prose in `settings.ts`) still describe custom headers as
> "not modeled as first-class fields". That is now stale after change #2, but those files are
> outside this task's allowed scope (`dsconfig.json` + `ui.parity.md` + `.gen.json` only) and were
> left untouched. `settings.go:75-77` remains accurate — it only states the *dynamic
> `httpHeaderValue<N>` secrets* are not in the static `SecureJsonDataKeys`, which is still true.
