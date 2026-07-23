# AWS IoT SiteWise configuration

Configuration reference for the **AWS IoT SiteWise** data source (`grafana-iot-sitewise-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-iot-sitewise-datasource/).

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
| `jsonData.endpoint` | string | jsonData | conditional | Optionally, specify a custom endpoint for the service |
| `jsonData.defaultRegion` | enum (us-east-2, us-east-1, us-west-2, ap-south-1, ap-northeast-2, ap-southeast-1, ap-southeast-2, ap-northeast-1, ca-central-1, eu-central-1, eu-west-1, us-gov-west-1, cn-north-1, Edge) | jsonData |  | Specify the region, such as for US West (Oregon) use `us-west-2` as the region. |
| `jsonData.edgeAuthMode` | enum (default, linux, ldap) | jsonData |  | Authentication Mode |
| `jsonData.edgeAuthUser` | string | jsonData | conditional | The username set to local authentication proxy |
| `secureJsonData.edgeAuthPass` 🔒 | string | secureJsonData | conditional | The password sent to local authentication proxy |
| `secureJsonData.cert` 🔒 | string (multiline) | secureJsonData | conditional | Certificate for SSL enabled authentication. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Workspace IAM Role (`ec2_iam_role`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: AWS IoT SiteWise
    type: grafana-iot-sitewise-datasource
    access: proxy
    jsonData:
      authType: ec2_iam_role
      edgeAuthMode: default
      edgeAuthUser: "<YOUR_USERNAME>"
      endpoint: "https://{service}.{region}.amazonaws.com"
    secureJsonData:
      cert: "<YOUR_SSL_CERTIFICATE>"
      edgeAuthPass: "<YOUR_PASSWORD>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_iot_sitewise_datasource_ec2_iam_role" {
  type = "grafana-iot-sitewise-datasource"
  name = "AWS IoT SiteWise"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "ec2_iam_role"
    edgeAuthMode = "default"
    edgeAuthUser = "<YOUR_USERNAME>"
    endpoint = "https://{service}.{region}.amazonaws.com"
  })

  secure_json_data_encoded = jsonencode({
    cert = "<YOUR_SSL_CERTIFICATE>"
    edgeAuthPass = "<YOUR_PASSWORD>"
  })
}
```

### Grafana Assume Role (`grafana_assume_role`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: AWS IoT SiteWise
    type: grafana-iot-sitewise-datasource
    access: proxy
    jsonData:
      authType: grafana_assume_role
      edgeAuthMode: default
      edgeAuthUser: "<YOUR_USERNAME>"
      endpoint: "https://{service}.{region}.amazonaws.com"
    secureJsonData:
      cert: "<YOUR_SSL_CERTIFICATE>"
      edgeAuthPass: "<YOUR_PASSWORD>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_iot_sitewise_datasource_grafana_assume_role" {
  type = "grafana-iot-sitewise-datasource"
  name = "AWS IoT SiteWise"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "grafana_assume_role"
    edgeAuthMode = "default"
    edgeAuthUser = "<YOUR_USERNAME>"
    endpoint = "https://{service}.{region}.amazonaws.com"
  })

  secure_json_data_encoded = jsonencode({
    cert = "<YOUR_SSL_CERTIFICATE>"
    edgeAuthPass = "<YOUR_PASSWORD>"
  })
}
```

### AWS SDK Default (`default`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: AWS IoT SiteWise
    type: grafana-iot-sitewise-datasource
    access: proxy
    jsonData:
      authType: default
      edgeAuthMode: default
      edgeAuthUser: "<YOUR_USERNAME>"
      endpoint: "https://{service}.{region}.amazonaws.com"
    secureJsonData:
      cert: "<YOUR_SSL_CERTIFICATE>"
      edgeAuthPass: "<YOUR_PASSWORD>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_iot_sitewise_datasource_default" {
  type = "grafana-iot-sitewise-datasource"
  name = "AWS IoT SiteWise"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "default"
    edgeAuthMode = "default"
    edgeAuthUser = "<YOUR_USERNAME>"
    endpoint = "https://{service}.{region}.amazonaws.com"
  })

  secure_json_data_encoded = jsonencode({
    cert = "<YOUR_SSL_CERTIFICATE>"
    edgeAuthPass = "<YOUR_PASSWORD>"
  })
}
```

### Access & secret key (`keys`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: AWS IoT SiteWise
    type: grafana-iot-sitewise-datasource
    access: proxy
    jsonData:
      authType: keys
      edgeAuthMode: default
      edgeAuthUser: "<YOUR_USERNAME>"
      endpoint: "https://{service}.{region}.amazonaws.com"
    secureJsonData:
      accessKey: "<YOUR_ACCESS_KEY_ID>"
      cert: "<YOUR_SSL_CERTIFICATE>"
      edgeAuthPass: "<YOUR_PASSWORD>"
      secretKey: "<YOUR_SECRET_ACCESS_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_iot_sitewise_datasource_keys" {
  type = "grafana-iot-sitewise-datasource"
  name = "AWS IoT SiteWise"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "keys"
    edgeAuthMode = "default"
    edgeAuthUser = "<YOUR_USERNAME>"
    endpoint = "https://{service}.{region}.amazonaws.com"
  })

  secure_json_data_encoded = jsonencode({
    accessKey = "<YOUR_ACCESS_KEY_ID>"
    cert = "<YOUR_SSL_CERTIFICATE>"
    edgeAuthPass = "<YOUR_PASSWORD>"
    secretKey = "<YOUR_SECRET_ACCESS_KEY>"
  })
}
```

### Credentials file (`credentials`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: AWS IoT SiteWise
    type: grafana-iot-sitewise-datasource
    access: proxy
    jsonData:
      authType: credentials
      edgeAuthMode: default
      edgeAuthUser: "<YOUR_USERNAME>"
      endpoint: "https://{service}.{region}.amazonaws.com"
    secureJsonData:
      cert: "<YOUR_SSL_CERTIFICATE>"
      edgeAuthPass: "<YOUR_PASSWORD>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_iot_sitewise_datasource_credentials" {
  type = "grafana-iot-sitewise-datasource"
  name = "AWS IoT SiteWise"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "credentials"
    edgeAuthMode = "default"
    edgeAuthUser = "<YOUR_USERNAME>"
    endpoint = "https://{service}.{region}.amazonaws.com"
  })

  secure_json_data_encoded = jsonencode({
    cert = "<YOUR_SSL_CERTIFICATE>"
    edgeAuthPass = "<YOUR_PASSWORD>"
  })
}
```

