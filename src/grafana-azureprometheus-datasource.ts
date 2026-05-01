export type grafanaAzurepromethousDatasourceConfig = {
  jsonData: {
    /** OPTIONAL. Azure authentication type. */
    azureAuthType?: string;
    /** OPTIONAL. Azure cloud name. */
    cloudName?: string;
    /** OPTIONAL. Azure AD tenant ID. */
    tenantId?: string;
    /** OPTIONAL. Azure AD client ID. */
    clientId?: string;
    /** OPTIONAL. Azure endpoint resource ID. */
    azureEndpointResourceId?: string;
    /** OPTIONAL. Prometheus scrape interval. */
    timeInterval?: string;
    /** OPTIONAL. HTTP method (GET or POST). */
    httpMethod?: string;
    /** OPTIONAL. Custom query parameters. */
    customQueryParameters?: string;
  };
  secureJsonData: {
    /** CONDITIONALLY REQUIRED: for client secret auth. Azure AD client secret. */
    clientSecret?: string;
  };
};
