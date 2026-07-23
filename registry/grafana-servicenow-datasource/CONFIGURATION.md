# ServiceNow configuration

How to configure the **ServiceNow** data source (`grafana-servicenow-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-servicenow-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Authentication](#authentication)
- [Additional settings](#additional-settings) — _optional_

## Connection

### URL

_**required** · string_

Your access method is Server, this means the URL needs to be accessible from the grafana backend/server.

| | |
|---|---|
| Example | `https://<YOUR INSTANCE ID>.service-now.com` |

## Authentication

### Authentication Type

_optional · radio_

Type of authentication to use. Defaults to basic auth.

| | |
|---|---|
| Default | `basicAuth` |
| Allowed values | `basicAuth` (Basic auth), `serviceNowOAuth` (ServiceNow OAuth) |

### Username

_**required** · string_

Username of the ServiceNow account.

| | |
|---|---|
| Example | `ServiceNow username` |

**Instructions to set permissions in servicenow**

The ServiceNow user provided requires read only access to the following tables/fields via ACL rules

| Table | ACL Rules | Details |
| --- | --- | --- |
| sys_db_object | `sys_db_object`, `sys_db_object.name`, `sys_db_object.label`, `sys_db_object.sys_name`/ Display Name, `sys_db_object.super_class`/ Extends Table | This table is used to populate the list of available tables/fields. |
| sys_dictionary, sys_glide_object | `sys_dictionary`, `sys_dictionary.*`, `sys_glide_object`, `sys_glide_object.*` | This table is used to populate the list of selectable fields per table, and is used as a schema for data responses. |
| sys_choice | `sys_choice`, `sys_choice.*` | This table is used to populate the available options when filtering using a `choice` type field. |
| incident | `incident` | This table is used in the health check to ensure that the plugin is able to communicate with ServiceNow. |

#### Limited permissions

> It is highly recommended that the service account being used only has access to the necessary tables

If the service account has access to too many tables, then you may encounter performance issues in the query editor. Please ensure that the ServiceNow account provided only has access to the necessary tables

### Password

_🔒 secret (write-only) · **required** · string_

Password for the ServiceNow account.

| | |
|---|---|
| Example | `Password for the ServiceNow account` |

### Client ID

_conditionally required · string_

Client ID for OAuth.

| | |
|---|---|
| Example | `OAuth Client ID` |
| Shown when | **Authentication Type** is **ServiceNow OAuth** (`serviceNowOAuth`) |

### Client Secret

_🔒 secret (write-only) · conditionally required · string_

Client Secret for OAuth.

| | |
|---|---|
| Example | `OAuth Client Secret` |
| Shown when | **Authentication Type** is **ServiceNow OAuth** (`serviceNowOAuth`) |

## Additional settings

_This section is optional._

### Use Sys Tables?

_optional · toggle_

Query sys tables for schema/meta lookups (requires elevated permissions).

| | |
|---|---|
| Default | `false` |

### Query Timeout

_optional · number_

Maximum time in seconds for queries to complete. Increase this for slow ServiceNow instances or large tables. Default is 30 seconds.

| | |
|---|---|
| Default | `30` |
| Example | `30` |

### Custom HTTP Headers

_optional · list_

Additional HTTP headers sent with every request. Header names are stored in jsonData (httpHeaderName<N>); header values are write-only in secureJsonData (httpHeaderValue<N>).

Each item has the following fields:

#### Header

_**required** · string_

| | |
|---|---|
| Example | `X-Custom-Header` |
| Must match | `^[A-Za-z][A-Za-z0-9-]*$` |

#### Value

_optional · string_

| | |
|---|---|
| Example | `Header Value` |

## Other settings

### oauthEnabled

_optional · boolean_

Deprecated legacy boolean that predates `authMethod`. Older plugin versions stored `oauthEnabled: true` to select ServiceNow OAuth. Not written by the current config editor, but still read for backwards compatibility by both the editor (initial auth-method derivation) and the backend (`GetAuthMethod`): when `authMethod` is empty, `oauthEnabled: true` selects `serviceNowOAuth`, otherwise `basicAuth`.

