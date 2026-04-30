export type grafanaAstradbDatasourceConfig = {
  jsonData: {
    /** REQUIRED. AstraDB URI. */
    uri: string;
    /** REQUIRED. Authentication kind. */
    authKind: number;
    /** REQUIRED. Username. */
    user: string;
    /** REQUIRED. gRPC endpoint. */
    grpcEndpoint: string;
    /** REQUIRED. Authentication endpoint. */
    authEndpoint: string;
    /** OPTIONAL. Enable TLS. */
    secure?: boolean;
    /** OPTIONAL. Database name. */
    database?: string;
  };
  secureJsonData: {
    /** OPTIONAL. AstraDB token. */
    token?: string;
    /** OPTIONAL. Database password. */
    password?: string;
  };
};
