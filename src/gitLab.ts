export type gitLabConfig = {
  /** OPTIONAL. GitLab API URL. Defaults to "https://gitlab.com/api/v4". */
  url?: string;
  jsonData: {
    /** OPTIONAL. Maximum pages to fetch per request. Defaults to 5. */
    pageLimit?: number;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    /** REQUIRED. GitLab personal access token. */
    accessToken: string;
  };
};
