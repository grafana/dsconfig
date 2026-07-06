# MySQL — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `mysql`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/bfranib5frta8a` (Grafana Enterprise, `@grafana/sql` `ConfigurationEditor`)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:mysql` (Storybook, `ConfigEditor/DatasourceConfigWizard`) — also exercised in `--wizard` mode
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/mysql/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the captures reflect the local schema without pushing.
- **Result:** **Parity achieved.** MySQL is a SQL datasource: every legacy field is present in the new UI, no field was missing, and the only schema change needed was making the two unconditionally-required fields (`url`, `user`) render in the wizard's **General** step. No Custom HTTP Headers, no `fileUpload`, no packs, and — unlike graphite — **no `conformance.go` or `plugin-ui` change was required**.

---

## TL;DR of changes

| #   | Change                                                                                        | File                             | Why                                                                                                                             |
| --- | --------------------------------------------------------------------------------------------- | -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| 1   | `root_url`: `"requiredWhen": "true"` → `"required": true`                                      | [`dsconfig.json`](dsconfig.json) | `url` is unconditionally required (legacy shows **Host URL \***); puts URL into the wizard's **General** step + OpenAPI `required` |
| 2   | `root_user`: `"requiredWhen": "true"` → `"required": true`                                     | [`dsconfig.json`](dsconfig.json) | `user` is unconditionally required (legacy shows **Username \***); puts User into **General** + OpenAPI `required`               |
| 3   | Regenerated the `.gen.json` artifacts (`schema.gen.json` + `settings.gen.json` changed; `settings.examples.gen.json` unchanged) | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`); `url`+`user` now in the spec's `required` array |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `schema/conformance.go`, or `plugin-ui`. All changes flow through `dsconfig.json` with the `.gen.json` artifacts produced by `go generate`.

The two edited fields were the **only** `requiredWhen: "true"` occurrences in the schema. All real `requiredWhen`/`dependsOn` conditions (the TLS cert reveals) were left untouched.

---

## Section layout

Verified rendering top-to-bottom in the new UI (tab mode, sidebar + accordion) and
matched against the legacy editor's section headings.

| Order | Section (`id`)                          | `optional` | Fields (display order)                                                                            |
| ----- | --------------------------------------- | ---------- | ------------------------------------------------------------------------------------------------- |
| 1     | **Connection** (`connection`)           | no         | Host URL \*, Database name                                                                        |
| 2     | **Authentication** (`authentication`)   | no         | Username \*, Password, Use TLS Client Auth, With CA Cert, Skip TLS Verification, Allow Cleartext Passwords |
| 3     | **TLS/SSL Auth Details** (`tls-details`) | yes        | TLS/SSL Client Certificate, TLS/SSL Root Certificate, TLS/SSL Client Key (all conditional)        |
| 4     | **Additional settings** (`additional-settings`) | yes  | Session timezone, Min time interval, Max open, Auto max idle, Max idle, Max lifetime              |

Notes:

- Legacy headings observed: **Connection**, **Authentication**, **Additional settings** (with two
  `<h6>` sub-groupings inside Additional settings — **MySQL Options** for timezone/interval and
  **Connection limits** for the pool fields). The new UI lists the Additional-settings fields flat
  under one section. This is a cosmetic sub-grouping difference only — **no field is missing** and
  all six fields are present.
- The **TLS/SSL Auth Details** section is empty until a TLS toggle is enabled (see
  [Conditional fields](#conditional-fields--effects--tested)); the legacy editor behaves the same way.
- The authentication group uses **`id: "authentication"`** (the registry convention). The `plugin-ui`
  wizard already recognises both `authentication` and `auth` as the auth group (from the earlier
  graphite fix, already merged), so the auth fields fold into the wizard's **General** step without any
  further `plugin-ui` change for mysql.

### Wizard mode: URL + User in the "General" step

In **wizard mode** the `plugin-ui` builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the
`authentication`/`auth` group's fields, plus their `dependsOn` parents/children.

- Before: `root_url` and `root_user` used `requiredWhen: "true"` (a CEL expression the resolver does
  **not** inspect), so no field seeded a General step.
- After: both are `required: true` — unconditionally required (the legacy editor marks Host URL and
  Username with a `*`, and `Config.Validate()` in `settings.go:181-186` rejects an empty URL/user). This
  puts both into General and emits a proper OpenAPI `required: ["url","user"]` in the generated settings
  spec (instead of two `x-dsconfig-required-when: "true"` extensions).

**Verified (capture `mysql-verify-wizard`):** the wizard opens on **General (1/5)** containing **Host URL**
(placeholder `localhost:3306`) and **Username** (placeholder `Username`) both marked required (2 required
`*` markers), plus the folded-in **Password** and the auth-group TLS toggles. Tab mode is unaffected — the
synthetic `_required` group is filtered out there, so it shows the four sections in order.

---

## Field-by-field parity

Legend: ✅ present & matching · 🔀 conditional (revealed by `dependsOn`)

| Legacy UI field              | Control (legacy)             | New UI (schema id)                 | Control (new)      | Storage target                     | Status |
| ---------------------------- | ---------------------------- | ---------------------------------- | ------------------ | ---------------------------------- | ------ |
| Host URL \*                  | text input                   | `root_url`                         | input              | `root.url` (required)              | ✅     |
| Database name                | text input                   | `jsonData_database`                | input              | `jsonData.database`                | ✅     |
| Username \*                  | text input                   | `root_user`                        | input              | `root.user` (required)             | ✅     |
| Password                     | password (`SecretFormField`) | `secureJsonData_password`          | secure input       | `secureJsonData.password`          | ✅     |
| Use TLS Client Auth          | switch                       | `jsonData_tlsAuth`                 | switch             | `jsonData.tlsAuth`                 | ✅     |
| With CA Cert                 | switch                       | `jsonData_tlsAuthWithCACert`       | switch             | `jsonData.tlsAuthWithCACert`       | ✅     |
| Skip TLS Verification        | switch                       | `jsonData_tlsSkipVerify`           | switch             | `jsonData.tlsSkipVerify`           | ✅     |
| Allow Cleartext Passwords    | switch                       | `jsonData_allowCleartextPasswords` | switch             | `jsonData.allowCleartextPasswords` | ✅     |
| TLS/SSL Client Certificate   | **textarea**                 | `secureJsonData_tlsClientCert`     | secure input¹      | `secureJsonData.tlsClientCert`     | ✅ 🔀  |
| TLS/SSL Root Certificate     | **textarea**                 | `secureJsonData_tlsCACert`         | secure input¹      | `secureJsonData.tlsCACert`         | ✅ 🔀  |
| TLS/SSL Client Key           | **textarea**                 | `secureJsonData_tlsClientKey`      | secure input¹      | `secureJsonData.tlsClientKey`      | ✅ 🔀  |
| Session timezone             | text input                   | `jsonData_timezone`                | input              | `jsonData.timezone`                | ✅     |
| Min time interval            | text input                   | `jsonData_timeInterval`            | input              | `jsonData.timeInterval`            | ✅     |
| Max open                     | number                       | `jsonData_maxOpenConns`            | number             | `jsonData.maxOpenConns`            | ✅     |
| Auto max idle                | switch                       | `jsonData_maxIdleConnsAuto`        | switch             | `jsonData.maxIdleConnsAuto`        | ✅     |
| Max idle                     | number                       | `jsonData_maxIdleConns`            | number             | `jsonData.maxIdleConns`            | ✅     |
| Max lifetime                 | number                       | `jsonData_connMaxLifetime`         | number             | `jsonData.connMaxLifetime`         | ✅     |

Verified counts (new UI, tab, all sections expanded): 6 placeholders (`localhost:3306`, `Database`,
`Username`, `Password`, `Europe/Berlin or +02:00`, `1m`) and 3 number inputs (Max open / Max idle /
Max lifetime), matching the legacy inventory (`legacy-mysql-fields`: `hasHeaders:false`,
`fileInputs:0`, `uploadBtns:0`).

¹ **Not a discrepancy.** The three TLS cert fields declare `ui.component: "textarea"`, but the new
renderer draws any `secureJsonData` field as a masked secure input (the `target === "secureJsonData"`
branch precedes the `textarea` branch — the same renderer policy documented for graphite). The legacy
editor uses plain textareas (verified: 3 textareas with PEM placeholders). Both UIs collect the same
PEM text into the same `secureJsonData` keys; only the widget affordance differs.

---

## Gaps found

**None missing.** Every legacy field is present in the new UI with a matching storage target. The only
schema change required was the `required: true` fix (see [TL;DR](#tldr-of-changes)).

The single layout nuance — the legacy `<h6>` sub-groupings **MySQL Options** and **Connection limits**
inside *Additional settings* vs. the new UI's flat list — is cosmetic and introduces no field gap. Fields
were kept **inline** (no packs), per the entry's modeling.

---

## `fileUpload` evaluation — not applicable to mysql

The task asks to use the `fileUpload` control **only if the legacy UI uses it**. It does not:

- The legacy `@grafana/sql` TLS section renders the client cert / client key / root cert fields as plain
  **textareas** (placeholders `-----BEGIN CERTIFICATE-----` / `-----BEGIN RSA PRIVATE KEY-----`).
- No file-upload button and no `<input type="file">` were found in the legacy DOM
  (`legacy-mysql-fields`: `fileInputs:0`, `uploadButtons:[]`; `mysql-legacy-tls`: 3 textareas revealed on toggle).

**Decision:** `fileUpload` is **not** added to any mysql field. The cert fields keep their current
`secureJsonData` string modeling.

---

## Custom HTTP Headers — not applicable to mysql

MySQL is a SQL datasource that connects over the MySQL wire protocol (TCP or Unix socket), **not** HTTP.
The legacy editor has **no** "Custom HTTP Headers" / "HTTP headers" section and no **Add header** button
(`legacy-mysql-fields` + `legacy-expand`: `hasCustomHeaders:false`, `addHeaderBtn:false`). The new UI
correctly renders **no headers editor** (`hasHeadersEditor:false` in both tab and wizard captures).

**Decision:** Custom HTTP Headers are **not** modeled. (This is the opposite of graphite, an HTTP
datasource, which needed them added.)

---

## Conditional fields & effects — tested

The TLS `dependsOn` conditionals were exercised in **both** UIs and confirmed to reveal the cert fields.

| Trigger                                    | Revealed field(s)                                | New UI (`mysql-verify-tls`)          | Legacy (`mysql-legacy-tls`)          |
| ------------------------------------------ | ------------------------------------------------ | ------------------------------------ | ------------------------------------ |
| _(both toggles off)_                       | —                                                | no cert fields ✅                    | 0 textareas ✅                       |
| `jsonData_tlsAuth == true` (Use TLS Client Auth) | TLS/SSL **Client Certificate** + **Client Key** | both revealed; Root Cert still hidden ✅ | textareas revealed ✅             |
| `jsonData_tlsAuthWithCACert == true` (With CA Cert) | TLS/SSL **Root Certificate**              | revealed ✅                          | textarea revealed ✅                 |

New-UI observation (DOM probe): with both off, `{clientCert:false, clientKey:false, rootCert:false}`;
after **Use TLS Client Auth**, `{clientCert:true, clientKey:true, rootCert:false}`; after **With CA Cert**,
`{clientCert:true, clientKey:true, rootCert:true}` — exactly the schema's `dependsOn` mapping
(`tlsClientCert`/`tlsClientKey` on `tlsAuth`; `tlsCACert` on `tlsAuthWithCACert`).

Legacy observation: toggling the same two switches reveals **3 textareas** with placeholders
`-----BEGIN CERTIFICATE-----`, `-----BEGIN CERTIFICATE-----`, `-----BEGIN RSA PRIVATE KEY-----`
(Client Cert, Root Cert, Client Key) — the same three fields, same reveal behaviour.

**Effects:** mysql's schema contains **no** `effects` blocks. `tlsAuth` and `tlsAuthWithCACert` are
independent direct toggles (enable either, both, or neither), not a virtual discriminator that fans out
to multiple fields, so there is nothing for `effects` to model and none were added.

---

## `required: true` / General-step fix

Changing `root_url` and `root_user` from `requiredWhen: "true"` to `required: true`:

1. **Wizard:** seeds the synthetic **General** step so URL + User (the two unconditionally-required
   fields) appear first — verified `General (1/5)` with both inputs and 2 required markers.
2. **Generated spec:** the settings spec now carries `required: ["url","user"]` at the top level instead
   of the `x-dsconfig-required-when: "true"` property extension (diff below), matching how graphite's
   `url` was handled.

```diff
     "spec": {
       "type": "object",
+      "required": [
+        "url",
+        "user"
+      ],
       "properties": {
         ...
         "url": {
-          "type": "string",
-          "x-dsconfig-required-when": "true"
+          "type": "string"
         },
         "user": {
-          "type": "string",
-          "x-dsconfig-required-when": "true"
+          "type": "string"
         }
```

This is faithful to the backend: `Config.Validate()` (`settings.go:181-186`) errors when `url` or `user`
is empty, i.e. they are truly required, not conditionally required.

---

## Verification

```
go generate ./registry/mysql/...     # regenerate schema.gen.json / settings.gen.json / settings.examples.gen.json
go test ./registry/mysql/...         # 8/8 conformance subtests PASS
go test ./registry/... ./schema/...  # entire suite PASS (no regressions)
```

Conformance subtests (`TestSchemaConformance`, mysql): `BaseFieldsResolved`, `SchemaRoundTrip`,
`SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`, `JSONDataMatchesStruct`,
`JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings` — all **PASS**.

Playwright evidence (in the scratch capture dir):

- `legacy-mysql-fields.{json,png}` / `legacy-expand-mysql-legacy.{json,png}` — legacy inventory (3 sections; no headers/upload/file inputs).
- `newgen-mysql-tab.{json,png}` — new UI tab mode (`hasHeadersEditor:false`, `urlPresent:true`, all sections render).
- `newgen-mysql-wizard.{json,png}` + `mysql-verify-wizard.png` — General step 1/5 with Host URL + Username (required).
- `mysql-newtab-full.{json,png}` — full field inventory with the optional section expanded.
- `mysql-verify-tls.{json,png}` (new UI) + `mysql-legacy-tls.{json,png}` (legacy) — TLS cert reveal in both UIs.

---

## Files changed

- [`registry/mysql/dsconfig.json`](dsconfig.json) — `root_url` and `root_user` changed from
  `requiredWhen: "true"` to `required: true` (so they render in the wizard's General step). No other
  edits: no Custom HTTP Headers (SQL datasource), no `fileUpload`, no packs; all real
  `requiredWhen`/`dependsOn` TLS conditions left intact.
- [`registry/mysql/schema.gen.json`](schema.gen.json), [`registry/mysql/settings.gen.json`](settings.gen.json) —
  regenerated by `go generate` (`url`+`user` now in the spec's `required` array; `x-dsconfig-required-when`
  removed).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `schema/conformance.go`, `plugin-ui`.
`settings.examples.gen.json` was re-run by `go generate` but is byte-identical (examples don't encode the
`required` array).

_Not fixable via `dsconfig.json` alone:_ **nothing** — mysql needed neither the `conformance.go`
`indexedPair` skip (no `indexedPair` field added) nor a `plugin-ui` change (the `authentication`-group
recognition landed with the earlier graphite work and already folds auth fields into General).
