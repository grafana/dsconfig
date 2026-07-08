# graphite

Declarative configuration schema for the [Graphite datasource plugin](https://github.com/grafana/grafana-graphite-datasource) (`graphite`).

## Upstream researched

- **Repo**: `github.com/grafana/grafana-graphite-datasource`
- **Ref**: `main`
- **Commit SHA**: `baa731870c5c1a5d8e47bd32e7134329ea3e2f04`

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders,
tooltips, option labels/values, section titles, help markdown, defaults,
validations, dependency and required-when expressions, storage keys, storage
targets, value types, group titles, and instructions — is traceable to a
specific `file:line` in the upstream repo at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone https://github.com/grafana/grafana-graphite-datasource
cd grafana-graphite-datasource
git checkout baa731870c5c1a5d8e47bd32e7134329ea3e2f04
```

If upstream `main` has advanced past this SHA, re-diff the sources listed
under [Sources researched](#sources-researched) before merging any changes to
this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: root URL/basicAuth/basicAuthUser/withCredentials tagged `json:"-"`, jsonData fields, `DecryptedSecureJSONData`), `PluginID`, `GraphiteVersion` / `GraphiteType` / `SecureJsonDataKey` typed constants, nested `GraphiteQueryImportConfiguration` struct, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth/TLS/feature variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the
shared [`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA (`baa7318`), plus
external editor components at the exact versions the plugin's `package.json`
pins.

### Plugin repo (`github.com/grafana/grafana-graphite-datasource@baa7318`)

| File | What was read |
| --- | --- |
| `src/plugin.json:1-49` | `pluginType` (`id` = `"graphite"`), `pluginName` (`name` = `"Graphite"`), docs URL (`info.links[3].url` = `"https://grafana.com/docs/grafana/latest/datasources/graphite/"`), `grafanaDependency: >=12.3.0-0` |
| `src/configuration/ConfigEditor.tsx:1-141` | Outer editor — composes `<Alert>` deprecation notice when `options.access === 'direct'` (`:54-58`), `<DataSourceHttpSettings defaultUrl="http://localhost:8080" ...>` (`:60-65`), a `<FieldSet>` "Graphite details" (`:66-107`) with Version (`:68-80`), Graphite backend type (`:82-94`), conditional Rollup indicator when `graphiteType === GraphiteType.Metrictank` (`:95-106`), then `<MappingsConfiguration>` (`:108-133`); `componentDidMount` writes `DEFAULT_GRAPHITE_VERSION` on load (`:43-45`) |
| `src/configuration/MappingsConfiguration.tsx:19-77` | `<h3 className="page-heading">Label mappings</h3>` (`:21`), per-row `<InlineField label="Mapping (${i + 1})"><Input placeholder="e.g. test.metric.(labelName).*"/></InlineField>` (`:33-48`), add button "Add label mapping" (`:65-74`) |
| `src/configuration/MappingsHelp.tsx:10-72` | Help drawer markdown — verbatim in the `help.markdown` of `jsonData_importConfiguration` |
| `src/configuration/parseLokiLabelMappings.ts:1-31` | `fromString` / `toString` — round-trip between the editor's string form (`servers.(cluster).*`) and the persisted `{ matchers: [{value, labelName?}] }` shape |
| `src/types.ts:7-71` | `GraphiteOptions extends DataSourceJsonData` with `graphiteVersion: string`, `graphiteType: GraphiteType`, `rollupIndicatorEnabled?: boolean`, `importConfiguration: GraphiteQueryImportConfiguration`; `GraphiteType` enum (`Default = 'default'`, `Metrictank = 'metrictank'`); `GraphiteQueryImportConfiguration`, `GraphiteLokiMapping`, `GraphiteMetricLokiMatcher` |
| `src/versions.ts:1-5` | `GRAPHITE_VERSIONS = ['0.9', '1.0', '1.1']`; `DEFAULT_GRAPHITE_VERSION = last(GRAPHITE_VERSIONS)!` — i.e. `'1.1'` |
| `pkg/graphite/graphite.go:37-61` | `NewDatasource` — the only jsonData/root read is `settings.URL` (`:51`) and `settings.ID` (`:52`); everything else is delegated to `settings.HTTPClientOptions(ctx)` (`:38`) |
| `pkg/graphite/admission_handler.go:30-53` | Validates that the payload is a `DataSourceInstanceSettings` protobuf (`:32`), `apiVersion` is `""` or `"v0alpha1"` (`:45-49`), and `settings.URL` is non-empty (`:51-53`) |
| `pkg/graphite/healthcheck.go:16-89` | Issues a `constantLine(100)` render request against the datasource URL as the CheckHealth probe |
| `pkg/graphite/types.go:1-83` | Query DTOs — no `Settings` struct at all |
| `go.mod:1-14` | `grafana-plugin-sdk-go v0.292.1`, `k8s.io/apimachinery v0.35.4` |
| `package.json` | External component versions (see next table) |

Notably absent: no `pkg/models/settings.go`, no upstream `LoadSettings` — the
Graphite plugin does not own a backend jsonData settings model. All
server-side reads of settings go through the SDK.

### External editor components

Read at the exact versions pinned in the plugin's `package.json`
(`@grafana/ui@13.1.0`, `@grafana/data@13.1.0`, `@grafana/runtime@13.1.0`).
Sources checked out at the corresponding grafana/grafana tag `v13.1.0`.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `DataSourceHttpSettings` | `@grafana/ui@13.1.0` | grafana/grafana `v13.1.0` `packages/grafana-ui/src/components/DataSourceSettings/DataSourceHttpSettings.tsx` | URL Field (label defaults to `"URL"`, description from `default-url-tooltip` "Specify a complete HTTP URL (for example http://your_server:8080)"), `Allowed cookies` Field + TagsInput (label + description strings from the `t()` catalog), `Timeout` Field (label "Timeout", description "HTTP request timeout in seconds", placeholder "Timeout in seconds"), `Basic auth` InlineSwitch, `With Credentials` InlineSwitch with tooltip "Whether credentials such as cookies or auth headers should be sent with cross-site requests." — Graphite does not pass `showAccessOptions`, so the Access radio group is never rendered |
| `BasicAuthSettings` | `@grafana/ui@13.1.0` | `packages/grafana-ui/src/components/DataSourceSettings/BasicAuthSettings.tsx` | `User` FormField (label "User", placeholder "user"), `Password` `SecretFormField` (defaults `label='Password'`, `placeholder='Password'`) |
| `HttpProxySettings` | `@grafana/ui@13.1.0` | `packages/grafana-ui/src/components/DataSourceSettings/HttpProxySettings.tsx` | `TLS Client Auth`, `With CA Cert` (tooltip "Needed for verifying self-signed TLS Certs"), `Skip TLS Verify`, `Forward OAuth Identity` (tooltip "Forward the user's upstream OAuth identity to the data source (Their access token gets passed along).") — all InlineSwitches writing to `jsonData.tlsAuth`, `jsonData.tlsAuthWithCACert`, `jsonData.tlsSkipVerify`, `jsonData.oauthPassThru` |
| `TLSAuthSettings` + `CertificationKey` | `@grafana/ui@13.1.0` | `packages/grafana-ui/src/components/DataSourceSettings/TLSAuthSettings.tsx` | `TLS/SSL Auth Details` heading + tooltip; `CA Cert` textarea placeholder `Begins with -----BEGIN CERTIFICATE-----`; `ServerName` FormField placeholder `domain.example.com`; `Client Cert` textarea (same placeholder as CA); `Client Key` textarea placeholder `Begins with -----BEGIN RSA PRIVATE KEY-----` — only rendered when `jsonData.tlsAuth` or `jsonData.tlsAuthWithCACert` is true |
| `CustomHeadersSettings` (excluded) | `@grafana/ui@13.1.0` | `packages/grafana-ui/src/components/DataSourceSettings/CustomHeadersSettings.tsx` | Indexed `httpHeaderName<N>` / `httpHeaderValue<N>` storage pattern; **not modeled** in this schema (see [Modeling decisions](#modeling-decisions)) |
| `SecureSocksProxySettings` (excluded) | `@grafana/ui@13.1.0` | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | Storage key it writes (`jsonData.enableSecureSocksProxy`) — confirmed and excluded per `AGENTS.md` |
| `Alert`, `Field`, `FieldSet`, `Select`, `Switch`, `Input`, `TagsInput`, `InlineField`, `InlineSwitch`, `Button`, `Icon`, `Box` | `@grafana/ui@13.1.0` | `packages/grafana-ui/src/components/` | Prop names (`label`, `description`, `placeholder`, `value`, `onChange`, `isConfigured`, `onReset`, `id`, `aria-label`) — needed to know which UI attributes to record |
| `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps`, `updateDatasourcePluginJsonDataOption`, `onUpdateDatasourceJsonDataOptionSelect`, `onUpdateDatasourceJsonDataOptionChecked`, `store` | `@grafana/data@13.1.0` | `packages/grafana-data/src/` | Base jsonData interface; the `componentDidMount` writer used at `ConfigEditor.tsx:44` |
| `config.secureSocksDSProxyEnabled` | `@grafana/runtime@13.1.0` | `packages/grafana-runtime/src/config.ts` | Feature flag toggling the excluded `SecureSocksProxySettings` widget |

Note: the DataSourceHttpSettings component carries a JSDoc `@deprecated` tag
pointing to `@grafana/plugin-ui`. The Graphite plugin continues to use the
deprecated component — see [Upstream findings](#upstream-findings) #1.

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line`
where each of its label, placeholder, tooltip, default, storage key, and
value type is defined. Where a field draws from multiple lines, all lines are
listed.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | `DataSourceHttpSettings.tsx:189` (default `urlLabel = 'URL'`; Graphite editor does not override) | `ConfigEditor.tsx:61` (`defaultUrl="http://localhost:8080"`) | `settings.URL string` — SDK base | Required per `admission_handler.go:51-53` (`requiredWhen: "true"`) |
| `root_basicAuth` | `basicAuth` | `root` | `DataSourceHttpSettings.tsx:314` (`Basic auth` InlineSwitch) | Default `false` | Root SDK bool | Toggle visible in editor Auth section |
| `root_withCredentials` | `withCredentials` | `root` | `DataSourceHttpSettings.tsx:329` (`With Credentials` InlineSwitch) | Tooltip `DataSourceHttpSettings.tsx:331-333` | Root SDK bool | Independent from `basicAuth` — both can be true |
| `root_basicAuthUser` | `basicAuthUser` | `root` | `BasicAuthSettings.tsx:41` (`t(..., 'User')`) | `BasicAuthSettings.tsx:43` (`t(..., 'user')`) | SDK `settings.BasicAuthUser string` | `dependsOn: root_basicAuth == true`; `requiredWhen: root_basicAuth == true` |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | `SecretFormField.tsx:38` (default `label = 'Password'`) | `SecretFormField.tsx:44` (default `placeholder = 'Password'`) | Role `auth.basic.password` | Same conditional/required as `basicAuthUser` |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | `HttpProxySettings.tsx:26` (`t(..., 'TLS Client Auth')`) | Default `false` | `bool` — SDK TLS pack | — |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | `HttpProxySettings.tsx:37` (`t(..., 'With CA Cert')`) | Tooltip `HttpProxySettings.tsx:38-41`; default `false` | `bool` | — |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | `HttpProxySettings.tsx:54` (`t(..., 'Skip TLS Verify')`) | Default `false` | Role `transport.tlsSkipVerify` | — |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | `HttpProxySettings.tsx:66` (`t(..., 'Forward OAuth Identity')`) | Tooltip `HttpProxySettings.tsx:67-70`; default `false` | Role `auth.forwardOAuthToken.enabled` | — |
| `jsonData_serverName` | `serverName` | `jsonData` | `TLSAuthSettings.tsx:93` (`t(..., 'ServerName')`) | `TLSAuthSettings.tsx:97` (untranslated `placeholder="domain.example.com"`) | Role `tls.serverName` | `dependsOn: jsonData_tlsAuth == true`; required for the mTLS contract |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | `TLSAuthSettings.tsx:81` (`t(..., 'CA Cert')`) | `TLSAuthSettings.tsx:75-79` (`Begins with -----BEGIN CERTIFICATE-----`) | Role `tls.caCert` | `dependsOn` / `requiredWhen`: `jsonData_tlsAuthWithCACert == true` |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | `TLSAuthSettings.tsx:104` (`t(..., 'Client Cert')`) | Same as `CA Cert` — `Begins with -----BEGIN CERTIFICATE-----` | Role `tls.clientCert` | `dependsOn`/`requiredWhen`: `jsonData_tlsAuth == true` |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | `TLSAuthSettings.tsx:115` (`t(..., 'Client Key')`) | `TLSAuthSettings.tsx:117-121` (`Begins with -----BEGIN RSA PRIVATE KEY-----`) | Role `tls.clientKey` | Same conditional/required as `tlsClientCert` |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | `DataSourceHttpSettings.tsx:230-233` (`t(..., 'Allowed cookies')`) | Description `DataSourceHttpSettings.tsx:231-234` | `string[]` | Rendered only when `access === 'proxy'` (default) |
| `jsonData_timeout` | `timeout` | `jsonData` | `DataSourceHttpSettings.tsx:246` (`t(..., 'Timeout')`) | `DataSourceHttpSettings.tsx:255` (`t(..., 'Timeout in seconds')`); description `DataSourceHttpSettings.tsx:247-250` | `number` — parsed with `parseInt` at `DataSourceHttpSettings.tsx:258` | Role `transport.timeoutSeconds` |
| `jsonData_graphiteVersion` | `graphiteVersion` | `jsonData` | `ConfigEditor.tsx:69` (`label="Version"`) | Options from `versions.ts:3` `['0.9','1.0','1.1']` — rendered as `${version}.x` at `ConfigEditor.tsx:22`; default `DEFAULT_GRAPHITE_VERSION = '1.1'` (`versions.ts:5`, written on mount `ConfigEditor.tsx:44`) | `string` — `types.ts:24` | The editor's `componentDidMount` guarantees non-empty on load |
| `jsonData_graphiteType` | `graphiteType` | `jsonData` | `ConfigEditor.tsx:83` (`label="Graphite backend type"`) | Options from `Object.entries(GraphiteType)` at `ConfigEditor.tsx:24-27` → `[{label:"Default",value:"default"},{label:"Metrictank",value:"metrictank"}]`; no default | `GraphiteType` — `types.ts:25,30-33` | Empty (undefined) is the untouched-editor state |
| `jsonData_rollupIndicatorEnabled` | `rollupIndicatorEnabled` | `jsonData` | `ConfigEditor.tsx:97` (`label="Rollup indicator"`) | Description `ConfigEditor.tsx:98`; default `false` | `bool` — `types.ts:26` | `dependsOn: jsonData_graphiteType == 'metrictank'` (mirrors the `{...&&}` render guard at `ConfigEditor.tsx:95`) |
| `jsonData_importConfiguration` | `importConfiguration` | `jsonData` | `MappingsConfiguration.tsx:21` (`<h3>Label mappings</h3>`) | Help markdown transcribed from `MappingsHelp.tsx:10-72`; per-row `Input placeholder="e.g. test.metric.(labelName).*"` at `MappingsConfiguration.tsx:45` | `GraphiteQueryImportConfiguration` — `types.ts:27,56-71` | Modeled as opaque `object` with rich `help.markdown`; the shape is `{loki:{mappings:[{matchers:[{value,labelName?}]}]}}` |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | URL | Yes (direct: `pkg/graphite/graphite.go:51`, `pkg/graphite/admission_handler.go:51`) |
| `root_basicAuth` | `basicAuth` | `root` | Basic auth | Yes (SDK via `HTTPClientOptions`) |
| `root_withCredentials` | `withCredentials` | `root` | With Credentials | Yes (SDK) |
| `root_basicAuthUser` | `basicAuthUser` | `root` | User | Yes (SDK) |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | Password | Yes (SDK) |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | TLS Client Auth | Yes (SDK) |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | With CA Cert | Yes (SDK) |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | Skip TLS Verify | Yes (SDK) |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | Forward OAuth Identity | Yes (SDK) |
| `jsonData_serverName` | `serverName` | `jsonData` | ServerName | Yes (SDK) |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | CA Cert | Yes (SDK) |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | Client Cert | Yes (SDK) |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | Client Key | Yes (SDK) |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | Allowed cookies | Yes (SDK) |
| `jsonData_timeout` | `timeout` | `jsonData` | Timeout | Yes (SDK) |
| `jsonData_graphiteVersion` | `graphiteVersion` | `jsonData` | Version | No — frontend-only (query editor function library selector) |
| `jsonData_graphiteType` | `graphiteType` | `jsonData` | Graphite backend type | No — frontend-only (query editor toggles Metrictank-specific rendering) |
| `jsonData_rollupIndicatorEnabled` | `rollupIndicatorEnabled` | `jsonData` | Rollup indicator | No — frontend-only (adds a badge to panel headers) |
| `jsonData_importConfiguration` | `importConfiguration` | `jsonData` | Label mappings | No — frontend-only (consumed by Explore's datasource-switch flow in `@grafana/data`) |

### Frontend-only settings

- **`jsonData.graphiteVersion`** — read by the query editor to select the
  Graphite function library exposed in autocomplete. The backend never reads
  it. The editor's `componentDidMount` writes `DEFAULT_GRAPHITE_VERSION` on
  every mount if missing (`ConfigEditor.tsx:43-45`).
- **`jsonData.graphiteType`** — used by the query editor to enable
  Metrictank-specific features (query-processing metadata rendering). The
  backend never reads it.
- **`jsonData.rollupIndicatorEnabled`** — purely visual: shows an info icon
  on panel headers when data is aggregated. Backend never reads it.
- **`jsonData.importConfiguration`** — cross-datasource migration hint used
  when Explore converts a Graphite query into a Loki query. Neither Graphite
  nor Loki reads it at query time; it is consumed by
  `@grafana/data`'s datasource-switch logic.

### Backend-only settings

None — the Graphite plugin's Go code only reads `settings.URL` and
`settings.ID` directly, and otherwise delegates to the SDK's
`HTTPClientOptions`. There is no upstream `pkg/models/settings.go`
unmarshaling jsonData server-side.

### Excluded settings

- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy` and associated
  socks-proxy fields) — rendered conditionally at `ConfigEditor.tsx:64` when
  `config.secureSocksDSProxyEnabled` is set on the Grafana instance.
  Deliberately omitted per `AGENTS.md`.
- **Custom HTTP headers** (`@grafana/ui`'s `CustomHeadersSettings`) — the
  editor writes indexed pairs `jsonData.httpHeaderName<N>` /
  `secureJsonData.httpHeaderValue<N>` starting at index 1. Not modeled as a
  first-class field because the storage keys are dynamic. Downstream tools
  should walk `jsonData` for the `httpHeaderName` prefix and pair up matching
  `httpHeaderValue<N>` secrets; the SDK's `HTTPClientOptions` already does
  this and forwards the resulting headers to Graphite.
- **`root.access`** — DataSourceHttpSettings stores `'proxy'` (Server) or
  `'direct'` (Browser) at the datasource root. Graphite does not pass
  `showAccessOptions`, so the editor never renders an Access control and new
  datasources always end up with `access: 'proxy'`. Legacy datasources with
  `access: 'direct'` still work but trigger a deprecation Alert
  (`ConfigEditor.tsx:54-58`). Not modeled as an editor field but included in
  `settings.ts` `RootConfig` for round-trip compatibility and demonstrated
  via the `legacyDirectAccess` example.

## Where the types are defined

The configuration types are spread across the plugin and its dependencies —
some fields and base types come from libraries/SDKs rather than the plugin
itself.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `GraphiteOptions`, `GraphiteType`, `GraphiteQueryImportConfiguration`, `GraphiteLokiMapping`, `GraphiteMetricLokiMatcher` | `src/types.ts:23-71` | plugin ([grafana/grafana-graphite-datasource](https://github.com/grafana/grafana-graphite-datasource)) |
| `GRAPHITE_VERSIONS`, `DEFAULT_GRAPHITE_VERSION` | `src/versions.ts:3-5` | plugin |
| `fromString`, `toString` (mapping string ↔ matchers) | `src/configuration/parseLokiLabelMappings.ts:7-31` | plugin |
| `DataSourceJsonData` (base interface: `authType`, `defaultRegion`, `profile`, `manageAlerts`, `allowAsRecordingRulesTarget`, `alertmanagerUid`, `disableGrafanaCache`) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `13.1.0` |
| `DataSourcePluginOptionsEditorProps`, `DataSourceSettings`, `updateDatasourcePluginJsonDataOption`, `onUpdateDatasourceJsonDataOptionSelect`, `onUpdateDatasourceJsonDataOptionChecked`, `store` | `packages/grafana-data/src/` | `@grafana/data` `13.1.0` |
| `DataSourceHttpSettings`, `BasicAuthSettings`, `HttpProxySettings`, `TLSAuthSettings`, `CertificationKey`, `CustomHeadersSettings`, `SecureSocksProxySettings`, `Alert`, `Field`, `FieldSet`, `Select`, `Switch`, `Input`, `TagsInput`, `InlineField`, `InlineSwitch`, `Button` | `packages/grafana-ui/src/components/` | `@grafana/ui` `13.1.0` |
| `config` (`secureSocksDSProxyEnabled`) | `packages/grafana-runtime/src/` | `@grafana/runtime` `13.1.0` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `NewDatasource`, `DataSource`, `datasourceInfo` (reads `settings.URL` and `settings.ID`), `CheckHealth`, `MutateAdmission` / `ValidateAdmission` (URL / apiVersion checks) | `pkg/graphite/graphite.go:37-111`, `pkg/graphite/admission_handler.go:17-72`, `pkg/graphite/healthcheck.go:16-89` | plugin |
| `backend.DataSourceInstanceSettings` (carries `URL`, `BasicAuthEnabled`, `BasicAuthUser`, `JSONData`, `DecryptedSecureJSONData`), `HTTPClientOptions(ctx)`, `AdmissionHandler` | `backend/common.go`, `backend/httpclient/`, `backend/admission.go` | `github.com/grafana/grafana-plugin-sdk-go v0.292.1` |

The models in this entry flatten the above into a single Go `Config` type
(root `URL`/`BasicAuth`/`BasicAuthUser`/`WithCredentials` tagged `json:"-"`,
plus the jsonData fields the editor writes and the SDK reads, plus
`DecryptedSecureJSONData`) and a `SecureJsonDataKey` typed constant list.
`settings.ts` keeps the three canonical TypeScript types (`RootConfig`,
`JsonDataConfig`, `SecureJsonDataConfig`).

## Modeling decisions

- **Direct-toggle auth model (no virtual selector)**: unlike Loki/Prometheus
  which use `@grafana/plugin-ui`'s `Auth` component with `visibleMethods` and
  render a Select, Graphite uses the older `@grafana/ui`
  `DataSourceHttpSettings`. That component exposes three independent
  switches — `root.basicAuth`, `root.withCredentials`, and
  `jsonData.oauthPassThru` — all of which can be true simultaneously.
  Modeling this as a `virtual_authMethod` discriminator would misrepresent
  the storage: there is no mutual-exclusion contract enforced by the editor
  or the SDK.
- **`root.access` intentionally omitted from the schema**: Graphite does not
  pass `showAccessOptions` to `DataSourceHttpSettings`, so the Access radio
  is never rendered. New datasources default to `access: 'proxy'`. The
  editor renders a deprecation Alert (`ConfigEditor.tsx:54-58`) when it
  observes `access: 'direct'` on a legacy datasource. The `RootConfig`
  TypeScript type keeps the field for round-trip compatibility, and the
  `legacyDirectAccess` example demonstrates a payload that carries it.
- **`requiredWhen` on `basicAuthUser` / `basicAuthPassword`**: keyed on the
  underlying storage field (`root_basicAuth == true`). The editor renders
  the Basic Auth Details block only when `basicAuth` is true, so the
  requirement is on the storage flag.
- **TLS pair requirements**: `TLSAuthSettings` only renders the ServerName
  and Client Cert / Client Key inputs when `jsonData.tlsAuth` is true, and
  the CA Cert input when `jsonData.tlsAuthWithCACert` is true. Encoded as
  `dependsOn` + `requiredWhen` on each field.
- **`graphiteVersion` defaulted to `'1.1'` in `ApplyDefaults`**: unlike the
  no-default Loki story, the Graphite editor's `componentDidMount` writes
  `DEFAULT_GRAPHITE_VERSION` on every load if the field is empty
  (`ConfigEditor.tsx:43-45`). Provisioned datasources that omit the field
  will get `'1.1'` written the first time the editor opens; we mirror that
  behaviour so `LoadConfig` gives callers editor-parity even before the
  editor has run.
- **`graphiteType` intentionally NOT defaulted**: the editor's `Select`
  shows no selection (`graphiteTypes.find(...) === undefined`) until the
  user picks one. Applying a default would fabricate a choice the user
  never made. Downstream code should treat empty as "use Graphite-flavour
  behaviour with no Metrictank features".
- **`rollupIndicatorEnabled` render-guarded on Metrictank**: the editor
  only shows the switch when `graphiteType === 'metrictank'`
  (`ConfigEditor.tsx:95`). Persisting `rollupIndicatorEnabled: true` with
  `graphiteType: 'default'` is legal storage but has no rendering effect.
  Captured as a `relationships` pair (informational) rather than a hard
  validation error.
- **`importConfiguration` modeled as an opaque `object` with rich
  `help.markdown`**: the storage shape is a 3-level nested tree
  (`{loki:{mappings:[{matchers:[{value,labelName?}]}]}}`) that the editor
  serializes from free-form strings via `parseLokiLabelMappings.ts`. Rather
  than decomposing every leaf into a dsconfig field, we describe the shape
  in the field's `help.markdown` (verbatim from `MappingsHelp.tsx` plus a
  storage-shape note) and keep the Go type as a full nested
  `GraphiteQueryImportConfiguration` struct so `LoadConfig` still gives
  callers typed access. This matches how the Jaeger entry models
  `tracesToLogsV2` / `tracesToMetrics`.
- **Field ID naming convention**: IDs are prefixed with their storage target
  for easy discoverability — `root_`, `jsonData_`, or `secureJsonData_` —
  followed by the camelCase storage key. The `key` property keeps the
  plugin's raw storage key.
- **Custom HTTP headers and Secure Socks Proxy excluded**: see
  [Excluded settings](#excluded-settings) above.
- **Flat `Config` in Go**: `settings.go` collapses jsonData fields and
  decrypted secrets onto a single `Config` struct. Root-level fields the
  editor and SDK both use (`URL`, `BasicAuth`, `BasicAuthUser`,
  `WithCredentials`) are carried with `json:"-"` tags so `LoadConfig`
  returns them alongside the jsonData shape. `access` is intentionally
  omitted from `Config` (see above) — callers who need it can read it from
  `backend.DataSourceInstanceSettings` directly.
- **`SecureJsonDataConfig` is a key list**: secure values are write-only, so
  the secure type is just the array of secret key names
  (`basicAuthPassword`, `tlsCACert`, `tlsClientCert`, `tlsClientKey`);
  consumers read `secureJsonFields` to see what is configured.
- **`ImportConfiguration` as a value struct (not a pointer)**: the
  conformance test's `JSONDataTypesMatchStruct` requires
  `valueType: "object"` fields to map to a Go `Struct` kind, not `Ptr`. A
  zero-value nested struct round-trips through unmarshal fine; callers can
  check `len(cfg.ImportConfiguration.Loki.Mappings) > 0` to detect
  configured mappings.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema`
bundle (the k8s-style schema Grafana's datasource API server serves as
`{apiVersion}.json`, `v0alpha1` today) from the embedded `dsconfig.json`:
root fields plus a nested `jsonData` object become the OpenAPI settings
`spec`, secure fields become `secureValues`, and virtual fields are skipped
(there are no virtual fields in this entry).

`SettingsExamples()` provides the default configuration plus one k8s-style
example per authentication method, TLS variant, backend flavour, and
Graphite-specific feature. Each example is a full instance-settings object
with the plugin configuration nested under `jsonData` and the relevant
write-only secrets under `secureJsonData` (placeholder values to be replaced
with real secrets; the default example — keyed by the empty string `""` —
carries an empty `basicAuthPassword` to show that no secret is required for
the default No-auth mode):

| Example | Auth | TLS | Extras | `secureJsonData` |
| --- | --- | --- | --- | --- |
| `""` (default) | None | — | `graphiteVersion=1.1` | `basicAuthPassword` (empty) |
| `noAuth` | None | — | `graphiteVersion=1.1`, `graphiteType=default` | `basicAuthPassword` (empty) |
| `basicAuth` | Basic | — | — | `basicAuthPassword` |
| `oauthForward` | OAuth Identity | — | — | `basicAuthPassword` (empty) |
| `tlsMutualAuth` | None | mTLS (serverName + client cert/key) | — | `tlsClientCert`, `tlsClientKey` |
| `tlsSelfSignedCA` | None | Custom CA | — | `tlsCACert` |
| `metrictank` | None | — | `graphiteType=metrictank`, `rollupIndicatorEnabled=true` | `basicAuthPassword` (empty) |
| `withLabelMappings` | None | — | `importConfiguration.loki.mappings` (single 4-matcher mapping) | `basicAuthPassword` (empty) |
| `legacyDirectAccess` | None | — | `access=direct`, `graphiteVersion=1.0` — the browser-access-mode legacy shape | `basicAuthPassword` (empty) |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and
returns a fully-defaulted, validated `Config`:

1. **Parse** — copy `settings.URL`, `settings.BasicAuthEnabled`,
   `settings.BasicAuthUser` into `Config`, unmarshal `settings.JSONData` into
   the jsonData portion of the same struct, and copy the four decrypted
   secrets into `DecryptedSecureJSONData`. The Graphite plugin has no
   upstream `LoadSettings` to mirror — `pkg/graphite/graphite.go:37-61` is
   the only server-side read of settings and it just uses `settings.URL` +
   `settings.ID` + `settings.HTTPClientOptions`.
2. **`ApplyDefaults`** — write `DefaultGraphiteVersion` (`'1.1'`) when the
   parsed `graphiteVersion` is empty. `graphiteType` is intentionally NOT
   defaulted.
3. **`Validate`** — enforce the runtime contract: URL is required (admission
   handler fails otherwise, `admission_handler.go:51-53`); `graphiteVersion`
   and `graphiteType` must belong to their enum sets (empty accepted
   because callers may call `Validate` before `ApplyDefaults`); Basic auth
   requires a username; mTLS requires serverName + client cert + client
   key; custom-CA requires the CA PEM; `timeout` must be non-negative.
   Errors are joined so every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with
`datasource_uid`, `datasource_name`, and `plugin` labels so log lines carry
request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are still
exported for callers that want to compose them themselves (e.g. provisioning
preview, schema-example round-trip, tests that need to distinguish
parse-level from policy-level errors). Skip them by never calling
`LoadConfig` in those flows — assemble a `Config` directly.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while
researching upstream. All preserved verbatim in the schema — the schema
records what the plugin **does**, not what it **should** do; these notes
exist so reviewers can reproduce each finding and decide separately whether
to fix upstream.

1. **`DataSourceHttpSettings` is deprecated but still used**: the component's
   JSDoc (`packages/grafana-ui/src/components/DataSourceSettings/DataSourceHttpSettings.tsx:73`)
   explicitly says "Use components from `@grafana/plugin-ui` instead"; every
   modern datasource plugin has migrated. Graphite has not. Migrating would
   introduce a virtual auth selector, an "Additional settings" collapse
   section, and a URL-format validator — and would change the storage of
   `withCredentials` (see below).
2. **`withCredentials` is orthogonal to auth method under `DataSourceHttpSettings`**:
   the switch is a standalone toggle in `DataSourceHttpSettings.tsx:329-343`,
   not part of an auth discriminator. Under `@grafana/plugin-ui`'s `Auth`
   component (used by newer plugins) it is folded into the auth method
   `CrossSiteCredentials`. Modeled here as a direct root switch to match
   what the Graphite editor actually renders.
3. **`graphiteVersion` is auto-written on every mount**: `ConfigEditor.tsx:43-45`
   calls `updateDatasourcePluginJsonDataOption(this.props, 'graphiteVersion', this.currentGraphiteVersion)`
   unconditionally in `componentDidMount`. That means opening the config
   editor on an untouched, provisioned datasource silently writes
   `graphiteVersion: '1.1'` to jsonData. Provisioning tooling that diffs
   against upstream state should expect this on first edit.
4. **`graphiteType` shows no default in the editor**: the Select at
   `ConfigEditor.tsx:87-93` uses `value={graphiteTypes.find(type => type.value === options.jsonData.graphiteType)}`.
   For an untouched datasource, that resolves to `undefined` and the Select
   renders empty. Users have to actively pick "Default" or "Metrictank" for
   `graphiteType` to be persisted. This differs from `graphiteVersion`,
   which is defaulted; the inconsistency is preserved by our schema (no
   default on `jsonData_graphiteType`).
5. **Multiline description with embedded whitespace preserved verbatim**:
   the Graphite backend type description at `ConfigEditor.tsx:84-85` is a
   JSX string attribute spanning two lines — the newline and eight leading
   spaces on the second line are part of the string that gets passed to
   `<Field description=...>` and rendered as-is. Preserved in the schema's
   description field.
6. **Deprecation Alert for `access === 'direct'` but the value is still
   stored**: `ConfigEditor.tsx:54-58` renders a warning Alert on load when
   the datasource has browser access, but the editor never rewrites
   `access` to `'proxy'`. A user acknowledging the warning still has to
   change it elsewhere (there is no visible control), or the datasource
   stays in the deprecated mode indefinitely. Modeled in `settings.ts`
   `RootConfig` and the `legacyDirectAccess` example, not in the schema
   fields (since it is not editor-editable).
7. **No frontend URL validation surfaced**: `DataSourceHttpSettings.tsx:167-171`
   computes `isValidUrl` against a regex and passes `invalid`/`error` props
   to the URL Field, so the editor shows an inline error for bad URLs. But
   the Graphite backend's admission handler only checks for empty URL, not
   URL well-formedness (`admission_handler.go:51-53`). A provisioning
   payload with a URL that fails the frontend regex will still be accepted
   at admission time and fail later at request time.
8. **`rollupIndicatorEnabled` can be persisted without effect**: setting
   the flag with `graphiteType !== 'metrictank'` is stored but never
   rendered anywhere. Not a bug per se, just dead configuration —
   captured as a `relationships.pair` between `jsonData_graphiteType` and
   `jsonData_rollupIndicatorEnabled`.
9. **`importConfiguration` is not read by either datasource backend**: it
   is only consumed by `@grafana/data`'s Explore query-conversion logic
   when the user switches from Graphite to Loki. A Graphite datasource
   with configured mappings but no Loki datasource in the org has stored
   the field for nothing.
10. **`SecureSocksProxySettings` also writes to jsonData**: the excluded
    Secure Socks Proxy widget writes `jsonData.enableSecureSocksProxy` and
    related fields. Provisioning payloads that include those keys will not
    round-trip through this schema — they will be preserved in the raw
    `JSONData` but not represented in `Config` or `SettingsExamples`.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in
  this repo) — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft 2020-12, `additionalProperties: false`) — passes (implicit via
  `NewSchema()` invocation inside conformance tests).
- `go test ./...` on this entry — passes (all 8 conformance subtests:
  `BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`,
  `SchemaSpecHasNoSecureJSON`, `ConfigSchemaValid`, `JSONDataMatchesStruct`,
  `JSONDataTypesMatchStruct`, `SecureValuesMatchLoadSettings`, plus 20+
  `LoadConfig`/`ApplyDefaults`/`Validate` table tests).
- `settings.go`/`schema.go`: `go build`, `go vet`, `gofmt` — clean.
- Full `registry/` module: `go build`, `go vet`, `go test`, `gofmt -l .` —
  clean; no regressions in the 13 pre-existing entries.
- `settings.ts`: exports the three canonical types (`RootConfig`,
  `JsonDataConfig`, `SecureJsonDataConfig`) plus the domain sub-types
  (`GraphiteVersion`, `GraphiteType`, `GraphiteLokiMapping`, etc.) —
  reviewed by hand against the frontend sources; no `tsc` runner is wired
  into the registry module.
