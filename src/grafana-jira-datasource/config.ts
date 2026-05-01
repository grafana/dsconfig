export type jiraConfig = {
  jsonData: {
    
    url: string;
    
    user?: string;
    
    hosting: string;
    
    scopedToken?: boolean;
    
    cloudId?: string;
    
    enableSecureSocksProxy?: boolean;
    
    authMethod?: JiraAuthMethod;
    
    oauthClientID?: string;
  };
  secureJsonData: {
    
    token?: string;
    
    oauthClientSecret?: string;
  };
};

export type JiraAuthMethod = "basicAuth" | "oauth2";
