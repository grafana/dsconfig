# DynamoDB configuration

Configuration reference for the **DynamoDB** data source (`grafana-dynamodb-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-dynamodb-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.authType` | enum (ec2_iam_role, default, keys, credentials) | jsonData |  | Specify which AWS credentials chain to use. |
| `jsonData.profile` | string | jsonData |  | Credentials profile name, as specified in ~/.aws/credentials, leave blank for default. |
| `secureJsonData.accessKey` 🔒 | string | secureJsonData | conditional | Access Key ID |
| `secureJsonData.secretKey` 🔒 | string | secureJsonData | conditional | Secret Access Key |
| `secureJsonData.sessionToken` 🔒 | string | secureJsonData |  |  |
| `jsonData.endpoint` | string | jsonData |  | Optionally, specify a custom endpoint for the service |
| `jsonData.defaultRegion` | enum | jsonData |  | Specify the region, such as for US West (Oregon) use ` us-west-2 ` as the region. |
| `jsonData.isV2` | boolean | jsonData |  |  |
| `jsonData.timeout` | string | jsonData |  |  |
| `jsonData.retries` | string | jsonData |  |  |
| `jsonData.pause` | string | jsonData |  |  |
| `jsonData.accessId` | string | jsonData |  |  |
| `jsonData.region` | string | jsonData |  |  |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Workspace IAM Role (`ec2_iam_role`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: DynamoDB
    type: grafana-dynamodb-datasource
    access: proxy
    jsonData:
      authType: ec2_iam_role
      isV2: true
      pause: "5"
      retries: "5"
      timeout: "60"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_dynamodb_datasource_ec2_iam_role" {
  type = "grafana-dynamodb-datasource"
  name = "DynamoDB"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "ec2_iam_role"
    isV2 = true
    pause = "5"
    retries = "5"
    timeout = "60"
  })
}
```

### AWS SDK Default (`default`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: DynamoDB
    type: grafana-dynamodb-datasource
    access: proxy
    jsonData:
      authType: default
      isV2: true
      pause: "5"
      retries: "5"
      timeout: "60"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_dynamodb_datasource_default" {
  type = "grafana-dynamodb-datasource"
  name = "DynamoDB"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "default"
    isV2 = true
    pause = "5"
    retries = "5"
    timeout = "60"
  })
}
```

### Access & secret key (`keys`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: DynamoDB
    type: grafana-dynamodb-datasource
    access: proxy
    jsonData:
      authType: keys
      isV2: true
      pause: "5"
      retries: "5"
      timeout: "60"
    secureJsonData:
      accessKey: "<YOUR_ACCESS_KEY_ID>"
      secretKey: "<YOUR_SECRET_ACCESS_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_dynamodb_datasource_keys" {
  type = "grafana-dynamodb-datasource"
  name = "DynamoDB"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "keys"
    isV2 = true
    pause = "5"
    retries = "5"
    timeout = "60"
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
  - name: DynamoDB
    type: grafana-dynamodb-datasource
    access: proxy
    jsonData:
      authType: credentials
      isV2: true
      pause: "5"
      retries: "5"
      timeout: "60"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_dynamodb_datasource_credentials" {
  type = "grafana-dynamodb-datasource"
  name = "DynamoDB"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "credentials"
    isV2 = true
    pause = "5"
    retries = "5"
    timeout = "60"
  })
}
```

