# yesoreyeram-infinity-datasource

Declarative configuration schema for the
[Grafana Infinity datasource plugin](https://github.com/grafana/grafana-infinity-datasource)
(`yesoreyeram-infinity-datasource`).

Infinity is a general-purpose datasource that talks to JSON / CSV / TSV / XML / GraphQL / HTML
endpoints, inline data, Azure Blob storage, and a few synthetic sources — with nine
authentication methods, four kinds of indexed key/value pairs, and a security allow-list. This
entry covers only the **datasource-level configuration**; the per-query editor state
(`InfinityQuery`) is deliberately out of scope.

## Upstream researched

- **Repo**: `github.com/grafana/grafana-infinity-datasource`
- **Ref**: `main`
- **Commit SHA**: `3aede2fe6be90bf5bae3ba3f53a32d7eb5e447a7` (2026-07-03, `Fix PEM line ending normalization for JWT OAuth2 private key, TLS certs and keys (#1673)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, option
labels/values, section titles, defaults, validations, dependency and required-when expressions,
storage keys, storage targets, value types, group titles, and instructions — is traceable to a
specific `file:line` in the upstream repo at this SHA. See [Field provenance](#field-provenance)
below.

To reproduce this research:

```bash
git clone https://github.com/grafana/grafana-infinity-datasource
cd grafana-infinity-datasource
git checkout 3aede2fe6be90bf5bae3ba3f53a32d7eb5e447a7
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under
[Sources researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all datasource-level config fields, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`, and the sub-objects (`OAuth2Props`, `AWSAuthProps`, …) |
| [`settings.go`](settings.go) | Go `Config` model (mirrors upstream `InfinitySettingsJson`), `PluginID`, discriminator typed constants, `SecureJsonDataKey` typed constants, and the `LoadConfig` utility (parse → `ApplyDefaults` → `Validate`) |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth/connection variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`3aede2fe6be90bf5bae3ba3f53a32d7eb5e447a7`).

### Plugin repo (`github.com/grafana/grafana-infinity-datasource@3aede2f`)

| File | What was read |
| --- | --- |
| `src/plugin.json:1-103` | `pluginType` (`id`), `pluginName` (`name`), `docURL` (`info.links[0].url`), `grafanaDependency` (`>=11.6.0-0`) |
| `src/editors/config.editor.tsx:141-150` | The eight config tabs: `Main`, `Authentication`, `URL, Headers & Params`, `Network`, `Security`, `Health check`, `Reference data`, `Global queries` |
| `src/editors/config.editor.tsx:22-75` | `MainEditor` — welcome / documentation / "Setup Authentication" pane (no storage fields) |
| `src/editors/config.editor.tsx:77-98` | `HeadersEditor` — Base URL + Custom HTTP Headers + URL Query Param + URL settings + Cookies |
| `src/editors/config.editor.tsx:100-129` | `NetworkEditor` — Timeout + TLS + Proxy |
| `src/editors/config.editor.tsx:131-139` | `SecurityEditor` — Allowed hosts + Query security |
| `src/editors/config/Auth.tsx:12-23` | Auth method options and labels ("No Auth", "Basic Authentication", "Bearer Token", "API Key Value pair", "Digest Auth", "Forward OAuth", "OAuth2", "AWS", "Azure Blob", "Other Auth Providers") |
| `src/editors/config/Auth.tsx:32-38` | The load-time discriminator (`jsonData.auth_method ?? (basicAuth ? 'basicAuth' : (oauthPassThru ? 'oauthPassThru' : 'none'))`), mirrored as a defaults-only back-fill in `settings.go` |
| `src/editors/config/Auth.tsx:68-86` | `onAuthTypeChange` — the multi-field write per selection (captured in `jsonData_authMethod.effects`) |
| `src/editors/config/Auth.tsx:145-274` | Every auth-method row: labels ("User Name" / "Password" / "Bearer token" / "Key" / "Value" / "Add to" / "Region" / "Service" / "Access Key" / "Secret Key"), placeholders, and the secureJsonData keys each `SecretFormField` writes to |
| `src/editors/config/Auth.tsx:278-283` | AllowedHosts is rendered by AuthEditor for every method except `none` and `azureBlob` |
| `src/editors/config/Auth.AzureBlob.tsx:8-70` | Azure Blob region combobox, storage-account-name / -key inputs |
| `src/editors/config/OAuthInput.tsx:9-190` | OAuth2 grant-type radio (`client_credentials` / `jwt` / `others`) with tooltip; client-credentials fields (Auth Style / Client ID / Client Secret / Token URL / Scopes / Endpoint params); JWT fields (Email / Private Key Identifier / Private Key / Token URL / Subject / Scopes); OAuth2 "others" placeholder |
| `src/editors/config/OAuthInput.tsx:192-234` | `TokenCustomization` — "Custom Token Header" and "Custom Token Template" inputs plus "Token request headers" secure-fields editor |
| `src/editors/config/URL.tsx:8-27` | Base URL editor; empty input is stored as `'__IGNORE_URL__'` (`src/constants.ts:72`) |
| `src/editors/config/URL.tsx:30-62` | URL settings switches: `ignoreStatusCodeCheck`, `allowDangerousHTTPMethods`, `pathEncodedUrlsEnabled` (Experimental badge) |
| `src/editors/config/TLSConfigEditor.tsx:11-125` | TLS toggles and PEM textareas: `Skip TLS Verify`, `With CA Cert`, `CA Cert`, `TLS Client Auth`, `Server Name`, `Client Cert`, `Client Key`; textarea placeholders and `rows: 5` |
| `src/editors/config/ProxyEditor.tsx:19-108` | Proxy Mode radio (`env` / `none` / `url`) and URL/User Name/Password inputs (labels/placeholders/tooltips come from `src/selectors.ts:59-88`) |
| `src/editors/config/ProxyEditor.tsx:109-137` | Secure Socks Proxy field, feature-gated on `secureSocksDSProxyEnabled` — **excluded** per AGENTS.md |
| `src/editors/config/AllowedHosts.tsx:9-26` | Allowed hosts list with placeholder `https://example.com`; hidden when `auth_method === 'azureBlob'` |
| `src/editors/config/SecurityConfigEditor.tsx:8-31` | Query security radio (`allow` / `warn` / `deny`) writing `jsonData.unsecuredQueryHandling` |
| `src/editors/config/CustomHealthCheckEditor.tsx:6-32` | `customHealthCheckEnabled` switch + conditional `customHealthCheckUrl` input |
| `src/editors/config/KeepCookies.tsx:8-24` | `keepCookies` `TagsInput` and tooltip |
| `src/editors/config/ReferenceData.tsx:1-60` | `refData` array of `{ name, data }` — inline reference datasets |
| `src/editors/config/GlobalQueryEditor.tsx:1-80` | `global_queries` array of `{ name, id, query }` where `query` is an opaque per-query editor state |
| `src/components/config/SecureFieldsEditor.tsx:75-113` | Shared indexed-pair writer used by the four SecureFieldsEditor call sites: Custom HTTP Headers (`httpHeaderName<N>` / `httpHeaderValue<N>`), URL Query Param (`secureQueryName<N>` / `secureQueryValue<N>`), OAuth2 endpoint params (`oauth2EndPointParamsName<N>` / `oauth2EndPointParamsValue<N>`), OAuth2 token headers (`oauth2TokenHeadersName<N>` / `oauth2TokenHeadersValue<N>`) |
| `src/selectors.ts:1-105` | Component labels, tooltips, placeholders sourced through the `Components` selector namespace (OAuth2 token customization, Azure Blob, URL settings, Proxy) |
| `src/constants.ts:72,98-134` | `IGNORE_URL` sentinel; `AWSRegions` list; `AzureBlobRegions` list; `AzureBlobCloudTypeDefault` |
| `src/types/config.types.ts:1-86` | Frontend types (`InfinityOptions`, `InfinitySecureOptions`, `AuthType`, `OAuth2Type`, `APIKeyType`, `ProxyType`, `UnsecureQueryHandling`, `AzureBlobCloudType`, `OAuth2Props`, `AWSAuthProps`, `InfinityReferenceData`, `GlobalInfinityQuery`) |
| `pkg/models/settings.go:17-27` | `AuthenticationMethod*` constants |
| `pkg/models/settings.go:29-38` | `AuthOAuthType*` and `ApiKeyType*` constants |
| `pkg/models/settings.go:40-67` | `OAuth2Settings`, `AWSSettings`, `AWSAuthType` structs |
| `pkg/models/settings.go:69-83` | `ProxyType` and `UnsecuredQueryHandlingMode` constants |
| `pkg/models/settings.go:85-133` | Flattened runtime `InfinitySettings` struct (not persisted directly — assembled by `LoadSettings` from `InfinitySettingsJson` + root fields + secrets + indexed-pair maps) |
| `pkg/models/settings.go:135-167` | `InfinitySettings.Validate` — the runtime contract mirrored by `Config.Validate` |
| `pkg/models/settings.go:169-203` | `DoesAllowedHostsRequired` — encoded verbatim in `Config.doesAllowedHostsRequired` and exposed via `Validate` |
| `pkg/models/settings.go:205-236` | `ValidateAllowedHosts` and `FixMissingURLSchema` — schema/protocol checks |
| `pkg/models/settings.go:261-290` | `InfinitySettingsJson` — the persisted jsonData shape (mirrored verbatim by `Config` in `settings.go`) |
| `pkg/models/settings.go:292-425` | `LoadSettings` — parses JSONData, back-fills legacy `basicAuth`/`oauthPassThru` into `auth_method`, defaults `oauth2_type` / `apiKeyType` / `timeoutInSeconds` / `proxy_type` / `unsecuredQueryHandling` / `azureBlobCloudType` + `azureBlobAccountUrl`, aggregates the four indexed-pair sets into maps, and normalizes the `__IGNORE_URL__` sentinel |
| `pkg/models/settings.go:427-443` | `GetSecrets` — the aggregator we mirror as `aggregateSecretPairs` |
| `pkg/pluginschema/dsconfig.json` | The plugin's own in-tree dsconfig schema (older format — no `$schema`, dot-notation IDs, `semanticType`). Used as a cross-reference; this entry's schema uses the current `<target>_<camelCaseKey>` field-id convention |
| `package.json` | External component versions (see next table) |

### External editor components

Read at the exact versions pinned in the plugin's `package.json`.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `RadioButtonGroup`, `Combobox`, `Switch`, `InlineSwitch`, `SecretInput`, `LegacyForms.FormField`, `LegacyForms.SecretFormField`, `TagsInput`, `TextArea`, `Input`, `InlineFormLabel`, `InlineLabel`, `InlineField`, `Stack`, `Grid`, `Card`, `Collapse`, `Badge`, `Divider` | `@grafana/ui@13.0.1` | grafana/grafana packages/grafana-ui | Prop names (`label`, `placeholder`, `tooltip`, `value`, `onChange`, `isConfigured`, `onReset`, `rows`) so we knew which UI attributes to record |
| `onUpdateDatasourceSecureJsonDataOption`, `DataSourcePluginOptionsEditorProps`, `SelectableValue`, `FeatureToggles` | `@grafana/data@13.0.1` | grafana/grafana packages/grafana-data | Storage-key semantics of the update helpers used by the config editor |
| `config` (feature toggles / buildInfo) | `@grafana/runtime@13.0.1` | grafana/grafana packages/grafana-runtime | Only consulted to confirm the Secure Socks Proxy gate (`secureSocksDSProxyEnabled` + `config.buildInfo.version >= "10.0.0"`) — excluded field |

The plugin does **not** depend on `@grafana/plugin-ui` or `@grafana/experimental`; the config
editor composes `@grafana/ui` primitives directly plus in-tree components under
`src/components/config/`.

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each of its
label, placeholder, tooltip, default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `root_url` | `url` | root | `URL.tsx:17` (`Base URL`) | `URL.tsx:21`; empty → `'__IGNORE_URL__'` per `constants.ts:72` + `URL.tsx:12` | `Settings.URL string` `pkg/models/settings.go:98`; TS: `DataSourceSettings.url` | Role `endpoint.baseUrl`; sentinel handling encoded in `Config.LoadConfig` and instruction #6 |
| `root_basicAuth` | `basicAuth` | root | — (managed by `jsonData_authMethod`) | Effect writes at `Auth.tsx:71` | `Settings.BasicAuthEnabled bool` `pkg/models/settings.go:99`; TS: `DataSourceSettings.basicAuth` | Role `auth.basic.enabled`; tagged `managed-by:jsonData_authMethod` |
| `root_basicAuthUser` | `basicAuthUser` | root | `Auth.tsx:148` (`User Name`) | `Auth.tsx:148` (`username`) | `Settings.UserName string` `pkg/models/settings.go:100`; TS: `DataSourceSettings.basicAuthUser` | Role `auth.basic.username`; `dependsOn`/`requiredWhen` from conditional render `Auth.tsx:145` + backend `Validate` `pkg/models/settings.go:136-138` |
| `jsonData_authMethod` | `auth_method` | jsonData | `Auth.tsx:108` (`Auth type`) | Options `Auth.tsx:12-23`; default `'none'` from `Auth.tsx:32-38` | `AuthType` union `src/types/config.types.ts:10`; backend `pkg/models/settings.go:17-27` | Role `auth.discriminator`; effects mirror `onAuthTypeChange` `Auth.tsx:68-86` |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | secureJsonData | `Auth.tsx:159` (`Password`) | `Auth.tsx:161` (`password`); tooltip `Auth.tsx:162` (`password`) | `Settings.Password string` `pkg/models/settings.go:101`; TS `InfinitySecureOptions.basicAuthPassword` | Role `auth.basic.password`; needed for `basicAuth` **and** `digestAuth` (`Auth.tsx:145`) |
| `secureJsonData_bearerToken` | `bearerToken` | secureJsonData | `Auth.tsx:178` (`Bearer token`) | `Auth.tsx:180` (`bearer token`); tooltip `Auth.tsx:181` (`bearer token`) | `Settings.BearerToken string` `pkg/models/settings.go:91`; TS `InfinitySecureOptions.bearerToken` | Role `auth.bearer.token` |
| `jsonData_apiKeyKey` | `apiKeyKey` | jsonData | `Auth.tsx:190` (`Key`) | `Auth.tsx:191` (`api key key`); tooltip `Auth.tsx:192` (`api key key`) | `Settings.ApiKeyKey string` `pkg/models/settings.go:92` | Role `auth.apiKey.key` |
| `secureJsonData_apiKeyValue` | `apiKeyValue` | secureJsonData | `Auth.tsx:207` (`Value`) | `Auth.tsx:209` (`api key value`); tooltip `Auth.tsx:210` | `Settings.ApiKeyValue string` `pkg/models/settings.go:94`; TS `InfinitySecureOptions.apiKeyValue` | Role `auth.apiKey.value` |
| `jsonData_apiKeyType` | `apiKeyType` | jsonData | `Auth.tsx:214` (`Add to`) | Options `Auth.tsx:216-219`; default `'header'` from `Auth.tsx:220` + backend `pkg/models/settings.go:315-317` | `APIKeyType` union `src/types/config.types.ts:12` | Radio |
| `jsonData_oauthPassThru` | `oauthPassThru` | jsonData | — (managed by `jsonData_authMethod`) | Effect writes at `Auth.tsx:74,84` | `Settings.ForwardOauthIdentity bool` `pkg/models/settings.go:102` | Role `auth.forwardOAuthToken.enabled`; tagged `managed-by:jsonData_authMethod` |
| `jsonData_awsRegion` | `region` (in `aws`) | jsonData | `Auth.tsx:229` (`Region`) | `Auth.tsx:230` (`us-east-2`); options `constants.ts:98-126` | `AWSSettings.Region string` `pkg/models/settings.go:65` | Nested under `jsonData.aws` |
| `jsonData_awsService` | `service` (in `aws`) | jsonData | `Auth.tsx:234` (`Service`) | `Auth.tsx:235` (`monitoring`) | `AWSSettings.Service string` `pkg/models/settings.go:66` | Nested under `jsonData.aws` |
| `jsonData_awsAuthType` | `authType` (in `aws`) | jsonData | — (no UI — implicit `'keys'`) | Only value `'keys'` per `AWSAuthType` `pkg/models/settings.go:59-61` | `AWSSettings.AuthType AWSAuthType` `pkg/models/settings.go:64` | Backend-only; the editor never writes it but `Validate` `pkg/models/settings.go:154` reads it |
| `secureJsonData_awsAccessKey` | `awsAccessKey` | secureJsonData | `Auth.tsx:250` (`Access Key`) | `Auth.tsx:252` (`aws access key`); tooltip `Auth.tsx:253` | `Settings.AWSAccessKey string` `pkg/models/settings.go:96`; TS `InfinitySecureOptions.awsAccessKey` | Role `auth.aws.accessKeyId` |
| `secureJsonData_awsSecretKey` | `awsSecretKey` | secureJsonData | `Auth.tsx:265` (`Secret Key`) | `Auth.tsx:267` (`aws secret key`); tooltip `Auth.tsx:268` | `Settings.AWSSecretKey string` `pkg/models/settings.go:97`; TS `InfinitySecureOptions.awsSecretKey` | Role `auth.aws.secretAccessKey` |
| `jsonData_oauth2Type` | `oauth2_type` (in `oauth2`) | jsonData | `OAuthInput.tsx:42` (`Grant Type`) | Options `OAuthInput.tsx:9-13`; default `'client_credentials'` from `OAuthInput.tsx:43` + backend `pkg/models/settings.go:309-311` | `OAuth2Type` union `src/types/config.types.ts:11`; backend `pkg/models/settings.go:30-33` | Nested under `jsonData.oauth2` |
| `jsonData_oauth2AuthStyle` | `authStyle` (in `oauth2`) | jsonData | `OAuthInput.tsx:63` (`Auth Style`) | Options `OAuthInput.tsx:66-70`; default `0` from `OAuthInput.tsx:72` | `OAuth2Settings.AuthStyle` int (0/1/2) `pkg/models/settings.go:48` | Radio; only shown for `client_credentials` |
| `jsonData_oauth2ClientId` | `client_id` (in `oauth2`) | jsonData | `OAuthInput.tsx:76` (`Client ID`) | `OAuthInput.tsx:77` (`Client ID`) | `OAuth2Settings.ClientID string` `pkg/models/settings.go:42` | Role `auth.oauth2.clientId` |
| `secureJsonData_oauth2ClientSecret` | `oauth2ClientSecret` | secureJsonData | `OAuthInput.tsx:88` (`Client Secret`) | `OAuthInput.tsx:90` (`Client secret`) | `OAuth2Settings.ClientSecret string` `pkg/models/settings.go:52` | Role `auth.oauth2.clientSecret` |
| `jsonData_oauth2TokenUrl` | `token_url` (in `oauth2`) | jsonData | `OAuthInput.tsx:94` (`Token URL`) | `OAuthInput.tsx:95` (`Token URL`) | `OAuth2Settings.TokenURL string` `pkg/models/settings.go:43` | Role `auth.oauth2.tokenUrl` |
| `jsonData_oauth2Scopes` | `scopes` (in `oauth2`) | jsonData | `OAuthInput.tsx:98,164` (`Scopes`) | `OAuthInput.tsx:103,170` (`Comma separated values of scopes`) | `OAuth2Settings.Scopes []string` `pkg/models/settings.go:47` | Persisted as a JSON array; editor splits on comma |
| `jsonData_oauth2Email` | `email` (in `oauth2`) | jsonData | `OAuthInput.tsx:126` (`Email`) | `OAuthInput.tsx:127` (`email`) | `OAuth2Settings.Email string` `pkg/models/settings.go:44` | JWT-only |
| `jsonData_oauth2PrivateKeyId` | `private_key_id` (in `oauth2`) | jsonData | `OAuthInput.tsx:132` (`Private Key Identifier`) | `OAuthInput.tsx:133` (`(optional) private key identifier`) | `OAuth2Settings.PrivateKeyID string` `pkg/models/settings.go:45` | JWT-only, optional |
| `secureJsonData_oauth2JWTPrivateKey` | `oauth2JWTPrivateKey` | secureJsonData | `OAuthInput.tsx:145` (`Private Key`) | `OAuthInput.tsx:147` (`Private Key`); tooltip `OAuthInput.tsx:141` | `OAuth2Settings.PrivateKey string` `pkg/models/settings.go:53` | Role `auth.oauth2.jwtPrivateKey`; backend normalizes PEM line endings `settings.go:362-364` |
| `jsonData_oauth2Subject` | `subject` (in `oauth2`) | jsonData | `OAuthInput.tsx:158` (`Subject`) | `OAuthInput.tsx:160` (`(optional) Subject`) | `OAuth2Settings.Subject string` `pkg/models/settings.go:46` | JWT-only, optional |
| `jsonData_oauth2AuthHeader` | `authHeader` (in `oauth2`) | jsonData | `selectors.ts:12` (`Custom Token Header`) | `selectors.ts:15` (`Authorization`); tooltip `selectors.ts:14` | `OAuth2Settings.AuthHeader string` `pkg/models/settings.go:49` | Shown for both `client_credentials` and `jwt` |
| `jsonData_oauth2TokenTemplate` | `tokenTemplate` (in `oauth2`) | jsonData | `selectors.ts:17` (`Custom Token Template`) | `selectors.ts:20` (`Bearer ${__oauth2.access_token}`); tooltip `selectors.ts:19` | `OAuth2Settings.TokenTemplate string` `pkg/models/settings.go:50` | Shown for both `client_credentials` and `jwt` |
| `jsonData_azureBlobCloudType` | `azureBlobCloudType` | jsonData | `selectors.ts:25` (`Azure cloud`) | Options `constants.ts:130-134`; default `'AzureCloud'` `constants.ts:128` + backend `settings.go:403-404` | `Settings.AzureBlobCloudType string` `pkg/models/settings.go:121`; TS `AzureBlobCloudType` union `types/config.types.ts:33` | Combobox |
| `jsonData_azureBlobAccountName` | `azureBlobAccountName` | jsonData | `selectors.ts:30` (`Storage account name`) | `selectors.ts:32` (`Azure blob storage account name`) | `Settings.AzureBlobAccountName string` `pkg/models/settings.go:123` | `requiredWhen` from backend `Validate` `settings.go:146-148` |
| `secureJsonData_azureBlobAccountKey` | `azureBlobAccountKey` | secureJsonData | `selectors.ts:36` (`Storage account key`) | `selectors.ts:38` (`Azure blob storage account key`) | `Settings.AzureBlobAccountKey string` `pkg/models/settings.go:124`; TS `InfinitySecureOptions.azureBlobAccountKey` | Role `auth.azureBlob.storageAccountKey` |
| `jsonData_azureBlobAccountUrl` | `azureBlobAccountUrl` | jsonData | — (no UI) | Backend derives from `azureBlobCloudType` `settings.go:406-415` | `Settings.AzureBlobAccountUrl string` `pkg/models/settings.go:122` | Tagged `backend-only`; provisioning may override the derived URL |
| `jsonData_allowedHosts` | `allowedHosts` | jsonData | `AllowedHosts.tsx:16` (`Allowed hosts`); tooltip `AllowedHosts.tsx:15` | `AllowedHosts.tsx:19` (`https://example.com`) | `Settings.AllowedHosts []string` `pkg/models/settings.go:117` | `dependsOn` mirrors AllowedHostsEditor visibility (`Auth.tsx:278`) + AllowedHostsEditor's own `azureBlob` guard (`AllowedHosts.tsx:10`) |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | jsonData | `TLSConfigEditor.tsx:66` (`Skip TLS Verify`); tooltip `TLSConfigEditor.tsx:65` (`Skip TLS Verify`) | Default `false` per `TLSConfigEditor.tsx:69` | `Settings.InsecureSkipVerify bool` `pkg/models/settings.go:105` | Role `transport.tlsSkipVerify` |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | jsonData | `TLSConfigEditor.tsx:74` (`With CA Cert`); tooltip `TLSConfigEditor.tsx:73` (`Needed for verifying self-signed TLS Certs`) | Default `false` per `TLSConfigEditor.tsx:77` | `Settings.TLSAuthWithCACert bool` `pkg/models/settings.go:109` | Switch |
| `secureJsonData_tlsCACert` | `tlsCACert` | secureJsonData | `TLSConfigEditor.tsx:84` (`CA Cert`) | `TLSConfigEditor.tsx:83` (`Begins with -----BEGIN CERTIFICATE-----`); `rows: 5` `TLSConfigEditor.tsx:86` | `Settings.TLSCACert string` `pkg/models/settings.go:110`; TS `InfinitySecureOptions.tlsCACert` | Role `tls.caCert`; backend normalizes PEM `settings.go:365-367` |
| `jsonData_tlsAuth` | `tlsAuth` | jsonData | `TLSConfigEditor.tsx:93` (`TLS Client Auth`); tooltip `TLSConfigEditor.tsx:92` | Default `false` per `TLSConfigEditor.tsx:96` | `Settings.TLSClientAuth bool` `pkg/models/settings.go:108` | Switch |
| `jsonData_serverName` | `serverName` | jsonData | `TLSConfigEditor.tsx:101` (`Server Name`); tooltip `'Server Name'` | `TLSConfigEditor.tsx:102` (`domain.example.com`) | `Settings.ServerName string` `pkg/models/settings.go:106` | Role `tls.serverName`; conditional on `tlsAuth === true` |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | secureJsonData | `TLSConfigEditor.tsx:107` (`Client Cert`) | `TLSConfigEditor.tsx:106` (`Begins with -----BEGIN CERTIFICATE-----`); `rows: 5` `TLSConfigEditor.tsx:109` | `Settings.TLSClientCert string` `pkg/models/settings.go:111` | Role `tls.clientCert`; backend normalizes PEM `settings.go:368-370` |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | secureJsonData | `TLSConfigEditor.tsx:116` (`Client Key`) | `TLSConfigEditor.tsx:115` (`Begins with -----BEGIN RSA PRIVATE KEY-----`); `rows: 5` `TLSConfigEditor.tsx:118` | `Settings.TLSClientKey string` `pkg/models/settings.go:112` | Role `tls.clientKey`; backend normalizes PEM `settings.go:371-373` |
| `jsonData_timeoutInSeconds` | `timeoutInSeconds` | jsonData | `config.editor.tsx:107` (`Timeout in seconds`) | `config.editor.tsx:111` (`timeout in seconds`); default `60` from `settings.go:326` + editor initial state `config.editor.tsx:102`; `min={0}, max={300}` `config.editor.tsx:112-113` | `Settings.TimeoutInSeconds int64` `pkg/models/settings.go:107` | Role `transport.timeoutSeconds`; range 0..300 |
| `jsonData_proxyType` | `proxy_type` | jsonData | `ProxyEditor.tsx:43` (`Proxy Mode`) | Options `ProxyEditor.tsx:46-50`; default `'env'` `ProxyEditor.tsx:45` + backend `settings.go:333-335` | `ProxyType` type `pkg/models/settings.go:69-75` | Radio |
| `jsonData_proxyUrl` | `proxy_url` | jsonData | `selectors.ts:63` (`Proxy URL`); tooltip `selectors.ts:64` | `selectors.ts:65` (`Example: https://localhost:3004`) | `Settings.ProxyUrl string` `pkg/models/settings.go:114` | `dependsOn` on `proxy_type == 'url'` |
| `jsonData_proxyUsername` | `proxy_username` | jsonData | `selectors.ts:69` (`Proxy User Name`); tooltip `selectors.ts:70-73` | `selectors.ts:74` (`Example: foo`) | `Settings.ProxyUserName string` `pkg/models/settings.go:115` | Optional |
| `secureJsonData_proxyUserPassword` | `proxyUserPassword` | secureJsonData | `selectors.ts:78` (`Proxy Password`); tooltip `selectors.ts:79-82` | `selectors.ts:83` (`Proxy Password`) | `Settings.ProxyUserPassword string` `pkg/models/settings.go:116` | No dedicated role |
| `jsonData_unsecuredQueryHandling` | `unsecuredQueryHandling` | jsonData | `SecurityConfigEditor.tsx:18` (`Query security`); tooltip `SecurityConfigEditor.tsx:17` | Options `SecurityConfigEditor.tsx:22-26`; default `'warn'` from `SecurityConfigEditor.tsx:21` + backend `settings.go:340-342` | `UnsecuredQueryHandlingMode` `pkg/models/settings.go:77-83` | Radio |
| `jsonData_ignoreStatusCodeCheck` | `ignoreStatusCodeCheck` | jsonData | `selectors.ts:47` (`Ignore status code check`); tooltip `selectors.ts:45-46` | Default `false` `URL.tsx:44` | `Settings.IgnoreStatusCodeCheck bool` `pkg/models/settings.go:127` | Switch |
| `jsonData_allowDangerousHTTPMethods` | `allowDangerousHTTPMethods` | jsonData | `selectors.ts:52` (`Allow dangerous HTTP methods`); tooltip `selectors.ts:50-51` | Default `false` `URL.tsx:51` | `Settings.AllowDangerousHTTPMethods bool` `pkg/models/settings.go:128` | Switch |
| `jsonData_pathEncodedUrlsEnabled` | `pathEncodedUrlsEnabled` | jsonData | `selectors.ts:56` (`Encode query parameters with %20`) | Default `false` `URL.tsx:57` | `Settings.PathEncodedURLsEnabled bool` `pkg/models/settings.go:126` | Marked Experimental in the UI (`URL.tsx:58`) |
| `jsonData_keepCookies` | `keepCookies` | jsonData | `KeepCookies.tsx:14` (`Include cookies`); tooltip `KeepCookies.tsx:13` | `KeepCookies.tsx:17` (`Enter the cookie names (enter key to add)`) | `Settings.KeepCookies []string` `pkg/models/settings.go:132` | TagsInput |
| `jsonData_customHealthCheckEnabled` | `customHealthCheckEnabled` | jsonData | `CustomHealthCheckEditor.tsx:12` (`Enable custom health check`) | Default `false` `CustomHealthCheckEditor.tsx:14` | `Settings.CustomHealthCheckEnabled bool` `pkg/models/settings.go:119` | Switch |
| `jsonData_customHealthCheckUrl` | `customHealthCheckUrl` | jsonData | `CustomHealthCheckEditor.tsx:21` (`Health check URL`) | `CustomHealthCheckEditor.tsx:24` (`https://jsonplaceholder.typicode.com/users`) | `Settings.CustomHealthCheckUrl string` `pkg/models/settings.go:120` | `dependsOn` on `customHealthCheckEnabled == true` |
| `jsonData_refData` | `refData` | jsonData | `ReferenceData.tsx` — array of objects | Item placeholders `ReferenceData.tsx:32,35` | `Settings.ReferenceData []RefData` `pkg/models/settings.go:118` | Array of `{name, data}` |
| `jsonData_globalQueries` | `global_queries` | jsonData | `GlobalQueryEditor.tsx` — array of `{name,id,query}` | No item placeholders — full query editor is embedded per row | Frontend-only shape `src/types/config.types.ts:5-9`; each `query` is an `InfinityQuery` | Deliberately opaque (`item.valueType: "any"`) — modeling the query editor jsonData is out of scope |
| `jsonData_isMock` | `is_mock` | jsonData | — (no UI) | Default `false` implicit | `Settings.IsMock bool` `pkg/models/settings.go:88`; `InfinitySettingsJson.IsMock bool` `pkg/models/settings.go:262` | Backend-only test flag |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `root_url` | `url` | root | Base URL | Yes |
| `root_basicAuth` | `basicAuth` | root | — (managed by `jsonData_authMethod`) | Yes |
| `root_basicAuthUser` | `basicAuthUser` | root | User Name | Yes |
| `jsonData_authMethod` | `auth_method` | jsonData | Auth type | Yes |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | secureJsonData | Password | Yes |
| `secureJsonData_bearerToken` | `bearerToken` | secureJsonData | Bearer token | Yes |
| `jsonData_apiKeyKey` | `apiKeyKey` | jsonData | Key | Yes |
| `secureJsonData_apiKeyValue` | `apiKeyValue` | secureJsonData | Value | Yes |
| `jsonData_apiKeyType` | `apiKeyType` | jsonData | Add to | Yes |
| `jsonData_oauthPassThru` | `oauthPassThru` | jsonData | — (managed by `jsonData_authMethod`) | Yes |
| `jsonData_awsRegion` | `aws.region` | jsonData | Region | Yes |
| `jsonData_awsService` | `aws.service` | jsonData | Service | Yes |
| `jsonData_awsAuthType` | `aws.authType` | jsonData | — (no UI) | Yes (backend-only) |
| `secureJsonData_awsAccessKey` | `awsAccessKey` | secureJsonData | Access Key | Yes |
| `secureJsonData_awsSecretKey` | `awsSecretKey` | secureJsonData | Secret Key | Yes |
| `jsonData_oauth2Type` | `oauth2.oauth2_type` | jsonData | Grant Type | Yes |
| `jsonData_oauth2AuthStyle` | `oauth2.authStyle` | jsonData | Auth Style | Yes |
| `jsonData_oauth2ClientId` | `oauth2.client_id` | jsonData | Client ID | Yes |
| `secureJsonData_oauth2ClientSecret` | `oauth2ClientSecret` | secureJsonData | Client Secret | Yes |
| `jsonData_oauth2TokenUrl` | `oauth2.token_url` | jsonData | Token URL | Yes |
| `jsonData_oauth2Scopes` | `oauth2.scopes` | jsonData | Scopes | Yes |
| `jsonData_oauth2Email` | `oauth2.email` | jsonData | Email | Yes |
| `jsonData_oauth2PrivateKeyId` | `oauth2.private_key_id` | jsonData | Private Key Identifier | Yes |
| `secureJsonData_oauth2JWTPrivateKey` | `oauth2JWTPrivateKey` | secureJsonData | Private Key | Yes |
| `jsonData_oauth2Subject` | `oauth2.subject` | jsonData | Subject | Yes |
| `jsonData_oauth2AuthHeader` | `oauth2.authHeader` | jsonData | Custom Token Header | Yes |
| `jsonData_oauth2TokenTemplate` | `oauth2.tokenTemplate` | jsonData | Custom Token Template | Yes |
| `jsonData_azureBlobCloudType` | `azureBlobCloudType` | jsonData | Azure cloud | Yes |
| `jsonData_azureBlobAccountName` | `azureBlobAccountName` | jsonData | Storage account name | Yes |
| `secureJsonData_azureBlobAccountKey` | `azureBlobAccountKey` | secureJsonData | Storage account key | Yes |
| `jsonData_azureBlobAccountUrl` | `azureBlobAccountUrl` | jsonData | — (no UI) | Yes (backend-only, derived) |
| `jsonData_allowedHosts` | `allowedHosts` | jsonData | Allowed hosts | Yes |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | jsonData | Skip TLS Verify | Yes |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | jsonData | With CA Cert | Yes |
| `secureJsonData_tlsCACert` | `tlsCACert` | secureJsonData | CA Cert | Yes |
| `jsonData_tlsAuth` | `tlsAuth` | jsonData | TLS Client Auth | Yes |
| `jsonData_serverName` | `serverName` | jsonData | Server Name | Yes |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | secureJsonData | Client Cert | Yes |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | secureJsonData | Client Key | Yes |
| `jsonData_timeoutInSeconds` | `timeoutInSeconds` | jsonData | Timeout in seconds | Yes |
| `jsonData_proxyType` | `proxy_type` | jsonData | Proxy Mode | Yes |
| `jsonData_proxyUrl` | `proxy_url` | jsonData | Proxy URL | Yes |
| `jsonData_proxyUsername` | `proxy_username` | jsonData | Proxy User Name | Yes |
| `secureJsonData_proxyUserPassword` | `proxyUserPassword` | secureJsonData | Proxy Password | Yes |
| `jsonData_unsecuredQueryHandling` | `unsecuredQueryHandling` | jsonData | Query security | Yes |
| `jsonData_ignoreStatusCodeCheck` | `ignoreStatusCodeCheck` | jsonData | Ignore status code check | Yes |
| `jsonData_allowDangerousHTTPMethods` | `allowDangerousHTTPMethods` | jsonData | Allow dangerous HTTP methods | Yes |
| `jsonData_pathEncodedUrlsEnabled` | `pathEncodedUrlsEnabled` | jsonData | Encode query parameters with %20 | Yes |
| `jsonData_keepCookies` | `keepCookies` | jsonData | Include cookies | Yes |
| `jsonData_customHealthCheckEnabled` | `customHealthCheckEnabled` | jsonData | Enable custom health check | Yes |
| `jsonData_customHealthCheckUrl` | `customHealthCheckUrl` | jsonData | Health check URL | Yes |
| `jsonData_refData` | `refData` | jsonData | Reference data | Yes |
| `jsonData_globalQueries` | `global_queries` | jsonData | — (Global queries tab) | Yes (query editor consumer) |
| `jsonData_isMock` | `is_mock` | jsonData | — (no UI) | Yes (backend-only) |

**Total: 54 fields — 3 root, 39 jsonData, 12 secureJsonData; 0 virtual.**

### Frontend-only settings

None. Every jsonData field this schema declares is read by the backend `LoadSettings`
(`pkg/models/settings.go:292-425`).

### Backend-only settings

- **`aws.authType`** — the editor never writes it (only presents Region/Service inputs plus
  the two secrets); the backend defaults it via provisioning payloads and `Validate` checks
  `AWSSettings.AuthType == "keys"` (`pkg/models/settings.go:154`).
- **`azureBlobAccountUrl`** — a URL template derived from `azureBlobCloudType` on load
  (`pkg/models/settings.go:406-415`). Provisioning payloads may override it, but the editor
  never writes it.
- **`is_mock`** — a boolean used only by the plugin's own tests (swaps in an in-memory mock
  client). Not exposed in the editor.

### Not modeled as first-class fields (indexed-pair storage)

Four sets of key/value pairs written by `SecureFieldsEditor`
(`src/components/config/SecureFieldsEditor.tsx:75-113`) are stored as dynamic keys with
1-based indices — the name in jsonData, the value in secureJsonData. Modeling them as
`indexedPair` storage would require them to appear as first-class schema fields, but the
struct model can't enumerate the resulting dynamic keys (name-value pairs are aggregated into
`map[string]string`s at load time by `pkg/models/settings.go:389-392,427-443`), so this
entry documents the pattern in [instruction #6](dsconfig.json) instead:

| Set | jsonData keys | secureJsonData keys | Aggregated field |
| --- | --- | --- | --- |
| Custom HTTP Headers | `httpHeaderName<N>` | `httpHeaderValue<N>` | `Config.CustomHeaders` |
| URL Query Params | `secureQueryName<N>` | `secureQueryValue<N>` | `Config.SecureQueryFields` |
| OAuth2 endpoint params (client_credentials only) | `oauth2EndPointParamsName<N>` | `oauth2EndPointParamsValue<N>` | `Config.OAuth2EndpointParams` |
| OAuth2 token request headers | `oauth2TokenHeadersName<N>` | `oauth2TokenHeadersValue<N>` | `Config.OAuth2TokenHeaders` |

## Where the types are defined

Only config type/field definitions are listed — UI components (e.g. `SecureFieldsEditor`,
`TLSConfigEditor`, `AuthEditor`) and functions/helpers (e.g. `LoadSettings`, `GetSecrets`,
`convertLegacyAuthProps`) are omitted even where they are the reason a field exists.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `InfinityOptions` (jsonData), `InfinitySecureOptions`, `AuthType`, `OAuth2Type`, `APIKeyType`, `OAuth2Props`, `AWSAuthProps`, `InfinityReferenceData`, `GlobalInfinityQuery`, `ProxyType`, `UnsecureQueryHandling`, `AzureBlobCloudType` | `src/types/config.types.ts:1-86` | plugin ([grafana/grafana-infinity-datasource](https://github.com/grafana/grafana-infinity-datasource)) |
| `SecureField` (rendered by the shared `SecureFieldsEditor`; not persisted directly) | `src/types/config.types.ts:79-84` | plugin |
| `DataSourceJsonData` (base interface `InfinityOptions` extends: `authType`, `defaultRegion`, `profile`, `manageAlerts`, …) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `13.0.1` |
| `DataSourceInstanceSettings.url`, `.basicAuth`, `.basicAuthUser`, `.withCredentials`, `.secureJsonFields` | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `13.0.1` |
| `SecureSocksProxyConfig` interface adding the `enableSecureSocksProxy` jsonData field (excluded from this entry) | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | `@grafana/ui` `13.0.1` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `InfinitySettings` (flattened runtime settings), `InfinitySettingsJson` (persisted jsonData), `OAuth2Settings`, `AWSSettings`, `RefData`, `AuthenticationMethod*` constants, `ProxyType`, `UnsecuredQueryHandlingMode`, `AWSAuthType` | `pkg/models/settings.go:17-290` | plugin ([grafana/grafana-infinity-datasource](https://github.com/grafana/grafana-infinity-datasource)) |
| `backend.DataSourceInstanceSettings` (carries `URL`, `BasicAuthEnabled`, `BasicAuthUser`, `JSONData`, `DecryptedSecureJSONData`) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |
| `httpclient.Options` (proxy / timeouts / TLS options) | `backend/httpclient` | `github.com/grafana/grafana-plugin-sdk-go` `v0.292.1` |
| `oauth2.AuthStyle` (int enum stored as `authStyle`: 0=Auto, 1=In Params, 2=In Header) | `golang.org/x/oauth2` | `golang.org/x/oauth2` (via plugin `go.mod`) |

The models in this entry flatten that spread into a single Go `Config` struct (jsonData
fields + root fields the plugin actually reads + aggregated indexed-pair maps +
`DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant list. `settings.ts` keeps
the three canonical TypeScript types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`).

## Modeling decisions

- **`auth_method` as the sole discriminator**: the effect block on `jsonData_authMethod`
  mirrors `Auth.tsx:68-86`'s `onAuthTypeChange` so that selecting `basicAuth` sets
  `root.basicAuth=true` + `jsonData.oauthPassThru=false`, `oauthPassThru` sets both flags
  accordingly, and every other method zeroes both flags. `Config.ApplyDefaults` also runs the
  reverse-direction back-fill (`basicAuth` → `auth_method='basicAuth'`, `oauthPassThru` →
  `auth_method='oauthPassThru'`) so legacy provisioning payloads that predate `auth_method`
  still resolve.
- **Nested `aws` and `oauth2` fields via `section`**: instead of modeling them as `object`
  fields with a nested `item.fields`, each leaf is declared with `section: "aws"` (or
  `"oauth2"`) plus a bare `key` (`region`, `service`, `oauth2_type`, `client_id`, …). This
  keeps schema-to-struct parity clean (the conformance walker joins `section.key` and
  compares against the struct's json tags on `AWSSettings` / `OAuth2Settings`).
- **Indexed-pair fields not modeled**: Custom HTTP Headers, URL Query Params, OAuth2 endpoint
  params, and OAuth2 token headers all use the shared `SecureFieldsEditor`, which writes
  dynamic `<prefix>Name<N>` / `<prefix>Value<N>` keys. Materializing them as first-class
  schema fields would break the struct-parity conformance check (their dynamic keys have no
  static struct counterparts — `LoadConfig` aggregates them into `map[string]string`s
  instead). They are documented in the "Connection, URL, and indexed-pair storage"
  instruction and via the `Config.CustomHeaders` / `SecureQueryFields` /
  `OAuth2EndpointParams` / `OAuth2TokenHeaders` aggregation fields.
- **`global_queries` as `array<any>`**: the individual `InfinityQuery` shape (owned by the
  query editor) is intentionally opaque at the datasource-config level. The schema records
  the storage-target and array-shape only.
- **`__IGNORE_URL__` sentinel**: preserved verbatim in `Config.LoadConfig` (normalized back
  to `""` on load, exactly as `pkg/models/settings.go:296-298` does). The schema's
  `root_url.role: "endpoint.baseUrl"` still applies; the sentinel is a wire-only concern.
- **Secure Socks Proxy excluded**: `ProxyEditor.tsx:109-137` conditionally renders the
  Grafana secure-socks toggle, writing `jsonData.enableSecureSocksProxy`. Per AGENTS.md the
  field is omitted from this entry.
- **6-instruction budget**: the AGENTS.md limit is honored. Coverage: (1) auth methods and
  discriminator, (2) minimal-payload recipes for the eight non-OAuth2 methods, (3) OAuth2
  recipes and defaults, (4) legacy auth interpretation, (5) write-only secrets and
  `secureJsonFields`, (6) connection / URL rules + indexed-pair storage.
- **Field ID naming convention**: `<target>_<camelCaseKey>` — `root_`, `jsonData_`, or
  `secureJsonData_` prefix followed by the camelCase storage key
  (e.g. `jsonData_authMethod`, `secureJsonData_bearerToken`). The `key` property keeps the
  plugin's raw storage key (`auth_method`, `bearerToken`) — `id` is the schema reference,
  `key` is the storage contract.
- **Flat `Config` in Go**: `settings.go` collapses jsonData fields, the nested `oauth2` and
  `aws` sub-objects, root-level fields the plugin reads (`URL`, `BasicAuth`, `BasicAuthUser`),
  and decrypted secrets onto a single `Config` struct with `OAuth2Settings`/`AWSSettings`
  sub-structs mirroring the upstream shape verbatim.
- **`SecureJsonDataConfig` is a key list**: secure values are write-only, so the secure type
  is just the array of fixed-name secret keys (dynamic indexed-value secrets are enumerated
  at runtime via `Config.CustomHeaders` etc., not statically).

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle (the
k8s-style schema Grafana's datasource API server serves as `{apiVersion}.json`, `v0alpha1`
today) from the embedded `dsconfig.json`.

`SettingsExamples()` provides the default configuration plus one k8s-style example per
authentication method and a handful of common connection/security variants.

| Example | Auth | Extra flavor | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | None | no URL / defaults only | `bearerToken` (empty) |
| `basicAuth` | Basic auth | `root.basicAuth=true` + basic user | `basicAuthPassword` |
| `bearerToken` | Bearer token | — | `bearerToken` |
| `apiKeyHeader` | API key | `apiKeyType='header'` | `apiKeyValue` |
| `apiKeyQuery` | API key | `apiKeyType='query'` | `apiKeyValue` |
| `digestAuth` | Digest auth | `root.basicAuth` stays false | `basicAuthPassword` |
| `forwardOAuth` | Forward OAuth | — | (none) |
| `oauth2ClientCredentials` | OAuth2 | client_credentials grant with `allowedHosts` | `oauth2ClientSecret` |
| `oauth2JWT` | OAuth2 | JWT grant (service account) | `oauth2JWTPrivateKey` |
| `awsSigV4` | AWS | authType='keys' + region + service | `awsAccessKey`, `awsSecretKey` |
| `azureBlob` | Azure Blob | `AzureCloud` + storage account name | `azureBlobAccountKey` |
| `tlsMutualAuth` | none | `tlsAuth=true` + serverName | `tlsClientCert`, `tlsClientKey` |
| `tlsCustomCA` | none | `tlsAuthWithCACert=true` | `tlsCACert` |
| `customProxy` | none | `proxy_type='url'` | `proxyUserPassword` |
| `customHeadersAndQueryParams` | none | 1× indexed HTTP header + 1× indexed query param | `httpHeaderValue1`, `secureQueryValue1` |
| `referenceData` | none | 1× `refData` entry | (none) |
| `customHealthCheck` | none | `customHealthCheckEnabled` + URL | (none) |
| `legacyBasicAuthWithoutMethod` | Legacy (empty `auth_method`) | only `root.basicAuth=true` — back-filled by LoadConfig | `basicAuthPassword` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns a
fully-defaulted, validated `Config`:

1. **Parse** — copy `URL`, `BasicAuthEnabled`, `BasicAuthUser` from the SDK
   `DataSourceInstanceSettings`; normalize `URL == "__IGNORE_URL__"` to `""` (mirrors
   `pkg/models/settings.go:296-298`); unmarshal `settings.JSONData` into `Config`; copy the
   fixed-name decrypted secrets into `DecryptedSecureJSONData`; aggregate the four indexed
   pair sets into `CustomHeaders` / `SecureQueryFields` / `OAuth2EndpointParams` /
   `OAuth2TokenHeaders` (mirrors `pkg/models/settings.go:427-443`).
2. **`ApplyDefaults`** — back-fill `auth_method` from `basicAuth` / `oauthPassThru` legacy
   flags; default `oauth2_type='client_credentials'` under `auth_method='oauth2'`; default
   `apiKeyType='header'`, `timeoutInSeconds=60`, `proxy_type='env'`,
   `unsecuredQueryHandling='warn'`; and for `auth_method='azureBlob'`, default
   `azureBlobCloudType='AzureCloud'` + fill the matching `azureBlobAccountUrl` template
   (mirrors `pkg/models/settings.go:307-416`).
3. **`Validate`** — enforce the plugin's runtime contract (mirrors
   `InfinitySettings.Validate` `pkg/models/settings.go:135-167` plus
   `DoesAllowedHostsRequired` `:169-203` plus `ValidateAllowedHosts` `:205-222`). Errors are
   joined so every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels so log lines carry request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are exported for callers that
want to compose them themselves (e.g. provisioning preview, schema-example round-trip, tests
that need to distinguish parse-level from policy-level errors).

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching upstream.
All preserved verbatim in the schema — the schema records what the plugin **does**, not what
it **should** do; these notes exist so reviewers can reproduce each finding and decide
separately whether to fix upstream.

1. **`auth_method === 'others'` is inert.** The Auth type card list includes an "Other Auth
   Providers" tile (`Auth.tsx:22`), but selecting it opens a discussion-link drawer instead
   of writing anything to storage (`Auth.tsx:117-133`, `OtherAuthProviders.tsx`). It never
   sets `jsonData.auth_method='others'`. The OAuth2 grant-type radio also has an "Others"
   value (`OAuthInput.tsx:12`), but selecting it renders a link-only message and no OAuth
   flow runs — the backend just leaves `OAuth2Settings.OAuth2Type` at `'others'` with no
   token acquisition.
2. **AllowedHosts is enforced by the backend but the editor doesn't validate protocol.**
   `AllowedHosts.tsx:19` accepts any string; the backend `ValidateAllowedHosts`
   (`pkg/models/settings.go:205-222`) rejects entries without `http://`/`https://` and with
   an empty hostname. A provisioning payload that reads `["example.com"]` breaks the
   datasource silently until the first query.
3. **AllowedHosts requirement is not shown as "required" in the editor.** `Auth.tsx:278-283`
   renders the AllowedHosts editor for every auth method except `none` and `azureBlob`,
   with no required-star marker. When `root.url` is empty, the backend hard-fails on load
   (`ErrInvalidConfigHostNotAllowed`, `pkg/models/settings.go:169-203`).
4. **`aws.authType` has no UI.** `Auth.tsx:226-272` presents Region / Service / Access Key
   / Secret Key inputs but never sets `jsonData.aws.authType`. Provisioning payloads must
   explicitly set `authType: 'keys'` because `Validate` `pkg/models/settings.go:154` checks
   for it. A missing `authType` skips the credential-required check silently.
5. **Timeout is clamped by the editor, not the backend.** `config.editor.tsx:112-113`
   restricts the timeout input to `0..300`, but `LoadSettings` accepts any positive int64
   (`pkg/models/settings.go:336-338`). Provisioning can set `timeoutInSeconds: 3600` and it
   will be honored.
6. **`azureBlobAccountUrl` is derived but overridable.** `LoadSettings` only fills it in
   when it is empty (`pkg/models/settings.go:406-415`). A provisioning payload that supplies
   a custom template (e.g. for a private Azure Stack endpoint) is preserved verbatim; the
   editor has no UI to set it and no UI to see whether it has been overridden.
7. **`oauth2.oauth2_type` back-fills to `'client_credentials'` only under oauth2 auth.**
   `pkg/models/settings.go:309-311` — a provisioning payload that sets `oauth2` fields
   without `oauth2_type` gets the default only when `auth_method === 'oauth2'`; leftover
   `oauth2` state on a datasource that switched auth methods is preserved but ignored.
8. **`unsecuredQueryHandling='allow'` bypasses the allow-list protection entirely.** The
   backend only aborts on per-query secrets when this is set to `'warn'` (default) or
   `'deny'` (`pkg/models/settings.go:77-83`). The editor exposes the choice as a
   radio without warning about the security implications; the tooltip
   `SecurityConfigEditor.tsx:17` only says "Option to handle insecure query content such as
   sensitive headers in the dashboard query".
9. **Scopes are stored as a JSON array but edited as a comma-separated string.**
   `OAuthInput.tsx:100-101` splits on `,` and stores as `string[]`, but the input has no
   validation — trailing whitespace, empty entries, or misplaced spaces all round-trip into
   the stored scope list.
10. **`pathEncodedUrlsEnabled` is marked Experimental in the UI.** `URL.tsx:58` renders an
    orange "Experimental" badge next to the switch. The backend
    (`pkg/models/settings.go:126,330`) treats it as a stable feature.
11. **`is_mock: true` bypasses all real HTTP calls.** `InfinitySettings.IsMock`
    (`pkg/models/settings.go:88`) is not editor-exposed; it exists for plugin unit tests
    only. A provisioning payload that sets `jsonData.is_mock: true` will silently return
    canned responses from every query.
12. **Duplicate secure-field naming between `oauth2EndPointParamsName` and
    `oauth2EndPointParamsValue`.** Note the mixed capitalization (`EndPoint`, not
    `Endpoint`), preserved verbatim throughout the plugin. Provisioning payloads must match
    this casing to be aggregated by `GetSecrets` (`pkg/models/settings.go:427-443`).

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft-07,
  `additionalProperties: false`) — passes (exercised by the shared conformance suite).
- `go test ./...` on the shared `registry/` module — passes (schema bundle shape, secure
  values, examples, `LoadConfig` incl. legacy back-fill and indexed-pair aggregation,
  `SchemaArtifactInSync` guard).
- `settings.go`/`schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: `tsc --noEmit --strict` — clean.
