# Azure Monitor Managed Service for Prometheus configuration

How to configure the **Azure Monitor Managed Service for Prometheus** data source (`grafana-azureprometheus-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/prometheus/configure-prometheus-data-source/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)
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

## Authentication

### Authentication

_optional_

Choose the type of authentication to Azure services.

Discriminated-union object written by the `@grafana/azure-sdk` `AzureCredentialsForm` (`src/configuration/AzureCredentialsForm.tsx`). Shape depends on `authType`:

- `clientsecret` — `{ authType: 'clientsecret', azureCloud, tenantId, clientId }` (secret in `secureJsonData.azureClientSecret`). Rendered as **App Registration** — always available.
- `msi` — `{ authType: 'msi' }`. Rendered as **Managed Identity** — only when Grafana has `azure.managedIdentityEnabled`.
- `workloadidentity` — `{ authType: 'workloadidentity' }`. Rendered as **Workload Identity** — only when `azure.workloadIdentityEnabled`.
- `currentuser` — `{ authType: 'currentuser', serviceCredentialsEnabled?: boolean, serviceCredentials?: { authType: 'msi' | 'workloadidentity' | 'clientsecret', ... } }`. Rendered as **Current User** — only when `azure.userIdentityEnabled`.

Only `App Registration (clientsecret)` exposes the tenant/client/secret inputs directly in this plugin's editor; the other authTypes have no additional sub-fields. See `src/configuration/AzureCredentialsForm.tsx:34-66` for the option list and `src/configuration/AppRegistrationCredentials.tsx:57-150` for the App Registration inputs. Backend parse is delegated to `github.com/grafana/grafana-azure-sdk-go/v2/azcredentials.FromDatasourceData` (invoked at `pkg/azureauth/azure.go:23`).

### Client Secret

_🔒 secret (write-only) · optional · string_

Client secret of the App Registration. Written write-only by `@grafana/azure-sdk`'s `AzureCredentialsForm`; check `secureJsonFields.azureClientSecret` on the read side. Used when `jsonData.azureCredentials.authType` is `clientsecret` or `currentuser` with a `clientsecret` service-credentials fallback.

| | |
|---|---|
| Example | `XXXXXXXX-XXXX-XXXX-XXXX-XXXXXXXXXXXX` |

### clientSecret

_🔒 secret (write-only) · optional · string_

Legacy client secret key. Preserved for backward compatibility with datasources provisioned before the credential migration to `azureClientSecret`. The backend reads it as a fallback via `grafana-azure-sdk-go/v2` `azcredentials/builder.go` `getFromCredentialsObject` when `secureJsonData.azureClientSecret` is missing.

### oauthPassThru

_optional · boolean_

Set to `true` by `@grafana/azure-sdk`'s `updateDatasourceCredentials` when the selected auth type is `currentuser`. `DataSourceHttpSettingsOverhaul.onAuthMethodSelect` (`src/configuration/DataSourceHttpSettingsOverhaul.tsx:121-131`) also always clears this to `false` on every save because the plugin's `visibleMethods=[azureAuthId]` locks the user to Azure auth. Consumed by the SDK's shared HTTP client and by `pkg/promlib` (`OauthPassThru` on PromOptions).

### azureEndpointResourceId

_optional · string_

Optional override of the Azure resource ID (audience) used to build the OAuth scope for Prometheus queries. Defined on `AzurePromDataSourceOptions` (`src/configuration/AzureCredentialsConfig.ts:68`) and cleared by `resetCredentials` (`AzureCredentialsConfig.ts:57-64`). Not rendered by the config editor — provisioning-only. If unset, the backend derives the scope from the resolved Azure cloud's `prometheusResourceId` property (`pkg/azureauth/azure.go:58-63`).

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

### Prometheus type

_optional · select_

Set this to the type of your prometheus database, e.g. Prometheus, Cortex, Mimir or Thanos. Changing this field will save your current settings. Certain types of Prometheus supports or does not support various APIs. For example, some types support regex matching for label queries to improve performance. Some types have an API for metadata. If you set this incorrectly you may experience odd behavior when querying metrics and labels. Please check your Prometheus documentation to ensure you enter the correct type.

| | |
|---|---|
| Allowed values | `Prometheus`, `Cortex`, `Mimir`, `Thanos` |

### Version

_optional · select_

Use this to set the version of your Prometheus instance if it is not automatically configured. The option list depends on the selected Prometheus type (see PromFlavorVersions.ts in @grafana/prometheus).

| | |
|---|---|
| Shown when | `jsonData_prometheusType != ''` |

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

### Use series endpoint

_optional · toggle_

Checking this option will favor the series endpoint with match[] parameter over the label values endpoint with match[] parameter. While the label values endpoint is considered more performant, some users may prefer the series because it has a POST method while the label values endpoint only has a GET method.

| | |
|---|---|
| Default | `false` |

## Exemplars

_This section is optional._

### Exemplars

_optional · list_

Exemplar trace ID destinations. For each configured destination, the plugin renders a link on exemplar labels — either to an internal Grafana tracing data source (datasourceUid) or an external URL. The label whose value carries the trace ID is 'name' (defaults to 'traceID' when new entries are added).

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

Sentinel flag set by the migration path when a vanilla Prometheus data source is migrated to Azure Monitor Managed Service for Prometheus. When true, `DataSourceHttpSettingsOverhaul.tsx:101-117` renders the 'Data source migrated' warning banner. Storage key uses a hyphen — the field ID is camelCased. Never rendered as an input; provisioning may set it to suppress or trigger the banner.

## Other settings

### maxSamplesProcessedWarningThreshold

_optional · number_

When set, Grafana appends this value to Prometheus query requests as the max_samples_processed_warning_threshold URL parameter. Not exposed in the Azure Prometheus config editor (the PromSettings component only renders this input when its showQuerySamplesProcessedThresholdFields prop is true — this plugin never passes it), but the field is parsed by `pkg/promlib/models/settings.go:41` (backend-only).

### maxSamplesProcessedErrorThreshold

_optional · number_

When set, Grafana appends this value to Prometheus query requests as the max_samples_processed_error_threshold URL parameter. Not exposed in the Azure Prometheus config editor (feature-flagged off — see maxSamplesProcessedWarningThreshold).

