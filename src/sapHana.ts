export type sapHanaConfig = {
  jsonData: {
    /** REQUIRED. SAP HANA server hostname. */
    server: string;
    /** REQUIRED. Database username. */
    username: string;
    /** OPTIONAL. Server port. Defaults based on instance number. */
    port?: number;
    /** OPTIONAL. Instance number (alternative to port). */
    instance?: number;
    /** OPTIONAL. Database name (for multi-tenant systems). */
    databaseName?: string;
    /** OPTIONAL. Default schema for queries. */
    defaultSchema?: string;
    /** OPTIONAL. Skip TLS certificate verification. */
    tlsSkipVerify?: boolean;
    /** OPTIONAL. Enable TLS client certificate authentication. */
    tlsAuth?: boolean;
    /** OPTIONAL. Enable TLS CA certificate verification. */
    tlsAuthWithCACert?: boolean;
    /** OPTIONAL. Connection timeout in seconds. Defaults to "30". */
    timeout?: string;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    /** REQUIRED (unless tlsAuth). Database password. */
    password: string;
    /** OPTIONAL. TLS CA certificate content. */
    tlsCACert?: string;
    /** OPTIONAL. TLS client certificate content. */
    tlsClientCert?: string;
    /** OPTIONAL. TLS client key content. */
    tlsClientKey?: string;
  };
};
