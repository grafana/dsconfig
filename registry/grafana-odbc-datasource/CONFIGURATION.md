# Sqlyze Datasource configuration

How to configure the **Sqlyze Datasource** data source (`grafana-odbc-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-odbc-datasource).

> This page is generated from [`dsconfig.json`](dsconfig.json). Do not edit it by hand — run `go generate ./...` to refresh.

## Configuration sections

- [Connection Settings](#connection-settings)

## Connection Settings

### Driver

_**required** · string_

This field either accepts '{mydb2}' for driver connections or a absolute path to the odbc driver of your database of choice, for example '/home/driver/db2/libs2.so'.

| | |
|---|---|
| Example | `DSN or path to ODBC Driver` |

### Timeout (seconds)

_optional · string_

| | |
|---|---|
| Default | `10` |
| Example | `10` |

### Driver Settings

_optional · key/value pairs_

These settings will be parsed into key value pairs and concatenated to create the Connection string the plugin will use. For additional settings, please check the keys match exactly what your database requires in a connection string.

Each item has the following fields:

#### Name

_optional · string_

| | |
|---|---|
| Example | `Setting name` |

#### Value

_optional · string_

| | |
|---|---|
| Example | `Setting value` |

## Other settings

### DSN

_optional · string_

Optional Data Source Name. Read only by the backend (pkg/database/connect.go:76-78): when non-empty, the connection string is built as 'DSN=<value>;' instead of 'Driver=<driver>;'. Not written by the configuration editor.

### pwd

_🔒 secret (write-only) · optional · string_

Representative secret for a driver setting whose 'secure' flag is enabled. Secret keys are dynamic and equal to the secure setting's Name; 'pwd' is the conventional password key from the plugin README's driver-settings table. There is no fixed secret key.

| | |
|---|---|
| Example | `Setting value` |

