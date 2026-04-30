export type grafanaAzureDataExplorerDatasourceConfig = {
  jsonData: {
    /** REQUIRED. Azure Data Explorer cluster URL. */
    clusterUrl: string;
    /** REQUIRED. Default database. */
    defaultDatabase: string;
    /** OPTIONAL. Azure authentication type. */
    azureAuthType?: string;
    /** OPTIONAL. Azure cloud name. */
    cloudName?: string;
    /** OPTIONAL. Azure AD tenant ID. */
    tenantId?: string;
    /** OPTIONAL. Azure AD client ID. */
    clientId?: string;
    /** OPTIONAL. Minimal cache duration in seconds. */
    minimalCache?: number;
    /** OPTIONAL. Default query editor mode. */
    defaultEditorMode?: string;
    /** OPTIONAL. Query timeout duration string. */
    queryTimeout?: string;
    /** OPTIONAL. Data consistency level. */
    dataConsistency?: string;
    /** OPTIONAL. Maximum cache age. */
    cacheMaxAge?: string;
    /** OPTIONAL. Enable dynamic caching. */
    dynamicCaching?: boolean;
    /** OPTIONAL. Enable schema mapping. */
    useSchemaMapping?: boolean;
    /** OPTIONAL. Enable user tracking. */
    enableUserTracking?: boolean;
    /** OPTIONAL. Application name identifier. */
    application?: string;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    /** CONDITIONALLY REQUIRED: for client secret auth. Azure AD client secret. */
    clientSecret?: string;
    /** OPTIONAL. OpenAI API key for AI features. */
    OpenAIAPIKey?: string;
  };
};
