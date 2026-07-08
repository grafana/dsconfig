# grafana-snowflake-datasource

Declarative configuration schema for the [Snowflake datasource plugin](https://grafana.com/docs/plugins/grafana-snowflake-datasource) (`grafana-snowflake-datasource`).

## Upstream researched

- **Monorepo**: `github.com/grafana/plugins-private`
- **Commit SHA**: `267f4937806ed6404b6628d13ae358a5d308e376` (2026-07-03)
- **Plugin path**: `plugins/grafana-snowflake-datasource`
- **Backend Go module**: `github.com/grafana/snowflake-datasource` (note: differs from the plugin id)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, descriptions (tooltips),
option labels/values, section titles, help markdown, defaults, validations, dependency and
required-when expressions, storage keys, storage targets, and value types — is traceable to a
specific `file:line` in the plugin at this SHA. See [Field provenance](#field-provenance) below.

To reproduce this research (the monorepo is large; a sparse/partial checkout is fine):

```bash
git clone https://github.com/grafana/plugins-private
cd plugins-private
git checkout 267f4937806ed6404b6628d13ae358a5d308e376
cd plugins/grafana-snowflake-datasource
```

If upstream advances past this SHA, re-diff the sources under [Sources researched](#sources-researched)
before merging changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, help, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `AuthType` + `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...`; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`). This entry's package is
`snowflakedatasource`.

## Sources researched

All read at the pinned monorepo SHA (`267f4937`), under `plugins/grafana-snowflake-datasource`.

### Plugin repo

| File | What was read |
| --- | --- |
| `src/plugin.json:3-5,30` | `pluginType` (`id`), `pluginName` (`name`), `docURL` (`info.links[0].url`), `type: datasource` |
| `src/editors/ConfigEditor.tsx:34-234` | Connection section: account, region, auth-type radio, username, and the per-auth secret fields (password / private key / passphrase / PAT / OAuth switch); conditional rendering |
| `src/editors/ConfigEditor.tsx:217-233` | OAuth "Forward OAuth Identity" `InlineSwitch` — inline label/tooltip (not from selectors) |
| `src/editors/ConfigEditor.tsx:238-286` | Environment section: role, warehouse, database, schema |
| `src/editors/ConfigEditor.tsx:287-389` | Customization section: Min Interval, Row Limit, Connection/Request Timeout, Variable Interpolation Format, Default Query, Default Variable Query |
| `src/editors/ConfigEditor.tsx:391-429` | Conditional `Secure Socks Proxy` switch — deliberately excluded from this entry |
| `src/components/SessionSettings.tsx:129-261` | "Connection settings" session-parameter editor (jsonData.settings; name/value/secure); the help-drawer `Card` content |
| `src/selectors.ts:4-151` | Every editor label, placeholder, and tooltip (`Components.ConfigEditor.*`) |
| `src/types.ts:5-70` | `SnowflakeDataSourceOptions` (jsonData), `SnowflakeSecureJsonData`, `Setting`, `AuthenticationType`, `InterpolationFormat` |
| `src/types.ts:91-146` | `defaultSQL`, `defaultVariableQuery`, `defaultTimeInterval`, `SelectableAuthenticationTypes`, `SelectableInterpolationFormat` (option labels/values, placeholder templates) |
| `src/datasource.ts:29-37` | Which jsonData fields the **frontend** reads (`timeInterval`, `defaultQuery`, `defaultVariableQuery`, `defaultInterpolation`, `rowLimit`) — proves the four frontend-only fields |
| `pkg/settings.go:17-52` | `Setting`, `Settings` (backend jsonData shape + parsed secrets), `ConnectionArgs` |
| `pkg/settings.go:56-192` | `LoadSettings`: default authType=password; password/keypair/pat/oauth branches; PEM decode/decrypt; session-setting secret resolution |
| `pkg/constants.go:5-42` | `authenticationType*` constants and the `Err*` messages mirrored in `settings.go` |
| `pkg/driver.go:51-147` | `Settings()`/`getCfg()`/`GetSessionParameters()` — how each field feeds the gosnowflake `sf.Config` and DSN (proves backend-read fields) |
| `pkg/driver.go:37-44,197-302` | GET/PUT file-transfer command blocking (security behavior) |
| `pkg/settings_test.go:14-218` | Backend expectations mirrored by `settings_test.go` (default/keypair/pat/oauth/rowLimit cases) |
| `pkg/main.go:24` | `ds.EnableRowLimit = true` |
| `package.json:30-43` | Frontend dependency versions (via `catalog:`) |
| `go.mod` | Backend dependency versions |

### External editor components

The plugin references `@grafana/*` deps via the `catalog:` protocol; versions resolved from the
monorepo root `.yarnrc.yml` catalog (plugin-local pins take precedence; none present here). The
plugin's `plugin.json` declares `grafanaDependency: ">=11.6.7-0"`.

| Component / type | Version | Source | What was read |
| --- | --- | --- | --- |
| `LegacyForms.FormField`, `LegacyForms.SecretFormField`, `Input`, `TextArea`, `Select`, `RadioButtonGroup`, `InlineFormLabel`, `InlineField`, `InlineSwitch`, `Switch`, `Button` | `@grafana/ui` `^11.6.7` | grafana/grafana `packages/grafana-ui` | Prop names (`label`, `placeholder`, `tooltip`, `value`, `onChange`, `onBlur`, `onReset`, `isConfigured`, `required`, `rows`) so the recorded UI attributes are accurate |
| `DataSourcePluginOptionsEditorProps`, `onUpdateDatasourceJsonDataOption`, `FeatureToggles` | `@grafana/data` `^11.6.7` | grafana/grafana `packages/grafana-data` | jsonData storage-key semantics of the update helper; `DataSourceJsonData` base of `SnowflakeDataSourceOptions` |
| `config`, `getTemplateSrv` | `@grafana/runtime` `^11.6.7` | grafana/grafana `packages/grafana-runtime` | `config.featureToggles.secureSocksDSProxyEnabled` gating the excluded proxy field |
| `E2ESelectors` | `@grafana/e2e-selectors` (not cataloged; swapped per Grafana version) | grafana/grafana `packages/grafana-e2e-selectors` | Type of the `selectors` export that carries all labels/placeholders/tooltips |
| `@emotion/css` (`css`), `semver` (`gte`) | `@emotion/css` `11.10.6`, `semver` (transitive) | — | Styling / version gate only; no config fields |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where its label,
placeholder, tooltip (→ `description`), value type, and storage key are defined. Labels,
placeholders, and tooltips come from `src/selectors.ts` unless noted; conditional visibility and
option lists come from `ConfigEditor.tsx` / `types.ts`.

| Schema `id` | Storage key | Target | Label / placeholder / tooltip source | Value type source | Conditional / notes |
| --- | --- | --- | --- | --- | --- |
| `jsonData_account` | `account` | `jsonData` | `selectors.ts:8,10,11-12`; rendered `ConfigEditor.tsx:83-92` | `SnowflakeDataSourceOptions.account` `types.ts:6`; `Settings.Account` `settings.go:29` | `requiredWhen:true` (data contract; not marked `required` in editor) |
| `jsonData_region` | `region` | `jsonData` | `selectors.ts:51-53`; `ConfigEditor.tsx:95-103` | `types.ts:8`; `settings.go:28` | Tooltip marks it deprecated |
| `jsonData_authType` | `authType` | `jsonData` | label/tooltip `selectors.ts:15,17`; options `types.ts:102-107`; `ConfigEditor.tsx:106-117` | `AuthenticationType` `types.ts:47-52`; `Settings.AuthType` `settings.go:39` | Role `auth.discriminator`; default `password` (`ConfigEditor.tsx:39`, `settings.go:60-61`) |
| `jsonData_username` | `username` | `jsonData` | `selectors.ts:20,22,23`; `ConfigEditor.tsx:118-131` | `types.ts:7`; `settings.go:33` | `dependsOn`/`requiredWhen` = authType != oauth (`ConfigEditor.tsx:118`) |
| `secureJsonData_password` | `password` | `secureJsonData` | `selectors.ts:26,28,29`; `ConfigEditor.tsx:132-149` | `SnowflakeSecureJsonData.password` `types.ts:31` | Role `auth.basic.password`; `dependsOn` authType==password; `requiredWhen` password/empty (`settings.go:76-85`) |
| `secureJsonData_privateKey` | `privateKey` | `secureJsonData` | `selectors.ts:32,34,35`; `ConfigEditor.tsx:150-181` (`TextArea rows={7}`) | `types.ts:32` | Role `auth.jwt.signingKey`; `dependsOn`/`requiredWhen` authType==keypair (`settings.go:88-125`) |
| `secureJsonData_privateKeyPassphrase` | `privateKeyPassphrase` | `secureJsonData` | `selectors.ts:38,40,41-42`; `ConfigEditor.tsx:182-198` | `types.ts:33` | `dependsOn` authType==keypair; optional (only for encrypted keys, `settings.go:91-92,101-115`) |
| `secureJsonData_patToken` | `patToken` | `secureJsonData` | `selectors.ts:45,47,48`; `ConfigEditor.tsx:199-216` | `types.ts:34` | Role `auth.bearer.token`; `dependsOn`/`requiredWhen` authType==pat (`settings.go:136-146`) |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | label/tooltip inline `ConfigEditor.tsx:221-222`; `InlineSwitch` `:225-229` | `types.ts:24`; `Settings.OAuthPassThrough` `settings.go:43` | Role `auth.forwardOAuthToken.enabled`; `dependsOn`/`requiredWhen` authType==oauth (`settings.go:148-151`) |
| `jsonData_settings` | `settings` | `jsonData` | section title `SessionSettings.tsx:132`; item labels/placeholders `selectors.ts:130-137`; help `SessionSettings.tsx:241-261` | `Setting[]` `types.ts:23,41-45`; `Settings.Settings` `settings.go:40` | Array of `{name,value,secure}`; consumed `settings.go:176-186`, `driver.go:72-86` |
| `jsonData_role` | `role` | `jsonData` | `selectors.ts:56,58,59-60`; `ConfigEditor.tsx:241-251` | `types.ts:9`; `settings.go:32` | Rendered under "Environment" (selector namespaced under `Connection`) |
| `jsonData_warehouse` | `warehouse` | `jsonData` | `selectors.ts:65-67`; `ConfigEditor.tsx:252-262` | `types.ts:12`; `settings.go:27` | — |
| `jsonData_database` | `database` | `jsonData` | `selectors.ts:70,72,73`; `ConfigEditor.tsx:263-274` | `types.ts:13`; `settings.go:30` | — |
| `jsonData_schema` | `schema` | `jsonData` | `selectors.ts:76-78`; `ConfigEditor.tsx:275-285` | `types.ts:14`; `settings.go:31` | — |
| `jsonData_timeInterval` | `timeInterval` | `jsonData` | `selectors.ts:100-102`; placeholder `defaultTimeInterval="10s"` `types.ts:91`; `ConfigEditor.tsx:289-306` | `types.ts:19` | **Frontend-only** (`datasource.ts:32`) |
| `jsonData_rowLimit` | `rowLimit` | `jsonData` | `selectors.ts:105,106-107`; `ConfigEditor.tsx:307-319` (number, `min=-1`) | `types.ts:25`; `Settings.RowLimit` `settings.go:44` | Consumed `driver.go:67` |
| `jsonData_loginTimeout` | `loginTimeout` | `jsonData` | `selectors.ts:110-112`; `ConfigEditor.tsx:320-334` | `types.ts:21`; `Settings.LoginTimeout` `settings.go:42` | Backend default 5s when 0 (`driver.go:97-100`) |
| `jsonData_requestTimeout` | `requestTimeout` | `jsonData` | `selectors.ts:114-117`; `ConfigEditor.tsx:335-349` | `types.ts:22`; `Settings.RequestTimeout` `settings.go:41` | Backend default 120s (`driver.go:54-57,101-104`) |
| `jsonData_defaultInterpolation` | `defaultInterpolation` | `jsonData` | `selectors.ts:94-97`; options `types.ts:109-125`; `ConfigEditor.tsx:350-360` | `InterpolationFormat` `types.ts:54-70` | **Frontend-only** (`datasource.ts:35`); default `""` (None) |
| `jsonData_defaultQuery` | `defaultQuery` | `jsonData` | `selectors.ts:83,85,86`; `ConfigEditor.tsx:361-374` (`TextArea rows={8}`) | `types.ts:16` | **Frontend-only** (`datasource.ts:33`); placeholder embeds `defaultSQL` `types.ts:92` |
| `jsonData_defaultVariableQuery` | `defaultVariableQuery` | `jsonData` | `selectors.ts:88,91,92`; `ConfigEditor.tsx:375-388` (`TextArea rows={4}`) | `types.ts:17` | **Frontend-only** (`datasource.ts:34`); placeholder embeds `defaultVariableQuery` `types.ts:93` |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_account` | `account` | `jsonData` | Account | Yes |
| `jsonData_region` | `region` | `jsonData` | Region | Yes (deprecated) |
| `jsonData_authType` | `authType` | `jsonData` | Authentication Type | Yes |
| `jsonData_username` | `username` | `jsonData` | Username | Yes |
| `secureJsonData_password` | `password` | `secureJsonData` | Password | Yes |
| `secureJsonData_privateKey` | `privateKey` | `secureJsonData` | Private key | Yes |
| `secureJsonData_privateKeyPassphrase` | `privateKeyPassphrase` | `secureJsonData` | Private key passphrase | Yes |
| `secureJsonData_patToken` | `patToken` | `secureJsonData` | Token | Yes |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | Forward OAuth Identity | Yes |
| `jsonData_settings` | `settings` | `jsonData` | Connection settings | Yes |
| `jsonData_role` | `role` | `jsonData` | Role | Yes |
| `jsonData_warehouse` | `warehouse` | `jsonData` | Warehouse | Yes |
| `jsonData_database` | `database` | `jsonData` | Database | Yes |
| `jsonData_schema` | `schema` | `jsonData` | Schema | Yes |
| `jsonData_timeInterval` | `timeInterval` | `jsonData` | Min Interval | **No — frontend-only** |
| `jsonData_rowLimit` | `rowLimit` | `jsonData` | Row Limit | Yes |
| `jsonData_loginTimeout` | `loginTimeout` | `jsonData` | Connection Timeout (sec) | Yes |
| `jsonData_requestTimeout` | `requestTimeout` | `jsonData` | Request Timeout (sec) | Yes |
| `jsonData_defaultInterpolation` | `defaultInterpolation` | `jsonData` | Variable Interpolation Format | **No — frontend-only** |
| `jsonData_defaultQuery` | `defaultQuery` | `jsonData` | Default Query | **No — frontend-only** |
| `jsonData_defaultVariableQuery` | `defaultVariableQuery` | `jsonData` | Default Variable Query | **No — frontend-only** |

### Frontend-only settings

Four jsonData fields are written by the config editor but never read by the plugin backend — they
are read only by the frontend datasource class (`src/datasource.ts:31-36`) to seed query editors
and interpolation:

- **`timeInterval`** — min interval for `$__interval` / `$__interval_ms` (`datasource.ts:32`).
- **`defaultQuery`** — seed SQL for a new panel query (`datasource.ts:33`).
- **`defaultVariableQuery`** — seed SQL for a new variable query (`datasource.ts:34`).
- **`defaultInterpolation`** — template variable interpolation format (`datasource.ts:35`).

They are tagged `frontend-only` in `dsconfig.json` and carried on the Go `Config` (required by the
jsonData/struct parity guard rail) but ignored by `LoadConfig`'s runtime contract.

### Backend-only settings

None. Every field the backend reads is also editor-visible.

### Excluded

- **`enableSecureSocksProxy`** (`jsonData.enableSecureSocksProxy`, `types.ts:27`) — the Secure Socks
  Proxy switch (`ConfigEditor.tsx:391-429`), deliberately omitted from registry entries. The
  backend does honor it (`driver.go:136-144,162-172`).
- **Dynamic secure session-setting keys** — when a session parameter is marked secure
  (`settings[i].secure == true`), its value is stored in `secureJsonData` under a key equal to the
  setting's `name` (`SessionSettings.tsx:73-86`; consumed `settings.go:176-186`). These keys are
  user-defined and cannot be enumerated, so they are not part of the fixed `SecureJsonDataConfig`.

## Where the types are defined

Only config type/field definitions are listed (UI components and functions/helpers such as
`LoadSettings`, `getCfg`, `onUpdateDatasourceJsonDataOption` are omitted even where they are the
reason a field exists).

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `SnowflakeDataSourceOptions` (jsonData), `SnowflakeSecureJsonData`, `Setting`, `SecureSnowflakeSettings`, `AuthenticationType`, `InterpolationFormat` | `src/types.ts:5-70` | plugin (`grafana-snowflake-datasource`) |
| `DataSourceJsonData` (base interface `SnowflakeDataSourceOptions` extends) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^11.6.7` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (jsonData + parsed secrets), `Setting`, `ConnectionArgs` | `pkg/settings.go:17-52` | plugin (`github.com/grafana/snowflake-datasource`) |
| `authenticationType*` discriminator constants; `Err*` messages | `pkg/constants.go:5-42` | plugin |
| `httpclient.Options` (`Settings.HttpClientOptions`), `backend.DataSourceInstanceSettings` | `backend/httpclient`, `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.290.1` |
| `*rsa.PrivateKey` (`Settings.PrivateKey`) | `crypto/rsa` | Go standard library |
| `sf.Config` (gosnowflake connection config the settings feed) | `pkg/driver.go:96-147` | `github.com/snowflakedb/gosnowflake` `v1.17.1` |

`InterpolationFormat` has **no backend equivalent** — it is a frontend-only type. This entry keeps
`defaultInterpolation` as a plain `string` on the Go `Config`. `settings.ts` keeps the three
canonical TS types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`); `settings.go` flattens
the backend `Settings` (jsonData fields) plus the four frontend-only fields into a single `Config`
with a `DecryptedSecureJSONData` map.

## Modeling decisions

- **No root fields**: the Snowflake backend authenticates entirely from `jsonData` +
  `secureJsonData` (`settings.go:56-192`, `driver.go:96-147`) and never reads root-level datasource
  fields (`url`, `user`, `basicAuth`). `RootConfig` is therefore a blank object and `Config` carries
  no `json:"-"` root fields.
- **Auth discriminator**: `authType` (role `auth.discriminator`) has four values —
  `password` (default), `keypair`, `pat`, `oauth`. A missing/empty `authType` is treated as
  `password` by the backend (`settings.go:76`); `ApplyDefaults` encodes that, and `Validate` treats
  `""` and `password` identically.
- **`dependsOn` vs `requiredWhen`**: `dependsOn` mirrors the editor's conditional rendering
  (`ConfigEditor.tsx`: username hidden for oauth; password/privateKey/passphrase/patToken/oauth
  switch shown per auth type). `requiredWhen` encodes the backend data contract — `account`
  (needed to connect), `username` (non-oauth), and each auth method's secret / `oauthPassThru`.
- **Editor `required` markers vs the data contract**: the editor marks only the secrets
  (`password`, `privateKey`, `patToken`) with `required` (`ConfigEditor.tsx:146,176,213`); it does
  **not** mark `account`/`username`. `LoadSettings` likewise only fails on a missing secret, not on
  a missing account/username (the backend test "Default settings" loads with only a password).
  To stay faithful, Go `Validate` mirrors `LoadSettings` (auth secret / `oauthPassThru` only) and
  does **not** require `account`/`username`; the schema still records `requiredWhen` for them as the
  data contract for consumers.
- **Session settings as an array-of-object**: `jsonData.settings` is modeled as `valueType: array`
  with an object `item` of `{name, value, secure}` (each `isItemField: true`), mirroring
  `SessionSettings.tsx`. The editor's info-drawer content is attached as the field's `help`.
- **Private-key validation is envelope-only**: `Validate` uses stdlib `encoding/pem` to confirm the
  key is a PKCS#8 PEM block (`PRIVATE KEY` or `ENCRYPTED PRIVATE KEY`, mirroring `settings.go:96-97`)
  without parsing/decrypting it — that avoids pulling in the `youmark/pkcs8` dependency the plugin
  uses at runtime. Passphrase correctness is not checked here.
- **Frontend-only fields carried on `Config`**: the four frontend-only jsonData fields are included
  on `Config` (with `frontend-only` tags in the schema) because the conformance guard rail requires
  the jsonData key set and the struct json tags to match exactly in both directions.
- **`Setting.secure` json tag**: this entry tags the field `json:"secure"` to round-trip the stored
  shape; the upstream backend `Setting` leaves `Secure` untagged and relies on case-insensitive
  unmarshal of the frontend's `secure` (see [Upstream findings](#upstream-findings) #1).
- **Groups mirror the editor sections verbatim**: `Connection`, `Connection settings`,
  `Environment`, `Customization` (the four `<h3>` blocks in render order). Authentication fields live
  inside the "Connection" section in the editor, so they are grouped there rather than in a separate
  "Authentication" group. The optional/expandable-feeling sections (`Connection settings`,
  `Customization`) are marked `optional`.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle from the
embedded `dsconfig.json`: jsonData fields become the OpenAPI settings `spec`, secure fields become
`secureValues`, and array/object item fields are nested.

`SettingsExamples()` provides the default configuration plus one example per authentication type
(and a session-parameters variant). Each is a full instance-settings object with `jsonData` and the
relevant write-only `secureJsonData` placeholders (the default `""` example carries an empty
`password` to show what must be filled in; the key-pair examples embed throwaway PKCS#8 PEM keys):

| Example | Auth | `secureJsonData` |
| --- | --- | --- |
| `""` (default) | Password (schema default) | `password` (empty) |
| `passwordAuth` | Password | `password` |
| `keyPair` | Key Pair (unencrypted) | `privateKey` |
| `keyPairEncrypted` | Key Pair (encrypted) | `privateKey`, `privateKeyPassphrase` |
| `programmaticAccessToken` | PAT | `patToken` |
| `oauth` | OAuth (forward identity) | — (none; `oauthPassThru: true`) |
| `sessionParameters` | Password + `settings[]` + timeouts + `rowLimit` | `password` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase flow and returns a fully-defaulted, validated `Config`:

1. **Parse** — unmarshal `settings.JSONData` into `Config` and copy the four known decrypted secrets
   into `DecryptedSecureJSONData`. Mirrors `LoadSettings` (`pkg/settings.go:56-192`).
2. **`ApplyDefaults`** — set `AuthType` to `password` when empty (mirrors `settings.go:60-61,76` and
   `ConfigEditor.tsx:39`). Curated: only the discriminator is defaulted.
3. **`Validate`** — enforce the runtime contract per auth method: `password`/empty requires
   `password`; `keypair` requires a PKCS#8 PEM `privateKey`; `pat` requires `patToken`; `oauth`
   requires `oauthPassThru == true`. Errors mirror the plugin's own (`invalid password`,
   `invalid private key`, `invalid programmatic access token`, `you must enable Forward OAuth
   Identity`) and are joined.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`, `datasource_name`,
and `plugin` labels. `ApplyDefaults` and `Validate` are exported for callers that assemble a
`Config` directly (provisioning preview, round-trip tooling, tests distinguishing parse-level from
policy-level errors).

## Upstream findings

Potential bugs, dead code, and inconsistencies found while researching upstream. All preserved
verbatim in the schema — the schema records what the plugin **does**, not what it **should** do.

1. **`Setting.Secure` has no json tag.** `pkg/settings.go:20` declares `Secure bool` with no json
   tag; the frontend writes `secure` (lowercase, `types.ts:44`). Unmarshal only works because Go's
   `encoding/json` is case-insensitive; marshaling the backend struct would emit `"Secure"`. This
   entry uses `json:"secure"` to model the stored shape faithfully.
2. **PKCS#1 keys are silently rejected.** The private-key placeholder says
   `Begins with -----BEGIN PRIVATE KEY-----` (`selectors.ts:34`), and `settings.go:97` accepts only
   block type `PRIVATE KEY` or `ENCRYPTED PRIVATE KEY` (PKCS#8). A valid PKCS#1 key
   (`-----BEGIN RSA PRIVATE KEY-----`) fails with `invalid private key`.
3. **Dead/ambiguous key-pair fall-through.** `pkg/settings.go:126-134` has a branch that logs
   `"could not decrypt secure JSON data for private key (but it is not empty or nil)"` and a
   `// TODO why are we not returning an error here?` — the happy path already returns earlier, so
   this code is effectively unreachable for the `ok` case and does not return an error.
4. **OAuth token parsing swallows malformed headers.** `pkg/settings.go:166-170` returns
   `settings, nil` (with another `// TODO why are we not returning an error here?`) when the
   forwarded `Authorization` header has fewer than two fields, instead of erroring.
5. **`region` is deprecated but still read.** The tooltip says
   `Deprecated; prefer including the region in the 'Account' field` (`selectors.ts:53`), yet the
   backend still passes `settings.Region` to `sf.Config.Region` (`driver.go:110`).
6. **`oauthPassThru` is the only non-optional jsonData field in TS.** `types.ts:24` declares
   `oauthPassThru: boolean;` (no `?`) while every other option is optional — a minor typing
   inconsistency. It is modeled as an optional `boolean` here.
7. **GET/PUT statements are blocked.** `driver.go:37-44,197-302` rewrites any query containing a
   Snowflake `GET`/`PUT` (client-side file transfer) to a sentinel that fails, returning
   `query blocked: PUT and GET commands are not permitted for security reasons`. Documented in an
   instruction so consumers do not expect file-transfer support.
8. **Secrets persist on blur.** The editor writes secrets via `onBlur` (`ConfigEditor.tsx:143,173,
   193,210`), so a typed secret is only staged when the field loses focus — a UX subtlety, not a
   storage difference.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes
  (via the `ConfigSchemaValid` conformance subtest).
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft-07,
  `additionalProperties: false`, strict) — passes.
- `go generate ./...` then `go test ./...` on this entry — passes: `TestSchemaConformance`
  (BaseFieldsResolved, SchemaRoundTrip, SchemaArtifactInSync, SchemaSpecHasNoSecureJSON,
  ConfigSchemaValid, JSONDataMatchesStruct, JSONDataTypesMatchStruct, SecureValuesMatchLoadSettings),
  `TestLoadConfig`, `TestApplyDefaults`, `TestValidate`.
- `gofmt -l .`, `go vet ./...`, `go build ./...`, `go test ./...` across the whole `registry`
  module — clean (all 46 entries pass).
- The pre-existing `dsconfig` and `schema` workspace modules still build and test — clean.
- `tsc --noEmit --strict` on `settings.ts` — clean.
