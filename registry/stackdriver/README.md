# stackdriver

Declarative configuration schema for the [Google Cloud Monitoring datasource plugin](https://github.com/grafana/grafana-cloudmonitoring-datasource) (plugin id `stackdriver` — the legacy name has been kept for backward compatibility with datasources provisioned under the product's original name).

## Upstream researched

- **Repo**: `github.com/grafana/grafana-cloudmonitoring-datasource`
- **Ref**: `main`
- **Commit SHA**: `f3bea86eac3289eabf82a0e62f3de9ba1512790b` (2026-07-02, `docs: add signed commits requirement to CONTRIBUTING.md (#98)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, option
labels/values, section titles, help markdown, defaults, validations, dependency and
required-when expressions, storage keys, storage targets, value types, group titles, and
instructions — is traceable to a specific `file:line` in the upstream repo at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone https://github.com/grafana/grafana-cloudmonitoring-datasource
cd grafana-cloudmonitoring-datasource
git checkout f3bea86eac3289eabf82a0e62f3de9ba1512790b
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under
[Sources researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, effects, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `SecureJsonDataKey` typed constants, `AuthType` enum, `LoadConfig` / `ApplyDefaults` / `Validate` |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth variant + additional-settings knobs |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`f3bea86eac3289eabf82a0e62f3de9ba1512790b`), plus external editor components at the exact
versions the plugin's `package.json` pins.

### Plugin repo (`github.com/grafana/grafana-cloudmonitoring-datasource@f3bea86`)

| File | What was read |
| --- | --- |
| `src/plugin.json:2-38` | `pluginType` (`id` = `stackdriver`), `pluginName` (`name` = `Google Cloud Monitoring`), docs URL from `info.links[]` |
| `src/components/ConfigEditor/ConfigEditor.tsx:1-19` | Imports; `AuthConfig` from `@grafana/google-sdk` renders the entire auth panel |
| `src/components/ConfigEditor/ConfigEditor.tsx:20-30` | `handleOnOptionsChange` — records `reportInteraction` telemetry when JWT credentials are supplied (no storage side-effect) |
| `src/components/ConfigEditor/ConfigEditor.tsx:32-38` | `authOptions` composition: `GOOGLE_AUTH_TYPE_OPTIONS` (JWT + GCE) plus, when `isCloud()`, `WIF_AUTH_TYPE_OPTION` + `OAUTH_PASSTHROUGH_AUTH_TYPE_OPTION`. Order in the editor radio: JWT, GCE, WIF, ForwardOAuthIdentity |
| `src/components/ConfigEditor/ConfigEditor.tsx:40-42` | `showServiceAccountImpersonation` — true when auth is JWT or GCE; passed to `AuthConfig.showServiceAccountImpersonationConfig` |
| `src/components/ConfigEditor/ConfigEditor.tsx:46-51` | `DataSourceDescription` with `dataSourceName="Google Cloud Monitoring"`, `docsLink` to the Grafana docs, and `hasRequiredFields` |
| `src/components/ConfigEditor/ConfigEditor.tsx:52-57` | `AuthConfig` invocation |
| `src/components/ConfigEditor/ConfigEditor.tsx:58-72` | Info box about service-account docs (hidden for `forwardOAuthIdentity`) |
| `src/components/ConfigEditor/ConfigEditor.tsx:73-77` | `<Alert severity="info">Verify GCE default service account by clicking Save & Test</Alert>` for GCE |
| `src/components/ConfigEditor/ConfigEditor.tsx:78-108` | "Additional settings" `ConfigSection` — **only rendered when `config.secureSocksDSProxyEnabled` is true** (an upstream coupling quirk). Contains the `Universe Domain` input (placeholder `"googleapis.com"`, `noMargin`) and the `SecureSocksProxySettings` component |
| `src/types/types.ts:38-45` | `CloudMonitoringOptions extends DataSourceOptions`: adds `gceDefaultProject`, `enableSecureSocksProxy`, `universeDomain`, `oauthPassThru`. `CloudMonitoringSecureJsonData extends DataSourceSecureJsonData` — inherits `privateKey`, adds nothing |
| `src/utils.ts:15-17` | `isCloud()` — returns `true` when `config.namespace` starts with `stacks-`, gating WIF + Forward OAuth Identity visibility |
| `src/datasource.ts:44-45` | Frontend defaults `authenticationType` to `'jwt'` when reading — mirrors the backend default |
| `src/datasource.ts:173-191` | `gceDefaultProject`: fetched from `/gceDefaultProject` resource endpoint on demand and cached back into `instanceSettings.jsonData.gceDefaultProject`. Frontend-only runtime cache |
| `pkg/plugin/plugin.go` (via `Magefile.go`) | Entry point; datasource type constructed by `pkg/cloudmonitoring/cloudmonitoring.go:NewDatasource` |
| `pkg/cloudmonitoring/cloudmonitoring.go:52-63` | Backend constants `gceAuthentication`, `jwtAuthentication`, `forwardOAuthIdentityAuthentication`, `workloadIdentityFederationAuthentication` — the four allowed discriminator values |
| `pkg/cloudmonitoring/cloudmonitoring.go:203-215` | `datasourceJSONData` struct — the source of truth for the flat `Config` fields in `settings.go` |
| `pkg/cloudmonitoring/cloudmonitoring.go:222-276` | `newDatasourceInfo`: json-unmarshal jsonData, default empty `authenticationType` to `'jwt'` (`:229-231`), call `utils.GetPrivateKey`, set `ForwardHTTPHeaders = true` for token-forwarding auth (`:260-262`), populate `services[]` map with routed HTTP clients |
| `pkg/cloudmonitoring/cloudmonitoring.go:108-171` | `CheckHealth` — refuses to run when `oauthPassThru && defaultProject == ""` (`:121-125`) and produces auth-specific 401/403 messages |
| `pkg/cloudmonitoring/cloudmonitoring.go:397-431` | `QueryData` — rejects alerting queries under token-forwarding auth (`:412-418`) via `fromAlertHeaderName == "true"` |
| `pkg/cloudmonitoring/cloudmonitoring.go:659-675` | `ensureProject` / `getDefaultProject` — for GCE, always calls `gceDefaultProjectGetter(ctx, cloudMonitorScope)`, ignoring any cached `jsonData.gceDefaultProject` |
| `pkg/cloudmonitoring/httpclient.go:11-16` | Constants: two routes (`cloudmonitoring`, `cloudresourcemanager`) with corresponding scopes |
| `pkg/cloudmonitoring/httpclient.go:18-35` | `routes` map — each service has a base URL `https://<service>.` that `buildURL` combines with `universeDomain` |
| `pkg/cloudmonitoring/httpclient.go:37-77` | `getMiddleware`: dispatches on `authenticationType` — GCE / GCE+impersonation / JWT / JWT+impersonation. Token-forwarding auth returns `nil` middleware so Grafana forwards the caller's Authorization header verbatim |
| `pkg/cloudmonitoring/httpclient.go:79-83` | `buildURL(route, universeDomain)` — empty universeDomain falls back to `"googleapis.com"` |
| `pkg/cloudmonitoring/httpclient.go:86-99` | `newHTTPClient` — WIF pool-provider validation (`:87-89`); appends middleware when not token-forwarding |

### External editor components

Read at the exact versions pinned in the plugin's `package.json`. Sources checked out at the
corresponding upstream refs.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `AuthConfig` | `@grafana/google-sdk@0.6.0` | `github.com/grafana/grafana-google-sdk-react`, `src/components/AuthConfig.tsx` | "Authentication type" `RadioButtonGroup` label (`:103`), default `GoogleAuthType.JWT` (`:40-48`), `oauthPassThru` side-effect on WIF/OAuth (`:66-78`), `Default project` GCE input (`:151-158`), impersonation UI (`:170-206`) rendered only when `showServiceAccountImpersonationConfig={true}` (and never for `ForwardOAuthIdentity`) |
| `JWTForm` | `@grafana/google-sdk@0.6.0` | `grafana-google-sdk-react`, `src/components/JWTForm.tsx` | Field labels (`Project ID`, `Client email`, `Token URI`, `Private key path`, `Private key`), the toggling description linking between the two private-key modes (`:44-62`), placeholders (`Enter Private key` `:67`, `File location of your private key (e.g. /etc/secrets/gce.pem)` `:109`) |
| `JWTConfigEditor` | `@grafana/google-sdk@0.6.0` | `grafana-google-sdk-react`, `src/components/JWTConfigEditor.tsx` | Paste / upload UI for the JWT service-account JSON; populates `clientEmail`, `defaultProject` (from `project_id`), `tokenUri`, and `secureJsonData.privateKey` |
| `WIFConfigEditor` | `@grafana/google-sdk@0.6.0` | `grafana-google-sdk-react`, `src/components/WIFConfigEditor.tsx` | `Workload Identity Pool Provider` field label + description + placeholder (`:18-30`), `Service account email` field label + description + placeholder (`:32-44`), `Default project` field (`:46-53`) |
| `OAuthPassthroughConfigEditor` | `@grafana/google-sdk@0.6.0` | `grafana-google-sdk-react`, `src/components/OAuthPassthroughConfigEditor.tsx` | Renders the `Default project` input with description `"Required when forwarding the signed-in user's OAuth identity…"` (`:12-27`) — confirms the field is genuinely required by the runtime |
| `GOOGLE_AUTH_TYPE_OPTIONS`, `OAUTH_PASSTHROUGH_AUTH_TYPE_OPTION`, `WIF_AUTH_TYPE_OPTION` | same | `grafana-google-sdk-react`, `src/constants.ts:4-27` | The label/value pairs — `Google JWT File`, `GCE Default Service Account`, `Workload Identity Federation`, `Forward OAuth Identity` |
| `DataSourceOptions`, `DataSourceSecureJsonData`, `GoogleAuthType` | same | `grafana-google-sdk-react`, `src/types.ts:3-25` | Base interfaces the plugin's TS types extend; discriminator values `jwt` / `gce` / `workloadIdentityFederation` / `forwardOAuthIdentity` |
| `GetPrivateKey` (backend) | `grafana-google-sdk-go` | `github.com/grafana/grafana-google-sdk-go`, `pkg/utils/utils.go:62-89` | Reads `privateKey` from a file when `privateKeyPath` is set (accepts raw PEM or a service-account JSON with a `private_key` field), otherwise reads `settings.DecryptedSecureJSONData["privateKey"]` and normalizes escaped newlines |
| `tokenprovider.NewJwtAccessTokenProvider`, `NewGceAccessTokenProvider`, `NewImpersonatedJwtAccessTokenProvider`, `NewImpersonatedGceAccessTokenProvider`, `AuthMiddleware` | `grafana-google-sdk-go` | `pkg/tokenprovider/` | The Google-token-provider stack that `pkg/cloudmonitoring/httpclient.go:37-77` dispatches into |
| `DataSourceDescription`, `ConfigSection` | `@grafana/plugin-ui@0.13.1` | `github.com/grafana/plugin-ui` — introspected via prop shape only | `ConfigEditor.tsx:46-51` and `:81-88` |
| `Field`, `Input`, `Alert`, `Divider`, `Stack`, `SecureSocksProxySettings` | `@grafana/ui@13.1.0` | `grafana/grafana` `packages/grafana-ui/src/components/` | Prop names (`label`, `description`, `value`, `onChange`, `placeholder`, `width`, `noMargin`, `title`, `severity`) |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each of its
label, placeholder, tooltip, default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_authenticationType` | `authenticationType` | `jsonData` | `AuthConfig.tsx:103` (`<Field label="Authentication type">`) | Options: `GOOGLE_AUTH_TYPE_OPTIONS` (`constants.ts:4-15` in google-sdk-react) + `WIF_AUTH_TYPE_OPTION` (`:17-21`) + `OAUTH_PASSTHROUGH_AUTH_TYPE_OPTION` (`:23-27`); composed at `ConfigEditor.tsx:32-38` (WIF and Forward OAuth Identity only when `isCloud()` is true); default `jwt` from `AuthConfig.tsx:40-48` `useEffect`, echoed by backend `cloudmonitoring.go:229-231` | `datasourceJSONData.AuthenticationType string`, `cloudmonitoring.go:204` | Role `auth.discriminator`; help drawer with the "Don't know how to get…" link from `ConfigEditor.tsx:60-71` |
| `jsonData_defaultProject` | `defaultProject` | `jsonData` | `AuthConfig.tsx:151` (`<Field label="Default project">`) for GCE; `JWTForm.tsx:76` (`<Field label="Project ID">`) for JWT; `WIFConfigEditor.tsx:46` for WIF; `OAuthPassthroughConfigEditor.tsx:14` for Forward OAuth Identity | Populated from uploaded JWT's `project_id` at `AuthConfig.tsx:140`; user input otherwise. Placeholder `my-gcp-project` from `OAuthPassthroughConfigEditor.tsx:22` | `datasourceJSONData.DefaultProject string`, `cloudmonitoring.go:205` | Required for `forwardOAuthIdentity` / `workloadIdentityFederation` (backend CheckHealth at `cloudmonitoring.go:121-125` and the token-forwarding branches); optional for GCE (backend resolves via metadata server) |
| `jsonData_clientEmail` | `clientEmail` | `jsonData` | `JWTForm.tsx:85` (`<Field label="Client email">`) | Populated from uploaded JWT's `client_email` at `AuthConfig.tsx:139` | `datasourceJSONData.ClientEmail string`, `cloudmonitoring.go:206` | `dependsOn: authenticationType == 'jwt'`; `requiredWhen: ...jwt && privateKeyPath == ''` |
| `jsonData_tokenUri` | `tokenUri` | `jsonData` | `JWTForm.tsx:94` (`<Field label="Token URI">`) | Populated from uploaded JWT's `token_uri` at `AuthConfig.tsx:141` | `datasourceJSONData.TokenURI string`, `cloudmonitoring.go:207` | `dependsOn: authenticationType == 'jwt'`; `requiredWhen: ...jwt && privateKeyPath == ''` |
| `jsonData_privateKeyPath` | `privateKeyPath` | `jsonData` | `JWTForm.tsx:104` (`<Field label="Private key path" description={Description}>`) — description at `:44-62` toggling "Paste private key or provide path to private key file" | `JWTForm.tsx:109` (`placeholder="File location of your private key (e.g. /etc/secrets/gce.pem)"`) | Not on `datasourceJSONData` directly; consumed by `grafana-google-sdk-go/pkg/utils/utils.go:62-89` (`GetPrivateKey`) which reads from `settings.JSONData.privateKeyPath` |
| `secureJsonData_privateKey` | `privateKey` | `secureJsonData` | `JWTForm.tsx:117` (`<Field label="Private key" description={Description}>`) | `JWTForm.tsx:67` (`placeholder: 'Enter Private key'`) | Decrypted by SDK; consumed by `GetPrivateKey` in `pkg/utils/utils.go:62-89` | Role `auth.jwt.signingKey`; required for JWT unless `privateKeyPath` is set |
| `jsonData_usingImpersonation` | `usingImpersonation` | `jsonData` | `AuthConfig.tsx:173` (`<Field label="Enable" ...>`) | AuthConfig only renders it when caller passes `showServiceAccountImpersonationConfig={true}`; the plugin passes it for `jwt` and `gce` (`ConfigEditor.tsx:41-42,56`). AuthConfig additionally suppresses it for `ForwardOAuthIdentity` (`AuthConfig.tsx:170`), which is a no-op here | `datasourceJSONData.UsingImpersonation bool`, `cloudmonitoring.go:209` | Backend consumer at `httpclient.go:56,68` |
| `jsonData_serviceAccountToImpersonate` | `serviceAccountToImpersonate` | `jsonData` | `AuthConfig.tsx:196` (`<Field label="Service account to impersonate" ...>`) | Rendered inside AuthConfig only when `usingImpersonation` is true (`:195`) | `datasourceJSONData.ServiceAccountToImpersonate string`, `cloudmonitoring.go:210` | Backend consumer at `httpclient.go:57,69` |
| `jsonData_workloadIdentityPoolProvider` | `workloadIdentityPoolProvider` | `jsonData` | `WIFConfigEditor.tsx:19` (`<Field label="Workload Identity Pool Provider" description="Full resource name…">`) | `WIFConfigEditor.tsx:26` (`placeholder="projects/<number>/…/providers/<provider>"`) | `datasourceJSONData.WorkloadIdentityPoolProvider string`, `cloudmonitoring.go:213` | Backend validates non-empty for WIF at `httpclient.go:87-89` |
| `jsonData_wifServiceAccountEmail` | `wifServiceAccountEmail` | `jsonData` | `WIFConfigEditor.tsx:33` (`<Field label="Service account email" description="Optional…">`) | `WIFConfigEditor.tsx:40` (`placeholder="name@project.iam.gserviceaccount.com"`) | `datasourceJSONData.WifServiceAccountEmail string`, `cloudmonitoring.go:214` | Optional impersonation for WIF (consumed by Grafana Cloud's auth middleware, not by the plugin itself) |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (no direct UI; side-effect only) | Set by `AuthConfig.tsx:73-74` to `true` when auth is `forwardOAuthIdentity` or `workloadIdentityFederation`, `false` otherwise | `datasourceJSONData.OAuthPassThru bool`, `cloudmonitoring.go:211`; TS `oauthPassThru?: boolean`, `src/types/types.ts:42` | Tagged `editor-managed`; backend consumer at `cloudmonitoring.go:260-262` (sets `ForwardHTTPHeaders`) and `:121-166` (CheckHealth error routing) |
| `jsonData_universeDomain` | `universeDomain` | `jsonData` | `ConfigEditor.tsx:90` (`<Field noMargin label="Universe Domain">`) | `ConfigEditor.tsx:101` (`placeholder="googleapis.com"`) | `datasourceJSONData.UniverseDomain string`, `cloudmonitoring.go:208` | Consumed by `buildURL` at `httpclient.go:79-83`. The editor field is only rendered when `config.secureSocksDSProxyEnabled` is true (see [Upstream findings](#upstream-findings)) |
| `jsonData_gceDefaultProject` | `gceDefaultProject` | `jsonData` | — (no UI) | Populated at runtime by `src/datasource.ts:186-191` via `/gceDefaultProject` resource endpoint | Not on `datasourceJSONData`; TS `gceDefaultProject?: string`, `src/types/types.ts:39` | Tagged `frontend-only, runtime-cache`; backend never reads this key (`cloudmonitoring.go:666-675` always calls `gceDefaultProjectGetter` fresh) |

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
| `jsonData_wifServiceAccountEmail` | `wifServiceAccountEmail` | `jsonData` | Service account email (WIF) | No — read by Grafana Cloud's auth middleware, not the plugin |
| `jsonData_oauthPassThru` | `oauthPassThru` | `jsonData` | — (editor-managed side-effect) | Yes (gates `ForwardHTTPHeaders`) |
| `jsonData_universeDomain` | `universeDomain` | `jsonData` | Universe Domain (only rendered when secure-socks proxy enabled) | Yes |
| `jsonData_gceDefaultProject` | `gceDefaultProject` | `jsonData` | — (frontend-populated runtime cache) | No |

## Frontend-only settings

- `jsonData.gceDefaultProject` — cached by the frontend at query time (`src/datasource.ts:186-191`). The backend re-resolves the GCE default project on every call and never reads this key.
- `jsonData.enableSecureSocksProxy` — the shared Secure Socks Proxy flag, deliberately excluded from every registry entry per AGENTS.md.

## Backend-only settings

None. Every field the backend reads is also written by the editor (or, for `oauthPassThru`, by the editor as a side-effect).

## Where the types are defined

Configuration types are spread across the plugin, the shared `@grafana/google-sdk` React
package, and its Go counterpart. Some fields come from libraries rather than the plugin itself.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `CloudMonitoringOptions`, `CloudMonitoringSecureJsonData`, `isCloud` | `src/types/types.ts:38-45`, `src/utils.ts:15-17` | plugin ([grafana/grafana-cloudmonitoring-datasource](https://github.com/grafana/grafana-cloudmonitoring-datasource)) |
| `GoogleAuthType`, `DataSourceOptions`, `DataSourceSecureJsonData`, `GOOGLE_AUTH_TYPE_OPTIONS`, `WIF_AUTH_TYPE_OPTION`, `OAUTH_PASSTHROUGH_AUTH_TYPE_OPTION` | `src/types.ts:3-25`, `src/constants.ts:4-27` | `@grafana/google-sdk` `0.6.0` ([grafana/grafana-google-sdk-react](https://github.com/grafana/grafana-google-sdk-react)) |
| `AuthConfig`, `JWTForm`, `JWTConfigEditor`, `WIFConfigEditor`, `OAuthPassthroughConfigEditor` | `src/components/` | `@grafana/google-sdk` `0.6.0` |
| `DataSourceDescription`, `ConfigSection` | `packages/plugin-ui/src/` | `@grafana/plugin-ui` `0.13.1` |
| `Field`, `Input`, `Alert`, `Divider`, `Stack`, `SecureSocksProxySettings` | `packages/grafana-ui/src/components/` | `@grafana/ui` `13.1.0` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `datasourceJSONData` (parsed jsonData struct), `datasourceInfo` (resolved info incl. HTTP clients), `newDatasourceInfo` | `pkg/cloudmonitoring/cloudmonitoring.go:184-276` | plugin |
| `getMiddleware`, `newHTTPClient`, `buildURL` (dispatch on `AuthenticationType`, join services with `universeDomain`, gate `ForwardHTTPHeaders`) | `pkg/cloudmonitoring/httpclient.go:11-99` | plugin |
| `GetPrivateKey` (reads `privateKey` from `privateKeyPath` when set, else from decrypted secure JSON) | `pkg/utils/utils.go:62-89` | `github.com/grafana/grafana-google-sdk-go` |
| `GCEDefaultProject` (resolves the GCE metadata server's default project) | `pkg/utils/` | `github.com/grafana/grafana-google-sdk-go` |
| `tokenprovider.NewJwtAccessTokenProvider`, `NewGceAccessTokenProvider`, `NewImpersonatedJwtAccessTokenProvider`, `NewImpersonatedGceAccessTokenProvider`, `AuthMiddleware` | `pkg/tokenprovider/` | `github.com/grafana/grafana-google-sdk-go` |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` |

The models in this entry flatten that spread into a single Go `Config` type (jsonData fields +
`DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant list. `settings.ts` keeps
the three canonical TypeScript types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`).

## Modeling decisions

- **All four auth types modeled unconditionally**: the plugin composes `authOptions` at
  `ConfigEditor.tsx:32-38` from JWT + GCE, plus WIF + Forward OAuth Identity only when
  `isCloud()` returns true. The `isCloud` gate is a UI concern (only Cloud stacks see the
  extra radio buttons); the backend at `pkg/cloudmonitoring/httpclient.go:41-77` accepts all
  four values regardless of deployment. The schema therefore lists all four in the
  `allowedValues` set and doesn't gate any storage on `isCloud`.
- **`oauthPassThru` is editor-managed, not user-driven**: `AuthConfig.tsx:73-74` sets it as a
  side-effect of the `Authentication type` radio, so we model it in the schema (backend reads
  it at `cloudmonitoring.go:260-262` to set `ForwardHTTPHeaders`) but tag it `editor-managed`
  and provide no UI. The Go `ApplyDefaults` mirrors that behaviour: it is always derived from
  `AuthenticationType`.
- **Impersonation only for JWT / GCE**: `AuthConfig` renders its impersonation section only
  when the caller passes `showServiceAccountImpersonationConfig={true}`, and the plugin gates
  that on `authenticationType === 'jwt' || 'gce'` (`ConfigEditor.tsx:41-42`). The schema
  encodes both conditions in `dependsOn`.
- **WIF pool provider validated on backend**: `httpclient.go:87-89` fails the request if
  `authenticationType === 'workloadIdentityFederation' && workloadIdentityPoolProvider ==
  ''`. Captured as `requiredWhen` and enforced in Go `Validate`.
- **`defaultProject` required for both token-forwarding auth types**: `cloudmonitoring.go:121-125` refuses to CheckHealth when `oauthPassThru && defaultProject == ""`. Modeled as `requiredWhen: authenticationType in ('forwardOAuthIdentity', 'workloadIdentityFederation')` and enforced in Go `Validate`. Not required for JWT/GCE because those flows can either derive the project from credentials (JWT) or resolve it via the metadata server (GCE).
- **`gceDefaultProject` modeled but tagged `frontend-only`**: the frontend caches it into
  `jsonData` (`datasource.ts:186-191`), but the backend never reads it — it always calls
  `utils.GCEDefaultProject` fresh (`cloudmonitoring.go:666-675`). Kept in the schema so
  provisioning payloads that inadvertently carry it don't fail JSON-Schema validation, but
  callers should leave it empty.
- **`universeDomain` modeled unconditionally**: the editor only surfaces the field when the
  Grafana instance has `secureSocksDSProxyEnabled` set (`ConfigEditor.tsx:78`), but the
  backend consumes the field regardless (`httpclient.go:79-83`). Modeled as a normal jsonData
  field so provisioning can set it anywhere; documented in the field description and
  `Upstream findings` below.
- **Secure Socks Proxy excluded**: `SecureSocksProxySettings` (`ConfigEditor.tsx:104`) writes
  `jsonData.enableSecureSocksProxy`; deliberately omitted from this registry entry per
  AGENTS.md.
- **Flat `Config` in Go**: `settings.go` mirrors the plugin's `datasourceJSONData`
  (`cloudmonitoring.go:203-215`) verbatim, plus `gceDefaultProject` and an
  `enableSecureSocksProxy`-shaped field is intentionally omitted (excluded from the schema so
  it doesn't need to be in the Config either). Enum-like `AuthType` constants mirror
  `cloudmonitoring.go:52-55` verbatim.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle from
the embedded `dsconfig.json`: root fields plus a nested `jsonData` object become the OpenAPI
settings `spec`, secure fields become `secureValues`, and virtual fields are skipped.

`SettingsExamples()` provides the default configuration plus one k8s-style example per
authentication type and connection variant. Each example is a full instance-settings object
with the plugin configuration nested under `jsonData` and the relevant write-only secrets
under `secureJsonData` (placeholder values — replace with real secrets; the default example —
keyed by the empty string `""` — carries an empty `privateKey` to show what must be filled
in):

| Example | Auth | Notes | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | JWT (schema defaults) | Must still supply `defaultProject`, `clientEmail`, `tokenUri`, and a private key | `privateKey` (empty) |
| `googleJWTFile` | JWT | Inline `privateKey` in secureJsonData | `privateKey` |
| `googleJWTFilePath` | JWT | Private key from `privateKeyPath` file on the Grafana server | `privateKey` (empty — supplied by file) |
| `gceDefaultServiceAccount` | GCE Default Service Account | Only works on a GCE VM | `privateKey` (empty) |
| `workloadIdentityFederation` | Workload Identity Federation | Requires `workloadIdentityPoolProvider` and `defaultProject`; only exposed in the editor for Cloud stacks | `privateKey` (empty) |
| `forwardOAuthIdentity` | Forward OAuth Identity | Requires `defaultProject`; caller's OAuth token is forwarded; `oauthPassThru` set to `true` | `privateKey` (empty) |
| `impersonation` | JWT + service account impersonation | Base SA impersonates another SA (`usingImpersonation`, `serviceAccountToImpersonate`) | `privateKey` |
| `universeDomain` | JWT + custom universe domain | Non-default Google Cloud universe (Trusted Partner Cloud, mTLS endpoint) | `privateKey` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns a
fully-defaulted, validated `Config`:

1. **Parse** — unmarshal jsonData into `Config` (mirrors
   `pkg/cloudmonitoring/cloudmonitoring.go:222-248`) and copy decrypted secrets into
   `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — fill a curated set of zero-valued discriminators with the same
   defaults the editor writes for a fresh datasource:
   - `AuthenticationType = AuthTypeJWT` (matches both `AuthConfig.tsx:40-48` and backend
     `cloudmonitoring.go:229-231`).
   - `OAuthPassthroughEnabled` derived from `AuthenticationType` (matches
     `AuthConfig.tsx:73-74`: true for `forwardOAuthIdentity` / `workloadIdentityFederation`,
     false otherwise).
3. **`Validate`** — enforce the runtime contract: auth method + its required inputs, WIF
   pool provider, `defaultProject` for token-forwarding auth types. Errors are joined so
   every problem surfaces at once.

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

1. **`Universe Domain` visibility coupled to Secure Socks Proxy flag.**
   `ConfigEditor.tsx:78` only renders the "Additional settings" `ConfigSection` (which
   contains the Universe Domain input **and** the SecureSocksProxySettings component) when
   `config.secureSocksDSProxyEnabled` is true. On a Grafana instance without the proxy
   feature flag, users can't change the universe domain via the UI at all — but the backend
   consumes the field regardless (`httpclient.go:79-83`). Two unrelated features share one
   visibility gate.
2. **Alerting silently unsupported for token-forwarding auth.** `cloudmonitoring.go:412-418`
   rejects `QueryData` requests carrying the `FromAlert: true` header when
   `oauthPassThru` is true, but nothing in the config editor warns the user that switching
   to `forwardOAuthIdentity` or `workloadIdentityFederation` will break any existing alert
   rules pointing at the datasource.
3. **`oauthPassThru` can drift from `authenticationType`.** `AuthConfig.tsx:73-74` writes
   `oauthPassThru` as a side-effect only when the radio changes. A provisioning payload
   that sets `authenticationType: "jwt"` and `oauthPassThru: true` bypasses that logic —
   the backend will honour `oauthPassThru` and set `ForwardHTTPHeaders = true` on the HTTP
   client (`cloudmonitoring.go:260-262`) before ever reaching the JWT middleware. Our
   `ApplyDefaults` fixes this by always deriving `OAuthPassthroughEnabled` from
   `AuthenticationType`.
4. **WIF + Forward OAuth Identity visibility gated only in the UI.** `isCloud()`
   (`src/utils.ts:15-17`) returns `true` when `config.namespace` starts with `stacks-`.
   On-prem editors will not see the two token-forwarding radios, but the backend
   (`httpclient.go:41-77`) accepts both values regardless — a provisioning API caller can
   enable them on any Grafana instance. Not a bug per se, but a capability boundary worth
   flagging.
5. **Impersonation guard is asymmetric with the SDK.** `showServiceAccountImpersonation`
   (`ConfigEditor.tsx:41-42`) is `true` for `jwt` and `gce`, and AuthConfig internally
   additionally checks `authenticationType !== GoogleAuthType.ForwardOAuthIdentity`
   (`AuthConfig.tsx:170`) — for this plugin that check is redundant (the plugin already
   passes `false` for `forwardOAuthIdentity`). A provisioning payload with
   `authenticationType: "workloadIdentityFederation"` and `usingImpersonation: true` is
   accepted by the backend but doesn't actually do impersonation via the token provider
   (the code branch in `getMiddleware` at `httpclient.go:54-74` handles only `gce` /
   `jwt`); the federated identity does its own optional impersonation via
   `wifServiceAccountEmail` instead.
6. **`gceDefaultProject` is a frontend leak into stored jsonData.** The frontend caches the
   GCE metadata server's default project into `jsonData.gceDefaultProject`
   (`datasource.ts:186-191`), which persists into the datasource's stored settings — the
   next user editing the datasource sees a hard-coded project id even though the backend
   ignores it. Not directly harmful, but the field looks like configuration when it isn't.
7. **`cloudmonitoring.go:37` typo preserved.** The check-health error message
   `"forwardOAuthIdentityUnauthorizedMessage"` (`:69-70`) refers to
   `"…may have expired — sign out and back in to refresh it."` — the em-dash gets sent
   verbatim to the client. Cosmetic, but noted for compatibility with any client that
   pattern-matches the message.
8. **`plugin.json:4` id is `stackdriver`, product is `Google Cloud Monitoring`.** Google
   renamed Stackdriver to Cloud Monitoring in 2020; the plugin id stayed `stackdriver` for
   backward compatibility, but the docs, the plugin name, and every user-facing string say
   "Google Cloud Monitoring". Callers matching on `pluginType` need the legacy id.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) —
  passes (via the `SchemaRoundTrip` conformance subtest).
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft 2020-12, strict — `additionalProperties: false`) — passes.
- `go test -count=1 ./stackdriver/...` inside `registry/` — passes for this entry
  (schema bundle shape, secure values, examples, `LoadConfig` covering all four auth
  branches, the inline-vs-path private-key choice, and the `SchemaArtifactInSync` guard).
- `settings.go`/`schema.go`/`conformance_test.go`/`settings_test.go`: `go build`, `go vet`,
  `gofmt -l` — clean.
- `settings.ts`: `tsc --noEmit --strict` — clean.
- The pre-existing `dsconfig` and `schema` workspace modules still build and their tests
  pass (`go test ./dsconfig/... ./schema/...`).
