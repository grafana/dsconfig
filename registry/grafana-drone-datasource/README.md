# grafana-drone-datasource

Configuration schema for the [Drone datasource plugin](https://grafana.com/docs/plugins/grafana-drone-datasource)
(`grafana-drone-datasource`), which lives in the [grafana/plugins](https://github.com/grafana/plugins)
monorepo. It has no hand-written `ConfigEditor.tsx` or per-plugin backend `Settings` model — both come from
shared packages (`@grafana/declarative-plugin` frontend, `github.com/grafana/plugins/sdk/pluginspec` backend)
specialized by the plugin's `src/spec.ts`. See [`grafana-zendesk-datasource`](../grafana-zendesk-datasource/README.md)
for the fullest description of the shared config/storage model.

## Upstream researched

- **Repo**: `github.com/grafana/plugins` (`plugins/grafana-drone-datasource/`)
- **Ref**: `main` — **Commit**: `4b176ec1f74d80c231be2aeb1ce4713c833a76af` (2026-07-02)

## Sources researched (grafana/plugins @ 4b176ec)

| File | What was read |
| --- | --- |
| `plugins/grafana-drone-datasource/src/plugin.json` | `id`/`name` → `pluginType`/`pluginName`; Docs link → `docURL` |
| `plugins/grafana-drone-datasource/src/spec.ts:6-37` | service `drone`; server `apiServer` (`{url}/api`, var `url`); bearer auth `auth_bearer` (name "Drone API token", description) |
| `sdk/pluginspec/pluginclient/pluginclient.go:52-99` | auth.id defaults to first method; bearer token read as `DecryptedSecureJSONData["drone.token"]` |
| `packages/declarative-plugin/.../common/VariablesForm.tsx:22-61` | variable label from `variable.name` |

## Field inventory

| Schema `id` | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_variables_url` | `variables.url` | `jsonData` | URL | Yes — substituted into `{url}/api` |
| `jsonData_services_drone_auth_id` | `services.drone.auth.id` | `jsonData` | (auth selector "Drone API token") | Yes — discriminator; defaults to `auth_bearer` |
| `secureJsonData_drone_token` | `drone.token` | `secureJsonData` | Token | Yes — bearer token |

**Not modeled:** single-server `services.drone.server.id` (backend-defaults), per-service `disabled`,
`enableSecureSocksProxy` (policy).

## Where the types are defined

- Frontend: `Spec`/`Config`/`SecureConfig` in `@grafana/declarative-plugin`; service/server/auth/variable data in the plugin's `src/spec.ts`.
- Backend: `Spec`/`AuthMethod` in `sdk/pluginspec/pluginspec.go`; `JsonData`/`ServiceConfig` in `sdk/pluginspec/pluginclient/config.go`.

`Config` in [`settings.go`](settings.go) projects the SDK's generic map-based `JsonData` into concrete structs.

## Modeling decisions

- `url` is marked required: it forms the server base URL (`{url}/api`) and the editor renders it required.
  The backend's `validateVariables` does not enforce it (the spec's ref lacks `required: true`), so a missing url
  surfaces as a broken request URL rather than a config error — recorded below.
- `auth.id` is the `auth.discriminator` (single allowed value `auth_bearer`, backend-parity default).
- `LoadConfig` mirrors the lenient `pluginclient.New` parse; `Validate` requires the token and the url.

## Settings examples

`""` (default, empty `url`/`drone.token`) and `apiToken` (populated).

## Upstream findings

- `url` is shown required in the editor but not enforced by the backend (`validateVariables` only enforces refs
  marked `required: true`; the `url` ref is not). Without it the base URL is malformed (`/api`).
- Credentials are not validated at instance creation (`pluginclient.New` is lenient); a missing token only fails
  at request time. `Validate` encodes the working-datasource contract.

## Validation performed

- dsconfig `Validate()` (via `ConfigSchemaValid`); ajv against `dsconfig/schema.json` (`--spec=draft7`) — valid.
- `go build/vet/gofmt/test -race` in `registry/` — clean; full conformance suite passes.
- `tsc --noEmit --strict` on `settings.ts` — clean.
