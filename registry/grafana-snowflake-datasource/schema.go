package snowflakedatasource

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
// truth for the Snowflake datasource configuration.
//
//go:embed dsconfig.json
var configSchemaJSON []byte

// ConfigSchema parses, resolves, and returns the declarative dsconfig schema
// (single source of truth) for the Snowflake datasource.
func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

// NewSchema returns the full k8s-style SDK plugin schema for the Snowflake
// datasource: the settings (configuration) spec derived from dsconfig.json, the
// secure values, and example configurations, stamped with TargetAPIVersion.
// Grafana's datasource API server serves this bundle as {TargetAPIVersion}.json.
func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

// SettingsExamples returns k8s-style example configurations for the Snowflake
// datasource, covering the default configuration and each authentication type
// plus a session-parameters variant. Each example value is a full instance
// settings object with the plugin configuration nested under jsonData and the
// relevant write-only secrets under secureJsonData (placeholder values —
// replace them with real secrets).
func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: password authentication. Only secureJsonData.password (empty here), plus jsonData.account and jsonData.username, need to be filled in to get a working datasource.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType": string(AuthTypePassword),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "",
						},
					},
				},
			},
			"passwordAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Password authentication",
					Description: "Authenticate with a Snowflake username and password. The password is provided in secureJsonData.password.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":  string(AuthTypePassword),
							"account":   "myorg-myaccount",
							"username":  "GRAFANA_READER",
							"role":      "GRAFANA_ROLE",
							"warehouse": "COMPUTE_WH",
							"database":  "ANALYTICS",
							"schema":    "PUBLIC",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"keyPair": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Key-pair authentication",
					Description: "Authenticate with an unencrypted RSA key pair (JWT). secureJsonData.privateKey is the PKCS#8 PEM including the BEGIN/END PRIVATE KEY lines.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":  string(AuthTypeKeyPair),
							"account":   "myorg-myaccount",
							"username":  "GRAFANA_READER",
							"warehouse": "COMPUTE_WH",
							"database":  "ANALYTICS",
							"schema":    "PUBLIC",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPrivateKey): examplePrivateKeyPEM,
						},
					},
				},
			},
			"keyPairEncrypted": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Key-pair authentication (encrypted key)",
					Description: "Authenticate with an encrypted RSA key pair. secureJsonData.privateKey is an ENCRYPTED PRIVATE KEY PEM and secureJsonData.privateKeyPassphrase is its passphrase.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":  string(AuthTypeKeyPair),
							"account":   "myorg-myaccount",
							"username":  "GRAFANA_READER",
							"warehouse": "COMPUTE_WH",
							"database":  "ANALYTICS",
							"schema":    "PUBLIC",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPrivateKey):           exampleEncryptedPrivateKeyPEM,
							string(SecureJsonDataKeyPrivateKeyPassphrase): "changeme",
						},
					},
				},
			},
			"programmaticAccessToken": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Programmatic Access Token (PAT)",
					Description: "Authenticate with a Snowflake programmatic access token provided in secureJsonData.patToken.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":  string(AuthTypePAT),
							"account":   "myorg-myaccount",
							"username":  "GRAFANA_READER",
							"warehouse": "COMPUTE_WH",
							"database":  "ANALYTICS",
							"schema":    "PUBLIC",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPATToken): "snowflake-pat-XXXXXXXXXXXX",
						},
					},
				},
			},
			"oauth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "OAuth (forward identity)",
					Description: "Authenticate by forwarding the user's upstream OAuth identity. Requires jsonData.oauthPassThru=true; no secret is stored (the access token is taken from the forwarded request at query time).",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":      string(AuthTypeOauth),
							"account":       "myorg-myaccount",
							"warehouse":     "COMPUTE_WH",
							"database":      "ANALYTICS",
							"schema":        "PUBLIC",
							"oauthPassThru": true,
						},
						"secureJsonData": map[string]any{},
					},
				},
			},
			"sessionParameters": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Password auth with session parameters and tuning",
					Description: "Password auth plus a jsonData.settings array of Snowflake session parameters, a plugin-side rowLimit, and login/request timeouts. Session parameters marked secure store their value in secureJsonData under a key equal to the setting name.",
					Value: map[string]any{
						"jsonData": map[string]any{
							"authType":       string(AuthTypePassword),
							"account":        "myorg-myaccount",
							"username":       "GRAFANA_READER",
							"warehouse":      "COMPUTE_WH",
							"database":       "ANALYTICS",
							"schema":         "PUBLIC",
							"loginTimeout":   30,
							"requestTimeout": 90,
							"rowLimit":       1000000,
							"settings": []map[string]any{
								{"name": "QUERY_TAG", "value": "grafana", "secure": false},
							},
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

// examplePrivateKeyPEM is a throwaway, unencrypted PKCS#8 RSA private key used
// only as a placeholder in the key-pair example. Replace it with a real key.
const examplePrivateKeyPEM = `-----BEGIN PRIVATE KEY-----
MIIEvwIBADANBgkqhkiG9w0BAQEFAASCBKkwggSlAgEAAoIBAQDsW28nNNl51DGc
YxMrmW/6FiFBWkECH1VF63yaDft/RVcgd039pWnQcwUx6Bcnblx6c9nf0dTspO3O
geEh5fv4p7HvUmeJkT2023MiYgxy2IqeGuP/Vd/wBdK9/9srTI6uQjnQyotAspjI
sQ8On51LpR8Au5Og81sEpWd0gQ05In6yciQJX2V+8aSqdxftrOGRYmx7UfTlM3yx
EuznUQ4z3bRVlLD7A2xxBT+1ZpHyBarMphwGDebG0L6WqHnk0d6AlPmpeO0Y/kc2
ChA6pdrH8HuJtwlgdHsjgmz+cWD+7hpjd4yQNls0KAgSnXQn/n77ew7Qu4Ka4LVy
Lhtl0O9zAgMBAAECggEANXKnvBHza91UKq0s1Jsp+z+t3X1vXe9/9QO/dYbb7Hfd
r/XKqSUSvJSGBDcbpUsLlpVEG5zzrV/Odvhf1K6RQDWLwza7Oxyg+5j0fD332rCl
CAPEsyTUMw7eDSEiirQRP86yDEkBHGxGqHuBkCkABO8eB6hjRe5CEtbkgi/8sYJw
GvLSXHnQhz62ht7La0P/Eqict+yn2UV6Lt6Fa6BpKca48970Iw1btV+gJcRToUu6
cO+dPSQdCJrNds1Jle2V0q6oQ7eSUvAaBnC8iq4DHdDEXucFPqNVjzDGdxXW27FQ
VZf8CxEUUpb00fwDE1wrOB9YTzdvGS66b8KGwxvbgQKBgQD+oh8B4Z+347S6d9t/
b9EgfYpwXOkfoBnHOmbbh1pSjlDKXg3vHPqo+94ugZcwFf1VGaU9aQnH0s1JXa/H
iakkQ62WKOFva+ZOShl/r5HRjcsu8CShhu5wwOKxEymlABFrFC3Y4tL2cHLJeBPc
22Saj43GkFoEYPx8uyDjeD9RowKBgQDtoDNlVBrvtnmM5BeVRxNWAJJHJcdt3ooR
bJJtplc/B3YQr1Korooh3UqgEYfGXaraSFphVF9/QqW6bchA+8CJCXz/0zuySKfQ
LiemzV268c3soaqn33SxsROUQq5gjinu90EbD5+EfDTOPNRD6VLRNTpRDS9tfoqK
sFUsOPnn8QKBgQDsu8lcNGoLyxYBruFRT6H8NPt6j8blcjHFOhTa0LI2wr12B4+o
2SZp6RCd6DmpqSgH8Hnh6EABmYjmRsXgG6o3XvyJ+KPutUA/VUDzp0VIsC1RDE7i
JdKU3Z9kxc4X60JGbVJarDc6iz0M9ihxUz/rOr+y9g3auFjFlixzjx4/RQKBgQDr
/KYuImmh2Ik8N6WIFY8JYQXkbIty/IgHp9h/1qtcqA9DoKopZTU/TmJ3NxGtGYa8
wxAnCsDQRKML000F5D8gmPCvq9rkQq2N3Nh6GgfUyaElOKSflRZyBZaZLeO5dlYE
wT7CHjDgRO24R8bSLtyVchQZPEv2pK338AiWI3tkMQKBgQC7e99+3GOUmEjlbuYw
u0neWJ//1SxUjVVvFCtqTHJfqKfi8RLqYaVxLVmtfQV8+JFqhpZRumPhIzl/c6Ka
MF+mwAvv1PaUqe8w26qZcp3c21uw76z8SqfWc9Gdt/p8A3qN7ZaPPLyb+loJU3FE
xvC4BjAdMQhnCly9mzST9KcfHQ==
-----END PRIVATE KEY-----
`

// exampleEncryptedPrivateKeyPEM is a throwaway, passphrase-encrypted PKCS#8 RSA
// private key (passphrase "changeme") used only as a placeholder in the
// encrypted key-pair example. Replace it with a real key.
const exampleEncryptedPrivateKeyPEM = `-----BEGIN ENCRYPTED PRIVATE KEY-----
MIIFNTBfBgkqhkiG9w0BBQ0wUjAxBgkqhkiG9w0BBQwwJAQQB9yTdeXRg6eG2IzI
s7gd2gICCAAwDAYIKoZIhvcNAgkFADAdBglghkgBZQMEASoEEBYoqUyN5bCx5KaP
+mTlWO0EggTQvRYs+1hfVAmAetx7nu0AbXevzkDC4KOA0Yt75stgR0xBr2B+GgP9
Z/pIrke2S/85MttrPuBHJoxXRvSUIF/gg+0YytptvIXC+2+FWCnjH0TWp78MwYXT
7afGdRu9azCj81VXr1/9+pIBmURLh1MuP6dQb+R+wG3u4i7LirDPUi5p/uosQ5Md
93HE5l7yoKFJDcoACkg2t27Q1nwQi5qL4oIABlWgPOygtGao2CUku96XsOtMdxR6
uOxe4U6KaQw9mCnFkFO+SYfmMJkOMhZuKQidFTSk6cVo08eGfVUqFwCzYmeVPJTc
XQy9587uQT/Gn1OkHuXddrsQajhWgoLKeQQ6zptxiCU3zi6lNQ87nBb7Y3MLOjjG
qMD3+N3WyCmM+vrl2ZisOTo0vSnR3BXZ98L2i4X4dOQ2nOnbR7GVKWPEFmg/GO1G
LtfcPRHrkkKhBRMTuZZicbMfMV21oKsIXE3ROD3HqZmda62QXvpZ6ACRhQyKMUbV
Xvaw1HijNmzaZihlHBP8Ji1Fdbaw5lzTp2Cd2hLpM0dHJpVmpRSoxeM324uH4ERL
MiIt551JplKjs4NgW1TZ8P6m0LFLvpBGU69bmOFHISupgMqOobymRGprpJxASltZ
dITG1gSXbT7VWJ//zf3V95I2Y7AR4BQO3x9AU/BwmMm8XCIUBDfZoVjAvY0/Ufyq
hMhFs1KcsRNUcmYnSouJC42+cm2RiFuPq4kqhQ4kYEeddI0rPBrmOy0NFGsHnHmM
qHvYcCgtb76MhzI/jRvybCSHMNfY2c4xUkhE0Q5UaKOpxk3VfmraYHmtgDAPYBsq
yR7BbqPR8xyC2ZMpriflh5b4SOxdEuGsBQ9IRgktVghNN31l7k5RN6XT8JPVyA/v
xNRw+8oK3OK9QSkbtsAmLGVKBb5yoSSOVvRFgkp2Lq5W4of0OVWX14+De1e8jDXx
J65MUZw4py2duNZINTEi6exrqmuMRtiU3xVXsGbTRXHPuAzDolPowXIfXAim9ks4
gOwpVvBra7C6Nb2+3ZMcsmwk9YyjuqGX8W0Q0OJTyavdVFqHvsVPseAFvvOsUfye
P/WHCMbj7uHZP57VgxHbiO7BzbmI8Hkv7ol5iqKpOh4bgKOfVxAdBr9z/QRZpK8F
k1/pNRNd6rPssL2xeCEb74cVpoq3G7GkDWCCKfdttzBuMmFxoF26hbpYZO7UWQdn
+VUhOAqo72zqaIU4EPPdm/SUE7RGjDbFVj+RHUVYsT9tA5bAWLpLPzuHE/cA/pct
yocor9eQFmy/sDMy9u4rrENjJuD1bcRKIUKMmxfLAX+VqBFTNduU/QygXWs+8rqU
QivIR8TTTnM4fiCm3i3Em6Ifi0dGCunoBX/vHi4FjNwJWuQv+DEKOMnm2r8bECcL
c9H61u6RLBWfB/LQlOioYu0akGQ6gG5sMUCmZFdFsqKBjhk6HJIwNAXeAY+Z/X21
3s1db9n/fIEGQ/7lwYX3DwGDBDPMv20vzJgL2XqLcVbpQEajheSvca0ZV3G90uTQ
IcVEvstCbxzR60QwLLobA7Sesi0WSvtDMhaxLuyzoR7B7mKMPUoP0TUV2caYq59y
AtSkfKB2NHKNJ6yML5X4Q4Z4pmdSmJvjtMHEk27BvNuOUNvd5lyGFIw=
-----END ENCRYPTED PRIVATE KEY-----
`
