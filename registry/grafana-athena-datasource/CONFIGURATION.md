# Amazon Athena configuration

Configuration reference for the **Amazon Athena** data source (`grafana-athena-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-athena-datasource/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.authType` | enum (ec2_iam_role, grafana_assume_role, default, keys, credentials) | jsonData |  | Specify which AWS credentials chain to use. |
| `jsonData.profile` | string | jsonData |  | Credentials profile name, as specified in ~/.aws/credentials, leave blank for default. |
| `secureJsonData.accessKey` 🔒 | string | secureJsonData | conditional | Access Key ID |
| `secureJsonData.secretKey` 🔒 | string | secureJsonData | conditional | Secret Access Key |
| `secureJsonData.sessionToken` 🔒 | string | secureJsonData |  |  |
| `jsonData.assumeRoleArn` | string | jsonData |  | Optional. Specifying the ARN of a role will ensure that the selected authentication provider is used to assume the role rather than the credentials directly. |
| `jsonData.externalId` | string | jsonData |  | If you are assuming a role in another account, that has been created with an external ID, specify the external ID here. |
| `jsonData.endpoint` | string | jsonData |  | Optionally, specify a custom endpoint for the service |
| `jsonData.defaultRegion` | enum | jsonData |  | Specify the region, such as for US West (Oregon) use ` us-west-2 ` as the region. |
| `jsonData.catalog` | enum | jsonData | yes | Data source |
| `jsonData.database` | enum | jsonData | yes | Database |
| `jsonData.workgroup` | enum | jsonData | yes | Workgroup |
| `jsonData.outputLocation` | string | jsonData |  | Optional. If not specified, the default query result location from the Workgroup configuration will be used. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Workspace IAM Role (`ec2_iam_role`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Athena
    type: grafana-athena-datasource
    access: proxy
    jsonData:
      authType: ec2_iam_role
      catalog: "<YOUR_DATA_SOURCE>"
      database: "<YOUR_DATABASE>"
      workgroup: "<YOUR_WORKGROUP>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_athena_datasource_ec2_iam_role" {
  type = "grafana-athena-datasource"
  name = "Amazon Athena"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "ec2_iam_role"
    catalog = "<YOUR_DATA_SOURCE>"
    database = "<YOUR_DATABASE>"
    workgroup = "<YOUR_WORKGROUP>"
  })
}
```

### Grafana Assume Role (`grafana_assume_role`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Athena
    type: grafana-athena-datasource
    access: proxy
    jsonData:
      authType: grafana_assume_role
      catalog: "<YOUR_DATA_SOURCE>"
      database: "<YOUR_DATABASE>"
      workgroup: "<YOUR_WORKGROUP>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_athena_datasource_grafana_assume_role" {
  type = "grafana-athena-datasource"
  name = "Amazon Athena"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "grafana_assume_role"
    catalog = "<YOUR_DATA_SOURCE>"
    database = "<YOUR_DATABASE>"
    workgroup = "<YOUR_WORKGROUP>"
  })
}
```

### AWS SDK Default (`default`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Athena
    type: grafana-athena-datasource
    access: proxy
    jsonData:
      authType: default
      catalog: "<YOUR_DATA_SOURCE>"
      database: "<YOUR_DATABASE>"
      workgroup: "<YOUR_WORKGROUP>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_athena_datasource_default" {
  type = "grafana-athena-datasource"
  name = "Amazon Athena"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "default"
    catalog = "<YOUR_DATA_SOURCE>"
    database = "<YOUR_DATABASE>"
    workgroup = "<YOUR_WORKGROUP>"
  })
}
```

### Access & secret key (`keys`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Athena
    type: grafana-athena-datasource
    access: proxy
    jsonData:
      authType: keys
      catalog: "<YOUR_DATA_SOURCE>"
      database: "<YOUR_DATABASE>"
      workgroup: "<YOUR_WORKGROUP>"
    secureJsonData:
      accessKey: "<YOUR_ACCESS_KEY_ID>"
      secretKey: "<YOUR_SECRET_ACCESS_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_athena_datasource_keys" {
  type = "grafana-athena-datasource"
  name = "Amazon Athena"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "keys"
    catalog = "<YOUR_DATA_SOURCE>"
    database = "<YOUR_DATABASE>"
    workgroup = "<YOUR_WORKGROUP>"
  })

  secure_json_data_encoded = jsonencode({
    accessKey = "<YOUR_ACCESS_KEY_ID>"
    secretKey = "<YOUR_SECRET_ACCESS_KEY>"
  })
}
```

### Credentials file (`credentials`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Athena
    type: grafana-athena-datasource
    access: proxy
    jsonData:
      authType: credentials
      catalog: "<YOUR_DATA_SOURCE>"
      database: "<YOUR_DATABASE>"
      workgroup: "<YOUR_WORKGROUP>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_athena_datasource_credentials" {
  type = "grafana-athena-datasource"
  name = "Amazon Athena"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "credentials"
    catalog = "<YOUR_DATA_SOURCE>"
    database = "<YOUR_DATABASE>"
    workgroup = "<YOUR_WORKGROUP>"
  })
}
```

