# Amazon Redshift configuration

Configuration reference for the **Amazon Redshift** data source (`grafana-redshift-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-redshift-datasource/).

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
| `jsonData.useManagedSecret` | enum (false, true) | jsonData |  |  |
| `jsonData.useServerless` | boolean | jsonData |  | Serverless |
| `jsonData.clusterIdentifier` | enum | jsonData | conditional | Cluster Identifier |
| `jsonData.workgroupName` | enum | jsonData | conditional | Workgroup |
| `jsonData.managedSecret.arn` | enum | jsonData | conditional | Managed Secret |
| `jsonData.managedSecret.name` | string | jsonData |  |  |
| `jsonData.dbUser` | string | jsonData | conditional | Database User |
| `jsonData.database` | string | jsonData | yes | Database |
| `jsonData.withEvent` | boolean | jsonData |  | Send events to Amazon EventBridge |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Workspace IAM Role (`ec2_iam_role`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Redshift
    type: grafana-redshift-datasource
    access: proxy
    jsonData:
      authType: ec2_iam_role
      clusterIdentifier: "<YOUR_CLUSTER_IDENTIFIER>"
      database: "<YOUR_DATABASE>"
      dbUser: "<YOUR_DATABASE_USER>"
      managedSecret:
        arn: "<YOUR_MANAGED_SECRET>"
      useManagedSecret: false
      useServerless: false
      withEvent: false
      workgroupName: "<YOUR_WORKGROUP>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_redshift_datasource_ec2_iam_role" {
  type = "grafana-redshift-datasource"
  name = "Amazon Redshift"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "ec2_iam_role"
    clusterIdentifier = "<YOUR_CLUSTER_IDENTIFIER>"
    database = "<YOUR_DATABASE>"
    dbUser = "<YOUR_DATABASE_USER>"
    managedSecret = {
      arn = "<YOUR_MANAGED_SECRET>"
    }
    useManagedSecret = false
    useServerless = false
    withEvent = false
    workgroupName = "<YOUR_WORKGROUP>"
  })
}
```

### Grafana Assume Role (`grafana_assume_role`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Redshift
    type: grafana-redshift-datasource
    access: proxy
    jsonData:
      authType: grafana_assume_role
      clusterIdentifier: "<YOUR_CLUSTER_IDENTIFIER>"
      database: "<YOUR_DATABASE>"
      dbUser: "<YOUR_DATABASE_USER>"
      managedSecret:
        arn: "<YOUR_MANAGED_SECRET>"
      useManagedSecret: false
      useServerless: false
      withEvent: false
      workgroupName: "<YOUR_WORKGROUP>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_redshift_datasource_grafana_assume_role" {
  type = "grafana-redshift-datasource"
  name = "Amazon Redshift"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "grafana_assume_role"
    clusterIdentifier = "<YOUR_CLUSTER_IDENTIFIER>"
    database = "<YOUR_DATABASE>"
    dbUser = "<YOUR_DATABASE_USER>"
    managedSecret = {
      arn = "<YOUR_MANAGED_SECRET>"
    }
    useManagedSecret = false
    useServerless = false
    withEvent = false
    workgroupName = "<YOUR_WORKGROUP>"
  })
}
```

### AWS SDK Default (`default`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Redshift
    type: grafana-redshift-datasource
    access: proxy
    jsonData:
      authType: default
      clusterIdentifier: "<YOUR_CLUSTER_IDENTIFIER>"
      database: "<YOUR_DATABASE>"
      dbUser: "<YOUR_DATABASE_USER>"
      managedSecret:
        arn: "<YOUR_MANAGED_SECRET>"
      useManagedSecret: false
      useServerless: false
      withEvent: false
      workgroupName: "<YOUR_WORKGROUP>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_redshift_datasource_default" {
  type = "grafana-redshift-datasource"
  name = "Amazon Redshift"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "default"
    clusterIdentifier = "<YOUR_CLUSTER_IDENTIFIER>"
    database = "<YOUR_DATABASE>"
    dbUser = "<YOUR_DATABASE_USER>"
    managedSecret = {
      arn = "<YOUR_MANAGED_SECRET>"
    }
    useManagedSecret = false
    useServerless = false
    withEvent = false
    workgroupName = "<YOUR_WORKGROUP>"
  })
}
```

### Access & secret key (`keys`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Redshift
    type: grafana-redshift-datasource
    access: proxy
    jsonData:
      authType: keys
      clusterIdentifier: "<YOUR_CLUSTER_IDENTIFIER>"
      database: "<YOUR_DATABASE>"
      dbUser: "<YOUR_DATABASE_USER>"
      managedSecret:
        arn: "<YOUR_MANAGED_SECRET>"
      useManagedSecret: false
      useServerless: false
      withEvent: false
      workgroupName: "<YOUR_WORKGROUP>"
    secureJsonData:
      accessKey: "<YOUR_ACCESS_KEY_ID>"
      secretKey: "<YOUR_SECRET_ACCESS_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_redshift_datasource_keys" {
  type = "grafana-redshift-datasource"
  name = "Amazon Redshift"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "keys"
    clusterIdentifier = "<YOUR_CLUSTER_IDENTIFIER>"
    database = "<YOUR_DATABASE>"
    dbUser = "<YOUR_DATABASE_USER>"
    managedSecret = {
      arn = "<YOUR_MANAGED_SECRET>"
    }
    useManagedSecret = false
    useServerless = false
    withEvent = false
    workgroupName = "<YOUR_WORKGROUP>"
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
  - name: Amazon Redshift
    type: grafana-redshift-datasource
    access: proxy
    jsonData:
      authType: credentials
      clusterIdentifier: "<YOUR_CLUSTER_IDENTIFIER>"
      database: "<YOUR_DATABASE>"
      dbUser: "<YOUR_DATABASE_USER>"
      managedSecret:
        arn: "<YOUR_MANAGED_SECRET>"
      useManagedSecret: false
      useServerless: false
      withEvent: false
      workgroupName: "<YOUR_WORKGROUP>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_redshift_datasource_credentials" {
  type = "grafana-redshift-datasource"
  name = "Amazon Redshift"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "credentials"
    clusterIdentifier = "<YOUR_CLUSTER_IDENTIFIER>"
    database = "<YOUR_DATABASE>"
    dbUser = "<YOUR_DATABASE_USER>"
    managedSecret = {
      arn = "<YOUR_MANAGED_SECRET>"
    }
    useManagedSecret = false
    useServerless = false
    withEvent = false
    workgroupName = "<YOUR_WORKGROUP>"
  })
}
```

