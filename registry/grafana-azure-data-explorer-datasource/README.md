# grafana-azure-data-explorer-datasource

Declarative configuration schema for the
[Azure Data Explorer datasource plugin](https://github.com/grafana/azure-data-explorer-datasource)
(`grafana-azure-data-explorer-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/azure-data-explorer-datasource`
- **Ref**: `main`
- **Commit SHA**: `febca70fe0814596ffb8a7e399d9dc62c2196e0b`
  (`docs: add signed commits requirement to CONTRIBUTING.md (#1784)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders,
tooltips, option labels/values, help markdown, defaults, storage keys,
storage targets, value types, group titles, and instructions — is
traceable to a specific `file:line` in the upstream repo (or a pinned
external dependency) at this SHA. See [Field provenance](#field-provenance)
below.

To reproduce this research:

```bash
git clone https://github.com/grafana/azure-data-explorer-datasource
cd azure-data-explorer-datasource
git checkout febca70fe0814596ffb8a7e399d9dc62c2196e0b
```

If upstream `main` has advanced past this SHA, re-diff the sources listed
under [Sources researched](#sources-researched) before merging any
changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and LLM instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig` (empty), `JsonDataConfig` (jsonData shape), `SecureJsonDataConfig` (secure key list), plus `AzureCredentials` union |
| [`settings.go`](settings.go) | Go `Config` (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `AuthType` / `DataConsistency` / `EditorMode` / `LegacyCloudName` / `SecureJsonDataKey` typed constants, and the `LoadConfig` utility with `ApplyDefaults` + `Validate` |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, exposes `NewSchema()` via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `Validate`, `EffectiveAuthType`, `ApplyDefaults` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...`; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of
the shared [`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

### Plugin repo (`github.com/grafana/azure-data-explorer-datasource@febca70`)

| File | What was read |
| --- | --- |
| `src/plugin.json:1-61` | `pluginType` (`id`), `pluginName` (`name`), docs URL (project site link) |
| `src/components/ConfigEditor/index.tsx:37-215` | Editor shell — renders `DataSourceDescription`, `ConfigHelp`, `AzureCredentialsForm`, `ConnectionConfig`, and the collapsible `Additional settings` section wrapping `QueryConfig`, `DatabaseConfig`, `ApplicationConfig`, `TrackingConfig` and the inline `keepCookies` TagsInput |
| `src/components/ConfigEditor/AzureCredentialsForm.tsx:31-311` | Authentication type select, per-Grafana-config visibility of `msi` / `workloadidentity` / `currentuser`, per-feature-toggle visibility of `clientsecret-obo`, and the App Registration sub-form (Azure Cloud select, Tenant ID, Client ID, Client Secret) |
| `src/components/ConfigEditor/AzureCredentialsConfig.ts:13-74` | `getOboEnabled`, `hasCredentials`, `getDefaultCredentials`, `getCredentials`, `getLegacyCredentials`, `updateCredentials` — the frontend default/legacy credential loader |
| `src/components/ConfigEditor/ConnectionConfig.tsx:16-36` | Default cluster URL input (label, description, placeholder `https://yourcluster.kusto.windows.net`) |
| `src/components/ConfigEditor/QueryConfig.tsx:16-141` | Query timeout, dynamic caching switch, cache max age, data consistency select, default editor mode select (defaults applied via `useEffect`) |
| `src/components/ConfigEditor/DatabaseConfig.tsx:135-234` | Default database select + Reload schema button, Use managed schema switch, Schema mappings list (target select → name input pairs) |
| `src/components/ConfigEditor/ApplicationConfig.tsx:13-37` | Application name input (label, description, placeholder `Grafana-ADX`) |
| `src/components/ConfigEditor/TrackingConfig.tsx:13-45` | Send username header switch — writes `jsonData.enableUserTracking` and documents the `x-ms-user-id` / `x-ms-client-request-id` header forwarding |
| `src/components/ConfigEditor/ConfigHelp.tsx:14-92` | 3-step configuration help (AAD app, DB user, config) rendered above the credentials form |
| `src/types/index.ts:50-135` | `EditorMode` / `AdxQueryType` enums, `SchemaMapping` / `SchemaMappingType`, `AdxDataSourceOptions` and `AdxDataSourceSecureOptions` |
| `pkg/plugin.go` + `pkg/azuredx/datasource.go:34-76` | `NewDatasource` — loads `DatasourceSettings`, reads `DecryptedSecureJSONData["OpenAIAPIKey"]`, resolves credentials via `adxcredentials.FromDatasourceData`, builds the ADX HTTP client |
| `pkg/azuredx/models/settings.go:18-125` | Backend `DatasourceSettings.Load` — unmarshals jsonData, sanitizes `ClusterURL`, parses `QueryTimeout` (default 30s), converts to a Kusto TimeSpan, reads `GF_PLUGIN_ENFORCE_TRUSTED_ENDPOINTS` / `GF_PLUGIN_ALLOW_USER_TRUSTED_ENDPOINTS` / `GF_PLUGIN_USER_TRUSTED_ENDPOINTS` from the environment |
| `pkg/azuredx/adxauth/adxcredentials/builder.go:12-121` | `FromDatasourceData` (modern-then-legacy), `getFromLegacy` (requires `tenantId + clientId + secureJsonData.clientSecret`), `ensureOnBehalfOfSupported` (rejects OBO without `oauthPassThru`), `resolveLegacyCloudName` (`azuremonitor` / `chinaazuremonitor` / `govazuremonitor` → modern names) |
| `pkg/azuredx/resource_handler.go:25-89` | `askOpenAI` handler consuming `settings.OpenAIAPIKey` |
| `go.mod:8-9` | Pinned SDK versions: `grafana-azure-sdk-go/v2 v2.4.1`, `grafana-plugin-sdk-go v0.292.1` |
| `package.json:103-110` | Pinned frontend SDK versions: `@grafana/azure-sdk 0.1.0`, `@grafana/plugin-ui 0.13.1`, `@grafana/ui 12.4.2` |

### External components (read at the exact pinned versions)

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `AzureAuthType`, `AzureCredentials` union | `@grafana/azure-sdk` `0.1.0` | `grafana/grafana-azure-sdk-react` `src/credentials/AzureCredentials.ts:1-91` | The seven `authType` values, the per-authType field shape (ADX only exposes five of them) |
| `AzureCredentialsConfig` (`getDatasourceCredentials`, `updateDatasourceCredentials`, `getClientSecret`, `isCredentialsComplete`) | `@grafana/azure-sdk` `0.1.0` | `grafana/grafana-azure-sdk-react` `src/credentials/AzureCredentialsConfig.ts` | Which storage keys the editor writes per authType (in particular that `oauthPassThru` is set for OBO), how "already-configured" is detected via `secureJsonFields` |
| `AzureDataSourceJsonData`, `AzureDataSourceSecureJsonData` | `@grafana/azure-sdk` `0.1.0` | `grafana/grafana-azure-sdk-react` `src/settings.ts:5-28` | Base jsonData / secureJsonData shape ADX extends |
| `getAzureClouds` / `getDefaultAzureCloud` / `resolveLegacyCloudName` | `@grafana/azure-sdk` `0.1.0` | `grafana/grafana-azure-sdk-react` `src/clouds.ts:1-48` | Which cloud identifiers the SDK renders + legacy cloud mapping |
| `DataSourceDescription`, `ConfigSection`, `ConfigSubSection` | `@grafana/plugin-ui` `0.13.1` | `grafana/plugin-ui` `src/components/ConfigEditor/` | Section titles and collapsible layout — no storage keys |
| `TagsInput` | `@grafana/ui` `12.4.2` | `grafana/grafana` `packages/grafana-ui/src/components/TagsInput` | Renders the `keepCookies` list inline in `ConfigEditor/index.tsx:146-164` |
| `azcredentials.FromDatasourceData` (backend) | `grafana-azure-sdk-go/v2` `v2.4.1` | `grafana/grafana-azure-sdk-go` `azcredentials/builder.go` | Per-authType parse and secret requirements — the checks `Validate()` mirrors |
| `azcredentials` constants (backend) | `grafana-azure-sdk-go/v2` `v2.4.1` | `grafana/grafana-azure-sdk-go` `azcredentials/credentials.go` | The backend-recognised authType constants (superset of what ADX exposes) |

## Field provenance

| Schema `id` | Storage key | Target | Editor label | Placeholder / default / options source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_clusterUrl` | `clusterUrl` | jsonData | `ConnectionConfig.tsx:18` (`Default cluster URL (Optional)`) | `ConnectionConfig.tsx:30` (`https://yourcluster.kusto.windows.net`) | `AdxDataSourceOptions.clusterUrl: string`, `src/types/index.ts:124`; backend `DatasourceSettings.ClusterURL`, `pkg/azuredx/models/settings.go:19` | Role `endpoint.baseUrl`; sanitized by backend at load |
| `jsonData_azureCredentials` | `azureCredentials` | jsonData | `AzureCredentialsForm.tsx:143-159` (`Authentication Method`) | Options `AzureCredentialsForm.tsx:31-71`; no default (Grafana-config-dependent — `AzureCredentialsConfig.ts:22-34`) | Union `AzureCredentials`, `azure-sdk/src/credentials/AzureCredentials.ts:74-81`; modeled as `valueType: any` | Role `auth.discriminator` |
| `secureJsonData_azureClientSecret` | `azureClientSecret` | secureJsonData | `AzureCredentialsForm.tsx:262,286` (`Client Secret`) | `AzureCredentialsForm.tsx:300` (`XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX`) | `AzureDataSourceSecureJsonData.azureClientSecret?: string`, `azure-sdk/src/settings.ts:18` | Managed by `@grafana/azure-sdk` |
| `secureJsonData_clientSecret` | `clientSecret` | secureJsonData | — (no UI; legacy fallback) | — | `AzureDataSourceSecureJsonData.clientSecret?: string`, `azure-sdk/src/settings.ts:27` | Backend fallback `pkg/azuredx/adxauth/adxcredentials/builder.go:57` |
| `jsonData_oauthPassThru` | `oauthPassThru` | jsonData | — (SDK-managed) | — | `AzureDataSourceJsonData.oauthPassThru?: boolean`, `azure-sdk/src/settings.ts:8` | Role `auth.forwardOAuthToken.enabled`; hard requirement for OBO |
| `jsonData_queryTimeout` | `queryTimeout` | jsonData | `QueryConfig.tsx:46` (`Query timeout`) | `QueryConfig.tsx:56` (`30s`); backend default `30s` (`pkg/azuredx/models/settings.go:65-71`) | `AdxDataSourceOptions.queryTimeout: string`, `src/types/index.ts:117`; backend `DatasourceSettings.QueryTimeoutRaw`, `settings.go:29` | Parsed as Go duration; >1h fails |
| `jsonData_dynamicCaching` | `dynamicCaching` | jsonData | `QueryConfig.tsx:63` (`Use dynamic caching`) | Description `QueryConfig.tsx:64-67` | `AdxDataSourceOptions.dynamicCaching: boolean`, `src/types/index.ts:120`; backend `DatasourceSettings.DynamicCaching`, `settings.go:23` | — |
| `jsonData_cacheMaxAge` | `cacheMaxAge` | jsonData | `QueryConfig.tsx:77` (`Cache max age`) | `QueryConfig.tsx:87` (`0m`) | `AdxDataSourceOptions.cacheMaxAge: string`, `src/types/index.ts:119`; backend `DatasourceSettings.CacheMaxAge`, `settings.go:22` | — |
| `jsonData_dataConsistency` | `dataConsistency` | jsonData | `QueryConfig.tsx:94` (`Data consistency`) | Options `QueryConfig.tsx:16-19`; default `strongconsistency` from `QueryConfig.tsx:28-29` | `AdxDataSourceOptions.dataConsistency: string`, `src/types/index.ts:118`; backend `DatasourceSettings.DataConsistency`, `settings.go:21` | Written verbatim into `queryconsistency` connection property |
| `jsonData_defaultEditorMode` | `defaultEditorMode` | jsonData | `QueryConfig.tsx:122` (`Default editor mode`) | Options `QueryConfig.tsx:21-24`; default `visual` from `QueryConfig.tsx:31-32` | `AdxDataSourceOptions.defaultEditorMode: EditorMode`, `src/types/index.ts:116` | Editor exposes only `visual` and `raw`; `openai` is a valid `EditorMode` for querying but not offered as a default |
| `jsonData_defaultDatabase` | `defaultDatabase` | jsonData | `DatabaseConfig.tsx:144` (`Default database`) | Options loaded via `refreshSchema` (`DatabaseConfig.tsx:91-107`); no schema default | `AdxDataSourceOptions.defaultDatabase: string`, `src/types/index.ts:114`; backend `DatasourceSettings.DefaultDatabase`, `settings.go:20` | `ui.allowCustom: true` since options are dynamic |
| `jsonData_useSchemaMapping` | `useSchemaMapping` | jsonData | `DatabaseConfig.tsx:159` (`Use managed schema`) | Description `DatabaseConfig.tsx:160-163` | `AdxDataSourceOptions.useSchemaMapping: boolean`, `src/types/index.ts:121` | Gates the schema mappings list |
| `jsonData_schemaMappings` | `schemaMappings` | jsonData | `DatabaseConfig.tsx:173` (`Schema mappings`) | Target select + Name input pairs (`DatabaseConfig.tsx:176-217`) | `AdxDataSourceOptions.schemaMappings?: Array<Partial<SchemaMapping>>`, `src/types/index.ts:122` | Frontend-only shape today — the backend uses schema mappings via the query path, not the datasource settings |
| `jsonData_application` | `application` | jsonData | `ApplicationConfig.tsx:18` (`Application name (Optional)`) | `ApplicationConfig.tsx:30` (`Grafana-ADX`) | `AdxDataSourceOptions.application: string`, `src/types/index.ts:125`; backend `DatasourceSettings.Application`, `settings.go:25` | Passed as `x-ms-app` in Kusto requests |
| `jsonData_enableUserTracking` | `enableUserTracking` | jsonData | `TrackingConfig.tsx:19` (`Send username header to host`) | Description `TrackingConfig.tsx:22-30` | `AdxDataSourceOptions.enableUserTracking: boolean`, `src/types/index.ts:123`; backend `DatasourceSettings.EnableUserTracking`, `settings.go:24` | — |
| `jsonData_keepCookies` | `keepCookies` | jsonData | `ConfigEditor/index.tsx:148` (`Allowed cookies`) | `ConfigEditor/index.tsx:160` (`New cookie (hit enter to add)`) | `AdxDataSourceOptions.keepCookies?: string[]`, `src/types/index.ts:130` | Consumed by Grafana's shared HTTP client |
| `secureJsonData_OpenAIAPIKey` | `OpenAIAPIKey` | secureJsonData | — (no editor UI) | — | `AdxDataSourceSecureOptions.OpenAIAPIKey?: string`, `src/types/index.ts:134` | Backend-only; consumed by `pkg/azuredx/resource_handler.go` |
| `jsonData_minimalCache` | `minimalCache` | jsonData | — (declared but never rendered) | — | `AdxDataSourceOptions.minimalCache: number`, `src/types/index.ts:115` | Dead field — no editor writes it, no backend reads it |
| `jsonData_azureCloud` / `jsonData_onBehalfOf` / `jsonData_tenantId` / `jsonData_clientId` | same | jsonData | — (legacy; cleared on save) | — | `AzureDataSourceJsonData.{cloudName,azureAuthType,tenantId,clientId}?: string` in `azure-sdk/src/settings.ts:11-14` plus ADX-specific `onBehalfOf: boolean` (`src/types/index.ts:129`) | Legacy fallback — `pkg/azuredx/adxauth/adxcredentials/builder.go:40-87` + `AzureCredentialsConfig.ts:46-67` |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_clusterUrl` | `clusterUrl` | jsonData | Default cluster URL (Optional) | Yes (`DatasourceSettings.ClusterURL`) |
| `jsonData_azureCredentials` | `azureCredentials` | jsonData | Authentication Method | Yes (via `azcredentials.FromDatasourceData`) |
| `secureJsonData_azureClientSecret` | `azureClientSecret` | secureJsonData | Client Secret | Yes |
| `secureJsonData_clientSecret` | `clientSecret` | secureJsonData | — (legacy) | Yes (fallback for `azureClientSecret`) |
| `jsonData_oauthPassThru` | `oauthPassThru` | jsonData | — (SDK-managed) | Yes (backend hard-requires it for OBO) |
| `jsonData_queryTimeout` | `queryTimeout` | jsonData | Query timeout | Yes (`DatasourceSettings.QueryTimeoutRaw`) |
| `jsonData_dynamicCaching` | `dynamicCaching` | jsonData | Use dynamic caching | Yes (`DatasourceSettings.DynamicCaching`) |
| `jsonData_cacheMaxAge` | `cacheMaxAge` | jsonData | Cache max age | Yes (`DatasourceSettings.CacheMaxAge`) |
| `jsonData_dataConsistency` | `dataConsistency` | jsonData | Data consistency | Yes (`DatasourceSettings.DataConsistency`) |
| `jsonData_defaultEditorMode` | `defaultEditorMode` | jsonData | Default editor mode | **No — frontend-only** |
| `jsonData_defaultDatabase` | `defaultDatabase` | jsonData | Default database | Yes (`DatasourceSettings.DefaultDatabase`) |
| `jsonData_useSchemaMapping` | `useSchemaMapping` | jsonData | Use managed schema | **No — frontend-only** |
| `jsonData_schemaMappings` | `schemaMappings` | jsonData | Schema mappings | **No — frontend-only** (schema mappings are applied client-side in the query path) |
| `jsonData_application` | `application` | jsonData | Application name (Optional) | Yes (`DatasourceSettings.Application`) |
| `jsonData_enableUserTracking` | `enableUserTracking` | jsonData | Send username header to host | Yes (`DatasourceSettings.EnableUserTracking`) |
| `jsonData_keepCookies` | `keepCookies` | jsonData | Allowed cookies | Consumed by SDK HTTP client |
| `secureJsonData_OpenAIAPIKey` | `OpenAIAPIKey` | secureJsonData | — (provisioning-only) | **Yes — backend-only** |
| `jsonData_minimalCache` | `minimalCache` | jsonData | — | **No — dead field** |
| `jsonData_azureCloud` | `azureCloud` | jsonData | — (legacy fallback) | Yes (legacy path) |
| `jsonData_onBehalfOf` | `onBehalfOf` | jsonData | — (legacy fallback) | Yes (legacy path) |
| `jsonData_tenantId` | `tenantId` | jsonData | — (legacy fallback) | Yes (legacy path) |
| `jsonData_clientId` | `clientId` | jsonData | — (legacy fallback) | Yes (legacy path) |

### Frontend-only settings

- **`defaultEditorMode`** — declared on `AdxDataSourceOptions`
  (`src/types/index.ts:116`) and written by `QueryConfig`, but no Go file
  under `pkg/` reads it. The value only drives the editor's initial mode.
- **`useSchemaMapping` / `schemaMappings`** — the frontend writes them
  through `DatabaseConfig`, but the backend `DatasourceSettings` never
  reads them; schema mappings are applied client-side when building the
  KQL request.
- **`minimalCache`** — declared on `AdxDataSourceOptions`
  (`src/types/index.ts:115`) but no editor writes it and no backend reads
  it. Dead field.

### Backend-only settings

- **`OpenAIAPIKey`** — consumed by
  `pkg/azuredx/resource_handler.go:55-89` for the `askOpenAI` endpoint.
  There is no config editor UI to set it; provisioning YAML or the
  datasource HTTP API is required.

## Where the types are defined

The ADX configuration types are spread across the plugin and its
dependencies — the entire authentication surface comes from
`@grafana/azure-sdk` and `grafana-azure-sdk-go`, not from the plugin
itself.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `AdxDataSourceOptions` (extends `AzureDataSourceJsonData`), `AdxDataSourceSecureOptions` (extends `AzureDataSourceSecureJsonData`) | `src/types/index.ts:113-135` | plugin (`grafana/azure-data-explorer-datasource`) |
| `EditorMode`, `AdxQueryType`, `SchemaMapping`, `SchemaMappingType` | `src/types/index.ts:50-111` | plugin |
| `AzureDataSourceJsonData` (base — `azureCredentials`, `oauthPassThru`, legacy `cloudName`/`azureAuthType`/`tenantId`/`clientId`), `AzureDataSourceSecureJsonData` (base — `azureClientSecret`/`clientSecret`/...) | `src/settings.ts:5-28` | `@grafana/azure-sdk` `0.1.0` (`grafana/grafana-azure-sdk-react`) |
| `AzureAuthType`, `AzureCredentials` discriminated union, `ConcealedSecret` symbol | `src/credentials/AzureCredentials.ts:1-91` | `@grafana/azure-sdk` `0.1.0` |
| `updateDatasourceCredentials` (writes `azureCredentials` + secrets + `oauthPassThru` side-effect; clears legacy top-level fields), `getDatasourceCredentials`, `getClientSecret`, `isCredentialsComplete` | `src/credentials/AzureCredentialsConfig.ts` | `@grafana/azure-sdk` `0.1.0` |
| `getAzureClouds`, `getDefaultAzureCloud`, `resolveLegacyCloudName` | `src/clouds.ts:1-48` | `@grafana/azure-sdk` `0.1.0` |
| `DataSourceDescription`, `ConfigSection`, `ConfigSubSection` | `src/components/ConfigEditor/` | `@grafana/plugin-ui` `0.13.1` |
| `TagsInput` (renders `jsonData.keepCookies` inline) | `packages/grafana-ui/src/components/TagsInput` | `@grafana/ui` `12.4.2` |
| `Switch`, `SecureSocksProxySettings` (writes `jsonData.enableSecureSocksProxy` — deliberately excluded from this entry) | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | `@grafana/ui` `12.4.2` |
| `DataSourcePluginOptionsEditorProps`, `DataSourceSettings`, `DataSourceInstanceSettings` | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `DatasourceSettings` (ClusterURL, DefaultDatabase, DataConsistency, CacheMaxAge, DynamicCaching, EnableUserTracking, Application, QueryTimeoutRaw, QueryTimeout, ServerTimeoutValue, OpenAIAPIKey, enforce/allow/user trusted-endpoints env-controlled fields) | `pkg/azuredx/models/settings.go:18-42` | plugin (`grafana/azure-data-explorer-datasource`) |
| `adxcredentials.FromDatasourceData` (modern-then-legacy), `getFromLegacy`, `ensureOnBehalfOfSupported`, `resolveLegacyCloudName` | `pkg/azuredx/adxauth/adxcredentials/builder.go:12-121` | plugin |
| Backend `AzureCredentials` interface + concrete types (`AzureManagedIdentityCredentials`, `AzureWorkloadIdentityCredentials`, `AadCurrentUserCredentials`, `AzureClientSecretCredentials`, `AzureClientSecretOboCredentials`, `AzureClientCertificateCredentials`, `AzureEntraPasswordCredentials`) and `AzureAuthType` constants | `azcredentials/credentials.go` | `github.com/grafana/grafana-azure-sdk-go/v2` `v2.4.1` |
| `azcredentials.FromDatasourceData` — the switch on `authType` that enforces per-authType secret requirements our `Validate` mirrors | `azcredentials/builder.go` | `github.com/grafana/grafana-azure-sdk-go/v2` `v2.4.1` |
| `azsettings.AzureSettings`, cloud resolution | `azsettings/` | `github.com/grafana/grafana-azure-sdk-go/v2` `v2.4.1` |
| Backend HTTP auth wiring (`azhttpclient.AddAzureAuthentication`) | `azhttpclient/` | `github.com/grafana/grafana-azure-sdk-go/v2` `v2.4.1` |
| `backend.DataSourceInstanceSettings` (`JSONData`, `DecryptedSecureJSONData`; root fields `URL` / `User` / `BasicAuthEnabled` — unused by this plugin) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |

The `Config` type in this entry flattens the plugin's own
`DatasourceSettings` plus every jsonData field the editor / SDK writes
into a single Go struct (jsonData fields + `DecryptedSecureJSONData`) —
that lets the conformance suite compare the schema's jsonData
keys/types against the settings model directly. `AzureCredentials` is
carried opaquely as `json.RawMessage` (declared `valueType: any`)
because its shape is a discriminated union owned by `@grafana/azure-sdk`
and the plugin's backend delegates its parse to
`azcredentials.FromDatasourceData`.

## Modeling decisions

- **`azureCredentials` is `valueType: any` (opaque)**. `@grafana/azure-sdk`
  owns the discriminated-union shape; the plugin's backend delegates the
  parse to `grafana-azure-sdk-go`'s `azcredentials.FromDatasourceData`.
  Modeling each nested field individually would double-source-of-truth
  against the SDK; instead we keep the raw object on the schema (with
  `help` markdown documenting the union shape) and match it in Go as
  `json.RawMessage`. Same choice as `grafana-azure-monitor-datasource`
  and `mssql`.
- **`schemaMappings` is `array` of `object` with a strict inner
  `fields` schema** — the item shape is fully known upstream
  (`src/types/index.ts:90-96`), and the object is data-only (no editor
  reads it back into a form). Using a strict item schema keeps
  conformance tight.
- **`RootConfig` is `Record<string, never>`** — the ADX backend never
  reads `settings.URL` / `settings.User` / `settings.BasicAuthEnabled`
  from `backend.DataSourceInstanceSettings`. All URLs come from
  `jsonData.clusterUrl` + `azureCloud`; auth comes from
  `jsonData.azureCredentials` and its secrets.
- **`ApplyDefaults` sets `dataConsistency` and `defaultEditorMode`**.
  These are the two `useEffect`-driven writes `QueryConfig` performs
  when the editor first mounts (`QueryConfig.tsx:27-34`). Credential
  defaults are deliberately NOT applied here because the editor's
  choice is Grafana-instance-dependent
  (`AzureCredentialsConfig.ts:22-34`).
- **`Validate` mirrors the SDK's per-authType checks plus the plugin's
  OBO-specific check** — the source of truth for "which secret is
  required" is `azcredentials.FromDatasourceData`; for OBO the plugin
  adds `ensureOnBehalfOfSupported`
  (`pkg/azuredx/adxauth/adxcredentials/builder.go:89-98`). We also
  reject `queryTimeout > 1h` (the `formatTimeout` cap) and unknown
  `dataConsistency` values so provisioning catches these early.
- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy` — rendered
  by `ConfigEditor/index.tsx:167-210`) is **deliberately excluded**
  per AGENTS.md.
- **Field ID naming**: IDs use the `<target>_<camelCaseKey>` convention;
  the `key` property keeps the plugin's raw storage key.
- **Legacy fields are kept alongside modern ones**, tagged `legacy` —
  the backend still reads them for pre-migration datasources
  (`pkg/azuredx/adxauth/adxcredentials/builder.go:40-87`). Removing them
  from the schema would misrepresent what the backend accepts.
  `@grafana/azure-sdk`'s `updateDatasourceCredentials` clears them
  whenever the editor saves.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go`
`pluginschema.PluginSchema` bundle (the k8s-style schema Grafana's
datasource API server serves as `{apiVersion}.json`, `v0alpha1` today)
from the embedded `dsconfig.json`.

`SettingsExamples()` covers the default configuration plus one example
per authentication path ADX supports:

| Example | AuthType | Config source | Secrets populated |
| --- | --- | --- | --- |
| `""` (default) | — (empty; user must complete auth) | schema defaults + ApplyDefaults | `azureClientSecret` (empty placeholder) |
| `clientSecret` | `clientsecret` | Config editor | `azureClientSecret` |
| `clientSecretOBO` | `clientsecret-obo` | Config editor (feature-toggled) | `azureClientSecret` |
| `managedIdentity` | `msi` | Config editor (when enabled) | — |
| `workloadIdentity` | `workloadidentity` | Config editor (when enabled) | — |
| `currentUser` | `currentuser` | Config editor (when enabled) | — |
| `openAI` | `clientsecret` + OpenAI API key | Config editor + provisioning-only secret | `azureClientSecret` + `OpenAIAPIKey` |
| `legacyClientSecret` | Legacy top-level tenant/client + secret | Pre-migration provisioning | `clientSecret` (legacy) |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx, settings)` runs the full three-phase load flow:

1. **Parse** — unmarshal `settings.JSONData` into `Config` (jsonData
   fields become struct fields; `azureCredentials` is captured as
   `json.RawMessage`), parse `queryTimeout` as a Go duration
   (empty → 30s default), copy decrypted secrets into
   `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — set `dataConsistency = strongconsistency` and
   `defaultEditorMode = visual` when zero.
3. **`Validate`** — mirror the runtime checks the plugin's backend
   performs (per-authType secret requirements, OBO requires
   `oauthPassThru`, legacy tuple must be complete, queryTimeout ≤ 1h,
   dataConsistency must be one of the two known values). Errors are
   joined.

Everything is logged via `backend.Logger.FromContext(ctx)` with
`datasource_uid`, `datasource_name`, and `plugin` labels.

## Upstream findings

Preserved verbatim in the schema — the schema records what the plugin
**does**, not what it **should** do. These notes exist so reviewers can
reproduce each finding and decide whether to file an upstream fix.

1. **`AzureAuthType` union lists `currentuser` twice.**
   `grafana-azure-sdk-react/src/credentials/AzureCredentials.ts:1-9` —
   the shared SDK's union declares `currentuser` at both position 1 and
   position 7. Harmless (TypeScript de-duplicates unions), but reads
   like a copy-paste artifact. Inherited from the SDK; ADX cannot fix
   it locally.
2. **ADX exposes a strict subset of the SDK's auth types.**
   `AzureCredentialsForm.tsx:31-71` builds options for only
   `clientsecret`, `msi`, `workloadidentity`, `currentuser`, and
   `clientsecret-obo`. `clientcertificate` and `ad-password` are valid
   backend authTypes accepted by `azcredentials.FromDatasourceData` but
   the ADX editor does not offer them; users who provision those types
   would have a working datasource with no editable auth in the UI.
3. **`defaultEditorMode` is frontend-only.** No file under `pkg/` reads
   it. The stored value only drives the query editor's initial mode.
4. **`useSchemaMapping` / `schemaMappings` are frontend-only.** The
   backend `DatasourceSettings` (`pkg/azuredx/models/settings.go`) does
   not carry them; schema mapping happens on the frontend when it
   builds queries.
5. **`minimalCache` is dead.** Declared on `AdxDataSourceOptions`
   (`src/types/index.ts:115`) but never written by the editor and never
   read by the backend `DatasourceSettings` (`settings.go`). Historical
   leftover.
6. **`OpenAIAPIKey` has no editor UI.** The ConfigEditor only reads
   `secureJsonFields['OpenAIAPIKey']` (`ConfigEditor/index.tsx:58`) to
   decide whether the Additional settings section starts open. There is
   no input to actually set the key — it must be set via provisioning
   YAML or the datasource HTTP API.
7. **On-Behalf-Of has a hard `oauthPassThru` prerequisite.**
   `pkg/azuredx/adxauth/adxcredentials/builder.go:89-98`
   (`ensureOnBehalfOfSupported`) refuses to create the instance if
   `jsonData.oauthPassThru != true` when the resolved auth type is OBO.
   `@grafana/azure-sdk` sets `oauthPassThru = true` automatically for
   OBO, so this only bites manual provisioning that forgets the flag.
8. **`getLegacyCredentials` silently ignores partial legacy tuples.**
   `pkg/azuredx/adxauth/adxcredentials/builder.go:60-62` — if any of
   `tenantId`, `clientId`, or `secureJsonData.clientSecret` is empty,
   the function returns `nil, nil` and the credentials fall through to
   `GetDefaultCredentials`. Provisioning YAML with a typo'd tenant ID
   quietly drops back to Grafana's default identity. Our `Validate`
   surfaces this as an error when partial legacy fields are present.
9. **Legacy `azureCloud` reuses Azure Monitor names.** The ADX legacy
   resolver (`builder.go:107-121`) recognises `azuremonitor` /
   `chinaazuremonitor` / `govazuremonitor`, not the ADX-native
   `AzureCloud` / `AzureChinaCloud` / `AzureUSGovernment` — the modern
   `azureCredentials.azureCloud` values are only understood by the
   modern (non-legacy) path.
10. **Backend `Load` fails hard on any `time.ParseDuration` error.**
    `pkg/azuredx/models/settings.go:68-70` — a `queryTimeout` value the
    frontend accepts as a free-text input can nevertheless kill instance
    creation. Suggest surfacing the parse in the editor.
11. **Backend `formatTimeout` caps `queryTimeout` at one hour.**
    `pkg/azuredx/models/settings.go:101-107` — timeouts >1h fail with
    "timeout must be one hour or less". Not documented in the editor's
    description; users who paste `2h` from other datasources will hit
    it. Our `Validate` surfaces it too.
12. **Editor `useEffect` in `QueryConfig` writes defaults on every
    mount.** `QueryConfig.tsx:27-34` — the effect runs unconditionally
    when `dataConsistency` or `defaultEditorMode` is falsy. This has the
    intended effect of seeding defaults but also causes a "dirty" state
    on first mount of legacy datasources, prompting users to save
    unrelated no-op changes.
13. **`Secure Socks Proxy` conditional render depends on a semver
    check.** `ConfigEditor/index.tsx:168-169` requires
    `gte(config.buildInfo.version, '10.0.0')` — datasources provisioned
    against older Grafana versions never see the toggle. The field is
    deliberately excluded from this entry per AGENTS.md.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` — passes.
- JSON Schema validation against
  [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft 2020-12,
  strict — `additionalProperties: false`) — passes via the conformance
  suite.
- `go build ./... && go vet ./... && gofmt -l . && go test ./...` inside
  `registry/` — passes (schema bundle shape, secure values, examples,
  `LoadConfig`, `Validate`, `EffectiveAuthType`, `ApplyDefaults`,
  `SchemaArtifactInSync` guard).
- `tsc --noEmit --strict` on `settings.ts` — passes.
- `dsconfig` and `schema` workspace modules still build.
