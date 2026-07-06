/**
 * Configuration models for the Graphite datasource plugin (`graphite`).
 *
 * Sources of truth (https://github.com/grafana/grafana-graphite-datasource @ baa7318):
 * - `src/plugin.json` — plugin id ("graphite"), name ("Graphite"), docs URL
 *   (info.links[3].url = "https://grafana.com/docs/grafana/latest/datasources/graphite/")
 * - `src/configuration/ConfigEditor.tsx` — the outer editor; composes @grafana/ui's
 *   `DataSourceHttpSettings` (defaultUrl "http://localhost:8080"), a FieldSet
 *   "Graphite details" (Version, Graphite backend type, conditional Rollup indicator),
 *   and `MappingsConfiguration` for the Graphite→Loki label mappings.
 * - `src/configuration/MappingsConfiguration.tsx:19-77` — "Label mappings" h3, per-row
 *   Input placeholder "e.g. test.metric.(labelName).*"
 * - `src/configuration/MappingsHelp.tsx` — help drawer markdown (verbatim)
 * - `src/configuration/parseLokiLabelMappings.ts` — fromString/toString converters
 *   between the editor's string form and the persisted matcher objects
 * - `src/types.ts:23-71` — `GraphiteOptions`, `GraphiteType`, `GraphiteQueryImportConfiguration`,
 *   `GraphiteLokiMapping`, `GraphiteMetricLokiMatcher`
 * - `src/versions.ts:3-5` — `GRAPHITE_VERSIONS = ['0.9','1.0','1.1']`,
 *   `DEFAULT_GRAPHITE_VERSION = '1.1'`
 * - `pkg/graphite/graphite.go:37-61` — `NewDatasource` reads `settings.URL` and
 *   `settings.ID` directly and builds an HTTP client via `settings.HTTPClientOptions(ctx)`;
 *   no other jsonData or secureJsonData field is unmarshaled server-side
 * - `pkg/graphite/admission_handler.go:45-53` — rejects an empty URL and any
 *   apiVersion other than "" / "v0alpha1"
 *
 * External components consulted at their pinned versions:
 * - `@grafana/ui@13.1.0` — `DataSourceHttpSettings` (URL input, Access help,
 *   Allowed cookies TagsInput, Timeout input), `BasicAuthSettings` (User + Password),
 *   `HttpProxySettings` (TLS Client Auth, With CA Cert, Skip TLS Verify, Forward
 *   OAuth Identity), `TLSAuthSettings` (ServerName, CA Cert, Client Cert, Client Key
 *   via `CertificationKey`), `CustomHeadersSettings` (dynamic httpHeaderName<N> /
 *   secureJsonData httpHeaderValue<N> — NOT modeled here), `SecureSocksProxySettings`
 *   (rendered when `config.secureSocksDSProxyEnabled` — excluded per AGENTS.md), plus
 *   `Alert`, `Field`, `FieldSet`, `Select`, `Switch`
 * - `@grafana/data@13.1.0` — `DataSourceJsonData` base interface,
 *   `DataSourcePluginOptionsEditorProps`, `updateDatasourcePluginJsonDataOption`,
 *   `onUpdateDatasourceJsonDataOptionSelect`, `onUpdateDatasourceJsonDataOptionChecked`,
 *   the `store` helper
 * - `@grafana/runtime@13.1.0` — `config.secureSocksDSProxyEnabled`
 */

/** Graphite semver-like schema selector — matches `src/versions.ts:3` verbatim. */
export type GraphiteVersion = '0.9' | '1.0' | '1.1';

/**
 * Graphite backend flavour, from `src/types.ts:30-33`. The editor renders a Select
 * whose labels are the enum keys ("Default", "Metrictank") and values are the enum
 * values ("default", "metrictank").
 */
export type GraphiteType = 'default' | 'metrictank';

/**
 * A single matcher inside a Graphite→Loki mapping. If `labelName` is present, the
 * segment is treated as a label extraction target; otherwise the segment is matched
 * literally by `value` (`src/types.ts:68-71`, `parseLokiLabelMappings.ts:9-19`).
 */
export type GraphiteMetricLokiMatcher = {
  value: string;
  labelName?: string;
};

/** A single mapping — an ordered list of matchers (`src/types.ts:64-66`). */
export type GraphiteLokiMapping = {
  matchers: GraphiteMetricLokiMatcher[];
};

/**
 * Cross-datasource migration hints written by the editor's MappingsConfiguration
 * component. Read by @grafana/data's Explore query-conversion logic when switching
 * from Graphite to Loki; neither the Graphite backend nor the Loki backend consume
 * this at query time (`src/types.ts:56-62`).
 */
export type GraphiteQueryImportConfiguration = {
  loki: {
    mappings: GraphiteLokiMapping[];
  };
};

/**
 * Root (top-level datasource settings) fields the Graphite editor writes.
 *
 * All root fields here are populated by @grafana/ui's `DataSourceHttpSettings`
 * component and consumed by the SDK's `settings.HTTPClientOptions(ctx)` call in
 * `pkg/graphite/graphite.go:38`. The Graphite plugin's own Go code touches `URL`
 * (and `ID` for tracing attributes) directly but never inspects `basicAuth` /
 * `basicAuthUser` / `withCredentials` by name — those are honored via the SDK's
 * transport builder.
 */
export type RootConfig = {
  /** Complete HTTP URL of the Graphite render endpoint. Backend admission rejects an empty URL. */
  url?: string;
  /**
   * `'proxy'` (Server, default) or `'direct'` (Browser). Graphite does not pass
   * `showAccessOptions` to `DataSourceHttpSettings`, so the editor never renders
   * an Access control; new datasources always end up with `access: 'proxy'`.
   * Legacy datasources may carry `access: 'direct'` and trigger the deprecation
   * Alert at `ConfigEditor.tsx:54-58`.
   */
  access?: 'proxy' | 'direct';
  /** True when HTTP Basic authentication is enabled (DataSourceHttpSettings basic-auth switch). */
  basicAuth?: boolean;
  /** Basic-auth username. Only meaningful when `basicAuth === true`. */
  basicAuthUser?: string;
  /**
   * Cross-site access-control toggle rendered as "With Credentials" by
   * DataSourceHttpSettings. Independent from `basicAuth` — both can be true.
   */
  withCredentials?: boolean;
};

/**
 * Fields stored in `jsonData`. Combines the plugin's `GraphiteOptions`
 * (`src/types.ts:23-28`) with the TLS/HTTP fields written by @grafana/ui's
 * `DataSourceHttpSettings`, `HttpProxySettings`, and `TLSAuthSettings`.
 *
 * The Graphite plugin's Go backend does not read any of these fields directly
 * (`pkg/graphite/graphite.go:38-61`); the SDK's `HTTPClientOptions` reads the
 * TLS-related fields and cookie/timeout knobs when building the HTTP client.
 * `graphiteVersion`, `graphiteType`, `rollupIndicatorEnabled`, and
 * `importConfiguration` are frontend-only.
 */
export type JsonDataConfig = {
  /** Enable TLS client authentication (mTLS). Requires `serverName` + `tlsClientCert` + `tlsClientKey`. */
  tlsAuth?: boolean;
  /** Enable custom CA verification. Requires `tlsCACert`. */
  tlsAuthWithCACert?: boolean;
  /** Skip TLS certificate validation. Not recommended outside testing. */
  tlsSkipVerify?: boolean;
  /** TLS SNI / cert-verification server name (@grafana/ui TLSAuthSettings.tsx:97-104). */
  serverName?: string;
  /** HTTP request timeout in seconds (@grafana/ui DataSourceHttpSettings.tsx timeout Field). */
  timeout?: number;
  /** Cookies to forward to the datasource, by name (@grafana/ui DataSourceHttpSettings.tsx keepCookies TagsInput). */
  keepCookies?: string[];
  /**
   * Written by the "Forward OAuth Identity" switch inside `HttpProxySettings`.
   * When true, the SDK forwards the signed-in user's OAuth identity to Graphite.
   */
  oauthPassThru?: boolean;
  /**
   * Graphite version — controls which functions are exposed in the query editor
   * (`src/versions.ts:3-5`). The editor's `componentDidMount` writes
   * `DEFAULT_GRAPHITE_VERSION` ('1.1') on load if this field is empty
   * (`ConfigEditor.tsx:43-45`), so any datasource saved through the editor
   * carries a non-empty value.
   */
  graphiteVersion?: GraphiteVersion;
  /**
   * Graphite backend flavour. Empty (undefined) is the initial state — the
   * editor's Select shows no selection until the user picks one.
   */
  graphiteType?: GraphiteType;
  /**
   * Metrictank-only visual affordance — the editor only renders the switch when
   * `graphiteType === 'metrictank'` (`ConfigEditor.tsx:95`). Persisting the flag
   * with a non-metrictank backend is legal but has no rendering effect.
   */
  rollupIndicatorEnabled?: boolean;
  /**
   * Cross-datasource migration hint used by Explore's datasource-switch flow.
   * Neither Graphite nor Loki reads this at query time.
   */
  importConfiguration?: GraphiteQueryImportConfiguration;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config
 * via `secureJsonFields`):
 * - `basicAuthPassword` — user password when Basic auth is enabled.
 * - `tlsCACert` — custom CA PEM when `tlsAuthWithCACert` is true.
 * - `tlsClientCert`, `tlsClientKey` — mTLS client credentials when `tlsAuth` is true.
 *
 * The editor also writes dynamic `httpHeaderValue<N>` secrets when the user
 * configures custom HTTP headers via @grafana/ui's `CustomHeadersSettings`
 * component. Those keys are indexed pairs — not modeled as first-class fields
 * in this schema; see the README.
 */
export type SecureJsonDataConfig = Array<
  'basicAuthPassword' | 'tlsCACert' | 'tlsClientCert' | 'tlsClientKey'
>;
