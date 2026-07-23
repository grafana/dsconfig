# Amazon Athena configuration

How to configure the **Amazon Athena** data source (`grafana-athena-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-athena-datasource/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Authentication](#authentication)
- [Assume Role](#assume-role)
- [Additional Settings](#additional-settings)
- [Athena Details](#athena-details)

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

Optional. Specifying the ARN of a role will ensure that the selected authentication provider is used to assume the role rather than the credentials directly.

| | |
|---|---|
| Example | `arn:aws:iam:*` |
| Must match | `^(arn:aws[a-zA-Z-]*:iam::[0-9]{12}:role/.+)?$` |
| Shown when | `jsonData_authType != 'grafana_assume_role'` |

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
| Example | `https://{service}.{region}.amazonaws.com` |
| Shown when | `jsonData_authType != 'grafana_assume_role'` |

### Default Region

_optional · select_

Specify the region, such as for US West (Oregon) use ` us-west-2 ` as the region.

## Athena Details

### Data source

_**required** · select_

### Database

_**required** · select_

### Workgroup

_**required** · select_

### Output Location

_optional · string_

Optional. If not specified, the default query result location from the Workgroup configuration will be used.

| | |
|---|---|
| Example | `s3://` |

