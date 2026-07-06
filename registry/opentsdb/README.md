# opentsdb

Declarative configuration schema for the [OpenTSDB datasource plugin](https://github.com/grafana/grafana-opentsdb-datasource) (`opentsdb`).

## Upstream researched

- **Repo**: `github.com/grafana/grafana-opentsdb-datasource`
- **Ref**: `main`
- **Commit SHA**: `569fe9dec38c0d2b90a4f0441040b393393fdbad`

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders,
tooltips, option labels/values, section titles, defaults, validations,
dependency and required-when expressions, storage keys, storage targets,
value types, group titles, and instructions — is traceable to a specific
`file:line` in the upstream repo at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone --depth 20 https://github.com/grafana/grafana-opentsdb-datasource \
  /var/folders/2l/6dq44779565gnyvkpmhhv1th0000gn/T/opencode/grafana-opentsdb-datasource
git -C /var/folders/2l/6dq44779565gnyvkpmhhv1th0000gn/T/opencode/grafana-opentsdb-datasource \
  checkout 569fe9dec38c0d2b90a4f0441040b393393fdbad
```

If upstream `main` has advanced past this SHA, re-diff the sources listed
under [Sources researched](#sources-researched) before merging any changes to
this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, relationships, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: root URL/basicAuth/basicAuthUser/withCredentials tagged `json:"-"`, jsonData fields including the three OpenTSDB-specific numeric knobs, and `DecryptedSecureJSONData`), `PluginID`, `OpenTsdbVersion` / `OpenTsdbResolution` / `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth / TLS / behaviour variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the
shared [`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA (`569fe9d`), plus
external editor components at the exact versions the plugin's `package.json`
pins.

### Plugin repo (`github.com/grafana/grafana-opentsdb-datasource@569fe9d`)

| File | What was read |
| --- | --- |
| `src/plugin.json:1-43` | `pluginType` (`id` = `"opentsdb"`), `pluginName` (`name` = `"OpenTSDB"`), docs URL (`info.links[1].url` = `"https://grafana.com/docs/grafana/latest/datasources/opentsdb/"`), `grafanaDependency: >=12.3.0-0` |
| `src/components/ConfigEditor.tsx:10-24` | Outer editor — composes `<DataSourceHttpSettings defaultUrl="http://localhost:4242" secureSocksDSProxyEnabled=...>` (`:15-20`) then `<OpenTsdbDetails value={options} onChange={onOptionsChange} />` (`:21`); no deprecation Alert, no help drawer, no custom `componentDidMount` |
| `src/components/OpenTsdbDetails.tsx:1-89` | OpenTSDB-specific settings panel — `<FieldSet label="OpenTSDB settings">` (`:32`) containing Version Select (`:33-41`, options `<=2.1`→1, `==2.2`→2, `==2.3`→3, `==2.4`→4, fallback `tsdbVersions[0]`), Resolution Select (`:42-53`, options `second`→1, `millisecond`→2, fallback `tsdbResolutions[0]`), Lookup limit Input `type="number"` (`:54-62`, `value ?? 1000`) |
| `src/types.ts:1-53` | `OpenTsdbOptions extends DataSourceJsonData` with `tsdbVersion: number`, `tsdbResolution: number`, `lookupLimit: number` (`:35-39`); `OpenTsdbQuery`, `LegacyAnnotation`, `OpenTsdbFilter` |
| `src/datasource.ts:37-71` | `OpenTsDatasource` constructor — reads `instanceSettings.url` (`:55`), `instanceSettings.name`, `instanceSettings.withCredentials` (`:57`), `instanceSettings.basicAuth` (`:58`), then falls back to `tsdbVersion || 1`, `tsdbResolution || 1`, `lookupLimit || 1000` (`:60-62`) |
| `src/datasource.ts:176-204,229-247,264-271` | Where the three jsonData knobs are actually used at runtime — msResolution from tsdbResolution, showQuery when tsdbVersion===3, lookupLimit as the `max`/`limit` param for suggest and lookup, and `_addCredentialOptions` (which treats `this.basicAuth` as a header value — see Upstream findings #2) |
| `pkg/opentsdb/opentsdb.go:24-69` | `NewDatasource` builds an HTTPClient via `settings.HTTPClientOptions(ctx)` (`:29`), reads `settings.URL` directly (`:47`), unmarshals `settings.JSONData` into a local `type JSONData struct { TSDBVersion float32; TSDBResolution int32; LookupLimit int32 }` (`:65-69`) whose fields feed the `datasourceInfo` cache |
| `pkg/opentsdb/opentsdb.go:87-138` | `CheckHealth` — parses `dsInfo.URL`, appends `/api/suggest?q=cpu&type=metrics`, does a GET, requires HTTP 200 |
| `pkg/opentsdb/utils.go:132-156` | `CreateRequest` — POSTs to `{url}/api/query`; adds `?arrays=true` when `TSDBVersion == 4` |
| `pkg/opentsdb/utils.go:225-311` | `ParseResponse` — dispatches to `ParseResponseLT24` or `ParseResponse24` based on `tsdbVersion == 4` |
| `pkg/opentsdb/callresource.go:1-433` | `HandleSuggestQuery`, `HandleAggregatorsQuery`, `HandleFiltersQuery`, `HandleLookupQuery` (→ `HandleKeyLookup`, `HandleKeyValueLookup`) — resource endpoints that also read `dsInfo.URL` and `dsInfo.LookupLimit` (`:361`) |
| `pkg/opentsdb/types.go:1-30` | Response DTOs — no `Settings` struct |
| `go.mod:1-14` | `grafana-plugin-sdk-go v0.292.1` |
| `package.json` | External component versions (see next table) |

Notably absent: no `pkg/models/settings.go` (upstream) — the OpenTSDB
plugin's backend `JSONData` struct is defined in-line at
`pkg/opentsdb/opentsdb.go:65-69` and has no `LoadSettings` helper. There is
also no `admission_handler.go` and no `UnmarshalJSON` overriding the raw
parsing.

### External editor components

Read at the exact versions pinned in the plugin's `package.json`
(`@grafana/ui@13.0.2`, `@grafana/data@13.0.2`, `@grafana/runtime@13.0.2`).

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `DataSourceHttpSettings` | `@grafana/ui@13.0.2` | `packages/grafana-ui/src/components/DataSourceSettings/DataSourceHttpSettings.tsx` | URL Field (default label `"URL"`, description "Specify a complete HTTP URL (for example http://your_server:8080)"), `Allowed cookies` Field + TagsInput (label + description), `Timeout` Field (label "Timeout", description "HTTP request timeout in seconds", placeholder "Timeout in seconds"), `Basic auth` InlineSwitch, `With Credentials` InlineSwitch with tooltip "Whether credentials such as cookies or auth headers should be sent with cross-site requests." — OpenTSDB does not pass `showAccessOptions`, so the Access radio group is never rendered |
| `BasicAuthSettings` | `@grafana/ui@13.0.2` | `packages/grafana-ui/src/components/DataSourceSettings/BasicAuthSettings.tsx` | `User` FormField (label "User", placeholder "user"), `Password` `SecretFormField` (defaults `label='Password'`, `placeholder='Password'`) |
| `HttpProxySettings` | `@grafana/ui@13.0.2` | `packages/grafana-ui/src/components/DataSourceSettings/HttpProxySettings.tsx` | `TLS Client Auth`, `With CA Cert` (tooltip "Needed for verifying self-signed TLS Certs"), `Skip TLS Verify`, `Forward OAuth Identity` (tooltip "Forward the user's upstream OAuth identity to the data source (Their access token gets passed along).") — writes `jsonData.tlsAuth`, `jsonData.tlsAuthWithCACert`, `jsonData.tlsSkipVerify`, `jsonData.oauthPassThru` |
| `TLSAuthSettings` + `CertificationKey` | `@grafana/ui@13.0.2` | `packages/grafana-ui/src/components/DataSourceSettings/TLSAuthSettings.tsx` | `TLS/SSL Auth Details` heading; `CA Cert` textarea placeholder `Begins with -----BEGIN CERTIFICATE-----`; `ServerName` FormField placeholder `domain.example.com`; `Client Cert` textarea; `Client Key` textarea placeholder `Begins with -----BEGIN RSA PRIVATE KEY-----` — only rendered when `jsonData.tlsAuth` or `jsonData.tlsAuthWithCACert` is true |
| `CustomHeadersSettings` (excluded) | `@grafana/ui@13.0.2` | `packages/grafana-ui/src/components/DataSourceSettings/CustomHeadersSettings.tsx` | Indexed `httpHeaderName<N>` / `httpHeaderValue<N>` storage pattern; **not modeled** in this schema (see [Modeling decisions](#modeling-decisions)) |
| `SecureSocksProxySettings` (excluded) | `@grafana/ui@13.0.2` | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | Storage key it writes (`jsonData.enableSecureSocksProxy`) — confirmed and excluded per `AGENTS.md` |
| `Field`, `FieldSet`, `Select`, `Input` | `@grafana/ui@13.0.2` | `packages/grafana-ui/src/components/` | Prop names (`label`, `htmlFor`, `inputId`, `options`, `value`, `onChange`, `type`, `width`) |
| `DataSourceJsonData`, `DataSourcePluginOptionsEditorProps`, `DataSourceSettings`, `SelectableValue` | `@grafana/data@13.0.2` | `packages/grafana-data/src/types/` | Base jsonData interface, editor prop shape, Select option shape |
| `config.secureSocksDSProxyEnabled` | `@grafana/runtime@13.0.2` | `packages/grafana-runtime/src/config.ts` | Feature flag toggling the excluded `SecureSocksProxySettings` widget |

Note: `DataSourceHttpSettings` carries a JSDoc `@deprecated` tag pointing to
`@grafana/plugin-ui`. OpenTSDB continues to use the deprecated component —
see [Upstream findings](#upstream-findings) #1.

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line`
where each of its label, placeholder, tooltip, default, storage key, and
value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | `DataSourceHttpSettings.tsx` (default `urlLabel = 'URL'`) | `ConfigEditor.tsx:16` (`defaultUrl="http://localhost:4242"`) | `settings.URL string` — SDK base | Required per `pkg/opentsdb/opentsdb.go:47` and CheckHealth (`:87-138`) |
| `root_basicAuth` | `basicAuth` | `root` | `DataSourceHttpSettings.tsx` (`Basic auth` InlineSwitch) | Default `false` | Root SDK bool | — |
| `root_withCredentials` | `withCredentials` | `root` | `DataSourceHttpSettings.tsx` (`With Credentials` InlineSwitch) | Tooltip from DataSourceHttpSettings.tsx | Root SDK bool | Independent from `basicAuth` |
| `root_basicAuthUser` | `basicAuthUser` | `root` | `BasicAuthSettings.tsx` (`t(..., 'User')`) | `BasicAuthSettings.tsx` (`t(..., 'user')`) | SDK `settings.BasicAuthUser string` | `dependsOn: root_basicAuth == true`; `requiredWhen: root_basicAuth == true` |
| `secureJsonData_basicAuthPassword` | `basicAuthPassword` | `secureJsonData` | `SecretFormField.tsx` (default `label = 'Password'`) | `SecretFormField.tsx` (default `placeholder = 'Password'`) | Role `auth.basic.password` | Same conditional/required as `basicAuthUser` |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | `HttpProxySettings.tsx` (`t(..., 'TLS Client Auth')`) | Default `false` | `bool` | — |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | `HttpProxySettings.tsx` (`t(..., 'With CA Cert')`) | Tooltip from HttpProxySettings.tsx; default `false` | `bool` | — |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | `HttpProxySettings.tsx` (`t(..., 'Skip TLS Verify')`) | Default `false` | Role `transport.tlsSkipVerify` | — |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | `HttpProxySettings.tsx` (`t(..., 'Forward OAuth Identity')`) | Tooltip from HttpProxySettings.tsx; default `false` | Role `auth.forwardOAuthToken.enabled` | — |
| `jsonData_serverName` | `serverName` | `jsonData` | `TLSAuthSettings.tsx` (`t(..., 'ServerName')`) | `TLSAuthSettings.tsx` (untranslated `placeholder="domain.example.com"`) | Role `tls.serverName` | `dependsOn: jsonData_tlsAuth == true`; required for the mTLS contract |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | `TLSAuthSettings.tsx` (`t(..., 'CA Cert')`) | `Begins with -----BEGIN CERTIFICATE-----` | Role `tls.caCert` | `dependsOn` / `requiredWhen`: `jsonData_tlsAuthWithCACert == true` |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | `TLSAuthSettings.tsx` (`t(..., 'Client Cert')`) | Same as CA — `Begins with -----BEGIN CERTIFICATE-----` | Role `tls.clientCert` | `dependsOn`/`requiredWhen`: `jsonData_tlsAuth == true` |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | `TLSAuthSettings.tsx` (`t(..., 'Client Key')`) | `Begins with -----BEGIN RSA PRIVATE KEY-----` | Role `tls.clientKey` | Same conditional/required as `tlsClientCert` |
| `jsonData_keepCookies` | `keepCookies` | `jsonData` | `DataSourceHttpSettings.tsx` (`t(..., 'Allowed cookies')`) | Description from DataSourceHttpSettings.tsx | `string[]` | Rendered only when `access === 'proxy'` (default) |
| `jsonData_timeout` | `timeout` | `jsonData` | `DataSourceHttpSettings.tsx` (`t(..., 'Timeout')`) | Placeholder `t(..., 'Timeout in seconds')`; description `t(..., 'HTTP request timeout in seconds')` | `number` — parsed with `parseInt` in DataSourceHttpSettings | Role `transport.timeoutSeconds` |
| `jsonData_tsdbVersion` | `tsdbVersion` | `jsonData` | `OpenTsdbDetails.tsx:33` (`label="Version"`) | Options from `OpenTsdbDetails.tsx:8-13` `[{label:"<=2.1",value:1},{label:"==2.2",value:2},{label:"==2.3",value:3},{label:"==2.4",value:4}]`; default 1 (`datasource.ts:60` `|| 1`) | `number` — `types.ts:36` `tsdbVersion: number`; backend `float32` (`opentsdb.go:66`) | Enum-constrained via `allowedValues` |
| `jsonData_tsdbResolution` | `tsdbResolution` | `jsonData` | `OpenTsdbDetails.tsx:42` (`label="Resolution"`) | Options from `OpenTsdbDetails.tsx:15-18` `[{label:"second",value:1},{label:"millisecond",value:2}]`; default 1 (`datasource.ts:61` `|| 1`) | `number` — `types.ts:37` `tsdbResolution: number`; backend `int32` (`opentsdb.go:67`) | Enum-constrained via `allowedValues` |
| `jsonData_lookupLimit` | `lookupLimit` | `jsonData` | `OpenTsdbDetails.tsx:54` (`label="Lookup limit"`) | `<Input type="number" value={value.jsonData.lookupLimit ?? 1000}>` (`OpenTsdbDetails.tsx:58`); default 1000 (`datasource.ts:62` `|| 1000`) | `number` — `types.ts:38` `lookupLimit: number`; backend `int32` (`opentsdb.go:68`) | `range: {min: 0}` validation; **stored as a string on first edit** — see [Upstream findings](#upstream-findings) #3 |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `root_url` | `url` | `root` | URL | Yes (direct: `pkg/opentsdb/opentsdb.go:47`, `:92`, `pkg/opentsdb/utils.go:133`, `pkg/opentsdb/callresource.go`) |
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
| `jsonData_tsdbVersion` | `tsdbVersion` | `jsonData` | Version | Yes (direct: `pkg/opentsdb/opentsdb.go:65-68`, `utils.go:138,247`) |
| `jsonData_tsdbResolution` | `tsdbResolution` | `jsonData` | Resolution | Yes (direct: `pkg/opentsdb/opentsdb.go:65-68`; but also frontend-only side that adds `msResolution` on outgoing bodies at `datasource.ts:178-180`) |
| `jsonData_lookupLimit` | `lookupLimit` | `jsonData` | Lookup limit | Yes (direct: `pkg/opentsdb/opentsdb.go:65-68`, `callresource.go:361`) |

### Frontend-only settings

None — every editor-writable field in this entry is either consumed by the
plugin's Go backend directly (URL, tsdbVersion, tsdbResolution, lookupLimit)
or by the SDK's `HTTPClientOptions` when constructing the HTTP client (auth,
TLS, cookies, timeout, OAuth forward).

### Backend-only settings

None — every jsonData field the backend reads (`tsdbVersion`,
`tsdbResolution`, `lookupLimit`) is also exposed in the config editor.

### Excluded settings

- **Secure Socks Proxy** (`jsonData.enableSecureSocksProxy` and associated
  socks-proxy fields) — rendered conditionally at `ConfigEditor.tsx:19` when
  `config.secureSocksDSProxyEnabled` is set on the Grafana instance.
  Deliberately omitted per `AGENTS.md`.
- **Custom HTTP headers** (`@grafana/ui`'s `CustomHeadersSettings`) — the
  editor writes indexed pairs `jsonData.httpHeaderName<N>` /
  `secureJsonData.httpHeaderValue<N>` starting at index 1. Not modeled as a
  first-class field because the storage keys are dynamic. Downstream tools
  should walk `jsonData` for the `httpHeaderName` prefix and pair up matching
  `httpHeaderValue<N>` secrets; the SDK's `HTTPClientOptions` already does
  this and forwards the resulting headers to OpenTSDB.
- **`root.access`** — DataSourceHttpSettings stores `'proxy'` (Server) or
  `'direct'` (Browser) at the datasource root. OpenTSDB does not pass
  `showAccessOptions`, so the editor never renders an Access control and new
  datasources always end up with `access: 'proxy'`. Not modeled as an editor
  field but included in `settings.ts` `RootConfig` for round-trip
  compatibility.

## Where the types are defined

The configuration types are spread across the plugin and its dependencies —
some fields and base types come from libraries/SDKs rather than the plugin
itself.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `OpenTsdbOptions`, `OpenTsdbQuery`, `OpenTsdbFilter`, `LegacyAnnotation` | `src/types.ts:3-53` | plugin ([grafana/grafana-opentsdb-datasource](https://github.com/grafana/grafana-opentsdb-datasource)) |
| `tsdbVersions`, `tsdbResolutions` option lists | `src/components/OpenTsdbDetails.tsx:8-18` | plugin |
| `DataSourceJsonData` (base interface: `authType`, `defaultRegion`, `profile`, `manageAlerts`, `allowAsRecordingRulesTarget`, `alertmanagerUid`, `disableGrafanaCache`) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `13.0.2` |
| `DataSourcePluginOptionsEditorProps`, `DataSourceSettings`, `SelectableValue` | `packages/grafana-data/src/` | `@grafana/data` `13.0.2` |
| `DataSourceHttpSettings`, `BasicAuthSettings`, `HttpProxySettings`, `TLSAuthSettings`, `CertificationKey`, `CustomHeadersSettings`, `SecureSocksProxySettings`, `Field`, `FieldSet`, `Select`, `Input` | `packages/grafana-ui/src/components/` | `@grafana/ui` `13.0.2` |
| `config` (`secureSocksDSProxyEnabled`) | `packages/grafana-runtime/src/` | `@grafana/runtime` `13.0.2` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `NewDatasource`, `DataSource`, `datasourceInfo`, `JSONData` (`TSDBVersion float32`, `TSDBResolution int32`, `LookupLimit int32`), `CheckHealth`, `CallResource`, `QueryData` | `pkg/opentsdb/opentsdb.go:24-208` | plugin |
| `BuildMetric`, `CreateRequest`, `DecodeResponseBody`, `CreateDataFrame`, `ParseResponse`, `ParseResponse24`, `ParseResponseLT24` | `pkg/opentsdb/utils.go:22-312` | plugin |
| `HandleSuggestQuery`, `HandleAggregatorsQuery`, `HandleFiltersQuery`, `HandleLookupQuery`, `HandleKeyLookup`, `HandleKeyValueLookup` | `pkg/opentsdb/callresource.go:15-433` | plugin |
| `backend.DataSourceInstanceSettings` (carries `URL`, `BasicAuthEnabled`, `BasicAuthUser`, `JSONData`, `DecryptedSecureJSONData`), `HTTPClientOptions(ctx)`, `httpclient.New`, `httpadapter.New` | `backend/common.go`, `backend/httpclient/`, `backend/resource/httpadapter/` | `github.com/grafana/grafana-plugin-sdk-go v0.292.1` |

The models in this entry flatten the above into a single Go `Config` type
(root `URL`/`BasicAuth`/`BasicAuthUser`/`WithCredentials` tagged `json:"-"`,
plus the jsonData fields the editor writes and the SDK reads, plus the
three OpenTSDB-specific numeric knobs that mirror the upstream `JSONData`
struct verbatim, plus `DecryptedSecureJSONData`). `settings.ts` keeps the
three canonical TypeScript types (`RootConfig`, `JsonDataConfig`,
`SecureJsonDataConfig`) plus the two numeric enums (`OpenTsdbVersion`,
`OpenTsdbResolution`).

## Modeling decisions

- **Direct-toggle auth model (no virtual selector)**: like Graphite, OpenTSDB
  uses the older `@grafana/ui` `DataSourceHttpSettings` (not the newer
  `@grafana/plugin-ui` `Auth` component). That exposes three independent
  switches — `root.basicAuth`, `root.withCredentials`, and
  `jsonData.oauthPassThru` — all of which can be true simultaneously.
  Modeling this as a `virtual_authMethod` discriminator would misrepresent
  the storage.
- **`root.access` intentionally omitted from the schema**: OpenTSDB does not
  pass `showAccessOptions`. New datasources default to `access: 'proxy'`.
  The `RootConfig` TypeScript type keeps the field for round-trip
  compatibility.
- **`requiredWhen` on `basicAuthUser` / `basicAuthPassword`**: keyed on
  `root_basicAuth == true` — the editor only renders the Basic Auth Details
  block when that flag is set.
- **TLS pair requirements**: `TLSAuthSettings` only renders ServerName and
  Client Cert / Client Key when `jsonData.tlsAuth` is true, and CA Cert when
  `jsonData.tlsAuthWithCACert` is true. Encoded as `dependsOn` +
  `requiredWhen` on each field.
- **`tsdbVersion`/`tsdbResolution`/`lookupLimit` defaulted in `ApplyDefaults`**:
  the plugin's frontend constructor at `src/datasource.ts:60-62` replaces
  zero/undefined values with `1`, `1`, and `1000` respectively at runtime.
  We mirror that behaviour in `Config.ApplyDefaults` so `LoadConfig` yields
  editor-parity even for a provisioned datasource that has never been
  opened in the editor. The editor itself does not write defaults on mount
  (unlike Graphite's `componentDidMount`), so raw jsonData in storage may
  legitimately be missing all three fields.
- **`tsdbVersion` typed as `float32` on `Config`**: mirrors
  `pkg/opentsdb/opentsdb.go:66` verbatim. `float32` accommodates both the
  frontend's `SelectableValue<number>.value` (JavaScript numbers are 64-bit
  floats) and the discrete integer enum the option list actually offers.
  The conformance test accepts any numeric Go kind against a `valueType:
  "number"` dsconfig field (`schema/conformance.go:377-380`).
- **`tsdbResolution` typed as `int32`, `lookupLimit` typed as `int32`**:
  mirror `pkg/opentsdb/opentsdb.go:67-68` verbatim.
- **Numeric options in the dsconfig `select`**: `FieldOption.value` accepts
  any non-null JSON value (`dsconfig/schema.json:1127-1131`) — using numeric
  values (`1..4`, `1..2`) matches the JSON storage type. Corresponding
  `allowedValues` validations also use numeric values.
- **`lookupLimit` with `range: {min: 0}` validation**: the frontend does not
  clamp the value at all (`OpenTsdbDetails.tsx:59` passes the raw string
  through), and the backend stores it in an `int32` used as an HTTP query
  parameter (`callresource.go:361`). A negative value would produce a
  nonsensical `limit=-N` on `/api/search/lookup`. The dsconfig `range`
  validation and `Config.Validate` both reject it.
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
  returns them alongside the jsonData shape.
- **`SecureJsonDataConfig` is a key list**: secure values are write-only, so
  the secure type is just the array of secret key names
  (`basicAuthPassword`, `tlsCACert`, `tlsClientCert`, `tlsClientKey`);
  consumers read `secureJsonFields` to see what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema`
bundle (the k8s-style schema Grafana's datasource API server serves as
`{apiVersion}.json`, `v0alpha1` today) from the embedded `dsconfig.json`:
root fields plus a nested `jsonData` object become the OpenAPI settings
`spec`, secure fields become `secureValues`, and virtual fields are skipped
(there are no virtual fields in this entry).

`SettingsExamples()` provides the default configuration plus one k8s-style
example per authentication method, TLS variant, and OpenTSDB-specific
behaviour knob. Each example is a full instance-settings object with the
plugin configuration nested under `jsonData` and the relevant write-only
secrets under `secureJsonData` (placeholder values to be replaced with real
secrets; the default example — keyed by the empty string `""` — carries an
empty `basicAuthPassword` to show that no secret is required for the default
No-auth mode):

| Example | Auth | TLS | Extras | `secureJsonData` |
| --- | --- | --- | --- | --- |
| `""` (default) | None | — | `tsdbVersion=1`, `tsdbResolution=1`, `lookupLimit=1000` | `basicAuthPassword` (empty) |
| `noAuth` | None | — | `tsdbVersion=4` (array-response parsing) | `basicAuthPassword` (empty) |
| `basicAuth` | Basic | — | `tsdbVersion=1` | `basicAuthPassword` |
| `oauthForward` | OAuth Identity | — | `tsdbVersion=1` | `basicAuthPassword` (empty) |
| `tlsMutualAuth` | None | mTLS (serverName + client cert/key) | `tsdbVersion=1` | `tlsClientCert`, `tlsClientKey` |
| `tlsSelfSignedCA` | None | Custom CA | `tsdbVersion=1` | `tlsCACert` |
| `millisecondResolution` | None | — | `tsdbVersion=3`, `tsdbResolution=2` (msResolution=true on outgoing bodies) | `basicAuthPassword` (empty) |
| `largeLookupLimit` | None | — | `tsdbVersion=4`, `lookupLimit=10000` | `basicAuthPassword` (empty) |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and
returns a fully-defaulted, validated `Config`:

1. **Parse** — copy `settings.URL`, `settings.BasicAuthEnabled`,
   `settings.BasicAuthUser` into `Config`, unmarshal `settings.JSONData`
   into the jsonData portion of the same struct (mirroring the upstream
   `JSONData` shape at `pkg/opentsdb/opentsdb.go:65-69` verbatim), and copy
   the four decrypted secrets into `DecryptedSecureJSONData`. The OpenTSDB
   plugin has no upstream `LoadSettings` — `pkg/opentsdb/opentsdb.go:28-53`
   inlines the parsing.
2. **`ApplyDefaults`** — write `1` when the parsed `TSDBVersion` is zero,
   `1` when `TSDBResolution` is zero, `1000` when `LookupLimit` is zero,
   matching `src/datasource.ts:60-62`.
3. **`Validate`** — enforce the runtime contract: URL is required (backend
   fails otherwise); `TSDBVersion` and `TSDBResolution` must belong to their
   enum sets (zero accepted because callers may call `Validate` before
   `ApplyDefaults`); `LookupLimit` and `Timeout` must be non-negative;
   Basic auth requires a username; mTLS requires serverName + client cert
   + client key; custom-CA requires the CA PEM. Errors are joined so every
   problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with
`datasource_uid`, `datasource_name`, and `plugin` labels so log lines carry
request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are still
exported for callers that want to compose them themselves (e.g. provisioning
preview, schema-example round-trip, tests that need to distinguish
parse-level from policy-level errors). Skip `LoadConfig` in those flows —
assemble a `Config` directly.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while
researching upstream. All preserved verbatim in the schema — the schema
records what the plugin **does**, not what it **should** do; these notes
exist so reviewers can reproduce each finding and decide separately whether
to fix upstream.

1. **`DataSourceHttpSettings` is deprecated but still used**: the component's
   JSDoc explicitly says to use `@grafana/plugin-ui` instead; every modern
   datasource plugin has migrated. OpenTSDB has not. Migrating would
   introduce a virtual auth selector, an "Additional settings" collapse
   section, and a URL-format validator — and would change the storage of
   `withCredentials`.
2. **`this.basicAuth` used as an Authorization header value**:
   `src/datasource.ts:41` types `basicAuth: string` and `:265-270` uses it
   as the value of the `Authorization` HTTP header. But `basicAuth` at
   root is universally a **boolean** flag in Grafana. This code path is
   effectively dead — for a normal datasource `this.basicAuth` will be
   `true`/`false`, and setting `Authorization: true` on outgoing requests
   is harmless (the SDK's proxy layer already handles Basic auth). Preserve
   as an upstream oddity; do not model `basicAuth` as a string.
3. **`lookupLimit` may be persisted as a string on first edit**:
   `OpenTsdbDetails.tsx:54-62` uses `<Input type="number" onChange={onInputChangeHandler(...)}>`,
   and `onInputChangeHandler` (`OpenTsdbDetails.tsx:79-88`) stores
   `event.currentTarget.value` — which is always a JavaScript string, even
   for `type="number"` inputs. The backend expects `int32`
   (`pkg/opentsdb/opentsdb.go:68`), so `json.Unmarshal` will fail on a
   stringified value. In practice this surfaces as an error at query time
   after any config edit that touched the field. Our schema declares the
   field as `valueType: "number"`; consumers writing provisioning payloads
   should pass a number, not a string.
4. **`tsdbVersion` typed as `float32` in the backend**: `types.ts:36`
   declares `tsdbVersion: number` (unbounded) while
   `pkg/opentsdb/opentsdb.go:66` uses `float32`. The Select only offers
   integer values (1..4), so this is safe today, but persisting a
   fractional value via provisioning would round-trip through
   `float32` without error — and only version `4` triggers the
   array-response parser (`utils.go:138`), so values like `3.999` would
   silently fall into the pre-2.4 code path.
5. **`tsdbResolution` has a runtime effect but no backend read that
   distinguishes it**: the backend `datasourceInfo` caches
   `TSDBResolution int32` (`pkg/opentsdb/opentsdb.go:59`), but the value
   is never inspected server-side. The distinction (second vs millisecond)
   is only enforced by the frontend at `src/datasource.ts:178-180` when
   building outgoing query bodies. Storing an incorrect resolution would
   produce misaligned timestamps on the frontend but no backend error.
6. **`ConfigEditor.tsx` has no `componentDidMount` defaulting**: unlike
   Graphite (which auto-writes `graphiteVersion` on mount), the OpenTSDB
   editor never writes defaults to jsonData when opened. A provisioned
   datasource may sit indefinitely with `{}` in jsonData, and `LoadConfig`
   must be prepared for zero values on all three OpenTSDB knobs (which is
   why `ApplyDefaults` fills them in).
7. **`CheckHealth` requires `q=cpu&type=metrics` to return HTTP 200**:
   `pkg/opentsdb/opentsdb.go:100-131` hard-codes the probe as a query for
   the metric `cpu`. On an OpenTSDB with no ingested `cpu` metric the
   probe still passes (the endpoint returns `[]` with status 200), but a
   deployment that rejects unknown-metric suggests would fail health
   without an actionable message.
8. **Lookup path uses hard-coded `limit=1000` in `HandleKeyLookup`**:
   `pkg/opentsdb/callresource.go:242` writes `lookupQueryParams.Set("limit", "1000")`
   ignoring `dsInfo.LookupLimit`. Only `HandleKeyValueLookup` (`:361`)
   honors the configured limit. So `lookupLimit` only takes effect for the
   `type=keyvalue` autocomplete path.
9. **`SecureSocksProxySettings` also writes to jsonData**: the excluded
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
  clean; no regressions in the 14 pre-existing entries.
- `settings.ts`: exports the three canonical types (`RootConfig`,
  `JsonDataConfig`, `SecureJsonDataConfig`) plus the two numeric enums
  (`OpenTsdbVersion`, `OpenTsdbResolution`) — reviewed by hand against the
  frontend sources; no `tsc` runner is wired into the registry module.
