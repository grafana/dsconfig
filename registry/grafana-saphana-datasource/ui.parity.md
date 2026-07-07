# SAP HANA® — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered
from this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-saphana-datasource` (a native **SQL** datasource — it speaks the SAP HANA wire protocol via the `github.com/SAP/go-hdb` driver over TCP/TLS, **not HTTP**)
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/ffrbqiy4qb9q8a` (Grafana Enterprise 13.2.x, upstream `src/components/ConfigEditor.tsx` using `@grafana/ui` `LegacyForms`)
- **New UI:** Storybook `ConfigEditor/DatasourceConfigWizard` (`--tab` / `--wizard`, `args=pluginType:grafana-saphana-datasource`), which fetches this entry's `dsconfig.json`
- **Method:** Static schema-vs-legacy analysis. The field/label/placeholder/section inventory is taken from the upstream plugin source researched in [`README.md`](README.md) (`src/selectors.ts`, `src/components/ConfigEditor.tsx`, `src/types.ts`, `pkg/models/settings.go`, `pkg/plugin/driver.go` at pinned commit `267f493…`), cross-checked against `dsconfig.json` and the backend contract in [`settings.go`](settings.go). The required-field change is proven by the `go test` conformance suite and the regenerated `*.gen.json` artifacts.
- **Result:** **Parity achieved (static).** All **15** modeled fields are present in the schema and route to the same `jsonData` / `secureJsonData` storage targets the legacy editor writes. **No missing fields** — the complete legacy field set was already modeled. SAP HANA is not an HTTP datasource, so it correctly has **no HTTP-headers editor** and **no file-upload** control (legacy confirmed none). The one change required was converting the single unconditionally-required field (`jsonData_server`) from `requiredWhen: "true"` → `required: true`, so the wizard's synthetic **General** step pulls it in and the OpenAPI spec emits a proper `required` array. The schema's one `effects` block (on the inverted **TLS** switch) is documented below.

> ⚠️ **Evidence caveats (no live screenshots this run).**
> - **Legacy live capture unavailable:** the `grafana-saphana-datasource` plugin is **not installed** on the target Grafana (`/api/plugins/grafana-saphana-datasource/settings` → *"Plugin not found"*; UID `ffrbqiy4qb9q8a` → *"Data source not found"*; 45 datasource plugins installed, none SAP HANA). The edit page only renders a spinner. Field parity is therefore asserted from the upstream-researched README + backend source, not a live DOM capture.
> - **New-UI capture deferred (storybook offline):** `http://192.168.1.241:58899` was unreachable on first try and was not retried, per instructions.
> - Both the field-by-field table and the `effects`/conditional behavior below are backed by static schema + upstream-source analysis and the passing `go test` conformance suite; they were **not** verified against a running UI this run.

---

## TL;DR of changes

| #   | Change | File | Why |
| --- | --- | --- | --- |
| 1   | Changed **`jsonData_server`** from `"requiredWhen": "true"` → `"required": true` | [`dsconfig.json`](dsconfig.json) | `server` is unconditionally required — the backend `Validate()` rejects an empty server (`settings.go:199-201`, `ErrInvalidServerName`). The wizard's synthetic **General** step only pulls fields with `required: true`; `requiredWhen` is a CEL expression the required-fields resolver does not inspect. Also emits a proper OpenAPI `required: ["server"]` array instead of the `x-dsconfig-required-when: "true"` extension. |
| 2   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-saphana-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`). `settings.examples.gen.json` was unaffected. |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `conformance_test.go`, or `plugin-ui`.
The **five conditional** `requiredWhen` expressions were **left untouched** — only the single literal
`"requiredWhen": "true"` was converted:

| Field | `requiredWhen` (unchanged) | Backend rule mirrored |
| --- | --- | --- |
| `jsonData_port` | `jsonData_instance == '' \|\| jsonData_databaseName == ''` | `Port == 0 && (Instance == "" \|\| DatabaseName == "")` → `ErrInvalidPort` (`settings.go:202-204`) |
| `jsonData_username` | `jsonData_tlsAuth != true` | username required unless TLS client auth (`settings.go:214-216`) |
| `secureJsonData_password` | `jsonData_tlsAuth != true` | password required unless TLS client auth (`settings.go:217-219`) |
| `secureJsonData_tlsClientCert` | `jsonData_tlsAuth == true` | client cert required when `tlsAuth` (`settings.go:206-209`) |
| `secureJsonData_tlsClientKey` | `jsonData_tlsAuth == true` | client key required when `tlsAuth` (`settings.go:210-212`) |

No `conformance_test.go` change was needed (saphana models no `indexedPair` field). No `plugin-ui` change
was needed for the fix itself — the `required: true` conversion is the whole change.

---

## Section layout

The schema declares **five groups**; the legacy editor composes the same fields in the same order
(`src/components/ConfigEditor.tsx`). saphana uses `@grafana/ui` `LegacyForms` (`FormField` /
`SecretFormField` / `Switch` / `InlineFormLabel`) rather than `<h3>` section headings, so a legacy
DOM heading probe returns no `h1–h6` — the grouping is visual, driven by `InlineFieldRow` blocks.

| Order | Group (`id`) | Fields (in display order) |
| --- | --- | --- |
| 1 | **HTTP** (`http`) | Server address, Server port |
| 2 | **Auth** (`auth`) | Username, Password |
| 3 | **TLS / SSL Settings** (`tls`) | TLS, Skip TLS Verify, TLS Client Auth, Client Cert, Client Key, With CA Cert, CA Cert |
| 4 | **Tenant Properties** (`tenant-properties`) | Tenant database name, Tenant instance number |
| 5 | **Additional Properties** (`additional-properties`) | Default schema, Timeout |

> Note: the connection group is titled **HTTP** in the schema, but SAP HANA carries **no** HTTP semantics
> (no URL, no headers, no basic-auth-over-HTTP). It is purely the host+port group; the label is cosmetic and
> was left as-is (out of scope for this fix).

### Wizard mode: the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General** from (a) every field
marked `required: true`, plus (b) auth-group fields and their `dependsOn` parents/children.

**Effect of the `required: true` fix:**

| Field | Before (`requiredWhen: "true"`) | After (`required: true`) |
| --- | --- | --- |
| Server address (`jsonData_server`, HTTP group) | **absent** from General (resolver ignores `requiredWhen`) | **present** in General ✅ |

`server` is the **only** unconditionally-required field, so it is the only one the fix promotes into
General. Every other required field on this datasource is *conditionally* required (see the table above)
and therefore correctly stays modeled with `requiredWhen`. Tab mode is unaffected — the synthetic
required group is filtered out there, so the five sections render in order.

> Not verified live this run (storybook offline / plugin not installed). The promotion of a
> `required: true` field into General is the same, already-shipped resolver mechanism validated for
> postgresql; the change here is structurally identical.

---

## Field-by-field parity

Legend: ✅ present & matching · 🔀 conditional (revealed by `dependsOn`) · ⭐ unconditionally required (`required: true`)

| Legacy UI field | Control (legacy) | New UI (schema id) | Control (new) | Storage target | Status |
| --- | --- | --- | --- | --- | --- |
| Server address \* | text input | `jsonData_server` | input | `jsonData.server` | ✅ ⭐ |
| Server port | text input | `jsonData_port` | input (number) | `jsonData.port` | ✅ 🔀¹ |
| Username | text input | `jsonData_username` | input | `jsonData.username` | ✅ 🔀¹ |
| Password | password (`SecretFormField`) | `secureJsonData_password` | secure input | `secureJsonData.password` | ✅ 🔀¹ |
| TLS | switch (inverted²) | `jsonData_tlsDisabled` | switch | `jsonData.tlsDisabled` | ✅ (has `effects`) |
| Skip TLS Verify | switch | `jsonData_tlsSkipVerify` | switch | `jsonData.tlsSkipVerify` | ✅ 🔀 |
| TLS Client Auth | switch | `jsonData_tlsAuth` | switch | `jsonData.tlsAuth` | ✅ 🔀 |
| Client Cert | textarea (rows=7) | `secureJsonData_tlsClientCert` | secure input³ | `secureJsonData.tlsClientCert` | ✅ 🔀¹ |
| Client Key | textarea (rows=7) | `secureJsonData_tlsClientKey` | secure input³ | `secureJsonData.tlsClientKey` | ✅ 🔀¹ |
| With CA Cert | switch | `jsonData_tlsAuthWithCACert` | switch | `jsonData.tlsAuthWithCACert` | ✅ 🔀 |
| CA Cert | textarea (rows=7) | `secureJsonData_tlsCACert` | secure input³ | `secureJsonData.tlsCACert` | ✅ 🔀 |
| Tenant database name | text input | `jsonData_databaseName` | input | `jsonData.databaseName` | ✅ |
| Tenant instance number | number input | `jsonData_instance` | input | `jsonData.instance` | ✅⁴ |
| Default schema | text input | `jsonData_defaultSchema` | input | `jsonData.defaultSchema` | ✅ |
| Timeout | text input | `jsonData_timeout` | input | `jsonData.timeout` (default `"30"`) | ✅ |

All 15 modeled fields map 1:1 to the upstream editor's fields and to the backend `Config` struct
(proven by the `JSONDataMatchesStruct` / `JSONDataTypesMatchStruct` conformance subtests). The legacy
`Name` and `Default` controls at the top are Grafana editor chrome (datasource name + default toggle),
not part of the datasource config, and are correctly **not** modeled.

¹ **Conditional required** (see the `requiredWhen` table in TL;DR) — the field always renders; it is
only *required* under its condition.
² **Inverted switch.** The editor renders the "TLS" toggle as `!tlsDisabled` (`ConfigEditor.tsx:173`):
switch **ON = TLS enabled = `tlsDisabled:false`** (the default). The schema stores the raw
`jsonData_tlsDisabled` boolean and attaches an `effects` block (below) to replicate the editor's reset.
³ **Secure-input footnote (not a discrepancy).** The three TLS cert/key fields declare
`ui.component: "textarea"`, but the new renderer draws any `secureJsonData` field as a masked secure
input (show/hide toggle) because it checks `target === "secureJsonData"` before the `textarea` branch —
the same renderer policy documented for postgresql/graphite. Both UIs collect the same PEM text into the
same `secureJsonData` keys; only the widget affordance differs. (Renderer policy in `plugin-ui`, a schema
non-issue; not re-verified live this run.)
⁴ **`instance` stored as string.** Despite the `type="number"` input (`HANAConfig.instance?: number`),
the editor writes it via the generic `onUpdateDatasourceJsonDataOption` with no numeric coercion and the
backend reads it as a `string` (`settings.go:85`), concatenating `3<instance>13` for the derived port.
The schema and Go `Config` correctly model it as `string` (see README "Potential upstream bugs" #1).

---

## Gaps found

**None beyond the required-field fix.** The complete legacy field set is already modeled in
`dsconfig.json`; no field had to be added. All fields are declared inline (no `packs`), matching how the
legacy SQL editor lays them out.

### Custom HTTP Headers — not applicable (confirmed)

SAP HANA talks the HANA SQL wire protocol over TCP (`go-hdb`), **not HTTP**, so it has no HTTP-headers
concept. The upstream editor renders no headers UI, the backend reads no `HTTPClientOptions`/headers,
and the schema declares no headers field (`grep` for `header` in `dsconfig.json` → none). Correctly
**not** added.

---

## `fileUpload` evaluation — not applicable to SAP HANA

The task asked to use the `fileUpload` control **only if the legacy UI uses it**. It does not:

- The legacy editor renders the three TLS cert/key fields as **textareas** (inline PEM,
  `src/components/ui/CertificationKey.tsx`, rows=7) — no `<input type="file">`, no upload button.
- The new UI's `fileUpload` component only activates when a field declares `ui.fileMapping` (multi-key
  JSON distribution, e.g. a GCP service-account file); it does not model single-PEM textareas.

**Decision:** do **not** add `fileUpload` to any saphana field. The cert fields keep their current
modeling (secure textareas → masked secure inputs); both UIs collect the same PEM into the same
`secureJsonData` keys.

---

## Conditional fields & `effects` — analysis

saphana drives its TLS visibility with **switches** (`tlsDisabled`, `tlsAuth`, `tlsAuthWithCACert`) via
`dependsOn`, and — unlike postgresql — it also carries one **`effects`** block.

### `dependsOn` visibility (from the schema)

| Field | `dependsOn` | Visible when |
| --- | --- | --- |
| `jsonData_tlsSkipVerify` | `jsonData_tlsDisabled != true` | TLS enabled |
| `jsonData_tlsAuth` | `jsonData_tlsDisabled != true` | TLS enabled |
| `jsonData_tlsAuthWithCACert` | `jsonData_tlsDisabled != true` | TLS enabled |
| `secureJsonData_tlsClientCert` | `jsonData_tlsDisabled != true && jsonData_tlsAuth == true` | TLS enabled **and** client auth on |
| `secureJsonData_tlsClientKey` | `jsonData_tlsDisabled != true && jsonData_tlsAuth == true` | TLS enabled **and** client auth on |
| `secureJsonData_tlsCACert` | `jsonData_tlsDisabled != true && jsonData_tlsAuthWithCACert == true` | TLS enabled **and** With-CA-Cert on |

So disabling TLS (`tlsDisabled = true`) hides all six TLS sub-fields; the client cert/key appear only
under **TLS Client Auth**, and the CA cert only under **With CA Cert**. This mirrors the upstream
conditional render (`ConfigEditor.tsx:178,206,224,238`).

### The `effects` block (on `jsonData_tlsDisabled`)

```json
"effects": [
  {
    "when": "value == true",
    "set": {
      "jsonData_tlsSkipVerify": false,
      "jsonData_tlsAuth": false,
      "jsonData_tlsAuthWithCACert": false
    }
  }
]
```

**Behavior (read from the schema):** when the stored `tlsDisabled` value becomes **`true`** — i.e. the
user turns the (inverted) **TLS** switch **off**, disabling TLS — the UI **fans out a write** that resets
the three dependent TLS switches (`tlsSkipVerify`, `tlsAuth`, `tlsAuthWithCACert`) to `false`. This
replicates the upstream editor's `onTLSSettingsChange` side-effect (`ConfigEditor.tsx:54-58`, per README),
so stale `true` flags cannot persist in `jsonData` while TLS is off. Because the cert fields are gated on
those switches via `dependsOn`, clearing the switches also removes any reason to show the cert inputs.

- **`when` semantics:** `value` refers to this field's own stored value (`tlsDisabled`). The effect fires
  on the *disable* transition, not the enable one — there is no reciprocal effect, matching upstream
  (re-enabling TLS leaves the sub-switches at their reset `false` state).
- **Not a virtual selector:** the switch maps 1:1 to a real stored boolean; `effects` only adds the reset
  side-effect. So it is modeled as a plain storage field **with** `effects`, not as a virtual field.
- **Spec emission:** `effects` is **UI-only** and consumed by `plugin-ui`; it is **not** emitted into
  `schema.gen.json` / `settings.gen.json` (confirmed — no `effects` / `x-dsconfig-effect` keys in either
  artifact). The `required: true` fix therefore does not interact with the `effects` block in any way.

**Contrast with postgresql:** the postgres schema has **no** `effects` (its TLS visibility is pure
`dependsOn` over two selects). saphana needs `effects` because its inverted TLS *master* switch must
actively *reset* sub-state, not merely hide it.

> The `effects` and `dependsOn` transitions above are read from the schema and the upstream editor source;
> they were **not** exercised against a running wizard this run (storybook offline; plugin not installed
> for a legacy capture).

---

## Verification

```
go generate ./registry/grafana-saphana-datasource/...   # regenerate schema.gen.json / settings.gen.json — OK
go test ./registry/grafana-saphana-datasource/...        # ok (0.357s) — all conformance subtests PASS
```

Conformance subtests (saphana), all **PASS**: `BaseFieldsResolved`, `SchemaRoundTrip`,
`SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`, `JSONDataMatchesStruct`,
`JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings`.

After regeneration, `schema.gen.json` and `settings.gen.json` moved `server` into the `jsonData`
`required` array and dropped its `x-dsconfig-required-when: "true"` extension; the five conditional
`x-dsconfig-required-when` expressions and the `x-dsconfig-depends-on` expressions are unchanged.
`settings.examples.gen.json` was not modified.

---

## Files changed

- [`registry/grafana-saphana-datasource/dsconfig.json`](dsconfig.json) — changed `jsonData_server` from
  `"requiredWhen": "true"` to `"required": true` (so it renders in the wizard's General step and emits
  OpenAPI `required`). The five conditional `requiredWhen` expressions and all `dependsOn` / `effects`
  were left untouched.
- [`registry/grafana-saphana-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-saphana-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`server` now in the spec `required` array).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `settings.examples.gen.json`,
`conformance_test.go`, and `plugin-ui`.

---

## Not dsconfig-fixable / open items

- **Live legacy capture** requires the `grafana-saphana-datasource` plugin to be **installed** on the
  target Grafana and a datasource at UID `ffrbqiy4qb9q8a` (currently absent). Neither installing the
  plugin nor creating a datasource is in scope for this task; the field inventory was therefore taken from
  the upstream-researched README + backend source.
- **New-UI (storybook) verification** is deferred until `http://192.168.1.241:58899` is reachable. When
  it is, capture tab + wizard for `pluginType:grafana-saphana-datasource` to visually confirm: (a) `server`
  appears in the **General** step, (b) whether the conditionally-required `auth`-group fields
  (`username`/`password`) also fold into General, and (c) the `effects` reset + `dependsOn` TLS transitions.
- No `conformance`/`plugin-ui` change was found to be required for the requested fix. Should a future
  requirement need the conditional auth fields promoted into General (a `plugin-ui` resolver concern) or
  the `auth` group id aligned to a resolver-recognised id, that would fall **outside** this dsconfig-only
  scope and should be raised separately.
