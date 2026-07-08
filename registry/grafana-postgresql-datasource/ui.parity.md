# PostgreSQL — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-postgresql-datasource` (a **SQL** datasource; `plugin.json` also declares the legacy alias id `postgres`)
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/dfranib2bedq8f` (Grafana Enterprise 13.x, `@grafana/sql` `DataSourceConfig` / PostgreSQL `ConfigurationEditor`)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-postgresql-datasource` (Storybook, `ConfigEditor/DatasourceConfigWizard`; also `--wizard`)
- **Method:** Playwright captured both UIs (full-page screenshots + DOM extraction + step/section probing). The Storybook story fetches the schema from `raw.githubusercontent.com/.../registry/grafana-postgresql-datasource/dsconfig.json`; the local (edited) `dsconfig.json` was served to the new UI by intercepting that request with `context.route(...)`, so the screenshots reflect the local schema without pushing.
- **Result:** **Parity achieved.** All 17 modeled fields are present in both UIs and route to identical storage targets. **No missing fields were found** (unlike graphite, no field had to be added). Postgres is a SQL datasource, so it correctly has **no HTTP-headers editor** and **no file-upload** control. The one change required was making the three unconditionally-required fields use `required: true` so the wizard's **General** step pulls them in. All SSL-mode / TLS-method conditionals were exercised and confirmed to reveal the right cert fields.

---

## TL;DR of changes

| #   | Change | File | Why |
| --- | --- | --- | --- |
| 1   | Changed **`root_url`**, **`jsonData_database`**, **`root_user`** from `"requiredWhen": "true"` → `"required": true` | [`dsconfig.json`](dsconfig.json) | These are unconditionally required (the backend `Validate()` rejects empty URL/user/database — `settings.go:163-174`). The wizard's synthetic **General** step only pulls fields with `required: true`; `requiredWhen` is a CEL expression the resolver does not inspect. Also emits proper OpenAPI `required` arrays instead of the `x-dsconfig-required-when` extension. |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-postgresql-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`) |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance.go`, or `plugin-ui`.
The `requiredWhen` conditions that are **real** conditions (the `dependsOn` expressions on
the TLS fields involving `sslmode` / `tlsConfigurationMethod`) were **left untouched** — only
the three literal `"requiredWhen": "true"` values were converted.

No `conformance.go` change was needed here (postgres models no `indexedPair` field, unlike
graphite). No `plugin-ui` change was needed either: the auth group already uses the
conventional id `authentication`, which the wizard's required-fields resolver already
recognises (that generalisation shipped with the graphite work), so the auth fields fold into
General correctly.

---

## Section layout

Verified rendering top-to-bottom in the new UI (tab mode) and matched to the legacy editor's
`h3` section headings.

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Connection** (`connection`) | no | Host URL, Database name |
| 2 | **Authentication** (`authentication`) | no | Username, Password, TLS/SSL Mode, TLS/SSL Method |
| 3 | **TLS/SSL Auth Details** (`tls-details`) | yes | (file-path) Root Certificate → Client Certificate → Client Key; (file-content) Root Certificate → Client Certificate → Client Key |
| 4 | **Additional settings** (`additional-settings`) | yes | Version, Min time interval, TimescaleDB, Max open, Max lifetime |

Both UIs use the **same four sections with identical titles**: legacy `h3` headings were
`Connection`, `Authentication`, `TLS/SSL Auth Details`, `Additional settings` (plus `h6`
sub-labels `PostgreSQL Options` / `Connection limits` inside Additional settings); the new UI
renders `Connection`, `Authentication`, `TLS/SSL Auth Details` (Optional), `Additional
settings` (Optional). The two optional sections are collapsible accordions.

### Wizard mode: the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) every
field in the auth group (`authentication`), plus their `dependsOn` parents/children.

Confirmed in `verify-postgres-wizard.js` (step **General 1/5**, fields captured with their
`*` required markers):

- **Host URL\*** (`root_url`), **Database name\*** (`jsonData_database`), **Username\***
  (`root_user`) — the three `required: true` fields.
- **Password**, **TLS/SSL Mode** (`require`), **TLS/SSL Method** (`File system path`) — folded
  in as auth-group members.
- **TLS/SSL Client Certificate** / **TLS/SSL Client Key** — the `dependsOn` children revealed
  by the default `require` + `file-path` state.

**Effect of the `required: true` fix (before/after, both captured):**

| Field | Before (`requiredWhen: "true"`) | After (`required: true`) |
| --- | --- | --- |
| Host URL (`root_url`, Connection group) | **absent** from General | **present** in General ✅ |
| Database name (`jsonData_database`, Connection group) | **absent** from General | **present** in General ✅ |
| Username (`root_user`, Authentication group) | present (via auth-group membership) | present ✅ |

Before the fix the wizard's General step was missing **Host URL** and **Database name** (they
were only reachable in the `Connection` step); after the fix all three required fields appear
in General. Tab mode is unaffected — the synthetic `_required` group is filtered out there, so
it still shows the four sections in order.

---

## Field-by-field parity

Legend: ✅ present & matching · 🔀 conditional (revealed by `dependsOn`)

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| Host URL \* | text input | `root_url` | input | `root.url` | ✅ |
| Database name \* | text input | `jsonData_database` | input | `jsonData.database` | ✅ |
| Username \* | text input | `root_user` | input | `root.user` | ✅ |
| Password | password (`SecretInput`) | `secureJsonData_password` | secure input | `secureJsonData.password` | ✅ |
| TLS/SSL Mode | select (`disable`/`require`/`verify-ca`/`verify-full`) | `jsonData_sslmode` | select | `jsonData.sslmode` | ✅ |
| TLS/SSL Method | select (File system path / Certificate content) | `jsonData_tlsConfigurationMethod` | select | `jsonData.tlsConfigurationMethod` | ✅ 🔀 |
| TLS/SSL Root Certificate (path) | text input | `jsonData_sslRootCertFile` | input | `jsonData.sslRootCertFile` | ✅ 🔀 |
| TLS/SSL Client Certificate (path) | text input | `jsonData_sslCertFile` | input | `jsonData.sslCertFile` | ✅ 🔀 |
| TLS/SSL Client Key (path) | text input | `jsonData_sslKeyFile` | input | `jsonData.sslKeyFile` | ✅ 🔀 |
| TLS/SSL Root Certificate (content) | **textarea** | `secureJsonData_tlsCACert` | secure input¹ | `secureJsonData.tlsCACert` | ✅ 🔀 |
| TLS/SSL Client Certificate (content) | **textarea** | `secureJsonData_tlsClientCert` | secure input¹ | `secureJsonData.tlsClientCert` | ✅ 🔀 |
| TLS/SSL Client Key (content) | **textarea** | `secureJsonData_tlsClientKey` | secure input¹ | `secureJsonData.tlsClientKey` | ✅ 🔀 |
| Version | select (9.0 … 15) | `jsonData_postgresVersion` | select | `jsonData.postgresVersion` | ✅ |
| Min time interval | text input | `jsonData_timeInterval` | input | `jsonData.timeInterval` | ✅ |
| TimescaleDB | switch | `jsonData_timescaledb` | switch | `jsonData.timescaledb` | ✅ |
| Max open | number | `jsonData_maxOpenConns` | number | `jsonData.maxOpenConns` | ✅ |
| Max lifetime | number | `jsonData_connMaxLifetime` | number | `jsonData.connMaxLifetime` | ✅ |

All 17 modeled fields render in the new UI and were located across the four sections
(`dump-postgres-additional.js` confirmed the Additional-settings fields: Version shown as
`9.3` = 903 default, Min time interval `1m`, TimescaleDB, Max open, Max lifetime). The legacy
`Name` and `Default` controls at the top are Grafana editor chrome (datasource name + default
toggle), not part of the datasource config, and are correctly **not** modeled in `dsconfig.json`.

¹ **Not a discrepancy.** The `dsconfig.json` declares `ui.component: "textarea"` for the three
`file-content` TLS cert fields, but the new renderer checks `target === "secureJsonData"`
_before_ the `textarea` branch, so any secure field is drawn as a masked secure input
(`input[type="password"]`) with a show/hide toggle. Verified directly
(`dump-postgres-filecontent.js`): the three fields render as password inputs carrying the same
placeholders (`-----BEGIN CERTIFICATE-----` / `-----BEGIN RSA PRIVATE KEY-----`) and labels
(TLS/SSL Root / Client Certificate / Client Key). Both UIs collect the same PEM text into the
same `secureJsonData` keys; only the widget affordance differs. This is a renderer policy in
`plugin-ui`, not a schema gap (same footnote as graphite's TLS cert fields).

---

## Gaps found

**None beyond the required-field fix.** The complete legacy field set is already modeled in
`dsconfig.json`, so — unlike graphite (which was missing Custom HTTP Headers) — no field had
to be added. Postgres is a SQL datasource and its schema keeps all fields **inline** (no
`packs`), which matches how the legacy SQL editor lays them out.

### Custom HTTP Headers — not applicable (verified)

Postgres talks the PostgreSQL wire protocol over TCP, **not HTTP**, so it has no HTTP-headers
concept. Legacy DOM capture (`legacy-expand-postgresql`, `legacy-postgres-parity`):
`hasCustomHeaders: false`, `addHeaderBtn: false`. New UI (tab + wizard):
`hasHeadersEditor: false`. Correctly **not** added.

---

## `fileUpload` evaluation — not applicable to postgres

The task asked to use the `fileUpload` control **if the legacy UI uses it**. It does not:

- The legacy PostgreSQL editor renders the file-path TLS fields (Root/Client Certificate,
  Client Key) as **plain text inputs** (paths on the Grafana host) and the file-content TLS
  fields as **textareas** (inline PEM). No `<input type="file">` and no upload button were
  found in the legacy DOM (`fileInputs: 0`, `uploadButtons: []` in every capture).
- The new UI's `fileUpload` component (`FileUploadField.tsx`) only activates when a field
  declares `ui.fileMapping` (multi-key JSON distribution, e.g. a GCP service-account file); it
  does not model single-PEM or path-string upload.

**Decision:** do **not** add `fileUpload` to any postgres field. The cert fields keep their
current modeling; both UIs collect the same values into the same `jsonData` / `secureJsonData`
keys.

---

## Conditional fields & effects — tested

Postgres drives its conditionals with **two `select` dropdowns** (`sslmode`,
`tlsConfigurationMethod`), not switches. Each scenario was run on a fresh page, the
`TLS/SSL Auth Details` accordion expanded, and the visible cert fields probed
(`verify-postgres-conditionals.js`; secure inputs detected as `input[type="password"]`):

| Scenario (sslmode, method) | Root cert (path/content) | Client cert (path/content) | Client key (path/content) | TLS/SSL Method select | Matches schema `dependsOn`? |
| --- | --- | --- | --- | --- | --- |
| **A** `require`, `file-path` | hidden | **shown** | **shown** | shown | ✅ root cert hidden — `require` isn't verify-ca/verify-full |
| **B** `disable` | hidden | hidden | hidden | **hidden** | ✅ method + all certs gated off by `sslmode != 'disable'` |
| **C** `verify-full`, `file-path` | **shown** | **shown** | **shown** | shown | ✅ root cert now revealed by `(sslmode == verify-ca \|\| verify-full)` |
| **D** `verify-full`, `file-content` | **shown** (secure input) | **shown** (secure input) | **shown** (secure input) | shown | ✅ path inputs vanish; the three `secureJsonData` PEM fields appear (2× `BEGIN CERTIFICATE`, 1× `BEGIN RSA PRIVATE KEY`) |

Observed transitions, exactly matching the schema:

- Selecting **`sslmode = disable`** hides the **TLS/SSL Method** select and every cert field
  (they all carry `jsonData_sslmode != 'disable'`).
- Selecting **`verify-full`** (or `verify-ca`) additionally reveals the **Root Certificate**
  field, which is otherwise hidden under `require` — the root-cert `dependsOn` includes
  `(jsonData_sslmode == 'verify-ca' || jsonData_sslmode == 'verify-full')`.
- Flipping **TLS/SSL Method** from `File system path` → `Certificate content` swaps the
  file-path text inputs (`sslRootCertFile` / `sslCertFile` / `sslKeyFile`, `jsonData`) for the
  inline PEM secure inputs (`tlsCACert` / `tlsClientCert` / `tlsClientKey`, `secureJsonData`) —
  the two mutually-exclusive `tlsConfigurationMethod == 'file-path'` vs `'file-content'`
  branches. This mirrors the legacy editor's file-path-vs-content toggle.

The legacy datasource under test is itself in a verify-mode + file-path state (legacy DOM shows
the Root Certificate path input, `textareas: 0`), i.e. it matches scenario **C** — consistent
with the new UI.

**Effects:** postgres's schema contains **no** `effects` blocks. Its TLS visibility is a set of
plain `dependsOn` CEL expressions over the two selects; there is no virtual selector that fans
out to write multiple fields, so nothing for `effects` to model, and none were added.

---

## Verification

```
go generate ./registry/grafana-postgresql-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-postgresql-datasource/...        # 8/8 conformance subtests PASS
go test ./registry/... ./schema/...                         # entire suite PASS (no regressions)
```

Conformance subtests (postgresql): `BaseFieldsResolved`, `SchemaRoundTrip`,
`SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`,
`JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings` — all
**PASS**.

After regeneration, `settings.gen.json` moved `url` + `user` into the root `required` array and
`database` into the `jsonData` `required` array, and dropped the three
`x-dsconfig-required-when: "true"` extensions.

---

## Files changed

- [`registry/grafana-postgresql-datasource/dsconfig.json`](dsconfig.json) — changed `root_url`,
  `jsonData_database`, and `root_user` from `"requiredWhen": "true"` to `"required": true`
  (so they render in the wizard's General step and emit OpenAPI `required`). The real
  `dependsOn` TLS conditions were left untouched.
- [`registry/grafana-postgresql-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-postgresql-datasource/settings.gen.json`](settings.gen.json) — regenerated
  by `go generate` (`url`/`user`/`database` now in the spec `required` arrays).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema/conformance.go`, and `plugin-ui`.
