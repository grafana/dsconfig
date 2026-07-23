# Snowflake configuration

How to configure the **Snowflake** data source (`grafana-snowflake-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-snowflake-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection](#connection)
- [Connection settings](#connection-settings) — _optional_
- [Environment](#environment)
- [Customization](#customization) — _optional_

## Connection

### Account

_**required** · string_

The name of the snowflake account (<account>.snowflakecomputing.com). If not on AWS us-west-2 region, include the region (e.g. <account>.us-east-1). If not on AWS, include the platform as well (e.g. <account>.us-east1.gcp).

| | |
|---|---|
| Example | `Snowflake Account` |

### Region

_optional · string_

Deprecated; prefer including the region in the 'Account' field.

| | |
|---|---|
| Example | `Default region` |

### Authentication Type

_optional · radio_

Authentication type of snowflake.

| | |
|---|---|
| Default | `password` |
| Allowed values | `password` (Password), `keypair` (Key Pair), `pat` (Programmatic Access Token), `oauth` (OAuth) |

### Username

_conditionally required · string_

The username assigned to the Snowflake user (via CREATE USER).

| | |
|---|---|
| Example | `Snowflake Username` |
| Shown when | `jsonData_authType != 'oauth'` |

### Password

_🔒 secret (write-only) · conditionally required · string_

The password assigned to the Snowflake account.

| | |
|---|---|
| Example | `Snowflake Password` |
| Shown when | **Authentication Type** is **Password** (`password`) |
| Required when | `jsonData_authType == 'password' || jsonData_authType == ''` |

### Private key

_🔒 secret (write-only) · conditionally required · multiline text_

Private Key for the key pair Authentication.

| | |
|---|---|
| Example | `Begins with -----BEGIN PRIVATE KEY-----` |
| Shown when | **Authentication Type** is **Key Pair** (`keypair`) |

### Private key passphrase

_🔒 secret (write-only) · optional · string_

Passphrase used to decrypt an encrypted private key. Leave empty if your private key is unencrypted.

| | |
|---|---|
| Example | `Passphrase for encrypted private key` |
| Shown when | **Authentication Type** is **Key Pair** (`keypair`) |

### Token

_🔒 secret (write-only) · conditionally required · string_

Programmatic Access Token secret.

| | |
|---|---|
| Example | `Programmatic access token secret` |
| Shown when | **Authentication Type** is **Programmatic Access Token** (`pat`) |

### Forward OAuth Identity

_conditionally required · toggle_

Forward the user's upstream OAuth identity to the data source (Their access token gets passed along).

| | |
|---|---|
| Shown when | **Authentication Type** is **OAuth** (`oauth`) |

## Connection settings

_This section is optional._

### Connection settings

_optional · list_

**Supported connection settings**

Session parameters can be added as a key/value pairs. List of all existing settings can be found in the [snowflake documentation](https://docs.snowflake.com/en/sql-reference/parameters.html#session-parameters)

You can also set secure connection parameters by checking the lock icon.

Each item has the following fields:

#### Name

_optional · string_

| | |
|---|---|
| Example | `Name of the setting` |

#### Value

_optional · string_

| | |
|---|---|
| Example | `Value of the setting` |

## Environment

### Role

_optional · string_

Assume a role other than the default role for queries sent by this datasource. The role specified here must still be granted to the user via 'GRANT ROLE'.

| | |
|---|---|
| Example | `Default role` |

### Warehouse

_optional · string_

The default warehouse for queries sent by this datasource.

| | |
|---|---|
| Example | `Default warehouse` |

### Database

_optional · string_

The default database for queries sent by this datasource.

| | |
|---|---|
| Example | `Default database` |

### Schema

_optional · string_

The default schema for queries sent by this datasource.

| | |
|---|---|
| Example | `Default schema` |

## Customization

_This section is optional._

### Min Interval

_optional · string_

A lower limit for the $__interval and $__interval_ms macros.

| | |
|---|---|
| Example | `10s` |

### Row Limit

_optional · number_

Limits the Max number of rows read from query results (applied by the plugin, not in the database). If unset, falls back to GF_DATAPROXY_ROW_LIMIT, or unlimited if not set.

### Connection Timeout (sec)

_optional · number_

Connection timeout in seconds. Suggested value: 5 - 120.

| | |
|---|---|
| Example | `5` |

### Request Timeout (sec)

_optional · number_

Request timeout in seconds. Suggested values: 30 - 120.

| | |
|---|---|
| Example | `120` |

### Variable Interpolation Format

_optional · select_

The formatting of the variable interpolation. Choose None for default behavior. For best results and simplified experience, choose SQL String.

| | |
|---|---|
| Default | `""` |
| Allowed values | `` (None), `raw` (Raw), `sqlstring` (Sql String), `regex` (Regex), `csv` (CSV), `distributed` (Distributed (OpenTSDB)), `doublequote` (Double Quote), `glob` (Glob (Graphite)), `json` (JSON), `lucene` (Lucene (Elasticsearch)), `percentencode` (Percentencode), `pipe` (Pipe), `singlequote` (Single Quote), `text` (Text), `queryparam` (Query Param) |

### Default Query

_optional · multiline text_

Default query to be used when adding a new snowflake query to the panel.

| | |
|---|---|
| Example | `-- OPTIONAL: Default template to be used for new query.

SELECT 
	 $__timeGroup(<time_column>, $__interval) as time,
	 <value_column>
 FROM <metric_table>
 WHERE $__timeFilter(time)` |

### Default Variable Query

_optional · multiline text_

Default query to be used when adding a new snowflake query to the dashboard variable.

| | |
|---|---|
| Example | `-- OPTIONAL: Default template to be used for new variable query.

SELECT DISTINCT <column_name> FROM <metric_table> LIMIT 1000;` |

