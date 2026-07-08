# Hello World (grafana-helloworld-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-helloworld-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (`<uid>` == plugin id; provisioned/read-only, but every field renders).
- **New UI:** Storybook `configeditor-datasourceconfigwizard--tab`, `args=pluginType:grafana-helloworld-datasource` (the local `dsconfig.json` was served by intercepting the remote schema fetch with Playwright `context.route(...)`).
- **Method:** Both captured with Playwright. The new UI is a stepper — each group is clicked and its fields unioned.
- **Result:** **Intentional, documented DIVERGENCE — not a bug, no change required.** The new UI renders one placeholder field that the legacy UI does not. See below.

> ⚠️ **Divergence (expected, not a defect).** The legacy config editor for Hello World renders **zero** configuration fields — only the static text **"Hello World Config Editor!"** (`bodyText` confirms `0` inputs, `0` labels, `0` headings). The new UI renders **one** field — **API key** (`secureJsonData_apiKey`, password input) under a default **Configuration** group. This field is a **conformance artifact, not a real credential**, and is deliberately retained. Do **not** treat it as a parity gap to fix.

## Why the divergence exists (and why it stays)

Hello World is an SDK **sample/template** plugin with **no configuration surface**:

- Its config editor renders static text and never calls `onOptionsChange`; its frontend config types are blank (`Config = {}`, `SecureConfig = {}`).
- Its backend ignores instance settings entirely — the instance factory returns an empty struct without reading `URL`, `jsonData`, or any secret.

The single `secureJsonData_apiKey` ("API key") field exists **only** to satisfy two hard constraints, both documented in the schema `instructions`:

1. `dsconfig` `Schema.Validate` rejects an empty `fields` array ("fields is required"), so the entry must declare **at least one field**.
2. The shared conformance suite requires **at least one `secureJsonData` key** (`schema.PluginUnderTest` rejects empty `SecureKeys`; `SchemaRoundTrip` asserts `SecureValues` is non-empty).

The backend **never reads** `apiKey` — setting or omitting it has no effect. There is **no honored "hidden" flag** in the schema to suppress a field from the wizard, so the placeholder necessarily renders. This is the expected state; leaving it in place is correct.

## Findings (no fixes needed)

- **No Custom HTTP Headers.** `hasCustomHeaders:false`, `addHeaderBtn:false` in **both**. Correctly not modeled.
- **No `fileUpload`.** `fileInputs:0` in **both**.
- **Required-field handling.** `secureJsonData_apiKey` is not marked required (no `required:true` / `requiredWhen`); it is an inert placeholder.

## Field-by-field parity

| Legacy field | schema id                 | Target           | Status                                              |
| ------------ | ------------------------- | ---------------- | --------------------------------------------------- |
| _(none)_     | `secureJsonData_apiKey`   | `secureJsonData` | ⚠️ new-UI-only placeholder (conformance artifact)   |

Legacy `bodyText`: "…Hello World Config Editor! Delete Test" (no fields). New-UI `bodyText`: "Configuration API key Save & Test". `groupTitles` is empty because the schema declares no groups, so the wizard falls back to a default **Configuration** group.

## Verification

```
go test -count=1 ./registry/grafana-helloworld-datasource/...   # ok (TestSchemaConformance 8/8 subtests PASS — incl. the SecureValues / SchemaRoundTrip checks that require the placeholder)
```

No schema edit was made, so no regeneration was needed; the committed artifacts remain in sync.

## Files changed

**None.** Validation-only report. The extra **API key** field in the new UI is an **intentional, explained divergence** (a conformance artifact required by the dsconfig validator and shared test suite, ignored by the backend), documented above rather than fixed.
