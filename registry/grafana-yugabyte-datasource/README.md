# grafana-yugabyte-datasource

Declarative configuration schema for the [Yugabyte datasource plugin](https://github.com/grafana/yugabyte-datasource) (`grafana-yugabyte-datasource`).

YugabyteDB is PostgreSQL wire-protocol compatible; the plugin uses [`jackc/pgx`](https://github.com/jackc/pgx) under the hood via `sqlds` and connects on the YSQL default port `5433`.

## Upstream researched

- **Repo**: `github.com/grafana/yugabyte-datasource`
- **Ref**: `main`
- **Commit SHA**: `d2c8a440caf53d87979a0a2f38491cafbebc7a5f` (2026-07-04)

Every value in [`dsconfig.json`](dsconfig.json) is traceable to a specific `file:line` in the upstream repo at this SHA.

```bash
git clone https://github.com/grafana/yugabyte-datasource
cd yugabyte-datasource
git checkout d2c8a440caf53d87979a0a2f38491cafbebc7a5f
```

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` (root URL/User + Connection + jsonData Database + `DecryptedSecureJSONData`), `LoadConfig` / `ApplyDefaults` / `Validate` |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema + 4 `SettingsExamples` |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, `Validate` |
| [`conformance_test.go`](conformance_test.go) | `schema.RunPluginTests` wrapper |
| `.gen.json` artifacts | Regenerate with `go generate ./...` |

## Sources researched

Read at the pinned upstream SHA:

| File | What was read |
| --- | --- |
| `src/plugin.json:2-9` | `pluginType` (`id: grafana-yugabyte-datasource`), `pluginName` (`Yugabyte`); `info.links[]` is empty (`:20`) |
| `src/types.ts:11-13` | `YugabyteOptions extends SQLOptions` (`@grafana/plugin-ui`) + `enableSecureSocksProxy?: boolean` |
| `src/types.ts:18-20` | `YugabyteSecureJsonData { password?: string }` |
| `src/components/ConfigEditor.tsx:29-35` | `DataSourceDescription dataSourceName="Yugabyte" docsLink=".../datasources/yugabyte/" hasRequiredFields={true}` |
| `src/components/ConfigEditor.tsx:39-57` | Connection section — `Host URL` (required, placeholder `localhost:5433`) + `Database` (required, placeholder `yb_demo`, stored in jsonData) |
| `src/components/ConfigEditor.tsx:61-80` | Authentication section — `Username` (required, root `user`, placeholder `yugabyte`) + `Password` (`SecretInput`, `secureJsonData.password`, placeholder `********`) |
| `src/components/ConfigEditor.tsx:82-84` | `Additional Settings` collapsible — `SecureSocksProxyToggle` (excluded from schema per policy) |
| `pkg/settings.go:11-16` | Backend `Settings` struct — `Connection` (derived), `User`, `Password`, `Database string json:"database"` |
| `pkg/settings.go:18-22` | `Connection` struct — `Url`, `Host`, `Port` (derived from `net.SplitHostPort`) |
| `pkg/settings.go:24-49` | `LoadSettings` — `net.SplitHostPort(s.URL)` first (fails on missing port), then jsonData unmarshal, then `Password = s.DecryptedSecureJSONData["password"]` |
| `pkg/settings.go:51-61` | `BuildConnectionString` — hardcodes `sslmode='allow'`; interpolates host/port/user/password/database as single-quoted libpq params |
| `pkg/driver.go:20-58` | How the settings are consumed: `LoadSettings` → `BuildConnectionString` → `pgx.ParseConfig` → optional Secure Socks Proxy dialer → `stdlib.OpenDB` |
| `pkg/main.go:15` | Confirms plugin id string `grafana-yugabyte-datasource` |

### External editor components

| Component | Version | Source | What was read |
| --- | --- | --- | --- |
| `DataSourceDescription`, `ConfigSection`, `SecureSocksProxyToggle` | `@grafana/plugin-ui@0.13.1` | [grafana/plugin-ui](https://github.com/grafana/plugin-ui) | Section headings + the (excluded) socks proxy switch (`SecureSocksProxyToggle.tsx:19` writes `jsonData.enableSecureSocksProxy`) |
| `Field`, `Input`, `SecretInput` | `@grafana/ui@13.1.0-27176567230` | grafana/grafana `packages/grafana-ui` | UI primitives |
| `onUpdateDatasourceJsonDataOption`, `onUpdateDatasourceSecureJsonDataOption`, `updateDatasourcePluginResetOption` | `@grafana/data@13.1.0-27176567230` | grafana/grafana `packages/grafana-data` | Editor write helpers |
| `SQLOptions` | `@grafana/plugin-ui@0.13.1` | grafana/plugin-ui `packages/plugin-ui/src/types/sql.ts` | Base jsonData shape — extended but only `database` is actually wired by the plugin |

## Field provenance

| Schema `id` | Storage key | Target | Editor label source | Placeholder | Read-by-backend |
| --- | --- | --- | --- | --- | --- |
| `root_url` | `url` | root | `ConfigEditor.tsx:40` (`<Field label="Host URL" required>`) | `:43` (`placeholder="localhost:5433"`) | Yes — `pkg/settings.go:25` (`net.SplitHostPort(s.URL)`) |
| `jsonData_database` | `database` | jsonData | `:49` (`<Field label="Database" required>`) | `:52` (`placeholder="yb_demo"`) | Yes — `pkg/settings.go:15,43` (`Settings.Database` populated by json.Unmarshal) |
| `root_user` | `user` | root | `:62` (`<Field label="Username" required>`) | `:65` (`placeholder="yugabyte"`) | Yes — `pkg/settings.go:38` (`Settings.User = s.User`) |
| `secureJsonData_password` | `password` | secureJsonData | `:71` (`<Field label="Password">`) — **not** marked required | `:74` (`placeholder="********"`) | Yes — `pkg/settings.go:39` (`s.DecryptedSecureJSONData["password"]`) |

## Frontend-only and backend-only settings

- **Frontend-only**: none — every schema field is read by the backend.
- **Backend-only**: none — every backend field surfaces in the editor.
- **Excluded (present in editor, deliberately not in schema)**: `jsonData.enableSecureSocksProxy` (`ConfigEditor.tsx:83`, `SecureSocksProxyToggle.tsx:19`, consumed at `pkg/driver.go:41-53`). Per AGENTS.md the Secure Socks Proxy field is omitted from every registry entry.

## Modeling decisions

- **Root-level URL and User on `Config`** because the backend reads them directly from `backend.DataSourceInstanceSettings` (`pkg/settings.go:25,38`). Both are tagged `json:"-"` to avoid colliding with jsonData unmarshal.
- **`Connection` struct on `Config`** mirrors `pkg/settings.go:18-22` verbatim so callers of `LoadConfig` get the same `Host`/`Port` split that the plugin's own `BuildConnectionString` uses.
- **`ApplyDefaults` is a no-op**: the plugin has no discriminator field (no auth-type selector, no license/plan, no TLS mode) so there is nothing to default. The method is kept exported for `LoadConfig` uniformity and for callers building a `Config` directly.
- **`Validate` covers URL/User/Database presence + `host:port` shape**: the editor's required markers give us `URL`, `User`, `Database`; the URL `host:port` check mirrors the backend's `SplitHostPort` contract so callers that skip `LoadConfig` still catch shape violations.
- **Password is NOT marked required** in either the editor or `Validate`. The password field in the editor has no `required` attribute (`ConfigEditor.tsx:71`) and Yugabyte deployments can be configured with `trust` authentication server-side. The backend interpolates whatever string sits in `secureJsonData.password` (empty or otherwise) into the connection string verbatim.
- **`Config.Database` mirrors the upstream field name**: the plugin's own `Settings` struct calls it `Database` (`pkg/settings.go:15`). We use the same name and `json:"database,omitempty"` tag to match the upstream shape.
- **No TLS fields at all** — the backend hardcodes `sslmode='allow'` in `BuildConnectionString` (`pkg/settings.go:52`). Diverges from `grafana-postgresql-datasource` which exposes `disable`/`require`/`verify-ca`/`verify-full` and paths/inline PEM certificates. See "Upstream findings" below.
- **`secureSocksProxy` example still surfaces the excluded field**: because the toggle is genuinely part of the plugin's stored jsonData shape (just outside the dsconfig registry contract), we ship one settings example demonstrating the wire format for provisioning callers — but no schema field for it.

## Where the types are defined

| Type | Where |
| --- | --- |
| `Settings` (backend, upstream) | `pkg/settings.go:11-16` in grafana/yugabyte-datasource |
| `Connection` (backend, upstream) | `pkg/settings.go:18-22` in grafana/yugabyte-datasource |
| `YugabyteOptions` (frontend, upstream) | `src/types.ts:11-13` in grafana/yugabyte-datasource |
| `YugabyteSecureJsonData` (frontend, upstream) | `src/types.ts:18-20` in grafana/yugabyte-datasource |
| `SQLOptions` (frontend, external) | `@grafana/plugin-ui@0.13.1` — extended but only `database` is wired |
| `SecureSocksProxyToggle` (frontend, external) | `@grafana/plugin-ui@0.13.1` `SecureSocksProxyToggle.tsx:19` — writes `jsonData.enableSecureSocksProxy` |
| `DataSourceDescription`, `ConfigSection` (frontend, external) | `@grafana/plugin-ui@0.13.1` |
| `Field`, `Input`, `SecretInput` (frontend, external) | `@grafana/ui@13.1.0-27176567230` |
| `Config` (this entry) | `settings.go` in this directory — flat mirror of upstream `Settings` |

## Settings examples matrix

| Key | Summary | Auth | URL | Extras |
| --- | --- | --- | --- | --- |
| `""` | Default configuration | Empty username + password | Empty | — |
| `localDev` | Local YugabyteDB (default port 5433) | `yugabyte` / `yugabyte` | `localhost:5433` | — |
| `remoteCluster` | Remote YugabyteDB cluster | `grafana_reader` / `changeme` | `yb.internal.example.com:5433` | — |
| `secureSocksProxy` | With Secure Socks Proxy enabled | `grafana_reader` / `changeme` | `yb.internal.example.com:5433` | `jsonData.enableSecureSocksProxy=true` (excluded from schema) |

## Upstream findings

1. **TLS is not configurable at all.** `pkg/settings.go:52` interpolates a hardcoded `sslmode='allow'` into the connection string. `allow` is a permissive libpq mode: pgx first tries plaintext and falls back to TLS only if the server refuses the plaintext handshake. There is no way from Grafana to require encryption, verify the server certificate, or supply a client certificate — a significant divergence from `grafana-postgresql-datasource`, which exposes all four sslmode variants plus file-path and inline-content TLS credentials.
2. **URL contract is unusual.** The backend calls `net.SplitHostPort(s.URL)` before doing anything else, so the URL MUST be `host:port` (no scheme, no path). The placeholder `localhost:5433` is a hint but the editor does not validate the shape, so users can save an invalid URL and only see the error on a health check. The frontend also has a lingering `// BUG: when delete "url" value and save, it will reset to the previous value??` comment at `ConfigEditor.tsx:18` — deleting the URL to force an update may not stick in the editor.
3. **Connection-string SQL-injection surface.** `BuildConnectionString` (`pkg/settings.go:51-61`) `fmt.Sprintf`s the user-supplied `Host`, `Port`, `User`, `Password`, and `Database` into a single-quoted libpq DSN with no escaping. A password (or database name) containing a single quote will break parsing at best and change semantics at worst. Provisioning callers should avoid single quotes in these values.
4. **`YugabyteOptions extends SQLOptions` is misleading.** The frontend type declaration suggests the full `SQLOptions` surface (timeInterval, maxOpenConns, connMaxLifetime, sslmode, sslRootCertFile, tlsConfigurationMethod, …) is available. In practice the plugin's backend reads exactly one jsonData field: `database`. Everything else from the SQLOptions surface is dead configuration that will be silently ignored. The registry schema models only what's wired.
5. **Password field has no `required` marker in the editor** even though the connection string always interpolates a `password=''` clause. That's fine for Yugabyte clusters configured with `trust` auth, but the editor does not distinguish "no password intended" from "user forgot to set the password" — the datasource will simply fail with an auth error at query time.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Validate()` on `dsconfig.json` — pass
- JSON Schema validation against `dsconfig/schema.json` (draft 2020-12, strict) — pass
- `go generate ./...` inside this entry — pass (regenerates the three `.gen.json` artifacts)
- `go build ./... && go vet ./... && gofmt -l . && go test ./...` inside `registry/` — pass across every entry
- Manual review of generated `settings.gen.json` — only `url` / `user` / `jsonData.database` in the spec; `password` is in `secureValues`, never in `spec`.
