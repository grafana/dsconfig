// Re-export schema types for convenience
export type { DatasourceConfigSchema, ConfigField, ConfigGroup } from "../schema/schema";

// Core Grafana datasources
export { prometheusConfig } from "./prometheus/config";
export { lokiConfig } from "./loki/config";
export { tempoConfig } from "./tempo/config";
export { cloudWatchConfig, AWSAuthType } from "./cloudwatch/config";
export { azureMonitorConfig } from "./grafana-azure-monitor-datasource/config";
export { alertmanagerConfig } from "./alertmanager/config";
export { elasticsearchConfig } from "./elasticsearch/config";
export { graphiteConfig } from "./graphite/config";
export { influxdbConfig } from "./influxdb/config";
export { jaegerConfig } from "./jaeger/config";
export { mssqlConfig } from "./mssql/config";
export { mysqlConfig } from "./mysql/config";
export { opentsdbConfig } from "./opentsdb/config";
export { grafanaPyroscopeConfig } from "./grafana-pyroscope-datasource/config";
export { zipkinConfig } from "./zipkin/config";
export { postgresConfig } from "./grafana-postgresql-datasource/config";
export { stackdriverConfig } from "./stackdriver/config";

// AWS datasources
export { grafanaAthenaDatasourceConfig } from "./grafana-athena-datasource/config";
export { grafanaRedshiftDatasourceConfig } from "./grafana-redshift-datasource/config";
export { grafanaTimestreamDatasourceConfig } from "./grafana-timestream-datasource/config";
export { grafanaXRayDatasourceConfig } from "./grafana-x-ray-datasource/config";
export { grafanaIotSitewiseDatasourceConfig } from "./grafana-iot-sitewise-datasource/config";
export { grafanaAmazonprometheusDatasourceConfig } from "./grafana-amazonprometheus-datasource/config";

// Azure datasources
export { grafanaAzureDataExplorerDatasourceConfig } from "./grafana-azure-data-explorer-datasource/config";
export { grafanaAzurepromethousDatasourceConfig } from "./grafana-azureprometheus-datasource/config";

// Google datasources
export { grafanaBigqueryDatasourceConfig } from "./grafana-bigquery-datasource/config";
export { googleSheetsConfig } from "./grafana-googlesheets-datasource/config";
export { GoogleAuthType } from "./grafana-googlesheets-datasource/config";

// Other Grafana-owned datasources
export { grafanaClickhouseDatasourceConfig } from "./grafana-clickhouse-datasource/config";
export { grafanaOpensearchDatasourceConfig } from "./grafana-opensearch-datasource/config";
export { yesoreyeramInfinityDatasourceConfig } from "./yesoreyeram-infinity-datasource/config";
export { grafanaFalconlogscaleDatasourceConfig } from "./grafana-falconlogscale-datasource/config";
export { grafanaSentryDatasourceConfig } from "./grafana-sentry-datasource/config";
export { grafanaMqttDatasourceConfig } from "./grafana-mqtt-datasource/config";
export { grafanaStravaDatasourceConfig } from "./grafana-strava-datasource/config";
export { grafanaSurrealdbDatasourceConfig } from "./grafana-surrealdb-datasource/config";
export { grafanaYugabyteDatasourceConfig } from "./grafana-yugabyte-datasource/config";
export { grafanaAstradbDatasourceConfig } from "./grafana-astradb-datasource/config";
export { marcusolssonJsonDatasourceConfig } from "./marcusolsson-json-datasource/config";
export { marcusolssonCsvDatasourceConfig } from "./marcusolsson-csv-datasource/config";
export { marcusolssonStaticDatasourceConfig } from "./marcusolsson-static-datasource/config";
export { githubConfig as grafanaGithubDatasourceConfig } from "./grafana-github-datasource/config";

// Enterprise datasources (plugins-private)
export { newRelicConfig } from "./grafana-newrelic-datasource/config";
export { appDynamicsConfig } from "./grafana-appdynamics-datasource/config";
export { azureDevOpsConfig } from "./grafana-azuredevops-datasource/config";
export { cockroachDbConfig } from "./grafana-cockroachdb-datasource/config";
export { databricksConfig } from "./grafana-databricks-datasource/config";
export { datadogConfig } from "./grafana-datadog-datasource/config";
export { dynatraceConfig } from "./grafana-dynatrace-datasource/config";
export { gitLabConfig } from "./grafana-gitlab-datasource/config";
export { honeycombConfig } from "./grafana-honeycomb-datasource/config";
export { jiraConfig } from "./grafana-jira-datasource/config";
export { lookerConfig } from "./grafana-looker-datasource/config";
export { mongoDbConfig } from "./grafana-mongodb-datasource/config";
export { odbcConfig } from "./grafana-odbc-datasource/config";
export { oracleConfig } from "./grafana-oracle-datasource/config";
export { pagerDutyConfig } from "./grafana-pagerduty-datasource/config";
export { salesforceConfig } from "./grafana-salesforce-datasource/config";
export { sapHanaConfig } from "./grafana-saphana-datasource/config";
export { serviceNowConfig } from "./grafana-servicenow-datasource/config";
export { snowflakeConfig } from "./grafana-snowflake-datasource/config";
export { splunkConfig } from "./grafana-splunk-datasource/config";
export { splunkObservabilityConfig } from "./grafana-splunkobservability-datasource/config";
export { sumoLogicConfig } from "./grafana-sumologic-datasource/config";
export { wavefrontConfig } from "./grafana-wavefront-datasource/config";

// ============================================================
// Plugin type constants — all known datasource plugin IDs
// ============================================================

export const DATASOURCE_PLUGIN_TYPES = [
    // Core (grafana/grafana)
    "prometheus",
    "loki",
    "tempo",
    "cloudwatch",
    "grafana-azure-monitor-datasource",
    "alertmanager",
    "elasticsearch",
    "graphite",
    "influxdb",
    "jaeger",
    "mssql",
    "mysql",
    "opentsdb",
    "grafana-pyroscope-datasource",
    "zipkin",
    "grafana-postgresql-datasource",
    "stackdriver",
    // AWS
    "grafana-athena-datasource",
    "grafana-redshift-datasource",
    "grafana-timestream-datasource",
    "grafana-x-ray-datasource",
    "grafana-iot-sitewise-datasource",
    "grafana-amazonprometheus-datasource",
    // Azure
    "grafana-azure-data-explorer-datasource",
    "grafana-azureprometheus-datasource",
    "grafana-azuredevops-datasource",
    // Google
    "grafana-bigquery-datasource",
    "grafana-googlesheets-datasource",
    // Community / Grafana-owned
    "grafana-clickhouse-datasource",
    "grafana-opensearch-datasource",
    "yesoreyeram-infinity-datasource",
    "grafana-falconlogscale-datasource",
    "grafana-mqtt-datasource",
    "grafana-surrealdb-datasource",
    "marcusolsson-csv-datasource",
    "marcusolsson-json-datasource",
    "grafana-github-datasource",
    "grafana-strava-datasource",
    "grafana-astradb-datasource",
    "grafana-yugabyte-datasource",
    "grafana-sentry-datasource",
    // Enterprise
    "grafana-appdynamics-datasource",
    "grafana-cockroachdb-datasource",
    "grafana-databricks-datasource",
    "grafana-datadog-datasource",
    "grafana-dynatrace-datasource",
    "grafana-gitlab-datasource",
    "grafana-honeycomb-datasource",
    "grafana-jira-datasource",
    "grafana-looker-datasource",
    "grafana-mongodb-datasource",
    "grafana-newrelic-datasource",
    "grafana-odbc-datasource",
    "grafana-oracle-datasource",
    "grafana-pagerduty-datasource",
    "grafana-salesforce-datasource",
    "grafana-saphana-datasource",
    "grafana-servicenow-datasource",
    "grafana-snowflake-datasource",
    "grafana-splunk-datasource",
    "grafana-splunkobservability-datasource",
    "grafana-sumologic-datasource",
    "grafana-wavefront-datasource",
] as const;

export type DatasourcePluginType = (typeof DATASOURCE_PLUGIN_TYPES)[number];
