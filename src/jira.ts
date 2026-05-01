export type jiraConfig = {
  jsonData: {
    /** REQUIRED. Jira instance URL (https:// prefix added automatically if missing). */
    url: string;
    /** OPTIONAL. Jira username for basic auth. */
    user?: string;
    /** REQUIRED. Jira hosting type ("cloud" or "server"). */
    hosting: string;
    /** OPTIONAL. Whether the token is a scoped Atlassian token. */
    scopedToken?: boolean;
    /** CONDITIONALLY REQUIRED: for OAuth2 with Jira Cloud. Atlassian Cloud ID. */
    cloudId?: string;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
    /** OPTIONAL. Authentication method. Defaults to "basicAuth". */
    authMethod?: JiraAuthMethod;
    /** CONDITIONALLY REQUIRED: for OAuth2. OAuth client ID. */
    oauthClientID?: string;
  };
  secureJsonData: {
    /** CONDITIONALLY REQUIRED: for basicAuth. API token or password. */
    token?: string;
    /** CONDITIONALLY REQUIRED: for OAuth2. OAuth client secret. */
    oauthClientSecret?: string;
  };
};

export type JiraAuthMethod = "basicAuth" | "oauth2";
