/**
 * Configuration models for the Grafana Pyroscope datasource plugin
 * (`grafana-pyroscope-datasource`).
 *
 * Sources of truth (https://github.com/grafana/grafana-pyroscope-datasource @ e5d6bfb):
 * - `src/plugin.json` — plugin id ("grafana-pyroscope-datasource"), aliasIDs
 *   (`["phlare"]`), name ("Grafana Pyroscope"), docs URL
 *   (`info.links[2].url` = "https://grafana.com/docs/grafana/latest/datasources/pyroscope/")
 * - `src/ConfigEditor.tsx:20-88` — the outermost editor (composes @grafana/plugin-ui
 *   `DataSourceDescription`, `ConnectionSettings` with `urlPlaceholder="http://localhost:4040"`,
 *   `Auth` via `convertLegacyAuthProps`, a collapsible "Additional settings" `ConfigSection`
 *   containing `AdvancedHttpSettings`, the conditional `SecureSocksProxySettings`, and a
 *   `ConfigSubSection title="Querying"` with a single `Field label="Minimal step"` /
 *   `Input placeholder="15s"` bound to `jsonData.minStep`)
 * - `src/ConfigEditor.tsx:63-65` — `Minimal step` description, error message, and the
 *   invalidity check `/^\d+(ms|[Mwdhmsy])$/.test(minStep)`
 * - `src/types.ts:17-19` — frontend `PyroscopeDataSourceOptions extends DataSourceJsonData`
 *   with a single field `minStep?: string`
 * - `pkg/grafana-pyroscope-datasource/instance.go:44-69` — backend datasource ctor
 *   reads `settings.URL` and delegates to `settings.HTTPClientOptions(ctx)`; the
 *   Connect/gRPC-over-HTTP profiling client (`NewPyroscopeClient`) wraps that HTTP
 *   client at `pyroscopeClient.go:1-40`
 * - `pkg/grafana-pyroscope-datasource/query.go:36-38` — the ad-hoc `dsJsonModel`
 *   used server-side (`type dsJsonModel struct { MinStep string \`json:"minStep"\` }`);
 *   parsed at query time only (`query.go:74-89`, `query.go:173-187`)
 * - `pkg/grafana-pyroscope-datasource/plugin.go:34-36` — `NewDatasource` entry point
 *   used by `datasource.Manage`
 *
 * External components consulted at their pinned versions:
 * - `@grafana/plugin-ui@0.15.0` — `ConnectionSettings` (URL input), `Auth` with
 *   `convertLegacyAuthProps` (AuthMethodSettings, BasicAuth, TLSSettings via
 *   SelfSignedCertificate/TLSClientAuth/SkipTLSVerification, CustomHeaders),
 *   `AdvancedHttpSettings` (Allowed cookies, Timeout), `DataSourceDescription`,
 *   `ConfigSection`/`ConfigSubSection`, `ConfigDescriptionLink`
 * - `@grafana/ui@13.0.2` — `Divider`, `Field`, `Input`, `Stack`, `SecureSocksProxySettings`
 *   (rendered conditionally, excluded here), `useStyles2`
 * - `@grafana/runtime@13.0.2` — `config` (reads `config.secureSocksDSProxyEnabled` to
 *   gate the excluded `SecureSocksProxySettings`)
 * - `@grafana/data@13.0.2` — `DataSourceJsonData` base interface,
 *   `DataSourcePluginOptionsEditorProps`, `GrafanaTheme2`
 *
 * No plugin-defined secrets exist beyond the standard HTTP-settings ones written
 * by `@grafana/plugin-ui`'s `Auth` component (basicAuthPassword, tlsCACert,
 * tlsClientCert, tlsClientKey, plus dynamic httpHeaderValue<N>).
 */

/**
 * Root (top-level datasource settings) fields the Grafana Pyroscope plugin
 * actually cares about.
 *
 * `url` is read directly by the backend (`pkg/grafana-pyroscope-datasource/instance.go:66`)
 * as the base URL of the profiling client. `basicAuth`, `basicAuthUser`, and
 * `withCredentials` are populated by @grafana/plugin-ui's `Auth` component and
 * consumed by the SDK's `settings.HTTPClientOptions(ctx)` call at
 * `instance.go:52`. The Pyroscope plugin's Go code never touches them via its
 * own typed struct — that's why they are not in the `Config` Go type.
 */
export type RootConfig = {
  /** Complete HTTP URL of the Pyroscope server. Backend fails at request time on empty URL. */
  url?: string;
  /** True when HTTP Basic authentication is enabled. Written by @grafana/plugin-ui `utils.ts:47`. */
  basicAuth?: boolean;
  /** Basic-auth username. Only meaningful when `basicAuth === true`. */
  basicAuthUser?: string;
  /**
   * Written by @grafana/plugin-ui's `getOnAuthMethodSelectHandler` when the
   * "Cross-site access control" method is selected. The Pyroscope editor does
   * not offer that method in its `visibleMethods`, so this stays `false` in
   * practice.
   */
  withCredentials?: boolean;
};

/**
 * Fields stored in `jsonData`. Combines the plugin's `PyroscopeDataSourceOptions`
 * (`src/types.ts:17-19`) and the TLS/HTTP fields written by @grafana/plugin-ui's
 * `Auth` + `AdvancedHttpSettings`.
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
  /** HTTP request timeout in seconds (`AdvancedHttpSettings.tsx:63`). */
  timeout?: number;
  /** Cookies to forward to the datasource, by name (`AdvancedHttpSettings.tsx:48-58`). */
  keepCookies?: string[];
  /**
   * Written by the "Forward OAuth Identity" auth method selector
   * (`@grafana/plugin-ui utils.ts:51`). When true, the SDK forwards the
   * signed-in user's OAuth identity to the datasource.
   */
  oauthPassThru?: boolean;
  /**
   * Minimal step used for metric queries against Pyroscope. Duration string
   * matching `/^\d+(ms|[Mwdhmsy])$/` (`ConfigEditor.tsx:65`) — e.g. `15s`,
   * `1m`, `500ms`. Parsed at query time by `backend/gtime.ParseDuration`
   * (`pkg/grafana-pyroscope-datasource/query.go:82-89, 182-186`). Empty or
   * unparseable values fall back to 15 seconds; the effective step per query
   * is `max(query.Interval, minStep)`. Should be the same as or higher than
   * the Pyroscope database's scrape interval.
   */
  minStep?: string;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config
 * via `secureJsonFields`):
 * - `basicAuthPassword` — user password when Basic auth is enabled.
 * - `tlsCACert` — custom CA PEM when `tlsAuthWithCACert` is true.
 * - `tlsClientCert`, `tlsClientKey` — mTLS client credentials when `tlsAuth` is
 *   true.
 *
 * The editor also writes dynamic `httpHeaderValue<N>` secrets when the user
 * configures custom HTTP headers via @grafana/plugin-ui's `CustomHeaders`
 * component. Those keys are indexed pairs — not modeled as first-class fields
 * in this schema; see the README.
 *
 * The Pyroscope datasource plugin itself defines no plugin-specific secrets
 * beyond this shared HTTP-settings set.
 */
export type SecureJsonDataConfig = Array<
  'basicAuthPassword' | 'tlsCACert' | 'tlsClientCert' | 'tlsClientKey'
>;
