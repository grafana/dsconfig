package mqttdatasource

import (
	_ "embed"

	"k8s.io/kube-openapi/pkg/spec3"

	"github.com/grafana/dsconfig/dsconfig"
	sdkschema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

//go:generate go test -generateArtifacts -run TestSchemaConformance ./...

// TargetAPIVersion is the API version this schema applies to.
const TargetAPIVersion = dsconfig.TargetAPIVersion

// configSchemaJSON is the declarative dsconfig schema — the single source
// of truth for the MQTT datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig
// schema (single source of truth) for the MQTT datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the MQTT
// datasource: the settings (configuration) spec derived from dsconfig.json,
// the secure values, and example configurations, stamped with
// TargetAPIVersion. Grafana's datasource API server serves this bundle as
// {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the MQTT
// datasource, covering the default configuration and each authentication /
// connection variant the config editor supports. Each example value is a
// full instance-settings object with the plugin configuration nested under
// jsonData and the relevant write-only secrets under secureJsonData
// (placeholder values — replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: every field at its Go zero value. jsonData.uri is empty — LoadConfig.Validate will reject this until a broker URI is supplied. The password placeholder is empty because secureJsonData values are write-only.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"tlsAuth":           false,
							"tlsAuthWithCACert": false,
							"tlsSkipVerify":     false,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "",
						},
					},
				},
			},
			"anonymousTCP": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Anonymous TCP broker",
					Description: "Connect to a broker that does not require MQTT-level auth over plaintext TCP. Only jsonData.uri is set; the empty secureJsonData.password placeholder records that no password is configured.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"uri":               "tcp://broker.example.com:1883",
							"tlsAuth":           false,
							"tlsAuthWithCACert": false,
							"tlsSkipVerify":     false,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "",
						},
					},
				},
			},
			"basicAuthTCP": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Basic auth over TCP",
					Description: "MQTT basic auth with username in jsonData and password in secureJsonData. The backend calls paho.SetUsername / paho.SetPassword only when both are non-empty (pkg/mqtt/client.go:54-60).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"uri":               "tcp://broker.example.com:1883",
							"username":          "grafana",
							"tlsAuth":           false,
							"tlsAuthWithCACert": false,
							"tlsSkipVerify":     false,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "s3cret",
						},
					},
				},
			},
			"tlsClientAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "TLS client certificate authentication",
					Description: "Mutual TLS. jsonData.tlsAuth toggles the visibility of the TLS Client Certificate / TLS Client Key inputs in the editor, but the backend loads the keypair (pkg/mqtt/client.go:66-73) purely based on whether the secrets are non-empty. Both tlsClientCert and tlsClientKey must be provided.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"uri":               "tls://broker.example.com:8883",
							"tlsAuth":           true,
							"tlsAuthWithCACert": false,
							"tlsSkipVerify":     false,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyTLSClientCert): "-----BEGIN CERTIFICATE-----\nMIIB...clientcert...==\n-----END CERTIFICATE-----",
							string(SecureJsonDataKeyTLSClientKey):  "-----BEGIN RSA PRIVATE KEY-----\nMIIE...clientkey...==\n-----END RSA PRIVATE KEY-----",
						},
					},
				},
			},
			"selfSignedCA": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Self-signed server with custom CA",
					Description: "TLS-encrypted broker whose server certificate is signed by a private CA. jsonData.tlsAuthWithCACert gates the editor input; the backend loads the CA cert pool (pkg/mqtt/client.go:75-79) whenever secureJsonData.tlsCACert is non-empty.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"uri":               "tls://broker.internal.corp:8883",
							"tlsAuth":           false,
							"tlsAuthWithCACert": true,
							"tlsSkipVerify":     false,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyTLSCACert): "-----BEGIN CERTIFICATE-----\nMIIC...cacert...==\n-----END CERTIFICATE-----",
						},
					},
				},
			},
			"mutualTLSWithCA": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Mutual TLS with custom CA and basic auth",
					Description: "Everything at once: private CA + client keypair + MQTT basic-auth username/password. Demonstrates that the three auth mechanisms compose freely and are gated only by which secrets are non-empty (pkg/mqtt/client.go:54-79).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"uri":               "tls://broker.internal.corp:8883",
							"clientID":          "grafana-mqtt-1",
							"username":          "grafana",
							"tlsAuth":           true,
							"tlsAuthWithCACert": true,
							"tlsSkipVerify":     false,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword):      "s3cret",
							string(SecureJsonDataKeyTLSCACert):     "-----BEGIN CERTIFICATE-----\nMIIC...cacert...==\n-----END CERTIFICATE-----",
							string(SecureJsonDataKeyTLSClientCert): "-----BEGIN CERTIFICATE-----\nMIIB...clientcert...==\n-----END CERTIFICATE-----",
							string(SecureJsonDataKeyTLSClientKey):  "-----BEGIN RSA PRIVATE KEY-----\nMIIE...clientkey...==\n-----END RSA PRIVATE KEY-----",
						},
					},
				},
			},
			"tlsSkipVerify": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "TLS broker with certificate verification disabled",
					Description: "Bypass server certificate verification via jsonData.tlsSkipVerify=true, which sets tls.Config.InsecureSkipVerify at pkg/mqtt/client.go:62-64. Only appropriate for trusted infrastructure — do NOT enable against public brokers.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"uri":               "tls://broker.example.com:8883",
							"tlsAuth":           false,
							"tlsAuthWithCACert": false,
							"tlsSkipVerify":     true,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "",
						},
					},
				},
			},
			"webSocket": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "WebSocket transport",
					Description: "MQTT over WebSocket. paho maps the ws:// scheme to default port 80 and wss:// to 443 (pkg/mqtt/proxy.go:66-80) when no port is specified — this example uses wss:// on the default port for MQTT-over-WSS.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"uri":               "wss://broker.example.com:443/mqtt",
							"tlsAuth":           false,
							"tlsAuthWithCACert": false,
							"tlsSkipVerify":     false,
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "",
						},
					},
				},
			},
		},
	}
}
