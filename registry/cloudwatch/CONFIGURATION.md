# CloudWatch configuration

How to configure the **CloudWatch** data source (`cloudwatch`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/aws-cloudwatch/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Authentication](#authentication)
- [Assume Role](#assume-role)
- [Proxy Configuration](#proxy-configuration) — _optional_
- [Additional Settings](#additional-settings)
- [Cloudwatch Logs](#cloudwatch-logs)
- [Application Signals trace link](#application-signals-trace-link) — _optional_

## Authentication

### Authentication Provider

_optional · select_

Specify which AWS credentials chain to use.

| | |
|---|---|
| Default | `default` |
| Allowed values | `ec2_iam_role` (Workspace IAM Role), `grafana_assume_role` (Grafana Assume Role), `default` (AWS SDK Default), `keys` (Access & secret key), `credentials` (Credentials file) |

### Credentials Profile Name

_optional · string_

Credentials profile name, as specified in ~/.aws/credentials, leave blank for default.

| | |
|---|---|
| Example | `default` |
| Shown when | **Authentication Provider** is **Credentials file** (`credentials`) |

### Access Key ID

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Shown when | **Authentication Provider** is **Access & secret key** (`keys`) |

### Secret Access Key

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Shown when | **Authentication Provider** is **Access & secret key** (`keys`) |

## Assume Role

### Assume Role ARN

_optional · string_

Optional. Specifying the ARN of a role will ensure that the
                     selected authentication provider is used to assume the role rather than the
                     credentials directly.

| | |
|---|---|
| Example | `arn:aws:iam:*` |
| Must match | `^(arn:aws[a-zA-Z-]*:iam::[0-9]{12}:role/.+)?$` |

### External ID

_optional · string_

If you are assuming a role in another account, that has been created with an external ID, specify the external ID here.

| | |
|---|---|
| Example | `External ID` |
| Shown when | `jsonData_authType != 'grafana_assume_role'` |

## Proxy Configuration

_This section is optional._

### Proxy Type

_optional · select_

Specify the type of proxy to use. This should not be set if Secure Socks Proxy is enabled.

| | |
|---|---|
| Default | `env` |
| Allowed values | `env` (Environment (default)), `none` (None), `url` (URL) |

### Proxy URL

_conditionally required · string_

Proxy URL. Don't set the username or password here.

| | |
|---|---|
| Example | `Example: https://localhost:3004` |
| Shown when | **Proxy Type** is **URL** (`url`) |

### Proxy Username

_optional · string_

Optional: Proxy Username. This functionality should only be used with legacy web sites. RFC 2396 warns that interpreting Userinfo this way "is NOT RECOMMENDED, because the passing of authentication information in clear text (such as URI) has proven to be a security risk in almost every case where it has been used.".

| | |
|---|---|
| Shown when | **Proxy Type** is **URL** (`url`) |

### Proxy Password

_🔒 secret (write-only) · optional · string_

Optional: Proxy Password. This functionality should only be used with legacy web sites. RFC 2396 warns that interpreting Userinfo this way "is NOT RECOMMENDED, because the passing of authentication information in clear text (such as URI) has proven to be a security risk in almost every case where it has been used.".

| | |
|---|---|
| Shown when | **Proxy Type** is **URL** (`url`) |

## Additional Settings

### Endpoint

_optional · string_

Optionally, specify a custom endpoint for the service.

| | |
|---|---|
| Example | `https://{service}.{region}.amazonaws.com` |
| Shown when | `jsonData_authType != 'grafana_assume_role'` |

### Default Region

_optional · select_

Specify the region, such as for US West (Oregon) use ` us-west-2 ` as the region.

### Namespaces of Custom Metrics

_optional · string_

| | |
|---|---|
| Example | `Namespace1,Namespace2` |

## Cloudwatch Logs

### Query Result Timeout

_optional · string_

Grafana will poll for Cloudwatch Logs results every second until Done status is returned from AWS or timeout is exceeded, in which case Grafana will return an error. Note: For Alerting, the timeout from Grafana config file will take precedence. Must be a valid duration string, such as "30m" (default) "30s" "2000ms" etc.

| | |
|---|---|
| Example | `30m` |

### Default Log Groups

_optional · list_

Optionally, specify default log groups for CloudWatch Logs queries.

Each item has the following fields:

#### ARN

_**required** · string_

#### Name

_**required** · string_

#### Account ID

_optional · string_

#### Account Label

_optional · string_

### Default Log Groups

_optional · list_

Deprecated. Use logGroups instead. Prior storage shape (array of log group names) for default log groups used in CloudWatch Logs queries.

## Application Signals trace link

Grafana will automatically create a link to a trace in Application Signals (formerly X-ray) data source if logs contain @xrayTraceId field

_This section is optional._

### Data source

_optional · string_

Application Signals data source containing traces.

