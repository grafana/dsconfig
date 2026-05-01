export type azureDevOpsConfig = {
  jsonData: {
    /** REQUIRED. Azure DevOps organization URL. */
    url: string;
    /** REQUIRED. Authentication type. Currently only "patToken" is supported. */
    authType: "patToken";
    /** OPTIONAL. Maximum number of projects to list. Defaults to 100. */
    projectsLimit?: number;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
    /** OPTIONAL. Username for authentication. */
    username?: string;
  };
  secureJsonData: {
    /** REQUIRED. Azure DevOps Personal Access Token. */
    patToken: string;
  };
};
