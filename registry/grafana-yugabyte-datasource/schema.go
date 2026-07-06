package yugabytedatasource

import (
	_ "embed"

	"k8s.io/kube-openapi/pkg/spec3"

	"github.com/grafana/dsconfig/dsconfig"
	sdkschema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

//go:generate go test -generateArtifacts -run TestSchemaConformance ./...

// TargetAPIVersion is the API version this schema applies to.
const TargetAPIVersion = dsconfig.TargetAPIVersion

// configSchemaJSON is the declarative dsconfig schema — the single source of
// truth for the Yugabyte datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Yugabyte datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Yugabyte
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Yugabyte
// datasource. The editor has a single fixed authentication path (username +
// password over the PostgreSQL wire protocol), so the variants below cover
// the empty default plus realistic connection payloads.
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The empty defaults a new datasource starts with. The user must supply url ('host:port'), user, jsonData.database, and secureJsonData.password before the datasource will connect.",
					Value: map[string]any{
						"url":  "",
						"user": "",
						"jsonData": map[string]any{
							"database": "",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "",
						},
					},
				},
			},
			"localDev": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Local YugabyteDB (default port 5433)",
					Description: "Typical local development setup pointing at a YugabyteDB YSQL node on the default port. TLS is not configurable; the backend hardcodes sslmode='allow' (pkg/settings.go:52).",
					Value: map[string]any{
						"url":  "localhost:5433",
						"user": "yugabyte",
						"jsonData": map[string]any{
							"database": "yb_demo",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "yugabyte",
						},
					},
				},
			},
			"remoteCluster": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Remote YugabyteDB cluster",
					Description: "Points at a remote YSQL endpoint (still 'host:port', no scheme). Whether the connection is encrypted depends entirely on the server: libpq's 'allow' mode tries plaintext first and falls back to TLS only if the server rejects the plaintext handshake.",
					Value: map[string]any{
						"url":  "yb.internal.example.com:5433",
						"user": "grafana_reader",
						"jsonData": map[string]any{
							"database": "metrics",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"secureSocksProxy": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "With Secure Socks Proxy enabled",
					Description: "The editor's SecureSocksProxyToggle (ConfigEditor.tsx:83) writes jsonData.enableSecureSocksProxy=true. The backend then routes the pgx connection through the configured secure socks proxy (pkg/driver.go:41-53). This field is stored in jsonData but is intentionally excluded from the dsconfig schema per repo policy.",
					Value: map[string]any{
						"url":  "yb.internal.example.com:5433",
						"user": "grafana_reader",
						"jsonData": map[string]any{
							"database":               "metrics",
							"enableSecureSocksProxy": true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
		},
	}
}
