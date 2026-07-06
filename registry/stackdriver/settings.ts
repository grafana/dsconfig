/**
 * Configuration models for the Google Cloud Monitoring datasource plugin (`stackdriver`).
 *
 * Sources of truth (https://github.com/grafana/grafana-cloudmonitoring-datasource @ f3bea86):
 * - `src/plugin.json` — plugin `id`, name, docs link
 * - `src/types/types.ts` — `CloudMonitoringOptions`, `CloudMonitoringSecureJsonData`
 * - `src/components/ConfigEditor/ConfigEditor.tsx` — the configuration editor
 * - `src/utils.ts` — `isCloud()` gates WIF + Forward OAuth Identity visibility
 * - `src/datasource.ts` — how `gceDefaultProject` gets populated at runtime
 * - `pkg/cloudmonitoring/cloudmonitoring.go` — backend `newDatasourceInfo` / `datasourceJSONData`
 * - `pkg/cloudmonitoring/httpclient.go` — how each auth type is consumed (`getMiddleware`,
 *   WIF validation at `:87-89`, `buildURL` combining routes with `universeDomain` at `:79-83`)
 * - External:
 *   - `@grafana/google-sdk` 0.6.0: `AuthConfig` (JWT/GCE/WIF/OAuth pass-through + impersonation UI),
 *     `GoogleAuthType`, `GOOGLE_AUTH_TYPE_OPTIONS`, `OAUTH_PASSTHROUGH_AUTH_TYPE_OPTION`,
 *     `WIF_AUTH_TYPE_OPTION`
 *   - `@grafana/plugin-ui` 0.13.1: `DataSourceDescription`, `ConfigSection`
 *   - `@grafana/ui` 13.1.0: `Field`, `Input`, `SecureSocksProxySettings`, `Divider`, `Alert`, `Stack`
 */

export type CloudMonitoringAuthType =
  | 'jwt'
  | 'gce'
  | 'workloadIdentityFederation'
  | 'forwardOAuthIdentity';

/**
 * Root (top-level datasource settings) fields.
 *
 * The Google Cloud Monitoring datasource stores no plugin-specific fields at the root
 * level (`url`, `basicAuth`, etc. are unused; the backend even ignores `settings.URL`
 * — it derives every Google URL from `routes[]` + `universeDomain` at
 * `pkg/cloudmonitoring/httpclient.go:79-83`), so this is a blank object rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Union of the plugin's `CloudMonitoringOptions`
 * (`src/types/types.ts:38-43`) and the fields from `@grafana/google-sdk`'s
 * `DataSourceOptions` that this plugin actually renders (via `AuthConfig`).
 */
export type JsonDataConfig = {
  /**
   * Discriminator for the auth flow. Written by `AuthConfig` (`AuthConfig.tsx:66-78`) and
   * defaulted to `'jwt'` for new datasources via the `useEffect` at `AuthConfig.tsx:40-48`.
   * The plugin composes `authOptions` at `ConfigEditor.tsx:32-38`: JWT + GCE, and (only when
   * `isCloud()`) WIF + ForwardOAuthIdentity.
   */
  authenticationType?: CloudMonitoringAuthType;

  /**
   * GCP project ID. Written by `JWTForm.tsx:76-83` (JWT), `AuthConfig.tsx:151-158` (GCE),
   * `WIFConfigEditor.tsx:46-53` (WIF), or `OAuthPassthroughConfigEditor.tsx:14-25`
   * (Forward OAuth Identity). Required at runtime for both token-forwarding auth types
   * (`pkg/cloudmonitoring/cloudmonitoring.go:121-125`).
   */
  defaultProject?: string;

  /**
   * JWT only. Written by `JWTForm.tsx:85-92` or populated from an uploaded JWT file's
   * `client_email` (`AuthConfig.tsx:139`).
   */
  clientEmail?: string;

  /**
   * JWT only. Written by `JWTForm.tsx:94-101` or populated from an uploaded JWT file's
   * `token_uri` (`AuthConfig.tsx:141`).
   */
  tokenUri?: string;

  /**
   * JWT only. When set, the backend reads the private key from this file path
   * (`grafana-google-sdk-go/pkg/utils/utils.go:62-89`), otherwise falls back to
   * `secureJsonData.privateKey`. Written by `JWTForm.tsx:103-113`.
   */
  privateKeyPath?: string;

  /**
   * Service account impersonation toggle. Rendered by `AuthConfig` only when the caller
   * passes `showServiceAccountImpersonationConfig` — for Cloud Monitoring, only when auth is
   * `jwt` or `gce` (`ConfigEditor.tsx:41-42`).
   */
  usingImpersonation?: boolean;

  /**
   * Email of the service account to impersonate. Written by `AuthConfig.tsx:196-203`.
   */
  serviceAccountToImpersonate?: string;

  /**
   * WIF only. Written by `WIFConfigEditor.tsx:22-30`. Required for WIF auth
   * (backend validates it at `pkg/cloudmonitoring/httpclient.go:87-89`).
   */
  workloadIdentityPoolProvider?: string;

  /**
   * WIF only. Written by `WIFConfigEditor.tsx:36-44`.
   */
  wifServiceAccountEmail?: string;

  /**
   * Editor-managed side-effect: `AuthConfig.tsx:73-74` sets this to `true` when the auth
   * type is `forwardOAuthIdentity` or `workloadIdentityFederation`; otherwise `false`.
   * Users should not toggle it directly. The backend uses it to gate token forwarding
   * (`pkg/cloudmonitoring/cloudmonitoring.go:260-262` sets `ForwardHTTPHeaders = true` on
   * the HTTP client) and to route CheckHealth error messages
   * (`pkg/cloudmonitoring/cloudmonitoring.go:121-166`).
   */
  oauthPassThru?: boolean;

  /**
   * Optional Google Cloud universe domain (Trusted Partner Cloud, Trusted Cloud by S3NS,
   * mTLS endpoints). The backend joins each Google API host with this suffix at
   * `pkg/cloudmonitoring/httpclient.go:79-83`; empty falls back to `'googleapis.com'`.
   *
   * Written by `ConfigEditor.tsx:90-103`, but the "Additional settings" section is only
   * rendered when the Grafana instance has `config.secureSocksDSProxyEnabled` set
   * (`ConfigEditor.tsx:78`). Provisioning payloads can set this regardless of that flag.
   */
  universeDomain?: string;

  /**
   * Frontend-managed cache of the GCE metadata server's default project ID. Written at
   * runtime by `datasource.ts:186-191` (`ensureGCEDefaultProject`) — the frontend calls
   * the `/gceDefaultProject` resource endpoint and stashes the result here so it isn't
   * refetched on every query.
   *
   * The backend never reads this key — for GCE auth it always calls
   * `utils.GCEDefaultProject` fresh
   * (`pkg/cloudmonitoring/cloudmonitoring.go:666-675`). Do not set this in provisioning
   * payloads; leave it empty and let the frontend fill it in.
   */
  gceDefaultProject?: string;

  /**
   * Toggle for the Secure Socks Proxy transport. Excluded from this registry entry per
   * AGENTS.md (the `SecureSocksProxySettings` component in `ConfigEditor.tsx:104` writes it).
   * Documented here for completeness only.
   *
   * @internal
   */
  enableSecureSocksProxy?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via
 * `secureJsonFields`):
 * - `privateKey` — set when authenticationType is 'jwt' and the user pastes or uploads a
 *   JWT (`AuthConfig.tsx:132-136`, `JWTForm.tsx:117-119` in `@grafana/google-sdk`).
 */
export type SecureJsonDataConfig = Array<'privateKey'>;
