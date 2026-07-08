/**
 * Configuration models for the Google BigQuery datasource plugin (`grafana-bigquery-datasource`).
 *
 * Sources of truth (https://github.com/grafana/google-bigquery-datasource @ 8c658f97):
 * - `src/types.ts` — `BigQueryOptions`, `BigQuerySecureJsonData`, `bigQueryAuthTypes`
 * - `src/constants.ts` — `PROCESSING_LOCATIONS` (34 GCP regions + `''` = automatic)
 * - `src/components/ConfigEditor.tsx` — the configuration editor (Additional Settings section)
 * - `src/components/ConfigurationHelp.tsx` — the top-level "How to configure" collapsible
 * - `pkg/bigquery/settings.go` — backend `loadSettings`
 * - `pkg/bigquery/types/types.go` — backend `BigQuerySettings` (the flat parsed shape)
 * - `pkg/bigquery/http_client.go` — how each auth type is consumed (`getMiddleware` switch,
 *   `newHTTPClient` for oauthPassThru / WIF)
 * - External:
 *   - `@grafana/google-sdk` 0.6.0: `AuthConfig` (JWT/GCE/WIF/impersonation UI),
 *     `GoogleAuthType`, `GOOGLE_AUTH_TYPE_OPTIONS`, `OAUTH_PASSTHROUGH_AUTH_TYPE_OPTION`,
 *     `WIF_AUTH_TYPE_OPTION`
 */

export type BigQueryAuthType =
  | 'jwt'
  | 'gce'
  | 'forwardOAuthIdentity'
  | 'workloadIdentityFederation';

export type BigQueryQueryPriority = 'INTERACTIVE' | 'BATCH';

/**
 * Root (top-level datasource settings) fields.
 *
 * The BigQuery datasource stores no plugin-specific fields at the root level
 * (`url`, `basicAuth`, etc. are unused), so this is a blank object rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Union of the plugin's `BigQueryOptions` and the fields
 * from `@grafana/google-sdk`'s `DataSourceOptions` that this plugin actually renders
 * (via `AuthConfig` + `JWTForm` + `WIFConfigEditor`).
 */
export type JsonDataConfig = {
  /**
   * Discriminator for the auth flow. Written by `AuthConfig` (`AuthConfig.tsx:66-78`) and
   * defaulted to `'jwt'` for new datasources via the `useEffect` at `AuthConfig.tsx:40-48`.
   * The plugin composes `bigQueryAuthTypes` at `src/types.ts:44-48`: JWT, GCE,
   * ForwardOAuthIdentity, and (only when `isCloud()`) WIF.
   */
  authenticationType?: BigQueryAuthType;

  /**
   * GCP project ID. Written by `JWTForm.tsx:76-83` (JWT), `AuthConfig.tsx:151-158` (GCE), or
   * `WIFConfigEditor.tsx:46-53` (WIF).
   */
  defaultProject?: string;

  /**
   * JWT only. Written by `JWTForm.tsx:85-92` or populated from an uploaded JWT file's
   * `client_email` (`AuthConfig.tsx:139` in `@grafana/google-sdk`).
   */
  clientEmail?: string;

  /**
   * JWT only. Written by `JWTForm.tsx:94-101` or populated from an uploaded JWT file's
   * `token_uri` (`AuthConfig.tsx:141` in `@grafana/google-sdk`).
   */
  tokenUri?: string;

  /**
   * JWT only. When set, the backend reads the private key from this file path
   * (`grafana-google-sdk-go/pkg/utils/utils.go:62-80`), otherwise falls back to
   * `secureJsonData.privateKey`. Written by `JWTForm.tsx:103-113`.
   */
  privateKeyPath?: string;

  /**
   * Service account impersonation toggle. Rendered by `AuthConfig` only when the caller
   * passes `showServiceAccountImpersonationConfig` — for BigQuery, only when auth is
   * `jwt` or `gce` (`ConfigEditor.tsx:35-36`).
   */
  usingImpersonation?: boolean;

  /**
   * Email of the service account to impersonate. Written by `AuthConfig.tsx:196-203`.
   */
  serviceAccountToImpersonate?: string;

  /**
   * WIF only. Written by `WIFConfigEditor.tsx:22-30`. Required for WIF auth
   * (backend validates it at `pkg/bigquery/http_client.go:95`).
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
   * (`pkg/bigquery/http_client.go:99-107`).
   */
  oauthPassThru?: boolean;

  /**
   * BigQuery processing location (empty string = automatic). One of 34 GCP regions plus
   * two multi-regionals (US, EU). Written by `ConfigEditor.tsx:62-85`; option list at
   * `src/constants.ts:11-65`.
   */
  processingLocation?: string;

  /**
   * Override the default BigQuery API endpoint (`https://bigquery.googleapis.com/bigquery/v2/`).
   * Written by `ConfigEditor.tsx:86-109`.
   */
  serviceEndpoint?: string;

  /**
   * Cap on bytes billed per query (integer). Written by `ConfigEditor.tsx:110-133` — the
   * `onMaxBytesBilledChange` handler casts the raw string to `Number` before writing.
   * Case-preserved key (`MaxBytesBilled`) matches the backend json tag at
   * `pkg/bigquery/types/types.go:17`.
   */
  MaxBytesBilled?: number;

  /**
   * Backend-only, unused at runtime. Defined at `pkg/bigquery/types/types.go:13` and
   * `src/types.ts:35` but never written by any editor and never read by any code path.
   */
  flatRateProject?: string;

  /**
   * Backend-only, unused at runtime. A `queryPriority` also exists on individual queries
   * (`src/types.ts:107`), but this datasource-level one is not consumed anywhere.
   */
  queryPriority?: BigQueryQueryPriority;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via
 * `secureJsonFields`):
 * - `privateKey` — set when authenticationType is 'jwt' and the user pastes or uploads a
 *   JWT (`AuthConfig.tsx:132-136`, `JWTForm.tsx:117-119` in `@grafana/google-sdk`).
 */
export type SecureJsonDataConfig = Array<'privateKey'>;
