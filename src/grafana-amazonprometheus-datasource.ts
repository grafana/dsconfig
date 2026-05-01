export type grafanaAmazonprometheusDatasourceConfig = {
  jsonData: {
    /** REQUIRED. AWS authentication type. */
    authType?: string;
    /** OPTIONAL. Default AWS region. */
    defaultRegion?: string;
    /** OPTIONAL. IAM role ARN to assume. */
    assumeRoleArn?: string;
    /** OPTIONAL. External ID for STS AssumeRole. */
    externalId?: string;
    /** OPTIONAL. AWS credentials profile name. */
    profile?: string;
    /** OPTIONAL. Custom AWS endpoint URL. */
    endpoint?: string;
    /** OPTIONAL. SigV4 service name override. */
    sigv4Service?: string;
    /** OPTIONAL. Prometheus scrape interval. */
    timeInterval?: string;
    /** OPTIONAL. HTTP method (GET or POST). */
    httpMethod?: string;
    /** OPTIONAL. Custom query parameters. */
    customQueryParameters?: string;
  };
  secureJsonData: {
    /** CONDITIONALLY REQUIRED: for authType "keys". AWS access key. */
    accessKey?: string;
    /** CONDITIONALLY REQUIRED: for authType "keys". AWS secret key. */
    secretKey?: string;
    /** OPTIONAL. AWS session token. */
    sessionToken?: string;
  };
};
