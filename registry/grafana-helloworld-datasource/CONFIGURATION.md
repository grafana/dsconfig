# Hello World configuration

How to configure the **Hello World** data source (`grafana-helloworld-datasource`) in Grafana.

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Other settings

### API key

_🔒 secret (write-only) · optional · string_

Placeholder secret. The Hello World datasource reads no configuration: its config editor renders static text and its backend ignores instance settings. This key exists only because a dsconfig entry must declare at least one field and the shared conformance suite requires at least one secureJsonData key. The plugin never reads it.

