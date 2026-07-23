# Amazon Managed Service for Prometheus configuration

Configuration reference for the **Amazon Managed Service for Prometheus** data source (`grafana-amazonprometheus-datasource`) in Grafana.

For more information, see the [official documentation](https://aws.amazon.com/prometheus/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `url` | string | root | yes | Prometheus server URL |
| `jsonData.sigV4Auth` | boolean | jsonData |  | SigV4 signing is mandatory for this plugin. `DataSourceHttpSettingsOverhaul.tsx:27-38` unconditionally sets `jsonData.sigV4Auth = true` on every editor mount and `visibleMethods=[sigV4Id]` (`:120`) hides every other authentication method. |
| `jsonData.sigV4AuthType` | enum (ec2_iam_role, grafana_assume_role, default, keys, credentials) | jsonData |  | Specify which AWS credentials chain to use. |
| `jsonData.sigV4Profile` | string | jsonData |  | Credentials profile name, as specified in ~/.aws/credentials, leave blank for default. |
| `secureJsonData.sigV4AccessKey` 🔒 | string | secureJsonData | conditional | Access Key ID |
| `secureJsonData.sigV4SecretKey` 🔒 | string | secureJsonData | conditional | Secret Access Key |
| `jsonData.sigV4AssumeRoleArn` | string | jsonData |  | Optional. Specifying the ARN of a role will ensure that the selected authentication provider is used to assume the role rather than the credentials directly. |
| `jsonData.sigV4ExternalId` | string | jsonData |  | If you are assuming a role in another account, that has been created with an external ID, specify the external ID here. |
| `jsonData.sigV4Region` | enum | jsonData |  | Specify the region, such as for US West (Oregon) use ` us-west-2 ` as the region. |
| `jsonData.sigv4Service` | string | jsonData |  | Specify the AWS service to sign requests against (e.g., 'aps' for Prometheus). |
| `jsonData.forwardGrafanaUserHeader` | boolean | jsonData |  | Forward the logged-in Grafana user's X-Grafana-User header to the workspace. Requires send_user_header to be enabled in the Grafana server configuration. |
| `jsonData.oauthPassThru` | boolean | jsonData |  | Set to `false` by `DataSourceHttpSettingsOverhaul.onAuthMethodSelect` (`:103-115`) on every save because `visibleMethods=[sigV4Id]` locks the editor to SigV4 auth. Consumed by the SDK's shared HTTP client. |
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
| `jsonData.prometheusType` | enum (Prometheus, Cortex, Mimir, Thanos) | jsonData |  | The Prometheus flavor. Not rendered by this plugin's editor (Amazon Prometheus passes `hidePrometheusTypeVersion={true}` to `PromSettings`, `ConfigEditor.tsx:100`). Provisioning can still set it — the backend `promlib` heuristics use it to enable flavor-specific query paths. |
| `jsonData.prometheusVersion` | string | jsonData |  | Free-form Prometheus version string. Editor-hidden alongside `prometheusType` (`ConfigEditor.tsx:100`). Provisioning-only. |
| `jsonData.cacheLevel` | enum (Low, Medium, High, None) | jsonData |  | Sets the browser caching level for editor queries. Higher cache settings are recommended for high cardinality data sources. |
| `jsonData.incrementalQuerying` | boolean | jsonData |  | This feature will change the default behavior of relative queries to always request fresh data from the prometheus instance, instead query results will be cached, and only new records are requested. Turn this on to decrease database and network load. |
| `jsonData.incrementalQueryOverlapWindow` | string | jsonData |  | Set a duration like 10m or 120s or 0s. Default of 10m. This duration will be added to the duration of each incremental request. |
| `jsonData.disableRecordingRules` | boolean | jsonData |  | This feature will disable recording rules. Turn this on to improve dashboard performance |
| `jsonData.customQueryParameters` | string | jsonData |  | Add custom parameters to the Prometheus query URL. For example timeout, partial_response, dedup, or max_source_resolution. Multiple parameters should be concatenated together with '&'. |
| `jsonData.httpMethod` | enum (POST, GET) | jsonData |  | You can use either POST or GET HTTP method to query your Prometheus data source. POST is the recommended method as it allows bigger queries. Change this to GET if you have a Prometheus version older than 2.1 or if POST requests are restricted in your network. |
| `jsonData.seriesLimit` | number | jsonData |  | The limit applies to all resources (metrics, labels, and values) for both endpoints (series and labels). Leave the field empty to use the default limit (40000). Set to 0 to disable the limit and fetch everything — this may cause performance issues. Default limit is 40000. |
| `jsonData.maxSamplesProcessedWarningThreshold` | number | jsonData |  | When set, Grafana appends this value to Prometheus query requests as the max_samples_processed_warning_threshold URL parameter. Leave empty to omit. |
| `jsonData.maxSamplesProcessedErrorThreshold` | number | jsonData |  | When set, Grafana appends this value to Prometheus query requests as the max_samples_processed_error_threshold URL parameter. Leave empty to omit. |
| `jsonData.seriesEndpoint` | boolean | jsonData |  | Checking this option will favor the series endpoint with match[] parameter over the label values endpoint with match[] parameter. While the label values endpoint is considered more performant, some users may prefer the series because it has a POST method while the label values endpoint only has a GET method. |
| `jsonData.exemplarTraceIdDestinations` | list | jsonData |  | Exemplar trace ID destinations. Not rendered by this plugin's editor (Amazon Prometheus passes `hideExemplars={true}` to `PromSettings`, `ConfigEditor.tsx:101`). Provisioning-only — the backend still parses and returns exemplar links from `promlib` query results. |
| `jsonData.exemplarTraceIdDestinations[].name` | string | jsonData | yes | Label name |
| `jsonData.exemplarTraceIdDestinations[].url` | string | jsonData |  | The URL of the trace backend the user would go to see its trace |
| `jsonData.exemplarTraceIdDestinations[].urlDisplayLabel` | string | jsonData |  | Use to override the button label on the exemplar traceID field. |
| `jsonData.exemplarTraceIdDestinations[].datasourceUid` | string | jsonData |  | The tracing data source the exemplar link should navigate to. Setting this makes the exemplar an internal link and takes precedence over url. |
| `jsonData.prometheus-type-migration` | boolean | jsonData |  | Sentinel flag set when a vanilla Prometheus data source is migrated to Amazon Managed Service for Prometheus. When true, `ConfigEditor.tsx:37-48` renders the 'Data source migrated' warning banner. Storage key uses a hyphen — the field ID is camelCased. Never rendered as an input; provisioning may set it to trigger the banner. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Workspace IAM Role (`ec2_iam_role`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Managed Service for Prometheus
    type: grafana-amazonprometheus-datasource
    access: proxy
    url: "http://localhost:9090"
    jsonData:
      cacheLevel: Low
      defaultEditor: builder
      disableMetricsLookup: false
      disableRecordingRules: false
      forwardGrafanaUserHeader: false
      httpMethod: POST
      incrementalQueryOverlapWindow: "10m"
      incrementalQuerying: false
      oauthPassThru: false
      seriesEndpoint: false
      sigV4Auth: true
      sigV4AuthType: ec2_iam_role
      sigv4Service: aps
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_amazonprometheus_datasource_ec2_iam_role" {
  type = "grafana-amazonprometheus-datasource"
  name = "Amazon Managed Service for Prometheus"
  url = "http://localhost:9090"

  json_data_encoded = jsonencode({
    cacheLevel = "Low"
    defaultEditor = "builder"
    disableMetricsLookup = false
    disableRecordingRules = false
    forwardGrafanaUserHeader = false
    httpMethod = "POST"
    incrementalQueryOverlapWindow = "10m"
    incrementalQuerying = false
    oauthPassThru = false
    seriesEndpoint = false
    sigV4Auth = true
    sigV4AuthType = "ec2_iam_role"
    sigv4Service = "aps"
  })
}
```

### Grafana Assume Role (`grafana_assume_role`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Managed Service for Prometheus
    type: grafana-amazonprometheus-datasource
    access: proxy
    url: "http://localhost:9090"
    jsonData:
      cacheLevel: Low
      defaultEditor: builder
      disableMetricsLookup: false
      disableRecordingRules: false
      forwardGrafanaUserHeader: false
      httpMethod: POST
      incrementalQueryOverlapWindow: "10m"
      incrementalQuerying: false
      oauthPassThru: false
      seriesEndpoint: false
      sigV4Auth: true
      sigV4AuthType: grafana_assume_role
      sigv4Service: aps
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_amazonprometheus_datasource_grafana_assume_role" {
  type = "grafana-amazonprometheus-datasource"
  name = "Amazon Managed Service for Prometheus"
  url = "http://localhost:9090"

  json_data_encoded = jsonencode({
    cacheLevel = "Low"
    defaultEditor = "builder"
    disableMetricsLookup = false
    disableRecordingRules = false
    forwardGrafanaUserHeader = false
    httpMethod = "POST"
    incrementalQueryOverlapWindow = "10m"
    incrementalQuerying = false
    oauthPassThru = false
    seriesEndpoint = false
    sigV4Auth = true
    sigV4AuthType = "grafana_assume_role"
    sigv4Service = "aps"
  })
}
```

### AWS SDK Default (`default`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Managed Service for Prometheus
    type: grafana-amazonprometheus-datasource
    access: proxy
    url: "http://localhost:9090"
    jsonData:
      cacheLevel: Low
      defaultEditor: builder
      disableMetricsLookup: false
      disableRecordingRules: false
      forwardGrafanaUserHeader: false
      httpMethod: POST
      incrementalQueryOverlapWindow: "10m"
      incrementalQuerying: false
      oauthPassThru: false
      seriesEndpoint: false
      sigV4Auth: true
      sigV4AuthType: default
      sigv4Service: aps
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_amazonprometheus_datasource_default" {
  type = "grafana-amazonprometheus-datasource"
  name = "Amazon Managed Service for Prometheus"
  url = "http://localhost:9090"

  json_data_encoded = jsonencode({
    cacheLevel = "Low"
    defaultEditor = "builder"
    disableMetricsLookup = false
    disableRecordingRules = false
    forwardGrafanaUserHeader = false
    httpMethod = "POST"
    incrementalQueryOverlapWindow = "10m"
    incrementalQuerying = false
    oauthPassThru = false
    seriesEndpoint = false
    sigV4Auth = true
    sigV4AuthType = "default"
    sigv4Service = "aps"
  })
}
```

### Access & secret key (`keys`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Managed Service for Prometheus
    type: grafana-amazonprometheus-datasource
    access: proxy
    url: "http://localhost:9090"
    jsonData:
      cacheLevel: Low
      defaultEditor: builder
      disableMetricsLookup: false
      disableRecordingRules: false
      forwardGrafanaUserHeader: false
      httpMethod: POST
      incrementalQueryOverlapWindow: "10m"
      incrementalQuerying: false
      oauthPassThru: false
      seriesEndpoint: false
      sigV4Auth: true
      sigV4AuthType: keys
      sigv4Service: aps
    secureJsonData:
      sigV4AccessKey: "<YOUR_ACCESS_KEY_ID>"
      sigV4SecretKey: "<YOUR_SECRET_ACCESS_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_amazonprometheus_datasource_keys" {
  type = "grafana-amazonprometheus-datasource"
  name = "Amazon Managed Service for Prometheus"
  url = "http://localhost:9090"

  json_data_encoded = jsonencode({
    cacheLevel = "Low"
    defaultEditor = "builder"
    disableMetricsLookup = false
    disableRecordingRules = false
    forwardGrafanaUserHeader = false
    httpMethod = "POST"
    incrementalQueryOverlapWindow = "10m"
    incrementalQuerying = false
    oauthPassThru = false
    seriesEndpoint = false
    sigV4Auth = true
    sigV4AuthType = "keys"
    sigv4Service = "aps"
  })

  secure_json_data_encoded = jsonencode({
    sigV4AccessKey = "<YOUR_ACCESS_KEY_ID>"
    sigV4SecretKey = "<YOUR_SECRET_ACCESS_KEY>"
  })
}
```

### Credentials file (`credentials`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Managed Service for Prometheus
    type: grafana-amazonprometheus-datasource
    access: proxy
    url: "http://localhost:9090"
    jsonData:
      cacheLevel: Low
      defaultEditor: builder
      disableMetricsLookup: false
      disableRecordingRules: false
      forwardGrafanaUserHeader: false
      httpMethod: POST
      incrementalQueryOverlapWindow: "10m"
      incrementalQuerying: false
      oauthPassThru: false
      seriesEndpoint: false
      sigV4Auth: true
      sigV4AuthType: credentials
      sigv4Service: aps
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_amazonprometheus_datasource_credentials" {
  type = "grafana-amazonprometheus-datasource"
  name = "Amazon Managed Service for Prometheus"
  url = "http://localhost:9090"

  json_data_encoded = jsonencode({
    cacheLevel = "Low"
    defaultEditor = "builder"
    disableMetricsLookup = false
    disableRecordingRules = false
    forwardGrafanaUserHeader = false
    httpMethod = "POST"
    incrementalQueryOverlapWindow = "10m"
    incrementalQuerying = false
    oauthPassThru = false
    seriesEndpoint = false
    sigV4Auth = true
    sigV4AuthType = "credentials"
    sigv4Service = "aps"
  })
}
```

