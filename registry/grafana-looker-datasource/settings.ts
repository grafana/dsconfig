/**
 * Configuration models for the Grafana Looker datasource plugin
 * (plugin id: `grafana-looker-datasource`).
 *
 * Sources of truth — grafana/plugins-private monorepo @ 267f4937806ed6404b6628d13ae358a5d308e376,
 * plugin path `plugins/grafana-looker-datasource/`:
 * - `src/plugin.json:3-4,24` — plugin id (`"grafana-looker-datasource"`), name
 *   (`"Looker"`), docs URL
 *   (`info.links[0].url` = `"https://grafana.com/docs/plugins/grafana-looker-datasource"`).
 * - `src/editors/configEditor.tsx` — the `LookerConfigEditor`:
 *   - `base_url` `Input` (`:34-48`), written to `jsonData.base_url` on blur
 *     (`:44`).
 *   - `auth_type` `RadioButtonGroup` (`:50-64`) — only rendered when
 *     `authOptions.length > 1` (`:50`); since `authOptions` has exactly one
 *     entry (`src/constants.ts:4-6`) the selector is never shown. Its value is
 *     `jsonData.auth_type || 'client_secret'` (`:60`).
 *   - `client_id` `Input` (`:67-81`) and `client_secret` `SecretInput`
 *     (`:82-97`), both rendered only when
 *     `jsonData.auth_type === 'client_secret' || !jsonData.auth_type` (`:65`).
 *     `client_id` writes `jsonData.client_id` on blur (`:77`); `client_secret`
 *     writes `secureJsonData.client_secret` (`:95`).
 *   - `BetaNotice` (`:100`, `src/editors/betaNotice.tsx`) — a public-preview
 *     info alert; no config fields.
 * - `src/selectors.ts:1-29` — `Components` supplies every label, tooltip, and
 *   placeholder the editor renders.
 * - `src/types.ts:12-24` — the frontend types `AuthenticationType`, `Config`,
 *   `SecureKey`, `SecureConfig`.
 * - `pkg/models/config.go:14-56` — backend `Config` (`base_url`, `auth_type`,
 *   `client_id`; `client_secret` and `HttpClientOptions` are `json:"-"`),
 *   `AuthType` (`:22-26`), `Validate` (`:28-45`), and `ApplyDefaults` (`:47-56`,
 *   defaults `auth_type` to `client_secret`, trims `base_url`/`client_id`/
 *   `client_secret` and strips a trailing slash from `base_url`).
 * - `pkg/models/config.go:58-74` — `LoadConfig`: unmarshal `jsonData`, read
 *   `client_secret` from `DecryptedSecureJSONData`, call `ApplyDefaults`.
 *   Validation runs separately in the health check
 *   (`pkg/handler_healthcheck.go:13`).
 * - `pkg/looker/client.go:22-38` — `NewClient(baseUrl, clientId, clientSecret)`
 *   builds an `rtl.AuthSession` targeting Looker API version `"4.0"`.
 *
 * External components consulted at the versions pinned by the workspace catalog
 * (`.yarnrc.yml:14-26`, referenced via `catalog:` in the plugin's
 * `package.json:34-42`):
 * - `@grafana/ui@^11.6.7` — `Input`, `RadioButtonGroup`, `InlineField`,
 *   `SecretInput`, `Alert`. `SecretInput` renders the write-only client secret
 *   and its "Reset" affordance; it does not add any storage keys beyond
 *   `secureJsonData.client_secret`.
 * - `@grafana/data@^11.6.7` — `DataSourceJsonData` (base of the frontend
 *   `Config` type), `DataSourcePluginOptionsEditorProps`,
 *   `onUpdateDatasourceSecureJsonDataOption`, `updateDatasourcePluginResetOption`.
 */

/**
 * Authentication type. Stored under `jsonData.auth_type`. The plugin exposes a
 * single value, `'client_secret'` (Looker API3 credentials). Mirrors
 * `src/types.ts:13` (`AuthenticationType`) and `pkg/models/config.go:22-26`
 * (`AuthType`).
 */
export type AuthenticationType = 'client_secret';

/**
 * Root (top-level datasource settings) fields.
 *
 * The Looker datasource stores nothing plugin-specific at the root level: the
 * backend reads `base_url` from `jsonData` (not the root `url`) and never reads
 * named root fields such as `url`, `basicAuth`, or `user`
 * (`pkg/models/config.go:58-74`). So this is a blank object rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Matches the frontend `Config` (`src/types.ts:14-18`)
 * and the backend `Config` json-tagged fields (`pkg/models/config.go:15-17`).
 */
export type JsonDataConfig = {
  /**
   * Looker instance base URL, e.g. `https://<instance>.looker.app`. Required.
   * The backend trims whitespace and a single trailing slash
   * (`pkg/models/config.go:51-52`) and the Looker SDK targets the `4.0` API
   * version against it (`pkg/looker/client.go:24-25`).
   */
  base_url: string;
  /**
   * Authentication discriminator. Editor default is `'client_secret'`
   * (`src/editors/configEditor.tsx:60`); the backend also defaults an empty
   * value to `'client_secret'` (`pkg/models/config.go:48-50`). The editor's
   * radio selector is currently never rendered because only one auth option
   * exists (`src/editors/configEditor.tsx:50`, `src/constants.ts:4-6`).
   */
  auth_type: AuthenticationType;
  /**
   * Looker API3 client ID. Required when `auth_type` is `'client_secret'`
   * (`pkg/models/config.go:33-36`).
   */
  client_id: string;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config
 * via `secureJsonFields`):
 * - `client_secret` — Looker API3 client secret. Read by the backend from
 *   `DecryptedSecureJSONData["client_secret"]` (`pkg/models/config.go:69`) and
 *   required when `auth_type` is `'client_secret'` (`pkg/models/config.go:37-39`).
 *
 * Mirrors `src/types.ts:22` (`SecureKey`).
 */
export type SecureJsonDataConfig = Array<'client_secret'>;
