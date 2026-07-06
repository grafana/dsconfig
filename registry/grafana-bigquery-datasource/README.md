# grafana-bigquery-datasource

Declarative configuration schema for the [Google BigQuery datasource plugin](https://github.com/grafana/google-bigquery-datasource) (`grafana-bigquery-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/google-bigquery-datasource`
- **Ref**: `main`
- **Commit SHA**: `8c658f97bc62e6c63a618dcf450ad504130a81c5` (2026-06-30, `Docs: BigQuery quarterly update (#496)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, option
labels/values, section titles, help markdown, defaults, validations, dependency and
required-when expressions, storage keys, storage targets, value types, group titles, and
instructions — is traceable to a specific `file:line` in the upstream repo at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone https://github.com/grafana/google-bigquery-datasource
cd google-bigquery-datasource
git checkout 8c658f97bc62e6c63a618dcf450ad504130a81c5
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under
[Sources researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, effects, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `SecureJsonDataKey` typed constants, `AuthType` / `QueryPriority` enums, `LoadConfig` / `ApplyDefaults` / `Validate` |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth variant + additional-settings knobs |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`8c658f97bc62e6c63a618dcf450ad504130a81c5`), plus external editor components at the exact
versions the plugin's `package.json` pins.

### Plugin repo (`github.com/grafana/google-bigquery-datasource@8c658f97`)

| File | What was read |
| --- | --- |
| `src/plugin.json:3-5` | `pluginType` (`id`), `pluginName` (`name`) |
| `src/plugin.json:37-50` | `routes[]` — deprecated Grafana-proxy definition; the modern backend consumes the same settings via `pkg/bigquery/http_client.go` |
| `src/components/ConfigEditor.tsx:1-19` | Imports; `AuthConfig` from `@grafana/google-sdk` renders the entire auth panel |
| `src/components/ConfigEditor.tsx:25-33` | `onMaxBytesBilledChange` — casts the raw string to `Number` before writing `MaxBytesBilled` |
| `src/components/ConfigEditor.tsx:35-36` | `showServiceAccountImpersonation` — true only when auth is JWT or GCE; drives whether AuthConfig renders its impersonation section |
| `src/components/ConfigEditor.tsx:40-44` | `DataSourceDescription` (`hasRequiredFields={false}`) — why no `required` marks in editor |
| `src/components/ConfigEditor.tsx:48` | `ConfigurationHelp` — top-level "How to configure Google BigQuery datasource?" collapsible |
| `src/components/ConfigEditor.tsx:52-57` | `AuthConfig` invocation with `authOptions={bigQueryAuthTypes}` and `showServiceAccountImpersonationConfig={showServiceAccountImpersonation}` |
| `src/components/ConfigEditor.tsx:61-138` | "Additional Settings" `ConfigSection` (`isCollapsible`) with `processingLocation`, `serviceEndpoint`, `MaxBytesBilled` |
| `src/components/ConfigEditor.tsx:62-85` | `Processing location` field: `Combobox`, description linking to https://cloud.google.com/bigquery/docs/locations, placeholder `"Automatic location selection"`, options list `PROCESSING_LOCATIONS` |
| `src/components/ConfigEditor.tsx:86-109` | `Service endpoint` field: description linking to https://cloud.google.com/bigquery/docs/reference/rest#service-endpoint, placeholder `"Optional, example https://bigquery.googleapis.com/bigquery/v2/"` |
| `src/components/ConfigEditor.tsx:110-133` | `Max bytes billed` field: description linking to https://cloud.google.com/bigquery/docs/best-practices-costs, placeholder `"Optional, example 5242880"`, `type={'number'}` |
| `src/components/ConfigEditor.tsx:135-137` | Conditional `SecureSocksProxySettings` — deliberately excluded from this entry per AGENTS.md |
| `src/types.ts:34-42` | `BigQueryOptions extends DataSourceOptions`: `flatRateProject`, `processingLocation`, `queryPriority`, `enableSecureSocksProxy`, `MaxBytesBilled`, `serviceEndpoint`, `oauthPassThru` |
| `src/types.ts:44-48` | `bigQueryAuthTypes` composition: `GOOGLE_AUTH_TYPE_OPTIONS` (JWT + GCE) + `OAUTH_PASSTHROUGH_AUTH_TYPE_OPTION` + WIF (only when `isCloud()`) |
| `src/types.ts:50` | `BigQuerySecureJsonData extends DataSourceSecureJsonData` — inherits `privateKey`, adds nothing |
| `src/constants.ts:11-65` | `PROCESSING_LOCATIONS`: 34 GCP regions plus `''` (Automatic location selection), `US`, and `EU` multi-regionals — 41 total entries |
| `src/utils.ts:323-327` | `isCloud()` — returns `true` when `config.namespace` starts with `stacks-`, gating WIF visibility |
| `src/components/ConfigurationHelp.tsx:9-53` | The top-level "How to configure Google BigQuery datasource?" `Collapse` markdown, captured verbatim in the `help` drawer of `jsonData_authenticationType` |
| `pkg/bigquery/settings.go:22-39` | `loadSettings`: json-unmarshal jsonData, call `utils.GetPrivateKey`, stamp `DatasourceId`/`Updated` |
| `pkg/bigquery/types/types.go:9-32` | `BigQuerySettings` struct fields and json tags (drives our `Config` fields verbatim) |
| `pkg/bigquery/http_client.go:41-92` | `getMiddleware`: switches on `authenticationType`, dispatches to GCE / ForwardOAuth / WIF / JWT token providers, with optional service-account impersonation |
| `pkg/bigquery/http_client.go:94-115` | `newHTTPClient`: WIF pool-provider validation (line 95), `oauthPassThru` handling (line 99-107) — sets `ForwardHTTPHeaders` and `Accept-Encoding: identity` |
| `pkg/bigquery/http_client.go:117-121` | `validateDataSourceSettings`: `DefaultProject`, `ClientEmail`, `PrivateKey`, `TokenUri` all required for JWT — encoded as `requiredWhen` in our schema |
| `pkg/bigquery/utils/auth.go:9-25` | `JWTConfigFromDataSourceSettings` — how the JWT settings feed the OAuth2 config (BigQuery + Drive + cloud-platform scopes) |

### External editor components

Read at the exact versions pinned in the plugin's `package.json`. Sources checked out at the
corresponding upstream refs.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `AuthConfig` | `@grafana/google-sdk@0.6.0` | `github.com/grafana/grafana-google-sdk-react`, `src/components/AuthConfig.tsx` | "Authentication type" `RadioButtonGroup` label (`:103`), default `GoogleAuthType.JWT` (`:40-48`), `oauthPassThru` side-effect on WIF/OAuth (`:66-78`), `Default project` GCE input (`:151-158`), impersonation UI (`:170-206`) rendered only when `showServiceAccountImpersonationConfig={true}` |
| `JWTForm` | `@grafana/google-sdk@0.6.0` | `grafana-google-sdk-react`, `src/components/JWTForm.tsx` | Field labels (`Project ID`, `Client email`, `Token URI`, `Private key path`, `Private key`), placeholders (`Enter Private key` `:67`, `File location of your private key (e.g. /etc/secrets/gce.pem)` `:109`) |
| `WIFConfigEditor` | `@grafana/google-sdk@0.6.0` | `grafana-google-sdk-react`, `src/components/WIFConfigEditor.tsx` | `Workload Identity Pool Provider` field label + description + placeholder (`:18-30`), `Service account email` field label + description + placeholder (`:32-44`), `Default project` field (`:46-53`) |
| `OAuthPassthroughConfigEditor` | `@grafana/google-sdk@0.6.0` | `grafana-google-sdk-react`, `src/components/OAuthPassthroughConfigEditor.tsx` | Confirms no additional user-supplied fields for `forwardOAuthIdentity` — the panel is informational only |
| `GOOGLE_AUTH_TYPE_OPTIONS`, `OAUTH_PASSTHROUGH_AUTH_TYPE_OPTION`, `WIF_AUTH_TYPE_OPTION` | same | `grafana-google-sdk-react`, `src/constants.ts:4-27` | The label/value pairs used to compose `bigQueryAuthTypes` — `Google JWT File`, `GCE Default Service Account`, `Forward OAuth Identity`, `Workload Identity Federation` |
| `DataSourceOptions`, `DataSourceSecureJsonData`, `GoogleAuthType` | same | `grafana-google-sdk-react`, `src/types.ts:3-25` | Base interfaces the plugin's TS types extend; discriminator values `jwt` / `gce` / `workloadIdentityFederation` / `forwardOAuthIdentity` |
| `GetPrivateKey` (backend) | `grafana-google-sdk-go` | `github.com/grafana/grafana-google-sdk-go`, `pkg/utils/utils.go:62-89` | Reads `privateKey` from a file when `privateKeyPath` is set (accepts raw PEM or a service-account JSON with a `private_key` field), otherwise reads `settings.DecryptedSecureJSONData["privateKey"]` and normalizes `\\n` → `\n` |
| `tokenprovider.NewJwtAccessTokenProvider`, `NewGceAccessTokenProvider`, `NewImpersonatedJwtAccessTokenProvider`, `NewImpersonatedGceAccessTokenProvider`, `AuthMiddleware` | `grafana-google-sdk-go` | `pkg/tokenprovider/` | The Google-token-provider stack `pkg/bigquery/http_client.go:41-92` dispatches into |
| `DataSourceDescription`, `ConfigSection` | `@grafana/plugin-ui@0.13.1` | `github.com/grafana/plugin-ui` — introspected via prop shape only | `ConfigEditor.tsx:40-44` and `:61` |
| `Combobox`, `Field`, `Input`, `SecureSocksProxySettings` | `@grafana/ui@13.1.0` | grafana/grafana `packages/grafana-ui/src/components/` | Prop names (`label`, `description`, `value`, `onChange`, `placeholder`, `options`, `width`, `type`, `className`) |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each of its
label, placeholder, tooltip, default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_authenticationType` | `authenticationType` | `jsonData` | `AuthConfig.tsx:103` (`<Field label="Authentication type">`) | Options: `GOOGLE_AUTH_TYPE_OPTIONS` (`constants.ts:4-15` in google-sdk-react) + `OAUTH_PASSTHROUGH_AUTH_TYPE_OPTION` (`:23-27`) + `WIF_AUTH_TYPE_OPTION` (`:17-21`); composed at `types.ts:44-48` (WIF only when `isCloud()` is true); default `jwt` from `AuthConfig.tsx:40-48` `useEffect` | `Settings.AuthenticationType string`, `pkg/bigquery/types/types.go:19` | Role `auth.discriminator`; help drawer verbatim from `ConfigurationHelp.tsx:9-53` |
| `jsonData_defaultProject` | `defaultProject` | `jsonData` | `AuthConfig.tsx:151` (`<Field label="Default project">`) for GCE; `JWTForm.tsx:76` (`<Field label="Project ID">`) for JWT; `WIFConfigEditor.tsx:46` for WIF | Populated from uploaded JWT's `project_id` at `AuthConfig.tsx:140`; user input otherwise | `Settings.DefaultProject string`, `pkg/bigquery/types/types.go:12` | Required for JWT (`http_client.go:118`); optional for GCE/WIF |
| `jsonData_clientEmail` | `clientEmail` | `jsonData` | `JWTForm.tsx:85` (`<Field label="Client email">`) | Populated from uploaded JWT's `client_email` at `AuthConfig.tsx:139` | `Settings.ClientEmail string`, `types.go:11` | `dependsOn: authenticationType == 'jwt'`; `requiredWhen: ...jwt && privateKeyPath == ''` |
| `jsonData_tokenUri` | `tokenUri` | `jsonData` | `JWTForm.tsx:94` (`<Field label="Token URI">`) | Populated from uploaded JWT's `token_uri` at `AuthConfig.tsx:141` | `Settings.TokenUri string`, `types.go:14` | Also referenced by the deprecated `plugin.json:44` route |
| `jsonData_privateKeyPath` | `privateKeyPath` | `jsonData` | `JWTForm.tsx:104` (`<Field label="Private key path" description={Description}>`) — description at `:44-62` "Paste private key or provide path to private key file" | `JWTForm.tsx:109` (`placeholder="File location of your private key (e.g. /etc/secrets/gce.pem)"`) | `Settings.PrivateKeyPath string`, `types.go:20`; consumer `grafana-google-sdk-go/pkg/utils/utils.go:62-80` |
| `secureJsonData_privateKey` | `privateKey` | `secureJsonData` | `JWTForm.tsx:117` (`<Field label="Private key" description={Description}>`) | `JWTForm.tsx:67` (`placeholder: 'Enter Private key'`) | `Settings.PrivateKey string`, `types.go:31` (`json:"-"` — decrypted separately) | Role `auth.jwt.signingKey`; required for JWT unless `privateKeyPath` is set |
| `jsonData_usingImpersonation` | `usingImpersonation` | `jsonData` | `AuthConfig.tsx:173` (`<Field label="Enable" ...>`) | AuthConfig only renders it when caller passes `showServiceAccountImpersonationConfig={true}`; the plugin passes it for `jwt` and `gce` only (`ConfigEditor.tsx:35-36,56`) | `Settings.UsingImpersonation bool`, `types.go:22` | Backend consumer at `http_client.go:53,73` |
| `jsonData_serviceAccountToImpersonate` | `serviceAccountToImpersonate` | `jsonData` | `AuthConfig.tsx:196` (`<Field label="Service account to impersonate" ...>`) | Rendered inside AuthConfig only when `usingImpersonation` is true (`:195`) | `Settings.ServiceAccountToImpersonate string`, `types.go:23` | Backend consumer at `http_client.go:54,74` |
| `jsonData_workloadIdentityPoolProvider` | `workloadIdentityPoolProvider` | `jsonData` | `WIFConfigEditor.tsx:19` (`<Field label="Workload Identity Pool Provider" description="Full resource name…">`) | `WIFConfigEditor.tsx:26` (`placeholder="projects/<number>/…/providers/<provider>"`) | `Settings.WorkloadIdentityPoolProvider string`, `types.go:27` | Backend validates non-empty for WIF at `http_client.go:95` |
| `jsonData_wifServiceAccountEmail` | `wifServiceAccountEmail` | `jsonData` | `WIFConfigEditor.tsx:33` (`<Field label="Service account email" description="Optional…">`) | `WIFConfigEditor.tsx:40` (`placeholder="name@project.iam.gserviceaccount.com"`) | `Settings.WifServiceAccountEmail string`, `types.go:28` | Optional impersonation for WIF |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (no direct UI; side-effect only) | Set by `AuthConfig.tsx:73-74` to `true` when auth is `forwardOAuthIdentity` or `workloadIdentityFederation`, `false` otherwise | `Settings.OAuthPassthroughEnabled bool`, `types.go:24`; TS `oauthPassThru?: boolean`, `src/types.ts:41` | Tagged `editor-managed`; backend consumer at `http_client.go:99-107` |
| `jsonData_processingLocation` | `processingLocation` | `jsonData` | `ConfigEditor.tsx:63` (`<Field label="Processing location" description={<a href="https://cloud.google.com/bigquery/docs/locations">…}>`) | `ConfigEditor.tsx:80` (`placeholder="Automatic location selection"`); options `constants.ts:11-65` (41 entries) | `Settings.ProcessingLocation string`, `types.go:16` | Default `""` = "Automatic location selection" (the first option) |
| `jsonData_serviceEndpoint` | `serviceEndpoint` | `jsonData` | `ConfigEditor.tsx:87` (`<Field label="Service endpoint" description={<a href="https://cloud.google.com/bigquery/docs/reference/rest#service-endpoint">…}>`) | `ConfigEditor.tsx:104` (`placeholder="Optional, example https://bigquery.googleapis.com/bigquery/v2/"`) | `Settings.ServiceEndpoint string`, `types.go:21` | Role `endpoint.baseUrl` |
| `jsonData_MaxBytesBilled` | `MaxBytesBilled` | `jsonData` | `ConfigEditor.tsx:111` (`<Field label="Max bytes billed" description={<a href="https://cloud.google.com/bigquery/docs/best-practices-costs">…}>`) | `ConfigEditor.tsx:128` (`placeholder="Optional, example 5242880"`); `type={'number'}` at `:129` | `Settings.MaxBytesBilled int64`, `types.go:17` (`omitempty`) | Case-preserved key matches the backend json tag; frontend casts to `Number` at `:30` |
| `jsonData_flatRateProject` | `flatRateProject` | `jsonData` | — (no UI) | — | `Settings.FlatRateProject string`, `types.go:13`; TS `flatRateProject?: string`, `src/types.ts:35` | Tagged `backend-only, unused`; not read by any code path |
| `jsonData_queryPriority` | `queryPriority` | `jsonData` | — (no UI at datasource level; a `queryPriority` also exists on the query at `src/types.ts:107`, but that is a separate storage location) | Values `INTERACTIVE` / `BATCH` from `QueryPriority` enum at `src/types.ts:22-25` | `Settings.QueryPriority string`, `types.go:15`; TS `queryPriority?: QueryPriority`, `src/types.ts:37` | Tagged `backend-only, unused` |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_authenticationType` | `authenticationType` | `jsonData` | Authentication type | Yes (backend discriminator) |
| `jsonData_defaultProject` | `defaultProject` | `jsonData` | Default project / Project ID | Yes |
| `jsonData_clientEmail` | `clientEmail` | `jsonData` | Client email | Yes (JWT) |
| `jsonData_tokenUri` | `tokenUri` | `jsonData` | Token URI | Yes (JWT) |
| `jsonData_privateKeyPath` | `privateKeyPath` | `jsonData` | Private key path | Yes (`grafana-google-sdk-go` `GetPrivateKey`) |
| `secureJsonData_privateKey` | `privateKey` | `secureJsonData` | Private key | Yes (`grafana-google-sdk-go` `GetPrivateKey`) |
| `jsonData_usingImpersonation` | `usingImpersonation` | `jsonData` | Enable (impersonation) | Yes (JWT / GCE branches) |
| `jsonData_serviceAccountToImpersonate` | `serviceAccountToImpersonate` | `jsonData` | Service account to impersonate | Yes |
| `jsonData_workloadIdentityPoolProvider` | `workloadIdentityPoolProvider` | `jsonData` | Workload Identity Pool Provider | Yes (WIF only) |
| `jsonData_wifServiceAccountEmail` | `wifServiceAccountEmail` | `jsonData` | Service account email (WIF) | Yes (optional impersonation) |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (editor-managed side-effect) | Yes (gates `ForwardHTTPHeaders`) |
| `jsonData_processingLocation` | `processingLocation` | `jsonData` | Processing location | Yes |
| `jsonData_serviceEndpoint` | `serviceEndpoint` | `jsonData` | Service endpoint | Yes |
| `jsonData_MaxBytesBilled` | `MaxBytesBilled` | `jsonData` | Max bytes billed | Yes |
| `jsonData_flatRateProject` | `flatRateProject` | `jsonData` | — | Loaded but never used |
| `jsonData_queryPriority` | `queryPriority` | `jsonData` | — | Loaded but never used |

## Where the types are defined

Configuration types are spread across the plugin, the shared `@grafana/google-sdk` React
package, and its Go counterpart. Some fields come from libraries rather than the plugin itself.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `BigQueryOptions`, `BigQuerySecureJsonData`, `bigQueryAuthTypes`, `QueryPriority`, `isCloud` | `src/types.ts:22-50`, `src/utils.ts:323` | plugin ([grafana/google-bigquery-datasource](https://github.com/grafana/google-bigquery-datasource)) |
| `PROCESSING_LOCATIONS`, `QUERY_PRIORITIES` | `src/constants.ts:11-70` | plugin |
| `GoogleAuthType`, `DataSourceOptions`, `DataSourceSecureJsonData`, `GOOGLE_AUTH_TYPE_OPTIONS`, `WIF_AUTH_TYPE_OPTION`, `OAUTH_PASSTHROUGH_AUTH_TYPE_OPTION` | `src/types.ts:3-25`, `src/constants.ts:4-27` | `@grafana/google-sdk` `0.6.0` ([grafana/grafana-google-sdk-react](https://github.com/grafana/grafana-google-sdk-react)) |
| `AuthConfig`, `JWTForm`, `JWTConfigEditor`, `WIFConfigEditor`, `OAuthPassthroughConfigEditor` | `src/components/` | `@grafana/google-sdk` `0.6.0` |
| `DataSourceDescription`, `ConfigSection` | `packages/plugin-ui/src/` | `@grafana/plugin-ui` `0.13.1` |
| `Combobox`, `Field`, `Input`, `SecureSocksProxySettings` | `packages/grafana-ui/src/components/` | `@grafana/ui` `13.1.0` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `BigQuerySettings` (flat parsed jsonData shape + decrypted secret staging), `Credentials`, `loadSettings`, `getConnectionSettings` | `pkg/bigquery/settings.go:14-67`, `pkg/bigquery/types/types.go:9-32` | plugin |
| `getMiddleware`, `newHTTPClient`, `validateDataSourceSettings` (dispatch on `AuthenticationType`, apply middleware, gate `oauthPassThru`) | `pkg/bigquery/http_client.go:41-121` | plugin |
| `JWTConfigFromDataSourceSettings` (JWT config with BigQuery + Drive + cloud-platform scopes) | `pkg/bigquery/utils/auth.go:9-25` | plugin |
| `GetPrivateKey` (reads `privateKey` from `privateKeyPath` when set, else from decrypted secure JSON) | `pkg/utils/utils.go:62-89` | `github.com/grafana/grafana-google-sdk-go` |
| `tokenprovider.NewJwtAccessTokenProvider`, `NewGceAccessTokenProvider`, `NewImpersonatedJwtAccessTokenProvider`, `NewImpersonatedGceAccessTokenProvider`, `AuthMiddleware` | `pkg/tokenprovider/` | `github.com/grafana/grafana-google-sdk-go` |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` |

The models in this entry flatten that spread into a single Go `Config` type (jsonData fields +
`DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant list. `settings.ts` keeps
the three canonical TypeScript types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`).

## Modeling decisions

- **All four auth types modeled unconditionally**: the plugin composes
  `bigQueryAuthTypes` at `src/types.ts:44-48` from JWT + GCE + Forward-OAuth, plus WIF only
  when `isCloud()` returns true. The `isCloud` gate is a UI concern (only Cloud stacks see
  the WIF button); the backend at `pkg/bigquery/http_client.go:51-89` accepts all four
  values regardless of deployment. The schema therefore lists all four in the
  `allowedValues` set.
- **`oauthPassThru` is editor-managed, not user-driven**: `AuthConfig.tsx:73-74` sets it as
  a side-effect of the `Authentication type` radio, so we model it in the schema (backend
  reads it at `http_client.go:99-107`) but tag it `editor-managed` and provide no UI. The
  Go `ApplyDefaults` mirrors that behaviour: it is derived from `AuthenticationType`.
- **Impersonation only for JWT / GCE**: `AuthConfig` renders its impersonation section only
  when the caller passes `showServiceAccountImpersonationConfig={true}`, and the plugin
  gates that on `authenticationType === 'jwt' || 'gce'` (`ConfigEditor.tsx:35-36`). The
  schema encodes both conditions in `dependsOn`.
- **WIF pool provider validated on backend**: `http_client.go:95` fails the request if
  `authenticationType === 'workloadIdentityFederation' && workloadIdentityPoolProvider ==
  ''`. Captured as `requiredWhen` and enforced in Go `Validate`.
- **`flatRateProject` and `queryPriority` modeled but tagged unused**: both are declared
  in the backend Settings struct (`types.go:13,15`) and the TS options
  (`src/types.ts:35,37`), but no runtime code path reads them (the query-time
  `queryPriority` at `src/types.ts:107` is on the query object, not the datasource). Kept
  in the schema so provisioning payloads are validated instead of silently accepting the
  keys, and tagged `backend-only, unused` so consumers know not to populate them for new
  datasources.
- **`processingLocation` full option list preserved**: 41 entries (empty-string automatic +
  US + EU multi-regionals + 34 regional) from `src/constants.ts:11-65`, verbatim.
- **`MaxBytesBilled` case preserved**: backend json tag is `"MaxBytesBilled,omitempty"`
  (leading capital, `types.go:17`), so both the schema `key` and Go field mirror the
  uppercase form. The editor casts to `Number` before storage (`ConfigEditor.tsx:30`).
- **Secure Socks Proxy excluded**: the editor conditionally renders `SecureSocksProxySettings`
  (`ConfigEditor.tsx:135-137`) writing `jsonData.enableSecureSocksProxy` when the Grafana
  instance has `secureSocksDSProxyEnabled`. The field is deliberately omitted from this
  registry entry per AGENTS.md.
- **Flat `Config` in Go**: `settings.go` mirrors the plugin's `pkg/bigquery/types/types.go`
  `BigQuerySettings` (minus the SDK back-references `DatasourceId` / `Updated` and the
  transient decrypted `PrivateKey`) with typed `AuthType`, `QueryPriority`, and
  `SecureJsonDataKey` enums. `settings.ts` keeps the three canonical TS types.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle from
the embedded `dsconfig.json`: root fields plus a nested `jsonData` object become the OpenAPI
settings `spec`, secure fields become `secureValues`, and virtual fields are skipped.

`SettingsExamples()` provides the default configuration plus one k8s-style example per
authentication type and connection variant. Each example is a full instance-settings object
with the plugin configuration nested under `jsonData` and the relevant write-only secrets
under `secureJsonData` (placeholder values — replace with real secrets; the default example
— keyed by the empty string `""` — carries an empty `privateKey` to show what must be filled
in):

| Example | Auth | Notes | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | JWT (schema defaults) | Must still supply `defaultProject`, `clientEmail`, `tokenUri`, and a private key | `privateKey` (empty) |
| `googleJWTFile` | JWT | Inline `privateKey` in secureJsonData | `privateKey` |
| `googleJWTFilePath` | JWT | Private key from `privateKeyPath` file on the Grafana server | `privateKey` (empty — supplied by file) |
| `gceDefaultServiceAccount` | GCE Default Service Account | Only works on a GCE VM | (none) |
| `forwardOAuthIdentity` | Forward OAuth Identity | Caller's OAuth token is forwarded; `oauthPassThru` set to `true` | (none) |
| `workloadIdentityFederation` | Workload Identity Federation | Requires `workloadIdentityPoolProvider`; only exposed in the editor for Cloud stacks | (none) |
| `impersonation` | JWT + service account impersonation | Base SA impersonates another SA (`usingImpersonation`, `serviceAccountToImpersonate`) | `privateKey` |
| `additionalSettings` | JWT + all Additional Settings knobs | `processingLocation`, `serviceEndpoint`, `MaxBytesBilled` populated | `privateKey` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns a
fully-defaulted, validated `Config`:

1. **Parse** — unmarshal jsonData into `Config` (mirrors `pkg/bigquery/settings.go:22-39`)
   and copy decrypted secrets into `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — fill a curated set of zero-valued discriminators with the same
   defaults the editor writes for a fresh datasource:
   - `AuthenticationType=AuthTypeJWT` (matches `AuthConfig.tsx:40-48`).
   - `OAuthPassthroughEnabled` derived from `AuthenticationType` (matches
     `AuthConfig.tsx:73-74`: true for `forwardOAuthIdentity` / `workloadIdentityFederation`,
     false otherwise).
3. **`Validate`** — enforce the runtime contract (auth method + its required inputs, WIF
   pool provider, `MaxBytesBilled` non-negative, `queryPriority` in the allowed set).
   Errors are joined so every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels so log lines carry request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are exported separately for
callers that want to compose them themselves (e.g. provisioning preview, schema-example
round-trip, tests that need to distinguish parse-level from policy-level errors).

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching upstream.
All preserved as-is in the schema — the schema records what the plugin **does**, not what it
**should** do; these notes exist so reviewers can reproduce each finding and decide
separately whether to fix upstream.

1. **`flatRateProject` is dead weight.** Declared in `src/types.ts:35` and
   `pkg/bigquery/types/types.go:13`, but no code path reads it. Persists in provisioned
   configs indefinitely.
2. **Datasource-level `queryPriority` is dead weight.** Same story — `types.go:15` and
   `src/types.ts:37` define it, but nothing reads it. A separate `queryPriority` exists on
   individual queries (`src/types.ts:107`) which IS consumed; the two share the name but
   not the storage location.
3. **`plugin.json:37-50` route is stale.** The `routes[]` block declares a `jwtTokenAuth`
   route to `https://www.googleapis.com/bigquery`, but the modern backend builds its own
   HTTP client (`pkg/bigquery/http_client.go`) and does not consume the Grafana-proxy route.
   Kept in `plugin.json` for backward compatibility.
4. **`oauthPassThru` can drift from `authenticationType`.** `AuthConfig.tsx:73-74` writes
   `oauthPassThru` as a side-effect only when the radio changes. A provisioning payload
   that sets `authenticationType: "jwt"` and `oauthPassThru: true` bypasses that logic —
   the backend will honour `oauthPassThru` and take the `ForwardHTTPHeaders` branch at
   `http_client.go:99-107` before ever reaching the JWT path. Our `ApplyDefaults` fixes
   this by always deriving `OAuthPassthroughEnabled` from `AuthenticationType`.
5. **WIF button visibility gated only in the UI.** `isCloud()` (`src/utils.ts:323-327`)
   returns `true` when `config.namespace` starts with `stacks-`. On-prem editors will not
   see the WIF radio, but the backend (`http_client.go:51-89`) accepts
   `workloadIdentityFederation` regardless — a provisioning API caller can enable it on any
   Grafana instance. Not a bug per se, but a capability boundary worth flagging.
6. **`Trailing slash tolerance on `serviceEndpoint`` is untested.** The placeholder
   suggests `https://bigquery.googleapis.com/bigquery/v2/` with a trailing slash; whether
   the Google client accepts that verbatim depends on the underlying library and is not
   validated by the plugin at save time.
7. **Impersonation guard is asymmetric with the SDK.** `showServiceAccountImpersonation`
   (`ConfigEditor.tsx:35-36`) is `true` for `jwt` and `gce`, but AuthConfig internally
   also checks `authenticationType !== GoogleAuthType.ForwardOAuthIdentity`
   (`AuthConfig.tsx:170`), which for BigQuery is a stricter subset. A provisioning payload
   with `authenticationType: "forwardOAuthIdentity"` and `usingImpersonation: true` is
   accepted by the backend but doesn't actually do impersonation — the code branch is not
   reached.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) —
  passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft-07, `additionalProperties: false`) — passes.
- `go test -race ./...` on the shared `registry/` module — passes for this entry (schema
  bundle shape, secure values, examples, `LoadConfig` incl. all four auth branches and the
  inline-vs-path private-key choice, `SchemaArtifactInSync` guard).
- `settings.go`/`schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: `tsc --noEmit --strict` — clean.
