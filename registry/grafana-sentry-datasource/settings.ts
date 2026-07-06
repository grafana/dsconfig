/**
 * Configuration models for the Sentry datasource plugin
 * (plugin id: `grafana-sentry-datasource`).
 *
 * Sources of truth (https://github.com/grafana/sentry-datasource @ cdb55de):
 * - `src/plugin.json:5` — plugin id (`"grafana-sentry-datasource"`),
 *   name (`"Sentry"` at `:4`). `info.links` is empty at `:22`; the editor
 *   hard-codes its docs link at `src/editors/SentryConfigEditor.tsx:49`.
 * - `src/editors/SentryConfigEditor.tsx:47-137` — the configuration editor
 *   itself:
 *   - `DataSourceDescription` with `dataSourceName="Sentry"`,
 *     `docsLink="https://grafana.com/grafana/plugins/grafana-sentry-datasource/"`,
 *     `hasRequiredFields` (`:47-51`).
 *   - `<ConfigSection title="Sentry Settings">` (`:53`) containing three
 *     required `<Field>`s that all write to `jsonData` (via
 *     `onOptionChange` at `:24-32`) or to `secureJsonData` (via
 *     `onSecureOptionChange` at `:33-43`):
 *       - `url` → jsonData.url, placeholder `https://sentry.io`
 *         (`:63,66-68`)
 *       - `orgSlug` → jsonData.orgSlug, placeholder `Sentry org slug`
 *         (`:80,83-85`)
 *       - `authToken` → secureJsonData.authToken, secret input with a
 *         Reset button when `secureJsonFields.authToken` is set (`:88-133`)
 *   - `<AdditionalSettings>` (`:135`) — a collapsible ConfigSection
 *     rendering the Secure Socks Proxy switch (only when Grafana has
 *     `secureSocksDSProxyEnabled` and version >= 10.0.0,
 *     `src/components/config-editor/AdditionalSettings.tsx:17,29-40`) and
 *     the `tlsSkipVerify` Switch (`:41-50`).
 * - `src/selectors.ts:6-38` — the E2E selector map that supplies every
 *   label, placeholder, and tooltip the editor renders. Descriptions in
 *   dsconfig.json are the tooltips (bound to `<Field description={...}>`).
 * - `src/constants.ts:105` — `DEFAULT_SENTRY_URL = 'https://sentry.io'`,
 *   used both as the URL placeholder and the initial value of the URL
 *   `useState` at `SentryConfigEditor.tsx:19`.
 * - `src/types.ts:54-62` — the frontend config types `SentryConfig` and
 *   `SentrySecureConfig`.
 * - `pkg/plugin/settings.go:11-52` — backend `SentryConfig` struct
 *   (`URL`, `OrgSlug`, `TLSSkipVerify`, unexported `authToken`) and
 *   `GetSettings`: `json.Unmarshal(s.JSONData, config)` (fatal on empty
 *   JSONData); default `URL = "https://sentry.io"` when empty
 *   (`:37-40`); required `OrgSlug` (`:41-43`); required `authToken`
 *   copied from `s.DecryptedSecureJSONData["authToken"]` (`:44-50`);
 *   final `Validate()` call.
 * - `pkg/plugin/plugin.go:44-73` — instance factory: `GetSettings`,
 *   `s.HTTPClientOptions(ctx)`, then `if settings.TLSSkipVerify` set
 *   `opt.TLS.InsecureSkipVerify = true` (`:57-63`), then
 *   `sentry.NewSentryClient(URL, OrgSlug, authToken, ...)`.
 * - `pkg/sentry/sentry.go:23-33` — `NewSentryClient` also defaults an
 *   empty base URL to `DefaultSentryURL`.
 * - `pkg/sentry/client.go:37-40` — every request adds
 *   `Authorization: Bearer <authToken>`.
 * - `pkg/errors/errors.go:15-19` — the three fatal errors
 *   (`ErrorUnmarshalingSettings`, `ErrorInvalidOrganizationSlug`,
 *   `ErrorInvalidAuthToken`).
 * - `pkg/plugin/settings_test.go:13-44` — confirms empty JSONData is a
 *   parse error and default URL is applied when empty.
 *
 * External components consulted at their pinned versions:
 * - `@grafana/plugin-ui@0.13.0` — `DataSourceDescription` (renders the
 *   header block with the docs link), `ConfigSection` (renders section
 *   titles with an optional collapsible chevron).
 * - `@grafana/ui@12.4.3` (resolved via `^12.2.0`) — `Field`, `Input`,
 *   `Button`, `Switch`, `Divider`.
 * - `@grafana/data@12.4.3` (resolved via `^12.2.0`) — `DataSourceJsonData`
 *   (the base interface `SentryConfig` extends),
 *   `DataSourcePluginOptionsEditorProps`.
 * - `@grafana/runtime@^12.2.0` — the `config` object read by
 *   `AdditionalSettings.tsx:17` to decide whether to render the Secure
 *   Socks Proxy switch.
 *
 * The Secure Socks Proxy switch is deliberately excluded from this
 * registry entry (AGENTS.md exclusion for `jsonData.enableSecureSocksProxy`).
 */

/**
 * Root (top-level datasource settings) fields.
 *
 * The Sentry plugin stores every configuration value in `jsonData` /
 * `secureJsonData`; nothing lives at the root level. `url` is a JSON-DATA
 * field (`jsonData.url`), NOT the root `url` — `pkg/plugin/settings.go:12`
 * declares `URL string \`json:"url"\`` on the struct that unmarshals
 * JSONData. So `RootConfig` is a blank object.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Matches the plugin's frontend
 * `SentryConfig` (`src/types.ts:54-59`) and backend `SentryConfig`
 * (`pkg/plugin/settings.go:11-16`).
 */
export type JsonDataConfig = {
  /**
   * Sentry base URL. Editor default is `https://sentry.io`
   * (`src/constants.ts:105`, applied via the URL useState at
   * `src/editors/SentryConfigEditor.tsx:19`). The backend also defaults
   * an empty value to `https://sentry.io` (`pkg/plugin/settings.go:37-40`).
   * Required in the editor (`invalid={!url}` at `:58`), but not strictly
   * required at the backend because of that default.
   */
  url: string;
  /**
   * Sentry organization slug (the last segment of
   * `https://sentry.io/organizations/{organization_slug}/`). Required at
   * both the editor (`invalid={!jsonData.orgSlug}` at `:75`) and the
   * backend (`pkg/plugin/settings.go:41-43` returns
   * `ErrorInvalidOrganizationSlug`).
   */
  orgSlug: string;
  /**
   * Skip TLS certificate verification when talking to a self-hosted
   * Sentry instance with a self-signed / private CA certificate. Read by
   * `pkg/plugin/plugin.go:57-63` to set `opt.TLS.InsecureSkipVerify = true`
   * on the SDK HTTP client options.
   */
  tlsSkipVerify?: boolean;
  /**
   * Written by `@grafana/plugin-ui`'s Secure Socks Proxy Switch inside
   * `AdditionalSettings` (`src/components/config-editor/AdditionalSettings.tsx:29-40`).
   * The Sentry plugin's own Go code never inspects this field by name;
   * the SDK-provided `s.HTTPClientOptions(ctx)` call in
   * `pkg/plugin/plugin.go:51` picks it up. Deliberately excluded from
   * the dsconfig registry entry per AGENTS.md.
   */
  enableSecureSocksProxy?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`).
 *
 * - `authToken` — Sentry API Bearer token, sent as
 *   `Authorization: Bearer <authToken>` on every outgoing request
 *   (`pkg/sentry/client.go:37-40`). Required (`pkg/plugin/settings.go:47-50`).
 */
export type SecureJsonDataConfig = Array<'authToken'>;
