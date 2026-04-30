export type mysqlConfig = {
  jsonData: {
    /** OPTIONAL. Database name. */
    database?: string;
    /** OPTIONAL. Maximum open connections. */
    maxOpenConns?: number;
    /** OPTIONAL. Maximum idle connections. */
    maxIdleConns?: number;
    /** OPTIONAL. Connection max lifetime in seconds. */
    connMaxLifetime?: number;
    /** OPTIONAL. Allow cleartext passwords. */
    allowCleartextPasswords?: boolean;
    /** OPTIONAL. Minimum time interval for auto-grouping. */
    timeInterval?: string;
  };
  secureJsonData: {
    /** REQUIRED. Database password. */
    password?: string;
  };
};
