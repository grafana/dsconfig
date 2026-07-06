/**
 * Configuration models for the MQTT datasource plugin
 * (plugin id: `grafana-mqtt-datasource`).
 *
 * Sources of truth (https://github.com/grafana/mqtt-datasource @ fed6376):
 * - `src/plugin.json:5` — plugin id (`"grafana-mqtt-datasource"`), name
 *   (`"MQTT"` at `:4`). `info.links` at `:22-30` points at the GitHub repo
 *   and license only, so the editor's hard-coded docsLink at
 *   `src/ConfigEditor.tsx:37` is the canonical docs URL.
 * - `src/ConfigEditor.tsx:33-129` — the configuration editor:
 *   - `DataSourceDescription` with `dataSourceName="MQTT"`,
 *     `docsLink="https://grafana.com/grafana/plugins/grafana-mqtt-datasource/?tab=overview"`,
 *     `hasRequiredFields` (`:35-39`).
 *   - `<ConfigSection title="Connection">` (`:43`) with a required URI
 *     `<Input>` writing `jsonData.uri` (`:44-53`, placeholder
 *     `"TCP (tcp://), TLS (tls://), or WebSocket (ws://)"`).
 *   - Client ID `<Field>` (`:56-64`) — sits between the Connection and
 *     Authentication sections with no section wrapper. Description
 *     `"If not set, a random client ID is used."`.
 *   - `<ConfigSection title="Authentication">` (`:68`) with:
 *     - Username `<Input>` writing `jsonData.username` (`:69-76`,
 *       placeholder `"Username"`).
 *     - Password `<SecretInput>` writing `secureJsonData.password`
 *       (`:78-86`, placeholder `"Password"`, reset via
 *       `updateDatasourcePluginResetOption` at `:22`).
 *     - Use TLS Client Auth `<Switch>` writing `jsonData.tlsAuth`
 *       (`:88-93`, description
 *       `"Enables TLS authentication using client cert configured in
 *       secure json data."`).
 *     - Skip TLS Verification `<Switch>` writing `jsonData.tlsSkipVerify`
 *       (`:95-100`, description
 *       `"When enabled, skips verification of the MQTT server's TLS
 *       certificate chain and host name."`).
 *     - With CA Cert `<Switch>` writing `jsonData.tlsAuthWithCACert`
 *       (`:102-104`, description
 *       `"Needed for verifying servers with self-signed TLS Certs."`).
 *   - Conditional `<ConfigSection title="TLS Configuration">`
 *     (`:107-122`) — only rendered when `jsonData.tlsAuth ||
 *     jsonData.tlsAuthWithCACert`. Hosts `TLSSecretsConfig`.
 *   - `<SecureSocksProxySettings>` (`:124-126`) — the Grafana-shared
 *     Secure Socks Proxy switch; deliberately excluded from this entry
 *     per AGENTS.md.
 * - `src/TLSConfig.tsx:20-114` — the `TLSSecretsConfig` component that
 *   renders the three secret text areas (`tlsCACert`, `tlsClientCert`,
 *   `tlsClientKey`) with tooltips (via `<Tooltip content=...>` next to
 *   an info icon) and PEM placeholders.
 * - `src/types.ts:10-24` — frontend `MqttDataSourceOptions` (jsonData
 *   shape) and `MqttSecureJsonData` (secret keys).
 * - `pkg/plugin/datasource.go:60-83` — `getDatasourceSettings`:
 *   unmarshals `s.JSONData` into `mqtt.Options` and then copies
 *   `s.DecryptedSecureJSONData["password" | "tlsClientCert" |
 *   "tlsClientKey" | "tlsCACert"]` onto the same struct fields (the
 *   struct's `json:"password"` / `json:"tlsCACert"` / etc. tags are only
 *   in-memory aliases — those keys are never stored in jsonData).
 * - `pkg/mqtt/client.go:26-109` — backend `Options` struct and
 *   `NewClient`: `AddBroker(URI)` unconditionally, generate random client
 *   id `"grafana_<int>"` when empty, `SetUsername/SetPassword` only when
 *   non-empty, `tls.X509KeyPair(tlsClientCert, tlsClientKey)` if either
 *   is non-empty, `AppendCertsFromPEM(tlsCACert)` when non-empty,
 *   `tlsConfig.InsecureSkipVerify = tlsSkipVerify`. Note the backend
 *   `Options` struct does NOT carry `tlsAuth` or `tlsAuthWithCACert` —
 *   those are frontend-only editor toggles.
 * - `pkg/mqtt/proxy.go:66-80` — default-port mapping for URI schemes
 *   (`tcp/mqtt=1883`, `ssl/tls/tcps/mqtts=8883`, `ws=80`, `wss=443`).
 *
 * External components consulted at their pinned versions:
 * - `@grafana/plugin-ui@0.13.1` — `DataSourceDescription` (renders the
 *   header block with the docs link), `ConfigSection` (renders section
 *   titles).
 * - `@grafana/ui@13.1.0-25893932881` — `Field`, `Input`, `SecretInput`,
 *   `SecretTextArea`, `Switch`, `SecureSocksProxySettings`, `Icon`,
 *   `Label`, `Stack`, `Tooltip`.
 * - `@grafana/data@13.1.0-25893932881` — `DataSourceJsonData` (base
 *   interface `MqttDataSourceOptions` extends),
 *   `DataSourcePluginOptionsEditorProps`,
 *   `onUpdateDatasourceJsonDataOption`,
 *   `onUpdateDatasourceSecureJsonDataOption`,
 *   `updateDatasourcePluginJsonDataOption`,
 *   `updateDatasourcePluginResetOption`.
 * - `@grafana/runtime@13.1.0-25893932881` — `config` object read at
 *   `ConfigEditor.tsx:124` to conditionally render Secure Socks Proxy.
 *
 * The Secure Socks Proxy switch is deliberately excluded from this
 * registry entry (AGENTS.md exclusion for
 * `jsonData.enableSecureSocksProxy`).
 */

/**
 * Root (top-level datasource settings) fields.
 *
 * The MQTT plugin stores nothing at the root level. `MqttDataSourceOptions`
 * (`src/types.ts:10-17`) extends `DataSourceJsonData` and every field
 * (`uri`, `username`, `clientID`, `tlsAuth`, `tlsAuthWithCACert`,
 * `tlsSkipVerify`) lives in jsonData. The backend `getDatasourceSettings`
 * (`pkg/plugin/datasource.go:60-83`) only reads `s.JSONData` and
 * `s.DecryptedSecureJSONData`. `RootConfig` is therefore a blank object.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Matches the plugin's frontend
 * `MqttDataSourceOptions` (`src/types.ts:10-17`).
 */
export type JsonDataConfig = {
  /**
   * MQTT broker URI. Passed directly to `paho.AddBroker` at
   * `pkg/mqtt/client.go:46`. Supported schemes are `tcp://`, `tls://`
   * (or `ssl://`, `tcps://`, `mqtts://` — all treated as MQTT-over-TLS by
   * `pkg/mqtt/proxy.go:66-80`), `ws://`, and `wss://`. The editor marks
   * this field required (`<Field label="URI" required>` at
   * `src/ConfigEditor.tsx:44`); the backend has no default and no
   * fallback.
   */
  uri: string;
  /**
   * Optional MQTT client identifier. When empty the backend generates a
   * random client id of the form `grafana_<int>` at
   * `pkg/mqtt/client.go:48-52`.
   */
  clientID?: string;
  /**
   * Optional MQTT basic-auth username. Passed to `paho.SetUsername` at
   * `pkg/mqtt/client.go:54-56` only when non-empty.
   */
  username?: string;
  /**
   * Editor-only toggle. `src/ConfigEditor.tsx:92,107-122` uses this
   * boolean to gate the visibility of the TLS Client Certificate and TLS
   * Client Key inputs inside the "TLS Configuration" section. The
   * backend `mqtt.Options` struct (`pkg/mqtt/client.go:26-35`) does NOT
   * include this field — the TLS keypair is loaded whenever either
   * `tlsClientCert` or `tlsClientKey` is non-empty
   * (`pkg/mqtt/client.go:66-73`).
   */
  tlsAuth: boolean;
  /**
   * Editor-only toggle. `src/ConfigEditor.tsx:103,107-122` uses this
   * boolean to gate the visibility of the TLS CA Certificate input. The
   * backend does not read this field — `pkg/mqtt/client.go:75-79`
   * builds a CA cert pool whenever `tlsCACert` is non-empty regardless of
   * the toggle.
   */
  tlsAuthWithCACert: boolean;
  /**
   * Skip verification of the MQTT server's TLS certificate chain and host
   * name. Flows directly into `tls.Config.InsecureSkipVerify` at
   * `pkg/mqtt/client.go:62-64`. Applies only to TLS-carrying URI schemes
   * (`tls://`, `wss://`, …).
   */
  tlsSkipVerify: boolean;
  /**
   * Written by `@grafana/ui`'s `SecureSocksProxySettings` at
   * `src/ConfigEditor.tsx:124-126`. Read transparently by the SDK's
   * `s.ProxyClient(ctx)` call in `pkg/mqtt/proxy.go:24-32`; the MQTT
   * plugin's own Go code never inspects it by name. Deliberately
   * excluded from the dsconfig registry entry per AGENTS.md.
   */
  enableSecureSocksProxy?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`):
 * - `password` — MQTT basic-auth password. Applied via `paho.SetPassword`
 *   at `pkg/mqtt/client.go:58-60` only when non-empty.
 * - `tlsCACert` — PEM-encoded CA certificate used to verify the MQTT
 *   server's certificate. Loaded via `x509.NewCertPool().AppendCertsFromPEM`
 *   at `pkg/mqtt/client.go:75-79` when non-empty.
 * - `tlsClientCert` — PEM-encoded client certificate used for mutual TLS.
 *   Loaded via `tls.X509KeyPair(tlsClientCert, tlsClientKey)` at
 *   `pkg/mqtt/client.go:66-73` when either the cert or the key is
 *   non-empty; both must be provided together for the call to succeed.
 * - `tlsClientKey` — PEM-encoded private key that pairs with
 *   `tlsClientCert`. Loaded via the same `tls.X509KeyPair` call.
 */
export type SecureJsonDataConfig = Array<'password' | 'tlsCACert' | 'tlsClientCert' | 'tlsClientKey'>;
