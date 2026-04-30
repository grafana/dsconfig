export type salesforceConfig = {
  jsonData: {
    /** REQUIRED. Salesforce username. */
    user: string;
    /** REQUIRED. Authentication type. */
    authType: SalesforceAuthType;
    /** DEPRECATED: use tokenUrl. When true, uses sandbox token URL. */
    sandbox?: boolean;
    /** OPTIONAL. Salesforce token URL override. */
    tokenUrl?: SalesforceTokenUrl;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    /** CONDITIONALLY REQUIRED: for authType "user". Salesforce password. */
    password?: string;
    /** CONDITIONALLY REQUIRED: for authType "user". Salesforce security token. */
    securityToken?: string;
    /** REQUIRED. Connected App client ID. */
    clientID: string;
    /** CONDITIONALLY REQUIRED: for authType "user". Connected App client secret. */
    clientSecret?: string;
    /** CONDITIONALLY REQUIRED: for authType "jwt". X.509 certificate (PEM). */
    cert?: string;
    /** CONDITIONALLY REQUIRED: for authType "jwt". Private key (PEM). */
    privateKey?: string;
  };
};

export type SalesforceAuthType = "user" | "jwt";
export type SalesforceTokenUrl = "https://login.salesforce.com" | "https://test.salesforce.com";
