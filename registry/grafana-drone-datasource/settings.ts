/**
 * Configuration models for the Drone datasource plugin (`grafana-drone-datasource`) from the
 * grafana/plugins monorepo. It has no hand-written config editor or backend settings model of its
 * own; both are provided by the shared `@grafana/declarative-plugin` package and specialized by the
 * plugin's `src/spec.ts`.
 *
 * Sources of truth (https://github.com/grafana/plugins @ 4b176ec1f74d80c231be2aeb1ce4713c833a76af):
 * - `plugins/grafana-drone-datasource/src/spec.ts` — one service (`drone`), one server (`apiServer`,
 *   `{url}/api`) with a `url` variable, and one bearer auth method (`auth_bearer`).
 * - `sdk/pluginspec/pluginclient/config.go` — the backend `JsonData`/`ServiceConfig` storage model.
 */

/** Root (top-level datasource settings) fields. The plugin stores nothing at the root level. */
export type RootConfig = Record<string, never>;

/** Auth block stored at `jsonData.services.drone.auth`. The bearer token itself is a secret. */
export type DroneAuthConfig = {
  /** Selected auth method id; `auth_bearer` (bearer). The backend defaults `auth.id` to it. */
  id?: 'auth_bearer';
};

/** Per-service config stored at `jsonData.services.drone`. */
export type DroneServiceConfig = {
  auth?: DroneAuthConfig;
};

/** Fields stored in `jsonData`. */
export type JsonDataConfig = {
  services?: {
    drone?: DroneServiceConfig;
  };
  variables?: {
    /** Drone server URL (including https://, no trailing slash); the base URL is `{url}/api`. */
    url?: string;
  };
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via
 * `secureJsonFields`):
 * - `drone.token` — the Drone API token (bearer) for the `drone` service.
 */
export type SecureJsonDataConfig = Array<'drone.token'>;
