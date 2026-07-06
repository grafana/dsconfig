# grafana-databricks-datasource

Declarative configuration schema for the Databricks datasource plugin (`grafana-databricks-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/plugins-private` (monorepo)
- **Monorepo path**: `plugins/grafana-databricks-datasource`
- **Ref**: `main`
- **Commit SHA**: `267f4937806ed6404b6628d13ae358a5d308e376`

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, option
labels/values, group titles, defaults, validations, dependency and required-when expressions,
storage keys, storage targets, value types, and instructions — is traceable to a specific
`file:line` in the upstream plugin at this SHA. See [Field provenance](#field-provenance) below.

To reproduce this research with a sparse checkout of just this plugin:

```bash
git clone --filter=blob:none --no-checkout https://github.com/grafana/plugins-private
cd plugins-private
git sparse-checkout init --cone
git sparse-checkout set plugins/grafana-databricks-datasource
git checkout 267f4937806ed6404b6628d13ae358a5d308e376
cd plugins/grafana-databricks-datasource
```

`catalog:` dependency versions are resolved from the monorepo root catalog in
`.yarnrc.yml` (the `catalog:` YAML block); the plugin-local `@grafana/azure-sdk` pin in
`package.json` takes precedence over any catalog entry. If upstream `main` has advanced past this
SHA, re-diff the sources under [Sources researched](#sources-researched) before merging changes.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `AuthType` + `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned SHA (`267f493…`), plus external editor components at the
exact versions the plugin's `package.json` / catalog pins.

### Plugin (`plugins-private/plugins/grafana-databricks-datasource@267f493`)

| File | What was read |
| --- | --- |
| `src/plugin.json:4-5,26` | `pluginName` (`name`), `pluginType` (`id`), `docURL` (`info.links[0].url`) |
| `src/ConfigEditor.tsx:64-74` | `useEffect` defaulting `jsonData.authType` → `Pat` for new datasources |
| `src/ConfigEditor.tsx:76-94` | `onChange`: `httpPath` slash-stripping (`80-82`), `authType` side-effect setting `oauthPassThru` (`89-91`) |
| `src/ConfigEditor.tsx:129-131` | `checkAuthType` — the conditional-render predicate |
| `src/ConfigEditor.tsx:147-167` | Host / Http Path fields |
| `src/ConfigEditor.tsx:169-198` | Authentication Type select + Token (Pat/Unknown) field |
| `src/ConfigEditor.tsx:200-244` | On-Behalf-Of `AzureCredentialsForm` and OAuth M2M (ClientId/ClientSecret) fields |
| `src/ConfigEditor.tsx:246-312` | Azure Entra ID M2M fields (AzureCloud/TenantId/ClientId/ClientSecret) |
| `src/ConfigEditor.tsx:314-402` | Retries / Pause / Timeout / Max Rows / Retry Timeout / Debug / Unity Catalog Support / Default Query Format |
| `src/ConfigEditor.tsx:404-416` | Conditional `Secure Socks Proxy` switch — deliberately excluded from this entry |
| `src/selectors.ts:5-160` | Field ids, labels, placeholders, tooltips (`Components.ConfigEditor`) |
| `src/types.ts:8-40` | `Settings` (jsonData) interface |
| `src/types.ts:45-49` | `SecureSettings` (secure keys) |
| `src/types.ts:51-85` | `AuthenticationType`, `AuthenticationTypeLabel`, `AzureAuthTypes`, `SelectableAuthenticationTypes`, `SelectableQueryFormats` |
| `src/authUtils.ts:1-118` | `getCredentials` / `updateCredentials` — the `azureCredentials` object + `azureClientSecret` secret + forced `oauthPassThru` for OBO |
| `src/AzureCredentialsForm.tsx:1-187` | OBO sub-fields (TenantId/ClientId/ClientSecret) and the `azure-auth-config-section` |
| `pkg/models/settings.go:21-42` | Backend `Settings` struct fields and json tags |
| `pkg/models/settings.go:45-171` | `LoadSettings`: host/httpPath/token validation (`50-65`), OAuth PT header parse (`67-83`), OAuth M2M (`85-99`), Azure M2M + azureCloud default (`101-119`), OBO via `azcredentials.FromDatasourceData` + oauthPassThru guard (`121-144`), timeout/pause/retries defaults (`151-159`), CloudFetch force-enable (`161-168`) |
| `pkg/models/constants.go:5-35` | Error sentinels (`ErrMissingHost`, `ErrMissingToken`, `ErrInvalidOAuth`, …), `RetryOnStrings`, `ConnectionArgs` |
| `pkg/authentication/constants.go:4-17` | `AuthenticationType*` string constants (`Pat`, `OauthM2M`, `OauthPT`, `OauthOBO`, `AzureM2M`, Unknown `""`) |
| `pkg/database/connect.go:44-94` | `getConnector`: `WithServerHostname(host)`, `WithPort(443)`, `WithHTTPPath(httpPath)`, `WithMaxRows`, `WithCloudFetch`, per-authType authenticator wiring; Community-Edition guard (`21,45`) |
| `pkg/authentication/authenticator.go:20-55` | `NewOBOAuthenticator` — reads azureCloud/tenantId/clientId/clientSecret |
| `pkg/main.go:48,55,69,77` | `LoadSettings(..., false, nil)` (instance + driver), `EnableUnitySupport` gating, `Debug` → driver log level |
| `pkg/driver/driver.go:33` | `LoadSettings(ctx, config, true, message)` — the `validate=true` path (health/connect) that enforces host/httpPath/token |

### External editor components

Resolved at the versions pinned in the plugin's `package.json` and the monorepo root
`.yarnrc.yml` catalog.

| Component / type | Version | Source | What was read |
| --- | --- | --- | --- |
| `AzureCredentials`, `AzureAuthType`, `getAzureClouds`, `ConcealedSecret` | `@grafana/azure-sdk@0.0.3` (plugin-local pin, `package.json`) | `dist/clouds.js` (from `src/clouds.ts`) | `predefinedClouds` → the Azure Cloud select option values/labels: `AzureCloud`→"Azure", `AzureChinaCloud`→"Azure China", `AzureUSGovernment`→"Azure US Government"; the `clientsecret-obo` authType and `azureClientSecret`/`azureCredentials` storage keys |
| `QueryFormat` | `@grafana/plugin-ui@^0.13.1` (catalog) | `dist/esm/index.d.ts:520-526` | Numeric enum backing `defaultQueryFormat` (`Timeseries=0`, `Table=1`, `Logs=2`, `Trace=3`, `OptionMulti=4`) |
| `DataSourceDescription` | `@grafana/plugin-ui@^0.13.1` (catalog) | plugin-ui `v0.13.1` | Config-editor intro block (no storage) |
| `Checkbox`, `Select`, `InlineField`, `Input`, `SecretInput`, `InlineSwitch` | `@grafana/ui@^11.6.7` (catalog) | grafana/grafana `v11.6.x` `packages/grafana-ui` | Prop names (`label`, `placeholder`, `value`, `onChange`, `isConfigured`, `onReset`) so the correct UI attributes were recorded |
| `DataSourcePluginOptionsEditorProps`, `SelectableValue`, `onUpdateDatasourceJsonDataOptionChecked` | `@grafana/data@^11.6.7` (catalog) | grafana/grafana `v11.6.x` `packages/grafana-data` | Storage-key semantics of the update helpers |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each of its
label, placeholder, tooltip, default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_host` | `host` | `jsonData` | `selectors.ts:8` (`label: 'Host'`) | `selectors.ts:9` (`placeholder: 'https://your-databricks-instance.com'`) | `Settings.host string`, `types.ts:9`; `Settings.Host string`, `settings.go:22` | Role `endpoint.baseUrl`; `requiredWhen: true` (backend `ErrMissingHost`, `settings.go:51-53`) |
| `jsonData_httpPath` | `httpPath` | `jsonData` | `selectors.ts:15` (`label: 'Http Path'`) | `selectors.ts:16` (`placeholder: '/sql/protocolv1/o/0/1234567890'`) | `Settings.httpPath string`, `types.ts:10`; `Settings.HttpPath string`, `settings.go:23` | Editor strips leading/trailing slashes (`ConfigEditor.tsx:80-82`); `requiredWhen: true` (`settings.go:54-56`) |
| `jsonData_authType` | `authType` | `jsonData` | `selectors.ts:67` (`label: 'Authentication Type'`) | Tooltip `selectors.ts:69` (`tip: 'Authentication type of Databricks'`); options `types.ts:74-80`; default `Pat` via `ConfigEditor.tsx:64-74` | `AuthenticationType`, `types.ts:51-58`; `Settings.AuthType string`, `settings.go:31` | Role `auth.discriminator`; `allowedValues` = the 5 editor options |
| `secureJsonData_token` | `token` | `secureJsonData` | `selectors.ts:22` (`label: 'Token'`) | `selectors.ts:23` (`placeholder: 'XXXXXXXX'`) | `SecureSettings.token`, `types.ts:46` | Role `auth.bearer.token`; shown for `Pat`/Unknown (`ConfigEditor.tsx:185`); required by backend (`settings.go:59-65`) |
| `jsonData_azureCredentials` | `azureCredentials` | `jsonData` | — (opaque object; sub-fields rendered by `AzureCredentialsForm`) | Object shape `authUtils.ts:72-77` | `Settings.azureCredentials?: AzureCredentials`, `types.ts:36`; parsed by `azcredentials.FromDatasourceData`, `settings.go:121-139` | `valueType: any` (modeled as `json.RawMessage`); OBO only |
| `secureJsonData_azureClientSecret` | `azureClientSecret` | `secureJsonData` | `selectors.ts:111` (OBO `ClientSecret.label: 'Client Secret'`) | `AzureCredentialsForm.tsx:168` (`placeholder={OauthOBO.ClientSecret.label}` = "Client Secret") | `SecureSettings.azureClientSecret`, `types.ts:48`; `authUtils.ts:79-89` | OBO only |
| `jsonData_clientId` | `clientId` | `jsonData` | `selectors.ts:119` (M2M `label: 'Client ID'`; AzureM2M label "Application (client) ID" `selectors.ts:148`) | `selectors.ts:122` (`placeholder: 'XXXXXXXX-XXXXXXXX-XXXX-XXXXXXXXXXXX'`) | `Settings.clientId?: string`, `types.ts:35`; `Settings.ClientID string json:"clientID"`, `settings.go:32` | Role `auth.oauth2.clientId`; shown for `OauthM2M`+`AzureM2M`; **backend json tag is `clientID` (see findings #1)** |
| `secureJsonData_clientSecret` | `clientSecret` | `secureJsonData` | `selectors.ts:126` (M2M `label: 'Client Secret'`) | `selectors.ts:127` (`placeholder: 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX'`) | `SecureSettings.clientSecret`, `types.ts:47` | Role `auth.oauth2.clientSecret`; loaded for `OauthM2M` (`settings.go:90-98`) and `AzureM2M` (`settings.go:109-113`) |
| `jsonData_tenantId` | `tenantId` | `jsonData` | `selectors.ts:141` (AzureM2M `label: 'Directory (tenant) ID'`) | `selectors.ts:143` (`placeholder: 'XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX'`) | `Settings.tenantId?: string`, `types.ts:34`; `Settings.TenantID string json:"tenantID"`, `settings.go:34` | AzureM2M only (`settings.go:101-104`); **backend json tag is `tenantID` (findings #1)** |
| `jsonData_azureCloud` | `azureCloud` | `jsonData` | `selectors.ts:135` (AzureM2M `label: 'Azure Cloud'`) | Options from `@grafana/azure-sdk` `getAzureClouds` (`clouds.ts`); default `AzureCloud` (`ConfigEditor.tsx:258`, `settings.go:116-118`) | `Settings.azureCloud?: string`, `types.ts:39`; `Settings.AzureCloud string`, `settings.go:36` | AzureM2M only |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (no UI) | Set true for `OauthPT`/`OauthOBO` (`ConfigEditor.tsx:89-91`, `authUtils.ts:94-115`) | `Settings.oauthPassThru?: boolean`, `types.ts:31`; `Settings.OAuthPassThrough bool`, `settings.go:38` | Role `auth.forwardOAuthToken.enabled`; OBO hard-fails without it (`settings.go:141-143`) |
| `jsonData_retries` | `retries` | `jsonData` | `selectors.ts:29` (`label: 'Retries'`) | `selectors.ts:31` (`placeholder: '5'`); backend default "5" (`settings.go:157-159`) | `Settings.retries?: string`, `types.ts:17`; `Settings.Retries string`, `settings.go:25` | Numeric knob stored as a **string** |
| `jsonData_pause` | `pause` | `jsonData` | `selectors.ts:37` (`label: 'Pause'`) | `selectors.ts:38` (`placeholder: '0'`); backend default "0" (`settings.go:154-156`) | `Settings.pause?: string`, `types.ts:18`; `Settings.Pause string`, `settings.go:26` | string |
| `jsonData_timeout` | `timeout` | `jsonData` | `selectors.ts:44` (`label: 'Timeout'`) | `selectors.ts:45` (`placeholder: '60'`); backend default "60" (`settings.go:151-153`) | `Settings.timeout?: string`, `types.ts:15`; `Settings.Timeout string`, `settings.go:24` | Role `transport.timeoutSeconds`; string |
| `jsonData_rows` | `rows` | `jsonData` | `selectors.ts:51` (`label: 'Max Rows'`) | `selectors.ts:52` (`placeholder: '10000'`); parse default 10000 (`connect.go:35`) | `Settings.rows?: string`, `types.ts:20`; `Settings.MaxRows string json:"rows"`, `settings.go:28` | string |
| `jsonData_retryTimeout` | `retryTimeout` | `jsonData` | `selectors.ts:58` (`label: 'Retry Timeout'`) | `selectors.ts:59` (`placeholder: '40'`) | `Settings.retryTimeout?: string`, `types.ts:21`; `Settings.RetryTimeout string`, `settings.go:29` | string |
| `jsonData_debug` | `debug` | `jsonData` | `ConfigEditor.tsx:369` (`label="Debug"`) | — (`Checkbox`) | `Settings.debug?: boolean`, `types.ts:19`; `Settings.Debug bool`, `settings.go:39` | `component: checkbox` |
| `jsonData_enableUnitySupport` | `enableUnitySupport` | `jsonData` | `ConfigEditor.tsx:381` (`label="Unity Catalog Support"`) | Tooltip `ConfigEditor.tsx:383` ("Enable Unity Catalog support for 3-level namespace (catalog.schema.table)") | `Settings.enableUnitySupport?: boolean`, `types.ts:25`; `Settings.EnableUnitySupport bool`, `settings.go:40` | `component: checkbox`; gates resource handler (`main.go:55`) |
| `jsonData_defaultQueryFormat` | `defaultQueryFormat` | `jsonData` | `ConfigEditor.tsx:394` (`label="Default Query Format"`) | Options `types.ts:82-85` (Timeseries=0, Table=1); `QueryFormat` enum `plugin-ui index.d.ts:520-526` | `Settings.defaultQueryFormat?: QueryFormat`, `types.ts:24`; `Settings.DefaultQueryFormat int`, `settings.go:37` | `valueType: number`; `allowedValues` [0,1] |
| `jsonData_cloudFetch` | `cloudFetch` | `jsonData` | — (no UI) | Default `true` mirrors force-enable `settings.go:161-168` | `Settings.CloudFetch bool`, `settings.go:41` | Tagged `backend-only`; no frontend type |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_host` | `host` | `jsonData` | Host | Yes |
| `jsonData_httpPath` | `httpPath` | `jsonData` | Http Path | Yes |
| `jsonData_authType` | `authType` | `jsonData` | Authentication Type | Yes |
| `secureJsonData_token` | `token` | `secureJsonData` | Token | Yes (Pat/Unknown/OauthPT header) |
| `jsonData_azureCredentials` | `azureCredentials` | `jsonData` | — (OBO form) | Yes (OBO) |
| `secureJsonData_azureClientSecret` | `azureClientSecret` | `secureJsonData` | Client Secret | Yes (OBO) |
| `jsonData_clientId` | `clientId` | `jsonData` | Client ID / Application (client) ID | Yes (OauthM2M/AzureM2M) |
| `secureJsonData_clientSecret` | `clientSecret` | `secureJsonData` | Client Secret | Yes (OauthM2M/AzureM2M) |
| `jsonData_tenantId` | `tenantId` | `jsonData` | Directory (tenant) ID | Yes (AzureM2M) |
| `jsonData_azureCloud` | `azureCloud` | `jsonData` | Azure Cloud | Yes (AzureM2M) |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (side-effect) | Yes (OBO guard + core) |
| `jsonData_retries` | `retries` | `jsonData` | Retries | Yes |
| `jsonData_pause` | `pause` | `jsonData` | Pause | Yes |
| `jsonData_timeout` | `timeout` | `jsonData` | Timeout | Yes |
| `jsonData_rows` | `rows` | `jsonData` | Max Rows | Yes |
| `jsonData_retryTimeout` | `retryTimeout` | `jsonData` | Retry Timeout | Yes |
| `jsonData_debug` | `debug` | `jsonData` | Debug | Yes |
| `jsonData_enableUnitySupport` | `enableUnitySupport` | `jsonData` | Unity Catalog Support | Yes |
| `jsonData_defaultQueryFormat` | `defaultQueryFormat` | `jsonData` | Default Query Format | No — frontend-only (declared `int`, never consumed) |
| `jsonData_cloudFetch` | `cloudFetch` | `jsonData` | — (no UI) | Yes (backend-only; force-set) |

### Frontend-only settings

- **`defaultQueryFormat`** is written by the editor (`ConfigEditor.tsx:394-402`) and consumed by
  the frontend query editor (`src/datasource.ts:54`, `src/QueryEditor.tsx:11-12`). The backend
  declares `DefaultQueryFormat int` (`settings.go:37`) but never reads it — it is effectively
  frontend-only. Modeled as `valueType: number` to match the backend struct kind.
- **Not modeled** (declared in `src/types.ts` but neither written by the current config editor nor
  read by the backend `Settings` struct — legacy/query-builder leftovers): `authMech`
  (`types.ts:11`), `ssl` (`types.ts:12`), `thriftTransport` (`types.ts:13`), `uid` (`types.ts:14`),
  `authKind` (`types.ts:16`), `database` (`types.ts:28`, query-builder default). They are documented
  in [`settings.ts`](settings.ts)'s `JsonDataConfig` (the full frontend picture) but intentionally
  excluded from `dsconfig.json`/`Config`, whose bidirectional parity guard scopes the schema to the
  backend-contract fields plus editor-visible config.

### Backend-only settings

- **`cloudFetch`** has no editor UI. It exists in `Settings` (`settings.go:41`), but `LoadSettings`
  force-sets it to `true` on every load (`settings.go:161-168`) unless the `disableCloudFetch`
  Grafana feature toggle is on — so the stored value is effectively ignored. See
  [Upstream findings](#upstream-findings) #4.

## Where the types are defined

Only config type/field definitions are listed — UI components (`AzureCredentialsForm`,
`DataSourceDescription`, `Checkbox`, …) and functions/helpers (`LoadSettings`, `getCredentials`,
`getConnector`, …) are omitted even where they are the reason a field exists.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (jsonData), `SecureSettings`, `AuthenticationType`, `AuthenticationTypeLabel`, `AzureAuthTypes`, `SelectableAuthenticationTypes`, `SelectableQueryFormats` | `src/types.ts:8-85` | plugin (`plugins-private/plugins/grafana-databricks-datasource`) |
| `DataSourceJsonData` (base interface `Settings` extends) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^11.6.7` (catalog) |
| `AzureCredentials`, `AzureAuthType`, `ConcealedSecret` (the `azureCredentials` object shape + `clientsecret-obo` discriminator) | `src/AzureCredentials.ts` | `@grafana/azure-sdk` `0.0.3` (plugin-local pin) |
| `QueryFormat` (numeric enum backing `defaultQueryFormat`) | `dist/esm/index.d.ts:520-526` | `@grafana/plugin-ui` `^0.13.1` (catalog) |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (jsonData + `Token`/`ClientSecret`/`TokenType` secret fields), `ConnectionArgs` | `pkg/models/settings.go:21-42`, `pkg/models/constants.go:30-35` | plugin |
| `AuthenticationType*` string constants | `pkg/authentication/constants.go:4-17` | plugin |
| `AzureClientSecretOboCredentials`, `AzureClientSecretCredentials` (OBO credential shape parsed from `azureCredentials`) | `azcredentials/credentials.go` | `github.com/grafana/grafana-azure-sdk-go/v2` |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`; its root `URL`/`User`/`BasicAuth*` are **unused** by this plugin) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` |
| `httpclient.Options` (the `HttpClientOptions` field, `settings.go:30`) | `backend/httpclient` | `github.com/grafana/grafana-plugin-sdk-go` |

This entry flattens that spread into a single Go `Config` (jsonData fields +
`DecryptedSecureJSONData`) plus `AuthType` and `SecureJsonDataKey` typed constant lists.
`settings.ts` keeps the three canonical TypeScript types (`RootConfig`, `JsonDataConfig`,
`SecureJsonDataConfig`). `RootConfig` is a blank object because the backend reads no root-level
datasource settings.

## Modeling decisions

- **`clientId`/`tenantId` storage keys vs backend json tags.** The config editor writes
  `jsonData.clientId` and `jsonData.tenantId` (`ConfigEditor.tsx:222,274,290`; `types.ts:34-35`),
  so those are the true storage keys used here. The backend `Settings` struct tags the same fields
  `clientID` / `tenantID` (`settings.go:32,34`); Go's case-insensitive JSON unmarshal bridges the
  two, so both the editor's `clientId`/`tenantId` and the upstream test's `clientID`/`tenantID`
  parse correctly (verified by a `LoadConfig` test). The schema and `Config` use the editor keys;
  see [Upstream findings](#upstream-findings) #1.
- **Secrets are `json:"-"` on `Config`, not jsonData fields.** The backend struct declares
  `ClientSecret string json:"clientSecret"` (`settings.go:33`), but that value is populated from
  `DecryptedSecureJSONData["clientSecret"]` (`settings.go:90-98,109-113`), not from jsonData. So
  `token`/`clientSecret`/`azureClientSecret` are modeled as `secureJsonData` keys and carried in
  `Config.DecryptedSecureJSONData` (tagged `json:"-"`), keeping them out of the jsonData spec.
- **Opaque `azureCredentials`.** The OBO credentials object is written atomically by the
  `@grafana/azure-sdk` `AzureCredentialsForm` and parsed downstream by
  `azcredentials.FromDatasourceData` (`settings.go:121-139`), so it is modeled as a single
  `valueType: any` field (`json.RawMessage` in Go), mirroring the azure-monitor entry — not split
  into leaf sub-fields.
- **`oauthPassThru` and `cloudFetch` are managed/backend fields.** Neither is a directly rendered
  editor input: `oauthPassThru` is a side-effect of the auth-type select (`ConfigEditor.tsx:89-91`,
  `authUtils.ts:94-115`) tagged `sdk-managed`; `cloudFetch` is force-enabled by the backend and
  tagged `backend-only`. Both are ungrouped (like the reference entry's backend-only field).
- **`requiredWhen` encodes the backend contract.** The editor shows no required markers, but the
  `validate=true` load path (`driver.go:33` → `settings.go:50-144`) hard-fails without
  host/httpPath and the selected method's inputs. `dependsOn` mirrors the editor's conditional
  render; `requiredWhen` mirrors the backend requirement.
- **Numeric knobs are strings.** `retries`/`pause`/`timeout`/`rows`/`retryTimeout` are stored as
  strings by both the editor (`Input` writes `target.value`) and the backend (`settings.go:24-29`),
  parsed to ints at connect time — so they are `valueType: string`.
- **`defaultQueryFormat` is a number.** `QueryFormat` is a numeric enum
  (`Timeseries=0`, `Table=1`), and the backend reads `int`, so the field is `valueType: number`
  with `allowedValues` [0, 1] (the only two options `SelectableQueryFormats` offers).
- **Secure Socks Proxy excluded.** `ConfigEditor.tsx:404-416` conditionally renders the
  `enableSecureSocksProxy` switch; per registry policy it is omitted from this entry.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle from the
embedded `dsconfig.json`: the `jsonData` fields become the OpenAPI settings `spec`, secure fields
become `secureValues`, and Grafana's datasource API server serves the bundle as
`{apiVersion}.json`.

`SettingsExamples()` provides the default configuration plus one example per authentication method.
Each example is a full instance-settings object with plugin config under `jsonData` and the relevant
write-only secrets under `secureJsonData` (placeholders — replace with real secrets; the default
example, keyed `""`, carries an empty `token` to show what must be filled in):

| Example | Auth (`jsonData.authType`) | `secureJsonData` |
| --- | --- | --- |
| `""` (default) | `Pat` (schema defaults: `cloudFetch:true`) | `token` (empty) |
| `personalAccessToken` | `Pat` | `token` |
| `oauthPassthrough` | `OauthPT` (`oauthPassThru:true`) | — (none required) |
| `oauthM2M` | `OauthM2M` (`clientId`) | `clientSecret` |
| `azureM2M` | `AzureM2M` (`tenantId`, `clientId`, `azureCloud`) | `clientSecret` |
| `azureOBO` | `OauthOBO` (`azureCredentials`, `oauthPassThru:true`) | `azureClientSecret` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase flow and returns a fully-defaulted, validated `Config`:

1. **Parse** — unmarshal jsonData into `Config` (case-insensitively absorbing the backend's
   `clientID`/`tenantID` casing), then copy the decrypted `token`/`clientSecret`/`azureClientSecret`
   into `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — a curated set of zero-valued fields: `AuthType→Pat`
   (`ConfigEditor.tsx:64-74`), `CloudFetch→true` (`settings.go:161-168`), and `AzureCloud→AzureCloud`
   when the method is Azure M2M (`settings.go:116-118`).
3. **`Validate`** — enforce the runtime contract: `host`/`httpPath` present, plus the selected auth
   method's inputs (Pat/Unknown→`token`; OauthM2M→`clientId`+`clientSecret`;
   AzureM2M→`tenantId`+`clientId`+`clientSecret`; OauthOBO→`oauthPassThru`+`azureCredentials`+
   `azureClientSecret`; OauthPT→none). Errors are joined so every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels. This is the intended shape for the plugin's own upstream
`LoadSettings` to sync to.

`(*Config).ApplyDefaults()` and `(Config).Validate() error` stay exported so callers that assemble
a `Config` directly (provisioning preview, schema-example round-trip, tests that distinguish
parse-level from policy-level errors) can invoke each phase individually.

## Upstream findings

Potential bugs, misleading UX, and consistency issues found while researching upstream. The schema
records what the plugin **does**, not what it **should** do; these notes exist so reviewers can
reproduce each finding.

1. **Frontend/backend key-casing mismatch for `clientId`/`tenantId`.** The editor writes
   `jsonData.clientId` and `jsonData.tenantId` (`ConfigEditor.tsx:222,274,290`; `types.ts:34-35`),
   but the backend `Settings` struct tags them `clientID` / `tenantID` (`settings.go:32,34`) and
   `pkg/models/settings_test.go:14-20` marshals the `clientID`/`tenantID` form. It only works
   because Go's `encoding/json` matches object keys case-insensitively. A JavaScript consumer that
   read `jsonData.clientID` (uppercase) would miss the editor-written value. The schema uses the
   editor keys (`clientId`/`tenantId`) as canonical.
2. **`ClientSecret` has a misleading jsonData-looking json tag.** `Settings.ClientSecret string
   json:"clientSecret,omitempty"` (`settings.go:33`) looks like a jsonData field but is only ever
   populated from `DecryptedSecureJSONData` (`settings.go:90-98,109-113`) — it is a secret, not a
   jsonData value. Modeled here as a `secureJsonData` key only.
3. **On-Behalf-Of silently requires "Forward OAuth Identity".** For `OauthOBO`, `LoadSettings`
   returns `ErrInvalidOAuth` ("you must enable Forward OAuth Identity") unless
   `jsonData.oauthPassThru` is true (`settings.go:141-143`). The editor sets `oauthPassThru` as a
   side-effect (`authUtils.ts:96-104`), but a provisioned datasource that omits it fails to load
   with a message that does not name the field.
4. **`cloudFetch` cannot be disabled per-datasource.** `settings.go:161-168` unconditionally sets
   `CloudFetch = true` after parsing, overriding any stored `jsonData.cloudFetch`; it can only be
   turned off globally via the `disableCloudFetch` Grafana feature toggle.
5. **`defaultQueryFormat` is dead on the backend.** `Settings.DefaultQueryFormat int`
   (`settings.go:37`) is parsed but never read anywhere in `pkg/`; the default query format is
   applied entirely in the frontend (`src/QueryEditor.tsx:11-12`).
6. **Community Edition only works with OAuth M2M.** `pkg/database/connect.go:45-47` rejects any
   `community.cloud.databricks.com` host unless `authType == OauthM2M`, returning
   `ErrCommunityEdition` ("databricks community edition doesn't support token based authentication.
   Use credentials instead", `constants.go:13`).
7. **`pause` default guard checks the wrong field.** `pkg/models/settings.go:195-200` (`pause`)
   returns 0 when `settings.Retries == ""` rather than when `settings.Pause == ""` — a likely
   copy/paste bug. It is effectively dead code today because `LoadSettings` defaults `Retries` to
   "5" (`settings.go:157-159`) before `pause` is ever called, so the guard never fires; but a
   refactor that reordered defaulting could regress it. Not encoded in the schema (which only
   carries the placeholder "0").
8. **Editor placeholder vs stored value for numeric knobs.** Retries/Pause/Timeout/Max Rows/Retry
   Timeout show numeric placeholders (`5`/`0`/`60`/`10000`/`40`) but are stored and parsed as
   strings; the backend re-derives defaults at load/connect time, so leaving them blank is valid.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes
  (via the `ConfigSchemaValid` conformance subtest).
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft 2020-12, strict — `additionalProperties: false`) — passes.
- `go test ./...` on the `registry` module — passes (schema round-trip, `SchemaArtifactInSync`
  drift guard, spec/secure separation, jsonData/struct parity both directions, secure-key parity,
  and `LoadConfig`/`ApplyDefaults`/`Validate` table tests incl. the legacy no-authType fallback and
  the `clientID`/`tenantID` casing case).
- `go build ./...`, `go vet ./...`, `gofmt -l .` in `registry/` — clean; the sibling `dsconfig` and
  `schema` workspace modules still build.
- `settings.ts`: `tsc --noEmit --strict` — clean.
