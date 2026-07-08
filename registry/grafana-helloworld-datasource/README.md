# grafana-helloworld-datasource

Declarative configuration schema for the **Hello World** datasource plugin
(`grafana-helloworld-datasource`), a minimal backend-datasource sample/template.

> **This plugin has no configuration surface.** Its config editor renders static
> text and persists nothing; its backend ignores instance settings entirely.
> The lone `secureJsonData.apiKey` field in this entry is a **placeholder** the
> plugin never reads — it exists only to satisfy hard, machine-checked
> constraints in the dsconfig validator and the shared conformance suite. See
> [Modeling decisions](#modeling-decisions) and [Upstream findings](#upstream-findings-and-modeling-caveats).

## Upstream researched

- **Monorepo**: `github.com/grafana/plugins-private`
- **Commit SHA**: `267f4937806ed6404b6628d13ae358a5d308e376`
- **Plugin path**: `plugins/grafana-helloworld-datasource/`

The monorepo is already on disk for this work; do **not** clone it. To reproduce
the research against the same tree:

```bash
git -C <plugins-private-checkout> fetch origin
git -C <plugins-private-checkout> checkout 267f4937806ed6404b6628d13ae358a5d308e376
# then read plugins/grafana-helloworld-datasource/
```

`@grafana/*` dependencies use the monorepo `catalog:` protocol; versions are
resolved from `.yarnrc.yml` at the monorepo root (see
[External components](#external-components)). If upstream `main` has advanced,
re-diff the sources below before merging changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth. One placeholder `secureJsonData` field + informational `instructions` |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig` (blank), `JsonDataConfig` (blank), `SecureJsonDataConfig` (`['apiKey']` placeholder) |
| [`settings.go`](settings.go) | Go `Config` (only `DecryptedSecureJSONData`), `PluginID`, `SecureJsonDataKey` typed constant, and the `LoadConfig` / `ApplyDefaults` / `Validate` utilities |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts via `dsconfig.NewSDKSchema`, defines `SettingsExamples` |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, `Validate`, and `PluginID`/schema parity |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| `schema.gen.json`, `settings.gen.json`, `settings.examples.gen.json` | Committed schema artifacts (regenerate with `go generate ./...`; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the
shared [`registry/`](..) module (`github.com/grafana/dsconfig/registry`). The Go
package name is `helloworlddatasource`.

## Sources researched

Every source below was read at the pinned monorepo SHA
(`267f4937806ed6404b6628d13ae358a5d308e376`).

### Plugin (`plugins/grafana-helloworld-datasource`)

| File | What was read |
| --- | --- |
| `src/plugin.json:3-8` | `type` (`datasource`), `name` (`Hello World`) → `pluginName`, `id` (`grafana-helloworld-datasource`) → `pluginType` + directory name, `backend: true`, `executable: gpx_helloworld` |
| `src/plugin.json:19` | `info.links` is `[]` — **no docs URL** to record, so `docURL` is omitted from `dsconfig.json` |
| `src/plugin.json:24-28` | `dependencies.grafanaDependency` (`>=11.6.7-0`), `grafanaVersion` (`11.x.x`), `plugins: []` |
| `src/module.tsx:13` | `export type Config = {} & DataSourceJsonData;` — **blank** jsonData interface (no plugin-specific fields) |
| `src/module.tsx:14` | `export type SecureConfig = {};` — **no secrets** defined |
| `src/module.tsx:15-16` | `Query = {} & DataQuery;`, `VariableQuery = {};` — query models, not config |
| `src/module.tsx:18-29` | `DataSource extends DataSourceWithBackend<Query, Config>`; constructor sets only `this.annotations = {}`; no settings read on the frontend |
| `src/module.tsx:31-34` | `ConfigEditor` renders `<>Hello World Config Editor!</>` — a static fragment. Receives `options` / `onOptionsChange` but **never** calls `onOptionsChange`, so it writes nothing to storage |
| `src/module.tsx:46-49` | `new DataSourcePlugin(...).setConfigEditor(ConfigEditor).setQueryEditor(QueryEditor).setVariableQueryEditor(VariablesEditor)` — the stub editors are wired in, replacing Grafana's default HTTP settings editor |
| `src/module.spec.tsx:6-13` | Config editor test only asserts the static string renders — confirms there are no inputs |
| `pkg/main.go:13` | `const PluginId = "grafana-helloworld-datasource"` — matches `plugin.json` id and this entry's `PluginID` |
| `pkg/main.go:17-22` | `CheckHealth` always returns `HealthStatusOk`, message `hello world datasource just works but does nothing` — never reads settings |
| `pkg/main.go:24-38` | `QueryData` returns a static single-string frame (`hello world response`) with a fixed notice for every query — never reads settings |
| `pkg/main.go:46-56` | `main`: the instance factory `func(_ context.Context, settings backend.DataSourceInstanceSettings) (...)` **ignores `settings`** and returns an empty `&DatasourceInstance{}` (`:49-51`). No `settings.JSONData`, `settings.URL`, or secret is ever read |
| `pkg/main_test.go:47-75` | `CheckHealth` test asserts the fixed OK message — confirms behavior is config-independent |

Notably absent (unlike most datasources): no `src/types.ts`, no
`pkg/models/settings.go`, and no `LoadSettings`. Nothing in the plugin parses
instance settings.

### External components

The plugin depends only on the four `@grafana/*` packages below (from
`package.json`), all via `catalog:`. It does **not** depend on
`@grafana/plugin-ui`, so none of the shared HTTP-settings / `Auth` editor
components are present — this is why there are no url / basicAuth / TLS fields.

| Package | `catalog:` range (`.yarnrc.yml`) | Resolved (`yarn.lock`) | What was read / used |
| --- | --- | --- | --- |
| `@grafana/data` | `^11.6.7` | `11.6.14` | `DataSourceJsonData` (base jsonData interface — `Config` extends it with no members), `DataSourcePlugin`, `DataSourcePluginOptionsEditorProps`, `DataQuery`, `DataSourceInstanceSettings`, `MetricFindValue`, `QueryEditorProps` — none write config storage |
| `@grafana/runtime` | `^11.6.7` | `11.6.14` | `DataSourceWithBackend` — query proxy base class; reads no plugin-specific settings |
| `@grafana/ui` | `^11.6.7` | `11.6.14` | Declared dependency but **not imported** by `module.tsx`; no editor components used |
| `@grafana/schema` | `^11.6.7` | `11.6.14` | Declared dependency but **not imported** by `module.tsx` |

## Field provenance

| Schema `id` | Storage key | Target | Label source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- |
| `secureJsonData_apiKey` | `apiKey` | `secureJsonData` | none upstream — synthetic placeholder (`label: "API key"` chosen for the entry) | `string` (arbitrary; the plugin reads no value) | **Placeholder only.** Upstream defines no secrets (`src/module.tsx:14` `SecureConfig = {}`). No `role` (the plugin uses no auth). No `requiredWhen` (nothing is required). No `ui` block (the editor renders no input for it) |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `secureJsonData_apiKey` | `apiKey` | `secureJsonData` | — (not rendered; placeholder) | **No** — never read by the plugin |

### Frontend-only settings

None. The editor persists nothing.

### Backend-only settings

None. The backend reads nothing.

### Excluded settings

None applicable. The plugin uses no `@grafana/plugin-ui` `Auth` component and no
Secure Socks Proxy field, so there is nothing to exclude.

## Where the types are defined

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Config` (`= {} & DataSourceJsonData`, blank), `SecureConfig` (`= {}`, blank) | `src/module.tsx:13-14` | plugin (`grafana/plugins-private`, `plugins/grafana-helloworld-datasource`) |
| `DataSourceJsonData` (base interface: `authType?`, `defaultRegion?`, `profile?`, `manageAlerts?`, `alertmanagerUid?`, …) — none written by this plugin | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^11.6.7` (resolved `11.6.14`) |
| `DataSourcePluginOptionsEditorProps` (the `ConfigEditor` props type; carries `options` / `onOptionsChange`, unused for storage here) | `packages/grafana-data/src/` | `@grafana/data` `^11.6.7` |
| `DataSourceInstanceSettings<Config>` (constructor arg; unused for storage) | `packages/grafana-data/src/` | `@grafana/data` `^11.6.7` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `PluginHost`, `DatasourceInstance` (empty), `PluginId` constant — none unmarshal `settings.JSONData` | `pkg/main.go` | plugin |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, root fields — all ignored by this plugin) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` |
| `datasource.NewInstanceManager`, `datasource.Serve`, `instancemgmt.Instance` | `backend/datasource`, `backend/instancemgmt` | `github.com/grafana/grafana-plugin-sdk-go` |

The models in this entry flatten the (empty) configuration into a single Go
`Config` type that carries only `DecryptedSecureJSONData` for the placeholder
secret. `settings.ts` keeps the three canonical TypeScript types; `RootConfig`
and `JsonDataConfig` are blank objects (`Record<string, never>`) because the
plugin stores nothing at root or in jsonData.

## Modeling decisions

- **`RootConfig` and `JsonDataConfig` are blank objects.** The plugin stores
  nothing at the datasource root or in jsonData (`Config = {} &
  DataSourceJsonData`, `src/module.tsx:13`).
- **A truly-empty entry is impossible in this repo — a placeholder is forced.**
  Two independent, machine-checked constraints — both in code I cannot modify
  (outside `registry/<plugin_id>/`) — require at least one field, and
  specifically at least one `secureJsonData` field:
  1. `dsconfig.Schema.Validate()` (`dsconfig/schema.go:59-61`) rejects an empty
     `fields` array with the error `fields is required`. This is invoked by
     `ParseAndResolveSchemaJSON` and `NewSDKSchema`, so even `go generate`
     fails without a field. (Verified empirically — see
     [Validation performed](#validation-performed).)
  2. The shared conformance suite requires a secret: `schema.PluginUnderTest`
     rejects empty `SecureKeys` (`schema/plugin_runner.go:54`) — and this check
     runs **before** the `-generateArtifacts` branch, so `go generate` fails
     too — and `SchemaRoundTrip` asserts the settings schema's `SecureValues`
     is non-empty (`schema/conformance.go:94`). `SecureValues` is populated
     only from `secureJsonData` fields (`dsconfig/convert.go:67-73`).

  Therefore the schema declares exactly one `secureJsonData` field, `apiKey`.
  It has no `role`, no `requiredWhen`, and no `ui` block, and its `description`
  states plainly that the plugin never reads it. Root and jsonData remain empty.
- **Why `secureJsonData` and not a root/jsonData field.** A single non-secure
  field would satisfy constraint (1) but not (2), so the one required field must
  be a secret. Keeping it in `secureJsonData` also keeps `RootConfig` and
  `JsonDataConfig` genuinely blank and makes `JSONDataMatchesStruct` compare two
  empty sets (schema jsonData fields = 0; `Config` json-tagged fields = 0).
- **`Config` carries only `DecryptedSecureJSONData`.** No jsonData fields (none
  exist) and no root fields (the backend reads none), mirroring the fact that
  the plugin's own code never touches `settings`.
- **`LoadConfig` still runs parse → `ApplyDefaults` → `Validate`.** `ApplyDefaults`
  is a no-op (the editor writes no defaults) and `Validate` always returns `nil`
  (nothing is required). Both are kept exported for uniformity with other
  entries. Malformed `JSONData` is still surfaced as a `parse jsonData` error
  for robustness, even though the plugin itself never unmarshals it.
- **No `docURL`.** `src/plugin.json:19` has an empty `info.links` array.
- **No groups / relationships.** The editor has no sections and there are no
  related fields; both arrays are omitted (they are optional in the schema).
- **Instructions are informational.** Three `llm`-tagged entries record that the
  plugin has no config surface, that `apiKey` is a harness-forced placeholder,
  and that health/query behavior is static and config-independent.

## Settings examples matrix (`schema.go`)

The plugin has no configuration, so there is a single default example.

| Example | `jsonData` | `secureJsonData` |
| --- | --- | --- |
| `""` (default) | `{}` (empty) | `apiKey` (empty string) |

Secret placeholders elsewhere in this entry (tests, docs) use obviously-fake
angle-bracket values (for example `<placeholder-not-read-by-plugin>`); the plugin
ignores them regardless.

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx, settings) (Config, error)` runs the standard three-phase flow:

1. **Parse** — initialize `DecryptedSecureJSONData`; unmarshal `settings.JSONData`
   into `Config` (a no-op for storage since `Config` has no jsonData fields, but
   malformed JSON is reported as `parse jsonData`); copy the placeholder secret
   by known key.
2. **`ApplyDefaults`** — no-op (guarded by `TestApplyDefaults`).
3. **`Validate`** — always `nil` (the plugin requires nothing).

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels. `ApplyDefaults` and `Validate` are
exported separately so callers can compose the phases themselves.

## Upstream findings and modeling caveats

1. **Zero config surface.** The `ConfigEditor` (`src/module.tsx:32-34`) renders a
   static string and never calls `onOptionsChange`; the backend
   (`pkg/main.go`) ignores instance settings. Nothing is configurable. This is
   expected for a "hello world" sample/template.
2. **`secureJsonData.apiKey` is synthetic.** It does not exist upstream
   (`SecureConfig = {}`). It is introduced solely to satisfy the dsconfig
   validator and the conformance suite (see [Modeling decisions](#modeling-decisions)).
   Downstream consumers (including LLMs) should treat it as **unused**: setting
   it has no effect. This is called out in the field `description`, in the
   `settings.ts` / `settings.go` doc comments, and in two `llm` instructions.
3. **Query/variable/annotation editors are also stubs.** `QueryEditor` and
   `VariablesEditor` (`src/module.tsx:37-44`) render static strings, and
   `annotations = {}`. Out of scope for a config schema, but confirms the whole
   plugin is a scaffold.
4. **No docs URL upstream.** `info.links` is empty (`src/plugin.json:19`), so no
   `docURL` is recorded.

## Validation performed

- **Empirical zero-field probe** (documenting the forced-placeholder rationale):
  `dsconfig.ParseAndResolveSchemaJSON` and `dsconfig.NewSDKSchema` on a
  zero-field schema both return `fields is required`; a one-`secureJsonData`-field
  schema yields `SecureValues = 1` with an empty settings spec (no
  `secureJsonData` in the spec). The scratch probe was removed after the
  experiment.
- `go generate ./...` inside this directory — writes the three `*.gen.json`
  artifacts; passes.
- Shared conformance suite (`schema.RunPluginTests`) — all 8 subtests pass:
  `BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`,
  `SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`, `JSONDataMatchesStruct`,
  `JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings`.
- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator) —
  passes (via `ConfigSchemaValid`).
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  with `additionalProperties: false` — passes under both Draft 7 (the schema's
  declared draft) and Draft 2020-12 (`jsonschema` 4.26.0).
- `gofmt -l .`, `go vet ./...`, `go test ./...` inside `registry/` — clean; all
  entries (including this one) pass.
- The sibling workspace modules (`dsconfig`, `schema`) still `go build ./...` and
  `go test ./...` cleanly.
- `tsc --noEmit --strict` on `settings.ts` (TypeScript 5) — clean (verified with
  a deliberately-broken control file to confirm the compiler actually ran).
