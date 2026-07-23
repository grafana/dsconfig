# Amazon Timestream configuration

How to configure the **Amazon Timestream** data source (`grafana-timestream-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-timestream-datasource/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Authentication](#authentication)
- [Assume Role](#assume-role)
- [Additional Settings](#additional-settings)
- [Timestream Details](#timestream-details)

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

## Additional Settings

### Endpoint

_optional · string_

Optionally, specify a custom endpoint for the service.

| | |
|---|---|
| Example | `https://query-{cell}.timestream.{region}.amazonaws.com` |
| Shown when | `jsonData_authType != 'grafana_assume_role'` |

### Default Region

_optional · select_

Specify the region, such as for US West (Oregon) use ` us-west-2 ` as the region.

| | |
|---|---|
| Allowed values | `us-east-1`, `us-east-2`, `us-west-2`, `eu-west-1`, `eu-central-1`, `ap-south-1`, `ap-southeast-2`, `ap-northeast-1`, `us-gov-west-1` |

## Timestream Details

### Database

_optional · select_

Default database to use as the {{database}} macro in queries.

### Table

_optional · select_

Default table to use as the {{table}} macro in queries. Depends on the selected database.

| | |
|---|---|
| Shown when | `jsonData_defaultDatabase != ''` |

### Measure

_optional · select_

Default measure to use as the {{measure}} macro in queries. Depends on the selected database and table.

| | |
|---|---|
| Shown when | `jsonData_defaultDatabase != '' && jsonData_defaultTable != ''` |

