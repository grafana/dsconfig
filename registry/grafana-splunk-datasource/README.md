# grafana-splunk-datasource

dsconfig registry entry for the **Splunk** datasource plugin.

- **Plugin ID / `pluginType`**: `grafana-splunk-datasource` (from `src/plugin.json:4`)
- **Plugin name**: `Splunk` (`src/plugin.json:3`)
- **Docs URL**: <https://grafana.com/docs/plugins/grafana-splunk-datasource> (`src/plugin.json:42`)
- **Import path**: `github.com/grafana/dsconfig/registry/grafana-splunk-datasource`
- **Go package**: `splunkdatasource`

## Files

| File | Purpose |
| --- | --- |
| `dsconfig.json` | dsconfig v1 schema — the single source of truth for the config surface |
| `settings.ts` | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| `settings.go` | Flat Go `Config` (jsonData + root + `DecryptedSecureJSONData`), enums, `SecureJsonDataKey`, `LoadConfig`/`ApplyDefaults`/`Validate` |
| `schema.go` | `//go:embed dsconfig.json`; `ConfigSchema()`, `NewSchema()`, `SettingsExamples()` |
| `conformance_test.go` | `schema.RunPluginTests` wrapper (also regenerates artifacts under `-generateArtifacts`) |
| `settings_test.go` | Table tests for `LoadConfig`, `ApplyDefaults`, `Validate` |
| `schema.gen.json`, `settings.gen.json`, `settings.examples.gen.json` | Generated artifacts (via `go generate ./...`) |

## Sources researched

Researched against **`github.com/grafana/plugins-private`** (a monorepo) at commit
**`267f4937806ed6404b6628d13ae358a5d308e376`**, plugin path
`plugins/grafana-splunk-datasource/`.

Frontend / editor:
- `src/plugin.json` — plugin id, name, docs URL.
- `src/types.ts` — `SplunkOptions` (jsonData), `SplunkSecureJsonData`, `AuthMethods`.
- `src/components/ConfigEditor.tsx` — outer editor (composes `DataSourceDescription`,
  `ConnectionSettings`, `SplunkAuthComponent`, `AdditionalSettingsEditor`, `DataLinks`).
- `src/components/SplunkAuthComponent.tsx` — auth method selector (`Auth` + `convertLegacyAuthProps`)
  with the custom `custom-splunk` method and its `authToken` field.
- `src/components/AdditionalSettingsEditor.tsx` — `AdvancedHttpSettings` + the "Advanced options" section.
- `src/components/selectors.ts` — URL label/placeholder (`URL`).
- `src/datasource/query.ts`, `src/datasource/Datasource.ts` — how config values become query defaults.

Backend:
- `pkg/models/settings.go` — `Settings` struct + `LoadSettings` (defaulting, streamMode migration).
- `pkg/models/settings_test.go` — expected defaulting behavior (used to mirror `LoadConfig`).
- `pkg/models/query_limits.go` — `GetSettingsResultsLimit` (`maxResultCount` resolution + env override).
- `pkg/splunk/client.go` — URL derivation (`/services/...`) and `authToken` → `Authorization: Bearer`.
- `pkg/splunk/auth.go` — vestigial `AuthenticationType` enum (unused by the config path).

External component libraries (pinned via the monorepo `.yarnrc.yml` catalog, read from the
extracted npm package `grafana-plugin-ui-0.13.1.tgz`):
- **`@grafana/plugin-ui@^0.13.1`** — `ConnectionSettings` (`dist/esm/.../Connection/ConnectionSettings.js`),
  `Auth` + `convertLegacyAuthProps` + `AuthMethodSettings` + `BasicAuth` + the TLS sub-components
  (`dist/esm/.../Auth/*`), `AdvancedHttpSettings` (`dist/esm/.../AdvancedSettings/AdvancedHttpSettings.js`),
  `DataLinks`/`DataLink` (`dist/esm/components/DataLinks/*`).
- **`@grafana/ui@^11.6.7`**, **`@grafana/data@^11.6.7`**, **`@grafana/runtime@^11.6.7`** — base UI
  primitives and update helpers.

## Field inventory

Field IDs use the `<target>_<camelCaseKey>` convention. "Read by backend" = read by the plugin's
own `pkg/models/settings.go` `LoadSettings`; fields marked *(SDK)* are consumed by the shared
`config.HTTPClientOptions(ctx)` path (`settings.go:125`) rather than the plugin's struct, and
*(core)*/*(FE)* denote Grafana-core-consumed and frontend-only fields.

| Schema ID | Storage key | Target | Editor label | Read by backend |
| --- | --- | --- | --- | --- |
| `root_url` | `url` | root | URL | yes (`settings.go:90`) |
| `jsonData_authType` | `authType` | jsonData | Authentication method | yes (`settings.go:34,95`) |
| `root_basicAuthUser` | `basicAuthUser` | root | User | no *(SDK basic auth)* |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | secureJsonData | Password | no *(SDK basic auth)* |
| `secureJsonData_authToken` | `authToken` | secureJsonData | Authentication token | yes (`settings.go:92`, `client.go:229`) |
| `jsonData_oauthPassThru` | `oauthPassThru` | jsonData | — (managed) | no *(core)* |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | jsonData | Add self-signed certificate | no *(SDK)* |
| `secureJsonData_tlsCACert` | `tlsCACert` | secureJsonData | CA Certificate | no *(SDK)* |
| `jsonData_tlsAuth` | `tlsAuth` | jsonData | TLS Client Authentication | no *(SDK)* |
| `jsonData_serverName` | `serverName` | jsonData | ServerName | no *(SDK)* |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | secureJsonData | Client Certificate | no *(SDK)* |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | secureJsonData | Client Key | no *(SDK)* |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | jsonData | Skip TLS certificate validation | no *(SDK)* |
| `jsonData_keepCookies` | `keepCookies` | jsonData | Allowed cookies | yes (`settings.go:38`, SDK) |
| `jsonData_timeout` | `timeout` | jsonData | Timeout | no *(SDK)* |
| `jsonData_maxResultCount` | `maxResultCount` | jsonData | Limit number of results | yes (`settings.go:42`) |
| `jsonData_previewMode` | `previewMode` | jsonData | Enable preview mode | yes (`settings.go:44`) |
| `jsonData_pollSearchResult` | `pollSearchResult` | jsonData | Enable async queries | yes (`settings.go:31`) |
| `jsonData_minPollInterval` | `minPollInterval` | jsonData | Min | no *(FE query.ts:259)* |
| `jsonData_maxPollInterval` | `maxPollInterval` | jsonData | Max | no *(FE query.ts:260)* |
| `jsonData_autoCancel` | `autoCancel` | jsonData | Auto cancel timeout | no *(FE query.ts:271)* |
| `jsonData_timeoutInSeconds` | `timeoutInSeconds` | jsonData | Timeout in seconds | yes (`settings.go:46`) |
| `jsonData_statusBuckets` | `statusBuckets` | jsonData | Set maximum status buckets | no *(FE query.ts:272)* |
| `jsonData_internalFieldsFiltration` | `internalFieldsFiltration` | jsonData | Internal fields filtration | yes (`settings.go:28`) |
| `jsonData_internalFieldPattern` | `internalFieldPattern` | jsonData | Internal field pattern | yes (`settings.go:29`) |
| `jsonData_tsField` | `tsField` | jsonData | Set time stamp field | yes (`settings.go:45`) |
| `jsonData_fieldSearchType` | `fieldSearchType` | jsonData | Set fields search mode | yes (`settings.go:43`) |
| `jsonData_variableSearchLevel` | `variableSearchLevel` | jsonData | Set variables search mode | yes (`settings.go:47`) |
| `jsonData_defaultEarliestTime` | `defaultEarliestTime` | jsonData | Set default earliest time | no *(FE query.ts:274)* |
| `jsonData_streamMode` | `streamMode` | jsonData | — (legacy, no UI) | yes (`settings.go:30,103-105`) |
| `jsonData_dataLinks` | `dataLinks` | jsonData | Data links | yes (`settings.go:33,62-85`) |

Data-link item fields (`isItemField: true`, under `jsonData_dataLinks.item.*`): `field`, `label`,
`matcherRegex`, `url`, `datasourceUid` (verbatim from `@grafana/plugin-ui` `DataLink.js`).

### Frontend-only settings

Written by the editor (or applied as query defaults) but **not** read by the plugin's backend
settings loader:

- `jsonData.oauthPassThru` — consumed by Grafana core for OAuth token forwarding.
- `jsonData.timeout` — consumed by the SDK HTTP client (`config.HTTPClientOptions`), not the plugin `Settings` struct.
- `jsonData.minPollInterval`, `jsonData.maxPollInterval` — used by the frontend query builder for async polling (`src/datasource/query.ts:259-262`).
- `jsonData.autoCancel`, `jsonData.statusBuckets`, `jsonData.defaultEarliestTime` — applied as per-query defaults by the frontend (`src/datasource/query.ts:271-274`).
- Root `basicAuthUser` + `secureJsonData.basicAuthPassword` — consumed by the SDK Basic-auth path, not the plugin `LoadSettings`.
- All TLS fields (`tlsAuth`, `tlsAuthWithCACert`, `serverName`, `tlsSkipVerify`, `tlsCACert`, `tlsClientCert`, `tlsClientKey`) — consumed by the SDK `HTTPClientOptions`.

### Backend-only / legacy settings

- `jsonData.streamMode` — no editor UI; the backend migrates a `true` value into `previewMode`
  (`settings.go:103-105`) and the frontend query builder treats it as preview (`query.ts:267`).
  Modeled and tagged `legacy` for round-trip fidelity.
- `secureJsonData.APIKey` — **dead secret**: the backend copies it into `settings.APIKey`
  (`settings.go:27,91`) but no code path consumes it, and the editor never writes it. **Not modeled**
  as a schema field (it is not part of the real auth surface); documented here only.

## Modeling decisions

- **`authType` is a real, stored discriminator** (`role: auth.discriminator`), unlike the virtual
  selectors used by Prometheus / Falcon LogScale. The `@grafana/plugin-ui` `Auth` component
  persists the selected method id verbatim into `jsonData.authType`, and the backend reads it
  directly. `jsonData.oauthPassThru` is modeled as a `managed-by:jsonData_authType` field (a `pair`
  relationship documents that it is set `true` only for `OAuthForward`).
- **`BasicAuth` ⇔ empty `authType`.** The editor's `selectedMethod` falls back to `BasicAuth` when
  `authType` is empty (`SplunkAuthComponent.tsx:11`), and the backend treats an empty `authType` as
  Basic auth (`settings.go:95`). Both `dependsOn`/`requiredWhen` for the Basic-auth fields therefore
  use `jsonData_authType == 'BasicAuth' || jsonData_authType == ''`, and `Validate` accepts `""`.
- **Root `basicAuth` is not stored.** Splunk overrides `onAuthMethodSelect`, so the editor never
  writes `root.basicAuth` (unlike the standard `convertLegacyAuthProps` handler); the backend derives
  it. So `root_basicAuth` is intentionally absent; only `root_basicAuthUser` is modeled at root.
- **TLS and custom-CA are modeled** because the `Auth` component receives the `TLS` prop from
  `convertLegacyAuthProps` and always renders the TLS sub-section. This adds `tlsCACert`,
  `tlsClientCert`, `tlsClientKey` to the secret set.
- **Custom HTTP headers are not modeled.** `convertLegacyAuthProps` also wires `@grafana/plugin-ui`'s
  `CustomHeaders`, which stores dynamic `httpHeaderName<N>` (jsonData) + `httpHeaderValue<N>`
  (secureJsonData). Following the Prometheus entry, these dynamic pairs are not modeled as
  first-class fields.
- **`minPollInterval`/`maxPollInterval` are modeled as `string`.** `SplunkOptions` declares them
  `number`, but the editor persists strings via `onUpdateDatasourceJsonDataOption`
  (`AdditionalSettingsEditor.tsx:221-231`); the storage reality is a string.
- **`autoCancel`/`statusBuckets` are `string`** (editor writes the raw input string) and modeled as such.
- **Secure Socks Proxy excluded** per registry policy (`jsonData.enableSecureSocksProxy`,
  `AdditionalSettingsEditor.tsx:131-146`).
- **`LoadConfig` = parse → `ApplyDefaults` → `Validate`.** The parse phase mirrors `LoadSettings`
  (`settings.go:102-123`): streamMode→previewMode migration, `internalFieldPattern` default `^_.+`
  (then cleared when filtration is off), `tsField` default `_time`, and `timeoutInSeconds < 1` → 30.
  `ApplyDefaults` fills the editor-parity discriminators (`authType`→`BasicAuth`,
  `fieldSearchType`→`quick`, `variableSearchLevel`→`fast`). `Validate` enforces the auth + TLS
  runtime contract. The env-driven `maxResultCount` resolution (`query_limits.go`) is a runtime
  detail and is **not** applied to the stored value.

### Where the types are defined

Frontend:
- `SplunkOptions` (jsonData), `SplunkSecureJsonData`, `AuthMethods` — `src/types.ts:96-128`.
- `AuthMethod` enum (`BasicAuth`, `OAuthForward`, `NoAuth`, `CrossSiteCredentials`) — `@grafana/plugin-ui`
  `dist/esm/.../Auth/types.js`.
- `DataLinkConfig` (frontend) — `@grafana/plugin-ui` `DataLinks` types.

Backend:
- `Settings`, `DataLinkConfig` — `pkg/models/settings.go:14-48`.
- `AuthenticationType` (vestigial; unused by the config path) — `pkg/splunk/auth.go:4-11`.

## Settings examples matrix

| Example key | Auth method | Notable fields | Secrets |
| --- | --- | --- | --- |
| `""` (default) | Basic (`authType=BasicAuth`) | url only | `basicAuthPassword: ""` (fails `Validate` — placeholder) |
| `basicAuth` | Basic | `basicAuthUser` | `basicAuthPassword` |
| `alternativeToken` | `custom-splunk` | — | `authToken` |
| `oauthForward` | `OAuthForward` | `oauthPassThru: true` | none |
| `basicAuthWithTLS` | Basic + mTLS + custom CA | `tlsAuth`, `serverName`, `tlsAuthWithCACert` | `basicAuthPassword`, `tlsClientCert`, `tlsClientKey`, `tlsCACert` |
| `tokenWithAdvancedOptions` | `custom-splunk` | preview/async, timeouts, filtering, cookies, data links | `authToken` |

All secret values use obvious angle-bracket placeholders (`<your-splunk-password>`,
`<splunk-authentication-token>`) or redacted PEM blocks; none are realistic token shapes.

## Potential upstream bugs / discrepancies

- **Dead `jsonData.apiURL`** — declared on `SplunkOptions` (`src/types.ts:97`) but never read or
  written by the current editor or backend.
- **Dead `jsonData.username`** — declared on `SplunkOptions` (`src/types.ts:109`) but never written
  (the Basic-auth username is stored at `root.basicAuthUser`).
- **Dead `secureJsonData.APIKey`** — read into `settings.APIKey` (`settings.go:27,91`) but never
  consumed and never written by the editor.
- **Two timeout fields** — `jsonData.timeout` (SDK, from `AdvancedHttpSettings`) and
  `jsonData.timeoutInSeconds` (plugin-specific). An upstream `TODO`
  (`AdditionalSettingsEditor.tsx:257`) notes they may want to merge them.
- **`minPollInterval`/`maxPollInterval` type mismatch** — declared `number`, persisted as `string`
  by the editor (see modeling decisions).
- **Data-link empty-string panic risk** — `LoadSettings` indexes `dataLinkEntry.Field[0]`
  (`settings.go:68`) and `dataLinkEntry.RawRegex[0]` (`settings.go:76`) without guarding against an
  empty string, so a saved data link with an empty `field` or `matcherRegex` would panic the backend.
  (This entry's `Validate` does not attempt to reproduce that panic.)
- **`OAuthForward` is feature-gated** by the `splunkEnableOAuthForwarding` feature toggle
  (`SplunkAuthComponent.tsx:26-33`); it is only offered in the editor when the toggle is on, though
  the backend handles the stored value regardless. It is kept as an allowed `authType` value.
- **`maxResultCount` env override** — resolved at runtime via
  `GF_PLUGIN_GRAFANA_SPLUNK_DATASOURCE_MAX_RESULT_LIMIT` (`query_limits.go:21`); `0` resolves to the
  `10000` safety limit rather than "unlimited".

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (via `conformance_test.go`).
- Strict JSON Schema validation of `dsconfig.json` against `dsconfig/schema.json`
  (`additionalProperties: false`) — passes.
- `schema.RunPluginTests`: schema round-trip, artifact drift, `secureJsonData` absent from the
  settings spec, `secureValues` == `SecureJsonDataKeys`, jsonData/struct parity (both directions),
  jsonData/struct type parity.
- `go build ./...`, `go vet ./...`, `gofmt -l .`, `go test ./...` inside `registry/` — all clean.
- `tsc --noEmit --strict settings.ts` — passes.
- `dsconfig` and `schema` workspace modules still build and test.
