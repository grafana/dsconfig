# Sentry configuration

Configuration reference for the **Sentry** data source (`grafana-sentry-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/grafana/plugins/grafana-sentry-datasource/).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.url` | string | jsonData |  | Sentry URL to be used. If left blank, https://sentry.io will be used |
| `jsonData.orgSlug` | string | jsonData | yes | Sentry Org slug. Typically this will be the last segment of the URL: https://sentry.io/organizations/{organization_slug}/ - only the slug should be entered here |
| `secureJsonData.authToken` 🔒 | string | secureJsonData | yes | Sentry authentication token. Auth tokens can be created from https://sentry.io/settings/{organization_slug}/developer-settings |
| `jsonData.tlsSkipVerify` | boolean | jsonData |  | Skip TLS certificate verification. Use this option for self-hosted Sentry instances with self-signed certificates. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Default configuration

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: Sentry
    type: grafana-sentry-datasource
    access: proxy
    jsonData:
      orgSlug: Sentry org slug
      tlsSkipVerify: false
      url: "https://sentry.io"
    secureJsonData:
      authToken: "<YOUR_SENTRY_AUTH_TOKEN>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_sentry_datasource" {
  type = "grafana-sentry-datasource"
  name = "Sentry"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    orgSlug = "Sentry org slug"
    tlsSkipVerify = false
    url = "https://sentry.io"
  })

  secure_json_data_encoded = jsonencode({
    authToken = "<YOUR_SENTRY_AUTH_TOKEN>"
  })
}
```

