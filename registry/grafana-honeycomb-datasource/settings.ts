/**
 * Configuration models for the Honeycomb datasource plugin
 * (plugin id: `grafana-honeycomb-datasource`).
 *
 * Sources of truth (github.com/grafana/plugins-private, monorepo SHA
 * `267f4937806ed6404b6628d13ae358a5d308e376`, plugin at
 * `plugins/grafana-honeycomb-datasource`):
 * - `src/plugin.json:4-5,25` — plugin name (`"Honeycomb"`), id
 *   (`"grafana-honeycomb-datasource"`), and docs link
 *   (`https://grafana.com/docs/plugins/grafana-honeycomb-datasource`).
 * - `src/types.ts:121-134` — the frontend config types `HoneycombOptions`
 *   (jsonData) and `HoneycombSecureOptions` (secureJsonData), plus
 *   `defaultConfigOptions` (`:128-130`) which seeds `hostname` to
 *   `https://api.honeycomb.io`.
 * - `src/Views/ConfigEditor.tsx:52-157` — the configuration editor. Three
 *   `<h3 className="page-heading">` sections — "Access" (`:70`),
 *   "Environment" (`:89`), and "Advanced Settings" (`:138`) — render:
 *     - `apiKey` → secureJsonData.apiKey, `LegacyForms.SecretFormField`
 *       label "Honeycomb API Key", placeholder "Honeycomb API Key"
 *       (`:72-85`). Written on blur via `onSecureJsonUpdate('apiKey')`
 *       (`:82`); reset via `onSettingReset('apiKey')` (`:83`).
 *     - `hostname` → jsonData.hostname, `Input` under `InlineFormLabel`
 *       "URL" with tooltip "Customize the api URL. By default this will be
 *       https://api.honeycomb.io", placeholder "https://api.honeycomb.io"
 *       (`:91-104`).
 *     - `team` → jsonData.team, `Input` under `InlineFormLabel` "Team Name"
 *       with tooltip "Specify the team name. This will be useful in data
 *       links" (`:106-120`).
 *     - `environment` → jsonData.environment, `Input` under
 *       `InlineFormLabel` "Environment Name" with tooltip "Optional.
 *       Specify the environment name. This will be useful in data links"
 *       (`:121-135`).
 *     - `retentionLimit` → jsonData.retentionLimit, `Input` under
 *       `InlineFormLabel` "Time Window (days)", placeholder "7"
 *       (`:137-155`). The change handler coerces an empty value to `7` via
 *       `parseInt(event.target.value || '7')` (`:45-46`).
 *   The `<InfoBox>` at the top (`:54-68`) is the API-key help drawer.
 * - `src/components/selectors.ts:3-20` — the E2E selector labels for each
 *   config input.
 * - `src/Datasource.ts:109-136` — `honeycombUiUrl` builds the "Open in
 *   Honeycomb" data-link URL from jsonData.hostname (`'api'`->`'ui'`),
 *   team, and environment.
 * - `pkg/models/settings.go:13-19` — backend `Settings` struct
 *   (`Env`/environment, `Hostname`/hostname, `RetentionLimit`/retentionLimit,
 *   `Team`/team; `APIKey` is `json:"-"`, copied from
 *   DecryptedSecureJSONData["apiKey"]).
 * - `pkg/models/settings.go:23-39` — `LoadSettings`: seeds
 *   hostname=https://api.honeycomb.io and retentionLimit=7, then unmarshals
 *   jsonData.
 * - `pkg/models/settings.go:45-71` — `Validate`: requires a non-empty
 *   https hostname, a non-empty apiKey, and a non-empty team.
 * - `pkg/main.go:40-66` — instance factory: `LoadSettings` -> apiKey ->
 *   requestor(hostname) -> handler(retentionLimit).
 * - `pkg/httpclient/client.go:39-42` — apiKey is sent as the
 *   `X-Honeycomb-Team` header on every request.
 *
 * External components consulted at their pinned versions (`.yarnrc.yml`
 * catalog in the monorepo; the plugin references them via `catalog:`):
 * - `@grafana/ui@^11.6.7` — `Input`, `InfoBox`, `Icon`, `InlineFormLabel`,
 *   and `LegacyForms.SecretFormField` (the API-key input). These render the
 *   labels/placeholders that are all defined inline in `ConfigEditor.tsx`;
 *   none of them introduce a hidden storage key.
 * - `@grafana/data@^11.6.7` — `DataSourceJsonData` (base interface
 *   `HoneycombOptions` extends) and `DataSourcePluginOptionsEditorProps`.
 * - `@grafana/e2e-selectors` (uncataloged; swapped per Grafana version via
 *   `resolutions`) — `E2ESelectors` used by `src/components/selectors.ts`.
 *
 * This plugin's config editor renders NO Secure Socks Proxy switch, so
 * there is no `jsonData.enableSecureSocksProxy` field to exclude.
 */

/**
 * Root (top-level datasource settings) fields.
 *
 * The Honeycomb plugin stores every configuration value in `jsonData` /
 * `secureJsonData`; nothing lives at the root level. The instance factory
 * (`pkg/main.go:40-66`) builds its HTTP client purely from
 * `jsonData.hostname` + the decrypted `apiKey` secret and never reads the
 * root `url`, `basicAuth`, etc. So `RootConfig` is a blank object.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Matches the plugin's frontend
 * `HoneycombOptions` (`src/types.ts:121-126`) and backend `Settings`
 * (`pkg/models/settings.go:13-19`). Every field here is read by the backend.
 */
export type JsonDataConfig = {
  /**
   * Honeycomb API base URL. Defaults to `https://api.honeycomb.io` in both
   * the editor (`src/types.ts:129`, applied via `defaultConfigOptions` at
   * `src/Views/ConfigEditor.tsx:12`) and the backend
   * (`pkg/models/settings.go:28`). Used as the request base URL
   * (`pkg/main.go:51`, `pkg/requestor/requestor.go:27-37`) and, with its
   * `api`->`ui` substring swapped, as the base of "Open in Honeycomb" data
   * links (`pkg/plugin/querydata.go:201`, `src/Datasource.ts:126`).
   * Required and must be https (`pkg/models/settings.go:48-60`).
   */
  hostname: string;
  /**
   * Honeycomb team slug. Required at the backend
   * (`pkg/models/settings.go:65-68` returns "enter a Honeycomb team name").
   * Functionally used only to build data-link URLs
   * (`pkg/plugin/querydata.go:209`, `src/Datasource.ts:127`) — it is not
   * part of API authentication.
   */
  team: string;
  /**
   * Optional Honeycomb environment name. Used only in data-link URLs when
   * set (`pkg/plugin/querydata.go:210-211`, `src/Datasource.ts:128-130`).
   * Not validated by the backend.
   */
  environment?: string;
  /**
   * Optional maximum query time window in days. Defaults to `7`
   * (`pkg/models/settings.go:29`; the editor coerces an empty input to `7`
   * at `src/Views/ConfigEditor.tsx:46`). The backend converts it to a
   * duration (`pkg/main.go:64`) and clamps query start times, attaching a
   * "Partial results" warning when data is clipped
   * (`pkg/plugin/querydata.go:281-318`). Not validated by the backend.
   */
  retentionLimit?: number;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`).
 *
 * - `apiKey` — Honeycomb Team API key, sent as the `X-Honeycomb-Team`
 *   header on every outgoing request (`pkg/httpclient/client.go:39-42`).
 *   Required (`pkg/models/settings.go:61-64`). Declared frontend-side as
 *   `HoneycombSecureOptions.apiKey` (`src/types.ts:132-134`).
 */
export type SecureJsonDataConfig = Array<'apiKey'>;
