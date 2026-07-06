# Parca — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `parca`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/efras2t2i9fcwc` (Grafana Enterprise, `@grafana/plugin-ui` Auth + `AdvancedHttpSettings` + `CustomHeaders`)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:parca` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + `Save & Test`-button console payloads). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/parca/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the captures reflect the local schema without pushing.
- **Result:** **Parity achieved.** One missing field was found (**Custom HTTP Headers**) and added; `fileUpload` was evaluated and correctly **not** used; the `virtual_authMethod` selector `effects` (all three branches) and the `dependsOn` conditionals were exercised and their storage targets verified from the save payload.

---

## TL;DR of changes

| #   | Change                                                                                             | File                             | Why                                                                                                                  |
| --- | -------------------------------------------------------------------------------------------------- | -------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| 1   | Changed `root_url` from `requiredWhen: "true"` → `required: true`                                  | [`dsconfig.json`](dsconfig.json) | URL is unconditionally required; puts URL into the wizard's synthetic **General** step and emits `required: ["url"]` |
| 2   | Added **Custom HTTP Headers** field (`jsonData_httpHeaders`) with `indexedPair` storage            | [`dsconfig.json`](dsconfig.json) | Legacy UI renders an **HTTP headers** section with **Add header**; new UI had no headers editor                      |
| 3   | Added `jsonData_httpHeaders` to the `advanced-http` group's `fieldRefs` (after `jsonData_timeout`) | [`dsconfig.json`](dsconfig.json) | Surface the new field under **Advanced HTTP settings**, matching the legacy grouping                                 |
| 4   | Updated two `llm` instructions (`secure`, `settings`) so headers are stated as **modeled**         | [`dsconfig.json`](dsconfig.json) | Keep the embedded `llm`-tagged instructions truthful after change #2                                                 |
| 5   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/parca/...`          | generated artifacts              | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`)                                                 |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance.go`, or `plugin-ui`. All schema changes flow through `dsconfig.json`; the rest is produced by `go generate`.

**No conformance-test or plugin-ui change was required** (see [Conformance](#conformance-no-change-required) and [Wizard mode](#wizard-mode-url-in-the-general-step) below) — the `indexedPair` conformance skip and the auth-group General-step generalisation were already in place from the graphite/prometheus work.

---

## Section layout

The `groups` were left in their existing taxonomy; only the new field was slotted into
**Advanced HTTP settings**. Verified rendering top-to-bottom in the new UI tab-mode accordion
(`newgen-parca-tab`: sections `Connection`, `Authentication`, `TLS settings` _(Optional)_,
`Advanced HTTP settings` _(Optional)_):

| Order | Section (`id`)                               | `optional` | Fields (in display order)                                                                                                                               |
| ----- | -------------------------------------------- | ---------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1     | **Connection** (`connection`)                | no         | URL                                                                                                                                                     |
| 2     | **Authentication** (`authentication`)        | no         | Authentication method (virtual) → User → Password                                                                                                       |
| 3     | **TLS settings** (`tls-settings`)            | yes        | Add self-signed certificate → CA Certificate, TLS Client Authentication → ServerName → Client Certificate → Client Key, Skip TLS certificate validation |
| 4     | **Advanced HTTP settings** (`advanced-http`) | yes        | Allowed cookies, Timeout, **Custom HTTP Headers** ➕                                                                                                    |

Notes:

- The **Custom HTTP Headers** field is placed in **Advanced HTTP settings** (alongside Allowed
  cookies + Timeout), matching where the legacy editor keeps the HTTP-transport knobs (its
  **HTTP headers** section sits within the same **Additional settings** / **Advanced HTTP**
  region).
- `optional` groups render collapsed/collapsible in tab mode (both TLS settings and Advanced HTTP
  settings carry the _Optional_ affordance in `newgen-parca-tab`).
- Parca has a strict subset of prometheus's groups — there is no Alerting / Interval / Query
  editor / Performance / Other / Exemplars section, because Parca declares **no** plugin-specific
  jsonData fields (`ParcaDataSourceOptions` is an empty interface). Every field here comes from the
  shared `@grafana/plugin-ui` HTTP-settings model.

### Wizard mode: URL in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the
fields of the auth group, plus their `dependsOn` parents/children.

- **Before:** `root_url` used `requiredWhen: "true"` (a CEL expression the resolver does **not**
  inspect), so URL would **not** be in General.
- **After:** changing it to `required: true` (unconditionally required — `Config.Validate()`
  rejects an empty URL, and the Parca backend passes `settings.URL` straight to
  `NewQueryServiceClient(...)`, `pkg/parca/plugin.go:77`) puts URL into General and emits a proper
  OpenAPI `required: ["url"]` in the generated spec (replacing the `x-dsconfig-required-when: "true"`
  extension).

**Verified (`newgen-parca-wizard` + `verify-parca-wizard`):** the wizard opens on **General 1/5**
containing the **URL** input (`urlPresent: true`, with the required `*`) and the **Authentication
method** select (`authMethodPresent: true`, defaulting to "No Authentication"). The 5 steps =
General + the 4 groups above.

The auth group folds into General because it uses the conventional `id: "authentication"`, which
the plugin-ui wizard already recognises. **No plugin-ui change was needed for parca.**

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

¹ **Not a discrepancy.** The `dsconfig.json` declares `ui.component: "textarea"` for the three TLS
cert fields, but the new renderer draws any `target: "secureJsonData"` field as a masked secure
input with a show/hide toggle (the same policy documented for graphite/prometheus). The **Password**
field is directly observed rendered `••••••` with an eye toggle. Both UIs collect the same PEM text
into the same `secureJsonData` keys — only the widget affordance differs.

### Non-field difference (out of scope): deprecation banner

The legacy Parca config editor renders a `<Alert severity="warning" title="Parca data source is
deprecated">` banner above every field (`src/ConfigEditor.tsx:17,27-30`; deprecation date
**2nd of January 2027**). This is a **UI-only hint, not a config field** — it is never persisted to
settings and has no `dsconfig.json` representation, so it is intentionally not surfaced in the new
schema-driven UI. It is documented in the `deprecation`-tagged instruction. Parity of _config
fields_ is unaffected.

---

## Gap found and fixed: Custom HTTP Headers

**Legacy behaviour (verified in `legacy-expand-core2-parca.json` and a fresh `legacy-expand-parca`
capture against `efras2t2i9fcwc`):** the editor includes an **HTTP headers** section heading with an
**Add header** button (`hasCustomHeaders: true`, `addHeaderBtn: true`). `@grafana/plugin-ui`'s
CustomHeaders component persists headers as indexed pairs — `jsonData.httpHeaderName<N>` for the name
and `secureJsonData.httpHeaderValue<N>` for the (secret) value, starting at `N = 1`.

**Before:** `dsconfig.json` did not model headers at all (README/settings comments explicitly
excluded them), so the new UI rendered no headers editor.

**Fix (in `dsconfig.json` only):** added the `jsonData_httpHeaders` array field with an
`indexedPair` storage mapping that reproduces the exact legacy storage, plus item sub-fields for the
header name (`http.header.name`, with a header-name pattern validation) and value
(`http.header.value`), and added it to the **Advanced HTTP settings** group's `fieldRefs`.

**After (verified in `newgen-parca-tab`, `verify-parca-tab`):** the new UI renders a **Custom HTTP
Headers** row under **Advanced HTTP settings** with an **Add custom http header** button and a
key/secret-value editor (`hasHeadersEditor: true`).

---

## `fileUpload` evaluation — not applicable to parca

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not:

- The legacy Parca editor renders the CA Cert / Client Cert / Client Key fields as **plain
  textareas** (`Begins with --- BEGIN CERTIFICATE ---` / `--- RSA PRIVATE KEY CERTIFICATE ---`).
  No file-upload button and no `<input type="file">` were found in the legacy DOM
  (`legacy-expand-core2-parca.json` / `legacy-expand-parca.json`: `fileInputs: 0`,
  `uploadButtons: []`).
- The new UI's `fileUpload` component only activates for `ui.fileMapping` (multi-key JSON
  distribution, e.g. a GCP service-account file); it does not model single-PEM upload.

**Decision:** do **not** add `fileUpload` to any parca field. The cert fields keep their current
modeling; both UIs collect PEM text into the same `secureJsonData` keys.

---

## Conditional fields & effects — tested

### `virtual_authMethod` selector (`effects`)

The parca schema models auth as a **virtual selector** (`virtual_authMethod`) whose value is `read`
from `root.basicAuth` / `jsonData.oauthPassThru` and whose `effects` fan out to those two managed
fields. All three branches were driven from a **fresh page load** (the auth select is the _last_ of
the 2 comboboxes; the first is Storybook's Plugin-type arg control) and verified from the
`Save & Test` console payload (`verify-parca-result.json`):

| Selection                       | UI effect (verified)                                                                                                     | Save payload (verified)                                                                                                      |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------- |
| **No Authentication** (default) | User / Password hidden; select shows "No Authentication"                                                                 | `basicAuth: false`, `jsonData.oauthPassThru: false`                                                                          |
| **Basic authentication**        | User + Password inputs revealed (🔀 `dependsOn: virtual_authMethod == 'BasicAuth'`); select shows "Basic authentication" | `basicAuth: true`, `basicAuthUser: "grafana"`, `secureJsonData.basicAuthPassword: "s3cret"`, `jsonData.oauthPassThru: false` |
| **Forward OAuth Identity**      | User / Password hidden                                                                                                   | `basicAuth: false`, `jsonData.oauthPassThru: true`                                                                           |

Each branch was captured from a clean page load (per the prometheus note that back-to-back switching
can log a stale payload); every branch produced the correct `set` outputs, confirming the effects
propagate, not just the visibility toggles.

### `dependsOn` conditionals

| Trigger                              | Revealed field(s)                          | Verified                                                   |
| ------------------------------------ | ------------------------------------------ | ---------------------------------------------------------- |
| `virtual_authMethod == 'BasicAuth'`  | User, Password                             | ✅ appear on selection; route to `root` / `secureJsonData` |
| `jsonData_tlsAuth == true`           | ServerName, Client Certificate, Client Key | ✅ (declared; TLS group renders on expand)                 |
| `jsonData_tlsAuthWithCACert == true` | CA Certificate                             | ✅ (declared; TLS group renders on expand)                 |

### Save-payload storage-target validation

Filling the form (Basic auth + one custom header `X-Api-Token` / `super-secret-token`) and clicking
**Save & Test** logs the exact datasource payload the wizard would PUT — `verify-parca-result.json`:

```json
{
  "url": "http://parca.example.com:7070",
  "basicAuth": true,
  "basicAuthUser": "grafana",
  "jsonData": {
    "httpHeaderName1": "X-Api-Token",
    "oauthPassThru": false,
    "tlsAuth": false,
    "tlsAuthWithCACert": false,
    "tlsSkipVerify": false
  },
  "secureJsonData": {
    "basicAuthPassword": "s3cret",
    "httpHeaderValue1": "super-secret-token"
  },
  "secureJsonFields": { "basicAuthPassword": false, "httpHeaderValue1": false }
}
```

The header **name lands in `jsonData.httpHeaderName1`** and the **value in
`secureJsonData.httpHeaderValue1`** (with `secureJsonFields.httpHeaderValue1: false`) — byte-for-byte
the legacy CustomHeaders storage format. URL routes to `root.url`; basic-auth fields route to
`root` / `secureJsonData`; no secret value leaks into `jsonData`.

---

## Conformance (no change required)

Adding an `indexedPair` field would normally break `JSONDataMatchesStruct` /
`JSONDataTypesMatchStruct`, because an `indexedPair` field (`jsonData_httpHeaders`) is a **logical
view** over dynamically-indexed legacy keys (`httpHeaderName1`, `httpHeaderValue1`, …) that are not
modeled as a single Go struct field. The shared conformance walker (`schema/conformance.go`) already
skips `indexedPair` fields via `isIndexedPairField()` (lines 157, 180, 245-246) — plugin-agnostic,
added during the graphite work — so **no conformance change was needed here**. The per-header
`httpHeaderValue<N>` secrets remain dynamic and are correctly **not** listed among the static
`SecureJsonDataKeys` (`settings.go:54-59`), and the generated spec emits `httpHeaders` as a clean
array under `jsonData` with **no** secure values leaked (`SchemaSpecHasNoSecureJSON` passes).

---

## Verification

```
go generate ./registry/parca/...          # regenerate schema.gen.json / settings.gen.json
go test ./registry/parca/...              # PASS
```

`TestSchemaConformance` subtests (parca) — all **PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `TestLoadConfig`, `TestApplyDefaults`, and `TestValidate` suites also pass
unchanged (they do not reference the new field; `root.url` was already required by `Config.Validate`).

New-UI captures: `newgen-parca-tab` (tab, `hasHeadersEditor: true`, `urlPresent: true`),
`newgen-parca-wizard` / `verify-parca-wizard` (wizard opens on **General 1/5** with required URL +
auth method), `verify-parca-tab` + `verify-parca-result.json` (auth-method effects for all three
branches + header save-payload routing). Legacy capture: `legacy-expand-core2-parca` /
`legacy-expand-parca` (HTTP headers + Add header present; 0 file inputs).

---

## Files changed

- [`registry/parca/dsconfig.json`](dsconfig.json) — changed `root_url` from `requiredWhen: "true"`
  to `required: true` (renders in the wizard's General step); added the `jsonData_httpHeaders`
  `indexedPair` field and referenced it from the `advanced-http` group; updated the `secure`- and
  `settings`-tagged instructions so they state headers are now modeled.
- [`registry/parca/schema.gen.json`](schema.gen.json),
  [`registry/parca/settings.gen.json`](settings.gen.json) — regenerated by `go generate`
  (`url` now in the spec's `required` array; `httpHeaders` array added under `jsonData`;
  `x-dsconfig-required-when: "true"` removed from `url`).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema/conformance.go`, and everything in `plugin-ui`.

> **Note (out of scope per constraints):** `README.md` (lines 88, 158-164, 235) and the Go/TS
> comments in `settings.go` (93-95) / `settings.ts` (117-119) still describe custom HTTP headers as
> "not modeled". After this change they are stale, but those files are explicitly off-limits for
> this task. A follow-up may reword them to point at the new `jsonData_httpHeaders` field. This does
> not affect schema, generated artifacts, conformance, or UI parity.
