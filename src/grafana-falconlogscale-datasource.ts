export type grafanaFalconlogscaleDatasourceConfig = {
  jsonData: {
    /** OPTIONAL. LogScale base URL. */
    baseUrl?: string;
    /** OPTIONAL. Enable OAuth passthrough. */
    oauthPassThru?: boolean;
    /** REQUIRED. Whether to authenticate with a personal token. */
    authenticateWithToken: boolean;
    /** OPTIONAL. Default repository. */
    defaultRepository?: string;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
    /** OPTIONAL. Enable incremental querying. */
    incrementalQuerying?: boolean;
    /** OPTIONAL. Incremental query overlap window. */
    incrementalQueryOverlapWindow?: string;
  };
  secureJsonData: {
    /** CONDITIONALLY REQUIRED: for token auth. Access token. */
    accessToken?: string;
  };
};
