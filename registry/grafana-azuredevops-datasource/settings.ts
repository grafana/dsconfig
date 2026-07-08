/**
 * Configuration models for the Azure DevOps datasource plugin
 * (plugin id: `grafana-azuredevops-datasource`).
 *
 * Sources of truth (github.com/grafana/plugins-private @ 267f4937, plugin path
 * `plugins/grafana-azuredevops-datasource`):
 * - `src/plugin.json:3,4,33` — plugin name (`"Azure DevOps"`), id
 *   (`"grafana-azuredevops-datasource"`), docs link
 *   (`"https://grafana.com/docs/plugins/grafana-azuredevops-datasource"`).
 * - `src/types.ts:4-11` — the frontend config types `AzDoConfig` (jsonData:
 *   `url`, `authType`, `projectsLimit`, `enableSecureSocksProxy`, `username`)
 *   and `AzDoSecureConfig` (secureJsonData: `patToken`).
 * - `src/editors/AzDoConfigEditor.tsx:19-185` — the configuration editor:
 *   - URL `Input` -> jsonData.url (`onOptionChange('url', …)`, `:60-68`).
 *   - PAT password `Input` (with Reset button when configured) ->
 *     secureJsonData.patToken (`onSecureOptionChange('patToken', …)`, `:70-101`).
 *   - every `onOptionChange` also stamps `authType: 'patToken'` (`:35-40`).
 *   - "Optional Configuration" `Collapse` (`:104`): Projects limit number
 *     `Input` -> jsonData.projectsLimit (`:106-123`), Username `Input` ->
 *     jsonData.username (`:124-140`), and the Secure Socks Proxy `Switch` ->
 *     jsonData.enableSecureSocksProxy (`:141-180`, excluded here).
 * - `src/selectors.ts:5-40` — the `Components.ConfigEditor.AzDoSettings` map:
 *   group titles, and every field's label / placeholder / ariaLabel / tooltip
 *   (the human-readable text the editor renders).
 * - `pkg/plugin/settings.go:10-51` — backend `AzDoConfig` struct (`authType`,
 *   `url`, `projectsLimit`, `enableSecureSocksProxy`, `username`; `PATToken`
 *   from secure `patToken`), `GetSettings`, and `Validate` (url + patToken
 *   required; projectsLimit < 1 -> 100).
 * - `pkg/plugin/plugin.go:67-138` — `GetInstance`: `azuredevops.NewPatConnection(url, patToken)`
 *   by default; when `username` is set, an explicit
 *   `azuredevops.CreateBasicAuthHeaderValue(username, patToken)` header with a
 *   `normalizeURL`d (lowercased, trailing-slash-trimmed) base URL.
 * - `pkg/plugin/constants.go:6,11-12` — `PluginID`, `ErrorInvalidURL`
 *   ("invalid URL"), `ErrorInvalidPATToken` ("invalid PAT").
 *
 * External components consulted at their pinned versions (plugin `package.json`
 * -> monorepo `.yarnrc.yml` catalog):
 * - `@grafana/ui@^11.6.7` — `Button`, `Collapse`, `InlineFormLabel`, `Input`,
 *   `Switch` (generic primitives; all display text comes from `src/selectors.ts`).
 * - `@grafana/data@^11.6.7` — `DataSourceJsonData` (the base interface
 *   `AzDoConfig` extends), `DataSourcePluginOptionsEditorProps`,
 *   `DataSourceSettings`, `FeatureToggles`.
 * - `@grafana/runtime@^11.6.7` — `config` (read to gate the Secure Socks Proxy
 *   switch).
 * - `@emotion/css@11.10.6` — `css`; `semver` — `gte` (Secure Socks Proxy gate).
 *
 * The Secure Socks Proxy switch (`jsonData.enableSecureSocksProxy`) is
 * deliberately excluded from this registry entry (AGENTS.md exclusion).
 */

/**
 * Authentication type discriminator (`src/types.ts:6`, written to
 * `jsonData.authType`). The frontend type pins it to the single literal
 * `'patToken'`, and the editor stamps that value on every jsonData write
 * (`src/editors/AzDoConfigEditor.tsx:38`). The backend declares the field but
 * does not branch on it ("Not in use yet", `pkg/plugin/settings.go:11`).
 */
export type AzDoAuthType = 'patToken';

/**
 * Root (top-level datasource settings) fields.
 *
 * The Azure DevOps plugin stores every configuration value in `jsonData` /
 * `secureJsonData` — including `url`, which lives in `jsonData`, not at the
 * datasource root. The backend only reads `s.JSONData` and
 * `s.DecryptedSecureJSONData` (`pkg/plugin/settings.go:29-51`) plus
 * `s.HTTPClientOptions(ctx)` for proxy options (`pkg/plugin/plugin.go:86`); it
 * never reads `settings.URL`, `BasicAuth`, etc. So `RootConfig` is a blank
 * object rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Matches the plugin's frontend `AzDoConfig`
 * (`src/types.ts:4-10`) and the json-tagged jsonData fields of the backend
 * `AzDoConfig` struct (`pkg/plugin/settings.go:10-17`).
 */
export type JsonDataConfig = {
  /**
   * Azure DevOps organization/collection URL, e.g. `https://dev.azure.com/XXXX`
   * (label/placeholder/tooltip `src/selectors.ts:8-13`;
   * `src/editors/AzDoConfigEditor.tsx:60-68`). Required: the backend errors with
   * `ErrorInvalidURL` ("invalid URL") when empty
   * (`pkg/plugin/settings.go:20-22,35-38`) and passes it to
   * `azuredevops.NewPatConnection` (`pkg/plugin/plugin.go:74`).
   */
  url: string;
  /**
   * Authentication type (`src/selectors.ts` has no UI entry — it is set
   * programmatically). The editor stamps `'patToken'` on every jsonData write
   * (`src/editors/AzDoConfigEditor.tsx:38`); the backend unmarshals it into
   * `AzDoConfig.AuthType` but never reads the value ("Not in use yet",
   * `pkg/plugin/settings.go:11`).
   */
  authType: AzDoAuthType;
  /**
   * Maximum number of items the projects-list query returns
   * (label/placeholder/tooltip `src/selectors.ts:25-30`;
   * `src/editors/AzDoConfigEditor.tsx:106-123`). Optional; editor default 100
   * (`:28`), and the backend coerces any value < 1 to 100
   * (`pkg/plugin/settings.go:46-48`).
   */
  projectsLimit?: number;
  /**
   * Username of the user that owns the PAT
   * (label/placeholder/tooltip `src/selectors.ts:31-37`;
   * `src/editors/AzDoConfigEditor.tsx:124-140`). Optional; when set the backend
   * authenticates with an explicit HTTP Basic header
   * `CreateBasicAuthHeaderValue(username, patToken)` and a normalized URL
   * instead of the default empty-username PAT connection — needed for some
   * Azure DevOps Server versions (`pkg/plugin/plugin.go:76-84`).
   */
  username?: string;
  /**
   * Written by the Secure Socks Proxy switch
   * (`src/editors/AzDoConfigEditor.tsx:141-180`, shown only when
   * `config.featureToggles.secureSocksDSProxyEnabled` and Grafana >= 10.0.0,
   * `:31-34`) and consumed transparently by the SDK's `s.HTTPClientOptions(ctx)`
   * call (`pkg/plugin/plugin.go:86`). The backend `AzDoConfig` struct carries it
   * as `ProxyEnabled` (`pkg/plugin/settings.go:15`) but never inspects it.
   * Deliberately excluded from the dsconfig registry entry per AGENTS.md.
   */
  enableSecureSocksProxy?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config
 * via `secureJsonFields`).
 *
 * - `patToken` — the Azure DevOps personal access token (`src/types.ts:11`).
 *   Copied from `DecryptedSecureJSONData['patToken']`
 *   (`pkg/plugin/settings.go:39-41`) and used as the HTTP Basic password by
 *   `azuredevops.NewPatConnection` / `CreateBasicAuthHeaderValue`
 *   (`pkg/plugin/plugin.go:74,77`). Required: an empty token makes settings load
 *   fail with `ErrorInvalidPATToken` ("invalid PAT",
 *   `pkg/plugin/settings.go:23-25,42-44`).
 */
export type SecureJsonDataConfig = Array<'patToken'>;
