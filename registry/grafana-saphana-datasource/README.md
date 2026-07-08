# grafana-saphana-datasource — dsconfig registry entry

Declarative configuration schema for the **SAP HANA®** Grafana datasource plugin
(`grafana-saphana-datasource`).

## Files

| File | Purpose |
| --- | --- |
| `dsconfig.json` | dsconfig v1 schema — the single source of truth for the config fields |
| `settings.ts` | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| `settings.go` | Flat Go `Config` (jsonData + `DecryptedSecureJSONData`), typed secret keys, `LoadConfig`/`ApplyDefaults`/`Validate` |
| `schema.go` | k8s-style SDK `PluginSchema`: embeds `dsconfig.json` + `SettingsExamples` |
| `conformance_test.go` | Wraps `schema.RunPluginTests` (schema round-trip, spec/secure separation, jsonData↔struct parity, artifact drift) |
| `settings_test.go` | `LoadConfig` / `ApplyDefaults` / `Validate` / `SettingsExamples` behavior tests |
| `schema.gen.json`, `settings.gen.json`, `settings.examples.gen.json` | Generated artifacts (`go generate ./...`) |

Import path: `github.com/grafana/dsconfig/registry/grafana-saphana-datasource` (package `saphanadatasource`).

## Sources researched

Researched against the on-disk monorepo **`github.com/grafana/plugins-private`** at commit
**`267f4937806ed6404b6628d13ae358a5d308e376`**, plugin path
`plugins/grafana-saphana-datasource/`.

Reproduce:

```bash
git -C <plugins-private> fetch origin
git -C <plugins-private> checkout 267f4937806ed6404b6628d13ae358a5d308e376
# then read plugins/grafana-saphana-datasource/{src,pkg}
```

| Source | What it provides |
| --- | --- |
| `src/plugin.json:3-4,23-28` | `id` (`grafana-saphana-datasource` → `pluginType`), `name` (`SAP HANA®`), docs URL |
| `src/types.ts:9-29` | `HANAConfig` (jsonData) and `HANASecureConfig` (secureJsonData) |
| `src/selectors.ts:2-74` | Every editor label / placeholder / tooltip (`Components.ConfigEditor`) — the verbatim source for `label`, `ui.placeholder`, `description` |
| `src/components/ConfigEditor.tsx` | Field composition, section headings, conditional rendering, the inverted TLS switch, and the `onTLSSettingsChange` reset side-effect |
| `src/components/ui/CertificationKey.tsx:12-27` | The plugin-local textarea (rows=7) used for the three TLS cert/key secrets (placeholder only, no tooltip) |
| `pkg/models/settings.go:18-98` | Backend `Settings` struct, `IsValid`, `LoadSettings`, `Timeout`/`ParseInt` |
| `pkg/models/errors.go:5-13` | Validation error messages mirrored in `settings.go` |
| `pkg/models/settings_test.go` | Backend parse/validation expectations |
| `pkg/plugin/driver.go:37-107` | `GetTLSConfig` / `GetConnection` — how every setting is consumed (server:port, `3<instance>13`, TLS, default schema) |

### External components (versions from `.yarnrc.yml` `catalog:`; `@grafana/*` pinned `catalog:` in `package.json`)

| Package | Version (catalog) | Used for |
| --- | --- | --- |
| `@grafana/data` | `^11.6.7` | `DataSourceJsonData` (base of `HANAConfig`), `onUpdateDatasourceJsonDataOption`, `onUpdateDatasourceSecureJsonDataOption`, `DataSourcePluginOptionsEditorProps` |
| `@grafana/schema` | `^11.6.7` | `DataSourceJsonData` type import in `src/types.ts:2` |
| `@grafana/ui` | `^11.6.7` | `LegacyForms.FormField`, `LegacyForms.SecretFormField` (password), `Switch`, `InlineFormLabel`, `InlineField` |
| `@grafana/runtime` | `^11.6.7` | `config` feature-toggle gate for the (excluded) Secure Socks Proxy toggle |
| `@grafana/e2e-selectors` | (intentionally not cataloged) | `E2ESelectors` type in `src/selectors.ts` |
| `github.com/grafana/grafana-plugin-sdk-go` | `v0.280.0` (plugin `go.mod:9`) | `backend`, `backend/proxy` (`proxy.Options`), `data` |
| `github.com/grafana/sqlds/v5` | (plugin `go.mod`) | `DriverSettings` |
| `github.com/SAP/go-hdb` | (plugin `go.mod`) | `driver` connector used in `driver.go` |

## Where the config types are defined

**Frontend**

- `HANAConfig` (jsonData) — `src/types.ts:9-22`, extends `DataSourceJsonData`.
- `HANASecureConfig` (secureJsonData) — `src/types.ts:24-29`.
- `DataSourceJsonData` (base type) — `@grafana/schema` `^11.6.7`.

**Backend**

- `Settings` (flat struct, all fields) — `pkg/models/settings.go:18-35`.
- `proxy.Options` (`Settings.ProxyOptions`, Secure Socks Proxy — excluded here) —
  `github.com/grafana/grafana-plugin-sdk-go/backend/proxy` `v0.280.0`.

## Field inventory

All fields live in `jsonData` or `secureJsonData`; the backend reads **no** root-level fields.

| Schema ID | Storage key | Target | Editor label | Read by backend |
| --- | --- | --- | --- | --- |
| `jsonData_server` | `server` | jsonData | Server address | Yes — `driver.go:68,73,82`; TLS ServerName `driver.go:40` |
| `jsonData_port` | `port` | jsonData | Server port | Yes — `driver.go:68` |
| `jsonData_username` | `username` | jsonData | Username | Yes — `driver.go:82` |
| `secureJsonData_password` | `password` | secureJsonData | Password | Yes — `settings.go:58-62`, `driver.go:82` |
| `jsonData_tlsDisabled` | `tlsDisabled` | jsonData | TLS | Yes — `driver.go:88` |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | jsonData | Skip TLS Verify | Yes — `driver.go:39` |
| `jsonData_tlsAuth` | `tlsAuth` | jsonData | TLS Client Auth | Yes — `driver.go:42,50,76,81` |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | secureJsonData | Client Cert | Yes — `settings.go:66-68`, `driver.go:51,77` |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | secureJsonData | Client Key | Yes — `settings.go:69-71`, `driver.go:51,77` |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | jsonData | With CA Cert | Yes — `driver.go:42-43` |
| `secureJsonData_tlsCACert` | `tlsCACert` | secureJsonData | CA Cert | Yes — `settings.go:63-65`, `driver.go:43-48` |
| `jsonData_databaseName` | `databaseName` | jsonData | Tenant database name | Yes — `driver.go:70,85-87` |
| `jsonData_instance` | `instance` | jsonData | Tenant instance number | Yes — `driver.go:70,73` |
| `jsonData_defaultSchema` | `defaultSchema` | jsonData | Default schema | Yes — `driver.go:101-103` |
| `jsonData_timeout` | `timeout` | jsonData | Timeout | Yes — `settings.go:72-74,96-98` |

### Frontend-only settings

None. Every editor-visible field is read by the backend.

### Backend-only settings

None. The backend `Settings` struct contains exactly the editor fields plus the four secrets
plus `ProxyOptions` (Secure Socks Proxy, sourced from `HTTPClientOptions`, excluded from this
entry per policy).

### Virtual fields

None. The inverted "TLS" switch maps 1:1 to the stored `tlsDisabled` boolean, so it is modeled
as a plain storage field (with `effects`) rather than a virtual selector.

### Excluded

- `jsonData.enableSecureSocksProxy` (Secure Socks Proxy) — excluded from registry entries per
  policy (`ConfigEditor.tsx:304-335`, backend `settings.go:76-80`).

## Modeling decisions

- **No root fields.** `RootConfig` is a blank object (`Record<string, never>`) and `Config`
  carries no `json:"-"` root fields — `LoadSettings` never reads `settings.URL`/`BasicAuth`/etc.
- **Inverted TLS switch → plain storage field.** The editor's "TLS" switch renders
  `!jsonData.tlsDisabled` (`ConfigEditor.tsx:173`), i.e. switch ON = TLS enabled =
  `tlsDisabled:false` (the default). We store `jsonData_tlsDisabled` (boolean, default `false`)
  and attach an `effects` block replicating `onTLSSettingsChange`'s reset of
  `tlsSkipVerify`/`tlsAuth`/`tlsAuthWithCACert` when TLS is disabled (`ConfigEditor.tsx:54-58`).
- **`instance` is `string`, not `number`.** Despite `HANAConfig.instance?: number`
  (`types.ts:13`) and the `type="number"` input, the editor writes it with the generic
  `onUpdateDatasourceJsonDataOption` (`ConfigEditor.tsx:269`, no numeric coercion) and the
  backend reads it as `string` (`settings.go:21`), concatenating it into the port as
  `3<instance>13` (`driver.go:73`). The schema and Go `Config` model it as `string`.
- **`port` is `number`.** The editor coerces it with `+port` (`ConfigEditor.tsx:40-48`); backend
  is `int64` (`settings.go:30`).
- **`dependsOn` = editor visibility, `requiredWhen` = backend contract.**
  - `jsonData_server` `requiredWhen: "true"` (always required — `IsValid` `settings.go:38-40`).
  - `jsonData_port` `requiredWhen: "jsonData_instance == '' || jsonData_databaseName == ''"`
    (backend rejects `Port == 0 && (Instance == "" || DatabaseName == "")` — `settings.go:41-43`).
  - `jsonData_username` / `secureJsonData_password` `requiredWhen: "jsonData_tlsAuth != true"`
    (`settings.go:44-49`).
  - TLS sub-fields `dependsOn` on `jsonData_tlsDisabled != true` (+`tlsAuth`/`tlsAuthWithCACert`)
    mirror the editor's conditional render (`ConfigEditor.tsx:178,206,224,238`).
- **`Validate` adds the X.509 connect-time contract.** `IsValid` does not check that a client
  cert/key are present when `tlsAuth` is on, but `driver.go:77` (`NewX509AuthConnector`) /
  `driver.go:51` (`tls.X509KeyPair`) fail on empty input, so `Validate` requires
  `tlsClientCert`+`tlsClientKey` when `tlsAuth`. It does **not** require `tlsCACert` when
  `tlsAuthWithCACert` is on, because `GetTLSConfig` only appends the CA when
  `len(TlsCACert) > 0` and otherwise silently uses system roots (`driver.go:43`).
- **Secrets in a map.** `password`, `tlsCACert`, `tlsClientCert`, `tlsClientKey` are held in
  `Config.DecryptedSecureJSONData` (write-only in Grafana); the upstream `Settings` keeps them as
  flat fields, but we follow the registry convention (and avoid replicating the upstream
  `json:"-,omitempty"` quirk — see below).
- **No `help` drawers.** The editor exposes only tooltips (no collapse/side panels), so no field
  carries a `help` object; tooltips map to `description`.
- **Roles** applied where meaning matches: `endpoint.domain` (server), `endpoint.port` (port),
  `auth.basic.username`/`auth.basic.password`, `transport.tlsSkipVerify`, `tls.clientCert`/
  `tls.clientKey`/`tls.caCert`, `transport.timeoutSeconds` (timeout). `tlsDisabled`, `tlsAuth`,
  `tlsAuthWithCACert`, `databaseName`, `instance`, `defaultSchema` have no matching role.

## Settings examples matrix

| Example key | Auth | Connection | TLS mode | secureJsonData keys |
| --- | --- | --- | --- | --- |
| `""` (default) | basic (empty) | port (empty) | TLS on | `password` (empty) |
| `basicAuthPort` | basic | host + port 443 | TLS on | `password` |
| `basicAuthInstance` | basic | instance + database | TLS on | `password` |
| `tlsClientAuth` | X.509 client cert | host + port 443 | TLS on, mutual | `tlsClientCert`, `tlsClientKey` |
| `tlsWithCACert` | basic | host + port 443 | TLS on, custom CA | `password`, `tlsCACert` |
| `tlsSkipVerify` | basic | host + port 443 | TLS on, skip verify | `password` |
| `tlsDisabled` | basic | host + port | TLS off (plaintext) | `password` |

All secret values are obviously-fake angle-bracket placeholders (`<your-password>`) or redacted
PEM blocks (`-----BEGIN CERTIFICATE-----\n<redacted>\n-----END CERTIFICATE-----`). The `""`
default example intentionally fails `LoadConfig`'s `Validate` (empty server/password).

## Potential upstream bugs / discrepancies

1. **`instance` type mismatch.** `HANAConfig.instance` is typed `number` (`types.ts:13`) and the
   editor uses a `type="number"` input, but it is stored and consumed as a **string**
   (`settings.go:21`; `driver.go:73` string-concatenates `3<instance>13`). Providing a
   JSON-number `instance` in provisioning would fail to unmarshal into the backend `string` field.
2. **`Password` json tag.** `Password string \`json:"-,omitempty"\`` (`settings.go:25`) names the
   field `"-"` rather than skipping it (the skip form is `json:"-"`). Harmless in practice because
   the password is only ever populated from `DecryptedSecureJSONData["password"]`, never from
   jsonData. This entry omits the field from `Config` and reads the secret from the map.
3. **`LoadSettings` early return.** When no password is supplied and `tlsAuth` is false, the
   backend returns `IsValid()` **before** defaulting the timeout, copying the TLS certs, or
   loading proxy options (`settings.go:58-61`). Since that branch always yields a validation
   error, this entry's `parse → ApplyDefaults → Validate` flow returns an equivalent error.
4. **Editor required markers are stricter/looser than the backend.** The password field is marked
   `required` unconditionally (`ConfigEditor.tsx:150`) even though the backend allows an empty
   password under TLS client auth; `port` and `instance` are marked `required={!databaseName}`
   (`ConfigEditor.tsx:122,268`), a simpler heuristic than the backend's
   "port OR (instance AND databaseName)" contract. `requiredWhen` encodes the backend contract.
5. **"With CA Cert" without a CA cert is silently tolerated.** `GetTLSConfig` only adds the CA to
   the pool when `len(TlsCACert) > 0` (`driver.go:43`); enabling `tlsAuthWithCACert` with no
   `tlsCACert` falls back to system roots rather than erroring.
6. **No dedicated TLS server-name field.** The TLS `ServerName` is derived from `jsonData.server`
   (`driver.go:40`); there is no separate server-name input, so none is modeled.
7. **Module path vs plugin id.** The Go module is `github.com/grafana/saphana-datasource`
   (`driver.go` imports) while the plugin id is `grafana-saphana-datasource`; `pluginType` uses
   the plugin id (`plugin.json:4`), as required.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + schema round-trip, spec/secure separation, jsonData↔`Config`
  parity, secure-key parity, and artifact-drift checks via `schema.RunPluginTests`
  (`conformance_test.go`) — **pass**.
- Strict JSON Schema validation of `dsconfig.json` against `dsconfig/schema.json` — **0 errors**
  under both the declared draft-07 validator and a Draft 2020-12 validator
  (`additionalProperties:false` enforced).
- `go generate ./...` (regenerates `*.gen.json`) — **pass**.
- From `registry/`: `gofmt -l .` (clean), `go vet ./...`, `go build ./...`, `go test ./...` — **pass**.
- `tsc --noEmit --strict settings.ts` (TypeScript `5.5.4`) — **pass**.
- Pre-existing `dsconfig` and `schema` workspace modules still build and test — **pass**.
