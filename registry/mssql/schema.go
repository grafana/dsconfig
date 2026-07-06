package mssqldatasource

import (
	_ "embed"

	"k8s.io/kube-openapi/pkg/spec3"

	"github.com/grafana/dsconfig/dsconfig"
	sdkschema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
)

//go:generate go test -generateArtifacts -run TestSchemaConformance ./...

const TargetAPIVersion = dsconfig.TargetAPIVersion

//go:embed dsconfig.json
var configSchemaJSON []byte

func ConfigSchema() (*dsconfig.Schema, error) {
	return dsconfig.ParseAndResolveSchemaJSON(configSchemaJSON)
}

func NewSchema() (*sdkschema.PluginSchema, error) {
	return dsconfig.NewSDKSchema(configSchemaJSON, SettingsExamples())
}

func SettingsExamples() *sdkschema.SettingsExamples {
	return &sdkschema.SettingsExamples{
		Examples: map[string]*spec3.Example{
			"": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Default configuration",
					Description: "The defaults a new datasource starts with: SQL Server Authentication, encrypt='false'. The user must still supply url, user, jsonData.database, and secureJsonData.password.",
					Value: map[string]any{
						"url":  "",
						"user": "",
						"jsonData": map[string]any{
							"authenticationType": string(AuthTypeSQL),
							"encrypt":            string(EncryptFalse),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "",
						},
					},
				},
			},
			"sqlAuth": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "SQL Server Authentication",
					Description: "The typical setup: SQL Server login with encryption negotiated as 'false' (login packet only).",
					Value: map[string]any{
						"url":  "mssql.internal:1433",
						"user": "grafana_reader",
						"jsonData": map[string]any{
							"database":           "metrics",
							"authenticationType": string(AuthTypeSQL),
							"encrypt":            string(EncryptFalse),
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"tlsEncrypted": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Fully encrypted TLS connection",
					Description: "encrypt='true' with TLS verification enabled — supply a root CA path and the server-cert Common Name.",
					Value: map[string]any{
						"url":  "mssql.internal:1433",
						"user": "grafana_reader",
						"jsonData": map[string]any{
							"database":           "metrics",
							"authenticationType": string(AuthTypeSQL),
							"encrypt":            string(EncryptTrue),
							"tlsSkipVerify":      false,
							"sslRootCertFile":    "/etc/secrets/mssql/ca.pem",
							"serverName":         "mssql.internal",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"windowsAuthentication": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Windows Authentication (SSO)",
					Description: "Integrated Security — no credentials stored in the datasource; the Grafana process's Windows identity is used.",
					Value: map[string]any{
						"url": "mssql.internal:1433",
						"jsonData": map[string]any{
							"database":           "metrics",
							"authenticationType": string(AuthTypeWindows),
							"encrypt":            string(EncryptFalse),
						},
						"secureJsonData": map[string]any{},
					},
				},
			},
			"kerberosUsernamePassword": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Windows AD: Username + password",
					Description: "Kerberos with a UPN-formatted username and password. Requires MIT krb5 on the Grafana host.",
					Value: map[string]any{
						"url":  "mssql.internal:1433",
						"user": "grafana_reader@EXAMPLE.COM",
						"jsonData": map[string]any{
							"database":           "metrics",
							"authenticationType": string(AuthTypeKerberosRaw),
							"configFilePath":     "/etc/krb5.conf",
							"UDPConnectionLimit": 1,
							"enableDNSLookupKDC": "true",
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyPassword): "changeme",
						},
					},
				},
			},
			"kerberosKeytab": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Windows AD: Keytab file",
					Description: "Kerberos authenticating via a keytab file on the Grafana host.",
					Value: map[string]any{
						"url":  "mssql.internal:1433",
						"user": "grafana_reader@EXAMPLE.COM",
						"jsonData": map[string]any{
							"database":           "metrics",
							"authenticationType": string(AuthTypeKerberosKeytab),
							"keytabFilePath":     "/etc/secrets/mssql.keytab",
							"configFilePath":     "/etc/krb5.conf",
							"UDPConnectionLimit": 1,
						},
						"secureJsonData": map[string]any{},
					},
				},
			},
			"kerberosCredentialCache": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Windows AD: Credential cache",
					Description: "Kerberos using a credential-cache path on disk.",
					Value: map[string]any{
						"url": "mssql.internal:1433",
						"jsonData": map[string]any{
							"database":           "metrics",
							"authenticationType": string(AuthTypeKerberosCache),
							"credentialCache":    "/tmp/krb5cc_1000",
							"configFilePath":     "/etc/krb5.conf",
							"UDPConnectionLimit": 1,
						},
						"secureJsonData": map[string]any{},
					},
				},
			},
			"kerberosCredentialCacheFile": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Windows AD: Credential cache file",
					Description: "Kerberos using a credential-cache lookup file.",
					Value: map[string]any{
						"url":  "mssql.internal:1433",
						"user": "grafana_reader@EXAMPLE.COM",
						"jsonData": map[string]any{
							"database":                  "metrics",
							"authenticationType":        string(AuthTypeKerberosCacheLookupFile),
							"credentialCacheLookupFile": "/home/grafana/cache.json",
							"configFilePath":            "/etc/krb5.conf",
							"UDPConnectionLimit":        1,
						},
						"secureJsonData": map[string]any{},
					},
				},
			},
			"azureAdManagedIdentity": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Azure AD: Managed Identity",
					Description: "Azure AD authentication using the Grafana host's managed identity. No secret required.",
					Value: map[string]any{
						"url": "mssql.database.windows.net:1433",
						"jsonData": map[string]any{
							"database":           "metrics",
							"authenticationType": string(AuthTypeAzureAD),
							"encrypt":            string(EncryptTrue),
							"azureCredentials": map[string]any{
								"authType": "msi",
							},
						},
						"secureJsonData": map[string]any{},
					},
				},
			},
			"azureAdClientSecret": {
				ExampleProps: spec3.ExampleProps{
					Summary:     "Azure AD: Client Secret",
					Description: "Azure AD service-principal authentication. The client secret is stored write-only as secureJsonData.azureClientSecret; the rest of the credential (tenant, client, cloud) lives in jsonData.azureCredentials.",
					Value: map[string]any{
						"url": "mssql.database.windows.net:1433",
						"jsonData": map[string]any{
							"database":           "metrics",
							"authenticationType": string(AuthTypeAzureAD),
							"encrypt":            string(EncryptTrue),
							"azureCredentials": map[string]any{
								"authType":   "clientsecret",
								"azureCloud": "AzureCloud",
								"tenantId":   "00000000-0000-0000-0000-000000000000",
								"clientId":   "00000000-0000-0000-0000-000000000000",
							},
						},
						"secureJsonData": map[string]any{
							string(SecureJsonDataKeyAzureClientSecret): "changeme",
						},
					},
				},
			},
		},
	}
}
