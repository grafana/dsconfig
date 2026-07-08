/**
 * Configuration models for the Netlify datasource plugin (`grafana-netlify-datasource`) from the
 * grafana/plugins monorepo. It has no hand-written config editor or backend settings model of its
 * own; both are provided by the shared `@grafana/declarative-plugin` package and specialized by the
 * plugin's `src/spec.ts`.
 *
 * Sources of truth (https://github.com/grafana/plugins @ 4b176ec1f74d80c231be2aeb1ce4713c833a76af):
 * - `plugins/grafana-netlify-datasource/src/spec.ts` — one service (`Netlify`), one server
 *   (`client_api`, https://api.netlify.com/api/v1, no variables), and one bearer auth method
 *   (`bearer_token`).
 * - `sdk/pluginspec/pluginclient/config.go` — the backend `JsonData`/`ServiceConfig` storage model.
 */

/** Root (top-level datasource settings) fields. The plugin stores nothing at the root level. */
export type RootConfig = Record<string, never>;

/** Auth block stored at `jsonData.services.Netlify.auth`. The bearer token itself is a secret. */
export type NetlifyAuthConfig = {
  /** Selected auth method id; `bearer_token` (bearer). The backend defaults `auth.id` to it. */
  id?: 'bearer_token';
};

/** Per-service config stored at `jsonData.services.Netlify` (the service id is capitalized). */
export type NetlifyServiceConfig = {
  auth?: NetlifyAuthConfig;
};

/** Fields stored in `jsonData`. */
export type JsonDataConfig = {
  services?: {
    Netlify?: NetlifyServiceConfig;
  };
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via
 * `secureJsonFields`):
 * - `Netlify.token` — the Netlify personal access token (bearer) for the `Netlify` service.
 */
export type SecureJsonDataConfig = Array<'Netlify.token'>;
