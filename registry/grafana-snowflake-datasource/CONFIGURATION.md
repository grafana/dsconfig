# Snowflake configuration

Configuration reference for the **Snowflake** data source (`grafana-snowflake-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-snowflake-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.account` | string | jsonData | yes | The name of the snowflake account (<account>.snowflakecomputing.com). If not on AWS us-west-2 region, include the region (e.g. <account>.us-east-1). If not on AWS, include the platform as well (e.g. <account>.us-east1.gcp) |
| `jsonData.region` | string | jsonData |  | Deprecated; prefer including the region in the 'Account' field |
| `jsonData.authType` | enum (password, keypair, pat, oauth) | jsonData |  | Authentication type of snowflake |
| `jsonData.username` | string | jsonData | conditional | The username assigned to the Snowflake user (via CREATE USER) |
| `secureJsonData.password` 🔒 | string | secureJsonData | conditional | The password assigned to the Snowflake account |
| `secureJsonData.privateKey` 🔒 | string (multiline) | secureJsonData | conditional | Private Key for the key pair Authentication |
| `secureJsonData.privateKeyPassphrase` 🔒 | string | secureJsonData |  | Passphrase used to decrypt an encrypted private key. Leave empty if your private key is unencrypted. |
| `secureJsonData.patToken` 🔒 | string | secureJsonData | conditional | Programmatic Access Token secret |
| `jsonData.oauthPassThru` | boolean | jsonData | conditional | Forward the user's upstream OAuth identity to the data source (Their access token gets passed along). |
| `jsonData.settings` | list | jsonData |  | Connection settings |
| `jsonData.settings[].name` | string | jsonData |  | Name |
| `jsonData.settings[].value` | string | jsonData |  | Value |
| `jsonData.settings[].secure` | boolean | jsonData |  |  |
| `jsonData.role` | string | jsonData |  | Assume a role other than the default role for queries sent by this datasource. The role specified here must still be granted to the user via 'GRANT ROLE' |
| `jsonData.warehouse` | string | jsonData |  | The default warehouse for queries sent by this datasource |
| `jsonData.database` | string | jsonData |  | The default database for queries sent by this datasource |
| `jsonData.schema` | string | jsonData |  | The default schema for queries sent by this datasource |
| `jsonData.timeInterval` | string | jsonData |  | A lower limit for the $__interval and $__interval_ms macros |
| `jsonData.rowLimit` | number | jsonData |  | Limits the Max number of rows read from query results (applied by the plugin, not in the database). If unset, falls back to GF_DATAPROXY_ROW_LIMIT, or unlimited if not set. |
| `jsonData.loginTimeout` | number | jsonData |  | Connection timeout in seconds. Suggested value: 5 - 120 |
| `jsonData.requestTimeout` | number | jsonData |  | Request timeout in seconds. Suggested values: 30 - 120 |
| `jsonData.defaultInterpolation` | enum (, raw, sqlstring, regex, csv, distributed, doublequote, glob, json, lucene, percentencode, pipe, singlequote, text, queryparam) | jsonData |  | The formatting of the variable interpolation. Choose None for default behavior. For best results and simplified experience, choose SQL String |
| `jsonData.defaultQuery` | string (multiline) | jsonData |  | Default query to be used when adding a new snowflake query to the panel |
| `jsonData.defaultVariableQuery` | string (multiline) | jsonData |  | Default query to be used when adding a new snowflake query to the dashboard variable |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Password (`password`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Snowflake
    type: grafana-snowflake-datasource
    access: proxy
    jsonData:
      account: Snowflake Account
      authType: password
      defaultInterpolation: ""
      username: Snowflake Username
    secureJsonData:
      password: "<YOUR_PASSWORD>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_snowflake_datasource_password" {
  type = "grafana-snowflake-datasource"
  name = "Snowflake"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    account = "Snowflake Account"
    authType = "password"
    defaultInterpolation = ""
    username = "Snowflake Username"
  })

  secure_json_data_encoded = jsonencode({
    password = "<YOUR_PASSWORD>"
  })
}
```

### Key Pair (`keypair`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Snowflake
    type: grafana-snowflake-datasource
    access: proxy
    jsonData:
      account: Snowflake Account
      authType: keypair
      defaultInterpolation: ""
      username: Snowflake Username
    secureJsonData:
      privateKey: "<YOUR_PRIVATE_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_snowflake_datasource_keypair" {
  type = "grafana-snowflake-datasource"
  name = "Snowflake"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    account = "Snowflake Account"
    authType = "keypair"
    defaultInterpolation = ""
    username = "Snowflake Username"
  })

  secure_json_data_encoded = jsonencode({
    privateKey = "<YOUR_PRIVATE_KEY>"
  })
}
```

### Programmatic Access Token (`pat`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Snowflake
    type: grafana-snowflake-datasource
    access: proxy
    jsonData:
      account: Snowflake Account
      authType: pat
      defaultInterpolation: ""
      username: Snowflake Username
    secureJsonData:
      patToken: "<YOUR_TOKEN>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_snowflake_datasource_pat" {
  type = "grafana-snowflake-datasource"
  name = "Snowflake"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    account = "Snowflake Account"
    authType = "pat"
    defaultInterpolation = ""
    username = "Snowflake Username"
  })

  secure_json_data_encoded = jsonencode({
    patToken = "<YOUR_TOKEN>"
  })
}
```

### OAuth (`oauth`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Snowflake
    type: grafana-snowflake-datasource
    access: proxy
    jsonData:
      account: Snowflake Account
      authType: oauth
      defaultInterpolation: ""
      oauthPassThru: false
      username: Snowflake Username
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_snowflake_datasource_oauth" {
  type = "grafana-snowflake-datasource"
  name = "Snowflake"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    account = "Snowflake Account"
    authType = "oauth"
    defaultInterpolation = ""
    oauthPassThru = false
    username = "Snowflake Username"
  })
}
```

