/**
 * Configuration models for the Adobe Analytics datasource plugin
 * (`grafana-adobeanalytics-datasource`) from the grafana/plugins monorepo. It has no hand-written
 * config editor or backend settings model of its own; both are provided by the shared
 * `@grafana/declarative-plugin` package and specialized by the plugin's `src/spec.ts`.
 *
 * Sources of truth (https://github.com/grafana/plugins @ 4b176ec1f74d80c231be2aeb1ce4713c833a76af):
 * - `plugins/grafana-adobeanalytics-datasource/src/spec.ts` — one service (`adobe_analytics`), one
 *   server (`adobeanalytics_api`, `https://analytics.adobe.io/api/{global_company_id}`) with a
 *   `global_company_id` variable, and one OAuth2 client-credentials auth method (`oauth2_m2m`).
 * - `sdk/pluginspec/pluginclient/config.go` — the backend `JsonData`/`ServiceConfig` storage model.
 */

/** Root (top-level datasource settings) fields. The plugin stores nothing at the root level. */
export type RootConfig = Record<string, never>;

/** Auth block stored at `jsonData.services.adobe_analytics.auth`. The client secret is a secret. */
export type AdobeAnalyticsAuthConfig = {
  /** Selected auth method id; `oauth2_m2m`. The backend defaults `auth.id` to it. */
  id?: 'oauth2_m2m';
  /** OAuth2 client id. */
  clientId?: string;
};

/** Per-service config stored at `jsonData.services.adobe_analytics`. */
export type AdobeAnalyticsServiceConfig = {
  auth?: AdobeAnalyticsAuthConfig;
};

/** Fields stored in `jsonData`. */
export type JsonDataConfig = {
  services?: {
    adobe_analytics?: AdobeAnalyticsServiceConfig;
  };
  variables?: {
    /** Adobe Global Company ID; base URL is `https://analytics.adobe.io/api/{global_company_id}`. */
    global_company_id?: string;
  };
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via
 * `secureJsonFields`):
 * - `adobe_analytics.clientSecret` — the OAuth2 client secret for the `adobe_analytics` service.
 */
export type SecureJsonDataConfig = Array<'adobe_analytics.clientSecret'>;
