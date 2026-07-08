# grafana-vercel-datasource

Configuration schema for the [Vercel datasource plugin](https://grafana.com/docs/plugins/grafana-vercel-datasource)
(`grafana-vercel-datasource`), which lives in the [grafana/plugins](https://github.com/grafana/plugins)
monorepo. It has no hand-written `ConfigEditor.tsx` or per-plugin backend `Settings` model — both are
provided by shared packages (`@grafana/declarative-plugin` frontend, `github.com/grafana/plugins/sdk/pluginspec`
backend) and specialized by the plugin's `src/spec.ts`. See
[`grafana-zendesk-datasource`](../grafana-zendesk-datasource/README.md) for the fullest description of
the shared config/storage model; this entry documents Vercel's specifics.

## Upstream researched

- **Repo**: `github.com/grafana/plugins` (monorepo; `plugins/grafana-vercel-datasource/`)
- **Ref**: `main` — **Commit**: `4b176ec1f74d80c231be2aeb1ce4713c833a76af` (2026-07-02)

## Files

Standard registry-entry layout: [`dsconfig.json`](dsconfig.json) (single source of truth),
[`settings.ts`](settings.ts), [`settings.go`](settings.go) (`Config` + `LoadConfig`/`ApplyDefaults`/`Validate`),
[`schema.go`](schema.go) (+ `SettingsExamples`), [`conformance_test.go`](conformance_test.go),
[`settings_test.go`](settings_test.go), and the generated `*.gen.json` artifacts. No per-entry `go.mod`.

## Sources researched (grafana/plugins @ 4b176ec)

| File | What was read |
| --- | --- |
| `plugins/grafana-vercel-datasource/src/plugin.json:3-5,info.links` | `id`/`name` → `pluginType`/`pluginName`; Docs link → `docURL` |
| `plugins/grafana-vercel-datasource/src/spec.ts:9,18-56` | service `vercel`; server `vercelServer` (`https://api.vercel.com`, var `team_id`); auth method `vercelApiKey` (type `bearer`, name "Access Token", description, read-more/generate links) |
| `sdk/pluginspec/pluginclient/pluginclient.go:52-99` | server.id/auth.id default to first; bearer token read as `DecryptedSecureJSONData["vercel.token"]` |
| `packages/declarative-plugin/.../config-editor/Auth.tsx:200-211` | bearer token → secure key `<serviceId>.token`; field label defaults to "Token", placeholder "Token value" |
| `packages/declarative-plugin/.../common/VariablesForm.tsx:22-61` | variable label from `variable.name`; placeholder from `variable.placeholder` |

## Field inventory

| Schema `id` | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_variables_team_id` | `variables.team_id` | `jsonData` | Team ID | Yes — optional team scope (not in base URL) |
| `jsonData_services_vercel_auth_id` | `services.vercel.auth.id` | `jsonData` | (auth selector "Access Token") | Yes — discriminator; defaults to `vercelApiKey` |
| `secureJsonData_vercel_token` | `vercel.token` | `secureJsonData` | Token | Yes — bearer token |

**Frontend-only:** none. **Not modeled:** single-server `services.vercel.server.id` (backend-defaults),
per-service `disabled`, and `enableSecureSocksProxy` (policy).

## Where the types are defined

- Frontend: `Spec`/`Config`/`SecureConfig` in `@grafana/declarative-plugin` (`packages/declarative-plugin/src/types/*`); the concrete service/server/auth/variable data in the plugin's `src/spec.ts`.
- Backend: `Spec`/`AuthMethod` in `sdk/pluginspec/pluginspec.go`; `JsonData`/`ServiceConfig` in `sdk/pluginspec/pluginclient/config.go`; `backend.DataSourceInstanceSettings` in `grafana-plugin-sdk-go`.

`Config` in [`settings.go`](settings.go) is a spec-specific projection of the SDK's generic map-based
`JsonData` (concrete structs so the jsonData⇔struct conformance guard holds).

## Modeling decisions

- Service-keyed storage → dotted `section`s (`services.vercel.auth`, `variables`); the bearer token is a
  flat dotted secure key (`vercel.token`).
- `auth.id` is the `auth.discriminator` with the single allowed value `vercelApiKey` and a backend-parity default.
- `team_id` is **not** marked required: it is not part of the base URL (`https://api.vercel.com`) and is only
  needed for team-scoped tokens (per the auth method description).
- `LoadConfig` mirrors the lenient `pluginclient.New` parse; `Validate` encodes the health-check contract
  (bearer requires the token).

## Settings examples

| Example | `jsonData` | `secureJsonData` |
| --- | --- | --- |
| `""` (default) | `services.vercel.auth.id` | `vercel.token` (empty) |
| `accessToken` | + `variables.team_id` | `vercel.token` (placeholder) |

## Upstream findings

- Credentials are not validated at instance creation (`pluginclient.New` is lenient); a missing token only
  fails at request time via `applyAuth`. `Validate` encodes the stricter working-datasource contract.
- `team_id` is declared as a server variable but the server URL contains no `{team_id}` placeholder, so it has
  no effect on the base URL; it is used only for team-scoped requests.

## Validation performed

- dsconfig `Validate()` (via `ConfigSchemaValid`); ajv against `dsconfig/schema.json` (`--spec=draft7`) — valid.
- `go build/vet/gofmt/test -race` in `registry/` — clean; full conformance suite passes.
- `tsc --noEmit --strict` on `settings.ts` — clean.
