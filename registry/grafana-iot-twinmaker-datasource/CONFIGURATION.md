# AWS IoT TwinMaker configuration

Configuration reference for the **AWS IoT TwinMaker** data source (`grafana-iot-twinmaker-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-iot-twinmaker-app/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.authType` | enum (ec2_iam_role, default, keys, credentials) | jsonData |  | Specify which AWS credentials chain to use. |
| `jsonData.profile` | string | jsonData |  | Credentials profile name, as specified in ~/.aws/credentials, leave blank for default. |
| `secureJsonData.accessKey` 🔒 | string | secureJsonData | conditional | Access Key ID |
| `secureJsonData.secretKey` 🔒 | string | secureJsonData | conditional | Secret Access Key |
| `jsonData.assumeRoleArn` | string | jsonData | yes | Optional. Specifying the ARN of a role will ensure that the                      selected authentication provider is used to assume the role rather than the                      credentials directly. |
| `jsonData.externalId` | string | jsonData |  | If you are assuming a role in another account, that has been created with an external ID, specify the external ID here. |
| `jsonData.endpoint` | string | jsonData |  | Optionally, specify a custom endpoint for the service |
| `jsonData.defaultRegion` | enum (ap-south-1, ap-northeast-1, ap-northeast-2, ap-southeast-1, ap-southeast-2, eu-central-1, eu-west-1, us-east-1, us-west-2, us-gov-west-1, cn-north-1) | jsonData |  | Specify the region, such as for US West (Oregon) use ` us-west-2 ` as the region. |
| `jsonData.workspaceId` | enum | jsonData | yes | Workspace |
| `jsonData.assumeRoleArnWriter` | string | jsonData |  | Specify the ARN of a role to assume when writing property values in IoT TwinMaker |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Workspace IAM Role (`ec2_iam_role`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: AWS IoT TwinMaker
    type: grafana-iot-twinmaker-datasource
    access: proxy
    jsonData:
      assumeRoleArn: "arn:aws:iam:*"
      authType: ec2_iam_role
      defaultRegion: us-east-1
      workspaceId: Select a workspace
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_iot_twinmaker_datasource_ec2_iam_role" {
  type = "grafana-iot-twinmaker-datasource"
  name = "AWS IoT TwinMaker"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    assumeRoleArn = "arn:aws:iam:*"
    authType = "ec2_iam_role"
    defaultRegion = "us-east-1"
    workspaceId = "Select a workspace"
  })
}
```

### AWS SDK Default (`default`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: AWS IoT TwinMaker
    type: grafana-iot-twinmaker-datasource
    access: proxy
    jsonData:
      assumeRoleArn: "arn:aws:iam:*"
      authType: default
      defaultRegion: us-east-1
      workspaceId: Select a workspace
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_iot_twinmaker_datasource_default" {
  type = "grafana-iot-twinmaker-datasource"
  name = "AWS IoT TwinMaker"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    assumeRoleArn = "arn:aws:iam:*"
    authType = "default"
    defaultRegion = "us-east-1"
    workspaceId = "Select a workspace"
  })
}
```

### Access & secret key (`keys`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: AWS IoT TwinMaker
    type: grafana-iot-twinmaker-datasource
    access: proxy
    jsonData:
      assumeRoleArn: "arn:aws:iam:*"
      authType: keys
      defaultRegion: us-east-1
      workspaceId: Select a workspace
    secureJsonData:
      accessKey: "<YOUR_ACCESS_KEY_ID>"
      secretKey: "<YOUR_SECRET_ACCESS_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_iot_twinmaker_datasource_keys" {
  type = "grafana-iot-twinmaker-datasource"
  name = "AWS IoT TwinMaker"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    assumeRoleArn = "arn:aws:iam:*"
    authType = "keys"
    defaultRegion = "us-east-1"
    workspaceId = "Select a workspace"
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
  - name: AWS IoT TwinMaker
    type: grafana-iot-twinmaker-datasource
    access: proxy
    jsonData:
      assumeRoleArn: "arn:aws:iam:*"
      authType: credentials
      defaultRegion: us-east-1
      workspaceId: Select a workspace
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_iot_twinmaker_datasource_credentials" {
  type = "grafana-iot-twinmaker-datasource"
  name = "AWS IoT TwinMaker"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    assumeRoleArn = "arn:aws:iam:*"
    authType = "credentials"
    defaultRegion = "us-east-1"
    workspaceId = "Select a workspace"
  })
}
```

