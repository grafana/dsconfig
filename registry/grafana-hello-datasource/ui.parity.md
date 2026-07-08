# Hello (grafana-hello-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-hello-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (`<uid>` == plugin id; provisioned/read-only, but every field renders).
- **New UI:** Storybook `configeditor-datasourceconfigwizard--tab`, `args=pluginType:grafana-hello-datasource` (the local `dsconfig.json` was served by intercepting the remote schema fetch with Playwright `context.route(...)`).
- **Method:** Both captured with Playwright. The new UI is a stepper — each group is clicked and its fields unioned.
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

## Findings (no fixes needed)

- **Experimental/test plugin, no secrets.** Hello is a framework test plugin with **two fixed-URL services** — `httpbin` (https://httpbin.org) and `postman_echo` (https://postman-echo.com) — each with a single-option **No Auth** selector and no secure fields. A working datasource needs no configuration at all (`{}` is valid).
- **No Custom HTTP Headers.** `hasCustomHeaders:false`, `addHeaderBtn:false` in **both**. Correctly not modeled.
- **No `fileUpload`.** `fileInputs:0` in **both**. There are no secrets to upload.
- **Required-field handling.** No `required:true` / `requiredWhen` fields — both selectors default to `none` (the backend default), so nothing is required.

## Conditional fields & auth selector — grouping difference (cosmetic)

Each service carries a single-option **No Auth** discriminator:

- `jsonData_services_httpbin_auth_id` (default `none`)
- `jsonData_services_postman_echo_auth_id` (default `none`)

The **legacy** editor renders these as **per-service subsections** — "HTTPBin" and "Postman echo", each with its own *No Authentication* selector and the note "Data source is available without authentication". The **new UI** renders both **No Auth** selectors together under one **Authentication** group. This is a purely **cosmetic grouping difference**: both controls are present and modeled in both UIs (`radios:0`, `switches:0`, `selects:0` in both — the selectors are single-option Grafana `Select` components). No control is missing.

## Field-by-field parity

| Legacy field (subsection)          | schema id                              | Target     | Status                        |
| ---------------------------------- | -------------------------------------- | ---------- | ----------------------------- |
| No Authentication (HTTPBin)        | `jsonData_services_httpbin_auth_id`    | `jsonData` | ✅ selector (single option)   |
| No Authentication (Postman echo)   | `jsonData_services_postman_echo_auth_id` | `jsonData` | ✅ selector (single option)   |

Group observed in the new UI: **Authentication** (holds both No Auth selectors). Legacy `bodyText` shows both service subsections; new-UI `bodyText` shows "No Auth" for both selectors under one group.

## Verification

```
go test -count=1 ./registry/grafana-hello-datasource/...   # ok (TestSchemaConformance 8/8 subtests PASS)
```

No schema edit was made, so no regeneration was needed; the committed artifacts remain in sync.

## Files changed

**None.** Validation-only report; Hello was already at parity (headers n/a, fileUpload n/a, no required fields). The only difference is a cosmetic grouping of the two per-service *No Auth* selectors under a single Authentication group — both controls are present.
