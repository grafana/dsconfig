# grafana-mqtt-datasource

Declarative configuration schema for the [MQTT datasource plugin](https://github.com/grafana/mqtt-datasource) (`grafana-mqtt-datasource`).

## Upstream researched

- **Repo**: `github.com/grafana/mqtt-datasource`
- **Ref**: `main`
- **Commit SHA**: `fed63768fae74645af3ab3a8ef3fc509d9dc8cb1`

Every value in [`dsconfig.json`](dsconfig.json) — labels, placeholders, descriptions, section
titles, defaults, `requiredWhen` expressions, storage keys, storage targets, value types, group
titles, and instructions — is traceable to a specific `file:line` in the upstream repo at this
SHA. See [Field provenance](#field-provenance) below.

To reproduce this research:

```bash
git clone https://github.com/grafana/mqtt-datasource
cd mqtt-datasource
git checkout fed63768fae74645af3ab3a8ef3fc509d9dc8cb1
```

If upstream `main` has advanced past this SHA, re-diff the sources listed under
[Sources researched](#sources-researched) before merging any changes to this entry.

## Files

| File | Purpose |
| --- | --- |
| [`dsconfig.json`](dsconfig.json) | dsconfig v1 schema — single source of truth for all config fields, groups, and instructions |
| [`settings.ts`](settings.ts) | TypeScript models: `RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig` |
| [`settings.go`](settings.go) | Go `Config` model (flat: jsonData fields + `DecryptedSecureJSONData`), `PluginID`, `SecureJsonDataKey` typed constants, and the `LoadConfig` utility |
| [`schema.go`](schema.go) | k8s-style SDK plugin schema: embeds `dsconfig.json`, converts it via `dsconfig.NewSDKSchema`, and defines `SettingsExamples` for each connection / authentication variant |
| [`settings_test.go`](settings_test.go) | Table tests for `LoadConfig`, `ApplyDefaults`, and `Validate` |
| [`conformance_test.go`](conformance_test.go) | Runs `schema.RunPluginTests` — the shared dsconfig conformance suite — against `dsconfig.json` |
| `schema.gen.json`, `settings.gen.json`, `settings.examples.gen.json` | Committed schema artifacts (regenerate with `go generate ./...` inside this directory; the `SchemaArtifactInSync` conformance subtest guards drift) |

There is no per-entry `go.mod` — every registry entry is a subpackage of the shared
[`registry/`](..) module (`github.com/grafana/dsconfig/registry`).

## Sources researched

Every source below was read at the pinned upstream SHA (`fed6376`), plus external editor
components at the exact versions the plugin's `package.json` / `yarn.lock` pins.

### Plugin repo (`github.com/grafana/mqtt-datasource@fed6376`)

| File | What was read |
| --- | --- |
| `src/plugin.json:4-5` | `pluginType` (`id` = `grafana-mqtt-datasource`), `pluginName` (`name` = `MQTT`). `info.links` at `:22-30` points at the GitHub repo and license only — no marketplace docs URL. |
| `src/ConfigEditor.tsx:33-129` | The complete config editor: `DataSourceDescription` with hard-coded docs link (`:35-39`), three `<ConfigSection>` blocks, a conditional TLS Configuration section, and the excluded Secure Socks Proxy switch. |
| `src/ConfigEditor.tsx:43-54` | `<ConfigSection title="Connection">` containing the required URI `<Input>`; `jsonData.uri` placeholder `"TCP (tcp://), TLS (tls://), or WebSocket (ws://)"`. |
| `src/ConfigEditor.tsx:56-64` | Client ID `<Field>` — rendered between the Connection and Authentication sections without its own section wrapper. Description `"If not set, a random client ID is used."`. |
| `src/ConfigEditor.tsx:68-105` | `<ConfigSection title="Authentication">`: Username `<Input>`, Password `<SecretInput>`, and the three TLS Switches (Use TLS Client Auth, Skip TLS Verification, With CA Cert). Placeholders and descriptions read verbatim. |
| `src/ConfigEditor.tsx:107-122` | Conditional `<ConfigSection title="TLS Configuration">` — only rendered when `jsonData.tlsAuth \|\| jsonData.tlsAuthWithCACert`. Wraps `TLSSecretsConfig`. |
| `src/ConfigEditor.tsx:124-126` | Grafana-shared `<SecureSocksProxySettings>` (excluded from this entry). |
| `src/ConfigEditor.tsx:21-29` | Reset / update handlers: `updateDatasourcePluginResetOption(props, 'password')` on password reset; `onSwitchChanged` writes to jsonData via `updateDatasourcePluginJsonDataOption`. |
| `src/TLSConfig.tsx:20-114` | `TLSSecretsConfig` component: three `<SecretTextArea>`s (`tlsCACert`, `tlsClientCert`, `tlsClientKey`) with `<Tooltip content=...>` labels and PEM placeholders (`-----BEGIN CERTIFICATE-----`, `-----BEGIN CERTIFICATE-----`, `-----BEGIN RSA PRIVATE KEY-----`). |
| `src/types.ts:10-17` | Frontend `MqttDataSourceOptions` (jsonData): `uri: string`, `username?: string`, `clientID?: string`, `tlsAuth: boolean`, `tlsAuthWithCACert: boolean`, `tlsSkipVerify: boolean`. |
| `src/types.ts:19-24` | Frontend `MqttSecureJsonData`: `password?`, `tlsCACert?`, `tlsClientKey?`, `tlsClientCert?`. |
| `pkg/plugin/datasource.go:60-83` | `getDatasourceSettings`: `json.Unmarshal(s.JSONData, settings)` unconditionally, then copy `s.DecryptedSecureJSONData["password" \| "tlsClientCert" \| "tlsClientKey" \| "tlsCACert"]` onto the same struct. Struct's `json:"password"` / `json:"tls*"` tags are in-memory aliases only. |
| `pkg/mqtt/client.go:26-35` | Backend `Options` struct: `URI`, `Username`, `Password`, `ClientID`, `TLSCACert`, `TLSClientCert`, `TLSClientKey`, `TLSSkipVerify`. Note the ABSENCE of `TLSAuth` / `TLSAuthWithCACert` — these frontend toggles are not read by the backend. |
| `pkg/mqtt/client.go:42-109` | `NewClient`: `AddBroker(URI)` unconditionally, random `grafana_<int>` client id when empty, `SetUsername/SetPassword` only when non-empty, `tls.X509KeyPair(tlsClientCert, tlsClientKey)` when either is non-empty, `AppendCertsFromPEM(tlsCACert)` when non-empty, `tlsConfig.InsecureSkipVerify = tlsSkipVerify`. Ping/keepalive/reconnect/etc. hard-coded, no defaults exposed. |
| `pkg/mqtt/proxy.go:24-58` | Secure Socks Proxy wiring (excluded field): `settings.ProxyClient(ctx)` and `NewSecureSocksProxyContextDialer` invoked transparently when `secureSocksProxyEnabled`. |
| `pkg/mqtt/proxy.go:66-80` | Default-port mapping used only for the proxy path but reflects paho scheme conventions: `tcp/mqtt=1883`, `ssl/tls/tcps/mqtts=8883`, `ws=80`, `wss=443`. |
| `package.json` / `yarn.lock` | External component versions (see next table). |

### External editor components

Read at the versions the plugin's `package.json` pins (and `yarn.lock` resolves).

| Component | Version | Source consulted | What was read |
| --- | --- | --- | --- |
| `DataSourceDescription`, `ConfigSection` | `@grafana/plugin-ui@0.13.1` | `github.com/grafana/plugin-ui` tag `v0.13.1` | `dataSourceName`, `docsLink`, `hasRequiredFields` header block; `title` / `isCollapsible` on ConfigSection (no storage fields written) |
| `SecureSocksProxySettings` (excluded) | `@grafana/ui@13.1.0-25893932881` | grafana/grafana build `13.1.0-25893932881` `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | Storage key written: `jsonData.enableSecureSocksProxy` — confirmed and excluded per AGENTS.md |
| `Field`, `Input`, `SecretInput`, `SecretTextArea`, `Switch`, `Icon`, `Label`, `Stack`, `Tooltip` | `@grafana/ui@13.1.0-25893932881` | grafana/grafana build `13.1.0-25893932881` `packages/grafana-ui/src/components/` | Prop names (`label`, `description`, `placeholder`, `required`, `value`, `onChange`, `onReset`, `onBlur`, `isConfigured`, `cols`, `rows`) — no storage fields written by the components themselves |
| `DataSourcePluginOptionsEditorProps`, `DataSourceJsonData`, `onUpdateDatasourceJsonDataOption`, `onUpdateDatasourceSecureJsonDataOption`, `updateDatasourcePluginJsonDataOption`, `updateDatasourcePluginResetOption` | `@grafana/data@13.1.0-25893932881` | grafana/grafana build `13.1.0-25893932881` `packages/grafana-data/src/types/datasource.ts` and `packages/grafana-data/src/utils/datasource.ts` | Base interface `MqttDataSourceOptions` extends; storage semantics of `onUpdateDatasource*Option` (writes to `jsonData` / `secureJsonData` respectively) and `updateDatasourcePluginResetOption` (clears the value and sets `secureJsonFields[key]=false`) |
| `config.secureSocksDSProxyEnabled` | `@grafana/runtime@13.1.0-25893932881` | grafana/grafana build `13.1.0-25893932881` `packages/grafana-runtime/src/config.ts` | Feature gate read at `src/ConfigEditor.tsx:124` — no storage field |

## Field provenance

Every dsconfig field, traced from its schema `id` to the upstream `file:line` where each of its
label, placeholder, description (tooltip), default, storage key, and value type is defined.

| Schema `id` | Storage key | Target | Label source | Placeholder / options / default source | Value type source | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| `jsonData_uri` | `uri` | `jsonData` | `<Field label="URI" required>` at `ConfigEditor.tsx:44` | Placeholder `"TCP (tcp://), TLS (tls://), or WebSocket (ws://)"` at `ConfigEditor.tsx:51` | `Options.URI string \`json:"uri"\`` at `pkg/mqtt/client.go:27`; TS `uri: string` at `src/types.ts:11` | Role `endpoint.baseUrl`; `requiredWhen: "true"` because `paho.AddBroker` is called unconditionally at `pkg/mqtt/client.go:46` with no fallback |
| `jsonData_clientID` | `clientID` | `jsonData` | `<Field label="Client ID" description="If not set, a random client ID is used.">` at `ConfigEditor.tsx:56` | No placeholder | `Options.ClientID string \`json:"clientID"\`` at `pkg/mqtt/client.go:30`; TS `clientID?: string` at `src/types.ts:13` | Description text is the only tooltip; backend generates `grafana_<int>` at `pkg/mqtt/client.go:48-52` when empty |
| `jsonData_username` | `username` | `jsonData` | `<Field label="Username">` at `ConfigEditor.tsx:69` | Placeholder `"Username"` at `ConfigEditor.tsx:73` | `Options.Username string \`json:"username"\`` at `pkg/mqtt/client.go:28`; TS `username?: string` at `src/types.ts:12` | Role `auth.basic.username`; only applied via `paho.SetUsername` when non-empty (`pkg/mqtt/client.go:54-56`) |
| `secureJsonData_password` | `password` | `secureJsonData` | `<Field label="Password">` at `ConfigEditor.tsx:78` | Placeholder `"Password"` at `ConfigEditor.tsx:81` | `Options.Password string \`json:"password"\`` at `pkg/mqtt/client.go:29`; TS `password?: string` at `src/types.ts:20` | Role `auth.basic.password`; only applied via `paho.SetPassword` when non-empty (`pkg/mqtt/client.go:58-60`) |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | `<Field label="Use TLS Client Auth" description="Enables TLS authentication using client cert configured in secure json data.">` at `ConfigEditor.tsx:88-91` | Default `false` (`ConfigEditor.tsx:92`: `value={jsonData.tlsAuth \|\| false}`) | Frontend-only: `MqttDataSourceOptions.tlsAuth: boolean` at `src/types.ts:14`. Backend `Options` struct does NOT include this field | No role — this is an editor-only visibility toggle; backend never reads it |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | `<Field label="Skip TLS Verification" description="When enabled, skips verification of the MQTT server's TLS certificate chain and host name.">` at `ConfigEditor.tsx:95-98` | Default `false` (`ConfigEditor.tsx:99`) | `Options.TLSSkipVerify bool \`json:"tlsSkipVerify"\`` at `pkg/mqtt/client.go:34`; TS `tlsSkipVerify: boolean` at `src/types.ts:16` | Role `transport.tlsSkipVerify`; feeds `tls.Config.InsecureSkipVerify` at `pkg/mqtt/client.go:62-64` |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | `<Field label="With CA Cert" description="Needed for verifying servers with self-signed TLS Certs.">` at `ConfigEditor.tsx:102-103` | Default `false` (`ConfigEditor.tsx:103`) | Frontend-only: `MqttDataSourceOptions.tlsAuthWithCACert: boolean` at `src/types.ts:15`. Backend `Options` struct does NOT include this field | No role — this is an editor-only visibility toggle; backend never reads it |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | Tooltip via `<Tooltip content="If a Certificate Authority certificate is required to verify the server's certificate, provide it here.">` at `TLSConfig.tsx:31-40` (rendered as info-icon label `"TLS CA Certificate"`) | Placeholder `"-----BEGIN CERTIFICATE-----"` at `TLSConfig.tsx:46` | `Options.TLSCACert string \`json:"tlsCACert"\`` at `pkg/mqtt/client.go:31`; TS `tlsCACert?: string` at `src/types.ts:21` | Role `tls.caCert`; `dependsOn: jsonData_tlsAuthWithCACert == true` mirrors editor visibility; loaded via `AppendCertsFromPEM` at `pkg/mqtt/client.go:75-79` when non-empty |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | Tooltip via `<Tooltip content="To authenticate with an TLS client certificate, provide the client certificate here.">` at `TLSConfig.tsx:63-71` (label `"TLS Client Certificate"`) | Placeholder `"-----BEGIN CERTIFICATE-----"` at `TLSConfig.tsx:75` | `Options.TLSClientCert string \`json:"tlsClientCert"\`` at `pkg/mqtt/client.go:32`; TS `tlsClientCert?: string` at `src/types.ts:23` | Role `tls.clientCert`; `dependsOn: jsonData_tlsAuth == true`; loaded via `tls.X509KeyPair` at `pkg/mqtt/client.go:66-73` |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | Tooltip via `<Tooltip content="To authenticate with a client TLS certificate, provide the private key here.">` at `TLSConfig.tsx:92-95` (label `"TLS Client Key"`) | Placeholder `"-----BEGIN RSA PRIVATE KEY-----"` at `TLSConfig.tsx:102` | `Options.TLSClientKey string \`json:"tlsClientKey"\`` at `pkg/mqtt/client.go:33`; TS `tlsClientKey?: string` at `src/types.ts:22` | Role `tls.clientKey`; `dependsOn: jsonData_tlsAuth == true`; paired with `tlsClientCert` in the `X509KeyPair` call |

## Field inventory summary

| Schema field | Storage key | Target | Editor label | Read by backend? |
| --- | --- | --- | --- | --- |
| `jsonData_uri` | `uri` | `jsonData` | URI | Yes (required, no fallback) |
| `jsonData_clientID` | `clientID` | `jsonData` | Client ID | Yes (random `grafana_<int>` when empty) |
| `jsonData_username` | `username` | `jsonData` | Username | Yes (applied when non-empty) |
| `secureJsonData_password` | `password` | `secureJsonData` | Password | Yes (applied when non-empty) |
| `jsonData_tlsAuth` | `tlsAuth` | `jsonData` | Use TLS Client Auth | **No — editor-only visibility toggle** |
| `jsonData_tlsSkipVerify` | `tlsSkipVerify` | `jsonData` | Skip TLS Verification | Yes (feeds `tls.Config.InsecureSkipVerify`) |
| `jsonData_tlsAuthWithCACert` | `tlsAuthWithCACert` | `jsonData` | With CA Cert | **No — editor-only visibility toggle** |
| `secureJsonData_tlsCACert` | `tlsCACert` | `secureJsonData` | TLS CA Certificate | Yes (applied when non-empty) |
| `secureJsonData_tlsClientCert` | `tlsClientCert` | `secureJsonData` | TLS Client Certificate | Yes (applied when non-empty; paired with `tlsClientKey`) |
| `secureJsonData_tlsClientKey` | `tlsClientKey` | `secureJsonData` | TLS Client Key | Yes (applied when non-empty; paired with `tlsClientCert`) |
| `jsonData_enableSecureSocksProxy` (excluded) | `enableSecureSocksProxy` | `jsonData` | Enable Secure Socks Proxy | Indirectly (via `settings.ProxyClient(ctx)` at `pkg/mqtt/proxy.go:24-32`) — excluded per AGENTS.md |

### Frontend-only settings

- `jsonData.tlsAuth` — written by the editor to gate the visibility of the TLS Client
  Certificate / TLS Client Key inputs in `TLSSecretsConfig`. The backend's `mqtt.Options` struct
  does not include this field.
- `jsonData.tlsAuthWithCACert` — same pattern for the TLS CA Certificate input.

Both are surfaced in the schema (they are legitimate editor-written jsonData keys), but neither
carries a role and neither influences the backend's TLS behavior directly. The backend's TLS
logic is triggered purely by the presence of `tlsClientCert`/`tlsClientKey`/`tlsCACert` secrets.

### Backend-only settings

None. Every backend-consumed setting has an editor UI, except the excluded Secure Socks Proxy
switch which is covered by Grafana's shared field pack.

## Where the types are defined

The configuration types are spread across the plugin and its dependencies — some fields and base
types come from libraries/SDKs rather than the plugin itself:

### Frontend (TypeScript)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `MqttDataSourceOptions` (jsonData: uri, username?, clientID?, tlsAuth, tlsAuthWithCACert, tlsSkipVerify), `MqttSecureJsonData` (password?, tlsCACert?, tlsClientKey?, tlsClientCert?) | `src/types.ts:10-24` | plugin ([grafana/mqtt-datasource](https://github.com/grafana/mqtt-datasource)) |
| `DataSourceJsonData` (base interface `MqttDataSourceOptions` extends: `authType?`, `defaultRegion?`, …) | `packages/grafana-data/src/types/datasource.ts` | `@grafana/data` `13.1.0-25893932881` (grafana/grafana build `13.1.0-25893932881`) |
| `DataSourcePluginOptionsEditorProps`, `onUpdateDatasourceJsonDataOption`, `onUpdateDatasourceSecureJsonDataOption`, `updateDatasourcePluginJsonDataOption`, `updateDatasourcePluginResetOption` | `packages/grafana-data/src/` | `@grafana/data` `13.1.0-25893932881` |
| `DataSourceDescription`, `ConfigSection` (no storage fields written) | `src/components/` | `@grafana/plugin-ui` `0.13.1` |
| `Field`, `Input`, `SecretInput`, `SecretTextArea`, `Switch`, `Icon`, `Label`, `Stack`, `Tooltip` (no storage fields written) | `packages/grafana-ui/src/components/` | `@grafana/ui` `13.1.0-25893932881` |
| Secure Socks Proxy — `SecureSocksProxySettings` writes `jsonData.enableSecureSocksProxy` (excluded from this entry) | `packages/grafana-ui/src/components/DataSourceSettings/SecureSocksProxySettings.tsx` | `@grafana/ui` `13.1.0-25893932881` |

### Backend (Go)

| Type / field | Defined in | Package |
| --- | --- | --- |
| `Options` (URI, Username, Password, ClientID, TLSCACert, TLSClientCert, TLSClientKey, TLSSkipVerify) | `pkg/mqtt/client.go:26-35` | plugin ([grafana/mqtt-datasource](https://github.com/grafana/mqtt-datasource)) |
| `NewClient` (paho.NewClientOptions wiring, `AddBroker`, random client id, `SetUsername/SetPassword`, `X509KeyPair`, `AppendCertsFromPEM`, `InsecureSkipVerify`) | `pkg/mqtt/client.go:42-109` | plugin |
| `getDatasourceSettings` (`json.Unmarshal(s.JSONData, settings)`, secret copy loop) | `pkg/plugin/datasource.go:60-83` | plugin |
| `configureProxyIfEnabled` (consumes `enableSecureSocksProxy` transparently via `settings.ProxyClient`) | `pkg/mqtt/proxy.go:19-58` | plugin (but the field is Grafana-defined) |
| `backend.DataSourceInstanceSettings` (carries `JSONData`, `DecryptedSecureJSONData`, `URL`, `BasicAuthEnabled` — all root fields unused by this plugin) | `backend/common.go` | `github.com/grafana/grafana-plugin-sdk-go` |
| `settings.ProxyClient(ctx)` / `SecureSocksProxyEnabled` / `NewSecureSocksProxyContextDialer` | `backend/proxy/` | `github.com/grafana/grafana-plugin-sdk-go` |
| `paho.mqtt.golang` `Client`, `ClientOptions` (upstream MQTT client library) | | `github.com/eclipse/paho.mqtt.golang` |

The models in this entry flatten that spread into a single Go `Config` type (jsonData fields +
`DecryptedSecureJSONData`) plus a `SecureJsonDataKey` typed constant list. `settings.ts` keeps
the three canonical TypeScript types (`RootConfig`, `JsonDataConfig`, `SecureJsonDataConfig`);
`RootConfig` is a blank object because the MQTT plugin stores nothing at the root level.

## Modeling decisions

- **`RootConfig` is a blank object**. `MqttDataSourceOptions` extends `DataSourceJsonData` and
  every plugin field lives in jsonData; the backend never touches `settings.URL`,
  `settings.User`, etc.
- **`requiredWhen` encodes the backend contract, not the editor markers**. Only `jsonData.uri`
  is `requiredWhen: "true"` because `paho.AddBroker(o.URI)` is called unconditionally at
  `pkg/mqtt/client.go:46` with no default. Basic-auth username/password, TLS material, and the
  toggles are all optional at the backend contract level.
- **Description = tooltip / Field description**. `<Field description={...}>` (Client ID, TLS
  toggles) is the only place descriptions appear inline in the editor. For the three TLS
  material fields, the "description" comes from `<Tooltip content=...>` in `TLSConfig.tsx` —
  these are the only tooltips shown by that component, so they are captured verbatim in the
  schema's `description` field.
- **Frontend-only toggles surfaced in schema**. `jsonData.tlsAuth` and
  `jsonData.tlsAuthWithCACert` are surfaced in the schema even though the backend does not read
  them, because they are legitimate editor-written jsonData keys. They are not given roles.
  Note this deliberately deviates from the backend's `mqtt.Options` struct
  (`pkg/mqtt/client.go:26-35`), which omits them.
- **TLS keypair paired via `pair` relationship**. `tlsClientCert` + `tlsClientKey` are
  declared as a `pair` because the backend's `tls.X509KeyPair` call fails when only one side
  is provided. `Validate` enforces this at the runtime-contract level.
- **`dependsOn` on secure TLS fields references frontend-only toggles**. The editor gates
  the visibility of the three TLS material inputs on `tlsAuth` (for cert/key) and
  `tlsAuthWithCACert` (for CA). The `dependsOn` expressions mirror the editor
  (`src/ConfigEditor.tsx:107-122`); the `requiredWhen` is absent because the backend does not
  require any of these secrets.
- **Role `tls.serverName` NOT used**. The MQTT plugin does not expose a TLS ServerName input;
  `crypto/tls` derives it from the URI at connect time.
- **`docURL` is not declared in `src/plugin.json`.** `info.links` at `src/plugin.json:22-30`
  points at GitHub only. The editor hard-codes its own docs link at
  `ConfigEditor.tsx:37` (`https://grafana.com/grafana/plugins/grafana-mqtt-datasource/?tab=overview`);
  this entry uses the canonical marketplace URL (no `?tab` query) for `docURL`.
- **Secure Socks Proxy excluded**. `jsonData.enableSecureSocksProxy` is deliberately omitted
  per AGENTS.md; the SDK-provided `SecureSocksProxySettings` component still renders in the
  editor via `src/ConfigEditor.tsx:124-126` when the Grafana feature is enabled.
- **Field ID naming convention**. IDs are prefixed with their storage target (`jsonData_` /
  `secureJsonData_`) followed by the camelCase storage key, e.g. `jsonData_uri`,
  `secureJsonData_tlsClientCert`. The `key` property keeps the plugin's raw storage key.
- **`SecureJsonDataConfig` is a key list**. Secure values are write-only, so the secure type is
  the array of secret key names (`password`, `tlsCACert`, `tlsClientCert`, `tlsClientKey`);
  consumers read `secureJsonFields` to see what is configured.

## SDK plugin schema and k8s-style examples (`schema.go`)

`NewSchema()` assembles the `grafana-plugin-sdk-go` `pluginschema.PluginSchema` bundle (the
k8s-style schema Grafana's datasource API server serves as `{apiVersion}.json`) from the
embedded `dsconfig.json`: root fields plus a nested `jsonData` object become the OpenAPI
settings `spec`, secure fields become `secureValues`.

`SettingsExamples()` provides the default configuration plus one k8s-style example per
authentication / connection variant. Each example is a full instance-settings object with the
plugin configuration nested under `jsonData` and the write-only secrets under `secureJsonData`
(placeholder values to be replaced with real secrets):

| Example | Connection | `secureJsonData` |
| --- | --- | --- |
| `""` (default) | Empty (URI unset) | `password` (empty) |
| `anonymousTCP` | `tcp://broker.example.com:1883` | `password` (empty) |
| `basicAuthTCP` | `tcp://broker.example.com:1883` + `username` | `password` |
| `tlsClientAuth` | `tls://broker.example.com:8883` + `tlsAuth=true` | `tlsClientCert`, `tlsClientKey` |
| `selfSignedCA` | `tls://broker.internal.corp:8883` + `tlsAuthWithCACert=true` | `tlsCACert` |
| `mutualTLSWithCA` | Everything (URI, client id, username, all three TLS toggles) | `password`, `tlsCACert`, `tlsClientCert`, `tlsClientKey` |
| `tlsSkipVerify` | `tls://broker.example.com:8883` + `tlsSkipVerify=true` | `password` (empty) |
| `webSocket` | `wss://broker.example.com:443/mqtt` | `password` (empty) |

## `LoadConfig` utility (`settings.go`)

`LoadConfig(ctx context.Context, settings backend.DataSourceInstanceSettings) (Config, error)`
runs the full three-phase load flow on a datasource instance's settings and returns a
fully-defaulted, validated `Config`:

1. **Parse** — unmarshal jsonData into `Config` when JSONData is non-empty (mirroring
   `pkg/plugin/datasource.go:63-65`; empty JSONData is not a parse error), then copy decrypted
   secrets by known key into `DecryptedSecureJSONData`.
2. **`ApplyDefaults`** — currently a no-op. Every schema default lands at the Go zero value
   (the three TLS booleans default to `false`, which is already the zero value). The method
   is kept as the sole entry point for editor-parity defaults so future non-zero defaults
   apply consistently.
3. **`Validate`** — enforce the runtime contract: URI must be non-empty; if either
   `tlsClientCert` or `tlsClientKey` is present, both must be. Errors are joined so every
   problem surfaces at once.

Everything is logged via `backend.Logger.FromContext(ctx)` with `datasource_uid`,
`datasource_name`, and `plugin` labels so log lines carry request context.

### Direct access to individual phases

`(*Config).ApplyDefaults()` and `(Config).Validate() error` are exported separately for callers
that want to compose them themselves (e.g. provisioning preview, schema-example round-trip,
tests that need to distinguish parse-level from policy-level errors).

## Upstream findings

Potential bugs, misleading UX, and consistency issues discovered while researching upstream. All
preserved verbatim in the schema — the schema records what the plugin **does**, not what it
**should** do.

1. **`tlsAuth` and `tlsAuthWithCACert` are editor-only view state, not backend-consumed
   fields.** The backend's `mqtt.Options` struct (`pkg/mqtt/client.go:26-35`) omits both toggles
   entirely. The TLS client keypair is loaded whenever either `tlsClientCert` or `tlsClientKey`
   is non-empty (`pkg/mqtt/client.go:66-73`), and the CA cert pool is loaded whenever
   `tlsCACert` is non-empty (`pkg/mqtt/client.go:75-79`) — the toggles have zero runtime
   effect. This means a datasource with `tlsClientCert`/`tlsClientKey` set but `tlsAuth: false`
   will still use TLS client authentication.
2. **URI is required at runtime but the editor has no `invalid` state.** `<Field label="URI"
   required>` (`ConfigEditor.tsx:44`) marks the field required visually, but there is no
   `invalid={!jsonData.uri}` prop and no reactive validation. A datasource saved with empty
   URI still passes the editor; the failure surfaces only at first Connect
   (`paho.AddBroker("")` succeeds silently but the subsequent Connect returns an error).
3. **The Password reset button clears the stored value.** `updateDatasourcePluginResetOption(props,
   'password')` at `ConfigEditor.tsx:22-23` explicitly resets the secret, which is standard.
   All three TLS `SecretTextArea` reset buttons behave the same way
   (`TLSConfig.tsx:51-53,80-82,107-109`).
4. **Description on "Use TLS Client Auth" is misleading.** The description reads "Enables TLS
   authentication using client cert configured in secure json data." — but the backend does
   NOT gate TLS client authentication on this toggle. Setting it to `true` without providing
   `tlsClientCert`/`tlsClientKey` does nothing; setting it to `false` while providing those
   secrets still uses TLS client authentication. Preserved verbatim.
5. **Description contains an English error: "an TLS" instead of "a TLS"**. `TLSConfig.tsx:65`
   reads "To authenticate with **an TLS** client certificate, provide the client certificate
   here." Preserved verbatim per AGENTS.md.
6. **Frontend/backend type divergence for TLS toggles.** `MqttDataSourceOptions` declares
   `tlsAuth: boolean` and `tlsAuthWithCACert: boolean` as required (`src/types.ts:14-15`), but
   the backend `Options` struct doesn't include them. This entry mirrors the frontend shape
   because it is the shape actually stored in jsonData.
7. **`docURL` is not declared in `src/plugin.json`.** `info.links` at `src/plugin.json:22-30`
   only contains the GitHub repo and license links; the editor hardcodes
   `https://grafana.com/grafana/plugins/grafana-mqtt-datasource/?tab=overview` at
   `ConfigEditor.tsx:37`. This entry uses the query-less canonical URL for `docURL`.
8. **Client ID is placed between two ConfigSections without its own section.** The editor
   renders it after `</ConfigSection>` for Connection and before `<ConfigSection
   title="Authentication">` (`ConfigEditor.tsx:56-64`), which visually looks like a Connection
   field but has no section wrapper. This entry groups it into the Connection group for
   semantic consistency.
9. **Paho client hard-codes ping/keepalive/reconnect/maxReconnectInterval.**
   `pkg/mqtt/client.go:82-86` sets `PingTimeout=60s`, `KeepAlive=60s`, `AutoReconnect=true`,
   `CleanSession=false`, `MaxReconnectInterval=10s` with no way to override from config. Not
   a schema issue — just an operational note.

## Validation performed

- `dsconfig.ParseAndResolveSchemaJSON` + `Schema.Validate()` (Go validator in this repo) — passes.
- `go generate ./...` inside this directory (regenerates the three `.gen.json` artifacts) — passes.
- `go build ./...` and `go vet ./...` from `registry/` — clean.
- `gofmt -l .` from `registry/` — clean.
- `go test ./...` from `registry/` — all packages pass (including this entry's `LoadConfig`,
  `ApplyDefaults`, `Validate` tables and the shared `TestSchemaConformance` guard rails:
  `BaseFieldsResolved`, `SchemaRoundTrip`, `SchemaArtifactInSync`, `SchemaSpecHasNoSecureJSON`,
  `ConfigSchemaValid`, `JSONDataMatchesStruct`, `JSONDataTypesMatchStruct`,
  `SecureValuesMatchLoadSettings`).
- `tsc --noEmit --strict --target es2020 --module esnext --moduleResolution node`
  (typescript@5.6) on `settings.ts` — clean.
