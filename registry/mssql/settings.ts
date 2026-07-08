/**
 * Configuration models for the Microsoft SQL Server datasource plugin (`mssql`).
 *
 * Sources of truth (https://github.com/grafana/grafana-mssql-datasource @ c4133924):
 * - `src/types.ts` — `MssqlOptions extends SQLOptions`, `MSSQLAuthenticationType`, `MSSQLEncryptOptions`
 * - `src/configuration/ConfigurationEditor.tsx` — the primary configuration editor
 * - `src/configuration/Kerberos.tsx` — Kerberos-specific sub-panels (KerberosConfig, KerberosAdvancedSettings)
 * - `src/azureauth/*` — Azure AD authentication form (delegates to @grafana/azure-sdk)
 * - `pkg/mssql/mssql.go` — backend Datasource setup
 * - `pkg/mssql/sqleng/sql_engine.go:48-69` — shared `sqleng.JsonData` shape
 * - `pkg/mssql/kerberos/kerberos.go:22-59` — Kerberos jsonData parsing (KeytabFilePath, CredentialCache, CredentialCacheLookupFile, ConfigFilePath, UDPConnectionLimit, EnableDNSLookupKDC)
 * - `pkg/mssql/azure/connection.go` — Azure credentials DSN fragment
 * - External:
 *   - `@grafana/sql` 13.0.2: `SQLOptions`, `ConnectionLimits`, `NumberInput`, `useMigrateDatabaseFields`
 *   - `@grafana/azure-sdk` 0.1.0: `AzureCredentials` type
 */

export type MSSQLAuthenticationType =
  | 'SQL Server Authentication'
  | 'Windows Authentication'
  | 'Azure AD Authentication'
  | 'Windows AD: Username + password'
  | 'Windows AD: Keytab'
  | 'Windows AD: Credential cache'
  | 'Windows AD: Credential cache file';

export type MSSQLEncryptOption = 'disable' | 'false' | 'true';

/**
 * Azure AD credential shape — mirrors @grafana/azure-sdk `AzureCredentials`
 * (github.com/grafana/grafana-azure-sdk-react src/credentials/AzureCredentials.ts).
 * Only populated when `jsonData.authenticationType === 'Azure AD Authentication'`.
 */
export type AzureCredentials =
  | { authType: 'msi' }
  | { authType: 'workloadidentity' }
  | { authType: 'currentuser'; serviceCredentials?: AzureCredentials; serviceCredentialsEnabled?: boolean }
  | { authType: 'clientsecret'; azureCloud?: string; tenantId?: string; clientId?: string }
  | { authType: 'clientsecret-obo'; azureCloud?: string; tenantId?: string; clientId?: string }
  | { authType: 'ad-password'; userId?: string; clientId?: string }
  | { authType: 'clientcertificate'; azureCloud?: string; clientId?: string; tenantId?: string; certificateFormat?: 'pem' | 'pfx' };

/**
 * Root (top-level datasource settings) fields the MSSQL backend actually reads.
 */
export type RootConfig = {
  /** MSSQL host+port. Read at `pkg/mssql/mssql.go:56`. */
  url?: string;
  /** Database user. Read at `pkg/mssql/mssql.go:57`. */
  user?: string;
  /**
   * Legacy: older datasources stored the database name at root level.
   * `useMigrateDatabaseFields` migrates it; the backend (`pkg/mssql/mssql.go:49-52`)
   * falls back to root.database when jsonData.database is empty.
   */
  database?: string;
};

/**
 * Fields stored in `jsonData`. Union of the plugin's `MssqlOptions` (`src/types.ts:31-44`)
 * and the shared `sqleng.JsonData` (`pkg/mssql/sqleng/sql_engine.go:48-69`).
 */
export type JsonDataConfig = {
  /** Preferred way to store the database name. */
  database?: string;
  /** Auth-type discriminator (`src/types.ts:16-24`). */
  authenticationType?: MSSQLAuthenticationType;
  /** SSL/TLS negotiation. Default `'false'` (`ConfigurationEditor.tsx:231`). */
  encrypt?: MSSQLEncryptOption;
  /** Only meaningful when `encrypt === 'true'`. */
  tlsSkipVerify?: boolean;
  /** File-path root cert; only used when `encrypt === 'true' && tlsSkipVerify === false`. */
  sslRootCertFile?: string;
  /** Expected server-cert Common Name; only used when `encrypt === 'true' && tlsSkipVerify === false`. */
  serverName?: string;
  /** Kerberos: keytab file path (auth type 'Windows AD: Keytab'). */
  keytabFilePath?: string;
  /** Kerberos: credential-cache path (auth type 'Windows AD: Credential cache'). */
  credentialCache?: string;
  /** Kerberos: credential-cache lookup file path (auth type 'Windows AD: Credential cache file'). */
  credentialCacheLookupFile?: string;
  /** Kerberos advanced: MIT krb5 config file path. Default `/etc/krb5.conf`. */
  configFilePath?: string;
  /** Kerberos advanced: UDP preference limit. Default `1` (always use TCP). */
  UDPConnectionLimit?: number;
  /** Kerberos advanced: whether DNS SRV lookup for KDCs is enabled. Default `'true'` (string, not bool). */
  enableDNSLookupKDC?: string;
  /** Azure AD credentials (object; only when authenticationType is 'Azure AD Authentication'). */
  azureCredentials?: AzureCredentials;
  /** Auto group-by lower bound, e.g. `'1m'`. */
  timeInterval?: string;
  /** Connection timeout in seconds (`0` = no timeout). */
  connectionTimeout?: number;
  /** Connection pool tuning. */
  maxOpenConns?: number;
  maxIdleConns?: number;
  maxIdleConnsAuto?: boolean;
  connMaxLifetime?: number;
};

/**
 * Secret key names stored in `secureJsonData` (write-only):
 * - `password` — SQL / Kerberos password
 * - `azureClientSecret` — Azure AD client secret (also `azureClientCertificate`,
 *   `azureClientCertificatePassword`, `azureClientCertificatePrivateKey` for cert-based auth,
 *   managed by @grafana/azure-sdk — see AzureCredentialsConfig.ts:428).
 */
export type SecureJsonDataConfig = Array<
  | 'password'
  | 'azureClientSecret'
>;
