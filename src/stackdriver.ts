export type stackdriverConfig = {
  jsonData: {
    /** OPTIONAL. Google authentication type (jwt, gce, key). */
    authenticationType?: string;
    /** OPTIONAL. Default GCP project ID. */
    defaultProject?: string;
    /** OPTIONAL. GCE default project (when using GCE auth). */
    gceDefaultProject?: string;
    /** OPTIONAL. Service account client email. */
    clientEmail?: string;
    /** OPTIONAL. OAuth2 token URI. */
    tokenUri?: string;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
    /** OPTIONAL. Google universe domain (for non-default environments). */
    universeDomain?: string;
  };
  secureJsonData: {
    /** CONDITIONALLY REQUIRED: for JWT auth. Private key (PEM). */
    privateKey?: string;
  };
};
