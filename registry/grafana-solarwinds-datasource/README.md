# grafana-solarwinds-datasource

Configuration schema for the [SolarWinds datasource plugin](https://grafana.com/docs/plugins/grafana-solarwinds-datasource)
(`grafana-solarwinds-datasource`), which lives in the [grafana/plugins](https://github.com/grafana/plugins)
monorepo. It has no hand-written `ConfigEditor.tsx` or per-plugin backend `Settings` model — both come from
shared packages (`@grafana/declarative-plugin` frontend, `github.com/grafana/plugins/sdk/pluginspec` backend)
specialized by the plugin's `src/spec.ts`. See [`grafana-zendesk-datasource`](../grafana-zendesk-datasource/README.md)
for the fullest description of the shared config/storage model.

## Upstream researched

- **Repo**: `github.com/grafana/plugins` (`plugins/grafana-solarwinds-datasource/`)
- **Ref**: `main` — **Commit**: `4b176ec1f74d80c231be2aeb1ce4713c833a76af` (2026-07-02)

## Sources researched (grafana/plugins @ 4b176ec)

| File | What was read |
| --- | --- |
| `plugins/grafana-solarwinds-datasource/src/plugin.json` | `id`/`name` ("Solarwinds") → `pluginType`/`pluginName`; Docs link → `docURL` |
| `plugins/grafana-solarwinds-datasource/src/spec.ts:8-55` | service `solarwinds`; server `api_server` (`{url}:17774/SolarWinds/InformationService/v3/Json`, var `url` `required: true`); basic auth `basic_auth` with `user`/`password` labels and `showTLSOptions: true` |
| `packages/declarative-plugin/.../config-editor/Auth.tsx:73-153,200-217` | basic auth (username jsonData + `<serviceId>.password` secret); TLS storage keys: `auth.tls.selfSignedCert.enabled` + `<serviceId>.tls.selfSignedCert`; `auth.tls.clientAuth.enabled`/`serverName` + `<serviceId>.tls.clientCert`/`clientKey`; `auth.tls.skipVerification` |
| `@grafana/plugin-ui` `Auth` component (`pui/package/dist/.../index.cjs`) | verbatim TLS labels/tooltips ("Add self-signed certificate", "CA Certificate", "TLS Client Authentication", "ServerName", "Client Certificate", "Client Key", "Skip TLS certificate validation") |
| `sdk/pluginspec/pluginclient/pluginclient.go:98-108` | TLS secrets read as `<serviceId>.tls.selfSignedCert`/`clientCert`/`clientKey` (only when the respective toggle is enabled) |

## Field inventory

| Schema `id` | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_variables_url` | `variables.url` | `jsonData` | URL | Yes — base URL (required) |
| `jsonData_services_solarwinds_auth_id` | `services.solarwinds.auth.id` | `jsonData` | (auth selector "Basic Auth") | Yes — discriminator; defaults to `basic_auth` |
| `jsonData_services_solarwinds_auth_username` | `services.solarwinds.auth.username` | `jsonData` | Username | Yes — basic auth username |
| `secureJsonData_solarwinds_password` | `solarwinds.password` | `secureJsonData` | Password | Yes — basic auth password |
| `jsonData_services_solarwinds_auth_tls_selfSignedCert_enabled` | `services.solarwinds.auth.tls.selfSignedCert.enabled` | `jsonData` | Add self-signed certificate | Yes |
| `secureJsonData_solarwinds_tls_selfSignedCert` | `solarwinds.tls.selfSignedCert` | `secureJsonData` | CA Certificate | Yes (when enabled) |
| `jsonData_services_solarwinds_auth_tls_clientAuth_enabled` | `services.solarwinds.auth.tls.clientAuth.enabled` | `jsonData` | TLS Client Authentication | Yes |
| `jsonData_services_solarwinds_auth_tls_clientAuth_serverName` | `services.solarwinds.auth.tls.clientAuth.serverName` | `jsonData` | ServerName | Yes |
| `secureJsonData_solarwinds_tls_clientCert` | `solarwinds.tls.clientCert` | `secureJsonData` | Client Certificate | Yes (when client auth enabled) |
| `secureJsonData_solarwinds_tls_clientKey` | `solarwinds.tls.clientKey` | `secureJsonData` | Client Key | Yes (when client auth enabled) |
| `jsonData_services_solarwinds_auth_tls_skipVerification` | `services.solarwinds.auth.tls.skipVerification` | `jsonData` | Skip TLS certificate validation | Yes |

**Not modeled:** single-server `services.solarwinds.server.id` (backend-defaults), per-service `disabled`,
`enableSecureSocksProxy` (policy).

## Where the types are defined

- Frontend: `Spec`/`Config`/`SecureConfig` in `@grafana/declarative-plugin`; the `Auth`/TLS control labels in `@grafana/plugin-ui`; service/server/auth/variable data in the plugin's `src/spec.ts`.
- Backend: `Spec`/`AuthMethod` in `sdk/pluginspec/pluginspec.go`; `JsonData`/`ServiceConfig` (incl. the nested `Auth.TLS` struct) in `sdk/pluginspec/pluginclient/config.go`.

`Config` in [`settings.go`](settings.go) projects the SDK's generic map-based `JsonData` into concrete structs, including the nested `Auth.TLS.{SelfSignedCert,ClientAuth}` blocks.

## Modeling decisions

- Basic auth → username jsonData field + `solarwinds.password` secret. TLS options (from the shared
  `@grafana/plugin-ui` `Auth` component, rendered because `showTLSOptions: true`) map to the nested
  `auth.tls.*` jsonData toggles and the `solarwinds.tls.*` secrets.
- TLS control labels/tooltips are taken verbatim from `@grafana/plugin-ui` (not from the plugin spec, which does
  not define them).
- `url` is marked required (base URL host; the spec's ref is `required: true`).
- `Validate` requires username + password + url, and — when a TLS toggle is enabled — the accompanying
  certificate(s)/key, mirroring how the backend reads those secrets only when the toggle is on.

## Settings examples

`""` (default), `basicAuth`, and `basicAuthMutualTLS` (basic auth + TLS client authentication).

## Upstream findings

- Credentials are not validated at instance creation (`pluginclient.New` is lenient); missing inputs fail at
  request time. `Validate` encodes the working-datasource contract.
- The TLS certificate secrets are read by the backend only when the matching toggle
  (`selfSignedCert.enabled` / `clientAuth.enabled`) is true.

## Validation performed

- dsconfig `Validate()` (via `ConfigSchemaValid`); ajv against `dsconfig/schema.json` (`--spec=draft7`) — valid.
- `go build/vet/gofmt/test -race` in `registry/` — clean; full conformance suite passes.
- `tsc --noEmit --strict` on `settings.ts` — clean.
