# grafana-oracle-datasource

Declarative configuration schema for the Oracle Database datasource plugin
(`grafana-oracle-datasource`), an enterprise plugin maintained in the private
`grafana/plugins-private` monorepo.

## Upstream researched

- **Monorepo**: `github.com/grafana/plugins-private`
- **Plugin path**: `plugins/grafana-oracle-datasource/`
- **Commit SHA**: `267f4937806ed6404b6628d13ae358a5d308e376`
- **Plugin version**: `3.4.4` (`package.json:3`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, option
labels/values/descriptions, section titles, defaults, validations, dependency and required-when
expressions, storage keys, storage targets, value types, group titles, and instructions — is
traceable to a specific `file:line` in the upstream plugin at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research (the monorepo is already on disk; do **not** clone):

```bash
git -C /path/to/plugins-private rev-parse HEAD   # 267f4937806ed6404b6628d13ae358a5d308e376
cd /path/to/plugins-private/plugins/grafana-oracle-datasource
```

If the monorepo advances past this SHA, re-diff the sources listed under
[Sources researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, effects, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: root `URL` + jsonData fields + `DecryptedSecureJSONData`), `PluginID`, typed enums/`SecureJsonDataKey` constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each connection/auth variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, `Validate`, `EffectiveTNSNamesEntry`, and the `SettingsExamples` shape |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`). Package name: `oracledatasource`.

## Sources researched

Every source below was read at the pinned monorepo SHA
(`267f4937806ed6404b6628d13ae358a5d308e376`), plus external editor components at the exact
versions the workspace catalog (`plugins-private/.yarnrc.yml`) pins for the `catalog:` protocol.

### Plugin (`plugins/grafana-oracle-datasource`)

| File | What was read |
| --- | --- |
| `src/plugin.json:3-5,24` | `type` (datasource), `name` (`Oracle Database` → `pluginName`), `id` (`grafana-oracle-datasource` → `pluginType`), `info.links[0].url` (`docURL`) |
| `src/components/ConfigEditor.tsx:164-183` | `connectionOptions` (Connection methods select labels/values/descriptions), `onConnectionChange`, `connectionsValue` derivation |
| `src/components/ConfigEditor.tsx:140-162` | `toggleUseTNSNames` / `toggleKerberos` — the multi-field writes captured as virtual `effects` |
| `src/components/ConfigEditor.tsx:185-204` | `authOptions` (auth-type select labels/values/descriptions), `onAuthChange`, `authValue`, Kerberos-only-when-TNS filter (`:202`) |
| `src/components/ConfigEditor.tsx:206-293` | Connection `ConfigSection` (title/description), `ConfigSubSection` "Connection methods", Host / Database / TNSName fields (labels, tooltips, placeholders, `required`, storage keys, conditional render) |
| `src/components/ConfigEditor.tsx:296-378` | `<Auth>` custom method "Oracle authentication" (label/description), auth-type `Select` (`disabled` when not TNS, `:319`), User / Password fields |
| `src/components/ConfigEditor.tsx:344-353` | Password `LegacyForms.SecretFormField` — tooltip, `onChange`/`onReset` storage key `password`; no `label`/`placeholder` props (both default to "Password" in `@grafana/ui`) |
| `src/components/ConfigEditor.tsx:98-114` | `useEffect` defaulting `timezone_name` → `UTC` and migrating `url` → `tnsNamesEntry` |
| `src/components/ConfigEditor.tsx:382-513` | Additional Settings `ConfigSection` (title/description) and its subsections: Timezone, Connections (pool size), Timeout, Prefetch Row Size, Row Limit |
| `src/components/ConfigEditor.tsx:404-448` | Conditional `Secure Socks Proxy` subsection (`enableSecureSocksProxy`) — deliberately excluded from this entry |
| `src/components/ConfigEditor.tsx:516-525` | "User Permission" info box (SELECT-only guidance, captured as an instruction) |
| `src/types.ts:6-23` | `OracleOptions` (jsonData) and `OracleSecureOptions` (secureJsonData) type members |
| `pkg/models/settings.go:16-33` | `DBConnectionOptions` struct + json tags (backend jsonData shape) |
| `pkg/models/settings.go:35-39` | Backend defaults (`defaultConnectionPoolSize=50`, `defaultDataProxyTimeout=120`, `defaultRowLimit=1000000`) |
| `pkg/models/settings.go:41-88` | `ConnectionOptions` loader: TZName→UTC default, pool/timeout/rowLimit defaults, password from decrypted secrets (`:77`), root URL read (`:79`), legacy TNSNames-in-URL fallback (`:82-84`) |
| `pkg/models/settings.go:90-157` | `buildConnectionString` / `GetGoDriverConnStr` — how URL/Database/TNSNamesEntry/User/Password/TZName/PrefetchRowsCount feed the connection string per connection+auth mode |
| `pkg/models/settings.go:159-195` | `getProxyTimeout` / `getConnectionPoolSize` — env-var overrides `GF_DATAPROXY_TIMEOUT` / `GF_PLUGINS_ORACLE_DATASOURCE_POOLSIZE` |
| `pkg/models/settings_test.go:68-216` | Backend expectations for each connection/auth mode + defaults + legacy migration (basis for the `LoadConfig` tests) |
| `pkg/oracle/utils.go:10` | Second read site of the decrypted `password` secret |
| `src/datasource.ts:101-119` | `use_dst` used only as a per-query/annotation default, confirming it is not a datasource-level setting |
| `package.json:29-40` | External component versions (all `catalog:`) |

### External editor components

Read at the versions the workspace catalog pins (`plugins-private/.yarnrc.yml:16-26`); the plugin's
`package.json` references them via `catalog:`.

| Component | Version (catalog) | What was read |
| --- | --- | --- |
| `LegacyForms.SecretFormField` | `@grafana/ui` `^11.6.7` | Password field: default `label` and `placeholder` are both "Password" (no explicit props passed); `isConfigured`/`onReset` semantics for write-only secrets |
| `Select`, `Input`, `Switch`, `InlineField`, `Alert`, `FieldValidationMessage` | `@grafana/ui` `^11.6.7` | Prop names (`label`, `placeholder`, `value`, `onChange`, `options`, `disabled`, `required`) so we knew which UI attributes to record |
| `ConfigSection`, `ConfigSubSection`, `DataSourceDescription`, `Auth` | `@grafana/plugin-ui` `^0.13.1` | Section titles/descriptions; the `Auth` custom-method card ("Oracle authentication" label + description) that wraps the auth-type select |
| `onUpdateDatasourceOption`, `onUpdateDatasourceJsonDataOption`, `onUpdateDatasourceJsonDataOptionSelect`, `onUpdateDatasourceSecureJsonDataOption`, `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps` | `@grafana/data` `^11.6.7` | Storage-key semantics: `onUpdateDatasourceOption(props,'url')` writes root `url`; the jsonData/secure helpers write to `jsonData`/`secureJsonData` |
| `config`, feature toggles | `@grafana/runtime` `^11.6.7` | Gate for the excluded Secure Socks Proxy subsection |
| `getAllTimezones` | `countries-and-timezones` `2.3.1` | Supplies the dynamic Time zone select options (why `jsonData_timezone_name` is a `select` with no static options) |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where its label,
placeholder, tooltip/description, default, storage key, and value type are defined.
`CE` = `src/components/ConfigEditor.tsx`, `S` = `pkg/models/settings.go`, `T` = `src/types.ts`.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `virtual_connectionType` | — | virtual | `CE:214` (`ConfigSubSection title="Connection methods"`) | Options `CE:164-176`; `defaultValue` from `connectionsValue` `CE:183` | Editor-local union, `CE:531-534` | `storage.computed.read` mirrors `CE:183`; effects mirror `toggleUseTNSNames` `CE:140-152` |
| `jsonData_useTNSNamesBasedConnection` | `useTNSNamesBasedConnection` | `jsonData` | — (managed by `virtual_connectionType`) | — | `bool` `S:26`; TS `boolean` `T:11` | Tagged `managed-by:virtual_connectionType` |
| `root_url` | `url` | `root` | `CE:229` (`label="Host"`) | `CE:239` (`placeholder="Host"`); tooltip `CE:231` | `s.URL` `S:79`; TS `options.url` `CE:240` | `dependsOn` from `CE:225`; `requiredWhen` from `CE:230` + backend `S:115` |
| `jsonData_database` | `database` | `jsonData` | `CE:248` (`label="Database"`) | `CE:258` (`placeholder="Database"`); tooltip `CE:249` | `string` `S:24`; TS `string` `T:7` | `dependsOn`/`requiredWhen` from `CE:225,250` |
| `jsonData_tnsNamesEntry` | `tnsNamesEntry` | `jsonData` | `CE:273` (`label="TNSName"`) | `CE:283` (`placeholder="server/DB"`); tooltip `CE:275` | `string` `S:27`; TS `string` `T:12` | `dependsOn`/`requiredWhen` from `CE:270,274`; legacy fallback `S:82-84` |
| `virtual_authType` | — | virtual | `CE:304` (`Auth` method `label: 'Oracle authentication'`) | Options `CE:190-202`; `defaultValue` from `authValue` `CE:204` | Editor-local union, `CE:535-538` | `disabledWhen` mirrors `disabled={!useTNSNamesBasedConnection}` `CE:319`; effects mirror `toggleKerberos` `CE:153-162` |
| `jsonData_useKerberosAuthentication` | `useKerberosAuthentication` | `jsonData` | — (managed by `virtual_authType`) | — | `bool` `S:28`; TS `boolean` `T:13` | Role `auth.discriminator`; tagged `managed-by:virtual_authType` |
| `jsonData_user` | `user` | `jsonData` | `CE:327` (`label="User"`) | `CE:337` (`placeholder="User"`); tooltip `CE:328` | `string` `S:25`; TS `string` `T:10` | Role `auth.basic.username`; `dependsOn`/`requiredWhen` from `CE:323,330` |
| `secureJsonData_password` | `password` | `secureJsonData` | `@grafana/ui` `SecretFormField` default "Password" (no prop at `CE:344-353`) | Default "Password"; tooltip `CE:349` | `s.DecryptedSecureJSONData["password"]` `S:77`; TS `OracleSecureOptions.password` `T:22` | Role `auth.basic.password`; `dependsOn`/`requiredWhen` from `CE:323,356` |
| `jsonData_timezone_name` | `timezone_name` | `jsonData` | `CE:389` (`label="Time zone"`) | Options dynamic (`getAllTimezones`, `CE:79-84`); default `UTC` `CE:108`, `S:51-53` | `string` `S:22`; TS `string` `T:8` | `select` with no static options (runtime-generated) |
| `jsonData_connectionPoolSize` | `connectionPoolSize` | `jsonData` | `CE:453` (`label="Connection Pool size"`) | tooltip `CE:454`; default 50 `S:36,63-65` | `int` `S:30`; TS `number` `T:15` | Env override `GF_PLUGINS_ORACLE_DATASOURCE_POOLSIZE` `S:180` |
| `jsonData_dataProxyTimeout` | `dataProxyTimeout` | `jsonData` | `CE:469` (`label="Dataproxy Timeout"`) | tooltip `CE:470`; default 120 `S:37,68-70` | `int` `S:29`; TS `number` `T:16` | Env override `GF_DATAPROXY_TIMEOUT` `S:162` |
| `jsonData_prefetchRowsCount` | `prefetchRowsCount` | `jsonData` | `CE:485` (`label="Prefetch Row Size"`) | tooltip `CE:486`; no default | `int` `S:31`; TS `number` `T:17` | Only applied to conn string when > 0 `S:130-156` |
| `jsonData_rowLimit` | `rowLimit` | `jsonData` | `CE:501` (`label="Row Limit"`) | tooltip `CE:502`; default 1000000 `S:38,73-75` | `int64` `S:32`; TS `number` `T:18` | Applied when <= 0 |
| `jsonData_use_dst` | `use_dst` | `jsonData` | — (no UI) | — | `bool` `S:23`; TS `boolean` `T:9` | Tagged `backend-only`; parsed but unused (see findings #4) |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `virtual_connectionType` | — (virtual) | — | Connection methods | — (editor-local state) |
| `jsonData_useTNSNamesBasedConnection` | `useTNSNamesBasedConnection` | `jsonData` | — (managed) | Yes |
| `root_url` | `url` | `root` | Host | Yes |
| `jsonData_database` | `database` | `jsonData` | Database | Yes |
| `jsonData_tnsNamesEntry` | `tnsNamesEntry` | `jsonData` | TNSName | Yes |
| `virtual_authType` | — (virtual) | — | Oracle authentication | — (editor-local state) |
| `jsonData_useKerberosAuthentication` | `useKerberosAuthentication` | `jsonData` | — (managed) | Yes |
| `jsonData_user` | `user` | `jsonData` | User | Yes |
| `secureJsonData_password` | `password` | `secureJsonData` | Password | Yes |
| `jsonData_timezone_name` | `timezone_name` | `jsonData` | Time zone | Yes |
| `jsonData_connectionPoolSize` | `connectionPoolSize` | `jsonData` | Connection Pool size | Yes |
| `jsonData_dataProxyTimeout` | `dataProxyTimeout` | `jsonData` | Dataproxy Timeout | Yes |
| `jsonData_prefetchRowsCount` | `prefetchRowsCount` | `jsonData` | Prefetch Row Size | Yes |
| `jsonData_rowLimit` | `rowLimit` | `jsonData` | Row Limit | Yes |
| `jsonData_use_dst` | `use_dst` | `jsonData` | — (no UI) | Parsed, unused |

Totals: **15 fields** (2 virtual, 1 root, 11 jsonData, 1 secureJsonData), **3 groups**,
**2 relationships**, **6 instructions**, **6 settings examples**, **1 secret key**.

### Frontend-only settings

None. Both discriminator booleans (`useTNSNamesBasedConnection`, `useKerberosAuthentication`) are
written by the editor **and** read by the backend (`pkg/models/settings.go:26,28`).

### Backend-only settings

- **`use_dst`** has no editor UI. It is declared in the frontend `OracleOptions` type
  (`src/types.ts:9`) and parsed by the backend into `DBConnectionOptions.DSTEnabled`
  (`pkg/models/settings.go:23`), but it is never referenced when building the connection string.
  `use_dst` is really a per-query / annotation option (`src/types.ts:30`, `src/datasource.ts:109`).
  See [Upstream findings](#upstream-findings) #4.

### Excluded

- **`enableSecureSocksProxy`** (Secure Socks Proxy). The editor conditionally renders the toggle
  (`CE:404-448`) and the backend honors it via `httpclient.Options`/`proxy.New`
  (`pkg/models/settings.go:59`), but per the registry convention this field is omitted. Note the
  backend reads the proxy state from `HTTPClientOptions`, not from a `DBConnectionOptions` json
  field, so excluding it introduces no jsonData/struct drift.

## Where the types are defined

Only config type/field definitions are listed. UI components (`ConfigSection`, `Auth`,
`SecretFormField`, …) and functions/helpers (`ConnectionOptions`, `onUpdateDatasourceOption`, …)
are omitted even where they are the reason a field exists.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `OracleOptions` (jsonData), `OracleSecureOptions` (secureJsonData), `OracleQuery.use_dst` | `src/types.ts:6-31` | plugin (`grafana-oracle-datasource`) |
| `DataSourceJsonData` (base interface `OracleOptions` extends) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^11.6.7` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `DBConnectionOptions` (jsonData tags + computed fields) | `pkg/models/settings.go:16-33` | plugin (`grafana-oracle-datasource`) |
| `backend.DataSourceInstanceSettings` (carries `URL`, `JSONData`, `DecryptedSecureJSONData`; only `URL` is read here) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` |
| `httpclient.Options` / `proxy` (Secure Socks Proxy state — excluded field) | `backend/httpclient`, `backend/proxy` | `github.com/grafana/grafana-plugin-sdk-go` |

This entry flattens that spread into a single Go `Config` (root `URL` + jsonData fields +
`DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant list. The `ConnectionType`
and `AuthType` Go enums are derived from the editor's selectors (`ConfigEditor.tsx:531-538`); the
backend has no named type for them (it stores raw booleans).

## Modeling decisions

- **Two virtual selectors**: the editor renders two `Select`s backed by React state rather than
  storage. `virtual_connectionType` (Host with TCP Port / TNSNames Entry) derives from
  `jsonData.useTNSNamesBasedConnection`; `virtual_authType` (Basic / Kerberos) derives from
  `jsonData.useKerberosAuthentication`. Each is modeled `kind: "virtual"` with a
  `storage.computed.read` expression and `effects` mirroring the editor's change handlers
  (`toggleUseTNSNames` clears `useKerberosAuthentication` when switching to TCP, `CE:140-152`;
  `toggleKerberos`, `CE:153-162`). The driven booleans carry `managed-by:<virtual_id>` tags.
- **`disabledWhen` on the auth selector**: the editor disables the auth-type `Select` when not
  using TNSNames (`disabled={!useTNSNamesBasedConnection}`, `CE:319`) and filters the Kerberos
  option out entirely for TCP (`CE:202`). This is encoded as
  `disabledWhen: "virtual_connectionType == 'tcp'"` on `virtual_authType`.
- **`requiredWhen` references storage fields**: `dependsOn` mirrors editor visibility (and may
  reference the virtual selectors), while `requiredWhen` references the storage discriminators
  (`jsonData_useTNSNamesBasedConnection`, `jsonData_useKerberosAuthentication`) so the backend
  data contract is expressible without the virtual fields.
- **Host in root, everything else in jsonData**: unusually for a SQL datasource, only the host is
  a root field (`url`); `user`, `database`, and `tnsNamesEntry` all live in `jsonData`
  (`CE:259,284,338`). `Config` carries `URL` tagged `json:"-"` and mirrors the backend
  `DBConnectionOptions` json tags for the rest.
- **Password label/placeholder from the SDK default**: the `SecretFormField` is rendered without
  `label`/`placeholder` props, so both fall back to the `@grafana/ui` default "Password".
- **Time zone as an option-less `select`**: the timezone list is generated at runtime from
  `countries-and-timezones` (`CE:79-84`), so enumerating hundreds of IANA zones would be
  unfaithful and brittle. Modeled as `select` with `defaultValue: "UTC"` and no static options.
- **Backend-parity defaults in `ApplyDefaults`**: the plugin applies its defaults inline in
  `ConnectionOptions` (not the editor), so `ApplyDefaults` mirrors them: `timezone_name`→UTC,
  `connectionPoolSize`→50, `dataProxyTimeout`→120, `rowLimit`→1000000. `prefetchRowsCount` is
  left at zero (only used when > 0). Env-var overrides are not resolved (runtime-only).
- **`Validate` follows the backend, not just the editor**: it requires host+database for TCP,
  a tnsNamesEntry (or the legacy root-url fallback) for TNSNames, and user+password for Basic;
  Kerberos needs no credentials. It does **not** reject Kerberos + Host with TCP Port because the
  backend accepts it (`pkg/models/settings_test.go:116-129`), even though the editor hides that
  combination — see findings #2.
- **Secure Socks Proxy excluded** per registry convention.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` from the embedded
`dsconfig.json`: root `url` plus a nested `jsonData` object become the OpenAPI settings `spec`,
`password` becomes the single `secureValues` entry, and the two virtual fields are skipped.

`SettingsExamples()` provides the default configuration plus one example per connection/auth
variant and the legacy storage shape. Each is a full instance-settings object with the root `url`
at the top level, plugin config under `jsonData`, and the write-only `password` under
`secureJsonData` (the default example — keyed `""` — carries an empty password):

| Example | Connection | Auth | Notable jsonData / secure |
| --- | --- | --- | --- |
| `""` (default) | Host with TCP Port | Basic | schema defaults; `password` empty |
| `basicAuthTcp` | Host with TCP Port | Basic | `url`, `database`, `user`, `password` |
| `basicAuthTns` | TNSNames Entry | Basic | `tnsNamesEntry`, `user`, `password` |
| `kerberosTns` | TNSNames Entry | Kerberos | `tnsNamesEntry`; no user/password |
| `tunedSettings` | Host with TCP Port | Basic | pool/timeout/prefetch/rowLimit/timezone knobs |
| `legacyTnsInUrl` | TNSNames Entry (legacy) | Basic | `tnsNamesEntry` in root `url` (v3.3.0 shape) |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase flow and returns a fully-defaulted, validated `Config`:

1. **Parse** — copy root `URL` from `settings.URL`, unmarshal `jsonData` into `Config`, copy the
   decrypted `password` into `DecryptedSecureJSONData`, and apply the legacy TNSNames-in-URL
   fallback (`pkg/models/settings.go:82-84`).
2. **`ApplyDefaults`** — fill the curated backend-parity defaults (UTC, pool 50, timeout 120,
   rowLimit 1000000).
3. **`Validate`** — enforce the connection + auth contract described above. Errors are joined so
   every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels. `ApplyDefaults` and `Validate` stay exported for callers
that assemble a `Config` directly (provisioning preview, tests). This is the intended shape for
the plugin's own `ConnectionOptions` to sync to.

## Upstream findings

Potential bugs, misleading UX, and consistency issues found while researching upstream. All are
preserved verbatim in the schema — the schema records what the plugin **does**, not what it
**should** do.

1. **Host stored at root while user/database are in jsonData.** Unlike the core SQL datasources
   (which put `user` at root), Oracle writes `user` and `database` to `jsonData`
   (`CE:259,338`) and only `url` to root (`CE:240`). Provisioning must not put `user` at root.
2. **Kerberos + Host with TCP Port is reachable via provisioning but not the editor.** The editor
   filters Kerberos out for TCP connections (`CE:202`) and disables the selector (`CE:319`), yet
   the backend builds a valid `host/database` Kerberos connection string
   (`pkg/models/settings.go:100-102`, verified at `settings_test.go:116-129`). A provisioned
   datasource can therefore use a combination the UI forbids.
3. **`required` markers in the editor are cosmetic.** The Host/Database/User/TNSName inputs carry
   `required` and inline `FieldValidationMessage`s (`CE:230,250,274,330`), but nothing blocks save
   and the backend never hard-fails on missing values — it builds a connection string with empty
   segments and lets Oracle reject the login. `requiredWhen` encodes the intent.
4. **`use_dst` is dead weight at datasource level.** It is declared in `OracleOptions`
   (`src/types.ts:9`) and parsed into `DBConnectionOptions.DSTEnabled` (`S:23`) but never used
   when building either connection string (`S:90-157`). It is only meaningful per-query
   (`src/types.ts:30`, `src/datasource.ts:109`).
5. **Time zone only affects display, never the wire.** The backend always connects with
   `loc=<tz>` but the comment and code note it "always connect using UTC and shift later"
   (`S:110-115`); the default is UTC (`S:51-53`, `CE:108`). Enabling Secure Socks Proxy makes the
   editor warn that time-zone config is ignored entirely (`CE:441`).
6. **Env-var overrides only apply when the stored value is 0.** `connectionPoolSize` and
   `dataProxyTimeout` read `GF_PLUGINS_ORACLE_DATASOURCE_POOLSIZE` / `GF_DATAPROXY_TIMEOUT`
   (`S:159-195`) **only** when the jsonData value is 0 (`S:63-70`); a stored non-zero value wins
   over the environment, the opposite of what "Takes precedence over environment variable" in the
   tooltips (`CE:454,470`) suggests.
7. **Legacy TNSNames migration is one-way and duplicated.** Both the editor `useEffect`
   (`CE:100-101,110`) and the backend (`S:82-84`) copy root `url` → `tnsNamesEntry` when a
   TNSNames datasource has no entry, but nothing clears the stale root `url` afterward.
8. **Grafana Cloud limitation is UX-only.** The TNSNames and Kerberos option descriptions warn
   "not supported in Grafana Cloud" (`CE:174,200`), but nothing in the settings model enforces it.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes
  (via the conformance suite).
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json) with
  `ajv --spec=draft7 --strict=false --all-errors -c ajv-formats` (the CI command) — `valid`.
- `go test ./...` in the `registry` module — passes (schema round-trip, `SchemaArtifactInSync`,
  spec/secure separation, jsonData/struct parity both directions, secure-key parity, and the
  `LoadConfig`/`ApplyDefaults`/`Validate`/`EffectiveTNSNamesEntry`/`SettingsExamples` tests).
- `go build ./...`, `go vet ./...`, `gofmt -l .` in the `registry` module — clean.
- `dsconfig` and `schema` workspace modules — still build.
- `settings.ts`: `tsc --noEmit --strict` (`typescript@5.5.4`) — clean.
