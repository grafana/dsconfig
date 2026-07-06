# Amazon Aurora — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-aurora-datasource` (an **AWS-auth SQL** datasource — Aurora Postgres / Aurora MySQL, authenticated with an RDS IAM auth token, so there is **no password field**)
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/efrbd051hnri8d` (Grafana Enterprise 13.0.1; `@grafana/aws-sdk` `ConnectionConfig` + the Aurora plugin's own `ConfigEditor`)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-aurora-datasource` (Storybook, `ConfigEditor/DatasourceConfigWizard`; also `--wizard`)
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + step/section probing). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/grafana-aurora-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the screenshots reflect the local schema without pushing.
- **Result:** **Parity achieved.** All 15 editor-facing legacy fields are present in both UIs and route to identical storage targets. **No missing fields were found** (no field had to be added). Aurora is an AWS-SDK datasource, so it correctly has **no HTTP-headers editor** and **no file-upload** control. The one change required was making the three unconditionally-required fields (`dbUser`, `dbHost`, `dbPort`) use `required: true` so the wizard's synthetic **General** step pulls them in. All AWS auth-type conditionals were exercised and confirmed to reveal/hide the right fields. One pre-existing, out-of-scope note: the new UI renders the backend-only `sessionToken` secret that the legacy editor does not (see _Gaps_).

---

## TL;DR of changes

| #   | Change                                                                                                                         | File                             | Why                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| --- | ------------------------------------------------------------------------------------------------------------------------------ | -------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Changed **`jsonData_dbUser`**, **`jsonData_dbHost`**, **`jsonData_dbPort`** from `"requiredWhen": "true"` → `"required": true` | [`dsconfig.json`](dsconfig.json) | These are unconditionally required (the backend `Validate()` rejects empty `dbUser`/`dbHost` and non-positive `dbPort` — `settings.go:312-320`; the RDS auth-token endpoint is built as `${dbHost}:${dbPort}` and `dbUser` is the DB principal the token impersonates). The wizard's synthetic **General** step only pulls fields with `required: true`; `requiredWhen` is a CEL expression the resolver does not inspect. Also emits proper OpenAPI `required` arrays instead of the `x-dsconfig-required-when` extension. |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-aurora-datasource/...`                  | generated artifacts              | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`)                                                                                                                                                                                                                                                                                                                                                                                                                                                        |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`conformance_test.go`, `schema/conformance.go`, or `plugin-ui`.

The `requiredWhen` conditions that are **real** conditions were **left untouched** — only the three
literal `"requiredWhen": "true"` values were converted. The two conditional
`"requiredWhen": "jsonData_authType == 'keys'"` values on `secureJsonData_accessKey` /
`secureJsonData_secretKey` (the AWS access-key pair, required only when auth = "Access & secret key")
are correct and remain as-is.

No `conformance.go` change was needed (Aurora models no `indexedPair` field). No `plugin-ui`
change was needed: the auth group already uses the conventional id `authentication`, which the
wizard's required-fields resolver recognises, so the auth fields fold into General correctly.

---

## Section layout

Verified rendering top-to-bottom in the new UI (tab mode) and matched to the legacy editor's
section headings. The legacy editor wraps everything under a single `h3` **Connection Details**
with sub-headings (Authentication / Assume Role / Additional Settings / Database Settings /
Advanced); the new UI promotes those five sub-headings to top-level sections.

| Order | Section (`id`)                                                           | `optional` | Fields (in display order)                                                                                        |
| ----- | ------------------------------------------------------------------------ | ---------- | ---------------------------------------------------------------------------------------------------------------- |
| 1     | **Authentication** (`authentication`)                                    | no         | Authentication Provider, [Credentials Profile Name 🔀], [Access Key ID 🔀], [Secret Access Key 🔀], sessionToken |
| 2     | **Assume Role** (`assume-role`)                                          | no         | Assume Role ARN 🔀, External ID 🔀                                                                               |
| 3     | **Additional Settings** (`additional-settings`)                          | no         | Endpoint 🔀, Default Region                                                                                      |
| 4     | **Database Settings** (`database-settings`)                              | no         | Engine, Database Name, **Database User\***, **Database Host\***, **Database Port\***                             |
| 5     | **Advanced: Separate Host and Port for Auth** (`advanced-auth-endpoint`) | yes        | Advanced: DB Host For Auth, Advanced: DB Port For Auth                                                           |

Both UIs use the **same five field groups with identical titles**. In the new UI (tab mode) the
first four render as expanded cards and section 5 as a collapsible **Optional** accordion, matching
the legacy `Advanced` sub-section. `hasHeadersEditor: false` in both tab and wizard (correct).

### Wizard mode: the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) every field
in the auth group (`authentication`), plus their `dependsOn` parents/children.

Confirmed in `verify-aurora-wizard.js` (step header **General 1/6**, fields captured with their
`*` required markers, default `authType = default`):

- **Database User\*** (`jsonData_dbUser`), **Database Host\*** (`jsonData_dbHost`),
  **Database Port\*** (`jsonData_dbPort`) — the three `required: true` fields.
- **Authentication Provider** (`jsonData_authType`) and **sessionToken** — auth-group members.
- **Assume Role ARN**, **External ID**, **Endpoint** — the `dependsOn` fields revealed by the
  default `authType != 'grafana_assume_role'` state.

**Effect of the `required: true` fix:**

| Field                                                      | Before (`requiredWhen: "true"`) | After (`required: true`)  |
| ---------------------------------------------------------- | ------------------------------- | ------------------------- |
| Database User (`jsonData_dbUser`, Database Settings group) | **absent** from General         | **present** in General ✅ |
| Database Host (`jsonData_dbHost`, Database Settings group) | **absent** from General         | **present** in General ✅ |
| Database Port (`jsonData_dbPort`, Database Settings group) | **absent** from General         | **present** in General ✅ |

Before the fix the wizard's General step was missing all three DB fields (they were only reachable
in the `Database Settings` step); after the fix all three appear in General with `*` markers. Tab
mode is unaffected — the synthetic `_required` group is filtered out there, so it still shows the
five sections in order (and the three fields still carry their `*` required markers in the
`Database Settings` card, matching the legacy `Database User *` / `Database Host *` /
`Database Port *`).

---

## Field-by-field parity

Legend: ✅ present & matching · 🔀 conditional (revealed by `dependsOn`) · ⚠️ note

| Legacy UI field            | Control (legacy)                 | New UI (schema id)            | Control (new) | Storage target                | Status                          |
| -------------------------- | -------------------------------- | ----------------------------- | ------------- | ----------------------------- | ------------------------------- |
| Authentication Provider    | select                           | `jsonData_authType`           | select        | `jsonData.authType`           | ✅ (discriminator)              |
| Credentials Profile Name   | text input                       | `jsonData_profile`            | input         | `jsonData.profile`            | ✅ 🔀 (`credentials`)           |
| Access Key ID              | text input                       | `secureJsonData_accessKey`    | secure input  | `secureJsonData.accessKey`    | ✅ 🔀 (`keys`)                  |
| Secret Access Key          | text input                       | `secureJsonData_secretKey`    | secure input  | `secureJsonData.secretKey`    | ✅ 🔀 (`keys`)                  |
| _(not in legacy editor)_   | —                                | `secureJsonData_sessionToken` | secure input  | `secureJsonData.sessionToken` | ⚠️ over-rendered (backend-only) |
| Assume Role ARN            | text input                       | `jsonData_assumeRoleArn`      | input         | `jsonData.assumeRoleArn`      | ✅ 🔀                           |
| External ID                | text input                       | `jsonData_externalId`         | input         | `jsonData.externalId`         | ✅ 🔀                           |
| Endpoint                   | text input                       | `jsonData_endpoint`           | input         | `jsonData.endpoint`           | ✅ 🔀                           |
| Default Region             | select                           | `jsonData_defaultRegion`      | select        | `jsonData.defaultRegion`      | ✅                              |
| Engine                     | select (Aurora Postgres / MySQL) | `jsonData_engine`             | select        | `jsonData.engine`             | ✅                              |
| Database Name              | text input                       | `jsonData_dbName`             | input         | `jsonData.dbName`             | ✅                              |
| Database User \*           | text input                       | `jsonData_dbUser`             | input         | `jsonData.dbUser`             | ✅ **(required fix)**           |
| Database Host \*           | text input                       | `jsonData_dbHost`             | input         | `jsonData.dbHost`             | ✅ **(required fix)**           |
| Database Port \*           | text input                       | `jsonData_dbPort`             | number        | `jsonData.dbPort`             | ✅ **(required fix)**           |
| Advanced: DB Host For Auth | text input                       | `jsonData_dbHostAuth`         | input         | `jsonData.dbHostAuth`         | ✅ (opt)                        |
| Advanced: DB Port For Auth | text input                       | `jsonData_dbPortAuth`         | number        | `jsonData.dbPortAuth`         | ✅ (opt)                        |

All 16 modeled fields render in the new UI and were located across the five sections. The legacy
`Name` and `Default` controls at the top are Grafana editor chrome (datasource name + default
toggle), not part of the datasource config, and are correctly **not** modeled in `dsconfig.json`.
There is no password/secret input beyond the AWS credentials — Aurora authenticates to the DB with
a generated RDS IAM auth token, so both UIs correctly omit any DB password field.

---

## Gaps found

**None beyond the required-field fix.** The complete legacy editor field set is already modeled in
`dsconfig.json`, so no field had to be added. Aurora keeps all fields **inline** (no `packs`),
matching how the legacy AWS/SQL editor lays them out.

### Custom HTTP Headers — not applicable (verified)

Aurora authenticates with AWS SigV4 / RDS IAM and talks the Postgres/MySQL wire protocol — it has
no HTTP-headers concept. Legacy DOM capture (`legacy-expand-aurora-parity`):
`hasCustomHeaders: false`, `addHeaderBtn: false`. New UI (tab + wizard): `hasHeadersEditor: false`.
Correctly **not** added.

### `fileUpload` — not applicable (verified)

AWS credentials are entered as text (Access Key ID / Secret Access Key) or resolved from the
environment / assumed role / credentials profile — there is **no** file-upload control
(`fileInputs: 0`, `uploadButtons: []` in the legacy capture). The new UI's `fileUpload` component
only activates for `ui.fileMapping` fields, which Aurora has none of. Correctly **not** used.

### ⚠️ `sessionToken` over-render — pre-existing, out of scope

The legacy editor does **not** render a `sessionToken` input (it is a provisioning/backend-only
secret — `settings.go:118-122`, read at `pkg/plugin/driver.go:112`, tagged `backend-only` in the
schema). The new UI renders it as a masked secure input because `secureJsonData_sessionToken` is a
member of the `authentication` group and carries no `dependsOn` gate. This is:

- **pre-existing** — not introduced by the required-field fix; and
- a **shared AWS-registry convention** — the sibling AWS SQL datasources model it identically
  (`grafana-athena-datasource` and `grafana-redshift-datasource` both place
  `secureJsonData_sessionToken` in their `authentication` group the same way).

Hiding it would be a schema/renderer decision **beyond the required-field/General-step fix this
task is scoped to** (it would mean removing it from the group's `fieldRefs` or introducing a hide
gate). It is therefore **left unchanged** and documented here. No `plugin-ui` change is required
for the required-field fix itself.

---

## Conditional fields & effects — tested

Aurora drives all its conditionals with a **single `select`** — the **Authentication Provider**
discriminator (`jsonData_authType`). Each scenario was run on a fresh page in tab mode
(`verify-aurora-conditionals.js`; secure fields detected as masked inputs) and confirmed with
full-page screenshots:

| Scenario (`authType`)                             | Credentials Profile | Access / Secret Key    | Assume Role ARN + External ID | Endpoint   | Matches schema `dependsOn`?                        |
| ------------------------------------------------- | ------------------- | ---------------------- | ----------------------------- | ---------- | -------------------------------------------------- |
| **A** `default` (AWS SDK Default)                 | hidden              | hidden                 | **shown**                     | **shown**  | ✅                                                 |
| **B** `keys` (Access & secret key)                | hidden              | **shown (required\*)** | **shown**                     | **shown**  | ✅ key pair revealed + `requiredWhen`              |
| **C** `credentials` (Credentials file)            | **shown**           | hidden                 | **shown**                     | **shown**  | ✅ profile revealed by `authType == 'credentials'` |
| **D** `grafana_assume_role` (Grafana Assume Role) | hidden              | hidden                 | **hidden**                    | **hidden** | ✅ Assume Role section + Endpoint gated off        |

Observed transitions, exactly matching the schema:

- Selecting **`keys`** reveals **Access Key ID** and **Secret Access Key** (both with `*` required
  markers, per `requiredWhen: "jsonData_authType == 'keys'"`) — `dependsOn` = `authType == 'keys'`.
- Selecting **`credentials`** reveals **Credentials Profile Name** — `dependsOn` =
  `authType == 'credentials'`.
- Selecting **`grafana_assume_role`** hides the entire **Assume Role** section
  (`assumeRoleArn`, `externalId`) and the **Endpoint** field — all three carry
  `dependsOn: "jsonData_authType != 'grafana_assume_role'"`, because Grafana Cloud derives those
  itself. This mirrors the plugin's own editor behaviour.
- Throughout every auth-type change, **Database User / Host / Port** keep their `*` required
  markers (unconditionally required, independent of auth type).

The legacy datasource under test is in the default `AWS SDK Default` state, i.e. it matches
scenario **A** — consistent with the new UI (Assume Role ARN / External ID / Endpoint shown, no key
or profile inputs).

**Effects:** Aurora's schema contains **no** `effects` blocks. Its field visibility is a set of
plain `dependsOn` CEL expressions over the single `authType` select; there is no virtual selector
that fans out to write multiple fields, so nothing for `effects` to model, and none were added.

---

## Verification

```
go generate ./registry/grafana-aurora-datasource/...          # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-aurora-datasource/...              # ALL subtests PASS
go test ./schema/... ./registry/grafana-aurora-datasource/... # PASS (no regressions)
```

Conformance subtests (`TestSchemaConformance`): `BaseFieldsResolved`, `SchemaRoundTrip`,
`SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`, `JSONDataMatchesStruct`,
`JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings` — all **PASS** (8/8). The settings
suites `TestLoadConfig`, `TestApplyDefaults`, `TestValidate` also **PASS**.

After regeneration, `schema.gen.json` / `settings.gen.json` moved `dbUser` + `dbHost` + `dbPort`
into the `jsonData` `required` array and dropped the three `x-dsconfig-required-when: "true"`
extensions.

---

## Files changed

- [`registry/grafana-aurora-datasource/dsconfig.json`](dsconfig.json) — changed `jsonData_dbUser`,
  `jsonData_dbHost`, and `jsonData_dbPort` from `"requiredWhen": "true"` to `"required": true`
  (so they render in the wizard's General step and emit OpenAPI `required`). The real `dependsOn` /
  conditional `requiredWhen` conditions (auth-type gating, the `keys` access-key pair) were left
  untouched.
- [`registry/grafana-aurora-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-aurora-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`dbUser`/`dbHost`/`dbPort` now in the spec `required` arrays).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`conformance_test.go`, `schema/conformance.go`, and `plugin-ui`.
