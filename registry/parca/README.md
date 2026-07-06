# parca

Declarative configuration schema for the [Grafana Parca datasource plugin](https://github.com/grafana/grafana-parca-datasource) (plugin id `parca`).

> **Deprecation notice.** The Parca datasource plugin is scheduled for deprecation on **2nd of January 2027** (`src/ConfigEditor.tsx:17,27-30`). No updates will ship after that date. This registry entry captures the plugin as-is on `main`; it is not being extended.

## Upstream researched

- **Repo**: `github.com/grafana/grafana-parca-datasource`
- **Ref**: `main`
- **Commit SHA**: `7d9b48a70e447ff37edc77c596547a5f5826032c`

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders,
tooltips, option labels/values, section titles, help markdown, defaults,
validations, dependency and required-when expressions, storage keys,
storage targets, value types, group titles, and instructions — is traceable
to a specific `file:line` in the upstream repo at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone --depth 20 https://github.com/grafana/grafana-parca-datasource
cd grafana-parca-datasource
git checkout 7d9b48a70e447ff37edc77c596547a5f5826032c
```

If upstream `main` has advanced past this SHA, re-diff the sources listed
under [Sources researched](#sources-researched) before merging any changes
to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, effects, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: root URL/basicAuth/basicAuthUser/withCredentials tagged `json:"-"`, jsonData fields, `DecryptedSecureJSONData`), `PluginID`, `DeprecationDate`, `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth / TLS variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of
the shared [`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA (`7d9b48a`), plus
external editor components at the exact versions the plugin's
`package.json` pins.

### Plugin repo (`github.com/grafana/grafana-parca-datasource@7d9b48a`)

| File | What was read |
| --- | --- |
| `src/plugin.json:1-41` | `pluginType` (`id` = `"parca"`), `pluginName` (`name` = `"Parca"`), docs URL (`info.links[2].url` = `"https://grafana.com/docs/grafana/latest/datasources/parca/"`). No `aliasIDs`. |
| `src/ConfigEditor.tsx:1-74` | Top-level editor. Composes an `<Alert severity="warning" title="Parca data source is deprecated">` deprecation banner (`:27-30`) using a hard-coded `DEPRECATION_DATE = '2nd of January 2027'` (`:17`), `DataSourceDescription` (docsLink hard-coded at `:34`), `ConnectionSettings` with `urlPlaceholder="http://localhost:7070"` (`:40`), `Auth` via `convertLegacyAuthProps` (`:43-48`), and a collapsible `ConfigSection title="Additional settings"` (`:51-64`, description at `:53`, `isCollapsible={true}`, `isInitiallyOpen={false}`) containing `AdvancedHttpSettings` and the conditional `SecureSocksProxySettings` (`:60-62`, gated on `config.secureSocksDSProxyEnabled`) — the socks-proxy widget is excluded per `AGENTS.md`. |
| `src/types.ts:17-21` | Frontend `ParcaDataSourceOptions extends DataSourceJsonData {}` — **blank interface**. Parca defines no plugin-specific jsonData fields. |
| `pkg/parca/plugin.go:58-79` | Backend `NewParcaDatasource`. The only server-side reads are `settings.HTTPClientOptions(ctx)` at `:65` for the HTTP client and `settings.URL` at `:77` as the base URL of the Connect/gRPC-web profiling client (`queryv1alpha1connect.NewQueryServiceClient(httpClient, settings.URL, connect.WithGRPCWeb())`). |
| `pkg/parca/query.go`, `pkg/parca/resources.go` | Additional server-side code paths. **Neither unmarshals `settings.JSONData`.** Every read of configured state goes through the profiling client built from `settings.URL` + `HTTPClientOptions`. |
| `package.json` | External component versions (see next table) |

Notably absent: no `pkg/models/settings.go`, no upstream `LoadSettings` —
the Parca plugin does not own a backend jsonData settings model at all. All
server-side reads of settings go through the SDK.

### External editor components

Read against the plugin-ui sources at HEAD in
`github.com/grafana/plugin-ui` (the exact versions pinned in
`package.json` — `@grafana/plugin-ui@0.13.1`,
`@grafana/ui@13.1.0-25893932881`,
`@grafana/runtime@13.1.0-25893932881`,
`@grafana/data@13.1.0-25893932881` — are npm-only releases; the plugin-ui
git repo tags stop at `v0.12.0` on `main`. Component contracts on the read
paths below have been stable across the 0.12→0.13→0.15 line: labels,
placeholders, tooltips, and the storage keys they write are unchanged, so
these sources are the authoritative reference for what the pinned version
renders and persists.)

| Component | Package @ pin | Source consulted | What was read |
| --- | --- | --- | --- |
| `ConnectionSettings` | `@grafana/plugin-ui@0.13.1` | `grafana/plugin-ui` `src/components/ConfigEditor/Connection/ConnectionSettings.tsx:17-75` | URL label defaults to `"URL"` (`:41`), placeholder passed by plugin (`ConfigEditor.tsx:40` — `urlPlaceholder="http://localhost:7070"`); required + built-in URL regex validation |
| `Auth`, `AuthMethodSettings`, `BasicAuth` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/Auth/{Auth,auth-method/AuthMethodSettings,auth-method/BasicAuth}.tsx` | Default `visibleMethods = [BasicAuth, OAuthForward, NoAuth]` at `AuthMethodSettings.tsx:59-64`; option labels/descriptions from `AuthMethodSettings.tsx:9-32`; BasicAuth `User`/`Password` labels + placeholders + tooltips from `BasicAuth.tsx:24-29` (defaults) |
| `convertLegacyAuthProps` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/Auth/utils.ts:8-55` | Maps `basicAuth` / `withCredentials` / `jsonData.oauthPassThru` ↔ AuthMethod enum; `onAuthMethodSelect` writes basicAuth+withCredentials+oauthPassThru in one shot (`:44-54`) |
| `TLSSettings`, `SelfSignedCertificate`, `TLSClientAuth`, `SkipTLSVerification` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/Auth/tls/*.tsx` | All TLS labels/placeholders/rows come verbatim from these files (see [Field provenance](#field-provenance)); note the RSA private key placeholder typo `Begins with --- RSA PRIVATE KEY CERTIFICATE ---` at `TLSClientAuth.tsx:109` — shared across every plugin that uses `Auth` |
| `CustomHeaders`, `CustomHeader` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/Auth/custom-headers/*.tsx` (via `utils.ts:188-233`) | Indexed `httpHeaderName<N>` / `httpHeaderValue<N>` storage pattern; **not modeled** in this schema (see [Excluded settings](#excluded-settings)) |
| `AdvancedHttpSettings` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/AdvancedSettings/AdvancedHttpSettings.tsx:44-81` | `Allowed cookies` label at `:48`, tooltip at `:50`, placeholder `New cookie (hit enter to add)` at `:56`; `Timeout` label at `:64`, tooltip `HTTP request timeout in seconds` at `:66`, placeholder `Timeout in seconds` at `:74`; timeout value parsed at `:33` |
| `DataSourceDescription`, `ConfigSection` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/DataSourceDescription.tsx`, `ConfigSection.tsx` | Intro text prop shape; section title/description/collapsible props (no storage keys — layout only) |
| `SecureSocksProxySettings` (excluded) | `@grafana/ui@13.1.0-…` | grafana/grafana `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | Storage key it writes (`jsonData.enableSecureSocksProxy`) — confirmed and excluded |
| `Alert`, `Divider`, `Stack`, `useStyles2` | `@grafana/ui@13.1.0-…` | grafana/grafana `packages/grafana-ui/src/components/` | Prop shape for the deprecation banner and layout primitives (no storage keys) |
| `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps`, `GrafanaTheme2` | `@grafana/data@13.1.0-…` | grafana/grafana `packages/grafana-data/src/` | Base jsonData interface; editor prop plumbing |
| `config` | `@grafana/runtime@13.1.0-…` | grafana/grafana `packages/grafana-runtime/src/` | Reads `config.secureSocksDSProxyEnabled` at `ConfigEditor.tsx:60` to gate the excluded SecureSocksProxySettings widget |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream
`file:line` where each of its label, placeholder, tooltip, default,
storage key, and value type is defined. Where a field draws from multiple
lines, all lines are listed.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | `ConnectionSettings.tsx:41` (default `urlLabel = 'URL'`; Parca editor does not override) | `ConfigEditor.tsx:40` (`urlPlaceholder="http://localhost:7070"`) | `settings.URL string` — SDK base | Required (`requiredWhen: "true"`) because `NewQueryServiceClient(httpClient, settings.URL, ...)` (`plugin.go:77`) fails at request time on empty URL |
| `virtual_authMethod` | — | virtual | Standard @grafana/plugin-ui Authentication method selector | Options from `AuthMethodSettings.tsx:9-32`; default `'NoAuth'` mirrors `getSelectedMethod` fallthrough (`utils.ts:37`) | Union of 3 strings | `storage.computed.read` mirrors `getSelectedMethod` (`utils.ts:27-38`) minus `CrossSiteCredentials`, which the Parca editor doesn't expose; `effects` mirror `onAuthMethodSelect` (`utils.ts:44-54`) |
| `root_basicAuth` | `basicAuth` | `root` | — (no UI; managed by `virtual_authMethod`) | Written by `utils.ts:47` | Root SDK bool | Tagged `managed-by:virtual_authMethod` |
| `root_basicAuthUser` | `basicAuthUser` | `root` | `BasicAuth.tsx:24` (default `userLabel = 'User'`) | `BasicAuth.tsx:26` (default `userPlaceholder = 'User'`) | SDK `settings.BasicAuthUser string` | `dependsOn: virtual_authMethod == 'BasicAuth'`; `requiredWhen: root_basicAuth == true` |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | `BasicAuth.tsx:27` (default `passwordLabel = 'Password'`) | `BasicAuth.tsx:29` (default `passwordPlaceholder = 'Password'`) | Role `auth.basic.password` | Same conditional/required as `basicAuthUser` |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (no visible UI; controlled by `virtual_authMethod == 'OAuthForward'`) | Written by `utils.ts:51` | `bool` (@grafana/plugin-ui writes it) | Tagged `managed-by:virtual_authMethod` |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | `SelfSignedCertificate.tsx:33` (`label="Add self-signed certificate"`) | `tooltipText` `SelfSignedCertificate.tsx:34`; default `false` | `bool` (SDK TLS pack) | — |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | `SelfSignedCertificate.tsx:39` (`label="CA Certificate"`) | `SelfSignedCertificate.tsx:54` (`placeholder="Begins with --- BEGIN CERTIFICATE ---"`, `rows: 6`); default tooltip `"Your self-signed certificate"` at `:41` | Role `tls.caCert` | `dependsOn` / `requiredWhen`: `jsonData_tlsAuthWithCACert == true` |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | `TLSClientAuth.tsx:45` (`label="TLS Client Authentication"`) | `tooltipText` `TLSClientAuth.tsx:46` | `bool` | — |
| `jsonData_serverName` | `serverName` | `jsonData` | `TLSClientAuth.tsx:51` (`label="ServerName"`) | `TLSClientAuth.tsx:63` (`placeholder="domain.example.com"`); default tooltip at `:53` | Role `tls.serverName` | `dependsOn: jsonData_tlsAuth == true`; required for the mTLS contract |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | `TLSClientAuth.tsx:70` (`label="Client Certificate"`) | `TLSClientAuth.tsx:88` (`placeholder="Begins with --- BEGIN CERTIFICATE ---"`, `rows: 6`); default tooltip at `:74` | Role `tls.clientCert` | `dependsOn`/`requiredWhen`: `jsonData_tlsAuth == true` |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | `TLSClientAuth.tsx:94` (`label="Client Key"`) | `TLSClientAuth.tsx:109` (`` placeholder=`Begins with --- RSA PRIVATE KEY CERTIFICATE ---` `` — upstream typo preserved, `rows: 6`); default tooltip at `:96` | Role `tls.clientKey` | Same conditional/required as `tlsClientCert`; see [Upstream findings](#upstream-findings) #1 for the placeholder typo |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | `SkipTLSVerification.tsx:14` (`label="Skip TLS certificate validation"`) | `tooltipText` `SkipTLSVerification.tsx:15` | Role `transport.tlsSkipVerify` | Default `false` |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | `AdvancedHttpSettings.tsx:48` (`label="Allowed cookies"`) | `AdvancedHttpSettings.tsx:56` (`placeholder="New cookie (hit enter to add)"`); tooltip `AdvancedHttpSettings.tsx:50` | `string[]` | — |
| `jsonData_timeout` | `timeout` | `jsonData` | `AdvancedHttpSettings.tsx:64` (`label="Timeout"`) | `AdvancedHttpSettings.tsx:74` (`placeholder="Timeout in seconds"`); tooltip `AdvancedHttpSettings.tsx:66` | `number` (int, parsed at `AdvancedHttpSettings.tsx:33`) | Role `transport.timeoutSeconds` |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | URL | Yes (direct: `pkg/parca/plugin.go:77`) |
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

### Frontend-only settings

None. Every editor-visible field is consumed by the SDK's
`HTTPClientOptions` path (root URL/basicAuth/TLS pairs/custom
headers/cookies). Parca has no plugin-owned jsonData fields to read at all.

### Backend-only settings

None. The Parca plugin's Go code reads only what the editor writes (and
only through the SDK's HTTP client + `settings.URL`).

### Excluded settings

- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy` and associated
  socks-proxy fields) — rendered conditionally at `ConfigEditor.tsx:60-62`
  when `config.secureSocksDSProxyEnabled` is set on the Grafana instance.
  Deliberately omitted per `AGENTS.md`.
- **Custom HTTP headers** (`@grafana/plugin-ui`'s `CustomHeaders`, exposed
  via `convertLegacyAuthProps` — see `utils.ts:188-233`) — the editor
  writes indexed pairs `jsonData.httpHeaderName<N>` /
  `secureJsonData.httpHeaderValue<N>` starting at index 1. Not modeled as a
  first-class field because the storage keys are dynamic. Downstream tools
  should walk `jsonData` for the `httpHeaderName` prefix and pair up
  matching `httpHeaderValue<N>` secrets; the SDK's `HTTPClientOptions`
  already does this and forwards the resulting headers to Parca.

## Where the types are defined

The configuration types are spread across the plugin and its dependencies
— nearly every field comes from libraries/SDKs rather than the plugin
itself, because Parca declares no plugin-specific config.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `ParcaDataSourceOptions` (blank `extends DataSourceJsonData`) | `src/types.ts:17-21` | plugin ([grafana/grafana-parca-datasource](https://github.com/grafana/grafana-parca-datasource)) |
| `DataSourceJsonData` (base interface: `authType`, `defaultRegion`, `profile`, `manageAlerts`, `alertmanagerUid`, `disableGrafanaCache`) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `13.1.0-25893932881` |
| `DataSourcePluginOptionsEditorProps`, `DataSourceSettings`, `GrafanaTheme2` | `packages/grafana-data/src/` | `@grafana/data` `13.1.0-25893932881` |
| `ConnectionSettings`, `Auth`, `AuthMethod`, `convertLegacyAuthProps`, `AdvancedHttpSettings`, `DataSourceDescription`, `ConfigSection` | `src/components/ConfigEditor/…` | `@grafana/plugin-ui` `0.13.1` |
| `Alert`, `SecureSocksProxySettings` (excluded), `Divider`, `Stack`, `useStyles2` | `packages/grafana-ui/src/components/` | `@grafana/ui` `13.1.0-25893932881` |
| `config` (reads `config.secureSocksDSProxyEnabled`) | `packages/grafana-runtime/src/` | `@grafana/runtime` `13.1.0-25893932881` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `NewParcaDatasource`, `ParcaDatasource` (reads `settings.URL`), `NewDatasource` entry, `query`/`callResource`/`CheckHealth` — **none unmarshal `settings.JSONData`** | `pkg/parca/{plugin.go,query.go,resources.go}` | plugin |
| `backend.DataSourceInstanceSettings` (carries `URL`, `BasicAuthEnabled`, `BasicAuthUser`, `JSONData`, `DecryptedSecureJSONData`), `HTTPClientOptions(ctx)` | `backend/common.go`, `backend/httpclient/` | `github.com/grafana/grafana-plugin-sdk-go` |
| `queryv1alpha1connect.NewQueryServiceClient(httpClient, settings.URL, connect.WithGRPCWeb())` — the Connect/gRPC-web profiling client that consumes the SDK HTTP client and `settings.URL` | `pkg/parca/plugin.go:77` | plugin |

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
  `jsonData.oauthPassThru`. That is the same virtual-selector pattern used
  by the Pyroscope, Loki, Tempo, and Prometheus entries. `withCredentials`
  is not in the Parca editor's default `visibleMethods`, so the virtual
  field's effects only write `basicAuth` and `oauthPassThru`. If a
  provisioning payload writes `withCredentials=true` directly, the SDK
  still honors it — the virtual's `storage.computed.read` doesn't preserve
  that state, but the underlying root storage does.
- **`requiredWhen` on `basicAuthUser` / `basicAuthPassword`**: keyed on the
  underlying storage field (`root_basicAuth == true`), not the virtual
  selector. The virtual is an editor-local convenience; the backend
  contract is "if basicAuth is on, we need a username and password".
- **TLS pair requirements**: `TLSClientAuth` and `SelfSignedCertificate`
  mark every field with `required` in the UI, but they only require the
  paired fields when the parent switch is on. Encoded as `dependsOn` +
  `requiredWhen` on each field.
- **No help drawer**: Parca's editor has no top-level Collapse/help panel,
  so there is no schema `help` object. The detailed guidance is captured
  in `description` on individual fields and in the `instructions` block —
  including a deprecation-focused instruction that mirrors the editor's
  warning banner.
- **Deprecation banner is not a config field**: the deprecation alert at
  `ConfigEditor.tsx:27-30` renders warning text ("The Parca plugin is
  scheduled for deprecation on {DEPRECATION_DATE} and will no longer
  receive updates after that time.") but writes nothing to storage.
  Captured in `instructions` and mirrored as the exported Go constant
  `DeprecationDate` so downstream tools can surface the same notice.
- **Field ID naming convention**: IDs are prefixed with their storage
  target for easy discoverability — `root_`, `jsonData_`, or
  `secureJsonData_` (and `virtual_` for virtual fields, which have no
  storage target) — followed by the camelCase storage key. The `key`
  property keeps the plugin's raw storage key.
- **Custom HTTP headers and Secure Socks Proxy excluded**: see
  [Excluded settings](#excluded-settings) above.
- **Flat `Config` in Go**: `settings.go` collapses jsonData fields and
  decrypted secrets onto a single `Config` struct. Root-level fields the
  editor and SDK both use (`URL`, `BasicAuth`, `BasicAuthUser`,
  `WithCredentials`) are carried with `json:"-"` tags so `LoadConfig`
  returns them alongside the jsonData shape. Base `DataSourceJsonData`
  fields (authType, defaultRegion, profile, alertmanagerUid,
  disableGrafanaCache) exist in Grafana core but are neither written by
  the Parca editor nor read by the Parca plugin, so they are omitted.
- **No `LegacyPluginID`**: `src/plugin.json` declares no `aliasIDs`. The
  plugin id is `parca` verbatim — that is the registry directory name,
  the `pluginType` in `dsconfig.json`, and the `PluginID` Go constant.
- **`ApplyDefaults` is a no-op**: the Parca editor writes nothing into
  jsonData on load. The `"Timeout in seconds"` placeholder on Timeout is a
  UI hint, not persisted state. `ApplyDefaults` intentionally does nothing
  so we don't clobber intentional zero values — the `TestApplyDefaults`
  test guards this.
- **`SecureJsonDataConfig` is a key list**: secure values are write-only,
  so the secure type is just the array of secret key names
  (`basicAuthPassword`, `tlsCACert`, `tlsClientCert`, `tlsClientKey`);
  consumers read `secureJsonFields` to see what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go`
`pluginschema.PluginSchema` bundle (the k8s-style schema Grafana's
datasource API server serves as `{apiVersion}.json`, `v0alpha1` today)
from the embedded `dsconfig.json`: root fields plus a nested `jsonData`
object become the OpenAPI settings `spec`, secure fields become
`secureValues`, and virtual fields are skipped.

`SettingsExamples()` provides the default configuration plus one k8s-style
example per authentication method and TLS variant. Each example is a full
instance-settings object with the plugin configuration nested under
`jsonData` and the relevant write-only secrets under `secureJsonData`
(placeholder values to be replaced with real secrets; the default example
— keyed by the empty string `""` — carries an empty `basicAuthPassword` to
show that no secret is required for the default No-auth mode):

| Example | Auth | TLS | Extras | `secureJsonData` |
| --- | --- | --- | --- | --- |
| `""` (default) | None | — | — | `basicAuthPassword` (empty) |
| `noAuth` | None | — | — | `basicAuthPassword` (empty) |
| `basicAuth` | Basic | — | — | `basicAuthPassword` |
| `oauthForward` | OAuth Identity | — | — | `basicAuthPassword` (empty) |
| `tlsMutualAuth` | None | mTLS (serverName + client cert/key) | — | `tlsClientCert`, `tlsClientKey` |
| `tlsSelfSignedCA` | None | Custom CA | — | `tlsCACert` |
| `advancedHttp` | Basic | — | `timeout=30`, `keepCookies=["session_id"]` | `basicAuthPassword` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings
and returns a fully-defaulted, validated `Config`:

1. **Parse** — copy `settings.URL`, `settings.BasicAuthEnabled`,
   `settings.BasicAuthUser` into `Config`, unmarshal `settings.JSONData`
   into the jsonData portion of the same struct, and copy the four
   decrypted secrets into `DecryptedSecureJSONData`. The Parca plugin has
   no upstream `LoadSettings` to mirror — Parca's server-side code never
   unmarshals `settings.JSONData` at all.
2. **`ApplyDefaults`** — intentionally a no-op (see
   [Modeling decisions](#modeling-decisions)).
3. **`Validate`** — enforce the runtime contract: URL is required, Basic
   auth requires a username, mTLS requires serverName + client cert +
   client key, custom-CA requires the CA PEM, and `timeout` must be
   non-negative. Errors are joined so every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with
`datasource_uid`, `datasource_name`, and `plugin` labels so log lines
carry request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are still
exported for callers that want to compose them themselves (e.g.
provisioning preview, schema-example round-trip, tests that need to
distinguish parse-level from policy-level errors). Skip them by never
calling `LoadConfig` in those flows — assemble a `Config` directly.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while
researching upstream. All preserved verbatim in the schema — the schema
records what the plugin **does**, not what it **should** do; these notes
exist so reviewers can reproduce each finding and decide separately
whether to fix upstream.

1. **Upstream placeholder typo preserved**: `@grafana/plugin-ui`'s
   `TLSClientAuth.tsx:109` sets the client key placeholder to
   `Begins with --- RSA PRIVATE KEY CERTIFICATE ---` — an RSA private key
   is not a "certificate". Preserved verbatim in
   `secureJsonData_tlsClientKey.ui.placeholder`. This is a plugin-ui typo
   shared across all data sources that use `Auth`.
2. **No `pkg/models/settings.go` and no jsonData reads**: unlike most
   sibling plugins the Parca datasource does not own a typed backend
   settings struct and never unmarshals `settings.JSONData`. All read
   paths go through `settings.URL` and the SDK HTTP client. This entry's
   `Config` fills that gap — it is the intended shape a plugin-owned
   settings loader would produce.
3. **No connection health check for URL**: `pkg/parca/plugin.go`'s
   `CheckHealth` (`:133-148`) does not pre-validate `settings.URL`; it
   just issues a `ProfileTypes` request and reports the transport error
   verbatim. We surface this as a `requiredWhen: "true"` constraint on
   `root_url` so provisioning tooling can reject an empty URL upfront.
4. **Base `DataSourceJsonData` fields are unused**: Grafana core embeds
   `DataSourceJsonData` into any jsonData shape and it carries `authType`,
   `defaultRegion`, `profile`, `manageAlerts`, `alertmanagerUid`,
   `disableGrafanaCache`. The Parca editor writes none of them, and the
   Parca backend reads none of them. They are omitted from the schema.
5. **`SecureSocksProxySettings` also writes to jsonData**: the excluded
   Secure Socks Proxy widget writes `jsonData.enableSecureSocksProxy` and
   related fields. Provisioning payloads that include those keys will not
   round-trip through this schema — they will be preserved in the raw
   `JSONData` but not be represented in `Config` or `SettingsExamples`.
6. **Deprecation date is a frontend constant**: the deprecation banner
   text is built from `DEPRECATION_DATE = '2nd of January 2027'`
   (`ConfigEditor.tsx:17`), a plain frontend constant. After that date
   the plugin still functions unchanged — the string is informational
   only. The same value is exposed as the `DeprecationDate` Go constant
   here so backend tooling can surface it consistently.
7. **npm-only plugin-ui version pin**: `package.json` pins
   `@grafana/plugin-ui@0.13.1`, but the plugin-ui git repository's tags
   stop at `v0.12.0` on `main`. The 0.13.x line is an npm-only release
   train. The external-component references above use plugin-ui's `main`
   sources; the labels/placeholders/tooltips/storage keys on the code
   paths this plugin exercises have been stable across 0.12 → 0.15, so
   the pinned version renders and persists exactly what is documented.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go
  validator in this repo) — passes (invoked by the conformance suite).
- JSON Schema validation against
  [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft 2020-12,
  `additionalProperties: false`) — passes (invoked by the conformance
  suite).
- `go test ./...` inside `registry/` — passes on every entry, including
  the new `parca` package (schema bundle shape, secure values, examples,
  `LoadConfig` incl. TLS variants and malformed input,
  `SchemaArtifactInSync` guard, `JSONDataMatchesStruct`,
  `JSONDataTypesMatchStruct`).
- `settings.go`/`schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: exports the three canonical types (`RootConfig`,
  `JsonDataConfig`, `SecureJsonDataConfig`) — reviewed by hand against
  the frontend sources; no `tsc` runner is wired into the registry
  module.
