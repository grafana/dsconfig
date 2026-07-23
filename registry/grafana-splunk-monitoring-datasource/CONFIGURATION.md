# Splunk Infrastructure Monitoring configuration

How to configure the **Splunk Infrastructure Monitoring** data source (`grafana-splunk-monitoring-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-splunk-monitoring-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Authentication](#authentication)
- [Custom URLs](#custom-urls) — _optional_

## Authentication

### Access Token

_🔒 secret (write-only) · **required** · string_

### Realm Name

_conditionally required · string_

| | |
|---|---|
| Example | `us1` |
| Required when | `jsonData_urlMetricsMetadata == '' || jsonData_urlSignalflow == ''` |

## Custom URLs

Optional URLs. Use this section only if you are using custom signalflow domains. Leave it blank for the default behavior

_This section is optional._

### Metrics MetaData URL

_optional · string_

Optional Metrics MetaData URL.

| | |
|---|---|
| Example | `https://api.us1.signalfx.com` |

### SignalFlow URL

_optional · string_

Optional SignalFlow URL.

| | |
|---|---|
| Example | `https://stream.us1.signalfx.com` |

