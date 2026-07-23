package dsconfig

import (
	"strings"
	"testing"
)

// ptr returns a pointer to v — helper for the small test schemas below.
func ptrTarget(t TargetLocation) *TargetLocation { return &t }

func TestRenderMarkdownDocs_FieldsTable(t *testing.T) {
	s := &Schema{
		SchemaVersion: "v1",
		PluginType:    "grafana-catchpoint-datasource",
		PluginName:    "Catchpoint",
		DocURL:        "https://example.com/docs",
		Fields: []ConfigField{
			{
				ID: "jsonData_services_catchpoint_auth_id", Key: "id",
				Section: "services.catchpoint.auth",
				Label:   "API v2 Key", Description: "Catchpoint REST API v2 Key.",
				ValueType: "string", Target: ptrTarget(JSONDataTarget),
				Role:         RoleAuthDiscriminator,
				DefaultValue: "bearer_token",
				Validations: []FieldValidationRule{
					{Type: AllowedValuesValidation, Values: []any{"bearer_token"}},
				},
				UI: &FieldUI{Component: UISelect, Options: []FieldOption{
					{Label: "API v2 Key", Value: "bearer_token"},
				}},
			},
			{
				ID: "secureJsonData_catchpoint_token", Key: "catchpoint.token",
				Label:     "Token",
				Description: "Token for accessing the datasource API",
				ValueType: "string", Target: ptrTarget(SecureJSONTarget),
				Role:         RoleAuthBearerToken,
				DependsOn:    "jsonData_services_catchpoint_auth_id == 'bearer_token'",
				RequiredWhen: "jsonData_services_catchpoint_auth_id == 'bearer_token'",
				UI:           &FieldUI{Component: UIInput, Placeholder: "Token value"},
			},
		},
	}
	out, err := RenderMarkdownDocs(s)
	if err != nil {
		t.Fatal(err)
	}

	// Header.
	mustContain(t, out, "# Catchpoint configuration")
	mustContain(t, out, "grafana-catchpoint-datasource")
	mustContain(t, out, "[official documentation](https://example.com/docs)")

	// Fields table.
	mustContain(t, out, "## Fields")
	mustContain(t, out, "| Field | Type | Target | Required | Description |")
	mustContain(t, out, "`jsonData.services.catchpoint.auth.id`")
	mustContain(t, out, "`secureJsonData.catchpoint.token` 🔒")
	mustContain(t, out, "enum (bearer_token)")
	mustContain(t, out, "conditional")

	// Provisioning examples — one scenario per auth option.
	mustContain(t, out, "## Provisioning examples")
	mustContain(t, out, "### API v2 Key (`bearer_token`)")
	mustContain(t, out, "**Grafana provisioning YAML**")
	mustContain(t, out, "apiVersion: 1")
	mustContain(t, out, "type: grafana-catchpoint-datasource")
	mustContain(t, out, "name: Catchpoint")
	mustContain(t, out, "services:")
	mustContain(t, out, "id: bearer_token")
	mustContain(t, out, "catchpoint.token: \"<YOUR_TOKEN>\"")

	mustContain(t, out, "**Terraform**")
	mustContain(t, out, `resource "grafana_data_source" "grafana_catchpoint_datasource_bearer_token"`)
	mustContain(t, out, `type = "grafana-catchpoint-datasource"`)
	mustContain(t, out, "json_data_encoded = jsonencode(")
	mustContain(t, out, "secure_json_data_encoded = jsonencode(")
	mustContain(t, out, `"catchpoint.token" = "<YOUR_TOKEN>"`)
}

func TestRenderMarkdownDocs_MultipleAuthScenarios(t *testing.T) {
	s := &Schema{
		SchemaVersion: "v1", PluginType: "x", PluginName: "X",
		Fields: []ConfigField{
			{
				ID: "jsonData.authType", Key: "authType",
				Label: "Auth", ValueType: "string",
				Target:       ptrTarget(JSONDataTarget),
				Role:         RoleAuthDiscriminator,
				DefaultValue: "basic",
				UI: &FieldUI{Component: UISelect, Options: []FieldOption{
					{Label: "Basic", Value: "basic"},
					{Label: "Token", Value: "token"},
				}},
			},
			{
				ID: "jsonData.user", Key: "user",
				Label: "User", ValueType: "string",
				Target:       ptrTarget(JSONDataTarget),
				DependsOn:    "jsonData.authType == 'basic'",
				RequiredWhen: "jsonData.authType == 'basic'",
			},
			{
				ID: "secureJsonData.password", Key: "password",
				Label: "Password", ValueType: "string",
				Target:       ptrTarget(SecureJSONTarget),
				DependsOn:    "jsonData.authType == 'basic'",
				RequiredWhen: "jsonData.authType == 'basic'",
			},
			{
				ID: "secureJsonData.token", Key: "token",
				Label: "Token", ValueType: "string",
				Target:       ptrTarget(SecureJSONTarget),
				DependsOn:    "jsonData.authType == 'token'",
				RequiredWhen: "jsonData.authType == 'token'",
			},
		},
	}
	out, err := RenderMarkdownDocs(s)
	if err != nil {
		t.Fatal(err)
	}
	// Two scenarios rendered.
	mustContain(t, out, "### Basic (`basic`)")
	mustContain(t, out, "### Token (`token`)")

	// The Basic scenario must contain user + password but NOT token.
	basic, tokenSec := splitScenarios(out)
	if !strings.Contains(basic, "user:") {
		t.Errorf("basic scenario should contain user: %s", basic)
	}
	if !strings.Contains(basic, "password:") {
		t.Errorf("basic scenario should contain password: %s", basic)
	}
	if strings.Contains(basic, "\n      token:") {
		t.Errorf("basic scenario should NOT contain token secret: %s", basic)
	}
	// The Token scenario must contain the token secret but NOT password.
	if !strings.Contains(tokenSec, "token:") {
		t.Errorf("token scenario should contain token: %s", tokenSec)
	}
	if strings.Contains(tokenSec, "password:") {
		t.Errorf("token scenario should NOT contain password: %s", tokenSec)
	}
}

// splitScenarios splits the doc at the two "### " scenario headings and
// returns (basicSection, tokenSection).
func splitScenarios(doc string) (string, string) {
	i := strings.Index(doc, "### Basic")
	j := strings.Index(doc, "### Token")
	if i < 0 || j < 0 {
		return "", ""
	}
	return doc[i:j], doc[j:]
}

func TestRenderMarkdownDocs_NoDiscriminator_DefaultScenario(t *testing.T) {
	s := &Schema{
		SchemaVersion: "v1", PluginType: "x", PluginName: "X",
		Fields: []ConfigField{
			{
				ID: "jsonData.url", Key: "url",
				Label: "URL", ValueType: "string",
				Target: ptrTarget(JSONDataTarget), Required: true,
				UI: &FieldUI{Component: UIInput, Placeholder: "http://example"},
			},
		},
	}
	out, err := RenderMarkdownDocs(s)
	if err != nil {
		t.Fatal(err)
	}
	mustContain(t, out, "### Default configuration")
	mustContain(t, out, "url: \"http://example\"")
}

func TestRenderMarkdownDocs_RootFieldsAndTerraformMapping(t *testing.T) {
	s := &Schema{
		SchemaVersion: "v1", PluginType: "prometheus", PluginName: "Prometheus",
		Fields: []ConfigField{
			{
				ID: "url", Key: "url",
				Label: "Prometheus server URL", ValueType: "string",
				Target: ptrTarget(RootTarget), Required: true, Role: RoleEndpointBaseURL,
				UI: &FieldUI{Component: UIInput, Placeholder: "http://localhost:9090"},
			},
			{
				ID: "basicAuth", Key: "basicAuth",
				Label: "Basic auth", ValueType: "boolean",
				Target: ptrTarget(RootTarget), DefaultValue: false,
			},
		},
	}
	out, err := RenderMarkdownDocs(s)
	if err != nil {
		t.Fatal(err)
	}
	// Root fields → top-level YAML keys.
	mustContain(t, out, "\n    url: \"http://localhost:9090\"\n")
	mustContain(t, out, "\n    basicAuth: false\n")
	// Terraform maps `url` root field to the provider's `url` attribute.
	mustContain(t, out, `url = "http://localhost:9090"`)
}

func TestRenderMarkdownDocs_ArrayItemFieldsInTable(t *testing.T) {
	s := &Schema{
		SchemaVersion: "v1", PluginType: "x", PluginName: "X",
		Fields: []ConfigField{
			{
				ID: "jsonData.headers", Key: "headers",
				Label: "Headers", ValueType: "array",
				Target: ptrTarget(JSONDataTarget),
				UI:     &FieldUI{Component: UIList},
				Item: &FieldItemSchema{
					Fields: []ConfigField{
						{ID: "jsonData.headers.item.name", Key: "name", Label: "Name",
							ValueType: "string", Required: true},
						{ID: "jsonData.headers.item.value", Key: "value", Label: "Value",
							ValueType: "string"},
					},
				},
			},
		},
	}
	out, err := RenderMarkdownDocs(s)
	if err != nil {
		t.Fatal(err)
	}
	mustContain(t, out, "`jsonData.headers`")
	mustContain(t, out, "`jsonData.headers[].name`")
	mustContain(t, out, "`jsonData.headers[].value`")
}

func TestRenderMarkdownDocs_MissingSchema(t *testing.T) {
	if _, err := RenderMarkdownDocs(nil); err == nil {
		t.Fatal("expected error for nil schema")
	}
	if _, err := RenderMarkdownDocs(&Schema{}); err == nil {
		t.Fatal("expected error for schema missing plugin metadata")
	}
}

func TestConditionMatches(t *testing.T) {
	cases := []struct {
		expr, id, val string
		want          bool
	}{
		{"authType == 'basic'", "authType", "basic", true},
		{"authType == 'basic'", "authType", "token", false},
		{"authType == 'basic' || authType == 'token'", "authType", "token", true},
		{"authType == 'basic' || authType == 'token'", "authType", "other", false},
		// Non-discriminator variable is treated as match (can't evaluate).
		{"someOther == 'x'", "authType", "basic", true},
		// Non-trivial operator: fall back to match.
		{"authType != 'basic'", "authType", "basic", true},
	}
	for _, c := range cases {
		got := conditionMatches(c.expr, c.id, c.val)
		if got != c.want {
			t.Errorf("conditionMatches(%q, %q, %q) = %v, want %v",
				c.expr, c.id, c.val, got, c.want)
		}
	}
}

func TestYAMLScalarQuoting(t *testing.T) {
	cases := map[string]string{
		"":         `""`,
		"true":     `"true"`,
		"1234":     `"1234"`,
		"foo":      "foo",
		"http://a": `"http://a"`,
		"foo bar":  "foo bar", // internal spaces are valid unquoted YAML
	}
	for in, want := range cases {
		got := yamlScalar(in)
		if got != want {
			t.Errorf("yamlScalar(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPlaceholderFor(t *testing.T) {
	f := &ConfigField{Label: "API v2 Key"}
	if got, want := placeholderFor(f), "<YOUR_API_V2_KEY>"; got != want {
		t.Errorf("placeholderFor label: got %q want %q", got, want)
	}
	f = &ConfigField{Key: "token"}
	if got, want := placeholderFor(f), "<YOUR_TOKEN>"; got != want {
		t.Errorf("placeholderFor key: got %q want %q", got, want)
	}
}

func mustContain(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected output to contain %q; got:\n%s", needle, haystack)
	}
}
