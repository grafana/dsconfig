export type snowflakeConfig = {
  jsonData: {
    /** REQUIRED. Snowflake account identifier. */
    account: string;
    /** REQUIRED. Snowflake username. */
    username: string;
    /** OPTIONAL. Authentication type. Defaults to "password". */
    authType?: SnowflakeAuthType;
    /** OPTIONAL. Snowflake region (if not using account URL format). */
    region?: string;
    /** OPTIONAL. Snowflake role to assume. */
    role?: string;
    /** OPTIONAL. Default warehouse. */
    warehouse?: string;
    /** OPTIONAL. Default database. */
    database?: string;
    /** OPTIONAL. Default schema. */
    schema?: string;
    /** OPTIONAL. Default query text for new panels. */
    defaultQuery?: string;
    /** OPTIONAL. Default variable query text. */
    defaultVariableQuery?: string;
    /** OPTIONAL. Default template variable interpolation format. */
    defaultInterpolation?: string;
    /** OPTIONAL. Minimum time interval for time-series queries. */
    timeInterval?: string;
    /** OPTIONAL. Login timeout in seconds. */
    loginTimeout?: number;
    /** OPTIONAL. Request timeout in seconds. */
    requestTimeout?: number;
    /** OPTIONAL. Enable OAuth passthrough. */
    oauthPassThru?: boolean;
    /** OPTIONAL. Maximum rows returned per query. */
    rowLimit?: number;
    /** OPTIONAL. Enable Grafana secure SOCKS proxy. */
    enableSecureSocksProxy?: boolean;
    /** OPTIONAL. Session-level Snowflake settings (key-value pairs). */
    settings?: SnowflakeSetting[];
  };
  secureJsonData: {
    /** CONDITIONALLY REQUIRED: for authType "password". Snowflake password. */
    password?: string;
    /** CONDITIONALLY REQUIRED: for authType "keypair". PEM-encoded private key. */
    privateKey?: string;
    /** OPTIONAL. Passphrase for encrypted private key. */
    privateKeyPassphrase?: string;
    /** CONDITIONALLY REQUIRED: for authType "pat". Programmatic Access Token. */
    patToken?: string;
  };
};

export type SnowflakeAuthType = "password" | "keypair" | "oauth" | "pat";

export type SnowflakeSetting = {
  name: string;
  value?: string;
  secure: boolean;
};
