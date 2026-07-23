# Amazon Managed Service for Prometheus configuration

How to configure the **Amazon Managed Service for Prometheus** data source (`grafana-amazonprometheus-datasource`) in Grafana.

For more information, see the [official documentation](https://aws.amazon.com/prometheus/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [SigV4 Auth Details](#sigv4-auth-details)
- [Alerting](#alerting) — _optional_
- [Advanced HTTP settings](#advanced-http-settings) — _optional_
- [Interval behaviour](#interval-behaviour) — _optional_
- [Query editor](#query-editor) — _optional_
- [Performance](#performance) — _optional_
- [Other](#other) — _optional_
- [Exemplars](#exemplars) — _optional_
- [Migration](#migration) — _optional_

## Connection

### Prometheus server URL

_**required** · string_

| | |
|---|---|
| Example | `http://localhost:9090` |

## SigV4 Auth Details

### sigV4Auth

_optional · boolean_

SigV4 signing is mandatory for this plugin. `DataSourceHttpSettingsOverhaul.tsx:27-38` unconditionally sets `jsonData.sigV4Auth = true` on every editor mount and `visibleMethods=[sigV4Id]` (`:120`) hides every other authentication method.

| | |
|---|---|
| Default | `true` |

### Authentication Provider

_optional · select_

Specify which AWS credentials chain to use.

| | |
|---|---|
| Allowed values | `ec2_iam_role` (Workspace IAM Role), `grafana_assume_role` (Grafana Assume Role), `default` (AWS SDK Default), `keys` (Access & secret key), `credentials` (Credentials file) |

### Credentials Profile Name

_optional · string_

Credentials profile name, as specified in ~/.aws/credentials, leave blank for default.

| | |
|---|---|
| Example | `default` |
| Shown when | **Authentication Provider** is **Credentials file** (`credentials`) |

### Access Key ID

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Shown when | **Authentication Provider** is **Access & secret key** (`keys`) |

### Secret Access Key

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Shown when | **Authentication Provider** is **Access & secret key** (`keys`) |

### Assume Role ARN

_optional · string_

Optional. Specifying the ARN of a role will ensure that the selected authentication provider is used to assume the role rather than the credentials directly.

| | |
|---|---|
| Example | `arn:aws:iam:*` |
| Must match | `^(arn:aws[a-zA-Z-]*:iam::[0-9]{12}:role/.+)?$` |

### External ID

_optional · string_

If you are assuming a role in another account, that has been created with an external ID, specify the external ID here.

| | |
|---|---|
| Example | `External ID` |
| Shown when | `jsonData_sigV4AuthType != 'grafana_assume_role'` |

### Default Region

_optional · select_

Specify the region, such as for US West (Oregon) use ` us-west-2 ` as the region.

| | |
|---|---|
| Example | `Choose` |

### Service

_optional · string_

Specify the AWS service to sign requests against (e.g., 'aps' for Prometheus).

| | |
|---|---|
| Default | `aps` |
| Example | `aps` |

### Forward Grafana User HTTP Header

_optional · toggle_

Forward the logged-in Grafana user's X-Grafana-User header to the workspace. Requires send_user_header to be enabled in the Grafana server configuration.

| | |
|---|---|
| Default | `false` |

### oauthPassThru

_optional · boolean_

Set to `false` by `DataSourceHttpSettingsOverhaul.onAuthMethodSelect` (`:103-115`) on every save because `visibleMethods=[sigV4Id]` locks the editor to SigV4 auth. Consumed by the SDK's shared HTTP client.

| | |
|---|---|
| Default | `false` |

## Alerting

_This section is optional._

### Manage alerts via Alerting UI

_optional · toggle_

Manage alert rules for this data source. To manage other alerting resources, add an Alertmanager data source.

### Allow as recording rules target

_optional · toggle_

Allow this data source to be selected as a target for writing recording rules.

## Advanced HTTP settings

_This section is optional._

### Timeout

_optional · number_

HTTP request timeout in seconds.

| | |
|---|---|
| Example | `Timeout in seconds` |

### Custom HTTP Headers

_optional · list_

Additional HTTP headers sent with every request. Header names are stored in jsonData (httpHeaderName<N>); header values are write-only in secureJsonData (httpHeaderValue<N>).

Each item has the following fields:

#### Header

_**required** · string_

| | |
|---|---|
| Example | `X-Custom-Header` |
| Must match | `^[A-Za-z][A-Za-z0-9-]*$` |

#### Value

_optional · string_

| | |
|---|---|
| Example | `Header Value` |

### Allowed cookies

_optional · list_

Grafana proxy deletes forwarded cookies by default. Specify cookies by name that should be forwarded to the data source.

| | |
|---|---|
| Example | `New cookie (hit enter to add)` |

## Interval behaviour

_This section is optional._

### Scrape interval

_optional · string_

This interval is how frequently Prometheus scrapes targets. Set this to the typical scrape and evaluation interval configured in your Prometheus config file. If you set this to a greater value than your Prometheus config file interval, Grafana will evaluate the data according to this interval and you will see less data points. Defaults to 15s.

| | |
|---|---|
| Example | `15s` |

### Query timeout

_optional · string_

Set the Prometheus query timeout.

| | |
|---|---|
| Example | `60s` |

## Query editor

_This section is optional._

### Default editor

_optional · select_

Set default editor option for all users of this data source.

| | |
|---|---|
| Default | `builder` |
| Allowed values | `builder` (Builder), `code` (Code) |

### Disable metrics lookup

_optional · toggle_

Checking this option will disable the metrics chooser and metric/label support in the query field's autocomplete. This helps if you have performance issues with bigger Prometheus instances.

| | |
|---|---|
| Default | `false` |

## Performance

_This section is optional._

### prometheusType

_optional · string_

The Prometheus flavor. Not rendered by this plugin's editor (Amazon Prometheus passes `hidePrometheusTypeVersion={true}` to `PromSettings`, `ConfigEditor.tsx:100`). Provisioning can still set it — the backend `promlib` heuristics use it to enable flavor-specific query paths.

| | |
|---|---|
| Allowed values | `Prometheus`, `Cortex`, `Mimir`, `Thanos` |

### prometheusVersion

_optional · string_

Free-form Prometheus version string. Editor-hidden alongside `prometheusType` (`ConfigEditor.tsx:100`). Provisioning-only.

### Cache level

_optional · select_

Sets the browser caching level for editor queries. Higher cache settings are recommended for high cardinality data sources.

| | |
|---|---|
| Default | `Low` |
| Allowed values | `Low`, `Medium`, `High`, `None` |

### Incremental querying (beta)

_optional · toggle_

This feature will change the default behavior of relative queries to always request fresh data from the prometheus instance, instead query results will be cached, and only new records are requested. Turn this on to decrease database and network load.

| | |
|---|---|
| Default | `false` |

### Query overlap window

_optional · string_

Set a duration like 10m or 120s or 0s. Default of 10m. This duration will be added to the duration of each incremental request.

| | |
|---|---|
| Default | `10m` |
| Example | `10m` |
| Shown when | **Incremental querying (beta)** is `true` |

### Disable recording rules (beta)

_optional · toggle_

This feature will disable recording rules. Turn this on to improve dashboard performance.

| | |
|---|---|
| Default | `false` |

## Other

_This section is optional._

### Custom query parameters

_optional · string_

Add custom parameters to the Prometheus query URL. For example timeout, partial_response, dedup, or max_source_resolution. Multiple parameters should be concatenated together with '&'.

| | |
|---|---|
| Example | `Example: max_source_resolution=5m&timeout=10` |

### HTTP method

_optional · select_

You can use either POST or GET HTTP method to query your Prometheus data source. POST is the recommended method as it allows bigger queries. Change this to GET if you have a Prometheus version older than 2.1 or if POST requests are restricted in your network.

| | |
|---|---|
| Default | `POST` |
| Allowed values | `POST`, `GET` |

### Series limit

_optional · number_

The limit applies to all resources (metrics, labels, and values) for both endpoints (series and labels). Leave the field empty to use the default limit (40000). Set to 0 to disable the limit and fetch everything — this may cause performance issues. Default limit is 40000.

| | |
|---|---|
| Example | `40000` |

### Query warning threshold

_optional · number_

When set, Grafana appends this value to Prometheus query requests as the max_samples_processed_warning_threshold URL parameter. Leave empty to omit.

| | |
|---|---|
| Example | `Example: 100000000` |

### Query error threshold

_optional · number_

When set, Grafana appends this value to Prometheus query requests as the max_samples_processed_error_threshold URL parameter. Leave empty to omit.

| | |
|---|---|
| Example | `Example: 200000000` |

### Use series endpoint

_optional · toggle_

Checking this option will favor the series endpoint with match[] parameter over the label values endpoint with match[] parameter. While the label values endpoint is considered more performant, some users may prefer the series because it has a POST method while the label values endpoint only has a GET method.

| | |
|---|---|
| Default | `false` |

## Exemplars

_This section is optional._

### exemplarTraceIdDestinations

_optional · list_

Exemplar trace ID destinations. Not rendered by this plugin's editor (Amazon Prometheus passes `hideExemplars={true}` to `PromSettings`, `ConfigEditor.tsx:101`). Provisioning-only — the backend still parses and returns exemplar links from `promlib` query results.

Each item has the following fields:

#### Label name

_**required** · string_

| | |
|---|---|
| Example | `traceID` |

#### URL

_optional · string_

The URL of the trace backend the user would go to see its trace.

| | |
|---|---|
| Example | `https://example.com/${__value.raw}` |

#### URL Label

_optional · string_

Use to override the button label on the exemplar traceID field.

| | |
|---|---|
| Example | `Go to example.com` |

#### Data source

_optional · string_

The tracing data source the exemplar link should navigate to. Setting this makes the exemplar an internal link and takes precedence over url.

## Migration

_This section is optional._

### prometheus-type-migration

_optional · boolean_

Sentinel flag set when a vanilla Prometheus data source is migrated to Amazon Managed Service for Prometheus. When true, `ConfigEditor.tsx:37-48` renders the 'Data source migrated' warning banner. Storage key uses a hyphen — the field ID is camelCased. Never rendered as an input; provisioning may set it to trigger the banner.

