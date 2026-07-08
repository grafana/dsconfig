# Amazon Athena — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-athena-datasource` (an **AWS-SDK auth** datasource — AWS SigV4 via
  the shared `@grafana/aws-sdk` `ConnectionConfig` frontend + `awsds.AWSDatasourceSettings` backend)
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/afrbd04oq6fwgd` (Grafana Enterprise 13.x)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-athena-datasource` (Storybook, `ConfigEditor/DatasourceConfigWizard`; also `--wizard`)
- **Method:** Playwright captured the legacy UI (`capture-legacy-expand.js afrbd04oq6fwgd`) and drove the
  new UI (`capture-new-generic.js` in `tab`/`wizard` modes, plus `verify-athena-wizard.js` and
  `verify-athena-conditionals.js`). The Storybook story fetches the schema from
  `raw.githubusercontent.com/.../registry/grafana-athena-datasource/dsconfig.json`; the local (edited)
  `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so
  the screenshots reflect the local schema without pushing.
- **Result:** **Parity achieved.** Every legacy-editor field is modeled in `dsconfig.json` and routes to
  the same storage target. Athena is an AWS-SDK datasource, so it correctly has **no HTTP-headers editor**
  and **no file-upload** control. The one change required was making the three unconditionally-required
  Athena selectors use `required: true` so the wizard's synthetic **General** step pulls them in. All four
  AWS auth-provider conditionals were exercised and confirmed to reveal/hide the right credential fields.

---

## TL;DR of changes

| #   | Change                                                                                                                               | File                             | Why                                                                                                                                                                                                                                                                                                                                                                                    |
| --- | ------------------------------------------------------------------------------------------------------------------------------------ | -------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| 1   | Changed **`jsonData_catalog`**, **`jsonData_database`**, **`jsonData_workgroup`** from `"requiredWhen": "true"` → `"required": true` | [`dsconfig.json`](dsconfig.json) | These are unconditionally required (the backend `Validate()` rejects an empty catalog/database/workgroup — `settings.go:258-266`). The wizard's synthetic **General** step only pulls fields with `required: true`; `requiredWhen` is a CEL expression the resolver does not inspect. Also emits proper OpenAPI `required` arrays instead of the `x-dsconfig-required-when` extension. |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-athena-datasource/...`                        | generated artifacts              | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`)                                                                                                                                                                                                                                                                                                                   |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance.go`, or `plugin-ui`.
The `requiredWhen` conditions that are **real** conditions (`secureJsonData_accessKey` /
`secureJsonData_secretKey`, both gated on `jsonData_authType == 'keys'`) were **left untouched** — only
the three literal `"requiredWhen": "true"` values were converted. No `plugin-ui` change was needed: the
auth group already uses the conventional id `authentication`, which the wizard's required-fields resolver
recognises, so the AWS auth fields fold into the General step correctly.

---

## Section layout

Verified rendering top-to-bottom in the new UI (tab mode) and matched to the legacy editor's section
headings.

| Order | Section (`id`)                                  | `optional` | Fields (in display order)                                                                           |
| ----- | ----------------------------------------------- | ---------- | --------------------------------------------------------------------------------------------------- |
| 1     | **Authentication** (`authentication`)           | no         | Authentication Provider, Credentials Profile Name, Access Key ID, Secret Access Key, (sessionToken) |
| 2     | **Assume Role** (`assume-role`)                 | no         | Assume Role ARN, External ID                                                                        |
| 3     | **Additional Settings** (`additional-settings`) | no         | Endpoint, Default Region                                                                            |
| 4     | **Athena Details** (`athena-details`)           | no         | Data source (catalog), Database, Workgroup, Output Location                                         |

Legacy DOM headings (`legacy-expand-athena-parity.json`): `Connection Details`, `Authentication`,
`Assume Role`, `Additional Settings`, `Athena Details`. The legacy `Connection Details` is the AWS
`ConnectionConfig` outer wrapper; the new schema splits its contents across `Authentication` /
`Assume Role` / `Additional Settings`, keeping the Athena-specific block under `Athena Details`. New UI
tab-mode section buttons (`newgen-athena-tab.json`): `Authentication`, `Assume Role`,
`Additional Settings`, `Athena Details` — same four groups, and `hasHeadersEditor: false`.

### Wizard mode: the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) every field in the
auth group (`authentication`), plus their `dependsOn` parents/children.

Confirmed in `verify-athena-wizard.js` (step **General 1/5**):

- **Data source\*** (`jsonData_catalog`), **Database\*** (`jsonData_database`), **Workgroup\***
  (`jsonData_workgroup`) — the three `required: true` fields (`requiredMarkerCount: 3`).
- **Authentication Provider** (`jsonData_authType`), **sessionToken**, and the auth discriminator's
  `dependsOn` children **Assume Role ARN** / **External ID** / **Endpoint** — folded in as auth-group /
  dependent members.
- **Default Region** and **Output Location** are **not** in General (they surface in the
  `Additional Settings` and `Athena Details` steps respectively).

**Effect of the `required: true` fix (before/after):**

| Field                                                  | Before (`requiredWhen: "true"`) | After (`required: true`)  |
| ------------------------------------------------------ | ------------------------------- | ------------------------- |
| Data source (`jsonData_catalog`, Athena Details group) | **absent** from General         | **present** in General ✅ |
| Database (`jsonData_database`, Athena Details group)   | **absent** from General         | **present** in General ✅ |
| Workgroup (`jsonData_workgroup`, Athena Details group) | **absent** from General         | **present** in General ✅ |

Before the fix the wizard's General step held only the auth fields; the three Athena selectors were
reachable only in the `Athena Details` step. After the fix all three appear in General with `*` required
markers. Tab mode is unaffected — the synthetic `_required` group is filtered out there, so it still shows
the four sections in order.

---

## Field-by-field parity

Legend: ✅ present & matching · 🔀 conditional (revealed by `dependsOn`)

| Legacy UI field                    | Control (legacy) | New UI (schema id)            | Storage target                | Status                                      |
| ---------------------------------- | ---------------- | ----------------------------- | ----------------------------- | ------------------------------------------- |
| Authentication Provider            | select           | `jsonData_authType`           | `jsonData.authType`           | ✅ 🔀 (discriminator)                       |
| Credentials Profile Name           | text input       | `jsonData_profile`            | `jsonData.profile`            | ✅ 🔀 (`authType == 'credentials'`)         |
| Access Key ID                      | secure input     | `secureJsonData_accessKey`    | `secureJsonData.accessKey`    | ✅ 🔀 (`authType == 'keys'`)                |
| Secret Access Key                  | secure input     | `secureJsonData_secretKey`    | `secureJsonData.secretKey`    | ✅ 🔀 (`authType == 'keys'`)                |
| _(no legacy field — backend-only)_ | —                | `secureJsonData_sessionToken` | `secureJsonData.sessionToken` | modeled (`backend-only` tag) — see note     |
| Assume Role ARN                    | text input       | `jsonData_assumeRoleArn`      | `jsonData.assumeRoleArn`      | ✅ 🔀 (`authType != 'grafana_assume_role'`) |
| External ID                        | text input       | `jsonData_externalId`         | `jsonData.externalId`         | ✅ 🔀 (`authType != 'grafana_assume_role'`) |
| Endpoint                           | text input       | `jsonData_endpoint`           | `jsonData.endpoint`           | ✅ 🔀 (`authType != 'grafana_assume_role'`) |
| Default Region                     | select           | `jsonData_defaultRegion`      | `jsonData.defaultRegion`      | ✅ (see note)                               |
| Data source                        | select           | `jsonData_catalog`            | `jsonData.catalog`            | ✅ **required**                             |
| Database                           | select           | `jsonData_database`           | `jsonData.database`           | ✅ **required**                             |
| Workgroup                          | select           | `jsonData_workgroup`          | `jsonData.workgroup`          | ✅ **required**                             |
| Output Location                    | text input       | `jsonData_outputLocation`     | `jsonData.outputLocation`     | ✅                                          |

All legacy fields are modeled and route to identical storage targets. The legacy `Name` and `Default`
controls at the top are Grafana editor chrome (datasource name + default toggle), not part of the
datasource config, and are correctly **not** modeled.

> **Note — `urlPresent: true` in the new-UI captures is a false positive.** Athena has **no** root URL
> field. The generic capture's `urlPresent` heuristic matches any `input[placeholder*="http"]`, which here
> is the optional **Endpoint** field (`placeholder="https://{service}.{region}.amazonaws.com"`), not a
> connection URL.

> **Note — `sessionToken` renders in the new UI though the legacy editor hides it.** The field carries
> `tags: ["backend-only"]` and the legacy editor deliberately renders no control for it (the backend still
> reads it; it is set via provisioning). The schema-driven `plugin-ui` renderer does not currently suppress
> `backend-only`-tagged fields, so a `sessionToken` label appears in both tab and wizard. This is a
> pre-existing renderer behaviour, **out of scope** for the required-field fix, and not fixable via
> `dsconfig.json` without deleting a field the backend/provisioning path needs (suppression belongs in
> `plugin-ui` honouring the `backend-only` tag). No change made.

> **Note — `Default Region` is backend-required but modeled without `required: true`.** The backend
> `Validate()` also rejects an empty `defaultRegion` (`settings.go:255-257`), but `jsonData_defaultRegion`
> carries neither `required: true` nor a `requiredWhen` in the schema, so it lands in the
> `Additional Settings` step rather than General. This task's scope was the three `"requiredWhen": "true"`
> selectors only (`catalog`/`database`/`workgroup`); `defaultRegion` was left as-is. It is a candidate for a
> future `required: true` if strict General-step parity for it is desired.

---

## Gaps found

**None beyond the required-field fix.** The complete legacy field set is already modeled, so no field had
to be added.

### Custom HTTP Headers — not applicable (verified)

Athena authenticates with AWS SigV4 (access keys / assume-role / credentials file / workspace IAM /
Grafana-managed temp credentials), **not** HTTP Basic/Bearer with custom headers. Legacy DOM capture
(`legacy-expand-athena-parity.json`): `hasCustomHeaders: false`, `addHeaderBtn: false`. New UI (tab +
wizard): `hasHeadersEditor: false`. Correctly **not** modeled.

---

## `fileUpload` evaluation — not applicable to Athena

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not:

- AWS credentials are entered as text (Access Key ID / Secret Access Key), read from a named
  `~/.aws/credentials` profile, or resolved from the environment / assumed role — there is **no**
  file-upload control. Legacy capture: `fileInputs: 0`, `uploadButtons: []`.
- The new UI's `fileUpload` component only activates when a field declares `ui.fileMapping` (multi-key JSON
  distribution, e.g. a GCP service-account file), which Athena has no use for.

**Decision:** do **not** add `fileUpload` to any Athena field.

---

## Conditional fields & effects — tested

Athena drives its conditionals with the **Authentication Provider** `select` discriminator
(`jsonData_authType`). Each scenario was run on a fresh page and the visible credential fields probed
(`verify-athena-conditionals.js`; conditionally-hidden fields are removed from the DOM by `dependsOn`):

| Scenario (`authType`)                             | Access Key ID | Secret Access Key | Credentials Profile | Assume Role ARN | External ID | Endpoint   | Matches schema?                                  |
| ------------------------------------------------- | ------------- | ----------------- | ------------------- | --------------- | ----------- | ---------- | ------------------------------------------------ |
| **A** AWS SDK Default (`default`)                 | hidden        | hidden            | hidden              | **shown**       | **shown**   | **shown**  | ✅ no keys/profile; assume-role + endpoint shown |
| **B** Access & secret key (`keys`)                | **shown**     | **shown**         | hidden              | **shown**       | **shown**   | **shown**  | ✅ key pair revealed (required); still assumable |
| **C** Credentials file (`credentials`)            | hidden        | hidden            | **shown**           | **shown**       | **shown**   | **shown**  | ✅ profile revealed; keys hidden                 |
| **D** Grafana Assume Role (`grafana_assume_role`) | hidden        | hidden            | hidden              | **hidden**      | **hidden**  | **hidden** | ✅ ARN/External ID/Endpoint gated off            |

Observed transitions, exactly matching the schema:

- Selecting **`keys`** reveals the **Access Key ID** + **Secret Access Key** secure inputs (their
  `dependsOn` / `requiredWhen` are both `jsonData_authType == 'keys'`), which the General step marks
  required.
- Selecting **`credentials`** reveals **Credentials Profile Name** (`dependsOn jsonData_authType ==
'credentials'`) and hides the key pair.
- Selecting **`grafana_assume_role`** hides **Assume Role ARN**, **External ID**, and **Endpoint** (all
  three carry `dependsOn jsonData_authType != 'grafana_assume_role'`, because Grafana derives the trust
  relationship itself).

The schema contains **no** `effects` blocks — visibility is a set of plain `dependsOn` CEL expressions over
the single auth discriminator, so there is nothing for `effects` to model and none were added.

---

## Verification

```
go generate ./registry/grafana-athena-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-athena-datasource/...        # conformance + settings tests PASS
go test ./registry/... ./schema/...                     # entire suite PASS (no regressions)
```

Conformance subtests (`TestSchemaConformance`): `BaseFieldsResolved`, `SchemaRoundTrip`,
`SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`, `JSONDataMatchesStruct`,
`JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings` — **8/8 PASS**. The plugin's own
`TestLoadConfig`, `TestApplyDefaults`, and `TestValidate` suites also pass (including the
`missing_catalog` / `missing_database` / `missing_workgroup` cases that assert the backend contract these
three `required: true` fields mirror).

After regeneration, `settings.gen.json` and `schema.gen.json` moved `catalog` / `database` / `workgroup`
into the `jsonData` `required` array and dropped the three `x-dsconfig-required-when: "true"` extensions.
The real `x-dsconfig-depends-on` conditions (profile, keys, assume-role, endpoint) are unchanged.

---

## Files changed

- [`registry/grafana-athena-datasource/dsconfig.json`](dsconfig.json) — changed `jsonData_catalog`,
  `jsonData_database`, and `jsonData_workgroup` from `"requiredWhen": "true"` to `"required": true` (so
  they render in the wizard's General step and emit OpenAPI `required`). The real `requiredWhen` conditions
  on `secureJsonData_accessKey` / `secureJsonData_secretKey` were left untouched.
- [`registry/grafana-athena-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-athena-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`catalog` / `database` / `workgroup` now in the spec `required` arrays).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema/conformance.go`, and `plugin-ui`.
