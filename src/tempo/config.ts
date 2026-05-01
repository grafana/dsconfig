export type tempoConfig = {
    url: string;
    basicAuth: boolean;
    basicAuthUser?: string;
    jsonData: {
        serviceMap?: {
            datasourceUid?: string;
        };
        search?: {
            hide?: boolean;
            filters?: TraceqlFilter[];
        };
        nodeGraph?: NodeGraphOptions;
        spanBar?: {
            tag: string;
        };
        tagLimit?: number;
        traceQuery?: {
            timeShiftEnabled?: boolean;
            spanStartTimeShift?: string;
            spanEndTimeShift?: string;
        };
        streamingEnabled?: {
            search?: boolean;
            metrics?: boolean;
        };
        timeRangeForTags?: number;
        tracesToMetrics?: TraceToMetricsOptions;
        tracesToLogs?: TraceToLogsOptions;
        tracesToLogsV2?: TraceToLogsOptionsV2;
        tracesToProfiles?: TraceToProfilesOptions;
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

export interface TraceToMetricsOptions {
    datasourceUid?: string;
    tags?: Array<{ key: string; value: string }>;
    queries: TraceToMetricQuery[];
    spanStartTimeShift?: string;
    spanEndTimeShift?: string;
}

export interface TraceToMetricQuery {
    name?: string;
    query?: string;
}

export interface TraceToLogsOptions {
    datasourceUid?: string;
    tags?: string[];
    mappedTags?: TraceToLogsTag[];
    mapTagNamesEnabled?: boolean;
    spanStartTimeShift?: string;
    spanEndTimeShift?: string;
    filterByTraceID?: boolean;
    filterBySpanID?: boolean;
    lokiSearch?: boolean; 
}

export interface TraceToLogsOptionsV2 {
    datasourceUid?: string;
    tags?: TraceToLogsTag[];
    spanStartTimeShift?: string;
    spanEndTimeShift?: string;
    filterByTraceID?: boolean;
    filterBySpanID?: boolean;
    query?: string;
    customQuery: boolean;
}

export interface TraceToLogsTag {
    key: string;
    value?: string;
}

export interface TraceToProfilesOptions {
    datasourceUid?: string;
    tags?: Array<{ key: string; value?: string }>;
    query?: string;
    profileTypeId?: string;
    customQuery: boolean;
}

export interface TraceqlFilter {
    id: string;
    isCustomValue?: boolean;
    operator?: string;
    scope?: TraceqlSearchScope;
    tag?: string;
    value?: (string | Array<string>);
    valueType?: string;
}

export enum TraceqlSearchScope {
    Event = 'event',
    Instrumentation = 'instrumentation',
    Intrinsic = 'intrinsic',
    Link = 'link',
    Resource = 'resource',
    Span = 'span',
    Unscoped = 'unscoped',
}

export interface NodeGraphOptions {
    enabled?: boolean;
}
