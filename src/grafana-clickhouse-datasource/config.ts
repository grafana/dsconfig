export type grafanaClickhouseDatasourceConfig = {
  jsonData: {

    host: string;

    port: number;

    protocol: ClickHouseProtocol;

    username?: string;

    version?: string;

    secure?: boolean;

    path?: string;

    tlsSkipVerify?: boolean;

    tlsAuth?: boolean;

    tlsAuthWithCACert?: boolean;

    defaultDatabase?: string;

    defaultTable?: string;

    connMaxLifetime?: string;

    dialTimeout?: string;

    maxIdleConns?: string;

    maxOpenConns?: string;

    queryTimeout?: string;

    validateSql?: boolean;

    forwardGrafanaHeaders?: boolean;

    enableRowLimit?: boolean;

    hideTableNameInAdhocFilters?: boolean;

    /** Logs configuration */
    logs?: ClickHouseLogsConfig;

    /** Traces configuration */
    traces?: ClickHouseTracesConfig;

    /** Column alias tables */
    aliasTables?: ClickHouseAliasTable[];

    /** Custom ClickHouse SETTINGS key-value pairs */
    customSettings?: ClickHouseCustomSetting[];
  };
  secureJsonData: {

    password?: string;

    tlsCACert?: string;

    tlsClientCert?: string;

    tlsClientKey?: string;
  };
};

export type ClickHouseProtocol = "native" | "http";

export type ClickHouseOtelVersion = "latest" | "1.29.0";

export type ClickHouseDurationUnit = "seconds" | "milliseconds" | "microseconds" | "nanoseconds";

export type ClickHouseLogsConfig = {
  defaultDatabase?: string;
  defaultTable?: string;
  otelEnabled?: boolean;
  otelVersion?: ClickHouseOtelVersion;
  filterTimeColumn?: string;
  timeColumn?: string;
  levelColumn?: string;
  messageColumn?: string;
  showLogLinks?: boolean;
  selectContextColumns?: boolean;
  contextColumns?: string[];
};

export type ClickHouseTracesConfig = {
  defaultDatabase?: string;
  defaultTable?: string;
  otelEnabled?: boolean;
  otelVersion?: ClickHouseOtelVersion;
  traceIdColumn?: string;
  spanIdColumn?: string;
  operationNameColumn?: string;
  parentSpanIdColumn?: string;
  serviceNameColumn?: string;
  durationColumn?: string;
  durationUnit?: ClickHouseDurationUnit;
  startTimeColumn?: string;
  tagsColumn?: string;
  serviceTagsColumn?: string;
  kindColumn?: string;
  statusCodeColumn?: string;
  statusMessageColumn?: string;
  stateColumn?: string;
  instrumentationLibraryNameColumn?: string;
  instrumentationLibraryVersionColumn?: string;
  flattenNested?: boolean;
  traceEventsColumnPrefix?: string;
  traceLinksColumnPrefix?: string;
  showTraceLinks?: boolean;
};

export type ClickHouseAliasTable = {
  targetDatabase?: string;
  targetTable: string;
  aliasDatabase?: string;
  aliasTable: string;
};

export type ClickHouseCustomSetting = {
  setting: string;
  value: string;
};
