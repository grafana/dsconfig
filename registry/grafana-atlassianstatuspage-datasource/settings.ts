/**
 * Configuration models for the Atlassian Statuspage datasource plugin
 * (`grafana-atlassianstatuspage-datasource`) from the grafana/plugins monorepo. It has no
 * hand-written config editor or backend settings model of its own; both are provided by the shared
 * `@grafana/declarative-plugin` package and specialized by the plugin's `src/spec.ts`.
 *
 * Sources of truth (https://github.com/grafana/plugins @ 4b176ec1f74d80c231be2aeb1ce4713c833a76af):
 * - `plugins/grafana-atlassianstatuspage-datasource/src/spec.ts` — one service (`atlassianstatuspage`),
 *   one server (`client_api`, `{url}/api/v2`) with a `url` variable, and NO auth methods (public API).
 * - `sdk/pluginspec/pluginclient/config.go` — the backend `JsonData` storage model.
 */

/** Root (top-level datasource settings) fields. The plugin stores nothing at the root level. */
export type RootConfig = Record<string, never>;

/** Fields stored in `jsonData`. This datasource has no per-service auth config. */
export type JsonDataConfig = {
  variables?: {
    /** Atlassian Statuspage URL (including https://, no trailing slash); base URL is `{url}/api/v2`. */
    url?: string;
  };
};

/**
 * Secret key names stored in `secureJsonData`. This datasource queries the public Statuspage API
 * and requires no authentication, so there are no secrets.
 */
export type SecureJsonDataConfig = never[];
