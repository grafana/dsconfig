# grafana-catchpoint-datasource

Configuration schema for the [Catchpoint datasource plugin](https://grafana.com/docs/plugins/grafana-catchpoint-datasource)
(`grafana-catchpoint-datasource`), which lives in the [grafana/plugins](https://github.com/grafana/plugins)
monorepo. It has no hand-written `ConfigEditor.tsx` or per-plugin backend `Settings` model — both come from
shared packages (`@grafana/declarative-plugin` frontend, `github.com/grafana/plugins/sdk/pluginspec` backend)
specialized by the plugin's `src/spec.ts`. See [`grafana-zendesk-datasource`](../grafana-zendesk-datasource/README.md)
for the fullest description of the shared config/storage model.

## Upstream researched

- **Repo**: `github.com/grafana/plugins` (`plugins/grafana-catchpoint-datasource/`)
- **Ref**: `main` — **Commit**: `4b176ec1f74d80c231be2aeb1ce4713c833a76af` (2026-07-02)

## Sources researched (grafana/plugins @ 4b176ec)

| File | What was read |
| --- | --- |
| `plugins/grafana-catchpoint-datasource/src/plugin.json` | `id`/`name` → `pluginType`/`pluginName`; Docs link → `docURL` |
| `plugins/grafana-catchpoint-datasource/src/spec.ts:8-22,538-548` | service `catchpoint`; server `client_api` (`https://io.catchpoint.com/api/v2`, no variables); bearer auth `bearer_token` (name "API v2 Key", description "Catchpoint REST API v2 Key.") |
| `sdk/pluginspec/pluginclient/pluginclient.go:52-99` | auth.id defaults to first method; bearer token read as `DecryptedSecureJSONData["catchpoint.token"]` |

## Field inventory

| Schema `id` | Storage key | Target | Read by backend? |
| --- | --- | --- | --- |
| `jsonData_services_catchpoint_auth_id` | `services.catchpoint.auth.id` | `jsonData` | Yes — discriminator; defaults to `bearer_token` |
| `secureJsonData_catchpoint_token` | `catchpoint.token` | `secureJsonData` | Yes — bearer API v2 key |

**Not modeled:** single-server `services.catchpoint.server.id` (backend-defaults), per-service `disabled`,
`enableSecureSocksProxy` (policy). No connection variables (single fixed server URL).

## Where the types are defined

- Frontend: `Spec`/`Config`/`SecureConfig` in `@grafana/declarative-plugin`; service/auth data in the plugin's `src/spec.ts`.
- Backend: `Spec`/`AuthMethod` in `sdk/pluginspec/pluginspec.go`; `JsonData`/`ServiceConfig` in `sdk/pluginspec/pluginclient/config.go`.

`Config` in [`settings.go`](settings.go) projects the SDK's generic map-based `JsonData` into concrete structs.

## Modeling decisions

- `auth.id` is the `auth.discriminator` (single allowed value `bearer_token`, backend-parity default).
- No connection group — fixed server URL, no variables.
- `LoadConfig` mirrors the lenient `pluginclient.New` parse; `Validate` requires the bearer key.

## Settings examples

`""` (default, empty `catchpoint.token`) and `apiKey` (populated).

## Upstream findings

- Credentials are not validated at instance creation (`pluginclient.New` is lenient); a missing key only fails at
  request time. `Validate` encodes the working-datasource contract.

## Validation performed

- dsconfig `Validate()` (via `ConfigSchemaValid`); ajv against `dsconfig/schema.json` (`--spec=draft7`) — valid.
- `go build/vet/gofmt/test -race` in `registry/` — clean; full conformance suite passes.
- `tsc --noEmit --strict` on `settings.ts` — clean.
