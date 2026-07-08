/**
 * Configuration models for the Hello datasource plugin (`grafana-hello-datasource`) from the
 * grafana/plugins monorepo. It has no hand-written config editor or backend settings model of its
 * own; both are provided by the shared `@grafana/declarative-plugin` package and specialized by the
 * plugin's `src/spec.ts`. Hello is an experimental plugin used for testing the framework.
 *
 * Sources of truth (https://github.com/grafana/plugins @ 4b176ec1f74d80c231be2aeb1ce4713c833a76af):
 * - `plugins/grafana-hello-datasource/src/spec.ts` — two services (`httpbin`, `postman_echo`), each
 *   with a fixed server URL and the `none` auth method (no authentication).
 * - `sdk/pluginspec/pluginclient/config.go` — the backend `JsonData`/`ServiceConfig` storage model.
 */

/** Root (top-level datasource settings) fields. The plugin stores nothing at the root level. */
export type RootConfig = Record<string, never>;

/** Auth block stored at `jsonData.services.<id>.auth`. Both services use the `none` auth method. */
export type HelloAuthConfig = {
  /** Selected auth method id; `none`. The backend defaults `auth.id` to it. */
  id?: 'none';
};

/** Per-service config. */
export type HelloServiceConfig = {
  auth?: HelloAuthConfig;
};

/** Fields stored in `jsonData`. */
export type JsonDataConfig = {
  services?: {
    httpbin?: HelloServiceConfig;
    postman_echo?: HelloServiceConfig;
  };
};

/**
 * Secret key names stored in `secureJsonData`. Neither service requires authentication, so there
 * are no secrets.
 */
export type SecureJsonDataConfig = never[];
