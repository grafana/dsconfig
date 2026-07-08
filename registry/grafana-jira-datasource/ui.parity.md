# Jira (grafana-jira-datasource) — UI parity report

Parity validation between the **legacy Grafana datasource config editor** and the
**new schema-driven config UI** (`plugin-ui` `DatasourceConfigWizard`, rendered from
[`dsconfig.json`](dsconfig.json)).

- **Plugin id:** `grafana-jira-datasource`
- **Legacy UI:** `http://localhost:3000/connections/datasources/edit/<uid>` (Grafana Enterprise, provisioned instance)
- **New UI:** Storybook `configeditor-datasourceconfigwizard--tab`, `args=pluginType:grafana-jira-datasource` (local `dsconfig.json` served by intercepting the remote schema fetch with Playwright `context.route(...)`).
- **Method:** Playwright captured both UIs (rendered field labels, radios, switches, file inputs). The new UI is a stepper, so each group was clicked and the fields unioned.
- **Result:** **One `dsconfig.json` fix applied** — the legacy editor's **TLS settings** section (3 switches + the conditional certificate fields) was **missing** from the schema and has been added. No Custom HTTP Headers (n/a). No `fileUpload` (n/a). No `required:true` fix needed (all required fields are conditionally required on the auth method).

---

## The gap: TLS settings (standard Grafana TLS block)

The legacy Jira editor renders the shared Grafana TLS panel (via `@grafana/plugin-ui`
`convertLegacyAuthProps`) under a **TLS settings** section with the three standard toggles:

| Legacy label (TLS settings)      | reveals                                   |
| -------------------------------- | ----------------------------------------- |
| Add self-signed certificate      | CA Cert (`secureJsonData.tlsCACert`)      |
| TLS Client Authentication        | Server Name + Client Cert + Client Key    |
| Skip TLS certificate validation  | —                                         |

The capture showed the legacy UI with **4 switches** (Scoped Token + the 3 TLS toggles) while the
new UI had only **1** (Scoped Token) and no TLS group — the schema modeled none of the TLS fields.

**Fix (schema + settings model):** added the 7 standard TLS fields with their canonical Grafana
storage keys and a `TLS settings` group (mirroring the `plugin_sdk_settings` pack definitions and
the pattern already used by e.g. `grafana-dynatrace-datasource`). The Jira module bundle was
confirmed to reference these exact keys (`tlsAuth`, `tlsAuthWithCACert`, `tlsSkipVerify`,
`serverName`, `tlsCACert`, `tlsClientCert`, `tlsClientKey`), and the plugin's `Settings` struct
carries an `HttpClientOptions` the SDK HTTP client populates from them — so they are real storage,
not invented fields.

Added fields (`jsonData` unless noted):

- `jsonData_tlsAuthWithCACert` (switch) → `secureJsonData_tlsCACert` (textarea, `dependsOn`/`requiredWhen` = `tlsAuthWithCACert`)
- `jsonData_tlsAuth` (switch) → `jsonData_serverName` (input), `secureJsonData_tlsClientCert`, `secureJsonData_tlsClientKey` (textareas, `dependsOn`/`requiredWhen` = `tlsAuth`)
- `jsonData_tlsSkipVerify` (switch)

Because the conformance suite enforces schema↔struct parity, the four TLS `jsonData` keys were
added to the `Config` struct (`TLSAuth`, `TLSAuthWithCACert`, `TLSSkipVerify`, `TLSServerName`) and
the three TLS secrets to `SecureJsonDataKeys` (`tlsCACert`, `tlsClientCert`, `tlsClientKey`), with
the TS model (`settings.ts`) kept in sync.

### Verified end-to-end (new UI, after fix)

- New UI now renders the **TLS settings** (Optional) group with all three toggles:
  `Add self-signed certificate`, `TLS Client Authentication`, `Skip TLS certificate validation`.
- `radios:4` (Provider ×2 + Authentication method ×2) matches legacy; switches now render per the
  stepper (Scoped Token in **Authentication**, the 3 TLS toggles in **TLS settings**).
- No console errors.

---

## Field-by-field parity

| Legacy field (section)                             | new (schema id)                                            | Target           | Status                 |
| ------------------------------------------------- | ---------------------------------------------------------- | ---------------- | ---------------------- |
| Provider (Jira Cloud / Data Center)               | `jsonData_hosting`                                         | `jsonData`       | ✅ (radio)             |
| URL                                               | `jsonData_url`                                             | `jsonData`       | ✅ (required)          |
| Authentication method (Basic / OAuth 2.0)         | `jsonData_authMethod`                                      | `jsonData`       | ✅ (radio)             |
| User email / API Token                            | `jsonData_user` / `secureJsonData_token`                  | `jsonData` / sec | ✅ 🔀 (basicAuth)      |
| Scoped Token / Jira App Cloud Id                  | `jsonData_scopedToken` / `jsonData_cloudId`               | `jsonData`       | ✅ 🔀                  |
| Client ID / Client Secret                         | `jsonData_oauthClientID` / `secureJsonData_oauthClientSecret` | `jsonData`/sec | ✅ 🔀 (oauth2)         |
| **Add self-signed certificate** → CA Cert         | `jsonData_tlsAuthWithCACert` / `secureJsonData_tlsCACert`  | `jsonData`/sec   | ➕ added               |
| **TLS Client Authentication** → SNI/Cert/Key      | `jsonData_tlsAuth` / `jsonData_serverName` / `secureJsonData_tlsClientCert` / `secureJsonData_tlsClientKey` | `jsonData`/sec | ➕ added |
| **Skip TLS certificate validation**               | `jsonData_tlsSkipVerify`                                   | `jsonData`       | ➕ added               |

`jsonData.enableSecureSocksProxy` (Secure Socks Proxy checkbox) remains intentionally excluded
(AGENTS.md exclusion), consistent with the rest of the registry.

## HTTP headers / fileUpload — not applicable

- **HTTP headers:** the legacy Jira editor has no Custom HTTP Headers section (`hasCustomHeaders:false`, `addHeaderBtn:false`); the new UI confirms `headersEditor:false`. Not modeled.
- **fileUpload:** credentials/certs are entered as text/textarea; legacy `fileInputs:0`. Not used.

## Verification

```
go generate ./registry/grafana-jira-datasource/...
go test ./registry/grafana-jira-datasource/...   # PASS
```

`TestSchemaConformance` 8/8 subtests PASS (including `JSONDataMatchesStruct`,
`JSONDataTypesMatchStruct`, and `SecureValuesMatchLoadSettings` after the struct/secret-key
additions); `TestLoadConfig`, `TestApplyDefaults`, and `TestValidate` all PASS. Artifacts
regenerated (`schema.gen.json`, `settings.gen.json`) and back in sync.

## Files changed

- [`dsconfig.json`](dsconfig.json) — added the 7 standard TLS fields + the `TLS settings` group (and a TLS client-auth relationship / TLS instructions).
- [`settings.go`](settings.go) — added the 4 TLS `jsonData` fields to `Config` and the 3 TLS secret keys to `SecureJsonDataKeys`.
- [`settings.ts`](settings.ts) — mirrored the TLS `jsonData` fields and secret keys.
- `schema.gen.json`, `settings.gen.json` — regenerated.
