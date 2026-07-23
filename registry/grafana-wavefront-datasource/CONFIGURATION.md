# Wavefront configuration

How to configure the **Wavefront** data source (`grafana-wavefront-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-wavefront-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Wavefront settings](#wavefront-settings)
- [Customization](#customization) — _optional_

## Wavefront settings

### API URL

_**required** · string_

URL to Wavefront API.

| | |
|---|---|
| Default | `https://try.wavefront.com` |
| Example | `https://try.wavefront.com` |

### Token

_🔒 secret (write-only) · **required** · string_

Wavefront token.

| | |
|---|---|
| Example | `Wavefront token` |

## Customization

_This section is optional._

### Request timeout in seconds

_optional · number_

Request timeout in seconds. Defaults to 30.

| | |
|---|---|
| Default | `30` |
| Example | `30` |

