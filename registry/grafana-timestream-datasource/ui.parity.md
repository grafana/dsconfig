# Amazon Timestream — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-timestream-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (Grafana v13.0.1, uid `bfrbd04sjiuwwd`)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-timestream-datasource`
- **Method:** Playwright captured the legacy UI (`legacy-expand-timestream-parity.png`) and drove the new UI in both `tab` and `wizard` modes (local schema served via `context.route(...)`). Auth conditionals were exercised dynamically (`verify-timestream-conditionals-result.json`).
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

---

## Findings (no fixes needed)

**No Custom HTTP Headers.** Timestream is an AWS-SDK auth datasource (AWS SigV4 via access
keys / assume-role / credentials file / workspace IAM / Grafana-managed assume-role). Its legacy
editor has **no** Custom HTTP Headers section (`hasCustomHeaders:false`, `addHeaderBtn:false`), and
the new UI exposes no headers editor (`hasHeadersEditor:false` in both tab and wizard). Headers are
correctly **not** modeled.

**No `fileUpload`.** AWS credentials are entered as text (Access Key ID / Secret Access Key) or
resolved from the environment / profile / assumed role — there is **no** file-upload control
(`fileInputs:0`, `uploadButtons:[]`). `fileUpload` is correctly **not** used.

**No unconditional `requiredWhen:"true"` (no `required:true` fix).** Timestream has **no**
`requiredWhen:"true"` and **no** `required:true` field. The only two `requiredWhen` expressions are
auth-type-gated (`accessKey` / `secretKey` → `requiredWhen: "jsonData_authType == 'keys'"`), which
is correct. No change.

## New UI verification

Tab mode renders all four sections: **Authentication**, **Assume Role**, **Additional Settings**,
**Timestream Details**, plus **Save & Test** (`newgen-timestream-tab.json`). `hasHeadersEditor:false`
(correct). `urlPresent:true` — the optional AWS **Endpoint** override field, whose placeholder
matches the legacy Timestream cell-endpoint form `https://query-{cell}.timestream.{region}.amazonaws.com`.
Wizard mode (`newgen-timestream-wizard.json`) renders the same schema as a stepped flow
(step 1/5 Authentication first) with `hasHeadersEditor:false`.

## Field-by-field parity (highlights)

| Legacy field                      | schema id                                | Target         | Status                                 |
| --------------------------------- | ---------------------------------------- | -------------- | -------------------------------------- |
| Authentication Provider           | `jsonData_authType` (auth.discriminator) | `jsonData`     | ✅ 🔀                                  |
| Credentials Profile Name          | `jsonData_profile`                       | `jsonData`     | ✅ 🔀 (credentials auth)               |
| Access Key ID / Secret Access Key | `secureJsonData_accessKey` / `secretKey` | secureJsonData | ✅ 🔀 (keys auth)                      |
| Assume Role ARN                   | `jsonData_assumeRoleArn`                 | `jsonData`     | ✅                                     |
| External ID                       | `jsonData_externalId`                    | `jsonData`     | ✅ 🔀 (hidden for grafana_assume_role) |
| Endpoint                          | `jsonData_endpoint`                      | `jsonData`     | ✅ 🔀 (hidden for grafana_assume_role) |
| Default Region                    | `jsonData_defaultRegion`                 | `jsonData`     | ✅                                     |
| Database                          | `jsonData_defaultDatabase`               | `jsonData`     | ✅                                     |
| Table                             | `jsonData_defaultTable`                  | `jsonData`     | ✅ 🔀 (depends on Database)            |
| Measure                           | `jsonData_defaultMeasure`                | `jsonData`     | ✅ 🔀 (depends on Database + Table)    |

## Conditional fields — tested

The Authentication Provider discriminator drives which credential fields are shown/required. All
four editor-selectable scenarios were verified live in the new UI
(`verify-timestream-conditionals-result.json`, `verify-timestream-cond-*.png`):

| authType                                    | accessKey/secretKey | profile   | assumeRoleArn | externalId | endpoint   | region | database |
| ------------------------------------------- | ------------------- | --------- | ------------- | ---------- | ---------- | ------ | -------- |
| AWS SDK Default (`default`)                 | hidden              | hidden    | shown         | shown      | shown      | shown  | shown    |
| Access & secret key (`keys`)                | **shown**           | hidden    | shown         | shown      | shown      | shown  | shown    |
| Credentials file (`credentials`)            | hidden              | **shown** | shown         | shown      | shown      | shown  | shown    |
| Grafana Assume Role (`grafana_assume_role`) | hidden              | hidden    | shown         | **hidden** | **hidden** | shown  | shown    |

`externalId` and `endpoint` correctly disappear under `grafana_assume_role`
(`dependsOn: "jsonData_authType != 'grafana_assume_role'"`). The Timestream macro chain renders per
schema: `Default Region` and `Database` are always shown; `Table` (`dependsOn defaultDatabase != ''`)
and `Measure` (`dependsOn defaultDatabase != '' && defaultTable != ''`) stay hidden until their
parent select has a value — matching the legacy editor, where Table/Measure are populated on demand
from `/resources/tables` and `/resources/measures` after a Database is chosen.

## Observation — `sessionToken` (plugin-ui behavior, not a dsconfig.json gap)

The new UI renders a **`sessionToken`** input in the Authentication section that the legacy editor
does **not** render. This is `secureJsonData_sessionToken`, tagged `backend-only` — it is written by
provisioning only, and the legacy `ConnectionConfig` editor never renders an input for it.

This is a **plugin-ui rendering behavior**: the `DatasourceConfigWizard` does not currently suppress
fields tagged `backend-only`. It is **identical to the reference plugin `cloudwatch`**, whose
`dsconfig.json` models `secureJsonData_sessionToken` the same way (same `backend-only` tag, same
inclusion in the `authentication` group's `fieldRefs`) and whose new UI likewise renders
`sessionToken` — yet `cloudwatch` was signed off as "parity already achieved." The schema is
correct (the field is real and correctly marked `backend-only`); honoring the tag at render time is a
`plugin-ui` concern, which is out of scope for this validation (no `plugin-ui`/conformance edits).
It is therefore **not fixable purely via `dsconfig.json`** without diverging from the established
cloudwatch pattern, and is reported here rather than "fixed."

---

## Verification

```
go test ./registry/grafana-timestream-datasource/...
# TestSchemaConformance   8/8  PASS  (incl. SchemaArtifactInSync — committed .gen.json in sync)
# TestLoadConfig         19/19 PASS
# TestApplyDefaults       3/3  PASS
# TestValidate           13/13 PASS
# TestEffectiveProfile    4/4  PASS
# ok  github.com/grafana/dsconfig/registry/grafana-timestream-datasource
```

No schema edit was made, so no regeneration was needed; the committed artifacts remain in sync
(`SchemaArtifactInSync` passes).

## Files changed

**None.** Timestream was already at parity for HTTP headers (n/a), file upload (n/a), and the
required/conditional behaviour (auth-gated `requiredWhen` is correct; no unconditional
`requiredWhen:"true"`). The one legacy-vs-new difference (`sessionToken` rendering) is a pre-existing
`plugin-ui` behavior identical to the accepted `cloudwatch` reference and is not fixable via
`dsconfig.json`. This report documents the validation only.
