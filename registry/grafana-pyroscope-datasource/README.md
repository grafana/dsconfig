# grafana-pyroscope-datasource

Declarative configuration schema for the [Grafana Pyroscope datasource plugin](https://github.com/grafana/grafana-pyroscope-datasource) (`grafana-pyroscope-datasource`, aliasID `phlare`).

## Upstream researched

- **Repo**: `github.com/grafana/grafana-pyroscope-datasource`
- **Ref**: `main`
- **Commit SHA**: `e5d6bfbbde415427f96f80fb8fff1f9003be4e6d`

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders,
tooltips, option labels/values, section titles, help markdown, defaults,
validations, dependency and required-when expressions, storage keys, storage
targets, value types, group titles, and instructions — is traceable to a
specific `file:line` in the upstream repo at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone --depth 20 https://github.com/grafana/grafana-pyroscope-datasource
cd grafana-pyroscope-datasource
git checkout e5d6bfbbde415427f96f80fb8fff1f9003be4e6d
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under
[Sources researched](#sources-researched) before merging any changes to this
entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, effects, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: root URL/basicAuth/basicAuthUser/withCredentials tagged `json:"-"`, jsonData fields, `DecryptedSecureJSONData`), `PluginID` / `LegacyPluginID`, `SecureJsonDataKey` typed constants, `MinStepPattern`, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth / TLS / minStep variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the
shared [`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA (`e5d6bfb`), plus
external editor components at the exact versions the plugin's `package.json`
pins.

### Plugin repo (`github.com/grafana/grafana-pyroscope-datasource@e5d6bfb`)

| File | What was read |
| --- | --- |
| `src/plugin.json:1-49` | `pluginType` (`id` = `"grafana-pyroscope-datasource"`), `aliasIDs` (`["phlare"]`), `pluginName` (`name` = `"Grafana Pyroscope"`), docs URL (`info.links[2].url` = `"https://grafana.com/docs/grafana/latest/datasources/pyroscope/"`) |
| `src/ConfigEditor.tsx:1-95` | Top-level editor — composes `DataSourceDescription` (docsLink hard-coded at `:28`), `ConnectionSettings` (with `urlPlaceholder="http://localhost:4040"` at `:34`), `Auth` (via `convertLegacyAuthProps` at `:37-42`), a collapsible `ConfigSection title="Additional settings"` (`:45-49`, description at `:47`) containing `AdvancedHttpSettings`, the conditional `SecureSocksProxySettings` (`:54-56`), and a `ConfigSubSection title="Querying"` (`:58`) with a single `Field label="Minimal step"` / `Input placeholder="15s"` bound to `jsonData.minStep` (`:59-82`). The `Field` renders description at `:63`, error text at `:64`, and invalidity is checked via `/^\d+(ms|[Mwdhmsy])$/` at `:65` |
| `src/types.ts:14-19` | Frontend `PyroscopeDataSourceOptions extends DataSourceJsonData` — a single field `minStep?: string`. No other plugin-defined jsonData fields. |
| `pkg/grafana-pyroscope-datasource/instance.go:44-69` | Backend `NewPyroscopeDatasource` — the only server-side reads are `settings.URL` (base URL for the profiling client at `:66`) and `settings.HTTPClientOptions(ctx)` (`:52`) for the HTTP client (that call is what pulls in root basicAuth / TLS fields / custom headers / cookies) |
| `pkg/grafana-pyroscope-datasource/query.go:36-38, 74-89, 173-187` | Ad-hoc `dsJsonModel struct { MinStep string \`json:"minStep"\` }` (`:36-38`); unmarshaled inline per query at `:74-75` and `:173-174`, parsed via `backend/gtime.ParseDuration` at `:84` / `:183`, empty or unparseable value silently falls back to 15s (`:82,86,182,186`), effective per-query step is `max(query.Interval, parsedInterval)` (`:96, 119, 132, 187`) |
| `pkg/grafana-pyroscope-datasource/plugin.go:14, 34-36` | `NewDatasource` entry point + `logger` (used across query paths); no additional settings reads |
| `pkg/grafana-pyroscope-datasource/pyroscopeClient.go` | `NewPyroscopeClient(httpClient, settings.URL)` — the Connect/gRPC-over-HTTP profiling client wraps the SDK HTTP client and uses `settings.URL` as the base |
| `package.json` | External component versions (see next table) |

Notably absent: no `pkg/models/settings.go`, no upstream `LoadSettings` — the
Pyroscope plugin does not own a backend jsonData settings model. All
server-side reads of settings go through the SDK and the ad-hoc `dsJsonModel`
inside query handling.

### External editor components

Read at the exact versions pinned in the plugin's `package.json`
(`@grafana/plugin-ui: 0.15.0`, `@grafana/ui/runtime/data: 13.0.2`). Sources
checked out at the corresponding upstream commits.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `ConnectionSettings` | `@grafana/plugin-ui@0.15.0` | `github.com/grafana/plugin-ui`, `src/components/ConfigEditor/Connection/ConnectionSettings.tsx:17-75` | URL label defaults to `"URL"`, placeholder passed by plugin (`ConfigEditor.tsx:34` — `urlPlaceholder="http://localhost:4040"`); required + built-in URL regex validation |
| `Auth`, `AuthMethodSettings`, `BasicAuth` | `@grafana/plugin-ui@0.15.0` | `src/components/ConfigEditor/Auth/{Auth,auth-method/AuthMethodSettings,auth-method/BasicAuth}.tsx` | Default `visibleMethods = [BasicAuth, OAuthForward, NoAuth]` at `AuthMethodSettings.tsx`; option labels/descriptions from `AuthMethodSettings.tsx`; BasicAuth `User`/`Password` labels + placeholders + tooltips from `BasicAuth.tsx` |
| `convertLegacyAuthProps` | `@grafana/plugin-ui@0.15.0` | `src/components/ConfigEditor/Auth/utils.ts:8-55` | Maps `basicAuth` / `withCredentials` / `jsonData.oauthPassThru` ↔ AuthMethod enum; `onAuthMethodSelect` writes basicAuth+withCredentials+oauthPassThru in one shot (`:44-54`) |
| `TLSSettings`, `SelfSignedCertificate`, `TLSClientAuth`, `SkipTLSVerification` | `@grafana/plugin-ui@0.15.0` | `src/components/ConfigEditor/Auth/tls/*.tsx` | All TLS labels/placeholders/rows come verbatim from these files (see [Field provenance](#field-provenance)); note the RSA private key placeholder typo `Begins with --- RSA PRIVATE KEY CERTIFICATE ---` — shared across every plugin that uses `Auth` |
| `CustomHeaders`, `CustomHeader` | `@grafana/plugin-ui@0.15.0` | `src/components/ConfigEditor/Auth/custom-headers/*.tsx` | Indexed `httpHeaderName<N>` / `httpHeaderValue<N>` storage pattern; **not modeled** in this schema (see [Modeling decisions](#modeling-decisions)) |
| `AdvancedHttpSettings` | `@grafana/plugin-ui@0.15.0` | `src/components/ConfigEditor/AdvancedSettings/AdvancedHttpSettings.tsx:44-82` | `Allowed cookies` and `Timeout` labels/tooltips/placeholders |
| `DataSourceDescription`, `ConfigSection`, `ConfigSubSection` | `@grafana/plugin-ui@0.15.0` | `src/components/ConfigEditor/DataSourceDescription.tsx`, `ConfigSection.tsx` | Intro text prop shape; section title/description props (no storage keys — layout only) |
| `SecureSocksProxySettings` (excluded) | `@grafana/ui@13.0.2` | grafana/grafana `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | Storage key it writes (`jsonData.enableSecureSocksProxy`) — confirmed and excluded |
| `Divider`, `Field`, `Input`, `Stack`, `useStyles2` | `@grafana/ui@13.0.2` | grafana/grafana `packages/grafana-ui/src/components/` | Prop names (`label`, `placeholder`, `value`, `onChange`, `description`, `error`, `invalid`, `noMargin`, `htmlFor`, `spellCheck`) — needed to know which UI attributes to record |
| `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps`, `GrafanaTheme2` | `@grafana/data@13.0.2` | grafana/grafana `packages/grafana-data/src/` | Base jsonData interface; editor prop plumbing |
| `config` | `@grafana/runtime@13.0.2` | grafana/grafana `packages/grafana-runtime/src/` | Reads `config.secureSocksDSProxyEnabled` at `ConfigEditor.tsx:54` to gate the excluded SecureSocksProxySettings widget |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line`
where each of its label, placeholder, tooltip, default, storage key, and
value type is defined. Where a field draws from multiple lines, all lines are
listed.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | `ConnectionSettings.tsx:41` (default `urlLabel = 'URL'`; Pyroscope editor does not override) | `ConfigEditor.tsx:34` (`urlPlaceholder="http://localhost:4040"`) | `settings.URL string` — SDK base | Required (`requiredWhen: "true"`) because `NewPyroscopeClient(httpClient, settings.URL)` (`instance.go:66`) fails at request time on empty URL |
| `virtual_authMethod` | — | virtual | Standard @grafana/plugin-ui Authentication method selector | Options from `AuthMethodSettings.tsx`; default `'NoAuth'` mirrors `getSelectedMethod` fallthrough (`utils.ts:37`) | Union of 3 strings | `storage.computed.read` mirrors `getSelectedMethod` (`utils.ts:27-38`) minus `CrossSiteCredentials`, which the Pyroscope editor doesn't expose; `effects` mirror `onAuthMethodSelect` (`utils.ts:44-54`) |
| `root_basicAuth` | `basicAuth` | `root` | — (no UI; managed by `virtual_authMethod`) | Written by `utils.ts:47` | Root SDK bool | Tagged `managed-by:virtual_authMethod` |
| `root_basicAuthUser` | `basicAuthUser` | `root` | `BasicAuth.tsx` (default `userLabel = 'User'`) | `BasicAuth.tsx` (default `userPlaceholder = 'User'`) | SDK `settings.BasicAuthUser string` | `dependsOn: virtual_authMethod == 'BasicAuth'`; `requiredWhen: root_basicAuth == true` |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | `BasicAuth.tsx` (default `passwordLabel = 'Password'`) | `BasicAuth.tsx` (default `passwordPlaceholder = 'Password'`) | Role `auth.basic.password` | Same conditional/required as `basicAuthUser` |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (no visible UI; controlled by `virtual_authMethod == 'OAuthForward'`) | Written by `utils.ts:51` | `bool` (@grafana/plugin-ui writes it) | Tagged `managed-by:virtual_authMethod` |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | `SelfSignedCertificate.tsx` (`label="Add self-signed certificate"`) | `tooltipText` `SelfSignedCertificate.tsx`; default `false` | `bool` (SDK TLS pack) | — |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | `SelfSignedCertificate.tsx` (`label="CA Certificate"`) | `SelfSignedCertificate.tsx` (`placeholder="Begins with --- BEGIN CERTIFICATE ---"`, `rows: 6`) | Role `tls.caCert` | `dependsOn` / `requiredWhen`: `jsonData_tlsAuthWithCACert == true` |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | `TLSClientAuth.tsx` (`label="TLS Client Authentication"`) | `tooltipText` `TLSClientAuth.tsx` | `bool` | — |
| `jsonData_serverName` | `serverName` | `jsonData` | `TLSClientAuth.tsx` (`label="ServerName"`) | `TLSClientAuth.tsx` (`placeholder="domain.example.com"`) | Role `tls.serverName` | `dependsOn: jsonData_tlsAuth == true`; required for the mTLS contract |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | `TLSClientAuth.tsx` (`label="Client Certificate"`) | `TLSClientAuth.tsx` (`placeholder="Begins with --- BEGIN CERTIFICATE ---"`, `rows: 6`) | Role `tls.clientCert` | `dependsOn`/`requiredWhen`: `jsonData_tlsAuth == true` |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | `TLSClientAuth.tsx` (`label="Client Key"`) | `TLSClientAuth.tsx` (`` placeholder=`Begins with --- RSA PRIVATE KEY CERTIFICATE ---` `` — upstream typo preserved, `rows: 6`) | Role `tls.clientKey` | Same conditional/required as `tlsClientCert`; see [Upstream findings](#upstream-findings) #1 for the placeholder typo |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | `SkipTLSVerification.tsx` (`label="Skip TLS certificate validation"`) | `tooltipText` `SkipTLSVerification.tsx` | Role `transport.tlsSkipVerify` | Default `false` |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | `AdvancedHttpSettings.tsx:48` (`label="Allowed cookies"`) | `AdvancedHttpSettings.tsx:56` (`placeholder="New cookie (hit enter to add)"`); tooltip `AdvancedHttpSettings.tsx:50` | `string[]` | — |
| `jsonData_timeout` | `timeout` | `jsonData` | `AdvancedHttpSettings.tsx:63` (`label="Timeout"`) | `AdvancedHttpSettings.tsx:74` (`placeholder="Timeout in seconds"`); tooltip `AdvancedHttpSettings.tsx:66` | `number` (int, parsed at `AdvancedHttpSettings.tsx:33`) | Role `transport.timeoutSeconds` |
| `jsonData_minStep` | `minStep` | `jsonData` | `ConfigEditor.tsx:61` (`label="Minimal step"`) | `ConfigEditor.tsx:71` (`placeholder="15s"`); description at `:63` (`"Minimal step used for metric query. Should be the same or higher as the scrape interval setting in the Pyroscope database."`); error at `:64` (`"Value is not valid, you can use number with time unit specifier: y, M, w, d, h, m, s"`) | `string`, `types.ts:18` | Validation via `pattern: ^\d+(ms|[Mwdhmsy])$` matching the frontend regex at `ConfigEditor.tsx:65`; parsed by `backend/gtime.ParseDuration` in `query.go:84`/`:183`; empty or unparseable silently falls back to 15s at query time |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | URL | Yes (direct: `pkg/grafana-pyroscope-datasource/instance.go:66`) |
| `virtual_authMethod` | — (virtual) | — | Authentication method | — (editor-local selector) |
| `root_basicAuth` | `basicAuth` | `root` | — (managed by virtual) | Yes (SDK via `HTTPClientOptions`) |
| `root_basicAuthUser` | `basicAuthUser` | `root` | User | Yes (SDK) |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | Password | Yes (SDK) |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (managed by virtual) | Yes (SDK) |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | Add self-signed certificate | Yes (SDK) |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | CA Certificate | Yes (SDK) |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | TLS Client Authentication | Yes (SDK) |
| `jsonData_serverName` | `serverName` | `jsonData` | ServerName | Yes (SDK) |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | Client Certificate | Yes (SDK) |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | Client Key | Yes (SDK) |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | Skip TLS certificate validation | Yes (SDK) |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | Allowed cookies | Yes (SDK) |
| `jsonData_timeout` | `timeout` | `jsonData` | Timeout | Yes (SDK) |
| `jsonData_minStep` | `minStep` | `jsonData` | Minimal step | Yes — but via ad-hoc unmarshal inside query handling (`query.go:74-89, 173-187`), not through a plugin-owned settings model |

### Frontend-only settings

None. Every editor-visible field is either consumed by the SDK's
`HTTPClientOptions` path (root URL/basicAuth/TLS pairs/custom headers/cookies)
or read directly by Pyroscope's query handling (`minStep`).

### Backend-only settings

None. The Pyroscope plugin's Go code reads only what the editor writes.

### Excluded settings

- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy` and associated
  socks-proxy fields) — rendered conditionally at `ConfigEditor.tsx:54-56`
  when `config.secureSocksDSProxyEnabled` is set on the Grafana instance.
  Deliberately omitted per `AGENTS.md`.
- **Custom HTTP headers** (`@grafana/plugin-ui`'s `CustomHeaders`) — the
  editor writes indexed pairs `jsonData.httpHeaderName<N>` /
  `secureJsonData.httpHeaderValue<N>` starting at index 1. Not modeled as a
  first-class field because the storage keys are dynamic. Downstream tools
  should walk `jsonData` for the `httpHeaderName` prefix and pair up matching
  `httpHeaderValue<N>` secrets; the SDK's `HTTPClientOptions` already does
  this and forwards the resulting headers to Pyroscope.

## Where the types are defined

The configuration types are spread across the plugin and its dependencies —
some fields and base types come from libraries/SDKs rather than the plugin
itself.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `PyroscopeDataSourceOptions` (with `minStep?: string`) | `src/types.ts:14-19` | plugin ([grafana/grafana-pyroscope-datasource](https://github.com/grafana/grafana-pyroscope-datasource)) |
| `DataSourceJsonData` (base interface: `authType`, `defaultRegion`, `profile`, `manageAlerts`, `alertmanagerUid`, `disableGrafanaCache`) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `13.0.2` |
| `DataSourcePluginOptionsEditorProps`, `DataSourceSettings`, `GrafanaTheme2` | `packages/grafana-data/src/` | `@grafana/data` `13.0.2` |
| `ConnectionSettings`, `Auth`, `AuthMethod`, `convertLegacyAuthProps`, `AdvancedHttpSettings`, `DataSourceDescription`, `ConfigSection`, `ConfigSubSection` | `src/components/ConfigEditor/…` | `@grafana/plugin-ui` `0.15.0` |
| `SecureSocksProxySettings` (excluded), `Divider`, `Field`, `Input`, `Stack`, `useStyles2` | `packages/grafana-ui/src/components/` | `@grafana/ui` `13.0.2` |
| `config` (reads `config.secureSocksDSProxyEnabled`) | `packages/grafana-runtime/src/` | `@grafana/runtime` `13.0.2` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `NewPyroscopeDatasource`, `PyroscopeDatasource` (reads `settings.URL`), `NewDatasource` entry, `dsJsonModel { MinStep string \`json:"minStep"\` }`, `query`/`callResource`/`CheckHealth` — none unmarshal jsonData outside `dsJsonModel` | `pkg/grafana-pyroscope-datasource/{instance.go,plugin.go,query.go,pyroscopeClient.go}` | plugin |
| `backend.DataSourceInstanceSettings` (carries `URL`, `BasicAuthEnabled`, `BasicAuthUser`, `JSONData`, `DecryptedSecureJSONData`), `HTTPClientOptions(ctx)`, `gtime.ParseDuration` | `backend/common.go`, `backend/httpclient/`, `backend/gtime/` | `github.com/grafana/grafana-plugin-sdk-go` |

The models in this entry flatten the above into a single Go `Config` type
(root `URL`/`BasicAuth`/`BasicAuthUser`/`WithCredentials` tagged `json:"-"`,
plus the jsonData fields the editor writes and the SDK reads, plus
`DecryptedSecureJSONData`) and a `SecureJsonDataKey` typed constant list.
`settings.ts` keeps the three canonical TypeScript types (`RootConfig`,
`JsonDataConfig`, `SecureJsonDataConfig`).

## Modeling decisions

- **Virtual auth method**: `convertLegacyAuthProps`'s `onAuthMethodSelect`
  (`@grafana/plugin-ui utils.ts:44-54`) writes three storage fields in one
  shot — `root.basicAuth`, `root.withCredentials`, and
  `jsonData.oauthPassThru`. That is the same virtual-selector pattern used by
  the Loki, Tempo, and Prometheus entries. `withCredentials` is not in the
  Pyroscope editor's default `visibleMethods`, so the virtual field's effects
  only write `basicAuth` and `oauthPassThru`. If a provisioning payload
  writes `withCredentials=true` directly, the SDK still honors it — the
  virtual's `storage.computed.read` doesn't preserve that state, but the
  underlying root storage does.
- **`requiredWhen` on `basicAuthUser` / `basicAuthPassword`**: keyed on the
  underlying storage field (`root_basicAuth == true`), not the virtual
  selector. The virtual is an editor-local convenience; the backend contract
  is "if basicAuth is on, we need a username and password".
- **TLS pair requirements**: `TLSClientAuth` and `SelfSignedCertificate` mark
  every field with `required` in the UI, but they only require the paired
  fields when the parent switch is on. Encoded as `dependsOn` + `requiredWhen`
  on each field.
- **`minStep` as a validated string, not a duration type**: dsconfig has no
  native duration `valueType`. The frontend stores the raw entry as a string
  (`options.jsonData.minStep`); we mirror that with a `pattern` validation
  that matches the frontend regex verbatim
  (`^\d+(ms|[Mwdhmsy])$`, `ConfigEditor.tsx:65`). The Go `Validate` runs the
  same regex plus a `gtime.ParseDuration` sanity check so provisioning
  callers get an explicit error instead of the query-path's silent 15s
  fallback (`query.go:82-89`). Unit specifiers: `ms` (milliseconds), `s`
  (seconds), `m` (minutes — lowercase), `h` (hours), `d` (days), `w` (weeks),
  `M` (months — uppercase), `y` (years).
- **No help drawer**: Pyroscope's editor has no top-level Collapse/help panel,
  so there is no schema `help` object. The detailed guidance is captured in
  `description` on individual fields and in the `instructions` block.
- **Field ID naming convention**: IDs are prefixed with their storage target
  for easy discoverability — `root_`, `jsonData_`, or `secureJsonData_` (and
  `virtual_` for virtual fields, which have no storage target) — followed by
  the camelCase storage key. The `key` property keeps the plugin's raw
  storage key.
- **Custom HTTP headers and Secure Socks Proxy excluded**: see
  [Excluded settings](#excluded-settings) above.
- **Flat `Config` in Go**: `settings.go` collapses jsonData fields and
  decrypted secrets onto a single `Config` struct. Root-level fields the
  editor and SDK both use (`URL`, `BasicAuth`, `BasicAuthUser`,
  `WithCredentials`) are carried with `json:"-"` tags so `LoadConfig` returns
  them alongside the jsonData shape. Base `DataSourceJsonData` fields
  (authType, defaultRegion, profile, alertmanagerUid, disableGrafanaCache)
  exist in Grafana core but are neither written by the Pyroscope editor nor
  read by the Pyroscope plugin, so they are omitted.
- **`PluginID` vs `LegacyPluginID`**: `src/plugin.json:6` declares
  `aliasIDs: ["phlare"]`, so provisioning payloads written against the old
  Phlare datasource still resolve to this plugin. `PluginID =
  "grafana-pyroscope-datasource"` is the authoritative id used for both the
  registry directory name and `pluginType` in `dsconfig.json`;
  `LegacyPluginID = "phlare"` is exposed as a documented constant so callers
  can resolve both names to the same schema.
- **`ApplyDefaults` is a no-op**: the Pyroscope editor writes nothing into
  jsonData on load. The `"15s"` placeholder on Minimal step and the
  `"Timeout in seconds"` placeholder on Timeout are UI hints, not persisted
  state; the 15s query-time fallback is applied per query, not baked into
  stored settings. `ApplyDefaults` intentionally does nothing so we don't
  clobber intentional zero values — the `TestApplyDefaults` test guards
  this.
- **`SecureJsonDataConfig` is a key list**: secure values are write-only, so
  the secure type is just the array of secret key names (`basicAuthPassword`,
  `tlsCACert`, `tlsClientCert`, `tlsClientKey`); consumers read
  `secureJsonFields` to see what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema`
bundle (the k8s-style schema Grafana's datasource API server serves as
`{apiVersion}.json`, `v0alpha1` today) from the embedded `dsconfig.json`:
root fields plus a nested `jsonData` object become the OpenAPI settings
`spec`, secure fields become `secureValues`, and virtual fields are skipped.

`SettingsExamples()` provides the default configuration plus one k8s-style
example per authentication method, TLS variant, and Pyroscope-specific
feature (minStep). Each example is a full instance-settings object with the
plugin configuration nested under `jsonData` and the relevant write-only
secrets under `secureJsonData` (placeholder values to be replaced with real
secrets; the default example — keyed by the empty string `""` — carries an
empty `basicAuthPassword` to show that no secret is required for the default
No-auth mode):

| Example | Auth | TLS | Extras | `secureJsonData` |
| --- | --- | --- | --- | --- |
| `""` (default) | None | — | — | `basicAuthPassword` (empty) |
| `noAuth` | None | — | `minStep=15s` | `basicAuthPassword` (empty) |
| `basicAuth` | Basic | — | — | `basicAuthPassword` |
| `oauthForward` | OAuth Identity | — | — | `basicAuthPassword` (empty) |
| `tlsMutualAuth` | None | mTLS (serverName + client cert/key) | — | `tlsClientCert`, `tlsClientKey` |
| `tlsSelfSignedCA` | None | Custom CA | — | `tlsCACert` |
| `withMinStep` | Basic | — | `minStep=1m` | `basicAuthPassword` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and
returns a fully-defaulted, validated `Config`:

1. **Parse** — copy `settings.URL`, `settings.BasicAuthEnabled`,
   `settings.BasicAuthUser` into `Config`, unmarshal `settings.JSONData` into
   the jsonData portion of the same struct, and copy the four decrypted
   secrets into `DecryptedSecureJSONData`. The Pyroscope plugin has no
   upstream `LoadSettings` to mirror — the only server-side reads of settings
   are the ad-hoc `dsJsonModel { MinStep string }` unmarshals inside query
   handling (`query.go:74-89, 173-187`) and `settings.URL`/`HTTPClientOptions`
   in `instance.go:52,66`.
2. **`ApplyDefaults`** — intentionally a no-op (see
   [Modeling decisions](#modeling-decisions)).
3. **`Validate`** — enforce the runtime contract: URL is required, Basic
   auth requires a username, mTLS requires serverName + client cert + client
   key, custom-CA requires the CA PEM, `timeout` must be non-negative, and
   `minStep` (when set) must match the same regex the editor enforces plus
   `gtime.ParseDuration` sanity. Errors are joined so every problem surfaces
   at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with
`datasource_uid`, `datasource_name`, and `plugin` labels so log lines carry
request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are still
exported for callers that want to compose them themselves (e.g. provisioning
preview, schema-example round-trip, tests that need to distinguish parse-
level from policy-level errors). Skip them by never calling `LoadConfig` in
those flows — assemble a `Config` directly.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while
researching upstream. All preserved verbatim in the schema — the schema
records what the plugin **does**, not what it **should** do; these notes
exist so reviewers can reproduce each finding and decide separately whether
to fix upstream.

1. **Upstream placeholder typo preserved**: `@grafana/plugin-ui`'s
   `TLSClientAuth.tsx` sets the client key placeholder to
   `Begins with --- RSA PRIVATE KEY CERTIFICATE ---` — an RSA private key is
   not a "certificate". Preserved verbatim in
   `secureJsonData_tlsClientKey.ui.placeholder`. This is a plugin-ui typo
   shared across all data sources that use `Auth`.
2. **Silent `minStep` fallback**: `query.go:82-89` and `:173-187` parse
   `dsJson.MinStep` via `gtime.ParseDuration`; on parse failure the code
   logs a warning and defaults to 15s per query with no user-visible error.
   The frontend regex at `ConfigEditor.tsx:65` catches malformed values in
   the editor, but a provisioning payload can bypass the editor and store
   any string. `Validate` on this entry surfaces the same regex check plus a
   `gtime.ParseDuration` sanity so callers get an explicit error rather than
   a silent runtime fallback.
3. **`minStep` regex accepts `500ms` alongside `1s`, `1m`, `1M`, `1y`**:
   the pattern `^\d+(ms|[Mwdhmsy])$` accepts millisecond precision only via
   the two-character `ms` suffix; a stray single `m` after digits means
   minutes (lowercase). This is easy to misread — e.g. `500m` is 500
   minutes, not 500 milliseconds — and there is no editor-side warning.
4. **Alias id (`phlare`) resolves to this plugin**: `src/plugin.json:6`
   declares `aliasIDs: ["phlare"]`. The alias is the pre-rename plugin id;
   both `grafana-pyroscope-datasource` and `phlare` land in the same code
   path. Provisioning payloads written against the old id still work, but
   the primary id used everywhere else in this schema is
   `grafana-pyroscope-datasource`.
5. **No `pkg/models/settings.go`**: unlike some sibling plugins the
   Pyroscope datasource does not own a typed backend settings struct. The
   only server-side read of `jsonData` is the ad-hoc `dsJsonModel` inside
   query handling. This entry's `Config` fills that gap — it is the intended
   shape a plugin-owned settings loader would produce.
6. **No connection health check for URL**: the Pyroscope backend does not
   pre-validate `settings.URL`; requests just fail when the profiling client
   issues them. We surface this as a `requiredWhen: "true"` constraint on
   `root_url` so provisioning tooling can reject an empty URL upfront.
7. **Base `DataSourceJsonData` fields are unused**: Grafana core embeds
   `DataSourceJsonData` into any jsonData shape and it carries `authType`,
   `defaultRegion`, `profile`, `manageAlerts`, `alertmanagerUid`,
   `disableGrafanaCache`. The Pyroscope editor writes none of them, and the
   Pyroscope backend reads none of them. They are omitted from the schema.
8. **`SecureSocksProxySettings` also writes to jsonData**: the excluded
   Secure Socks Proxy widget writes `jsonData.enableSecureSocksProxy` and
   related fields. Provisioning payloads that include those keys will not
   round-trip through this schema — they will be preserved in the raw
   `JSONData` but not be represented in `Config` or `SettingsExamples`.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in
  this repo) — passes (invoked by the conformance suite).
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft 2020-12, `additionalProperties: false`) — passes (invoked by the
  conformance suite).
- `go test ./...` inside `registry/` — passes on every entry, including the
  new `grafana-pyroscope-datasource` package (schema bundle shape, secure
  values, examples, `LoadConfig` incl. TLS variants and malformed input,
  `SchemaArtifactInSync` guard, `JSONDataMatchesStruct`,
  `JSONDataTypesMatchStruct`).
- `settings.go`/`schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: exports the three canonical types (`RootConfig`,
  `JsonDataConfig`, `SecureJsonDataConfig`) — reviewed by hand against the
  frontend sources; no `tsc` runner is wired into the registry module.
