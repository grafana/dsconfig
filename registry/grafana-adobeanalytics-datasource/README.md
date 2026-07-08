# grafana-adobeanalytics-datasource

Configuration schema for the [Adobe Analytics datasource plugin](https://grafana.com/docs/plugins/grafana-adobeanalytics-datasource)
(`grafana-adobeanalytics-datasource`), which lives in the [grafana/plugins](https://github.com/grafana/plugins)
monorepo. It has no hand-written `ConfigEditor.tsx` or per-plugin backend `Settings` model — both come from
shared packages (`@grafana/declarative-plugin` frontend, `github.com/grafana/plugins/sdk/pluginspec` backend)
specialized by the plugin's `src/spec.ts`. See [`grafana-zendesk-datasource`](../grafana-zendesk-datasource/README.md)
for the fullest description of the shared config/storage model.

## Upstream researched

- **Repo**: `github.com/grafana/plugins` (`plugins/grafana-adobeanalytics-datasource/`)
- **Ref**: `main` — **Commit**: `4b176ec1f74d80c231be2aeb1ce4713c833a76af` (2026-07-02)

## Sources researched (grafana/plugins @ 4b176ec)

| File | What was read |
| --- | --- |
| `plugins/grafana-adobeanalytics-datasource/src/plugin.json` | `id`/`name` → `pluginType`/`pluginName`; Docs link → `docURL` |
| `plugins/grafana-adobeanalytics-datasource/src/spec.ts:7-35,357-371` | service `adobe_analytics`; server `adobeanalytics_api` (`https://analytics.adobe.io/api/{global_company_id}`, var `global_company_id`); auth `oauth2_m2m` (type `oauth2_client_credentials`, `tokenUrl`, `scopes`, no `clientId` → client-id field shown) |
| `sdk/pluginspec/pluginclient/pluginclient.go:59-100` | oauth2: `tokenUrl`/`clientId` sourced from spec or config; `clientSecret` read as `DecryptedSecureJSONData["adobe_analytics.clientSecret"]` |
| `sdk/pluginspec/pluginclient/serviceclient.go:268-274` | oauth2 requires non-empty clientId and clientSecret |
| `packages/declarative-plugin/.../rest/auth/OAuthClientCredentials.tsx:52-85` | Client ID input (label "Client ID") shown when the auth method has no fixed clientId; Client Secret secret (label "Client Secret") |

## Field inventory

| Schema `id` | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_variables_global_company_id` | `variables.global_company_id` | `jsonData` | Global Company ID | Yes — substituted into the base URL (required) |
| `jsonData_services_adobe_analytics_auth_id` | `services.adobe_analytics.auth.id` | `jsonData` | (auth selector) | Yes — discriminator; defaults to `oauth2_m2m` |
| `jsonData_services_adobe_analytics_auth_clientId` | `services.adobe_analytics.auth.clientId` | `jsonData` | Client ID | Yes — OAuth2 client id |
| `secureJsonData_adobe_analytics_clientSecret` | `adobe_analytics.clientSecret` | `secureJsonData` | Client Secret | Yes — OAuth2 client secret |

**Not modeled:** single-server `services.adobe_analytics.server.id` (backend-defaults), per-service `disabled`,
`enableSecureSocksProxy` (policy). The token URL and scopes are fixed in the spec (not user-configurable, so not
stored fields).

## Where the types are defined

- Frontend: `Spec`/`Config`/`SecureConfig` in `@grafana/declarative-plugin`; service/server/auth/variable data in the plugin's `src/spec.ts`.
- Backend: `Spec`/`AuthMethod` in `sdk/pluginspec/pluginspec.go`; `JsonData`/`ServiceConfig` in `sdk/pluginspec/pluginclient/config.go`.

`Config` in [`settings.go`](settings.go) projects the SDK's generic map-based `JsonData` into concrete structs.

## Modeling decisions

- OAuth2 client-credentials → a `clientId` jsonData field (`auth.oauth2.clientId`) plus a `clientSecret` secret
  (`auth.oauth2.clientSecret`); the discriminator `auth.id` = `oauth2_m2m`.
- `global_company_id` is marked required: it forms the base URL and the editor renders it required.
- `LoadConfig` mirrors the lenient `pluginclient.New` parse; `Validate` requires clientId, clientSecret, and global_company_id.

## Settings examples

`""` (default, empty fields) and `oauth2ClientCredentials` (populated).

## Upstream findings

- The token URL (`https://ims-na1.adobelogin.com/ims/token/v3`) and OAuth scopes are hard-coded in the spec, so
  they are not user-configurable and are not modeled as stored fields.
- Credentials are not validated at instance creation (`pluginclient.New` is lenient); missing inputs fail at
  request time. `Validate` encodes the working-datasource contract.

## Validation performed

- dsconfig `Validate()` (via `ConfigSchemaValid`); ajv against `dsconfig/schema.json` (`--spec=draft7`) — valid.
- `go build/vet/gofmt/test -race` in `registry/` — clean; full conformance suite passes.
- `tsc --noEmit --strict` on `settings.ts` — clean.
