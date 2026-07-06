/**
 * Configuration models for the GitHub datasource plugin (`grafana-github-datasource`).
 *
 * Sources of truth (https://github.com/grafana/github-datasource):
 * - `src/types/config.ts` — `GitHubLicenseType`, `GitHubAuthType`, `GitHubDataSourceOptions`,
 *   `GitHubSecureJsonDataKeys`
 * - `src/views/ConfigEditor.tsx` — the configuration editor
 * - `pkg/models/settings.go` — backend `Settings` has a `CachingEnabled` field, but
 *   `pkg/plugin/instance.go` unconditionally sets it to `true` after `LoadSettings` runs,
 *   so the stored value is effectively ignored.
 */

export type GitHubLicenseType = 'github-basic' | 'github-enterprise-cloud' | 'github-enterprise-server';

export type GitHubAuthType = 'personal-access-token' | 'github-app';

/**
 * Root (top-level datasource settings) fields.
 *
 * The GitHub datasource stores no plugin-specific fields at the root level
 * (`url`, `basicAuth`, etc. are unused), so this is a blank object rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Matches the plugin's `GitHubDataSourceOptions`
 * plus `cachingEnabled`, which only exists in the backend model.
 */
export type JsonDataConfig = {
  /**
   * Frontend-only: written and read by the config editor to drive the
   * "GitHub License Type" radio; never read by the plugin backend, which
   * infers Enterprise Server solely from a non-empty `githubUrl`.
   */
  githubPlan?: GitHubLicenseType;
  /** GitHub Enterprise Server base URL; the backend derives `<url>/api/v3` (REST) and `<url>/api/graphql` (GraphQL). */
  githubUrl?: string;
  /** Defaults to 'personal-access-token' in the editor; backend also defaults to it when only accessToken is set. */
  selectedAuthType?: GitHubAuthType;
  /** GitHub App ID (github-app auth). Stored as a string by the editor; backend also accepts a JSON number. */
  appId?: string;
  /** GitHub App installation ID (github-app auth). Stored as a string by the editor; backend also accepts a JSON number. */
  installationId?: string;
  /** Backend-only: not exposed in the config editor; the backend currently forces caching on for every instance. */
  cachingEnabled?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`):
 * - `accessToken` — set if the user is using a Personal Access Token to connect to GitHub
 * - `privateKey` — set if the user is using a GitHub App to connect to GitHub
 */
export type SecureJsonDataConfig = Array<'accessToken' | 'privateKey'>;
