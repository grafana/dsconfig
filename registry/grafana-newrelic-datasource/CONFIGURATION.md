# New Relic configuration

How to configure the **New Relic** data source (`grafana-newrelic-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-newrelic-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [New Relic API Credentials](#new-relic-api-credentials)

## New Relic API Credentials

### Personal API Key / User API key

_🔒 secret (write-only) · **required** · string_

Used for NRQL queries.

| | |
|---|---|
| Example | `Personal API Key` |

### Account ID

_🔒 secret (write-only) · **required** · string_

Your New Relic Account ID.

| | |
|---|---|
| Example | `Account ID` |

### Region

_optional · select_

Region hosting your service.

| | |
|---|---|
| Example | `default` |
| Allowed values | `EU`, `US` |

### Timeout in Seconds

_optional · number_

Enter the timeout in seconds. Defaults to 300.

| | |
|---|---|
| Default | `300` |
| Example | `300` |

## Other settings

### restBaseURL

_optional · string_

Backend-only override for the New Relic REST API base URL. Used for internal testing and mocking; not exposed in the configuration editor.

### infrastructureBaseURL

_optional · string_

Backend-only override for the New Relic Infrastructure API base URL. Used for internal testing and mocking; not exposed in the configuration editor.

### nerdGraphBaseURL

_optional · string_

Backend-only override for the New Relic NerdGraph (GraphQL) API base URL. Used for internal testing and mocking; not exposed in the configuration editor.

