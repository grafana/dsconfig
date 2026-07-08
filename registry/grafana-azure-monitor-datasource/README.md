# grafana-azure-monitor-datasource

Declarative configuration schema for the
[Azure Monitor datasource plugin](https://github.com/grafana/grafana-azure-monitor-datasource)
(`grafana-azure-monitor-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/grafana-azure-monitor-datasource`
- **Ref**: `main`
- **Commit SHA**: `87f88ede8122295a7b671420460535d75a4c02bf` (`chore: remove monorepo package.json artifact and sync deps to root (#84)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders,
tooltips, option labels/values, help markdown, defaults, storage keys,
storage targets, value types, group titles, and instructions — is
traceable to a specific `file:line` in the upstream repo (or a pinned
external dependency) at this SHA. See [Field provenance](#field-provenance)
below.

To reproduce this research:

```bash
git clone https://github.com/grafana/grafana-azure-monitor-datasource
cd grafana-azure-monitor-datasource
git checkout 87f88ede8122295a7b671420460535d75a4c02bf
```

If upstream `main` has advanced past this SHA, re-diff the sources listed
under [Sources researched](#sources-researched) before merging any changes
to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and LLM instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig` (empty), `JsonDataConfig` (jsonData shape), `SecureJsonDataConfig` (secure key list), plus `AzureCredentials` union |
| [`settings.go`](settings.go) | Go `Config` (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `AuthType` / `CertificateFormat` / `LegacyCloudName` / `SecureJsonDataKey` typed constants, and the `LoadConfig` utility with `ApplyDefaults` + `Validate` |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, exposes `NewSchema()` via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `Validate`, `EffectiveAuthType`, `ApplyDefaults` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...`; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of
the shared [`registry/`](..) module
(`github.com/grafana/dsconfig/registry`).

## Sources researched

### Plugin repo (`github.com/grafana/grafana-azure-monitor-datasource@87f88ed`)

| File | What was read |
| --- | --- |
| `src/plugin.json:4,146-162` | `pluginType` (`id`), `pluginName` (`name`), `docURL` (first `info.links[]` — Learn more) |
| `src/components/ConfigEditor/ConfigEditor.tsx:36-127` | Editor shell — renders `DataSourceDescription`, `MonitorConfig`, and the collapsible **Additional settings** section (`AdvancedHttpSettings` + optional `SecureSocksProxySettings`) |
| `src/components/ConfigEditor/MonitorConfig.tsx:21-77` | Composition of `AzureCredentialsForm`, `DefaultSubscription`, `BasicLogsToggle`, and the `useEffectOnce` that seeds credentials on first load |
| `src/components/ConfigEditor/AzureCredentialsForm.tsx:44-149` | Authentication type select, per-Grafana-config visibility of `msi` / `workloadidentity` / `currentuser`, and the sub-panel dispatch |
| `src/components/ConfigEditor/AppRegistrationCredentials.tsx:52-370` | Azure Cloud select, Tenant / Client ID fields, Client Secret input, Certificate Format select, Client Certificate / Private Key / Certificate Password textareas — plus their placeholders (`XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX`, `-----BEGIN CERTIFICATE-----`, `-----BEGIN PRIVATE KEY-----`) |
| `src/components/ConfigEditor/CurrentUserFallbackCredentials.tsx:24-200` | Fallback Service Credentials section, `serviceCredentialsEnabled` radio, fallback auth types |
| `src/components/ConfigEditor/DefaultSubscription.tsx:21-101` | Default Subscription select + Load Subscriptions button (writes `jsonData.subscriptionId`) |
| `src/components/ConfigEditor/BasicLogsToggle.tsx:14-61` | Enable Basic Logs switch (writes `jsonData.basicLogsEnabled`; description string preserved verbatim minus the anchor link) |
| `src/types/types.ts:30-53` | `AzureMonitorDataSourceJsonData` — the jsonData shape (subscriptionId, basicLogsEnabled, deprecated log-analytics fields, deprecated appInsightsAppId, timeout, keepCookies, enableSecureSocksProxy) |
| `src/types/types.ts:55-57` | `AzureMonitorDataSourceSecureJsonData` — adds `appInsightsApiKey` on top of the azure-sdk secret keys |
| `src/credentials.ts:13-74` | `getCredentials` / `updateCredentials` / `getLegacyCredentials` / `getDefaultCredentials` — the frontend legacy fallback and Grafana-config-dependent defaults |
| `pkg/plugin.json` — plugin.json backend section | Backend enabled (`executable: gpx_grafana-azure-monitor-datasource`) |
| `pkg/azuremonitor/azuremonitor.go:145-201` | `NewInstanceSettings` — parses jsonData twice (once as a generic map for the credentials builder, once into `AzureMonitorSettings`), fails hard on `currentuser` without the feature toggle |
| `pkg/azuremonitor/types/types.go:22-58` | `AzRoute`, `AzureMonitorSettings` (subscriptionId, logAnalyticsDefaultWorkspace, appInsightsAppId), `AzureMonitorCustomizedCloudSettings` (customizedRoutes), `DatasourceInfo` |
| `pkg/azuremonitor/azmoncredentials/builder.go:13-142` | `FromDatasourceData` — modern-then-legacy credential parse, `getFromLegacy`, `resolveLegacyCloudName` mapping (`azuremonitor` → `AzureCloud`, `chinaazuremonitor` → `AzureChinaCloud`, `govazuremonitor` → `AzureUSGovernment`, `customizedazuremonitor` → `AzureCustomized`) |
| `pkg/azuremonitor/azmoncredentials/default.go:8-15` | `GetDefaultCredentials` — backend default credentials mirror the frontend `getDefaultCredentials` (managedIdentity → workloadIdentity → clientSecret) |
| `pkg/azuremonitor/routes.go:25-106` | Cloud → routes resolution; `AzureCustomized` requires `jsonData.customizedRoutes` |
| `pkg/azuremonitor/loganalytics/azure-log-analytics-datasource.go:328,415-432` | `basicLogsEnabled` gate and the `azureLogAnalyticsSameAs=false` hard-fail |
| `pkg/azuremonitor/httpclient.go:25-50` | Auth wiring — `AzureAuthentication` middleware built from `model.Credentials` |
| `go.mod:8` | Pinned Azure SDK version: `grafana-azure-sdk-go/v2 v2.4.1` |
| `package.json:76,79` | Pinned frontend SDK versions: `@grafana/azure-sdk 0.1.0`, `@grafana/plugin-ui 0.13.1` |

### External components (read at the exact pinned versions)

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `AzureAuthType`, `AzureCredentials` union | `@grafana/azure-sdk` `0.1.0` | `grafana/grafana-azure-sdk-react` `src/credentials/AzureCredentials.ts:1-91` | The seven `authType` values, the per-authType field shape, `certificateFormat` enum, `ConcealedSecret` symbol |
| `AzureCredentialsConfig` (`updateDatasourceCredentials`, `getDatasourceCredentials`, `getClientSecret`, `getAdPassword`, `getClientCertificate`, `getPrivateKey`, `getCertificatePassword`, `isCredentialsComplete`) | `@grafana/azure-sdk` `0.1.0` | `grafana/grafana-azure-sdk-react` `src/credentials/AzureCredentialsConfig.ts:14-453` | Which storage keys the editor writes per authType (in particular the `oauthPassThru` / `disableGrafanaCache` side-effects and the clearing of legacy top-level fields), and how it decides "already-configured" via `secureJsonFields` |
| `AzureDataSourceJsonData`, `AzureDataSourceSecureJsonData` | `@grafana/azure-sdk` `0.1.0` | `grafana/grafana-azure-sdk-react` `src/settings.ts:5-28` | Base jsonData / secureJsonData shape Azure Monitor extends |
| `getAzureClouds` / `resolveLegacyCloudName` | `@grafana/azure-sdk` `0.1.0` | `grafana/grafana-azure-sdk-react` `src/clouds.ts:1-48` | Which cloud identifiers the SDK renders (Azure / Azure China / Azure US Government — plus custom clouds injected via Grafana config); mapping from legacy `cloudName` values |
| `AdvancedHttpSettings` | `@grafana/plugin-ui` `0.13.1` | `grafana/plugin-ui` `src/components/ConfigEditor/AdvancedSettings/AdvancedHttpSettings.tsx:13-82` | Storage keys it writes (`jsonData.timeout`, `jsonData.keepCookies`) and their labels (`Timeout`, `Allowed cookies`) |
| `DataSourceDescription`, `ConfigSection` | `@grafana/plugin-ui` `0.13.1` | `grafana/plugin-ui` `src/components/ConfigEditor/` | Intro layout and section titles — no storage keys |
| `azcredentials.FromDatasourceData` (backend) | `grafana-azure-sdk-go/v2` `v2.4.1` | `grafana/grafana-azure-sdk-go` `azcredentials/builder.go` | Per-authType parse and secret requirements — the checks `Validate()` mirrors |
| `azcredentials.AzureAuthType` constants (backend) | `grafana-azure-sdk-go/v2` `v2.4.1` | `grafana/grafana-azure-sdk-go` `azcredentials/credentials.go:3-11` | The seven backend-recognised authType constants |

## Field provenance

| Schema `id` | Storage key | Target | Label source | Placeholder / default / options source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_azureCredentials` | `azureCredentials` | `jsonData` | `AzureCredentialsForm.tsx:110` (`<Field label="Authentication type">`) | Options `AzureCredentialsForm.tsx:44-78`; no default (Grafana-config-dependent — `src/credentials.ts:66-74`) | Union `AzureCredentials`, `AzureCredentials.ts:74-81`; modeled as `valueType: any` (opaque, see [Modeling decisions](#modeling-decisions)) | Role `auth.discriminator`; `description` mirrors `AzureCredentialsForm.tsx:112-114` verbatim |
| `jsonData_subscriptionId` | `subscriptionId` | `jsonData` | `DefaultSubscription.tsx:73` (`<Field label="Default Subscription">`) | Options loaded from `/subscriptions?api-version=2019-03-01` (`ConfigEditor.tsx:65-72`); no schema default | `AzureMonitorDataSourceJsonData.subscriptionId?: string`, `types.ts:32`; backend `AzureMonitorSettings.SubscriptionId`, `pkg/azuremonitor/types/types.go:29` | `ui.allowCustom: true` since options are dynamic (loaded from Azure) |
| `jsonData_basicLogsEnabled` | `basicLogsEnabled` | `jsonData` | `BasicLogsToggle.tsx:50` (`<Field label="Enable Basic Logs">`) | Description text `BasicLogsToggle.tsx:35-43` (link text preserved as inline text since we can't render an anchor here); default `false` from `BasicLogsToggle.tsx:56` (`options.basicLogsEnabled ?? false`) | `AzureMonitorDataSourceJsonData.basicLogsEnabled?: boolean`, `types.ts:33` | Read by backend at `loganalytics/azure-log-analytics-datasource.go:328` |
| `jsonData_timeout` | `timeout` | `jsonData` | `AdvancedHttpSettings.tsx:64` (`label="Timeout"`) | `AdvancedHttpSettings.tsx:74` (`placeholder="Timeout in seconds"`) | `AzureMonitorDataSourceJsonData.timeout?: number`, `types.ts:51` | Role `transport.timeoutSeconds` |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | `AdvancedHttpSettings.tsx:48` (`label="Allowed cookies"`) | `AdvancedHttpSettings.tsx:56` (`placeholder="New cookie (hit enter to add)"`) | `AzureMonitorDataSourceJsonData.keepCookies?: string[]`, `types.ts:52` | `ui.component: list` (mirrors the TagsInput UX) |
| `secureJsonData_azureClientSecret` | `azureClientSecret` | `secureJsonData` | `AppRegistrationCredentials.tsx:182` (`label="Client Secret"`) | `AppRegistrationCredentials.tsx:194` (`placeholder="XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"`) | `AzureDataSourceSecureJsonData.azureClientSecret?: string`, `azure-sdk/src/settings.ts:18` | Managed by `@grafana/azure-sdk` — see `AzureCredentialsConfig.ts:302-311` |
| `secureJsonData_clientCertificate` | `clientCertificate` | `secureJsonData` | `AppRegistrationCredentials.tsx:296` (`label="Client Certificate"`) | `AppRegistrationCredentials.tsx:306` (`placeholder="-----BEGIN CERTIFICATE-----"`) | `AzureDataSourceSecureJsonData.clientCertificate?: string`, `azure-sdk/src/settings.ts:22` | — |
| `secureJsonData_privateKey` | `privateKey` | `secureJsonData` | `AppRegistrationCredentials.tsx:324` (`label="Private Key"`) | `AppRegistrationCredentials.tsx:334` (`placeholder="-----BEGIN PRIVATE KEY-----"`) | `AzureDataSourceSecureJsonData.privateKey?: string`, `azure-sdk/src/settings.ts:23` | Only used when `certificateFormat == 'pem'` |
| `secureJsonData_certificatePassword` | `certificatePassword` | `secureJsonData` | `AppRegistrationCredentials.tsx:349` (`label-symbol-private-key-password: "Certificate Password"`) | `AppRegistrationCredentials.tsx:355` (`placeholder="XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX"`) | `AzureDataSourceSecureJsonData.certificatePassword?: string`, `azure-sdk/src/settings.ts:24` | Only used when `certificateFormat == 'pfx'` |
| `secureJsonData_clientSecret` | `clientSecret` | `secureJsonData` | — (no UI; legacy fallback) | — | `AzureDataSourceSecureJsonData.clientSecret?: string`, `azure-sdk/src/settings.ts:27` | Preserved by `AzureCredentialsConfig.ts:311` as `concealedLegacy` for read-back |
| `secureJsonData_password` | `password` | `secureJsonData` | — (no editor UI today) | — | `AzureDataSourceSecureJsonData.password?: string`, `azure-sdk/src/settings.ts:19` | Consumed by backend for `authType == 'ad-password'` (`grafana-azure-sdk-go` `azcredentials/builder.go` case `ad-password`) |
| `secureJsonData_appInsightsApiKey` | `appInsightsApiKey` | `secureJsonData` | — (no editor UI) | — | `AzureMonitorDataSourceSecureJsonData.appInsightsApiKey?: string`, `types.ts:56` | Deprecated App Insights key preserved on migrated datasources |
| `jsonData_appInsightsAppId` | `appInsightsAppId` | `jsonData` | — (deprecated; no UI) | — | `AzureMonitorSettings.AppInsightsAppId`, `pkg/azuremonitor/types/types.go:31` | Backend-only |
| `jsonData_logAnalyticsDefaultWorkspace` | `logAnalyticsDefaultWorkspace` | `jsonData` | — (deprecated; no UI) | — | `AzureMonitorSettings.LogAnalyticsDefaultWorkspace`, `pkg/azuremonitor/types/types.go:30` | Backend-only |
| `jsonData_azureLogAnalyticsSameAs` | `azureLogAnalyticsSameAs` | `jsonData` | — (deprecated; no UI) | — | `AzureMonitorDataSourceJsonData.azureLogAnalyticsSameAs?: boolean`, `types.ts:37`; backend accepts bool OR bool-parsable string at `loganalytics/azure-log-analytics-datasource.go:415-427` — hence `valueType: any` | Backend-only; hard-fails Log Analytics if not-truthy |
| `jsonData_logAnalyticsTenantId` / `jsonData_logAnalyticsClientId` / `jsonData_logAnalyticsSubscriptionId` | same | `jsonData` | — (deprecated; no UI) | — | `types.ts:39-43` | Frontend-only (no backend reader) |
| `jsonData_azureAuthType` / `jsonData_cloudName` / `jsonData_tenantId` / `jsonData_clientId` | same | `jsonData` | — (legacy; cleared on save) | — | `AzureDataSourceJsonData.{azureAuthType,cloudName,tenantId,clientId}?: string` in `azure-sdk/src/settings.ts:11-14` | Legacy fallback path only — `azmoncredentials/builder.go:33-116` + `src/credentials.ts:40-57` |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (no UI; SDK-managed) | — | `AzureDataSourceJsonData.oauthPassThru?: boolean`, `azure-sdk/src/settings.ts:8` | Role `auth.forwardOAuthToken.enabled`; set to `true` for `clientsecret-obo` (`AzureCredentialsConfig.ts:320-322`) and `currentuser` (`:423`) |
| `jsonData_disableGrafanaCache` | `disableGrafanaCache` | `jsonData` | — (no UI; SDK-managed) | — | Extension by `AzureCredentialsConfig.ts:424` | Set to `true` only for `currentuser` |
| `jsonData_customizedRoutes` | `customizedRoutes` | `jsonData` | — (provisioning-only) | — | Backend `AzureMonitorCustomizedCloudSettings.CustomizedRoutes` `map[string]AzRoute`, `pkg/azuremonitor/types/types.go:36`; here modeled as `valueType: map` with `item.valueType: any` | Backend consults it only when `azureCloud == 'AzureCustomizedCloud'` |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_azureCredentials` | `azureCredentials` | jsonData | Authentication type | Yes (via `azcredentials.FromDatasourceData`) |
| `jsonData_subscriptionId` | `subscriptionId` | jsonData | Default Subscription | Yes (`AzureMonitorSettings`) |
| `jsonData_basicLogsEnabled` | `basicLogsEnabled` | jsonData | Enable Basic Logs | Yes (`loganalytics/…:328`) |
| `jsonData_timeout` | `timeout` | jsonData | Timeout | Consumed by SDK HTTP client |
| `jsonData_keepCookies` | `keepCookies` | jsonData | Allowed cookies | Consumed by SDK HTTP client |
| `jsonData_oauthPassThru` | `oauthPassThru` | jsonData | — (SDK-managed) | Consumed by SDK HTTP auth |
| `jsonData_disableGrafanaCache` | `disableGrafanaCache` | jsonData | — (SDK-managed) | Consumed by Grafana caching layer |
| `jsonData_customizedRoutes` | `customizedRoutes` | jsonData | — (provisioning-only) | **Yes — backend-only** |
| `jsonData_appInsightsAppId` | `appInsightsAppId` | jsonData | — (deprecated) | **Yes — backend-only** |
| `jsonData_logAnalyticsDefaultWorkspace` | `logAnalyticsDefaultWorkspace` | jsonData | — (deprecated) | **Yes — backend-only** |
| `jsonData_azureLogAnalyticsSameAs` | `azureLogAnalyticsSameAs` | jsonData | — (deprecated) | **Yes — backend-only (hard-fails when not truthy)** |
| `jsonData_logAnalyticsTenantId` | `logAnalyticsTenantId` | jsonData | — (deprecated) | **No — frontend-only** |
| `jsonData_logAnalyticsClientId` | `logAnalyticsClientId` | jsonData | — (deprecated) | **No — frontend-only** |
| `jsonData_logAnalyticsSubscriptionId` | `logAnalyticsSubscriptionId` | jsonData | — (deprecated) | **No — frontend-only** |
| `jsonData_azureAuthType` | `azureAuthType` | jsonData | — (legacy fallback) | Yes (legacy `azmoncredentials/builder.go:33-116`) |
| `jsonData_cloudName` | `cloudName` | jsonData | — (legacy fallback) | Yes (legacy path + routes gate) |
| `jsonData_tenantId` | `tenantId` | jsonData | — (legacy fallback) | Yes (legacy path) |
| `jsonData_clientId` | `clientId` | jsonData | — (legacy fallback) | Yes (legacy path) |
| `secureJsonData_azureClientSecret` | `azureClientSecret` | secureJsonData | Client Secret | Yes |
| `secureJsonData_clientCertificate` | `clientCertificate` | secureJsonData | Client Certificate | Yes |
| `secureJsonData_privateKey` | `privateKey` | secureJsonData | Private Key | Yes |
| `secureJsonData_certificatePassword` | `certificatePassword` | secureJsonData | Certificate Password | Yes |
| `secureJsonData_password` | `password` | secureJsonData | — | Yes (backend `ad-password` only) |
| `secureJsonData_clientSecret` | `clientSecret` | secureJsonData | — (legacy) | Yes (fallback for `azureClientSecret`) |
| `secureJsonData_appInsightsApiKey` | `appInsightsApiKey` | secureJsonData | — (deprecated) | Preserved for migrated datasources; the current backend does not authenticate against App Insights directly |

### Frontend-only settings

- **`logAnalyticsTenantId`**, **`logAnalyticsClientId`**, **`logAnalyticsSubscriptionId`**
  are marked `@deprecated` in `src/types/types.ts:39-43`. The frontend types still list
  them so provisioned datasources with these fields parse cleanly, but no
  Go file in `pkg/` reads them.

### Backend-only settings

- **`appInsightsAppId`** (`AzureMonitorSettings.AppInsightsAppId`,
  `pkg/azuremonitor/types/types.go:31`) — the backend still parses it for
  migrated datasources; no editor UI writes it any more.
- **`logAnalyticsDefaultWorkspace`** (`AzureMonitorSettings.LogAnalyticsDefaultWorkspace`,
  `types.go:30`) — same story.
- **`azureLogAnalyticsSameAs`** — a deprecated gate flag; if defined AND
  not truthy the backend hard-fails Log Analytics queries
  (`azure-log-analytics-datasource.go:415-432`).
- **`customizedRoutes`** (`AzureMonitorCustomizedCloudSettings.CustomizedRoutes`,
  `types.go:36`) — provisioning-only. The config editor cannot express
  this shape; the backend consults it only when the resolved cloud is
  `AzureCustomizedCloud`.

## Where the types are defined

The Azure Monitor configuration types are spread across the plugin and
its dependencies — most fields on the editor come from
`@grafana/azure-sdk`, not from the plugin itself.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `AzureMonitorDataSourceJsonData` (extends `AzureDataSourceJsonData`), `AzureMonitorDataSourceSecureJsonData` (extends `AzureDataSourceSecureJsonData`) | `src/types/types.ts:30-57` | plugin (`grafana/grafana-azure-monitor-datasource`) |
| `AzureDataSourceJsonData` (base — `azureCredentials`, `oauthPassThru`, legacy `cloudName`/`azureAuthType`/`tenantId`/`clientId`), `AzureDataSourceSecureJsonData` (base — `azureClientSecret`/`password`/`clientCertificate`/`privateKey`/`certificatePassword`/legacy `clientSecret`) | `src/settings.ts:5-28` | `@grafana/azure-sdk` `0.1.0` (`grafana/grafana-azure-sdk-react`) |
| `AzureAuthType`, `AzureCredentials` discriminated union (msi / workloadidentity / clientsecret / clientsecret-obo / clientcertificate / currentuser / ad-password), `CertificateFormat` enum, `ConcealedSecret` symbol | `src/credentials/AzureCredentials.ts:1-91` | `@grafana/azure-sdk` `0.1.0` |
| `updateDatasourceCredentials` (writes `azureCredentials` + secrets + `oauthPassThru`/`disableGrafanaCache` side-effects; clears legacy top-level fields), `getDatasourceCredentials` (reader) | `src/credentials/AzureCredentialsConfig.ts:244-449` | `@grafana/azure-sdk` `0.1.0` |
| `getAzureClouds`, `getDefaultAzureCloud`, `resolveLegacyCloudName` | `src/clouds.ts:1-48` | `@grafana/azure-sdk` `0.1.0` |
| `AdvancedHttpSettings` (writes `jsonData.timeout` + `jsonData.keepCookies`) | `src/components/ConfigEditor/AdvancedSettings/AdvancedHttpSettings.tsx:13-82` | `@grafana/plugin-ui` `0.13.1` (`grafana/plugin-ui`) |
| `SecureSocksProxySettings` (writes `jsonData.enableSecureSocksProxy` — deliberately excluded from this entry) | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | `@grafana/ui` `13.1.0` |
| `DataSourcePluginOptionsEditorProps`, `DataSourceSettings`, `DataSourceInstanceSettings` | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `13.1.0` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `AzureMonitorSettings` (subscriptionId, logAnalyticsDefaultWorkspace, appInsightsAppId), `AzureMonitorCustomizedCloudSettings` (customizedRoutes), `AzRoute`, `DatasourceInfo` | `pkg/azuremonitor/types/types.go:22-58` | plugin (`grafana/grafana-azure-monitor-datasource`) |
| `azmoncredentials.FromDatasourceData` (modern-then-legacy), `getFromLegacy`, `resolveLegacyCloudName`, `GetDefaultCredentials` | `pkg/azuremonitor/azmoncredentials/{builder.go,default.go}` | plugin |
| Backend `AzureCredentials` interface + concrete types (`AadCurrentUserCredentials`, `AzureManagedIdentityCredentials`, `AzureWorkloadIdentityCredentials`, `AzureClientSecretCredentials`, `AzureClientCertificateCredentials`, `AzureClientSecretOboCredentials`, `AzureEntraPasswordCredentials`) and their `AzureAuthType` constants | `azcredentials/credentials.go:1-97` | `github.com/grafana/grafana-azure-sdk-go/v2` `v2.4.1` |
| `azcredentials.FromDatasourceData` (the switch on `authType` that enforces the per-authType secret requirements our `Validate` mirrors) | `azcredentials/builder.go:9-209` | `github.com/grafana/grafana-azure-sdk-go/v2` `v2.4.1` |
| `azsettings.AzureSettings`, `azcredentials.GetAzureCloud` (cloud resolution) | `azsettings/` and `azcredentials/cloud.go` | `github.com/grafana/grafana-azure-sdk-go/v2` `v2.4.1` |
| Backend HTTP auth wiring (`azhttpclient.AddAzureAuthentication`, `AllowUserIdentity`, `AddRateLimitSession`, `Scopes`) | `azhttpclient/` | `github.com/grafana/grafana-azure-sdk-go/v2` `v2.4.1` |
| `backend.DataSourceInstanceSettings` (`JSONData`, `DecryptedSecureJSONData`; root fields `URL` / `User` / `BasicAuthEnabled` — unused by this plugin) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |

The `Config` type in this entry flattens the plugin's own `Settings` /
`DatasourceInfo` split into a single Go struct (jsonData fields +
`DecryptedSecureJSONData`) — that lets the conformance suite compare the
schema's jsonData keys/types against the settings model directly.
`AzureCredentials` is carried opaquely as `json.RawMessage` (declared
`valueType: any`) because its shape is a discriminated union owned by
`@grafana/azure-sdk` and the plugin's backend delegates its parse to
`azcredentials.FromDatasourceData`.

## Modeling decisions

- **`azureCredentials` is `valueType: any` (opaque)**. `@grafana/azure-sdk`
  owns the discriminated-union shape (`AzureCredentials.ts:74-81`); the
  plugin's backend delegates its parse to
  `grafana-azure-sdk-go`'s `azcredentials.FromDatasourceData`. Modeling
  each nested field individually would double-source-of-truth against the
  SDK; instead we keep the raw object on the schema (with a descriptive
  `help` markdown documenting the union shape) and match it in Go as
  `json.RawMessage`. This mirrors the pattern used by `mssql`.
- **`customizedRoutes` is `valueType: map` with `item.valueType: any`**.
  The value shape is the backend `AzRoute` struct
  (`pkg/azuremonitor/types/types.go:22-26`); modeling each field
  individually would fight the map's dynamic keys (arbitrary route names).
- **`azureLogAnalyticsSameAs` is `valueType: any`**. The frontend type
  declares `boolean` (`types.ts:37`) but the backend explicitly parses
  bool OR bool-parsable string
  (`azure-log-analytics-datasource.go:415-427`). We surface both by using
  `any` on the schema + `json.RawMessage` on `Config`, with
  `Validate()` running the same parse the backend does.
- **`RootConfig` is `Record<string, never>`** — Azure Monitor's backend
  never reads `settings.URL` / `settings.User` / `settings.BasicAuthEnabled`
  etc. from `backend.DataSourceInstanceSettings` (all URLs are derived
  from `azureCloud` + `customizedRoutes`; auth comes from
  `jsonData.azureCredentials`). AGENTS.md requires "blank object, never
  null" in that case.
- **`ApplyDefaults` is a no-op**. Azure Monitor's editor picks its default
  credential shape at load time based on Grafana's runtime config
  (`config.azure.managedIdentityEnabled` / `workloadIdentityEnabled` — see
  `src/credentials.ts:66-74`). We can't reproduce that decision on the
  schema side without knowing the Grafana instance, so we leave
  `AzureCredentials` un-defaulted and let the backend fall back to
  `GetDefaultCredentials` (`azmoncredentials/default.go:8-15`) when the
  field is empty. The `ApplyDefaults` method is still exported for symmetry
  with `LoadConfig` and future defaults.
- **`Validate` mirrors the SDK's per-authType checks** rather than adding
  new ones — the source of truth for "which secret is required" is the
  shared `azcredentials.FromDatasourceData` in `grafana-azure-sdk-go`. Any
  drift is a bug in this entry.
- **`SecureSocksProxySettings` (`jsonData.enableSecureSocksProxy`) is
  deliberately excluded**, per AGENTS.md.
- **Field ID naming**: IDs use the `<target>_<camelCaseKey>` convention
  (`jsonData_subscriptionId`, `secureJsonData_azureClientSecret`, etc.);
  the `key` property keeps the plugin's raw storage key.
- **Legacy fields are kept alongside modern ones**, tagged `legacy` — the
  backend still reads them for pre-migration datasources
  (`azmoncredentials/builder.go:33-116`). Removing them from the schema
  would misrepresent what the backend accepts. `@grafana/azure-sdk`'s
  `updateDatasourceCredentials` clears these fields whenever the editor
  saves, so newly-configured datasources never have them populated.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go`
`pluginschema.PluginSchema` bundle (the k8s-style schema Grafana's
datasource API server serves as `{apiVersion}.json`, `v0alpha1` today)
from the embedded `dsconfig.json`.

`SettingsExamples()` covers the default configuration plus one example per
authentication path Azure Monitor supports (config-editor-reachable AND
provisioning-only):

| Example | AuthType | Config source | Secrets populated |
| --- | --- | --- | --- |
| `""` (default) | — (empty; user must complete auth) | schema defaults | `azureClientSecret` (empty placeholder) |
| `clientSecret` | `clientsecret` | Config editor | `azureClientSecret` |
| `clientCertificatePEM` | `clientcertificate` (PEM) | Config editor | `clientCertificate` + `privateKey` |
| `clientCertificatePFX` | `clientcertificate` (PFX) | Config editor | `clientCertificate` + `certificatePassword` |
| `managedIdentity` | `msi` | Config editor (when enabled) | — |
| `workloadIdentity` | `workloadidentity` | Config editor (when enabled) | — |
| `currentUser` | `currentuser` + `clientsecret` fallback | Config editor (feature-toggled) | `azureClientSecret` (fallback secret) |
| `customizedCloud` | `clientsecret` + `AzureCustomizedCloud` + `customizedRoutes` | Provisioning-only | `azureClientSecret` |
| `legacyClientSecret` | Legacy top-level `azureAuthType=clientsecret` | Pre-migration provisioning | `clientSecret` (legacy) |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx, settings)` runs the full three-phase load flow:

1. **Parse** — unmarshal `settings.JSONData` into `Config` (jsonData
   fields become struct fields; `azureCredentials` is captured as
   `json.RawMessage` and `azureLogAnalyticsSameAs` as `json.RawMessage`
   to accept both bool and string forms), copy decrypted secrets into
   `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — no-op today (see [Modeling decisions](#modeling-decisions)).
3. **`Validate`** — mirror the runtime checks the plugin's backend
   performs during instance creation and query execution (per-authType
   secret requirements, `AzureCustomizedCloud` → `customizedRoutes`, and
   the `azureLogAnalyticsSameAs=false` hard-fail). Errors are joined.

Everything is logged via `backend.Logger.FromContext(ctx)` with
`datasource_uid`, `datasource_name`, and `plugin` labels so log lines
carry request context.

## Upstream findings

Preserved verbatim in the schema — the schema records what the plugin
**does**, not what it **should** do. These notes exist so reviewers can
reproduce each finding and decide whether to file an upstream fix.

1. **`AzureAuthType` type union lists `currentuser` twice.**
   `grafana-azure-sdk-react/src/credentials/AzureCredentials.ts:1-9` —
   `currentuser` appears at both position 1 and position 7 in the string
   union. Harmless (TypeScript de-duplicates unions), but reads like a
   copy-paste artifact.
2. **`azureLogAnalyticsSameAs`: bool-typed on the frontend, bool-or-string
   on the backend.** `src/types/types.ts:37` declares `boolean?`;
   `azure-log-analytics-datasource.go:415-427` accepts a JSON bool OR a
   string that `strconv.ParseBool` can parse. Provisioning YAML written by
   hand can (and historically did) supply `"false"` as a string, so the
   backend supports both. Our schema uses `valueType: any` to reflect
   what the backend actually accepts.
3. **Feature toggle name drift.** The frontend gates `currentuser` on
   `config.featureToggles.azureMonitorEnableUserAuth`
   (`MonitorConfig.tsx:56`); the backend enforces the same toggle at
   `azuremonitor.go:177` but with a `backend.DownstreamError` that
   effectively hard-fails instance creation. If Grafana ever removes the
   toggle without updating both sides, the plugin will refuse to load
   even when the toggle is on by default.
4. **`getLegacyCredentials` reports the wrong error on unrecognised
   `authType`.** `src/credentials.ts:39-63` — if
   `jsonData.azureAuthType` is set but unrecognised, the catch prints
   `Unable to restore legacy credentials: <msg>` and returns `undefined`,
   which silently falls through to `getDefaultCredentials()`. There is no
   surfaced error path.
5. **`AzureCustomizedCloud` is a frontend enum value but not a legacy
   `cloudName` value.** The backend's `resolveLegacyCloudName`
   (`azmoncredentials/builder.go:126-142`) maps only
   `azuremonitor` / `chinaazuremonitor` / `govazuremonitor` /
   `customizedazuremonitor` (legacy) to modern names. So a datasource
   with modern `azureCredentials.azureCloud = 'AzureCustomizedCloud'`
   but no legacy `cloudName = 'customizedazuremonitor'` still resolves to
   `AzureCustomized` via `GetAzureCloud` (`azcredentials/cloud.go:20-22`).
   The `customizedRoutes` gate in `routes.go:31-37` checks the resolved
   cloud, so both paths work — but the legacy `cloudName` field is what
   `Validate` gates on in this entry (matches the historical provisioning
   contract).
6. **Deprecated `appInsightsAppId` / `logAnalyticsDefaultWorkspace` still
   parse into `AzureMonitorSettings`.** `pkg/azuremonitor/types/types.go:29-31`
   — the backend still exposes these on `types.AzureMonitorSettings`
   even though the current query paths don't consult them, so a
   provisioning file that sets these fields will not error out but the
   values are effectively dead weight.
7. **`ConfigEditor.tsx:71` triggers a subscription API call on every
   save.** `getSubscriptions` calls `saveOptions` before making the HTTP
   request, so clicking "Load Subscriptions" persists any in-flight
   editor state as a side-effect — not obvious from the button label.
8. **`clientsecret-obo` and `ad-password` have no editor UI.**
   `AzureCredentialsForm.tsx:44-78` builds the auth-type dropdown from
   only the enabled Grafana identity toggles plus `clientsecret` and
   `clientcertificate`. `clientsecret-obo` and `ad-password` are only
   reachable via provisioning YAML. Both are still fully parsed and
   enforced by `grafana-azure-sdk-go`'s builder, so the schema keeps
   them.
9. **Reset-secret buttons desync `secureJsonFields`.**
   `AppRegistrationCredentials.tsx:143-151,231-266` — `onReset` handlers
   set the secret value to an empty string but do not clear the
   corresponding entry in `secureJsonFields`, which the editor uses to
   decide whether to render the "configured/Reset" affordance or a live
   input. Effect is usually invisible because saving-and-reload
   normalizes the state, but a datasource with `clientSecret = ''` and
   `secureJsonFields.azureClientSecret = true` is a legal
   pre-save shape that reads back as "still configured".
10. **`SecureSocksProxySettings` write path is asymmetric.** The plugin
    conditionally renders it (`ConfigEditor.tsx:120-123`) and it writes
    `jsonData.enableSecureSocksProxy` when Grafana has
    `secureSocksDSProxyEnabled`, but the plugin's backend HTTP client
    (`httpclient.go:25-50`) does not explicitly wire the secure-socks
    proxy — it relies on the shared SDK `httpclient.New` picking it up
    from `settings.HTTPClientOptions`. Works today; noted for future
    HTTP client refactors. The field is deliberately excluded from this
    entry per AGENTS.md.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` — passes.
- JSON Schema validation against
  [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft 2020-12,
  strict — `additionalProperties: false`) — passes via the conformance
  suite.
- `go build ./... && go vet ./... && gofmt -l . && go test ./...` inside
  `registry/` — passes (schema bundle shape, secure values, examples,
  `LoadConfig`, `Validate`, `EffectiveAuthType`, `ApplyDefaults` no-op
  contract, `SchemaArtifactInSync` guard).
- `tsc --noEmit --strict` on `settings.ts` — passes.
- `dsconfig` and `schema` workspace modules still build.
