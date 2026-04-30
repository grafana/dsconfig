export type grafanaAthenaDatasourceConfig = {
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
    /** OPTIONAL. Athena data catalog name. */
    catalog?: string;
    /** OPTIONAL. Default database. */
    database?: string;
    /** OPTIONAL. Athena workgroup. */
    workgroup?: string;
    /** OPTIONAL. S3 output location for query results. */
    outputLocation?: string;
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
