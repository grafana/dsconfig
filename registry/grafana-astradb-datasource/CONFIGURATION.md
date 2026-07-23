# AstraDB configuration

Configuration reference for the **AstraDB** data source (`grafana-astradb-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-astradb-datasource/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.authKind` | enum (0, 1) | jsonData |  | Authentication |
| `jsonData.uri` | string | jsonData | conditional | URI |
| `secureJsonData.token` 🔒 | string | secureJsonData | conditional | Token |
| `jsonData.grpcEndpoint` | string | jsonData | conditional | GRPC Endpoint |
| `jsonData.authEndpoint` | string | jsonData | conditional | Auth Endpoint |
| `jsonData.user` | string | jsonData | conditional | User Name |
| `secureJsonData.password` 🔒 | string | secureJsonData | conditional | Password |
| `jsonData.secure` | boolean | jsonData |  | Secure |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Token (`0`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: AstraDB
    type: grafana-astradb-datasource
    access: proxy
    jsonData:
      authEndpoint: "localhost:8081"
      authKind: "0"
      grpcEndpoint: "localhost:8090"
      secure: false
      uri: "$ASTRA_CLUSTER_ID-$ASTRA_REGION.apps.astra.datastax.com:443"
      user: "localhost:8090"
    secureJsonData:
      password: "<YOUR_PASSWORD>"
      token: "<YOUR_TOKEN>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_astradb_datasource_0" {
  type = "grafana-astradb-datasource"
  name = "AstraDB"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authEndpoint = "localhost:8081"
    authKind = "0"
    grpcEndpoint = "localhost:8090"
    secure = false
    uri = "$ASTRA_CLUSTER_ID-$ASTRA_REGION.apps.astra.datastax.com:443"
    user = "localhost:8090"
  })

  secure_json_data_encoded = jsonencode({
    password = "<YOUR_PASSWORD>"
    token = "<YOUR_TOKEN>"
  })
}
```

### Credentials (`1`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: AstraDB
    type: grafana-astradb-datasource
    access: proxy
    jsonData:
      authEndpoint: "localhost:8081"
      authKind: "1"
      grpcEndpoint: "localhost:8090"
      secure: false
      uri: "$ASTRA_CLUSTER_ID-$ASTRA_REGION.apps.astra.datastax.com:443"
      user: "localhost:8090"
    secureJsonData:
      password: "<YOUR_PASSWORD>"
      token: "<YOUR_TOKEN>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_astradb_datasource_1" {
  type = "grafana-astradb-datasource"
  name = "AstraDB"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    authEndpoint = "localhost:8081"
    authKind = "1"
    grpcEndpoint = "localhost:8090"
    secure = false
    uri = "$ASTRA_CLUSTER_ID-$ASTRA_REGION.apps.astra.datastax.com:443"
    user = "localhost:8090"
  })

  secure_json_data_encoded = jsonencode({
    password = "<YOUR_PASSWORD>"
    token = "<YOUR_TOKEN>"
  })
}
```

