/**
 * Configuration models for the OpenTSDB datasource plugin (`opentsdb`).
 *
 * Sources of truth (https://github.com/grafana/grafana-opentsdb-datasource @ 569fe9d):
 * - `src/plugin.json:1-43` — plugin id (`"opentsdb"`), name (`"OpenTSDB"`),
 *   docs URL (`info.links[1].url` = `"https://grafana.com/docs/grafana/latest/datasources/opentsdb/"`),
 *   grafanaDependency `>=12.3.0-0`
 * - `src/components/ConfigEditor.tsx:10-24` — outer editor; composes
 *   `<DataSourceHttpSettings defaultUrl="http://localhost:4242" secureSocksDSProxyEnabled=...>`
 *   followed by `<OpenTsdbDetails value={options} onChange={onOptionsChange} />`
 * - `src/components/OpenTsdbDetails.tsx:8-66` — the plugin-specific settings
 *   panel: `<FieldSet label="OpenTSDB settings">` with three inputs — Version
 *   (Select), Resolution (Select), Lookup limit (Input type="number")
 * - `src/types.ts:35-39` — `OpenTsdbOptions extends DataSourceJsonData` with
 *   `tsdbVersion: number`, `tsdbResolution: number`, `lookupLimit: number`
 * - `src/datasource.ts:37-71` — `OpenTsDatasource` constructor consumes
 *   `instanceSettings.jsonData.tsdbVersion || 1`, `tsdbResolution || 1`,
 *   `lookupLimit || 1000`
 * - `pkg/opentsdb/opentsdb.go:24-69` — backend `NewDatasource` unmarshals
 *   `settings.JSONData` into a local `JSONData` struct
 *   (`TSDBVersion float32`, `TSDBResolution int32`, `LookupLimit int32`) and
 *   reads `settings.URL` directly; everything else is delegated to
 *   `settings.HTTPClientOptions(ctx)`
 * - `pkg/opentsdb/opentsdb.go:87-138` — `CheckHealth` issues
 *   `GET {url}/api/suggest?q=cpu&type=metrics` as the connectivity probe
 * - `pkg/opentsdb/utils.go:132-156` — `CreateRequest` posts to `{url}/api/query`
 *   and adds `?arrays=true` when `TSDBVersion == 4`
 * - `pkg/opentsdb/utils.go:225-311` — `ParseResponse` picks the array-shape
 *   response parser when `tsdbVersion == 4`, else the map-shape parser
 * - `pkg/opentsdb/callresource.go:361` — `HandleKeyValueLookup` passes
 *   `dsInfo.LookupLimit` as the `limit` query param to `/api/search/lookup`
 *
 * External components consulted at their pinned versions:
 * - `@grafana/ui@13.0.2` — `DataSourceHttpSettings` (URL input, Access help,
 *   Allowed cookies TagsInput, Timeout input, Basic auth switch, With
 *   Credentials switch), `BasicAuthSettings` (User + Password inputs),
 *   `HttpProxySettings` (TLS Client Auth, With CA Cert, Skip TLS Verify,
 *   Forward OAuth Identity switches), `TLSAuthSettings` + `CertificationKey`
 *   (ServerName, CA Cert, Client Cert, Client Key textareas),
 *   `CustomHeadersSettings` (dynamic httpHeaderName<N> / secureJsonData
 *   httpHeaderValue<N> — NOT modeled here), `SecureSocksProxySettings`
 *   (rendered when `config.secureSocksDSProxyEnabled` — excluded per
 *   AGENTS.md), plus `Field`, `FieldSet`, `Select`, `Input`
 * - `@grafana/data@13.0.2` — `DataSourceJsonData` base interface,
 *   `DataSourcePluginOptionsEditorProps`, `DataSourceSettings`,
 *   `SelectableValue`
 * - `@grafana/runtime@13.0.2` — `config.secureSocksDSProxyEnabled`
 */

/**
 * OpenTSDB "Version" selector — numeric enum. Stored as a JSON number under
 * `jsonData.tsdbVersion`. Options come from `OpenTsdbDetails.tsx:8-13`:
 *   1 → "<=2.1"  (default; `datasource.ts:60` fallback)
 *   2 → "==2.2"
 *   3 → "==2.3"
 *   4 → "==2.4"  (unlocks array-response parsing at `utils.go:138,247-254`)
 */
export type OpenTsdbVersion = 1 | 2 | 3 | 4;

/**
 * OpenTSDB "Resolution" selector — numeric enum. Stored as a JSON number under
 * `jsonData.tsdbResolution`. Options come from `OpenTsdbDetails.tsx:15-18`:
 *   1 → "second"      (default; `datasource.ts:61` fallback)
 *   2 → "millisecond" (sets `msResolution=true` on outgoing queries,
 *                      `datasource.ts:178-180`)
 */
export type OpenTsdbResolution = 1 | 2;

/**
 * Root (top-level datasource settings) fields the OpenTSDB editor writes.
 *
 * All root fields here are populated by @grafana/ui's `DataSourceHttpSettings`
 * component and consumed by the SDK's `settings.HTTPClientOptions(ctx)` call
 * in `pkg/opentsdb/opentsdb.go:29`. The plugin's own Go code reads only
 * `settings.URL` directly (`pkg/opentsdb/opentsdb.go:47`); the rest is honored
 * by the SDK's transport builder.
 */
export type RootConfig = {
  /** Complete HTTP URL of the OpenTSDB HTTP API. Default `http://localhost:4242`. */
  url?: string;
  /**
   * `'proxy'` (Server, default) or `'direct'` (Browser). OpenTSDB does not
   * pass `showAccessOptions` to `DataSourceHttpSettings`, so the editor
   * never renders an Access control; new datasources always end up with
   * `access: 'proxy'`. Legacy datasources may carry `access: 'direct'`.
   */
  access?: 'proxy' | 'direct';
  /** True when HTTP Basic authentication is enabled (DataSourceHttpSettings basic-auth switch). */
  basicAuth?: boolean;
  /** Basic-auth username. Only meaningful when `basicAuth === true`. */
  basicAuthUser?: string;
  /**
   * Cross-site access-control toggle rendered as "With Credentials" by
   * DataSourceHttpSettings. Independent from `basicAuth` — both can be true.
   */
  withCredentials?: boolean;
};

/**
 * Fields stored in `jsonData`. Combines the plugin's `OpenTsdbOptions`
 * (`src/types.ts:35-39`) with the TLS/HTTP fields written by @grafana/ui's
 * `DataSourceHttpSettings`, `HttpProxySettings`, and `TLSAuthSettings`.
 *
 * The plugin's Go backend only unmarshals the three OpenTSDB-specific fields
 * (`tsdbVersion`, `tsdbResolution`, `lookupLimit`) at
 * `pkg/opentsdb/opentsdb.go:65-69`. The TLS-related fields and cookie/timeout
 * knobs are read by the SDK's `HTTPClientOptions` when building the HTTP
 * client.
 */
export type JsonDataConfig = {
  /** Enable TLS client authentication (mTLS). Requires `serverName` + `tlsClientCert` + `tlsClientKey`. */
  tlsAuth?: boolean;
  /** Enable custom CA verification. Requires `tlsCACert`. */
  tlsAuthWithCACert?: boolean;
  /** Skip TLS certificate validation. Not recommended outside testing. */
  tlsSkipVerify?: boolean;
  /** TLS SNI / cert-verification server name (@grafana/ui TLSAuthSettings.tsx). */
  serverName?: string;
  /** HTTP request timeout in seconds (@grafana/ui DataSourceHttpSettings.tsx timeout Field). */
  timeout?: number;
  /** Cookies to forward to the datasource, by name (@grafana/ui DataSourceHttpSettings.tsx keepCookies TagsInput). */
  keepCookies?: string[];
  /**
   * Written by the "Forward OAuth Identity" switch inside `HttpProxySettings`.
   * When true, the SDK forwards the signed-in user's OAuth identity to
   * OpenTSDB.
   */
  oauthPassThru?: boolean;
  /**
   * OpenTSDB version selector — stored as a JSON number.
   * Default `1` (`datasource.ts:60` fallback; the editor's Select falls back
   * to `tsdbVersions[0]` visually at `OpenTsdbDetails.tsx:37` but does not
   * write it to storage until the user picks an option).
   */
  tsdbVersion?: OpenTsdbVersion;
  /**
   * OpenTSDB resolution selector — stored as a JSON number.
   * Default `1` (`datasource.ts:61` fallback).
   */
  tsdbResolution?: OpenTsdbResolution;
  /**
   * Row cap for `/api/search/lookup` responses when the query editor resolves
   * tag values. Default `1000` (`datasource.ts:62` fallback). Stored as a JSON
   * number, but see the upstream findings in README — the editor's
   * `onInputChangeHandler` writes `event.currentTarget.value` (a string), so
   * a datasource that has been edited once may hold a stringified number.
   */
  lookupLimit?: number;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`):
 * - `basicAuthPassword` — user password when Basic auth is enabled.
 * - `tlsCACert` — custom CA PEM when `tlsAuthWithCACert` is true.
 * - `tlsClientCert`, `tlsClientKey` — mTLS client credentials when `tlsAuth`
 *   is true.
 *
 * The editor also writes dynamic `httpHeaderValue<N>` secrets when the user
 * configures custom HTTP headers via @grafana/ui's `CustomHeadersSettings`
 * component. Those keys are indexed pairs — not modeled as first-class fields
 * in this schema; see the README.
 */
export type SecureJsonDataConfig = Array<
  'basicAuthPassword' | 'tlsCACert' | 'tlsClientCert' | 'tlsClientKey'
>;
