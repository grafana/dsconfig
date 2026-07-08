# grafana-cloudflare-datasource

Configuration schema for the [Cloudflare datasource plugin](https://grafana.com/docs/plugins/grafana-cloudflare-datasource)
(`grafana-cloudflare-datasource`), which lives in the [grafana/plugins](https://github.com/grafana/plugins)
monorepo. It has no hand-written `ConfigEditor.tsx` or per-plugin backend `Settings` model — both come from
shared packages (`@grafana/declarative-plugin` frontend, `github.com/grafana/plugins/sdk/pluginspec` backend)
specialized by the plugin's `src/spec.ts`. See [`grafana-zendesk-datasource`](../grafana-zendesk-datasource/README.md)
for the fullest description of the shared config/storage model.

## Upstream researched

- **Repo**: `github.com/grafana/plugins` (`plugins/grafana-cloudflare-datasource/`)
- **Ref**: `main` — **Commit**: `4b176ec1f74d80c231be2aeb1ce4713c833a76af` (2026-07-02)

## Sources researched (grafana/plugins @ 4b176ec)

| File | What was read |
| --- | --- |
| `plugins/grafana-cloudflare-datasource/src/plugin.json` | `id`/`name` → `pluginType`/`pluginName`; Docs link → `docURL` |
| `plugins/grafana-cloudflare-datasource/src/spec.ts:8-32,1034-1076` | service `cloudflare`; server `client_api` (`https://api.cloudflare.com/client/v4`, no variables); bearer auth `bearer_token` (name "API Key", description, api-tokens links) |
| `sdk/pluginspec/pluginclient/pluginclient.go:52-99` | auth.id defaults to first method; bearer token read as `DecryptedSecureJSONData["cloudflare.token"]` |

## Field inventory

| Schema `id` | Storage key | Target | Read by backend? |
| --- | --- | --- | --- |
| `jsonData_services_cloudflare_auth_id` | `services.cloudflare.auth.id` | `jsonData` | Yes — discriminator; defaults to `bearer_token` |
| `secureJsonData_cloudflare_token` | `cloudflare.token` | `secureJsonData` | Yes — bearer API token |

**Not modeled:** single-server `services.cloudflare.server.id` (backend-defaults), per-service `disabled`,
`enableSecureSocksProxy` (policy). No connection variables (single fixed server URL). The spec's `bearer_token`
auth method carries read-more/generate links to the Cloudflare API tokens page, captured as the fields' `docURL`.

## Where the types are defined

- Frontend: `Spec`/`Config`/`SecureConfig` in `@grafana/declarative-plugin`; service/auth data in the plugin's `src/spec.ts`.
- Backend: `Spec`/`AuthMethod` in `sdk/pluginspec/pluginspec.go`; `JsonData`/`ServiceConfig` in `sdk/pluginspec/pluginclient/config.go`.

`Config` in [`settings.go`](settings.go) projects the SDK's generic map-based `JsonData` into concrete structs.

## Modeling decisions

- `auth.id` is the `auth.discriminator` (single allowed value `bearer_token`, backend-parity default).
- No connection group — fixed server URL, no variables.
- `LoadConfig` mirrors the lenient `pluginclient.New` parse; `Validate` requires the bearer token.

## Settings examples

`""` (default, empty `cloudflare.token`) and `apiToken` (populated).

## Upstream findings

- Credentials are not validated at instance creation (`pluginclient.New` is lenient); a missing token only fails
  at request time. `Validate` encodes the working-datasource contract.

## Validation performed

- dsconfig `Validate()` (via `ConfigSchemaValid`); ajv against `dsconfig/schema.json` (`--spec=draft7`) — valid.
- `go build/vet/gofmt/test -race` in `registry/` — clean; full conformance suite passes.
- `tsc --noEmit --strict` on `settings.ts` — clean.
