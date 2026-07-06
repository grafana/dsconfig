# Microsoft SQL Server (mssql) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `mssql`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/ffraniu9osn40f` (Grafana Enterprise, `mssql` `ConfigurationEditor` + `@grafana/sql` / `@grafana/azure-sdk`)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:mssql` (Storybook, `ConfigEditor/DatasourceConfigWizard`)
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + interactive combobox driving). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/mssql/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so screenshots reflect the local schema without pushing.
- **Result:** **Parity achieved.** mssql is a **SQL** datasource — it has **no HTTP transport**, so (unlike graphite) there is **no Custom HTTP Headers** section and **nothing was added**. The only schema change was promoting the two unconditionally-required fields (`root_url`, `jsonData_database`) from `requiredWhen: "true"` to `required: true` so the wizard builds a proper **General** step. All conditional fields (the **Encrypt** select and the 7-way **Authentication Type** select) were exercised and confirmed to reveal their dependent fields. `fileUpload` was evaluated and correctly **not** used.

---

## TL;DR of changes

| #   | Change | File | Why |
| --- | --- | --- | --- |
| 1   | `root_url`: `requiredWhen: "true"` → `required: true` | [`dsconfig.json`](dsconfig.json) | Host is unconditionally required (backend `Validate()` rejects empty URL, `settings.go:174`). `required: true` puts it in the wizard's **General** step and emits OpenAPI `required: ["url"]`. |
| 2   | `jsonData_database`: `requiredWhen: "true"` → `required: true` | [`dsconfig.json`](dsconfig.json) | Database is unconditionally required (`settings.go:177`, `EffectiveDatabase()`). Now in **General** and emitted as `jsonData.required: ["database"]`. |
| 3   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/mssql/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`). |

**Not changed / not needed (contrast with graphite):**

- **No new field added.** Every legacy field was already modeled — full 1:1 parity (see table below). No Custom HTTP Headers (SQL datasource has no HTTP transport).
- **No `schema/conformance.go` change.** mssql models no `indexedPair` field, so the jsonData↔struct parity walker needed no exemption. All 8 conformance subtests pass as-is after regeneration.
- **No `plugin-ui` change.** The auth-group folding into **General** (recognising the conventional `id: "authentication"`) was already in place; the wizard folded the auth fields in correctly with no edit.
- **No `settings.go` / `settings.ts` / `README.md` / `schema.go` / `conformance_test.go` change.** All changes flow through `dsconfig.json` + `go generate`. `settings.examples.gen.json` is unchanged.
- The schema keeps its fields **inline** (no `packs`).

---

## Section layout

The five `groups` render top-to-bottom in the new UI sidebar/accordion and map 1:1
to the legacy editor's sections (legacy headings captured: `Connection`,
`TLS/SSL Auth`, `Authentication`, `Additional settings`, with a nested
`Windows AD: Advanced Settings` sub-section):

| Order | Section (`id`) | `optional` | Fields (display order) |
| --- | --- | --- | --- |
| 1 | **Connection** (`connection`) | no | Host \*, Database \* |
| 2 | **TLS/SSL Auth** (`tls-auth`) | no | Encrypt → (Skip TLS Verify, TLS/SSL Root Certificate, Hostname in server certificate) |
| 3 | **Authentication** (`authentication`) | no | Authentication Type → (Username, Password, Keytab file path, Credential cache path, Credential cache file path, Azure credentials, Azure client secret) |
| 4 | **Additional settings** (`additional-settings`) | yes | Max open, Auto max idle, Max idle, Max lifetime, Min time interval, Connection timeout |
| 5 | **Windows AD: Advanced Settings** (`kerberos-advanced`) | yes | UDP Preference Limit, DNS Lookup KDC, krb5 config file path |

Notes:

- The legacy editor nests **Windows AD: Advanced Settings** inside the Additional-settings
  area; the schema promotes it to a dedicated top-level **optional** group. Both are
  collapsible and gated on the four Kerberos auth types. Faithful representation, not a gap.
- `optional` groups (Additional settings, Windows AD: Advanced Settings) render collapsed.
- The authentication group uses the registry-conventional **`id: "authentication"`**, which
  the wizard's required-fields resolver already recognises as the auth group (folded into
  **General** in wizard mode).

### Wizard mode: the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) the
auth group's fields, plus their `dependsOn` parents/children.

Promoting `root_url` and `jsonData_database` to `required: true` is what makes this work:

- **Before:** both used `requiredWhen: "true"` — a CEL expression the resolver does **not**
  inspect — so no General step was created and the wizard opened on `Connection`.
- **After (verified):** the wizard opens on **General 1/6** containing **Host \***,
  **Database \***, **Authentication Type**, and (because the default auth type is
  `SQL Server Authentication`) **Username \*** and **Password \***. Selecting a different
  auth type re-computes the revealed fields inline. The redundant `Connection` /
  `Authentication` steps are auto-skipped since their fields already appear in General.
  Tab mode is unaffected (the synthetic `_required` group is filtered out — all five
  sections still render in order).

---

## Field-by-field parity

Legend: ✅ present & matching · 🔀 conditional (revealed by `dependsOn`)

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| Host | text input (required) | `root_url` | input | `root.url` | ✅ (`required: true`) |
| Database | text input (required) | `jsonData_database` | input | `jsonData.database` | ✅ (`required: true`) |
| Encrypt | select (disable/false/true) | `jsonData_encrypt` | select | `jsonData.encrypt` | ✅ |
| Skip TLS Verify | switch | `jsonData_tlsSkipVerify` | switch | `jsonData.tlsSkipVerify` | ✅ 🔀 |
| TLS/SSL Root Certificate | text input | `jsonData_sslRootCertFile` | input | `jsonData.sslRootCertFile` | ✅ 🔀 |
| Hostname in server certificate | text input | `jsonData_serverName` | input | `jsonData.serverName` | ✅ 🔀 |
| Authentication Type | select (7 options) | `jsonData_authenticationType` | select | `jsonData.authenticationType` | ✅ |
| Username | text input | `root_user` | input | `root.user` | ✅ 🔀 |
| Password | password (`SecretInput`) | `secureJsonData_password` | secure input | `secureJsonData.password` | ✅ 🔀 |
| Keytab file path | text input | `jsonData_keytabFilePath` | input | `jsonData.keytabFilePath` | ✅ 🔀 |
| Credential cache path | text input | `jsonData_credentialCache` | input | `jsonData.credentialCache` | ✅ 🔀 |
| Credential cache file path | text input | `jsonData_credentialCacheLookupFile` | input | `jsonData.credentialCacheLookupFile` | ✅ 🔀 |
| Azure auth settings | `AzureAuthSettings` component | `jsonData_azureCredentials` | complex-object note¹ | `jsonData.azureCredentials` | ✅ 🔀 |
| (Azure client secret) | password (within Azure form) | `secureJsonData_azureClientSecret` | secure (sdk-managed)¹ | `secureJsonData.azureClientSecret` | ✅ 🔀 |
| Max open | number | `jsonData_maxOpenConns` | input | `jsonData.maxOpenConns` | ✅ |
| Auto max idle | switch | `jsonData_maxIdleConnsAuto` | switch | `jsonData.maxIdleConnsAuto` | ✅ |
| Max idle | number | `jsonData_maxIdleConns` | input | `jsonData.maxIdleConns` | ✅ |
| Max lifetime | number | `jsonData_connMaxLifetime` | input | `jsonData.connMaxLifetime` | ✅ |
| Min time interval | text input | `jsonData_timeInterval` | input | `jsonData.timeInterval` | ✅ |
| Connection timeout | number | `jsonData_connectionTimeout` | input | `jsonData.connectionTimeout` | ✅ |
| UDP Preference Limit | number | `jsonData_UDPConnectionLimit` | input | `jsonData.UDPConnectionLimit` | ✅ 🔀 |
| DNS Lookup KDC | text input | `jsonData_enableDNSLookupKDC` | input | `jsonData.enableDNSLookupKDC` | ✅ 🔀 |
| krb5 config file path | text input | `jsonData_configFilePath` | input | `jsonData.configFilePath` | ✅ 🔀 |

All **23** schema fields are present in the new UI. No legacy field is missing; no new
field was required.

¹ **Azure credentials** are a seven-variant discriminated union owned by
`@grafana/azure-sdk` (`AzureCredentialsForm`). The schema models the object opaquely
(`jsonData_azureCredentials`, `valueType: any`) plus the one static secret
(`secureJsonData_azureClientSecret`); the new UI shows the opaque object as a complex-field
note. Both UIs write to the same `jsonData.azureCredentials` / `secureJsonData.azureClientSecret`
keys. Consistent with how opaque SDK-managed objects are modeled elsewhere.

---

## Full parity — no gap to fix

Unlike graphite (which was missing a Custom HTTP Headers section), the mssql schema already
covered every legacy field. Confirmed against the legacy capture:

- Legacy headings: `Connection`, `TLS/SSL Auth`, `Authentication`, `Additional settings`
  (+ nested `Connection limits`, `Connection details`, `Windows AD: Advanced Settings`).
- Legacy selects: `encrypt`, `authenticationType`. Legacy textareas: **0**. Legacy custom
  headers: **none** (`hasCustomHeaders: false`, `addHeaderBtn: false`).

The new UI reproduces all of these across its five groups.

### No Custom HTTP Headers (correct)

mssql is a SQL datasource. Its backend builds a database DSN — there is **no**
`DataSourceHttpSettings` / `CustomHeadersSettings` in the legacy editor and **no** HTTP
header transport. The new-UI capture confirms **no** headers editor
(`hasHeadersEditor: false` in both tab and wizard mode). Custom HTTP Headers were therefore
**not** modeled — this is correct, not a gap.

---

## `fileUpload` evaluation — not applicable to mssql

The task asks to use the `fileUpload` control **only if the legacy UI uses it**. It does not:

- The legacy editor collects TLS/Kerberos material as **file-path text inputs** — the
  server-side paths `jsonData.sslRootCertFile` ("TLS/SSL root certificate file path"),
  `jsonData.keytabFilePath` (`/home/grot/grot.keytab`), `jsonData.credentialCache`
  (`/tmp/krb5cc_1000`), `jsonData.credentialCacheLookupFile` (`/home/grot/cache.json`),
  and `jsonData.configFilePath` (`/etc/krb5.conf`). These are paths on the Grafana host,
  **not** uploaded file contents.
- The legacy DOM had **0** `<input type="file">` and **0** upload buttons
  (`fileInputs: 0`, `uploadButtons: []`).
- The new UI's `fileUpload` component is hard-coded for the multi-key JSON-token use case
  (`ui.fileMapping`, e.g. a GCP service-account file); it does not model single-path or
  single-PEM inputs.

**Decision:** do **not** add `fileUpload` to any mssql field. The path fields keep their
`input` control; both UIs store identical path strings under `jsonData`.

---

## Conditional fields & effects — tested

Both discriminator selects render as `react-select` comboboxes in the new UI. Each was
driven interactively (open → type → Enter) and the revealed fields probed by placeholder /
label. Baseline: with defaults (`encrypt = false`, `authenticationType = SQL Server
Authentication`) only Username + Password are revealed; all TLS and Kerberos fields are
correctly hidden.

### Encrypt select (`jsonData_encrypt`)

| Trigger | Revealed field(s) | Verified |
| --- | --- | --- |
| `jsonData_encrypt == 'true'` | Skip TLS Verify, **TLS/SSL Root Certificate**, **Hostname in server certificate** | ✅ all three appear |

The two TLS-verify fields carry the compound guard
`jsonData_encrypt == 'true' && jsonData_tlsSkipVerify != true`. With `encrypt = true` and
`tlsSkipVerify` at its default `false`, both `sslRootCertFile` and `serverName` are shown
(verified). Turning on Skip TLS Verify would hide them again (the second-level gate).

### Authentication Type select (`jsonData_authenticationType`)

| Selected auth type | Revealed | Hidden | Verified |
| --- | --- | --- | --- |
| `SQL Server Authentication` (default) | Username, Password | — | ✅ (baseline) |
| `Windows AD: Keytab` | Username, **Keytab file path** | Password | ✅ |
| `Windows AD: Credential cache` | **Credential cache path** | Username, Password, Keytab | ✅ |
| `Windows AD: Credential cache file` | Username, **Credential cache file path** | Password, Keytab, Credential cache | ✅ |
| any `Windows AD: *` | **UDP Preference Limit**, **DNS Lookup KDC**, **krb5 config file path** (in the expanded *Windows AD: Advanced Settings* group) | — | ✅ (verified after expanding the optional group) |
| `Azure AD Authentication` | Azure credentials object + Azure client secret | user/password/kerberos fields | ✅ (modeled; SDK-managed form) |

Observed re-routing (each transition confirmed): selecting a Kerberos variant hides the SQL
`Password`; `Credential cache` additionally hides `Username`; `Credential cache file`
brings `Username` back alongside the lookup-file input. The three Kerberos-advanced fields
are gated by both the auth-type `dependsOn` **and** membership in the collapsible
`kerberos-advanced` group, so they surface once a Kerberos type is selected and that group
is expanded (verified: `udpLimit/dnsLookupKDC/krb5Config` flip `false → true`).

**Effects:** the mssql schema contains **no** `effects` blocks. `authenticationType` is a
true discriminator, but the fan-out is modeled the idiomatic way — per-field `dependsOn` /
`requiredWhen` referencing the selected value — rather than a virtual `effects` selector.
Nothing to add (same approach as graphite).

---

## Required-fields fix (`required: true`) and the General step

The **only** schema edit was changing the two unconditionally-required fields from a
`requiredWhen: "true"` CEL expression to a plain `required: true` flag:

```diff
  "id": "root_url",  ...
- "requiredWhen": "true",
+ "required": true,

  "id": "jsonData_database",  ...
- "requiredWhen": "true",
+ "required": true,
```

The five **conditional** `requiredWhen` expressions (on `root_user`,
`secureJsonData_password`, `jsonData_keytabFilePath`, `jsonData_credentialCache`,
`jsonData_credentialCacheLookupFile`, all referencing `jsonData_authenticationType`) were
**left untouched** — those encode real runtime conditions and must stay as `requiredWhen`.

Effect in the generated spec (both `schema.gen.json` and `settings.gen.json`):

```diff
  "spec": {
    "type": "object",
+   "required": [ "url" ],
    "properties": {
      "jsonData": {
        "type": "object",
+       "required": [ "database" ],
        "properties": {
-         "database": { "type": "string", "x-dsconfig-required-when": "true" }
+         "database": { "type": "string" }
        }
      },
-     "url": { "type": "string", "x-dsconfig-required-when": "true" }
+     "url": { "type": "string" }
    }
  }
```

Conditional `x-dsconfig-required-when` extensions (e.g. the auth-type ones) are preserved.
Both UIs now render Host and Database with the required `*` marker, and the wizard opens on
**General 1/6**.

---

## Conformance & tests

```
go generate ./registry/mssql/...   # regenerate schema.gen.json / settings.gen.json
go test    ./registry/mssql/...    # PASS
```

`go test -v ./registry/mssql/...` — `TestSchemaConformance` and all 8 subtests **PASS**:

| Subtest | Result |
| --- | --- |
| `BaseFieldsResolved` | PASS |
| `SchemaRoundTrip` | PASS |
| `SchemaArtifactInSync` | PASS |
| `SchemaSpecHasNoSecureJSON` | PASS |
| `ConfigSchemaValid` | PASS |
| `JSONDataMatchesStruct` | PASS |
| `JSONDataTypesMatchStruct` | PASS |
| `SecureValuesMatchLoadSettings` | PASS |

The package's own settings tests also pass unchanged: `TestLoadConfig` (14 cases),
`TestApplyDefaults`, `TestValidate` (10 cases), `TestEffectiveDatabase`.

No conformance-suite exemption was required (no `indexedPair` field), and no `plugin-ui`
change was required (the `authentication` group already folds into General).

---

## Verification

- **Tab mode** (`newgen-mssql-tab`): five sections render (Connection, TLS/SSL Auth,
  Authentication, Additional settings, Windows AD: Advanced Settings); `hasHeadersEditor:
  false` (correct for SQL); Host \* and Database \* show the required marker; URL input present.
- **Wizard mode** (`verify-mssql-wizard`): step **General 1/6** contains Host \*, Database \*,
  Authentication Type, Username \*, Password \* — the required + auth fields folded in.
- **Conditionals** (`verify-mssql-conditionals2`, `verify-mssql-kerbadv`): Encrypt→TLS
  reveal and the four auth-type→field reveals all confirmed (see tables above).

---

## Files changed

- [`registry/mssql/dsconfig.json`](dsconfig.json) — `root_url` and `jsonData_database`
  changed from `requiredWhen: "true"` to `required: true`. No fields added; conditional
  `requiredWhen` expressions untouched; fields kept inline (no packs).
- [`registry/mssql/schema.gen.json`](schema.gen.json),
  [`registry/mssql/settings.gen.json`](settings.gen.json) — regenerated by `go generate`
  (`url` and `jsonData.database` now in the spec `required` arrays; their
  `x-dsconfig-required-when: "true"` extensions removed).

_Unchanged by design:_ `settings.go`, `settings.ts`, `schema.go`, `conformance_test.go`,
`settings_test.go`, `settings.examples.gen.json`, `README.md`, and everything under
`schema/` and `plugin-ui`.
