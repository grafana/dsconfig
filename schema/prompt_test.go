package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/grafana/dsconfig/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// ExtractLiteralFromWhen
// ============================================================

func TestExtractLiteralFromWhen_SingleQuotedString(t *testing.T) {
	assert.Equal(t, "basic-auth", schema.ExtractLiteralFromWhen("value == 'basic-auth'"))
}

func TestExtractLiteralFromWhen_DoubleQuotedString(t *testing.T) {
	assert.Equal(t, "forward-oauth", schema.ExtractLiteralFromWhen(`value == "forward-oauth"`))
}

func TestExtractLiteralFromWhen_True(t *testing.T) {
	assert.Equal(t, true, schema.ExtractLiteralFromWhen("value == true"))
}

func TestExtractLiteralFromWhen_False(t *testing.T) {
	assert.Equal(t, false, schema.ExtractLiteralFromWhen("value == false"))
}

func TestExtractLiteralFromWhen_Integer(t *testing.T) {
	assert.Equal(t, float64(42), schema.ExtractLiteralFromWhen("value == 42"))
}

func TestExtractLiteralFromWhen_NegativeFloat(t *testing.T) {
	assert.Equal(t, float64(-3.14), schema.ExtractLiteralFromWhen("value == -3.14"))
}

func TestExtractLiteralFromWhen_Complex(t *testing.T) {
	assert.Nil(t, schema.ExtractLiteralFromWhen("value.startsWith('http')"))
}

func TestExtractLiteralFromWhen_Whitespace(t *testing.T) {
	assert.Equal(t, "spaced", schema.ExtractLiteralFromWhen("value  ==  'spaced'"))
}

// ============================================================
// ToPromptSchema — simple fields
// ============================================================

func TestToPromptSchema_BasicField(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "url", Key: "url", ValueType: schema.StringType,
				SemanticType: schema.URLType,
				Target:       ptr(schema.RootTarget),
				Required:     true,
				Description:  "Base URL",
			},
		},
	}
	ps := schema.ToPromptSchema(s)
	assert.Equal(t, "test", ps.PluginType)
	assert.Equal(t, "Test", ps.PluginName)
	require.Len(t, ps.Fields, 1)

	f := ps.Fields[0]
	assert.Equal(t, "url", f.ID)
	assert.Equal(t, "root.url", f.Path)
	assert.Equal(t, schema.StringType, f.Type)
	assert.Equal(t, schema.URLType, f.SemanticType)
	assert.True(t, f.Required)
	assert.Equal(t, "Base URL", f.Description)
}

func TestToPromptSchema_StripsUI(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "url", Key: "url", ValueType: schema.StringType,
				Target: ptr(schema.RootTarget),
				UI:     &schema.FieldUI{Component: schema.UIInput, Placeholder: "https://..."},
			},
		},
	}
	ps := schema.ToPromptSchema(s)
	// Verify by JSON round-trip: no UI fields in output
	data, err := json.Marshal(ps.Fields[0])
	require.NoError(t, err)
	var raw map[string]any
	require.NoError(t, json.Unmarshal(data, &raw))
	assert.Nil(t, raw["ui"])
	assert.Nil(t, raw["placeholder"])
}

func TestToPromptSchema_FlattensAllowedValues(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "method", Key: "httpMethod", ValueType: schema.StringType,
				Target: ptr(schema.JSONDataTarget),
				Validations: []schema.FieldValidationRule{
					{Type: schema.AllowedValuesValidation, Values: []any{"GET", "POST"}},
				},
			},
		},
	}
	ps := schema.ToPromptSchema(s)
	assert.Equal(t, []any{"GET", "POST"}, ps.Fields[0].AllowedValues)
}

func TestToPromptSchema_FlattensPattern(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "url", Key: "url", ValueType: schema.StringType,
				Target: ptr(schema.RootTarget),
				Validations: []schema.FieldValidationRule{
					{Type: schema.PatternValidation, Pattern: "^https?://"},
				},
			},
		},
	}
	ps := schema.ToPromptSchema(s)
	assert.Equal(t, "^https?://", ps.Fields[0].Pattern)
}

func TestToPromptSchema_FlattensRange(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "timeout", Key: "timeout", ValueType: schema.NumberType,
				Target: ptr(schema.JSONDataTarget),
				Validations: []schema.FieldValidationRule{
					{Type: schema.RangeValidation, Min: ptrF(1), Max: ptrF(600)},
				},
			},
		},
	}
	ps := schema.ToPromptSchema(s)
	require.NotNil(t, ps.Fields[0].Range)
	assert.Equal(t, 1.0, *ps.Fields[0].Range.Min)
	assert.Equal(t, 600.0, *ps.Fields[0].Range.Max)
}

func TestToPromptSchema_PathWithSection(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "nested", Key: "datasourceUid", ValueType: schema.StringType,
				Target: ptr(schema.JSONDataTarget), Section: "tracesToLogs",
			},
		},
	}
	ps := schema.ToPromptSchema(s)
	assert.Equal(t, "jsonData.tracesToLogs.datasourceUid", ps.Fields[0].Path)
}

func TestToPromptSchema_DefaultValue(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "method", Key: "httpMethod", ValueType: schema.StringType,
				Target: ptr(schema.JSONDataTarget), DefaultValue: "POST",
			},
		},
	}
	ps := schema.ToPromptSchema(s)
	assert.Equal(t, "POST", ps.Fields[0].DefaultValue)
}

// ============================================================
// ToPromptSchema — managed fields
// ============================================================

func TestToPromptSchema_ExcludesManagedFields(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{ID: "visible", Key: "url", ValueType: schema.StringType, Target: ptr(schema.RootTarget)},
			{ID: "hidden", Key: "basicAuth", ValueType: schema.BooleanType, Target: ptr(schema.RootTarget), Tags: []string{"managed-by:auth.method"}},
		},
	}
	ps := schema.ToPromptSchema(s)
	require.Len(t, ps.Fields, 1)
	assert.Equal(t, "visible", ps.Fields[0].ID)
}

// ============================================================
// ToPromptSchema — effects → options
// ============================================================

func TestToPromptSchema_EffectsToOptions(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "auth.method", Key: "authMethod", ValueType: schema.StringType,
				Kind: schema.VirtualField,
				UI: &schema.FieldUI{
					Component: schema.UISelect,
					Options: []schema.FieldOption{
						{Label: "No Authentication", Value: "no-auth"},
						{Label: "Basic authentication", Value: "basic-auth"},
					},
				},
				Effects: []schema.FieldEffect{
					{When: "value == 'no-auth'", Set: map[string]any{"ba": false}},
					{When: "value == 'basic-auth'", Set: map[string]any{"ba": true}},
				},
			},
			{ID: "ba", Key: "basicAuth", ValueType: schema.BooleanType, Target: ptr(schema.RootTarget), Tags: []string{"managed-by:auth.method"}},
		},
	}
	ps := schema.ToPromptSchema(s)
	require.Len(t, ps.Fields, 1)

	opts := ps.Fields[0].Options
	require.Len(t, opts, 2)
	assert.Equal(t, "no-auth", opts[0].Value)
	assert.Equal(t, "No Authentication", opts[0].Label)
	assert.Equal(t, false, opts[0].Sets["ba"])
	assert.Equal(t, "Basic authentication", opts[1].Label)
}

// ============================================================
// ToPromptSchema — array items
// ============================================================

func TestToPromptSchema_ArrayItems(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "queries", Key: "queries", ValueType: schema.ArrayType,
				Target: ptr(schema.JSONDataTarget),
				Item: &schema.FieldItemSchema{
					ValueType: schema.ObjectType,
					Fields: []schema.ConfigField{
						{ID: "queries.item.name", Key: "name", ValueType: schema.StringType, IsItemField: ptr(true)},
						{ID: "queries.item.query", Key: "query", ValueType: schema.StringType, IsItemField: ptr(true), Description: "PromQL query"},
					},
				},
			},
		},
	}
	ps := schema.ToPromptSchema(s)
	require.Len(t, ps.Fields[0].Items, 2)
	assert.Equal(t, "queries.item.name", ps.Fields[0].Items[0].ID)
	assert.Equal(t, "PromQL query", ps.Fields[0].Items[1].Description)
}

// ============================================================
// ToPromptString
// ============================================================

func TestToPromptString_ValidJSON(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{ID: "url", Key: "url", ValueType: schema.StringType, Target: ptr(schema.RootTarget), Required: true},
		},
	}
	str, err := schema.ToPromptString(s)
	require.NoError(t, err)

	var parsed schema.PromptSchema
	require.NoError(t, json.Unmarshal([]byte(str), &parsed))
	assert.Equal(t, "test", parsed.PluginType)
	assert.Len(t, parsed.Fields, 1)
}

func TestToPromptString_PrettyPrinted(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{ID: "url", Key: "url", ValueType: schema.StringType, Target: ptr(schema.RootTarget)},
		},
	}
	str, err := schema.ToPromptString(s)
	require.NoError(t, err)
	assert.Contains(t, str, "\n")
	assert.Contains(t, str, "  ")
}

// ============================================================
// ToPromptSchema — auth-selector example round-trip
// ============================================================

func TestToPromptSchema_AuthSelectorExample(t *testing.T) {
	s := loadExample(t, "auth-selector.schema.json")
	ps := schema.ToPromptSchema(s)

	// auth.basicAuth and auth.oauthPassThru are managed → excluded
	// Remaining: url, auth.method, auth.basicAuthUser, auth.basicAuthPassword
	assert.Len(t, ps.Fields, 4)

	// auth.method has 3 options from effects
	var authMethod *schema.PromptField
	for i := range ps.Fields {
		if ps.Fields[i].ID == "auth.method" {
			authMethod = &ps.Fields[i]
			break
		}
	}
	require.NotNil(t, authMethod)
	require.Len(t, authMethod.Options, 3)
	assert.Equal(t, "no-auth", authMethod.Options[0].Value)
	assert.Equal(t, "No Authentication", authMethod.Options[0].Label)

	// Verify JSON round-trip
	str, err := schema.ToPromptString(s)
	require.NoError(t, err)
	var decoded schema.PromptSchema
	require.NoError(t, json.Unmarshal([]byte(str), &decoded))
	assert.Equal(t, ps.PluginType, decoded.PluginType)
	assert.Len(t, decoded.Fields, 4)
}

func ptrF(v float64) *float64 { return &v }

// ============================================================
// ToPromptText — human-friendly output
// ============================================================

func TestToPromptText_Header(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "prometheus", PluginName: "Prometheus",
		Fields: []schema.ConfigField{
			{ID: "url", Key: "url", ValueType: schema.StringType, Target: ptr(schema.RootTarget)},
		},
	}
	text := schema.ToPromptText(s)
	assert.Contains(t, text, "Prometheus (pluginType: prometheus)")
	assert.Contains(t, text, "Fields:")
}

func TestToPromptText_RequiredFieldWithDescription(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "url", Key: "url", Label: "URL", ValueType: schema.StringType,
				SemanticType: schema.URLType, Target: ptr(schema.RootTarget),
				Required: true, Description: "Base URL of the server",
			},
		},
	}
	text := schema.ToPromptText(s)
	assert.Contains(t, text, "- URL (root.url) [string, url] REQUIRED — Base URL of the server")
}

func TestToPromptText_DefaultValue(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "method", Key: "httpMethod", ValueType: schema.StringType,
				Target: ptr(schema.JSONDataTarget), DefaultValue: "POST",
			},
		},
	}
	text := schema.ToPromptText(s)
	assert.Contains(t, text, `default: "POST"`)
}

func TestToPromptText_UsesLabelFallsBackToID(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{ID: "jsonData.timeout", Key: "timeout", Label: "Timeout", ValueType: schema.NumberType, Target: ptr(schema.JSONDataTarget)},
			{ID: "jsonData.debug", Key: "debug", ValueType: schema.BooleanType, Target: ptr(schema.JSONDataTarget)},
		},
	}
	text := schema.ToPromptText(s)
	assert.Contains(t, text, "- Timeout (jsonData.timeout)")
	assert.Contains(t, text, "- jsonData.debug (jsonData.debug)")
}

func TestToPromptText_AllowedValues(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "method", Key: "httpMethod", ValueType: schema.StringType,
				Target: ptr(schema.JSONDataTarget),
				Validations: []schema.FieldValidationRule{
					{Type: schema.AllowedValuesValidation, Values: []any{"GET", "POST"}},
				},
			},
		},
	}
	text := schema.ToPromptText(s)
	assert.Contains(t, text, `Allowed: "GET", "POST"`)
}

func TestToPromptText_Pattern(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "url", Key: "url", ValueType: schema.StringType,
				Target: ptr(schema.RootTarget),
				Validations: []schema.FieldValidationRule{
					{Type: schema.PatternValidation, Pattern: "^https?://"},
				},
			},
		},
	}
	text := schema.ToPromptText(s)
	assert.Contains(t, text, "Pattern: ^https?://")
}

func TestToPromptText_Range(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "timeout", Key: "timeout", ValueType: schema.NumberType,
				Target: ptr(schema.JSONDataTarget),
				Validations: []schema.FieldValidationRule{
					{Type: schema.RangeValidation, Min: ptrF(1), Max: ptrF(600)},
				},
			},
		},
	}
	text := schema.ToPromptText(s)
	assert.Contains(t, text, "Range: min: 1, max: 600")
}

func TestToPromptText_DependsOn(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "user", Key: "basicAuthUser", ValueType: schema.StringType,
				Target:    ptr(schema.RootTarget),
				DependsOn: "auth.method == 'basic-auth'",
			},
		},
	}
	text := schema.ToPromptText(s)
	assert.Contains(t, text, "Visible when: auth.method == 'basic-auth'")
}

func TestToPromptText_RequiredWhen(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "user", Key: "basicAuthUser", ValueType: schema.StringType,
				Target:       ptr(schema.RootTarget),
				RequiredWhen: "auth.method == 'basic-auth'",
			},
		},
	}
	text := schema.ToPromptText(s)
	assert.Contains(t, text, "Required when: auth.method == 'basic-auth'")
}

func TestToPromptText_EffectsAsOptions(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "sel", Key: "sel", Label: "Selector", ValueType: schema.StringType,
				Kind: schema.VirtualField,
				UI: &schema.FieldUI{
					Component: schema.UISelect,
					Options: []schema.FieldOption{
						{Label: "Off", Value: "off"},
						{Label: "On", Value: "on"},
					},
				},
				Effects: []schema.FieldEffect{
					{When: "value == 'off'", Set: map[string]any{"flag": false}},
					{When: "value == 'on'", Set: map[string]any{"flag": true}},
				},
			},
			{ID: "flag", Key: "flag", ValueType: schema.BooleanType, Target: ptr(schema.JSONDataTarget), Tags: []string{"managed-by:sel"}},
		},
	}
	text := schema.ToPromptText(s)
	assert.Contains(t, text, "Options:")
	assert.Contains(t, text, `"off" (Off)`)
	assert.Contains(t, text, `"on" (On)`)
	assert.Contains(t, text, "flag=true")
	// No Allowed: line when options are present
	assert.NotContains(t, text, "Allowed:")
}

func TestToPromptText_ExcludesManagedFields(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{ID: "visible", Key: "url", ValueType: schema.StringType, Target: ptr(schema.RootTarget)},
			{ID: "hidden", Key: "basicAuth", ValueType: schema.BooleanType, Target: ptr(schema.RootTarget), Tags: []string{"managed-by:auth"}},
		},
	}
	text := schema.ToPromptText(s)
	assert.Contains(t, text, "visible")
	assert.NotContains(t, text, "hidden")
	assert.NotContains(t, text, "basicAuth")
}

func TestToPromptText_ArrayItemFields(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "queries", Key: "queries", Label: "Queries", ValueType: schema.ArrayType,
				Target: ptr(schema.JSONDataTarget),
				Item: &schema.FieldItemSchema{
					ValueType: schema.ObjectType,
					Fields: []schema.ConfigField{
						{ID: "queries.item.name", Key: "name", Label: "Name", ValueType: schema.StringType, IsItemField: ptr(true)},
						{ID: "queries.item.query", Key: "query", Label: "Query", ValueType: schema.StringType, IsItemField: ptr(true), Description: "PromQL"},
					},
				},
			},
		},
	}
	text := schema.ToPromptText(s)
	assert.Contains(t, text, "Item fields:")
	assert.Contains(t, text, "- Name")
	assert.Contains(t, text, "- Query")
	assert.Contains(t, text, "— PromQL")
}

func TestToPromptText_BooleanDefault(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{ID: "flag", Key: "flag", ValueType: schema.BooleanType, Target: ptr(schema.JSONDataTarget), DefaultValue: false},
		},
	}
	text := schema.ToPromptText(s)
	assert.Contains(t, text, "default: false")
}

func TestToPromptText_AuthSelectorExampleFile(t *testing.T) {
	s := loadExample(t, "auth-selector.schema.json")
	text := schema.ToPromptText(s)

	// Header
	assert.Contains(t, text, "Auth Selector Datasource (pluginType: example-auth-selector)")

	// URL field
	assert.Contains(t, text, "- URL (root.url) [string, url] REQUIRED")

	// Auth method with options
	assert.Contains(t, text, "Authentication method")
	assert.Contains(t, text, "Options:")
	assert.Contains(t, text, `"no-auth" (No Authentication)`)
	assert.Contains(t, text, `"basic-auth" (Basic authentication)`)
	assert.Contains(t, text, `"forward-oauth" (Forward OAuth Identity)`)

	// Managed fields excluded
	assert.NotContains(t, text, "managed-by")

	// Conditional fields
	assert.Contains(t, text, "- Username (root.basicAuthUser)")
	assert.Contains(t, text, "Required when: auth.method == 'basic-auth'")
	assert.Contains(t, text, "- Password (secureJsonData.basicAuthPassword) [string, password]")
}
