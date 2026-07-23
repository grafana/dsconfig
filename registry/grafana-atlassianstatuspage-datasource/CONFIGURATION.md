# Atlassian Statuspage configuration

How to configure the **Atlassian Statuspage** data source (`grafana-atlassianstatuspage-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-atlassianstatuspage-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)

## Connection

Provide information to connect to the data source

### URL

_**required** · string_

The URL of the Atlassian Statuspage, including `https://` and without trailing `/`.

