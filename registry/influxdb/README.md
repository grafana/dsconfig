# influxdb

Declarative configuration schema for the [InfluxDB datasource plugin](https://github.com/grafana/grafana-influxdb-datasource) (`influxdb`).

## Upstream researched

- **Repo**: `github.com/grafana/grafana-influxdb-datasource`
- **Ref**: `main`
- **Commit SHA**: `a3e5fe3abfa3fa0b1f9e56c1d4d2e97fda6e898f` (`docs: add signed commits requirement to CONTRIBUTING.md`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders,
tooltips, option labels/values, section titles, defaults, dependency and
required-when expressions, storage keys, storage targets, value types, group
titles, and instructions — is traceable to a specific `file:line` in the
upstream repo at this SHA.

To reproduce this research:

```bash
git clone https://github.com/grafana/grafana-influxdb-datasource
cd grafana-influxdb-datasource
git checkout a3e5fe3abfa3fa0b1f9e56c1d4d2e97fda6e898f
```

If upstream `main` has advanced past this SHA, re-diff the sources listed
under [Sources researched](#sources-researched) before merging any changes
to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: root URL/BasicAuth/BasicAuthUser/User/Database/WithCredentials tagged `json:"-"`, jsonData fields, `DecryptedSecureJSONData`), `PluginID`, `InfluxVersion` / `InfluxHTTPMode` / `InfluxProduct` / `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each query language / auth / TLS variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| `schema.gen.json`, `settings.gen.json`, `settings.examples.gen.json` | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the
shared [`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Two editors, one storage shape

`src/module.ts:10-14` picks between two config editors at runtime:

- **V1** (`src/components/editor/config/ConfigEditor.tsx`) — the default; a
  Select for query language, `@grafana/ui`'s `DataSourceHttpSettings`, and a
  language-specific details tab (InfluxQL / Flux / SQL).
- **V2** (`src/components/editor/config-v2/ConfigEditor.tsx`) — behind the
  `newInfluxDSConfigPageDesign` feature toggle; a new URL + product + query
  language wizard, `@grafana/plugin-ui`'s `AuthMethod` radio, and per-language
  database connection sections.

Both editors write into the same jsonData + secureJsonData storage keys. This
schema captures the **union** of what either editor writes, with V2-only
fields (`jsonData.product`, `jsonData.pdcInjected`, and the switch to
`root.basicAuthUser` + `secureJsonData.basicAuthPassword` under BasicAuth) and
V1-only legacy fields (`root.user` + `secureJsonData.password`) both modeled
verbatim.

## Sources researched

### Plugin repo (`github.com/grafana/grafana-influxdb-datasource@a3e5fe3`)

| File | What was read |
| --- | --- |
| `src/plugin.json:1-40` | `pluginType` (`id`=`"influxdb"`), `pluginName` (`name`=`"InfluxDB"`), `docURL` (`info.links[1].url`) |
| `src/types.ts:5-9` | `InfluxVersion` enum values — `'InfluxQL' \| 'Flux' \| 'SQL'` (labels match verbatim) |
| `src/types.ts:11-31` | `InfluxOptions extends DataSourceJsonData` — jsonData shape: `version`, `timeInterval`, `httpMode`, `showTagTime`, `dbName`, `product`, `pdcInjected`, `oauthPassThru`, `organization`, `defaultBucket`, `maxSeries`, `insecureGrpc` |
| `src/types.ts:36-39` | Deprecated `InfluxOptionsV1` shape — adds `user` and `database` as jsonData fields (but the v1 editor writes them at root, not jsonData) |
| `src/types.ts:41-47` | `InfluxSecureJsonData` — `token` and `password` |
| `src/module.ts:4-14` | Config editor selection: `newInfluxDSConfigPageDesign` toggle picks V1 vs V2 |
| `src/components/editor/config/ConfigEditor.tsx:20-42` | V1 query-language Select options with the exact labels `"InfluxQL"`/`"SQL"`/`"Flux"` and their descriptions |
| `src/components/editor/config/ConfigEditor.tsx:62-73` | `onVersionChanged` — selecting Flux forces `access='proxy'`, `basicAuth=true`, `jsonData.httpMode='POST'` and deletes root `user` + `database` |
| `src/components/editor/config/ConfigEditor.tsx:107-119` | Renders inline `Alert` when `access==='direct'`; wires `DataSourceHttpSettings defaultUrl="http://localhost:8086"` |
| `src/components/editor/config/ConfigEditor.tsx:120-142` | Max series input — `placeholder="1000"`, verbatim tooltip |
| `src/components/editor/config/InfluxInfluxQLConfig.tsx:41-51` | Database Access Alert copy (informational only, not schema-modeled) |
| `src/components/editor/config/InfluxInfluxQLConfig.tsx:53-76` | Database input: label `"Database"`, writes `jsonData.dbName` and clears `root.database` on change |
| `src/components/editor/config/InfluxInfluxQLConfig.tsx:78-91` | User input: label `"User"`, writes `root.user` (not `basicAuthUser`) |
| `src/components/editor/config/InfluxInfluxQLConfig.tsx:92-108` | Password `SecretInput`: writes `secureJsonData.password` (not `basicAuthPassword`) |
| `src/components/editor/config/InfluxInfluxQLConfig.tsx:109-133` | HTTP Method Select: verbatim GET/POST tooltip; writes `jsonData.httpMode` |
| `src/components/editor/config/InfluxInfluxQLConfig.tsx:135-154` | Min time interval input: label + tooltip + placeholder `"10s"` |
| `src/components/editor/config/InfluxInfluxQLConfig.tsx:156-175` | Autocomplete range input: label + tooltip + placeholder `"12h"` |
| `src/components/editor/config/InfluxFluxConfig.tsx:22-84` | V1 Flux tab: Organization, Token (secureJsonData.token), Default Bucket, Min time interval |
| `src/components/editor/config/InfluxSQLConfig.tsx:19-90` | V1 SQL tab: Database, Token, Insecure Connection |
| `src/components/editor/config-v2/UrlAndAuthenticationSection.tsx:38-315` | V2 URL + Product + Query language section, product/version validation logic, DBRP mapping warning for OSS 1.x / 2.x / Enterprise 1.x |
| `src/components/editor/config-v2/versions.ts:24-155` | V2 product option values — the 10 `InfluxDBProduct.name` strings |
| `src/components/editor/config-v2/AuthSettings.tsx:50-304` | V2 auth radio (NoAuth / BasicAuth / OAuthForward), TLS toggles, `AuthMethod` enum from `@grafana/plugin-ui` |
| `src/components/editor/config-v2/AdvancedHttpSettings.tsx:15-108` | V2 Allowed cookies / Timeout / CustomHeaders (dynamic secrets not modeled) |
| `src/components/editor/config-v2/AdvancedDBConnectionSettings.tsx:24-147` | V2 HTTP Method / Min time interval / Autocomplete range / Max series / Insecure Connection controls (mirror the same storage keys as v1) |
| `src/components/editor/config-v2/InfluxInfluxQLDBConnection.tsx:19-88` | V2 InfluxQL DB details: writes `root.user` (matching v1) but validates against `root.basicAuth` for Basic mode |
| `src/components/editor/config-v2/InfluxFluxDBConnection.tsx:18-103` | V2 Flux DB details |
| `src/components/editor/config-v2/InfluxSQLDBConnection.tsx:17-71` | V2 SQL DB details |
| `src/components/editor/config-v2/LeftSideBar.tsx:8-12` | Reads `jsonData.pdcInjected` (backend-set) to switch section headers |
| `src/datasource.ts:48-102` | `InfluxDatasource` constructor — reads root.url, root.username (legacy `username`), root.password (legacy `password`), root.basicAuth, root.access, jsonData.dbName (fallback root.database), jsonData.timeInterval, jsonData.showTagTime, jsonData.httpMode (default `'GET'`), jsonData.version (default `'InfluxQL'`) |
| `src/datasource.ts:105-108` | Rejects `access==='direct'` at query time |
| `pkg/influxdb/settings.go:3-7` | Backend `influxVersionInfluxQL` / `influxVersionFlux` / `influxVersionSQL` constants |
| `pkg/influxdb/influxdb.go:26-87` | Backend `NewDatasource`: `settings.HTTPClientOptions(ctx)`, unmarshals `settings.JSONData` into `DatasourceInfo`, defaults httpMode `'GET'`, maxSeries `1000`, version `'InfluxQL'`, falls back to `settings.Database` when jsonData.dbName is empty, reads `settings.DecryptedSecureJSONData["token"]` |
| `pkg/influxdb/influxdb.go:97-107` | `QueryData` dispatches by `Version` to Flux / InfluxQL / FSQL |
| `pkg/influxdb/models/datasource_info.go:11-32` | `DatasourceInfo` shape: `DbName`, `Version`, `HTTPMode`, `TimeInterval`, `DefaultBucket`, `Organization`, `MaxSeries`, `InsecureGrpc` (each with json tags) |
| `pkg/influxdb/influxql/influxql.go:140-194` | `createRequest` — writes `?db={DbName}`, GET or POST, form-encoded body for POST |

### External editor components

Read at the exact versions pinned in the plugin's `package.json`.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `DataSourceHttpSettings`, `BasicAuthSettings`, `HttpProxySettings`, `TLSAuthSettings`, `CertificationKey`, `CustomHeadersSettings`, `Alert`, `Input`, `SecretInput`, `Select`, `Field`, `FieldSet`, `InlineField`, `InlineLabel`, `InlineSwitch`, `Combobox`, `Box`, `Stack`, `Text`, `TextLink`, `TagsInput`, `Checkbox`, `Button`, `Space` | `@grafana/ui@13.1.0` | grafana/grafana `packages/grafana-ui/src/components/DataSourceSettings/` and `packages/grafana-ui/src/components/…` | Label / placeholder conventions for the "URL", "Basic auth", "With Credentials", "TLS Client Auth", "With CA Cert", "Skip TLS Verify", "Forward OAuth Identity", "ServerName", "CA Cert", "Client Cert", "Client Key", "Allowed cookies", "Timeout" fields — same conventions used by opentsdb / graphite entries. `SecureSocksProxySettings` rendered conditionally; excluded per AGENTS.md |
| `AuthMethod` enum + `convertLegacyAuthProps` | `@grafana/plugin-ui@0.13.1` (resolved from package-lock, not a direct dependency) | grafana/plugin-ui `src/components/ConfigEditor/Auth/…` | AuthMethod values (`NoAuth`, `BasicAuth`, `OAuthForward`, `CrossSiteCredentials`); mapping to `basicAuth` / `withCredentials` / `oauthPassThru`. Used only by v2's `AuthSettings.tsx` |
| `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps`, `updateDatasourcePluginJsonDataOption`, `onUpdateDatasourceOption`, `onUpdateDatasourceJsonDataOption`, `onUpdateDatasourceJsonDataOptionSelect`, `onUpdateDatasourceSecureJsonDataOption`, `updateDatasourcePluginResetOption` | `@grafana/data@13.1.0` | grafana/grafana `packages/grafana-data/src/` | Base jsonData interface + editor-option helpers |
| `config.secureSocksDSProxyEnabled`, `config.featureToggles.newInfluxDSConfigPageDesign`, `config.featureToggles.influxDBConfigValidation` | `@grafana/runtime@13.1.0` | grafana/grafana `packages/grafana-runtime/src/config` | Feature toggles that gate v2 editor + Secure Socks Proxy rendering |

## Field inventory

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | URL | Yes (direct + SDK) |
| `root_basicAuth` | `basicAuth` | `root` | Basic auth | Yes (SDK) |
| `root_withCredentials` | `withCredentials` | `root` | With Credentials | Yes (SDK) |
| `root_basicAuthUser` | `basicAuthUser` | `root` | User (v2 Basic mode) | Yes (SDK) |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | Password (v2 Basic mode) | Yes (SDK) |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | TLS Client Auth | Yes (SDK) |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | With CA Cert | Yes (SDK) |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | Skip TLS Verify | Yes (SDK) |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | Forward OAuth Identity | Yes (SDK) |
| `jsonData_serverName` | `serverName` | `jsonData` | ServerName | Yes (SDK) |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | CA Cert | Yes (SDK) |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | Client Cert | Yes (SDK) |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | Client Key | Yes (SDK) |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | Allowed cookies | Yes (SDK) |
| `jsonData_timeout` | `timeout` | `jsonData` | Timeout | Yes (SDK) |
| `jsonData_version` | `version` | `jsonData` | Query language | Yes (query dispatch) |
| `jsonData_product` | `product` | `jsonData` | Product (v2 only) | No (editor-only) |
| `jsonData_pdcInjected` | `pdcInjected` | `jsonData` | — (no UI) | Yes (backend-writable indicator) |
| `jsonData_dbName` | `dbName` | `jsonData` | Database | Yes (?db= param) |
| `root_user` | `user` | `root` | User (v1 InfluxQL) | Partially (SDK legacy path only) |
| `secureJsonData_password` | `password` | `secureJsonData` | Password (v1 InfluxQL) | Partially (SDK legacy path only) |
| `jsonData_httpMode` | `httpMode` | `jsonData` | HTTP Method | Yes |
| `jsonData_timeInterval` | `timeInterval` | `jsonData` | Min time interval | Yes |
| `jsonData_showTagTime` | `showTagTime` | `jsonData` | Autocomplete range | Consumed by frontend (metadata queries) |
| `jsonData_organization` | `organization` | `jsonData` | Organization | Yes (Flux) |
| `secureJsonData_token` | `token` | `secureJsonData` | Token | Yes (Flux / SQL) |
| `jsonData_defaultBucket` | `defaultBucket` | `jsonData` | Default Bucket | Yes (Flux) |
| `jsonData_insecureGrpc` | `insecureGrpc` | `jsonData` | Insecure Connection | Yes (SQL/FlightSQL) |
| `jsonData_maxSeries` | `maxSeries` | `jsonData` | Max series | Yes |

### Frontend-only settings

- **`jsonData.product`** — written by the v2 editor to drive its query-language
  Combobox filtering and the "requires DBRP mapping" alert
  (`UrlAndAuthenticationSection.tsx:80-89`). Not read by the backend.

### Backend-only / editor-invisible

- **`jsonData.pdcInjected`** — populated by the Grafana backend when a PDC
  proxy is injected; the v2 editor reads it to render a different sidebar
  (`LeftSideBar.tsx:12`) but does not write it. Tagged `backend-only` in the
  schema.

### Partially-consumed (editor-writable, backend-inert)

- **`root.user`** + **`secureJsonData.password`** — the v1 InfluxQL editor
  writes these (`InfluxInfluxQLConfig.tsx:87-107`), but neither the backend's
  own settings unmarshal nor the SDK's HTTPClientOptions auth handler
  automatically attaches them to outgoing HTTP requests. They flow through
  as `settings.User` and `secureJsonData["password"]`. For the SDK's HTTP
  Basic transport to pick them up, `root.basicAuth` must also be true — which
  the v1 editor does NOT auto-enable when the User/Password fields are used.

### Excluded settings

- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy` and associated
  fields) — rendered conditionally when the Grafana instance has
  `config.secureSocksDSProxyEnabled`. Deliberately omitted per `AGENTS.md`.
- **Custom HTTP headers** — the v2 editor's `CustomHeadersSettings`
  (`AdvancedHttpSettings.tsx:102`) writes indexed pairs
  `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>` starting
  at index 1. Not modeled as a first-class field because the storage keys are
  dynamic. The SDK's `HTTPClientOptions` handles them transparently.
- **`jsonData.metadata`** — declared on `InfluxOptions` (`src/types.ts:29`)
  but there is no editor UI writing it and no backend code reading it
  (grep for `metadata` in `pkg/` returns zero jsonData reads). Omitted.

## Where the types are defined

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `InfluxVersion`, `InfluxOptions`, `InfluxOptionsV1`, `InfluxSecureJsonData` | `src/types.ts:5-47` | plugin ([grafana/grafana-influxdb-datasource](https://github.com/grafana/grafana-influxdb-datasource)) |
| `InfluxDBProduct`, `INFLUXDB_VERSION_MAP` | `src/components/editor/config-v2/versions.ts:17-155` | plugin |
| `DataSourceJsonData` (base interface) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data@13.1.0` |
| `DataSourceHttpSettings`, `BasicAuthSettings`, `HttpProxySettings`, `TLSAuthSettings`, `CustomHeadersSettings` | `packages/grafana-ui/src/components/DataSourceSettings/` | `@grafana/ui@13.1.0` |
| `AuthMethod` enum, `convertLegacyAuthProps` | `src/components/ConfigEditor/Auth/…` | `@grafana/plugin-ui@0.13.1` (transitively resolved) |
| `SecureSocksProxySettings` (excluded) | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | `@grafana/ui@13.1.0` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `DatasourceInfo` (jsonData shape) | `pkg/influxdb/models/datasource_info.go:11-32` | plugin |
| `influxVersion*` constants | `pkg/influxdb/settings.go:3-7` | plugin |
| `NewDatasource` (settings unmarshal + defaults) | `pkg/influxdb/influxdb.go:26-87` | plugin |
| `Query` dispatch by version | `pkg/influxdb/influxdb.go:97-107` | plugin |
| InfluxQL `createRequest` | `pkg/influxdb/influxql/influxql.go:140-194` | plugin |
| Backend `settings.HTTPClientOptions(ctx)` | `backend/httpclient/` | `github.com/grafana/grafana-plugin-sdk-go` |
| `backend.DataSourceInstanceSettings` (carries `URL`, `BasicAuthEnabled`, `BasicAuthUser`, `User`, `Database`, `JSONData`, `DecryptedSecureJSONData`) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` |

The models in this entry flatten that spread into a single Go `Config` type
(root `URL` / `BasicAuth` / `BasicAuthUser` / `User` / `Database` /
`WithCredentials` tagged `json:"-"`, plus the jsonData fields, plus
`DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant list.

## Modeling decisions

- **No virtual auth discriminator**: InfluxDB doesn't have a single-field auth
  model. The v1 editor uses `DataSourceHttpSettings` (independent
  `basicAuth`, `withCredentials`, `oauthPassThru` switches) plus per-language
  auth (`user`/`password` for InfluxQL, `token` for Flux/SQL). The v2 editor
  adds a "NoAuth / BasicAuth / OAuthForward" radio, but token is still stored
  separately from any radio choice. Following the OpenTSDB entry's pattern,
  we model each auth toggle as its own field and describe the ensemble in
  `instructions`.
- **`jsonData.version` as the primary discriminator**: `dependsOn` and
  `requiredWhen` expressions on the InfluxQL / Flux / SQL fields all key off
  `jsonData_version`. `role: "auth.discriminator"` marks it in the schema
  vocabulary even though it discriminates the query language (which
  transitively determines the auth path).
- **`requiredWhen` on `dbName`, `organization`, `defaultBucket`, `token`**:
  keyed on the underlying storage field (`jsonData_version ==
  'InfluxQL'|'Flux'|'SQL'`). The V2 editor's own runtime validators enforce
  these when `influxDBConfigValidation` is on; our `Config.Validate` enforces
  them unconditionally at load time so provisioning payloads surface missing
  fields immediately.
- **Both root.user and root.basicAuthUser modeled**: the two are written by
  different editor paths (v1 InfluxQL vs v2 Basic mode) but land in different
  root fields on the datasource. Both flow through `LoadConfig` so callers
  see whichever the operator provided.
- **Both secureJsonData.password and secureJsonData.basicAuthPassword
  modeled**: same reasoning as above. `Validate()` enforces the pair with
  `basicAuthPassword` when `basicAuth == true`, but does not require
  `password` (since the current backend never uses it — see
  [Upstream findings](#upstream-findings) #1).
- **`root.database` carried on `Config` but tagged `json:"-"`**: the backend
  reads `settings.Database` directly (`pkg/influxdb/influxdb.go:60`) as a
  fallback for `jsonData.dbName`. `ApplyDefaults` copies it into `DbName`
  when the latter is empty, mirroring the backend's behavior — this way
  `Config.DbName` is always the effective value.
- **TLS pair requirements**: `TLSAuthSettings` only requires the paired
  fields when the parent switch is on. Encoded as `dependsOn` +
  `requiredWhen` on each field.
- **Field ID naming convention**: `<target>_<camelCaseKey>` — `root_`,
  `jsonData_`, or `secureJsonData_` prefix followed by the raw storage key.
- **Enum defaults on discriminator fields**: `jsonData_version.defaultValue
  = "InfluxQL"` and `jsonData_httpMode.defaultValue = "GET"` mirror the
  backend fallbacks (`pkg/influxdb/influxdb.go:43-56`). `jsonData_maxSeries.
  defaultValue = 1000` mirrors the same file's line 48-51 fallback.

## Settings examples

| Example key | Query language | Auth | TLS | secureJsonData |
| --- | --- | --- | --- | --- |
| `""` (default) | InfluxQL | None | — | `basicAuthPassword` (empty) |
| `influxqlBasicAuth` | InfluxQL | HTTP Basic | — | `basicAuthPassword` |
| `influxqlLegacyUserPassword` | InfluxQL | Legacy user/password | — | `password` |
| `fluxToken` | Flux | HTTP Basic + Token | — | `token` |
| `sqlFlightSQL` | SQL | Token | — | `token` |
| `tlsMutualAuth` | InfluxQL | None | mTLS | `tlsClientCert`, `tlsClientKey` |
| `tlsSelfSignedCA` | InfluxQL | None | Custom CA | `tlsCACert` |
| `oauthForward` | InfluxQL | Forward OAuth | — | `basicAuthPassword` (empty) |
| `legacyRootDatabase` | InfluxQL | None | — | `basicAuthPassword` (empty) |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings)
(Config, error)` runs the full three-phase load flow on a datasource instance's
settings and returns a fully-defaulted, validated `Config`:

1. **Parse** — copy `settings.URL`, `settings.BasicAuthEnabled`,
   `settings.BasicAuthUser`, `settings.User`, `settings.Database` into
   `Config`, unmarshal `settings.JSONData` into the jsonData portion of the
   same struct, and copy the six decrypted secrets into
   `DecryptedSecureJSONData`. Mirrors the split reads that the plugin's
   `NewDatasource` performs (`pkg/influxdb/influxdb.go:37-51`).
2. **`ApplyDefaults`** — fill `Version='InfluxQL'`, `HTTPMode='GET'`,
   `MaxSeries=1000` when zero, and copy `Database` → `DbName` when the latter
   is empty. All four mirror the backend's fallbacks
   (`pkg/influxdb/influxdb.go:43-61`).
3. **`Validate`** — enforce the runtime contract: URL is required, `Version`
   must be one of the three known languages, `HTTPMode` must be empty/GET/POST,
   per-language required fields (InfluxQL: `dbName`; Flux: `organization` +
   `defaultBucket` + `token`; SQL: `dbName` + `token`), Basic auth requires a
   username, mTLS requires serverName + client cert + client key, custom-CA
   requires the CA PEM, and numeric fields (timeout, maxSeries) must be
   non-negative. Errors are joined so every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with
`datasource_uid`, `datasource_name`, and `plugin` labels so log lines carry
request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are still
exported for callers that want to compose them themselves (provisioning
preview, schema-example round-trip, tests that need to distinguish
parse-level from policy-level errors).

## Upstream findings

All preserved verbatim in the schema — the schema records what the plugin
**does**, not what it **should** do; these notes exist so reviewers can
reproduce each finding and decide separately whether to fix upstream.

1. **`root.user` + `secureJsonData.password` are inert in the backend**:
   the v1 InfluxQL editor writes them (`InfluxInfluxQLConfig.tsx:87-107`)
   and the v2 editor's InfluxQL DB connection editor also writes `user`
   (`InfluxInfluxQLDBConnection.tsx:66-74`) + `password`
   (`InfluxInfluxQLDBConnection.tsx:76-85`). But `pkg/influxdb/influxdb.go`
   does not read these — it only reads `settings.DecryptedSecureJSONData["token"]`
   — and `settings.HTTPClientOptions` only applies Basic auth when
   `settings.BasicAuthEnabled == true` (using `BasicAuthUser` +
   `secureJsonData["basicAuthPassword"]`, not `User` + `password`). Result:
   filling in only User + Password in the v1 InfluxQL tab yields
   unauthenticated requests. Operators must additionally enable
   `basicAuth` at the top of the config editor for the credentials to be
   attached — but if they do, the SDK uses `basicAuthUser` +
   `basicAuthPassword` (which the v1 editor does NOT populate from the
   User/Password inputs). Preserved as first-class fields with a
   description that flags the discrepancy.
2. **`onVersionChanged` for Flux force-sets multiple fields
   simultaneously**: selecting Flux writes `access='proxy'`,
   `basicAuth=true`, `jsonData.httpMode='POST'` and deletes `root.user` +
   `root.database` in a single onOptionsChange call
   (`ConfigEditor.tsx:62-73`). A provisioning payload can skip `basicAuth`
   entirely and Flux queries will still work if a valid `token` is
   provided — but the editor UI will show a "basicAuth was force-enabled"
   inconsistency the next time an operator edits the datasource.
3. **`access='direct'` (Browser mode) is rejected at query time, not at
   admission**: `ConfigEditor.tsx:107-111` shows an inline error banner,
   and `datasource.ts:105-108` throws `BROWSER_MODE_DISABLED_MESSAGE` from
   `query()`. But nothing prevents the datasource from being provisioned
   with `access: "direct"` — the datasource just returns errors for every
   query. Preserved: `root_access` is not modeled as a schema field (it
   defaults to `'proxy'` via SDK); operators shouldn't set it to `direct`.
4. **`jsonData.pdcInjected` is a backend-controlled indicator**: the v2
   editor reads it (`LeftSideBar.tsx:12`) but the codebase does not include
   the writer — it appears to be set by Grafana's PDC injection
   middleware, not the plugin itself. Documented as `backend-only` +
   `defaultValue: false`.
5. **`jsonData.metadata` in `InfluxOptions` is unused**: declared
   (`src/types.ts:29`) but no writer in the editors and no reader in the
   backend Go code. Omitted from the schema; adding it would fail the
   `JSONDataMatchesStruct` conformance test unless we also added it to
   `Config`.
6. **`InfluxOptionsV1` declares `user` and `database` as jsonData fields**
   (`src/types.ts:36-39`), but the actual editors write them to the
   **root** of the datasource settings, not jsonData. This is a stale
   frontend type — the deprecation comment is accurate. Modeled as
   `root_user` and `root_database` (the latter is not first-class in the
   schema because there is no editor UI; it lives on `Config` for the
   fallback logic).
7. **The v1 InfluxQL User input doesn't have a placeholder or tooltip**
   (`InfluxInfluxQLConfig.tsx:78-91`); the v2 editor's uses `placeholder="myuser"`
   (`InfluxInfluxQLDBConnection.tsx:69`). The schema uses `"myuser"` to
   provide operator guidance without violating fidelity (the v1 editor
   simply has no placeholder to preserve).
8. **`onVersionChanged` for Flux blanks `root.user`+`root.database` but
   NOT `secureJsonData.password`**: the legacy password secret can leak
   into a Flux configuration. The current backend ignores it (see #1) but
   downstream tooling that echoes `secureJsonFields` may still surface it.
9. **Frontend and backend disagree on legacy `dbName` fallback semantics**:
   the backend prefers `jsonData.dbName` and falls back to
   `settings.Database` (`pkg/influxdb/influxdb.go:58-61`). The frontend
   InfluxQL editor prefers `jsonData.dbName ?? database` for display
   (`InfluxInfluxQLConfig.tsx:63`) but writes only to `jsonData.dbName`
   and blanks `root.database`. So a legacy provisioned datasource with
   only `root.database` set will display and query correctly, but as soon
   as an operator opens the config editor and types a value the frontend
   will silently blank `root.database`. Preserved: `ApplyDefaults` mirrors
   the backend, populating `Config.DbName` from `Database` when needed.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator
  in this repo) — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft-07, `additionalProperties: false`) — passes.
- `go test ./...` on the `registry` module — passes (schema bundle shape,
  secure values, examples, `LoadConfig` incl. per-language contracts + TLS
  variants + malformed input, `SchemaArtifactInSync` guard,
  `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
  `SecureValuesMatchLoadSettings`).
- `settings.go`/`schema.go`: `go build`, `go vet`, `gofmt` — clean.
