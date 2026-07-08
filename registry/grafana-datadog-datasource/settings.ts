/**
 * Configuration models for the Grafana Datadog datasource plugin
 * (plugin id: `grafana-datadog-datasource`).
 *
 * Sources of truth — grafana/plugins-private monorepo @ 267f4937806ed6404b6628d13ae358a5d308e376,
 * plugin path `plugins/grafana-datadog-datasource/`:
 * - `src/plugin.json:3-5,31` — plugin name (`"Datadog"`), id
 *   (`"grafana-datadog-datasource"`), docs URL
 *   (`info.links[0].url` = `"https://grafana.com/docs/plugins/grafana-datadog-datasource"`).
 * - `src/components/ConfigEditor.tsx` — the config editor:
 *   - `ConnectionEditor` (`:237-331`): a `Mode` RadioButtonGroup toggling the
 *     effective mode between `'default'` and `'hosted-metrics'`. The displayed
 *     value comes from `getPluginMode(jsonData, basicAuth)` (`:527-535`): use
 *     `jsonData.pluginMode` if set, else `'hosted-metrics'` when root
 *     `basicAuth` is true, else `'default'`. `onModeChange` (`:254-265`) writes
 *     `jsonData.pluginMode`, toggles root `basicAuth`, and swaps `jsonData.url`.
 *     In `'default'` mode the URL field is a region `Select` with
 *     `allowCustomValue` (`:311-327`, options from `src/constants.ts:6-12`); in
 *     `'hosted-metrics'` mode it is an `Input` (`:295-310`). Both write
 *     `jsonData.url`.
 *   - `Auth` from `@grafana/plugin-ui` (`:134-222`): a single visible method
 *     driven by mode — BasicAuth for `'hosted-metrics'`, or a custom
 *     `API_AND_APP_KEY` method for `'default'` (`:138-140`). The custom method
 *     renders two `SecretInput`s writing `secureJsonData.apiKey` and
 *     `secureJsonData.appKey` (`:146-220`). `onAuthMethodSelect` is a no-op
 *     (`:137`) — the method is chosen by the Mode radio, not the Auth widget.
 *   - `AdditionalSettingsEditor` (`:333-502`): a collapsible section with
 *     `logApiRateLimits` (Checkbox `:385-391`), `rateLimitEnabled`
 *     (Checkbox `:396-402`), `rateLimitMetrics` (number Input, min 0 max 100,
 *     shown only when `rateLimitEnabled`, `:406-425`), `disableDataLinks`
 *     (Checkbox `:431-437`), `size` (number Input `:445-461`), and the Secure
 *     Socks Proxy switch (`:463-498`, EXCLUDED — see below).
 * - `src/components/tooltips.tsx:3-137` — `ConfigComponentProps` supplies every
 *   label, placeholder, and tooltip the editor renders.
 * - `src/constants.ts:6-12` — `regions` (US1/Default, US3, US5, EU, US1-FED).
 * - `src/types.ts:193-220` — the frontend types `pluginMode`,
 *   `DataDogJsonData`, and `SecureSettings`.
 * - `pkg/models/settings.go:19-133` — backend `Settings` + internal
 *   `jsonSettings` parse struct, `LoadSettings`, `getPluginMode`, the legacy
 *   `migrateToSecureKey` (api_key/app_key -> secureJsonData), the
 *   `boolMaybeQuoted` lenient bool parser, and `defaultJSONSettings`
 *   (url -> DefaultDatadogAPIURL, size -> 100).
 * - `pkg/models/constants.go:4-7` — `DefaultDatadogAPIURL =
 *   "https://api.datadoghq.com"`, `DefaultDatadogAPIResponseSize = 100`.
 * - `pkg/datadog/health_diagnostics.go:81-105` — `CheckSettings`: default mode
 *   requires apiKey + appKey; hosted-metrics requires url != default + basic
 *   auth username + password; url is required in all modes.
 * - `pkg/datadog/client/client_v1.go:85-99,222-248` — DD-API-KEY /
 *   DD-APPLICATION-KEY headers in every mode; URL userinfo basic auth in
 *   hosted-metrics mode; `/api/v1` path join.
 *
 * External components consulted at the versions pinned by the workspace
 * catalog (`.yarnrc.yml:14-26`, referenced via `catalog:` in the plugin's
 * `package.json:34-51`):
 * - `@grafana/plugin-ui@^0.13.1` —
 *   - `Auth` + `convertLegacyAuthProps`
 *     (`dist/esm/components/ConfigEditor/Auth/utils.js`): `getSelectedMethod`
 *     maps root `basicAuth` -> BasicAuth; `getBasicAuthProps` writes
 *     `config.basicAuthUser` (root) and `secureJsonData.basicAuthPassword`.
 *   - `BasicAuth`
 *     (`dist/esm/components/ConfigEditor/Auth/auth-method/BasicAuth.js`):
 *     default labels/placeholders `"User"` / `"Password"`; the Datadog editor
 *     overrides only the tooltips with the hosted-metrics help text
 *     (ConfigEditor.tsx:141-145).
 *   - `ConfigSection`, `ConfigSubSection`, `DataSourceDescription`.
 * - `@grafana/ui@^11.6.7` — `Input`, `Select`, `RadioButtonGroup`, `Checkbox`,
 *   `SecretInput`, `InlineField`, `Switch`, `Tooltip`, `Icon`, `LinkButton`.
 * - `@grafana/data@^11.6.7` — `DataSourceJsonData` (base of `DataDogJsonData`),
 *   `DataSourcePluginOptionsEditorProps`.
 * - `@grafana/runtime@^11.6.7` — `config` (feature toggle read for the Secure
 *   Socks Proxy switch).
 *
 * The Secure Socks Proxy switch (`jsonData.enableSecureSocksProxy`,
 * ConfigEditor.tsx:463-498) is deliberately excluded per AGENTS.md.
 */

/**
 * Connection/authentication mode. Stored under `jsonData.pluginMode`.
 * `'default'` connects directly to the Datadog API (apiKey + appKey);
 * `'hosted-metrics'` connects through a Grafana Cloud proxy (root basic auth).
 * Mirrors `src/types.ts:193` (`pluginMode`) and
 * `pkg/models/settings.go:12-17` (`PluginMode`).
 */
export type PluginMode = 'default' | 'hosted-metrics';

/**
 * Root (top-level datasource settings) fields the Datadog plugin uses.
 *
 * Unlike most datasources, the Datadog backend reads two root-level fields:
 * `basicAuth` (`config.BasicAuthEnabled`) and `basicAuthUser`
 * (`config.BasicAuthUser`), both at `pkg/models/settings.go:68-72`. They are
 * only meaningful in `'hosted-metrics'` mode. The root `url` is NOT used — the
 * backend reads `jsonData.url` (`pkg/models/settings.go:56`), so it is modeled
 * under `JsonDataConfig`, not here.
 */
export type RootConfig = {
  /**
   * True in `'hosted-metrics'` mode. Written by the Mode radio's `onModeChange`
   * (`src/components/ConfigEditor.tsx:256-264`, `basicAuth: isHostedMetrics`),
   * not by the Auth widget (whose `onAuthMethodSelect` is a no-op). Read by the
   * backend as `config.BasicAuthEnabled` and used by `getPluginMode` as the
   * legacy signal for hosted-metrics (`pkg/models/settings.go:84-92`).
   */
  basicAuth?: boolean;
  /**
   * Grafana Cloud Prometheus username, only used in `'hosted-metrics'` mode.
   * Written by `@grafana/plugin-ui`'s BasicAuth via `convertLegacyAuthProps`
   * (`getBasicAuthProps.onUserChange`). Read by the backend as
   * `config.BasicAuthUser` (`pkg/models/settings.go:69`) and injected as URL
   * userinfo (`pkg/datadog/client/client_v1.go:86`).
   */
  basicAuthUser?: string;
};

/**
 * Fields stored in `jsonData`. Matches the frontend `DataDogJsonData`
 * (`src/types.ts:195-214`) and the backend `jsonSettings` parse struct
 * (`pkg/models/settings.go:94-105`) for the current (non-legacy) keys.
 */
export type JsonDataConfig = {
  /**
   * Connection/auth mode discriminator. Empty on legacy datasources, where the
   * backend infers the mode from root `basicAuth`
   * (`pkg/models/settings.go:84-92`). Editor default is `'default'`.
   */
  pluginMode?: PluginMode;
  /**
   * Datadog API base URL. In `'default'` mode it is a regional API endpoint
   * (default `https://api.datadoghq.com`); in `'hosted-metrics'` mode it is the
   * Grafana Cloud proxy URL. The backend joins `/api/v1` (and `/api/v2`) onto
   * it (`pkg/datadog/client/client_v1.go:88,118`). Backend default is
   * `https://api.datadoghq.com` (`pkg/models/settings.go:108-111`).
   */
  url: string;
  /**
   * "Show API rate limits" checkbox. Surfaces per-endpoint Datadog rate-limit
   * headers in the DataFrame meta. Parsed leniently (accepts `true` or
   * `"true"`) by the backend `boolMaybeQuoted` type
   * (`pkg/models/settings.go:97,120-126`).
   */
  logApiRateLimits: boolean;
  /**
   * "Disable data links" checkbox. When true, panels do not emit Datadog deep
   * links. Parsed leniently (`boolMaybeQuoted`).
   */
  disableDataLinks: boolean;
  /**
   * "Enable API rate limit threshold" checkbox. When true, queries stop once
   * the configured percentage of the API rate limit is reached. Parsed
   * leniently (`boolMaybeQuoted`).
   */
  rateLimitEnabled: boolean;
  /**
   * "API rate limit threshold %" (0-100). Only shown when `rateLimitEnabled` is
   * true (`src/components/ConfigEditor.tsx:406-425`). When enabled but 0, the
   * backend coerces it to 100 (`pkg/models/settings.go:63-65`).
   */
  rateLimitMetrics: number;
  /**
   * "Response Size" — maximum items to retrieve per API request. Editor and
   * backend default 100 (`pkg/models/constants.go:7`,
   * `pkg/models/settings.go:110`).
   */
  size?: number;
  /**
   * Written by the Secure Socks Proxy switch
   * (`src/components/ConfigEditor.tsx:463-498`) and consumed transparently by
   * the SDK's `config.HTTPClientOptions(ctx)` call
   * (`pkg/models/settings.go:74`). The Datadog plugin's own Go code never reads
   * it by name. Deliberately EXCLUDED from the dsconfig registry entry per
   * AGENTS.md.
   */
  enableSecureSocksProxy?: boolean;

  /*
   * Legacy 1.x jsonData keys, kept here only for reference — they are NOT
   * modeled as schema fields. The editor (src/components/ConfigEditor.tsx:41-68)
   * and backend (pkg/models/settings.go:52-58,103-104) migrate api_key/app_key
   * into secureJsonData.apiKey/appKey on load and stop reading the jsonData
   * copies. cacheInterval/cacheSize/naming_strategy are dead keys
   * (src/types.ts:205-213).
   *
   * api_key?: string;
   * app_key?: string;
   * cacheInterval?: string | number;
   * cacheSize?: string | number;
   * naming_strategy?: string;
   */
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`):
 * - `apiKey` — Datadog API key, sent as the `DD-API-KEY` header. Required in
 *   `'default'` mode (`pkg/datadog/health_diagnostics.go:83-85`;
 *   `pkg/datadog/client/client_v1.go:98`).
 * - `appKey` — Datadog Application key, sent as the `DD-APPLICATION-KEY`
 *   header. Required in `'default'` mode
 *   (`pkg/datadog/health_diagnostics.go:86-88`;
 *   `pkg/datadog/client/client_v1.go:99`).
 * - `basicAuthPassword` — Grafana Cloud API key used as the basic-auth
 *   password in `'hosted-metrics'` mode. Read as
 *   `secureSettings["basicAuthPassword"]` (`pkg/models/settings.go:70`) and
 *   injected as URL userinfo (`pkg/datadog/client/client_v1.go:86`).
 */
export type SecureJsonDataConfig = Array<'apiKey' | 'appKey' | 'basicAuthPassword'>;
