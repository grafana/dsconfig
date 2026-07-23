# Sqlyze Datasource configuration

Configuration reference for the **Sqlyze Datasource** data source (`grafana-odbc-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-odbc-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.driver` | string | jsonData | yes | This field either accepts '{mydb2}' for driver connections or a absolute path to the odbc driver of your database of choice, for example '/home/driver/db2/libs2.so'. |
| `jsonData.timeout` | string | jsonData |  | Timeout (seconds) |
| `jsonData.settings` | map<string,string> | jsonData |  | These settings will be parsed into key value pairs and concatenated to create the Connection string the plugin will use. For additional settings, please check the keys match exactly what your database requires in a connection string. |
| `jsonData.settings[].name` | string | jsonData |  | Name |
| `jsonData.settings[].value` | string | jsonData |  | Value |
| `jsonData.settings[].secure` | boolean | jsonData |  |  |
| `jsonData.DSN` | string | jsonData |  | Optional Data Source Name. Read only by the backend (pkg/database/connect.go:76-78): when non-empty, the connection string is built as 'DSN=<value>;' instead of 'Driver=<driver>;'. Not written by the configuration editor. |
| `secureJsonData.pwd` 🔒 | string | secureJsonData |  | Representative secret for a driver setting whose 'secure' flag is enabled. Secret keys are dynamic and equal to the secure setting's Name; 'pwd' is the conventional password key from the plugin README's driver-settings table. There is no fixed secret key. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Sqlyze Datasource
    type: grafana-odbc-datasource
    access: proxy
    jsonData:
      driver: DSN or path to ODBC Driver
      timeout: "10"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_odbc_datasource" {
  type = "grafana-odbc-datasource"
  name = "Sqlyze Datasource"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    driver = "DSN or path to ODBC Driver"
    timeout = "10"
  })
}
```

