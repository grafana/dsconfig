/**
 * Configuration models for the Grafana Infinity datasource plugin
 * (plugin id: `yesoreyeram-infinity-datasource`).
 *
 * Sources of truth (https://github.com/grafana/grafana-infinity-datasource @ 3aede2f):
 * - `src/plugin.json:1-103` — plugin id (`"yesoreyeram-infinity-datasource"`),
 *   name (`"Infinity"`), docs URL (`info.links[0].url`
 *   = `"https://grafana.com/docs/plugins/yesoreyeram-infinity-datasource"`),
 *   `backend: true`, `grafanaDependency: ">=11.6.0-0"`.
 * - `src/editors/config.editor.tsx:141-150` — the tab list (`Main`,
 *   `Authentication`, `URL, Headers & Params`, `Network`, `Security`,
 *   `Health check`, `Reference data`, `Global queries`).
 * - `src/editors/config/Auth.tsx:12-23,68-105` — auth method options and the
 *   `onAuthTypeChange` handler that writes `root.basicAuth`,
 *   `jsonData.oauthPassThru`, and `jsonData.auth_method` together.
 * - `src/editors/config/Auth.tsx:145-274` — every auth-method field: labels
 *   (`User Name`, `Password`, `Bearer token`, `Key`, `Value`, `Add to`,
 *   `Region`, `Service`, `Access Key`, `Secret Key`), placeholders, and the
 *   secureJsonData keys each SecretFormField writes to.
 * - `src/editors/config/Auth.AzureBlob.tsx:8-70` — Azure Blob storage account
 *   name / key inputs and `azureBlobCloudType` combobox.
 * - `src/editors/config/OAuthInput.tsx:9-234` — OAuth2 grant-type selector
 *   (`client_credentials` / `jwt` / `others`), auth style radio (0/1/2), and
 *   the JWT/client-credentials input geometry.
 * - `src/editors/config/URL.tsx:8-27` — Base URL editor; empty input is
 *   written as `'__IGNORE_URL__'` (`src/constants.ts:72`).
 * - `src/editors/config/URL.tsx:30-62` — URL settings switches:
 *   `ignoreStatusCodeCheck`, `allowDangerousHTTPMethods`,
 *   `pathEncodedUrlsEnabled` (Experimental).
 * - `src/editors/config/TLSConfigEditor.tsx:11-125` — TLS toggles and PEM
 *   textareas (`tlsSkipVerify`, `tlsAuthWithCACert`, `tlsAuth`, `serverName`,
 *   `tlsCACert`, `tlsClientCert`, `tlsClientKey`).
 * - `src/editors/config/ProxyEditor.tsx:19-141` — Proxy Mode radio
 *   (`env` / `none` / `url`), URL/User Name/Password inputs, and the
 *   feature-gated Secure Socks Proxy switch (excluded here per AGENTS.md).
 * - `src/editors/config/AllowedHosts.tsx:9-26` — Allowed hosts list; hidden
 *   when `jsonData.auth_method === 'azureBlob'`.
 * - `src/editors/config/SecurityConfigEditor.tsx:8-31` — Query security
 *   radio (`allow` / `warn` / `deny`) writing `jsonData.unsecuredQueryHandling`.
 * - `src/editors/config/CustomHealthCheckEditor.tsx:6-32` — Enable custom
 *   health check switch + Health check URL input.
 * - `src/editors/config/KeepCookies.tsx:8-24` — `keepCookies` `TagsInput`.
 * - `src/editors/config/ReferenceData.tsx:1-60` — `refData` list of
 *   `{ name, data }` objects.
 * - `src/editors/config/GlobalQueryEditor.tsx:1-80` — `global_queries` list
 *   of `{ name, id, query }` objects. The nested `query` is the full
 *   InfinityQuery object (per-query editor state); this config schema keeps
 *   it opaque at the datasource-config level.
 * - `src/components/config/SecureFieldsEditor.tsx:75-113` — the shared
 *   indexed-pair writer used by Custom HTTP Headers
 *   (`httpHeaderName<N>` / `httpHeaderValue<N>`), URL Query Params
 *   (`secureQueryName<N>` / `secureQueryValue<N>`), OAuth2 endpoint params
 *   (`oauth2EndPointParamsName<N>` / `oauth2EndPointParamsValue<N>`) and
 *   OAuth2 token headers (`oauth2TokenHeadersName<N>` /
 *   `oauth2TokenHeadersValue<N>`).
 * - `src/types/config.types.ts:1-86` — the frontend types
 *   (`InfinityOptions`, `InfinitySecureOptions`, `AuthType`, `OAuth2Type`,
 *   `APIKeyType`, `ProxyType`, `UnsecureQueryHandling`, `AzureBlobCloudType`,
 *   `OAuth2Props`, `AWSAuthProps`, `InfinityReferenceData`,
 *   `GlobalInfinityQuery`).
 * - `pkg/models/settings.go:17-133,261-425` — backend `InfinitySettings`
 *   (flattened runtime shape), `InfinitySettingsJson` (the persisted jsonData
 *   shape), and `LoadSettings` (parses jsonData, back-fills legacy
 *   basicAuth/oauthPassThru into `auth_method`, defaults `timeoutInSeconds`
 *   to 60, `proxy_type` to 'env', `unsecuredQueryHandling` to 'warn',
 *   `apiKeyType` to 'header', `azureBlobCloudType` to 'AzureCloud',
 *   normalizes PEM secrets, aggregates `httpHeader*`, `secureQuery*`,
 *   `oauth2EndPointParams*`, and `oauth2TokenHeaders*` indexed pairs into
 *   maps).
 *
 * External components consulted at their pinned versions
 * (`package.json`): `@grafana/ui@13.0.1`, `@grafana/data@13.0.1`,
 * `@grafana/runtime@13.0.1`. The plugin does NOT depend on
 * `@grafana/plugin-ui` or `@grafana/experimental`; the config editor
 * composes @grafana/ui primitives directly plus in-tree components under
 * `src/components/config/`.
 *
 * The Secure Socks Proxy field
 * (`jsonData.enableSecureSocksProxy`) is written by
 * `src/editors/config/ProxyEditor.tsx:109-137` when the
 * `secureSocksDSProxyEnabled` feature toggle is on. It is excluded from
 * this schema per AGENTS.md.
 */

/** Discriminator for the selected authentication method — jsonData.auth_method. */
export type AuthType =
  | 'none'
  | 'basicAuth'
  | 'bearerToken'
  | 'apiKey'
  | 'digestAuth'
  | 'oauthPassThru'
  | 'oauth2'
  | 'aws'
  | 'azureBlob';

/** OAuth2 grant type — jsonData.oauth2.oauth2_type. */
export type OAuth2Type = 'client_credentials' | 'jwt' | 'others';

/** Where the API key value is sent — jsonData.apiKeyType. */
export type APIKeyType = 'header' | 'query';

/** Outbound proxy mode — jsonData.proxy_type. */
export type ProxyType = 'none' | 'env' | 'url';

/**
 * Query security discriminator — jsonData.unsecuredQueryHandling. Controls
 * how the backend treats queries that carry per-query secrets (e.g. inline
 * headers) that bypass the allowed-hosts protection.
 */
export type UnsecureQueryHandling = 'allow' | 'warn' | 'deny';

/** Azure Blob cloud discriminator — jsonData.azureBlobCloudType. */
export type AzureBlobCloudType = 'AzureCloud' | 'AzureUSGovernment' | 'AzureChinaCloud';

/**
 * AWS SigV4 sub-object stored at jsonData.aws (`OAuth2Props`-style,
 * `src/types/config.types.ts:25-29`). Only `authType: 'keys'` exists today.
 */
export type AWSAuthProps = {
  authType?: 'keys';
  region?: string;
  service?: string;
};

/**
 * OAuth2 sub-object stored at jsonData.oauth2
 * (`src/types/config.types.ts:13-24`). The `authStyle` field encodes the
 * `oauth2.AuthStyle` enum: 0=Auto, 1=In Params, 2=In Header.
 */
export type OAuth2Props = {
  oauth2_type?: OAuth2Type;
  client_id?: string;
  email?: string;
  private_key_id?: string;
  subject?: string;
  token_url?: string;
  scopes?: string[];
  authStyle?: number;
  authHeader?: string;
  tokenTemplate?: string;
};

/** Reference-data list entry stored at jsonData.refData. */
export type InfinityReferenceData = { name: string; data: string };

/**
 * Global-query list entry stored at jsonData.global_queries. `query` is
 * the full InfinityQuery object from the query editor; at the
 * datasource-config layer it is intentionally opaque (`unknown`), because
 * modeling the entire query editor jsonData is outside the datasource
 * config schema's scope.
 */
export type GlobalInfinityQuery = {
  name: string;
  id: string;
  query: unknown;
};

/**
 * Root (top-level datasource settings) fields the Grafana Infinity plugin
 * uses.
 *
 * `url` is the Base URL prefixed to every query URL. When left blank in the
 * editor the frontend writes the sentinel `'__IGNORE_URL__'`
 * (`src/constants.ts:72`); `pkg/models/settings.go:296-298` normalizes it
 * back to `''` on the backend, so a provisioning payload can safely use
 * either.
 *
 * `basicAuth` and `basicAuthUser` are populated by
 * `src/editors/config/Auth.tsx:70-88`'s `onAuthTypeChange` /
 * `onUserNameChange` handlers whenever the selected auth method is
 * `'basicAuth'` or `'digestAuth'`, and are consumed by the SDK's
 * `HTTPClientOptions` (via `settings.HTTPClientOptions(ctx)` in
 * `pkg/models/settings.go:419-423`). The Infinity plugin's own code also
 * mirrors them onto `InfinitySettings.BasicAuthEnabled` and `.UserName`.
 */
export type RootConfig = {
  /** Base URL prefixed to every query URL. Empty in editor → stored as `'__IGNORE_URL__'`. */
  url?: string;
  /** True when the auth method is `basicAuth`. Written by `Auth.tsx:71`. */
  basicAuth?: boolean;
  /** Basic-auth / digest-auth username. Only meaningful when `basicAuth === true`. */
  basicAuthUser?: string;
};

/**
 * Fields stored in `jsonData`. Mirrors the persisted
 * `InfinitySettingsJson` struct in `pkg/models/settings.go:261-290` plus the
 * two nested objects (`oauth2`, `aws`), the `refData` / `global_queries` /
 * `allowedHosts` / `keepCookies` arrays, and the dynamic indexed-pair keys
 * written by `SecureFieldsEditor`.
 *
 * The plugin does not use TypeScript index signatures for the dynamic pairs
 * (`httpHeaderName<N>`, `secureQueryName<N>`,
 * `oauth2EndPointParamsName<N>`, `oauth2TokenHeadersName<N>`), so they are
 * documented as a comment rather than represented in the type.
 */
export type JsonDataConfig = {
  /** Authentication method discriminator. Default `'none'`. */
  auth_method?: AuthType;

  // ─── API key ─────────────────────────────────────────────────────────
  /** API key name (header or query-param name). */
  apiKeyKey?: string;
  /** Where the API key value is sent. Default `'header'`. */
  apiKeyType?: APIKeyType;

  // ─── OAuth2 ──────────────────────────────────────────────────────────
  /** OAuth2 sub-object. `oauth2_type` defaults to `'client_credentials'`. */
  oauth2?: OAuth2Props;

  // ─── AWS SigV4 ───────────────────────────────────────────────────────
  /** AWS SigV4 sub-object. `authType` is currently always `'keys'`. */
  aws?: AWSAuthProps;

  // ─── Forward OAuth ───────────────────────────────────────────────────
  /**
   * Written together with `root.basicAuth` by `Auth.tsx:74`. `true` when
   * `auth_method` is `'oauthPassThru'`.
   */
  oauthPassThru?: boolean;

  // ─── Azure Blob ──────────────────────────────────────────────────────
  azureBlobCloudType?: AzureBlobCloudType;
  azureBlobAccountName?: string;
  /**
   * Backend-only: filled in by `LoadSettings` from `azureBlobCloudType`
   * (`pkg/models/settings.go:406-415`), not written by the editor.
   */
  azureBlobAccountUrl?: string;

  // ─── TLS ─────────────────────────────────────────────────────────────
  tlsSkipVerify?: boolean;
  tlsAuth?: boolean;
  tlsAuthWithCACert?: boolean;
  /** SNI / cert verification server name. Only meaningful when `tlsAuth === true`. */
  serverName?: string;

  // ─── Network ─────────────────────────────────────────────────────────
  /** HTTP request timeout in seconds. Backend clamps to `[0, 300]`; default 60. */
  timeoutInSeconds?: number;
  /** Proxy mode. Default `'env'` (uses `HTTP_PROXY` / `HTTPS_PROXY`). */
  proxy_type?: ProxyType;
  proxy_url?: string;
  proxy_username?: string;

  // ─── Security ────────────────────────────────────────────────────────
  /** Hosts the datasource is allowed to query when `root.url` is blank. */
  allowedHosts?: string[];
  /**
   * How to handle per-query secrets that bypass the allowed-hosts
   * protection. Default `'warn'`.
   */
  unsecuredQueryHandling?: UnsecureQueryHandling;

  // ─── URL / request behavior ──────────────────────────────────────────
  ignoreStatusCodeCheck?: boolean;
  allowDangerousHTTPMethods?: boolean;
  pathEncodedUrlsEnabled?: boolean;
  keepCookies?: string[];

  // ─── Custom health check ─────────────────────────────────────────────
  customHealthCheckEnabled?: boolean;
  customHealthCheckUrl?: string;

  // ─── Reference data & global queries ─────────────────────────────────
  refData?: InfinityReferenceData[];
  global_queries?: GlobalInfinityQuery[];

  // ─── Backend-only ────────────────────────────────────────────────────
  /**
   * Backend-only test flag: when `true` the plugin swaps in the in-memory
   * mock client. Not exposed in the editor
   * (`pkg/models/settings.go:262`).
   */
  is_mock?: boolean;

  // ─── Dynamic indexed-pair keys ───────────────────────────────────────
  // The following keys are written dynamically by
  // `src/components/config/SecureFieldsEditor.tsx:98-113` and are NOT
  // represented as first-class fields in this type:
  //
  //   - httpHeaderName<N>          (paired with secureJsonData.httpHeaderValue<N>)
  //   - secureQueryName<N>         (paired with secureJsonData.secureQueryValue<N>)
  //   - oauth2EndPointParamsName<N> (paired with secureJsonData.oauth2EndPointParamsValue<N>)
  //   - oauth2TokenHeadersName<N>  (paired with secureJsonData.oauth2TokenHeadersValue<N>)
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`):
 *
 * - `basicAuthPassword` — Basic / Digest auth password
 * - `bearerToken` — Bearer token auth
 * - `apiKeyValue` — API-key value paired with jsonData.apiKeyKey
 * - `awsAccessKey`, `awsSecretKey` — AWS SigV4 credentials
 * - `oauth2ClientSecret` — OAuth2 client-credentials grant
 * - `oauth2JWTPrivateKey` — OAuth2 JWT grant private key PEM
 * - `azureBlobAccountKey` — Azure Blob storage account key
 * - `tlsCACert`, `tlsClientCert`, `tlsClientKey` — TLS PEMs
 * - `proxyUserPassword` — Custom proxy password
 *
 * In addition, `SecureFieldsEditor` writes dynamic indexed-pair secrets
 * whose names are NOT enumerated as static string literals here:
 * `httpHeaderValue<N>`, `secureQueryValue<N>`,
 * `oauth2EndPointParamsValue<N>`, `oauth2TokenHeadersValue<N>`. Consumers
 * enumerate them by inspecting `secureJsonFields` on the datasource
 * instance.
 */
export type SecureJsonDataConfig = Array<
  | 'basicAuthPassword'
  | 'bearerToken'
  | 'apiKeyValue'
  | 'awsAccessKey'
  | 'awsSecretKey'
  | 'oauth2ClientSecret'
  | 'oauth2JWTPrivateKey'
  | 'azureBlobAccountKey'
  | 'tlsCACert'
  | 'tlsClientCert'
  | 'tlsClientKey'
  | 'proxyUserPassword'
>;
