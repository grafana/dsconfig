# Yugabyte — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-yugabyte-datasource` (a **SQL** datasource; PostgreSQL-wire compatible — the frontend `YugabyteOptions extends SQLOptions`, `src/types.ts:11`)
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/ffqvcosvjlfcwe` (Grafana Enterprise 13.2.0, `@grafana/sql`-style `ConfigurationEditor`)
- **New UI:** Storybook `ConfigEditor/DatasourceConfigWizard` (`--tab` / `--wizard`, `args=pluginType:grafana-yugabyte-datasource`) — **capture deferred (storybook offline)** this run; see note below.
- **Method:** Playwright captured the legacy UI (full-page screenshot + DOM extraction: headings, labels, inputs, header/upload probing) after authenticating (`admin`/`admin`). The new UI (Storybook `192.168.1.241:58899`) was **unreachable on the first attempt**, so per run policy the new-UI screen captures were not taken; parity for the new UI is established from the shared `plugin-ui` resolver behaviour (identical to the verified postgresql case), the committed schema, and the `go test` conformance suite.
- **Result:** **Parity achieved (legacy-verified; new-UI capture deferred).** All **4 modeled fields** are present in the legacy UI and route to identical storage targets; the schema models exactly those 4 and nothing more. Yugabyte is a SQL datasource over the PostgreSQL wire protocol, so it correctly has **no HTTP-headers editor** and **no file-upload** control (confirmed absent in legacy DOM). The one change required was making the three unconditionally-required fields use `required: true` so the wizard's synthetic **General** step pulls them in. Unlike postgresql, Yugabyte exposes **no TLS/SSL UI at all** (the backend hardcodes `sslmode='allow'`), so there are **no conditional cert fields** to exercise.

---

## TL;DR of changes

| #   | Change | File | Why |
| --- | --- | --- | --- |
| 1   | Changed **`root_url`**, **`jsonData_database`**, **`root_user`** from `"requiredWhen": "true"` → `"required": true` | [`dsconfig.json`](dsconfig.json) | These are unconditionally required (the backend `Validate()` rejects empty URL/user/database — `pkg/settings.go`; mirrored by `settings_test.go` `missing_user` / `missing_database` / `missing_url` cases). The wizard's synthetic **General** step only pulls fields with `required: true`; `requiredWhen` is a CEL expression the resolver does not inspect. Also emits proper OpenAPI `required` arrays instead of the `x-dsconfig-required-when` extension. |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-yugabyte-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`) |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance_test.go`, `schema.go`, or `plugin-ui`.
There were **no conditional `requiredWhen` expressions** in this schema to preserve — all three
`requiredWhen` values were the literal `"true"` (Yugabyte has no TLS-mode-gated cert fields),
so all three were converted and none were left behind.

`settings.examples.gen.json` was **not** regenerated (examples do not depend on the required flag).

---

## Section layout

Verified rendering top-to-bottom in the legacy editor and matched to its section headings.

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Connection** (`connection`) | no | Host URL, Database |
| 2 | **Authentication** (`authentication`) | no | Username, Password |
| — | *Additional Settings* (legacy-only, **empty here**) | yes | *(none rendered — see below)* |

The legacy editor's `h3` headings were **`Connection`**, **`Authentication`**, and
**`Additional Settings`**. The `dsconfig.json` models the first two groups with identical
titles (`Connection` order 1, `Authentication` order 2). The legacy **Additional Settings**
accordion was expanded during capture and found **empty** — the only control it can host is the
**Secure Socks Proxy** toggle (`ConfigEditor.tsx:83`, `jsonData.enableSecureSocksProxy`), which
(a) is gated behind Grafana's `secureSocksDSProxyEnabled` server config and did not render in
this instance, and (b) is intentionally **excluded from the registry schema per repo policy**.
So no schema group is needed for it, and the two modeled groups fully cover every rendered field.

### Wizard mode: the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from (a) every field marked `required: true`, plus (b) every
field in the auth group (`authentication`), plus their `dependsOn` parents/children.

For Yugabyte that resolves to:

- **Host URL\*** (`root_url`), **Database\*** (`jsonData_database`), **Username\*** (`root_user`)
  — the three `required: true` fields.
- **Password** (`secureJsonData_password`) — folded in as an `authentication`-group member.

There are no `dependsOn` children (no TLS conditionals), so the General step is exactly these
four fields.

**Effect of the `required: true` fix (reasoned from the resolver; identical mechanism to the
verified postgresql case):**

| Field | Before (`requiredWhen: "true"`) | After (`required: true`) |
| --- | --- | --- |
| Host URL (`root_url`, Connection group) | **absent** from General | **present** in General ✅ |
| Database (`jsonData_database`, Connection group) | **absent** from General | **present** in General ✅ |
| Username (`root_user`, Authentication group) | present (via auth-group membership) | present ✅ |

Before the fix the wizard's General step was missing **Host URL** and **Database** (reachable
only in the `Connection` step); after the fix all three required fields appear in General. Tab
mode is unaffected — the synthetic `_required` group is filtered out there, so it still shows the
two sections in order.

---

## Field-by-field parity

Legend: ✅ present & matching

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| Host URL \* | text input (`localhost:5433`) | `root_url` | input | `root.url` | ✅ |
| Database \* | text input (`yb_demo`) | `jsonData_database` | input | `jsonData.database` | ✅ |
| Username \* | text input (`yugabyte`) | `root_user` | input | `root.user` | ✅ |
| Password | password (`SecretInput`, `********`) | `secureJsonData_password` | secure input | `secureJsonData.password` | ✅ |

All **4 modeled fields** map 1:1 to the legacy editor. Legacy DOM capture
(`legacy-auth-yugabyte.json`) found exactly these four inputs and the labels
`Host URL *`, `Database *`, `Username *`, `Password` under the "Fields marked with * are
required" note — the three asterisked fields are precisely the three converted to
`required: true`, and Password is correctly **not** required. The legacy `Name` and `Default`
controls (datasource name + default toggle) are Grafana editor chrome, not part of the
datasource config, and are correctly **not** modeled.

Note on placeholders: the legacy Host URL placeholder is `localhost:5433` (YugabyteDB YSQL
default port **5433**, not PostgreSQL's 5432) — matched exactly by the schema.

---

## Gaps found

**None beyond the required-field fix.** The complete legacy field set (4 fields) is already
modeled in `dsconfig.json`. No field had to be added. The schema keeps all fields **inline**
(no `packs`), matching the flat legacy SQL editor layout.

### Custom HTTP Headers — not applicable (verified)

Yugabyte talks the PostgreSQL wire protocol over TCP, **not HTTP**, so it has no HTTP-headers
concept. Legacy DOM capture: `hasCustomHeaders: false`, `addHeaderBtn: false`. Correctly **not**
modeled.

---

## `fileUpload` evaluation — not applicable to Yugabyte

The task calls for the `fileUpload` control **only if the legacy UI uses it**. It does not:

- The legacy Yugabyte editor renders only plain text inputs and one password input; there is
  **no `<input type="file">`** and **no upload button** (`fileInputs: 0`, `uploadButtons: []`
  in every capture).
- Yugabyte has no TLS cert fields at all (see below), so there is nothing a file/upload widget
  could target.

**Decision:** do **not** add `fileUpload` to any Yugabyte field.

---

## Conditional fields & effects — none (TLS is hardcoded)

This is the key divergence from `grafana-postgresql-datasource`. Yugabyte exposes **no TLS/SSL
UI**: the backend hardcodes `sslmode='allow'` in `BuildConnectionString` (`pkg/settings.go:52`).
There is no `sslmode` select, no `tlsConfigurationMethod` select, and no root/client cert fields
in either the legacy editor or the schema.

Consequently:

- The schema contains **no `dependsOn`** expressions and **no `effects`** blocks — there are no
  conditional cert fields to reveal/hide, so there was nothing to exercise or model.
- Legacy DOM capture confirms **no TLS/SSL section** and no certificate inputs (the "Additional
  Settings" accordion is empty in this instance).
- libpq's `allow` mode tries plaintext first and falls back to TLS only if the server rejects
  it, so whether the connection is encrypted is determined entirely by cluster policy, not by
  any editor field. (Documented in `dsconfig.json` `instructions[tls]`.)

---

## Verification

```
go generate ./registry/grafana-yugabyte-datasource/...   # regenerate schema.gen.json / settings.gen.json
go test ./registry/grafana-yugabyte-datasource/...        # 8/8 conformance subtests + settings tests PASS
go test ./schema/...                                      # shared conformance engine PASS (no regressions)
```

Conformance subtests (yugabyte): `BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`,
`SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`, `JSONDataMatchesStruct`,
`JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings` — all **PASS**. The plugin's own
`settings_test.go` (`TestLoadConfig`, `TestValidate`, etc.) also **PASS**, confirming the
backend really does require `url`, `user`, and `jsonData.database`.

After regeneration, `settings.gen.json` / `schema.gen.json` moved `url` + `user` into the spec
`required` array and `database` into the `jsonData` `required` array, and dropped the three
`x-dsconfig-required-when: "true"` extensions.

### New-UI capture deferred (storybook offline)

The Storybook host (`http://192.168.1.241:58899`) was unreachable on the first request this run,
so — per run policy (do not retry >1×) — the new-UI screenshots were skipped. New-UI parity is
therefore asserted from: (1) the shared `plugin-ui` `DatasourceConfigWizard` resolver, whose
behaviour on `required: true` + auth-group folding is identical to the directly-verified
`grafana-postgresql-datasource` case; (2) the regenerated schema, which now carries the correct
OpenAPI `required` arrays the wizard reads; and (3) the passing conformance suite. When Storybook
is back up, re-run:

```
node capture-new-generic.js grafana-yugabyte-datasource <dsconfig.json> tab    yugabyte-tab
node capture-new-generic.js grafana-yugabyte-datasource <dsconfig.json> wizard yugabyte-wiz
```

to attach the tab/wizard screenshots and confirm the General step shows Host URL / Database /
Username / Password.

### Legacy UID note

The UID supplied for this run (`bfrbqiyqfxhxcf`) does **not** exist in the target instance
(`/api/datasources/uid/bfrbqiyqfxhxcf` → "Data source not found"). The actual Yugabyte
datasource is at UID **`ffqvcosvjlfcwe`** (type `grafana-yugabyte-datasource`), which is what
was captured.

---

## Files changed

- [`registry/grafana-yugabyte-datasource/dsconfig.json`](dsconfig.json) — changed `root_url`,
  `jsonData_database`, and `root_user` from `"requiredWhen": "true"` to `"required": true`
  (so they render in the wizard's General step and emit OpenAPI `required`). No conditional
  `requiredWhen`/`dependsOn` existed to preserve.
- [`registry/grafana-yugabyte-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-yugabyte-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`url`/`user`/`database` now in the spec `required` arrays).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`schema.go`, `conformance_test.go`, and `plugin-ui`.
