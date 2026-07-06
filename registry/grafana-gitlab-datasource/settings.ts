/**
 * Configuration models for the GitLab datasource plugin
 * (plugin id: `grafana-gitlab-datasource`).
 *
 * Sources of truth (github.com/grafana/plugins-private @ 267f4937, plugin path
 * `plugins/grafana-gitlab-datasource`):
 * - `src/plugin.json:4,5,24` — plugin name (`"GitLab"`), id
 *   (`"grafana-gitlab-datasource"`), docs link
 *   (`"https://grafana.com/docs/plugins/grafana-gitlab-datasource"`).
 * - `src/types.ts:35-42,63` — the frontend config types `GitLabDataSourceOptions`
 *   (jsonData) and `GitLabSecureJsonData` (secureJsonData), plus
 *   `DefaultURL = 'https://gitlab.com/api/v4'`.
 * - `src/views/ConfigEditor.tsx:13-217` — the configuration editor:
 *   - `<ConfigSection title="Connection">` (`:79`) with the URL `Input` ->
 *     root `options.url` (`onURLChange`, `:14-18`); placeholder
 *     `` `Default: ${DefaultURL}` `` (`:93`); tooltip `:82`.
 *   - `<Auth visibleMethods={['custom-gitlab']} ...>` (`:100-148`) from
 *     `@grafana/plugin-ui`, which renders a `ConfigSection title="Authentication"`
 *     containing the single custom method "Gitlab authentication"; its component
 *     is the Access token `SecretInput` -> secureJsonData.accessToken
 *     (`onAccessTokenChange`, `:21-29`; `onResetAccessToken`, `:41-54`).
 *   - `<ConfigSection title="Additional Settings" ...>` (`:155-158`) ->
 *     `ConfigSubSection title="Page limit"` -> Page limit `Input` ->
 *     jsonData.pageLimit (`onPageLimitChange`, `:31-34`); and the Secure Socks
 *     Proxy switch -> jsonData.enableSecureSocksProxy (`:36-39,184-211`, excluded here).
 * - `src/components/selectors.ts:3-14` — the E2E selector map (aria-labels only;
 *   the human labels/placeholders/tooltips are inline in ConfigEditor.tsx).
 * - `pkg/models/settings.go:16-63` — backend `Settings` struct (`URL`,
 *   `AccessToken`, `PageLimit`, `SdkClientOptions`) and `LoadSettings`.
 * - `pkg/gitlab/datasource.go:166-186` — `NewDatasource`: `gitlab.WithBaseURL(settings.URL)`
 *   and `gitlab.NewClient(settings.AccessToken, ...)` (go-gitlab v0.105.0, whose
 *   `NewClient` sends the token as the `PRIVATE-TOKEN` header — gitlab.go:855-858).
 * - `pkg/errors/errors.go:28` — `ErrorEmptyAccessToken` ("access token can not be blank").
 *
 * External components consulted at their pinned versions (plugin `package.json`
 * -> monorepo `.yarnrc.yml` catalog):
 * - `@grafana/plugin-ui@^0.13.1` — `Auth` (renders `ConfigSection title="Authentication"`
 *   + `AuthMethodSettings`; a single visible method drops the method `Select` and
 *   renders the custom component under a `ConfigSubSection` titled with the method
 *   label), `ConfigSection`, `ConfigSubSection`, `DataSourceDescription`.
 * - `@grafana/ui@^11.6.7` — `Input`, `SecretInput`, `InlineField`, `InlineSwitch`,
 *   `FieldValidationMessage`, `useTheme2`.
 * - `@grafana/data@^11.6.7` — `DataSourceJsonData` (the base interface
 *   `GitLabDataSourceOptions` extends), `DataSourcePluginOptionsEditorProps`,
 *   `FeatureToggles`.
 * - `@grafana/runtime@^11.6.7` — `config` (read to gate the Secure Socks Proxy switch).
 *
 * The Secure Socks Proxy switch (`jsonData.enableSecureSocksProxy`) is deliberately
 * excluded from this registry entry (AGENTS.md exclusion).
 */

/**
 * Root (top-level datasource settings) fields the GitLab plugin actually reads.
 *
 * `url` is the GitLab API base URL. The config editor writes it to the datasource
 * root `options.url` (`src/views/ConfigEditor.tsx:14-18`), NOT to `jsonData.url`,
 * and the backend reads `config.URL` (`pkg/models/settings.go:34`). Note that
 * `LoadSettings` first unmarshals `config.JSONData` into a struct whose `URL` is
 * tagged `json:"url"` (`pkg/models/settings.go:17,30`) but then immediately
 * overwrites it with `config.URL` (`:34`), so any `jsonData.url` is dead and the
 * value is purely the root url. When the root url is empty the backend defaults it
 * to `https://gitlab.com/api/v4` (`pkg/models/settings.go:23,35-37`).
 */
export type RootConfig = {
  /**
   * GitLab API base URL, e.g. `https://gitlab.com/api/v4` (the default) or, for a
   * self-hosted instance, `https://gitlab.example.com` (go-gitlab appends
   * `api/v4/` when missing). Required in the editor (`hasRequiredFields` +
   * `required`, `src/views/ConfigEditor.tsx:76,84`), but not strictly required by
   * the backend because an empty value defaults to `https://gitlab.com/api/v4`.
   */
  url?: string;
};

/**
 * Fields stored in `jsonData`. Matches the plugin's frontend
 * `GitLabDataSourceOptions` (`src/types.ts:35-38`, minus `enableSecureSocksProxy`)
 * and the json tag the backend `LoadSettings` reads from `config.JSONData`
 * (`pkg/models/settings.go:19,30`).
 */
export type JsonDataConfig = {
  /**
   * Maximum number of API pages a query fetches (`src/types.ts:36`). Editor input
   * writes `parseInt(value, 10)` (`src/views/ConfigEditor.tsx:31-34`); the backend
   * defaults `0` to `5` (`basePageLimit`, `pkg/models/settings.go:25,39-41`) and
   * passes it to every list handler (`pkg/gitlab/datasource.go:31` etc.).
   */
  pageLimit?: number;
  /**
   * Written by the Secure Socks Proxy switch (`src/views/ConfigEditor.tsx:36-39`,
   * shown only when `config.featureToggles.secureSocksDSProxyEnabled` and Grafana
   * >= 10.0.0, `:64-68,184`) and consumed transparently by the SDK's
   * `config.HTTPClientOptions(ctx)` call in `pkg/models/settings.go:53`. The
   * plugin's own Go code never inspects it by name. Deliberately excluded from the
   * dsconfig registry entry per AGENTS.md.
   */
  enableSecureSocksProxy?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config
 * via `secureJsonFields`).
 *
 * - `accessToken` — the GitLab personal (or project/group) access token
 *   (`src/types.ts:40-42`). Copied from `DecryptedSecureJSONData['accessToken']`
 *   in `LoadSettings` (`pkg/models/settings.go:47`) and sent by go-gitlab as the
 *   `PRIVATE-TOKEN` request header (`pkg/gitlab/datasource.go:176`). Required: an
 *   empty token makes `LoadSettings` fail with `ErrorEmptyAccessToken`
 *   (`pkg/models/settings.go:48-50`).
 */
export type SecureJsonDataConfig = Array<'accessToken'>;
