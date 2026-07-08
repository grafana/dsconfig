# grafana-newrelic-datasource

Declarative configuration schema for the New Relic datasource plugin (`grafana-newrelic-datasource`).

## Upstream researched

- **Monorepo**: `github.com/grafana/plugins-private` (private)
- **Commit SHA**: `267f4937806ed6404b6628d13ae358a5d308e376`
- **Plugin path**: `plugins/grafana-newrelic-datasource/`

The plugin lives inside the `plugins-private` monorepo (its `package.json` `repository`
points at the public mirror `github:grafana/newrelic-datasource`, but this entry was authored
against the monorepo sources at the SHA above). Every value in [`dsconfig.json`](dsconfig.json) —
labels, placeholders, tooltips, option labels/values, section title, defaults, required-when
expressions, storage keys, storage targets, value types, group title, and instructions — is
traceable to a specific `file:line` in the plugin at this SHA. See [Field provenance](#field-provenance).

To reproduce this research (monorepo already on disk — do not clone):

```bash
git -C <plugins-private> rev-parse HEAD    # 267f4937806ed6404b6628d13ae358a5d308e376
cd <plugins-private>/plugins/grafana-newrelic-datasource
```

`@grafana/*` dependencies use the `catalog:` protocol; versions are resolved from the workspace
catalog block in `<plugins-private>/.yarnrc.yml:14-26`. If the monorepo advances past this SHA,
re-diff the sources under [Sources researched](#sources-researched) before merging changes.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, group, relationship, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + parsed `AccountID` + `DecryptedSecureJSONData`), `PluginID`, `Region`/`SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each region/legacy variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| [`schema.gen.json`](schema.gen.json), [`settings.gen.json`](settings.gen.json), [`settings.examples.gen.json`](settings.examples.gen.json) | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`). The Go package name is
`newrelicdatasource`.

## Sources researched

Every source below was read at the pinned monorepo SHA
(`267f4937806ed6404b6628d13ae358a5d308e376`), plus external editor components at the versions
the workspace catalog pins.

### Plugin (`plugins/grafana-newrelic-datasource/`)

| File | What was read |
| --- | --- |
| `src/plugin.json:3-4,24` | `pluginName` (`name` = "New Relic"), `pluginType` (`id`), `docURL` (`info.links[0].url`) |
| `src/components/ConfigEditor.tsx:67` | Section heading `New Relic API Credentials` (single `<h3>` section) |
| `src/components/ConfigEditor.tsx:71-92` | Personal API Key field — label `Personal API Key / User API key`, tooltip `Used for NRQL queries`, placeholder `Personal API Key`, `required`, `Input type="password"`, `onUpdateDatasourceSecureJsonDataOption(..., 'personalApiKey')` |
| `src/components/ConfigEditor.tsx:109-148` | Account ID field — label `Account ID`, tooltip `Your New Relic Account ID`, placeholder `Account ID`, `required`, `Input type="number"`, `onUpdateDatasourceSecureJsonDataOption(..., 'accountId')` |
| `src/components/ConfigEditor.tsx:150-165` | Region field — label `Region`, tooltip `Region hosting your service`, placeholder `default`, `Select` with `options={regions}`, `onRegionChanged` writes `jsonData.region` (`:42-52`) |
| `src/components/ConfigEditor.tsx:167-186` | Timeout field — label `Timeout in Seconds`, tooltip/placeholder from selectors, `Input type="number"`, default `300` (`:57`, `|| 300` at `:181`) → `jsonData.timeoutInSeconds` |
| `src/components/ConfigEditor.tsx:20-40` | `componentDidMount` legacy migration: moves plaintext `jsonData.accountId` into `secureJsonData.accountId` and deletes `jsonData.accountId` / `jsonData.accountIdConfigured` |
| `src/components/ConfigEditor.tsx:60-61` | `secureJsonFields.personalApiKey` / `secureJsonFields.accountId` "configured" checks (write-only read side) |
| `src/components/selectors.ts:9-30` | ConfigEditor selectors — the Timeout tooltip (`Enter the timeout in seconds. Defaults to 300`) and placeholder (`300`), and the aria-labels |
| `src/types.ts:4-12` | `NewRelicSupportedRegion` (`'US' | 'EU'`), `NewRelicJsonData` (`region`, `timeoutInSeconds`), `NewRelicSecureJsonData` (`accountId`, `personalApiKey`) |
| `src/types.ts:187-190` | `regions` option list — order `EU`, `US` |
| `src/components/ConfigEditor.test.tsx:44-101` | Confirms: account ID only updates for numeric input; region renders as `default`; region dropdown lists the `regions` |
| `pkg/models/settings.go:12-24` | Backend `Settings` struct + json tags (`region`, `timeoutInSeconds`, `restBaseURL`, `infrastructureBaseURL`, `nerdGraphBaseURL`; `PersonalAPIKey`/`AccountID` as `json:"-"`) |
| `pkg/models/settings.go:27-48` | `LoadSettings`: unconditional `json.Unmarshal` (empty = error), copy `personalApiKey`, read + `strconv.Atoi` `accountId`, default `TimeoutInSeconds` to 300 when `< 1` |
| `pkg/models/settings_test.go:11-68` | Confirms jsonData parse, secret copy, 300s timeout default |
| `pkg/datasource/datasource.go:39-69` | `NewInstance`: `LoadSettings` → `CheckSettings` (fails instance creation) → `GetNewRelicClient` |
| `pkg/datasource/handler_checkhealth.go:26-41` | Error message constants (`NoPersonalAPIError`, `NoAccountIDError`, …) |
| `pkg/datasource/handler_checkhealth.go:99-101,138-148` | `isEmpty` (TrimSpace) and `CheckSettings`: requires non-empty `PersonalAPIKey`, non-zero `AccountID` |
| `pkg/datasource/newrelic_client.go:30-59` | `GetNewRelicClient`: `ConfigPersonalAPIKey` (`:43`), `ConfigRegion` when region set (`:47-49`), `ConfigBaseURL`/`ConfigInfrastructureBaseURL`/`ConfigNerdGraphBaseURL` from the three overrides (`:50-58`), timeout (`:35-44`) |
| `pkg/datasource/insights/insights_client.go:20-54` | `GetInsightsData`: `AccountID` is the NerdGraph `$accountId: Int!` NRQL variable |
| `pkg/main.go:8` | `dsID = "grafana-newrelic-datasource"` |
| `package.json:34-44` | External component versions via `catalog:` |
| `go.mod:11,13` | Backend deps: `grafana-plugin-sdk-go v0.279.0`, `newrelic-client-go/v2 v2.22.0` |

### External editor components

Read at the versions the workspace catalog pins (`<plugins-private>/.yarnrc.yml:14-26`, referenced
via `catalog:` in `package.json:34-44`).

| Component / API | Catalog version | Source | What was read |
| --- | --- | --- | --- |
| `InlineFormLabel`, `Input`, `Button`, `Select` | `@grafana/ui@^11.6.7` | `ConfigEditor.tsx:9` imports | Prop names (`tooltip`, `placeholder`, `value`, `type`, `onChange`, `options`, `width`) — which UI attributes to record |
| `DataSourcePluginOptionsEditorProps`, `onUpdateDatasourceSecureJsonDataOption`, `onUpdateDatasourceResetOption`, `SelectableValue` | `@grafana/data@^11.6.7` | `ConfigEditor.tsx:2-7` imports | Storage-key semantics of the secure update/reset helpers (which `secureJsonData` key each writes) |
| `DataSourceJsonData` | `@grafana/data@^11.6.7` | base interface `NewRelicJsonData extends` (`src/types.ts:5`) | Base jsonData interface fields |

Other cataloged deps (`@grafana/plugin-ui@^0.13.1`, `@grafana/runtime@^11.6.7`,
`@grafana/schema@^11.6.7`) are declared in `package.json` but are **not** used by
`ConfigEditor.tsx`; the config editor is hand-rolled from `@grafana/ui` primitives with no
`ConfigSection`/`DataSourceDescription` wrapper and no Secure Socks Proxy switch.

## Field provenance

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `secureJsonData_personalApiKey` | `personalApiKey` | `secureJsonData` | `ConfigEditor.tsx:72` (`Personal API Key / User API key`) | `ConfigEditor.tsx:88` (`placeholder="Personal API Key"`); tooltip `:71` (`Used for NRQL queries`) | `NewRelicSecureJsonData.personalApiKey` `types.ts:11`; `Settings.PersonalAPIKey` `settings.go:17` | Role `auth.apiKey.value`; `Input type="password"` (`:84`); `requiredWhen:"true"` from `required` (`:89`) + `CheckSettings` (`handler_checkhealth.go:139-141`) |
| `secureJsonData_accountId` | `accountId` | `secureJsonData` | `ConfigEditor.tsx:112` (`Account ID`) | `ConfigEditor.tsx:129` (`placeholder="Account ID"`); tooltip `:111` (`Your New Relic Account ID`) | `NewRelicSecureJsonData.accountId` `types.ts:10` (string); parsed to `Settings.AccountID int` `settings.go:18,42-46` | `Input type="number"` (`:126`) but stored as a `secureJsonData` string; `requiredWhen:"true"` from `required` (`:130`) + `CheckSettings` (`:143-145`) |
| `jsonData_region` | `region` | `jsonData` | `ConfigEditor.tsx:153` (`Region`) | Options `types.ts:187-190` (`EU`, `US`); placeholder `default` `ConfigEditor.tsx:161`; tooltip `:152` (`Region hosting your service`) | `NewRelicJsonData.region` `types.ts:6`; `Settings.Region string` `settings.go:13` | `Select` (`:156-162`); no default written; empty → New Relic client default (US) |
| `jsonData_timeoutInSeconds` | `timeoutInSeconds` | `jsonData` | `ConfigEditor.tsx:170` (`Timeout in Seconds`) | Placeholder `selectors.ts:27` (`300`); tooltip `selectors.ts:24`; default `300` `ConfigEditor.tsx:57,181` / `settings.go:38-40` | `NewRelicJsonData.timeoutInSeconds` `types.ts:7`; `Settings.TimeoutInSeconds int64` `settings.go:14` | Role `transport.timeoutSeconds`; `Input type="number"` (`:174`); `defaultValue:300` |
| `jsonData_restBaseURL` | `restBaseURL` | `jsonData` | — (no UI) | — | `Settings.RestBaseUrl string` `settings.go:21` | Backend-only; description from code comment `settings.go:20`; applied via `ConfigBaseURL` (`newrelic_client.go:50-52`) |
| `jsonData_infrastructureBaseURL` | `infrastructureBaseURL` | `jsonData` | — (no UI) | — | `Settings.InfraBaseUrl string` `settings.go:22` | Backend-only; applied via `ConfigInfrastructureBaseURL` (`newrelic_client.go:53-55`) |
| `jsonData_nerdGraphBaseURL` | `nerdGraphBaseURL` | `jsonData` | — (no UI) | — | `Settings.NerdGraphBaseURL string` `settings.go:23` | Backend-only; applied via `ConfigNerdGraphBaseURL` (`newrelic_client.go:56-58`) |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `secureJsonData_personalApiKey` | `personalApiKey` | `secureJsonData` | Personal API Key / User API key | Yes |
| `secureJsonData_accountId` | `accountId` | `secureJsonData` | Account ID | Yes (parsed to `int`) |
| `jsonData_region` | `region` | `jsonData` | Region | Yes |
| `jsonData_timeoutInSeconds` | `timeoutInSeconds` | `jsonData` | Timeout in Seconds | Yes |
| `jsonData_restBaseURL` | `restBaseURL` | `jsonData` | — (no UI) | Yes (backend-only) |
| `jsonData_infrastructureBaseURL` | `infrastructureBaseURL` | `jsonData` | — (no UI) | Yes (backend-only) |
| `jsonData_nerdGraphBaseURL` | `nerdGraphBaseURL` | `jsonData` | — (no UI) | Yes (backend-only) |

### Frontend-only settings

None. Every editor-visible field is read by the backend. (The legacy `jsonData.accountId` /
`jsonData.accountIdConfigured` keys are removed by the editor migration and are not modeled — see
[Upstream findings](#upstream-findings) #3.)

### Backend-only settings

- **`restBaseURL`, `infrastructureBaseURL`, `nerdGraphBaseURL`** have no editor UI. They exist only
  in the backend `Settings` struct (`pkg/models/settings.go:20-23`, commented "Used for internal
  testing and mocking. not exposed in the UI") and override the corresponding New Relic client
  base URLs when non-empty (`pkg/datasource/newrelic_client.go:50-58`).

## Where the types are defined

Only config type/field definitions are listed (UI components and helper functions are omitted even
where they are the reason a field exists).

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `NewRelicJsonData` (`region`, `timeoutInSeconds`), `NewRelicSecureJsonData` (`accountId`, `personalApiKey`), `NewRelicSupportedRegion` | `src/types.ts:4-12` | plugin (`grafana-newrelic-datasource`) |
| `DataSourceJsonData` (base interface `NewRelicJsonData` extends) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `^11.6.7` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Settings` (`Region`, `TimeoutInSeconds`, `PersonalAPIKey`, `AccountID`, `RestBaseUrl`, `InfraBaseUrl`, `NerdGraphBaseURL`) | `pkg/models/settings.go:12-24` | plugin (`grafana-newrelic-datasource`) |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, and root fields like `URL`/`BasicAuthEnabled` — unused by this plugin) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` `v0.279.0` |
| `newrelic.ConfigOption` (`ConfigPersonalAPIKey`, `ConfigRegion`, `ConfigBaseURL`, `ConfigInfrastructureBaseURL`, `ConfigNerdGraphBaseURL`); `region.Name` values | `newrelic/config.go`, `pkg/region/` | `github.com/newrelic/newrelic-client-go/v2` `v2.22.0` |

This entry flattens that spread into a single Go `Config` (jsonData fields verbatim + the parsed
`AccountID int` + `DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant list.
`settings.ts` keeps the three canonical TypeScript types. The `Region` string constants in
`settings.go` mirror the frontend `NewRelicSupportedRegion` union (the backend keeps `Region` as a
plain `string`).

## Modeling decisions

- **Single group, verbatim heading**: the editor renders one section
  (`<h3>New Relic API Credentials</h3>`, `ConfigEditor.tsx:67`) containing all four visible fields
  in order (Personal API Key, Account ID, Region, Timeout in Seconds). The schema mirrors that as a
  single `credentials` group titled `New Relic API Credentials`; the backend-only base-URL fields
  have no UI and are left out of the group (as GitHub does with `cachingEnabled`).
- **`accountId` modeled as a `secureJsonData` string, `AccountID` as the parsed number**: the editor
  writes `secureJsonData.accountId` (a string) and the backend parses it to `Settings.AccountID int`
  via `strconv.Atoi` (`settings.go:42-46`). `Config` keeps both the raw string (in
  `DecryptedSecureJSONData`) and the parsed `AccountID int` (`json:"-"`), analogous to how the GitHub
  entry keeps `AppIdInt64`. `Validate` checks the parsed `AccountID`, mirroring `CheckSettings`
  (`AccountID == 0`).
- **Secrets in `DecryptedSecureJSONData`**: following the gold-standard GitHub entry, the raw
  decrypted secrets (`personalApiKey`, `accountId`) live in the map rather than as separate struct
  fields, even though upstream `Settings` carries `PersonalAPIKey`/`AccountID` as `json:"-"` fields.
- **No root fields**: `LoadSettings` reads only `JSONData` + `DecryptedSecureJSONData` and
  `GetNewRelicClient` builds the client purely from the parsed `Settings`; `settings.URL`, basic auth,
  etc. are never read, so `RootConfig` is a blank object and `Config` carries no root fields.
- **`requiredWhen` vs the editor**: the editor marks both secret inputs `required`
  (`ConfigEditor.tsx:89,130`), and the backend hard-fails instance creation without them
  (`handler_checkhealth.go:139-145` via `NewInstance` at `datasource.go:50-54`). Both secrets get
  `requiredWhen:"true"`; region and timeout are optional.
- **`region` as a fixed `select`**: modeled with `ui.component:"select"` and options `EU`, `US`
  (order per `types.ts:187-190`); the converter derives the `enum` from the options. No `defaultValue`
  is set because the editor writes none (placeholder `default`) and an empty region falls back to the
  New Relic client default (US).
- **`LoadConfig` mirrors upstream verbatim**: `json.Unmarshal` is unconditional (empty/`nil` JSONData
  is a parse error, exactly as upstream `LoadSettings` behaves), the account ID is parsed only when
  `Atoi` succeeds, the 300s timeout default is applied in `ApplyDefaults`, and the
  personalApiKey/AccountID contract is enforced in `Validate` — together reproducing what
  `NewInstance` does (`LoadSettings` + `CheckSettings`).
- **Backend-only base URLs included**: `restBaseURL`/`infrastructureBaseURL`/`nerdGraphBaseURL` are
  modeled as `backend-only`-tagged jsonData fields (with descriptions from the code comment) so the
  jsonData/struct parity conformance holds and the internal overrides are documented.
- **No Secure Socks Proxy field**: the editor renders none, so the AGENTS.md
  `enableSecureSocksProxy` exclusion is not applicable here.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle from the
embedded `dsconfig.json`: the `jsonData` fields become the OpenAPI settings `spec` (with `region`
carrying an `enum` and `timeoutInSeconds` a `default` of 300), and the two secrets become
`secureValues`.

`SettingsExamples()` provides the default configuration plus region and legacy variants. Secret
placeholders are obviously-fake (`<your-newrelic-api-key>`); the numeric `accountId` is a plain
account number (not a credential):

| Example | jsonData | secureJsonData | Loads? |
| --- | --- | --- | --- |
| `""` (default) | `timeoutInSeconds: 300` | `personalApiKey` (empty), `accountId` (empty) | No — empty secrets fail `Validate` |
| `usRegion` | `region: US` (timeout defaults to 300) | `personalApiKey`, `accountId` | Yes |
| `euRegion` | `region: EU`, `timeoutInSeconds: 600` | `personalApiKey`, `accountId` | Yes |
| `legacyAccountIdInJsonData` | `region: US`, `accountId` (plaintext) | `personalApiKey` | No — backend reads only `secureJsonData.accountId`, so this fails until the editor migrates it |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx, settings) (Config, error)` runs the full three-phase flow and returns a
fully-defaulted, validated `Config`:

1. **Parse** — unmarshal `JSONData` into `Config` (unconditional, mirroring
   `pkg/models/settings.go:30-32`; empty/`nil` JSONData is a parse error), copy decrypted secrets by
   known key, and parse `secureJsonData.accountId` into `AccountID` via `strconv.Atoi` (only when it
   succeeds, `settings.go:42-46`).
2. **`ApplyDefaults`** — set `TimeoutInSeconds` to 300 when `< 1` (`settings.go:38-40`;
   `ConfigEditor.tsx:57`). Region has no default.
3. **`Validate`** — enforce the runtime contract from `CheckSettings`
   (`handler_checkhealth.go:138-148`): non-empty (trimmed) `personalApiKey` and non-zero `AccountID`.
   Errors are joined so every problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels. `ApplyDefaults` and `Validate` stay exported so callers that
assemble a `Config` directly can invoke each phase.

## Upstream findings

Potential bugs, misleading UX, and consistency issues found while researching upstream. All are
preserved verbatim in the schema — the schema records what the plugin **does**, not what it
**should** do.

1. **`accountId` is stored as a "secret" but is not one.** The New Relic account ID is a public
   numeric identifier, yet the editor persists it under `secureJsonData.accountId`
   (`ConfigEditor.tsx:131`) and renders it write-only with a Reset button (`:135-147`). It is modeled
   as a secure value here to match the plugin, but it carries no confidentiality.
2. **The account-ID check message overstates the constraint.** `NoAccountIDError` reads "Enter an
   account ID. This must be a valid, positive number." (`handler_checkhealth.go:36`), but
   `CheckSettings` only rejects `AccountID == 0` (`:143`) and `strconv.Atoi` accepts negatives — so a
   negative account ID (e.g. `-5`) passes settings validation even though it is not positive (and
   later NRQL queries against a negative `$accountId` would fail downstream). `Validate` here mirrors
   the `== 0` behavior; `TestValidate` documents the negative case.
3. **The legacy `accountId` migration is frontend-only.** `componentDidMount`
   (`ConfigEditor.tsx:20-40`) moves a plaintext `jsonData.accountId` into `secureJsonData.accountId`,
   but the backend only ever reads `secureJsonData.accountId` (`settings.go:36`). A provisioned or
   legacy datasource that has `accountId` under `jsonData` (and never had the config page re-saved)
   fails the account-ID check. Provisioning YAML must put the account ID in `secureJsonData.accountId`,
   not `jsonData`. The `legacyAccountIdInJsonData` example captures this (it is intentionally
   non-loadable).
4. **`LoadSettings` never validates.** It returns `nil` for any input it can unmarshal
   (`settings.go:27-48`); the required-field contract lives separately in `CheckSettings`, which
   `NewInstance` (`datasource.go:50-54`) and `CheckHealth` (`handler_checkhealth.go:60-83`) invoke.
   This entry composes both into `LoadConfig`.
5. **Empty/`nil` JSONData is a hard parse error.** `LoadSettings` calls `json.Unmarshal(config.JSONData, …)`
   unconditionally (`settings.go:30`), so a datasource with no `jsonData` at all fails to load with an
   unmarshal error. In practice Grafana always sends at least `{}` (the `NewInstance` test uses
   `[]byte("{}")`, `handler_checkhealth_test.go:276`).
6. **The Personal API Key tooltip is narrow.** The tooltip says "Used for NRQL queries"
   (`ConfigEditor.tsx:71`), but the key authenticates **every** New Relic request — NerdGraph, plus
   the REST/APM "info" calls used by the health check (`ConfigPersonalAPIKey` at
   `newrelic_client.go:43`; `CheckAPIKey` at `handler_checkhealth.go:156-174`), not only NRQL.
   Preserved verbatim as the field description.
7. **Region has no stored default and the dropdown is EU-first.** The `Select` shows placeholder
   `default` with options ordered `EU`, `US` (`types.ts:187-190`); nothing is written until the user
   picks one, and an empty region relies on the New Relic client's own default (US) via the
   `len(settings.Region) > 0` guard (`newrelic_client.go:47-49`). An invalid region provided out-of-band
   (e.g. provisioning) is not caught by `LoadSettings`/`CheckSettings` and would instead fail later in
   `newrelic.New` inside `GetNewRelicClient`.
8. **Undocumented base-URL overrides are provisionable.** `restBaseURL`, `infrastructureBaseURL`, and
   `nerdGraphBaseURL` are "internal testing and mocking" fields (`settings.go:20-23`) with no editor
   UI, but they are ordinary `jsonData` keys — anything provisioning `jsonData` can set them and
   redirect the plugin's API traffic (`newrelic_client.go:50-58`).

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes
  (via the `ConfigSchemaValid` conformance subtest).
- JSON Schema validation against [`dsconfig/schema.json`](../../dsconfig/schema.json) (draft-07,
  `additionalProperties: false`) — passes.
- `go generate ./...` (regenerates the three artifacts), then
  `gofmt -l . && go vet ./... && go build ./... && go test ./...` inside `registry/` — all clean;
  the New Relic entry's conformance + `LoadConfig`/`ApplyDefaults`/`Validate` tests pass, and the
  `SchemaArtifactInSync` guard confirms the committed artifacts match.
- The pre-existing `dsconfig` and `schema` workspace modules still build and test.
- `settings.ts`: `tsc --noEmit --strict` — clean.
