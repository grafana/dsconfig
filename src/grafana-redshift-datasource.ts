export type grafanaRedshiftDatasourceConfig = {
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
    /** OPTIONAL. Include CloudTrail events. */
    withEvent?: boolean;
    /** OPTIONAL. Use AWS Secrets Manager for credentials. */
    useManagedSecret?: boolean;
    /** OPTIONAL. Use Redshift Serverless. */
    useServerless?: boolean;
    /** OPTIONAL. Redshift Serverless workgroup name. */
    workgroupName?: string;
    /** OPTIONAL. Redshift cluster identifier. */
    clusterIdentifier?: string;
    /** OPTIONAL. Database name. */
    database?: string;
    /** OPTIONAL. Database user for GetClusterCredentials. */
    dbUser?: string;
    /** OPTIONAL. AWS Secrets Manager secret reference. */
    managedSecret?: { name: string; arn: string };
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
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
