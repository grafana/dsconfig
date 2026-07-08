# grafana-opensearch-datasource

Declarative configuration schema for the [OpenSearch datasource plugin](https://github.com/grafana/opensearch-datasource) (`grafana-opensearch-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/opensearch-datasource`
- **Ref**: `main`
- **Commit SHA**: `6881bb4218ee9924f8ca6330c5b558b913ab5f19` (2026-07-04, `docs: add signed commits requirement to CONTRIBUTING.md (#1124)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, option
labels/values, section titles, defaults, validations, dependency and required-when
expressions, storage keys, storage targets, value types, group titles, and instructions
— is traceable to a specific `file:line` in the upstream repo at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone https://github.com/grafana/opensearch-datasource
cd opensearch-datasource
git checkout 6881bb4218ee9924f8ca6330c5b558b913ab5f19
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under
[Sources researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and instructions |
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
(`6881bb4218ee9924f8ca6330c5b558b913ab5f19`), plus external editor components at the
exact versions the plugin's `package.json` pins.

### Plugin repo (`github.com/grafana/opensearch-datasource@6881bb4`)

| File | What was read |
| --- | --- |
| `src/plugin.json:1-41` | `pluginType` (`id`=`grafana-opensearch-datasource`), `pluginName` (`name`=`OpenSearch`), `docURL` (`info.links[0].url`=`https://github.com/grafana/opensearch-datasource`) |
| `src/configuration/ConfigEditor.tsx:1-112` | Outer editor; `DataSourceHttpSettings` invocation (defaultUrl=`http://localhost:9200`, `showAccessOptions=true`, `sigV4AuthToggleEnabled` gated on `config.sigV4AuthEnabled`, `renderSigV4Editor=<SIGV4ConnectionConfig ...>`), conditional `SecureSocksProxySettings` (excluded), `coerceOptions` invocation, `OpenSearchDetails`, `LogsConfig`, `DataLinks` |
| `src/configuration/OpenSearchDetails.tsx:1-327` | Index name field (`jsonData.database`, placeholder `es-index-name`, required), Pattern (`indexPatternTypes` at :10-17), Time field name (required, no explicit default in the FormField), Serverless toggle (`getServerlessSettings` at :66-80 hard-codes flavor=`opensearch`, version=`1.0.0`, maxConcurrentShardRequests=5, pplEnabled=true), Version disabled input + `Get Version and Save` button, Max concurrent Shard Requests (`shouldRenderMaxConcurrentShardRequests` gates on version/flavor at :286-301), Min time interval (validation regex `^\d+(ms|[Mwdhmsy])$` with the exact message string), PPL enabled toggle (default true) |
| `src/configuration/OpenSearchDetails.tsx:251-284` | `intervalHandler`: reads/writes `root.database` (NOT `jsonData.database`) — a latent upstream inconsistency; the backend index-pattern generator reads `jsonData.database` at `client.go:120` |
| `src/configuration/OpenSearchDetails.tsx:322-327` | `defaultMaxConcurrentShardRequests(flavor, version)` = 256 for Elasticsearch <7.0.0, 5 otherwise |
| `src/configuration/LogsConfig.tsx:1-48` | Message field name (placeholder `_source`), Level field name (no placeholder), no tooltip on either |
| `src/configuration/DataLinks.tsx:1-84` | Section title `Data links`, description verbatim, seed `{ field: '', url: '' }` for new entries |
| `src/configuration/DataLink.tsx:1-159` | Field input (tooltip `Can be exact field name or a regex pattern that will match on the field name.`), Title input (labelWidth=6, no placeholder), URL/Query switch (label depends on `showInternalLink`), Internal link switch drives `datasourceUid` |
| `src/configuration/utils.ts:6-48` | `coerceOptions` — writes `timeField='@timestamp'`, `maxConcurrentShardRequests` via `defaultMaxConcurrentShardRequests(flavor, version)`, `logMessageField=''`, `logLevelField=''`, `pplEnabled=true` when unset; `isValidOptions` triggers re-coercion on every render whenever any invariant is violated |
| `src/configuration/utils.ts:56-99` | `AVAILABLE_VERSIONS` map (OpenSearch 1.0.x, Elasticsearch 7.0+/6.0+/5.6+/5.0+/2.0+) — documentation only; the editor never renders a version picker (users hit "Get Version and Save" instead) |
| `src/types.ts:12-28` | `OpenSearchOptions` (jsonData shape) with `database`, `timeField`, `version`, `flavor`, `versionLabel`, `interval`, `timeInterval`, `maxConcurrentShardRequests`, `logMessageField`, `logLevelField`, `dataLinks`, `pplEnabled`, `sigV4Auth`, `serverless`, `enableSecureSocksProxy` |
| `src/types.ts:101-106` | `DataLinkConfig` = `{ field: string; url: string; datasourceUid?: string; title?: string }` — the OpenSearch shape uses `title` (not `urlDisplayLabel`) |
| `src/types.ts:113-116` | `Flavor` enum: `elasticsearch` / `opensearch` |
| `pkg/opensearch/opensearch.go:28-39` | `NewOpenSearchDatasource` — delegates to `client.NewDatasourceHttpClient`; no upstream `pkg/models/settings.go` exists |
| `pkg/opensearch/opensearch.go:45-174` | `CheckHealth` — hard-fails on missing flavor/version ("No version set") or missing timeField ("time field name is required"); uses `jsonData.database` + `jsonData.interval` to build the health-check index pattern |
| `pkg/opensearch/client/client.go:30-63` | `NewDatasourceHttpClient` — unmarshals `serverless`+`oauthPassThru` into an anonymous struct; wires `ForwardHTTPHeaders=true` when `oauthPassThru=true`; forces `httpCliOpts.SigV4.Service='es'` (or `'aoss'` when serverless is true) whenever the SDK builds a SigV4 transport |
| `pkg/opensearch/client/client.go:96-153` | `NewClient` — reads `jsonData.version` (required), `jsonData.flavor` (defaults to `opensearch`), `jsonData.timeField` (required), `jsonData.logLevelField`, `jsonData.logMessageField` (defaults to `_source`), `jsonData.database`, `jsonData.interval` |
| `pkg/opensearch/client/client.go:288-302,520-534` | Per-request Basic-auth wiring: `secureJsonData.basicAuthPassword` when `basicAuth=true`; legacy `secureJsonData.password` when `basicAuth=false` and `settings.User != ""` |
| `pkg/opensearch/client/client.go:300-302,532-534` | Adds `x-amz-content-sha256` on non-GET requests when `jsonData.serverless=true` |
| `pkg/opensearch/client/client.go:411-433` | Backend defaults for `maxConcurrentShardRequests`: 256 for Elasticsearch [5.6.0, 7.0.0), 5 for OpenSearch and Elasticsearch >=7.0.0 |
| `pkg/opensearch/client/client.go:557-560` | PPL endpoint choice: `/_plugins/_ppl` for OpenSearch, `/_opendistro/_ppl` for Elasticsearch; the choice is by flavor, not by `pplEnabled` |
| `pkg/opensearch/client/models.go:13-18` | `Flavor` string constants (`elasticsearch`, `opensearch`) matching the frontend enum |
| `package.json` | External component versions (see next table) |
| `go.mod` | Backend dep versions (`github.com/grafana/grafana-plugin-sdk-go v0.291.1`, `github.com/grafana/grafana-aws-sdk v1.4.3`, `github.com/Masterminds/semver v1.5.0`) |

### External editor components

Read at the exact versions pinned in the plugin's `package.json`
(`@grafana/aws-sdk@0.10.2`, `@grafana/data@12.4.2`, `@grafana/plugin-ui@0.13.1`,
`@grafana/runtime@12.4.2`, `@grafana/schema@12.4.2`, `@grafana/ui@12.4.2`,
`semver@7.7.4`).

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `DataSourceHttpSettings` | `@grafana/ui@12.4.2` | `dist/esm/components/DataSourceSettings/DataSourceHttpSettings.mjs` | Section title (`HTTP`, `Auth`), URL Field label (`URL`, override via `urlLabel` — not used here), default URL tooltip and placeholder (overridden here to `http://localhost:9200`), Access radio (`Server (default)` / `Browser`), Access help copy, Allowed cookies (`Grafana proxy deletes forwarded cookies by default. Specify cookies by name that should be forwarded to the data source.`), Timeout (`HTTP request timeout in seconds`, placeholder `Timeout in seconds`, parsed via `parseInt`), Basic auth toggle (`Basic auth`), With Credentials toggle (`Whether credentials such as cookies or auth headers should be sent with cross-site requests.`), SigV4 auth toggle (`SigV4 auth`, gated on `sigV4AuthToggleEnabled`) |
| `BasicAuthSettings` | `@grafana/ui@12.4.2` | `dist/esm/components/DataSourceSettings/BasicAuthSettings.mjs` | User Field label (`User`, placeholder `user`), Password Field (via SecretFormField, label `Password`, placeholder `Password` from SecretFormField defaults) |
| `HttpProxySettings` | `@grafana/ui@12.4.2` | `dist/esm/components/DataSourceSettings/HttpProxySettings.mjs` | TLS Client Auth toggle (`TLS Client Auth`), With CA Cert toggle (`With CA Cert`, tooltip `Needed for verifying self-signed TLS Certs`), Skip TLS Verify toggle (`Skip TLS Verify`), Forward OAuth Identity toggle (`Forward OAuth Identity`, tooltip `Forward the user's upstream OAuth identity to the data source (Their access token gets passed along).`) |
| `TLSAuthSettings` | `@grafana/ui@12.4.2` | `dist/esm/components/DataSourceSettings/TLSAuthSettings.mjs` | Section title (`TLS/SSL Auth Details`), ServerName Field (label `ServerName`, placeholder `domain.example.com`), CA Cert (label `CA Cert`, placeholder `Begins with -----BEGIN CERTIFICATE-----`), Client Cert (label `Client Cert`, placeholder `Begins with -----BEGIN CERTIFICATE-----`), Client Key (label `Client Key`, placeholder `Begins with -----BEGIN RSA PRIVATE KEY-----`) |
| `CustomHeadersSettings` (dynamic keys, excluded) | `@grafana/ui@12.4.2` | `dist/esm/components/DataSourceSettings/CustomHeadersSettings.mjs` | Renders when `access='proxy'`; writes indexed `jsonData.httpHeaderName<N>` and `secureJsonData.httpHeaderValue<N>` pairs. Excluded because keys are dynamically indexed |
| `SIGV4ConnectionConfig` (contributed only, not modeled) | `@grafana/aws-sdk@0.10.2` | Read via the plugin's usage in `src/configuration/ConfigEditor.tsx:60-61` | Contributes `sigV4Auth`, `sigV4AuthType`, `sigV4Region`, `sigV4Profile`, `sigV4AssumeRoleArn`, `sigV4ExternalId` (jsonData) and `sigV4AccessKey`, `sigV4SecretKey` (secureJsonData) when the Grafana instance has `config.sigV4AuthEnabled` |
| `SecureSocksProxySettings` (excluded) | `@grafana/ui@12.4.2` | Rendered conditionally at `ConfigEditor.tsx:64-66`. Storage key `jsonData.enableSecureSocksProxy` deliberately excluded per repo `AGENTS.md` |
| `LegacyForms.FormField`, `LegacyForms.Input`, `LegacyForms.Select`, `LegacyForms.Switch`, `Alert`, `Button`, `VerticalGroup`, `DataLinkInput` | `@grafana/ui@12.4.2` | Prop names (`label`, `labelWidth`, `inputWidth`, `placeholder`, `required`, `value`, `onChange`, `tooltip`, `disabled`) so we knew which UI attributes to record |
| `DataSourcePluginOptionsEditorProps`, `DataSourceSettings`, `DataLinkBuiltInVars`, `VariableOrigin` | `@grafana/data@12.4.2` | Editor prop shape and data-link variable suggestions surfaced by `DataLinks.tsx` |
| `config.sigV4AuthEnabled`, `config.secureSocksDSProxyEnabled`, `getBackendSrv`, `getDataSourceSrv` | `@grafana/runtime@12.4.2` | Feature gates for the conditional SigV4 toggle and secure-socks proxy render; datasource lookup for `useDatasource` |
| `semver@7.7.4` | `semver@7.7.4` | `valid`, `gte`, `lt` — validate `jsonData.version` and drive `defaultMaxConcurrentShardRequests`; the backend uses `github.com/Masterminds/semver v1.5.0` for the same checks |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each
of its label, placeholder, tooltip, default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `root_url` | `url` | root | @grafana/ui `DataSourceHttpSettings.mjs:143` (`URL`) | Placeholder from `defaultUrl` prop = `http://localhost:9200` (`ConfigEditor.tsx:56`) | Backend reads `settings.URL` at `opensearch.go:99, client.go:245,479` | Role `endpoint.baseUrl`; `requiredWhen: "true"` because the health check parses it and every msearch/PPL request builds URLs from it |
| `root_access` | `access` | root | `DataSourceHttpSettings.mjs:178` (`Access`) | Options `Server (default)`/`Browser` (`DataSourceHttpSettings.mjs:69-79`); default `proxy` from `DEFAULT_ACCESS_OPTION` (`DataSourceHttpSettings.mjs:81`) | @grafana/ui `DataSourceHttpSettings`; SDK does not surface it on `backend.DataSourceInstanceSettings` (documented; not modeled on `Config`) | Not read by backend but gates editor sections at `DataSourceHttpSettings.mjs:204,341,357` |
| `root_basicAuth` | `basicAuth` | root | `DataSourceHttpSettings.mjs:259` (`Basic auth`) | Default `false` | `settings.BasicAuthEnabled` on the SDK settings struct | Role `auth.basic.enabled` |
| `root_basicAuthUser` | `basicAuthUser` | root | @grafana/ui `BasicAuthSettings.mjs:36` (`User`) | Placeholder `user` (`BasicAuthSettings.mjs:39`) | `settings.BasicAuthUser` | Role `auth.basic.username`; `requiredWhen: "root_basicAuth == true"` |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | secureJsonData | `SecretFormField` default label (`Password`) invoked by `BasicAuthSettings.mjs:44-54` | Placeholder `Password` (SecretFormField default) | Backend `secureJsonData["basicAuthPassword"]` at `client.go:290-291` | Role `auth.basic.password`; `requiredWhen: "root_basicAuth == true"` |
| `root_withCredentials` | `withCredentials` | root | `DataSourceHttpSettings.mjs:277` (`With Credentials`) | Tooltip `Whether credentials such as cookies or auth headers should be sent with cross-site requests.` (`DataSourceHttpSettings.mjs:279-281`); default `false` | @grafana/ui `DataSourceHttpSettings`; SDK does not surface it — browser-only | — |
| `jsonData_oauthPassThru` | `oauthPassThru` | jsonData | @grafana/ui `HttpProxySettings.mjs:75` (`Forward OAuth Identity`) | Tooltip `Forward the user's upstream OAuth identity to the data source (Their access token gets passed along).` (`HttpProxySettings.mjs:76-79`); default `false`; only shown when `access='proxy'` | Backend `client.go:33` reads it and sets `httpCliOpts.ForwardHTTPHeaders=true` | Role `auth.forwardOAuthToken.enabled`; `dependsOn: "root_access == 'proxy'"` |
| `jsonData_sigV4Auth` | `sigV4Auth` | jsonData | @grafana/ui `DataSourceHttpSettings.mjs:323` (`SigV4 auth`) | Default `false`; only offered when `config.sigV4AuthEnabled` is true (`ConfigEditor.tsx:60`) | Backend `client.go:49-55` — presence of `httpCliOpts.SigV4 != nil` triggers the SigV4 middleware and service-namespace override | Role `auth.awsSigV4.enabled`; SigV4 companion fields are contributed by `@grafana/aws-sdk`'s `SIGV4ConnectionConfig` and not modeled |
| `jsonData_tlsAuth` | `tlsAuth` | jsonData | @grafana/ui `HttpProxySettings.mjs:22` (`TLS Client Auth`) | Default `false` | @grafana/ui `HttpProxySettings`; consumed by SDK HTTPClient | — |
| `jsonData_serverName` | `serverName` | jsonData | @grafana/ui `TLSAuthSettings.mjs:80` (`ServerName`) | Placeholder `domain.example.com` (`TLSAuthSettings.mjs:83`) | @grafana/ui `TLSAuthSettings`; consumed by SDK HTTPClient | Role `tls.serverName`; `requiredWhen: "jsonData_tlsAuth == true"` |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | secureJsonData | @grafana/ui `TLSAuthSettings.mjs:92` (`Client Cert`) | Placeholder `Begins with -----BEGIN CERTIFICATE-----` (`TLSAuthSettings.mjs:94-97`) | @grafana/ui `TLSAuthSettings`; consumed by SDK HTTPClient | Role `tls.clientCert`; `requiredWhen: "jsonData_tlsAuth == true"` |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | secureJsonData | @grafana/ui `TLSAuthSettings.mjs:106` (`Client Key`) | Placeholder `Begins with -----BEGIN RSA PRIVATE KEY-----` (`TLSAuthSettings.mjs:107-110`) | @grafana/ui `TLSAuthSettings`; consumed by SDK HTTPClient | Role `tls.clientKey`; `requiredWhen: "jsonData_tlsAuth == true"` |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | jsonData | @grafana/ui `HttpProxySettings.mjs:38` (`With CA Cert`) | Tooltip `Needed for verifying self-signed TLS Certs` (`HttpProxySettings.mjs:39-42`); default `false` | @grafana/ui `HttpProxySettings`; consumed by SDK HTTPClient | — |
| `secureJsonData_tlsCACert` | `tlsCACert` | secureJsonData | @grafana/ui `TLSAuthSettings.mjs:72` (`CA Cert`) | Placeholder `Begins with -----BEGIN CERTIFICATE-----` (`TLSAuthSettings.mjs:67-71`) | @grafana/ui `TLSAuthSettings`; consumed by SDK HTTPClient | Role `tls.caCert`; `requiredWhen: "jsonData_tlsAuthWithCACert == true"` |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | jsonData | @grafana/ui `HttpProxySettings.mjs:59` (`Skip TLS Verify`) | Default `false` | @grafana/ui `HttpProxySettings`; consumed by SDK HTTPClient | Role `transport.tlsSkipVerify` |
| `jsonData_keepCookies` | `keepCookies` | jsonData | @grafana/ui `DataSourceHttpSettings.mjs:208` (`Allowed cookies`) | Description verbatim (`DataSourceHttpSettings.mjs:210-212`); only shown when `access='proxy'` | @grafana/ui `DataSourceHttpSettings` (TagsInput); consumed by SDK HTTPClient | `dependsOn: "root_access == 'proxy'"` |
| `jsonData_timeout` | `timeout` | jsonData | @grafana/ui `DataSourceHttpSettings.mjs:227` (`Timeout`) | Placeholder `Timeout in seconds` (`DataSourceHttpSettings.mjs:238`); description `HTTP request timeout in seconds` (`DataSourceHttpSettings.mjs:229-231`); parsed via `parseInt` (`DataSourceHttpSettings.mjs:242`); only shown when `access='proxy'` | @grafana/ui `DataSourceHttpSettings`; consumed by SDK HTTPClient | Role `transport.timeoutSeconds`; `dependsOn: "root_access == 'proxy'"` |
| `jsonData_database` | `database` | jsonData | `OpenSearchDetails.tsx:108` (`Index name`) | Placeholder `es-index-name` (`OpenSearchDetails.tsx:111`); `required` (`OpenSearchDetails.tsx:112`); no tooltip | Backend reads `jsonData.database` at `opensearch.go:77, client.go:120-124` | Editor also erroneously writes `root.database` via `intervalHandler` — that field is a dead echo (see [Upstream findings](#upstream-findings)) |
| `jsonData_interval` | `interval` | jsonData | `OpenSearchDetails.tsx:119` (`Pattern`) | Options from `indexPatternTypes` at `OpenSearchDetails.tsx:10-17`; "No pattern" maps to `undefined` on the wire (`OpenSearchDetails.tsx:254`); no tooltip | Backend `client.go:126` — read as a string; empty means "no pattern" | Editor renders "No pattern" as `value: 'none'` but writes `undefined`; the schema uses `value: ""` for parity with what lands in JSON |
| `jsonData_timeField` | `timeField` | jsonData | `OpenSearchDetails.tsx:139` (`Time field name`) | No placeholder; default `@timestamp` from `coerceOptions` (`utils.ts:20`); `required` (`OpenSearchDetails.tsx:142`) | Backend hard-fail on empty timeField at `opensearch.go:70-75, client.go:111-114` | — |
| `jsonData_serverless` | `serverless` | jsonData | `OpenSearchDetails.tsx:147` (`Serverless`) | Tooltip `If this is a DataSource to query a serverless OpenSearch service.` (`OpenSearchDetails.tsx:149`); default `false` | Backend `client.go:32` (unmarshal), `client.go:51-53` (SigV4 service swap), `client.go:300-302,532-534` (x-amz-content-sha256) | — |
| `jsonData_flavor` | `flavor` | jsonData | — (no editor label; set implicitly by the "Get Version and Save" button and by `getServerlessSettings`) | Options `opensearch`/`elasticsearch` from `Flavor` (`src/types.ts:113-116`) | Backend hard-fail at `opensearch.go:56-61` if flavor is neither `opensearch` nor `elasticsearch` | Role `auth.discriminator` — pairs with `jsonData.version`; both are required by the health check |
| `jsonData_version` | `version` | jsonData | `OpenSearchDetails.tsx:161` (`Version`) | Placeholder `version required` (`OpenSearchDetails.tsx:163`); the input is `disabled` (:164) — the value is written by the version-detection button; `required` (:165) | Backend hard-fail at `opensearch.go:63-68, client.go:104-107` when the value is not a valid semver | `dependsOn: "jsonData_serverless != true"` (editor hides Version for serverless) |
| `jsonData_versionLabel` | `versionLabel` | jsonData | — (no editor label; used as the input's display string) | Set by `setVersion` in `OpenSearchDetails.tsx:39-41` | Frontend-only display string | — |
| `jsonData_maxConcurrentShardRequests` | `maxConcurrentShardRequests` | jsonData | `OpenSearchDetails.tsx:177` (`Max concurrent Shard Requests`) | Default per flavor/version via `defaultMaxConcurrentShardRequests` (`OpenSearchDetails.tsx:322-327`); backend also coerces at `client.go:411-433` | Backend supports both JSON number and JSON string; mirrored in `Config.UnmarshalJSON` | Hidden by editor for Elasticsearch <5.6.0 and for serverless (`shouldRenderMaxConcurrentShardRequests` at :286-301); `dependsOn: "jsonData_serverless != true"` |
| `jsonData_timeInterval` | `timeInterval` | jsonData | `OpenSearchDetails.tsx:187` (`Min time interval`) | Placeholder `10s` (`OpenSearchDetails.tsx:193`); tooltip verbatim (`OpenSearchDetails.tsx:204-209`); validation regex `^\d+(ms|[Mwdhmsy])$` with message `Value is not valid, you can use number with time unit specifier: y, M, w, d, h, m, s` (`OpenSearchDetails.tsx:196-200`) | Frontend-only: not consumed by the backend; passed via `panelPluginJsonData` to Grafana's Interval calculation | — |
| `jsonData_pplEnabled` | `pplEnabled` | jsonData | `OpenSearchDetails.tsx:215` (`PPL enabled`) | Tooltip `Allow Piped Processing Language as an alternative query syntax in the OpenSearch query editor.` (`OpenSearchDetails.tsx:217`); default `true` from `coerceOptions` (`utils.ts:27`) and the switch's `?? true` fallback (`OpenSearchDetails.tsx:218`) | Frontend-only: the backend's PPL endpoint choice is by flavor (`client.go:557-560`), not by this flag | — |
| `jsonData_logMessageField` | `logMessageField` | jsonData | `LogsConfig.tsx:30` (`Message field name`) | Placeholder `_source` (`LogsConfig.tsx:33`); no tooltip | Backend `client.go:118` — defaults to `_source` when empty | — |
| `jsonData_logLevelField` | `logLevelField` | jsonData | `LogsConfig.tsx:40` (`Level field name`) | No placeholder; no tooltip | Backend `client.go:116` — empty means "no log level field" | — |
| `jsonData_dataLinks` | `dataLinks` | jsonData | `DataLinks.tsx:28` (`Data links`) | Description verbatim (`DataLinks.tsx:30-32`); default is empty array; new entries seeded with `{ field: '', url: '' }` (`DataLinks.tsx:76`) | Frontend-only: consumed by log row viewer; the backend does not read it | Item fields: `field` (required, `DataLink.tsx:48-58`, tooltip verbatim), `title` (`DataLink.tsx:71-86`), `url` (`DataLink.tsx:87-107`; label switches between `URL` and `Query` based on internal-link toggle; placeholder `http://example.com/${__value.raw}` when external, `${__value.raw}` when internal), `datasourceUid` (`DataLink.tsx:126-138`) |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `root_url` | `url` | root | URL | Yes (`opensearch.go:99, client.go:245,479`) |
| `root_access` | `access` | root | Access | **No — frontend-only** (SDK does not surface it) |
| `root_basicAuth` | `basicAuth` | root | Basic auth | Yes (`client.go:288-292`) |
| `root_basicAuthUser` | `basicAuthUser` | root | User | Yes (`client.go:291`) |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | secureJsonData | Password | Yes (`client.go:290`) |
| `root_withCredentials` | `withCredentials` | root | With Credentials | **No — frontend/browser only** |
| `jsonData_oauthPassThru` | `oauthPassThru` | jsonData | Forward OAuth Identity | Yes (`client.go:33,45-47`) |
| `jsonData_sigV4Auth` | `sigV4Auth` | jsonData | SigV4 auth | Yes (via SDK; service forced to `es`/`aoss` at `client.go:49-55`) |
| `jsonData_tlsAuth` | `tlsAuth` | jsonData | TLS Client Auth | Yes (via SDK) |
| `jsonData_serverName` | `serverName` | jsonData | ServerName | Yes (via SDK) |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | secureJsonData | Client Cert | Yes (via SDK) |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | secureJsonData | Client Key | Yes (via SDK) |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | jsonData | With CA Cert | Yes (via SDK) |
| `secureJsonData_tlsCACert` | `tlsCACert` | secureJsonData | CA Cert | Yes (via SDK) |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | jsonData | Skip TLS Verify | Yes (via SDK) |
| `jsonData_keepCookies` | `keepCookies` | jsonData | Allowed cookies | Yes (via SDK) |
| `jsonData_timeout` | `timeout` | jsonData | Timeout | Yes (via SDK) |
| `jsonData_database` | `database` | jsonData | Index name | Yes (`opensearch.go:77, client.go:120`) |
| `jsonData_interval` | `interval` | jsonData | Pattern | Yes (`opensearch.go:78, client.go:126`) |
| `jsonData_timeField` | `timeField` | jsonData | Time field name | Yes (`opensearch.go:70-75, client.go:111-114` — hard-required in both) |
| `jsonData_serverless` | `serverless` | jsonData | Serverless | Yes (`client.go:32,51-53,300-302,532-534`) |
| `jsonData_flavor` | `flavor` | jsonData | — (not directly editable) | Yes (`opensearch.go:56-61, client.go:109,557-560` — hard-required) |
| `jsonData_version` | `version` | jsonData | Version | Yes (`opensearch.go:63-68, client.go:104-107` — hard-required) |
| `jsonData_versionLabel` | `versionLabel` | jsonData | — (display string) | **No — frontend-only** |
| `jsonData_maxConcurrentShardRequests` | `maxConcurrentShardRequests` | jsonData | Max concurrent Shard Requests | Yes (`client.go:411,428`) |
| `jsonData_timeInterval` | `timeInterval` | jsonData | Min time interval | **No — frontend-only** (see below) |
| `jsonData_pplEnabled` | `pplEnabled` | jsonData | PPL enabled | **No — frontend-only** (see below) |
| `jsonData_logMessageField` | `logMessageField` | jsonData | Message field name | Yes (`client.go:118`) |
| `jsonData_logLevelField` | `logLevelField` | jsonData | Level field name | Yes (`client.go:116`) |
| `jsonData_dataLinks` | `dataLinks` | jsonData | Data links | **No — frontend-only** |

### Frontend-only settings

- **`root.access`** picks whether the browser or the Grafana backend fetches
  from the datasource. The plugin's Go code never reads it; the SDK does not
  even surface it on `backend.DataSourceInstanceSettings`.
- **`root.withCredentials`** governs cross-site cookie forwarding in the
  browser fetch. Never consumed by the backend HTTP client.
- **`jsonData.versionLabel`** is the human-readable version string rendered
  in the disabled Version input. Never read by the backend.
- **`jsonData.timeInterval`** is the "Min time interval" duration. It is a
  Grafana panel-plugin knob passed via `panelPluginJsonData` to the query
  engine's `Interval` calculation and does not flow through `NewClient`.
- **`jsonData.pplEnabled`** gates the PPL query editor's availability. The
  backend chooses the PPL endpoint by flavor
  (`_plugins/_ppl` for OpenSearch, `_opendistro/_ppl` for Elasticsearch,
  `client.go:557-560`) and never inspects this flag.
- **`jsonData.dataLinks`** is consumed exclusively by the frontend log-row
  viewer to render buttons on matching fields; no backend code path reads it.

### Backend-only settings

- **`secureJsonData.password`** — used by `client.go:294-298,528-530` when
  `settings.User != ""` and `settings.BasicAuthEnabled` is false. No editor UI
  writes it; only very old provisioning payloads that set `user` at the root
  and provide `password` as a secret ever exercise this path. Not modeled in
  either the schema or `SecureJsonDataKeys` because the conformance test
  requires schema/secure-key parity and this key has no editor field.

### Fields deliberately excluded

- **`jsonData.enableSecureSocksProxy`** — written by `@grafana/ui`'s
  `SecureSocksProxySettings`, rendered conditionally at
  `ConfigEditor.tsx:64-66`. Per repo `AGENTS.md`, this field is deliberately
  omitted from every registry entry.
- **`sigV4*` jsonData fields and their `sigV4AccessKey` / `sigV4SecretKey`
  secrets** — contributed by `@grafana/aws-sdk`'s `SIGV4ConnectionConfig` (not
  by this plugin) and only offered by the editor when the Grafana instance has
  `sigV4AuthEnabled`. The `sigV4Auth` discriminator boolean is modeled because
  it is a plugin-owned discriminator; the companion fields are external and
  out of scope for this schema.
- **Dynamic `httpHeaderName<N>` (jsonData) / `httpHeaderValue<N>`
  (secureJsonData) pairs** — written by `@grafana/ui`'s `CustomHeadersSettings`
  component when the user configures custom HTTP headers. Their keys are
  dynamically indexed, so they are not modeled as first-class fields; see
  `settings.ts` and `settings.go` for the note.
- **Legacy `root.user` + `secureJsonData.password`** — a very old Basic-auth
  path that the current editor never writes. Retained in the backend but not
  reachable from the config UI.

## Where the types are defined

The configuration types are spread across the plugin and its dependencies — some
fields and base types come from libraries/SDKs rather than the plugin itself:

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `OpenSearchOptions` (jsonData), `Flavor`, `QueryType`, `DataLinkConfig` | `src/types.ts:12-116` | plugin ([grafana/opensearch-datasource](https://github.com/grafana/opensearch-datasource)) |
| `DataSourceJsonData` (base interface `OpenSearchOptions` extends) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `12.4.2` |
| `DataSourceSettings`, `DataSourcePluginOptionsEditorProps`, `DataLinkBuiltInVars`, `VariableOrigin` | `packages/grafana-data/src/` | `@grafana/data` `12.4.2` |
| `DataSourceHttpSettings`, `BasicAuthSettings`, `HttpProxySettings`, `TLSAuthSettings`, `CustomHeadersSettings`, `SecureSocksProxySettings` | `packages/grafana-ui/src/components/DataSourceSettings/` | `@grafana/ui` `12.4.2` |
| `LegacyForms.FormField`, `LegacyForms.Input`, `LegacyForms.Select`, `LegacyForms.Switch`, `Alert`, `Button`, `VerticalGroup`, `DataLinkInput`, `TagsInput`, `RadioButtonGroup`, `InlineField`, `InlineSwitch` | `packages/grafana-ui/src/components/` | `@grafana/ui` `12.4.2` |
| `SecureSocksProxyConfig` / `enableSecureSocksProxy` jsonData field (excluded from this entry) | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | `@grafana/ui` `12.4.2` |
| `SIGV4ConnectionConfig` — contributes `sigV4Auth`, `sigV4AuthType`, `sigV4Region`, `sigV4Profile`, `sigV4AssumeRoleArn`, `sigV4ExternalId` jsonData and `sigV4AccessKey` / `sigV4SecretKey` secrets | `@grafana/aws-sdk` `pkg/sigV4/` | `@grafana/aws-sdk` `0.10.2` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| No plugin-owned settings struct — `NewOpenSearchDatasource` delegates to `client.NewDatasourceHttpClient`, which unmarshals a tiny anonymous struct (`serverless`, `oauthPassThru`) | `pkg/opensearch/opensearch.go:28-39`, `pkg/opensearch/client/client.go:30-63` | plugin |
| `client.ConfiguredFields`, `client.Client`, `client.baseClientImpl` — the internal, post-parsing shape the client uses | `pkg/opensearch/client/client.go:66-166` | plugin |
| `client.Flavor` string constants (`elasticsearch`, `opensearch`) | `pkg/opensearch/client/models.go:13-18` | plugin |
| `backend.DataSourceInstanceSettings` (carries `URL`, `User`, `Database`, `BasicAuthEnabled`, `BasicAuthUser`, `JSONData`, `DecryptedSecureJSONData`) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.291.1` |
| `settings.HTTPClientOptions(ctx)` — reads TLS, timeout, cookies, basic auth, SigV4 config | `backend/httpclient` | `github.com/grafana/grafana-plugin-sdk-go` |
| `awsauth.NewSigV4Middleware` — attaches on top of the SDK's SigV4 transport | `pkg/awsauth` | `github.com/grafana/grafana-aws-sdk` `v1.4.3` |
| `semver.NewVersion` / `semver.NewConstraint` — used by both `ExtractVersion` and the shard-request coercion | | `github.com/Masterminds/semver` `v1.5.0` |

The models in this entry flatten that spread into a single Go `Config` type (root
fields tagged `json:"-"` + jsonData fields + `DecryptedSecureJSONData`) plus a
`SecureJsonDataKey` typed constant list. `settings.ts` keeps the three canonical
TypeScript types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`). Enum
constants (`Flavor`, `Interval`, `SecureJsonDataKey`) mirror the plugin's
frontend sources; there are no backend equivalents to sync to for `Interval`
(the backend consumes `interval` as a raw string).

## Modeling decisions

- **Independent auth toggles (no discriminator)**: OpenSearch uses the older
  `@grafana/ui` `DataSourceHttpSettings` which renders independent boolean
  switches for `basicAuth`, `withCredentials`, `sigV4Auth`, `tlsAuth`,
  `tlsAuthWithCACert`, `tlsSkipVerify`, and `oauthPassThru`. Unlike the newer
  `@grafana/plugin-ui` `Auth` component (used by the Elasticsearch plugin),
  no discriminator constrains combinations. The schema therefore models each
  toggle as an independent field with role annotations rather than the
  `virtual_authMethod` selector pattern the Elasticsearch entry uses.
- **`root.access` modeled but not on `Config`**: the editor writes it and it
  gates several editor sections (Additional HTTP, Forward OAuth, Custom
  Headers), so the schema declares it as a root field. But
  `backend.DataSourceInstanceSettings` does not surface it, so it is not
  reachable from Go and is intentionally absent from `Config`.
- **`root.withCredentials` modeled but not on `Config`**: same story as
  `access` — the editor writes it, the backend HTTP client does not read it.
- **SigV4 companion fields not modeled**: `sigV4AuthType`, `sigV4Region`, etc.
  are contributed by `@grafana/aws-sdk`'s `SIGV4ConnectionConfig`, not by
  this plugin. Only the plugin-owned `sigV4Auth` boolean discriminator is
  modeled; provisioning payloads that use SigV4 should include the companion
  fields directly in jsonData (see the `sigV4Managed` / `sigV4Serverless`
  settings examples) but they are out of scope for the schema.
- **`root_url` marked `requiredWhen: "true"`**: the backend does not
  explicitly check the URL, but the health check parses it (opensearch.go:99)
  and every msearch / PPL / resource call builds URLs from it; empty URL is
  unusable.
- **`jsonData_flavor` + `jsonData_version` both required**: the health check
  hard-fails on either (opensearch.go:56-68) and the per-query client factory
  also requires version (client.go:104-107). Both are wired together by the
  editor's "Get Version and Save" button, which is why they are declared as a
  `pair` relationship.
- **`jsonData_version` uses `dependsOn: "jsonData_serverless != true"`**: the
  editor hides the Version input for serverless (OpenSearchDetails.tsx:156)
  because `getServerlessSettings` hard-codes version to `1.0.0`. The
  requirement still fires — `ApplyDefaults` writes `1.0.0` when serverless is
  true so the Validate step passes.
- **`timeField` defaults to `@timestamp`**: `coerceOptions` writes this on
  mount (`utils.ts:20`); the backend hard-fails on an empty timeField
  (`opensearch.go:70-75`, `client.go:111-114`). We apply the default in
  `ApplyDefaults` for editor parity and enforce presence in `Validate`.
- **`maxConcurrentShardRequests` string-or-number normalization**: the
  backend's `simplejson.MustInt` accepts the field as either a JSON number or
  a JSON string. `Config.UnmarshalJSON` mirrors that: try number, then string,
  then fall through to `ApplyDefaults`' flavor/version-aware fallback.
- **Flavor/version-aware default for `maxConcurrentShardRequests`**: 256 for
  Elasticsearch <7.0.0, 5 for OpenSearch and Elasticsearch >=7.0.0
  (`OpenSearchDetails.tsx:322-327`, `client.go:411-433`). We reproduce that
  in `defaultMaxConcurrentShardRequestsFor` (Go) using a lightweight
  major-version parse; the plugin uses `semver` for the same check.
- **`interval` uses `""` for "No pattern"**: `indexPatternTypes[0]` has
  `value: 'none'` in the editor (`OpenSearchDetails.tsx:11`), but
  `intervalHandler` converts `'none'` to `undefined` before storing
  (`OpenSearchDetails.tsx:254`). The schema uses `value: ""` for the "No
  pattern" option, matching the wire format the backend reads.
- **`pplEnabled` is a plain `bool`**: The dsconfig conformance walker does
  not handle `*bool` fields, so we use `bool`. This means we cannot
  distinguish "unset" from "explicitly false" on load, so `ApplyDefaults`
  does not attempt to default it to `true`. The field is frontend-only, so
  the loss is documentation-only — see `TestApplyDefaults` and the comment
  on `Config.PPLEnabled` for the note.
- **`dataLinks.url` label uses `URL` (not `URL/Query`)**: the editor's label
  flips between `URL` and `Query` at runtime based on the internal-link
  switch (`DataLink.tsx:89`). We record the neutral `URL` label because
  consumers use `datasourceUid` to distinguish external vs internal.
- **`dataLinks` uses `title` not `urlDisplayLabel`**: OpenSearch's data-link
  shape (`src/types.ts:101-106`) diverges from the Elasticsearch plugin's
  `DataLinkConfig` (which uses `urlDisplayLabel`). We honor the plugin's
  own type verbatim.
- **Browser access mode allowed**: unlike the Elasticsearch plugin which
  renders a persistent Alert when `access='direct'`, OpenSearch supports
  browser mode (there is no such Alert). The editor still disables the
  Additional-HTTP, Forward-OAuth, and Custom-Headers panels for direct mode,
  which is why we mark those fields with `dependsOn: "root_access == 'proxy'"`.
- **Secure Socks Proxy excluded**: the editor conditionally renders
  `SecureSocksProxySettings` (`ConfigEditor.tsx:64-66`) writing
  `jsonData.enableSecureSocksProxy` when the Grafana instance has
  `secureSocksDSProxyEnabled`. Per repo `AGENTS.md`, the field is
  deliberately omitted from this registry entry.
- **Field ID naming convention**: IDs are prefixed with their storage target
  (`root_`, `jsonData_`, `secureJsonData_`) followed by the camelCase storage
  key. The `key` property keeps the plugin's raw storage key.
- **Flat `Config` in Go**: `settings.go` collapses root fields (`URL`,
  `BasicAuth`, `BasicAuthUser`, `User`, `Database`, all tagged `json:"-"`) and
  jsonData fields onto a single `Config` struct with the same json tags the
  backend reads via its `simplejson.Get(...)` calls.
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

`SettingsExamples()` provides the default configuration plus one k8s-style example
per authentication method and connection variant:

| Example | Auth | Notes | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | Anonymous | Editor defaults + placeholder index (OpenSearch 1.0.0) | `basicAuthPassword` (empty) |
| `noAuth` | Anonymous | Daily logstash-style time-pattern index; OpenSearch 2.11.0 | `basicAuthPassword` (empty) |
| `basicAuth` | Basic auth | Root-level `basicAuth`/`basicAuthUser`; password is secret | `basicAuthPassword` |
| `sigV4Managed` | SigV4 | AWS-managed OpenSearch (`Service='es'`); SigV4 companion fields in jsonData | `basicAuthPassword` (empty) |
| `sigV4Serverless` | SigV4 | AWS OpenSearch Serverless (`Service='aoss'`, `jsonData.serverless=true`) | `basicAuthPassword` (empty) |
| `oauthForward` | Forward OAuth Identity | Sets `jsonData.oauthPassThru=true`; no secret | `basicAuthPassword` (empty) |
| `tlsMutualAuth` | mTLS | Client cert + key + serverName | `tlsClientCert`, `tlsClientKey` |
| `tlsSelfSignedCA` | CA verification | `tlsAuthWithCACert=true` | `tlsCACert` |
| `elasticsearchLegacy` | Anonymous | Elasticsearch 6.8.0 flavor — PPL endpoint switches, default shards = 256 | `basicAuthPassword` (empty) |
| `logsWithDataLinks` | Basic auth | Logs mode + two data link entries (external + internal) | `basicAuthPassword` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and
returns a fully-defaulted, validated `Config`:

1. **Parse** — copy root fields (`URL`, `BasicAuth`, `BasicAuthUser`, `User`,
   `Database`) from `settings`, unmarshal jsonData into `Config` (with the
   custom `UnmarshalJSON` that accepts `maxConcurrentShardRequests` as JSON
   number OR string), and copy decrypted secrets by known key into
   `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — write the curated set of editor-parity defaults
   (`TimeField='@timestamp'`, flavor/version-aware
   `MaxConcurrentShardRequests`, serverless overrides). `PPLEnabled` is
   intentionally not defaulted because we cannot distinguish "unset" from
   "explicitly false" on a `bool` and the field is frontend-only.
3. **`Validate`** — enforce the runtime contract: URL required, flavor
   required (with enum check), version required, timeField required, valid
   interval enum, per-auth-method secret requirements, TLS field pairs,
   non-negative timeout / maxConcurrentShardRequests. Errors are joined so
   every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with
`datasource_uid`, `datasource_name`, and `plugin` labels so log lines carry
request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are exported for
callers that want to compose them themselves (provisioning preview,
schema-example round-trip, tests that distinguish parse-level from
policy-level errors). Skip them by never calling `LoadConfig` in those flows —
assemble a `Config` directly.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while
researching upstream. All preserved verbatim in the schema — the schema
records what the plugin **does**, not what it **should** do; these notes
exist so reviewers can reproduce each finding and decide separately whether
to fix upstream.

1. **`intervalHandler` writes to `root.database`, not `jsonData.database`.**
   The Index name input reads and writes `value.jsonData.database`
   (`OpenSearchDetails.tsx:109-110`), and the backend reads
   `jsonData.database` (`opensearch.go:77, client.go:120`). But
   `intervalHandler` reads and writes `value.database` at the root
   (`OpenSearchDetails.tsx:252, 269`) — a latent inconsistency preserved by
   the test suite (`OpenSearchDetails.test.tsx:35` asserts on the root
   `database` field). Selecting a Pattern therefore writes a dead root-level
   value while the actual index-name field remains untouched.
2. **Version is set implicitly by a button, not typed by the user.** The
   Version input is `disabled` (`OpenSearchDetails.tsx:164`); the value only
   changes when the user presses "Get Version and Save", which HTTPs the
   cluster and unpacks the result into jsonData.version, jsonData.flavor,
   jsonData.versionLabel, and jsonData.maxConcurrentShardRequests. Users can
   still provision a datasource with an arbitrary version/flavor pair
   (bypassing the health probe) via YAML.
3. **`serverless=true` forces flavor/version/maxConcurrentShardRequests on
   save.** Toggling Serverless writes `flavor='opensearch'`, `version='1.0.0'`,
   `maxConcurrentShardRequests=5`, `pplEnabled=true`
   (`OpenSearchDetails.tsx:66-80`), even if the user just came off Elasticsearch
   6.x. This is a deliberate reset but is undocumented in the tooltip.
4. **`jsonData.pplEnabled` gates nothing on the backend.** The backend
   chooses the PPL endpoint by flavor alone (`client.go:557-560`). The
   toggle only affects whether the PPL query editor is offered in the UI. A
   PPL query submitted with `pplEnabled=false` still runs.
5. **`jsonData.timeInterval` regex is editor-only.** The invalid-state check
   at `OpenSearchDetails.tsx:196-200` flags bad inputs in the UI, but the
   backend never validates this string — invalid values silently pass through
   to Grafana's `Interval` calculation.
6. **`jsonData.maxConcurrentShardRequests` accepted as a string.** The
   editor's Input writes it as a string (`OpenSearchDetails.tsx:178-179` sets
   `value` to `value.jsonData.maxConcurrentShardRequests || ''`), and the
   backend accepts both number and string via `MustInt`, but provisioning
   docs only show the numeric form.
7. **Backend silently coerces `flavor` to `opensearch` when missing at
   `client.go:109`.** Only the health check hard-fails on missing flavor
   (`opensearch.go:56-61`); actual queries default to OpenSearch semantics
   even if the cluster is Elasticsearch. Health-check success is therefore
   an important precondition for correct query behavior.
8. **Legacy `root.user` + `secureJsonData.password` path is dead but still
   active.** `client.go:294-298,528-530` will still send those credentials
   when Basic Auth is off, but no current editor path writes them; only
   very-old provisioning payloads exercise it. Provisioning tools should
   avoid setting `user` at the root.
9. **`isValidOptions` re-runs `coerceOptions` on every change.** The editor's
   `useEffect` at `ConfigEditor.tsx:19-26` invokes `coerceOptions` whenever
   `isValidOptions` is false. Fields covered by coerceOptions (timeField,
   maxConcurrentShardRequests, logMessageField, logLevelField, pplEnabled)
   reappear on save whenever a user clears them, because `isValidOptions`
   treats `undefined` as invalid.
10. **The docs URL points to the GitHub repo, not to grafana.com/docs.** Unlike
    other plugins whose `info.links` includes a `Documentation` link to
    `grafana.com/docs/plugins/...`, OpenSearch only ships a `Github` link
    (`plugin.json:24-28`). That is what we set as `docURL`.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in
  this repo) — passes.
- JSON Schema validation against
  [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft 2020-12,
  `additionalProperties: false`) — passes.
- `go test ./...` on this module — passes (schema bundle shape, secure values,
  examples, `LoadConfig` incl. serverless overrides, flavor/version-aware
  maxConcurrentShardRequests default, string-or-number
  maxConcurrentShardRequests, `SchemaArtifactInSync` guard).
- `settings.go` / `schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: to be validated with `tsc --noEmit --strict`.
