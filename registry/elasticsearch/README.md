# elasticsearch

Declarative configuration schema for the [Elasticsearch datasource plugin](https://github.com/grafana/grafana-elasticsearch-datasource) (`elasticsearch`).

## Upstream researched

- **Repo**: `github.com/grafana/grafana-elasticsearch-datasource`
- **Ref**: `main`
- **Commit SHA**: `265a51efe4b2fa3d025c816d1bec7d6abc8d5d89` (2026-07-02, `docs: add signed commits requirement to CONTRIBUTING.md (#351)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, option
labels/values, section titles, defaults, validations, dependency and required-when
expressions, storage keys, storage targets, value types, group titles, and instructions
— is traceable to a specific `file:line` in the upstream repo at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone https://github.com/grafana/grafana-elasticsearch-datasource
cd grafana-elasticsearch-datasource
git checkout 265a51efe4b2fa3d025c816d1bec7d6abc8d5d89
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under
[Sources researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, effects, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (root fields + jsonData + `DecryptedSecureJSONData`), `PluginID`, `SecureJsonDataKey` typed constants, `LoadConfig`, `ApplyDefaults`, `Validate` |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth/connection variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`265a51efe4b2fa3d025c816d1bec7d6abc8d5d89`), plus external editor components at the exact
versions the plugin's `package.json` pins.

### Plugin repo (`github.com/grafana/grafana-elasticsearch-datasource@265a51e`)

| File | What was read |
| --- | --- |
| `src/plugin.json:1-53` | `pluginType` (`id`), `pluginName` (`name`), `docURL` (`info.links[2].url`) |
| `src/configuration/ConfigEditor.tsx:36-132` | Auth wiring (`convertLegacyAuthProps`, `customMethods` for API Key and SigV4, `onAuthMethodSelect`), `ConnectionSettings` URL placeholder override, browser-access deprecation Alert, "Additional settings" collapsible ConfigSection |
| `src/configuration/ConfigEditor.tsx:80-92` | `onAuthMethodSelect` — the multi-field discriminator write recorded as `effects` on `virtual_authMethod` |
| `src/configuration/ConfigEditor.tsx:65-69` | "Browser access mode ... is no longer available" error banner triggered by `options.access === 'direct'` |
| `src/configuration/ConfigEditor.tsx:103-105` | Conditional `SecureSocksProxySettings` — deliberately excluded from this entry |
| `src/configuration/ApiKeyConfig.tsx:7-23` | Custom API Key auth field — labels, placeholder (`Enter your API key`), tooltip (`API Key authentication`), reset behavior |
| `src/configuration/ElasticDetails.tsx:11-18` | Index-pattern selector options (`No pattern` / `Hourly` / `Daily` / `Weekly` / `Monthly` / `Yearly`) and their example templates |
| `src/configuration/ElasticDetails.tsx:36-154` | All Elasticsearch-detail fields: Index name (`required`, placeholder `es-index-name`), Pattern, Time field name (`required`, placeholder `@timestamp`), Max concurrent Shard Requests, Min time interval (with `^\d+(ms|[Mwdhmsy])$` validation), Include Frozen Indices, Default query mode |
| `src/configuration/ElasticDetails.tsx:159-232` | `indexChangeHandler` (always clears `database`), `intervalHandler` (auto-fills a bracketed index template when interval changes), `jsonDataChangeHandler`, `jsonDataSwitchChangeHandler` |
| `src/configuration/ElasticDetails.tsx:234-239` | `defaultMaxConcurrentShardRequests()` -> 5, `defaultQueryMode()` -> `'metrics'` |
| `src/configuration/LogsConfig.tsx:22-59` | "Logs" sub-section — Message field name (placeholder `_source`, tooltip verbatim), Level field name (no placeholder), tooltips |
| `src/configuration/DataLinks.tsx:33-91` | "Data links" sub-section — description, empty state, Add button |
| `src/configuration/DataLink.tsx:41-136` | Individual data link entry — Field input (tooltip verbatim), URL/Query switch (label depends on `showInternalLink`), URL Label input, "Internal link" InlineSwitch (drives datasourceUid) |
| `src/configuration/utils.ts:7-29` | `QUERY_TYPE_SELECTOR_OPTIONS` (verbatim label/value pairs), `coerceOptions` — the on-mount defaulter that writes `timeField='@timestamp'`, `maxConcurrentShardRequests=5`, `logMessageField=''`, `logLevelField=''`, `includeFrozen=false`, `defaultQueryMode='metrics'` |
| `src/types.ts:60-82` | `ElasticsearchOptions`, `Interval`, `QueryType`, `ElasticsearchSecureJsonData` |
| `src/types.ts:135-140` | `DataLinkConfig` shape |
| `pkg/plugin.json`, `pkg/main.go:11-24` | Confirmed backend `Manage("elasticsearch", ...)` binding matches the plugin ID |
| `pkg/elasticsearch/elasticsearch.go:106-243` | `NewDatasource` — jsonData is unmarshaled as `map[string]any`, apiKeyAuth is checked before every request, timeField is hard-required, `maxConcurrentShardRequests` supports JSON number OR string, empty index falls back to `settings.Database`, SigV4.Service is forced to `"es"` |
| `pkg/elasticsearch/healthcheck.go:21-52` | Health check probes `<url>/_cluster/health?wait_for_status=yellow` — implicit URL requirement encoded as `requiredWhen: "true"` on `root_url` |
| `pkg/elasticsearch/client/client.go:30-77` | `DatasourceInfo` — confirms which jsonData fields the client consumes (`Database`, `Interval`, `MaxConcurrentShardRequests`, `IncludeFrozen`, `ConfiguredFields`, `URL`) |
| `package.json` | External component versions (see next table) |

### External editor components

Read at the exact versions pinned in the plugin's `package.json`
(`@grafana/aws-sdk@^0.10.0`, `@grafana/data@^13.1.0`, `@grafana/plugin-ui@^0.13.1`,
`@grafana/runtime@^13.1.0`, `@grafana/schema@^13.1.0`, `@grafana/ui@^13.1.0`).

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `ConnectionSettings` | `@grafana/plugin-ui@0.13.1` | `dist/esm/components/ConfigEditor/Connection/ConnectionSettings.js` | Section title (`"Connection"`), URL Field `label`/`labelWidth`/`required`, default URL tooltip and placeholder (overridden here to `http://localhost:9200`) |
| `Auth`, `AuthMethod`, `AuthMethodSettings`, `convertLegacyAuthProps` | `@grafana/plugin-ui@0.13.1` | `dist/esm/components/ConfigEditor/Auth/{Auth,utils,auth-method/AuthMethodSettings,auth-method/BasicAuth,types}.js` | Auth section title (`"Authentication"`), default visibleMethods (`BasicAuth`, `OAuthForward`, `NoAuth`, plus custom method IDs), the load-time `getSelectedMethod` derivation used by our `virtual_authMethod` storage/read expression, BasicAuth `User`/`Password` labels and placeholders |
| `TLSSettings`, `SelfSignedCertificate`, `TLSClientAuth`, `SkipTLSVerification` | `@grafana/plugin-ui@0.13.1` | `dist/esm/components/ConfigEditor/Auth/tls/*.js` | TLS section labels and tooltips verbatim: "Add self-signed certificate", "CA Certificate" (placeholder `Begins with --- BEGIN CERTIFICATE ---`, rows 6), "TLS Client Authentication", "ServerName" (placeholder `domain.example.com`), "Client Certificate", "Client Key" (placeholder `Begins with --- RSA PRIVATE KEY CERTIFICATE ---`), "Skip TLS certificate validation" |
| `AdvancedHttpSettings` | `@grafana/plugin-ui@0.13.1` | `dist/esm/components/ConfigEditor/AdvancedSettings/AdvancedHttpSettings.js` | Section title "Advanced HTTP settings" (the plugin wraps this inside its own "Additional settings" ConfigSection), "Allowed cookies" TagsInput (placeholder `New cookie (hit enter to add)`), "Timeout" numeric input (placeholder `Timeout in seconds`) |
| `DataSourceDescription`, `ConfigSection`, `ConfigSubSection`, `ConfigDescriptionLink` | `@grafana/plugin-ui@0.13.1` | `dist/esm/components/ConfigEditor/*` | Layout, `title`/`isCollapsible`/`isInitiallyOpen` props |
| `SIGV4ConnectionConfig` (contributed only, not modeled) | `@grafana/aws-sdk@0.10.0` | Read via the plugin's usage in `src/configuration/ConfigEditor.tsx:50-61` | Contributes `sigV4Auth`, `sigV4AuthType`, `sigV4Region`, `sigV4Profile`, `sigV4AssumeRoleArn`, `sigV4ExternalId` (jsonData) and `sigV4AccessKey`, `sigV4SecretKey` (secureJsonData) when the Grafana instance has `config.sigV4AuthEnabled` |
| `SecureSocksProxySettings` (excluded) | `@grafana/ui@13.1.0` | Rendered conditionally at `ConfigEditor.tsx:103-105`. Storage key `jsonData.enableSecureSocksProxy` deliberately excluded per repo `AGENTS.md` |
| `Alert`, `Divider`, `Stack`, `InlineField`, `Input`, `InlineSwitch`, `Select`, `SecretInput`, `TagsInput`, `Button`, `DataLinkInput` | `@grafana/ui@13.1.0` | Prop names (`label`, `placeholder`, `value`, `onChange`, `isConfigured`, `onReset`, `width`, `labelWidth`, `rows`) so we knew which UI attributes to record |
| `DataSourcePluginOptionsEditorProps`, `updateDatasourcePluginResetOption`, `onUpdateDatasourceSecureJsonDataOption` | `@grafana/data@13.1.0` | Storage-key semantics of the update helpers used by `ApiKeyConfig` |
| `config.sigV4AuthEnabled`, `config.secureSocksDSProxyEnabled` | `@grafana/runtime@13.1.0` | Feature gates for the conditional SigV4 method and secure-socks proxy toggle |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each
of its label, placeholder, tooltip, default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `root_url` | `url` | root | @grafana/plugin-ui `ConnectionSettings.js:26` (`"URL"`) | Overridden placeholder `"http://localhost:9200"` at `ConfigEditor.tsx:76` | Backend reads `settings.URL` at `elasticsearch.go:196,213` | Role `endpoint.baseUrl`; `requiredWhen: "true"` because health check parses it |
| `virtual_authMethod` | — | virtual | `AuthMethodSettings.js:126` (`"Authentication method"`) | Options built from @grafana/plugin-ui `defaultOptions` + `customMethods` at `AuthMethodSettings.js:10-31, 66-77`; default `NoAuth` from `getSelectedMethod` (`utils.js:24-35`) | Union of 5 strings; `AuthMethod` at `types.js:3-8` + custom method IDs at `ConfigEditor.tsx:42,54` | Load-time derivation mirrors `getSelectedMethod` (checked in this order: apiKeyAuth → sigV4Auth → basicAuth → oauthPassThru → NoAuth). Write effects mirror `onAuthMethodSelect` (`ConfigEditor.tsx:80-92`) |
| `root_basicAuth` | `basicAuth` | root | — (managed by `virtual_authMethod`) | Default `false` mirrors editor's onAuthMethodSelect | `settings.BasicAuthEnabled` on the SDK settings struct | Role `auth.basic.enabled` |
| `root_basicAuthUser` | `basicAuthUser` | root | `BasicAuth.js:9` (`"User"`, default label) | `"User"` placeholder (`BasicAuth.js:11`); tooltip verbatim `"The username of the data source account"` (`BasicAuth.js:10`) | `settings.BasicAuthUser` | Role `auth.basic.username`; `dependsOn`/`requiredWhen` from BasicAuth `required: true` (`BasicAuth.js:33,46`) |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | secureJsonData | `BasicAuth.js:12` (`"Password"`) | Placeholder `"Password"` (`BasicAuth.js:14`); tooltip `"The password of the data source account"` (`BasicAuth.js:13`) | Handled by SDK HTTPClient auth wiring; secret key `basicAuthPassword` (`utils.js:52-63`) | Role `auth.basic.password` |
| `jsonData_oauthPassThru` | `oauthPassThru` | jsonData | — (managed by `virtual_authMethod`) | Default `false` mirrors editor's onAuthMethodSelect | `map[string]any` unmarshal in `elasticsearch.go` (SDK consumes the field via `settings.HTTPClientOptions`) | Role `auth.forwardOAuthToken.enabled` |
| `jsonData_apiKeyAuth` | `apiKeyAuth` | jsonData | — (managed by `virtual_authMethod`) | Default `false` mirrors editor's onAuthMethodSelect | Backend `jsonData["apiKeyAuth"].(bool)` at `elasticsearch.go:125` | Role `auth.discriminator` — the boolean gate the backend actually checks |
| `secureJsonData_apiKey` | `apiKey` | secureJsonData | `ApiKeyConfig.tsx:11` (`"API Key"`) | Placeholder `"Enter your API key"` (`ApiKeyConfig.tsx:16`); tooltip `"API Key authentication"` (`ApiKeyConfig.tsx:11`) | Backend `settings.DecryptedSecureJSONData["apiKey"]` at `elasticsearch.go:127` | Role `auth.apiKey.value`; `dependsOn`/`requiredWhen` mirrors ApiKeyConfig `required` (`ApiKeyConfig.tsx:13`) |
| `jsonData_sigV4Auth` | `sigV4Auth` | jsonData | — (managed by `virtual_authMethod`) | Default `false` mirrors editor's onAuthMethodSelect | Backend `httpCliOpts.SigV4 != nil` check at `elasticsearch.go:120-123` | Role `auth.awsSigV4.enabled`; the sigV4* companion fields are contributed by @grafana/aws-sdk's SIGV4ConnectionConfig and are not modeled here |
| `jsonData_tlsAuth` | `tlsAuth` | jsonData | `TLSClientAuth.js:27` (`"TLS Client Authentication"`) | Tooltip `"Validate using TLS client authentication ..."` (`TLSClientAuth.js:28`); default `false` | @grafana/plugin-ui `TLSClientAuth`, consumed by SDK HTTPClient | — |
| `jsonData_serverName` | `serverName` | jsonData | `TLSClientAuth.js:35` (`"ServerName"`) | Placeholder `"domain.example.com"` (`TLSClientAuth.js:49`); tooltip verbatim (`TLSClientAuth.js:37`) | @grafana/plugin-ui TLSClientAuth | Role `tls.serverName` |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | secureJsonData | `TLSClientAuth.js:59` (`"Client Certificate"`) | Placeholder `"Begins with --- BEGIN CERTIFICATE ---"` (`TLSClientAuth.js:77`); rows 6 (`TLSClientAuth.js:78`) | @grafana/plugin-ui TLSClientAuth | Role `tls.clientCert` |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | secureJsonData | `TLSClientAuth.js:86` (`"Client Key"`) | Placeholder `"Begins with --- RSA PRIVATE KEY CERTIFICATE ---"` (`TLSClientAuth.js:104`); rows 6 | @grafana/plugin-ui TLSClientAuth | Role `tls.clientKey`; upstream placeholder typo (see [Upstream findings](#upstream-findings)) preserved |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | jsonData | `SelfSignedCertificate.js:22` (`"Add self-signed certificate"`) | Tooltip `"Add your own Certificate Authority ..."` (`SelfSignedCertificate.js:23`); default `false` | @grafana/plugin-ui SelfSignedCertificate | — |
| `secureJsonData_tlsCACert` | `tlsCACert` | secureJsonData | `SelfSignedCertificate.js:30` (`"CA Certificate"`) | Placeholder `"Begins with --- BEGIN CERTIFICATE ---"` (`SelfSignedCertificate.js:48`); rows 6 (`SelfSignedCertificate.js:49`) | @grafana/plugin-ui SelfSignedCertificate | Role `tls.caCert` |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | jsonData | `SkipTLSVerification.js:9` (`"Skip TLS certificate validation"`) | Tooltip `"Skipping TLS certificate validation ..."` (`SkipTLSVerification.js:10`); default `false` | @grafana/plugin-ui SkipTLSVerification | Role `transport.tlsSkipVerify` |
| `jsonData_keepCookies` | `keepCookies` | jsonData | `AdvancedHttpSettings.js:39` (`"Allowed cookies"`) | Placeholder `"New cookie (hit enter to add)"` (`AdvancedHttpSettings.js:49`); tooltip verbatim (`AdvancedHttpSettings.js:41`) | @grafana/plugin-ui AdvancedHttpSettings | — |
| `jsonData_timeout` | `timeout` | jsonData | `AdvancedHttpSettings.js:58` (`"Timeout"`) | Placeholder `"Timeout in seconds"` (`AdvancedHttpSettings.js:70`); tooltip `"HTTP request timeout in seconds"` (`AdvancedHttpSettings.js:60`) | @grafana/plugin-ui AdvancedHttpSettings; parsed via `parseInt` (`AdvancedHttpSettings.js:26`) | Role `transport.timeoutSeconds` |
| `jsonData_index` | `index` | jsonData | `ElasticDetails.tsx:37` (`"Index name"`) | Placeholder `"es-index-name"` (`ElasticDetails.tsx:47`); tooltip verbatim (`ElasticDetails.tsx:40`); `required` in the input (`ElasticDetails.tsx:48`) | Backend `jsonData["index"].(string)` at `elasticsearch.go:164-170` (with legacy `settings.Database` fallback) | — |
| `jsonData_interval` | `interval` | jsonData | `ElasticDetails.tsx:53` (`"Pattern"`) | Options `indexPatternTypes` at `ElasticDetails.tsx:11-18`; tooltip verbatim (`ElasticDetails.tsx:56`); "No pattern" is the empty-value placeholder | Backend `jsonData["interval"].(string)` at `elasticsearch.go:159-162` | Editor stores `''`/undefined for "No pattern" and the enum string otherwise |
| `jsonData_timeField` | `timeField` | jsonData | `ElasticDetails.tsx:70` (`"Time field name"`) | Placeholder `"@timestamp"` (`ElasticDetails.tsx:80`); tooltip verbatim (`ElasticDetails.tsx:73`); default `"@timestamp"` from `coerceOptions` (`utils.ts:21`); `required` input (`ElasticDetails.tsx:81`) | Backend hard-fail on empty timeField at `elasticsearch.go:145-147` | — |
| `jsonData_maxConcurrentShardRequests` | `maxConcurrentShardRequests` | jsonData | `ElasticDetails.tsx:86` (`"Max concurrent Shard Requests"`) | Tooltip verbatim (`ElasticDetails.tsx:89`); default `5` from `defaultMaxConcurrentShardRequests()` (`ElasticDetails.tsx:234-236`); backend coerces to 5 when missing/non-positive (`elasticsearch.go:172-189`) | Backend supports both JSON number and JSON string; mirrored in `Config.UnmarshalJSON` | — |
| `jsonData_timeInterval` | `timeInterval` | jsonData | `ElasticDetails.tsx:100` (`"Min time interval"`) | Placeholder `"10s"` (`ElasticDetails.tsx:117`); tooltip verbatim (`ElasticDetails.tsx:103-108`); validation regex `^\d+(ms|[Mwdhmsy])$` (`ElasticDetails.tsx:110`) | Stored as string; backend does not enforce the regex (editor-only validation) | — |
| `jsonData_includeFrozen` | `includeFrozen` | jsonData | `ElasticDetails.tsx:121` (`"Include Frozen Indices"`) | Tooltip `"Include frozen indices in searches."` (`ElasticDetails.tsx:124`); default `false` | Backend `jsonData["includeFrozen"].(bool)` at `elasticsearch.go:191-194`, adds `ignore_throttled=false` to msearch query (`client.go:215-217`) | — |
| `jsonData_defaultQueryMode` | `defaultQueryMode` | jsonData | `ElasticDetails.tsx:134` (`"Default query mode"`) | Options `QUERY_TYPE_SELECTOR_OPTIONS` at `utils.ts:7-12`; tooltip verbatim (`ElasticDetails.tsx:137`); default `'metrics'` from `defaultQueryMode()` (`ElasticDetails.tsx:237-239`) | Frontend-only: consumed by `datasource.ts` to pick the query editor's default type | — |
| `jsonData_logMessageField` | `logMessageField` | jsonData | `LogsConfig.tsx:34` (`"Message field name"`) | Placeholder `"_source"` (`LogsConfig.tsx:42`); tooltip verbatim (`LogsConfig.tsx:36`) | Backend `jsonData["logMessageField"].(string)` at `elasticsearch.go:154-157`, feeds `ConfiguredFields.LogMessageField` | — |
| `jsonData_logLevelField` | `logLevelField` | jsonData | `LogsConfig.tsx:48` (`"Level field name"`) | No placeholder; tooltip verbatim (`LogsConfig.tsx:50`) | Backend `jsonData["logLevelField"].(string)` at `elasticsearch.go:149-152` | — |
| `jsonData_dataLinks` | `dataLinks` | jsonData | `DataLinks.tsx:36` (`"Data links"`) | Description verbatim (`DataLinks.tsx:39-43`); default is empty array; new entries seeded with `{ field: '', url: '' }` (`DataLinks.tsx:85`) | Frontend-only: consumed by log row viewer; neither the Elasticsearch backend nor any other backend reads this | Item fields: `field` (required, `DataLink.tsx:45-56`), `url` (`DataLink.tsx:69-84`; label switches between "URL" and "Query" based on internal-link toggle), `urlDisplayLabel` (`DataLink.tsx:87-100`), `datasourceUid` (`DataLink.tsx:121-133`) |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `root_url` | `url` | root | URL | Yes (`elasticsearch.go:196,213`) |
| `virtual_authMethod` | — (virtual) | — | Authentication method | — (editor-local state) |
| `root_basicAuth` | `basicAuth` | root | — (managed by `virtual_authMethod`) | Yes (via SDK `HTTPClientOptions`) |
| `root_basicAuthUser` | `basicAuthUser` | root | User | Yes (via SDK) |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | secureJsonData | Password | Yes (via SDK) |
| `jsonData_oauthPassThru` | `oauthPassThru` | jsonData | — (managed by `virtual_authMethod`) | Yes (via SDK) |
| `jsonData_apiKeyAuth` | `apiKeyAuth` | jsonData | — (managed by `virtual_authMethod`) | **Yes** (`elasticsearch.go:125`) |
| `secureJsonData_apiKey` | `apiKey` | secureJsonData | API Key | Yes (`elasticsearch.go:127`) |
| `jsonData_sigV4Auth` | `sigV4Auth` | jsonData | — (managed by `virtual_authMethod`) | Yes (via SDK; service forced to `"es"` at `elasticsearch.go:120-123`) |
| `jsonData_tlsAuth` | `tlsAuth` | jsonData | TLS Client Authentication | Yes (via SDK) |
| `jsonData_serverName` | `serverName` | jsonData | ServerName | Yes (via SDK) |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | secureJsonData | Client Certificate | Yes (via SDK) |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | secureJsonData | Client Key | Yes (via SDK) |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | jsonData | Add self-signed certificate | Yes (via SDK) |
| `secureJsonData_tlsCACert` | `tlsCACert` | secureJsonData | CA Certificate | Yes (via SDK) |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | jsonData | Skip TLS certificate validation | Yes (via SDK) |
| `jsonData_keepCookies` | `keepCookies` | jsonData | Allowed cookies | Yes (via SDK) |
| `jsonData_timeout` | `timeout` | jsonData | Timeout | Yes (via SDK) |
| `jsonData_index` | `index` | jsonData | Index name | Yes (`elasticsearch.go:164-170`) |
| `jsonData_interval` | `interval` | jsonData | Pattern | Yes (`elasticsearch.go:159-162`) |
| `jsonData_timeField` | `timeField` | jsonData | Time field name | Yes (`elasticsearch.go:140-147` — hard-required) |
| `jsonData_maxConcurrentShardRequests` | `maxConcurrentShardRequests` | jsonData | Max concurrent Shard Requests | Yes (`elasticsearch.go:172-189`) |
| `jsonData_timeInterval` | `timeInterval` | jsonData | Min time interval | **No — frontend-only** (see below) |
| `jsonData_includeFrozen` | `includeFrozen` | jsonData | Include Frozen Indices | Yes (`elasticsearch.go:191-194`) |
| `jsonData_defaultQueryMode` | `defaultQueryMode` | jsonData | Default query mode | **No — frontend-only** (query editor default only) |
| `jsonData_logMessageField` | `logMessageField` | jsonData | Message field name | Yes (`elasticsearch.go:154-157`) |
| `jsonData_logLevelField` | `logLevelField` | jsonData | Level field name | Yes (`elasticsearch.go:149-152`) |
| `jsonData_dataLinks` | `dataLinks` | jsonData | Data links | **No — frontend-only** |

### Frontend-only settings

- **`defaultQueryMode`** picks the query editor's default query type
  (`datasource.ts:135, 164, 167`) but is never used by the backend's msearch or
  ES|QL execution paths. Provisioning a value here changes only the editor.
- **`timeInterval`** is the "Min time interval" duration. It is a Grafana
  panel-plugin knob passed via `panelPluginJsonData` to the query engine's
  `Interval` calculation and does not flow through `NewDatasource`.
- **`dataLinks`** is consumed exclusively by the frontend log-row viewer to
  render buttons on matching fields; no backend code path reads it.

### Backend-only settings

None. Every jsonData key the backend reads (`apiKeyAuth`, `timeField`,
`logLevelField`, `logMessageField`, `interval`, `index`, `maxConcurrentShardRequests`,
`includeFrozen`) has a corresponding editor field.

### Fields deliberately excluded

- **`jsonData.enableSecureSocksProxy`** — written by `@grafana/ui`'s
  `SecureSocksProxySettings`, rendered conditionally at
  `ConfigEditor.tsx:103-105`. Per repo `AGENTS.md`, this field is deliberately
  omitted from every registry entry.
- **`sigV4*` jsonData fields and their `sigV4AccessKey` / `sigV4SecretKey`
  secrets** — contributed by `@grafana/aws-sdk`'s `SIGV4ConnectionConfig` (not
  by this plugin) and only offered by the editor when the Grafana instance has
  `sigV4AuthEnabled`. The `sigV4Auth` discriminator boolean is modeled because
  it is a plugin-owned discriminator; the companion fields are external and out
  of scope for this schema.
- **Dynamic `httpHeaderName<N>` (jsonData) / `httpHeaderValue<N>`
  (secureJsonData) pairs** — written by `@grafana/plugin-ui`'s `CustomHeaders`
  component when the user configures custom HTTP headers. Their keys are
  dynamically indexed, so they are not modeled as first-class fields; see
  `settings.ts` and `settings.go` for the note.

## Where the types are defined

The configuration types are spread across the plugin and its dependencies — some fields
and base types come from libraries/SDKs rather than the plugin itself:

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `ElasticsearchOptions` (jsonData), `Interval`, `QueryType`, `ElasticsearchSecureJsonData`, `DataLinkConfig` | `src/types.ts:60-140` | plugin ([grafana/grafana-elasticsearch-datasource](https://github.com/grafana/grafana-elasticsearch-datasource)) |
| `DataSourceJsonData` (base interface `ElasticsearchOptions` extends) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `13.1.0` |
| `DataSourcePluginOptionsEditorProps`, `onUpdateDatasourceSecureJsonDataOption`, `updateDatasourcePluginResetOption` | `packages/grafana-data/src/` | `@grafana/data` `13.1.0` |
| `AuthMethod` (`BasicAuth`, `OAuthForward`, `NoAuth`, `CrossSiteCredentials`) and `CustomMethodId` | `packages/plugin-ui/src/components/ConfigEditor/Auth/types.ts` | `@grafana/plugin-ui` `0.13.1` |
| `convertLegacyAuthProps`, `getSelectedMethod`, `getOnAuthMethodSelectHandler`, `getBasicAuthProps`, `getTLSProps`, `getCustomHeaders` | `packages/plugin-ui/src/components/ConfigEditor/Auth/utils.ts` | `@grafana/plugin-ui` `0.13.1` |
| `ConnectionSettings`, `Auth`, `AuthMethodSettings`, `BasicAuth`, `TLSSettings`, `SelfSignedCertificate`, `TLSClientAuth`, `SkipTLSVerification`, `AdvancedHttpSettings`, `DataSourceDescription`, `ConfigSection`, `ConfigSubSection`, `ConfigDescriptionLink`, `CustomHeaders` | `packages/plugin-ui/src/components/ConfigEditor/*` | `@grafana/plugin-ui` `0.13.1` |
| `SecureSocksProxyConfig` / `enableSecureSocksProxy` jsonData field (excluded from this entry) | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | `@grafana/ui` `13.1.0` |
| `SIGV4ConnectionConfig` — contributes `sigV4Auth`, `sigV4AuthType`, `sigV4Region`, `sigV4Profile`, `sigV4AssumeRoleArn`, `sigV4ExternalId` jsonData and `sigV4AccessKey` / `sigV4SecretKey` secrets | `@grafana/aws-sdk` `pkg/sigV4/` | `@grafana/aws-sdk` `0.10.0` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| No plugin-owned settings struct — `NewDatasource` reads jsonData as `map[string]any` | `pkg/elasticsearch/elasticsearch.go:106-243` | plugin |
| `client.DatasourceInfo`, `client.ConfiguredFields` — the internal, post-parsing shape the ES client uses | `pkg/elasticsearch/client/client.go:30-46` | plugin |
| `backend.DataSourceInstanceSettings` (carries `URL`, `Database`, `BasicAuthEnabled`, `BasicAuthUser`, `JSONData`, `DecryptedSecureJSONData`) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.294.0` |
| `settings.HTTPClientOptions(ctx)` — reads TLS, timeout, cookies, basic auth, SigV4 config | `backend/httpclient` | `github.com/grafana/grafana-plugin-sdk-go` |
| SigV4 middleware namespace override (`Service = "es"`), attaches on top of the SDK's SigV4 transport | `pkg/awsauth` in `github.com/grafana/grafana-aws-sdk` | `github.com/grafana/grafana-aws-sdk` |

The models in this entry flatten that spread into a single Go `Config` type (root
fields tagged `json:"-"` + jsonData fields + `DecryptedSecureJSONData`) plus a
`SecureJsonDataKey` typed constant list. `settings.ts` keeps the three canonical
TypeScript types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`). Auth-method
constants (Interval, QueryType, SecureJsonDataKey) mirror the plugin's frontend
sources; there are no backend equivalents to sync to.

## Modeling decisions

- **Virtual auth-method selector**: `onAuthMethodSelect` (`ConfigEditor.tsx:80-92`)
  writes ALL five discriminators on every selection (basicAuth, apiKeyAuth, sigV4Auth,
  oauthPassThru, withCredentials) so exactly one is true at a time. This is captured
  as five `effects` on the virtual `virtual_authMethod` field, with the driven storage
  fields tagged `managed-by:virtual_authMethod`. The load-time `read` expression
  mirrors `getSelectedMethod` (`@grafana/plugin-ui/utils.js:24-35`) with the two custom
  method IDs (`custom-api-key`, `custom-sigv4`) checked first, matching
  `ConfigEditor.tsx:48,60`.
- **SigV4 companion fields not modeled**: `sigV4AuthType`, `sigV4Region`, etc. are
  contributed by `@grafana/aws-sdk`'s `SIGV4ConnectionConfig`, not by this plugin.
  Only the plugin-owned `sigV4Auth` boolean discriminator is modeled; provisioning
  payloads that use SigV4 should include the companion fields directly in jsonData
  (see the `sigV4` settings example) but they are out of scope for the schema.
- **`requiredWhen` for `root_url` and `jsonData_index`**: the editor marks the URL
  and Index inputs as `required` in the UI. The backend does not explicitly check the
  URL, but the health check parses it and any msearch will fail without one; the
  Index has a legacy fallback to `settings.Database` (`elasticsearch.go:164-170`), so
  callers assembling a `Config` directly may satisfy the requirement by pre-filling
  either. `requiredWhen: "true"` reflects the editor's data contract.
- **`timeField` defaults to `@timestamp`**: `coerceOptions` writes this on mount
  (`utils.ts:21`); the backend hard-fails on an empty timeField
  (`elasticsearch.go:145-147`). We apply the default in `ApplyDefaults` for editor
  parity and enforce presence in `Validate`.
- **`maxConcurrentShardRequests` string-or-number normalization**: the backend
  supports the field as a JSON number (float64 on the wire) OR a JSON string
  (`elasticsearch.go:172-185`). `Config.UnmarshalJSON` mirrors that: try number,
  then string, then fall through to `ApplyDefaults`' non-positive coercion to 5.
- **`interval` uses `""` for "No pattern"**: `indexPatternTypes[0]` has
  `value: 'none'` in the editor (`ElasticDetails.tsx:12`), but the change handler
  `intervalHandler` converts `'none'` to `undefined` before storing
  (`ElasticDetails.tsx:200-201`). We model that in the schema by giving the "No
  pattern" option `value: ""` (empty string), which matches the wire format the
  backend reads.
- **`dataLinks.url` label mirrors editor duality**: the editor renders "URL" when
  `datasourceUid` is unset and "Query" when it is set (`DataLink.tsx:71-72`). We
  record `"URL/Query"` as the schema label to reflect both states; consumers should
  treat the string as a query when `datasourceUid` is non-empty.
- **Browser access mode not modeled**: `access: 'direct'` triggers a persistent
  error Alert (`ConfigEditor.tsx:65-69`) and no request will succeed. We keep
  `access` out of the schema entirely because there is no editor field for it and
  provisioning must always use `'proxy'` (the SDK default). The finding is recorded
  under [Upstream findings](#upstream-findings).
- **Secure Socks Proxy excluded**: the editor conditionally renders
  `SecureSocksProxySettings` (`ConfigEditor.tsx:103-105`) writing
  `jsonData.enableSecureSocksProxy` when the Grafana instance has
  `secureSocksDSProxyEnabled`. Per repo `AGENTS.md`, the field is deliberately
  omitted from this registry entry.
- **Field ID naming convention**: IDs are prefixed with their storage target
  (`root_`, `jsonData_`, `secureJsonData_`, `virtual_`) followed by the camelCase
  storage key. The `key` property keeps the plugin's raw storage key.
- **Flat `Config` in Go**: `settings.go` collapses root fields (`URL`, `BasicAuth`,
  `BasicAuthUser`, `WithCredentials`, `Database`, all tagged `json:"-"`) and jsonData
  fields onto a single `Config` struct with the same json tags the backend reads via
  its `map[string]any` unmarshal. `settings.ts` keeps the three canonical TS types.
- **`SecureJsonDataConfig` is a key list**: secure values are write-only, so the
  secure type is just the array of secret key names (`basicAuthPassword`, `apiKey`,
  `tlsCACert`, `tlsClientCert`, `tlsClientKey`); consumers read `secureJsonFields`
  to see what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema`
bundle (the k8s-style schema Grafana's datasource API server serves as
`{apiVersion}.json`, `v0alpha1` today) from the embedded `dsconfig.json`: root
fields plus a nested `jsonData` object become the OpenAPI settings `spec`,
secure fields become `secureValues`, and virtual fields are skipped.

`SettingsExamples()` provides the default configuration plus one k8s-style example
per authentication method and connection variant:

| Example | Auth | Notes | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | NoAuth | Editor defaults with a placeholder index | `basicAuthPassword` (empty) |
| `noAuth` | NoAuth | Daily logstash-style time-pattern index | `basicAuthPassword` (empty) |
| `basicAuth` | BasicAuth | Root-level `basicAuth`/`basicAuthUser`; password is secret | `basicAuthPassword` |
| `apiKey` | API Key | Sets `jsonData.apiKeyAuth=true`; logs-mode query default | `apiKey` |
| `sigV4` | SigV4 | AWS OpenSearch/managed ES; SigV4 companion fields in jsonData | `basicAuthPassword` (empty) |
| `oauthForward` | Forward OAuth Identity | Sets `jsonData.oauthPassThru=true`; no secret | `basicAuthPassword` (empty) |
| `tlsMutualAuth` | mTLS | Client cert + key + serverName; PAT unset | `tlsClientCert`, `tlsClientKey` |
| `tlsSelfSignedCA` | CA verification | `tlsAuthWithCACert=true` | `tlsCACert` |
| `logsWithDataLinks` | BasicAuth | Logs mode + two data link entries (external + internal) | `basicAuthPassword` |
| `legacyDatabaseFallback` | NoAuth | Index in root.database (legacy shape) | `basicAuthPassword` (empty) |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns
a fully-defaulted, validated `Config`:

1. **Parse** — copy root fields (`URL`, `BasicAuth`, `BasicAuthUser`, `Database`)
   from `settings`, unmarshal jsonData into `Config` (with the custom
   `UnmarshalJSON` that accepts `maxConcurrentShardRequests` as JSON number OR
   string), promote a legacy `settings.Database` to `Index` when `jsonData.index`
   is empty (`elasticsearch.go:164-170`), and copy decrypted secrets by known key
   into `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — write the curated set of editor-parity defaults
   (`TimeField='@timestamp'`, `MaxConcurrentShardRequests=5`,
   `DefaultQueryMode=metrics`) for zero-valued fields only.
3. **`Validate`** — enforce the runtime contract: URL required, index required,
   timeField required, valid interval / defaultQueryMode enums, per-auth-method
   secret requirements, TLS field pairs, non-negative timeout /
   maxConcurrentShardRequests. Errors are joined so every problem surfaces at
   once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels so log lines carry request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are exported for
callers that want to compose them themselves (provisioning preview, schema-example
round-trip, tests that distinguish parse-level from policy-level errors). Skip
them by never calling `LoadConfig` in those flows — assemble a `Config` directly.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching
upstream. All preserved verbatim in the schema — the schema records what the plugin
**does**, not what it **should** do; these notes exist so reviewers can reproduce
each finding and decide separately whether to fix upstream.

1. **`Client Key` placeholder says `RSA PRIVATE KEY CERTIFICATE`** (@grafana/plugin-ui
   `TLSClientAuth.js:104`) — an RSA private key PEM header is
   `-----BEGIN RSA PRIVATE KEY-----`, and a certificate is a different artefact.
   Preserved verbatim in the schema to match the UI.
2. **Custom method IDs are hard-coded strings, not enum members.** `AuthMethod`
   (`@grafana/plugin-ui/types.js:3-8`) only exposes `NoAuth`, `BasicAuth`,
   `OAuthForward`, `CrossSiteCredentials`; the API Key and SigV4 methods use
   free-form IDs `'custom-api-key'` and `'custom-sigv4'` in `ConfigEditor.tsx:42,54`.
   Renaming an ID upstream without coordinating would silently break provisioned
   datasources.
3. **SigV4 fields are contributed by an external SDK, not by the plugin.** All
   `sigV4*` jsonData fields (and their `sigV4AccessKey` / `sigV4SecretKey` secrets)
   come from `@grafana/aws-sdk`'s `SIGV4ConnectionConfig` (`ConfigEditor.tsx:57`).
   The plugin itself only stores the discriminator `jsonData.sigV4Auth`. Provisioning
   payloads must include the companion fields directly.
4. **`onAuthMethodSelect` also touches `withCredentials`.** The editor's handler
   sets `withCredentials = method === AuthMethod.CrossSiteCredentials`
   (`ConfigEditor.tsx:84`) even though the default `visibleMethods` from
   `@grafana/plugin-ui/AuthMethodSettings.js:47-51` does not include
   `CrossSiteCredentials`. Selecting any of the five visible methods will thus
   overwrite a legacy `withCredentials: true` to `false`.
5. **`jsonData.timeInterval` regex is editor-only.** The invalid-state check at
   `ElasticDetails.tsx:110` uses `^\d+(ms|[Mwdhmsy])$` to flag bad inputs in the
   editor, but the backend never validates this string — invalid values silently
   pass through to Grafana's `TimeInterval` calculation. Provisioning malformed
   values will fail at query time, not at load time.
6. **`intervalHandler` auto-modifies the index name.** When the user changes the
   Pattern selector and the current index is empty or a legacy `[logstash-]`
   template, the handler overwrites the Index input with the pattern's example
   template (`ElasticDetails.tsx:200-212`). This can silently discard a legitimate
   configuration if the user only wanted to switch off patterning.
7. **`coerceOptions` runs on every render, not once.** The editor's `useEffect`
   at `ConfigEditor.tsx:30-34` re-invokes `coerceOptions` whenever any option
   changes and `isValidOptions` is false. Values only written by `coerceOptions`
   (empty `logMessageField`/`logLevelField`, default `defaultQueryMode`) reappear
   on save whenever a user clears them, because `isValidOptions` treats
   `undefined` as invalid.
8. **Browser access mode has no code path but is not blocked.** `ConfigEditor.tsx:65-69`
   only renders an Alert; nothing prevents provisioning a datasource with
   `access: 'direct'`. The msearch/ES|QL client paths do not check `access`; they
   fail later because the Grafana proxy machinery is bypassed. Provisioning tools
   should reject `access: 'direct'` on their side.
9. **`maxConcurrentShardRequests` accepted as a string is undocumented.** The
   editor writes it as a string (Input's `value` -> `currentTarget.value`,
   `ElasticDetails.tsx:94`), the backend accepts both number and string
   (`elasticsearch.go:174-185`), but provisioning docs only show the numeric form.
10. **Backend silently drops `interval` when it is not a string.** The type
    assertion at `elasticsearch.go:159-162` coerces the field to `""` if it is
    any non-string JSON value, without warning. A provisioned datasource with a
    numeric `interval` would silently behave as "No pattern".

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this
  repo) — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft 2020-12, `additionalProperties: false`) — passes.
- `go test ./...` on this module — passes (schema bundle shape, secure values,
  examples, `LoadConfig` incl. legacy database fallback and string-or-number
  maxConcurrentShardRequests, `SchemaArtifactInSync` guard).
- `settings.go` / `schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: to be validated with `tsc --noEmit --strict`.
