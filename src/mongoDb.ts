export type mongoDbConfig = {
  jsonData: {
    /** REQUIRED. MongoDB connection string (mongodb:// or mongodb+srv://). */
    connection: string;
    /** OPTIONAL. Authentication type. Defaults to "BasicAuth". */
    authType?: MongoDbAuthType;
    /** DEPRECATED: use basicAuth fields instead. MongoDB username. */
    user?: string;
    /** OPTIONAL. TLS server name for certificate validation. */
    serverName?: string;
    /** OPTIONAL. Enable TLS client certificate authentication. */
    tlsAuth?: boolean;
    /** OPTIONAL. Enable TLS CA certificate verification. */
    tlsAuthWithCACert?: boolean;
    /** OPTIONAL. Skip TLS certificate verification. */
    tlsSkipVerify?: boolean;
    /** DEPRECATED: use tlsSkipVerify. Skip TLS validation. */
    skipTLSValidation?: boolean;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
    /** OPTIONAL. Maximum rows returned per query. Defaults to "10000". */
    responseRowsLimit?: string;
    /** OPTIONAL. Kerberos username. */
    kerberosUser?: string;
    /** OPTIONAL. Kerberos ccache lookup file path. */
    ccacheLookupFile?: string;
    /** OPTIONAL. Kerberos keytab file path. */
    keyTabFilePath?: string;
    /** OPTIONAL. Kerberos global ccache file path. */
    globalCcacheFilePath?: string;
    /** OPTIONAL. Validate MongoDB credentials on save. */
    validate?: boolean;
  };
  secureJsonData: {
    /** CONDITIONALLY REQUIRED: for BasicAuth. MongoDB password. */
    password?: string;
    /** OPTIONAL. Kerberos password. */
    kerberosPassword?: string;
    /** OPTIONAL. Basic auth password (alternative). */
    basicAuthPassword?: string;
    /** OPTIONAL. TLS certificate key file password. */
    tlsCertificateKeyFilePassword?: string;
  };
};

export type MongoDbAuthType = "NoAuth" | "BasicAuth" | "custom-Kerberos";
