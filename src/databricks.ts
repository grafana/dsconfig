export type databricksConfig = {
  jsonData: {
    /** REQUIRED. Databricks workspace hostname. */
    host: string;
    /** REQUIRED. SQL warehouse or cluster HTTP path. */
    httpPath: string;
    /** REQUIRED. Authentication type. */
    authType: DatabricksAuthType;
    /** OPTIONAL. Request timeout in seconds. Defaults to "60". */
    timeout?: string;
    /** OPTIONAL. Number of retries on transient failures. Defaults to "5". */
    retries?: string;
    /** OPTIONAL. Pause between retries in seconds. Defaults to "0". */
    pause?: string;
    /** OPTIONAL. Maximum rows returned. */
    rows?: string;
    /** OPTIONAL. Retry timeout. */
    retryTimeout?: string;
    /** OPTIONAL. Enable debug logging. */
    debug?: boolean;
    /** OPTIONAL. Default query format. */
    defaultQueryFormat?: number;
    /** OPTIONAL. Enable Unity Catalog support. */
    enableUnitySupport?: boolean;
    /** OPTIONAL. Default database/catalog. */
    database?: string;
    /** OPTIONAL. Enable OAuth passthrough. */
    oauthPassThru?: boolean;
    /** CONDITIONALLY REQUIRED: for AzureM2M/OauthOBO. Azure AD tenant ID. */
    tenantId?: string;
    /** CONDITIONALLY REQUIRED: for OauthM2M/AzureM2M. OAuth client ID. */
    clientId?: string;
    /** OPTIONAL. Azure cloud name. Defaults to "AzureCloud". */
    azureCloud?: string;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    /** CONDITIONALLY REQUIRED: for authType "Pat". Personal access token. */
    token?: string;
    /** CONDITIONALLY REQUIRED: for authType "OauthM2M" or "AzureM2M". OAuth client secret. */
    clientSecret?: string;
    /** CONDITIONALLY REQUIRED: for Azure credentials format. Azure client secret. */
    azureClientSecret?: string;
  };
};

export type DatabricksAuthType = "" | "Pat" | "OauthM2M" | "OauthPT" | "OauthOBO" | "AzureM2M";
