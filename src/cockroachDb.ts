export type cockroachDbConfig = {
  url: string;
  user?: string;
  jsonData: {
    /** REQUIRED. Database name. */
    database: string;
    /** OPTIONAL. SSL mode for connections. */
    sslmode?: CockroachDbTlsMode;
    /** OPTIONAL. Method for providing TLS certificates. */
    tlsConfigurationMethod?: CockroachDbTlsMethod;
    /** OPTIONAL. Path to SSL root certificate file. */
    sslRootCertFile?: string;
    /** OPTIONAL. Path to SSL client certificate file. */
    sslCertFile?: string;
    /** OPTIONAL. Path to SSL client key file. */
    sslKeyFile?: string;
    /** OPTIONAL. Authentication type (SQL, Kerberos, TLS). */
    authType?: string;
    /** OPTIONAL. Maximum open connections. Defaults to 5. */
    maxOpenConns?: number;
    /** OPTIONAL. Maximum idle connections. Defaults to 2. */
    maxIdleConns?: number;
    /** OPTIONAL. Auto-set max idle connections. */
    maxIdleConnsAuto?: boolean;
    /** OPTIONAL. Connection max lifetime in seconds. Defaults to 300. */
    connMaxLifetime?: number;
    /** OPTIONAL. Query timeout in seconds (5-600). Defaults to 30. */
    queryTimeout?: number;
    /** OPTIONAL. Kerberos configuration file path. */
    configFilePath?: string;
    /** OPTIONAL. Kerberos credential cache path. */
    credentialCache?: string;
    /** OPTIONAL. Kerberos server name. */
    kerberosServerName?: string;
    /** OPTIONAL. PostgreSQL version compatibility. */
    postgresVersion?: number;
    /** OPTIONAL. Enable TimescaleDB support. */
    timescaledb?: boolean;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    /** REQUIRED. Database password. */
    password: string;
  };
};

export type CockroachDbTlsMode = "disable" | "require" | "verify-ca" | "verify-full";
export type CockroachDbTlsMethod = "file-path" | "file-content";
