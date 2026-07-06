/**
 * Configuration models for the Grafana Hello World datasource plugin
 * (plugin id: `grafana-helloworld-datasource`).
 *
 * Sources of truth (github.com/grafana/plugins-private @ 267f4937806ed6404b6628d13ae358a5d308e376,
 * plugins/grafana-helloworld-datasource):
 * - `src/plugin.json:3-8` — `type` (`datasource`), `name` (`Hello World`),
 *   `id` (`grafana-helloworld-datasource`), `backend: true`,
 *   `executable: gpx_helloworld`. `info.links` is an empty array, so there is
 *   no docs URL to record.
 * - `src/module.tsx:13-14` — the frontend config types are BLANK:
 *   `export type Config = {} & DataSourceJsonData;` (no plugin-specific
 *   jsonData) and `export type SecureConfig = {};` (no secrets).
 * - `src/module.tsx:32-34` — the `ConfigEditor` renders the static fragment
 *   `<>Hello World Config Editor!</>`. It receives `options` /
 *   `onOptionsChange` props but NEVER calls `onOptionsChange`, so it persists
 *   nothing to root, jsonData, or secureJsonData.
 * - `src/module.tsx:18-29` — `DataSource extends DataSourceWithBackend<Query,
 *   Config>`; the constructor only sets `this.annotations = {}`. No settings
 *   are read on the frontend.
 * - `pkg/main.go:46-56` — the backend datasource instance factory
 *   `datasource.NewInstanceManager(func(_ context.Context, settings
 *   backend.DataSourceInstanceSettings) ...)` ignores `settings` and returns
 *   an empty `&DatasourceInstance{}`. `CheckHealth` / `QueryData`
 *   (pkg/main.go:17-38) never read instance settings.
 *
 * There is no `src/types.ts`, no `pkg/models/settings.go`, and no
 * `LoadSettings` in this plugin — nothing parses instance settings at all.
 *
 * External components consulted at their pinned versions:
 * - `@grafana/data` (catalog pin) — `DataSourcePlugin`,
 *   `DataSourcePluginOptionsEditorProps`, `DataSourceJsonData`, `DataQuery`,
 *   `DataSourceInstanceSettings`, `MetricFindValue`, `QueryEditorProps`. Only
 *   the base `DataSourceJsonData` interface is relevant to config storage, and
 *   `Config` extends it with no additional members.
 * - `@grafana/runtime` (catalog pin) — `DataSourceWithBackend`. Provides the
 *   query proxy but reads no plugin-specific settings.
 * - `@grafana/*` versions resolve via the monorepo `catalog:` protocol; see
 *   `.yarnrc.yml` at the repo root. The plugin pins no `@grafana/plugin-ui`,
 *   so NONE of the shared HTTP-settings / Auth editor components are used —
 *   that is why (unlike most datasources) there are no url / basicAuth / TLS
 *   fields here.
 */

/**
 * Root (top-level datasource settings) fields the Hello World plugin reads.
 *
 * The plugin stores and reads NOTHING at the datasource root — its config
 * editor writes no fields and its backend never inspects
 * `backend.DataSourceInstanceSettings` (pkg/main.go:46-51). Modeled as a blank
 * object (never `null`) per the registry convention.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`.
 *
 * Blank: upstream declares `export type Config = {} & DataSourceJsonData`
 * (src/module.tsx:13) — an empty interface over the base `DataSourceJsonData`.
 * The Hello World editor persists no jsonData and the backend reads none.
 */
export type JsonDataConfig = Record<string, never>;

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`).
 *
 * Upstream defines NO secrets: `export type SecureConfig = {}`
 * (src/module.tsx:14). The single `apiKey` key below is a PLACEHOLDER that is
 * never read by the plugin. It exists only because:
 *   1. a dsconfig entry must declare at least one field (the dsconfig
 *      validator rejects an empty `fields` array), and
 *   2. the shared conformance suite requires at least one `secureJsonData`
 *      key (`schema.PluginUnderTest` rejects empty `SecureKeys`, and
 *      `SchemaRoundTrip` asserts the settings schema's `SecureValues` is
 *      non-empty).
 * See the entry README for the full rationale.
 */
export type SecureJsonDataConfig = Array<'apiKey'>;
