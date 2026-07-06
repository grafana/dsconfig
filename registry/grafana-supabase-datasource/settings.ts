/**
 * Configuration models for the Supabase datasource plugin (`grafana-supabase-datasource`) from the
 * grafana/plugins monorepo. It has no hand-written config editor or backend settings model of its
 * own; both are provided by the shared `@grafana/declarative-plugin` package and specialized by the
 * plugin's `src/spec.ts`.
 *
 * Sources of truth (https://github.com/grafana/plugins @ 4b176ec1f74d80c231be2aeb1ce4713c833a76af):
 * - `plugins/grafana-supabase-datasource/src/spec.ts` — one service (`mgmt`), one server (`mgmt`,
 *   https://api.supabase.com, no variables), and one bearer auth method (`mgmt_bearer`).
 * - `sdk/pluginspec/pluginclient/config.go` — the backend `JsonData`/`ServiceConfig` storage model.
 */

/** Root (top-level datasource settings) fields. The plugin stores nothing at the root level. */
export type RootConfig = Record<string, never>;

/** Auth block stored at `jsonData.services.mgmt.auth`. The bearer token itself is a secret. */
export type SupabaseAuthConfig = {
  /** Selected auth method id; `mgmt_bearer` (bearer). The backend defaults `auth.id` to it. */
  id?: 'mgmt_bearer';
};

/** Per-service config stored at `jsonData.services.mgmt`. */
export type SupabaseServiceConfig = {
  auth?: SupabaseAuthConfig;
};

/** Fields stored in `jsonData`. */
export type JsonDataConfig = {
  services?: {
    mgmt?: SupabaseServiceConfig;
  };
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via
 * `secureJsonFields`):
 * - `mgmt.token` — the Supabase personal token (bearer) for the `mgmt` service.
 */
export type SecureJsonDataConfig = Array<'mgmt.token'>;
