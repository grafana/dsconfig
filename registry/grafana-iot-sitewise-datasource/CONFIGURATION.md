# AWS IoT SiteWise configuration

How to configure the **AWS IoT SiteWise** data source (`grafana-iot-sitewise-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-iot-sitewise-datasource/).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Authentication](#authentication)
- [Assume Role](#assume-role)
- [Additional Settings](#additional-settings)
- [Edge settings](#edge-settings) — _optional_

## Authentication

### Authentication Provider

_optional · select_

Specify which AWS credentials chain to use.

| | |
|---|---|
| Default | `default` |
| Allowed values | `ec2_iam_role` (Workspace IAM Role), `grafana_assume_role` (Grafana Assume Role), `default` (AWS SDK Default), `keys` (Access & secret key), `credentials` (Credentials file) |
| Shown when | `jsonData_defaultRegion != 'Edge' || jsonData_edgeAuthMode == 'default'` |

### Credentials Profile Name

_optional · string_

Credentials profile name, as specified in ~/.aws/credentials, leave blank for default.

| | |
|---|---|
| Example | `default` |
| Shown when | `(jsonData_defaultRegion != 'Edge' || jsonData_edgeAuthMode == 'default') && jsonData_authType == 'credentials'` |

### Access Key ID

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Shown when | `(jsonData_defaultRegion != 'Edge' || jsonData_edgeAuthMode == 'default') && jsonData_authType == 'keys'` |
| Required when | **Authentication Provider** is **Access & secret key** (`keys`) |

### Secret Access Key

_🔒 secret (write-only) · conditionally required · string_

| | |
|---|---|
| Shown when | `(jsonData_defaultRegion != 'Edge' || jsonData_edgeAuthMode == 'default') && jsonData_authType == 'keys'` |
| Required when | **Authentication Provider** is **Access & secret key** (`keys`) |

## Assume Role

### Assume Role ARN

_optional · string_

Optional. Specifying the ARN of a role will ensure that the selected authentication provider is used to assume the role rather than the credentials directly.

| | |
|---|---|
| Example | `arn:aws:iam:*` |
| Must match | `^(arn:aws[a-zA-Z-]*:iam::[0-9]{12}:role/.+)?$` |
| Shown when | `(jsonData_defaultRegion != 'Edge' || jsonData_edgeAuthMode == 'default') && jsonData_authType != 'grafana_assume_role'` |

### External ID

_optional · string_

If you are assuming a role in another account, that has been created with an external ID, specify the external ID here.

| | |
|---|---|
| Example | `External ID` |
| Shown when | `(jsonData_defaultRegion != 'Edge' || jsonData_edgeAuthMode == 'default') && jsonData_authType != 'grafana_assume_role'` |

## Additional Settings

### Endpoint

_conditionally required · string_

Optionally, specify a custom endpoint for the service.

| | |
|---|---|
| Example | `https://{service}.{region}.amazonaws.com` |
| Shown when | `(jsonData_defaultRegion == 'Edge' && jsonData_edgeAuthMode != 'default') || jsonData_authType != 'grafana_assume_role'` |
| Required when | **Default Region** is `Edge` |

### Default Region

_optional · select_

Specify the region, such as for US West (Oregon) use `us-west-2` as the region.

| | |
|---|---|
| Allowed values | `us-east-2`, `us-east-1`, `us-west-2`, `ap-south-1`, `ap-northeast-2`, `ap-southeast-1`, `ap-southeast-2`, `ap-northeast-1`, `ca-central-1`, `eu-central-1`, `eu-west-1`, `us-gov-west-1`, `cn-north-1`, `Edge` |

## Edge settings

_This section is optional._

### Authentication Mode

_optional · select_

| | |
|---|---|
| Default | `default` |
| Allowed values | `default` (Standard), `linux` (Linux), `ldap` (LDAP) |
| Shown when | **Default Region** is `Edge` |

### Username

_conditionally required · string_

The username set to local authentication proxy.

| | |
|---|---|
| Shown when | `jsonData_defaultRegion == 'Edge' && jsonData_edgeAuthMode != 'default'` |

### Password

_🔒 secret (write-only) · conditionally required · string_

The password sent to local authentication proxy.

| | |
|---|---|
| Shown when | `jsonData_defaultRegion == 'Edge' && jsonData_edgeAuthMode != 'default'` |

### SSL Certificate

_🔒 secret (write-only) · conditionally required · multiline text_

Certificate for SSL enabled authentication.

| | |
|---|---|
| Example | `Begins with -----BEGIN CERTIFICATE------` |
| Shown when | **Default Region** is `Edge` |

