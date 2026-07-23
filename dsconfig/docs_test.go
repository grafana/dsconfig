package dsconfig

import (
	"strings"
	"testing"
)

func TestRenderMarkdownDocs_Minimal(t *testing.T) {
	target := SecureJSONTarget
	s := &Schema{
		SchemaVersion: "v1",
		PluginType:    "grafana-hello-datasource",
		PluginName:    "Hello",
		DocURL:        "https://example.com/docs",
		Fields: []ConfigField{
			{
				ID:          "secureJsonData_apiKey",
				Key:         "apiKey",
				Label:       "API Key",
				Description: "Your API key",
				ValueType:   "string",
				Target:      &target,
				Required:    true,
				UI:          &FieldUI{Component: UIInput, Placeholder: "sk-..."},
			},
		},
		Groups: []ConfigGroup{
			{ID: "auth", Title: "Authentication", FieldRefs: []string{"secureJsonData_apiKey"}},
		},
	}

	out, err := RenderMarkdownDocs(s)
	if err != nil {
		t.Fatalf("RenderMarkdownDocs: %v", err)
	}
	mustContain(t, out, "# Hello configuration")
	mustContain(t, out, "grafana-hello-datasource")
	mustContain(t, out, "https://example.com/docs")
	mustContain(t, out, "## Authentication")
	mustContain(t, out, "### API Key")
	mustContain(t, out, "🔒 secret")
	mustContain(t, out, "**required**")
	mustContain(t, out, "`sk-...`")
	// Consumer facing: must NOT expose internal IDs or storage targets.
	mustNotContain(t, out, "secureJsonData_apiKey")
	mustNotContain(t, out, "secureJsonData\"")
}

func TestRenderMarkdownDocs_Ungrouped(t *testing.T) {
	target := JSONDataTarget
	s := &Schema{
		SchemaVersion: "v1",
		PluginType:    "x",
		PluginName:    "X",
		Fields: []ConfigField{
			{
				ID: "jsonData.url", Key: "url", Label: "Server URL",
				ValueType: "string", Target: &target,
				UI: &FieldUI{Component: UIInput},
			},
		},
	}
	out, err := RenderMarkdownDocs(s)
	if err != nil {
		t.Fatal(err)
	}
	mustContain(t, out, "## Other settings")
	mustContain(t, out, "### Server URL")
}

func TestRenderMarkdownDocs_SelectOptionsAndConditional(t *testing.T) {
	jd := JSONDataTarget
	sd := SecureJSONTarget
	s := &Schema{
		SchemaVersion: "v1",
		PluginType:    "x",
		PluginName:    "X",
		Fields: []ConfigField{
			{
				ID: "jsonData_auth_id", Key: "id", Label: "Auth method",
				ValueType: "string", Target: &jd, DefaultValue: "bearer_token",
				UI: &FieldUI{
					Component: UISelect,
					Options: []FieldOption{
						{Label: "******", Value: "bearer_token"},
						{Label: "Basic", Value: "basic"},
					},
				},
			},
			{
				ID: "secureJsonData_token", Key: "token", Label: "Token",
				ValueType: "string", Target: &sd,
				DependsOn:    "jsonData_auth_id == 'bearer_token'",
				RequiredWhen: "jsonData_auth_id == 'bearer_token'",
				UI:           &FieldUI{Component: UIInput},
			},
		},
		Groups: []ConfigGroup{
			{ID: "auth", Title: "Authentication",
				FieldRefs: []string{"jsonData_auth_id", "secureJsonData_token"}},
		},
	}
	out, err := RenderMarkdownDocs(s)
	if err != nil {
		t.Fatal(err)
	}
	mustContain(t, out, "`bearer_token` (******")
	mustContain(t, out, "`basic` (Basic)")
	mustContain(t, out, "Default")
	mustContain(t, out, "`bearer_token`")
	mustContain(t, out, "conditionally required")
	mustContain(t, out, "**Auth method** is **")
	mustContain(t, out, "(`bearer_token`)")
}

func TestRenderMarkdownDocs_ValidationsAndRange(t *testing.T) {
	jd := JSONDataTarget
	min := 1.0
	max := 60.0
	s := &Schema{
		SchemaVersion: "v1", PluginType: "x", PluginName: "X",
		Fields: []ConfigField{
			{
				ID: "jsonData.timeout", Key: "timeout", Label: "Timeout (s)",
				ValueType: "number", Target: &jd,
				Validations: []FieldValidationRule{
					{Type: RangeValidation, Min: &min, Max: &max},
				},
				UI: &FieldUI{Component: UIInput},
			},
			{
				ID: "jsonData.mode", Key: "mode", Label: "Mode",
				ValueType: "string", Target: &jd,
				Validations: []FieldValidationRule{
					{Type: AllowedValuesValidation, Values: []any{"a", "b", "c"}},
				},
			},
			{
				ID: "jsonData.slug", Key: "slug", Label: "Slug",
				ValueType: "string", Target: &jd,
				Validations: []FieldValidationRule{
					{Type: PatternValidation, Pattern: "^[a-z0-9-]+$"},
				},
			},
		},
	}
	out, err := RenderMarkdownDocs(s)
	if err != nil {
		t.Fatal(err)
	}
	mustContain(t, out, "1 – 60")
	mustContain(t, out, "`a`, `b`, `c`")
	mustContain(t, out, "`^[a-z0-9-]+$`")
}

func TestRenderMarkdownDocs_HiddenVirtualField(t *testing.T) {
	s := &Schema{
		SchemaVersion: "v1", PluginType: "x", PluginName: "X",
		Fields: []ConfigField{
			// Virtual field with no UI → hidden from consumer doc.
			{ID: "v1", Key: "v", Label: "internal", Kind: VirtualField, ValueType: "string"},
			// Visible field so the doc has content.
			{ID: "jsonData.a", Key: "a", Label: "A", ValueType: "string",
				Target: func() *TargetLocation { t := JSONDataTarget; return &t }(),
				UI:     &FieldUI{Component: UIInput}},
		},
	}
	out, err := RenderMarkdownDocs(s)
	if err != nil {
		t.Fatal(err)
	}
	mustNotContain(t, out, "internal")
	mustContain(t, out, "### A")
}

func TestRenderMarkdownDocs_ArrayItemFields(t *testing.T) {
	jd := JSONDataTarget
	s := &Schema{
		SchemaVersion: "v1", PluginType: "x", PluginName: "X",
		Fields: []ConfigField{
			{
				ID: "jsonData.headers", Key: "headers", Label: "Custom headers",
				ValueType: "array", Target: &jd,
				UI: &FieldUI{Component: UIList},
				Item: &FieldItemSchema{
					Fields: []ConfigField{
						{ID: "jsonData.headers.item.name", Key: "name", Label: "Name",
							ValueType: "string"},
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
	mustContain(t, out, "### Custom headers")
	mustContain(t, out, "Each item has the following fields:")
	mustContain(t, out, "#### Name")
	mustContain(t, out, "#### Value")
}

func TestRenderMarkdownDocs_MissingSchema(t *testing.T) {
	if _, err := RenderMarkdownDocs(nil); err == nil {
		t.Fatal("expected error for nil schema")
	}
	if _, err := RenderMarkdownDocs(&Schema{}); err == nil {
		t.Fatal("expected error for schema missing plugin metadata")
	}
}

func TestRenderMarkdownDocs_ResolvedBaseFields(t *testing.T) {
	// Parse a small dsconfig JSON without baseFields to verify that
	// ParseAndResolveSchemaJSON + RenderMarkdownDocs work together.
	jsonIn := `{
	  "schemaVersion": "v1",
	  "pluginType": "x",
	  "pluginName": "X",
	  "fields": [
	    {"id": "jsonData.custom", "key": "custom", "label": "Custom",
	     "valueType": "string", "target": "jsonData",
	     "ui": {"component": "input"}}
	  ],
	  "groups": [
	    {"id": "connection", "title": "Connection",
	     "fieldRefs": ["jsonData.custom"]}
	  ]
	}`
	s, err := ParseAndResolveSchemaJSON([]byte(jsonIn))
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	out, err := RenderMarkdownDocs(s)
	if err != nil {
		t.Fatal(err)
	}
	mustContain(t, out, "## Connection")
	mustContain(t, out, "### Custom")
}

func mustContain(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Fatalf("expected output to contain %q; got:\n%s", needle, haystack)
	}
}

func mustNotContain(t *testing.T, haystack, needle string) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		t.Fatalf("expected output NOT to contain %q; got:\n%s", needle, haystack)
	}
}
