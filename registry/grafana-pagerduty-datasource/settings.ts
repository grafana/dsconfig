/**
 * Configuration models for the PagerDuty datasource plugin
 * (plugin id: `grafana-pagerduty-datasource`).
 *
 * PagerDuty is built on the shared "OpenAPI datasource" framework
 * (`src/openapids/`), so its configuration UI and storage shape are driven by a
 * bundled OpenAPI spec (`pkg/spec.json`) plus a customization file
 * (`pkg/customization.json`) rather than a hand-written ConfigEditor.
 *
 * Sources of truth (github.com/grafana/plugins-private @ 267f493, path
 * `plugins/grafana-pagerduty-datasource`):
 * - `src/plugin.json:5,4,24` — plugin id (`grafana-pagerduty-datasource`), name
 *   (`PagerDuty`), and docs link.
 * - `pkg/customization.json:4-12` — the security block: `supportsNoAuth:false`
 *   (`:5`) and the single `api_key` scheme with `apiKeyPrefix:"Token token="`
 *   (`:7-10`) and description "PagerDuty REST API Key (prefer generating
 *   read-only key)" (`:8`).
 * - `pkg/spec.json:3699-3706` — the `api_key` apiKey scheme: `name:"Authorization"`,
 *   `in:"header"`. `pkg/spec.json:164-169` — the single server
 *   `https://api.pagerduty.com` with no variables.
 * - `src/openapids/types.ts:4-22` — the generic frontend `Config` /
 *   `SecureConfig` types (jsonData: `servers?`, `auth?`, `enableSecureSocksProxy?`).
 * - `src/openapids/components/config-editor/Auth/Auth.tsx:51-64,108-129,190-199`
 *   — the auth section: `onAuthMethodChange` writes `jsonData.auth.id`,
 *   `onApiKeyChange` writes `secureJsonData["auth.<scheme>.apiKey"]`, and the
 *   mount effect auto-selects the only method (sets `auth.id`).
 * - `src/openapids/components/config-editor/Auth/ApiKey.tsx:16-31` — the API key
 *   field: `InlineField label="API key"`, tooltip from the scheme description,
 *   `SecretInput` (write-only).
 * - `src/openapids/components/config-editor/Connection.tsx:43-45` — the
 *   Connection section renders nothing when there is a single server with no
 *   variables (PagerDuty's case).
 * - `pkg/openapids/options.go:21-70` — backend `Options` (jsonData: `servers`,
 *   `auth.id`) and `loadOptionsFromPluginSettings`.
 * - `pkg/openapids/httpclient.go:42-48` — reads
 *   `DecryptedSecureJSONData["auth.<scheme>.apiKey"]`, prepends the
 *   `Token token=` prefix, and sets the `Authorization` header.
 *
 * External components consulted at their catalog-resolved versions
 * (`.yarnrc.yml` catalog; `package.json` uses `catalog:`):
 * - `@grafana/plugin-ui@^0.13.1` — `Auth` renders
 *   `<ConfigSection title="Authentication">`; `AuthMethodSettings` renders a
 *   `ConfigSubSection` whose title is the single method's label ("API key")
 *   when only one method is visible (no method-selector dropdown is shown).
 *   Read from github.com/grafana/plugin-ui
 *   (`src/components/ConfigEditor/Auth/Auth.tsx`,
 *   `src/components/ConfigEditor/Auth/auth-method/AuthMethodSettings.tsx`).
 * - `@grafana/ui@^11.6.7` — `SecretInput`, `InlineField` (no storage keys of
 *   their own). `@grafana/data@^11.6.7` — `DataSourceJsonData` base interface,
 *   `DataSourcePluginOptionsEditorProps`. `@grafana/runtime@^11.6.7` — feature
 *   toggles gating the (excluded) Secure Socks Proxy switch.
 *
 * The Secure Socks Proxy switch is deliberately excluded from this registry
 * entry (AGENTS.md exclusion for `jsonData.enableSecureSocksProxy`).
 */

/**
 * The selected OpenAPI security scheme id, stored in `jsonData.auth.id`.
 * PagerDuty declares exactly one scheme (`api_key`) and disables no-auth
 * (`pkg/customization.json:5-11`).
 */
export type PagerDutyAuthSchemeId = 'api_key';

/**
 * Root (top-level datasource settings) fields.
 *
 * PagerDuty stores nothing at the datasource root level. The backend
 * (`pkg/openapids/options.go`) reads only `jsonData` and the decrypted
 * `secureJsonData`; `settings.URL` and other root fields are unused (the base
 * URL comes from the OpenAPI spec, `pkg/openapids/httpclient.go:80-124`). So
 * this is a blank object rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. PagerDuty only meaningfully stores the auth
 * scheme id; the other members of the generic framework `Config`
 * (`src/openapids/types.ts:4-18`) are unused or excluded for this plugin.
 */
export type JsonDataConfig = {
  /**
   * The nested auth object. Only the scheme id is stored for the `api_key`
   * scheme; the per-scheme credential sub-objects
   * (`jsonData.auth.<scheme>.{username,clientId}`) are only written for the
   * basic / oauth2 schemes
   * (`src/openapids/components/config-editor/Auth/Auth.tsx:131-145`), which
   * PagerDuty does not use.
   */
  auth: {
    /**
     * Selected security scheme id (`jsonData.auth.id`). Auto-set by the editor
     * on mount to the only available method
     * (`src/openapids/components/config-editor/Auth/Auth.tsx:190-199`); read by
     * the backend to choose the scheme (`pkg/openapids/httpclient.go:34-35`).
     */
    id: PagerDutyAuthSchemeId;
  };
  /**
   * Generic framework field (`src/openapids/types.ts:5-8`,
   * `pkg/openapids/options.go:11-14,22`). UNUSED for PagerDuty: the spec has a
   * single server and no variables, so the editor renders no connection
   * section (`Connection.tsx:43-45`) and never writes `servers`; the base URL
   * always resolves to `https://api.pagerduty.com` (`pkg/spec.json:164-169`).
   * Not modeled in `dsconfig.json` or the Go `Config`.
   */
  servers?: {
    url: string;
    variables?: Record<string, string | number>;
  };
  /**
   * SDK-managed Secure Socks Proxy toggle
   * (`src/openapids/components/config-editor/EditorForm.tsx:98-111`), rendered
   * only when the `secureSocksDSProxyEnabled` feature toggle is on. Deliberately
   * excluded from this registry entry per AGENTS.md.
   */
  enableSecureSocksProxy?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing config
 * via `secureJsonFields`).
 *
 * The framework namespaces secrets by scheme id, so PagerDuty's single secret
 * is stored under the literal, dotted key `auth.api_key.apiKey`
 * (`src/openapids/components/config-editor/Auth/Auth.tsx:108-116`,
 * `pkg/openapids/httpclient.go:43`). It is the raw PagerDuty REST API key; the
 * backend prepends `Token token=` and sends it as the `Authorization` header.
 */
export type SecureJsonDataConfig = Array<'auth.api_key.apiKey'>;
