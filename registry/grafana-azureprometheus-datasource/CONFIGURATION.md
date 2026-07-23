# Azure Monitor Managed Service for Prometheus configuration

Configuration reference for the **Azure Monitor Managed Service for Prometheus** data source (`grafana-azureprometheus-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/prometheus/configure-prometheus-data-source/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `url` | string | root | yes | Prometheus server URL |
| `jsonData.azureCredentials` | any | jsonData |  | Choose the type of authentication to Azure services |
| `secureJsonData.azureClientSecret` 🔒 | string | secureJsonData |  | Client secret of the App Registration. Written write-only by `@grafana/azure-sdk`'s `AzureCredentialsForm`; check `secureJsonFields.azureClientSecret` on the read side. Used when `jsonData.azureCredentials.authType` is `clientsecret` or `currentuser` with a `clientsecret` service-credentials fallback. |
| `secureJsonData.clientSecret` 🔒 | string | secureJsonData |  | Legacy client secret key. Preserved for backward compatibility with datasources provisioned before the credential migration to `azureClientSecret`. The backend reads it as a fallback via `grafana-azure-sdk-go/v2` `azcredentials/builder.go` `getFromCredentialsObject` when `secureJsonData.azureClientSecret` is missing. |
| `jsonData.oauthPassThru` | boolean | jsonData |  | Set to `true` by `@grafana/azure-sdk`'s `updateDatasourceCredentials` when the selected auth type is `currentuser`. `DataSourceHttpSettingsOverhaul.onAuthMethodSelect` (`src/configuration/DataSourceHttpSettingsOverhaul.tsx:121-131`) also always clears this to `false` on every save because the plugin's `visibleMethods=[azureAuthId]` locks the user to Azure auth. Consumed by the SDK's shared HTTP client and by `pkg/promlib` (`OauthPassThru` on PromOptions). |
| `jsonData.azureEndpointResourceId` | string | jsonData |  | Optional override of the Azure resource ID (audience) used to build the OAuth scope for Prometheus queries. Defined on `AzurePromDataSourceOptions` (`src/configuration/AzureCredentialsConfig.ts:68`) and cleared by `resetCredentials` (`AzureCredentialsConfig.ts:57-64`). Not rendered by the config editor — provisioning-only. If unset, the backend derives the scope from the resolved Azure cloud's `prometheusResourceId` property (`pkg/azureauth/azure.go:58-63`). |
| `jsonData.prometheus-type-migration` | boolean | jsonData |  | Sentinel flag set by the migration path when a vanilla Prometheus data source is migrated to Azure Monitor Managed Service for Prometheus. When true, `DataSourceHttpSettingsOverhaul.tsx:101-117` renders the 'Data source migrated' warning banner. Storage key uses a hyphen — the field ID is camelCased. Never rendered as an input; provisioning may set it to suppress or trigger the banner. |
| `jsonData.manageAlerts` | boolean | jsonData |  | Manage alert rules for this data source. To manage other alerting resources, add an Alertmanager data source. |
| `jsonData.allowAsRecordingRulesTarget` | boolean | jsonData |  | Allow this data source to be selected as a target for writing recording rules. |
| `jsonData.timeout` | number | jsonData |  | HTTP request timeout in seconds |
| `jsonData.httpHeaders` | list | jsonData |  | Additional HTTP headers sent with every request. Header names are stored in jsonData (httpHeaderName<N>); header values are write-only in secureJsonData (httpHeaderValue<N>). |
| `jsonData.httpHeaders[].name` | string | jsonData | yes | Header |
| `jsonData.httpHeaders[].value` | string | jsonData |  | Value |
| `jsonData.keepCookies` | list | jsonData |  | Grafana proxy deletes forwarded cookies by default. Specify cookies by name that should be forwarded to the data source. |
| `jsonData.timeInterval` | string | jsonData |  | This interval is how frequently Prometheus scrapes targets. Set this to the typical scrape and evaluation interval configured in your Prometheus config file. If you set this to a greater value than your Prometheus config file interval, Grafana will evaluate the data according to this interval and you will see less data points. Defaults to 15s. |
| `jsonData.queryTimeout` | string | jsonData |  | Set the Prometheus query timeout. |
| `jsonData.defaultEditor` | enum (builder, code) | jsonData |  | Set default editor option for all users of this data source. |
| `jsonData.disableMetricsLookup` | boolean | jsonData |  | Checking this option will disable the metrics chooser and metric/label support in the query field's autocomplete. This helps if you have performance issues with bigger Prometheus instances. |
| `jsonData.prometheusType` | enum (Prometheus, Cortex, Mimir, Thanos) | jsonData |  | Set this to the type of your prometheus database, e.g. Prometheus, Cortex, Mimir or Thanos. Changing this field will save your current settings. Certain types of Prometheus supports or does not support various APIs. For example, some types support regex matching for label queries to improve performance. Some types have an API for metadata. If you set this incorrectly you may experience odd behavior when querying metrics and labels. Please check your Prometheus documentation to ensure you enter the correct type. |
| `jsonData.prometheusVersion` | enum | jsonData |  | Use this to set the version of your Prometheus instance if it is not automatically configured. The option list depends on the selected Prometheus type (see PromFlavorVersions.ts in @grafana/prometheus). |
| `jsonData.cacheLevel` | enum (Low, Medium, High, None) | jsonData |  | Sets the browser caching level for editor queries. Higher cache settings are recommended for high cardinality data sources. |
| `jsonData.incrementalQuerying` | boolean | jsonData |  | This feature will change the default behavior of relative queries to always request fresh data from the prometheus instance, instead query results will be cached, and only new records are requested. Turn this on to decrease database and network load. |
| `jsonData.incrementalQueryOverlapWindow` | string | jsonData |  | Set a duration like 10m or 120s or 0s. Default of 10m. This duration will be added to the duration of each incremental request. |
| `jsonData.disableRecordingRules` | boolean | jsonData |  | This feature will disable recording rules. Turn this on to improve dashboard performance |
| `jsonData.customQueryParameters` | string | jsonData |  | Add custom parameters to the Prometheus query URL. For example timeout, partial_response, dedup, or max_source_resolution. Multiple parameters should be concatenated together with '&'. |
| `jsonData.httpMethod` | enum (POST, GET) | jsonData |  | You can use either POST or GET HTTP method to query your Prometheus data source. POST is the recommended method as it allows bigger queries. Change this to GET if you have a Prometheus version older than 2.1 or if POST requests are restricted in your network. |
| `jsonData.seriesLimit` | number | jsonData |  | The limit applies to all resources (metrics, labels, and values) for both endpoints (series and labels). Leave the field empty to use the default limit (40000). Set to 0 to disable the limit and fetch everything — this may cause performance issues. Default limit is 40000. |
| `jsonData.seriesEndpoint` | boolean | jsonData |  | Checking this option will favor the series endpoint with match[] parameter over the label values endpoint with match[] parameter. While the label values endpoint is considered more performant, some users may prefer the series because it has a POST method while the label values endpoint only has a GET method. |
| `jsonData.exemplarTraceIdDestinations` | list | jsonData |  | Exemplar trace ID destinations. For each configured destination, the plugin renders a link on exemplar labels — either to an internal Grafana tracing data source (datasourceUid) or an external URL. The label whose value carries the trace ID is 'name' (defaults to 'traceID' when new entries are added). |
| `jsonData.exemplarTraceIdDestinations[].name` | string | jsonData | yes | Label name |
| `jsonData.exemplarTraceIdDestinations[].url` | string | jsonData |  | The URL of the trace backend the user would go to see its trace |
| `jsonData.exemplarTraceIdDestinations[].urlDisplayLabel` | string | jsonData |  | Use to override the button label on the exemplar traceID field. |
| `jsonData.exemplarTraceIdDestinations[].datasourceUid` | string | jsonData |  | The tracing data source the exemplar link should navigate to. Setting this makes the exemplar an internal link and takes precedence over url. |
| `jsonData.maxSamplesProcessedWarningThreshold` | number | jsonData |  | When set, Grafana appends this value to Prometheus query requests as the max_samples_processed_warning_threshold URL parameter. Not exposed in the Azure Prometheus config editor (the PromSettings component only renders this input when its showQuerySamplesProcessedThresholdFields prop is true — this plugin never passes it), but the field is parsed by `pkg/promlib/models/settings.go:41` (backend-only). |
| `jsonData.maxSamplesProcessedErrorThreshold` | number | jsonData |  | When set, Grafana appends this value to Prometheus query requests as the max_samples_processed_error_threshold URL parameter. Not exposed in the Azure Prometheus config editor (feature-flagged off — see maxSamplesProcessedWarningThreshold). |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Azure Monitor Managed Service for Prometheus
    type: grafana-azureprometheus-datasource
    access: proxy
    url: "http://localhost:9090"
    jsonData:
      cacheLevel: Low
      defaultEditor: builder
      disableMetricsLookup: false
      disableRecordingRules: false
      httpMethod: POST
      incrementalQueryOverlapWindow: "10m"
      incrementalQuerying: false
      seriesEndpoint: false
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_azureprometheus_datasource" {
  type = "grafana-azureprometheus-datasource"
  name = "Azure Monitor Managed Service for Prometheus"
  url = "http://localhost:9090"

  json_data_encoded = jsonencode({
    cacheLevel = "Low"
    defaultEditor = "builder"
    disableMetricsLookup = false
    disableRecordingRules = false
    httpMethod = "POST"
    incrementalQueryOverlapWindow = "10m"
    incrementalQuerying = false
    seriesEndpoint = false
  })
}
```

