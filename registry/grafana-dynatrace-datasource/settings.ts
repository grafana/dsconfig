/**
 * Configuration models for the Grafana Dynatrace datasource plugin
 * (plugin id: `grafana-dynatrace-datasource`).
 *
 * Sources of truth — grafana/plugins-private monorepo @ 267f4937806ed6404b6628d13ae358a5d308e376,
 * plugin path `plugins/grafana-dynatrace-datasource/`:
 * - `src/plugin.json:3-4,29` — plugin name (`"Dynatrace"`), id
 *   (`"grafana-dynatrace-datasource"`), docs URL
 *   (`info.links[0].url` = `"https://grafana.com/docs/plugins/grafana-dynatrace-datasource"`).
 * - `src/components/config/ConfigEditor.tsx` — the config editor
 *   (`DynatraceConfigEditor`). A flat list of `@grafana/ui` `InlineField`s
 *   (no `@grafana/plugin-ui` sections):
 *   - `apiType` RadioButtonGroup (`:65-71`), options `DynatraceConfigAPITypes`
 *     (`:21-25`: SaaS Environment/`saas`, Managed Cluster/`managed`,
 *     Raw URL/`url`); default `jsonData.apiType || 'saas'` (`:68`).
 *   - `environmentId` Input (`:73-90`): dynamic label
 *     `apiType === 'url' ? 'URL' : 'Environment ID'` (`:75`), dynamic tooltip
 *     (`:77-81`) and placeholder (`:88`).
 *   - `domain` Input (`:92-105`): rendered only when `apiType === 'managed'`,
 *     label `'Domain'`, placeholder `'Your Domain'`.
 *   - `apiToken` SecretInput (`:107-124`) -> secureJsonData.apiToken.
 *   - `platformToken` SecretInput (`:127-144`) -> secureJsonData.platformToken.
 *   - `httpClientTimeout` number Input (`:147-162`): default 30, min 0.
 *   - `tlsSkipVerify` Checkbox (`:164-169`), label `'Skip TLS Verify'`.
 *   - `tlsAuthWithCACert` Checkbox (`:171-182`), label `'With CA Cert'`.
 *   - `tlsCACert` SecretTextArea (`:184-200`): rendered only when
 *     `tlsAuthWithCACert`, label `'CA Cert'`, placeholder
 *     `'Begins with -----BEGIN CERTIFICATE-----'`, rows 5.
 *   - Secure Socks Proxy Checkbox (`:202-231`, writes
 *     `jsonData.enableSecureSocksProxy`) — EXCLUDED, see below.
 * - `src/selectors.ts:8-51` — the E2E selector map supplying the static
 *   labels/tooltips/placeholders the editor renders (Descriptions in the
 *   dsconfig entry are these tooltips).
 * - `src/types.ts:6-20` — the frontend types `DynatraceConfigAPIType`,
 *   `DynatraceDataSourceOptions` (jsonData), `DynatraceDataSourceSecureOptions`.
 * - `pkg/models/settings.go:16-78` — backend `Settings` struct + `LoadSettings`
 *   (unmarshal jsonData, copy apiToken/platformToken/tlsCACert from
 *   secureJsonData, default httpClientTimeout to 30 when <= 0, load proxy opts)
 *   and `Validate` (token presence + CA-cert PEM checks).
 * - `pkg/dynatrace/client/rest.go:24-53,62-68,170-179` — URL construction per
 *   apiType (saas/managed/url), the SaaS Grail `.live` -> `.apps` host switch,
 *   and the `Api-Token` (apiToken) vs `Bearer` (platformToken) auth headers.
 * - `pkg/dynatrace/handler_healthcheck.go:141-159` — `CheckSettings`:
 *   environmentId required, domain required when managed, at least one token,
 *   then `Settings.Validate()`.
 *
 * External components consulted at the versions pinned by the workspace catalog
 * (`.yarnrc.yml:14-26`, referenced via `catalog:` in the plugin's
 * `package.json:30-43`):
 * - `@grafana/ui@^11.6.7` — `RadioButtonGroup`, `Input`, `Checkbox`,
 *   `SecretInput`, `SecretTextArea`, `InlineField`, `InlineFormLabel` (the
 *   config editor's rendered labels/placeholders/tooltips).
 * - `@grafana/data@^11.6.7` — `DataSourceJsonData` (the base interface
 *   `DynatraceDataSourceOptions` extends), `DataSourcePluginOptionsEditorProps`,
 *   `SelectableValue`.
 * - `@grafana/runtime@^11.6.7` — `config` (feature toggle + build version read
 *   at `ConfigEditor.tsx:202` to decide whether to render the Secure Socks
 *   Proxy switch).
 * - `@grafana/e2e-selectors` (NOT cataloged — swapped per Grafana version) —
 *   `E2ESelectors` type backing `src/selectors.ts`.
 *
 * The Secure Socks Proxy switch (`jsonData.enableSecureSocksProxy`,
 * `ConfigEditor.tsx:202-231`) is deliberately excluded per AGENTS.md.
 */

/**
 * Connection type. Stored under `jsonData.apiType`. Selects how the API base
 * URL is built (`pkg/dynatrace/client/rest.go:24-53`). Mirrors
 * `src/types.ts:6` (`DynatraceConfigAPIType`) and
 * `pkg/models/settings.go:16-20` (`SettingsAPIType*`). Editor default `'saas'`
 * (`ConfigEditor.tsx:68`).
 */
export type DynatraceConfigAPIType = 'saas' | 'managed' | 'url';

/**
 * Root (top-level datasource settings) fields.
 *
 * The Dynatrace plugin stores every configuration value in `jsonData` /
 * `secureJsonData`; nothing lives at the root level. `pkg/models/settings.go`
 * unmarshals only `config.JSONData` and reads only decrypted secrets — the
 * root `url`, `basicAuth`, etc. are never read (URLs are built from
 * `jsonData.apiType`/`environmentId`/`domain`). So `RootConfig` is a blank
 * object rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Matches the frontend
 * `DynatraceDataSourceOptions` (`src/types.ts:7-15`) and the json-tagged fields
 * of the backend `Settings` struct (`pkg/models/settings.go:23-37`).
 */
export type JsonDataConfig = {
  /**
   * Connection type discriminator. Selects the API base URL shape
   * (`pkg/dynatrace/client/rest.go:24-53`). Editor default `'saas'`
   * (`ConfigEditor.tsx:68`); the backend `GetHostURL` also treats an empty
   * value as SaaS (`rest.go:44-52`).
   */
  apiType: DynatraceConfigAPIType;
  /**
   * The environment identifier. Its meaning depends on `apiType`:
   * - `'saas'`: the SaaS environment/tenant ID; URL becomes
   *   `https://<environmentId>.live.dynatrace.com/api/...`.
   * - `'managed'`: the environment ID under a Managed cluster; URL becomes
   *   `https://<domain>/e/<environmentId>/api/...`.
   * - `'url'`: the full base URL (the editor relabels the field to `'URL'`),
   *   with `/api/...` appended.
   * Always required (`pkg/dynatrace/handler_healthcheck.go:142-144`).
   */
  environmentId: string;
  /**
   * Managed cluster host. Only used (and shown) when `apiType === 'managed'`
   * (`ConfigEditor.tsx:92-105`); required in that mode
   * (`pkg/dynatrace/handler_healthcheck.go:147-149`). Ignored for `saas`/`url`.
   */
  domain: string;
  /**
   * Skip TLS certificate verification. Read by `getHTTPClient` to set
   * `httpclient.TLSOptions.InsecureSkipVerify`
   * (`pkg/dynatrace/client/rest.go:159-161`).
   */
  tlsSkipVerify?: boolean;
  /**
   * Enable a custom CA certificate. When true, the editor shows the `tlsCACert`
   * textarea (`ConfigEditor.tsx:184`) and the backend requires a valid PEM
   * (`pkg/models/settings.go:68-76`).
   */
  tlsAuthWithCACert?: boolean;
  /**
   * HTTP client timeout in seconds. Editor default 30 (`ConfigEditor.tsx:156`);
   * the backend defaults it to 30 when unset or `<= 0`
   * (`pkg/models/settings.go:50-53`).
   */
  httpClientTimeout?: number;
  /**
   * Written by the Secure Socks Proxy Checkbox
   * (`src/components/config/ConfigEditor.tsx:202-231`) and consumed
   * transparently by the SDK's `config.HTTPClientOptions(ctx)` call
   * (`pkg/models/settings.go:55-59`). The Dynatrace plugin's own Go code never
   * reads it by name, and the backend `Settings` struct does not carry it.
   * Deliberately EXCLUDED from the dsconfig registry entry per AGENTS.md.
   */
  enableSecureSocksProxy?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config
 * via `secureJsonFields`). Mirrors `DynatraceDataSourceSecureOptions`
 * (`src/types.ts:16-20`) and the keys `LoadSettings` copies
 * (`pkg/models/settings.go:46-48`):
 * - `apiToken` — classic Dynatrace API token, sent as
 *   `Authorization: Api-Token <token>` for the classic API endpoints
 *   (`pkg/dynatrace/client/rest.go:171-172`). Required unless `platformToken`
 *   is set.
 * - `platformToken` — Dynatrace platform token, sent as
 *   `Authorization: Bearer <token>` for the Grail platform API
 *   (`pkg/dynatrace/client/rest.go:173-174`). Required unless `apiToken` is
 *   set.
 * - `tlsCACert` — PEM CA certificate, used when `tlsAuthWithCACert` is true
 *   (`pkg/dynatrace/client/rest.go:162-164`; validated in
 *   `pkg/models/settings.go:68-76`).
 */
export type SecureJsonDataConfig = Array<'apiToken' | 'platformToken' | 'tlsCACert'>;
