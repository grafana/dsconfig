export type prometheusConfig = {
    url: string;
    basicAuth: boolean;
    basicAuthUser?: string;
    jsonData: {
        timeInterval?: string;
        queryTimeout?: string;
        httpMethod?: string; 
        customQueryParameters?: string;
        disableMetricsLookup?: boolean;
        exemplarTraceIdDestinations?: ExemplarTraceIdDestination[];
        prometheusType?: PromApplication; 
        prometheusVersion?: string;
        cacheLevel?: PrometheusCacheLevel; 
        defaultEditor?: QueryEditorMode; 
        incrementalQuerying?: boolean;
        incrementalQueryOverlapWindow?: string;
        disableRecordingRules?: boolean;
        manageAlerts?: boolean;
        allowAsRecordingRulesTarget?: boolean;
        oauthPassThru?: boolean;
        seriesEndpoint?: boolean;
        seriesLimit?: number;
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

enum PromApplication {
    Cortex = "Cortex",
    Mimir = "Mimir",
    Prometheus = "Prometheus",
    Thanos = "Thanos",
}

enum PrometheusCacheLevel {
    Low = "Low",
    Medium = "Medium",
    High = "High",
    None = "None",
}

enum QueryEditorMode {
    Code = "code",
    Builder = "builder",
}

type ExemplarTraceIdDestination = {
    name: string;
    url?: string;
    urlDisplayLabel?: string;
    datasourceUid?: string;
};
