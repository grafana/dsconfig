/**
 * Configuration models for the LogicMonitor Devices datasource plugin
 * (`grafana-logicmonitor-datasource`) from the grafana/plugins monorepo. It has no hand-written
 * config editor or backend settings model of its own; both are provided by the shared
 * `@grafana/declarative-plugin` package and specialized by the plugin's `src/spec.ts`.
 *
 * Sources of truth (https://github.com/grafana/plugins @ 4b176ec1f74d80c231be2aeb1ce4713c833a76af):
 * - `plugins/grafana-logicmonitor-datasource/src/spec.ts` — one service (`logicmonitor`), one server
 *   (`apiServer`, `https://{account_name}.logicmonitor.com/santaba/rest`) with a required
 *   `account_name` variable, and one bearer auth method (`auth_bearer`).
 * - `sdk/pluginspec/pluginclient/config.go` — the backend `JsonData`/`ServiceConfig` storage model.
 */

/** Root (top-level datasource settings) fields. The plugin stores nothing at the root level. */
export type RootConfig = Record<string, never>;

/** Auth block stored at `jsonData.services.logicmonitor.auth`. The bearer token itself is a secret. */
export type LogicMonitorAuthConfig = {
  /** Selected auth method id; `auth_bearer` (bearer). The backend defaults `auth.id` to it. */
  id?: 'auth_bearer';
};

/** Per-service config stored at `jsonData.services.logicmonitor`. */
export type LogicMonitorServiceConfig = {
  auth?: LogicMonitorAuthConfig;
};

/** Fields stored in `jsonData`. */
export type JsonDataConfig = {
  services?: {
    logicmonitor?: LogicMonitorServiceConfig;
  };
  variables?: {
    /** LogicMonitor account subdomain; base URL is `https://{account_name}.logicmonitor.com/santaba/rest`. */
    account_name?: string;
  };
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via
 * `secureJsonFields`):
 * - `logicmonitor.token` — the LogicMonitor REST API v3 bearer token for the `logicmonitor` service.
 */
export type SecureJsonDataConfig = Array<'logicmonitor.token'>;
