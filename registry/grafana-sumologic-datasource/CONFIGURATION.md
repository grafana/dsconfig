# Sumo Logic configuration

How to configure the **Sumo Logic** data source (`grafana-sumologic-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-sumologic-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [API Region](#api-region)
- [Authentication](#authentication)

## API Region

### API region / URL

_**required** · string_

SumoLogic API URL. [Read how to find your deployment.](https://help.sumologic.com/docs/api/getting-started/#which-endpoint-should-i-should-use).

| | |
|---|---|
| Default | `https://api.sumologic.com/api/` |

### Timeout

_optional · number_

Timeout in seconds for the data requests.

| | |
|---|---|
| Default | `30` |
| Range | at least 1 |

### Interval

_optional · number_

Interval in milliseconds for the log polling requests. Min value is 200.

| | |
|---|---|
| Default | `1000` |
| Range | at least 200 |

## Authentication

### AccessID

_conditionally required · string_

Sumo Logic Access Id.

| | |
|---|---|
| Example | `Sumo Logic Access Id` |
| Required when | **authMethod** is `accessKey` |

### AccessKey

_🔒 secret (write-only) · conditionally required · string_

Sumo Logic Access Key.

| | |
|---|---|
| Example | `Access key` |
| Required when | **authMethod** is `accessKey` |

## Other settings

### authMethod

_optional · string_

Authentication method discriminator. The only supported value is 'accessKey' (HTTP basic auth with an access ID + access key). The configuration editor never writes this key — it renders a single fixed authentication method whose selector handler is a no-op — and the backend defaults it to 'accessKey' when empty and rejects any other value.

| | |
|---|---|
| Default | `accessKey` |
| Allowed values | `accessKey` |

