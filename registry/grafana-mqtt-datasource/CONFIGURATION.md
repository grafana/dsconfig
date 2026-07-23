# MQTT configuration

Configuration reference for the **MQTT** data source (`grafana-mqtt-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-mqtt-datasource/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.uri` | string | jsonData | yes | URI |
| `jsonData.clientID` | string | jsonData |  | If not set, a random client ID is used. |
| `jsonData.username` | string | jsonData |  | Username |
| `secureJsonData.password` 🔒 | string | secureJsonData |  | Password |
| `jsonData.tlsAuth` | boolean | jsonData |  | Enables TLS authentication using client cert configured in secure json data. |
| `jsonData.tlsSkipVerify` | boolean | jsonData |  | When enabled, skips verification of the MQTT server's TLS certificate chain and host name. |
| `jsonData.tlsAuthWithCACert` | boolean | jsonData |  | Needed for verifying servers with self-signed TLS Certs. |
| `secureJsonData.tlsCACert` 🔒 | string (multiline) | secureJsonData |  | If a Certificate Authority certificate is required to verify the server's certificate, provide it here. |
| `secureJsonData.tlsClientCert` 🔒 | string (multiline) | secureJsonData |  | To authenticate with an TLS client certificate, provide the client certificate here. |
| `secureJsonData.tlsClientKey` 🔒 | string (multiline) | secureJsonData |  | To authenticate with a client TLS certificate, provide the private key here. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: MQTT
    type: grafana-mqtt-datasource
    access: proxy
    jsonData:
      tlsAuth: false
      tlsAuthWithCACert: false
      tlsSkipVerify: false
      uri: "TCP (tcp://), TLS (tls://), or WebSocket (ws://)"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_mqtt_datasource" {
  type = "grafana-mqtt-datasource"
  name = "MQTT"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    tlsAuth = false
    tlsAuthWithCACert = false
    tlsSkipVerify = false
    uri = "TCP (tcp://), TLS (tls://), or WebSocket (ws://)"
  })
}
```

