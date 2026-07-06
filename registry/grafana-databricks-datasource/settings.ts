/**
 * Configuration models for the Databricks datasource plugin
 * (`grafana-databricks-datasource`).
 *
 * Sources of truth (github.com/grafana/plugins-private, monorepo path
 * `plugins/grafana-databricks-datasource`, commit
 * 267f4937806ed6404b6628d13ae358a5d308e376):
 * - `src/types.ts` — `Settings` (jsonData), `SecureSettings`,
 *   `AuthenticationType`, `AuthenticationTypeLabel`, `SelectableQueryFormats`
 * - `src/ConfigEditor.tsx` — the configuration editor
 * - `src/selectors.ts` — field labels / placeholders / ids
 * - `src/authUtils.ts` + `src/AzureCredentialsForm.tsx` — the Azure
 *   On-Behalf-Of `azureCredentials` object and `azureClientSecret` secret
 * - `pkg/models/settings.go` — backend `Settings` struct; note it declares a
 *   `cloudFetch` field that `LoadSettings` force-sets to `true` on every load
 *   (line 161-168), so the stored value is effectively ignored.
 */

/**
 * jsonData.authType discriminator. Mirrors the `AuthenticationType` enum in
 * `src/types.ts:51-58`. `''` (Unknown) is the pre-migration value the backend
 * treats as `Pat` (`pkg/models/settings.go:59`).
 */
export type AuthenticationType = '' | 'Pat' | 'OauthM2M' | 'OauthPT' | 'OauthOBO' | 'AzureM2M';

/**
 * jsonData.defaultQueryFormat. The editor stores the numeric `QueryFormat`
 * enum from `@grafana/plugin-ui` (`Timeseries = 0`, `Table = 1`, `Logs = 2`,
 * `Trace = 3`, `OptionMulti = 4`); `SelectableQueryFormats`
 * (`src/types.ts:82-85`) only offers Timeseries (0) and Table (1). The backend
 * reads it as `int` (`pkg/models/settings.go:37`) but never consumes it.
 */
export type QueryFormat = number;

/**
 * Opaque Azure On-Behalf-Of credentials object written by the
 * `@grafana/azure-sdk` `AzureCredentialsForm` (`src/AzureCredentialsForm.tsx`)
 * into `jsonData.azureCredentials` when authType is `OauthOBO`. Shape:
 * `{ authType: 'clientsecret-obo', azureCloud, tenantId, clientId }`; the
 * secret lives write-only in `secureJsonData.azureClientSecret`.
 */
export type AzureCredentials = Record<string, unknown>;

/**
 * Root (top-level datasource settings) fields.
 *
 * The Databricks backend authenticates entirely through `jsonData` +
 * `secureJsonData` and never reads root-level datasource settings (`url`,
 * `user`, `basicAuth`, …), so this is a blank object rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Mirrors `Settings` in `src/types.ts:8-40`
 * (the full frontend picture) plus `cloudFetch`, which only exists in the
 * backend model (`pkg/models/settings.go:41`).
 *
 * `enableSecureSocksProxy` (`src/types.ts:22`) is intentionally omitted — the
 * Secure Socks Proxy field is excluded from registry entries.
 */
export type JsonDataConfig = {
  /** Databricks server hostname; the backend connects on port 443 (`src/selectors.ts:6-12`, `pkg/database/connect.go:59-60`). */
  host: string;
  /** SQL warehouse / cluster HTTP path; the editor strips leading/trailing slashes (`src/ConfigEditor.tsx:80-82`, `src/selectors.ts:13-19`). */
  httpPath: string;
  /** Frontend-only, legacy: declared in `src/types.ts:11`; never written by the current editor nor read by the backend. */
  authMech?: number;
  /** Frontend-only, legacy: declared in `src/types.ts:12`; unused. */
  ssl?: number;
  /** Frontend-only, legacy: declared in `src/types.ts:13`; unused. */
  thriftTransport?: number;
  /** Frontend-only, legacy: declared in `src/types.ts:14`; unused. */
  uid?: string;
  /** Connection timeout in seconds, stored as a string; backend defaults to "60" (`src/selectors.ts:42-48`, `pkg/models/settings.go:151-153`). */
  timeout?: string;
  /** Frontend-only, legacy: declared required in `src/types.ts:16`; never written by the current editor nor read by the backend (appears only in test fixtures). */
  authKind: number;
  /** Connection retry count, stored as a string; backend defaults to "5" (`src/selectors.ts:27-34`, `pkg/models/settings.go:157-159`). */
  retries?: string;
  /** Pause between retries, stored as a string; backend defaults to "0" (`src/selectors.ts:35-41`, `pkg/models/settings.go:154-156`). */
  pause?: string;
  /** Enables verbose driver logging (`src/ConfigEditor.tsx:369-377`, `pkg/main.go:76-79`). */
  debug?: boolean;
  /** Max rows per query, stored as a string; backend defaults to 10000 at parse time (`src/selectors.ts:49-55`, `pkg/database/connect.go:35`). */
  rows?: string;
  /** Retry timeout, stored as a string (`src/selectors.ts:56-62`, `pkg/database/connect.go:37`). */
  retryTimeout?: string;
  /** Authentication method discriminator; defaults to `Pat` via the editor's useEffect (`src/ConfigEditor.tsx:64-74`). */
  authType: AuthenticationType;
  /** Default query result format (`src/ConfigEditor.tsx:394-402`). */
  defaultQueryFormat?: QueryFormat;
  /** Enables Unity Catalog 3-level namespace support (`src/ConfigEditor.tsx:379-392`, `pkg/main.go:55`). */
  enableUnitySupport?: boolean;
  /** Frontend-only query-builder default database/schema (`src/types.ts:28`, `src/datasource.ts`); not part of the backend `Settings` struct. */
  database?: string;
  /** Set to true automatically for `OauthPT`/`OauthOBO`; required true for On-Behalf-Of (`src/ConfigEditor.tsx:89-91`, `src/authUtils.ts:94-115`, `pkg/models/settings.go:141-143`). */
  oauthPassThru?: boolean;
  /** Azure Entra ID M2M directory (tenant) ID (`src/selectors.ts:139-145`, `pkg/models/settings.go:34,101-104`). */
  tenantId?: string;
  /** OAuth M2M / Azure Entra ID M2M application (client) ID (`src/types.ts:35`, `src/selectors.ts:116-131,146-152`, `pkg/models/settings.go:32`). */
  clientId?: string;
  /** Azure On-Behalf-Of credentials object; written by `@grafana/azure-sdk` (`src/authUtils.ts:61-92`, `pkg/models/settings.go:121-139`). */
  azureCredentials?: AzureCredentials;
  /** Azure Entra ID M2M cloud; defaults to `AzureCloud` (`src/ConfigEditor.tsx:258`, `pkg/models/settings.go:116-118`). */
  azureCloud?: string;
  /** Backend-only: declared in `pkg/models/settings.go:41`; `LoadSettings` force-sets it to true unless the `disableCloudFetch` feature toggle is on (lines 161-168). No editor UI. */
  cloudFetch?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`). Mirrors `SecureSettings`
 * (`src/types.ts:45-49`):
 * - `token` — Personal Access Token (`Pat` / legacy Unknown auth)
 * - `clientSecret` — OAuth M2M and Azure Entra ID M2M client secret
 * - `azureClientSecret` — Azure On-Behalf-Of client secret
 */
export type SecureJsonDataConfig = Array<'token' | 'clientSecret' | 'azureClientSecret'>;
