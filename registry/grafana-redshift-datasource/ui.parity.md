# Amazon Redshift — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-redshift-datasource` (an **AWS-SDK auth** datasource; composes `@grafana/aws-sdk` `ConnectionConfig` + a Redshift-specific block)
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/bfrbd04qqlhxce` (Grafana Enterprise 13.0.1, Redshift `ConfigEditor` over AWS `ConnectionConfig`)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-redshift-datasource` (Storybook, `ConfigEditor/DatasourceConfigWizard`; also `--wizard`)
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + step/section probing + per-scenario conditional toggling). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/grafana-redshift-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the screenshots reflect the local schema without pushing.
- **Result:** **Parity achieved for the sanctioned fix.** All modeled fields are present in both UIs and route to identical storage targets — **no field had to be added** (Redshift is AWS-SDK auth, so it correctly has **no HTTP-headers editor** and **no file-upload** control). The one change required was making the unconditionally-required **Database** field use `required: true` so the wizard's synthetic **General** step pulls it in. All AWS-auth (`authType`) conditionals and the `useServerless` (Provisioned↔Serverless) conditional were exercised and reveal the correct fields. **Two pre-existing new-UI (plugin-ui) discrepancies were found that are not fixable via `dsconfig.json`** and are reported below (they are independent of this change): the boolean-valued `useManagedSecret` **radio** does not drive its dependent fields, and the `backend-only` `sessionToken` field renders in the editor.

---

## TL;DR of changes

| #   | Change                                                                                                          | File                             | Why                                                                                                                                                                                                                                                                                                                                                                                                                                                                             |
| --- | --------------------------------------------------------------------------------------------------------------- | -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Changed **`jsonData_database`** from `"requiredWhen": "true"` → `"required": true`                              | [`dsconfig.json`](dsconfig.json) | Database is unconditionally required — the backend `Validate()` rejects an empty database for every auth × shape × credential combination (`settings.go:328-330`, `"database is required"`). The wizard's synthetic **General** step only pulls fields with `required: true`; `requiredWhen` is a CEL expression the required-fields resolver does not inspect. Also emits a proper OpenAPI `required: ["database"]` array instead of the `x-dsconfig-required-when` extension. |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-redshift-datasource/...` | generated artifacts              | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`).                                                                                                                                                                                                                                                                                                                                                                                                           |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance.go`, or `plugin-ui`.
The **real** conditional `requiredWhen` values were **left untouched** — only the literal
`"requiredWhen": "true"` on `database` was converted. Every other `requiredWhen` is a genuine
gate and was preserved:

- `secureJsonData_accessKey` / `secureJsonData_secretKey`: `jsonData_authType == 'keys'`
- `jsonData_clusterIdentifier`: `jsonData_useServerless != true`
- `jsonData_workgroupName`: `jsonData_useServerless == true`
- `jsonData_managedSecret_arn`: `jsonData_useManagedSecret == true`
- `jsonData_dbUser`: `jsonData_useServerless != true && jsonData_useManagedSecret != true`

The auth group already uses the conventional id `authentication`, which the wizard's
required-fields resolver recognises, so the AWS auth fields fold into General correctly (no
`plugin-ui` change needed for that).

---

## Section layout

Verified rendering top-to-bottom in the new UI (tab mode) and matched to the legacy editor's
section headings.

| Order | Section (`id`)                                  | Fields (in display order)                                                                                                                                                              |
| ----- | ----------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1     | **Authentication** (`authentication`)           | Authentication Provider, (Access Key ID / Secret Access Key — keys), (Credentials Profile Name — credentials), sessionToken¹                                                           |
| 2     | **Assume Role** (`assume-role`)                 | Assume Role ARN, External ID                                                                                                                                                           |
| 3     | **Additional Settings** (`additional-settings`) | Endpoint, Default Region                                                                                                                                                               |
| 4     | **Redshift Details** (`redshift-details`)       | useManagedSecret (Temporary credentials / AWS Secrets Manager), Serverless, Cluster Identifier / Workgroup, Managed Secret, Database User, Database, Send events to Amazon EventBridge |

The legacy editor renders a top-level **Connection Details** container heading that wraps the
AWS `ConnectionConfig` block (Authentication / Assume Role / Additional Settings), then a
plugin-specific **Redshift Details** section. The new UI renders the four `dsconfig.json`
groups directly (**Authentication**, **Assume Role**, **Additional Settings**, **Redshift
Details**) — the `Connection Details` wrapper title is legacy chrome and is correctly not
modeled as a separate group. Both UIs surface the same fields.

### Wizard mode: the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) every
field in the auth group (`authentication`), plus their `dependsOn` parents/children.

Confirmed in `verify-redshift-wizard.js` (step **General 1/5**):

- **Database\*** (`jsonData_database`) — the `required: true` field, now pulled into General with its `*` marker.
- **Authentication Provider** (`jsonData_authType`) — the AWS auth discriminator (auth-group member).
- **sessionToken**, **Assume Role ARN**, **External ID**, **Endpoint** — folded in as auth-group members / `dependsOn` children of `authType` (the assume-role/endpoint fields carry `jsonData_authType != 'grafana_assume_role'`).

**Effect of the `required: true` fix (before/after):**

| Field                                                  | Before (`requiredWhen: "true"`) | After (`required: true`)             |
| ------------------------------------------------------ | ------------------------------- | ------------------------------------ |
| Database (`jsonData_database`, Redshift Details group) | **absent** from General         | **present** in General ✅ (with `*`) |

Before the fix the wizard's General step was missing **Database** (it was only reachable in the
`Redshift Details` step); after the fix it appears in General. Tab mode is unaffected — the
synthetic `_required` group is filtered out there, so it still shows the four sections in order.

---

## Field-by-field parity

Legend: ✅ present & matching · 🔀 conditional (revealed by `dependsOn`) · ⚠️ new-UI discrepancy (see below)

| Legacy UI field                             | Control (legacy) | New UI (schema id)            | Storage target                | Status                                                         |
| ------------------------------------------- | ---------------- | ----------------------------- | ----------------------------- | -------------------------------------------------------------- |
| Authentication Provider                     | select           | `jsonData_authType`           | `jsonData.authType`           | ✅ (discriminator)                                             |
| Access Key ID                               | text/secure      | `secureJsonData_accessKey`    | `secureJsonData.accessKey`    | ✅ 🔀 (`authType == 'keys'`)                                   |
| Secret Access Key                           | text/secure      | `secureJsonData_secretKey`    | `secureJsonData.secretKey`    | ✅ 🔀 (`authType == 'keys'`)                                   |
| Credentials Profile Name                    | text input       | `jsonData_profile`            | `jsonData.profile`            | ✅ 🔀 (`authType == 'credentials'`)                            |
| Assume Role ARN                             | text input       | `jsonData_assumeRoleArn`      | `jsonData.assumeRoleArn`      | ✅ 🔀 (`!= 'grafana_assume_role'`)                             |
| External ID                                 | text input       | `jsonData_externalId`         | `jsonData.externalId`         | ✅ 🔀                                                          |
| Endpoint                                    | text input       | `jsonData_endpoint`           | `jsonData.endpoint`           | ✅ 🔀                                                          |
| Default Region                              | select           | `jsonData_defaultRegion`      | `jsonData.defaultRegion`      | ✅                                                             |
| Temporary credentials / AWS Secrets Manager | RadioButtonGroup | `jsonData_useManagedSecret`   | `jsonData.useManagedSecret`   | ⚠️ renders, but does not drive dependents (plugin-ui)          |
| Serverless                                  | switch           | `jsonData_useServerless`      | `jsonData.useServerless`      | ✅                                                             |
| Cluster Identifier                          | select           | `jsonData_clusterIdentifier`  | `jsonData.clusterIdentifier`  | ✅ 🔀 (`useServerless != true`)                                |
| Workgroup                                   | select           | `jsonData_workgroupName`      | `jsonData.workgroupName`      | ✅ 🔀 (`useServerless == true`)                                |
| Managed Secret                              | select           | `jsonData_managedSecret_arn`  | `jsonData.managedSecret.arn`  | ⚠️ 🔀 (`useManagedSecret == true` — never revealed, see below) |
| (secret name, written from Select label)    | —                | `jsonData_managedSecret_name` | `jsonData.managedSecret.name` | ✅ (managed by the arn Select)                                 |
| Database User                               | text input       | `jsonData_dbUser`             | `jsonData.dbUser`             | ✅ 🔀 (`useServerless != true \|\| useManagedSecret == true`)  |
| Database \*                                 | text input       | `jsonData_database`           | `jsonData.database`           | ✅ **required (fixed)**                                        |
| Send events to Amazon EventBridge           | switch           | `jsonData_withEvent`          | `jsonData.withEvent`          | ✅                                                             |
| _(none — backend-only)_                     | _(not shown)_    | `secureJsonData_sessionToken` | `secureJsonData.sessionToken` | ⚠️ renders in new UI (legacy hides)                            |

All Redshift Details required markers render in the new UI: **Cluster Identifier\***,
**Database User\***, **Database\*** (tab screenshot `newgen-redshift-tab.png`). The legacy
`Name` and `Default` controls at the top are Grafana editor chrome (datasource name + default
toggle), not part of the datasource config, and are correctly **not** modeled.

---

## Custom HTTP Headers — not applicable (verified)

Redshift is an **AWS-SDK auth** datasource (SigV4 via access keys / assume-role / credentials
file / workspace IAM / Grafana assume-role) and talks the Redshift Data API, **not** arbitrary
HTTP with user headers. Legacy DOM capture (`legacy-expand-p3-redshift.json`,
`legacy-expand-redshift.json`): `hasCustomHeaders: false`, `addHeaderBtn: false`. New UI
(tab + wizard): `hasHeadersEditor: false`. Correctly **not** added.

---

## `fileUpload` evaluation — not used

AWS credentials are entered as text (Access Key ID / Secret Access Key) or resolved from the
environment / assumed role / a named credentials profile — there is **no** file-upload control.
Legacy DOM: `fileInputs: 0`, `uploadButtons: []`. The new UI's `fileUpload` component activates
only for a field declaring `ui.fileMapping` (multi-key JSON distribution), which Redshift does
not have. **Decision:** do **not** add `fileUpload` to any Redshift field.

---

## Conditional fields & effects — tested

Each scenario was run on a fresh page (`verify-redshift-conditionals.js`); secure inputs
detected by placeholder, select-labels by rendered text, db fields by their `*` marker.

| Scenario                                                    | Reveals                                                                                                   | Hides                                                 | Matches schema `dependsOn`?                              |
| ----------------------------------------------------------- | --------------------------------------------------------------------------------------------------------- | ----------------------------------------------------- | -------------------------------------------------------- |
| **A** default (`authType=default`, Provisioned, Temp creds) | Assume Role ARN, External ID, Endpoint, Default Region, Cluster Identifier\*, Database User\*, Database\* | Access/Secret Key, Profile, Workgroup, Managed Secret | ✅ base state                                            |
| **B** `authType=keys`                                       | **Access Key ID\***, **Secret Access Key\***                                                              | Profile                                               | ✅ `authType == 'keys'`                                  |
| **C** `authType=credentials`                                | **Credentials Profile Name**                                                                              | Access/Secret Key                                     | ✅ `authType == 'credentials'`                           |
| **D** Serverless switch ON                                  | **Workgroup**                                                                                             | **Cluster Identifier**, **Database User**             | ✅ `useServerless == true` / `!= true`; dbUser gated off |
| **E** useManagedSecret → AWS Secrets Manager                | _(nothing — see discrepancy)_                                                                             | _(nothing changes)_                                   | ⚠️ **`useManagedSecret == true` does not re-evaluate**   |
| **F** `authType=grafana_assume_role`                        | —                                                                                                         | **Assume Role ARN**, **External ID**, **Endpoint**    | ✅ `authType != 'grafana_assume_role'`                   |

Observed transitions matching the schema exactly:

- **AWS auth discriminator** (`jsonData_authType`, a `select`) drives the credential fields:
  `keys` reveals the access/secret key pair, `credentials` reveals the profile, and
  `grafana_assume_role` hides the assume-role ARN / external ID / endpoint block. The
  `authType` select propagates its **string** value into the expression evaluator correctly.
- **Provisioning shape** (`jsonData_useServerless`, a `switch`) swaps **Cluster Identifier**
  (Provisioned) ↔ **Workgroup** (Serverless) and gates **Database User** off in Serverless. The
  `useServerless` switch propagates its **boolean** value correctly.

**Effects:** Redshift's schema contains **no** `effects` blocks — visibility is driven by plain
`dependsOn`/`requiredWhen` CEL expressions over `authType`, `useServerless`, and
`useManagedSecret`; there is no virtual selector that fans out to write multiple fields, so
nothing for `effects` to model, and none were added.

---

## Discrepancies not fixable via `dsconfig.json` (reported, not changed)

These are **pre-existing new-UI (plugin-ui) issues**, independent of the `required: true` fix.
Per the task constraints (never edit `plugin-ui`/`conformance.go`; the only sanctioned schema
change is the required-field/General-step fix), they are documented here rather than worked
around in the schema.

1. **Boolean-valued `radio` (`useManagedSecret`) does not drive its dependent fields.**
   Toggling the **AWS Secrets Manager** radio flips the underlying radio's `checked` state to
   `true` and the RadioButtonGroup re-renders (verified: `option-true-radiogroup-*`
   `checked: true` after the click — `debug-redshift-radio.js` / `debug-redshift-radio2.js`),
   **but** the dependency `jsonData_useManagedSecret == true` never re-evaluates:
   - **Managed Secret** (`jsonData_managedSecret_arn`, `dependsOn: useManagedSecret == true`)
     never appears (combobox count stays 4 before/after).
   - **Database User** (`jsonData_dbUser`, `requiredWhen: useServerless != true && useManagedSecret != true`)
     keeps its `*` (stays required) instead of relaxing.

   The sibling `useServerless` **switch** — same `boolean` type, same `== true` expression —
   works correctly, so the fault is specific to how the wizard reads **boolean-valued `radio`**
   selections (it appears to read the DOM value `"on"` rather than the option's declared boolean
   `true`, so `useManagedSecret == true` is never satisfied). Fixing this requires a **plugin-ui**
   change (honor a radio option's declared value in the expression evaluator). It is **not**
   fixable in `dsconfig.json` without changing the stored type from `boolean` to `string`, which
   would break the backend contract (`Config.UseManagedSecret bool`, `settings.go:170`) and the
   `JSONDataTypesMatchStruct` conformance test. **No schema change made; needs a plugin-ui fix.**

2. **`backend-only` field (`sessionToken`) renders in the new UI.** `secureJsonData_sessionToken`
   carries `tags: ["backend-only"]` (no editor writes it; the backend reads it —
   `settings.go:90-95`). The legacy editor does **not** render it, but the new UI shows it as a
   visible secure input (label `sessionToken`) in the Authentication section and the wizard's
   General step. Honoring the `backend-only` tag to hide it is a **plugin-ui** concern; removing
   the field or its group membership in `dsconfig.json` would drop it from the schema the backend
   relies on. **No schema change made.** (Cosmetic sibling note: `useManagedSecret` and
   `sessionToken` display their raw key as the field label because they have no `label` in the
   schema; legacy shows no label for the radio. Left as-is — outside the sanctioned fix.)

---

## Verification

```
go generate ./registry/grafana-redshift-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-redshift-datasource/...        # ALL subtests PASS
go test ./registry/grafana-redshift-datasource/... ./schema/...   # PASS (no regressions)
```

Conformance subtests (redshift), all **PASS**: `BaseFieldsResolved`, `SchemaRoundTrip`,
`SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`, `JSONDataMatchesStruct`,
`JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings`. The `TestLoadConfig`,
`TestApplyDefaults`, and `TestValidate` suites also pass — including
`TestValidate/missing_database_errors`, which confirms `database` is unconditionally required.

After regeneration, `schema.gen.json` and `settings.gen.json` moved `database` into the `jsonData`
`required: ["database"]` array and dropped the `x-dsconfig-required-when: "true"` extension.

---

## Files changed

- [`registry/grafana-redshift-datasource/dsconfig.json`](dsconfig.json) — changed
  `jsonData_database` from `"requiredWhen": "true"` to `"required": true` (so it renders in the
  wizard's General step and emits OpenAPI `required`). The real `dependsOn` / `requiredWhen`
  conditions on `authType`, `useServerless`, and `useManagedSecret` fields were left untouched.
- [`registry/grafana-redshift-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-redshift-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`database` now in the spec `required` array).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema/conformance.go`, and `plugin-ui`. The two discrepancies above require a `plugin-ui` fix
and are reported, not worked around.
