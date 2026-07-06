# prometheus

Declarative configuration schema for the [Prometheus datasource plugin](https://github.com/grafana/grafana-prometheus-datasource) (`prometheus`).

## Upstream researched

- **Repo**: `github.com/grafana/grafana-prometheus-datasource`
- **Ref**: `main`
- **Commit SHA**: `be19ceb85264b27d299cb9316833c9f535d3ef6f` (2026-06-01, `docs: add signed commits requirement to CONTRIBUTING.md`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips,
option labels/values, section titles, help markdown, defaults, validations,
dependency and required-when expressions, storage keys, storage targets, value
types, group titles, and instructions — is traceable to a specific `file:line`
in the upstream repo at this SHA. See [Field provenance](#field-provenance)
below.

To reproduce this research:

```bash
git clone https://github.com/grafana/grafana-prometheus-datasource
cd grafana-prometheus-datasource
git checkout be19ceb85264b27d299cb9316833c9f535d3ef6f
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under
[Sources researched](#sources-researched) before merging any changes to this
entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, effects, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: root URL/basicAuth/basicAuthUser/withCredentials tagged `json:"-"`, jsonData fields, `DecryptedSecureJSONData`), `PluginID`, `HTTPMethod` / `PromApplication` / `PrometheusCacheLevel` / `QueryEditorMode` / `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth/TLS/HTTP variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the
shared [`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`be19ceb85264b27d299cb9316833c9f535d3ef6f`), plus external editor components
at the exact versions the plugin's `package.json` pins.

### Plugin repo (`github.com/grafana/grafana-prometheus-datasource@be19ceb`)

| File | What was read |
| --- | --- |
| `src/plugin.json:2-4,91-103` | `pluginType` (`id` = `"prometheus"`), `pluginName` (`name` = `"Prometheus"`), `docURL` (`info.links[2].url` = `"https://grafana.com/docs/grafana/latest/datasources/prometheus/"`) |
| `src/configuration/ConfigEditor.tsx:13-51` | Top-level editor — composes `DataSourceDescription`, `HttpSettings`, `ConfigSection` "Advanced settings" (containing `AdvancedHttpSettings`, `AlertingSettingsOverhaul<PromOptions>`, `PromSettings`). Note that `PromSettings` is instantiated **without** `showQuerySamplesProcessedThresholdFields`, `hidePrometheusTypeVersion`, or `hideExemplars`, so those two threshold inputs are hidden but Prometheus type/version and exemplars are visible. |
| `src/configuration/HttpSettings.tsx:47-83` | `ConnectionSettings` (URL) and `Auth` — `onAuthMethodSelect` writes `basicAuth`, `withCredentials`, and `jsonData.oauthPassThru` in one shot |
| `src/configuration/HttpSettings.tsx:50,53` | URL placeholder `"http://localhost:9090"` and label `"Prometheus server URL"` passed to `ConnectionSettings` |
| `packages/grafana-prometheus/src/types.ts:28-33,35-55,57-62` | `PromApplication` enum, `PromOptions extends DataSourceJsonData`, `ExemplarTraceIdDestination` |
| `packages/grafana-prometheus/src/types.ts:21-26` | `PrometheusCacheLevel` enum (`Low`, `Medium`, `High`, `None`) |
| `packages/grafana-prometheus/src/configuration/PromSettings.tsx:44-72` | HTTP-method options, cache-level options, Prometheus flavour options |
| `packages/grafana-prometheus/src/configuration/PromSettings.tsx:74-82` | `getOptionsWithDefaults` — defaults `httpMethod` to `POST` when unset |
| `packages/grafana-prometheus/src/configuration/PromSettings.tsx:126-208` | "Interval behaviour" section — `timeInterval` and `queryTimeout` inputs, labels, placeholders, tooltips |
| `packages/grafana-prometheus/src/configuration/PromSettings.tsx:211-273` | "Query editor" section — `defaultEditor` select (Builder/Code) and `disableMetricsLookup` switch |
| `packages/grafana-prometheus/src/configuration/PromSettings.tsx:276-499` | "Performance" section — `prometheusType`, `prometheusVersion`, `cacheLevel`, `incrementalQuerying`, `incrementalQueryOverlapWindow`, `disableRecordingRules` |
| `packages/grafana-prometheus/src/configuration/PromSettings.tsx:503-773` | "Other" section — `customQueryParameters`, `httpMethod`, `seriesLimit`, gated `maxSamplesProcessed*Threshold` inputs, `seriesEndpoint` |
| `packages/grafana-prometheus/src/configuration/PromSettings.tsx:619` | The two `maxSamplesProcessed*Threshold` inputs are wrapped in `showQuerySamplesProcessedThresholdFields` — not passed by this plugin, so they never render |
| `packages/grafana-prometheus/src/configuration/AlertingSettingsOverhaul.tsx:24-108` | `manageAlerts` and `allowAsRecordingRulesTarget` toggle labels, tooltips, and read-defaults from `config.default*` |
| `packages/grafana-prometheus/src/configuration/ExemplarSetting.tsx:22-205` | Exemplar per-item editor: `name`, `url`, `urlDisplayLabel`, `datasourceUid`, mutual exclusivity between url and datasourceUid |
| `packages/grafana-prometheus/src/configuration/ExemplarsSettings.tsx:20-77` | List wrapper — new entries default to `{ name: 'traceID' }` |
| `packages/grafana-prometheus/src/configuration/PromFlavorVersions.ts:2-99` | `PromFlavorVersions[prometheusType]` provides the version select options (Prometheus 2.0.0 - 2.50.1, Cortex 0.0.0 - 1.14.0, Mimir 2.0.0 - 2.9.1, Thanos 0.0.0 - 0.31.1); values are duplicated across many entries — we don't enumerate the full list in the schema |
| `packages/grafana-prometheus/src/constants.ts:5,15,19` | `PROM_CONFIG_LABEL_WIDTH = 30`, `NON_NEGATIVE_INTEGER_REGEX`, `DEFAULT_SERIES_LIMIT = 40000` |
| `packages/grafana-prometheus/src/querycache/QueryCache.ts:32` | `defaultPrometheusQueryOverlapWindow = '10m'` — used as the `incrementalQueryOverlapWindow` placeholder/default |
| `pkg/promlib/models/settings.go:14-22,24-53,55-61,63-96` | Backend `DataSourceJsonData`, `PromOptions` shape, `ExemplarTraceIDDestination`, `ParsePromOptions` + `ApplyDefaults` + `Validate` |
| `pkg/promlib/models/settings.go:82-87` | Backend also uppercases + defaults `HTTPMethod` to `POST`; our `Config.ApplyDefaults` mirrors this verbatim |
| `pkg/promlib/models/settings.go:92-95` | Backend rejects any HTTPMethod other than empty, `GET`, or `POST` |
| `pkg/promlib/admission_handler.go:45-53` | Admission handler hard-fails on missing `settings.URL` (encoded as `requiredWhen: "true"` on `root_url`) |
| `pkg/promlib/client/transport.go:17-43` | `CreateTransportOptions` — calls `settings.HTTPClientOptions(ctx)` (SDK reads all root/jsonData/secureJsonData HTTP fields), attaches `CustomQueryParameters` middleware and (for GET) `ForceHttpGet`, sets `ForwardHTTPHeaders=true` |
| `pkg/promlib/querydata/request.go:50-77` | Datasource instance construction — reads `settings.URL`, `jsonData.HTTPMethod`, `jsonData.QueryTimeout`, `jsonData.TimeInterval` |
| `package.json:84-98` | External component versions (see next table) |

### External editor components

Read at the exact versions pinned in the plugin's `package.json`. Sources
checked out at the corresponding upstream commits.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `ConnectionSettings` | `@grafana/plugin-ui@0.13.1` | `github.com/grafana/plugin-ui` @ `4d2f196` (post-`0.13.1`, `defaultProps` migration), `src/components/ConfigEditor/Connection/ConnectionSettings.tsx:36-75` | URL label (`urlLabel` prop) and placeholder (`urlPlaceholder` prop) come from the calling plugin — we use the values `HttpSettings.tsx:50,53` passes |
| `Auth`, `AuthMethodSettings`, `BasicAuth` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/Auth/{Auth,auth-method/AuthMethodSettings,auth-method/BasicAuth}.tsx` | Default `visibleMethods` = [BasicAuth, OAuthForward, NoAuth]; option labels/descriptions from `AuthMethodSettings.tsx:9-32`; BasicAuth `User`/`Password` labels + placeholders from `BasicAuth.tsx:24-29` |
| `convertLegacyAuthProps` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/Auth/utils.ts:27-55` | Maps `basicAuth` / `withCredentials` / `jsonData.oauthPassThru` ↔ AuthMethod enum |
| `TLSSettings`, `SelfSignedCertificate`, `TLSClientAuth`, `SkipTLSVerification` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/Auth/tls/*.tsx` | All TLS labels/placeholders/rows come verbatim from these files (see [Field provenance](#field-provenance)) |
| `CustomHeaders`, `CustomHeader` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/Auth/custom-headers/*.tsx` | Indexed `httpHeaderName<N>` / `httpHeaderValue<N>` storage pattern; **not modeled** in this schema (see [Modeling decisions](#modeling-decisions)) |
| `AdvancedHttpSettings` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/AdvancedSettings/AdvancedHttpSettings.tsx:44-82` | `Allowed cookies` and `Timeout` labels/tooltips/placeholders |
| `DataSourceDescription`, `ConfigSection`, `ConfigSubSection` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/{DataSourceDescription,ConfigSection}/*.tsx` | Intro text prop shape, section title/description props (no storage keys — layout only) |
| `SecureSocksProxySettings` (excluded) | `@grafana/ui@13.1.0-25893932881` | grafana/grafana `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | Storage key it writes (`jsonData.enableSecureSocksProxy`) — confirmed and excluded |
| `Alert`, `Input`, `SecretInput`, `SecretTextArea`, `Select`, `Switch`, `TagsInput`, `InlineField`, `Box`, `Stack`, `TextLink`, `useTheme2` | `@grafana/ui@13.1.0-25893932881` | grafana/grafana `packages/grafana-ui/src/components/` | Prop names (`label`, `placeholder`, `value`, `onChange`, `isConfigured`, `onReset`, `rows`, `width`) — needed to know which UI attributes to record |
| `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps`, `onUpdateDatasourceJsonDataOptionChecked`, `updateDatasourcePluginJsonDataOption` | `@grafana/data@13.1.0-25893932881` | grafana/grafana `packages/grafana-data/src/` | Base jsonData interface (`authType`, `defaultRegion`, `profile`, `manageAlerts`, …) that `PromOptions` extends but the Prometheus editor never touches |

Note: `@grafana/plugin-ui` published `v0.13.1` as an npm tag but did not push a
git tag; commit `4d2f196` on `main` corresponds to the changelog entry
"v0.13.1 - 2026-02-10 - Replace defaultProps with es6 defaults for React 19
compatibility" and is what npm resolved when the plugin's yarn.lock was
generated.

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line`
where each of its label, placeholder, tooltip, default, storage key, and value
type is defined. Where a field draws from multiple lines, all lines are listed.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | `HttpSettings.tsx:53` (`urlLabel="Prometheus server URL"`) | `HttpSettings.tsx:50` (`urlPlaceholder="http://localhost:9090"`) | `settings.URL string` — SDK base | Required per `admission_handler.go:51` (`requiredWhen: "true"`) |
| `virtual_authMethod` | — | virtual | `AuthMethodSettings.tsx:145` (`<Field label="Authentication method">`) | Options from `AuthMethodSettings.tsx:9-32`; default `'NoAuth'` mirrors `getSelectedMethod` fallthrough at `utils.ts:37` for a freshly-provisioned datasource | Union of 3 strings | `storage.computed.read` mirrors `getSelectedMethod` (`utils.ts:27-38`) minus `CrossSiteCredentials`, which the Prometheus editor doesn't expose; `effects` mirror `onAuthMethodSelect` (`HttpSettings.tsx:59-69`) |
| `root_basicAuth` | `basicAuth` | `root` | — (no UI; managed by `virtual_authMethod`) | Written by `HttpSettings.tsx:62`; SDK writes `settings.BasicAuthEnabled` | Root SDK bool | Tagged `managed-by:virtual_authMethod` |
| `root_basicAuthUser` | `basicAuthUser` | `root` | `BasicAuth.tsx:24` (default `userLabel = 'User'`) | `BasicAuth.tsx:26` (default `userPlaceholder = 'User'`); tooltip `BasicAuth.tsx:25` | SDK `settings.BasicAuthUser string` | `dependsOn: virtual_authMethod == 'BasicAuth'`; `requiredWhen: root_basicAuth == true` |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | `BasicAuth.tsx:27` (default `passwordLabel = 'Password'`) | `BasicAuth.tsx:29` (default `passwordPlaceholder = 'Password'`); tooltip `BasicAuth.tsx:28` | Role `auth.basic.password` | Same conditional/required as `basicAuthUser` |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (no visible UI; controlled by `virtual_authMethod == 'OAuthForward'`, and PromOptions `types.ts:50`) | Written by `HttpSettings.tsx:66` | `bool`, `types.ts:50` | Tagged `managed-by:virtual_authMethod` |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | `SelfSignedCertificate.tsx:33` (`label="Add self-signed certificate"`) | `tooltipText` `SelfSignedCertificate.tsx:34`; default `false` (utils.ts:88 read) | `bool` (SDK TLS pack) | — |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | `SelfSignedCertificate.tsx:39` (`label="CA Certificate"`) | `SelfSignedCertificate.tsx:54` (`placeholder="Begins with --- BEGIN CERTIFICATE ---"`); `rows: 6` `SelfSignedCertificate.tsx:55` | Role `tls.caCert` | `dependsOn` / `requiredWhen`: `jsonData_tlsAuthWithCACert == true` |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | `TLSClientAuth.tsx:45` (`label="TLS Client Authentication"`) | `tooltipText` `TLSClientAuth.tsx:46` | `bool` | — |
| `jsonData_serverName` | `serverName` | `jsonData` | `TLSClientAuth.tsx:51` (`label="ServerName"`) | `TLSClientAuth.tsx:63` (`placeholder="domain.example.com"`); tooltip `TLSClientAuth.tsx:53` | Role `tls.serverName` | `dependsOn: jsonData_tlsAuth == true`; required for the mTLS contract |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | `TLSClientAuth.tsx:70` (`label="Client Certificate"`) | `TLSClientAuth.tsx:88` (`placeholder="Begins with --- BEGIN CERTIFICATE ---"`); `rows: 6` `TLSClientAuth.tsx:89` | Role `tls.clientCert` | `dependsOn`/`requiredWhen`: `jsonData_tlsAuth == true` |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | `TLSClientAuth.tsx:94` (`label="Client Key"`) | `TLSClientAuth.tsx:109` (`` placeholder=`Begins with --- RSA PRIVATE KEY CERTIFICATE ---` `` — upstream typo preserved); `rows: 6` `TLSClientAuth.tsx:110` | Role `tls.clientKey` | Same conditional/required as `tlsClientCert`; see [Upstream findings](#upstream-findings) #4 for the placeholder typo |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | `SkipTLSVerification.tsx:14` (`label="Skip TLS certificate validation"`) | `tooltipText` `SkipTLSVerification.tsx:15` | Role `transport.tlsSkipVerify` | Default `false` |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | `AdvancedHttpSettings.tsx:48` (`label="Allowed cookies"`) | `AdvancedHttpSettings.tsx:56` (`placeholder="New cookie (hit enter to add)"`); tooltip `AdvancedHttpSettings.tsx:50` | `string[]` | — |
| `jsonData_timeout` | `timeout` | `jsonData` | `AdvancedHttpSettings.tsx:63` (`label="Timeout"`) | `AdvancedHttpSettings.tsx:74` (`placeholder="Timeout in seconds"`); tooltip `AdvancedHttpSettings.tsx:66` | `number` (int, parsed at `AdvancedHttpSettings.tsx:33`) | Role `transport.timeoutSeconds` |
| `jsonData_manageAlerts` | `manageAlerts` | `jsonData` | `AlertingSettingsOverhaul.tsx:43` (`label='Manage alerts via Alerting UI'`) | Read default `config.defaultDatasourceManageAlertsUiToggle` (`AlertingSettingsOverhaul.tsx:59`) — not written on load | `bool`, `types.ts` DataSourceJsonData | — |
| `jsonData_allowAsRecordingRulesTarget` | `allowAsRecordingRulesTarget` | `jsonData` | `AlertingSettingsOverhaul.tsx:77` (`label='Allow as recording rules target'`) | Read default `config.defaultAllowRecordingRulesTargetAlertsUiToggle` (`AlertingSettingsOverhaul.tsx:93`) — not written on load | `bool` | — |
| `jsonData_timeInterval` | `timeInterval` | `jsonData` | `PromSettings.tsx:135` (`label='Scrape interval'`) | `PromSettings.tsx:160` (`placeholder='15s'`); tooltip `PromSettings.tsx:143-146` | `string`, `types.ts:36` | Duration regex validated on blur `PromSettings.tsx:170` |
| `jsonData_queryTimeout` | `queryTimeout` | `jsonData` | `PromSettings.tsx:175` (`label='Query timeout'`) | `PromSettings.tsx:195` (`placeholder='60s'`) | `string`, `types.ts:37` | Duration regex validated on blur |
| `jsonData_defaultEditor` | `defaultEditor` | `jsonData` | `PromSettings.tsx:218` (`label='Default editor'`) | Options `PromSettings.tsx:90-97`; default fallback to Builder `PromSettings.tsx:239` — not written on load | `string` (`QueryEditorMode`) | `defaultValue: 'builder'` mirrors the editor's visual fallback |
| `jsonData_disableMetricsLookup` | `disableMetricsLookup` | `jsonData` | `PromSettings.tsx:250` (`label='Disable metrics lookup'`) | tooltip `PromSettings.tsx:253-256` | `bool`, `types.ts:40` | Default `false` |
| `jsonData_prometheusType` | `prometheusType` | `jsonData` | `PromSettings.tsx:299` (`label='Prometheus type'`) | Options `PromSettings.tsx:67-72` | `PromApplication`, `types.ts:28-33` | — |
| `jsonData_prometheusVersion` | `prometheusVersion` | `jsonData` | `PromSettings.tsx:336` (`label='{{promType}} version'`; the schema uses static label `"Version"` because the label is dynamic per selected type) | Options come from `PromFlavorVersions[prometheusType]` (`PromFlavorVersions.ts:2-99`) — 100+ dynamic values, so the schema uses `allowCustom: true` instead of enumerating them | `string`, `types.ts:43` | `dependsOn: jsonData_prometheusType != ''` mirrors `PromSettings.tsx:332` |
| `jsonData_cacheLevel` | `cacheLevel` | `jsonData` | `PromSettings.tsx:378` (`label='Cache level'`) | Options `PromSettings.tsx:49-54` | `PrometheusCacheLevel`, `types.ts:21-26` | Editor falls back to `Low` visually (`PromSettings.tsx:396`); default `'Low'` mirrored |
| `jsonData_incrementalQuerying` | `incrementalQuerying` | `jsonData` | `PromSettings.tsx:406` (`label='Incremental querying (beta)'`) | tooltip `PromSettings.tsx:411-414` | `bool`, `types.ts:46` | Default `false` |
| `jsonData_incrementalQueryOverlapWindow` | `incrementalQueryOverlapWindow` | `jsonData` | `PromSettings.tsx:432` (`label='Query overlap window'`) | Editor fallback `defaultPrometheusQueryOverlapWindow = '10m'` (`QueryCache.ts:32`) | `string`, `types.ts:47` | `dependsOn: jsonData_incrementalQuerying == true` |
| `jsonData_disableRecordingRules` | `disableRecordingRules` | `jsonData` | `PromSettings.tsx:479` (`label='Disable recording rules (beta)'`) | tooltip `PromSettings.tsx:484-486` | `bool`, `types.ts:48` | Default `false` |
| `jsonData_customQueryParameters` | `customQueryParameters` | `jsonData` | `PromSettings.tsx:512` (`label='Custom query parameters'`) | `PromSettings.tsx:544-547` (`placeholder='Example: max_source_resolution=5m&timeout=10'`) | `string`, `types.ts:39` | — |
| `jsonData_httpMethod` | `httpMethod` | `jsonData` | `PromSettings.tsx:565` (`label='HTTP method'`) | Options `PromSettings.tsx:44-47`; default `POST` written by `getOptionsWithDefaults` (`PromSettings.tsx:74-82`) | `string` (HTTP verb), `types.ts:38` | Backend also uppercases + defaults to `POST` at `pkg/promlib/models/settings.go:82-87` |
| `jsonData_seriesLimit` | `seriesLimit` | `jsonData` | `PromSettings.tsx:582` (`label='Series limit'`) | `PromSettings.tsx:602` (`placeholder='40000'`; `DEFAULT_SERIES_LIMIT = 40000`, `constants.ts:19`) | `number`, `types.ts:52` | — |
| `jsonData_seriesEndpoint` | `seriesEndpoint` | `jsonData` | `PromSettings.tsx:747` (`label='Use series endpoint'`) | tooltip `PromSettings.tsx:751-757` | `bool`, `types.ts:51` | Default `false` |
| `jsonData_exemplarTraceIdDestinations` | `exemplarTraceIdDestinations` | `jsonData` | `ExemplarsSettings.tsx:26` (`title='Exemplars'`) | Item defaults from `ExemplarsSettings.tsx:60` (`{ name: 'traceID' }`); item fields from `ExemplarSetting.tsx:56-179` | `ExemplarTraceIdDestination[]`, `types.ts:57-62` | Item has 4 fields (name required + url/urlDisplayLabel/datasourceUid); url and datasourceUid are mutually exclusive at the UI level (`ExemplarSetting.tsx:80-86`) |
| `jsonData_maxSamplesProcessedWarningThreshold` | `maxSamplesProcessedWarningThreshold` | `jsonData` | — (no UI in this plugin) | Parsed by `pkg/promlib/models/settings.go:41` | `float64` | Tagged `backend-only`; see [Upstream findings](#upstream-findings) #1 |
| `jsonData_maxSamplesProcessedErrorThreshold` | `maxSamplesProcessedErrorThreshold` | `jsonData` | — (no UI in this plugin) | Parsed by `pkg/promlib/models/settings.go:42` | `float64` | Tagged `backend-only` |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | Prometheus server URL | Yes (direct + SDK) |
| `virtual_authMethod` | — (virtual) | — | Authentication method | — (editor-local selector) |
| `root_basicAuth` | `basicAuth` | `root` | — (managed by virtual) | Yes (SDK) |
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
| `jsonData_manageAlerts` | `manageAlerts` | `jsonData` | Manage alerts via Alerting UI | Yes (backend `DataSourceJsonData.ManageAlerts`) |
| `jsonData_allowAsRecordingRulesTarget` | `allowAsRecordingRulesTarget` | `jsonData` | Allow as recording rules target | Yes |
| `jsonData_timeInterval` | `timeInterval` | `jsonData` | Scrape interval | Yes |
| `jsonData_queryTimeout` | `queryTimeout` | `jsonData` | Query timeout | Yes |
| `jsonData_defaultEditor` | `defaultEditor` | `jsonData` | Default editor | Parsed but only used by editor UI |
| `jsonData_disableMetricsLookup` | `disableMetricsLookup` | `jsonData` | Disable metrics lookup | Parsed; consumed by editor UI |
| `jsonData_prometheusType` | `prometheusType` | `jsonData` | Prometheus type | Parsed; consumed by heuristics/query hints |
| `jsonData_prometheusVersion` | `prometheusVersion` | `jsonData` | Version | Parsed; consumed by heuristics |
| `jsonData_cacheLevel` | `cacheLevel` | `jsonData` | Cache level | Parsed; consumed by editor caching |
| `jsonData_incrementalQuerying` | `incrementalQuerying` | `jsonData` | Incremental querying (beta) | Parsed; consumed by editor query cache |
| `jsonData_incrementalQueryOverlapWindow` | `incrementalQueryOverlapWindow` | `jsonData` | Query overlap window | Parsed; consumed by editor query cache |
| `jsonData_disableRecordingRules` | `disableRecordingRules` | `jsonData` | Disable recording rules (beta) | Parsed; consumed by resource client |
| `jsonData_customQueryParameters` | `customQueryParameters` | `jsonData` | Custom query parameters | Yes (`middleware/custom_query_params.go`) |
| `jsonData_httpMethod` | `httpMethod` | `jsonData` | HTTP method | Yes (`querydata/request.go:61`) |
| `jsonData_seriesLimit` | `seriesLimit` | `jsonData` | Series limit | Parsed; consumed by resource endpoints |
| `jsonData_seriesEndpoint` | `seriesEndpoint` | `jsonData` | Use series endpoint | Parsed; consumed by resource endpoints |
| `jsonData_exemplarTraceIdDestinations` | `exemplarTraceIdDestinations` | `jsonData` | Exemplars | Parsed; consumed by result transformer |
| `jsonData_maxSamplesProcessedWarningThreshold` | `maxSamplesProcessedWarningThreshold` | `jsonData` | — (no UI) | Yes (backend-only) |
| `jsonData_maxSamplesProcessedErrorThreshold` | `maxSamplesProcessedErrorThreshold` | `jsonData` | — (no UI) | Yes (backend-only) |

### Frontend-only settings

None of Prometheus's editor fields are pure frontend-only — every jsonData/root
field written by the editor is either parsed by the backend `PromOptions` or
consumed by the SDK's `HTTPClientOptions`.

### Backend-only settings

- **`maxSamplesProcessedWarningThreshold`** / **`maxSamplesProcessedErrorThreshold`**
  live in `PromOptions` (`pkg/promlib/models/settings.go:41-42`) and are read by
  `ParsePromOptions`, but the Prometheus plugin's own ConfigEditor never renders
  them: `PromSettings` gates them on
  `showQuerySamplesProcessedThresholdFields` (`PromSettings.tsx:619`), and
  `ConfigEditor.tsx:47` does not pass that prop. Provisioning can still set
  them.

### Excluded settings

- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy` and the associated
  socks-proxy fields) — rendered conditionally at `HttpSettings.tsx:75-79`
  when the Grafana instance has `config.secureSocksDSProxyEnabled`. Deliberately
  omitted per `AGENTS.md`.
- **Custom HTTP headers** (`@grafana/plugin-ui` `CustomHeaders`) — the editor
  writes indexed pairs `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>`
  starting at index 1. Not modeled as a first-class field because the storage
  keys are dynamic. Downstream tools should walk `jsonData` for the
  `httpHeaderName` prefix and pair up matching `httpHeaderValue<N>` secrets;
  the SDK's `HTTPClientOptions` already does this and forwards the resulting
  headers to Prometheus.
- **`access`** — the Prometheus plugin explicitly rejects `access === 'direct'`
  (Browser mode) with an inline `Alert` at `ConfigEditor.tsx:20-24`. All
  requests go through the Grafana backend (`proxy`). The field is left at the
  SDK default and not exposed here.

## Where the types are defined

The configuration types are spread across the plugin and its dependencies — some
fields and base types come from libraries/SDKs rather than the plugin itself.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `PromOptions`, `PromApplication`, `PrometheusCacheLevel`, `ExemplarTraceIdDestination` | `packages/grafana-prometheus/src/types.ts:21-62` | plugin ([grafana/grafana-prometheus-datasource](https://github.com/grafana/grafana-prometheus-datasource)) |
| `QueryEditorMode` (values `'builder'` / `'code'`) | `packages/grafana-prometheus/src/querybuilder/shared/types.ts` | plugin |
| `defaultPrometheusQueryOverlapWindow` | `packages/grafana-prometheus/src/querycache/QueryCache.ts:32` | plugin |
| `DataSourceJsonData` (base interface: `authType`, `defaultRegion`, `profile`, `manageAlerts`, `allowAsRecordingRulesTarget`, `alertmanagerUid`, `disableGrafanaCache`) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `13.1.0-25893932881` |
| `DataSourcePluginOptionsEditorProps`, `DataSourceSettings`, `updateDatasourcePluginJsonDataOption`, `onUpdateDatasourceJsonDataOptionChecked` | `packages/grafana-data/src/` | `@grafana/data` `13.1.0-25893932881` |
| `ConnectionSettings`, `Auth`, `AuthMethod`, `convertLegacyAuthProps`, `AdvancedHttpSettings`, `DataSourceDescription`, `ConfigSection`, `ConfigSubSection` | `src/components/ConfigEditor/…` | `@grafana/plugin-ui` `0.13.1` (grafana/plugin-ui @ `4d2f196`) |
| `SecureSocksProxySettings` (excluded) | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | `@grafana/ui` `13.1.0-25893932881` |
| `Alert`, `Input`, `SecretInput`, `SecretTextArea`, `Select`, `Switch`, `TagsInput`, `TextLink`, `useTheme2` | `packages/grafana-ui/src/components/` | `@grafana/ui` `13.1.0-25893932881` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `PromOptions` (jsonData), `DataSourceJsonData` (base struct), `ExemplarTraceIDDestination`, `ParsePromOptions`, `ApplyDefaults`, `Validate` | `pkg/promlib/models/settings.go:14-96` | plugin |
| `Service`, `newInstanceSettings`, `QueryData.New`, `Resource.New` (all read `settings.URL` + `jsonData.HTTPMethod` + `jsonData.QueryTimeout`) | `pkg/promlib/{library.go,querydata/request.go,resource/resource.go}` | plugin |
| `CreateTransportOptions` (delegates HTTP options to the SDK, adds `CustomQueryParameters` and `ForceHttpGet` middlewares, sets `ForwardHTTPHeaders=true`) | `pkg/promlib/client/transport.go:17-43` | plugin |
| Admission handler (rejects missing `settings.URL`) | `pkg/promlib/admission_handler.go:45-53` | plugin |
| `backend.DataSourceInstanceSettings` (carries `URL`, `BasicAuthEnabled`, `BasicAuthUser`, `JSONData`, `DecryptedSecureJSONData`), `HTTPClientOptions(ctx)` | `backend/common.go`, `backend/httpclient/` | `github.com/grafana/grafana-plugin-sdk-go` |

The models in this entry flatten that spread into a single Go `Config` type
(root `URL`/`BasicAuth`/`BasicAuthUser`/`WithCredentials` tagged `json:"-"`,
plus the jsonData fields, plus `DecryptedSecureJSONData`) plus a
`SecureJsonDataKey` typed constant list. `settings.ts` keeps the three
canonical TypeScript types (`RootConfig`, `JsonDataConfig`,
`SecureJsonDataConfig`).

## Modeling decisions

- **Virtual auth method**: `Auth`'s `onAuthMethodSelect` (`HttpSettings.tsx:59-69`)
  writes three storage fields in one shot — `root.basicAuth`,
  `root.withCredentials`, and `jsonData.oauthPassThru`. That is exactly the
  virtual-selector pattern from the GitHub datasource. `withCredentials` is not
  in the Prometheus editor's `visibleMethods` (defaults per
  `AuthMethodSettings.tsx:57-66` are `[BasicAuth, OAuthForward, NoAuth]`), so
  the virtual field's effects only write `basicAuth` and `oauthPassThru`, and
  `withCredentials` is left off the schema entirely. If a provisioning payload
  writes `withCredentials=true` directly, the SDK still honors it — the
  virtual's `storage.computed.read` doesn't preserve that state, but the
  underlying storage does.
- **`requiredWhen` on `basicAuthUser` / `basicAuthPassword`**: keyed on the
  underlying storage field (`root_basicAuth == true`), not the virtual selector.
  The virtual is an editor-local convenience; the backend contract is "if
  basicAuth is on, we need a username and password".
- **TLS pair requirements**: `TLSClientAuth` and `SelfSignedCertificate` mark
  every field with `required` in the UI, but they only require the paired
  fields when the parent switch is on. Encoded as `dependsOn` + `requiredWhen`
  on each field.
- **No help drawer**: Prometheus's editor has no top-level `Collapse`/`help`
  panel (all inline tooltips), so there is no schema `help` object. The
  detailed guidance is captured in `description` on individual fields and in
  the `instructions` block.
- **PrometheusVersion has no `allowedValues`**: the option list is dynamic per
  `prometheusType` and includes 100+ values across the four flavours
  (`PromFlavorVersions.ts:2-99`). We set `allowCustom: true` on the UI and skip
  enumerating them; downstream tools that need the list should read
  `PromFlavorVersions.ts` directly.
- **Field ID naming convention**: IDs are prefixed with their storage target
  for easy discoverability — `root_`, `jsonData_`, or `secureJsonData_` (and
  `virtual_` for virtual fields, which have no storage target) — followed by
  the camelCase storage key. The `key` property keeps the plugin's raw storage
  key.
- **Custom HTTP headers excluded**: see [Excluded settings](#excluded-settings)
  above.
- **Flat `Config` in Go**: `settings.go` collapses jsonData fields and
  decrypted secrets onto a single `Config` struct. Base `DataSourceJsonData`
  fields (authType, defaultRegion, profile, alertmanagerUid,
  disableGrafanaCache) exist in the upstream `PromOptions` via embedding but
  are omitted here because the Prometheus editor never writes them and the
  Prometheus code never reads them — the `JSONDataMatchesStruct` conformance
  test would fail if we included them without schema fields. Root-level
  fields the editor and SDK both use (`URL`, `BasicAuth`, `BasicAuthUser`,
  `WithCredentials`) are carried with `json:"-"` tags so `LoadConfig` returns
  them alongside the jsonData shape.
- **`SecureJsonDataConfig` is a key list**: secure values are write-only, so
  the secure type is just the array of secret key names (`basicAuthPassword`,
  `tlsCACert`, `tlsClientCert`, `tlsClientKey`); consumers read
  `secureJsonFields` to see what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema`
bundle (the k8s-style schema Grafana's datasource API server serves as
`{apiVersion}.json`, `v0alpha1` today) from the embedded `dsconfig.json`: root
fields plus a nested `jsonData` object become the OpenAPI settings `spec`,
secure fields become `secureValues`, and virtual fields are skipped.

`SettingsExamples()` provides the default configuration plus one k8s-style
example per authentication method, TLS variant, HTTP method, and one Mimir /
exemplars combination. Each example is a full instance-settings object with
the plugin configuration nested under `jsonData` and the relevant write-only
secrets under `secureJsonData` (placeholder values to be replaced with real
secrets; the default example — keyed by the empty string `""` — carries an
empty `basicAuthPassword` to show that no secret is required for the default
No-auth mode):

| Example | Auth | HTTP method | TLS | `secureJsonData` |
| --- | --- | --- | --- | --- |
| `""` (default) | None | POST | — | `basicAuthPassword` (empty) |
| `noAuth` | None | POST | — | `basicAuthPassword` (empty) |
| `basicAuth` | Basic | POST | — | `basicAuthPassword` |
| `oauthForward` | OAuth Identity | POST | — | `basicAuthPassword` (empty) |
| `tlsMutualAuth` | None | POST | mTLS (serverName + client cert/key) | `tlsClientCert`, `tlsClientKey` |
| `tlsSelfSignedCA` | None | POST | Custom CA | `tlsCACert` |
| `getHTTPMethod` | None | GET | — | `basicAuthPassword` (empty) |
| `mimirWithExemplars` | Basic | POST | — | `basicAuthPassword` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and
returns a fully-defaulted, validated `Config`:

1. **Parse** — copy `settings.URL`, `settings.BasicAuthEnabled`,
   `settings.BasicAuthUser` into `Config`, unmarshal `settings.JSONData` into
   the jsonData portion of the same struct, and copy the four decrypted secrets
   into `DecryptedSecureJSONData`. This mirrors the split reads the upstream
   `ParsePromOptions` (`pkg/promlib/models/settings.go:65-79`) plus
   `CreateTransportOptions` (`pkg/promlib/client/transport.go:17-43`) perform.
2. **`ApplyDefaults`** — uppercase and default `HTTPMethod` to `POST`, matching
   both `getOptionsWithDefaults` (`PromSettings.tsx:74-82`) and the backend's
   own `ApplyDefaults` (`pkg/promlib/models/settings.go:82-87`). No other
   fields are defaulted on load — the editor's other visual fallbacks (cache
   level Low, default editor Builder, series limit 40000, overlap window 10m)
   are never written to storage on load.
3. **`Validate`** — enforce the runtime contract: URL is required
   (`admission_handler.go:51`), `HTTPMethod` must be empty/POST/GET
   (`pkg/promlib/models/settings.go:92-95`), Basic auth requires a username,
   mTLS requires serverName + client cert + client key, custom-CA requires the
   CA PEM, and numeric fields (timeout, seriesLimit) must be non-negative.
   Errors are joined so every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with
`datasource_uid`, `datasource_name`, and `plugin` labels so log lines carry
request context.

This is the intended shape for the plugin's own upstream `ParsePromOptions` to
sync to (with the addition of root-field capture that `ParsePromOptions` itself
doesn't do — the SDK does it separately).

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are still exported
for callers that want to compose them themselves (e.g. provisioning preview,
schema-example round-trip, tests that need to distinguish parse-level from
policy-level errors). Skip them by never calling `LoadConfig` in those flows —
assemble a `Config` directly.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while
researching upstream. All preserved verbatim in the schema — the schema records
what the plugin **does**, not what it **should** do; these notes exist so
reviewers can reproduce each finding and decide separately whether to fix
upstream.

1. **Hidden threshold inputs**: `pkg/promlib/models/settings.go:41-42` parses
   `maxSamplesProcessedWarningThreshold` and `maxSamplesProcessedErrorThreshold`
   from jsonData, but the `PromSettings` component only renders the inputs
   when `showQuerySamplesProcessedThresholdFields` is true
   (`PromSettings.tsx:619`), and `ConfigEditor.tsx:47` does not pass that prop.
   The fields therefore work only for provisioning. Preserved as
   `backend-only`-tagged fields in the schema without editor UI.
2. **Editor visual defaults are not persisted**: `PromSettings.tsx` renders
   cache level as `... ?? PrometheusCacheLevel.Low` (`:396`), default editor as
   `... ?? Builder` (`:239`), and series limit's placeholder as `40000`. These
   look like defaults but are never written on load. A datasource created via
   the API with `jsonData: {}` will not have `cacheLevel`, `defaultEditor`, or
   `seriesLimit` set, and the backend never fills them. The schema records
   them as visual defaults (`defaultValue`) so downstream tools can render the
   same UI, but callers should be aware the storage stays sparse.
3. **`httpMethod` normalisation is inconsistent between load paths**: the
   backend's `ApplyDefaults` (`pkg/promlib/models/settings.go:82-87`)
   uppercases and defaults to POST. The editor's `getOptionsWithDefaults`
   (`PromSettings.tsx:74-82`) only assigns POST when the field is falsy — it
   does not uppercase. A datasource stored with `"httpMethod":"get"` will
   render fine (the Select's `.find(o => o.value === 'get')` returns
   `undefined` and the select shows blank) but the backend will treat it as
   GET. Our `Config.ApplyDefaults` mirrors the backend behavior.
4. **Upstream typo preserved**: `TLSClientAuth.tsx:109` sets the client key
   placeholder to `` `Begins with --- RSA PRIVATE KEY CERTIFICATE ---` `` — an
   RSA private key is not a "certificate". Preserved verbatim in
   `secureJsonData_tlsClientKey.ui.placeholder`.
5. **Browser access mode is rejected but still storable**: `ConfigEditor.tsx:20-24`
   renders an inline `Alert` when `options.access === 'direct'`, but nothing
   prevents storing that value. The backend's admission handler
   (`pkg/promlib/admission_handler.go`) does not check `access`. A provisioned
   datasource with `access: "direct"` will be accepted by the API and rendered
   with the error banner every time an admin edits it.
6. **`incrementalQueryOverlapWindow` default only exists in the editor**:
   `PromSettings.tsx:465` sets the input value to
   `optionsWithDefaults.jsonData.incrementalQueryOverlapWindow ?? defaultPrometheusQueryOverlapWindow`
   (`'10m'`). The value is only persisted if the user edits the input. If
   `incrementalQuerying` is true and `incrementalQueryOverlapWindow` was never
   touched, the query cache falls back to `'10m'` at request time.
7. **Empty `queryTimeout` means "no timeout param"**: `client.go:45-47` skips
   the `timeout` URL parameter when `c.queryTimeout == ""`. The editor's
   placeholder `'60s'` (`PromSettings.tsx:195`) is only cosmetic — an empty
   value does NOT mean "60s"; it means "let Prometheus decide".
8. **Prometheus `Version` label is dynamic**: `PromSettings.tsx:336` renders
   the label as `'{{promType}} version'` interpolated with the currently
   selected `prometheusType`. The schema uses a static label `"Version"`
   because dsconfig has no way to declare dynamic labels; downstream tools can
   substitute the type name if desired.
9. **Base `DataSourceJsonData` fields are unused**: the backend `PromOptions`
   embeds `DataSourceJsonData` (`pkg/promlib/models/settings.go:14-22`) which
   carries `authType`, `defaultRegion`, `profile`, `alertmanagerUid`, and
   `disableGrafanaCache`. None of these are written by the Prometheus editor
   and none are read by the Prometheus plugin's own code (`manageAlerts` and
   `allowAsRecordingRulesTarget` in the same struct are the only two that are
   used). The unused base fields are omitted from the schema.
10. **Custom headers can leak plaintext**: `@grafana/plugin-ui`
    `getCustomHeaders` (`utils.ts:203-232`) writes header values to
    `secureJsonData` when the user creates a new header, but sets `configured:
    false` while typing (`CustomHeader.tsx:74-75`), so `secureJsonFields`
    reports the header as unconfigured until the datasource is saved. This is
    intentional (keeps the input editable) but downstream tooling that trusts
    `secureJsonFields.httpHeaderValueN` to know whether a value is configured
    can be misled during editing.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in
  this repo) — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft-07, `additionalProperties: false`) — passes.
- `go test ./...` on this module — passes (schema bundle shape, secure values,
  examples, `LoadConfig` incl. TLS variants and malformed input,
  `SchemaArtifactInSync` guard, `JSONDataMatchesStruct`,
  `JSONDataTypesMatchStruct`).
- `settings.go`/`schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: `tsc --noEmit --strict` clean at the pinned versions.
