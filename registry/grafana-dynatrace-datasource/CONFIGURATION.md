# Dynatrace configuration

How to configure the **Dynatrace** data source (`grafana-dynatrace-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-dynatrace-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)
- [Additional settings](#additional-settings) — _optional_

## Connection

### Dynatrace API Type

_optional · radio_

| | |
|---|---|
| Default | `saas` |
| Allowed values | `saas` (SaaS Environment), `managed` (Managed Cluster), `url` (Raw URL) |

### Environment ID

_**required** · string_

Get environment ID from your instance URL: [environmentId].live.dynatrace.com.

| | |
|---|---|
| Example | `Your Environment ID` |

### Domain

_conditionally required · string_

| | |
|---|---|
| Example | `Your Domain` |
| Shown when | **Dynatrace API Type** is **Managed Cluster** (`managed`) |

## Authentication

### Dynatrace API Token

_🔒 secret (write-only) · conditionally required · string_

The API token for the Dynatrace API. This is required for Older api endpoints on Dynatrace like Metrics, Problems, Logs, etc.

| | |
|---|---|
| Example | `Your API Token` |
| Required when | **Dynatrace Platform Token** is `` |

### Dynatrace Platform Token

_🔒 secret (write-only) · conditionally required · string_

The Platform token for the Dynatrace Platform API. This is required for Newer api endpoints on Dynatrace like Grail.

| | |
|---|---|
| Example | `Your Platform Token` |
| Required when | **Dynatrace API Token** is `` |

## Additional settings

_This section is optional._

### Timeout

_optional · number_

The timeout for the HTTP client in seconds. Default is 30 seconds.

| | |
|---|---|
| Default | `30` |
| Example | `30` |
| Range | at least 0 |

### Skip TLS Verify

_optional · toggle_

| | |
|---|---|
| Default | `false` |

### With CA Cert

_optional · toggle_

Needed for verifying self-signed TLS Certificates.

| | |
|---|---|
| Default | `false` |

### CA Cert

_🔒 secret (write-only) · conditionally required · multiline text_

TLS/SSL Certs are encrypted and stored in the Grafana database.

| | |
|---|---|
| Example | `Begins with -----BEGIN CERTIFICATE-----` |
| Shown when | **With CA Cert** is `true` |

