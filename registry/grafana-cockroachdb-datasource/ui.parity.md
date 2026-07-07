# CockroachDB — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-cockroachdb-datasource` (a **SQL** datasource — PostgreSQL wire protocol, CockroachDB default port `26257`)
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/cfrbqiwv5nym8f` (Grafana Enterprise 13.x, CockroachDB `ConfigEditor` built on `@grafana/plugin-ui` SQL components)
- **New UI:** `http://192.168.1.241:58899/iframe.html?args=pluginType:grafana-cockroachdb-datasource&id=configeditor-datasourceconfigwizard--tab` (Storybook, `ConfigEditor/DatasourceConfigWizard`; also `--wizard`)
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + step/section/conditional probing). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/grafana-cockroachdb-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the screenshots reflect the local schema without pushing.
- **Result:** **Parity achieved.** All 21 modeled fields are present in both UIs and route to identical storage targets. **No missing fields were found** (unlike graphite, no field had to be added). CockroachDB is a SQL datasource, so it correctly has **no HTTP-headers editor** and **no file-upload** control. The one change required was making the three unconditionally-required fields use `required: true` so the wizard's **General** step pulls them in. All `authType` / SSL-mode / TLS-method conditionals were exercised and confirmed to reveal the right cert fields.

---

## TL;DR of changes

| #   | Change | File | Why |
| --- | --- | --- | --- |
| 1   | Changed **`jsonData_url`**, **`jsonData_database`**, **`jsonData_user`** from `"requiredWhen": "true"` → `"required": true` | [`dsconfig.json`](dsconfig.json) | These are unconditionally required (the backend `Validate()` rejects empty URL/user/database — `settings.go:231-239`). The wizard's synthetic **General** step only pulls fields with `required: true`; `requiredWhen` is a CEL expression the resolver does not inspect. Also emits proper OpenAPI `required` arrays instead of the `x-dsconfig-required-when` extension. |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-cockroachdb-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`) |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance_test.go`, or `plugin-ui`.
The `requiredWhen` conditions that are **real** conditions were **left untouched** — only the
three literal `"requiredWhen": "true"` values were converted. The real conditionals kept as-is:

- `secureJsonData_password` — `jsonData_authType == 'SQL Authentication' || jsonData_authType == 'TLS/SSL Authentication'`
- `jsonData_credentialCache`, `jsonData_configFilePath` — `jsonData_authType == 'Kerberos Authentication'`
- `jsonData_sslRootCertFile` / `jsonData_sslCertFile` / `jsonData_sslKeyFile` (file-path) and
  `secureJsonData_tlsCACert` / `secureJsonData_tlsClientCert` / `secureJsonData_tlsClientKey`
  (file-content) — `jsonData_authType == 'TLS/SSL Authentication' && jsonData_sslmode != 'disable' && jsonData_tlsConfigurationMethod == '<method>'`

No `conformance_test.go` change was needed (this plugin models no `indexedPair`/HTTP-header field).
No `plugin-ui` change was needed either: the auth group uses the conventional id
`authentication`, which the wizard's required-fields resolver already recognises, so the auth
fields fold into General correctly.

**Note (differs from postgres):** here `url`, `database` **and** `user` all target `jsonData`
(there are no root fields — the backend reads every connection field from `config.JSONData`,
`settings.go:100-125`, and `conformance_test.go` documents this). So after regeneration all three
land in the **`jsonData` `required` array** (postgres split them across the root and jsonData
required arrays because its `url`/`user` are root fields).

---

## Section layout

Verified rendering top-to-bottom in the new UI (tab mode) and matched to the legacy editor's
section headings.

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Connection** (`connection`) | no | Host URL, Database |
| 2 | **Authentication** (`authentication`) | no | authType, User, Password, Credential cache path, Kerberos server name, TLS/SSL Method |
| 3 | **TLS/SSL Auth Details** (`tls-details`) | yes | (file-path) Root Certificate → Client Certificate → Client Key; (file-content) Root Certificate → Client Certificate → Client Key |
| 4 | **Additional settings** (`additional-settings`) | yes | Max open, Auto max idle, Max idle, Max lifetime, Query timeout, krb5 config file path, TLS/SSL Mode |

New UI (tab) section buttons captured: `Connection`, `Authentication`, `TLS/SSL Auth Details`
(Optional), `Additional settings` (Optional). The legacy editor shows the `h*` headings
**Connection**, **Authentication**, **Additional settings** (captured: `["Connection","Authentication","Additional settings"]`).

**One benign grouping difference (not a gap):** the legacy editor has **no dedicated "TLS/SSL
Auth Details" heading** — its cert fields render *inline* (with no wrapping `ConfigSection`)
between Authentication and Additional settings, and only when `authType = 'TLS/SSL Authentication'`
and `sslmode != 'disable'`. The schema buckets those same fields into an optional
`TLS/SSL Auth Details` accordion so provisioning/UX consumers get a coherent group. Both UIs
collect the identical fields into the identical storage keys. (This is documented as a modeling
decision in `README.md`.)

### Wizard mode: the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) every
field in the auth group (`authentication`), plus their `dependsOn` parents/children. The wizard
renders **5 steps** (`1/5` = General, then the four sections above).

**Effect of the `required: true` fix (before/after, both captured via route interception of the
pre-edit vs edited schema):**

| Field | Before (`requiredWhen: "true"`) | After (`required: true`) |
| --- | --- | --- |
| Host URL (`jsonData_url`, Connection group) | **absent** from General | **present** in General ✅ |
| Database (`jsonData_database`, Connection group) | **absent** from General | **present** in General ✅ |
| User (`jsonData_user`, Authentication group) | present (via auth-group membership) | present ✅ |

Captured probe (`verify-cockroachdb-wizard.js` / `verify-cockroachdb-extra.js`), General step
with `authType = 'SQL Authentication'` selected:

- **Before:** `hostUrl:false, database:false, user:true, password:true` — Host URL and Database
  were missing (they carried only `requiredWhen:"true"` and live in the `connection` group, not
  `authentication`).
- **After:** `hostUrl:true, database:true, user:true, password:true` — all three required fields
  appear.

`jsonData_user` keeps its `dependsOn: "jsonData_authType != ''"`, so it is `required: true`
(backend contract) **and** conditionally visible (editor behaviour): on a fresh load with
`authType` empty, Host URL + Database + the authType select show immediately, and **User**
appears once an auth type is chosen (its `dependsOn` parent, `authType`, is itself in the General
step). Tab mode is unaffected — the synthetic `_required` group is filtered out there, so it
still shows the four sections in order.

---

## Field-by-field parity

Legend: ✅ present & matching · 🔀 conditional (revealed by `dependsOn`)

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| Host URL \* | text input | `jsonData_url` | input | `jsonData.url` | ✅ |
| Database \* | text input | `jsonData_database` | input | `jsonData.database` | ✅ |
| authType | select (SQL / Kerberos / TLS-SSL) | `jsonData_authType` | select | `jsonData.authType` | ✅ |
| User \* | text input | `jsonData_user` | input | `jsonData.user` | ✅ 🔀 |
| Password | password (`SecretInput`) | `secureJsonData_password` | secure input | `secureJsonData.password` | ✅ 🔀 |
| Credential cache path | text input | `jsonData_credentialCache` | input | `jsonData.credentialCache` | ✅ 🔀 |
| Kerberos server name | text input | `jsonData_kerberosServerName` | input | `jsonData.kerberosServerName` | ✅ 🔀 |
| TLS/SSL Method | select (file-content / file-path) | `jsonData_tlsConfigurationMethod` | select | `jsonData.tlsConfigurationMethod` | ✅ 🔀 |
| TLS/SSL Root Certificate (path) | text input | `jsonData_sslRootCertFile` | input | `jsonData.sslRootCertFile` | ✅ 🔀 |
| TLS/SSL Client Certificate (path) | text input | `jsonData_sslCertFile` | input | `jsonData.sslCertFile` | ✅ 🔀 |
| TLS/SSL Client Key (path) | text input | `jsonData_sslKeyFile` | input | `jsonData.sslKeyFile` | ✅ 🔀 |
| TLS/SSL Root Certificate (content) | **textarea** (`SecretTextArea`) | `secureJsonData_tlsCACert` | secure input¹ | `secureJsonData.tlsCACert` | ✅ 🔀 |
| TLS/SSL Client Certificate (content) | **textarea** (`SecretTextArea`) | `secureJsonData_tlsClientCert` | secure input¹ | `secureJsonData.tlsClientCert` | ✅ 🔀 |
| TLS/SSL Client Key (content) | **textarea** (`SecretTextArea`) | `secureJsonData_tlsClientKey` | secure input¹ | `secureJsonData.tlsClientKey` | ✅ 🔀 |
| Max open | number | `jsonData_maxOpenConns` | number | `jsonData.maxOpenConns` | ✅ |
| Auto max idle | switch | `jsonData_maxIdleConnsAuto` | switch | `jsonData.maxIdleConnsAuto` | ✅ |
| Max idle | number | `jsonData_maxIdleConns` | number | `jsonData.maxIdleConns` | ✅ |
| Max lifetime | number | `jsonData_connMaxLifetime` | number | `jsonData.connMaxLifetime` | ✅ |
| Query timeout | number | `jsonData_queryTimeout` | number | `jsonData.queryTimeout` | ✅ |
| krb5 config file path | text input | `jsonData_configFilePath` | input | `jsonData.configFilePath` | ✅ 🔀 |
| TLS/SSL Mode | select (disable/require/verify-ca/verify-full) | `jsonData_sslmode` | select | `jsonData.sslmode` | ✅ 🔀 |

All 21 modeled fields render in the new UI and were located across the four sections. The legacy
`Name` and `Default` controls at the top are Grafana editor chrome (datasource name + default
toggle), not part of the datasource config, and are correctly **not** modeled in `dsconfig.json`.
Editor-only/excluded fields (`enableSecureSocksProxy`, and the vestigial
`postgresVersion`/`timescaledb` inherited from the PostgreSQL type) are intentionally not modeled
(see `README.md` → "Frontend-only and backend-only settings").

¹ **Not a discrepancy.** The `dsconfig.json` declares `ui.component: "textarea"` for the three
`file-content` TLS cert fields, but the new renderer checks `target === "secureJsonData"`
_before_ the `textarea` branch, so any secure field is drawn as a masked secure input
(`input[type="password"]`) with a show/hide toggle. Verified directly
(`verify-cockroachdb-extra.js`): `certPasswordInputs:2, certTextareas:0, keyPasswordInputs:1,
keyTextareas:0` — the three fields render as password inputs carrying the PEM placeholders
(`-----BEGIN CERTIFICATE-----` ×2, `-----BEGIN RSA PRIVATE KEY-----` ×1). Both UIs collect the
same PEM text into the same `secureJsonData` keys; only the widget affordance differs. This is a
renderer policy in `plugin-ui`, not a schema gap (same footnote as postgres's TLS cert fields).

---

## Gaps found

**None beyond the required-field fix.** The complete legacy field set is already modeled in
`dsconfig.json`, so — unlike graphite (which was missing Custom HTTP Headers) — no field had to be
added. CockroachDB is a SQL datasource and its schema keeps all fields **inline** (no `packs`),
which matches how the legacy SQL editor lays them out.

### Custom HTTP Headers — not applicable (verified)

CockroachDB talks the PostgreSQL wire protocol over TCP, **not HTTP**, so it has no HTTP-headers
concept. Legacy DOM capture (`capture-legacy-expand.js` on `cfrbqiwv5nym8f`):
`hasCustomHeaders: false`, `addHeaderBtn: false`. New UI (tab + wizard):
`hasHeadersEditor: false`. Correctly **not** added.

---

## `fileUpload` evaluation — not applicable to cockroachdb

The task noted the legacy UI has no file upload. Confirmed:

- The legacy CockroachDB editor renders the file-path TLS fields (Root/Client Certificate,
  Client Key) as **plain text inputs** (paths on the Grafana host) and the file-content TLS
  fields as **textareas** (`SecretTextArea`, inline PEM). No `<input type="file">` and no upload
  button were found in the legacy DOM (`fileInputs: 0`, `uploadButtons: []`).
- The new UI's `fileUpload` component only activates when a field declares `ui.fileMapping`
  (multi-key JSON distribution, e.g. a GCP service-account file); it does not model single-PEM or
  path-string upload.

**Decision:** do **not** add `fileUpload` to any cockroachdb field. The cert fields keep their
current modeling; both UIs collect the same values into the same `jsonData` / `secureJsonData`
keys.

---

## Conditional fields & effects — tested

CockroachDB drives its conditionals with a top-level **`authType` discriminator** plus two
`select` dropdowns (`sslmode`, `tlsConfigurationMethod`). Unlike postgres (which has no auth
discriminator and always shows `sslmode`), **all** TLS surface here is gated first by
`authType == 'TLS/SSL Authentication'`. Each scenario was run on a fresh page
(`verify-cockroachdb-conditionals.js`), the `Additional settings` accordion expanded to reach
`sslmode`, then the `TLS/SSL Auth Details` accordion expanded and the visible cert fields probed
(secure inputs detected as `input[type="password"]`):

| Scenario (authType, sslmode, method) | sslmode select | method select | Root/Client/Key **path** | Root/Client/Key **content** | Matches schema `dependsOn`? |
| --- | --- | --- | --- | --- | --- |
| **A** TLS, `require` (default), `file-content` (default) | shown | shown | hidden | **shown** (2× CERT + 1× KEY) | ✅ file-content revealed by default method |
| **B** TLS, `disable` | shown | **hidden** | hidden | hidden | ✅ method + all certs gated off by `sslmode != 'disable'` |
| **C** TLS, `verify-full`, `file-path` | shown | shown | **shown** (all 3) | hidden | ✅ all three file-path fields appear together |
| **D** TLS, `verify-full`, `file-content` | shown | shown | hidden | **shown** (2× CERT + 1× KEY) | ✅ path inputs vanish; the three `secureJsonData` PEM fields appear |
| **E** `SQL Authentication` | **hidden** | **hidden** | hidden | hidden | ✅ entire TLS surface gated off by `authType != 'TLS/SSL Authentication'` |

Observed transitions, exactly matching the schema:

- Choosing an **`authType`** other than TLS (scenario **E**, SQL) hides `sslmode`,
  `tlsConfigurationMethod` and every cert field (they all begin with
  `jsonData_authType == 'TLS/SSL Authentication'`).
- Under TLS auth, selecting **`sslmode = disable`** (scenario **B**) hides the **TLS/SSL Method**
  select and every cert field (`jsonData_sslmode != 'disable'`).
- Flipping **TLS/SSL Method** from `file-content` (default) → `file-path` swaps the inline PEM
  secure inputs (`tlsCACert` / `tlsClientCert` / `tlsClientKey`, `secureJsonData`) for the
  file-path text inputs (`sslRootCertFile` / `sslCertFile` / `sslKeyFile`, `jsonData`) — the two
  mutually-exclusive `tlsConfigurationMethod == 'file-content'` vs `'file-path'` branches.

**Difference vs postgres (verified):** postgres gates its **root** cert field with an extra
`(sslmode == 'verify-ca' || sslmode == 'verify-full')` term, so the root cert hides under
`require`. CockroachDB's cert `dependsOn` expressions carry **no** such term — so all three
file-path (or all three file-content) fields appear **together** whenever `sslmode != 'disable'`,
regardless of verify level. Scenario **C** confirms root+client+key all shown at `verify-full`
with `file-path`; this matches the backend's `IsValidFilePathTLS` / `IsValidFileContentTLS`,
which require all three together (`tlsmanager.go:25-50`).

Also confirmed: `sslmode` defaults to `require` and `tlsConfigurationMethod` defaults to
`file-content` (scenario **A** shows both), which is deliberately different from postgres's
`file-path` default (`README.md` → Modeling decisions).

**Effects:** cockroachdb's schema contains **no** `effects` blocks. Its visibility is a set of
plain `dependsOn` CEL expressions over `authType` + the two selects; there is no virtual selector
that fans out to write multiple fields, so nothing for `effects` to model, and none were added.

---

## Verification

```
go generate ./registry/grafana-cockroachdb-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-cockroachdb-datasource/...        # 8/8 conformance subtests PASS (+ settings_test.go)
go test ./registry/... ./schema/...                          # entire suite PASS (no regressions)
```

Conformance subtests (cockroachdb): `BaseFieldsResolved`, `SchemaRoundTrip`,
`SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`,
`JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings` — all
**PASS**.

After regeneration, `settings.gen.json` added `url`, `database`, `user` to the `jsonData`
`required` array and dropped the three `x-dsconfig-required-when: "true"` extensions;
`jsonData_user` correctly retains its `x-dsconfig-depends-on: "jsonData_authType != ''"`.

---

## Files changed

- [`registry/grafana-cockroachdb-datasource/dsconfig.json`](dsconfig.json) — changed
  `jsonData_url`, `jsonData_database`, and `jsonData_user` from `"requiredWhen": "true"` to
  `"required": true` (so they render in the wizard's General step and emit OpenAPI `required`).
  The real `dependsOn` / conditional `requiredWhen` expressions (auth-, Kerberos-, and
  TLS-gated) were left untouched.
- [`registry/grafana-cockroachdb-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-cockroachdb-datasource/settings.gen.json`](settings.gen.json) — regenerated
  by `go generate` (`url`/`database`/`user` now in the spec `required` array).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema.go`, `conformance_test.go`, and `plugin-ui`.
