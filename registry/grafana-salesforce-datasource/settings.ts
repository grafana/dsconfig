/**
 * Configuration models for the Salesforce datasource plugin
 * (plugin id: `grafana-salesforce-datasource`).
 *
 * Sources of truth (github.com/grafana/plugins-private @ 267f4937, plugin path
 * `plugins/grafana-salesforce-datasource`):
 * - `src/plugin.json:5,4,28` — plugin id (`"grafana-salesforce-datasource"`),
 *   name (`"Salesforce"`), docs link
 *   (`"https://grafana.com/docs/plugins/grafana-salesforce-datasource"`).
 * - `src/types.ts:4-33` — the frontend config types `AuthType`, `TokenURL`,
 *   `DatasourceOptions` (jsonData) and `SecureJsonData`.
 * - `src/constants.ts:4-12` — `AuthTypes` (radio: Credentials=`user`,
 *   JWT=`jwt`) and `TokenURLs` (Production=`https://login.salesforce.com`,
 *   SandBox=`https://test.salesforce.com`).
 * - `src/views/SFConfigEditor.tsx:161-364` — the configuration editor:
 *   - `<FieldSet label="Connection Settings">` (`:163`) with the
 *     `Authentication` radio (`:165-166`) bound to `jsonData.authType`.
 *   - authType==='user' block (`:168-241`): User Name -> jsonData.user
 *     (`:170-178`), then the SecretFormFields Password -> password
 *     (`:181-193`), Security Token -> securityToken (`:196-208`),
 *     Consumer Key -> clientID (`:211-223`), Consumer Secret -> clientSecret
 *     (`:226-238`) all writing secureJsonData.
 *   - authType==='jwt' block (`:242-305`): Certificate -> cert (`:253-259`),
 *     Private Key -> privateKey (`:267-273`), User Name -> jsonData.user
 *     (`:278-287`), Consumer Key -> clientID (`:289-301`).
 *   - `<FieldSet label="Optional Settings">` (`:307`): Environment select ->
 *     jsonData.tokenUrl (`:309-325`) and the Secure Socks Proxy switch ->
 *     jsonData.enableSecureSocksProxy (`:327-362`, excluded here).
 * - `src/selectors.ts:3-20` — the E2E aria-label selector map for the config
 *   editor inputs (labels/placeholders themselves are hard-coded in the editor).
 * - `pkg/models/settings.go:25-125` — backend `Settings` struct, `GetSettings`,
 *   `normalizeAuthType`, `normalizeTokenUrl`, and `Validate`.
 * - `pkg/plugin/client.go:76-147` — `fetchToken`: how each field feeds the
 *   OAuth2 password grant (username/password+securityToken/client_id/
 *   client_secret) vs the JWT bearer grant (assertion built from privateKey/
 *   cert/clientID/user/tokenUrl).
 * - `pkg/jwt/jwt.go:23-73` — `CreateJWT`: privateKey signs the assertion,
 *   clientID is the issuer, user is the subject, tokenUrl is the audience.
 * - `pkg/plugin/datasource.go:14-28` — instance factory: `GetSettings` then
 *   `NewClient`; the datasource's root `url` field is never read.
 *
 * External components consulted at their pinned versions (plugin
 * `package.json` -> monorepo `.yarnrc.yml` catalog):
 * - `@grafana/ui@^11.6.7` — `RadioButtonGroup`, `Input`, `Select`,
 *   `LegacyForms.SecretFormField` (Password/Security Token/Consumer Key/
 *   Consumer Secret), `CertificationKey` (Certificate/Private Key textareas),
 *   `FieldSet`, `InlineField`, `InlineFormLabel`, `InlineSwitch`.
 * - `@grafana/data@^11.6.7` — `DataSourceJsonData` (the base interface
 *   `DatasourceOptions` extends), `DataSourcePluginOptionsEditorProps`,
 *   `onUpdateDatasourceJsonDataOption`, `SelectableValue`.
 * - `@grafana/runtime@^11.6.7` — `config` (read to gate the Secure Socks Proxy
 *   switch) and `getDataSourceSrv` (used by the JWT "Generate"/"Download"
 *   certificate buttons).
 * - `@grafana/plugin-ui@^0.13.1` — test helpers only for this editor.
 *
 * The Secure Socks Proxy switch (`jsonData.enableSecureSocksProxy`) is
 * deliberately excluded from this registry entry (AGENTS.md exclusion).
 */

/** Authentication method discriminator (`src/types.ts:4`). */
export type AuthType = 'user' | 'jwt';

/** Salesforce OAuth host options (`src/types.ts:5`, `src/constants.ts:9-12`). */
export type TokenURL = 'https://login.salesforce.com' | 'https://test.salesforce.com';

/**
 * Root (top-level datasource settings) fields.
 *
 * The Salesforce plugin stores every configuration value in `jsonData` /
 * `secureJsonData`; nothing plugin-specific lives at the root level. The
 * backend never reads `settings.URL`, `BasicAuth`, etc. — `pkg/plugin/datasource.go:14-19`
 * builds the client from jsonData + decrypted secrets only. So `RootConfig`
 * is a blank object rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Matches the plugin's frontend
 * `DatasourceOptions` (`src/types.ts:7-22`) and the json-tagged jsonData
 * fields of the backend `Settings` (`pkg/models/settings.go:25-39`).
 */
export type JsonDataConfig = {
  /**
   * Authentication method (`src/types.ts:14`). Radio: Credentials=`user`,
   * JWT=`jwt` (`src/constants.ts:4-7`). Written directly by the editor
   * (`src/views/SFConfigEditor.tsx:148-154`). The backend treats an empty
   * value as `user` (`pkg/models/settings.go:82-90`).
   */
  authType: AuthType;
  /**
   * Salesforce username. Shown in both auth modes
   * (`src/views/SFConfigEditor.tsx:170-178,278-287`). Used as the OAuth2
   * password-grant username and as the JWT subject
   * (`pkg/plugin/client.go:103,96`, `pkg/jwt/jwt.go:63`). Required by
   * `Validate` only for `user` auth (`pkg/models/settings.go:112-113`).
   */
  user: string;
  /**
   * Legacy sandbox toggle (`src/types.ts:9-13`). Deprecated in favor of
   * `tokenUrl`; the editor no longer writes it but `getTokenUrl`
   * (`src/views/SFConfigEditor.tsx:39-47`) reads it for the initial
   * Environment selection, and the backend `normalizeTokenUrl`
   * (`pkg/models/settings.go:92-100`) uses it to derive the token host when
   * `tokenUrl` is empty (`true` -> test.salesforce.com).
   */
  sandbox?: boolean;
  /**
   * Salesforce OAuth host, written by the "Environment" select
   * (`src/views/SFConfigEditor.tsx:309-325`). Marked `@internal` in the type
   * (`src/types.ts:15-20`) yet driven by the editor; used to build
   * `<tokenUrl>/services/oauth2/token` (`pkg/plugin/client.go:92`) and as the
   * JWT audience (`pkg/jwt/jwt.go:62`). Empty defaults to
   * https://login.salesforce.com (or https://test.salesforce.com when
   * `sandbox` is true).
   */
  tokenUrl?: TokenURL;
  /**
   * Written by the Secure Socks Proxy switch
   * (`src/views/SFConfigEditor.tsx:350-358`) and consumed transparently by
   * the SDK's `s.HTTPClientOptions(ctx)` call in
   * `pkg/models/settings.go:73`. The plugin's own Go code never inspects it by
   * name and the backend `Settings` struct does not carry it. Deliberately
   * excluded from the dsconfig registry entry per AGENTS.md.
   */
  enableSecureSocksProxy?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`). All six are declared on the frontend
 * `SecureJsonData` (`src/types.ts:26-33`) and copied from
 * `DecryptedSecureJSONData` by auth method in `GetSettings`
 * (`pkg/models/settings.go:48-72`):
 * - `password` — Salesforce password (user auth).
 * - `securityToken` — Salesforce security token, concatenated onto the
 *   password (user auth; used but not validated).
 * - `clientID` — connected app consumer key (user auth `client_id`; jwt issuer).
 * - `clientSecret` — connected app consumer secret (user auth `client_secret`).
 * - `cert` — connected app digital-signature certificate (jwt auth).
 * - `privateKey` — connected app RSA private key signing the JWT (jwt auth).
 */
export type SecureJsonDataConfig = Array<
  'password' | 'securityToken' | 'clientID' | 'clientSecret' | 'cert' | 'privateKey'
>;
