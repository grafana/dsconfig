# Wavefront — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
this entry's [`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-wavefront-datasource` (Wavefront / VMware Aria Operations for Applications)
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (Grafana Enterprise 13.2.0) — the Wavefront `ConfigEditor` (`@grafana/ui` `LegacyForms.FormField` + `LegacyForms.SecretFormField`)
- **New UI:** `http://192.168.1.241:58899/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-wavefront-datasource` (Storybook, `ConfigEditor/DatasourceConfigWizard`), also the `--wizard` story.
- **Method:** `go generate` + the shared conformance suite (`go test`) validated the schema and the regenerated artifacts. The legacy inventory was taken from this entry's **source-cited** in-repo docs (`README.md` "Sources researched" + `settings.ts` header), pinned to upstream SHA `267f4937806ed6404b6628d13ae358a5d308e376` — a live legacy capture was not possible in this instance (see [Capture notes](#capture-notes)).
- **Result:** **Parity fix applied.** The two connection fields — **API URL** (`jsonData.url`) and **Token** (`secureJsonData.token`) — were promoted from `requiredWhen: "true"` to `required: true` so they render in the wizard's synthetic **General** step and emit proper OpenAPI `required`. **Custom HTTP Headers** and `fileUpload` were evaluated and correctly **not** modeled (the legacy editor has neither).

---

## TL;DR of changes

| #   | Change                                                                                        | File                             | Why                                                                                                                        |
| --- | --------------------------------------------------------------------------------------------- | -------------------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| 1   | Changed `jsonData_url` from `requiredWhen: "true"` → `required: true`                          | [`dsconfig.json`](dsconfig.json) | URL is unconditionally required (`pkg/models/settings.go:32-34` returns `"invalid url"` when empty); puts URL into the wizard's **General** step and emits `required: ["url"]` in the spec |
| 2   | Changed `secureJsonData_token` from `requiredWhen: "true"` → `required: true`                  | [`dsconfig.json`](dsconfig.json) | Token is unconditionally required (`pkg/models/settings.go:35-37` returns `"invalid credentials"` when empty); puts Token into the wizard's **General** step and emits `required: true` on the `token` secure value |
| 3   | Regenerated `schema.gen.json`, `settings.gen.json` via `go generate ./registry/grafana-wavefront-datasource/...` | generated artifacts | Keep committed artifacts in sync (guarded by `SchemaArtifactInSync`)                                                        |

No changes were made to `settings.go`, `settings.ts`, `README.md`, `schema.go`, `conformance_test.go`, or `plugin-ui`. All schema changes flow through `dsconfig.json`; the rest is produced by `go generate`.

**No conformance-test or plugin-ui change was required.** The change is a pure required-field promotion (already the established pattern — e.g. honeycomb's `secureJsonData_apiKey`, newrelic, sentry), which the shared conformance walker and the plugin-ui General-step resolver already handle.

> Note: neither field had a `required: true` already present, and neither carried a *conditional* `requiredWhen` — both used the literal `requiredWhen: "true"` — so this was a clean 1:1 swap: no redundant `requiredWhen` to delete, no conditional `requiredWhen` to preserve.

---

## Section layout

The `groups` were left in their existing taxonomy (they mirror the two editor `<h3>` sections —
`ConfigEditor.tsx:55,86`). No field was moved.

| Order | Section (`id`) | `optional` | Fields (in display order) |
| --- | --- | --- | --- |
| 1 | **Wavefront settings** (`wavefront-settings`) | no | API URL, Token |
| 2 | **Customization** (`customization`) | yes | Request timeout in seconds |

### Wizard mode: URL + Token in the "General" step

In **wizard mode** the plugin-ui builds a synthetic first step titled **General**
(`resolveRequiredFieldsGroup`) from every field marked `required: true` (plus any auth-group /
`dependsOn` fields). Wavefront has no `authentication` group and no conditionals, so General is
driven purely by the required fields.

- **Before:** both `jsonData_url` and `secureJsonData_token` used `requiredWhen: "true"` (a CEL
  expression the resolver does **not** inspect), so neither was placed in General and the generated
  spec carried only the `x-dsconfig-required-when: "true"` extension.
- **After:** promoting both to `required: true` puts **API URL** and **Token** into General and emits
  a proper OpenAPI `required: ["url"]` under `jsonData` plus `required: true` on the `token` secure
  value. This is unconditionally correct — `LoadSettings` (`pkg/models/settings.go:32-37`) and the
  health check (`pkg/datasource/handler_healthcheck.go:100-111`, `"Enter an API URL."` / `"Enter a
  token."`) both hard-fail a datasource that is missing either value.

Verification of the rendered wizard was **deferred** (Storybook offline — see
[Capture notes](#capture-notes)); the General-step behaviour is asserted from the regenerated spec
(`schema.gen.json` / `settings.gen.json`) plus the plugin-ui resolver behaviour documented for the
prior required-field promotions (appdynamics, splunk, honeycomb).

---

## Field-by-field parity

Legend: ✅ present & matching · 🔒 secure (write-only, `secureJsonData`) · ⤴ promoted to `required: true` by this change · 🚫 excluded per AGENTS.md

| Legacy UI field | Control (legacy) | New UI (schema id) | Storage target | Status |
| --- | --- | --- | --- | --- |
| API URL | text input (`LegacyForms.FormField`) | `jsonData_url` | `jsonData.url` (required) | ✅ ⤴ |
| Token | secure input (`LegacyForms.SecretFormField`) | `secureJsonData_token` | `secureJsonData.token` (required) | ✅ 🔒 ⤴ |
| Request timeout in seconds | number input (`LegacyForms.FormField type="number"`) | `jsonData_requestTimeout` | `jsonData.requestTimeout` (optional, default 30) | ✅ |
| Enable Secure Socks Proxy | `InlineSwitch` (feature-gated) | — | `jsonData.enableSecureSocksProxy` | 🚫 |

**Secure input affordance:** `dsconfig.json` declares `ui.component: "input"` for the token, but the
new renderer draws any `target: "secureJsonData"` field as a masked secure input with a show/hide
toggle — matching the legacy `SecretFormField`. Both UIs collect the same secret into
`secureJsonData.token`; only the widget affordance differs.

**Secure Socks Proxy (excluded):** the legacy editor renders a feature-flag- and version-gated
`enableSecureSocksProxy` switch (`ConfigEditor.tsx:100-132`). It is deliberately **excluded** from
this registry entry per AGENTS.md (it is not even part of the upstream `Settings` struct — it is
consumed transparently via `config.HTTPClientOptions(ctx)`). Not modeled; not in scope.

---

## No Custom HTTP Headers

The Wavefront `ConfigEditor` (`src/components/ConfigEditor.tsx:52-135`) renders only the three
`LegacyForms` fields above plus the excluded proxy switch. It does **not** use
`DataSourceHttpSettings` and has **no** Custom HTTP Headers section or **Add header** button.
Authentication is a single bearer token sent as `Authorization: Bearer <token>`
(`pkg/datasource/datasource.go:45-47`, `pkg/wavefront/client.go:39-44`) — there is nowhere to attach
custom headers. Headers are correctly **not** modeled. (Matches the task premise "legacy confirmed
none".)

## `fileUpload` evaluation — not applicable to Wavefront

The task asked to use the `fileUpload` control **only if the legacy UI uses it**. It does not:

- Every legacy input is a text / secure-text `LegacyForms.FormField` — the URL and request-timeout
  are plain text/number inputs and the Token is a masked `SecretFormField`. There is **no**
  `<input type="file">` and **no** upload button anywhere in the editor.
- The new UI's `fileUpload` component only activates for `ui.fileMapping` (multi-key JSON
  distribution, e.g. a GCP service-account file); nothing here needs it.

**Decision:** do **not** add `fileUpload` to any Wavefront field.

---

## Conformance (no change required)

Promoting the two fields to `required: true` is a pure declarative change:

- **`jsonData_url`** now emits `required: ["url"]` under the `jsonData` spec object and drops the
  `x-dsconfig-required-when: "true"` extension.
- **`secureJsonData_token`** now emits `"required": true` on its `secureValues` entry.

This mirrors the established pattern (honeycomb's required `secureJsonData_apiKey`, newrelic, sentry),
so **no conformance change was needed**. The token remains write-only and correctly listed among
`SecureJsonDataKeys` (`settings.go`); `SchemaSpecHasNoSecureJSON` still passes (no secure value leaks
into the settings spec). The Go `Config` struct is unchanged — `required` is a UI/spec-level
constraint, not a new field — so `JSONDataMatchesStruct` / `JSONDataTypesMatchStruct` are unaffected.

---

## Verification

```
go generate ./registry/grafana-wavefront-datasource/...   # regenerate schema.gen.json / settings.gen.json  → PASS
go test     ./registry/grafana-wavefront-datasource/...   # PASS
go test     ./registry/grafana-wavefront-datasource/... ./schema/...   # PASS (no regressions)
```

`TestSchemaConformance` subtests — all **PASS**:
`BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
`ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
`SecureValuesMatchLoadSettings`.

The hand-authored `settings_test.go` suites (`LoadConfig`, `ApplyDefaults`, `Validate`) also pass
unchanged — they do not reference the required-field metadata.

Generated-spec deltas (`schema.gen.json` / `settings.gen.json`):
`required: ["url"]` added to the `jsonData` spec; `x-dsconfig-required-when: "true"` removed from
`url`; `"required": true` added to the `token` secure value.

---

## Capture notes

- **New-UI capture deferred (Storybook offline).** `http://192.168.1.241:58899` was unreachable on
  the first attempt (HTTP 000); per the task constraints it was not retried. New-UI rendering was
  therefore not captured live; parity is verified via `go test` + the regenerated artifacts and the
  documented plugin-ui resolver behaviour.
- **Legacy live capture not possible in this instance.** The provided legacy UID
  (`afrbqiyoain7kd`) returned **"Data source not found"**, and the `grafana-wavefront-datasource`
  plugin is **not installed** here (`GET /api/plugins/grafana-wavefront-datasource/settings` → HTTP
  404; absent from `/api/plugins?type=datasource`). The legacy inventory instead comes from this
  entry's **source-cited** in-repo docs at the pinned upstream SHA
  (`267f4937806ed6404b6628d13ae358a5d308e376`): `ConfigEditor.tsx:52-135`, `selectors.ts:3-27`,
  `pkg/models/settings.go:32-37`, `handler_healthcheck.go:100-111`. This corroborates the task's
  stated confirmation that the legacy editor has no HTTP headers and no file upload. Environment
  status recorded in `legacy-wavefront-NOTE.json`.

---

## Files changed

- [`registry/grafana-wavefront-datasource/dsconfig.json`](dsconfig.json) — changed `jsonData_url`
  and `secureJsonData_token` from `requiredWhen: "true"` to `required: true` (both now render in the
  wizard's General step and emit proper OpenAPI `required`).
- [`registry/grafana-wavefront-datasource/schema.gen.json`](schema.gen.json),
  [`registry/grafana-wavefront-datasource/settings.gen.json`](settings.gen.json) — regenerated by
  `go generate` (`required: ["url"]` under `jsonData`; `x-dsconfig-required-when: "true"` removed
  from `url`; `"required": true` added to the `token` secure value).

_Unchanged by design:_ `settings.go`, `settings.ts`, `README.md`, `schema.go`,
`settings.examples.gen.json`, `conformance_test.go`, `schema/`, and everything in `plugin-ui`.
