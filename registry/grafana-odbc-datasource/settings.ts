/**
 * Configuration models for the Sqlyze (ODBC) datasource plugin (`grafana-odbc-datasource`).
 *
 * Sources of truth (github.com/grafana/plugins-private @ 267f4937, plugin
 * `plugins/grafana-odbc-datasource`):
 * - `src/types.ts:8-22` — `Setting` (name/value/secure), `ODBCSettings` (jsonData),
 *   `SecureODBCSettings` (dynamic secret map)
 * - `src/ConfigEditor.tsx` — the configuration editor; writes jsonData keys `driver`,
 *   `timeout`, `settings` and routes secure setting values to `secureJsonData[name]`
 * - `src/selectors.ts:3-35` — editor labels/placeholders
 * - `pkg/models/settings.go:15-54` — backend `Settings` (`Driver`, `DSN`, `Timeout`,
 *   `Settings`) and `LoadSettings` (resolves secure settings, defaults timeout to "10")
 * - `pkg/database/connect.go:75-86` — `DSN()` builds the connection string consumed by the backend
 *
 * Note on storage keys: the backend `Settings` struct has no json tags and relies on Go's
 * case-insensitive matching, so the editor's lowercase keys (`driver`, `timeout`, `settings`,
 * and per-setting `name`/`value`/`secure`) unmarshal into the capitalized Go fields. The
 * canonical stored keys used here are the lowercase ones the editor actually writes.
 */

/**
 * A single driver setting. Non-secure settings store their value inline; secure settings
 * (`secure: true`) omit `value` and store it in `secureJsonData[name]` instead
 * (`src/ConfigEditor.tsx:89-114`, `src/types.ts:8-12`).
 */
export type Setting = {
  /** Connection-string key, e.g. `host`, `port`, `uid`, `pwd`. */
  name: string;
  /** Inline value for non-secure settings; absent when `secure` is true. */
  value?: string;
  /** When true, the value lives in `secureJsonData[name]` rather than here. */
  secure: boolean;
};

/**
 * Root (top-level datasource settings) fields.
 *
 * The Sqlyze datasource reads nothing at the root level — the backend (`pkg/driver.go:27-36`)
 * only consumes `JSONData` and `DecryptedSecureJSONData`. This is a blank object rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Combines the frontend `ODBCSettings` (`src/types.ts:14-18`)
 * with the backend-only `DSN` field from the Go `Settings` struct (`pkg/models/settings.go:21-26`).
 */
export type JsonDataConfig = {
  /**
   * Driver alias in braces (e.g. `{MySQLDB}`) or an absolute path to the ODBC driver shared
   * library. Required (`src/ConfigEditor.tsx:174-186`, `pkg/models/settings.go:68-71`).
   */
  driver?: string;
  /** Connection timeout in seconds, stored as a string. Defaults to `"10"` (`pkg/models/settings.go:50-52`). */
  timeout?: string;
  /** Dynamic list of driver settings concatenated into the connection string (`src/types.ts:17`). */
  settings?: Setting[];
  /**
   * Backend-only: optional Data Source Name. Read by `DSN()` (`pkg/database/connect.go:76-78`)
   * — when set, the connection string uses `DSN=<value>;` instead of `Driver=<driver>;`.
   * Never written by the configuration editor (absent from `src/types.ts`).
   */
  DSN?: string;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via
 * `secureJsonFields`).
 *
 * Secret keys are DYNAMIC: each equals the `name` of a driver setting whose `secure` flag is
 * enabled (`src/ConfigEditor.tsx:99-114`, `pkg/models/settings.go:33-41`). There is no fixed
 * secret key. `'pwd'` is the conventional password key shown in the plugin README's
 * driver-settings table and is listed here as the representative example.
 */
export type SecureJsonDataConfig = Array<'pwd'>;
