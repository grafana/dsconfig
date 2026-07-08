# grafana-logicmonitor-datasource

Configuration schema for the [LogicMonitor Devices datasource plugin](https://grafana.com/docs/plugins/grafana-logicmonitor-datasource)
(`grafana-logicmonitor-datasource`), which lives in the [grafana/plugins](https://github.com/grafana/plugins)
monorepo. It has no hand-written `ConfigEditor.tsx` or per-plugin backend `Settings` model — both come from
shared packages (`@grafana/declarative-plugin` frontend, `github.com/grafana/plugins/sdk/pluginspec` backend)
specialized by the plugin's `src/spec.ts`. See [`grafana-zendesk-datasource`](../grafana-zendesk-datasource/README.md)
for the fullest description of the shared config/storage model.

## Upstream researched

- **Repo**: `github.com/grafana/plugins` (`plugins/grafana-logicmonitor-datasource/`)
- **Ref**: `main` — **Commit**: `4b176ec1f74d80c231be2aeb1ce4713c833a76af` (2026-07-02)

## Sources researched (grafana/plugins @ 4b176ec)

| File | What was read |
| --- | --- |
| `plugins/grafana-logicmonitor-datasource/src/plugin.json` | `id`/`name` ("LogicMonitor Devices") → `pluginType`/`pluginName`; Docs link → `docURL` |
| `plugins/grafana-logicmonitor-datasource/src/spec.ts:8-46` | service `logicmonitor`; server `apiServer` (`https://{account_name}.logicmonitor.com/santaba/rest`, var `account_name` `required: true`); bearer auth `auth_bearer` (name "API v3 Key", description) |
| `sdk/pluginspec/pluginclient/pluginclient.go:52-99` | auth.id defaults to first method; bearer token read as `DecryptedSecureJSONData["logicmonitor.token"]` |

## Field inventory

| Schema `id` | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_variables_account_name` | `variables.account_name` | `jsonData` | Account Name | Yes — substituted into the base URL (required) |
| `jsonData_services_logicmonitor_auth_id` | `services.logicmonitor.auth.id` | `jsonData` | (auth selector "API v3 Key") | Yes — discriminator; defaults to `auth_bearer` |
| `secureJsonData_logicmonitor_token` | `logicmonitor.token` | `secureJsonData` | Token | Yes — bearer token |

**Not modeled:** single-server `services.logicmonitor.server.id` (backend-defaults), per-service `disabled`,
`enableSecureSocksProxy` (policy).

## Where the types are defined

- Frontend: `Spec`/`Config`/`SecureConfig` in `@grafana/declarative-plugin`; service/server/auth/variable data in the plugin's `src/spec.ts`.
- Backend: `Spec`/`AuthMethod` in `sdk/pluginspec/pluginspec.go`; `JsonData`/`ServiceConfig` in `sdk/pluginspec/pluginclient/config.go`.

`Config` in [`settings.go`](settings.go) projects the SDK's generic map-based `JsonData` into concrete structs.

## Modeling decisions

- `account_name` is marked required: the spec's server variable ref is `required: true` and it forms the base
  URL host. `Validate` enforces it (and the bearer token).
- `auth.id` is the `auth.discriminator` (single allowed value `auth_bearer`, backend-parity default).
- `LoadConfig` mirrors the lenient `pluginclient.New` parse.

## Settings examples

`""` (default, empty `account_name`/`logicmonitor.token`) and `apiKey` (populated).

## Upstream findings

- The `account_name` description ends with a stray backtick (`...logicmonitor.com/\``) in the spec; preserved verbatim.
- Credentials are not validated at instance creation (`pluginclient.New` is lenient); a missing token only fails
  at request time. `Validate` encodes the working-datasource contract.

## Validation performed

- dsconfig `Validate()` (via `ConfigSchemaValid`); ajv against `dsconfig/schema.json` (`--spec=draft7`) — valid.
- `go build/vet/gofmt/test -race` in `registry/` — clean; full conformance suite passes.
- `tsc --noEmit --strict` on `settings.ts` — clean.
