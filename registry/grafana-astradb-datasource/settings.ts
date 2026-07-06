/**
 * Configuration models for the AstraDB datasource plugin (`grafana-astradb-datasource`).
 *
 * Sources of truth (https://github.com/grafana/astradb-datasource @ 5c4a0400):
 * - `src/plugin.json` — plugin id, name, docs URL.
 * - `src/types.ts` — `AstraSettings` (jsonData shape), `SecureSettings` (secret keys).
 * - `src/components/ConfigEditor.tsx` — the configuration editor; defines the
 *   `Connection` numeric enum (`TOKEN = 0`, `CREDENTIALS = 1`) that jsonData.authKind stores.
 * - `pkg/models/settings.go` — backend `Settings` and `LoadSettings` (secure values are
 *   copied into the Settings struct via mapstructure, keyed by field name — so
 *   secureJsonData keys are literally `token` and `password`).
 * - `pkg/plugin/handlers_checkhealth.go` — the mandatory-field checks (URI/Token when
 *   authKind==0; grpcEndpoint/authEndpoint/user/password when authKind==1).
 * - `pkg/plugin/handlers_querydata.go:105-130` — how each jsonData field is actually
 *   consumed when building the gRPC client and the token/table-based auth providers.
 */

/**
 * AstraDB authentication mode. Numeric enum (JSON number, not string) matching
 * `Connection.TOKEN = 0` / `Connection.CREDENTIALS = 1` in
 * `src/components/ConfigEditor.tsx:11-14` and `AuthType uint8` +
 * `AuthTypeToken = iota` / `AuthTypeCredentials` in `pkg/models/settings.go:11-16`.
 */
export type AstraAuthKind = 0 | 1;

/**
 * Root (top-level datasource settings) fields.
 *
 * The AstraDB datasource stores no plugin-specific fields at the root level
 * (`url`, `basicAuth`, `user`, etc. are unused — the plugin puts its user/URI
 * inside jsonData), so this is a blank object rather than null.
 */
export type RootConfig = Record<string, never>;

/**
 * Fields stored in `jsonData`. Matches `AstraSettings` in `src/types.ts:5-14`
 * minus fields the frontend never writes and the backend never reads:
 *
 * - The frontend `AstraSettings` interface declares `password: string` and
 *   `database?: string`, but the config editor never writes either into
 *   jsonData: the password lives in `secureJsonData`, and `database` is not
 *   present anywhere in the editor or the backend Settings — it's dead weight
 *   in the frontend type only.
 * - The backend `Settings` struct in `pkg/models/settings.go:18-27` tags
 *   `Token string` as `json:"token"` and `Password string` as `json:"password"`
 *   even though the plugin only stores those secrets in `secureJsonData`. The
 *   `LoadSettings` flow later copies the decrypted secret onto the same struct
 *   field. Those aliases are backend implementation artifacts, not real
 *   jsonData storage keys, so they are not modeled as jsonData here.
 */
export type JsonDataConfig = {
  /**
   * Authentication mode discriminator. Stored as a JSON number
   * (`Connection.TOKEN = 0` / `Connection.CREDENTIALS = 1` from
   * `ConfigEditor.tsx:11-14`). Defaults to 0 for a fresh datasource because
   * the editor uses `jsonData.authKind || Connection.TOKEN` on load
   * (`ConfigEditor.tsx:71`).
   */
  authKind?: AstraAuthKind;
  /**
   * Astra Cloud gRPC URI in host:port form (no scheme), e.g.
   * `$ASTRA_CLUSTER_ID-$ASTRA_REGION.apps.astra.datastax.com:443`.
   * Passed directly to `grpc.NewClient` in Token mode
   * (`pkg/plugin/handlers_querydata.go:106`). Only used when `authKind === 0`.
   */
  uri?: string;
  /**
   * Self-hosted Stargate gRPC listener in host:port form (no scheme), e.g.
   * `localhost:8090`. Passed to `grpc.NewClient` in Credentials mode
   * (`pkg/plugin/handlers_querydata.go:114,122`). Only used when `authKind === 1`.
   */
  grpcEndpoint?: string;
  /**
   * Self-hosted Stargate REST auth listener in host:port form (no scheme), e.g.
   * `localhost:8081`. The backend prepends `https://` (or `http://` when
   * `secure` is false) and appends `/v1/auth` before calling
   * `NewTableBasedTokenProvider` (`pkg/plugin/handlers_querydata.go:117,125`).
   * Only used when `authKind === 1`.
   */
  authEndpoint?: string;
  /**
   * Basic-auth username for the Stargate table-based auth endpoint. Consumed
   * at `pkg/plugin/handlers_querydata.go:117,125`. Only used when
   * `authKind === 1`. Upstream editor placeholder is `localhost:8090` — a
   * copy-paste error preserved verbatim in the schema.
   */
  user?: string;
  /**
   * TLS toggle for the Credentials-mode Stargate connection. `true` uses
   * `credentials.NewTLS` on the gRPC channel and `https://` on the auth
   * endpoint; `false` uses `insecure.NewCredentials` and `http://`
   * (`pkg/plugin/handlers_querydata.go:112-129`). Ignored when
   * `authKind === 0` (Token mode always uses TLS).
   */
  secure?: boolean;
};

/**
 * Secret key names stored in `secureJsonData` (write-only; read existing
 * config via `secureJsonFields`):
 * - `token` — Astra application token (`AstraCS:...`), used in Token mode.
 * - `password` — Basic-auth password for the Stargate REST auth endpoint,
 *   used in Credentials mode.
 */
export type SecureJsonDataConfig = Array<'token' | 'password'>;
