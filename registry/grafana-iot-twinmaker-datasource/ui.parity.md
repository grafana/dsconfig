# AWS IoT TwinMaker — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-iot-twinmaker-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/afrbd04zr8av4e` (Grafana Enterprise 13.0.1)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-iot-twinmaker-datasource`
- **Method:** Playwright captured the legacy UI (`legacy-expand-twinmaker-verify.png`) and drove the new UI (local schema served via `context.route(...)`), in both `tab` and `wizard` modes, plus a per-auth-provider conditional sweep.
- **Result:** **One fix applied** — `assumeRoleArn` and `workspaceId` converted from `requiredWhen:"true"` to `required:true` so they render as unconditionally-required. Parity otherwise already held.

---

## Findings

**No Custom HTTP Headers.** IoT TwinMaker is an AWS-SDK auth datasource (AWS SigV4 via the
`@grafana/aws-sdk` `ConnectionConfig`: access keys / credentials file / workspace IAM /
AWS SDK default, then STS assume-role). Its legacy editor has **no** Custom HTTP Headers
section (`hasCustomHeaders:false`, `addHeaderBtn:false`). Headers are correctly **not**
modeled, and the new UI shows `hasHeadersEditor:false`.

**No `fileUpload`.** AWS credentials are entered as text (Access Key ID / Secret Access Key)
or resolved from the environment / a named profile / an assumed role — there is **no**
file-upload control (`fileInputs:0`, `uploadButtons:[]`). `fileUpload` is correctly **not**
used.

**Required-field fix (the General-step fix).** Legacy TwinMaker requires both `assumeRoleArn`
and `workspaceId` regardless of auth type — the backend `CheckHealth`
(`pkg/plugin/datasource.go:171-184`) fails with `Missing WorkspaceID configuration` and
`Assume Role ARN is required` when either is empty, and the editor shows a red Alert directing
users to the IoT TwinMaker dashboard-IAM-role guide. The schema encoded this as
`requiredWhen:"true"`, which the generator emitted as the non-standard
`x-dsconfig-required-when:"true"` extension rather than a JSON-Schema `required` array, so the
new UI did **not** render the required marker. Changed both to `required:true`; the generated
artifacts now carry a standard `"required": ["assumeRoleArn", "workspaceId"]` array under
`jsonData`, matching the sibling AWS pattern (`grafana-athena-datasource`).

## New UI verification

**Tab mode** (`newgen-twinmaker-fixed-tab.png`) renders all four groups: **Authentication**,
**Assume Role**, **Additional Settings**, **Twinmaker Settings**. `hasHeadersEditor:false`
(correct). `urlPresent:true` refers to the optional AWS **Endpoint** override field
(placeholder `https://{service}.{region}.amazonaws.com`), not a datasource URL. The header note
"Fields marked with \* are required" appears, and **Assume Role ARN \*** and **Workspace \***
carry the required asterisk; all other fields are unmarked.

**Wizard mode** (`verify-twinmaker-wizard-step1.png`) synthesizes a **"General" step (1/5)**
containing the AWS auth surface (Authentication Provider) plus the now-required **Assume Role
ARN \*** and **Workspace \*** fields (with External ID and Endpoint). The step reports exactly
`requiredMarkerCount:2`, attributable precisely to the two fixed fields:

```
Assume Role ARN:        REQUIRED(*)
Workspace:              REQUIRED(*)
External ID:            present(no *)
Endpoint:               present(no *)
Default Region:         present(no *)
Authentication Provider: present(no *)
```

## Field-by-field parity (highlights)

| Legacy field                      | schema id                                | Target         | Status                       |
| --------------------------------- | ---------------------------------------- | -------------- | ---------------------------- |
| Authentication Provider           | `jsonData_authType` (auth discriminator) | `jsonData`     | ✅ 🔀                        |
| Access Key ID / Secret Access Key | `secureJsonData_accessKey` / `secretKey` | secureJsonData | ✅ 🔀 (keys auth)            |
| Credentials Profile Name          | `jsonData_profile`                       | `jsonData`     | ✅ 🔀 (credentials auth)     |
| Assume Role ARN                   | `jsonData_assumeRoleArn`                 | `jsonData`     | ✅ **required:true (fixed)** |
| External ID                       | `jsonData_externalId`                    | `jsonData`     | ✅ 🔀                        |
| Endpoint                          | `jsonData_endpoint`                      | `jsonData`     | ✅ 🔀                        |
| Default Region                    | `jsonData_defaultRegion`                 | `jsonData`     | ✅                           |
| Workspace                         | `jsonData_workspaceId`                   | `jsonData`     | ✅ **required:true (fixed)** |
| Define write permissions (switch) | `virtual_alarmConfigEnabled`             | virtual        | ✅ (effects)                 |
| Assume Role ARN Write             | `jsonData_assumeRoleArnWriter`           | `jsonData`     | ✅ 🔀 (switch-gated)         |

## Conditional fields — tested

The **Authentication Provider** discriminator drives which credential fields are shown. A
four-scenario Playwright sweep (`verify-twinmaker-conditionals-result.json`) confirmed every
editor-selectable provider reveals correctly:

| authType (label)                    | Access Key ID | Secret Access Key | Credentials Profile |
| ----------------------------------- | ------------- | ----------------- | ------------------- |
| `default` (AWS SDK Default)         | hidden        | hidden            | hidden              |
| `keys` (Access & secret key)        | **shown**     | **shown**         | hidden              |
| `credentials` (Credentials file)    | hidden        | hidden            | **shown**           |
| `ec2_iam_role` (Workspace IAM Role) | hidden        | hidden            | hidden              |

`assumeRoleArn`, `externalId`, `endpoint`, `defaultRegion`, `workspaceId`, and the Alarm
Configuration switch stay visible across all four (no auth-type dependency); `assumeRoleArn`
and `workspaceId` stay required throughout. `grafana_assume_role` and legacy `arn` are
storage-valid but intentionally **not** editor-selectable (`@grafana/aws-sdk@0.8.3` restricts
Grafana Assume Role to a cloudwatch/athena/amazonprometheus allow-list). The
`assumeRoleArnWriter` input is gated behind the `virtual_alarmConfigEnabled` switch via
`dependsOn`, with an `effects` clause clearing it when the switch is toggled off.

---

## Verification

```
go generate ./registry/grafana-iot-twinmaker-datasource/...   # regenerated .gen.json artifacts
go test ./registry/grafana-iot-twinmaker-datasource/...       # PASS — 8/8 conformance subtests + settings tests
go test ./registry/grafana-iot-twinmaker-datasource/... ./schema/...   # PASS
```

`TestSchemaConformance` (BaseFieldsResolved, SchemaRoundTrip, SchemaArtifactInSync,
SchemaSpecHasNoSecureJSON, ConfigSchemaValid, JSONDataMatchesStruct, JSONDataTypesMatchStruct,
SecureValuesMatchLoadSettings) all pass, as do `TestLoadConfig`, `TestApplyDefaults`, and
`TestValidate`.

## Files changed

- **`dsconfig.json`** — `jsonData_assumeRoleArn` and `jsonData_workspaceId`:
  `requiredWhen:"true"` → `required:true`. Conditional `requiredWhen` on
  `secureJsonData_accessKey` / `secretKey` (`jsonData_authType == 'keys'`) left unchanged.
- **`schema.gen.json`** / **`settings.gen.json`** — regenerated: both now carry
  `"required": ["assumeRoleArn", "workspaceId"]` under `jsonData` and drop the
  `x-dsconfig-required-when:"true"` extension from the two fields.

No HTTP headers (n/a), no file upload (n/a), no packs. No `settings.go` / `settings.ts` /
`README` / `schema` / `conformance` / `plugin-ui` changes were required.
