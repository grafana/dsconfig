/**
 * Configuration models for the Vercel datasource plugin (`grafana-vercel-datasource`) from the
 * grafana/plugins monorepo. It has no hand-written config editor or backend settings model of its
 * own; both are provided by the shared `@grafana/declarative-plugin` package and specialized by the
 * plugin's `src/spec.ts`.
 *
 * Sources of truth (https://github.com/grafana/plugins @ 4b176ec1f74d80c231be2aeb1ce4713c833a76af):
 * - `plugins/grafana-vercel-datasource/src/spec.ts` — one service (`vercel`), one server
 *   (`vercelServer`, https://api.vercel.com) with an optional `team_id` variable, and one bearer
 *   auth method (`vercelApiKey`).
 * - `packages/declarative-plugin/src/components/config-editor/*` — the shared config editor.
 * - `sdk/pluginspec/pluginclient/config.go` — the backend `JsonData`/`ServiceConfig` storage model.
 */

/** Root (top-level datasource settings) fields. The plugin stores nothing at the root level. */
export type RootConfig = Record<string, never>;

/** Auth block stored at `jsonData.services.vercel.auth`. The bearer token itself is a secret. */
export type VercelAuthConfig = {
  /** Selected auth method id; `vercelApiKey` (bearer). The backend defaults `auth.id` to it. */
  id?: 'vercelApiKey';
};

/** Per-service config stored at `jsonData.services.vercel`. */
export type VercelServiceConfig = {
  auth?: VercelAuthConfig;
};

/** Fields stored in `jsonData`. */
export type JsonDataConfig = {
  services?: {
    vercel?: VercelServiceConfig;
  };
  variables?: {
    /** Optional Vercel team ID; only needed for team-scoped tokens (not part of the base URL). */
    team_id?: string;
  };
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via
 * `secureJsonFields`):
 * - `vercel.token` — the Vercel Access Token (bearer) for the `vercel` service.
 */
export type SecureJsonDataConfig = Array<'vercel.token'>;
