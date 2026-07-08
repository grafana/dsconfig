/**
 * Configuration models for the Jenkins datasource plugin
 * (plugin id: `grafana-jenkins-datasource`).
 *
 * Sources of truth (https://github.com/grafana/jenkins-datasource @ 7f8efb4):
 * - `src/plugin.json:5` — plugin id (`"grafana-jenkins-datasource"`),
 *   name (`"Jenkins"` at `:3`), `info.links[0]` Docs URL at `:35`
 *   (`https://grafana.com/docs/plugins/grafana-jenkins-datasource`).
 * - `src/components/ConfigEditor.tsx:65-122` — the configuration editor:
 *   - `DataSourceDescription` with `dataSourceName="Jenkins"`,
 *     `docsLink="https://grafana.com/docs/plugins/grafana-jenkins-datasource/"`,
 *     `hasRequiredFields={true}` (`:67-71`).
 *   - `<Legend>Connection</Legend>` (`:73`) and the URL `<InlineField>`
 *     (`:74-83`): label `"URL"`, `required`, `invalid={!jsonData.url}`,
 *     `error={'URL is required'}`, tooltip and placeholder both
 *     `"Jenkins URL, e.g. https://jenkins.example.com"`, writing
 *     `jsonData.url` via `onUrlChange` (`:16-24`).
 *   - `<Legend>Authentication</Legend>` (`:85`) and:
 *       - User `<InlineField>` (`:86-95`): label `"User"`, tooltip
 *         `"The username to use for authentication"`, placeholder
 *         `"Username"`, `autoComplete="off"`, writing `jsonData.username`
 *         via `onUsernameChange` (`:26-34`).
 *       - Password `<InlineField>` (`:96-106`): label `"Password"`,
 *         tooltip `"The password to use for authentication"`,
 *         placeholder `"Password"`, a `<SecretInput>` bound to
 *         `secureJsonFields.password` / `secureJsonData?.password`,
 *         writing `secureJsonData.password` via `onPasswordChange`
 *         (`:36-43`) and clearing it via `onResetPassword` (`:45-57`).
 *   - Optional Secure Socks Proxy `<ConfigSubSection>` (`:108-121`),
 *     gated on `config.featureToggles.secureSocksDSProxyEnabled` and
 *     Grafana `>=10.0.0` (`:59-63`). Writes `jsonData.enableSecureSocksProxy`
 *     via `onUpdateDatasourceJsonDataOptionChecked(props, 'enableSecureSocksProxy')`
 *     — deliberately excluded from this registry entry per AGENTS.md.
 * - `src/types.ts:35-46` — the frontend config types `JenkinsConfig`
 *   (extends `DataSourceJsonData`; `url?`, `username?`,
 *   `enableSecureSocksProxy?`) and `JenkinsSecureConfig` (`password?`).
 * - `pkg/plugin/settings.go:10-29` — backend `Settings` struct
 *   (`URL`, `Username`, and an unexported `Password`) and `LoadSettings`:
 *   unconditional `json.Unmarshal(source.JSONData, &settings)`; fatal
 *   `DownstreamError("URL is missing")` when `settings.URL == ""`
 *   (`:23-25`); then `settings.Password = source.DecryptedSecureJSONData["password"]`.
 * - `pkg/plugin/datasource.go:50-87` — instance factory: `LoadSettings`,
 *   `dss.HTTPClientOptions(ctx)`, set a 5-minute timeout, then
 *   `if settings.Username != ""` wire `httpclient.BasicAuthOptions{User,
 *   Password}` (`:66-71`). Without a username the client is
 *   unauthenticated even when a password is set.
 * - `pkg/jenkins/client.go:222-257` — `NewClient` normalizes the base
 *   URL with `strings.TrimRight(baseURL, "/")` (`:226`), so a trailing
 *   slash is tolerated. API paths (`/api/json`, `/computer/api/json`,
 *   `/queue/api/json`, `/job/<name>/api/json`,
 *   `/label/<name>/api/json`) are joined via `urlJoin` (`:525-527`).
 * - `pkg/jenkins/client.go:498-500` — every outgoing request that has
 *   credentials sets `req.SetBasicAuth(username, password)`.
 *
 * External editor components consulted at their pinned versions:
 * - `@grafana/plugin-ui@0.10.7` (resolved via `^0.10.4`) —
 *   `DataSourceDescription`, `ConfigSubSection`.
 * - `@grafana/ui@10.4.19` (resolved via `^10.4.8`) — `Divider`,
 *   `InlineField`, `InlineFormLabel`, `InlineSwitch`, `Input`, `Legend`,
 *   `SecretInput`.
 * - `@grafana/data@10.4.19` (resolved via `^10.4.8`) —
 *   `DataSourceJsonData` (the base interface `JenkinsConfig` extends),
 *   `DataSourcePluginOptionsEditorProps`, `FeatureToggles`,
 *   `onUpdateDatasourceJsonDataOptionChecked` (writes the Secure Socks
 *   Proxy toggle).
 * - `@grafana/runtime@10.4.19` (resolved via `^10.4.8`) — `config`,
 *   read at `ConfigEditor.tsx:61-62` to decide whether to render the
 *   Secure Socks Proxy sub-section.
 *
 * The Secure Socks Proxy switch is deliberately excluded from this
 * registry entry per AGENTS.md
 * (`jsonData.enableSecureSocksProxy`).
 */

/**
 * Root (top-level datasource settings) fields.
 *
 * The Jenkins plugin stores every configuration value in `jsonData` /
 * `secureJsonData`; nothing lives at the root level. `url` is a
 * JSON-DATA field (`jsonData.url`), NOT the root `url` — the frontend
 * `JenkinsConfig` (`src/types.ts:35-39`) declares `url?: string` on the
 * type that becomes `options.jsonData`, and the backend `Settings`
 * struct (`pkg/plugin/settings.go:10-14`) unmarshals `url` off
 * `source.JSONData`, not off `source.URL`. So `RootConfig` is a blank
 * object.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Matches the plugin's frontend
 * `JenkinsConfig` (`src/types.ts:35-39`) and backend `Settings`
 * (`pkg/plugin/settings.go:10-14`).
 */
export type JsonDataConfig = {
  /**
   * Jenkins base URL. Editor placeholder / tooltip:
   * `"Jenkins URL, e.g. https://jenkins.example.com"`
   * (`src/components/ConfigEditor.tsx:75,80`). Required in the editor
   * (`required`, `invalid={!jsonData.url}`, error text `"URL is required"`
   * at `:74`) AND at the backend (`pkg/plugin/settings.go:23-25`
   * returns `DownstreamError("URL is missing")` when empty).
   * `pkg/jenkins/client.go:226` normalizes the value with
   * `strings.TrimRight(baseURL, "/")`, so a trailing slash is
   * tolerated but not necessary.
   */
  url?: string;
  /**
   * Basic-auth username. Editor label `"User"`, tooltip
   * `"The username to use for authentication"`, placeholder
   * `"Username"` (`src/components/ConfigEditor.tsx:86-95`). Optional at
   * both the editor and the backend: `pkg/plugin/datasource.go:66-71`
   * only wires `httpclient.BasicAuthOptions{User, Password}` when
   * `settings.Username != ""`. Leaving it blank makes the client
   * anonymous.
   */
  username?: string;
  /**
   * Written by the SDK-provided Secure Socks Proxy switch inside the
   * conditional `<ConfigSubSection>` at
   * `src/components/ConfigEditor.tsx:108-121`; the toggle is only
   * rendered when `config.featureToggles.secureSocksDSProxyEnabled` is
   * on and Grafana `>=10.0.0`. The Jenkins plugin's own Go code never
   * inspects this field by name; `dss.HTTPClientOptions(ctx)` at
   * `pkg/plugin/datasource.go:57` picks it up transparently.
   * Deliberately excluded from the dsconfig registry entry per
   * AGENTS.md.
   */
  enableSecureSocksProxy?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read
 * existing config via `secureJsonFields`).
 *
 * - `password` — Jenkins user password (or API token). Consumed at
 *   `pkg/plugin/settings.go:27` via
 *   `source.DecryptedSecureJSONData["password"]`, then passed to
 *   `httpclient.BasicAuthOptions.Password` at
 *   `pkg/plugin/datasource.go:66-71`. Only actually used when
 *   `jsonData.username != ""`.
 */
export type SecureJsonDataConfig = Array<'password'>;
