# grafana-atlassianstatuspage-datasource

Configuration schema for the [Atlassian Statuspage datasource plugin](https://grafana.com/docs/plugins/grafana-atlassianstatuspage-datasource)
(`grafana-atlassianstatuspage-datasource`), which lives in the [grafana/plugins](https://github.com/grafana/plugins)
monorepo. It has no hand-written `ConfigEditor.tsx` or per-plugin backend `Settings` model — both come from
shared packages (`@grafana/declarative-plugin` frontend, `github.com/grafana/plugins/sdk/pluginspec` backend)
specialized by the plugin's `src/spec.ts`. See [`grafana-zendesk-datasource`](../grafana-zendesk-datasource/README.md)
for the fullest description of the shared config/storage model.

**This datasource queries the public Statuspage API and has no authentication, so it stores no
`secureJsonData` secrets.**

## Upstream researched

- **Repo**: `github.com/grafana/plugins` (`plugins/grafana-atlassianstatuspage-datasource/`)
- **Ref**: `main` — **Commit**: `4b176ec1f74d80c231be2aeb1ce4713c833a76af` (2026-07-02)

## Sources researched (grafana/plugins @ 4b176ec)

| File | What was read |
| --- | --- |
| `plugins/grafana-atlassianstatuspage-datasource/src/plugin.json` | `id`/`name` → `pluginType`/`pluginName`; Docs link → `docURL` |
| `plugins/grafana-atlassianstatuspage-datasource/src/spec.ts:8-29,130-147` | service `atlassianstatuspage`; server `client_api` (`{url}/api/v2`, var `url`); `authMethods: {}` and `server.authMethods: []` (no auth) |
| `packages/declarative-plugin/.../rest/ServiceConfig.tsx:63,90` | the Auth section is only rendered when the server has auth methods (it does not here) |

## Field inventory

| Schema `id` | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_variables_url` | `variables.url` | `jsonData` | URL | Yes — substituted into `{url}/api/v2` |

There are **no** secure fields. **Not modeled:** single-server `services.atlassianstatuspage.server.id`
(backend-defaults), per-service `disabled`, `enableSecureSocksProxy` (policy).

## Where the types are defined

- Frontend: `Spec`/`Config` in `@grafana/declarative-plugin`; service/server/variable data in the plugin's `src/spec.ts`.
- Backend: `Spec` in `sdk/pluginspec/pluginspec.go`; `JsonData` in `sdk/pluginspec/pluginclient/config.go`.

`Config` in [`settings.go`](settings.go) projects the SDK's generic map-based `JsonData` into a concrete struct
holding only the `url` variable.

## Modeling decisions

- No authentication ⇒ no `secureJsonData` and no auth discriminator field; the only configurable field is the
  `url` variable, marked required (base URL).
- **Conformance test deviation:** the shared `schema.RunPluginTests` requires at least one secure key (its
  `SchemaRoundTrip` subtest asserts a non-empty `SecureValues`). Because this datasource has no secrets,
  [`conformance_test.go`](conformance_test.go) runs the applicable subset of the conformance checks directly and
  skips only the secure-value assertion; all other guard rails still apply.
- `ApplyDefaults` is a no-op (no discriminators); kept exported for API parity with the other entries.

## Settings examples

`""` (default, empty `url`) and `configured` (populated). Examples carry only `jsonData` (no secrets).

## Upstream findings

- The Statuspage API is public, so the plugin defines no auth methods; a datasource is functional with only the
  `url` set.
- `url` is shown required in the editor but not enforced by the backend's `validateVariables` (the spec's ref
  lacks `required: true`); `Validate` requires it since the base URL is malformed without it.

## Validation performed

- dsconfig `Validate()` (via `ConfigSchemaValid`); ajv against `dsconfig/schema.json` (`--spec=draft7`) — valid.
- `go build/vet/gofmt/test -race` in `registry/` — clean; the adapted conformance suite passes.
- `tsc --noEmit --strict` on `settings.ts` — clean.
