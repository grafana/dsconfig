# grafana-supabase-datasource

Configuration schema for the Supabase datasource plugin (`grafana-supabase-datasource`), which lives in the
[grafana/plugins](https://github.com/grafana/plugins) monorepo. It has no hand-written `ConfigEditor.tsx` or
per-plugin backend `Settings` model — both come from shared packages (`@grafana/declarative-plugin` frontend,
`github.com/grafana/plugins/sdk/pluginspec` backend) specialized by the plugin's `src/spec.ts`. See
[`grafana-zendesk-datasource`](../grafana-zendesk-datasource/README.md) for the fullest description of the
shared config/storage model.

## Upstream researched

- **Repo**: `github.com/grafana/plugins` (`plugins/grafana-supabase-datasource/`)
- **Ref**: `main` — **Commit**: `4b176ec1f74d80c231be2aeb1ce4713c833a76af` (2026-07-02)

## Sources researched (grafana/plugins @ 4b176ec)

| File | What was read |
| --- | --- |
| `plugins/grafana-supabase-datasource/src/plugin.json` | `id`/`name` → `pluginType`/`pluginName` (no Docs link, so `docURL` omitted) |
| `plugins/grafana-supabase-datasource/src/spec.ts:8-30,61` | service `mgmt`; server `mgmt` (`https://api.supabase.com`, no variables); bearer auth `mgmt_bearer` (name/description "Supabase personal token") |
| `sdk/pluginspec/pluginclient/pluginclient.go:52-99` | auth.id defaults to first method; bearer token read as `DecryptedSecureJSONData["mgmt.token"]` |
| `packages/declarative-plugin/.../config-editor/Auth.tsx:200-211` | bearer token → secure key `<serviceId>.token`; field label "Token", placeholder "Token value" |

## Field inventory

| Schema `id` | Storage key | Target | Read by backend? |
| --- | --- | --- | --- |
| `jsonData_services_mgmt_auth_id` | `services.mgmt.auth.id` | `jsonData` | Yes — discriminator; defaults to `mgmt_bearer` |
| `secureJsonData_mgmt_token` | `mgmt.token` | `secureJsonData` | Yes — bearer token |

**Not modeled:** single-server `services.mgmt.server.id` (backend-defaults), per-service `disabled`,
`enableSecureSocksProxy` (policy). There are no connection variables (single fixed server URL).

## Where the types are defined

- Frontend: `Spec`/`Config`/`SecureConfig` in `@grafana/declarative-plugin`; service/auth data in the plugin's `src/spec.ts`.
- Backend: `Spec`/`AuthMethod` in `sdk/pluginspec/pluginspec.go`; `JsonData`/`ServiceConfig` in `sdk/pluginspec/pluginclient/config.go`.

`Config` in [`settings.go`](settings.go) projects the SDK's generic map-based `JsonData` into concrete structs.

## Modeling decisions

- `auth.id` is the `auth.discriminator` with the single allowed value `mgmt_bearer` and a backend-parity default.
- No connection group — the server URL is fixed and there are no variables.
- `LoadConfig` mirrors the lenient `pluginclient.New` parse; `Validate` requires the bearer token.

## Settings examples

`""` (default, empty `mgmt.token`) and `token` (populated bearer token).

## Upstream findings

- Credentials are not validated at instance creation (`pluginclient.New` is lenient); a missing token only fails
  at request time. `Validate` encodes the working-datasource contract.
- `plugin.json` has no Docs link and an empty description; `docURL` is therefore omitted.

## Validation performed

- dsconfig `Validate()` (via `ConfigSchemaValid`); ajv against `dsconfig/schema.json` (`--spec=draft7`) — valid.
- `go build/vet/gofmt/test -race` in `registry/` — clean; full conformance suite passes.
- `tsc --noEmit --strict` on `settings.ts` — clean.
