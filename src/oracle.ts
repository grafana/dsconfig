export type oracleConfig = {
  url: string;
  jsonData: {
    /** OPTIONAL. Oracle database name / service name. */
    database?: string;
    /** OPTIONAL. Oracle timezone name. Defaults to "UTC". */
    timezone_name?: string;
    /** OPTIONAL. Enable daylight saving time adjustment. */
    use_dst?: boolean;
    /** OPTIONAL. Database username. */
    user?: string;
    /** OPTIONAL. Use TNS Names-based connection. */
    useTNSNamesBasedConnection?: boolean;
    /** CONDITIONALLY REQUIRED: when useTNSNamesBasedConnection is true. TNS Names entry. */
    tnsNamesEntry?: string;
    /** OPTIONAL. Enable Kerberos authentication. */
    useKerberosAuthentication?: boolean;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
    /** OPTIONAL. Connection pool size. Defaults to 50. */
    connectionPoolSize?: number;
    /** OPTIONAL. Data proxy timeout in seconds. Defaults to 120. */
    dataProxyTimeout?: number;
    /** OPTIONAL. Number of rows to prefetch. */
    prefetchRowsCount?: number;
    /** OPTIONAL. Maximum rows returned. Defaults to 1000000. */
    rowLimit?: number;
  };
  secureJsonData: {
    /** REQUIRED. Database password. */
    password: string;
  };
};
