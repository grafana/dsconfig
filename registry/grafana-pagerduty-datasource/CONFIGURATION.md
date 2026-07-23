# PagerDuty configuration

How to configure the **PagerDuty** data source (`grafana-pagerduty-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-pagerduty-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Authentication](#authentication)

## Authentication

### API key

_🔒 secret (write-only) · conditionally required · string_

PagerDuty REST API Key (prefer generating read-only key).

| | |
|---|---|
| Required when | **id** is `api_key` |

