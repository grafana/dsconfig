# AWS IoT TwinMaker configuration

How to configure the **AWS IoT TwinMaker** data source (`grafana-iot-twinmaker-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-iot-twinmaker-app/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Authentication](#authentication)
- [Assume Role](#assume-role)
- [Additional Settings](#additional-settings)
- [Twinmaker Settings](#twinmaker-settings)

## Authentication

### Authentication Provider

_optional · select_

Specify which AWS credentials chain to use.

| | |
|---|---|
| Default | `default` |
| Allowed values | `ec2_iam_role` (Workspace IAM Role), `default` (AWS SDK Default), `keys` (Access & secret key), `credentials` (Credentials file) |

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

_**required** · string_

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
| Example | `https://{service}.{region}.amazonaws.com` |
| Shown when | `jsonData_authType != 'grafana_assume_role'` |

### Default Region

_optional · select_

Specify the region, such as for US West (Oregon) use ` us-west-2 ` as the region.

| | |
|---|---|
| Default | `us-east-1` |
| Allowed values | `ap-south-1`, `ap-northeast-1`, `ap-northeast-2`, `ap-southeast-1`, `ap-southeast-2`, `eu-central-1`, `eu-west-1`, `us-east-1`, `us-west-2`, `us-gov-west-1`, `cn-north-1` |

## Twinmaker Settings

### Workspace

_**required** · select_

| | |
|---|---|
| Example | `Select a workspace` |

### Define write permissions for Alarm Configuration Panel

_optional · toggle_

| | |
|---|---|
| Default | `false` |

### Assume Role ARN Write

_optional · string_

Specify the ARN of a role to assume when writing property values in IoT TwinMaker.

| | |
|---|---|
| Example | `arn:aws:iam:*` |
| Must match | `^(arn:aws[a-zA-Z-]*:iam::[0-9]{12}:role/.+)?$` |
| Shown when | **Define write permissions for Alarm Configuration Panel** is `true` |

