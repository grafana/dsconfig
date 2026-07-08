/**
 * Configuration models for the Zendesk datasource plugin (`grafana-zendesk-datasource`) from the
 * grafana/plugins monorepo. It has no hand-written config editor or backend settings model of its
 * own; both are provided by the shared `@grafana/declarative-plugin` package and specialized by the
 * plugin's `src/spec.ts`.
 *
 * Sources of truth (https://github.com/grafana/plugins @ 4b176ec1f74d80c231be2aeb1ce4713c833a76af):
 * - `plugins/grafana-zendesk-datasource/src/spec.ts` — the plugin spec: one service (`zendesk`),
 *   one server (`zendesk_api`) with a `subdomain` variable, and one auth method (`basic_auth`).
 * - `packages/declarative-plugin/src/components/config-editor/*` — the shared config editor that
 *   renders and stores this shape (`EditorForm.tsx`, `rest/ServiceConfig.tsx`, `rest/Connection.tsx`,
 *   `config-editor/Auth.tsx`, `common/VariablesForm.tsx`).
 * - `sdk/pluginspec/pluginclient/config.go` — the backend `JsonData`/`ServiceConfig` storage model.
 *
 * The shared package's own frontend types (`Config`, `SecureConfig` in `@grafana/declarative-plugin`)
 * are generic (`services: Record<string, ServiceConfig>`, `variables: Record<string, string>`).
 * These entry types narrow that generic shape to the concrete keys this plugin's spec declares.
 */

/**
 * Root (top-level datasource settings) fields.
 *
 * The plugin stores nothing at the root level (`url`, `basicAuth`, etc. are unused — the base URL
 * is derived from the `subdomain` variable), so this is a blank object, never null.
 */
export type RootConfig = Record<string, never>;

/**
 * Per-service auth block stored at `jsonData.services.zendesk.auth`.
 * Only the non-secret parts live here; the API token is a write-only secret (see
 * `SecureJsonDataConfig`).
 */
export type ZendeskAuthConfig = {
  /**
   * Selected authentication method id (the `$defs.authMethods` key). Zendesk exposes a single
   * method, `basic_auth`; the backend defaults `auth.id` to it when unset.
   */
  id?: 'basic_auth';
  /** Basic-auth username — the Zendesk login email (editor label "Email"). */
  username?: string;
};

/** Per-service config stored at `jsonData.services.zendesk`. */
export type ZendeskServiceConfig = {
  auth?: ZendeskAuthConfig;
};

/**
 * Fields stored in `jsonData`. Mirrors the shared package's service-keyed storage shape,
 * narrowed to the keys the Zendesk spec declares.
 */
export type JsonDataConfig = {
  /** Per-service configuration keyed by service id. Zendesk declares one service: `zendesk`. */
  services?: {
    zendesk?: ZendeskServiceConfig;
  };
  /** Connection variables shared across services and referenced by server URLs / auth. */
  variables?: {
    /** Zendesk subdomain; the server URL is `https://{subdomain}.zendesk.com/api/v2/`. */
    subdomain?: string;
  };
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via
 * `secureJsonFields`). Secrets use flat dotted keys of the form
 * `<serviceId>.<secret>`:
 * - `zendesk.password` — the Zendesk API token used for basic auth on the `zendesk` service.
 */
export type SecureJsonDataConfig = Array<'zendesk.password'>;
