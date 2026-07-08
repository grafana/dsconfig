/**
 * Configuration models for the SolarWinds datasource plugin (`grafana-solarwinds-datasource`) from
 * the grafana/plugins monorepo. It has no hand-written config editor or backend settings model of
 * its own; both are provided by the shared `@grafana/declarative-plugin` package and specialized by
 * the plugin's `src/spec.ts`.
 *
 * Sources of truth (https://github.com/grafana/plugins @ 4b176ec1f74d80c231be2aeb1ce4713c833a76af):
 * - `plugins/grafana-solarwinds-datasource/src/spec.ts` — one service (`solarwinds`), one server
 *   (`api_server`, `{url}:17774/...`) with a required `url` variable, and one basic auth method
 *   (`basic_auth`) with `showTLSOptions: true`.
 * - `packages/declarative-plugin/.../config-editor/Auth.tsx` — basic auth + TLS storage keys.
 * - `sdk/pluginspec/pluginclient/config.go` — the backend `JsonData`/`ServiceConfig` storage model.
 */

/** Root (top-level datasource settings) fields. The plugin stores nothing at the root level. */
export type RootConfig = Record<string, never>;

/** Self-signed CA certificate toggle stored at `...auth.tls.selfSignedCert`. */
export type SelfSignedCertConfig = {
  enabled?: boolean;
};

/** TLS client-authentication settings stored at `...auth.tls.clientAuth`. */
export type ClientAuthConfig = {
  enabled?: boolean;
  serverName?: string;
};

/** TLS settings stored at `jsonData.services.solarwinds.auth.tls`. Certificates/keys are secrets. */
export type TLSConfig = {
  selfSignedCert?: SelfSignedCertConfig;
  clientAuth?: ClientAuthConfig;
  skipVerification?: boolean;
};

/** Auth block stored at `jsonData.services.solarwinds.auth`. Password/certificates are secrets. */
export type SolarWindsAuthConfig = {
  /** Selected auth method id; `basic_auth`. The backend defaults `auth.id` to it. */
  id?: 'basic_auth';
  /** Basic-auth username (editor label "Username"). */
  username?: string;
  tls?: TLSConfig;
};

/** Per-service config stored at `jsonData.services.solarwinds`. */
export type SolarWindsServiceConfig = {
  auth?: SolarWindsAuthConfig;
};

/** Fields stored in `jsonData`. */
export type JsonDataConfig = {
  services?: {
    solarwinds?: SolarWindsServiceConfig;
  };
  variables?: {
    /** SolarWinds instance URL; base URL is `{url}:17774/SolarWinds/InformationService/v3/Json`. */
    url?: string;
  };
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config via
 * `secureJsonFields`):
 * - `solarwinds.password` — basic auth password.
 * - `solarwinds.tls.selfSignedCert` — self-signed CA certificate (when enabled).
 * - `solarwinds.tls.clientCert` — TLS client certificate (when client auth enabled).
 * - `solarwinds.tls.clientKey` — TLS client key (when client auth enabled).
 */
export type SecureJsonDataConfig = Array<
  'solarwinds.password' | 'solarwinds.tls.selfSignedCert' | 'solarwinds.tls.clientCert' | 'solarwinds.tls.clientKey'
>;
