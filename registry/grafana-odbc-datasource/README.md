# grafana-odbc-datasource

Declarative configuration schema for the **Sqlyze Datasource** plugin
(`grafana-odbc-datasource`) — a generic Grafana datasource that connects to SQL databases via an
ODBC driver.

## Upstream researched

- **Repo**: `github.com/grafana/plugins-private` (monorepo)
- **Ref**: `main`
- **Commit SHA**: `267f4937806ed6404b6628d13ae358a5d308e376` (2026-07-03)
- **Plugin path**: `plugins/grafana-odbc-datasource/`

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, the driver tooltip,
section titles, defaults, the required-when expression, storage keys, storage targets, value
types, group title, and instructions — is traceable to a specific `file:line` in the upstream
plugin at this SHA. See [Field provenance](#field-provenance) below.

To reproduce this research (the repo is a private monorepo; access required):

```bash
git clone https://github.com/grafana/plugins-private
cd plugins-private
git checkout 267f4937806ed6404b6628d13ae358a5d308e376
cd plugins/grafana-odbc-datasource
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under
[Sources researched](#sources-researched) before merging changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, the driver-settings array, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` (plus the `Setting` element type) |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `SecureJsonDataKey` typed constant, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each connection variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, `Validate`, and dynamic secret resolution |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`). Import path:
`github.com/grafana/dsconfig/registry/grafana-odbc-datasource` (package `odbcdatasource`).

## Sources researched

Every source below was read at the pinned monorepo SHA
(`267f4937806ed6404b6628d13ae358a5d308e376`), plus external editor components at the versions the
plugin's `package.json` resolves through the workspace `catalog:` protocol.

### Plugin (`plugins/grafana-odbc-datasource` @ 267f4937)

| File | What was read |
| --- | --- |
| `src/plugin.json:4-5,18` | `pluginName` (`name` = "Sqlyze Datasource"), `pluginType` (`id`), `docURL` (`info.links[0].url`) |
| `src/ConfigEditor.tsx:161-165` | `DataSourceDescription` (`dataSourceName="Sqlyze"`, `hasRequiredFields`) |
| `src/ConfigEditor.tsx:169-172` | `ConfigSection` title "Connection Settings" + description (used as the `settings` field description) |
| `src/ConfigEditor.tsx:174-185` | Driver `FormField`: label, `onChange(e,'driver')` storage key, placeholder, tooltip |
| `src/ConfigEditor.tsx:188-200` | Timeout `FormField`: hardcoded label "Timeout (seconds)", `onChange(e,'timeout')`, placeholder "10" |
| `src/ConfigEditor.tsx:202-212` | "Driver Settings" `<h5>` heading and Add-setting button |
| `src/ConfigEditor.tsx:73-142` | `onChangeSetting`/`onSaveSetting`/`onRemoveSetting`: how a setting object `{name,value,secure}` is written and how a secure value is routed to `secureJsonData[name]` |
| `src/ConfigEditor.tsx:259-292` | Setting row rendering: `label={s.name}`, secure rows use `type="password"` |
| `src/selectors.ts:3-35` | Editor labels/placeholders/aria-labels and the driver `ToolTip` (with the "a absolute path" typo) |
| `src/types.ts:8-22` | `Setting` (`name`/`value?`/`secure`), `ODBCSettings` (jsonData: `driver`/`timeout`/`settings`), `SecureODBCSettings` |
| `pkg/models/settings.go:15-26` | `Setting` and `Settings` (`Driver`, `DSN`, `Timeout`, `Settings`) — no json tags (case-insensitive matching) |
| `pkg/models/settings.go:28-54` | `LoadSettings`: unmarshal, per-setting secure resolution from `DecryptedSecureJSONData[name]`, `CheckDriverFileExists`, timeout default "10" |
| `pkg/models/settings.go:68-114` | `CheckDriverFileExists`: empty-driver rejection, `{...}` DSN-alias regex, path existence/executable checks |
| `pkg/database/connect.go:16-73` | `Connect`: `strconv.Atoi(Timeout)`, ping bounded by the timeout |
| `pkg/database/connect.go:75-87` | `DSN()`: builds `Driver=<driver>;` (or `DSN=<DSN>;`) + `name=value;` per setting |
| `pkg/driver.go:27-36` | `ODBC.Connect`: only `config.JSONData` + `config.DecryptedSecureJSONData` are read — no root fields |
| `pkg/driver.go:48-55` | `Settings()`: sqlds `DriverSettings.Timeout` hardcoded to 30s (independent of `jsonData.timeout`) |
| `pkg/models/settings_test.go:14-121` | Upstream fixtures using capitalized keys (`Driver`, `DSN`, `Settings`) and a secure `Password` setting |
| `README.md`, `DEV-GUIDE-DB2.md` | Driver-settings table (`pwd` shown as the masked secret) and the DB2 walkthrough (`uid`/`pwd`) |

### External editor components

Read at the versions the plugin resolves through the monorepo `catalog:` (`.yarnrc.yml`).

| Component / type | Version (catalog) | Package | What was read |
| --- | --- | --- | --- |
| `ConfigSection`, `DataSourceDescription` | `^0.13.1` | `@grafana/plugin-ui` | Section `title`/`description` props; `DataSourceDescription` `dataSourceName`/`docsLink`/`hasRequiredFields` |
| `LegacyForms.FormField`, `IconButton` | `^11.6.7` | `@grafana/ui` | `label`/`placeholder`/`value`/`onChange`/`type` prop names driving each field's presentation |
| `DataSourcePluginOptionsEditorProps` | `^11.6.7` | `@grafana/data` | Editor props carrying `options.jsonData`/`secureJsonData`/`secureJsonFields` storage semantics |
| `DataSourceJsonData` (base of `ODBCSettings`) | `^11.6.7` | `@grafana/data` | Confirms the plugin adds only `driver`/`timeout`/`settings` on top of the base jsonData |
| `E2ESelectors` | not cataloged (swapped per Grafana version) | `@grafana/e2e-selectors` | Type wrapper around the `Components` selector object in `src/selectors.ts` |
| `css` | `11.10.6` | `@emotion/css` | Styling only (no storage impact) |

### Backend dependencies (`plugins/grafana-odbc-datasource/go.mod`)

| Module | Version | Role |
| --- | --- | --- |
| `github.com/grafana/grafana-plugin-sdk-go` | `v0.280.0` | `backend.DataSourceInstanceSettings` (`JSONData`, `DecryptedSecureJSONData`) |
| `github.com/grafana/sqlds/v5` | `v5.0.3` | SQL datasource framework (`sqlds.DriverSettings`) |
| `github.com/polytomic/odbc` | `v0.0.0-20211101212313-e87882d56ba3` | `database/sql` ODBC driver registered as `"odbc"` |
| `github.com/pkg/errors` | `v0.9.1` | error wrapping in `LoadSettings` |
| `golang.org/x/sys` | `v0.46.0` | `unix.Access` executable check in `CheckDriverFileExists` |

## Field provenance

| Schema `id` | Storage key | Target | Label source | Placeholder / default / description source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_driver` | `driver` | `jsonData` | `selectors.ts:6` (`ConfigEditor.tsx:176`) | placeholder `selectors.ts:7` (`ConfigEditor.tsx:181`); description = driver `ToolTip` `selectors.ts:33-34` (`ConfigEditor.tsx:184`) | `ODBCSettings.driver` `types.ts:15` (string); `Settings.Driver` `settings.go:22` | `requiredWhen: "true"` — `LoadSettings` rejects empty driver (`settings.go:43,69-71`) |
| `jsonData_timeout` | `timeout` | `jsonData` | `ConfigEditor.tsx:191` (hardcoded; also `selectors.ts:11`) | placeholder "10" `ConfigEditor.tsx:196`; default "10" `settings.go:50-52` | `ODBCSettings.timeout` `types.ts:16` (string); `Settings.Timeout` `settings.go:24` | role `transport.timeoutSeconds` |
| `jsonData_settings` | `settings` | `jsonData` | `ConfigEditor.tsx:203` (`<h5>Driver Settings</h5>`) | description = `ConfigSection` description `ConfigEditor.tsx:171` | `ODBCSettings.settings` `types.ts:17` (`Setting[]`); `Settings.Settings` `settings.go:25` | array of `{name,value,secure}` |
| `jsonData_settings.item.name` | `name` | (item) | `selectors.ts:20` (`ConfigEditor.tsx:218`) | placeholder `selectors.ts:22` | `Setting.name` `types.ts:9` (string) | item field |
| `jsonData_settings.item.value` | `value` | (item) | `selectors.ts:26` (`ConfigEditor.tsx:227`) | placeholder `selectors.ts:27` | `Setting.value?` `types.ts:10` (string) | absent for secure settings |
| `jsonData_settings.item.secure` | `secure` | (item) | — (lock toggle, no label; `ConfigEditor.tsx:237-246`) | default `false` | `Setting.secure` `types.ts:11` (boolean) | `true` routes the value to `secureJsonData[name]` |
| `jsonData_dsn` | `DSN` | `jsonData` | — (no editor UI) | description = backend behavior `connect.go:76-78` | `Settings.DSN` `settings.go:23` (string) | tagged `backend-only` |
| `secureJsonData_pwd` | `pwd` | `secureJsonData` | plugin `README.md` driver-settings table (`pwd`) | placeholder "Setting value" `selectors.ts:27` | dynamic — resolved by name `settings.go:33-41` | **representative** secret; role `auth.basic.password` |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_driver` | `driver` | `jsonData` | Driver | Yes (`connect.go:76`) |
| `jsonData_timeout` | `timeout` | `jsonData` | Timeout (seconds) | Yes (`connect.go:18`) |
| `jsonData_settings` | `settings` | `jsonData` | Driver Settings | Yes (`connect.go:83-85`) |
| `jsonData_settings.item.name` | `name` | (item) | Name | Yes |
| `jsonData_settings.item.value` | `value` | (item) | Value | Yes (non-secure) |
| `jsonData_settings.item.secure` | `secure` | (item) | — (lock toggle) | Yes (`settings.go:34`) |
| `jsonData_dsn` | `DSN` | `jsonData` | — (no UI) | Yes, backend-only (`connect.go:77-78`) |
| `secureJsonData_pwd` | `pwd` (dynamic) | `secureJsonData` | (setting name) | Yes (`settings.go:35`) |

### Frontend-only settings

None. Every editor field (`driver`, `timeout`, `settings`, and each setting's `secure` flag) is
read by the backend when building the connection string.

### Backend-only settings

- **`DSN`** exists in the Go `Settings` struct (`pkg/models/settings.go:23`) and is read by `DSN()`
  (`pkg/database/connect.go:77-78`) to switch the connection-string prefix from `Driver=` to
  `DSN=`, but it has no editor UI and is absent from `src/types.ts`. It can only be set via
  provisioning. Tagged `backend-only`.

## Where the types are defined

Only config type/field definitions are listed — UI components (`ConfigSection`,
`DataSourceDescription`, `FormField`, `IconButton`) and functions/helpers (`LoadSettings`,
`CheckDriverFileExists`, `DSN`, `Connect`) are omitted even where they are the reason a field exists.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `ODBCSettings` (jsonData: `driver`/`timeout`/`settings`), `Setting` (`name`/`value?`/`secure`), `SecureODBCSettings` (dynamic secret map) | `src/types.ts:8-22` | plugin (`grafana-odbc-datasource`) |
| `DataSourceJsonData` (base interface `ODBCSettings` extends) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^11.6.7` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (`Driver`, `DSN`, `Timeout`, `Settings`), `Setting` (`Name`/`Value`/`Secure`) | `pkg/models/settings.go:15-26` | plugin (`grafana-odbc-datasource`) |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`; root fields like `URL` are unused by this plugin) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.280.0` |

This entry flattens that spread into a single Go `Config` (jsonData fields + `DecryptedSecureJSONData`)
plus a representative `SecureJsonDataKey` constant. `settings.ts` keeps the three canonical
TypeScript types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`).

## Modeling decisions

- **Dynamic secrets → a representative `pwd` key.** This plugin has **no fixed secret keys**: a
  driver setting whose `secure` flag is enabled stores its value in `secureJsonData` under a key
  equal to the setting's `name` (`ConfigEditor.tsx:99-114`, `settings.go:33-41`). The dsconfig
  conformance suite, however, requires at least one concrete `secureJsonData` field and matching
  `SecureKeys`. We declare a single representative secret `pwd` — the conventional password key
  from the plugin's own README driver-settings table — and document the dynamic mechanism in
  `instructions`, `settings.ts`, and `settings.go`. `LoadConfig` resolves **whatever** secure
  setting names are present at runtime, not just `pwd` (see the "dynamic secret name" test).
- **`settings` as an array of objects.** Modeled as `valueType: "array"` with an object `item`
  whose sub-fields are `name` (string), `value` (string), and `secure` (boolean), mirroring
  `Setting` (`types.ts:8-12`, `settings.go:15-19`). The Go `Config.Settings` is `[]*Setting`
  (Go kind `slice` → `array`), satisfying the type-parity conformance check.
- **Storage-key casing.** The upstream Go `Settings` struct has **no json tags** and relies on Go's
  case-insensitive unmarshal, so the editor's lowercase keys (`driver`, `timeout`, `settings`,
  `name`, `value`, `secure`) and the upstream test fixtures' capitalized keys (`Driver`, `DSN`,
  `Settings`) both work. This entry standardizes on the **lowercase keys the editor actually
  writes** and gives `Config` explicit lowercase json tags (still case-insensitive-compatible with
  the capitalized form — see the "capitalized upstream keys" test). `DSN` keeps its upstream
  capitalization because the editor never writes it and the only precedents (`Settings.DSN`, the
  test fixtures) are capitalized.
- **No root fields.** The backend reads only `JSONData` and `DecryptedSecureJSONData`
  (`pkg/driver.go:29`), so `RootConfig` is a blank object and `Config` carries no root fields.
- **`requiredWhen: "true"` on `driver`.** `LoadSettings` calls `CheckDriverFileExists`
  unconditionally (`settings.go:43`), which rejects an empty driver (`settings.go:69-71`) before
  any connection string is built — so the driver is effectively always required, which the
  schema encodes as an unconditional `requiredWhen`.
- **`Validate` omits filesystem checks.** `CheckDriverFileExists` also stats path-style drivers for
  existence and the executable bit (`settings.go:79-112`); those depend on the runtime host and
  cannot be evaluated at config-load time, so `(Config).Validate` only enforces driver-presence and
  integer-timeout.
- **Field ID naming convention.** IDs are prefixed with their storage target — `jsonData_` or
  `secureJsonData_` — followed by the camelCase storage key (`jsonData_driver`,
  `secureJsonData_pwd`); item fields use `<parent>.item.<key>`. The `key` property keeps the raw
  storage key (including `DSN`'s capitalization).

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle (served by
Grafana's datasource API server as `{apiVersion}.json`, `v0alpha1` today): the `jsonData` object
becomes the OpenAPI settings `spec`, the `pwd` secure field becomes `secureValues`, and item
fields are nested under the `settings` array items.

`SettingsExamples()` provides the default configuration plus one example per connection variant.
Each is a full instance-settings object with plugin config under `jsonData` and the write-only
secret under `secureJsonData` (obviously-fake `<your-password>` placeholders; the `""` default
carries an empty `pwd`):

| Example | `driver` | Connection settings | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | — (unset) | schema defaults (`timeout` "10") | `pwd` (empty) |
| `driverPathDB2` | `/opt/db2/clidriver/lib/libdb2.so.1` | `host`, `port`, `database`, `uid`, secure `pwd` | `pwd` |
| `driverAlias` | `{MySQLDB}` | `uid`, secure `pwd` | `pwd` |
| `connectionStringDSN` | `{TESTDB}` + backend-only `DSN=TESTDB` | `uid`, secure `pwd` | `pwd` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow and returns a fully-defaulted, validated `Config`:

1. **Parse** — unmarshal `settings.JSONData` into `Config`, then, for each driver setting whose
   `Secure` flag is set, resolve its value from `settings.DecryptedSecureJSONData[name]` (failing
   with `Missing <name>` if absent) and record it under `DecryptedSecureJSONData` keyed by the
   setting name. Mirrors `LoadSettings` (`pkg/models/settings.go:28-41`) verbatim.
2. **`ApplyDefaults`** — default `Timeout` to `"10"` when empty (`pkg/models/settings.go:50-52`).
3. **`Validate`** — enforce the runtime contract: driver present, timeout parseable by
   `strconv.Atoi` (`pkg/database/connect.go:18`). Errors are joined so every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels. `ApplyDefaults` and `Validate` are exported so callers that
assemble a `Config` directly (provisioning preview, tests that distinguish parse-level from
policy-level errors) can invoke each phase independently.

## Upstream findings

Potential bugs / discrepancies discovered while researching upstream. All behavior is preserved
verbatim in the schema — these notes exist so reviewers can reproduce each finding and decide
separately whether to fix upstream.

1. **A DSN-only configuration can never connect.** `DSN()` (`pkg/database/connect.go:77-78`)
   supports building `DSN=<value>;` with an empty `Driver`, but `LoadSettings` calls
   `CheckDriverFileExists(settings.Driver)` unconditionally first (`settings.go:43`), which returns
   `"Driver is empty"` for a blank driver (`settings.go:69-71`). So the `DSN`-only code path is
   unreachable — `jsonData.driver` must always be non-empty. Encoded as `requiredWhen: "true"` on
   `driver` and noted in the `connectionStringDSN` example description.
2. **Backend struct has no json tags.** `pkg/models/settings.go:21-26` relies on Go's
   case-insensitive matching, so the editor's lowercase keys and the test fixtures' capitalized
   keys both unmarshal. This is fragile: a stricter parser, or two settings differing only in case,
   would behave differently between frontend and backend.
3. **`jsonData.timeout` only bounds the initial ping, not queries.** `Connect` uses it to bound the
   ping (`connect.go:41,67`), but `Settings()` hardcodes the sqlds `DriverSettings.Timeout` to
   `time.Second * 30` (`pkg/driver.go:51`) regardless of the configured timeout.
4. **Driver tooltip typo, preserved verbatim.** `src/selectors.ts:34` reads "a absolute path"
   (should be "an absolute path"); it is reproduced exactly in the `driver` field `description`.
5. **Two different documentation URLs.** `src/plugin.json:18` links
   `https://grafana.com/docs/plugins/grafana-odbc-datasource` while the editor's
   `DataSourceDescription` `docsLink` (`ConfigEditor.tsx:163`) points at
   `https://grafana.com/grafana/plugins/grafana-odbc-datasource/`. This entry uses the `plugin.json`
   link for `docURL` per the authoring convention.
6. **Secure settings hard-fail load when the secret is absent.** `LoadSettings` returns
   `Missing <name>` (`settings.go:36-38`) if a `secure: true` setting has no matching
   `DecryptedSecureJSONData` entry — reachable via provisioning that lists a secure setting without
   supplying its secret. Mirrored in `LoadConfig` and covered by a test.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes
  (via the `TestSchemaConformance` suite).
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft 2020-12, strict `additionalProperties: false`) — passes.
- `go generate ./...` (regenerates the three `*.gen.json` artifacts) — clean.
- `gofmt -l .`, `go vet ./...`, `go build ./...`, `go test ./...` inside `registry/` — all pass
  (schema round-trip, artifact-in-sync, spec/secure separation, jsonData↔struct key + type parity,
  secure-key parity, and `LoadConfig`/`ApplyDefaults`/`Validate` behavior incl. dynamic secret
  resolution and case-insensitive parsing).
- `tsc --noEmit --strict settings.ts` (TypeScript `5.5.4`) — clean.
- The pre-existing `dsconfig` and `schema` workspace modules still build and test.
