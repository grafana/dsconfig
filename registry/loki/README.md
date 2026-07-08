# loki

Declarative configuration schema for the [Loki datasource plugin](https://github.com/grafana/grafana-loki-datasource) (`loki`).

## Upstream researched

- **Repo**: `github.com/grafana/grafana-loki-datasource`
- **Ref**: `main`
- **Commit SHA**: `882588ba81944057b3eefa8f2fa55e76156352d7` (Merge pull request #113: `docs: add signed commits requirement to CONTRIBUTING.md`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders,
tooltips, option labels/values, section titles, help markdown, defaults,
validations, dependency and required-when expressions, storage keys, storage
targets, value types, group titles, and instructions — is traceable to a
specific `file:line` in the upstream repo at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone https://github.com/grafana/grafana-loki-datasource
cd grafana-loki-datasource
git checkout 882588ba81944057b3eefa8f2fa55e76156352d7
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under
[Sources researched](#sources-researched) before merging any changes to this
entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, effects, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: root URL/basicAuth/basicAuthUser/withCredentials tagged `json:"-"`, jsonData fields, `DecryptedSecureJSONData`), `PluginID`, `DerivedFieldMatcherType` / `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth/TLS/feature variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the
shared [`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA (`882588b`), plus
external editor components at the exact versions the plugin's `package.json`
pins.

### Plugin repo (`github.com/grafana/grafana-loki-datasource@882588b`)

| File | What was read |
| --- | --- |
| `src/plugin.json:1-58` | `pluginType` (`id` = `"loki"`), `pluginName` (`name` = `"Loki"`), docs URL (`info.links[3].url` = `"https://grafana.com/docs/grafana/latest/datasources/loki/"`) |
| `src/configuration/ConfigEditor.tsx:37-80` | Top-level editor — composes `DataSourceDescription`, `ConnectionSettings` (with `urlPlaceholder="http://localhost:3100"`), `Auth` (via `convertLegacyAuthProps`), a collapsible "Additional settings" section containing `AdvancedHttpSettings`, the conditional `SecureSocksProxySettings`, `AlertingSettings`, `QuerySettings`, and `DerivedFields` |
| `src/configuration/AlertingSettings.tsx:12-38` | `ConfigSubSection title="Alerting"`, `InlineField label="Manage alert rules in Alerting UI"`, tooltip at `:26`, `InlineSwitch` bound to `jsonData.manageAlerts` |
| `src/configuration/QuerySettings.tsx:14-46` | `ConfigSubSection title="Queries"`, `InlineField label="Maximum lines"`, tooltip at `:28-34`, `Input type="number" placeholder="1000"` bound to `jsonData.maxLines` (stored as string) |
| `src/configuration/DerivedFields.tsx:46-125` | `ConfigSubSection title="Derived fields"`, description at `:51`, per-item render via `DerivedField`, add-button default `{ name: '', matcherRegex: '', urlDisplayLabel: '', url: '', matcherType: 'regex' }` at `:93-99` |
| `src/configuration/DerivedField.tsx:55-232` | Per-item fields — `Name` (`Field label="Name" placeholder="Field name"`, `:85-87`, duplicate-name validation at `DerivedFields.tsx:39-44`), `Type` (Select `Regex in log line`/`Label`, tooltip at `:92-94`), `Regex`/`Label` (`:117-131`, tooltips at `:120-122` and `:126`), `URL`/`Query` (label toggles with internal link, `:146-158`, placeholder `${__value.raw}` when internal / `http://example.com/${__value.raw}` when external), `URL Label` (`:159-169`, tooltip at `:163-165`), `Internal link` toggle (`:173-186`), `Data source` picker (`:190-202`), `Open in new tab` toggle (`:206-220`) |
| `src/types.ts:36-64` | Frontend `LokiOptions extends DataSourceJsonData` with `maxLines?: string`, `derivedFields?: DerivedFieldConfig[]`, `alertmanager?: string`, `keepCookies?: string[]`; `DerivedFieldConfig` shape |
| `src/datasource.ts:110-168,397` | Frontend consumption — `DEFAULT_MAX_LINES = 1000`, constructor parses `settingsData.maxLines ?? '0'` into an int limit, `derivedFields` fed into the result transformer |
| `pkg/loki/loki.go:48-72` | `NewDatasource` — the only jsonData/root read is `settings.URL`; everything else is delegated to `settings.HTTPClientOptions(ctx)` (SDK reads basicAuth/TLS/custom headers/cookies from the same instance settings) |
| `pkg/loki/loki.go:116-170` | `callResource` builds `/loki/api/v1/<url>` off the base URL — confirms Loki hard-requires the URL to be set |
| `package.json` | External component versions (see next table) |

Notably absent: no `pkg/models/settings.go`, no upstream `LoadSettings` — the
Loki plugin does not own a backend jsonData settings model. All server-side
reads of settings go through the SDK.

### External editor components

Read at the exact versions pinned in the plugin's `package.json`
(`@grafana/plugin-ui: ^0.13.1`, `@grafana/ui/runtime/data: ^12.4.0`). Sources
checked out at the corresponding upstream commits.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `ConnectionSettings` | `@grafana/plugin-ui@0.13.1` | `github.com/grafana/plugin-ui` @ `4d2f196` (release: v0.13.1, `defaultProps` migration), `src/components/ConfigEditor/Connection/ConnectionSettings.tsx:17-75` | URL label defaults to `"URL"`, placeholder passed by plugin (`ConfigEditor.tsx:47` — `urlPlaceholder="http://localhost:3100"`); required + built-in URL regex validation |
| `Auth`, `AuthMethodSettings`, `BasicAuth` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/Auth/{Auth,auth-method/AuthMethodSettings,auth-method/BasicAuth}.tsx` | Default `visibleMethods = [BasicAuth, OAuthForward, NoAuth]` at `AuthMethodSettings.tsx:57-66`; option labels/descriptions from `AuthMethodSettings.tsx:9-32`; BasicAuth `User`/`Password` labels + placeholders + tooltips from `BasicAuth.tsx:24-29` |
| `convertLegacyAuthProps` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/Auth/utils.ts:8-55` | Maps `basicAuth` / `withCredentials` / `jsonData.oauthPassThru` ↔ AuthMethod enum; `onAuthMethodSelect` writes basicAuth+withCredentials+oauthPassThru in one shot (`:44-54`) |
| `TLSSettings`, `SelfSignedCertificate`, `TLSClientAuth`, `SkipTLSVerification` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/Auth/tls/*.tsx` | All TLS labels/placeholders/rows come verbatim from these files (see [Field provenance](#field-provenance)); note the RSA private key placeholder typo `Begins with --- RSA PRIVATE KEY CERTIFICATE ---` at `TLSClientAuth.tsx:109` |
| `CustomHeaders`, `CustomHeader` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/Auth/custom-headers/*.tsx` | Indexed `httpHeaderName<N>` / `httpHeaderValue<N>` storage pattern; **not modeled** in this schema (see [Modeling decisions](#modeling-decisions)) |
| `AdvancedHttpSettings` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/AdvancedSettings/AdvancedHttpSettings.tsx:44-82` | `Allowed cookies` and `Timeout` labels/tooltips/placeholders |
| `DataSourceDescription`, `ConfigSection`, `ConfigSubSection`, `ConfigDescriptionLink` | `@grafana/plugin-ui@0.13.1` | `src/components/ConfigEditor/DataSourceDescription.tsx`, `ConfigSection.tsx` | Intro text prop shape; section title/description props (no storage keys — layout only) |
| `SecureSocksProxySettings` (excluded) | `@grafana/ui@^12.4.0` | grafana/grafana `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | Storage key it writes (`jsonData.enableSecureSocksProxy`) — confirmed and excluded |
| `DataSourcePicker` | `@grafana/runtime@^12.4.0` | grafana/grafana `packages/grafana-runtime/src/components/DataSourcePicker.tsx` | Used by `DerivedField.tsx:190-202` for the internal-link tracing data source; writes `datasourceUid` (a UID string) to the item |
| `DataLinkInput`, `Input`, `Switch`, `Select`, `SecretInput`, `SecretTextArea`, `TagsInput`, `Field`, `InlineField`, `InlineSwitch`, `Button` | `@grafana/ui@^12.4.0` | grafana/grafana `packages/grafana-ui/src/components/` | Prop names (`label`, `placeholder`, `value`, `onChange`, `isConfigured`, `onReset`, `rows`, `width`) — needed to know which UI attributes to record |
| `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps`, `VariableOrigin`, `DataLinkBuiltInVars` | `@grafana/data@^12.4.0` | grafana/grafana `packages/grafana-data/src/` | Base jsonData interface (`manageAlerts` is defined here, not on `LokiOptions`); variable suggestion metadata for `DataLinkInput` |

Note: `@grafana/plugin-ui` published `v0.13.1` as an npm tag but did not push a
git tag; commit `4d2f196` on `main` corresponds to the changelog entry
"v0.13.1 - 2026-02-10 - Replace defaultProps with es6 defaults for React 19
compatibility" and is what npm resolved when the plugin's package-lock was
generated.

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line`
where each of its label, placeholder, tooltip, default, storage key, and value
type is defined. Where a field draws from multiple lines, all lines are listed.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | `ConnectionSettings.tsx:41` (default `urlLabel = 'URL'`; Loki editor does not override) | `ConfigEditor.tsx:47` (`urlPlaceholder="http://localhost:3100"`) | `settings.URL string` — SDK base | Required per `loki.go:66,119` (`requiredWhen: "true"`) |
| `virtual_authMethod` | — | virtual | `AuthMethodSettings.tsx:145` (`<Field label="Authentication method">`) | Options from `AuthMethodSettings.tsx:9-32`; default `'NoAuth'` mirrors `getSelectedMethod` fallthrough at `utils.ts:37` for a fresh datasource | Union of 3 strings | `storage.computed.read` mirrors `getSelectedMethod` (`utils.ts:27-38`) minus `CrossSiteCredentials`, which the Loki editor doesn't expose; `effects` mirror `onAuthMethodSelect` (`utils.ts:44-54`) |
| `root_basicAuth` | `basicAuth` | `root` | — (no UI; managed by `virtual_authMethod`) | Written by `utils.ts:47` | Root SDK bool | Tagged `managed-by:virtual_authMethod` |
| `root_basicAuthUser` | `basicAuthUser` | `root` | `BasicAuth.tsx:24` (default `userLabel = 'User'`) | `BasicAuth.tsx:26` (default `userPlaceholder = 'User'`); tooltip `BasicAuth.tsx:25` | SDK `settings.BasicAuthUser string` | `dependsOn: virtual_authMethod == 'BasicAuth'`; `requiredWhen: root_basicAuth == true` |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | `BasicAuth.tsx:27` (default `passwordLabel = 'Password'`) | `BasicAuth.tsx:29` (default `passwordPlaceholder = 'Password'`); tooltip `BasicAuth.tsx:28` | Role `auth.basic.password` | Same conditional/required as `basicAuthUser` |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (no visible UI; controlled by `virtual_authMethod == 'OAuthForward'`) | Written by `utils.ts:51` | `bool` (@grafana/plugin-ui writes it) | Tagged `managed-by:virtual_authMethod` |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | `SelfSignedCertificate.tsx:33` (`label="Add self-signed certificate"`) | `tooltipText` `SelfSignedCertificate.tsx:34`; default `false` | `bool` (SDK TLS pack) | — |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | `SelfSignedCertificate.tsx:39` (`label="CA Certificate"`) | `SelfSignedCertificate.tsx:54` (`placeholder="Begins with --- BEGIN CERTIFICATE ---"`); `rows: 6` `SelfSignedCertificate.tsx:55` | Role `tls.caCert` | `dependsOn` / `requiredWhen`: `jsonData_tlsAuthWithCACert == true` |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | `TLSClientAuth.tsx:45` (`label="TLS Client Authentication"`) | `tooltipText` `TLSClientAuth.tsx:46` | `bool` | — |
| `jsonData_serverName` | `serverName` | `jsonData` | `TLSClientAuth.tsx:51` (`label="ServerName"`) | `TLSClientAuth.tsx:63` (`placeholder="domain.example.com"`); tooltip `TLSClientAuth.tsx:53` | Role `tls.serverName` | `dependsOn: jsonData_tlsAuth == true`; required for the mTLS contract |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | `TLSClientAuth.tsx:70` (`label="Client Certificate"`) | `TLSClientAuth.tsx:88` (`placeholder="Begins with --- BEGIN CERTIFICATE ---"`); `rows: 6` `TLSClientAuth.tsx:89` | Role `tls.clientCert` | `dependsOn`/`requiredWhen`: `jsonData_tlsAuth == true` |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | `TLSClientAuth.tsx:94` (`label="Client Key"`) | `TLSClientAuth.tsx:109` (`` placeholder=`Begins with --- RSA PRIVATE KEY CERTIFICATE ---` `` — upstream typo preserved); `rows: 6` `TLSClientAuth.tsx:110` | Role `tls.clientKey` | Same conditional/required as `tlsClientCert`; see [Upstream findings](#upstream-findings) #2 for the placeholder typo |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | `SkipTLSVerification.tsx:14` (`label="Skip TLS certificate validation"`) | `tooltipText` `SkipTLSVerification.tsx:15` | Role `transport.tlsSkipVerify` | Default `false` |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | `AdvancedHttpSettings.tsx:48` (`label="Allowed cookies"`) | `AdvancedHttpSettings.tsx:56` (`placeholder="New cookie (hit enter to add)"`); tooltip `AdvancedHttpSettings.tsx:50` | `string[]` | — |
| `jsonData_timeout` | `timeout` | `jsonData` | `AdvancedHttpSettings.tsx:63` (`label="Timeout"`) | `AdvancedHttpSettings.tsx:74` (`placeholder="Timeout in seconds"`); tooltip `AdvancedHttpSettings.tsx:66` | `number` (int, parsed at `AdvancedHttpSettings.tsx:33`) | Role `transport.timeoutSeconds` |
| `jsonData_manageAlerts` | `manageAlerts` | `jsonData` | `AlertingSettings.tsx:24` (`label="Manage alert rules in Alerting UI"`) | Tooltip `AlertingSettings.tsx:26`; read default `config.defaultDatasourceManageAlertsUiToggle` (`AlertingSettings.tsx:29`) — not written on load | `bool`, base `DataSourceJsonData` | Loki-specific label — differs from Prometheus's "Manage alerts via Alerting UI" |
| `jsonData_maxLines` | `maxLines` | `jsonData` | `QuerySettings.tsx:25` (`label="Maximum lines"`) | `QuerySettings.tsx:42` (`placeholder="1000"`); tooltip `QuerySettings.tsx:28-34` | `string`, `types.ts:37` | Stored as a string despite being numeric; the frontend parses it via `parseInt(... ?? '0', 10) \|\| DEFAULT_MAX_LINES` (`datasource.ts:168`) |
| `jsonData_derivedFields` | `derivedFields` | `jsonData` | `DerivedFields.tsx:48` (`title="Derived fields"`) | Item defaults from `DerivedFields.tsx:93-99` (`{ name: '', matcherRegex: '', urlDisplayLabel: '', url: '', matcherType: 'regex' }`); item fields from `DerivedField.tsx:55-232` | `DerivedFieldConfig[]`, `types.ts:56-64` | Item has 7 fields (name+matcherType+matcherRegex required, url/urlDisplayLabel/datasourceUid/targetBlank optional); URL vs Query label toggles with `datasourceUid` |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | URL | Yes (direct: `pkg/loki/loki.go:66`) |
| `virtual_authMethod` | — (virtual) | — | Authentication method | — (editor-local selector) |
| `root_basicAuth` | `basicAuth` | `root` | — (managed by virtual) | Yes (SDK via `HTTPClientOptions`) |
| `root_basicAuthUser` | `basicAuthUser` | `root` | User | Yes (SDK) |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | Password | Yes (SDK) |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (managed by virtual) | Yes (SDK) |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | Add self-signed certificate | Yes (SDK) |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | CA Certificate | Yes (SDK) |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | TLS Client Authentication | Yes (SDK) |
| `jsonData_serverName` | `serverName` | `jsonData` | ServerName | Yes (SDK) |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | Client Certificate | Yes (SDK) |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | Client Key | Yes (SDK) |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | Skip TLS certificate validation | Yes (SDK) |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | Allowed cookies | Yes (SDK) |
| `jsonData_timeout` | `timeout` | `jsonData` | Timeout | Yes (SDK) |
| `jsonData_manageAlerts` | `manageAlerts` | `jsonData` | Manage alert rules in Alerting UI | Consumed by Grafana core Alerting, not by the Loki plugin's Go code |
| `jsonData_maxLines` | `maxLines` | `jsonData` | Maximum lines | No — frontend-only; parsed at `datasource.ts:168` |
| `jsonData_derivedFields` | `derivedFields` | `jsonData` | Derived fields | No — frontend-only; consumed by `transformBackendResult` (`datasource.ts:397`) |

### Frontend-only settings

- **`jsonData.maxLines`** — parsed by `datasource.ts:168` into a JavaScript
  number used to cap each Loki query's `MaxLines`. The Loki backend never
  reads `settings.JSONData.maxLines`; per-query limits come from the query
  model itself.
- **`jsonData.derivedFields`** — applied by the frontend result transformer
  (`datasource.ts:397`, `transformBackendResult`) to inject data links onto
  log rows. The Loki backend never reads it.

### Backend-only settings

None — the Loki plugin's Go code only reads `settings.URL` directly, and
otherwise delegates to the SDK's `HTTPClientOptions`. There is no upstream
`pkg/models/settings.go` unmarshaling jsonData server-side.

### Excluded settings

- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy` and associated
  socks-proxy fields) — rendered conditionally at `ConfigEditor.tsx:64-66`
  when `config.secureSocksDSProxyEnabled` is set on the Grafana instance.
  Deliberately omitted per `AGENTS.md`.
- **Custom HTTP headers** (`@grafana/plugin-ui`'s `CustomHeaders`) — the
  editor writes indexed pairs `jsonData.httpHeaderName<N>` /
  `secureJsonData.httpHeaderValue<N>` starting at index 1. Not modeled as a
  first-class field because the storage keys are dynamic. Downstream tools
  should walk `jsonData` for the `httpHeaderName` prefix and pair up matching
  `httpHeaderValue<N>` secrets; the SDK's `HTTPClientOptions` already does
  this and forwards the resulting headers to Loki.
- **`jsonData.alertmanager`** — declared on the frontend `LokiOptions` type
  (`src/types.ts:39`) but never rendered by the editor and never read by the
  Loki datasource. See [Upstream findings](#upstream-findings) #1 — kept out
  of the schema because it is dead storage, and retained in the TypeScript
  `JsonDataConfig` type only for round-trip compatibility.

## Where the types are defined

The configuration types are spread across the plugin and its dependencies —
some fields and base types come from libraries/SDKs rather than the plugin
itself.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `LokiOptions`, `DerivedFieldConfig` | `src/types.ts:36-64` | plugin ([grafana/grafana-loki-datasource](https://github.com/grafana/grafana-loki-datasource)) |
| `DEFAULT_MAX_LINES`, `DEFAULT_MAX_LINES_SAMPLE` | `src/datasource.ts:110-111` | plugin |
| `DataSourceJsonData` (base interface: `authType`, `defaultRegion`, `profile`, `manageAlerts`, `allowAsRecordingRulesTarget`, `alertmanagerUid`, `disableGrafanaCache`) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^12.4.0` |
| `DataSourcePluginOptionsEditorProps`, `DataSourceSettings`, `VariableOrigin`, `DataLinkBuiltInVars` | `packages/grafana-data/src/` | `@grafana/data` `^12.4.0` |
| `ConnectionSettings`, `Auth`, `AuthMethod`, `convertLegacyAuthProps`, `AdvancedHttpSettings`, `DataSourceDescription`, `ConfigSection`, `ConfigSubSection`, `ConfigDescriptionLink` | `src/components/ConfigEditor/…` | `@grafana/plugin-ui` `0.13.1` (grafana/plugin-ui @ `4d2f196`) |
| `SecureSocksProxySettings` (excluded), `DataLinkInput`, `Input`, `Switch`, `SecretInput`, `SecretTextArea`, `Select`, `TagsInput`, `Field`, `InlineField`, `InlineSwitch`, `Button` | `packages/grafana-ui/src/components/` | `@grafana/ui` `^12.4.0` |
| `DataSourcePicker`, `config` (`defaultDatasourceManageAlertsUiToggle`) | `packages/grafana-runtime/src/` | `@grafana/runtime` `^12.4.0` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `NewDatasource`, `datasourceInfo` (reads `settings.URL`), `queryData` / `callResource` / `CheckHealth` (all take `dsInfo` by reference — no jsonData unmarshal anywhere) | `pkg/loki/loki.go:36-285` | plugin |
| `backend.DataSourceInstanceSettings` (carries `URL`, `BasicAuthEnabled`, `BasicAuthUser`, `JSONData`, `DecryptedSecureJSONData`), `HTTPClientOptions(ctx)` | `backend/common.go`, `backend/httpclient/` | `github.com/grafana/grafana-plugin-sdk-go` |

The models in this entry flatten the above into a single Go `Config` type
(root `URL`/`BasicAuth`/`BasicAuthUser`/`WithCredentials` tagged `json:"-"`,
plus the jsonData fields the editor writes and the SDK reads, plus
`DecryptedSecureJSONData`) and a `SecureJsonDataKey` typed constant list.
`settings.ts` keeps the three canonical TypeScript types (`RootConfig`,
`JsonDataConfig`, `SecureJsonDataConfig`).

## Modeling decisions

- **Virtual auth method**: `convertLegacyAuthProps`'s `onAuthMethodSelect`
  (`@grafana/plugin-ui utils.ts:44-54`) writes three storage fields in one
  shot — `root.basicAuth`, `root.withCredentials`, and
  `jsonData.oauthPassThru`. That is the same virtual-selector pattern used by
  the Prometheus entry. `withCredentials` is not in the Loki editor's default
  `visibleMethods` (`AuthMethodSettings.tsx:57-66`), so the virtual field's
  effects only write `basicAuth` and `oauthPassThru`. If a provisioning
  payload writes `withCredentials=true` directly, the SDK still honors it —
  the virtual's `storage.computed.read` doesn't preserve that state, but the
  underlying root storage does.
- **`requiredWhen` on `basicAuthUser` / `basicAuthPassword`**: keyed on the
  underlying storage field (`root_basicAuth == true`), not the virtual
  selector. The virtual is an editor-local convenience; the backend contract
  is "if basicAuth is on, we need a username and password".
- **TLS pair requirements**: `TLSClientAuth` and `SelfSignedCertificate` mark
  every field with `required` in the UI, but they only require the paired
  fields when the parent switch is on. Encoded as `dependsOn` + `requiredWhen`
  on each field.
- **`maxLines` typed as string**: the storage type is `string` because
  `QuerySettings.tsx` binds an HTML `<Input type="number">` with
  `value={maxLines}` back to the raw string form via
  `event.currentTarget.value` (`:40`). The frontend converts to a number at
  read time. Persisting as a string keeps round-trip parity with the editor.
- **Derived fields modeled as an object array with 7 item fields**: each
  field maps 1:1 to a control in `DerivedField.tsx`. `matcherType` gets a
  `defaultValue: "regex"` because the "Add" button (`DerivedFields.tsx:93-99`)
  writes that on new entries, and `matcherType` may also be absent on
  legacy entries (the frontend defaults to `'regex'` at
  `DerivedField.tsx:61`) — the schema validation accepts both.
- **No help drawer**: Loki's editor has no top-level `Collapse`/`help` panel
  (the sub-sections each have their own `ConfigDescriptionLink` at the top,
  pointing back to the datasource docs), so there is no schema `help` object.
  The detailed guidance is captured in `description` on individual fields and
  in the `instructions` block.
- **Field ID naming convention**: IDs are prefixed with their storage target
  for easy discoverability — `root_`, `jsonData_`, or `secureJsonData_` (and
  `virtual_` for virtual fields, which have no storage target) — followed by
  the camelCase storage key. The `key` property keeps the plugin's raw
  storage key.
- **Custom HTTP headers and Secure Socks Proxy excluded**: see
  [Excluded settings](#excluded-settings) above.
- **Flat `Config` in Go**: `settings.go` collapses jsonData fields and
  decrypted secrets onto a single `Config` struct. Base `DataSourceJsonData`
  fields (authType, defaultRegion, profile, alertmanagerUid,
  disableGrafanaCache) exist in Grafana core but are neither written by the
  Loki editor nor read by the Loki plugin, so they are omitted. Root-level
  fields the editor and SDK both use (`URL`, `BasicAuth`, `BasicAuthUser`,
  `WithCredentials`) are carried with `json:"-"` tags so `LoadConfig` returns
  them alongside the jsonData shape.
- **`ApplyDefaults` is a no-op**: unlike Prometheus (which defaults
  `httpMethod` to POST), the Loki editor writes nothing into jsonData on
  load. Every visual default (Maximum lines placeholder `"1000"`, Manage
  alerts fallback to `config.defaultDatasourceManageAlertsUiToggle`,
  new-derived-field `matcherType: "regex"`) is a render-time `??` fallback,
  not persisted state. `ApplyDefaults` intentionally does nothing so we
  don't clobber intentional zero values on a stored datasource — the
  `TestApplyDefaults` test guards this.
- **`SecureJsonDataConfig` is a key list**: secure values are write-only, so
  the secure type is just the array of secret key names (`basicAuthPassword`,
  `tlsCACert`, `tlsClientCert`, `tlsClientKey`); consumers read
  `secureJsonFields` to see what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema`
bundle (the k8s-style schema Grafana's datasource API server serves as
`{apiVersion}.json`, `v0alpha1` today) from the embedded `dsconfig.json`:
root fields plus a nested `jsonData` object become the OpenAPI settings
`spec`, secure fields become `secureValues`, and virtual fields are skipped.

`SettingsExamples()` provides the default configuration plus one k8s-style
example per authentication method, TLS variant, and Loki-specific feature.
Each example is a full instance-settings object with the plugin configuration
nested under `jsonData` and the relevant write-only secrets under
`secureJsonData` (placeholder values to be replaced with real secrets; the
default example — keyed by the empty string `""` — carries an empty
`basicAuthPassword` to show that no secret is required for the default
No-auth mode):

| Example | Auth | TLS | Extras | `secureJsonData` |
| --- | --- | --- | --- | --- |
| `""` (default) | None | — | — | `basicAuthPassword` (empty) |
| `noAuth` | None | — | `maxLines=5000`, `manageAlerts=true` | `basicAuthPassword` (empty) |
| `basicAuth` | Basic | — | — | `basicAuthPassword` |
| `oauthForward` | OAuth Identity | — | — | `basicAuthPassword` (empty) |
| `tlsMutualAuth` | None | mTLS (serverName + client cert/key) | — | `tlsClientCert`, `tlsClientKey` |
| `tlsSelfSignedCA` | None | Custom CA | — | `tlsCACert` |
| `withDerivedFields` | Basic | — | `derivedFields` (regex + internal link, regex + external link), `maxLines=1500` | `basicAuthPassword` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and
returns a fully-defaulted, validated `Config`:

1. **Parse** — copy `settings.URL`, `settings.BasicAuthEnabled`,
   `settings.BasicAuthUser` into `Config`, unmarshal `settings.JSONData` into
   the jsonData portion of the same struct, and copy the four decrypted
   secrets into `DecryptedSecureJSONData`. The Loki plugin has no upstream
   `LoadSettings` to mirror — `pkg/loki/loki.go:48-72` is the only server-side
   read of settings and it just uses `settings.URL` + `settings.HTTPClientOptions`.
2. **`ApplyDefaults`** — intentionally a no-op (see [Modeling decisions](#modeling-decisions)).
3. **`Validate`** — enforce the runtime contract: URL is required
   (Loki fails at request-time otherwise, `loki.go:119`), Basic auth requires
   a username, mTLS requires serverName + client cert + client key, custom-CA
   requires the CA PEM, `timeout` must be non-negative, and each configured
   `derivedField` must have a non-empty name + matcher regex and a valid
   `matcherType`. Errors are joined so every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with
`datasource_uid`, `datasource_name`, and `plugin` labels so log lines carry
request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are still exported
for callers that want to compose them themselves (e.g. provisioning preview,
schema-example round-trip, tests that need to distinguish parse-level from
policy-level errors). Skip them by never calling `LoadConfig` in those flows —
assemble a `Config` directly.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while
researching upstream. All preserved verbatim in the schema — the schema
records what the plugin **does**, not what it **should** do; these notes
exist so reviewers can reproduce each finding and decide separately whether
to fix upstream.

1. **Dead `alertmanager` field**: `src/types.ts:39` declares
   `alertmanager?: string` on `LokiOptions`, but nothing in `src/**` or
   `pkg/**` writes or reads it. It is dead storage — likely a leftover from
   an earlier design where Loki datasources could point to a specific
   Alertmanager. The schema does not include it as an editor field; the
   TypeScript `JsonDataConfig` keeps it only for round-trip parity, and the
   Go `Config` intentionally omits it (jsonData with `alertmanager` would
   round-trip through `LoadConfig` and be silently dropped on marshal).
2. **Upstream typo preserved**: `TLSClientAuth.tsx:109` sets the client key
   placeholder to `` `Begins with --- RSA PRIVATE KEY CERTIFICATE ---` `` —
   an RSA private key is not a "certificate". Preserved verbatim in
   `secureJsonData_tlsClientKey.ui.placeholder`. This is a plugin-ui typo
   shared across all data sources that use `Auth`.
3. **`maxLines` typed as string despite being numeric**: the editor renders
   `Input type="number"` (`QuerySettings.tsx:37`) but binds the raw
   `event.currentTarget.value` (`:40`), so an empty box saves as `""` and any
   value round-trips as a JavaScript string. The frontend parses it at read
   time with `parseInt(... ?? '0', 10) || DEFAULT_MAX_LINES` (`datasource.ts:168`),
   which means an unparseable value silently falls back to 1000 without any
   error.
4. **`manageAlerts` reads a global default but never persists it**:
   `AlertingSettings.tsx:29` renders the switch as
   `options.jsonData.manageAlerts ?? config.defaultDatasourceManageAlertsUiToggle`.
   Loading an untouched datasource shows the global toggle's state, but
   `jsonData.manageAlerts` is only written when the user flips the switch
   at least once. A datasource that has always followed the global default
   will have `manageAlerts` undefined in storage even if the UI showed it
   as "on".
5. **No connection health check for URL**: the Loki backend does not
   pre-validate `settings.URL`; requests just fail when `callResource`
   builds `/loki/api/v1/<url>` (`loki.go:119`). We surface this as a
   `requiredWhen: "true"` constraint on `root_url` so provisioning tooling
   can reject an empty URL upfront.
6. **`derivedFields` per-item `matcherType` may be absent**: `DerivedFields.tsx:93-99`
   writes `matcherType: 'regex'` when adding a new entry, but legacy entries
   (pre-`matcherType`, before the label matcher landed) may lack the field.
   `DerivedField.tsx:61` defaults to `'regex'` at render time. Our
   validation accepts empty `matcherType` for parity.
7. **Base `DataSourceJsonData` fields are mostly unused**: Grafana core
   embeds `DataSourceJsonData` into any jsonData shape and it carries
   `authType`, `defaultRegion`, `profile`, `manageAlerts`,
   `allowAsRecordingRulesTarget`, `alertmanagerUid`, `disableGrafanaCache`.
   The Loki editor only writes `manageAlerts`; the rest are neither shown
   nor consumed. `allowAsRecordingRulesTarget` in particular does not apply
   to Loki (it is a Prometheus/Mimir alerting concept). The unused base
   fields are omitted from the schema.
8. **`SecureSocksProxySettings` also writes to jsonData**: the excluded
   Secure Socks Proxy widget writes `jsonData.enableSecureSocksProxy` and
   related fields. Provisioning payloads that include those keys will not
   round-trip through this schema — they will be preserved in the raw
   `JSONData` but not be represented in `Config` or `SettingsExamples`.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in
  this repo) — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft 2020-12, `additionalProperties: false`) — passes.
- `go test ./...` on this entry — passes (schema bundle shape, secure values,
  examples, `LoadConfig` incl. TLS variants and malformed input,
  `SchemaArtifactInSync` guard, `JSONDataMatchesStruct`,
  `JSONDataTypesMatchStruct`).
- `settings.go`/`schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: exports the three canonical types (`RootConfig`,
  `JsonDataConfig`, `SecureJsonDataConfig`) — reviewed by hand against the
  frontend sources; no `tsc` runner is wired into the registry module.
