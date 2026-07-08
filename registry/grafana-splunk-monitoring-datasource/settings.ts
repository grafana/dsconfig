/**
 * Configuration models for the Splunk Infrastructure Monitoring (SignalFx)
 * datasource plugin (plugin id: `grafana-splunk-monitoring-datasource`).
 *
 * Sources of truth (github.com/grafana/plugins-private @ 267f4937806ed6404b6628d13ae358a5d308e376,
 * plugins/grafana-splunk-monitoring-datasource):
 * - `src/plugin.json:3-5` — plugin type (`type: "datasource"`), name
 *   (`"Splunk Infrastructure Monitoring"` at `:4`), id
 *   (`"grafana-splunk-monitoring-datasource"` at `:5`). Docs URL from
 *   `info.links[0]` (`:31`).
 * - `src/components/ConfigEditor.tsx:20-143` — the configuration editor:
 *   - `DataSourceDescription` with `dataSourceName="Splunk Infrastructure Monitoring"`,
 *     `docsLink="https://grafana.com/grafana/plugins/grafana-splunk-monitoring-datasource/"`,
 *     `hasRequiredFields` (`:52-56`).
 *   - `<ConfigSection title="Authentication">` (`:60-62`) containing:
 *       - `accessToken` → secureJsonData.accessToken, `<SecretInput>` with
 *         `required` (`:63-71`); reset handler at `:35-48`.
 *       - `realmName` → jsonData.realmName, `<Input placeholder="us1">`
 *         (`:72-81`).
 *       - `<ConfigSubSection title="Custom URLs" description="...">` (`:83-108`):
 *           - `url_metrics_metadata` → jsonData.url_metrics_metadata,
 *             `<Input placeholder="https://api.us1.signalfx.com">`, tooltip
 *             `"Optional Metrics MetaData URL."` (`:88-97`)
 *           - `url_signalflow` → jsonData.url_signalflow,
 *             `<Input placeholder="https://stream.us1.signalfx.com">`, tooltip
 *             `"Optional SignalFlow URL"` (`:98-107`)
 *       - `<ConfigSubSection title="Secure Socks Proxy">` (`:110-140`) — only
 *         rendered when Grafana has `secureSocksDSProxyEnabled` and version
 *         >= 10.0.0 (`:29-33`); writes `jsonData.enableSecureSocksProxy`.
 *         Deliberately excluded from the dsconfig registry entry per AGENTS.md.
 * - `src/types.ts:3-20` — frontend config types `SignalFxJsonData`
 *   (jsonData: `realmName`, `url_metrics_metadata?`, `url_signalflow?`,
 *   `enableSecureSocksProxy?`) and `SignalFxSecureJsonData` (`accessToken`).
 * - `pkg/models/settings.go:13-42` — backend `Settings` struct
 *   (`AccessToken`, `Realm` `json:"realmName"`, `URLMetricsMetaData`
 *   `json:"url_metrics_metadata"`, `URLSignalFlow` `json:"url_signalflow"`,
 *   `HttpClientOptions` `json:"-"`) and `LoadSettings`: `json.Unmarshal`
 *   (fatal on empty/malformed JSONData); require decrypted
 *   `accessToken` (`:27-30`); load HTTP client options for the proxy (`:34-39`).
 * - `pkg/client/client.go:61-76` — `NewSignalFxClient`: hard-fail
 *   `"required access token is missing"` when the token is empty (`:62-63`).
 * - `pkg/client/rest.go:225` — every request adds the `X-SF-TOKEN` header.
 * - `pkg/client/rest.go:339-353` — `GetBaseURL`: derives
 *   `https://api.{realm}.signalfx.com` (metrics-metadata) and
 *   `https://stream.{realm}.signalfx.com` (SignalFlow) from the realm, or
 *   uses `URLMetricsMetaData` / `URLSignalFlow` when set.
 *
 * External components consulted at the catalog versions the plugin pins
 * (plugins-private `.yarnrc.yml` catalog; `package.json` uses `catalog:`):
 * - `@grafana/plugin-ui@^0.13.1` — `DataSourceDescription` (renders the
 *   header block with the docs link), `ConfigSection` / `ConfigSubSection`
 *   (section titles + descriptions; no storage fields written).
 * - `@grafana/ui@^11.6.7` — `Input`, `InlineField`, `InlineSwitch`,
 *   `InlineFormLabel`, `SecretInput`, `useStyles2`.
 * - `@grafana/data@^11.6.7` — `DataSourceJsonData` (the base interface
 *   `SignalFxJsonData` extends), `DataSourcePluginOptionsEditorProps`, and
 *   the `onUpdateDatasource*` storage-key helpers.
 * - `@grafana/runtime@^11.6.7` — the `config` object read at
 *   `ConfigEditor.tsx:29-33` to decide whether to render the Secure Socks
 *   Proxy switch.
 */

/**
 * Root (top-level datasource settings) fields.
 *
 * The Splunk Infrastructure Monitoring plugin stores every configuration
 * value in `jsonData` / `secureJsonData`; nothing lives at the root level
 * (`url`, `basicAuth`, etc. are unused — the backend derives its endpoints
 * from `jsonData.realmName` and the custom URL overrides, and never reads
 * `settings.URL`). So `RootConfig` is a blank object.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Matches the plugin's frontend
 * `SignalFxJsonData` (`src/types.ts:3-10`) and backend `Settings`
 * (`pkg/models/settings.go:13-19`).
 */
export type JsonDataConfig = {
  /**
   * Splunk Observability realm (e.g. `us0`, `us1`, `eu0`, `jp0`). The
   * backend derives `https://api.{realm}.signalfx.com` (metrics-metadata)
   * and `https://stream.{realm}.signalfx.com` (SignalFlow) from it
   * (`pkg/client/rest.go:346,351`). Not marked required in the editor, but
   * an empty realm with no custom URLs yields the broken host
   * `https://api..signalfx.com`. Declared non-optional in the upstream
   * frontend type (`src/types.ts:4`).
   */
  realmName: string;
  /**
   * Optional override for the metrics-metadata REST base URL (default
   * `https://api.{realm}.signalfx.com`). Used only for custom SignalFlow
   * domains (`pkg/client/rest.go:342-346`).
   */
  url_metrics_metadata?: string;
  /**
   * Optional override for the SignalFlow streaming base URL (default
   * `https://stream.{realm}.signalfx.com`). Used only for custom SignalFlow
   * domains (`pkg/client/rest.go:347-352`).
   */
  url_signalflow?: string;
  /**
   * Written by the editor's Secure Socks Proxy `InlineSwitch`
   * (`src/components/ConfigEditor.tsx:132-137`), which is only rendered when
   * Grafana has `secureSocksDSProxyEnabled` and version >= 10.0.0. Consumed
   * transparently by the SDK proxy plumbing (`pkg/client/rest.go:307-314`
   * reads `settings.HttpClientOptions.ProxyOptions`); the plugin's own Go
   * code never inspects this field by name. Deliberately excluded from the
   * dsconfig registry entry per AGENTS.md.
   */
  enableSecureSocksProxy?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`).
 *
 * - `accessToken` — Splunk Observability API access token, sent as the
 *   `X-SF-TOKEN` header on every outgoing request (`pkg/client/rest.go:225`).
 *   Required (`pkg/models/settings.go:27-30`, `pkg/client/client.go:62-63`).
 */
export type SecureJsonDataConfig = Array<'accessToken'>;
