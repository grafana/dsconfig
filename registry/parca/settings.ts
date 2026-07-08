/**
 * Configuration models for the Grafana Parca datasource plugin
 * (plugin id: `parca`).
 *
 * Sources of truth (https://github.com/grafana/grafana-parca-datasource @ 7d9b48a):
 * - `src/plugin.json` — plugin id (`"parca"`), name (`"Parca"`), docs URL
 *   (`info.links[2].url` = `"https://grafana.com/docs/grafana/latest/datasources/parca/"`).
 *   The plugin declares no `aliasIDs`.
 * - `src/ConfigEditor.tsx:1-74` — the config editor. Composes @grafana/plugin-ui
 *   `DataSourceDescription` (docsLink hard-coded at `:34`), `ConnectionSettings`
 *   with `urlPlaceholder="http://localhost:7070"` (`:40`), `Auth` via
 *   `convertLegacyAuthProps` (`:43-48`), and a collapsible
 *   `ConfigSection title="Additional settings"` (`:51-64`, description at
 *   `:53`, `isCollapsible={true}` and `isInitiallyOpen={false}`) containing
 *   `AdvancedHttpSettings` and the conditional `SecureSocksProxySettings`
 *   (excluded, gated on `config.secureSocksDSProxyEnabled` at `:60-62`). A
 *   deprecation banner (`<Alert severity="warning" title="Parca data source
 *   is deprecated">`) is rendered above every other field with a hard-coded
 *   `DEPRECATION_DATE = '2nd of January 2027'` (`:17,27-30`).
 * - `src/types.ts:17-21` — frontend `ParcaDataSourceOptions extends
 *   DataSourceJsonData {}` — a **blank** interface. Parca defines no
 *   plugin-specific jsonData fields.
 * - `pkg/parca/plugin.go:58-79` — backend `NewParcaDatasource`. The only
 *   server-side reads are `settings.HTTPClientOptions(ctx)` (`:65`) for the
 *   HTTP client and `settings.URL` (`:77`) as the base URL of the Connect/
 *   gRPC-web profiling client (`queryv1alpha1connect.NewQueryServiceClient
 *   (httpClient, settings.URL, connect.WithGRPCWeb())`).
 * - `pkg/parca/query.go`, `pkg/parca/resources.go` — additional server-side
 *   code paths. Neither unmarshals `settings.JSONData` at all: every read of
 *   configured state goes through the profiling client built from
 *   `settings.URL` + `HTTPClientOptions`.
 *
 * External components consulted at their pinned versions:
 * - `@grafana/plugin-ui@0.13.1` — `ConnectionSettings` (URL input), `Auth`
 *   with `convertLegacyAuthProps` (AuthMethodSettings, BasicAuth, TLSSettings
 *   via SelfSignedCertificate / TLSClientAuth / SkipTLSVerification,
 *   CustomHeaders), `AdvancedHttpSettings` (Allowed cookies, Timeout),
 *   `DataSourceDescription`, `ConfigSection`.
 * - `@grafana/ui@13.1.0-25893932881` — `Alert`, `Divider`, `Stack`,
 *   `SecureSocksProxySettings` (rendered conditionally, excluded here),
 *   `useStyles2`.
 * - `@grafana/runtime@13.1.0-25893932881` — `config` (reads
 *   `config.secureSocksDSProxyEnabled` at `ConfigEditor.tsx:60` to gate the
 *   excluded `SecureSocksProxySettings`).
 * - `@grafana/data@13.1.0-25893932881` — `DataSourceJsonData` base interface,
 *   `DataSourcePluginOptionsEditorProps`, `GrafanaTheme2`.
 *
 * No plugin-defined secrets exist beyond the standard HTTP-settings ones
 * written by `@grafana/plugin-ui`'s `Auth` component (basicAuthPassword,
 * tlsCACert, tlsClientCert, tlsClientKey, plus dynamic httpHeaderValue<N>).
 */

/**
 * Root (top-level datasource settings) fields the Grafana Parca plugin
 * actually cares about.
 *
 * `url` is read directly by the backend (`pkg/parca/plugin.go:77`) as the
 * base URL of the profiling client. `basicAuth`, `basicAuthUser`, and
 * `withCredentials` are populated by @grafana/plugin-ui's `Auth` component
 * and consumed by the SDK's `settings.HTTPClientOptions(ctx)` call at
 * `pkg/parca/plugin.go:65`. The Parca plugin's Go code never touches them by
 * name — that's why they are not in the `Config` Go type.
 */
export type RootConfig = {
  /** Complete HTTP URL of the Parca server. Backend fails at request time on empty URL. */
  url?: string;
  /** True when HTTP Basic authentication is enabled. Written by @grafana/plugin-ui `utils.ts:47`. */
  basicAuth?: boolean;
  /** Basic-auth username. Only meaningful when `basicAuth === true`. */
  basicAuthUser?: string;
  /**
   * Written by @grafana/plugin-ui's `getOnAuthMethodSelectHandler` when the
   * "Cross-site access control" method is selected. The Parca editor does
   * not offer that method in its `visibleMethods`, so this stays `false` in
   * practice.
   */
  withCredentials?: boolean;
};

/**
 * Fields stored in `jsonData`. Parca defines **no** plugin-specific jsonData
 * fields (`src/types.ts:21` declares `ParcaDataSourceOptions extends
 * DataSourceJsonData {}` — empty). Every field below is written by
 * @grafana/plugin-ui's `Auth` (`convertLegacyAuthProps`) or
 * `AdvancedHttpSettings` and read by the SDK's `HTTPClientOptions`, not by
 * any Parca-owned code path.
 */
export type JsonDataConfig = {
  /** Enable TLS client authentication (mTLS). Requires `serverName`, `tlsClientCert`, `tlsClientKey`. Written by @grafana/plugin-ui `utils.ts:123`. */
  tlsAuth?: boolean;
  /** Enable custom CA verification. Requires `tlsCACert`. Written by @grafana/plugin-ui `utils.ts:94`. */
  tlsAuthWithCACert?: boolean;
  /** Skip TLS certificate validation. Not recommended outside testing. Written by @grafana/plugin-ui `utils.ts:182`. */
  tlsSkipVerify?: boolean;
  /** TLS SNI / cert-verification server name (`TLSClientAuth.tsx:51`). */
  serverName?: string;
  /** HTTP request timeout in seconds (`AdvancedHttpSettings.tsx:64`). */
  timeout?: number;
  /** Cookies to forward to the datasource, by name (`AdvancedHttpSettings.tsx:48-59`). */
  keepCookies?: string[];
  /**
   * Written by the "Forward OAuth Identity" auth method selector
   * (`@grafana/plugin-ui utils.ts:51`). When true, the SDK forwards the
   * signed-in user's OAuth identity to the datasource.
   */
  oauthPassThru?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`):
 * - `basicAuthPassword` — user password when Basic auth is enabled.
 * - `tlsCACert` — custom CA PEM when `tlsAuthWithCACert` is true.
 * - `tlsClientCert`, `tlsClientKey` — mTLS client credentials when `tlsAuth`
 *   is true.
 *
 * The editor also writes dynamic `httpHeaderValue<N>` secrets when the user
 * configures custom HTTP headers via @grafana/plugin-ui's `CustomHeaders`
 * component. Those keys are indexed pairs — not modeled as first-class
 * fields in this schema; see the README.
 *
 * The Parca datasource plugin itself defines no plugin-specific secrets
 * beyond this shared HTTP-settings set.
 */
export type SecureJsonDataConfig = Array<
  'basicAuthPassword' | 'tlsCACert' | 'tlsClientCert' | 'tlsClientKey'
>;
