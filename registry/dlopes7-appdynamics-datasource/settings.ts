/**
 * Configuration models for the AppDynamics datasource plugin
 * (plugin id: `dlopes7-appdynamics-datasource`).
 *
 * Sources of truth — grafana/plugins-private monorepo @ 267f4937806ed6404b6628d13ae358a5d308e376,
 * plugin path `plugins/dlopes7-appdynamics-datasource/`:
 * - `src/plugin.json:3-4,24` — plugin name (`"AppDynamics"`), id
 *   (`"dlopes7-appdynamics-datasource"`), docs URL
 *   (`info.links[0].url` = `"https://grafana.com/docs/plugins/dlopes7-appdynamics-datasource"`).
 * - `src/components/ConfigEditor.tsx` — the config editor:
 *   - `DataSourceHttpSettings` from `@grafana/ui` (`:96-105`, `defaultUrl` =
 *     `HTTP_URL_PLACEHOLDER`): renders the Controller URL (root `url`), the
 *     Basic auth toggle (root `basicAuth`) + User (root `basicAuthUser`) +
 *     Password (`secureJsonData.basicAuthPassword`), and the Skip TLS Verify
 *     toggle (`jsonData.tlsSkipVerify`).
 *   - "Metrics" `FieldSet` (`:106-185`): Client Name (`jsonData.clientName`),
 *     Client Domain (`jsonData.clientDomain`), Client Secret
 *     (`secureJsonData.clientSecret`), and the excluded Secure Socks Proxy
 *     switch (`jsonData.enableSecureSocksProxy`, `:146-184`).
 *   - "Analytics" `FieldSet` (`:186-256`): Analytics API URL
 *     (`jsonData.analyticsURL`, Select with `ANALYTICS_URLS` `:28-32`), Global
 *     Account Name (`jsonData.globalAccountName`), Analytics API Key
 *     (`secureJsonData.analyticsAPIKey`) + a Help drawer (`:241-255`).
 * - `src/types.ts:77-89` — the frontend types `AppDOptions` (jsonData) and
 *   `AppDSecureJsonData` (secureJsonData).
 * - `src/components/selectors.ts:1` — `HTTP_URL_PLACEHOLDER = 'http://localhost:8086'`.
 * - `pkg/models/settings.go:14-64` — backend `Settings` + `LoadSettings`:
 *   `MetricsURL` from root `config.URL` (`:44`); `ClientSecret` from
 *   `secureJsonData.clientSecret`, else `BasicAuthUsername` from root
 *   `config.BasicAuthUser` + `BasicAuthPassword` from
 *   `secureJsonData.basicAuthPassword` (`:46-51`); `AnalyticsAPIKey` from
 *   `secureJsonData.analyticsAPIKey` (`:53-55`); `TLSSkipVerify`,
 *   `ClientName`, `ClientDomain`, `AnalyticsURL`, `AccountName` from jsonData
 *   (`:15-27`).
 * - `pkg/appd/auth/auth_provider.go:55-89` — `NewMetricsProvider`: Basic auth
 *   when username+password set; API Client (OAuth2 client-credentials) when
 *   clientSecret+url+clientName+clientDomain set.
 * - `pkg/appd/analytics/client.go:39-54` — Analytics uses analyticsURL +
 *   X-Events-API-Key (analyticsAPIKey) + X-Events-API-AccountName
 *   (globalAccountName).
 * - `pkg/appd/health_diagnostics.go:70-135` — `IsAnalyticsConfigured` and
 *   `CheckSettings`: url required; at least one Controller auth method;
 *   per-group completeness for API Client and Basic auth.
 * - `pkg/appd/client/client.go:29-57` — HTTP client honors only
 *   `TLSSkipVerify` and the SDK proxy options; other DataSourceHttpSettings
 *   TLS/header fields are discarded.
 *
 * External components consulted at the versions pinned by the workspace catalog
 * (`.yarnrc.yml:19-26`, referenced via `catalog:` in the plugin's
 * `package.json:79-84`):
 * - `@grafana/ui@^11.6.7`:
 *   - `DataSourceHttpSettings`
 *     (`packages/grafana-ui/src/components/DataSourceSettings/DataSourceHttpSettings.tsx`):
 *     URL field label `'URL'`, placeholder = `defaultUrl`; Basic auth toggle
 *     label `'Basic auth'` writing root `basicAuth`.
 *   - `BasicAuthSettings`
 *     (`.../DataSourceSettings/BasicAuthSettings.tsx`): User label `'User'`,
 *     placeholder `'user'` (root `basicAuthUser`); Password via `SecretFormField`.
 *   - `SecretFormField`
 *     (`packages/grafana-ui/src/components/SecretFormField/SecretFormField.tsx`):
 *     default label/placeholder `'Password'` (`secureJsonData.basicAuthPassword`).
 *   - `HttpProxySettings`
 *     (`.../DataSourceSettings/HttpProxySettings.tsx`): Skip TLS Verify toggle
 *     label `'Skip TLS Verify'` writing `jsonData.tlsSkipVerify`.
 *   - `SecretInput`, `Input`, `Select`, `InlineField`, `Alert`, `Button`, `Icon`.
 * - `@grafana/data@^11.6.7` — `DataSourceJsonData` (base of `AppDOptions`),
 *   `DataSourcePluginOptionsEditorProps`, `DataSourceSettings`.
 * - `@grafana/runtime@^11.6.7` — `config` (feature toggle read for the Secure
 *   Socks Proxy switch).
 *
 * The Secure Socks Proxy switch (`jsonData.enableSecureSocksProxy`,
 * ConfigEditor.tsx:146-184) is deliberately excluded per AGENTS.md.
 */

/**
 * Root (top-level datasource settings) fields the AppDynamics plugin uses.
 *
 * Unlike most datasources, the AppDynamics backend reads root-level fields:
 * `url` (`config.URL` -> `settings.MetricsURL`, `pkg/models/settings.go:44`)
 * and `basicAuthUser` (`config.BasicAuthUser` -> `settings.BasicAuthUsername`,
 * `pkg/models/settings.go:49`). `basicAuth` is the DataSourceHttpSettings
 * enabler toggle (written by the editor, `DataSourceHttpSettings.tsx`); the
 * backend does not read the flag directly but gates the User/Password UI on it.
 */
export type RootConfig = {
  /**
   * AppDynamics Controller base URL. Written by DataSourceHttpSettings (URL
   * field, placeholder `http://localhost:8086`). Read by the backend as
   * `config.URL` -> `settings.MetricsURL` (`pkg/models/settings.go:44`); the
   * backend path-joins Controller paths (e.g. `/controller/rest/applications`,
   * `/controller/api/oauth/access_token`) onto it.
   */
  url?: string;
  /**
   * Basic auth enabler toggle (`'Basic auth'`, DataSourceHttpSettings Auth
   * section). When true the editor reveals the User/Password fields. Not read
   * directly by the backend, but selects the Basic auth credential path when
   * no clientSecret is set.
   */
  basicAuth?: boolean;
  /**
   * Basic auth username. Written by `@grafana/ui` BasicAuthSettings (label
   * `'User'`). Read by the backend as `config.BasicAuthUser` ->
   * `settings.BasicAuthUsername` (`pkg/models/settings.go:49`), but only when
   * no clientSecret is present.
   */
  basicAuthUser?: string;
};

/**
 * Fields stored in `jsonData`. Matches the frontend `AppDOptions`
 * (`src/types.ts:77-83`) plus `tlsSkipVerify` (declared on the backend
 * `Settings`, `pkg/models/settings.go:15`, and written by
 * DataSourceHttpSettings' HttpProxySettings), minus the excluded Secure Socks
 * Proxy switch.
 */
export type JsonDataConfig = {
  /**
   * API Client name. Combined with `clientDomain` as the OAuth2
   * `client_id = clientName@clientDomain` (`pkg/appd/auth/auth_provider.go:78`).
   * Part of the API Client (OAuth) auth method.
   */
  clientName?: string;
  /**
   * API Client account/domain. Combined with `clientName` as the OAuth2
   * `client_id` (`pkg/appd/auth/auth_provider.go:78`). Part of the API Client
   * (OAuth) auth method.
   */
  clientDomain?: string;
  /**
   * Analytics (Events) API URL, independent of the Controller `url`. Editor
   * offers `https://analytics.api.appdynamics.com`,
   * `https://fra-ana-api.saas.appdynamics.com`,
   * `https://syd-ana-api.saas.appdynamics.com` (ConfigEditor.tsx:28-32) plus a
   * custom value. Read by the backend as `settings.AnalyticsURL`
   * (`pkg/models/settings.go:26`).
   */
  analyticsURL?: string;
  /**
   * Analytics global account name, sent as the `X-Events-API-AccountName`
   * header (`pkg/appd/analytics/client.go:54`). Read by the backend as
   * `settings.AccountName` (`pkg/models/settings.go:27`).
   */
  globalAccountName?: string;
  /**
   * Skip TLS certificate validation. Written by DataSourceHttpSettings
   * (HttpProxySettings, label `'Skip TLS Verify'`). Read by the backend as
   * `settings.TLSSkipVerify` and applied as `InsecureSkipVerify`
   * (`pkg/models/settings.go:15`, `pkg/appd/client/client.go:40`).
   */
  tlsSkipVerify?: boolean;

  /*
   * Excluded per AGENTS.md: `enableSecureSocksProxy` (src/types.ts:82,
   * ConfigEditor.tsx:146-184) — consumed transparently by the SDK's
   * `config.HTTPClientOptions(ctx)` proxy options (pkg/models/settings.go:57-61),
   * never read by name. Not modeled as a schema field.
   */
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`):
 * - `basicAuthPassword` — Basic auth password for the Controller API. Read as
 *   `secureJsonData.basicAuthPassword` (`pkg/models/settings.go:50`), used only
 *   when no clientSecret is set.
 * - `clientSecret` — API Client (OAuth2) client secret for the Controller API.
 *   Read as `secureJsonData.clientSecret` (`pkg/models/settings.go:46-47`);
 *   takes precedence over basic auth.
 * - `analyticsAPIKey` — Analytics (Events) API key, sent as the
 *   `X-Events-API-Key` header (`pkg/models/settings.go:53-55`;
 *   `pkg/appd/analytics/client.go:53`).
 */
export type SecureJsonDataConfig = Array<'basicAuthPassword' | 'clientSecret' | 'analyticsAPIKey'>;
