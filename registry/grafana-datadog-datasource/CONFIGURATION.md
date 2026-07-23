# Datadog configuration

How to configure the **Datadog** data source (`grafana-datadog-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-datadog-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)
- [Additional settings](#additional-settings) — _optional_

## Connection

### Mode

_optional · radio_

Choose **Hosted Datadog Metrics**, if you want to connect using a Datadog proxy through [Hosted Datadog Metrics](https://grafana.com/docs/grafana-cloud/data-configuration/metrics/metrics-datadog/). Otherwise choose **Default** to directly connect to DataDog API endpoints.

| | |
|---|---|
| Default | `default` |
| Allowed values | `default` (Default), `hosted-metrics` (Hosted Datadog Metrics) |

### API URL / Region

_**required** · string_

A URL to the Datadog API (e.g.: https://api.datadoghq.com).

| | |
|---|---|
| Default | `https://api.datadoghq.com` |
| Example | `https://api.datadoghq.com` |

## Authentication

### API key

_🔒 secret (write-only) · conditionally required · string_

An API key is unique to your organization. [Learn more](https://grafana.com/docs/plugins/grafana-datadog-datasource/latest/#get-an-api-key-and-application-key-from-datadog).

| | |
|---|---|
| Example | `Datadog API key` |
| Shown when | **Mode** is **Default** (`default`) |
| Required when | `jsonData_pluginMode != 'hosted-metrics'` |

### App key

_🔒 secret (write-only) · conditionally required · string_

An application key is used with the API key to give access to the Datadog API. By default, application keys have the permissions of the user who created them. [Learn more](https://grafana.com/docs/plugins/grafana-datadog-datasource/latest/#get-api-key-and-application-key-from-datadog). You can also customize the scope of the application key in the [Datadog docs](https://docs.datadoghq.com/api/latest/scopes/).

| | |
|---|---|
| Example | `Datadog App key` |
| Shown when | **Mode** is **Default** (`default`) |
| Required when | `jsonData_pluginMode != 'hosted-metrics'` |

### User

_conditionally required · string_

Your username is your Grafana Cloud Prometheus username. This can be found in the Prometheus details in your cloud portal.

| | |
|---|---|
| Example | `User` |
| Shown when | **Mode** is **Hosted Datadog Metrics** (`hosted-metrics`) |
| Required when | **pluginMode** is `hosted-metrics` |

### Password

_🔒 secret (write-only) · conditionally required · string_

Your password is your Grafana Cloud API Key with read permissions. This can be found in the Prometheus details in your cloud portal.

| | |
|---|---|
| Example | `Password` |
| Shown when | **Mode** is **Hosted Datadog Metrics** (`hosted-metrics`) |
| Required when | **pluginMode** is `hosted-metrics` |

## Additional settings

_This section is optional._

### Show API rate limits

_optional · toggle_

Show Datadog API limits for each queried endpoint. To view the API rate limits, go to the **Query Inspector**, select **JSON**, and set **select source** to **DataFrame structure**.

| | |
|---|---|
| Default | `false` |

### Enable API rate limit threshold

_optional · toggle_

Enable rate limit. Datadog query will stop once it reaches entered threshold.

| | |
|---|---|
| Default | `false` |

### API rate limit threshold %

_optional · number_

Enter percentage of threshold. (If the API hit the % of rate limit, plugin will block subsequent requests till next reset).

| | |
|---|---|
| Default | `100` |
| Range | 0 – 100 |
| Shown when | **Enable API rate limit threshold** is `true` |

### Disable data links

_optional · toggle_

Data links take users directly to the relevant location in the Datadog app when they interact with panels.

| | |
|---|---|
| Default | `false` |

### Response Size

_optional · number_

Set maximum number of items to retrieve in a single API request (default is 100).

| | |
|---|---|
| Default | `100` |
| Example | `100` |

