# Amazon Timestream configuration

Configuration reference for the **Amazon Timestream** data source (`grafana-timestream-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-timestream-datasource/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.authType` | enum (ec2_iam_role, grafana_assume_role, default, keys, credentials) | jsonData |  | Specify which AWS credentials chain to use. |
| `jsonData.profile` | string | jsonData |  | Credentials profile name, as specified in ~/.aws/credentials, leave blank for default. |
| `secureJsonData.accessKey` 🔒 | string | secureJsonData | conditional | Access Key ID |
| `secureJsonData.secretKey` 🔒 | string | secureJsonData | conditional | Secret Access Key |
| `secureJsonData.sessionToken` 🔒 | string | secureJsonData |  |  |
| `jsonData.assumeRoleArn` | string | jsonData |  | Optional. Specifying the ARN of a role will ensure that the                      selected authentication provider is used to assume the role rather than the                      credentials directly. |
| `jsonData.externalId` | string | jsonData |  | If you are assuming a role in another account, that has been created with an external ID, specify the external ID here. |
| `jsonData.endpoint` | string | jsonData |  | Optionally, specify a custom endpoint for the service |
| `jsonData.defaultRegion` | enum (us-east-1, us-east-2, us-west-2, eu-west-1, eu-central-1, ap-south-1, ap-southeast-2, ap-northeast-1, us-gov-west-1) | jsonData |  | Specify the region, such as for US West (Oregon) use ` us-west-2 ` as the region. |
| `jsonData.defaultDatabase` | enum | jsonData |  | Default database to use as the {{database}} macro in queries. |
| `jsonData.defaultTable` | enum | jsonData |  | Default table to use as the {{table}} macro in queries. Depends on the selected database. |
| `jsonData.defaultMeasure` | enum | jsonData |  | Default measure to use as the {{measure}} macro in queries. Depends on the selected database and table. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Workspace IAM Role (`ec2_iam_role`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Timestream
    type: grafana-timestream-datasource
    access: proxy
    jsonData:
      authType: ec2_iam_role
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_timestream_datasource_ec2_iam_role" {
  type = "grafana-timestream-datasource"
  name = "Amazon Timestream"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "ec2_iam_role"
  })
}
```

### Grafana Assume Role (`grafana_assume_role`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Timestream
    type: grafana-timestream-datasource
    access: proxy
    jsonData:
      authType: grafana_assume_role
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_timestream_datasource_grafana_assume_role" {
  type = "grafana-timestream-datasource"
  name = "Amazon Timestream"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "grafana_assume_role"
  })
}
```

### AWS SDK Default (`default`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Timestream
    type: grafana-timestream-datasource
    access: proxy
    jsonData:
      authType: default
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_timestream_datasource_default" {
  type = "grafana-timestream-datasource"
  name = "Amazon Timestream"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "default"
  })
}
```

### Access & secret key (`keys`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Timestream
    type: grafana-timestream-datasource
    access: proxy
    jsonData:
      authType: keys
    secureJsonData:
      accessKey: "<YOUR_ACCESS_KEY_ID>"
      secretKey: "<YOUR_SECRET_ACCESS_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_timestream_datasource_keys" {
  type = "grafana-timestream-datasource"
  name = "Amazon Timestream"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "keys"
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
  - name: Amazon Timestream
    type: grafana-timestream-datasource
    access: proxy
    jsonData:
      authType: credentials
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_timestream_datasource_credentials" {
  type = "grafana-timestream-datasource"
  name = "Amazon Timestream"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "credentials"
  })
}
```

