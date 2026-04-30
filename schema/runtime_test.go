package schema_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/grafana/dsconfig/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// NewDatasourceConfig
// ============================================================

func TestNewDatasourceConfig_ParsesAllSections(t *testing.T) {
	raw := map[string]any{
		"url":       "https://example.com",
		"basicAuth": true,
		"jsonData": map[string]any{
			"timeout": float64(30),
		},
		"secureJsonData": map[string]any{
			"password": "secret",
		},
		"secureJsonFields": map[string]any{
			"password": true,
		},
	}

	dc := schema.NewDatasourceConfig(raw)

	assert.Equal(t, "https://example.com", dc.Root["url"])
	assert.Equal(t, true, dc.Root["basicAuth"])
	assert.Equal(t, float64(30), dc.JSONData["timeout"])
	assert.Equal(t, "secret", dc.SecureJSONData["password"])
	assert.True(t, dc.SecureJSONFields["password"])
}

// ============================================================
// LoadAndValidate — simple field extraction
// ============================================================

func TestLoadAndValidate_RootField(t *testing.T) {
	s := minimalSchema(schema.ConfigField{
		ID: "url", Key: "url", ValueType: schema.StringType,
		Target: ptr(schema.RootTarget), Required: true,
	})

	config := schema.DatasourceConfig{
		Root: map[string]any{"url": "https://example.com"},
	}

	result, err := schema.LoadAndValidate(s, config, schema.ReadMode)
	require.NoError(t, err)
	assert.False(t, result.HasErrors())
	assert.Equal(t, "https://example.com", result.Values["url"].Value)
	assert.Equal(t, schema.SourceConfig, result.Values["url"].Source)
}

func TestLoadAndValidate_JSONDataField(t *testing.T) {
	s := minimalSchema(schema.ConfigField{
		ID: "timeout", Key: "timeout", ValueType: schema.NumberType,
		Target: ptr(schema.JSONDataTarget),
	})

	config := schema.DatasourceConfig{
		JSONData: map[string]any{"timeout": float64(30)},
	}

	result, err := schema.LoadAndValidate(s, config, schema.ReadMode)
	require.NoError(t, err)
	assert.False(t, result.HasErrors())
	assert.Equal(t, float64(30), result.Values["timeout"].Value)
}

func TestLoadAndValidate_SectionField(t *testing.T) {
	s := minimalSchema(schema.ConfigField{
		ID: "oauth.clientId", Key: "clientId", ValueType: schema.StringType,
		Target: ptr(schema.JSONDataTarget), Section: "oauth2",
	})

	config := schema.DatasourceConfig{
		JSONData: map[string]any{
			"oauth2": map[string]any{
				"clientId": "my-client",
			},
		},
	}

	result, err := schema.LoadAndValidate(s, config, schema.ReadMode)
	require.NoError(t, err)
	assert.Equal(t, "my-client", result.Values["oauth.clientId"].Value)
}

func TestLoadAndValidate_DottedSectionField(t *testing.T) {
	s := minimalSchema(schema.ConfigField{
		ID: "oauth.tokenUrl", Key: "tokenUrl", ValueType: schema.StringType,
		Target: ptr(schema.JSONDataTarget), Section: "oauth2.endpoints",
	})

	config := schema.DatasourceConfig{
		JSONData: map[string]any{
			"oauth2": map[string]any{
				"endpoints": map[string]any{
					"tokenUrl": "https://auth.example.com/token",
				},
			},
		},
	}

	result, err := schema.LoadAndValidate(s, config, schema.ReadMode)
	require.NoError(t, err)
	assert.Equal(t, "https://auth.example.com/token", result.Values["oauth.tokenUrl"].Value)
}

// ============================================================
// Defaults
// ============================================================

func TestLoadAndValidate_AppliesDefault(t *testing.T) {
	s := minimalSchema(schema.ConfigField{
		ID: "timeout", Key: "timeout", ValueType: schema.NumberType,
		Target: ptr(schema.JSONDataTarget), DefaultValue: float64(30),
	})

	config := schema.DatasourceConfig{
		JSONData: map[string]any{},
	}

	result, err := schema.LoadAndValidate(s, config, schema.ReadMode)
	require.NoError(t, err)
	assert.Equal(t, float64(30), result.Values["timeout"].Value)
	assert.Equal(t, schema.SourceDefault, result.Values["timeout"].Source)
}

func TestLoadAndValidate_ConfigOverridesDefault(t *testing.T) {
	s := minimalSchema(schema.ConfigField{
		ID: "timeout", Key: "timeout", ValueType: schema.NumberType,
		Target: ptr(schema.JSONDataTarget), DefaultValue: float64(30),
	})

	config := schema.DatasourceConfig{
		JSONData: map[string]any{"timeout": float64(60)},
	}

	result, err := schema.LoadAndValidate(s, config, schema.ReadMode)
	require.NoError(t, err)
	assert.Equal(t, float64(60), result.Values["timeout"].Value)
	assert.Equal(t, schema.SourceConfig, result.Values["timeout"].Source)
}

// ============================================================
// Required validation
// ============================================================

func TestLoadAndValidate_RequiredFieldMissing(t *testing.T) {
	s := minimalSchema(schema.ConfigField{
		ID: "url", Key: "url", ValueType: schema.StringType,
		Target: ptr(schema.RootTarget), Required: true,
	})

	config := schema.DatasourceConfig{Root: map[string]any{}}

	result, err := schema.LoadAndValidate(s, config, schema.ReadMode)
	require.NoError(t, err)
	assert.True(t, result.HasErrors())
	assert.Equal(t, "required", result.Errors[0].Code)
	assert.Equal(t, "url", result.Errors[0].FieldID)
}

// ============================================================
// Type validation
// ============================================================

func TestLoadAndValidate_TypeMismatch(t *testing.T) {
	s := minimalSchema(schema.ConfigField{
		ID: "timeout", Key: "timeout", ValueType: schema.NumberType,
		Target: ptr(schema.JSONDataTarget),
	})

	config := schema.DatasourceConfig{
		JSONData: map[string]any{"timeout": "not-a-number"},
	}

	result, err := schema.LoadAndValidate(s, config, schema.ReadMode)
	require.NoError(t, err)
	assert.True(t, result.HasErrors())
	assert.Equal(t, "type_mismatch", result.Errors[0].Code)
}

// ============================================================
// Validation rules
// ============================================================

func TestLoadAndValidate_PatternValidation(t *testing.T) {
	s := minimalSchema(schema.ConfigField{
		ID: "url", Key: "url", ValueType: schema.StringType,
		Target: ptr(schema.RootTarget),
		Validations: []schema.FieldValidationRule{
			{Type: schema.PatternValidation, Pattern: "^https?://"},
		},
	})

	t.Run("valid", func(t *testing.T) {
		config := schema.DatasourceConfig{Root: map[string]any{"url": "https://example.com"}}
		result, err := schema.LoadAndValidate(s, config, schema.ReadMode)
		require.NoError(t, err)
		assert.False(t, result.HasErrors())
	})

	t.Run("invalid", func(t *testing.T) {
		config := schema.DatasourceConfig{Root: map[string]any{"url": "ftp://example.com"}}
		result, err := schema.LoadAndValidate(s, config, schema.ReadMode)
		require.NoError(t, err)
		assert.True(t, result.HasErrors())
		assert.Equal(t, "pattern", result.Errors[0].Code)
	})
}

func TestLoadAndValidate_RangeValidation(t *testing.T) {
	s := minimalSchema(schema.ConfigField{
		ID: "timeout", Key: "timeout", ValueType: schema.NumberType,
		Target: ptr(schema.JSONDataTarget),
		Validations: []schema.FieldValidationRule{
			{Type: schema.RangeValidation, Min: ptr(1.0), Max: ptr(300.0)},
		},
	})

	t.Run("in range", func(t *testing.T) {
		config := schema.DatasourceConfig{JSONData: map[string]any{"timeout": float64(30)}}
		result, _ := schema.LoadAndValidate(s, config, schema.ReadMode)
		assert.False(t, result.HasErrors())
	})

	t.Run("below min", func(t *testing.T) {
		config := schema.DatasourceConfig{JSONData: map[string]any{"timeout": float64(0)}}
		result, _ := schema.LoadAndValidate(s, config, schema.ReadMode)
		assert.True(t, result.HasErrors())
		assert.Equal(t, "range", result.Errors[0].Code)
	})

	t.Run("above max", func(t *testing.T) {
		config := schema.DatasourceConfig{JSONData: map[string]any{"timeout": float64(999)}}
		result, _ := schema.LoadAndValidate(s, config, schema.ReadMode)
		assert.True(t, result.HasErrors())
	})
}

func TestLoadAndValidate_AllowedValuesValidation(t *testing.T) {
	s := minimalSchema(schema.ConfigField{
		ID: "method", Key: "httpMethod", ValueType: schema.StringType,
		Target: ptr(schema.JSONDataTarget),
		Validations: []schema.FieldValidationRule{
			{Type: schema.AllowedValuesValidation, Values: []any{"GET", "POST"}},
		},
	})

	t.Run("allowed", func(t *testing.T) {
		config := schema.DatasourceConfig{JSONData: map[string]any{"httpMethod": "GET"}}
		result, _ := schema.LoadAndValidate(s, config, schema.ReadMode)
		assert.False(t, result.HasErrors())
	})

	t.Run("not allowed", func(t *testing.T) {
		config := schema.DatasourceConfig{JSONData: map[string]any{"httpMethod": "DELETE"}}
		result, _ := schema.LoadAndValidate(s, config, schema.ReadMode)
		assert.True(t, result.HasErrors())
		assert.Equal(t, "allowedValues", result.Errors[0].Code)
	})
}

func TestLoadAndValidate_ItemCountValidation(t *testing.T) {
	s := minimalSchema(schema.ConfigField{
		ID: "tags", Key: "tags", ValueType: schema.ArrayType,
		Target: ptr(schema.JSONDataTarget),
		Item:   &schema.FieldItemSchema{ValueType: schema.StringType},
		Validations: []schema.FieldValidationRule{
			{Type: schema.ItemCountValidation, Max: ptr(3.0)},
		},
	})

	t.Run("within limit", func(t *testing.T) {
		config := schema.DatasourceConfig{JSONData: map[string]any{"tags": []any{"a", "b"}}}
		result, _ := schema.LoadAndValidate(s, config, schema.ReadMode)
		assert.False(t, result.HasErrors())
	})

	t.Run("exceeds limit", func(t *testing.T) {
		config := schema.DatasourceConfig{JSONData: map[string]any{"tags": []any{"a", "b", "c", "d"}}}
		result, _ := schema.LoadAndValidate(s, config, schema.ReadMode)
		assert.True(t, result.HasErrors())
		assert.Equal(t, "itemCount", result.Errors[0].Code)
	})
}

// ============================================================
// Secure fields
// ============================================================

func TestLoadAndValidate_SecureField_ReadMode(t *testing.T) {
	s := minimalSchema(schema.ConfigField{
		ID: "password", Key: "password", ValueType: schema.StringType,
		Target: ptr(schema.SecureJSONTarget), Required: true,
	})

	t.Run("configured", func(t *testing.T) {
		config := schema.DatasourceConfig{
			SecureJSONFields: map[string]bool{"password": true},
		}
		result, _ := schema.LoadAndValidate(s, config, schema.ReadMode)
		assert.False(t, result.HasErrors())
		assert.Equal(t, schema.SecureConfigured, result.SecureFields["password"])
	})

	t.Run("missing required", func(t *testing.T) {
		config := schema.DatasourceConfig{
			SecureJSONFields: map[string]bool{},
		}
		result, _ := schema.LoadAndValidate(s, config, schema.ReadMode)
		assert.True(t, result.HasErrors())
		assert.Equal(t, schema.SecureUnset, result.SecureFields["password"])
	})
}

func TestLoadAndValidate_SecureField_WriteMode(t *testing.T) {
	s := minimalSchema(schema.ConfigField{
		ID: "password", Key: "password", ValueType: schema.StringType,
		Target: ptr(schema.SecureJSONTarget), Required: true,
	})

	t.Run("provided", func(t *testing.T) {
		config := schema.DatasourceConfig{
			SecureJSONData: map[string]any{"password": "s3cret"},
		}
		result, _ := schema.LoadAndValidate(s, config, schema.WriteMode)
		assert.False(t, result.HasErrors())
		assert.Equal(t, schema.SecureUpdated, result.SecureFields["password"])
		assert.Equal(t, "s3cret", result.Values["password"].Value)
	})

	t.Run("missing required", func(t *testing.T) {
		config := schema.DatasourceConfig{
			SecureJSONData: map[string]any{},
		}
		result, _ := schema.LoadAndValidate(s, config, schema.WriteMode)
		assert.True(t, result.HasErrors())
		assert.Equal(t, schema.SecureUnset, result.SecureFields["password"])
	})
}

// ============================================================
// Virtual fields are skipped
// ============================================================

func TestLoadAndValidate_SkipsVirtualFields(t *testing.T) {
	s := minimalSchema(
		schema.ConfigField{
			ID: "url", Key: "url", ValueType: schema.StringType,
			Target: ptr(schema.RootTarget),
		},
		schema.ConfigField{
			ID: "derived", Key: "derived", ValueType: schema.BooleanType,
			Kind: schema.VirtualField,
		},
	)

	config := schema.DatasourceConfig{Root: map[string]any{"url": "https://example.com"}}
	result, err := schema.LoadAndValidate(s, config, schema.ReadMode)
	require.NoError(t, err)
	assert.Contains(t, result.Values, "url")
	assert.NotContains(t, result.Values, "derived")
}

// ============================================================
// IndexedPair expansion
// ============================================================

func TestLoadAndValidate_IndexedPairExpansion(t *testing.T) {
	trueVal := true
	s := minimalSchema(schema.ConfigField{
		ID: "httpHeaders", Key: "httpHeaders", ValueType: schema.ArrayType,
		Target: ptr(schema.JSONDataTarget),
		Item: &schema.FieldItemSchema{
			ValueType: schema.ObjectType,
			Fields: []schema.ConfigField{
				{ID: "httpHeaders.item.name", Key: "name", ValueType: schema.StringType, IsItemField: &trueVal},
				{ID: "httpHeaders.item.value", Key: "value", ValueType: schema.StringType, IsItemField: &trueVal},
			},
		},
		Storage: &schema.StorageMapping{
			Type:       schema.IndexedPairMapping,
			Key:        &schema.MappingField{Target: schema.JSONDataTarget, Pattern: "httpHeaderName{index}"},
			Value:      &schema.MappingField{Target: schema.SecureJSONTarget, Pattern: "httpHeaderValue{index}"},
			StartIndex: ptr(1),
		},
	})

	config := schema.DatasourceConfig{
		JSONData: map[string]any{
			"httpHeaderName1": "X-Custom",
			"httpHeaderName2": "X-Token",
		},
		SecureJSONData: map[string]any{
			"httpHeaderValue1": "value1",
			"httpHeaderValue2": "value2",
		},
	}

	result, err := schema.LoadAndValidate(s, config, schema.WriteMode)
	require.NoError(t, err)
	assert.False(t, result.HasErrors())

	arr, ok := result.Values["httpHeaders"].Value.([]any)
	require.True(t, ok)
	require.Len(t, arr, 2)

	item0 := arr[0].(map[string]any)
	assert.Equal(t, "X-Custom", item0["name"])
	assert.Equal(t, "value1", item0["value"])

	item1 := arr[1].(map[string]any)
	assert.Equal(t, "X-Token", item1["name"])
	assert.Equal(t, "value2", item1["value"])
}

func TestLoadAndValidate_IndexedPairEmpty(t *testing.T) {
	trueVal := true
	s := minimalSchema(schema.ConfigField{
		ID: "httpHeaders", Key: "httpHeaders", ValueType: schema.ArrayType,
		Target: ptr(schema.JSONDataTarget),
		Item: &schema.FieldItemSchema{
			ValueType: schema.ObjectType,
			Fields: []schema.ConfigField{
				{ID: "httpHeaders.item.name", Key: "name", ValueType: schema.StringType, IsItemField: &trueVal},
				{ID: "httpHeaders.item.value", Key: "value", ValueType: schema.StringType, IsItemField: &trueVal},
			},
		},
		Storage: &schema.StorageMapping{
			Type:       schema.IndexedPairMapping,
			Key:        &schema.MappingField{Target: schema.JSONDataTarget, Pattern: "httpHeaderName{index}"},
			Value:      &schema.MappingField{Target: schema.SecureJSONTarget, Pattern: "httpHeaderValue{index}"},
			StartIndex: ptr(1),
		},
	})

	config := schema.DatasourceConfig{
		JSONData: map[string]any{},
	}

	result, err := schema.LoadAndValidate(s, config, schema.ReadMode)
	require.NoError(t, err)
	assert.Nil(t, result.Values["httpHeaders"].Value)
	assert.Equal(t, schema.SourceNone, result.Values["httpHeaders"].Source)
}

// ============================================================
// File-based integration tests using existing testdata
// ============================================================

func TestLoadAndValidate_FromTestdata(t *testing.T) {
	cases := []struct {
		dir      string
		mode     schema.LoadMode
		hasError bool
	}{
		{dir: "simple-url", mode: schema.ReadMode},
		{dir: "root-jsondata-secure-mix", mode: schema.WriteMode},
		{dir: "bearer-token-auth", mode: schema.WriteMode},
		{dir: "indexed-headers-storage", mode: schema.WriteMode},
		{dir: "nested-object-jsondata", mode: schema.ReadMode},
	}

	for _, tc := range cases {
		t.Run(tc.dir, func(t *testing.T) {
			dir := filepath.Join("testdata", "convert", tc.dir)

			inputBytes, err := os.ReadFile(filepath.Join(dir, "input.json"))
			require.NoError(t, err)

			configBytes, err := os.ReadFile(filepath.Join(dir, "config.json"))
			require.NoError(t, err)

			var s schema.DatasourceConfigSchema
			require.NoError(t, json.Unmarshal(inputBytes, &s))

			var rawConfig map[string]any
			require.NoError(t, json.Unmarshal(configBytes, &rawConfig))
			delete(rawConfig, "_comment")

			config := schema.NewDatasourceConfig(rawConfig)

			result, err := schema.LoadAndValidate(&s, config, tc.mode)
			require.NoError(t, err)

			if tc.hasError {
				assert.True(t, result.HasErrors(), "expected errors for %s", tc.dir)
			} else {
				assert.False(t, result.HasErrors(),
					"unexpected errors for %s: %v", tc.dir, result.Errors)
			}

			// Verify at least some values were extracted
			assert.NotEmpty(t, result.Values, "expected values for %s", tc.dir)
		})
	}
}
