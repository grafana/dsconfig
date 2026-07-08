/**
 * Configuration models for the Wavefront (VMware Aria Operations for
 * Applications) datasource plugin (plugin id: `grafana-wavefront-datasource`).
 *
 * Sources of truth (github.com/grafana/plugins-private @
 * 267f4937806ed6404b6628d13ae358a5d308e376, plugin path
 * `plugins/grafana-wavefront-datasource`):
 * - `src/plugin.json:5` — plugin id (`"grafana-wavefront-datasource"`),
 *   name `"Wavefront"` (`:3`), docs link
 *   `https://grafana.com/docs/plugins/grafana-wavefront-datasource`
 *   (`info.links[0]`, `:28`).
 * - `src/components/ConfigEditor.tsx:52-135` — the configuration editor:
 *   - `<h3>Wavefront settings</h3>` (`:55`) section with:
 *     - `LegacyForms.FormField` for the API URL (`:57-67`) — value is a
 *       local `url` useState initialised to `jsonData.url || DEFAULT_API_URL`
 *       (`:16`); written to `jsonData.url` on blur (`onURLUpdate`, `:36-38`).
 *     - `LegacyForms.SecretFormField` for the token (`:70-82`) — writes
 *       `secureJsonData.token` and toggles `secureJsonFields.token`
 *       (`onSettingUpdate`, `:18-32`; `isConfigured={secureJsonFields['token']}`
 *       at `:81`).
 *   - `<h3>Customization</h3>` (`:86`) section with:
 *     - `LegacyForms.FormField type="number"` for the request timeout
 *       (`:88-98`) — writes `jsonData.requestTimeout`, or `null` when the
 *       value is falsy (`onRequestTimeoutUpdate`, `:39-51`).
 *     - the Secure Socks Proxy `InlineSwitch` (`:100-132`), Grafana
 *       feature-flag + version gated, writing `jsonData.enableSecureSocksProxy`
 *       — deliberately excluded from this registry entry per AGENTS.md.
 * - `src/selectors.ts:3-27` — `DEFAULT_API_URL = 'https://try.wavefront.com'`
 *   (`:3`) plus every label / input aria-label / tooltip / placeholder the
 *   editor renders (ApiUrl `:7-12`, Token `:13-18`, RequestTimeout `:20-25`).
 *   Descriptions in dsconfig.json are the `tooltip` props verbatim.
 * - `src/types.ts:73-81` — the frontend config types `WavefrontJsonData`
 *   (`url`, `requestTimeout?`, `enableSecureSocksProxy?`) and `SecureSettings`
 *   (`token`).
 * - `pkg/models/settings.go:13-46` — backend `Settings` struct (`URL
 *   json:"url"`, `RequestTimeout int64 json:"requestTimeout"`, `Token`,
 *   `ProxyOptions`) and `LoadSettings`: seed `RequestTimeout =
 *   defaultRequestTimeout` (`:23-24`), `json.Unmarshal(config.JSONData)`
 *   (`:26-28`), copy `config.DecryptedSecureJSONData["token"]` (`:29-31`),
 *   require non-empty url (`:32-34`, `"invalid url"`) and token (`:35-37`,
 *   `"invalid credentials"`), then derive `ProxyOptions` from
 *   `config.HTTPClientOptions(ctx)` (`:39-43`).
 * - `pkg/models/constant.go:4` — `defaultRequestTimeout = 30`.
 * - `pkg/datasource/datasource.go:36-65` — instance factory: `LoadSettings`,
 *   `strings.TrimSuffix(settings.URL, "/")` (`:42`), the
 *   `Authorization: Bearer <token>` header (`:45-47`), and HTTP client
 *   construction with `settings.RequestTimeout` / `settings.ProxyOptions`.
 * - `pkg/datasource/client.go:19-22` — `getHTTPClient` coerces any timeout
 *   `<= 0` to `30` seconds.
 * - `pkg/datasource/handler_healthcheck.go:100-111` — `CheckSettings` rejects
 *   an empty URL (`"Enter an API URL."`) or token (`"Enter a token."`).
 * - `pkg/wavefront/client.go:39-44` — every outgoing request adds
 *   `Authorization: Bearer <token>`.
 *
 * External components consulted at the versions the plugin resolves through
 * the monorepo catalog (`.yarnrc.yml`) + `yarn.lock`:
 * - `@grafana/ui@11.6.14` (catalog `^11.6.7`) — `LegacyForms.FormField`,
 *   `LegacyForms.SecretFormField`, `InlineField`, `InlineSwitch`. These render
 *   the labels/tooltips/placeholders passed from `selectors.ts`; they do not
 *   write storage keys themselves (the editor's onChange handlers do).
 * - `@grafana/runtime@11.6.14` (catalog `^11.6.7`) — the `config` object read
 *   at `ConfigEditor.tsx:100` to gate the Secure Socks Proxy switch.
 * - `@grafana/data@11.6.14` (catalog `^11.6.7`) — `DataSourceJsonData` (the
 *   base interface `WavefrontJsonData` extends) and
 *   `DataSourcePluginOptionsEditorProps`.
 * - `@grafana/e2e-selectors@11.6.7` (pinned via workflow resolutions, not
 *   cataloged) — `E2ESelectors` typing for `src/selectors.ts`.
 *
 * The Secure Socks Proxy switch is deliberately excluded from this registry
 * entry (AGENTS.md exclusion for `jsonData.enableSecureSocksProxy`).
 */

/**
 * Root (top-level datasource settings) fields.
 *
 * The Wavefront plugin stores every configuration value in `jsonData` /
 * `secureJsonData`; nothing lives at the datasource root. `url` is a
 * JSON-DATA field (`jsonData.url`), NOT the root `url` —
 * `pkg/models/settings.go:15` declares `URL string \`json:"url"\`` on the
 * struct that unmarshals `config.JSONData`, and the editor writes it via
 * `onOptionsChange({ ...options, jsonData: { ...jsonData, url } })`
 * (`src/components/ConfigEditor.tsx:37`). So `RootConfig` is a blank object.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Matches the plugin's frontend
 * `WavefrontJsonData` (`src/types.ts:73-77`) and the backend `Settings`
 * jsonData-backed fields (`pkg/models/settings.go:15-16`).
 */
export type JsonDataConfig = {
  /**
   * Wavefront cluster / API base URL. The editor pre-fills the input with
   * `DEFAULT_API_URL` (`https://try.wavefront.com`, `src/selectors.ts:3`)
   * when `jsonData.url` is empty (`ConfigEditor.tsx:16`), and writes it to
   * `jsonData.url` on blur (`:36-38`). Required at both the editor
   * (health check `handler_healthcheck.go:102-104`) and the backend
   * (`pkg/models/settings.go:32-34` returns `"invalid url"` when empty). The
   * backend does NOT supply a default — the pre-fill is a frontend
   * convenience only.
   */
  url: string;
  /**
   * Per-request timeout in seconds. Backend default is `30`
   * (`pkg/models/constant.go:4`, seeded at `pkg/models/settings.go:23-24`
   * before unmarshal). The editor writes `jsonData.requestTimeout`, or `null`
   * when the entered value is falsy (`ConfigEditor.tsx:39-51`). Optional: a
   * missing/null/`<= 0` value is treated as 30 (`getHTTPClient` coerces
   * `<= 0` to 30 at `pkg/datasource/client.go:20-22`).
   */
  requestTimeout?: number;
  /**
   * Written by the Secure Socks Proxy `InlineSwitch`
   * (`src/components/ConfigEditor.tsx:120-128`), which is gated on the
   * `secureSocksDSProxyEnabled` feature toggle and Grafana >= 10.0.0
   * (`:100`). The Wavefront plugin's own Go code never inspects this field by
   * name — the SDK-provided `config.HTTPClientOptions(ctx)` call at
   * `pkg/models/settings.go:39-43` picks it up transparently. Deliberately
   * excluded from the dsconfig registry entry per AGENTS.md.
   */
  enableSecureSocksProxy?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`).
 *
 * - `token` — Wavefront API token, sent as `Authorization: Bearer <token>`
 *   on every outgoing request (`pkg/datasource/datasource.go:45-47`,
 *   `pkg/wavefront/client.go:39-44`). Required
 *   (`pkg/models/settings.go:35-37`). Declared as `SecureSettings.token`
 *   (`src/types.ts:79-81`).
 */
export type SecureJsonDataConfig = Array<'token'>;
