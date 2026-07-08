/**
 * Configuration models for the Google Sheets datasource plugin (`grafana-googlesheets-datasource`).
 *
 * Sources of truth (https://github.com/grafana/google-sheets-datasource @ 7619fa04):
 * - `src/types.ts` — `GoogleSheetsDataSourceOptions`, `GoogleSheetsSecureJSONData`, `GoogleSheetsAuth`, `googleSheetsAuthTypes`
 * - `src/components/ConfigEditor.tsx` — the configuration editor and the API-Key field
 * - `src/components/ConfigurationHelp.tsx` — collapsible per-auth-type help drawer
 * - `src/utils.ts` — `getBackwardCompatibleOptions` (legacy `authType` → `authenticationType` migration)
 * - `pkg/models/settings.go` — backend `DatasourceSettings` and `LoadSettings`
 * - External:
 *   - `@grafana/google-sdk` 0.6.1: `AuthConfig` (`src/components/AuthConfig.tsx`) provides the JWT and
 *     GCE auth types and renders the `defaultProject`, `clientEmail`, `tokenUri`, `privateKeyPath`,
 *     and `privateKey` fields via nested `JWTForm`/`JWTConfigEditor`/`WIFConfigEditor` components.
 *     `constants.ts` provides `GOOGLE_AUTH_TYPE_OPTIONS` (JWT + GCE labels).
 */

export type GoogleSheetsAuthType = 'key' | 'jwt' | 'gce';

/**
 * Root (top-level datasource settings) fields.
 *
 * The Google Sheets datasource stores no plugin-specific fields at the root level
 * (`url`, `basicAuth`, etc. are unused), so this is a blank object rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Matches the plugin's `GoogleSheetsDataSourceOptions` plus the
 * fields from `@grafana/google-sdk`'s `DataSourceOptions` that this plugin actually renders.
 */
export type JsonDataConfig = {
  /**
   * Discriminator for the auth flow: 'key' (API Key), 'jwt' (Google JWT File), or 'gce'
   * (GCE Default Service Account). Written by `AuthConfig` at `AuthConfig.tsx:66-78`, defaulted
   * to `'jwt'` for new datasources via the `useEffect` at `AuthConfig.tsx:40-48`.
   */
  authenticationType?: GoogleSheetsAuthType;

  /**
   * Legacy: older provisioned datasources stored the auth type here. `LoadSettings` copies
   * `authType` into `authenticationType` on load (`pkg/models/settings.go:49-51`). Prefer
   * `authenticationType` for new configurations.
   */
  authType?: GoogleSheetsAuthType;

  /**
   * JWT / GCE. Written by `JWTForm.tsx:76-83` (JWT) or by the top-level `Default project` field
   * `AuthConfig.tsx:151-158` (GCE). Populated from an uploaded/pasted JWT file's `project_id`
   * (`AuthConfig.tsx:140`).
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
   * (`grafana-google-sdk-go/pkg/utils/utils.go:62-80`), otherwise falls back to
   * `secureJsonData.privateKey`. Written by `JWTForm.tsx:103-113`.
   */
  privateKeyPath?: string;

  /**
   * Optional spreadsheet ID to use as default when creating new queries. Written by
   * `ConfigEditor.tsx:146` via the `SegmentAsync` `Default Spreadsheet ID` control. Loaded
   * by the backend (`pkg/models/settings.go:22`) but not consumed at query time.
   */
  defaultSheetID?: string;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via
 * `secureJsonFields`):
 * - `apiKey` — set when authenticationType is 'key' (`ConfigEditor.tsx:30-43`).
 * - `privateKey` — set when authenticationType is 'jwt' and the user pastes or uploads a JWT
 *   (`AuthConfig.tsx:132-136`, `JWTForm.tsx:117-119`).
 * - `jwt` — legacy blob of a full JWT service-account JSON; kept for backward compatibility only
 *   (`pkg/models/settings.go:17,45`), no runtime code path depends on it.
 */
export type SecureJsonDataConfig = Array<'apiKey' | 'privateKey' | 'jwt'>;
