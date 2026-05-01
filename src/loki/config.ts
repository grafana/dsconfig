export type lokiConfig = {
    url: string;
    basicAuth: boolean;
    basicAuthUser?: string;
    jsonData: {
        manageAlerts?: boolean;
        maxLines?: string;
        derivedFields?: DerivedFieldConfig[];
        alertmanager?: string;
        timeout?: number;
        serverName?: string;
        tlsAuth?: boolean;
        tlsAuthWithCACert?: boolean;
        tlsSkipVerify?: boolean;
        keepCookies?: string[];
        pdcInjected?: boolean;
    };
    secureJsonData: {
        basicAuthPassword?: string;
        tlsCACert?: string;
        tlsClientCert?: string;
        tlsClientKey?: string;
    };
};

type DerivedFieldConfig = {
    matcherRegex: string;
    name: string;
    url?: string;
    urlDisplayLabel?: string;
    datasourceUid?: string;
    matcherType?: "label" | "regex";
    targetBlank?: boolean;
};
