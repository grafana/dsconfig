/**
 * Configuration models for the Loki datasource plugin (`loki`).
 *
 * Sources of truth (https://github.com/grafana/grafana-loki-datasource @ 882588b):
 * - `src/plugin.json` — plugin id ("loki"), name ("Loki"), docs URL
 * - `src/configuration/ConfigEditor.tsx` — the outermost editor (composes @grafana/plugin-ui
 *   `DataSourceDescription`, `ConnectionSettings`, `Auth` (via `convertLegacyAuthProps`),
 *   `AdvancedHttpSettings`, plus plugin-local `AlertingSettings`, `QuerySettings`, `DerivedFields`)
 * - `src/configuration/AlertingSettings.tsx` — `manageAlerts` toggle (label
 *   "Manage alert rules in Alerting UI", tooltip at line 26)
 * - `src/configuration/QuerySettings.tsx` — `maxLines` input (label "Maximum lines",
 *   placeholder "1000", tooltip at lines 28-34)
 * - `src/configuration/DerivedFields.tsx` — list wrapper (title "Derived fields",
 *   description at line 51)
 * - `src/configuration/DerivedField.tsx` — per-item fields (name, matcherType,
 *   matcherRegex, url, urlDisplayLabel, datasourceUid, targetBlank)
 * - `src/types.ts:36-64` — frontend `LokiOptions extends DataSourceJsonData` and
 *   `DerivedFieldConfig` types
 * - `src/datasource.ts:158,168,397` — frontend consumption of `maxLines` and `derivedFields`
 * - `pkg/loki/loki.go:48-72` — backend `NewDatasource`; reads `settings.URL` directly and
 *   builds an HTTP client via the SDK's `settings.HTTPClientOptions(ctx)` — no other
 *   jsonData field is unmarshaled server-side
 *
 * External components consulted at their pinned versions:
 * - `@grafana/plugin-ui@0.13.1` — `ConnectionSettings` (URL input), `Auth` with
 *   `convertLegacyAuthProps` (AuthMethodSettings, BasicAuth, TLSSettings via
 *   SelfSignedCertificate/TLSClientAuth/SkipTLSVerification, CustomHeaders),
 *   `AdvancedHttpSettings` (Allowed cookies, Timeout), `DataSourceDescription`,
 *   `ConfigSection`/`ConfigSubSection`, `ConfigDescriptionLink`
 * - `@grafana/ui@^12.4.0` — `Input`, `Switch`, `Select`, `TagsInput`, `Field`,
 *   `InlineField`, `InlineSwitch`, `SecretInput`, `SecretTextArea`, `DataLinkInput`,
 *   `Button`, `SecureSocksProxySettings` (rendered conditionally, excluded here)
 * - `@grafana/runtime@^12.4.0` — `DataSourcePicker` (used inside DerivedField for the
 *   internal link target), `config` (reads `config.defaultDatasourceManageAlertsUiToggle`)
 * - `@grafana/data@^12.4.0` — `DataSourceJsonData` base interface, `VariableOrigin`,
 *   `DataLinkBuiltInVars`
 */

/** Matcher type for a derived field: extract from a log line via regex, or from a label. */
export type DerivedFieldMatcherType = 'regex' | 'label';

/**
 * A single derived-field configuration entry, as rendered by `DerivedField.tsx`.
 * When `datasourceUid` is set the editor treats the derived field as an internal
 * link (the URL is interpolated as a query on that data source); otherwise the
 * URL becomes an external hyperlink template.
 */
export type DerivedFieldConfig = {
  /** Field name shown on the derived data link (`DerivedField.tsx:85-87`). Required and must be unique across the list (`DerivedFields.tsx:39-44`). */
  name: string;
  /**
   * Either a regex applied to the log line (when `matcherType === 'regex'`) or a
   * label name (when `matcherType === 'label'`). Required (`DerivedField.tsx:115-131`).
   */
  matcherRegex: string;
  /** `'regex'` (default, `DerivedField.tsx:61,99`) or `'label'` (`DerivedField.tsx:100`). */
  matcherType?: DerivedFieldMatcherType;
  /** URL template (external) or query text (internal). Supports interpolation like `${__value.raw}` (`DerivedField.tsx:146-158`). */
  url?: string;
  /** Optional custom label rendered on the derived link's button (`DerivedField.tsx:159-169`). */
  urlDisplayLabel?: string;
  /** UID of a Grafana data source; when set makes the derived field an internal link (`DerivedField.tsx:188-203`). Mutually exclusive with `url` semantically — the editor toggles between the two. */
  datasourceUid?: string;
  /** Open the derived link in a new browser tab (`DerivedField.tsx:206-220`). */
  targetBlank?: boolean;
};

/**
 * Root (top-level datasource settings) fields the Loki plugin actually cares about.
 *
 * `url` is read directly by the backend (`pkg/loki/loki.go:66`). `basicAuth`,
 * `basicAuthUser`, and `withCredentials` are populated by @grafana/plugin-ui's
 * `Auth` component and consumed by the SDK's `settings.HTTPClientOptions(ctx)` in
 * `pkg/loki/loki.go:51`. The Loki plugin's Go code never touches them via its own
 * typed struct — that's why they are not in the `Config` Go type.
 */
export type RootConfig = {
  /** Complete HTTP URL of the Loki server. Backend hard-fails at request time on empty URL. */
  url?: string;
  /** True when HTTP Basic authentication is enabled. Written by @grafana/plugin-ui `utils.ts:47`. */
  basicAuth?: boolean;
  /** Basic-auth username. Only meaningful when `basicAuth === true`. */
  basicAuthUser?: string;
  /**
   * Written by @grafana/plugin-ui's `getOnAuthMethodSelectHandler` when the
   * "Cross-site access control" method is selected. The Loki editor does not
   * offer that method in its `visibleMethods`, so this stays `false` in practice.
   */
  withCredentials?: boolean;
};

/**
 * Fields stored in `jsonData`. Combines the plugin's `LokiOptions` (`src/types.ts:36-41`),
 * the TLS/HTTP fields written by @grafana/plugin-ui's `Auth` + `AdvancedHttpSettings`,
 * and the `manageAlerts` toggle from Grafana core's `DataSourceJsonData`.
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
   * Written by the "Forward OAuth Identity" auth method selector (`@grafana/plugin-ui utils.ts:51`).
   * When true, the SDK forwards the signed-in user's OAuth identity to the datasource.
   */
  oauthPassThru?: boolean;
  /**
   * Manage alert rules for this data source from Grafana's Alerting UI
   * (`AlertingSettings.tsx:24-37`). Frontend-only from Loki's perspective — the setting
   * lives on Grafana core's `DataSourceJsonData` and is consumed by Grafana Alerting,
   * not by the Loki plugin's Go code.
   */
  manageAlerts?: boolean;
  /**
   * Default maximum lines returned by log queries against this datasource, as a
   * string (`QuerySettings.tsx:36-42`). Parsed by the frontend at
   * `datasource.ts:168` into an integer with `DEFAULT_MAX_LINES = 1000` as the
   * fallback. The Loki backend never reads this field from settings — the limit
   * is carried on each query via `LokiDataQuery.MaxLines`.
   */
  maxLines?: string;
  /**
   * Derived-field configurations rendered on log rows by the frontend result
   * transformer (`datasource.ts:397`, `transformBackendResult`). Frontend-only.
   */
  derivedFields?: DerivedFieldConfig[];
  /**
   * Declared on the frontend `LokiOptions` type (`src/types.ts:39`) but never
   * written by the editor and never read by the datasource. Treat as dead
   * storage — kept in the type for round-trip compatibility only.
   */
  alertmanager?: string;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via
 * `secureJsonFields`):
 * - `basicAuthPassword` — user password when Basic auth is enabled.
 * - `tlsCACert` — custom CA PEM when `tlsAuthWithCACert` is true.
 * - `tlsClientCert`, `tlsClientKey` — mTLS client credentials when `tlsAuth` is true.
 *
 * The editor also writes dynamic `httpHeaderValue<N>` secrets when the user configures
 * custom HTTP headers via @grafana/plugin-ui's `CustomHeaders` component. Those keys are
 * indexed pairs — not modeled as first-class fields in this schema; see the README.
 */
export type SecureJsonDataConfig = Array<
  'basicAuthPassword' | 'tlsCACert' | 'tlsClientCert' | 'tlsClientKey'
>;
