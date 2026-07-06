# grafana-azureprometheus-datasource

Declarative configuration schema for the [Azure Monitor Managed Service for
Prometheus datasource plugin](https://github.com/grafana/azure-prometheus-datasource)
(`grafana-azureprometheus-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/azure-prometheus-datasource`
- **Ref**: `main`
- **Commit SHA**: `fe45d2eea9c7d923fbef1a98b8e0be468781525b` (HEAD at time of
  authoring — `Updating plugin-ci-workflows (#257)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders,
tooltips, option labels/values, section titles, help markdown, defaults,
validations, dependency and required-when expressions, storage keys, storage
targets, value types, group titles, and instructions — is traceable to a
specific `file:line` in the upstream repo at this SHA or in a pinned external
component. See [Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone --depth 20 https://github.com/grafana/azure-prometheus-datasource
cd azure-prometheus-datasource
git checkout fe45d2eea9c7d923fbef1a98b8e0be468781525b
```

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: root `URL` tagged `json:"-"`, Prometheus + Azure jsonData fields, `DecryptedSecureJSONData`), `PluginID`, `AuthType` / `HTTPMethod` / `PromApplication` / `PrometheusCacheLevel` / `QueryEditorMode` / `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, produces the `pluginschema.PluginSchema` bundle via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, `Validate`, and `EffectiveAuthType` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the
shared [`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

### Plugin repo (`github.com/grafana/azure-prometheus-datasource@fe45d2e`)

| File | What was read |
| --- | --- |
| `src/plugin.json:2-6,94-108` | `pluginType` (`id` = `"grafana-azureprometheus-datasource"`), `pluginName` (`name` = `"Azure Monitor Managed Service for Prometheus"`), empty `info.links` — no dedicated docs URL |
| `src/configuration/ConfigEditor.tsx:20-80` | Top-level editor composition: `DataSourceDescription` (docs link points at the vanilla Prometheus docs — flagged as `// TODO Update this to Azure prom docs when available`), `DataSourceHttpSettingsOverhaul`, then a collapsible `ConfigSection` "Advanced settings" wrapping `AdvancedHttpSettings`, `AlertingSettingsOverhaul<PromOptions>`, and `PromSettings` |
| `src/configuration/ConfigEditor.tsx:41-48` | Inline error banner rejecting `options.access === 'direct'` (Browser access mode); nothing prevents storing that value, only the UI complains |
| `src/configuration/DataSourceHttpSettingsOverhaul.tsx:36-144` | The `Auth` wrapper: registers a single `customMethods` entry keyed `azureAuthId = 'custom-azureAuthId'`, sets `visibleMethods=[azureAuthId]` (`:134`), and `onAuthMethodSelect` (`:121-131`) always writes `basicAuth: false`, `withCredentials: false`, and `jsonData.oauthPassThru: false` |
| `src/configuration/DataSourceHttpSettingsOverhaul.tsx:37,101-117` | `jsonData['prometheus-type-migration']` sentinel + the `"Data source migrated"` warning banner that renders when the sentinel is truthy |
| `src/configuration/AzureAuthSettings.tsx:17-43` | Wraps `AzureCredentialsForm` with `managedIdentityEnabled`, `workloadIdentityEnabled`, `userIdentityEnabled` from `config.azure` |
| `src/configuration/AzureCredentialsForm.tsx:33-79` | `authType` select — options `App Registration` (`clientsecret`, always), `Managed Identity` (`msi`, config-gated), `Workload Identity` (`workloadidentity`, config-gated), `Current User` (`currentuser`, config-gated). Default fallthrough favours `currentuser` when userIdentity is enabled, then `workloadidentity`, then `clientsecret` (`:69-72`) |
| `src/configuration/AzureCredentialsForm.tsx:131-273` | Inline `clientsecret` inputs (`Azure Cloud` select, `Directory (tenant) ID` input, `Application (client) ID` input, `Client Secret` input) rendered directly inside the form for `authType === 'clientsecret'`. `Client Secret` renders as an `Input` with placeholder `configured` + a `Reset` button when the current value is a `Symbol` (i.e. `secureJsonFields.azureClientSecret == true`) |
| `src/configuration/CurrentUserFallbackCredentials.tsx:21-208` | `serviceCredentialsEnabled` radio + fallback `authType` select rendered inside a `ConfigSection` when `authType === 'currentuser'`; delegates the `clientsecret` fallback to `AppRegistrationCredentials` |
| `src/configuration/AppRegistrationCredentials.tsx:14-153` | The stand-alone App-Registration sub-editor used inside `CurrentUserFallbackCredentials`. Fields: `Azure Cloud`, `Directory (tenant) ID` (required), `Application (client) ID` (required), `Client Secret` (required) with the same `configured`/`Reset` pattern |
| `src/configuration/AzureCredentialsConfig.ts:15-72` | Helpers (`getAzureCloudOptions`, `getDefaultCredentials`, `getCredentials`, `updateCredentials`, `setDefaultCredentials`, `resetCredentials`) and `AzurePromDataSourceOptions` (`extends PromOptions, AzureDataSourceJsonData` plus `azureEndpointResourceId?: string` and `'prometheus-type-migration'?: boolean`) |
| `pkg/azureauth/azure.go:18-75` | `ConfigureAzureAuthentication`: `azcredentials.FromDatasourceData(jsonData, DecryptedSecureJSONData)` → `getPrometheusScopes` reads the resolved Azure cloud's `prometheusResourceId` property, appends `.default` to build the OAuth scope → `AddAzureAuthentication` with `AllowUserIdentity()` |
| `pkg/datasource.go:19-86` | `NewDatasource` builds a `promlib.Service` whose `extendClientOpts` invokes `azureauth.ConfigureAzureAuthentication` when `azureSettings.AzureAuthEnabled` (`:78-83`). All query/resource/health-check execution is delegated to `promlib` |

### External components (pinned to `package.json` / `go.mod`)

- **`@grafana/azure-sdk@0.1.0`** (`package.json:76`) — `AzureCredentials`,
  `AzureAuthType`, `AzureDataSourceJsonData`, `AzureDataSourceSecureJsonData`,
  `updateDatasourceCredentials`, `getDatasourceCredentials`, `getAzureClouds`,
  `getDefaultAzureCloud`.
- **`@grafana/plugin-ui@0.13.1`** (`package.json:79`) — `Auth`, `AuthMethod`,
  `ConnectionSettings`, `convertLegacyAuthProps`, `AdvancedHttpSettings`,
  `DataSourceDescription`, `ConfigSection`.
- **`@grafana/prometheus@12.4.2`** (`package.json:80`) — `PromOptions`,
  `PromSettings`, `AlertingSettingsOverhaul`, `overhaulStyles`. Every
  Prometheus jsonData field this schema carries (except the Azure-specific
  three) is defined in `packages/grafana-prometheus/src/types.ts` and rendered
  by `PromSettings.tsx`.
- **`@grafana/ui@12.4.2`** — `Input`, `Select`, `Switch`, `TagsInput`, `Alert`,
  `SecureSocksProxySettings` (excluded), `TextLink`.
- **`@grafana/data@12.4.2`**, **`@grafana/runtime@12.4.2`** — editor plumbing.
- **`github.com/grafana/grafana-prometheus-datasource/pkg/promlib@v0.0.12`**
  (`go.mod:6`) — backend `PromOptions` (`pkg/promlib/models/settings.go`)
  parsed by `ParsePromOptions` + `ApplyDefaults` + `Validate`. This entry's
  `Config` struct mirrors that shape verbatim.
- **`github.com/grafana/grafana-azure-sdk-go/v2@v2.4.1`** (`go.mod:70`) —
  `azcredentials/credentials.go` (the `AzureAuthType` constants),
  `azcredentials/builder.go` (`FromDatasourceData` and its per-authType
  parse), `azsettings.AzureSettings.GetCloud(...)`.

## Field inventory

| Schema `id` | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | Prometheus server URL | Yes (via `promlib`) |
| `jsonData_azureCredentials` | `azureCredentials` | `jsonData` | Authentication | Yes (`pkg/azureauth/azure.go:23`) |
| `secureJsonData_azureClientSecret` | `azureClientSecret` | `secureJsonData` | Client Secret | Yes (Azure SDK) |
| `secureJsonData_clientSecret` | `clientSecret` | `secureJsonData` | — (legacy) | Yes (Azure SDK legacy fallback) |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (managed by editor) | Yes (SDK forward) |
| `jsonData_azureEndpointResourceId` | `azureEndpointResourceId` | `jsonData` | — (no UI) | Yes (backend-only override) |
| `jsonData_prometheusTypeMigration` | `prometheus-type-migration` | `jsonData` | — (banner sentinel) | No (frontend-only banner) |
| `jsonData_manageAlerts` | `manageAlerts` | `jsonData` | Manage alerts via Alerting UI | Yes (`promlib`) |
| `jsonData_allowAsRecordingRulesTarget` | `allowAsRecordingRulesTarget` | `jsonData` | Allow as recording rules target | Yes |
| `jsonData_timeout` | `timeout` | `jsonData` | Timeout | Yes (SDK) |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | Allowed cookies | Yes (SDK) |
| `jsonData_timeInterval` | `timeInterval` | `jsonData` | Scrape interval | Yes (`promlib`) |
| `jsonData_queryTimeout` | `queryTimeout` | `jsonData` | Query timeout | Yes (`promlib`) |
| `jsonData_defaultEditor` | `defaultEditor` | `jsonData` | Default editor | Parsed; UI-only |
| `jsonData_disableMetricsLookup` | `disableMetricsLookup` | `jsonData` | Disable metrics lookup | Parsed; UI-only |
| `jsonData_prometheusType` | `prometheusType` | `jsonData` | Prometheus type | Parsed; heuristics |
| `jsonData_prometheusVersion` | `prometheusVersion` | `jsonData` | Version | Parsed; heuristics |
| `jsonData_cacheLevel` | `cacheLevel` | `jsonData` | Cache level | Parsed; editor caching |
| `jsonData_incrementalQuerying` | `incrementalQuerying` | `jsonData` | Incremental querying (beta) | Parsed |
| `jsonData_incrementalQueryOverlapWindow` | `incrementalQueryOverlapWindow` | `jsonData` | Query overlap window | Parsed |
| `jsonData_disableRecordingRules` | `disableRecordingRules` | `jsonData` | Disable recording rules (beta) | Parsed |
| `jsonData_customQueryParameters` | `customQueryParameters` | `jsonData` | Custom query parameters | Yes (middleware) |
| `jsonData_httpMethod` | `httpMethod` | `jsonData` | HTTP method | Yes |
| `jsonData_seriesLimit` | `seriesLimit` | `jsonData` | Series limit | Parsed |
| `jsonData_seriesEndpoint` | `seriesEndpoint` | `jsonData` | Use series endpoint | Parsed |
| `jsonData_exemplarTraceIdDestinations` | `exemplarTraceIdDestinations` | `jsonData` | Exemplars | Parsed; result transformer |
| `jsonData_maxSamplesProcessedWarningThreshold` | `maxSamplesProcessedWarningThreshold` | `jsonData` | — (no UI) | Yes (backend-only) |
| `jsonData_maxSamplesProcessedErrorThreshold` | `maxSamplesProcessedErrorThreshold` | `jsonData` | — (no UI) | Yes (backend-only) |

### Frontend-only settings

- **`prometheus-type-migration`** — sentinel flag toggling the migration
  banner at `DataSourceHttpSettingsOverhaul.tsx:101-117`. Nothing in the
  runtime depends on it; the editor reads it and renders a warning.

### Backend-only settings

- **`maxSamplesProcessedWarningThreshold` / `maxSamplesProcessedErrorThreshold`**
  live in `PromOptions` (`pkg/promlib/models/settings.go:41-42`) but are gated
  by `PromSettings`'s `showQuerySamplesProcessedThresholdFields` prop, which
  this plugin's `ConfigEditor` never passes. Provisioning can still set them.
- **`azureEndpointResourceId`** — read by `pkg/azureauth/azure.go` (through
  `azcredentials.FromDatasourceData`, which passes the whole jsonData to the
  SDK); the editor never renders an input.
- **`oauthPassThru`** — set by `@grafana/azure-sdk`
  `updateDatasourceCredentials` for `currentuser` and cleared by
  `DataSourceHttpSettingsOverhaul.onAuthMethodSelect` on every save. Consumed
  by the SDK's shared HTTP client.

### Excluded settings

- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy`) — rendered by
  `DataSourceHttpSettingsOverhaul.tsx:137-142` when
  `config.secureSocksDSProxyEnabled` is true. Excluded per AGENTS.md.
- **Basic auth, OAuth Forward, and TLS** — `visibleMethods=[azureAuthId]`
  locks the editor to Azure auth only, and `onAuthMethodSelect` clears
  `basicAuth` / `withCredentials` / `oauthPassThru` on every save. TLS
  settings are not rendered because they live inside the `Auth` component's
  method sub-panels, and no auth method other than Azure is visible. These
  fields can still be set via provisioning (SDK consumes them) but are not
  modeled in this schema. If a future plugin change re-enables them, add
  them here mirroring the `prometheus` registry entry.
- **Custom HTTP headers** (`jsonData.httpHeaderName<N>` /
  `secureJsonData.httpHeaderValue<N>`) — not rendered by this plugin's
  editor (custom-headers UI lives inside `@grafana/plugin-ui`'s `Auth`
  method sub-panels, which are not exposed here). Same rationale as basic
  auth / TLS.

## Where the types are defined

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `AzurePromDataSourceOptions`, `AzurePromDataSourceSettings` | `src/configuration/AzureCredentialsConfig.ts:67-72` | plugin |
| `AzureCredentials`, `AzureAuthType`, `AzureDataSourceJsonData`, `AzureDataSourceSecureJsonData` | `src/credentials/AzureCredentials.ts`, `src/settings.ts` | `@grafana/azure-sdk` `0.1.0` |
| `PromOptions`, `PromApplication`, `PrometheusCacheLevel`, `ExemplarTraceIdDestination` | `packages/grafana-prometheus/src/types.ts` | `@grafana/prometheus` `12.4.2` |
| `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps`, `DataSourceSettings` | `packages/grafana-data/src/` | `@grafana/data` `12.4.2` |
| `Auth`, `AuthMethod`, `ConnectionSettings`, `convertLegacyAuthProps`, `AdvancedHttpSettings`, `DataSourceDescription`, `ConfigSection` | `src/components/ConfigEditor/…` | `@grafana/plugin-ui` `0.13.1` |
| `Input`, `Select`, `Switch`, `TagsInput`, `Alert`, `SecureSocksProxySettings` (excluded), `TextLink` | `packages/grafana-ui/src/components/` | `@grafana/ui` `12.4.2` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `PromOptions` (jsonData), `DataSourceJsonData` (base struct), `ExemplarTraceIDDestination`, `ParsePromOptions`, `ApplyDefaults`, `Validate` | `pkg/promlib/models/settings.go` | `github.com/grafana/grafana-prometheus-datasource/pkg/promlib` `v0.0.12` |
| `ConfigureAzureAuthentication`, `getPrometheusScopes`, `audienceToScopes` | `pkg/azureauth/azure.go:18-76` | plugin |
| `NewDatasource`, `extendClientOpts`, `Datasource` | `pkg/datasource.go:19-86` | plugin |
| `AzureCredentials`, `AzureAuthType`, `FromDatasourceData` | `azcredentials/credentials.go`, `azcredentials/builder.go` | `github.com/grafana/grafana-azure-sdk-go/v2` `v2.4.1` |
| `AzureSettings`, `GetCloud`, `cloud.Properties["prometheusResourceId"]` | `azsettings/…` | `github.com/grafana/grafana-azure-sdk-go/v2` `v2.4.1` |
| `backend.DataSourceInstanceSettings`, `HTTPClientOptions(ctx)` | `backend/common.go`, `backend/httpclient/` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |

The models in this entry flatten that spread into a single Go `Config` type
(root `URL` tagged `json:"-"`, Prometheus + Azure jsonData fields,
`DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant list.
`settings.ts` keeps the three canonical TypeScript types (`RootConfig`,
`JsonDataConfig`, `SecureJsonDataConfig`).

## Modeling decisions

- **Opaque `azureCredentials`**: modeled as a single `valueType: any` field
  with `role: auth.discriminator`, following the same pattern as
  `grafana-azure-monitor-datasource`. The nested union is exercised by
  `@grafana/azure-sdk`'s `AzureCredentialsForm` on the write side and by
  `grafana-azure-sdk-go/v2/azcredentials.FromDatasourceData` on the read
  side; flattening the union into individual dsconfig fields would require
  discriminator-driven conditional visibility that the `dsconfig` vocabulary
  does not currently express for object subfields.
- **No basic-auth / TLS / OAuth Forward fields**: locked out by
  `visibleMethods=[azureAuthId]` and cleared on every save. See
  [Excluded settings](#excluded-settings).
- **Prometheus knobs mirrored from vanilla Prometheus**: everything under
  Advanced settings comes from `@grafana/prometheus`'s `PromSettings`
  component, which is identical to the vanilla Prometheus plugin. Kept the
  same labels/descriptions/defaults so tools that already understand the
  Prometheus schema can consume this one without re-authoring.
- **`prometheus-type-migration` field ID**: storage key has a hyphen; the
  schema field ID is camelCased (`jsonData_prometheusTypeMigration`) per
  AGENTS.md, and the raw storage key stays on the `key` property.
- **Root fields**: only `url` is carried. Basic-auth root fields are omitted
  because the plugin actively clears them on save; this differs from the
  vanilla Prometheus entry, which carries them because the vanilla Prometheus
  editor uses them.
- **Field ID naming convention**: IDs are prefixed with their storage target
  for discoverability — `root_`, `jsonData_`, or `secureJsonData_` — followed
  by the camelCase form of the storage key.
- **Flat `Config` in Go**: mirrors backend `PromOptions` verbatim plus the
  three Azure-specific jsonData fields; keeps `URL` at root with `json:"-"`.
  See [`settings.go`](settings.go).
- **`SecureJsonDataConfig` is a key list**: secure values are write-only, so
  the type is just the array of secret key names (`azureClientSecret`,
  `clientSecret`).
- **`LoadConfig` phases**: `parse → ApplyDefaults → Validate`, matching the
  registry-wide convention. `ApplyDefaults` uppercases + defaults
  `httpMethod` to POST (mirroring `pkg/promlib/models/settings.go:82-87`);
  `Validate` enforces URL, HTTP method, `azureCredentials.authType`, and
  per-authType secret requirements.

## Settings examples matrix

`SettingsExamples()` (`schema.go`) provides:

| Example key | authType | Secret key | Notes |
| --- | --- | --- | --- |
| `""` (default) | none | `azureClientSecret` (empty) | Bare defaults; users must add credentials |
| `clientSecret` | `clientsecret` | `azureClientSecret` | App Registration |
| `managedIdentity` | `msi` | — (empty `azureClientSecret`) | Requires `azure.managedIdentityEnabled` |
| `workloadIdentity` | `workloadidentity` | — (empty `azureClientSecret`) | Requires `azure.workloadIdentityEnabled` |
| `currentUser` | `currentuser` + `clientsecret` fallback | `azureClientSecret` | Requires `azure.userIdentityEnabled` |
| `customEndpointResourceID` | `clientsecret` | `azureClientSecret` | Provisioning-only `azureEndpointResourceId` override |
| `legacyClientSecret` | `clientsecret` | `clientSecret` | Legacy secure key |
| `migratedFromPrometheus` | `clientsecret` | `azureClientSecret` | Includes `prometheus-type-migration: true` |

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while
researching upstream. Preserved verbatim in the schema — the schema records
what the plugin **does**, not what it **should** do; these notes exist so
reviewers can reproduce each finding and decide separately whether to fix
upstream.

1. **`docsLink` still points at vanilla Prometheus** —
   `src/configuration/ConfigEditor.tsx:52` sets
   `docsLink="https://grafana.com/docs/grafana/latest/datasources/prometheus/configure-prometheus-data-source/"`
   with a `// TODO Update this to Azure prom docs when available`. The
   `docsTip` helper (`ConfigEditor.tsx:88`) references
   `https://grafana.com/grafana/plugins/grafana-azureprometheus-datasource/`
   instead — the two are out of sync. `pluginType_id.info.links` is empty.
2. **Auth picker locked to Azure but state is still cleared on save** —
   `DataSourceHttpSettingsOverhaul.onAuthMethodSelect` (`:121-131`) is a
   full state reset that always writes `basicAuth: false`,
   `withCredentials: false`, and `oauthPassThru: false`, even though the
   only selectable method is `azureAuthId`. Provisioning that sets any of
   those fields will see them cleared the first time an operator saves in
   the UI.
3. **`useEffectOnce` re-writes `jsonData` on first mount** — both
   `ConfigEditor.tsx:33-37` and `DataSourceHttpSettingsOverhaul.tsx:25-32`
   register `useEffectOnce` hooks that call `onOptionsChange` with an
   otherwise-unchanged object. This dirties the datasource on every editor
   open and can trigger auto-save side effects in some Grafana UIs.
4. **`prometheus-type-migration` is a magic string** — the sentinel key
   contains a hyphen and lives in `jsonData` alongside camelCased fields.
   Storage-format-wise it works, but it's an outlier that scripts scanning
   `jsonData` for known keys may miss.
5. **Empty `docsLink` for the "TODO Azure prom docs"** — combined with the
   empty `info.links` in `plugin.json`, downstream tooling that expects a
   dedicated docs URL for this plugin has no canonical value to render.
   This entry uses the same vanilla-Prometheus URL the editor already ships
   to keep parity.
6. **`getDefaultCredentials` favours `currentuser` when userIdentity is
   enabled** — `AzureCredentialsConfig.ts:24-29` returns
   `{ authType: 'currentuser' }` whenever `config.azure?.userIdentityEnabled`
   is truthy, even without any hint the user wants user-scoped auth. This
   makes a first-time save on a Grafana instance with user identity enabled
   silently commit `currentuser` auth — the config editor's own default
   auth-type-change fallback (`AzureCredentialsForm.tsx:69-72`) preserves
   this. See also grafana-azure-monitor-datasource which has the same quirk.
7. **`ConfigureAzureAuthentication` swallows `nil` credentials** —
   `pkg/azureauth/azure.go:29` checks `if credentials != nil` and only then
   installs Azure auth; a config with no `azureCredentials` object produces
   an unauthenticated Prometheus client. The datasource still initialises
   successfully but every query fails at Azure Monitor with a 401 —
   surface-level "auth failed" rather than "credentials missing".
8. **`azureEndpointResourceId` never referenced in Go code** — it exists on
   the frontend `AzurePromDataSourceOptions` (`AzureCredentialsConfig.ts:68`)
   and is cleared by `resetCredentials`, but `pkg/azureauth/azure.go` never
   reads it directly — the scope derivation only consults
   `azureCloud.Properties["prometheusResourceId"]`. Preserved in the schema
   as backend-only because the value flows through
   `azcredentials.FromDatasourceData` and may still influence scope handling
   in the shared azure-sdk-go library, but its effect on this specific
   plugin is unclear.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft 2020-12, strict) — passes via conformance tests.
- `go generate ./...` inside this directory regenerates the schema artifacts
  cleanly.
- `go test ./...` in the shared `registry/` module — passes (schema bundle
  shape, secure values, examples, `LoadConfig` incl. Azure auth variants
  and malformed input, `SchemaArtifactInSync` guard,
  `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`).
- `go build`, `go vet`, `gofmt` on this package — clean.
