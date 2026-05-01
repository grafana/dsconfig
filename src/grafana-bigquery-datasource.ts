export type grafanaBigqueryDatasourceConfig = {
  jsonData: {
    /** OPTIONAL. Google authentication type (jwt, gce, key). */
    authenticationType?: string;
    /** OPTIONAL. Default GCP project ID. */
    defaultProject?: string;
    /** OPTIONAL. Service account client email. */
    clientEmail?: string;
    /** OPTIONAL. OAuth2 token URI. */
    tokenUri?: string;
    /** OPTIONAL. Flat-rate billing project override. */
    flatRateProject?: string;
    /** OPTIONAL. BigQuery processing location. */
    processingLocation?: string;
    /** OPTIONAL. Query priority (INTERACTIVE or BATCH). */
    queryPriority?: string;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
    /** OPTIONAL. Maximum bytes billed per query. */
    MaxBytesBilled?: number;
    /** OPTIONAL. Custom BigQuery service endpoint. */
    serviceEndpoint?: string;
    /** OPTIONAL. Enable OAuth passthrough. */
    oauthPassThru?: boolean;
  };
  secureJsonData: {
    /** CONDITIONALLY REQUIRED: for JWT auth. Private key (PEM). */
    privateKey?: string;
  };
};
