export type odbcConfig = {
  jsonData: {
    /** REQUIRED. Path to the ODBC driver file. */
    driver: string;
    /** OPTIONAL. Query timeout in seconds. Defaults to "10". */
    timeout?: string;
    /** OPTIONAL. Connection settings (key-value pairs, some may be secure). */
    settings?: OdbcSetting[];
  };
  secureJsonData: {
    /** Dynamic keys: secure setting values are stored here by name. */
    [key: string]: string;
  };
};

export type OdbcSetting = {
  /** Setting name (also used as key in secureJsonData when secure=true). */
  name: string;
  /** Setting value (only for non-secure settings). */
  value?: string;
  /** Whether this setting's value should be stored in secureJsonData. */
  secure: boolean;
};
