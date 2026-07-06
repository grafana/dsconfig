/**
 * Configuration models for the Jira datasource plugin
 * (plugin id: `grafana-jira-datasource`).
 *
 * Sources of truth (github.com/grafana/plugins-private @ 267f4937, plugin path
 * `plugins/grafana-jira-datasource`):
 * - `src/plugin.json:5,4,28` — plugin id (`"grafana-jira-datasource"`), name
 *   (`"Jira"`), docs link
 *   (`"https://grafana.com/docs/plugins/grafana-jira-datasource"`).
 * - `src/types.ts:3-8` — `JiraAuthMethod` and the `authMethodOptions` radio
 *   (Basic Auth=`basicAuth`, OAuth 2.0 — Service Account=`oauth2`).
 * - `src/types.ts:31-48` — `JiraOptions` (jsonData) and `JiraSecureOptions`
 *   (secureJsonData `token`, `oauthClientSecret`).
 * - `src/types.ts:57-60` — `Provider` enum (Jira Cloud=`cloud`, Jira Data
 *   Center / Jira Server=`server`), the values written to `jsonData.hosting`.
 * - `src/components/ConfigEditor.tsx:34-465` — the configuration editor:
 *   - `<ConfigSection title="Connection">` (`:175`) with the Provider radio ->
 *     jsonData.hosting (`:176-202`) and the URL input -> jsonData.url
 *     (`:203-225`).
 *   - `<Auth ... customMethods={[{ id: 'custom-jira', label: 'Jira
 *     authentication' ...}]}>` (`:235-412`): the Authentication method radio ->
 *     jsonData.authMethod (`:248-262`); the basicAuth block (`:264-349`): User
 *     email -> jsonData.user (`:266-282`), API Token -> secureJsonData.token
 *     (`:283-317`), Scoped Token switch -> jsonData.scopedToken (`:318-328`),
 *     Jira App Cloud Id -> jsonData.cloudId (`:329-347`); the oauth2 block
 *     (`:351-408`): Client ID -> jsonData.oauthClientID (`:353-369`), Client
 *     Secret -> secureJsonData.oauthClientSecret (`:370-389`), Jira App Cloud Id
 *     -> jsonData.cloudId (`:390-406`).
 *   - `<ConfigSection title="Additional settings">` (`:419-459`): the Secure
 *     Socks Proxy checkbox -> jsonData.enableSecureSocksProxy (excluded here).
 * - `src/components/selectors.ts:3-33` — the E2E aria-label / input-id selector
 *   map for the config editor inputs (labels/placeholders themselves are
 *   hard-coded in the editor).
 * - `pkg/models/settings.go:14-86` — backend `Settings` struct and
 *   `LoadSettings` (json parsing, `GetAuthMethod` resolution, URL scheme
 *   normalization, per-auth-method secret copy + required-field checks).
 * - `pkg/models/auth_method.go:3-16` — `AuthMethod` alias, `AuthMethodBasicAuth`
 *   / `AuthMethodOAuth2` constants, and `GetAuthMethod` (defaults to basicAuth).
 * - `pkg/plugin.go:171-267` — instance factory: how each setting is consumed
 *   (`hosting` -> REST API version; `getEndpoint` -> base URL from url / cloudId;
 *   `getHttpClient` -> Basic vs Bearer transport vs OAuth2 client-credentials).
 *
 * External components consulted at their pinned versions (plugin
 * `package.json` -> monorepo `.yarnrc.yml` catalog):
 * - `@grafana/ui@^11.6.7` — `RadioButtonGroup`, `Input`, `SecretInput`,
 *   `Switch`, `InlineField`, `InlineLabel`, `Checkbox`, `Tooltip`, `Icon`.
 * - `@grafana/data@^11.6.7` — `DataSourceJsonData` (the base interface
 *   `JiraOptions` extends), `DataSourcePluginOptionsEditorProps`,
 *   `SelectableValue`, `FeatureToggles`.
 * - `@grafana/runtime@^11.6.7` — `config` (read to gate the Secure Socks Proxy
 *   section) and `reportInteraction`.
 * - `@grafana/plugin-ui@^0.13.1` — `Auth`, `ConfigSection`, `ConfigSubSection`,
 *   `DataSourceDescription`, `convertLegacyAuthProps` (the `custom-jira`
 *   authentication container and the legacy TLS panel).
 *
 * The Secure Socks Proxy checkbox (`jsonData.enableSecureSocksProxy`) is
 * deliberately excluded from this registry entry (AGENTS.md exclusion).
 */

/** Authentication method discriminator (`src/types.ts:3`). */
export type JiraAuthMethod = 'basicAuth' | 'oauth2';

/** Hosting provider (`src/types.ts:57-60`, written to `jsonData.hosting`). */
export type JiraHosting = 'cloud' | 'server';

/**
 * Root (top-level datasource settings) fields.
 *
 * The Jira plugin stores every configuration value in `jsonData` /
 * `secureJsonData` — including `url` and `user`, which live in `jsonData`, not at
 * the datasource root. The backend never reads `settings.URL`, `BasicAuth`, etc.;
 * `pkg/plugin.go:171-228` builds the client from jsonData + decrypted secrets
 * only. So `RootConfig` is a blank object rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Matches the plugin's frontend `JiraOptions`
 * (`src/types.ts:31-40`) and the json-tagged jsonData fields of the backend
 * `Settings` (`pkg/models/settings.go:14-21`). Every field below is read by the
 * backend.
 */
export type JsonDataConfig = {
  /**
   * Jira base URL, e.g. `https://mycompany.atlassian.net`
   * (`src/components/ConfigEditor.tsx:203-225`). Required; the backend prepends
   * `https://` when no scheme is present and errors when empty
   * (`pkg/models/settings.go:39-45`). Used verbatim as the API base URL for
   * non-scoped Basic Auth (`pkg/plugin.go:227`).
   */
  url?: string;
  /**
   * Atlassian account email (`src/components/ConfigEditor.tsx:266-282`, basicAuth
   * only). Used as the HTTP Basic username; when empty the backend switches to a
   * Bearer-token transport (`pkg/plugin.go:253-264`). Not enforced by
   * `LoadSettings`.
   */
  user?: string;
  /**
   * Hosting provider (`src/components/ConfigEditor.tsx:176-202`). Radio: Jira
   * Cloud=`cloud`, Jira Data Center / Jira Server=`server`
   * (`src/components/ConfigEditor.tsx:23-26`). Defaults to `cloud`
   * (`src/components/ConfigEditor.tsx:160`); forced to `cloud` and disabled when
   * `authMethod` is `oauth2` (`src/components/ConfigEditor.tsx:94-96,200`).
   * Selects the REST API version — `cloud` -> `/rest/api/3`, `server` ->
   * `/rest/api/2` (`pkg/plugin.go:177-180`).
   */
  hosting?: JiraHosting;
  /**
   * Whether the API token is a scoped Atlassian token
   * (`src/components/ConfigEditor.tsx:318-328`, basicAuth only). Defaults to
   * `false` (`src/components/ConfigEditor.tsx:161`). When `true`, requests route
   * through `https://api.atlassian.com/ex/jira/<cloudId>` and `cloudId` is
   * required (`pkg/plugin.go:221-226`).
   */
  scopedToken?: boolean;
  /**
   * Jira App Cloud Id (`src/components/ConfigEditor.tsx:329-347,390-406`). Shown
   * for scoped Basic Auth tokens and for OAuth 2.0. Required by the backend for
   * OAuth 2.0 (`pkg/models/settings.go:62-64`) and for scoped tokens
   * (`pkg/plugin.go:221-224`); selects the `api.atlassian.com` gateway base URL.
   */
  cloudId?: string;
  /**
   * Authentication method (`src/components/ConfigEditor.tsx:248-262`). Radio:
   * Basic Auth=`basicAuth`, OAuth 2.0 — Service Account=`oauth2`
   * (`src/types.ts:5-8`). Defaults to `basicAuth`
   * (`src/components/ConfigEditor.tsx:162`); the backend resolves any empty or
   * unknown value to `basicAuth` (`pkg/models/auth_method.go:11-15`).
   */
  authMethod?: JiraAuthMethod;
  /**
   * OAuth 2.0 client ID for the Jira service account
   * (`src/components/ConfigEditor.tsx:353-369`, oauth2 only). Required for OAuth
   * 2.0 (`pkg/models/settings.go:56-58`); used as `ClientID` in the
   * client-credentials grant (`pkg/plugin.go:240`).
   */
  oauthClientID?: string;
  /**
   * Written by the Secure Socks Proxy checkbox
   * (`src/components/ConfigEditor.tsx:427-437`) and consumed transparently by the
   * SDK's `config.HTTPClientOptions(ctx)` call (`pkg/models/settings.go:75`). The
   * plugin's own Go code never inspects it by name and the backend `Settings`
   * struct does not carry it. Deliberately excluded from the dsconfig registry
   * entry per AGENTS.md.
   */
  enableSecureSocksProxy?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config
 * via `secureJsonFields`). Both are declared on the frontend `JiraSecureOptions`
 * (`src/types.ts:45-48`) and copied from `DecryptedSecureJSONData` by auth
 * method in `LoadSettings` (`pkg/models/settings.go:52-72`):
 * - `token` — the Jira API token / personal access token (Basic Auth). Sent as
 *   the HTTP Basic password, or as a Bearer token when `user` is empty.
 * - `oauthClientSecret` — the OAuth 2.0 client secret for the service account
 *   (OAuth 2.0 auth).
 */
export type SecureJsonDataConfig = Array<'token' | 'oauthClientSecret'>;
