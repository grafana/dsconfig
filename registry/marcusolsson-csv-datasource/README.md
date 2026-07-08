# marcusolsson-csv-datasource

Declarative configuration schema for the [Grafana CSV datasource plugin](https://github.com/grafana/grafana-csv-datasource) (`marcusolsson-csv-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/grafana-csv-datasource`
- **Ref**: `main`
- **Commit SHA**: `5de046675ad9d359f5253d454f8ec35e8462a515`

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders,
tooltips, option labels/values, section titles, defaults, validations,
dependency and required-when expressions, storage keys, storage targets,
value types, group titles, and instructions — is traceable to a specific
`file:line` in the upstream repo at this SHA. See [Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone --depth 20 https://github.com/grafana/grafana-csv-datasource \
  /var/folders/2l/6dq44779565gnyvkpmhhv1th0000gn/T/opencode/grafana-csv-datasource
git -C /var/folders/2l/6dq44779565gnyvkpmhhv1th0000gn/T/opencode/grafana-csv-datasource \
  checkout 5de046675ad9d359f5253d454f8ec35e8462a515
```

If upstream `main` has advanced past this SHA, re-diff the sources listed
under [Sources researched](#sources-researched) before merging any changes
to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`, `StorageMode` |
| [`settings.go`](settings.go) | Go `Config` model (flat: root URL/basicAuth/basicAuthUser/withCredentials tagged `json:"-"`, jsonData fields including the two CSV-specific keys `storage` and `queryParams`, and `DecryptedSecureJSONData`), `PluginID`, `StorageMode` / `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for the default configuration, HTTP variants (no-auth, basic auth, OAuth-forward, mTLS, self-signed CA, advanced HTTP), local-file storage, and a legacy empty-storage example |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of
the shared [`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA (`5de0466`), plus
external editor components at the exact versions the plugin's
`package.json` pins.

### Plugin repo (`github.com/grafana/grafana-csv-datasource@5de0466`)

| File | What was read |
| --- | --- |
| `src/plugin.json:1-47` | `pluginType` (`id` = `"marcusolsson-csv-datasource"`), `pluginName` (`name` = `"CSV"`), docs URL (`info.links[0].url` = `"https://grafana.com/docs/plugins/marcusolsson-csv-datasource/latest/"`), `backend: true`, `alerting: true`, `logs: true`, `annotations: true`, `metrics: true`, `grafanaDependency: ">=11.6.0-0"` |
| `src/ConfigEditor.tsx:1-136` | The full config editor. `DataSourceDescription` (`:56-60`, `hasRequiredFields={false}`), Storage Location RadioButtonGroup (`:64-73`), HTTP branch (`:77-112`) composing `ConnectionSettings` (urlPlaceholder `:82`), `Auth` via `convertLegacyAuthProps` (`:87-92`), collapsible `ConfigSection title="Additional settings"` (`:96-110`) containing `AdvancedHttpSettings` and the Custom query parameters Field/Input (`:101-108`, label / description / placeholder), Local branch (`:114-124`, `<Field label="Path">` + Input placeholder) |
| `src/types.ts:1-51` | Frontend types. `CSVDataSourceOptions extends DataSourceJsonData { storage?: string; queryParams?: string; }` (`:42-45`) — only two datasource-level jsonData fields. `defaultOptions = { storage: 'http' }` (`:47-49`). Every other field (delimiter, header, decimalSeparator, ignoreUnknown, skipRows, schema, timezone, method, path, params, headers, body, experimental.regex) is per-query on `CSVQuery` (`:9-28`), not persisted on the datasource |
| `src/utils.ts:4-10` | `getOptionsWithDefaults` — merges `defaultOptions` into `jsonData` only when `jsonData.storage` is missing, applying `storage: 'http'` for backwards compatibility |
| `pkg/settings.go:10-28` | Backend `PluginSettings { Storage string, QueryParams string }`. `LoadPluginSettings` unmarshals `settings.JSONData`, then normalizes empty Storage to `"http"` for backwards compatibility (`:22-24`) |
| `pkg/datasource.go:24-47` | `NewDatasource` — reads `os.Getenv("GF_PLUGIN_ALLOW_LOCAL_MODE") == "true"` into `allowLocalMode` (`:44`), builds an HTTP client via `settings.HTTPClientOptions(ctx)` + `httpclient.New` (`:35-42`) |
| `pkg/datasource.go:109-145` | `CheckHealth` — requires `settings.URL != ""` when `Storage == "http"` (`:121-125`), then calls `store.Stat` (`:135-140`) which for local storage `os.Stat`s the URL as a filesystem path |
| `pkg/datasource.go:147-171` | `newStorage` — rejects local mode when `!allowLocalMode` with `"local mode has been disabled by your administrator"` (`:158-160`), else dispatches to `newLocalStorage` or `newHTTPStorage` (default HTTP) |
| `pkg/http_storage.go:1-131` | HTTP storage. `newRequestFromQuery` (`:73-131`) parses `settings.URL + query.Path` (`:79-93`), then merges `customSettings.QueryParams` (parsed with `url.ParseQuery`) INTO the request query params with the admin's values overriding per-query on collision (`:102-109`) |
| `pkg/local_storage.go:1-51` | Local storage. `Open` uses `settings.URL` as either the file path directly (when `query.Path == ""`, `:34-36`) or as a base directory that `query.Path` is joined onto with a `HasPrefix` escape check (`:37-45`). `Stat` `os.Stat`s the URL (`:48-50`) |
| `pkg/csv.go:19-344` | `csvOptions` — per-query CSV parsing options (delimiter, header, decimalSeparator, skipRows, schema, ignoreUnknown, timezone). Not datasource-level; not modeled here |
| `go.mod:5` | `grafana-plugin-sdk-go v0.292.0` |
| `package.json:31-38` | External component versions (see next table) |

Notably absent: no `LoadSettings` on the frontend that writes defaults into
`jsonData` on load; no admission handler; no `UnmarshalJSON` overriding
the raw parsing. Both the frontend (`utils.ts`) and backend
(`pkg/settings.go`) treat an empty `storage` value as `"http"` at read
time only — a datasource stored with `jsonData: {}` and no storage key
continues to work.

### External editor components

Read at the exact versions pinned in the plugin's `package.json`
(`@grafana/plugin-ui@0.13.1`, `@grafana/ui@13.1.0-25893932881`,
`@grafana/data@13.1.0-25893932881`, `@grafana/schema@13.1.0-25893932881`,
`@grafana/runtime@13.1.0-25893932881`).

| Component | Version | Source consulted (from `npm pack @grafana/plugin-ui@0.13.1`) | What was read |
| --- | --- | --- | --- |
| `DataSourceDescription` | `@grafana/plugin-ui@0.13.1` | `dist/esm/components/ConfigEditor/DataSourceDescription.js` | Renders "Before you can use the {dataSourceName} data source, you must configure it below or in the config file. For detailed instructions, view the documentation." Suppresses the "Fields marked with * are required" note when `hasRequiredFields=false` |
| `ConnectionSettings` | `@grafana/plugin-ui@0.13.1` | `dist/esm/components/ConfigEditor/Connection/ConnectionSettings.js` | Default `urlLabel = 'URL'` (`:26`), default tooltip "Specify a complete HTTP URL (for example https://example.com:8080)" (`:28`), URL regex validator (`:16`), required=true (`:31`) |
| `Auth` + `convertLegacyAuthProps` | `@grafana/plugin-ui@0.13.1` | `dist/esm/components/ConfigEditor/Auth/Auth.js`, `dist/esm/components/ConfigEditor/Auth/utils.js`, `dist/esm/components/ConfigEditor/Auth/auth-method/AuthMethodSettings.js` | ConfigSection title `"Authentication"` (`Auth.js:27`). Default `visibleMethods = [BasicAuth, OAuthForward, NoAuth]` (`AuthMethodSettings.js:47-52`) — CrossSiteCredentials is deliberately hidden. Option labels: "Basic authentication", "Enable cross-site access control requests" (hidden), "Forward OAuth Identity", "No Authentication" (`AuthMethodSettings.js:9-30`). `convertLegacyAuthProps` writes BOTH `root.basicAuth` and `jsonData.oauthPassThru` on every selection (`utils.js:36-48`) |
| `BasicAuth` | `@grafana/plugin-ui@0.13.1` | `dist/esm/components/ConfigEditor/Auth/auth-method/BasicAuth.js` | Labels default to `User` / `Password` with tooltips "The username of the data source account" / "The password of the data source account", placeholders `User` / `Password`. Writes `root.basicAuthUser` + `secureJsonData.basicAuthPassword` |
| `TLSSettings` | `@grafana/plugin-ui@0.13.1` | `dist/esm/components/ConfigEditor/Auth/tls/TLSSettings.js` | ConfigSubSection title "TLS settings" with description "Additional security measures that can be applied on top of authentication" |
| `SelfSignedCertificate` | `@grafana/plugin-ui@0.13.1` | `dist/esm/components/ConfigEditor/Auth/tls/SelfSignedCertificate.js` | Label "Add self-signed certificate", tooltip "Add your own Certificate Authority (CA) certificate on top of one generated by the certificate authorities for additional security measures". CA Certificate textarea label "CA Certificate", tooltip "Your self-signed certificate", placeholder `Begins with --- BEGIN CERTIFICATE ---`. Writes `jsonData.tlsAuthWithCACert` + `secureJsonData.tlsCACert` |
| `TLSClientAuth` | `@grafana/plugin-ui@0.13.1` | `dist/esm/components/ConfigEditor/Auth/tls/TLSClientAuth.js` | Label "TLS Client Authentication", tooltip "Validate using TLS client authentication, in which the server authenticates the client". Inner labels: "ServerName" (placeholder `domain.example.com`), "Client Certificate" (placeholder `Begins with --- BEGIN CERTIFICATE ---`), "Client Key" (placeholder `Begins with --- RSA PRIVATE KEY CERTIFICATE ---`). Writes `jsonData.tlsAuth`, `jsonData.serverName`, `secureJsonData.tlsClientCert`, `secureJsonData.tlsClientKey` |
| `SkipTLSVerification` | `@grafana/plugin-ui@0.13.1` | `dist/esm/components/ConfigEditor/Auth/tls/SkipTLSVerification.js` | Label "Skip TLS certificate validation", tooltip "Skipping TLS certificate validation is not recommended unless absolutely necessary or for testing". Writes `jsonData.tlsSkipVerify` |
| `CustomHeaders` (excluded) | `@grafana/plugin-ui@0.13.1` | `dist/esm/components/ConfigEditor/Auth/custom-headers/…` and `dist/esm/components/ConfigEditor/Auth/utils.js:160-197` | Indexed `httpHeaderName<N>` (jsonData) / `httpHeaderValue<N>` (secureJsonData) storage pattern starting at N=1. **Not modeled** as first-class fields (see [Modeling decisions](#modeling-decisions)) |
| `AdvancedHttpSettings` | `@grafana/plugin-ui@0.13.1` | `dist/esm/components/ConfigEditor/AdvancedSettings/AdvancedHttpSettings.js` | ConfigSubSection title "Advanced HTTP settings". Allowed cookies TagsInput label "Allowed cookies" (`:39`), tooltip "Grafana proxy deletes forwarded cookies by default. Specify cookies by name that should be forwarded to the data source." (`:41`), placeholder "New cookie (hit enter to add)" (`:49`). Timeout Input label "Timeout" (`:58`), tooltip "HTTP request timeout in seconds" (`:60`), placeholder "Timeout in seconds" (`:70`), `type="number"` `min={0}` (`:68-69`), `parseInt(value, 10)` (`:26`) |
| `ConfigSection` | `@grafana/plugin-ui@0.13.1` | `dist/esm/components/ConfigEditor/ConfigSection/ConfigSection.js` | Renders a titled section with `isCollapsible` support — used verbatim as the "Additional settings" wrapper (`ConfigEditor.tsx:96`) |
| `Field`, `Input`, `RadioButtonGroup`, `Divider`, `useStyles2` | `@grafana/ui@13.1.0-25893932881` | `packages/grafana-ui/…` | Prop names (`label`, `description`, `options`, `value`, `onChange`, `spellCheck`, `width`, `placeholder`) — used directly by `src/ConfigEditor.tsx` |
| `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps`, `GrafanaTheme2` | `@grafana/data@13.1.0-25893932881` | `packages/grafana-data/…` | Base jsonData interface, editor prop shape, theme type |
| `DataQuery` | `@grafana/schema@13.1.0-25893932881` | `packages/grafana-schema/…` | Base of `CSVQuery` — not directly relevant to datasource config, listed only because `types.ts:2` imports it |

Note: the CSV editor does **not** use `@grafana/ui`'s deprecated
`DataSourceHttpSettings` at all. It composes the newer
`@grafana/plugin-ui@0.13.1` components — `ConnectionSettings`, `Auth`
(with `convertLegacyAuthProps`), `AdvancedHttpSettings`,
`DataSourceDescription`, `ConfigSection` — the same modern set used by
Parca and other newer datasource plugins.

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream
`file:line` where each of its label, placeholder, tooltip, default,
storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_storage` | `storage` | `jsonData` | `ConfigEditor.tsx:64` (`Field label="Storage Location"`) | RadioButtonGroup options at `ConfigEditor.tsx:66-69`; default `"http"` from `types.ts:47-49` + `pkg/settings.go:22-24` | `string` — `types.ts:43`, `pkg/settings.go:11` | `allowedValues` restricts to `"http"` / `"local"` |
| `root_url` | `url` | `root` | `ConnectionSettings.js:26` (`urlLabel || 'URL'`); Local branch `ConfigEditor.tsx:115` (`label="Path"`, not overridable in dsconfig) | HTTP: `ConfigEditor.tsx:82` (`urlPlaceholder="http://localhost:8080"`); Local: `ConfigEditor.tsx:121` (`placeholder="Path to the CSV file"`) | Root SDK `settings.URL string` | Required per `pkg/datasource.go:121-125` (http). Dual-purpose per `pkg/local_storage.go:33-45`. `overrides` swap the placeholder + description under `jsonData_storage == 'local'` — dsconfig cannot swap the field label to `"Path"` today; see [Upstream findings](#upstream-findings) #4 |
| `virtual_authMethod` | `authMethod` (virtual) | — | Convention (mirrors the parca / mock entries and the plugin-ui Auth dropdown title `"Authentication method"`) | Options from `AuthMethodSettings.js:9-30` and default `visibleMethods` `[BasicAuth, OAuthForward, NoAuth]` at `:47-52`; default `"NoAuth"` per `utils.js:24-34` | `string` (discriminator) | Computed from `root.basicAuth` and `jsonData.oauthPassThru`; effects flip both flags per selection |
| `root_basicAuth` | `basicAuth` | `root` | Managed by virtual field (no user-facing label) | Default `false` | SDK bool | `tags: ["managed-by:virtual_authMethod"]` |
| `root_basicAuthUser` | `basicAuthUser` | `root` | `BasicAuth.js:9` (`userLabel = "User"`) | `BasicAuth.js:11` (`userPlaceholder = "User"`); tooltip `BasicAuth.js:10` (`"The username of the data source account"`) | SDK string | `dependsOn: virtual_authMethod == 'BasicAuth'`; `requiredWhen: root_basicAuth == true` |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | `BasicAuth.js:13` (`passwordLabel = "Password"`) | `BasicAuth.js:15` (`passwordPlaceholder = "Password"`); tooltip `BasicAuth.js:14` (`"The password of the data source account"`) | Role `auth.basic.password` | Same conditional/required as `basicAuthUser` |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | Managed by virtual field (no user-facing label) | Default `false` | Role `auth.forwardOAuthToken.enabled` | `tags: ["managed-by:virtual_authMethod"]` |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | `SelfSignedCertificate.js:22` (`label: "Add self-signed certificate"`) | Tooltip `SelfSignedCertificate.js:23`; default `false` | `bool` | `dependsOn: jsonData_storage == 'http'` — TLS only for HTTP storage |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | `SelfSignedCertificate.js:30` (`label: "CA Certificate"`) | `SelfSignedCertificate.js:48` (`placeholder="Begins with --- BEGIN CERTIFICATE ---"`); tooltip default `"Your self-signed certificate"` (`:32`) | Role `tls.caCert` | `dependsOn`/`requiredWhen`: `jsonData_tlsAuthWithCACert == true` |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | `TLSClientAuth.js:27` (`label: "TLS Client Authentication"`) | Tooltip `TLSClientAuth.js:28`; default `false` | `bool` | `dependsOn: jsonData_storage == 'http'` |
| `jsonData_serverName` | `serverName` | `jsonData` | `TLSClientAuth.js:35` (`label: "ServerName"`) | `TLSClientAuth.js:49` (`placeholder: "domain.example.com"`); tooltip default `"A Servername is used to verify the hostname on the returned certificate"` (`:37`) | Role `tls.serverName` | `dependsOn`/`requiredWhen`: `jsonData_tlsAuth == true` |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | `TLSClientAuth.js:59` (`label: "Client Certificate"`) | `TLSClientAuth.js:77` (`placeholder: "Begins with --- BEGIN CERTIFICATE ---"`); tooltip default `"The client certificate can be generated from a Certificate Authority or be self-signed"` (`:61`) | Role `tls.clientCert` | `dependsOn`/`requiredWhen`: `jsonData_tlsAuth == true` |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | `TLSClientAuth.js:86` (`label: "Client Key"`) | `TLSClientAuth.js:104` (`placeholder: "Begins with --- RSA PRIVATE KEY CERTIFICATE ---"`); tooltip default `"The client key can be generated from a Certificate Authority or be self-signed"` (`:88`) | Role `tls.clientKey` | Same conditional/required as `tlsClientCert` |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | `SkipTLSVerification.js:9` (`label: "Skip TLS certificate validation"`) | Tooltip `SkipTLSVerification.js:10`; default `false` | Role `transport.tlsSkipVerify` | `dependsOn: jsonData_storage == 'http'` |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | `AdvancedHttpSettings.js:39` (`label: "Allowed cookies"`) | `AdvancedHttpSettings.js:49` (`placeholder: "New cookie (hit enter to add)"`); tooltip `AdvancedHttpSettings.js:41` | `string[]` | `dependsOn: jsonData_storage == 'http'` |
| `jsonData_timeout` | `timeout` | `jsonData` | `AdvancedHttpSettings.js:58` (`label: "Timeout"`) | `AdvancedHttpSettings.js:70` (`placeholder: "Timeout in seconds"`); tooltip `AdvancedHttpSettings.js:60` (`"HTTP request timeout in seconds"`) | `number` — parsed with `parseInt` at `AdvancedHttpSettings.js:26`; `min={0}` at `:69` | Role `transport.timeoutSeconds`; `dependsOn: jsonData_storage == 'http'` |
| `jsonData_queryParams` | `queryParams` | `jsonData` | `ConfigEditor.tsx:101` (`label="Custom query parameters"`) | `ConfigEditor.tsx:107` (`placeholder="limit=100"`); description `ConfigEditor.tsx:101` (`"Add custom parameters to your queries."`) | `string` — `types.ts:44`, `pkg/settings.go:12` | `dependsOn: jsonData_storage == 'http'`. Admin values OVERRIDE per-query on collision (`pkg/http_storage.go:102-109`) |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_storage` | `storage` | `jsonData` | Storage Location | Yes (direct: `pkg/settings.go:11`, `pkg/datasource.go:158-170`) |
| `root_url` | `url` | `root` | URL (http) / Path (local) | Yes (direct: `pkg/http_storage.go:79-93`, `pkg/local_storage.go:33-49`, `pkg/datasource.go:121-125`) |
| `virtual_authMethod` | `authMethod` | — (virtual) | Authentication method | No — computed from other fields |
| `root_basicAuth` | `basicAuth` | `root` | — (managed) | Yes (SDK via `HTTPClientOptions` at `pkg/datasource.go:35`) |
| `root_basicAuthUser` | `basicAuthUser` | `root` | User | Yes (SDK) |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | Password | Yes (SDK) |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (managed) | Yes (SDK) |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | Add self-signed certificate | Yes (SDK) |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | CA Certificate | Yes (SDK) |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | TLS Client Authentication | Yes (SDK) |
| `jsonData_serverName` | `serverName` | `jsonData` | ServerName | Yes (SDK) |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | Client Certificate | Yes (SDK) |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | Client Key | Yes (SDK) |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | Skip TLS certificate validation | Yes (SDK) |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | Allowed cookies | Yes (SDK) |
| `jsonData_timeout` | `timeout` | `jsonData` | Timeout | Yes (SDK) |
| `jsonData_queryParams` | `queryParams` | `jsonData` | Custom query parameters | Yes (direct: `pkg/http_storage.go:74,103-108`) |

### Frontend-only settings

None — every editor-writable field is either consumed by the plugin's Go
backend directly (`storage`, `queryParams`, `url`) or by the SDK's
`HTTPClientOptions` when constructing the HTTP client (auth, TLS,
cookies, timeout, OAuth forward).

### Backend-only settings

None — every jsonData field the backend reads (`storage`, `queryParams`)
is also exposed in the config editor.

### Excluded settings

- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy` and associated
  socks-proxy fields) — the CSV editor does not even attempt to render
  `SecureSocksProxySettings` (no reference in `src/ConfigEditor.tsx`), so
  the field is absent both by policy (AGENTS.md) and in practice.
- **Custom HTTP headers** (`@grafana/plugin-ui`'s `CustomHeaders`) — even
  though the CSV editor **does not opt in** to `Auth`'s `customHeaders`
  prop (`ConfigEditor.tsx:87-92`), a provisioning payload can still
  populate `jsonData.httpHeaderName<N>` / `secureJsonData.httpHeaderValue<N>`
  pairs and the SDK's `HTTPClientOptions` will forward them. The keys
  are dynamic — not modeled as first-class fields; downstream tools
  should walk `jsonData` for the `httpHeaderName` prefix.
- **`root.access`** — the CSV editor never renders an Access control. New
  datasources default to `access: 'proxy'`. Not modeled as an editor
  field; `RootConfig` in `settings.ts` omits it because it is not part of
  what the editor writes.
- **`root.withCredentials`** — recognised by `convertLegacyAuthProps`'
  `getSelectedMethod` (`utils.js:28-30`) but not included in the default
  `visibleMethods` list `[BasicAuth, OAuthForward, NoAuth]`
  (`AuthMethodSettings.js:47-52`). Kept in `RootConfig` for round-trip
  fidelity — a provisioning payload can still set it — but not
  user-selectable through the editor UI, so it is not a target of the
  virtual auth selector's effects.
- **Per-query fields** (delimiter, header, decimalSeparator, skipRows,
  schema, ignoreUnknown, timezone, method, path, params, headers, body,
  experimental.regex) — defined on `CSVQuery` (`types.ts:9-28`), not on
  the datasource. Not part of `dsconfig.json`.

## Where the types are defined

The configuration types are spread across the plugin and its dependencies
— some fields and base types come from libraries/SDKs rather than the
plugin itself.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `CSVDataSourceOptions`, `CSVQuery`, `Pair`, `FieldSchema`, `defaultOptions`, `defaultQuery` | `src/types.ts:1-51` | plugin ([grafana/grafana-csv-datasource](https://github.com/grafana/grafana-csv-datasource)) |
| `getOptionsWithDefaults`, `getQueryWithDefaults` | `src/utils.ts:1-20` | plugin |
| `DataSourceJsonData` (base interface), `DataSourcePluginOptionsEditorProps`, `GrafanaTheme2` | `packages/grafana-data/src/types/` | `@grafana/data` `13.1.0-25893932881` |
| `DataQuery` | `packages/grafana-schema/src/` | `@grafana/schema` `13.1.0-25893932881` |
| `DataSourceDescription`, `ConnectionSettings`, `Auth`, `convertLegacyAuthProps`, `AuthMethodSettings`, `BasicAuth`, `TLSSettings`, `SelfSignedCertificate`, `TLSClientAuth`, `SkipTLSVerification`, `CustomHeaders` (not opted-in), `AdvancedHttpSettings`, `ConfigSection`, `ConfigSubSection` | `dist/esm/components/ConfigEditor/…` | `@grafana/plugin-ui` `0.13.1` |
| `Field`, `Input`, `RadioButtonGroup`, `Divider`, `useStyles2`, `InlineField`, `TagsInput`, `SecretInput`, `SecretTextArea`, `Select` | `packages/grafana-ui/src/components/` | `@grafana/ui` `13.1.0-25893932881` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `PluginSettings { Storage, QueryParams }`, `LoadPluginSettings` | `pkg/settings.go:10-28` | plugin |
| `Datasource { allowLocalMode, HTTPClient }`, `NewDatasource`, `QueryData`, `CheckHealth`, `newStorage`, `queryModel`, `storage` interface | `pkg/datasource.go:1-171` | plugin |
| `httpStorage`, `newHTTPStorage`, `newRequestFromQuery` | `pkg/http_storage.go:14-131` | plugin |
| `localStorage`, `newLocalStorage`, `Open`, `Stat` | `pkg/local_storage.go:14-51` | plugin |
| `csvOptions`, `fieldSchema`, `parseCSV` | `pkg/csv.go:19-344` | plugin (per-query — not modeled) |
| `backend.DataSourceInstanceSettings` (carries `URL`, `BasicAuthEnabled`, `BasicAuthUser`, `JSONData`, `DecryptedSecureJSONData`), `HTTPClientOptions(ctx)`, `httpclient.New`, `backend.Logger.FromContext` | `backend/common.go`, `backend/httpclient/`, `backend/log/` | `github.com/grafana/grafana-plugin-sdk-go v0.292.0` (plugin) / `v0.292.1` (registry) |

The models in this entry flatten the above into a single Go `Config` type
(root `URL`/`BasicAuth`/`BasicAuthUser`/`WithCredentials` tagged
`json:"-"`, plus the jsonData fields the editor writes and the SDK
reads, plus the two CSV-specific `Storage` / `QueryParams` fields that
mirror `pkg/settings.go:10-13` verbatim, plus `DecryptedSecureJSONData`).
`settings.ts` keeps the three canonical TypeScript types (`RootConfig`,
`JsonDataConfig`, `SecureJsonDataConfig`) plus the `StorageMode` string
union.

## Modeling decisions

- **`jsonData_storage` first, as a first-class field**: it is a real
  storage key with a discrete two-value enum (`"http"` / `"local"`), a
  default (`"http"`), and it drives visibility for every other field via
  `dependsOn: jsonData_storage == 'http'` / `== 'local'`. Not modeled as
  a virtual selector because the value **is** the storage.
- **Root URL is dual-purpose** — the same `settings.URL` root field is
  both the HTTP endpoint (in HTTP mode) and the filesystem base path (in
  local mode). We model it as a single `root_url` field with an
  `overrides` block that swaps the placeholder + description when
  `jsonData_storage == 'local'`. **The dsconfig `overrides` schema does
  not support a `label` swap**, so the field label stays `"URL"` even in
  Local mode — see [Upstream findings](#upstream-findings) #4.
- **Virtual auth selector**: the CSV editor uses
  `@grafana/plugin-ui@0.13.1`'s modern `Auth` widget with default
  `visibleMethods = [BasicAuth, OAuthForward, NoAuth]`, and
  `convertLegacyAuthProps.getOnAuthMethodSelectHandler` writes BOTH
  `root.basicAuth` and `jsonData.oauthPassThru` on every selection
  (mutually exclusive by construction). Modeled as
  `virtual_authMethod` with a `storage.computed.read` expression and
  `effects` mirroring the parca / mock entries.
- **`virtual_authMethod` gated on HTTP storage**: added
  `dependsOn: jsonData_storage == 'http'` so the auth selector disappears
  when Local mode is selected — matching the editor which unmounts the
  `Auth` block for `storage === 'local'` (`src/ConfigEditor.tsx:77-124`).
- **`requiredWhen` on `basicAuthUser` / `basicAuthPassword`**: keyed on
  `root_basicAuth == true` — the editor only renders the Basic Auth
  fields when the Basic auth method is selected, and
  `getOnAuthMethodSelectHandler` sets `basicAuth = (method ===
  BasicAuth)`.
- **TLS pair requirements**: `TLSClientAuth` only reveals ServerName +
  Client Cert + Client Key when `jsonData.tlsAuth` is true;
  `SelfSignedCertificate` only reveals the CA field when
  `jsonData.tlsAuthWithCACert` is true. Encoded as `dependsOn` +
  `requiredWhen` on each field. All three TLS toggles also depend on
  `jsonData_storage == 'http'` so the entire TLS group hides when Local
  mode is selected.
- **`Storage` defaulted in `ApplyDefaults`**: mirrors `src/utils.ts:4-10`
  (frontend `getOptionsWithDefaults`) and `pkg/settings.go:22-24`
  (backend `LoadPluginSettings`). Both replace an empty `storage` with
  `"http"` at read time. `LoadConfig` performs the same normalization so
  callers see editor-parity even for a legacy datasource that never
  wrote the field.
- **`Storage` typed as `StorageMode` (string alias)**: mirrors
  `pkg/settings.go:11` (`Storage string`) with a language-level enum for
  discoverability. The `Validate` method allows `""` explicitly so
  callers may call `Validate` before `ApplyDefaults`; `LoadConfig`
  always applies defaults first.
- **`QueryParams` as a plain string, not a parsed map**: mirrors the
  plugin's own type at `pkg/settings.go:12` (`QueryParams string`) — the
  backend parses it with `url.ParseQuery` at query time
  (`pkg/http_storage.go:103`). Encoding it as a `map`/`array` here would
  diverge from the storage shape.
- **URL required in both storage modes** — Validate rejects an empty
  URL under `http` (matching `pkg/datasource.go:121-125` CheckHealth)
  and under `local` (empty base path would make `os.Open("")` fail on
  every query, `pkg/local_storage.go:35,49`). The
  `dsconfig.json`-level `requiredWhen: "true"` on `root_url` enforces
  the same at the declarative schema layer.
- **Local-mode admin gate NOT enforced by `Validate`**: the
  `GF_PLUGIN_ALLOW_LOCAL_MODE=true` gate lives on the plugin process
  (`pkg/datasource.go:44,158-160`), not in the datasource settings. A
  Config with `Storage == "local"` may still be valid at admission but
  fail at every query time when the plugin process doesn't have the
  env var. The instruction and the `localFile` example call this out.
- **Field ID naming convention**: IDs are prefixed with their storage
  target for easy discoverability — `root_`, `jsonData_`,
  `secureJsonData_`, or `virtual_` — followed by the camelCase storage
  key. The `key` property keeps the plugin's raw storage key.
- **Custom HTTP headers and Secure Socks Proxy excluded**: see
  [Excluded settings](#excluded-settings) above.
- **Flat `Config` in Go**: `settings.go` collapses jsonData fields and
  decrypted secrets onto a single `Config` struct. Root-level fields
  the plugin's own code doesn't read (`BasicAuth`, `BasicAuthUser`,
  `WithCredentials`) are still carried on `Config` because
  `Validate()` needs them to enforce the Basic-auth pair contract, and
  round-trip fidelity requires them.
- **`SecureJsonDataConfig` is a key list**: secure values are
  write-only, so the secure type is just the array of secret key names
  (`basicAuthPassword`, `tlsCACert`, `tlsClientCert`, `tlsClientKey`);
  consumers read `secureJsonFields` to see what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema`
bundle (the k8s-style schema Grafana's datasource API server serves as
`{apiVersion}.json`, `v0alpha1` today) from the embedded `dsconfig.json`:
root fields plus a nested `jsonData` object become the OpenAPI settings
`spec`, secure fields become `secureValues`, and virtual fields are
skipped.

`SettingsExamples()` provides the default configuration plus one
k8s-style example per storage mode / authentication method / TLS
variant. Each example is a full instance-settings object with the
plugin configuration nested under `jsonData` and the relevant
write-only secrets under `secureJsonData` (placeholder values to be
replaced with real secrets; the default example — keyed by the empty
string `""` — carries an empty `basicAuthPassword` to show that no
secret is required for the default No-auth HTTP mode):

| Example | Storage | Auth | TLS | Extras | `secureJsonData` |
| --- | --- | --- | --- | --- | --- |
| `""` (default) | http | None | — | — | `basicAuthPassword` (empty) |
| `httpNoAuth` | http | None | — | `queryParams=limit=100` | `basicAuthPassword` (empty) |
| `httpBasicAuth` | http | Basic | — | — | `basicAuthPassword` |
| `httpOAuthForward` | http | OAuth Identity | — | — | `basicAuthPassword` (empty) |
| `httpTLSMutualAuth` | http | None | mTLS (serverName + client cert/key) | — | `tlsClientCert`, `tlsClientKey` |
| `httpTLSSelfSignedCA` | http | None | Custom CA | — | `tlsCACert` |
| `httpAdvanced` | http | None | — | `timeout=30`, `keepCookies=[session_id]`, `queryParams=format=csv&limit=1000` | `basicAuthPassword` (empty) |
| `localFile` | local | — | — | filesystem base path `/var/lib/csv-data` (requires `GF_PLUGIN_ALLOW_LOCAL_MODE=true` on the plugin process) | `basicAuthPassword` (empty) |
| `legacyEmptyStorage` | (missing → defaulted to http) | None | — | Empty `jsonData` — mirrors a pre-storage-selector provisioned datasource | `basicAuthPassword` (empty) |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings
and returns a fully-defaulted, validated `Config`:

1. **Parse** — copy `settings.URL`, `settings.BasicAuthEnabled`,
   `settings.BasicAuthUser` into `Config`, unmarshal `settings.JSONData`
   into the jsonData portion of the same struct (mirroring the upstream
   `PluginSettings` shape at `pkg/settings.go:10-13` verbatim), and copy
   the four decrypted secrets into `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — write `"http"` when the parsed `Storage` is
   empty, matching `src/utils.ts:9` and `pkg/settings.go:22-24`.
3. **`Validate`** — enforce the runtime contract: URL is required (both
   storage modes fail without it); Storage must be `""`, `"http"`, or
   `"local"` (empty accepted because callers may call `Validate` before
   `ApplyDefaults`); Basic auth requires a username; mTLS requires
   serverName + client cert + client key; custom-CA requires the CA
   PEM; Timeout must be non-negative. Errors are joined so every problem
   surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with
`datasource_uid`, `datasource_name`, and `plugin` labels so log lines
carry request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are still
exported for callers that want to compose them themselves (e.g.
provisioning preview, schema-example round-trip, tests that need to
distinguish parse-level from policy-level errors). Skip `LoadConfig`
in those flows — assemble a `Config` directly.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while
researching upstream. All preserved verbatim in the schema — the schema
records what the plugin **does**, not what it **should** do; these notes
exist so reviewers can reproduce each finding and decide separately
whether to fix upstream.

1. **`root.url` is dual-purpose across storage modes**: the same
   `settings.URL` root storage field serves as an HTTP base URL
   (`pkg/http_storage.go:79`) or a filesystem path
   (`pkg/local_storage.go:33,49`) depending on
   `jsonData.storage`. The editor swaps label ("URL" → "Path") and
   placeholder to hint at this, but the underlying storage is the same.
   Provisioning tools and the k8s API surface see only `url` and have to
   infer intent from `jsonData.storage`.
2. **Admin-gated local mode is a runtime-only check**: an admin can
   provision a datasource with `jsonData.storage = "local"` and a valid
   base path, but every query returns `"local mode has been disabled by
   your administrator"` if the plugin process wasn't started with
   `GF_PLUGIN_ALLOW_LOCAL_MODE=true`. The gate is checked at query time
   in `pkg/datasource.go:158-160`, never at admission or CheckHealth
   (which the current code path executes only after the gate check
   fails). A misconfigured stack will surface this as a query error, not
   a datasource-level error.
3. **`queryParams` overrides per-query params on key collision**:
   `pkg/http_storage.go:102-109` calls `params[k] = v` **after**
   populating `params` from the per-query `query.Params`, so an
   admin-configured `queryParams=format=json` overrides an editor
   `params: [["format", "csv"]]`. This is likely intentional (admin
   trumps user) but not obvious from the "Custom query parameters"
   editor description.
4. **dsconfig `overrides` cannot swap `label`**: `src/ConfigEditor.tsx:64`
   labels the URL field "URL" via `ConnectionSettings` in HTTP mode and
   `<Field label="Path">` in Local mode (`:115`). The dsconfig schema
   does not support a `label` swap under `overrides` (only
   `description`, `placeholder`, `tooltip`, `readOnly`, `validations`,
   `options`, `secureKey`, `defaultValue`, per `dsconfig/schema.json:1063-1092`).
   We use `overrides` to swap the placeholder and description under
   `jsonData_storage == 'local'` and note the label mismatch here.
5. **Local mode does no path escape check when `query.Path == ""`**:
   `pkg/local_storage.go:33-45` only checks the `HasPrefix(fullPath, base)`
   escape when `query.Path` is non-empty. A query with an empty Path
   opens `settings.URL` directly — which is fine because that is the
   admin-configured base path — but the code shape is asymmetric.
6. **HTTP path/host cross-check on every request**:
   `pkg/http_storage.go:85-93` refuses to send a request whose host
   differs from `settings.URL`'s host after concatenating `query.Path`.
   This mitigates a query editor trying to redirect requests to
   arbitrary hosts by embedding `//other.example.com` in `path`. It is
   a defence-in-depth measure and not documented in the editor UI.
7. **`CheckHealth` gate order**: `pkg/datasource.go:127-140` calls
   `d.newStorage(…)` (which for local mode returns the "local mode has
   been disabled" error when the env var is absent) BEFORE the
   `store.Stat` check. So a local-storage datasource with
   `GF_PLUGIN_ALLOW_LOCAL_MODE` unset reports "local mode has been
   disabled by your administrator" at CheckHealth — correctly — but a
   local-storage datasource with the env var set AND a missing file
   reports the `os.Stat` error at CheckHealth (surfaces via
   `store.Stat`).
8. **Frontend defaults are read-only**: `src/utils.ts:4-10`
   `getOptionsWithDefaults` returns a merged object **without** calling
   `onOptionsChange`, so opening the config editor never persists a
   default value into storage. The only way `jsonData.storage` ends up
   `"http"` in storage is if the user explicitly clicked the HTTP
   radio button (or a provisioning payload set it). Legacy datasources
   thus keep `jsonData = {}`, and the backend's own default kicks in.
9. **CustomHeaders is opted out at the editor but writable via
   provisioning**: `src/ConfigEditor.tsx:87-92` calls `Auth` without a
   `customHeaders` prop — so `convertLegacyAuthProps.getCustomHeaders`
   is never wired into the editor UI. But if a provisioning payload
   sets `jsonData.httpHeaderName1 = "X-API-Key"` +
   `secureJsonData.httpHeaderValue1`, the SDK's `HTTPClientOptions`
   still forwards it. The dsconfig schema doesn't model this because
   the keys are dynamically indexed.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go
  validator in this repo) — passes.
- JSON Schema validation against
  [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft 2020-12,
  `additionalProperties: false`) — passes (implicit via `NewSchema()`
  invocation inside conformance tests).
- `go test ./...` on this entry — passes.
- `settings.go`/`schema.go`: `go build`, `go vet`, `gofmt` — clean.
- Full `registry/` module: `go build`, `go vet`, `go test`, `gofmt -l .`
  — clean; no regressions in pre-existing entries.
- `settings.ts`: exports the three canonical types (`RootConfig`,
  `JsonDataConfig`, `SecureJsonDataConfig`) plus the `StorageMode`
  string union — reviewed by hand against the frontend sources; no
  `tsc` runner is wired into the registry module.
