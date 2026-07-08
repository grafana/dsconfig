/**
 * Configuration models for the Amazon Managed Service for Prometheus
 * datasource plugin (`grafana-amazonprometheus-datasource`).
 *
 * Sources of truth (https://github.com/grafana/grafana-amazonprometheus-datasource
 * @ 34eb30afef47d8550382dd23b99deb81c32471a9):
 * - `src/plugin.json:2-8` — plugin ID (`"grafana-amazonprometheus-datasource"`),
 *   name (`"Amazon Managed Service for Prometheus"`), and `info.links[0].url`
 *   (`https://aws.amazon.com/prometheus/`).
 * - `src/configuration/ConfigEditor.tsx:17-107` — top-level editor: renders a
 *   `DataSourceDescription` with `docsLink` pointing at the plugin's public
 *   docs, an inline `Alert` rejecting browser-mode access, and a warning
 *   `Alert` when `jsonData['prometheus-type-migration']` is truthy. Composes
 *   `DataSourceHttpSettingsOverhaul` (URL + SigV4-locked `Auth` block +
 *   `sigv4Service` input) and a collapsible `ConfigSection` "Advanced
 *   settings" wrapping `AdvancedHttpSettings`, `AlertingSettingsOverhaul<DataSourceOptions>`,
 *   and `PromSettings` (with `hidePrometheusTypeVersion={true}`,
 *   `hideExemplars={true}`, `showQuerySamplesProcessedThresholdFields={true}`).
 * - `src/configuration/DataSourceHttpSettingsOverhaul.tsx:17-153` — the
 *   `Auth` wrapper with `visibleMethods=[sigV4Id]` (only SigV4 auth is
 *   selectable), `useEffectOnce` forces `jsonData.sigV4Auth = true` on every
 *   mount (`:27-38`), and `onAuthMethodSelect` (`:103-115`) always writes
 *   `basicAuth: false`, `withCredentials: false`, `jsonData.oauthPassThru: false`
 *   because SigV4 is the only visible method. Also renders the
 *   `forwardGrafanaUserHeader` InlineSwitch (`:122-143`).
 * - `src/configuration/ConfigEditor.tsx:52-83` — the SigV4 editor slot: wraps
 *   `SIGV4ConnectionConfig` from `@grafana/aws-sdk` and adds a manually
 *   rendered `Service` `Field`/`Input` for `jsonData.sigv4Service` (note the
 *   lowercase `v` — different from the `sigV4Auth` boolean and the other
 *   sigV4-prefixed fields). Placeholder and `defaultValue` are both `"aps"`.
 * - `src/configuration/DataSourceOptions.ts:1-8` — the plugin's
 *   `DataSourceOptions extends PromOptions` with three additions:
 *   `'prometheus-type-migration'?: boolean`, `sigV4Auth?: boolean`,
 *   `sigv4Service?: string`, `forwardGrafanaUserHeader?: boolean`.
 * - `pkg/datasource.go:24-131` — backend: `NewDatasource` reads
 *   `jsonData.forwardGrafanaUserHeader` via `promlib/utils.GetJsonData` +
 *   `maputil.GetBoolOptional`, builds a `promlib.Service` whose
 *   `extendClientOpts` (`:117-131`) reads `jsonData.sigv4Service` (falls back
 *   to `"aps"` when empty/missing) and writes it onto
 *   `clientOpts.SigV4.Service` for `awsauth.NewSigV4Middleware()` to sign
 *   against. All query/resource/health-check execution is delegated to
 *   `promlib`.
 *
 * The Prometheus knobs (`httpMethod`, `timeInterval`, `queryTimeout`,
 * `prometheusType`, `prometheusVersion`, `cacheLevel`, `incrementalQuerying`,
 * `incrementalQueryOverlapWindow`, `disableRecordingRules`,
 * `customQueryParameters`, `seriesLimit`, `seriesEndpoint`, `defaultEditor`,
 * `disableMetricsLookup`, `exemplarTraceIdDestinations`, `manageAlerts`,
 * `allowAsRecordingRulesTarget`, `timeout`, `keepCookies`, `oauthPassThru`,
 * `maxSamplesProcessedWarningThreshold`, `maxSamplesProcessedErrorThreshold`)
 * come from the shared `@grafana/prometheus` package (`PromOptions` in
 * `packages/grafana-prometheus/src/types.ts`) at version `13.1.6` — the pin
 * in this plugin's `package.json:80`.
 *
 * External components consulted at their pinned versions (from
 * `package.json` at the pinned SHA):
 * - `@grafana/aws-sdk@0.11.0` — `SIGV4ConnectionConfig`, `ConnectionConfig`,
 *   `AwsAuthType`, `AwsAuthDataSourceJsonData`,
 *   `AwsAuthDataSourceSecureJsonData`, `awsAuthProviderOptions`.
 *   `SIGV4ConnectionConfig` wraps `ConnectionConfig` with `skipHeader` and
 *   `skipEndpoint` and maps the standard `authType`/`profile`/`assumeRoleArn`/
 *   `externalId`/`defaultRegion`/`endpoint` fields onto their `sigV4`-prefixed
 *   counterparts in jsonData, plus `accessKey`/`secretKey` in secureJsonData.
 * - `@grafana/plugin-ui@0.16.0` — `Auth`, `AuthMethod`, `ConnectionSettings`,
 *   `convertLegacyAuthProps`, `AdvancedHttpSettings`, `DataSourceDescription`,
 *   `ConfigSection`, `ConfigSubSection`.
 * - `@grafana/prometheus@13.1.6` — `PromOptions`, `PromSettings`,
 *   `AlertingSettingsOverhaul`, `overhaulStyles`, `docsTip`.
 * - `@grafana/ui@13.0.2` — `Input`, `Select`, `Switch`, `TagsInput`, `Alert`,
 *   `Box`, `Field`, `InlineField`, `InlineSwitch`, `TextLink`,
 *   `SecureSocksProxySettings` (excluded), `useTheme2`.
 * - `@grafana/data@13.0.2` — `DataSourcePluginOptionsEditorProps`,
 *   `DataSourceSettings`, `SelectableValue`.
 * - `@grafana/runtime@13.0.2` — `config` (reads `config.secureSocksDSProxyEnabled`).
 *
 * The Secure Socks Proxy switch (rendered from
 * `DataSourceHttpSettingsOverhaul.tsx:145-150` when
 * `config.secureSocksDSProxyEnabled` is true) is deliberately excluded from
 * this registry entry per AGENTS.md.
 */

/**
 * AWS authentication type union (`AwsAuthType`).
 * Source: `@grafana/aws-sdk` `src/types.ts:3-13` (v0.11.0). Mirrored in
 * `github.com/grafana/grafana-aws-sdk` `pkg/awsds/settings.go`.
 * `arn` is deprecated and preserved for round-trip fidelity with datasources
 * provisioned before the value was renamed to `default`.
 */
export type SigV4AuthType =
  | 'default'
  | 'keys'
  | 'credentials'
  | 'ec2_iam_role'
  | 'grafana_assume_role'
  | 'arn';

/** Prometheus flavor type ("Prometheus" | "Cortex" | "Mimir" | "Thanos"). */
export type PromApplication = 'Prometheus' | 'Cortex' | 'Mimir' | 'Thanos';

/** Query editor mode ("builder" | "code"). */
export type QueryEditorMode = 'builder' | 'code';

/** Browser query cache level ("Low" | "Medium" | "High" | "None"). */
export type PrometheusCacheLevel = 'Low' | 'Medium' | 'High' | 'None';

/** HTTP method used for Prometheus range/instant queries. Defaults to "POST". */
export type PromHTTPMethod = 'POST' | 'GET';

/**
 * A single exemplar trace ID destination entry. When `datasourceUid` is set
 * the editor treats the exemplar as an internal link and takes precedence
 * over `url`. Amazon Prometheus does not render the exemplar editor UI
 * (`hideExemplars={true}`), but provisioning may still set the field and the
 * backend still emits exemplar links from `promlib` query results.
 */
export type ExemplarTraceIdDestination = {
  name: string;
  url?: string;
  urlDisplayLabel?: string;
  datasourceUid?: string;
};

/**
 * Root (top-level datasource settings) fields the Amazon Prometheus plugin
 * reads.
 *
 * Only `url` matters. Basic-auth and cross-site-credentials root fields are
 * not exposed by this plugin's editor — `visibleMethods=[sigV4Id]` locks the
 * auth picker to SigV4 auth, and `onAuthMethodSelect` clears `basicAuth` /
 * `withCredentials` on every save
 * (`src/configuration/DataSourceHttpSettingsOverhaul.tsx:103-115`).
 * `options.access === 'direct'` (Browser mode) is rejected with an inline
 * banner (`src/configuration/ConfigEditor.tsx:26-31`).
 */
export type RootConfig = {
  /** Complete HTTP URL of the AWS-hosted Prometheus workspace query endpoint. */
  url?: string;
};

/**
 * Fields stored in `jsonData`. Union of the plugin's own `DataSourceOptions`
 * (`extends PromOptions` plus `prometheus-type-migration`, `sigV4Auth`,
 * `sigv4Service`, `forwardGrafanaUserHeader`) plus the sigV4-prefixed fields
 * `SIGV4ConnectionConfig` writes and every field the `promlib` backend
 * parses.
 */
export type JsonDataConfig = {
  /**
   * Enabling flag for SigV4 signing. Forced to `true` on every editor mount
   * via `DataSourceHttpSettingsOverhaul.tsx:27-38` because Amazon Prometheus
   * hides every non-SigV4 auth method (`visibleMethods=[sigV4Id]`). Consumed
   * by the SDK's shared HTTP client to install the SigV4 middleware.
   */
  sigV4Auth?: boolean;

  /**
   * AWS credentials chain to use for signing. Written by
   * `@grafana/aws-sdk`'s `SIGV4ConnectionConfig` — the underlying
   * `ConnectionConfig` writes to `authType` and the SigV4 wrapper renames it
   * to `sigV4AuthType` on the way out (`SIGV4ConnectionConfig.tsx:20-27`).
   */
  sigV4AuthType?: SigV4AuthType;

  /** Credentials profile name for `sigV4AuthType === 'credentials'`. */
  sigV4Profile?: string;

  /** Optional STS role ARN the selected provider should assume. */
  sigV4AssumeRoleArn?: string;

  /**
   * Optional STS external ID for cross-account assume-role. Hidden when
   * `sigV4AuthType === 'grafana_assume_role'` because Grafana Cloud injects
   * the external ID automatically (`ConnectionConfig.tsx:274`).
   */
  sigV4ExternalId?: string;

  /**
   * Default AWS region to sign against. Amazon Prometheus does not surface an
   * `endpoint` override in the editor (`skipEndpoint`), so the region alone
   * (plus the workspace URL) determines the signing target.
   */
  sigV4Region?: string;

  /**
   * AWS service namespace to sign requests against. NOTE: lowercase `v` —
   * `sigv4Service`, not `sigV4Service`. Placeholder and default are both
   * `"aps"` (Amazon Managed Prometheus). Read by `pkg/datasource.go:117-131`
   * via `promlib/utils.GetJsonData` + `maputil.GetStringOptional`; the
   * backend falls back to `"aps"` when the field is empty or missing.
   */
  sigv4Service?: string;

  /**
   * Forward the logged-in Grafana user's `X-Grafana-User` header to the
   * workspace. Requires `send_user_header` to be enabled server-side.
   * Consumed by `pkg/datasource.go:95-99`.
   */
  forwardGrafanaUserHeader?: boolean;

  /**
   * Sentinel flag set to `true` when a vanilla Prometheus data source is
   * migrated to Amazon Prometheus. Triggers the migration banner at
   * `ConfigEditor.tsx:37-48`. Storage key uses a hyphen, hence the string
   * key here.
   */
  'prometheus-type-migration'?: boolean;

  /**
   * Cleared to `false` on every save by
   * `DataSourceHttpSettingsOverhaul.onAuthMethodSelect` because
   * `visibleMethods=[sigV4Id]` never selects `AuthMethod.OAuthForward`.
   * Consumed by the SDK's shared HTTP client and by `pkg/promlib`.
   */
  oauthPassThru?: boolean;

  /** Manage alert rules for this data source (`AlertingSettingsOverhaul`). */
  manageAlerts?: boolean;
  /** Allow this datasource as a target for writing recording rules. */
  allowAsRecordingRulesTarget?: boolean;

  /** HTTP request timeout in seconds (`AdvancedHttpSettings`). */
  timeout?: number;
  /** Cookies to forward to the datasource, by name (`AdvancedHttpSettings`). */
  keepCookies?: string[];

  /** Scrape interval like `'15s'` (`PromSettings.tsx`). */
  timeInterval?: string;
  /** Prometheus query timeout like `'60s'`. */
  queryTimeout?: string;
  /** Default editor when opening a query — Builder or Code. */
  defaultEditor?: QueryEditorMode;
  /** Disable metrics chooser + metric/label autocomplete. */
  disableMetricsLookup?: boolean;

  /**
   * Prometheus flavor. Not rendered by this plugin's editor
   * (`hidePrometheusTypeVersion={true}` at `ConfigEditor.tsx:100`) but
   * parseable via provisioning and consumed by the backend `promlib`
   * heuristics.
   */
  prometheusType?: PromApplication;
  /** Free-form version string; editor-hidden alongside `prometheusType`. */
  prometheusVersion?: string;

  /** Browser cache level for editor queries. */
  cacheLevel?: PrometheusCacheLevel;
  /** Turn on incremental query caching (beta). */
  incrementalQuerying?: boolean;
  /** Duration string; defaults to `'10m'` when `incrementalQuerying` is true. */
  incrementalQueryOverlapWindow?: string;
  /** Disable recording rules (beta). */
  disableRecordingRules?: boolean;
  /** URL query-parameter string appended to Prometheus requests. */
  customQueryParameters?: string;
  /** POST (default) or GET. Uppercased on save by `promlib` backend. */
  httpMethod?: PromHTTPMethod;
  /** Series/label endpoint limit; empty = 40000, 0 = no limit. */
  seriesLimit?: number;

  /**
   * Query warning threshold. Visible in this plugin's editor because
   * `ConfigEditor.tsx:102` passes `showQuerySamplesProcessedThresholdFields={true}`
   * to `PromSettings` — unlike vanilla Prometheus and Azure Prometheus, which
   * hide it.
   */
  maxSamplesProcessedWarningThreshold?: number;
  /** Query error threshold. See `maxSamplesProcessedWarningThreshold`. */
  maxSamplesProcessedErrorThreshold?: number;

  /** Prefer /api/v1/series over /api/v1/label/*&#47;values. */
  seriesEndpoint?: boolean;

  /**
   * Exemplar trace-ID destinations. Not rendered by this plugin's editor
   * (`hideExemplars={true}` at `ConfigEditor.tsx:101`). Provisioning-only —
   * the backend still parses and returns exemplar links from `promlib`
   * query results.
   */
  exemplarTraceIdDestinations?: ExemplarTraceIdDestination[];
};

/**
 * Secure key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`):
 *
 * - `sigV4AccessKey` — AWS access key ID, set when
 *   `sigV4AuthType === 'keys'`.
 * - `sigV4SecretKey` — AWS secret access key, set when
 *   `sigV4AuthType === 'keys'`.
 *
 * Both secrets are written by `@grafana/aws-sdk`'s `SIGV4ConnectionConfig`
 * (`SIGV4ConnectionConfig.tsx:28-35`) which maps the underlying
 * `ConnectionConfig`'s `accessKey`/`secretKey` onto their sigV4-prefixed
 * forms.
 */
export type SecureJsonDataConfig = Array<'sigV4AccessKey' | 'sigV4SecretKey'>;
