# grafana-googlesheets-datasource

Declarative configuration schema for the [Google Sheets datasource plugin](https://github.com/grafana/google-sheets-datasource) (`grafana-googlesheets-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/google-sheets-datasource`
- **Ref**: `main`
- **Commit SHA**: `7619fa04edd6fceac19d90e621cadab944c7512c` (2026-07-01, `Updating plugin-ci-workflows (#610)`)

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, tooltips, option
labels/values, section titles, help markdown, defaults, validations, dependency and
required-when expressions, storage keys, storage targets, value types, group titles, and
instructions — is traceable to a specific `file:line` in the upstream repo at this SHA. See
[Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone https://github.com/grafana/google-sheets-datasource
cd google-sheets-datasource
git checkout 7619fa04edd6fceac19d90e621cadab944c7512c
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under
[Sources researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, effects, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `SecureJsonDataKey` typed constants, `AuthType`/`LoadConfig`/`ApplyDefaults`/`Validate` |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each auth variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA
(`7619fa04edd6fceac19d90e621cadab944c7512c`), plus external editor components at the exact
versions the plugin's `package.json` pins.

### Plugin repo (`github.com/grafana/google-sheets-datasource@7619fa04`)

| File | What was read |
| --- | --- |
| `src/plugin.json:4-6,37` | `pluginType` (`id`), `pluginName` (`name`), `docURL` (`info.links[0].url`) |
| `src/components/ConfigEditor.tsx:1-19` | Editor imports; `AuthConfig` from `@grafana/google-sdk` handles all auth-related fields |
| `src/components/ConfigEditor.tsx:21-24` | `ConfigEditorProps` shape and `getBackwardCompatibleOptions` wrapping the incoming options |
| `src/components/ConfigEditor.tsx:30-43` | `apiKeyProps` — `SecretInput` placeholder `"Enter API key"`, `id="apiKey"`, `onReset`/`onChange` semantics |
| `src/components/ConfigEditor.tsx:45-82` | `loadSheetIDs`/`useEffect` for the default-sheet selector (state-only, not schema-modeled) |
| `src/components/ConfigEditor.tsx:85-89` | `DataSourceDescription` (`hasRequiredFields={false}`) — why no `required` marks in editor |
| `src/components/ConfigEditor.tsx:92-115` | The `grafana-info-box` intro text ("Choosing an authentication type") — captured verbatim in the `help` drawer of `jsonData_authenticationType` |
| `src/components/ConfigEditor.tsx:116` | `ConfigurationHelp` — the per-auth-type collapsible help block |
| `src/components/ConfigEditor.tsx:120` | `AuthConfig authOptions={googleSheetsAuthTypes}` — this plugin exposes 3 auth types (`key`, `jwt`, `gce`); the SDK's WIF and ForwardOAuthIdentity options are NOT added, and `showServiceAccountImpersonationConfig` is NOT passed, so impersonation fields never render |
| `src/components/ConfigEditor.tsx:122-126` | Conditional API-Key `Field` (label `"API Key"`, inner `SecretInput` label `"API key"`, `width={40}`) — only when `authenticationType === 'key'` |
| `src/components/ConfigEditor.tsx:130-151` | `Default Spreadsheet ID` field: label, description `"Optional spreadsheet ID to use as default when creating new queries"`, placeholder `"Select Spreadsheet ID"`, `SegmentAsync` with `allowCustomValue` |
| `src/types.ts:9-14` | `GoogleSheetsAuth` (`API: 'key'` extends SDK's `GoogleAuthType`) and `googleSheetsAuthTypes` (composition: API-Key first, then SDK's JWT+GCE options) |
| `src/types.ts:16-22` | `GoogleSheetsSecureJSONData` (adds `apiKey`) and `GoogleSheetsDataSourceOptions` (adds `defaultSheetID`) both extending the SDK's shape |
| `src/utils.ts:4-21` | `getBackwardCompatibleOptions`: sets `authenticationType = authenticationType || authType`, and when JWT is set and `secureJsonFields.jwt` is true, marks JWT-related jsonData fields as `configured` (frontend display only) |
| `src/components/ConfigurationHelp.tsx:9-186` | The `Collapse` "Configure Google Sheets Authentication" body — three help variants (API key, GCE, JWT default) with links to GCP console pages; captured in the help drawer of the discriminator |
| `pkg/models/settings.go:12-26` | Backend `DatasourceSettings` struct fields and json tags (drives our `Config` fields verbatim) |
| `pkg/models/settings.go:29-53` | `LoadSettings`: json-unmarshal jsonData, call `utils.GetPrivateKey`, copy `apiKey` and legacy `jwt` decrypted secrets, and migrate `authType` → `authenticationType` (lines 49-51) |
| `pkg/googlesheets/googleclient.go:20-24` | `authenticationTypeAPIKey = "key"` constant (backend discriminator value) |
| `pkg/googlesheets/googleclient.go:31-40` | Backend uses two Google APIs: Sheets (Spreadsheets read-only scope) and Drive (Drive read-only scope) |
| `pkg/googlesheets/googleclient.go:54-77` | `NewGoogleClient` — builds sheets + drive services from the settings |
| `pkg/googlesheets/googleclient.go:132-157` | `createSheetsService`: hard-fail with `"missing AuthenticationType setting"` (line 135), `"missing API Key"` for key auth (line 141); otherwise build an HTTP client with a token-provider middleware |
| `pkg/googlesheets/googleclient.go:185-228` | `getMiddleware`: switches on `authenticationType`, sets up `gce`/`jwt` token providers, parses JWT config from inline `settings.JWT` blob OR falls back to individual field/private-key path |
| `pkg/googlesheets/googleclient.go:240-246` | `validateDataSourceSettings`: `defaultProject`, `clientEmail`, `privateKey`, `tokenUri` all required for JWT — encoded as `requiredWhen` in our schema |

### External editor components

Read at the exact versions pinned in the plugin's `package.json`. Sources checked out at the
corresponding upstream refs.

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `AuthConfig` | `@grafana/google-sdk@0.6.1` | `github.com/grafana/grafana-google-sdk-react` HEAD (`package.json` reports `0.6.1`), `src/components/AuthConfig.tsx` | The whole auth flow: `authOptions` prop (`:16-21`), the "Authentication type" `RadioButtonGroup` label (`:103`), the `useEffect` default of `GoogleAuthType.JWT` (`:40-48`), the `onAuthTypeChange` behavior that sets `oauthPassThru` for WIF/OAuth but not others (`:66-78`), the `Default project` GCE field (`:151-158`), and the fact that WIF/OAuth panels only render when `authenticationType` matches (`:150-169`) — none of which the plugin exposes options for |
| `AuthConfig` impersonation UI | same | `AuthConfig.tsx:170-206` | Only rendered when `showServiceAccountImpersonationConfig` is passed; the plugin (`ConfigEditor.tsx:120`) does NOT pass it → `usingImpersonation` and `serviceAccountToImpersonate` never render → excluded from schema |
| `JWTForm` | `@grafana/google-sdk@0.6.1` | `grafana-google-sdk-react` `src/components/JWTForm.tsx` | Field labels (`Project ID`, `Client email`, `Token URI`, `Private key path`, `Private key`), placeholders (`Enter Private key` `:67`, `File location of your private key (e.g. /etc/secrets/gce.pem)` `:109`), the "Paste private key or provide path to private key file" description toggle (`:44-62`) |
| `JWTConfigEditor` | same | `grafana-google-sdk-react` `src/components/JWTConfigEditor.tsx` | JWT upload/paste flow that writes into `secureJsonData.privateKey` + `jsonData.{clientEmail,defaultProject,tokenUri}` (called in `AuthConfig.tsx:129-144`) — informs why we mark `privateKey` as the JWT signing key |
| `GOOGLE_AUTH_TYPE_OPTIONS` | same | `grafana-google-sdk-react` `src/constants.ts:4-15` | The label/value pairs for JWT and GCE options that the plugin's `googleSheetsAuthTypes` composes with its `key` option |
| `DataSourceSecureJsonData`, `DataSourceOptions`, `GoogleAuthType` | same | `grafana-google-sdk-react` `src/types.ts:3-25` | Base interfaces this plugin's TS types extend; discriminator values `jwt` / `gce` |
| `GetPrivateKey` (backend) | `grafana-google-sdk-go` (any version >= v0.4.1; matches the plugin's `go.mod`) | `grafana-google-sdk-go/pkg/utils/utils.go:62-89` | Reads `privateKey` from a file when `privateKeyPath` is set (accepts raw PEM or a service-account JSON with a `private_key` field), otherwise reads `settings.DecryptedSecureJSONData["privateKey"]` and normalizes `\\n` → `\n` |
| `DataSourceDescription`, `Divider` | `@grafana/plugin-ui@0.16.0` | `github.com/grafana/plugin-ui` — introspected via component prop shape only (no schema-relevant text besides `dataSourceName`, `docsLink`, `hasRequiredFields`) | `ConfigEditor.tsx:85-89` uses these — no additional storage fields introduced |
| `Field`, `SecretInput`, `SegmentAsync`, `Collapse`, `Input`, `RadioButtonGroup`, `FieldSet` | `@grafana/ui@13.1.0` | grafana/grafana `packages/grafana-ui/src/components/` | Prop names (`label`, `placeholder`, `value`, `onChange`, `isConfigured`, `onReset`, `width`) — so we knew which UI attributes to capture |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each of its
label, placeholder, tooltip, default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_authenticationType` | `authenticationType` | `jsonData` | `AuthConfig.tsx:103` (`<Field label="Authentication type">`) | Options composed at `types.ts:14` from `GoogleSheetsAuth.API` (`types.ts:11`) + `constants.ts:4-15` from `@grafana/google-sdk`; default `jwt` from `AuthConfig.tsx:40-48` `useEffect` | `Settings.AuthenticationType string`, `pkg/models/settings.go:20`; TS extends SDK's `authenticationType: string`, `grafana-google-sdk-react/src/types.ts:11` | Role `auth.discriminator`; help drawer verbatim from `ConfigEditor.tsx:92-115` |
| `jsonData_authType` | `authType` | `jsonData` | — (no UI at HEAD; provided for legacy configs) | Migration in `utils.ts:11` and `pkg/models/settings.go:49-51` | `Settings.AuthType string`, `pkg/models/settings.go:14` (`// jwt \| key \| gce`) | Tagged `legacy`; documented `description` since no editor label exists |
| `secureJsonData_apiKey` | `apiKey` | `secureJsonData` | `ConfigEditor.tsx:123` (`<Field label="API Key">`); inner label `"API key"` `ConfigEditor.tsx:124` | `ConfigEditor.tsx:33` (`placeholder: 'Enter API key'`) | `Settings.APIKey string`, `pkg/models/settings.go:15`; TS `apiKey?: string`, `types.ts:17` | Role `auth.apiKey.value`; `dependsOn` from `ConfigEditor.tsx:122` conditional render; `requiredWhen` from backend `googleclient.go:138-142` |
| `jsonData_defaultProject` | `defaultProject` | `jsonData` | `AuthConfig.tsx:151` (`<Field label="Default project">`) for GCE; `JWTForm.tsx:76` (`<Field label="Project ID">`) for JWT | Populated from uploaded JWT's `project_id` at `AuthConfig.tsx:140`; user input otherwise | `Settings.DefaultProject string`, `pkg/models/settings.go:16`; SDK `defaultProject?: string`, `grafana-google-sdk-react/src/types.ts:14` | Required for JWT (backend `googleclient.go:241`); not required for GCE (optional there) |
| `jsonData_clientEmail` | `clientEmail` | `jsonData` | `JWTForm.tsx:85` (`<Field label="Client email">`) | Populated from uploaded JWT's `client_email` at `AuthConfig.tsx:139`; user input otherwise | `Settings.ClientEmail string`, `pkg/models/settings.go:18`; SDK `clientEmail?: string`, `grafana-google-sdk-react/src/types.ts:13` | `dependsOn: authenticationType == 'jwt'`; `requiredWhen: ...jwt && privateKeyPath == ''` (backend validates it at `googleclient.go:241` when the JWT blob isn't provided) |
| `jsonData_tokenUri` | `tokenUri` | `jsonData` | `JWTForm.tsx:94` (`<Field label="Token URI">`) | Populated from uploaded JWT's `token_uri` at `AuthConfig.tsx:141` | `Settings.TokenURI string`, `pkg/models/settings.go:19`; SDK `tokenUri?: string`, `grafana-google-sdk-react/src/types.ts:12` | Same conditional/required story as `clientEmail` |
| `jsonData_privateKeyPath` | `privateKeyPath` | `jsonData` | `JWTForm.tsx:104` (`<Field label="Private key path" description={Description}>`) — description at `:44-62` "Paste private key or provide path to private key file" | `JWTForm.tsx:109` (`placeholder="File location of your private key (e.g. /etc/secrets/gce.pem)"`) | `Settings.PrivateKeyPath string`, `pkg/models/settings.go:21`; SDK `privateKeyPath?: string`, `grafana-google-sdk-react/src/types.ts:15`; backend consumer `grafana-google-sdk-go/pkg/utils/utils.go:62-80` (`GetPrivateKey`) | Alternative to inline `privateKey` (see `relationships`) |
| `secureJsonData_privateKey` | `privateKey` | `secureJsonData` | `JWTForm.tsx:117` (`<Field label="Private key" description={Description}>`) | `JWTForm.tsx:67` (`placeholder: 'Enter Private key'`) | `Settings.PrivateKey string`, `pkg/models/settings.go:25` (tag `json:"-"` — decrypted separately); backend reads via `utils.GetPrivateKey` `utils.go:82` | Role `auth.jwt.signingKey`; required for JWT unless `privateKeyPath` is set |
| `secureJsonData_jwt` | `jwt` | `secureJsonData` | — (no UI at HEAD) | Written by legacy JWT-upload flow (see the "Leaving this here for backward compatibility" comment) | `Settings.JWT string`, `pkg/models/settings.go:17`; decrypted at `:45` | Tagged `legacy`; documented `description` since no editor label exists |
| `jsonData_defaultSheetID` | `defaultSheetID` | `jsonData` | `ConfigEditor.tsx:131` (`<Field label="Default Spreadsheet ID" description="Optional spreadsheet ID to use as default when creating new queries">`) | `ConfigEditor.tsx:136` (`placeholder="Select Spreadsheet ID"`) | `Settings.DefaultSheetID string`, `pkg/models/settings.go:22`; TS `defaultSheetID?: string`, `types.ts:21` | Loaded by backend but not consumed at query time — UX hint only |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_authenticationType` | `authenticationType` | `jsonData` | Authentication type | Yes (backend discriminator) |
| `jsonData_authType` | `authType` | `jsonData` | — (legacy; no UI at HEAD) | Yes (migrated to `authenticationType`) |
| `secureJsonData_apiKey` | `apiKey` | `secureJsonData` | API Key | Yes (when auth is `key`) |
| `jsonData_defaultProject` | `defaultProject` | `jsonData` | Default project / Project ID | Yes (JWT + GCE token provider) |
| `jsonData_clientEmail` | `clientEmail` | `jsonData` | Client email | Yes (JWT token provider) |
| `jsonData_tokenUri` | `tokenUri` | `jsonData` | Token URI | Yes (JWT token provider) |
| `jsonData_privateKeyPath` | `privateKeyPath` | `jsonData` | Private key path | Yes (via `grafana-google-sdk-go` `GetPrivateKey`) |
| `secureJsonData_privateKey` | `privateKey` | `secureJsonData` | Private key | Yes (via `grafana-google-sdk-go` `GetPrivateKey`) |
| `secureJsonData_jwt` | `jwt` | `secureJsonData` | — (legacy) | Decrypted for BC but never consumed |
| `jsonData_defaultSheetID` | `defaultSheetID` | `jsonData` | Default Spreadsheet ID | Loaded but not used at query time |

### Legacy / backward-compat settings

- **`jsonData.authType`** was the original discriminator name; the current code writes
  `jsonData.authenticationType` and migrates `authType` → `authenticationType` on load
  (`pkg/models/settings.go:49-51`; `src/utils.ts:11`).
- **`secureJsonData.jwt`** was the original way to store the full JWT service-account JSON as
  a single secret; new configurations use `secureJsonData.privateKey` + the individual
  `defaultProject`/`clientEmail`/`tokenUri` jsonData fields. The backend still decrypts the
  legacy `jwt` value (`pkg/models/settings.go:45` — with the comment "Leaving this here for
  backward compatibility") but the middleware at `pkg/googlesheets/googleclient.go:199-210`
  only reads it when it is explicitly non-empty, and there is no editor UI that writes it.

## Where the types are defined

Configuration types are spread across the plugin, the shared `@grafana/google-sdk` React
package, and its Go counterpart — some fields come from libraries rather than the plugin
itself.

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `GoogleSheetsAuth`, `googleSheetsAuthTypes`, `GoogleSheetsSecureJSONData`, `GoogleSheetsDataSourceOptions` | `src/types.ts:9-22` | plugin ([grafana/google-sheets-datasource](https://github.com/grafana/google-sheets-datasource)) |
| `getBackwardCompatibleOptions` (frontend `authType` → `authenticationType` migration) | `src/utils.ts:4-21` | plugin |
| `GoogleAuthType`, `DataSourceOptions`, `DataSourceSecureJsonData` (base interfaces the plugin extends), `GOOGLE_AUTH_TYPE_OPTIONS` | `src/types.ts:3-25`, `src/constants.ts:4-15` | `@grafana/google-sdk` `0.6.1` ([grafana/grafana-google-sdk-react](https://github.com/grafana/grafana-google-sdk-react)) |
| `AuthConfig` (the whole auth panel: type radio, JWT form, GCE default-project input, WIF/OAuth panels for other plugins, optional impersonation) | `src/components/AuthConfig.tsx` | `@grafana/google-sdk` `0.6.1` |
| `JWTForm` (Project ID / Client email / Token URI / Private key or path) | `src/components/JWTForm.tsx` | `@grafana/google-sdk` `0.6.1` |
| `JWTConfigEditor` (upload/paste JSON key file, populates jsonData + secureJsonData) | `src/components/JWTConfigEditor.tsx` | `@grafana/google-sdk` `0.6.1` |
| `WIFConfigEditor` (Workload Identity Federation fields) | `src/components/WIFConfigEditor.tsx` | `@grafana/google-sdk` `0.6.1` — NOT exposed by this plugin |
| `OAuthPassthroughConfigEditor` | `src/components/OAuthPassthroughConfigEditor.tsx` | `@grafana/google-sdk` `0.6.1` — NOT exposed by this plugin |
| `DataSourceDescription`, `Divider` | `packages/plugin-ui/src/` | `@grafana/plugin-ui` `0.16.0` |
| `Field`, `SecretInput`, `SegmentAsync`, `Collapse`, `Input`, `RadioButtonGroup`, `FieldSet`, `Switch` | `packages/grafana-ui/src/components/` | `@grafana/ui` `13.1.0` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `DatasourceSettings` (all jsonData + decrypted secret staging), `LoadSettings` (parse + legacy migration) | `pkg/models/settings.go:12-53` | plugin |
| `NewGoogleClient`, `createSheetsService`, `createDriveService`, `getMiddleware`, `validateDataSourceSettings` (how the settings feed into Google API clients + token providers) | `pkg/googlesheets/googleclient.go:20-246` | plugin |
| `GetPrivateKey` (reads `privateKey` from disk via `privateKeyPath` when set, else from decrypted secure JSON; normalizes `\\n` → `\n`) | `pkg/utils/utils.go:62-89` | `github.com/grafana/grafana-google-sdk-go` |
| `tokenprovider.NewJwtAccessTokenProvider`, `tokenprovider.NewGceAccessTokenProvider`, `tokenprovider.AuthMiddleware` | `pkg/tokenprovider/` | `github.com/grafana/grafana-google-sdk-go` |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` |
| Google Sheets + Drive clients (`sheets.NewService`, `drive.NewService`) | — | `google.golang.org/api/sheets/v4`, `.../drive/v3` |

The models in this entry flatten that spread into a single Go `Config` type (jsonData fields +
`DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant list. `settings.ts` keeps
the three canonical TypeScript types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`).
The `AuthType` enum and its constants mirror the plugin's runtime values (`key`, `jwt`, `gce`)
and are the only auth types this plugin actually exposes — WIF and Forward-OAuth are defined
in the SDK but not added to `googleSheetsAuthTypes` (`src/types.ts:14`), so they are
deliberately absent from the schema.

## Modeling decisions

- **Three auth types, not four/five**: the SDK's `AuthConfig` component supports JWT, GCE,
  WIF, and Forward-OAuth. This plugin composes its option list at `src/types.ts:14` from just
  the API-Key option plus `GOOGLE_AUTH_TYPE_OPTIONS` (JWT + GCE), so only three buttons render.
  The backend token provider at `pkg/googlesheets/googleclient.go:195-225` only has branches
  for `gce`, `jwt`, and `key` — WIF and Forward-OAuth are unreachable. The schema therefore
  offers only three values in `jsonData_authenticationType`.
- **Impersonation not modeled**: `AuthConfig.tsx:170-206` renders the impersonation UI only
  when its caller passes `showServiceAccountImpersonationConfig={true}`. The plugin does not
  pass that prop (`ConfigEditor.tsx:120`), so `usingImpersonation` and
  `serviceAccountToImpersonate` never render and never end up in storage. Excluded from the
  schema.
- **API key is a plugin-only field**: `apiKey` is not part of the shared SDK — the plugin adds
  it in `src/types.ts:17` and renders its own `Field` at `ConfigEditor.tsx:122-126`. Modeled
  as `secureJsonData_apiKey` with `role: "auth.apiKey.value"` and gated on
  `authenticationType == 'key'`.
- **Legacy `authType` and `jwt` secret**: both are backward-compat fields that the current
  editor never writes. Modeled explicitly (no UI label, `description` explaining their status,
  `tags: ["legacy"]`) so that provisioning-style API consumers see them and know they
  shouldn't populate them for new datasources.
- **Alternative sources for the JWT private key**: the JWT flow accepts EITHER an inline
  `secureJsonData.privateKey` OR a `jsonData.privateKeyPath` pointing at a file on the Grafana
  server. `grafana-google-sdk-go/pkg/utils/utils.go:74-80` prefers `privateKeyPath` when set
  and falls back to the decrypted secret otherwise. Captured as a `relationships` entry of
  `type: "group"` (the schema doesn't have an `"alternative"` primitive) with the trade-off
  documented in the description, and enforced in Go `Validate` at
  `settings.go:Validate` for the `AuthTypeJWT` branch.
- **`defaultProject` is a JWT field AND a GCE field**: `AuthConfig.tsx:151-158` renders it as
  "Default project" in the GCE branch; `JWTForm.tsx:76-83` renders it as "Project ID" in the
  JWT branch. Schema uses the label `"Default project"` (the more general one) with a
  `dependsOn` that covers both auth types. The two labels differ upstream but capture the
  same field.
- **Help drawers**: `ConfigEditor.tsx:92-115` and `ConfigurationHelp.tsx:9-186` provide
  ~200 lines of guidance across two collapsibles. Consolidated as `help` markdown on the two
  most relevant fields — `jsonData_authenticationType` (top-level "Choosing an authentication
  type" summary) and `secureJsonData_apiKey` (the API-Key-specific "Generate an API key"
  steps). The GCE and JWT walkthroughs are long — surfaced via instructions rather than
  duplicated verbatim on each JWT field.
- **Field ID naming convention**: IDs are prefixed with their storage target — `jsonData_` or
  `secureJsonData_` — followed by the camelCase storage key, matching the pattern used by
  every other entry in this registry.
- **Flat `Config` in Go**: `settings.go` mirrors the plugin's `pkg/models/settings.go`
  `DatasourceSettings` (minus the SDK back-reference `InstanceSettings`) with typed
  `AuthType` and `SecureJsonDataKey` enums. `settings.ts` keeps the three canonical TS types.
- **`SecureJsonDataConfig` is a key list**: secure values are write-only, so the secure type
  is just the array of secret key names (`apiKey`, `privateKey`, `jwt`); consumers read
  `secureJsonFields` to see what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle (the
k8s-style schema Grafana's datasource API server serves as `{apiVersion}.json`, `v0alpha1`
today) from the embedded `dsconfig.json`: root fields plus a nested `jsonData` object become
the OpenAPI settings `spec`, secure fields become `secureValues`, and virtual fields are
skipped.

`SettingsExamples()` provides the default configuration plus one k8s-style example per
authentication type / connection variant. Each example is a full instance-settings object with
the plugin configuration nested under `jsonData` and the relevant write-only secrets under
`secureJsonData` (placeholder values — replace with real secrets; the default example — keyed
by the empty string `""` — carries an empty `privateKey` to show what must be filled in):

| Example | Auth | Notes | `secureJsonData` |
| --- | --- | --- | --- |
| `""` (default) | JWT (schema defaults) | Must still supply `defaultProject`, `clientEmail`, `tokenUri`, and a private key to be functional | `privateKey` (empty) |
| `apiKey` | API Key | Public spreadsheets only | `apiKey` |
| `googleJWTFile` | JWT | Inline `privateKey` in secureJsonData | `privateKey` |
| `googleJWTFilePath` | JWT | Private key from `privateKeyPath` file on the Grafana server | `privateKey` (empty — supplied by file) |
| `gceDefaultServiceAccount` | GCE Default Service Account | Only works on a GCE VM; no secret needed | (none) |
| `legacyAuthType` | Legacy `authType`-only shape | `authenticationType` absent; backend migrates it | `apiKey` |
| `withDefaultSheetID` | API Key + `defaultSheetID` | Illustrates the UX-hint `defaultSheetID` | `apiKey` |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns a
fully-defaulted, validated `Config`:

1. **Parse** — unmarshal jsonData into `Config`, run the upstream legacy fallback that copies
   `AuthType` into `AuthenticationType` when the former is set and the latter is empty
   (mirrors `pkg/models/settings.go:49-51`), and copy decrypted secrets into
   `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — fill a curated set of zero-valued discriminators with the same
   defaults the editor writes for a fresh datasource (`AuthenticationType=AuthTypeJWT` per
   `AuthConfig.tsx:40-48`).
3. **`Validate`** — enforce the runtime contract (auth method + its required inputs, and the
   inline-vs-path private-key choice for JWT). Errors are joined so every problem surfaces at
   once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels so log lines carry request context.

This is the intended shape for the plugin's own upstream `LoadSettings` to sync to: a load
returns a config that is safe to use, or an error explaining why it isn't.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are still exported for callers
that want to compose them themselves (e.g. provisioning preview, schema-example round-trip,
tests that need to distinguish parse-level from policy-level errors). Skip them by never
calling `LoadConfig` in those flows — assemble a `Config` directly.

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching upstream.
All preserved as-is in the schema — the schema records what the plugin **does**, not what it
**should** do; these notes exist so reviewers can reproduce each finding and decide separately
whether to fix upstream.

1. **`defaultSheetID` is stored but never consumed by the backend.** `pkg/models/settings.go:22`
   defines the field and `LoadSettings` unmarshals it, but no code path in
   `pkg/googlesheets/` reads it (query editor uses its own `spreadsheet` field on
   `SheetsQuery`, `src/types.ts:43`). Functionally a UX hint only.
2. **`secureJsonData.jwt` is dead weight for the backend at HEAD.** `pkg/models/settings.go:45`
   decrypts it "for backward compatibility", and `pkg/googlesheets/googleclient.go:199` only
   uses it when the middleware happens to be called with `settings.JWT != ""`. The current
   editor never writes it — `AuthConfig.tsx:129-144` writes the individual `clientEmail` /
   `defaultProject` / `tokenUri` fields and the `secureJsonData.privateKey` secret instead.
3. **JWT credential validation is split across three places.** `pkg/googlesheets/googleclient.go:240-245`
   (`validateDataSourceSettings`) requires `DefaultProject`, `ClientEmail`, `PrivateKey`,
   `TokenURI`, but is only invoked in the JWT branch when `settings.JWT` is empty
   (`googleclient.go:212-217`). When a legacy `jwt` blob is present, none of the individual
   fields are checked at all. Our `Validate` requires all four regardless because that is the
   supported code path.
4. **`authType` and `authenticationType` can disagree.** Both are unmarshaled from jsonData,
   and `LoadSettings` at `pkg/models/settings.go:49-51` overwrites `authenticationType` with
   `authType` when `authType` is non-empty — meaning a provisioning payload that sets both
   fields to different values silently uses `authType`. The schema's `authType` `description`
   documents this precedence.
5. **Impersonation, WIF, and Forward-OAuth code paths are dead in this plugin.** The SDK's
   `AuthConfig` component supports all four extra flows, but the plugin does not add their
   option buttons and the backend has no branches for them. If they ever appear in a
   provisioning payload (e.g. copy-pasted from a Prometheus datasource), the backend will hit
   `"missing AuthenticationType setting"` or `"missing API Key"`-style errors because none of
   the code branches match.
6. **Legacy JWT-migration UI never renders the "configured" placeholder for new datasources.**
   `src/utils.ts:14-19` sets `clientEmail`/`defaultProject`/`tokenUri` to the literal string
   `"configured"` when a JWT blob exists in `secureJsonFields.jwt`. Any consumer inspecting
   jsonData on such a legacy datasource sees those placeholders in memory, though they never
   round-trip back to storage because the editor only writes on user input.
7. **`privateKeyPath` is trusted verbatim; no path sanitization.**
   `grafana-google-sdk-go/pkg/utils/utils.go:36` does a plain `os.ReadFile` on whatever the
   user provides. This is a deliberate ops-only escape hatch, but worth flagging: whoever
   controls the datasource config can read any file the Grafana process can read, as long as
   they can trigger a JWT auth flow. Not a bug per se, but a capability boundary reviewers
   should be aware of.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes.
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json)
  (draft-07, `additionalProperties: false`) — passes.
- `go test ./...` on the shared `registry/` module — passes for this entry (schema bundle
  shape, secure values, examples, `LoadConfig` incl. legacy migration and inline-vs-path
  private-key choice, `SchemaArtifactInSync` guard).
- `settings.go`/`schema.go`: `go build`, `go vet`, `gofmt` — clean.
- `settings.ts`: `tsc --noEmit --strict` — clean.
