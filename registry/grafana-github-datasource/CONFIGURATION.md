# GitHub configuration

Configuration reference for the **GitHub** data source (`grafana-github-datasource`) in Grafana.

For more information, see the [official documentation](https://grafana.com/docs/plugins/grafana-github-datasource).

> Generated from [`dsconfig.json`](dsconfig.json). Do not edit by hand — run `go generate ./...` to refresh.

## Fields

| Field | Type | Target | Required | Description |
|---|---|---|---|---|
| `jsonData.githubPlan` | enum (github-basic, github-enterprise-cloud, github-enterprise-server) | jsonData |  |  |
| `jsonData.githubUrl` | string | jsonData | conditional | GitHub Enterprise Server URL |
| `jsonData.selectedAuthType` | enum (personal-access-token, github-app) | jsonData |  | Authentication Type |
| `secureJsonData.accessToken` 🔒 | string | secureJsonData | conditional | Personal Access Token |
| `jsonData.appId` | string | jsonData | conditional | App ID |
| `jsonData.installationId` | string | jsonData | conditional | Installation ID |
| `secureJsonData.privateKey` 🔒 | string (multiline) | secureJsonData | conditional | Private Key |
| `jsonData.cachingEnabled` | boolean | jsonData |  | Enables the query caching wrapper in the plugin backend. Not exposed in the configuration editor; the backend currently enables caching for every datasource instance. |

## Provisioning examples

Each scenario below shows how to provision the data source in Grafana using a YAML file (loaded by Grafana's [file provisioner](https://grafana.com/docs/grafana/latest/administration/provisioning/#data-sources)) and using the [Grafana Terraform provider](https://registry.terraform.io/providers/grafana/grafana/latest/docs/resources/data_source).

Placeholders like `<YOUR_TOKEN>` must be replaced with real values before use.

### Personal Access Token (`personal-access-token`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: GitHub
    type: grafana-github-datasource
    access: proxy
    jsonData:
      cachingEnabled: true
      githubUrl: "http(s)://HOSTNAME/"
      selectedAuthType: personal-access-token
    secureJsonData:
      accessToken: "<YOUR_PERSONAL_ACCESS_TOKEN>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_github_datasource_personal_access_token" {
  type = "grafana-github-datasource"
  name = "GitHub"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    cachingEnabled = true
    githubUrl = "http(s)://HOSTNAME/"
    selectedAuthType = "personal-access-token"
  })

  secure_json_data_encoded = jsonencode({
    accessToken = "<YOUR_PERSONAL_ACCESS_TOKEN>"
  })
}
```

### GitHub App (`github-app`)

**Grafana provisioning YAML**

```yaml
apiVersion: 1
datasources:
  - name: GitHub
    type: grafana-github-datasource
    access: proxy
    jsonData:
      appId: App ID
      cachingEnabled: true
      githubUrl: "http(s)://HOSTNAME/"
      installationId: Installation ID
      selectedAuthType: github-app
    secureJsonData:
      privateKey: "<YOUR_PRIVATE_KEY>"
```

**Terraform**

```hcl
resource "grafana_data_source" "grafana_github_datasource_github_app" {
  type = "grafana-github-datasource"
  name = "GitHub"
  url = "https://example.com"

  json_data_encoded = jsonencode({
    appId = "App ID"
    cachingEnabled = true
    githubUrl = "http(s)://HOSTNAME/"
    installationId = "Installation ID"
    selectedAuthType = "github-app"
  })

  secure_json_data_encoded = jsonencode({
    privateKey = "<YOUR_PRIVATE_KEY>"
  })
}
```

