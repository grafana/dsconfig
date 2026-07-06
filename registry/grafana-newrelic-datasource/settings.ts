/**
 * Configuration models for the New Relic datasource plugin
 * (plugin id: `grafana-newrelic-datasource`).
 *
 * Sources of truth — grafana/plugins-private monorepo @ 267f4937806ed6404b6628d13ae358a5d308e376,
 * plugin path `plugins/grafana-newrelic-datasource/`:
 * - `src/plugin.json:3-4,24` — plugin name (`"New Relic"`), id
 *   (`"grafana-newrelic-datasource"`), docs URL
 *   (`info.links[0].url` = `"https://grafana.com/docs/plugins/grafana-newrelic-datasource"`).
 * - `src/components/ConfigEditor.tsx` — the config editor. A single
 *   `<h3>New Relic API Credentials</h3>` section (`:67`) with four fields:
 *   - Personal API Key / User API key (`:71-73` label + `tooltip="Used for NRQL queries"`,
 *     `:88` placeholder `"Personal API Key"`) → `secureJsonData.personalApiKey`
 *     via `onUpdateDatasourceSecureJsonDataOption(this.props, 'personalApiKey')` (`:90`),
 *     rendered as an `Input type="password"`, `required` (`:89`).
 *   - Account ID (`:111-113` label + `tooltip="Your New Relic Account ID"`,
 *     `:129` placeholder `"Account ID"`) → `secureJsonData.accountId`
 *     via `onUpdateDatasourceSecureJsonDataOption(this.props, 'accountId')` (`:131`),
 *     rendered as an `Input type="number"`, `required` (`:130`).
 *   - Region (`:152-153` label + `tooltip="Region hosting your service"`,
 *     `:161` placeholder `"default"`) → `jsonData.region` via `onRegionChanged`
 *     (`:42-52`); a `Select` whose options come from `regions` (`src/types.ts:187-190`).
 *   - Timeout in Seconds (`:169-170` label + tooltip/placeholder from
 *     `selectors.ts:22-28`) → `jsonData.timeoutInSeconds` (`:178-183`), an
 *     `Input type="number"` defaulting to 300 (`:57`, `|| 300` at `:181`).
 *   - `componentDidMount` (`:20-40`) migrates a legacy plaintext
 *     `jsonData.accountId` into `secureJsonData.accountId` and deletes
 *     `jsonData.accountId` / `jsonData.accountIdConfigured`.
 * - `src/components/selectors.ts:22-28` — the Timeout label tooltip
 *   (`"Enter the timeout in seconds. Defaults to 300"`) and placeholder (`"300"`).
 * - `src/types.ts:4-12,187-190` — the frontend config types
 *   `NewRelicSupportedRegion` (`'US' | 'EU'`), `NewRelicJsonData`
 *   (`region`, `timeoutInSeconds`), `NewRelicSecureJsonData`
 *   (`accountId`, `personalApiKey`), and the `regions` option list.
 * - `pkg/models/settings.go:12-49` — backend `Settings` struct and
 *   `LoadSettings`: parses `jsonData` (region, timeoutInSeconds,
 *   restBaseURL, infrastructureBaseURL, nerdGraphBaseURL), copies
 *   `personalApiKey` from `DecryptedSecureJSONData` (`:34`), reads and
 *   `strconv.Atoi`-parses `accountId` from `DecryptedSecureJSONData` into
 *   the numeric `AccountID` (`:36,42-46`), and defaults `TimeoutInSeconds`
 *   to 300 when `< 1` (`:38-40`).
 * - `pkg/datasource/newrelic_client.go:42-59` — builds the New Relic client:
 *   `ConfigPersonalAPIKey` (`:43`), `ConfigRegion` when region non-empty
 *   (`:47-49`), and `ConfigBaseURL` / `ConfigInfrastructureBaseURL` /
 *   `ConfigNerdGraphBaseURL` from the three backend-only URL overrides when
 *   non-empty (`:50-58`).
 * - `pkg/datasource/insights/insights_client.go:24-46` — the numeric
 *   `AccountID` is the NerdGraph `$accountId: Int!` NRQL query variable.
 * - `pkg/datasource/handler_checkhealth.go:35-36,139-145` — instance/health
 *   contract: non-empty `personalApiKey` (trimmed) and non-zero `AccountID`.
 * - `pkg/models/settings_test.go:11-68` — confirms jsonData parsing, secret
 *   copy, and the 300s timeout default.
 *
 * External components consulted at the versions pinned by the workspace
 * catalog (`.yarnrc.yml:14-26`, referenced via `catalog:` in the plugin's
 * `package.json:34-44`):
 * - `@grafana/ui@^11.6.7` — `InlineFormLabel`, `Input`, `Button`, `Select`
 *   (the config editor's field widgets).
 * - `@grafana/data@^11.6.7` — `DataSourceJsonData` (base interface of
 *   `NewRelicJsonData`), `DataSourcePluginOptionsEditorProps`,
 *   `onUpdateDatasourceSecureJsonDataOption`, `onUpdateDatasourceResetOption`
 *   (the storage-key update helpers the editor uses).
 *
 * The New Relic config editor renders no Secure Socks Proxy switch, so there
 * is nothing to exclude on that front (AGENTS.md `enableSecureSocksProxy`
 * exclusion is not applicable here).
 */

/**
 * New Relic data-center region. Stored under `jsonData.region`. Mirrors
 * `NewRelicSupportedRegion` (`src/types.ts:4`) and the `regions` option list
 * (`src/types.ts:187-190`). Passed to the New Relic client's `ConfigRegion`
 * only when non-empty (`pkg/datasource/newrelic_client.go:47-49`).
 */
export type NewRelicSupportedRegion = 'US' | 'EU';

/**
 * Root (top-level datasource settings) fields.
 *
 * The New Relic plugin stores every configuration value in `jsonData` /
 * `secureJsonData`; nothing lives at the root level. `LoadSettings`
 * (`pkg/models/settings.go:27-48`) reads only `config.JSONData` and
 * `config.DecryptedSecureJSONData`, and `GetNewRelicClient`
 * (`pkg/datasource/newrelic_client.go:30-59`) builds the client purely from
 * those parsed settings — `settings.URL`, basic auth, etc. are never read.
 * So `RootConfig` is a blank object.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Matches the backend `Settings` struct's
 * json-tagged fields (`pkg/models/settings.go:12-24`). The frontend
 * `NewRelicJsonData` (`src/types.ts:5-8`) only declares `region` and
 * `timeoutInSeconds`; the three base-URL overrides exist only in the backend
 * `Settings` struct.
 */
export type JsonDataConfig = {
  /**
   * New Relic data-center region (`'US' | 'EU'`). Editor `Select` with
   * placeholder `"default"` (`src/components/ConfigEditor.tsx:152-162`); no
   * value is written until a region is picked. Passed to `ConfigRegion` only
   * when non-empty (`pkg/datasource/newrelic_client.go:47-49`); an empty
   * region falls back to the New Relic client default (US).
   */
  region?: NewRelicSupportedRegion;
  /**
   * HTTP timeout in seconds. Editor default 300
   * (`src/components/ConfigEditor.tsx:57`), and the backend forces 300 when
   * the stored value is `< 1` (`pkg/models/settings.go:38-40`). Applied to
   * both the SDK HTTP client and the New Relic client
   * (`pkg/datasource/newrelic_client.go:35-44`).
   */
  timeoutInSeconds?: number;
  /**
   * Backend-only: override for the New Relic REST API base URL. "Used for
   * internal testing and mocking. not exposed in the UI"
   * (`pkg/models/settings.go:20-21`). Applied via `ConfigBaseURL` when
   * non-empty (`pkg/datasource/newrelic_client.go:50-52`).
   */
  restBaseURL?: string;
  /**
   * Backend-only: override for the New Relic Infrastructure API base URL.
   * Internal testing/mocking only (`pkg/models/settings.go:22`). Applied via
   * `ConfigInfrastructureBaseURL` when non-empty
   * (`pkg/datasource/newrelic_client.go:53-55`).
   */
  infrastructureBaseURL?: string;
  /**
   * Backend-only: override for the New Relic NerdGraph (GraphQL) API base URL.
   * Internal testing/mocking only (`pkg/models/settings.go:23`). Applied via
   * `ConfigNerdGraphBaseURL` when non-empty
   * (`pkg/datasource/newrelic_client.go:56-58`).
   */
  nerdGraphBaseURL?: string;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`):
 * - `personalApiKey` — New Relic user/personal API key. Authenticates every
 *   request via `ConfigPersonalAPIKey` (`pkg/datasource/newrelic_client.go:43`).
 *   Required: the backend rejects the instance with "Enter a personal API key."
 *   when empty/whitespace (`pkg/datasource/handler_checkhealth.go:35,139-141`).
 * - `accountId` — New Relic account ID, stored as a string but parsed to a
 *   numeric `AccountID` via `strconv.Atoi` (`pkg/models/settings.go:36,42-46`)
 *   and used as the NerdGraph `$accountId: Int!` NRQL variable
 *   (`pkg/datasource/insights/insights_client.go:24-46`). Required: the backend
 *   rejects the instance with "Enter an account ID. This must be a valid,
 *   positive number." when it does not parse to a non-zero integer
 *   (`pkg/datasource/handler_checkhealth.go:36,143-145`). Migrated out of the
 *   legacy `jsonData.accountId` by the editor
 *   (`src/components/ConfigEditor.tsx:20-40`).
 */
export type SecureJsonDataConfig = Array<'personalApiKey' | 'accountId'>;
