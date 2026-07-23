# Amazon Aurora configuration

Configuration reference for the **Amazon Aurora** data source (`grafana-aurora-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-aurora-datasource).

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
| `jsonData.defaultRegion` | enum | jsonData |  | Specify the region, such as for US West (Oregon) use `us-west-2` as the region. |
| `jsonData.engine` | enum (aurora-postgres, aurora-mysql) | jsonData |  | Engine |
| `jsonData.dbName` | string | jsonData |  | Database Name |
| `jsonData.dbUser` | string | jsonData | yes | Database User |
| `jsonData.dbHost` | string | jsonData | yes | You can connect to all the read replicas on your Amazon Aurora cluster through a single reader endpoint. Read more about it here (https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/Aurora.Overview.Endpoints.html#Aurora.Endpoints.Reader). |
| `jsonData.dbPort` | number | jsonData | yes | Database Port |
| `jsonData.dbHostAuth` | string | jsonData |  | Optional, if not provided, the dbHost above will be used. |
| `jsonData.dbPortAuth` | number | jsonData |  | Optional, if not provided, the dbPort above will be used. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Workspace IAM Role (`ec2_iam_role`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Aurora
    type: grafana-aurora-datasource
    access: proxy
    jsonData:
      authType: ec2_iam_role
      dbHost: Host
      dbPort: 0
      dbUser: postgres
      engine: aurora-postgres
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_aurora_datasource_ec2_iam_role" {
  type = "grafana-aurora-datasource"
  name = "Amazon Aurora"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "ec2_iam_role"
    dbHost = "Host"
    dbPort = 0
    dbUser = "postgres"
    engine = "aurora-postgres"
  })
}
```

### Grafana Assume Role (`grafana_assume_role`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Aurora
    type: grafana-aurora-datasource
    access: proxy
    jsonData:
      authType: grafana_assume_role
      dbHost: Host
      dbPort: 0
      dbUser: postgres
      engine: aurora-postgres
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_aurora_datasource_grafana_assume_role" {
  type = "grafana-aurora-datasource"
  name = "Amazon Aurora"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "grafana_assume_role"
    dbHost = "Host"
    dbPort = 0
    dbUser = "postgres"
    engine = "aurora-postgres"
  })
}
```

### AWS SDK Default (`default`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Aurora
    type: grafana-aurora-datasource
    access: proxy
    jsonData:
      authType: default
      dbHost: Host
      dbPort: 0
      dbUser: postgres
      engine: aurora-postgres
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_aurora_datasource_default" {
  type = "grafana-aurora-datasource"
  name = "Amazon Aurora"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "default"
    dbHost = "Host"
    dbPort = 0
    dbUser = "postgres"
    engine = "aurora-postgres"
  })
}
```

### Access & secret key (`keys`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Amazon Aurora
    type: grafana-aurora-datasource
    access: proxy
    jsonData:
      authType: keys
      dbHost: Host
      dbPort: 0
      dbUser: postgres
      engine: aurora-postgres
    secureJsonData:
      accessKey: "<YOUR_ACCESS_KEY_ID>"
      secretKey: "<YOUR_SECRET_ACCESS_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_aurora_datasource_keys" {
  type = "grafana-aurora-datasource"
  name = "Amazon Aurora"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "keys"
    dbHost = "Host"
    dbPort = 0
    dbUser = "postgres"
    engine = "aurora-postgres"
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
  - name: Amazon Aurora
    type: grafana-aurora-datasource
    access: proxy
    jsonData:
      authType: credentials
      dbHost: Host
      dbPort: 0
      dbUser: postgres
      engine: aurora-postgres
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_aurora_datasource_credentials" {
  type = "grafana-aurora-datasource"
  name = "Amazon Aurora"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authType = "credentials"
    dbHost = "Host"
    dbPort = 0
    dbUser = "postgres"
    engine = "aurora-postgres"
  })
}
```

