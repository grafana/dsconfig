/**
 * Configuration models for the Grafana Sumo Logic datasource plugin
 * (plugin id: `grafana-sumologic-datasource`).
 *
 * Sources of truth — grafana/plugins-private monorepo @ 267f4937806ed6404b6628d13ae358a5d308e376,
 * plugin path `plugins/grafana-sumologic-datasource/`:
 * - `src/plugin.json:3-4,25` — plugin id (`"grafana-sumologic-datasource"`),
 *   name (`"Sumo Logic"`), docs URL
 *   (`info.links[0].url` = `"https://grafana.com/docs/plugins/grafana-sumologic-datasource"`).
 * - `src/types.ts:6` — `AuthenticationMethod = 'accessKey'`.
 * - `src/types.ts:8-14` — the frontend `Config` (jsonData) type.
 * - `src/types.ts:20-22` — the frontend `SecureConfig` (`accessKey`).
 * - `src/constants.ts:3-41` — `ApiEndpointOptions` (nine regional API URLs).
 * - `src/constants.ts:53,55` — `DefaultTimeout = 30`, `DefaultInterval = 1000`.
 * - `src/editor/ConfigEditor.tsx`:
 *   - `79-133` — the "API Region" `ConfigSection` with the `apiUrl` `Select`
 *     (`allowCustomValue`, options from `ApiEndpointOptions`, `82`/`100-109`),
 *     `timeout` number `Input` (`111-119`), and `interval` number `Input`
 *     (`120-132`).
 *   - `137-185` — the `Auth` (`@grafana/plugin-ui`) widget with a single fixed
 *     `custom-sumo` method; `onAuthMethodSelect` is a no-op (`139`). The custom
 *     method renders the `accessId` `Input` (`149-162`) and the `accessKey`
 *     `SecretInput` (`163-180`). `authMethod` itself is never written here.
 * - `pkg/models/settings.go:11-28` — backend `AuthenticationMethod` +
 *   `Settings` struct (the loaded shape and its json tags).
 * - `pkg/models/settings.go:30-54` — `LoadSettings` (defaulting: authMethod →
 *   accessKey, apiUrl → DefaultApiURL, timeout → 30, interval → 1000).
 * - `pkg/models/settings.go:56-72` — `Validate` (apiUrl, authMethod, accessId,
 *   accessKey required).
 * - `pkg/sumo/client.go:45-58` — access ID/key wired as `httpclient`
 *   `BasicAuthOptions{User: AccessID, Password: AccessKey}`; the `timeout`
 *   feeds the HTTP client timeout (`45-48`, `158-163`).
 * - `pkg/sumo/logs_query.go:80` — `interval` used as the log-polling interval.
 *
 * External components consulted at the versions pinned by the workspace catalog
 * (`.yarnrc.yml:14-26`, referenced via `catalog:` in the plugin's
 * `package.json:73-84`):
 * - `@grafana/plugin-ui@^0.13.1` — `Auth` renders a `ConfigSection` titled
 *   "Authentication" (`dist/esm/components/ConfigEditor/Auth/Auth.js`); with a
 *   single visible method `AuthMethodSettings` shows the custom method's label
 *   ("Authentication method") and description without a method `Select`
 *   (`dist/esm/.../auth-method/AuthMethodSettings.js`). `ConfigSection`,
 *   `DataSourceDescription`.
 * - `@grafana/ui@^11.6.7` — `Input`, `Select`, `SecretInput`, `InlineField`.
 * - `@grafana/data@^11.6.7` — `DataSourceJsonData` (base of the frontend
 *   `Config`), `DataSourcePluginOptionsEditorProps`, `SelectableValue`.
 */

/**
 * Authentication method discriminator. The only value the plugin supports is
 * `'accessKey'` (HTTP basic auth). Mirrors `src/types.ts:6`
 * (`AuthenticationMethod`) and `pkg/models/settings.go:11-15`.
 */
export type AuthenticationMethod = 'accessKey';

/**
 * Root (top-level datasource settings) fields.
 *
 * The Sumo Logic backend reads nothing plugin-specific at the root level: it
 * builds its own HTTP basic auth from `jsonData.accessId` and
 * `secureJsonData.accessKey` (`pkg/sumo/client.go:51-55`) rather than from the
 * datasource's root `basicAuth`/`user`/`url` fields. So this is a blank object
 * rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Matches the frontend `Config` (`src/types.ts:8-14`)
 * and the backend `Settings` json tags (`pkg/models/settings.go:21-28`).
 */
export type JsonDataConfig = {
  /**
   * Authentication method discriminator (`'accessKey'`). Declared in the
   * frontend `Config` type (`src/types.ts:9`) but the config editor never
   * writes it — the `Auth` widget uses one fixed method and a no-op
   * `onAuthMethodSelect` (`src/editor/ConfigEditor.tsx:138-139`). The backend
   * reads it, defaults it to `'accessKey'` when empty
   * (`pkg/models/settings.go:40-42`), and switches on it in
   * `pkg/sumo/client.go:51-58`.
   */
  authMethod?: AuthenticationMethod;
  /**
   * Regional Sumo Logic API base URL (`src/editor/ConfigEditor.tsx:82,100-109`;
   * options in `src/constants.ts:3-41`). Backend default is
   * `https://api.sumologic.com/api/` (`pkg/models/settings.go:17,43-45`); it is
   * used as the request base URL with one trailing slash trimmed
   * (`pkg/sumo/client.go:145`).
   */
  apiUrl?: string;
  /**
   * Sumo Logic Access Id — the HTTP basic-auth username
   * (`pkg/sumo/client.go:53`). Editor label "AccessID"
   * (`src/editor/ConfigEditor.tsx:149-162`).
   */
  accessId?: string;
  /**
   * Timeout in seconds for the data requests
   * (`src/editor/ConfigEditor.tsx:111-119`). Backend default 30
   * (`src/constants.ts:53`; `pkg/models/settings.go:18,46-48`); feeds the HTTP
   * client timeout (`pkg/sumo/client.go:45-48,158-163`). Editor enforces a min
   * of 1.
   */
  timeout?: number;
  /**
   * Interval in milliseconds for the log polling requests (min 200)
   * (`src/editor/ConfigEditor.tsx:120-132`). Backend default 1000
   * (`src/constants.ts:55`; `pkg/models/settings.go:19,49-51`); used as the
   * log-polling interval (`pkg/sumo/logs_query.go:80`).
   */
  interval?: number;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config
 * via `secureJsonFields`):
 * - `accessKey` — Sumo Logic Access Key, used as the HTTP basic-auth password
 *   (`src/types.ts:21`; `pkg/models/settings.go:39`; `pkg/sumo/client.go:54`).
 */
export type SecureJsonDataConfig = Array<'accessKey'>;
