export type grafanaTimestreamDatasourceConfig = {
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
    /** OPTIONAL. Default Timestream database. */
    defaultDatabase?: string;
    /** OPTIONAL. Default Timestream table. */
    defaultTable?: string;
    /** OPTIONAL. Default measure name. */
    defaultMeasure?: string;
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
