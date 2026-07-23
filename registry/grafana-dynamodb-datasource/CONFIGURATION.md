# DynamoDB configuration

How to configure the **DynamoDB** data source (`grafana-dynamodb-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-dynamodb-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Authentication](#authentication)
- [Additional Settings](#additional-settings)
- [Driver Settings](#driver-settings) — _optional_
- [Legacy Migration](#legacy-migration) — _optional_

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

## Additional Settings

### Endpoint

_optional · string_

Optionally, specify a custom endpoint for the service.

| | |
|---|---|
| Example | `https://{service}.{region}.amazonaws.com` |

### Default Region

_optional · select_

Specify the region, such as for US West (Oregon) use ` us-west-2 ` as the region.

## Driver Settings

_This section is optional._

_No user-configurable fields in this section._

## Legacy Migration

_This section is optional._

_No user-configurable fields in this section._

