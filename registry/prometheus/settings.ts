/**
 * Configuration models for the Prometheus datasource plugin (`prometheus`).
 *
 * Sources of truth (https://github.com/grafana/grafana-prometheus-datasource @ be19ceb):
 * - `src/plugin.json` — plugin id ("prometheus"), name ("Prometheus"), docs URL
 * - `src/configuration/ConfigEditor.tsx` — the outermost editor (composes plugin-ui + PromSettings)
 * - `src/configuration/HttpSettings.tsx` — wires @grafana/plugin-ui's `ConnectionSettings` and `Auth`
 * - `packages/grafana-prometheus/src/configuration/PromSettings.tsx` — Prometheus-specific knobs
 * - `packages/grafana-prometheus/src/configuration/AlertingSettingsOverhaul.tsx` — alerting toggles
 * - `packages/grafana-prometheus/src/configuration/ExemplarSetting.tsx` — exemplars sub-editor
 * - `packages/grafana-prometheus/src/configuration/PromFlavorVersions.ts` — Prometheus type/version options
 * - `packages/grafana-prometheus/src/types.ts:35-55` — `PromOptions extends DataSourceJsonData`
 * - `pkg/promlib/models/settings.go` — backend `PromOptions` (jsonData shape) + `ParsePromOptions`
 * - `pkg/promlib/client/transport.go` — how settings are consumed (SDK HTTPClientOptions)
 *
 * External components consulted at their pinned versions:
 * - `@grafana/plugin-ui@0.13.1` — `ConnectionSettings`, `Auth` (`AuthMethodSettings`, `BasicAuth`,
 *   `TLSSettings`/`SelfSignedCertificate`/`TLSClientAuth`/`SkipTLSVerification`, `CustomHeaders`),
 *   `AdvancedHttpSettings`, `DataSourceDescription`, `ConfigSection`
 * - `@grafana/ui@13.1.0-25893932881` — `Input`, `Switch`, `Select`, `SecretInput`, `SecretTextArea`,
 *   `TagsInput`, `Alert`, `SecureSocksProxySettings` (rendered conditionally, excluded here)
 * - `@grafana/data@13.1.0-25893932881` — `DataSourceJsonData` base interface
 */

/** Prometheus flavor type ("Prometheus" | "Cortex" | "Mimir" | "Thanos"). */
export type PromApplication = 'Prometheus' | 'Cortex' | 'Mimir' | 'Thanos';

/** Query editor mode ("builder" | "code"). */
export type QueryEditorMode = 'builder' | 'code';

/** Browser query cache level ("Low" | "Medium" | "High" | "None"). */
export type PrometheusCacheLevel = 'Low' | 'Medium' | 'High' | 'None';

/** HTTP method used for Prometheus range/instant queries. Defaults to "POST". */
export type PromHTTPMethod = 'POST' | 'GET';

/**
 * A single exemplar trace ID destination entry, as rendered by `ExemplarSetting.tsx`.
 * When `datasourceUid` is set the editor treats the exemplar as an internal link and
 * clears `url`; when `url` is set `datasourceUid` is cleared.
 */
export type ExemplarTraceIdDestination = {
  /** Label name that carries the trace ID. Defaults to `'traceID'` when new entries are added (`ExemplarsSettings.tsx:60`). */
  name: string;
  /** Trace backend URL. Mutually exclusive with `datasourceUid` (`ExemplarSetting.tsx:80-86`). */
  url?: string;
  /** Optional custom label rendered on the exemplar's link button. */
  urlDisplayLabel?: string;
  /** Grafana tracing data source UID. Mutually exclusive with `url` (`ExemplarSetting.tsx:80-86`). */
  datasourceUid?: string;
};

/**
 * Root (top-level datasource settings) fields the Prometheus plugin actually cares about.
 *
 * `url` is read directly by the backend (`pkg/promlib/admission_handler.go:51`,
 * `pkg/promlib/querydata/request.go:61`). `basicAuth`, `basicAuthUser`, and `withCredentials`
 * are populated by @grafana/plugin-ui's `Auth` component and consumed by the SDK's
 * `settings.HTTPClientOptions(ctx)` in `pkg/promlib/client/transport.go:18`. The Prometheus
 * plugin's Go code never touches them via its own typed struct — that's why they are not
 * in the `Config` Go type.
 */
export type RootConfig = {
  /** Complete HTTP URL of the Prometheus server. Backend rejects an empty URL. */
  url?: string;
  /** True when HTTP Basic authentication is enabled. Written by the editor at `HttpSettings.tsx:60-69`. */
  basicAuth?: boolean;
  /** Basic-auth username. Only meaningful when `basicAuth === true`. */
  basicAuthUser?: string;
  /**
   * Legacy: written by @grafana/plugin-ui's `getOnAuthMethodSelectHandler` when the
   * "Cross-site access control" method is selected. The Prometheus editor does not
   * offer that method in its `visibleMethods`, so this stays `false` in practice.
   */
  withCredentials?: boolean;
};

/**
 * Fields stored in `jsonData`. Combines the plugin's `PromOptions` (`types.ts:35-55`),
 * the TLS/HTTP fields written by @grafana/plugin-ui, alerting toggles, and the two
 * backend-only threshold fields the backend parses but the editor never renders.
 */
export type JsonDataConfig = {
  /** Enable TLS client authentication (mTLS). Requires `serverName`, `tlsClientCert`, `tlsClientKey`. */
  tlsAuth?: boolean;
  /** Enable custom CA verification. Requires `tlsCACert`. */
  tlsAuthWithCACert?: boolean;
  /** Skip TLS certificate validation. Not recommended outside testing. */
  tlsSkipVerify?: boolean;
  /** TLS SNI / cert-verification server name (`TLSClientAuth.tsx:51`). */
  serverName?: string;
  /** HTTP request timeout in seconds (`AdvancedHttpSettings.tsx:63`). */
  timeout?: number;
  /** Cookies to forward to the datasource, by name (`AdvancedHttpSettings.tsx:48-58`). */
  keepCookies?: string[];
  /**
   * Written by the "Forward OAuth Identity" auth method selector. Also invoked by
   * `HttpSettings.tsx:66`. When true, the SDK forwards the signed-in user's OAuth
   * identity to the datasource.
   */
  oauthPassThru?: boolean;
  /** Manage alert rules for this data source (`AlertingSettingsOverhaul.tsx:41`). */
  manageAlerts?: boolean;
  /** Allow this datasource as a target for writing recording rules. */
  allowAsRecordingRulesTarget?: boolean;
  /** Scrape interval like `'15s'` (`PromSettings.tsx:135`). */
  timeInterval?: string;
  /** Prometheus query timeout like `'60s'` (`PromSettings.tsx:175`). */
  queryTimeout?: string;
  /** Default editor when opening a query — Builder or Code (`PromSettings.tsx:218`). */
  defaultEditor?: QueryEditorMode;
  /** Disable metrics chooser + metric/label autocomplete (`PromSettings.tsx:250`). */
  disableMetricsLookup?: boolean;
  /** Prometheus flavor (`PromSettings.tsx:299`). */
  prometheusType?: PromApplication;
  /** Free-form version string; the editor's Select options come from `PromFlavorVersions[prometheusType]`. */
  prometheusVersion?: string;
  /** Browser cache level for editor queries (`PromSettings.tsx:378`). */
  cacheLevel?: PrometheusCacheLevel;
  /** Turn on incremental query caching (beta) (`PromSettings.tsx:406`). */
  incrementalQuerying?: boolean;
  /** Duration string, defaults to `'10m'` (`PromSettings.tsx:432`, `QueryCache.ts:32`). Only used when `incrementalQuerying` is true. */
  incrementalQueryOverlapWindow?: string;
  /** Disable recording rules (beta) (`PromSettings.tsx:479`). */
  disableRecordingRules?: boolean;
  /** URL query-parameter string appended to Prometheus requests (`PromSettings.tsx:512`). */
  customQueryParameters?: string;
  /** POST (default) or GET. Backend validates in `pkg/promlib/models/settings.go:92-95`. */
  httpMethod?: PromHTTPMethod;
  /** Series/label endpoint limit; empty = 40000, 0 = no limit (`PromSettings.tsx:583`, `constants.ts:19`). */
  seriesLimit?: number;
  /** Prefer /api/v1/series over /api/v1/label/*/values (`PromSettings.tsx:747`). */
  seriesEndpoint?: boolean;
  /** Exemplar trace-ID destinations (`ExemplarsSettings.tsx`). */
  exemplarTraceIdDestinations?: ExemplarTraceIdDestination[];
  /**
   * Backend-only: parsed by `ParsePromOptions` (`pkg/promlib/models/settings.go:41`), but the
   * Prometheus editor never renders the input (the PromSettings prop
   * `showQuerySamplesProcessedThresholdFields` is not set by this plugin — `ConfigEditor.tsx:47`).
   */
  maxSamplesProcessedWarningThreshold?: number;
  /** Backend-only: see `maxSamplesProcessedWarningThreshold`. */
  maxSamplesProcessedErrorThreshold?: number;
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
