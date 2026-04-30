export type mssqlConfig = {
  jsonData: {
    /** OPTIONAL. Authentication method. */
    authenticationType?: MSSQLAuthenticationType;
    /** OPTIONAL. Connection encryption mode. */
    encrypt?: MSSQLEncryptOption;
    /** OPTIONAL. SSL root certificate file path. */
    sslRootCertFile?: string;
    /** OPTIONAL. TLS server name for certificate verification. */
    serverName?: string;
    /** OPTIONAL. Connection timeout in seconds. */
    connectionTimeout?: number;
    /** OPTIONAL. Database name. */
    database?: string;
    /** OPTIONAL. Maximum open connections. */
    maxOpenConns?: number;
    /** OPTIONAL. Maximum idle connections. */
    maxIdleConns?: number;
    /** OPTIONAL. Connection max lifetime in seconds. */
    connMaxLifetime?: number;
    /** OPTIONAL. Kerberos keytab file path. */
    keytabFilePath?: string;
    /** OPTIONAL. Kerberos credential cache. */
    credentialCache?: string;
    /** OPTIONAL. Kerberos credential cache lookup file. */
    credentialCacheLookupFile?: string;
    /** OPTIONAL. Kerberos config file path. */
    configFilePath?: string;
    /** OPTIONAL. UDP connection limit for Kerberos. */
    UDPConnectionLimit?: number;
    /** OPTIONAL. Enable DNS lookup for KDC. */
    enableDNSLookupKDC?: string;
    /** OPTIONAL. Minimum time interval for auto-grouping. */
    timeInterval?: string;
  };
  secureJsonData: {
    /** REQUIRED. Database password. */
    password?: string;
  };
};

export type MSSQLAuthenticationType =
  | "SQL Server Authentication"
  | "Windows Authentication"
  | "Azure AD Authentication"
  | "Windows AD: Username + password"
  | "Windows AD: Keytab"
  | "Windows AD: Credential cache"
  | "Windows AD: Credential cache file";

export type MSSQLEncryptOption = "disable" | "false" | "true";
