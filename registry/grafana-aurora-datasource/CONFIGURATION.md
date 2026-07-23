# Amazon Aurora configuration

How to configure the **Amazon Aurora** data source (`grafana-aurora-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-aurora-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Authentication](#authentication)
- [Assume Role](#assume-role)
- [Additional Settings](#additional-settings)
- [Database Settings](#database-settings)
- [Advanced: Separate Host and Port for Auth](#advanced-separate-host-and-port-for-auth) — _optional_

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

Specify the region, such as for US West (Oregon) use `us-west-2` as the region.

## Database Settings

### Engine

_optional · select_

| | |
|---|---|
| Default | `aurora-postgres` |
| Allowed values | `aurora-postgres` (Aurora (PostgreSQL Compatible)), `aurora-mysql` (Aurora (MySQL Compatible)) |

### Database Name

_optional · string_

| | |
|---|---|
| Example | `Database` |

### Database User

_**required** · string_

| | |
|---|---|
| Example | `postgres` |

### Database Host

_**required** · string_

You can connect to all the read replicas on your Amazon Aurora cluster through a single reader endpoint. Read more about it here (https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/Aurora.Overview.Endpoints.html#Aurora.Endpoints.Reader).

| | |
|---|---|
| Example | `Host` |

### Database Port

_**required** · number_

| | |
|---|---|
| Example | `5432` |

## Advanced: Separate Host and Port for Auth

_This section is optional._

### Advanced: DB Host For Auth

_optional · string_

Optional, if not provided, the dbHost above will be used.

| | |
|---|---|
| Example | `separate host for generating auth token` |

### Advanced: DB Port For Auth

_optional · number_

Optional, if not provided, the dbPort above will be used.

| | |
|---|---|
| Example | `separate port for generating auth token` |

