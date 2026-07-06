# Azure Monitor — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-azure-monitor-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (Grafana Enterprise 13.0.1)
- **New UI:** `http://localhost:64433/?path=/story/configeditor-datasourceconfigwizard--tab&args=pluginType:grafana-azure-monitor-datasource`
- **Method:** Playwright captured the legacy UI (screenshot `legacy-expand-core2-azure.png`) and drove the new UI (local schema served via `context.route(...)`).
- **Result:** **Parity already achieved — no `dsconfig.json` changes required.**

---

## Findings (no fixes needed)

**No Custom HTTP Headers.** Azure Monitor is a cloud-auth datasource (Azure AD / App
Registration / Managed Identity). Its legacy "Additional settings → Advanced HTTP settings"
section contains only **Allowed cookies** and **Timeout** — there is **no** Custom HTTP Headers
control. (The "Add" button observed in the expanded capture is the _Allowed cookies_ TagsInput,
not headers.) So headers are correctly **not** modeled.

**No `fileUpload`.** Authentication uses Directory (tenant) ID, Application (client) ID, and
Client Secret (or Managed Identity / Workload Identity / client certificate as a textarea). The
legacy UI has **0 file inputs / 0 upload buttons** → `fileUpload` is correctly **not** used.

**No `required:true` fix.** Azure has **no** `requiredWhen:"true"` fields. Its required fields —
Tenant ID, Client ID, Client Secret — are _conditionally_ required on the auth type (App
Registration / Client Secret) via `requiredWhen: "jsonData_authType == '...'"`, which is correct.
The legacy UI marks them required (`*`) only for that auth type. No change.

## New UI verification

Tab mode renders all sections: **Authentication**, **Azure Monitor**, **Deprecated Application
Insights / Log Analytics**, **Legacy top-level credentials** (opt), **Customized cloud** (opt).
`hasHeadersEditor:false` (correct), `urlPresent:false` (Azure has no URL field — correct).

## Field-by-field parity (highlights)

| Legacy field              | schema id                                  | Target           | Status             |
| ------------------------- | ------------------------------------------ | ---------------- | ------------------ |
| Authentication type       | `jsonData_azureAuthType` (discriminator)   | `jsonData`       | ✅ 🔀              |
| Azure Cloud               | `jsonData_cloudName` (or customized cloud) | `jsonData`       | ✅                 |
| Directory (tenant) ID     | tenantId field                             | `jsonData`       | ✅ 🔀 (App Reg)    |
| Application (client) ID   | clientId field                             | `jsonData`       | ✅ 🔀 (App Reg)    |
| Client Secret             | secret field                               | `secureJsonData` | ✅ 🔀 (App Reg)    |
| Default Subscription      | subscription field                         | `jsonData`       | ✅                 |
| Enable Basic Logs         | basicLogsEnabled                           | `jsonData`       | ✅                 |
| Allowed cookies / Timeout | keepCookies / timeout                      | `jsonData`       | ✅ (Advanced HTTP) |

(The full auth-method matrix — Managed Identity, Workload Identity, Current User, App
Registration client-secret/certificate — is modeled via the auth discriminator + conditional
fields; all render per their `dependsOn`.)

## Conditional fields — tested

The auth-type discriminator drives the credential fields (tenant/client/secret for App
Registration; none for Managed Identity; etc.). Legacy top-level credentials and Customized cloud
are optional/collapsed groups. All conditionals render per the schema.

---

## Verification

```
go test ./registry/grafana-azure-monitor-datasource/...   # 8/8 conformance subtests PASS
```

No schema edit was made, so no regeneration was needed; the committed artifacts remain in sync
and the full `./registry/... ./schema/...` suite passes.

## Files changed

**None.** Azure Monitor was already at parity for HTTP headers (n/a), file upload (n/a), and the
required/General-step behaviour (conditional required is correct). This report documents the
validation only.
