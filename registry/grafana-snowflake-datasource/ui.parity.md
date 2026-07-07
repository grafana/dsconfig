# Snowflake — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-snowflake-datasource` (an enterprise/community datasource with its **own** React `ConfigEditor`, not `@grafana/sql`; it talks to Snowflake over the Snowflake Go driver, not a generic HTTP client)
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/snowflake-test` (Grafana Enterprise v13.2.0, Snowflake plugin `ConfigEditor`)
- **New UI:** Storybook `ConfigEditor/DatasourceConfigWizard` (`--tab` / `--wizard`, `args=pluginType:grafana-snowflake-datasource`) — **not captured this run (see note below)**
- **Method:** Playwright captured the legacy UI (authenticated as `admin`, full-page screenshot + DOM extraction of headings/labels/inputs + header/file-upload probing). The new UI (Storybook at `http://192.168.1.241:58899`) was **unreachable** (HTTP 000) on the first attempt, so per the run constraints **new-UI capture is deferred (storybook offline)** and was not retried. The authoritative verification is therefore the Go conformance suite (`go test`), corroborated by the legacy capture and a schema read.
- **Result:** **Parity achieved (schema + legacy verified; new-UI wizard rendering deferred).** All 21 modeled fields are present in the legacy editor and route to identical storage targets. **No missing fields, no extra fields.** Snowflake has **no HTTP-headers editor** and **no file-upload** control in the legacy UI (confirmed), so none are modeled — correct. The one change required was converting the unconditionally-required **Account** field from `"requiredWhen": "true"` to `"required": true` so the wizard's synthetic **General** step pulls it in. The four auth-mode conditionals (`authType` = password / keypair / pat / oauth) are modeled as `dependsOn` + `requiredWhen` CEL and were **left untouched**.

> **NOTE — new-UI capture deferred (storybook offline).** The Storybook host was down
> (HTTP 000, single attempt, not retried per run policy). Claims about the new wizard's
> live rendering (the General-step composition in particular) are therefore asserted from
> the schema plus the documented `resolveRequiredFieldsGroup` contract (verified for the
> PostgreSQL entry), **not** from a screenshot captured this run. Re-run
> `capture-new-generic.js grafana-snowflake-datasource <localSchema> wizard snowflake` when
> Storybook is back to attach the visual confirmation.

---

## TL;DR of changes

| #   | Change | File | Why |
| --- | --- | --- | --- |
| 1   | Changed **`jsonData_account`** from `"requiredWhen": "true"` → `"required": true` | [`dsconfig.json`](dsconfig.json) | Account is unconditionally required by the connection contract (the plugin docs/instructions state `jsonData.account` is required; the legacy editor shows it as the first Connection field). The wizard's synthetic **General** step only pulls fields with `required: true`; `requiredWhen` is a CEL expression the resolver does not inspect. The conversion also emits a proper OpenAPI `jsonData.required: ["account"]` array (`convert.go:78-79,96-97`) instead of the `x-dsconfig-required-when: "true"` extension (`convert.go:210-211`). |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-snowflake-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`). `settings.examples.gen.json` was unchanged. |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `schema.go`,
`conformance_test.go`, or `plugin-ui`.

Only the **one literal** `"requiredWhen": "true"` was converted. The **five conditional**
`requiredWhen` expressions — all auth-mode-gated — were **left untouched**:

| Field (id) | Kept `requiredWhen` (unchanged) |
| --- | --- |
| `jsonData_username` | `jsonData_authType != 'oauth'` |
| `secureJsonData_password` | `jsonData_authType == 'password' \|\| jsonData_authType == ''` |
| `secureJsonData_privateKey` | `jsonData_authType == 'keypair'` |
| `secureJsonData_patToken` | `jsonData_authType == 'pat'` |
| `jsonData_oauthPassThru` | `jsonData_authType == 'oauth'` |

> **Backend-validation nuance (differs from PostgreSQL).** For postgres the backend
> `Validate()` hard-rejects an empty URL/user/database. For snowflake the backend
> `Validate()` **does not** require `account` — it only enforces the selected auth method's
> credential (`settings.go:194-226`, and its comment: _"It does not require
> account/username"_). So `required: true` on Account here encodes an **editor / connection-
> contract** requirement (matching the legacy editor and the plugin's own instructions),
> not a backend load failure. This is the intended, documented meaning of `required` for the
> wizard and does not change backend behaviour.

---

## Section layout

Legacy sections captured top-to-bottom (`legacy-expand-snowflake.json`, authenticated).
The legacy `h3` headings are **exactly** the four schema group titles, in order:

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Connection** (`connection`) | no | Account, Region, Authentication Type, Username, Password / Private key / Token / Forward OAuth Identity (auth-gated) |
| 2 | **Connection settings** (`connection-settings`) | yes | Connection settings (repeatable `{name, value, secure}` list) |
| 3 | **Environment** (`environment`) | no | Role, Warehouse, Database, Schema |
| 4 | **Customization** (`customization`) | yes | Min Interval, Row Limit, Connection Timeout (sec), Request Timeout (sec), Variable Interpolation Format, Default Query, Default Variable Query |

Legacy DOM headings captured: `["Connection","Connection settings","Environment","Customization"]` — a 1:1 match to the schema `groups`. The legacy editor was in its **default `password`** auth state, so it showed **Account, Region, Authentication Type, Username, Password** (Connection) and hid Private key / Private key passphrase / Token / Forward OAuth Identity — confirming the auth conditionals behave in the legacy UI exactly as the schema `dependsOn` prescribes.

### Wizard mode: the "General" step (schema-level; live capture deferred)

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the
auth fields, plus their `dependsOn` parents/children. This behaviour was verified live for
the PostgreSQL entry; for snowflake it is inferred from the schema because Storybook was
offline this run.

**Effect of the `required: true` fix (schema-level):**

| Field | Before (`requiredWhen: "true"`) | After (`required: true`) |
| --- | --- | --- |
| Account (`jsonData_account`, Connection group) | **absent** from General — `requiredWhen` is a CEL extension the required-fields resolver does not inspect | **eligible** for General — now a `required: true` field the resolver pulls in (matches how Host URL/Database/Username were fixed for postgres) ✅ *(new-UI visual check deferred)* |

The auth discriminator `jsonData_authType` (role `auth.discriminator`, a **radio** with
password/keypair/pat/oauth) and its gated credential fields carry `role: auth.*` markers and
are expected to fold into General via the wizard's auth resolution. **Caveat:** unlike
postgres (whose auth fields sat in a group literally named `authentication`), snowflake's
auth fields live in the **`connection`** group and are identified by their `auth.*` roles;
the exact General-step membership of the auth credential fields is the part that would
benefit from the deferred new-UI capture. Tab mode is unaffected either way — the synthetic
`_required` group is filtered out there, so tabs show the four sections in order.

---

## Field-by-field parity

Legend: ✅ present & matching · 🔀 conditional (revealed by `dependsOn` on `jsonData_authType`) · ⚙️ frontend-only (`tags: ["frontend-only"]`)

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| Account \* | text input | `jsonData_account` | input | `jsonData.account` | ✅ (**fixed → `required: true`**) |
| Region | text input | `jsonData_region` | input | `jsonData.region` | ✅ (deprecated; prefer region-in-account) |
| Authentication Type | radio (Password / Key Pair / Programmatic Access Token / OAuth) | `jsonData_authType` | radio | `jsonData.authType` | ✅ (discriminator, default `password`) |
| Username | text input | `jsonData_username` | input | `jsonData.username` | ✅ 🔀 (`authType != 'oauth'`) |
| Password | password (`SecretInput`) | `secureJsonData_password` | secure input | `secureJsonData.password` | ✅ 🔀 (`authType == 'password'`) |
| Private key | textarea | `secureJsonData_privateKey` | secure input¹ | `secureJsonData.privateKey` | ✅ 🔀 (`authType == 'keypair'`) |
| Private key passphrase | password | `secureJsonData_privateKeyPassphrase` | secure input | `secureJsonData.privateKeyPassphrase` | ✅ 🔀 (`authType == 'keypair'`, optional) |
| Token | password (`SecretInput`) | `secureJsonData_patToken` | secure input | `secureJsonData.patToken` | ✅ 🔀 (`authType == 'pat'`) |
| Forward OAuth Identity | switch | `jsonData_oauthPassThru` | switch | `jsonData.oauthPassThru` | ✅ 🔀 (`authType == 'oauth'`) |
| Connection settings | repeatable list + `+` | `jsonData_settings` | list (item: name / value / secure) | `jsonData.settings[]` | ✅ (array of `{name,value,secure}`) |
| Role | text input | `jsonData_role` | input | `jsonData.role` | ✅ |
| Warehouse | text input | `jsonData_warehouse` | input | `jsonData.warehouse` | ✅ |
| Database | text input | `jsonData_database` | input | `jsonData.database` | ✅ |
| Schema | text input | `jsonData_schema` | input | `jsonData.schema` | ✅ |
| Min Interval | text input | `jsonData_timeInterval` | input | `jsonData.timeInterval` | ✅ ⚙️ |
| Row Limit | number | `jsonData_rowLimit` | input (number) | `jsonData.rowLimit` | ✅ |
| Connection Timeout (sec) | number | `jsonData_loginTimeout` | input (number) | `jsonData.loginTimeout` | ✅ |
| Request Timeout (sec) | number | `jsonData_requestTimeout` | input (number) | `jsonData.requestTimeout` | ✅ |
| Variable Interpolation Format | select | `jsonData_defaultInterpolation` | select | `jsonData.defaultInterpolation` | ✅ ⚙️ |
| Default Query | textarea | `jsonData_defaultQuery` | textarea | `jsonData.defaultQuery` | ✅ ⚙️ |
| Default Variable Query | textarea | `jsonData_defaultVariableQuery` | textarea | `jsonData.defaultVariableQuery` | ✅ ⚙️ |

All 21 modeled fields (plus the 3 `Connection settings` item fields: `name` / `value` /
`secure`) map to the legacy editor. Legacy label capture in the default password state:
`Account, Region, Authentication Type, Username, Password, Role, Warehouse, Database, Schema,
Min Interval, Row Limit, Connection Timeout (sec), Request Timeout (sec), Variable
Interpolation Format, Default Query, Default Variable Query` (plus the four
Authentication-Type radio option labels). The Grafana editor chrome at the top (datasource
`Name`, `Default` toggle, `Save & test` / `Delete`) is not part of the datasource config and
is correctly **not** modeled.

¹ **Not a discrepancy.** `dsconfig.json` declares `ui.component: "textarea"` for
`secureJsonData_privateKey`, but the new renderer draws any `target === "secureJsonData"`
field as a masked secure input (show/hide toggle) regardless of the declared component — the
same renderer policy documented for the postgres TLS cert fields. Both UIs collect the same
PEM text into `secureJsonData.privateKey`; only the widget affordance differs.

---

## Gaps found

**None beyond the required-field fix.** The complete legacy field set is already modeled in
`dsconfig.json`; no field had to be added or removed. Snowflake keeps all fields **inline**
(no `packs`), matching how the plugin's own `ConfigEditor` lays them out.

### Custom HTTP Headers — not applicable (verified)

The Snowflake plugin connects through the Snowflake Go driver, not a generic HTTP client, so
it has no custom-HTTP-headers concept. Authenticated legacy DOM capture
(`legacy-expand-snowflake.json`): `hasCustomHeaders: false`, `addHeaderBtn: false`. Correctly
**not** modeled — matching the run scope ("no HTTP headers … legacy confirmed none").

---

## `fileUpload` evaluation — not applicable to snowflake

The `fileUpload` control is used only when the legacy UI offers file upload. It does not:

- The legacy Snowflake editor renders the RSA **Private key** as a plain **textarea** (inline
  PEM paste), not a file picker. Authenticated legacy capture: `fileInputs: 0`,
  `uploadButtons: []` (no `<input type="file">`, no upload button anywhere in the form).
- The new UI's `fileUpload` component only activates when a field declares `ui.fileMapping`
  (multi-key JSON distribution, e.g. a GCP service-account file); snowflake models no such
  field.

**Decision:** do **not** add `fileUpload` to any snowflake field — matching the run scope
("no fileUpload … legacy confirmed none").

---

## Conditional fields & effects — from schema (auth-mode gated)

Snowflake drives its conditionals with a single **radio discriminator**,
`jsonData_authType` (`role: auth.discriminator`, default `password`, allowed values
`password` / `keypair` / `pat` / `oauth`). Each credential field is gated by a `dependsOn`
CEL over `authType`, with a matching `requiredWhen` where the credential is mandatory:

| `authType` | Visible auth fields (besides Account/Region/Role/Warehouse/Database/Schema) | Required credential |
| --- | --- | --- |
| **password** (default) | Username, **Password** | `secureJsonData.password` |
| **keypair** | Username, **Private key**, Private key passphrase (optional) | `secureJsonData.privateKey` (PKCS#8 PEM) |
| **pat** | Username, **Token** | `secureJsonData.patToken` |
| **oauth** | **Forward OAuth Identity** (Username hidden) | `jsonData.oauthPassThru == true` |

Observed / prescribed transitions (legacy confirmed for the default `password` state; the
other three states are schema-prescribed and pending the deferred new-UI capture):

- **`password`** (legacy default, captured): Username + Password shown; Private key / Token /
  Forward OAuth Identity hidden. ✅ matches `dependsOn`.
- **`keypair`**: reveals Private key (+ optional passphrase); hides Password / Token / OAuth.
- **`pat`**: reveals Token; hides Password / Private key / OAuth.
- **`oauth`**: reveals Forward OAuth Identity and **hides Username**
  (`jsonData_username` `dependsOn: authType != 'oauth'`); hides all stored-secret fields.

**Effects:** snowflake's schema contains **no** `effects` blocks. Its auth visibility is a set
of plain `dependsOn` CEL expressions over one discriminator; there is no virtual selector that
fans out to write multiple fields, so nothing for `effects` to model, and none were added.

**Relationships:** the schema declares one `group` relationship binding
`secureJsonData_privateKey` + `secureJsonData_privateKeyPassphrase` (key-pair credentials used
together) — unchanged.

---

## Verification

```
go generate ./registry/grafana-snowflake-datasource/...   # regenerate schema.gen.json / settings.gen.json (settings.examples.gen.json unchanged)
go test ./registry/grafana-snowflake-datasource/...        # PASS (0.418s)
```

`TestSchemaConformance` — all 8 subtests **PASS**: `BaseFieldsResolved`, `SchemaRoundTrip`,
`SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`,
`JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings`.
`TestLoadConfig`, `TestApplyDefaults`, `TestValidate` (settings_test.go) also **PASS** (all
four auth-mode examples covered).

After regeneration, `schema.gen.json` / `settings.gen.json` moved `account` into the
`jsonData` `required: ["account"]` array and dropped the `x-dsconfig-required-when: "true"`
extension.

_Legacy capture note:_ the run-provided UID `afrbqiyap2jggc` did not exist on the legacy
instance; the actual Snowflake datasource UID is **`snowflake-test`** (type
`grafana-snowflake-datasource`, name "Snowflake"), which is what was captured. The default
`capture-legacy-expand.js` landed on Grafana's login screen (unauthenticated Playwright
context, empty headings); a login step (admin/admin) was added to reach the real config form.

---

## Files changed

- [`registry/grafana-snowflake-datasource/dsconfig.json`](dsconfig.json) — changed
  `jsonData_account` from `"requiredWhen": "true"` to `"required": true` (so it renders in the
  wizard's General step and emits OpenAPI `required`). The five conditional auth-mode
  `requiredWhen` expressions were left untouched.
- [`registry/grafana-snowflake-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-snowflake-datasource/settings.gen.json`](settings.gen.json) — regenerated
  by `go generate` (`account` now in the spec `required` arrays; `x-dsconfig-required-when`
  removed).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`,
`settings.examples.gen.json`, `schema.go`, `conformance_test.go`, `settings_test.go`, and
`plugin-ui`.
