export type snowflakeConfig = {
  jsonData: {
    
    account: string;
    
    username: string;
    
    authType?: SnowflakeAuthType;
    
    region?: string;
    
    role?: string;
    
    warehouse?: string;
    
    database?: string;
    
    schema?: string;
    
    defaultQuery?: string;
    
    defaultVariableQuery?: string;
    
    defaultInterpolation?: string;
    
    timeInterval?: string;
    
    loginTimeout?: number;
    
    requestTimeout?: number;
    
    oauthPassThru?: boolean;
    
    rowLimit?: number;
    
    enableSecureSocksProxy?: boolean;
    
    settings?: SnowflakeSetting[];
  };
  secureJsonData: {
    
    password?: string;
    
    privateKey?: string;
    
    privateKeyPassphrase?: string;
    
    patToken?: string;
  };
};

export type SnowflakeAuthType = "password" | "keypair" | "oauth" | "pat";

export type SnowflakeSetting = {
  name: string;
  value?: string;
  secure: boolean;
};
