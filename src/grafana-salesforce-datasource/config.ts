export type salesforceConfig = {
  jsonData: {
    
    user: string;
    
    authType: SalesforceAuthType;
    
    sandbox?: boolean;
    
    tokenUrl?: SalesforceTokenUrl;
    
    enableSecureSocksProxy?: boolean;
  };
  secureJsonData: {
    
    password?: string;
    
    securityToken?: string;
    
    clientID: string;
    
    clientSecret?: string;
    
    cert?: string;
    
    privateKey?: string;
  };
};

export type SalesforceAuthType = "user" | "jwt";
export type SalesforceTokenUrl = "https://login.salesforce.com" | "https://test.salesforce.com";
