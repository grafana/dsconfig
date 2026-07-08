# grafana-hello-datasource

Configuration schema for the Hello datasource plugin (`grafana-hello-datasource`), which lives in the
[grafana/plugins](https://github.com/grafana/plugins) monorepo. It has no hand-written `ConfigEditor.tsx` or
per-plugin backend `Settings` model — both come from shared packages (`@grafana/declarative-plugin` frontend,
`github.com/grafana/plugins/sdk/pluginspec` backend) specialized by the plugin's `src/spec.ts`. See
[`grafana-zendesk-datasource`](../grafana-zendesk-datasource/README.md) for the fullest description of the shared
config/storage model.

**Hello is an experimental plugin used for testing the framework.** It has two services (`httpbin`,
`postman_echo`), both with fixed server URLs and the `none` auth method, so it stores **no `secureJsonData`
secrets** and needs essentially no configuration.

## Upstream researched

- **Repo**: `github.com/grafana/plugins` (`plugins/grafana-hello-datasource/`)
- **Ref**: `main` — **Commit**: `4b176ec1f74d80c231be2aeb1ce4713c833a76af` (2026-07-02)

## Sources researched (grafana/plugins @ 4b176ec)

| File | What was read |
| --- | --- |
| `plugins/grafana-hello-datasource/src/plugin.json` | `id`/`name` → `pluginType`/`pluginName` (no Docs link, so `docURL` omitted) |
| `plugins/grafana-hello-datasource/src/spec.ts:8-45` | services `httpbin`/`postman_echo`; fixed server URLs (https://httpbin.org, https://postman-echo.com); auth method `none` |
| `packages/declarative-plugin/.../config-editor/Auth.tsx:43-46` | the `none` auth method renders as "No Auth" and writes `auth.id = 'none'` |
| `sdk/pluginspec/pluginclient/pluginclient.go:52-57` | auth.id defaults to the server's first method (`none`) |

## Field inventory

| Schema `id` | Storage key | Target | Read by backend? |
| --- | --- | --- | --- |
| `jsonData_services_httpbin_auth_id` | `services.httpbin.auth.id` | `jsonData` | Yes — discriminator; defaults to `none` |
| `jsonData_services_postman_echo_auth_id` | `services.postman_echo.auth.id` | `jsonData` | Yes — discriminator; defaults to `none` |

There are **no** secure fields and **no** connection variables. **Not modeled:** single-server
`services.<id>.server.id` (backend-defaults), per-service `disabled`, `enableSecureSocksProxy` (policy).

## Where the types are defined

- Frontend: `Spec`/`Config` in `@grafana/declarative-plugin`; service/auth data in the plugin's `src/spec.ts`.
- Backend: `Spec` in `sdk/pluginspec/pluginspec.go`; `JsonData`/`ServiceConfig` in `sdk/pluginspec/pluginclient/config.go`.

## Modeling decisions

- With no authentication and no variables, the only stored fields are the per-service `auth.id` discriminators
  (value `none`). They are included because the dsconfig schema requires at least one field and they are the
  genuine stored/defaulted discriminators; a fully empty datasource is still valid.
- **Conformance test deviation:** the shared `schema.RunPluginTests` requires at least one secure key. Because
  this datasource has no secrets, [`conformance_test.go`](conformance_test.go) runs the applicable subset of the
  conformance checks directly and skips only the secure-value assertion.
- `Validate` accepts the `none` (or empty) auth method for each service and has no credentials to require, so the
  default example is fully valid (unlike the secret-bearing entries).

## Settings examples

`""` (default) — both services with `auth.id: none`, no secrets.

## Upstream findings

- Hello is explicitly experimental ("WIP / Experiments required for testing the framework" in the monorepo
  README); its config surface is minimal by design.

## Validation performed

- dsconfig `Validate()` (via `ConfigSchemaValid`); ajv against `dsconfig/schema.json` (`--spec=draft7`) — valid.
- `go build/vet/gofmt/test -race` in `registry/` — clean; the adapted conformance suite passes.
- `tsc --noEmit --strict` on `settings.ts` — clean.
