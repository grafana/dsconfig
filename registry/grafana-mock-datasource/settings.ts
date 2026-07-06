/**
 * Configuration models for the Grafana Mock datasource plugin
 * (plugin id: `grafana-mock-datasource`).
 *
 * Sources of truth (https://github.com/grafana/mock-datasource @ 090f6f2):
 * - `src/plugin.json` — plugin id (`"grafana-mock-datasource"`), name
 *   (`"Mock"`). No `info.links`, no `aliasIDs`; the plugin ships with no
 *   canonical docs URL, so this entry uses the upstream repository URL.
 * - `src/editors/MockConfigEditor.tsx:1-115` — the config editor. Composes
 *   three sections: a bare `ConnectionSettings` (no `urlPlaceholder` override
 *   — the plugin-ui default `'URL'` placeholder applies, `:20`), an
 *   `Auth` wired via `convertLegacyAuthProps` (`:24` — visibleMethods
 *   default `[BasicAuth, OAuthForward, NoAuth]`), and a plugin-owned
 *   `ConfigSection title="Custom HealthCheck"` (`:27-45`) with an
 *   `InlineSwitch` to toggle `customHealthCheckEnabled` (`:33-36`) and a
 *   nested `CustomHealthCheckOptionsEditor` (`:38-43`, `:50-114`) exposing
 *   `customHealthCheck.{status,message,details,skipBackend}`.
 * - `src/selectors.ts:1-26` — verbatim labels and tooltips consumed by
 *   `MockConfigEditor.tsx` (`selectors.ConfigEditor.customHealthCheck.*`).
 * - `src/types/config.types.ts:1-19` — frontend `MockConfig` type
 *   (`customHealthCheckEnabled?: boolean`, `customHealthCheck?: CustomHealthCheck`
 *   extending `DataSourceJsonData`) and `MockSecureConfig` (a
 *   `Partial<Record<never, string>>` derived from an empty
 *   `mockSecureConfigKeys` tuple — no plugin-owned secrets).
 * - `src/datasource.ts:24-60` — frontend `MockDS` overrides `testDatasource`
 *   to short-circuit the backend health call when
 *   `jsonData.customHealthCheck.skipBackend` is true (returns a fake
 *   `{status, message, details}` synthesised from the same jsonData).
 * - `pkg/models/settings.go:1-28` — backend `Config` (`customHealthCheckEnabled`
 *   bool, `customHealthCheck` `CustomHealthCheckConfig{status,message,details,skipBackend}`)
 *   and `LoadSettings` (verbatim `json.Unmarshal(settings.JSONData, &config)`).
 * - `pkg/client/client.go:20-38` — the SDK plumbing. `New` builds the HTTP
 *   client via `setting.HTTPClientOptions(ctx)` (`:21`) — which reads the
 *   root/TLS/OAuth fields — and calls `models.LoadSettings` for the plugin-owned
 *   jsonData.
 * - `pkg/client/handler_checkhealth.go:20-40` — how the CustomHealthCheck
 *   fields are consumed: `settings.CustomHealthCheck.Message` (falls back
 *   to `"health check message not specified"` on blank), `.Status`
 *   (converted via `backend.HealthStatus(...)`), and `.Details` (opaque
 *   bytes). `skipBackend` is parsed but never read by the backend.
 *
 * External components consulted at their pinned versions:
 * - `@grafana/plugin-ui@0.14.0` — `ConnectionSettings`, `Auth` (with
 *   `convertLegacyAuthProps`), `ConfigSection`. Read against
 *   `grafana/plugin-ui` main.
 * - `@grafana/ui@12.4.3` — `Stack`, `InlineFormLabel`, `RadioButtonGroup`,
 *   `Input`, `CodeEditor`, `InlineSwitch`.
 * - `@grafana/data@12.4.3` — `DataSourceJsonData` (base interface),
 *   `DataSourceSettings`, `DataSourcePluginOptionsEditorProps`.
 * - `@grafana/runtime@12.4.3` — `DataSourceWithBackend` (parent class of
 *   `MockDS`).
 *
 * No plugin-defined secrets exist. The secureJsonData keys tracked here
 * are all written by @grafana/plugin-ui's `Auth` component (`basicAuthPassword`,
 * `tlsCACert`, `tlsClientCert`, `tlsClientKey`) plus dynamic
 * `httpHeaderValue<N>` entries when custom headers are configured.
 */

/**
 * Root (top-level datasource settings) fields the Grafana Mock plugin's
 * editor persists.
 *
 * `url` is written by the plugin-ui `ConnectionSettings` component
 * (`src/editors/MockConfigEditor.tsx:20`). `basicAuth`, `basicAuthUser`,
 * and `withCredentials` are written by @grafana/plugin-ui's `Auth`
 * component via `convertLegacyAuthProps` (`Auth/utils.ts:44-54`). None of
 * these are read by the Mock plugin's own Go code — they are consumed by
 * the SDK's `HTTPClientOptions(ctx)` call in `pkg/client/client.go:21`.
 */
export type RootConfig = {
  /** Complete HTTP URL of the mock backend. Not dialed by the Mock plugin itself. */
  url?: string;
  /** True when HTTP Basic authentication is enabled. Written by @grafana/plugin-ui `utils.ts:47`. */
  basicAuth?: boolean;
  /** Basic-auth username. Only meaningful when `basicAuth === true`. */
  basicAuthUser?: string;
  /**
   * Written by @grafana/plugin-ui's `getOnAuthMethodSelectHandler` when
   * the "Cross-site access control" method is selected. The Mock editor
   * does not offer that method in its default `visibleMethods`
   * (`[BasicAuth, OAuthForward, NoAuth]`), so this stays `false` in
   * practice.
   */
  withCredentials?: boolean;
};

/**
 * Fields stored in `jsonData`. Includes the plugin-owned CustomHealthCheck
 * override (`src/types/config.types.ts:3-6` mirrored by `pkg/models/settings.go:11-21`)
 * plus the standard HTTP-settings jsonData fields written by
 * @grafana/plugin-ui's `Auth` (`convertLegacyAuthProps`), which the SDK's
 * `HTTPClientOptions` reads at request time.
 */
export type JsonDataConfig = {
  /**
   * Toggle for the custom health check override. When true, the backend's
   * `CheckHealth` handler (`pkg/client/handler_checkhealth.go:25-35`)
   * returns the values under `customHealthCheck` instead of a generic OK
   * status.
   */
  customHealthCheckEnabled?: boolean;
  /**
   * Nested config object for the custom health check response. Modeled as
   * a `section` in the dsconfig schema so each field is validated
   * individually.
   */
  customHealthCheck?: {
    /** Numeric health status: 0 = UNKNOWN, 1 = OK, 2 = ERROR. Mapped via `backend.HealthStatus(...)` in the backend. */
    status?: number;
    /** Custom message returned by CheckHealth. Blank falls back to `"health check message not specified"` server-side. */
    message?: string;
    /** Opaque details string passed through to `jsonDetails` on the CheckHealth response. Expected to be JSON. */
    details?: string;
    /**
     * When true, the frontend short-circuits `testDatasource` (`src/datasource.ts:32-58`)
     * and synthesises the health response locally. Parsed by the backend
     * (`pkg/models/settings.go:20`) but never consumed there.
     */
    skipBackend?: boolean;
  };
  /** Enable TLS client authentication (mTLS). Requires `serverName`, `tlsClientCert`, `tlsClientKey`. Written by @grafana/plugin-ui `utils.ts:123`. */
  tlsAuth?: boolean;
  /** Enable custom CA verification. Requires `tlsCACert`. Written by @grafana/plugin-ui `utils.ts:94`. */
  tlsAuthWithCACert?: boolean;
  /** Skip TLS certificate validation. Not recommended outside testing. Written by @grafana/plugin-ui `utils.ts:182`. */
  tlsSkipVerify?: boolean;
  /** TLS SNI / cert-verification server name (`TLSClientAuth.tsx:51`). */
  serverName?: string;
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
 * - `tlsClientCert`, `tlsClientKey` — mTLS client credentials when
 *   `tlsAuth` is true.
 *
 * The editor also writes dynamic `httpHeaderValue<N>` secrets when the
 * user configures custom HTTP headers via @grafana/plugin-ui's
 * `CustomHeaders` component. Those keys are indexed pairs — not modeled
 * as first-class fields in this schema; see the README.
 *
 * The Mock plugin itself defines no plugin-specific secrets
 * (`src/types/config.types.ts:8` declares `mockSecureConfigKeys = [] as const`).
 */
export type SecureJsonDataConfig = Array<
  'basicAuthPassword' | 'tlsCACert' | 'tlsClientCert' | 'tlsClientKey'
>;
