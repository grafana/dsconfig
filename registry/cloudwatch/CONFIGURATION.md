# CloudWatch configuration

Configuration reference for the **CloudWatch** data source (`cloudwatch`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/grafana/latest/datasources/aws-cloudwatch/).

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
| `jsonData.proxyType` | enum (env, none, url) | jsonData |  | Specify the type of proxy to use. This should not be set if Secure Socks Proxy is enabled. |
| `jsonData.proxyUrl` | string | jsonData | conditional | Proxy URL. Don't set the username or password here |
| `jsonData.proxyUsername` | string | jsonData |  | Optional: Proxy Username. This functionality should only be used with legacy web sites. RFC 2396 warns that interpreting Userinfo this way "is NOT RECOMMENDED, because the passing of authentication information in clear text (such as URI) has proven to be a security risk in almost every case where it has been used." |
| `secureJsonData.proxyPassword` 🔒 | string | secureJsonData |  | Optional: Proxy Password. This functionality should only be used with legacy web sites. RFC 2396 warns that interpreting Userinfo this way "is NOT RECOMMENDED, because the passing of authentication information in clear text (such as URI) has proven to be a security risk in almost every case where it has been used." |
| `jsonData.endpoint` | string | jsonData |  | Optionally, specify a custom endpoint for the service |
| `jsonData.defaultRegion` | enum | jsonData |  | Specify the region, such as for US West (Oregon) use ` us-west-2 ` as the region. |
| `jsonData.customMetricsNamespaces` | string | jsonData |  | Namespaces of Custom Metrics |
| `jsonData.logsTimeout` | string | jsonData |  | Grafana will poll for Cloudwatch Logs results every second until Done status is returned from AWS or timeout is exceeded, in which case Grafana will return an error. Note: For Alerting, the timeout from Grafana config file will take precedence. Must be a valid duration string, such as "30m" (default) "30s" "2000ms" etc. |
| `jsonData.logGroups` | list | jsonData |  | Optionally, specify default log groups for CloudWatch Logs queries. |
| `jsonData.logGroups[].arn` | string | jsonData | yes | ARN |
| `jsonData.logGroups[].name` | string | jsonData | yes | Name |
| `jsonData.logGroups[].accountId` | string | jsonData |  | Account ID |
| `jsonData.logGroups[].accountLabel` | string | jsonData |  | Account Label |
| `jsonData.defaultLogGroups` | list | jsonData |  | Deprecated. Use logGroups instead. Prior storage shape (array of log group names) for default log groups used in CloudWatch Logs queries. |
| `jsonData.tracingDatasourceUid` | string | jsonData |  | Application Signals data source containing traces |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Workspace IAM Role (`ec2_iam_role`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: CloudWatch
    type: cloudwatch
    access: proxy
    jsonData:
      authType: ec2_iam_role
      proxyType: env
      proxyUrl: "Example: https://localhost:3004"
```

**Terraform**

```hcl
resource "grafana_data_source" "cloudwatch_ec2_iam_role" {
  type = "cloudwatch"
  name = "CloudWatch"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "ec2_iam_role"
    proxyType = "env"
    proxyUrl = "Example: https://localhost:3004"
  })
}
```

### Grafana Assume Role (`grafana_assume_role`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: CloudWatch
    type: cloudwatch
    access: proxy
    jsonData:
      authType: grafana_assume_role
      proxyType: env
      proxyUrl: "Example: https://localhost:3004"
```

**Terraform**

```hcl
resource "grafana_data_source" "cloudwatch_grafana_assume_role" {
  type = "cloudwatch"
  name = "CloudWatch"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "grafana_assume_role"
    proxyType = "env"
    proxyUrl = "Example: https://localhost:3004"
  })
}
```

### AWS SDK Default (`default`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: CloudWatch
    type: cloudwatch
    access: proxy
    jsonData:
      authType: default
      proxyType: env
      proxyUrl: "Example: https://localhost:3004"
```

**Terraform**

```hcl
resource "grafana_data_source" "cloudwatch_default" {
  type = "cloudwatch"
  name = "CloudWatch"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "default"
    proxyType = "env"
    proxyUrl = "Example: https://localhost:3004"
  })
}
```

### Access & secret key (`keys`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: CloudWatch
    type: cloudwatch
    access: proxy
    jsonData:
      authType: keys
      proxyType: env
      proxyUrl: "Example: https://localhost:3004"
    secureJsonData:
      accessKey: "<YOUR_ACCESS_KEY_ID>"
      secretKey: "<YOUR_SECRET_ACCESS_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "cloudwatch_keys" {
  type = "cloudwatch"
  name = "CloudWatch"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "keys"
    proxyType = "env"
    proxyUrl = "Example: https://localhost:3004"
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
  - name: CloudWatch
    type: cloudwatch
    access: proxy
    jsonData:
      authType: credentials
      proxyType: env
      proxyUrl: "Example: https://localhost:3004"
```

**Terraform**

```hcl
resource "grafana_data_source" "cloudwatch_credentials" {
  type = "cloudwatch"
  name = "CloudWatch"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "credentials"
    proxyType = "env"
    proxyUrl = "Example: https://localhost:3004"
  })
}
```

