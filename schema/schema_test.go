package schema_test

import (
	"encoding/json"
	"testing"

	"github.com/grafana/dsconfig/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ============================================================
// DatasourceConfigSchema.Validate
// ============================================================

// TestSchemaValidate_MinimalValid verifies that the simplest possible
// schema (one storage field) passes validation without errors.
func TestSchemaValidate_MinimalValid(t *testing.T) {
	s := minimalSchema(validStorageField("url", "url"))
	require.NoError(t, s.Validate())
}

// TestSchemaValidate_EmptyFields confirms that a schema with zero
// fields is valid — plugins may start with no config.
func TestSchemaValidate_EmptyFields(t *testing.T) {
	s := minimalSchema()
	require.NoError(t, s.Validate())
}

// TestSchemaValidate_PropagatesFieldError ensures that a validation
// error on an individual field bubbles up through the root Validate().
func TestSchemaValidate_PropagatesFieldError(t *testing.T) {
	s := minimalSchema(schema.ConfigField{ID: "", Key: "x", ValueType: schema.StringType})
	assert.ErrorContains(t, s.Validate(), "field id is required")
}

// ============================================================
// DatasourceConfigSchema.FieldIDs
// ============================================================

// TestFieldIDs_CollectsTopLevel verifies that FieldIDs returns all
// top-level field IDs as a set.
func TestFieldIDs_CollectsTopLevel(t *testing.T) {
	s := minimalSchema(
		validStorageField("a", "a"),
		validStorageField("b", "b"),
	)
	ids, err := s.FieldIDs()
	require.NoError(t, err)
	assert.Equal(t, map[string]bool{"a": true, "b": true}, ids)
}

// TestFieldIDs_CollectsItemFields verifies that FieldIDs recursively
// collects IDs from nested array item fields.
func TestFieldIDs_CollectsItemFields(t *testing.T) {
	s := minimalSchema(schema.ConfigField{
		ID:        "headers",
		Key:       "headers",
		ValueType: schema.ArrayType,
		Target:    ptr(schema.JSONDataTarget),
		Item: &schema.FieldItemSchema{
			ValueType: schema.ObjectType,
			Fields: []schema.ConfigField{
				{ID: "headers.item.key", Key: "key", ValueType: schema.StringType, IsItemField: ptr(true)},
				{ID: "headers.item.value", Key: "value", ValueType: schema.StringType, IsItemField: ptr(true)},
			},
		},
	})
	ids, err := s.FieldIDs()
	require.NoError(t, err)
	assert.True(t, ids["headers"])
	assert.True(t, ids["headers.item.key"])
	assert.True(t, ids["headers.item.value"])
}

// TestFieldIDs_DuplicateID ensures that two top-level fields sharing
// the same ID are rejected.
func TestFieldIDs_DuplicateID(t *testing.T) {
	s := minimalSchema(
		validStorageField("dup", "a"),
		validStorageField("dup", "b"),
	)
	_, err := s.FieldIDs()
	assert.ErrorContains(t, err, "duplicate field id: dup")
}

// TestFieldIDs_EmptyID ensures that a field with an empty ID is caught.
func TestFieldIDs_EmptyID(t *testing.T) {
	s := minimalSchema(schema.ConfigField{Key: "x", ValueType: schema.StringType})
	_, err := s.FieldIDs()
	assert.ErrorContains(t, err, "field id is required")
}

// TestFieldIDs_DuplicateBetweenTopAndItem verifies that an item field
// cannot reuse a top-level field's ID (global uniqueness).
func TestFieldIDs_DuplicateBetweenTopAndItem(t *testing.T) {
	s := minimalSchema(
		validStorageField("conflict", "x"),
		schema.ConfigField{
			ID: "arr", Key: "arr", ValueType: schema.ArrayType, Target: ptr(schema.JSONDataTarget),
			Item: &schema.FieldItemSchema{
				ValueType: schema.ObjectType,
				Fields: []schema.ConfigField{
					{ID: "conflict", Key: "k", ValueType: schema.StringType, IsItemField: ptr(true)},
				},
			},
		},
	)
	_, err := s.FieldIDs()
	assert.ErrorContains(t, err, "duplicate field id: conflict")
}

// ============================================================
// DatasourceConfigSchema.ValidateRefs
// ============================================================

// TestValidateRefs_ValidGroupRefs ensures groups referencing existing
// field IDs pass validation.
func TestValidateRefs_ValidGroupRefs(t *testing.T) {
	s := minimalSchema(
		validStorageField("a", "a"),
		validStorageField("b", "b"),
	)
	s.Groups = []schema.ConfigGroup{{ID: "g1", Title: "G", FieldRefs: []string{"a", "b"}}}
	require.NoError(t, s.Validate())
}

// TestValidateRefs_InvalidGroupRef ensures a group referencing a
// non-existent field ID is rejected.
func TestValidateRefs_InvalidGroupRef(t *testing.T) {
	s := minimalSchema(validStorageField("a", "a"))
	s.Groups = []schema.ConfigGroup{{ID: "g1", Title: "G", FieldRefs: []string{"missing"}}}
	assert.ErrorContains(t, s.Validate(), "group g1 references unknown field id: missing")
}

// TestValidateRefs_ValidRelationshipRefs ensures relationships
// referencing existing field IDs pass validation.
func TestValidateRefs_ValidRelationshipRefs(t *testing.T) {
	s := minimalSchema(
		validStorageField("user", "user"),
		validStorageField("pass", "pass"),
	)
	s.Relationships = []schema.FieldRelationship{
		{Type: schema.PairRelationship, Fields: []string{"user", "pass"}},
	}
	require.NoError(t, s.Validate())
}

// TestValidateRefs_InvalidRelationshipRef ensures a relationship
// referencing a non-existent field ID is rejected.
func TestValidateRefs_InvalidRelationshipRef(t *testing.T) {
	s := minimalSchema(validStorageField("a", "a"))
	s.Relationships = []schema.FieldRelationship{
		{Type: schema.PairRelationship, Fields: []string{"a", "ghost"}},
	}
	assert.ErrorContains(t, s.Validate(), "relationship references unknown field id: ghost")
}

// TestValidateRefs_GroupRefToItemField verifies that groups can reference
// nested item field IDs (not just top-level).
func TestValidateRefs_GroupRefToItemField(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1",
		PluginType:    "test",
		PluginName:    "Test",
		Fields: []schema.ConfigField{
			{
				ID: "arr", Key: "arr", ValueType: schema.ArrayType, Target: ptr(schema.JSONDataTarget),
				Item: &schema.FieldItemSchema{
					ValueType: schema.ObjectType,
					Fields: []schema.ConfigField{
						{ID: "arr.item.name", Key: "name", ValueType: schema.StringType, IsItemField: ptr(true)},
					},
				},
			},
		},
		Groups: []schema.ConfigGroup{
			{ID: "g1", Title: "G", FieldRefs: []string{"arr.item.name"}},
		},
	}
	require.NoError(t, s.Validate())
}

// ============================================================
// ConfigField.Validate — identity fields
// ============================================================

// TestFieldValidate_EmptyID ensures that a field without an ID is
// rejected, since ID is the primary reference key.
func TestFieldValidate_EmptyID(t *testing.T) {
	f := schema.ConfigField{Key: "x", ValueType: schema.StringType, Target: ptr(schema.JSONDataTarget)}
	assert.ErrorContains(t, f.Validate(), "field id is required")
}

// TestFieldValidate_EmptyKey ensures that a field without a key is
// rejected, since key is required for storage mapping.
func TestFieldValidate_EmptyKey(t *testing.T) {
	f := schema.ConfigField{ID: "x", ValueType: schema.StringType, Target: ptr(schema.JSONDataTarget)}
	assert.ErrorContains(t, f.Validate(), "key is required")
}

// ============================================================
// ConfigField.Validate — valueType
// ============================================================

// TestFieldValidate_InvalidValueType ensures that an unrecognized
// valueType string is rejected.
func TestFieldValidate_InvalidValueType(t *testing.T) {
	f := schema.ConfigField{ID: "x", Key: "x", ValueType: "blob", Target: ptr(schema.JSONDataTarget)}
	assert.ErrorContains(t, f.Validate(), "invalid valueType")
}

// TestFieldValidate_AllValueTypes verifies that every valid ValueType
// constant passes field validation.
func TestFieldValidate_AllValueTypes(t *testing.T) {
	for _, vt := range []schema.ValueType{
		schema.StringType, schema.NumberType, schema.BooleanType,
		schema.ArrayType, schema.ObjectType, schema.MapType, schema.AnyType,
	} {
		f := validStorageField("x", "x")
		f.ValueType = vt
		if vt == schema.ArrayType || vt == schema.MapType {
			f.Item = &schema.FieldItemSchema{ValueType: schema.StringType}
		}
		assert.NoError(t, f.Validate(), "valueType %s should be valid", vt)
	}
}

// ============================================================
// ConfigField.Validate — target requirement
// ============================================================

// TestFieldValidate_StorageFieldRequiresTarget verifies that a storage
// field (default kind) without a target is rejected.
func TestFieldValidate_StorageFieldRequiresTarget(t *testing.T) {
	f := schema.ConfigField{ID: "x", Key: "x", ValueType: schema.StringType}
	assert.ErrorContains(t, f.Validate(), "target is required for storage fields")
}

// TestFieldValidate_VirtualFieldOmitsTarget confirms that virtual
// fields do not require a target.
func TestFieldValidate_VirtualFieldOmitsTarget(t *testing.T) {
	f := schema.ConfigField{ID: "x", Key: "x", ValueType: schema.StringType, Kind: schema.VirtualField}
	require.NoError(t, f.Validate())
}

// TestFieldValidate_ItemFieldOmitsTarget confirms that item fields
// (isItemField=true) do not require a target.
func TestFieldValidate_ItemFieldOmitsTarget(t *testing.T) {
	f := schema.ConfigField{ID: "x", Key: "x", ValueType: schema.StringType, IsItemField: ptr(true)}
	require.NoError(t, f.Validate())
}

// TestFieldValidate_SectionOnItemFieldRejected ensures that section
// is not allowed on item fields (they inherit storage from the parent).
func TestFieldValidate_SectionOnItemFieldRejected(t *testing.T) {
	f := schema.ConfigField{
		ID: "x", Key: "x", ValueType: schema.StringType,
		IsItemField: ptr(true), Section: "nested",
	}
	assert.ErrorContains(t, f.Validate(), "section is not allowed on item fields")
}

// TestFieldValidate_SectionOnVirtualFieldRejected ensures that section
// is not allowed on virtual fields (they have no storage target).
func TestFieldValidate_SectionOnVirtualFieldRejected(t *testing.T) {
	f := schema.ConfigField{
		ID: "x", Key: "x", ValueType: schema.StringType,
		Kind: schema.VirtualField, Section: "nested",
	}
	assert.ErrorContains(t, f.Validate(), "section is not allowed on virtual fields")
}

// TestFieldValidate_SectionOnStorageFieldAllowed confirms that section
// is valid on a normal storage field with a target.
func TestFieldValidate_SectionOnStorageFieldAllowed(t *testing.T) {
	f := schema.ConfigField{
		ID: "x", Key: "x", ValueType: schema.StringType,
		Target: ptr(schema.JSONDataTarget), Section: "oauth2.endpoints",
	}
	require.NoError(t, f.Validate())
}

// TestFieldValidate_InvalidTarget ensures that an unrecognized target
// location string is rejected.
func TestFieldValidate_InvalidTarget(t *testing.T) {
	bad := schema.TargetLocation("badTarget")
	f := schema.ConfigField{ID: "x", Key: "x", ValueType: schema.StringType, Target: &bad}
	assert.ErrorContains(t, f.Validate(), "invalid target")
}

// TestFieldValidate_AllTargets verifies that every valid TargetLocation
// constant passes field validation.
func TestFieldValidate_AllTargets(t *testing.T) {
	for _, tgt := range []schema.TargetLocation{
		schema.RootTarget, schema.JSONDataTarget, schema.SecureJSONTarget,
	} {
		f := schema.ConfigField{ID: "x", Key: "x", ValueType: schema.StringType, Target: &tgt}
		assert.NoError(t, f.Validate(), "target %s should be valid", tgt)
	}
}

// ============================================================
// ConfigField.Validate — kind
// ============================================================

// TestFieldValidate_InvalidKind ensures that an unrecognized kind
// string is rejected even if target is provided.
func TestFieldValidate_InvalidKind(t *testing.T) {
	f := schema.ConfigField{
		ID: "x", Key: "x", ValueType: schema.StringType,
		Kind: "unknown", Target: ptr(schema.JSONDataTarget),
	}
	assert.ErrorContains(t, f.Validate(), "invalid kind")
}

// TestFieldValidate_ValidKinds verifies that both storage and virtual
// field kinds pass validation when properly configured.
func TestFieldValidate_ValidKinds(t *testing.T) {
	for _, k := range []schema.FieldKind{schema.StorageField, schema.VirtualField} {
		f := schema.ConfigField{ID: "x", Key: "x", ValueType: schema.StringType, Kind: k}
		if k == schema.StorageField {
			f.Target = ptr(schema.JSONDataTarget)
		}
		assert.NoError(t, f.Validate(), "kind %s should be valid", k)
	}
}

// ============================================================
// ConfigField.Validate — array / item
// ============================================================

// TestFieldValidate_ArrayRequiresItem ensures that an array field
// without an item schema is rejected.
func TestFieldValidate_ArrayRequiresItem(t *testing.T) {
	f := schema.ConfigField{
		ID: "x", Key: "x", ValueType: schema.ArrayType, Target: ptr(schema.JSONDataTarget),
	}
	assert.ErrorContains(t, f.Validate(), "item is required for array and map fields")
}

// TestFieldValidate_MapRequiresItem ensures that a map field
// without an item schema is rejected.
func TestFieldValidate_MapRequiresItem(t *testing.T) {
	f := schema.ConfigField{
		ID: "x", Key: "x", ValueType: schema.MapType, Target: ptr(schema.JSONDataTarget),
	}
	assert.ErrorContains(t, f.Validate(), "item is required for array and map fields")
}

// TestFieldValidate_MapWithStringItem confirms that a map field
// with a string item schema passes (Record<string, string>).
func TestFieldValidate_MapWithStringItem(t *testing.T) {
	f := schema.ConfigField{
		ID: "x", Key: "x", ValueType: schema.MapType, Target: ptr(schema.JSONDataTarget),
		Item: &schema.FieldItemSchema{ValueType: schema.StringType},
	}
	require.NoError(t, f.Validate())
}

// TestFieldValidate_MapWithObjectItem confirms that a map field
// with an object item schema passes (Record<string, SomeObj>).
func TestFieldValidate_MapWithObjectItem(t *testing.T) {
	f := schema.ConfigField{
		ID: "routes", Key: "routes", ValueType: schema.MapType, Target: ptr(schema.JSONDataTarget),
		Item: &schema.FieldItemSchema{
			ValueType: schema.ObjectType,
			Fields: []schema.ConfigField{
				{ID: "routes.item.url", Key: "url", ValueType: schema.StringType, IsItemField: ptr(true)},
			},
		},
	}
	require.NoError(t, f.Validate())
}

// TestFieldValidate_AnyFieldValid confirms that an any-typed field
// passes validation without item.
func TestFieldValidate_AnyFieldValid(t *testing.T) {
	f := schema.ConfigField{
		ID: "x", Key: "x", ValueType: schema.AnyType, Target: ptr(schema.JSONDataTarget),
	}
	require.NoError(t, f.Validate())
}

// TestFieldValidate_AnyItemFieldValid confirms that any-typed item
// fields pass validation.
func TestFieldValidate_AnyItemFieldValid(t *testing.T) {
	f := schema.ConfigField{
		ID: "x", Key: "x", ValueType: schema.AnyType, IsItemField: ptr(true),
	}
	require.NoError(t, f.Validate())
}

// TestFieldValidate_ArrayWithItem confirms that an array field
// with a valid item schema passes.
func TestFieldValidate_ArrayWithItem(t *testing.T) {
	f := schema.ConfigField{
		ID: "x", Key: "x", ValueType: schema.ArrayType, Target: ptr(schema.JSONDataTarget),
		Item: &schema.FieldItemSchema{ValueType: schema.StringType},
	}
	require.NoError(t, f.Validate())
}

// TestFieldValidate_ItemInvalidValueType ensures that an item schema
// with an unrecognized valueType is rejected.
func TestFieldValidate_ItemInvalidValueType(t *testing.T) {
	f := schema.ConfigField{
		ID: "x", Key: "x", ValueType: schema.ArrayType, Target: ptr(schema.JSONDataTarget),
		Item: &schema.FieldItemSchema{ValueType: "invalid"},
	}
	assert.ErrorContains(t, f.Validate(), "invalid item valueType")
}

// TestFieldValidate_ItemFieldsOnlyForObject ensures that item.fields
// are only allowed when item.valueType is "object". A string array
// with nested fields should be rejected.
func TestFieldValidate_ItemFieldsOnlyForObject(t *testing.T) {
	f := schema.ConfigField{
		ID: "x", Key: "x", ValueType: schema.ArrayType, Target: ptr(schema.JSONDataTarget),
		Item: &schema.FieldItemSchema{
			ValueType: schema.StringType,
			Fields: []schema.ConfigField{
				{ID: "sub", Key: "sub", ValueType: schema.StringType, IsItemField: ptr(true)},
			},
		},
	}
	assert.ErrorContains(t, f.Validate(), "item fields are only allowed when item valueType is object")
}

// TestFieldValidate_ItemFieldMustHaveIsItemField ensures that every
// field inside item.fields must have isItemField=true.
func TestFieldValidate_ItemFieldMustHaveIsItemField(t *testing.T) {
	f := schema.ConfigField{
		ID: "x", Key: "x", ValueType: schema.ArrayType, Target: ptr(schema.JSONDataTarget),
		Item: &schema.FieldItemSchema{
			ValueType: schema.ObjectType,
			Fields: []schema.ConfigField{
				{ID: "sub", Key: "sub", ValueType: schema.StringType},
			},
		},
	}
	assert.ErrorContains(t, f.Validate(), "must have isItemField=true")
}

// TestFieldValidate_ItemFieldValidationPropagates ensures that
// validation errors in nested item fields bubble up through the
// parent field's Validate().
func TestFieldValidate_ItemFieldValidationPropagates(t *testing.T) {
	f := schema.ConfigField{
		ID: "x", Key: "x", ValueType: schema.ArrayType, Target: ptr(schema.JSONDataTarget),
		Item: &schema.FieldItemSchema{
			ValueType: schema.ObjectType,
			Fields: []schema.ConfigField{
				{ID: "sub", Key: "", ValueType: schema.StringType, IsItemField: ptr(true)},
			},
		},
	}
	assert.ErrorContains(t, f.Validate(), "key is required")
}

// TestFieldValidate_ObjectItemWithValidFields confirms that a
// well-formed array-of-objects field passes validation.
func TestFieldValidate_ObjectItemWithValidFields(t *testing.T) {
	f := schema.ConfigField{
		ID: "headers", Key: "headers", ValueType: schema.ArrayType, Target: ptr(schema.JSONDataTarget),
		Item: &schema.FieldItemSchema{
			ValueType: schema.ObjectType,
			Fields: []schema.ConfigField{
				{ID: "headers.key", Key: "key", ValueType: schema.StringType, IsItemField: ptr(true)},
				{ID: "headers.val", Key: "val", ValueType: schema.StringType, IsItemField: ptr(true)},
			},
		},
	}
	require.NoError(t, f.Validate())
}

// ============================================================
// ConfigField.Validate — validation rules
// ============================================================

// TestFieldValidate_ValidValidationRules confirms that a field with
// multiple well-formed validation rules passes.
func TestFieldValidate_ValidValidationRules(t *testing.T) {
	f := validStorageField("x", "x")
	f.Validations = []schema.FieldValidationRule{
		{Type: schema.PatternValidation, Pattern: "^https?://"},
		{Type: schema.RangeValidation, Min: ptr(0.0), Max: ptr(100.0)},
	}
	require.NoError(t, f.Validate())
}

// TestFieldValidate_InvalidValidationRule ensures that an invalid
// validation rule (missing required fields) causes the parent field
// validation to fail.
func TestFieldValidate_InvalidValidationRule(t *testing.T) {
	f := validStorageField("x", "x")
	f.Validations = []schema.FieldValidationRule{
		{Type: schema.PatternValidation}, // missing pattern
	}
	assert.ErrorContains(t, f.Validate(), "pattern validation requires pattern")
}

// TestFieldValidate_OverrideValidationRulePropagates ensures that
// invalid validation rules inside overrides are caught.
func TestFieldValidate_OverrideValidationRulePropagates(t *testing.T) {
	f := validStorageField("x", "x")
	f.Overrides = []schema.FieldOverride{
		{
			When: "authType == 'basic'",
			Validations: []schema.FieldValidationRule{
				{Type: schema.CustomValidation}, // missing expression
			},
		},
	}
	assert.ErrorContains(t, f.Validate(), "custom validation requires expression")
}

// ============================================================
// ConfigField.Validate — storage mapping integration
// ============================================================

// TestFieldValidate_DirectStorageMapping confirms that a field with
// a valid direct storage mapping passes.
func TestFieldValidate_DirectStorageMapping(t *testing.T) {
	f := validStorageField("x", "x")
	f.Storage = &schema.StorageMapping{Type: schema.DirectMapping}
	require.NoError(t, f.Validate())
}

// TestFieldValidate_InvalidStorageMapping ensures that a field with
// an invalid storage mapping (e.g. computed with no read/write)
// causes validation to fail.
func TestFieldValidate_InvalidStorageMapping(t *testing.T) {
	f := validStorageField("x", "x")
	f.Storage = &schema.StorageMapping{Type: schema.ComputedMapping} // missing read/write
	assert.ErrorContains(t, f.Validate(), "computed mapping requires read or write")
}

// TestFieldValidate_ComputedStorageMappingOnVirtual confirms that
// a virtual field with a computed storage mapping passes.
func TestFieldValidate_ComputedStorageMappingOnVirtual(t *testing.T) {
	f := schema.ConfigField{
		ID: "derived", Key: "derived", ValueType: schema.StringType, Kind: schema.VirtualField,
		Storage: &schema.StorageMapping{Type: schema.ComputedMapping, Read: "jsonData.a + jsonData.b"},
	}
	require.NoError(t, f.Validate())
}

// ============================================================
// ConfigField.Path
// ============================================================

// TestFieldPath_WithTarget verifies that Path() returns
// "target.key" when a target is set.
func TestFieldPath_WithTarget(t *testing.T) {
	f := schema.ConfigField{Target: ptr(schema.JSONDataTarget), Key: "timeout"}
	assert.Equal(t, "jsonData.timeout", f.Path())
}

// TestFieldPath_WithoutTarget verifies that Path() returns just
// the key when no target is set (e.g. virtual fields).
func TestFieldPath_WithoutTarget(t *testing.T) {
	f := schema.ConfigField{Key: "url"}
	assert.Equal(t, "url", f.Path())
}

// TestFieldPath_RootTarget verifies the path for root-level fields.
func TestFieldPath_RootTarget(t *testing.T) {
	f := schema.ConfigField{Target: ptr(schema.RootTarget), Key: "url"}
	assert.Equal(t, "root.url", f.Path())
}

// TestFieldPath_SecureTarget verifies the path for secure fields.
func TestFieldPath_SecureTarget(t *testing.T) {
	f := schema.ConfigField{Target: ptr(schema.SecureJSONTarget), Key: "password"}
	assert.Equal(t, "secureJsonData.password", f.Path())
}

// ============================================================
// ValueType.IsValid
// ============================================================

// TestValueType_Valid verifies that all defined ValueType constants
// are recognized as valid.
func TestValueType_Valid(t *testing.T) {
	for _, v := range []schema.ValueType{
		schema.StringType, schema.NumberType, schema.BooleanType,
		schema.ArrayType, schema.ObjectType, schema.MapType, schema.AnyType,
	} {
		assert.True(t, v.IsValid(), "%s should be valid", v)
	}
}

// TestValueType_Invalid verifies that empty strings and unknown
// type names are rejected.
func TestValueType_Invalid(t *testing.T) {
	assert.False(t, schema.ValueType("").IsValid())
	assert.False(t, schema.ValueType("int").IsValid())
	assert.False(t, schema.ValueType("union").IsValid())
}

// ============================================================
// FieldKind.IsValid
// ============================================================

// TestFieldKind_Valid verifies that storage and virtual are
// accepted as valid kinds.
func TestFieldKind_Valid(t *testing.T) {
	assert.True(t, schema.StorageField.IsValid())
	assert.True(t, schema.VirtualField.IsValid())
}

// TestFieldKind_Invalid verifies that empty strings and unknown
// kind names are rejected.
func TestFieldKind_Invalid(t *testing.T) {
	assert.False(t, schema.FieldKind("").IsValid())
	assert.False(t, schema.FieldKind("computed").IsValid())
	assert.False(t, schema.FieldKind("derived").IsValid())
}

// ============================================================
// TargetLocation.IsValid
// ============================================================

// TestTargetLocation_Valid verifies that all defined target
// location constants are recognized as valid.
func TestTargetLocation_Valid(t *testing.T) {
	for _, tgt := range []schema.TargetLocation{
		schema.RootTarget, schema.JSONDataTarget, schema.SecureJSONTarget,
	} {
		assert.True(t, tgt.IsValid(), "%s should be valid", tgt)
	}
}

// TestTargetLocation_Invalid verifies that empty strings and unknown
// target names are rejected.
func TestTargetLocation_Invalid(t *testing.T) {
	assert.False(t, schema.TargetLocation("").IsValid())
	assert.False(t, schema.TargetLocation("metadata").IsValid())
}

// ============================================================
// FieldValidationRule.Validate — pattern
// ============================================================

// TestValidationRule_Pattern_Valid confirms that a pattern rule
// with a non-empty regex string passes.
func TestValidationRule_Pattern_Valid(t *testing.T) {
	r := schema.FieldValidationRule{Type: schema.PatternValidation, Pattern: "^[a-z]+$"}
	require.NoError(t, r.Validate())
}

// TestValidationRule_Pattern_MissingPattern ensures that a pattern
// rule without a pattern string is rejected.
func TestValidationRule_Pattern_MissingPattern(t *testing.T) {
	r := schema.FieldValidationRule{Type: schema.PatternValidation}
	assert.ErrorContains(t, r.Validate(), "pattern validation requires pattern")
}

// ============================================================
// FieldValidationRule.Validate — range
// ============================================================

// TestValidationRule_Range_MinOnly verifies a range rule with
// only a minimum bound.
func TestValidationRule_Range_MinOnly(t *testing.T) {
	r := schema.FieldValidationRule{Type: schema.RangeValidation, Min: ptr(1.0)}
	require.NoError(t, r.Validate())
}

// TestValidationRule_Range_MaxOnly verifies a range rule with
// only a maximum bound.
func TestValidationRule_Range_MaxOnly(t *testing.T) {
	r := schema.FieldValidationRule{Type: schema.RangeValidation, Max: ptr(100.0)}
	require.NoError(t, r.Validate())
}

// TestValidationRule_Range_BothBounds verifies a range rule with
// both min and max.
func TestValidationRule_Range_BothBounds(t *testing.T) {
	r := schema.FieldValidationRule{Type: schema.RangeValidation, Min: ptr(1.0), Max: ptr(300.0)}
	require.NoError(t, r.Validate())
}

// TestValidationRule_Range_NeitherMinNorMax ensures that a range
// rule with no bounds is rejected.
func TestValidationRule_Range_NeitherMinNorMax(t *testing.T) {
	r := schema.FieldValidationRule{Type: schema.RangeValidation}
	assert.ErrorContains(t, r.Validate(), "range validation requires min or max")
}

// ============================================================
// FieldValidationRule.Validate — length
// ============================================================

// TestValidationRule_Length_Valid verifies a length rule with both
// min and max bounds.
func TestValidationRule_Length_Valid(t *testing.T) {
	r := schema.FieldValidationRule{Type: schema.LengthValidation, Min: ptr(1.0), Max: ptr(255.0)}
	require.NoError(t, r.Validate())
}

// TestValidationRule_Length_NeitherMinNorMax ensures that a length
// rule with no bounds is rejected.
func TestValidationRule_Length_NeitherMinNorMax(t *testing.T) {
	r := schema.FieldValidationRule{Type: schema.LengthValidation}
	assert.ErrorContains(t, r.Validate(), "length validation requires min or max")
}

// ============================================================
// FieldValidationRule.Validate — itemCount
// ============================================================

// TestValidationRule_ItemCount_Valid verifies an itemCount rule
// with a maximum.
func TestValidationRule_ItemCount_Valid(t *testing.T) {
	r := schema.FieldValidationRule{Type: schema.ItemCountValidation, Max: ptr(10.0)}
	require.NoError(t, r.Validate())
}

// TestValidationRule_ItemCount_NeitherMinNorMax ensures that an
// itemCount rule with no bounds is rejected.
func TestValidationRule_ItemCount_NeitherMinNorMax(t *testing.T) {
	r := schema.FieldValidationRule{Type: schema.ItemCountValidation}
	assert.ErrorContains(t, r.Validate(), "itemCount validation requires min or max")
}

// ============================================================
// FieldValidationRule.Validate — allowedValues
// ============================================================

// TestValidationRule_AllowedValues_Valid confirms that a rule
// with a non-empty values list passes.
func TestValidationRule_AllowedValues_Valid(t *testing.T) {
	r := schema.FieldValidationRule{
		Type: schema.AllowedValuesValidation, Values: []any{"GET", "POST"},
	}
	require.NoError(t, r.Validate())
}

// TestValidationRule_AllowedValues_Empty ensures that an empty
// values slice is rejected.
func TestValidationRule_AllowedValues_Empty(t *testing.T) {
	r := schema.FieldValidationRule{Type: schema.AllowedValuesValidation, Values: []any{}}
	assert.ErrorContains(t, r.Validate(), "allowedValues validation requires values")
}

// TestValidationRule_AllowedValues_Nil ensures that a nil values
// field is rejected (same as empty).
func TestValidationRule_AllowedValues_Nil(t *testing.T) {
	r := schema.FieldValidationRule{Type: schema.AllowedValuesValidation}
	assert.ErrorContains(t, r.Validate(), "allowedValues validation requires values")
}

// ============================================================
// FieldValidationRule.Validate — custom
// ============================================================

// TestValidationRule_Custom_Valid confirms that a custom rule
// with a non-empty CEL expression passes.
func TestValidationRule_Custom_Valid(t *testing.T) {
	r := schema.FieldValidationRule{
		Type: schema.CustomValidation, Expression: "self.startsWith('http')",
	}
	require.NoError(t, r.Validate())
}

// TestValidationRule_Custom_MissingExpression ensures that a
// custom rule without an expression is rejected.
func TestValidationRule_Custom_MissingExpression(t *testing.T) {
	r := schema.FieldValidationRule{Type: schema.CustomValidation}
	assert.ErrorContains(t, r.Validate(), "custom validation requires expression")
}

// ============================================================
// FieldValidationRule.Validate — unknown type & optional fields
// ============================================================

// TestValidationRule_UnknownType ensures that an unrecognized
// validation rule type is rejected.
func TestValidationRule_UnknownType(t *testing.T) {
	r := schema.FieldValidationRule{Type: "banana"}
	assert.ErrorContains(t, r.Validate(), "unknown validation rule type: banana")
}

// TestValidationRule_WithOptionalIDAndMessage confirms that the
// optional id and message fields do not interfere with validation.
func TestValidationRule_WithOptionalIDAndMessage(t *testing.T) {
	r := schema.FieldValidationRule{
		Type:    schema.PatternValidation,
		ID:      "url-format",
		Message: "Must be a valid URL",
		Pattern: "^https?://",
	}
	require.NoError(t, r.Validate())
}

// ============================================================
// StorageMapping.Validate — direct
// ============================================================

// TestStorageMapping_Direct_Valid confirms that a bare direct
// mapping with no extra fields passes.
func TestStorageMapping_Direct_Valid(t *testing.T) {
	m := schema.StorageMapping{Type: schema.DirectMapping}
	require.NoError(t, m.Validate())
}

// TestStorageMapping_Direct_WithRead ensures that a direct mapping
// with unexpected read/write fields is rejected.
func TestStorageMapping_Direct_WithRead(t *testing.T) {
	m := schema.StorageMapping{Type: schema.DirectMapping, Read: "something"}
	assert.ErrorContains(t, m.Validate(), "direct mapping must not have")
}

// TestStorageMapping_Direct_WithKey ensures that a direct mapping
// with unexpected key field is rejected.
func TestStorageMapping_Direct_WithKey(t *testing.T) {
	m := schema.StorageMapping{
		Type: schema.DirectMapping,
		Key:  &schema.MappingField{Target: schema.JSONDataTarget, Pattern: "x{index}"},
	}
	assert.ErrorContains(t, m.Validate(), "direct mapping must not have")
}

// TestStorageMapping_Direct_WithStartIndex ensures that a direct
// mapping with an unexpected startIndex is rejected.
func TestStorageMapping_Direct_WithStartIndex(t *testing.T) {
	m := schema.StorageMapping{Type: schema.DirectMapping, StartIndex: ptr(1)}
	assert.ErrorContains(t, m.Validate(), "direct mapping must not have")
}

// ============================================================
// StorageMapping.Validate — indexedPair
// ============================================================

// TestStorageMapping_IndexedPair_Valid confirms that a properly
// configured indexed pair mapping passes.
func TestStorageMapping_IndexedPair_Valid(t *testing.T) {
	m := schema.StorageMapping{
		Type:  schema.IndexedPairMapping,
		Key:   &schema.MappingField{Target: schema.JSONDataTarget, Pattern: "httpHeaderName{index}"},
		Value: &schema.MappingField{Target: schema.SecureJSONTarget, Pattern: "httpHeaderValue{index}"},
	}
	require.NoError(t, m.Validate())
}

// TestStorageMapping_IndexedPair_WithStartIndex confirms that
// startIndex is allowed on indexed pair mappings.
func TestStorageMapping_IndexedPair_WithStartIndex(t *testing.T) {
	m := schema.StorageMapping{
		Type:       schema.IndexedPairMapping,
		Key:        &schema.MappingField{Target: schema.JSONDataTarget, Pattern: "k{index}"},
		Value:      &schema.MappingField{Target: schema.JSONDataTarget, Pattern: "v{index}"},
		StartIndex: ptr(1),
	}
	require.NoError(t, m.Validate())
}

// TestStorageMapping_IndexedPair_MissingKey ensures that an indexed
// pair mapping without a key field is rejected.
func TestStorageMapping_IndexedPair_MissingKey(t *testing.T) {
	m := schema.StorageMapping{
		Type:  schema.IndexedPairMapping,
		Value: &schema.MappingField{Target: schema.JSONDataTarget, Pattern: "x{index}"},
	}
	assert.ErrorContains(t, m.Validate(), "indexedPair requires key and value")
}

// TestStorageMapping_IndexedPair_MissingValue ensures that an indexed
// pair mapping without a value field is rejected.
func TestStorageMapping_IndexedPair_MissingValue(t *testing.T) {
	m := schema.StorageMapping{
		Type: schema.IndexedPairMapping,
		Key:  &schema.MappingField{Target: schema.JSONDataTarget, Pattern: "x{index}"},
	}
	assert.ErrorContains(t, m.Validate(), "indexedPair requires key and value")
}

// TestStorageMapping_IndexedPair_WithRead ensures that indexed pair
// mappings with read/write (computed fields) are rejected.
func TestStorageMapping_IndexedPair_WithRead(t *testing.T) {
	m := schema.StorageMapping{
		Type:  schema.IndexedPairMapping,
		Key:   &schema.MappingField{Target: schema.JSONDataTarget, Pattern: "k{i}"},
		Value: &schema.MappingField{Target: schema.JSONDataTarget, Pattern: "v{i}"},
		Read:  "expr",
	}
	assert.ErrorContains(t, m.Validate(), "indexedPair must not have read/write")
}

// TestStorageMapping_IndexedPair_InvalidKeyTarget ensures that an
// invalid target on the key mapping field is caught.
func TestStorageMapping_IndexedPair_InvalidKeyTarget(t *testing.T) {
	m := schema.StorageMapping{
		Type:  schema.IndexedPairMapping,
		Key:   &schema.MappingField{Target: "bad", Pattern: "k{i}"},
		Value: &schema.MappingField{Target: schema.JSONDataTarget, Pattern: "v{i}"},
	}
	assert.ErrorContains(t, m.Validate(), "indexedPair key")
}

// TestStorageMapping_IndexedPair_EmptyValuePattern ensures that an
// empty pattern on the value mapping field is caught.
func TestStorageMapping_IndexedPair_EmptyValuePattern(t *testing.T) {
	m := schema.StorageMapping{
		Type:  schema.IndexedPairMapping,
		Key:   &schema.MappingField{Target: schema.JSONDataTarget, Pattern: "k{i}"},
		Value: &schema.MappingField{Target: schema.JSONDataTarget, Pattern: ""},
	}
	assert.ErrorContains(t, m.Validate(), "indexedPair value")
}

// ============================================================
// StorageMapping.Validate — computed
// ============================================================

// TestStorageMapping_Computed_ReadOnly confirms a computed mapping
// with only a read expression.
func TestStorageMapping_Computed_ReadOnly(t *testing.T) {
	m := schema.StorageMapping{Type: schema.ComputedMapping, Read: "jsonData.x + jsonData.y"}
	require.NoError(t, m.Validate())
}

// TestStorageMapping_Computed_WriteOnly confirms a computed mapping
// with only a write expression.
func TestStorageMapping_Computed_WriteOnly(t *testing.T) {
	m := schema.StorageMapping{Type: schema.ComputedMapping, Write: "split(value)"}
	require.NoError(t, m.Validate())
}

// TestStorageMapping_Computed_Both confirms a computed mapping
// with both read and write expressions.
func TestStorageMapping_Computed_Both(t *testing.T) {
	m := schema.StorageMapping{Type: schema.ComputedMapping, Read: "r", Write: "w"}
	require.NoError(t, m.Validate())
}

// TestStorageMapping_Computed_Neither ensures that a computed
// mapping with neither read nor write is rejected.
func TestStorageMapping_Computed_Neither(t *testing.T) {
	m := schema.StorageMapping{Type: schema.ComputedMapping}
	assert.ErrorContains(t, m.Validate(), "computed mapping requires read or write")
}

// TestStorageMapping_Computed_WithKey ensures that computed mappings
// reject key/value/startIndex fields meant for indexedPair.
func TestStorageMapping_Computed_WithKey(t *testing.T) {
	m := schema.StorageMapping{
		Type: schema.ComputedMapping,
		Read: "expr",
		Key:  &schema.MappingField{Target: schema.JSONDataTarget, Pattern: "x"},
	}
	assert.ErrorContains(t, m.Validate(), "computed mapping must not have key/value/startIndex")
}

// ============================================================
// StorageMapping.Validate — unknown type
// ============================================================

// TestStorageMapping_UnknownType ensures that an unrecognized
// mapping type string is rejected.
func TestStorageMapping_UnknownType(t *testing.T) {
	m := schema.StorageMapping{Type: "magic"}
	assert.ErrorContains(t, m.Validate(), "unknown mapping type: magic")
}

// ============================================================
// MappingField.Validate
// ============================================================

// TestMappingField_Valid confirms that a mapping field with a
// valid target and non-empty pattern passes.
func TestMappingField_Valid(t *testing.T) {
	m := schema.MappingField{Target: schema.JSONDataTarget, Pattern: "httpHeaderName{index}"}
	require.NoError(t, m.Validate())
}

// TestMappingField_InvalidTarget ensures that an unrecognized
// target on a mapping field is rejected.
func TestMappingField_InvalidTarget(t *testing.T) {
	m := schema.MappingField{Target: "bad", Pattern: "x{i}"}
	assert.ErrorContains(t, m.Validate(), "invalid target")
}

// TestMappingField_EmptyPattern ensures that a mapping field
// with an empty pattern is rejected.
func TestMappingField_EmptyPattern(t *testing.T) {
	m := schema.MappingField{Target: schema.JSONDataTarget, Pattern: ""}
	assert.ErrorContains(t, m.Validate(), "pattern is required")
}

// ============================================================
// Integration: full schema validation
// ============================================================

// TestFullSchemaValidation_Prometheus exercises a realistic
// Prometheus-like schema with multiple field types, groups,
// relationships, array items, storage mappings, and validation
// rules — validating end-to-end correctness.
func TestFullSchemaValidation_Prometheus(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1",
		PluginType:    "prometheus",
		PluginName:    "Prometheus",
		Fields: []schema.ConfigField{
			{
				ID: "url", Key: "url", ValueType: schema.StringType,
				Target: ptr(schema.RootTarget), Required: true,
				Validations: []schema.FieldValidationRule{
					{Type: schema.PatternValidation, Pattern: "^https?://", Message: "Must be HTTP(S) URL"},
				},
			},
			{
				ID: "auth.basicAuth", Key: "basicAuth",
				ValueType: schema.BooleanType, Target: ptr(schema.RootTarget),
			},
			{
				ID: "auth.basicAuthUser", Key: "basicAuthUser",
				ValueType: schema.StringType, Target: ptr(schema.RootTarget),
				RequiredWhen: "auth.basicAuth == true",
			},
			{
				ID: "auth.basicAuthPassword", Key: "basicAuthPassword",
				ValueType: schema.StringType, Target: ptr(schema.SecureJSONTarget),
				SemanticType: schema.PasswordType,
			},
			{
				ID: "jsonData.httpMethod", Key: "httpMethod",
				ValueType: schema.StringType, Target: ptr(schema.JSONDataTarget),
				Validations: []schema.FieldValidationRule{
					{Type: schema.AllowedValuesValidation, Values: []any{"GET", "POST"}},
				},
				UI: &schema.FieldUI{
					Component: schema.UISelect,
					Options: []schema.FieldOption{
						{Label: "GET", Value: "GET"},
						{Label: "POST", Value: "POST"},
					},
				},
			},
			{
				ID: "jsonData.timeout", Key: "timeout",
				ValueType: schema.NumberType, Target: ptr(schema.JSONDataTarget),
				Validations: []schema.FieldValidationRule{
					{Type: schema.RangeValidation, Min: ptr(1.0), Max: ptr(300.0)},
				},
			},
			{
				ID: "httpHeaders", Key: "httpHeaders",
				ValueType: schema.ArrayType, Target: ptr(schema.JSONDataTarget),
				Item: &schema.FieldItemSchema{
					ValueType: schema.ObjectType,
					Fields: []schema.ConfigField{
						{ID: "httpHeaders.item.key", Key: "key", ValueType: schema.StringType, IsItemField: ptr(true)},
						{ID: "httpHeaders.item.value", Key: "value", ValueType: schema.StringType, IsItemField: ptr(true)},
					},
				},
				Storage: &schema.StorageMapping{
					Type:  schema.IndexedPairMapping,
					Key:   &schema.MappingField{Target: schema.JSONDataTarget, Pattern: "httpHeaderName{index}"},
					Value: &schema.MappingField{Target: schema.SecureJSONTarget, Pattern: "httpHeaderValue{index}"},
				},
			},
			{
				ID: "derived.hasAuth", Key: "hasAuth",
				ValueType: schema.BooleanType, Kind: schema.VirtualField,
				DependsOn: "auth.basicAuth == true",
			},
		},
		Groups: []schema.ConfigGroup{
			{ID: "connection", Title: "Connection", FieldRefs: []string{"url", "jsonData.httpMethod", "jsonData.timeout"}},
			{ID: "auth", Title: "Authentication", FieldRefs: []string{"auth.basicAuth", "auth.basicAuthUser", "auth.basicAuthPassword"}},
		},
		Relationships: []schema.FieldRelationship{
			{Type: schema.PairRelationship, Fields: []string{"auth.basicAuthUser", "auth.basicAuthPassword"}, Description: "Basic auth credentials"},
		},
	}

	require.NoError(t, s.Validate())

	// Verify all 10 fields (8 top-level + 2 item fields) are collected
	ids, err := s.FieldIDs()
	require.NoError(t, err)
	assert.Len(t, ids, 10)
}

// ============================================================
// SemanticType.IsValid
// ============================================================

// TestSemanticType_Valid verifies all defined SemanticType constants
// are recognized as valid.
func TestSemanticType_Valid(t *testing.T) {
	for _, st := range []schema.SemanticType{
		schema.URLType, schema.PasswordType, schema.TokenType,
		schema.HostnameType, schema.DurationType,
	} {
		assert.True(t, st.IsValid(), "%s should be valid", st)
	}
}

// TestSemanticType_Invalid verifies that empty and unknown semantic
// types are rejected.
func TestSemanticType_Invalid(t *testing.T) {
	assert.False(t, schema.SemanticType("").IsValid())
	assert.False(t, schema.SemanticType("email").IsValid())
}

// TestFieldValidate_InvalidSemanticType ensures that a field with an
// unrecognized semanticType is rejected during validation.
func TestFieldValidate_InvalidSemanticType(t *testing.T) {
	f := validStorageField("x", "x")
	f.SemanticType = "email"
	assert.ErrorContains(t, f.Validate(), "invalid semanticType")
}

// TestFieldValidate_ValidSemanticType confirms that known semantic
// types pass validation.
func TestFieldValidate_ValidSemanticType(t *testing.T) {
	f := validStorageField("x", "x")
	f.SemanticType = schema.PasswordType
	require.NoError(t, f.Validate())
}

// ============================================================
// Lifecycle.IsValid
// ============================================================

// TestLifecycle_Valid verifies all defined Lifecycle constants
// are recognized as valid.
func TestLifecycle_Valid(t *testing.T) {
	for _, l := range []schema.Lifecycle{
		schema.StableLifecycle, schema.DeprecatedLifecycle, schema.ExperimentalLifecycle,
	} {
		assert.True(t, l.IsValid(), "%s should be valid", l)
	}
}

// TestLifecycle_Invalid verifies that empty and unknown lifecycle
// values are rejected.
func TestLifecycle_Invalid(t *testing.T) {
	assert.False(t, schema.Lifecycle("").IsValid())
	assert.False(t, schema.Lifecycle("beta").IsValid())
}

// TestFieldValidate_InvalidLifecycle ensures that a field with an
// unrecognized lifecycle is rejected during validation.
func TestFieldValidate_InvalidLifecycle(t *testing.T) {
	f := validStorageField("x", "x")
	f.Lifecycle = "beta"
	assert.ErrorContains(t, f.Validate(), "invalid lifecycle")
}

// TestFieldValidate_ValidLifecycle confirms that known lifecycle
// values pass validation.
func TestFieldValidate_ValidLifecycle(t *testing.T) {
	f := validStorageField("x", "x")
	f.Lifecycle = schema.DeprecatedLifecycle
	require.NoError(t, f.Validate())
}

// ============================================================
// UIComponent.IsValid
// ============================================================

// TestUIComponent_Valid verifies all defined UIComponent constants
// are recognized as valid.
func TestUIComponent_Valid(t *testing.T) {
	for _, c := range []schema.UIComponent{
		schema.UIInput, schema.UITextarea, schema.UISelect, schema.UIMultiselect,
		schema.UIRadio, schema.UICheckbox, schema.UISwitch, schema.UICode,
		schema.UIKeyValue, schema.UIList,
	} {
		assert.True(t, c.IsValid(), "%s should be valid", c)
	}
}

// TestUIComponent_Invalid verifies that empty and unknown component
// names are rejected.
func TestUIComponent_Invalid(t *testing.T) {
	assert.False(t, schema.UIComponent("").IsValid())
	assert.False(t, schema.UIComponent("datepicker").IsValid())
}

// TestFieldValidate_InvalidUIComponent ensures that a field with
// an unrecognized UI component is rejected during validation.
func TestFieldValidate_InvalidUIComponent(t *testing.T) {
	f := validStorageField("x", "x")
	f.UI = &schema.FieldUI{Component: "datepicker"}
	assert.ErrorContains(t, f.Validate(), "invalid ui component")
}

// TestFieldValidate_ValidUIComponent confirms that a field with
// a known UI component passes validation.
func TestFieldValidate_ValidUIComponent(t *testing.T) {
	f := validStorageField("x", "x")
	f.UI = &schema.FieldUI{Component: schema.UIInput}
	require.NoError(t, f.Validate())
}

// ============================================================
// UIWidth.IsValid
// ============================================================

// TestUIWidth_Valid verifies that full and half are accepted.
func TestUIWidth_Valid(t *testing.T) {
	assert.True(t, schema.FullWidth.IsValid())
	assert.True(t, schema.HalfWidth.IsValid())
}

// TestUIWidth_Invalid verifies that empty and unknown widths
// are rejected.
func TestUIWidth_Invalid(t *testing.T) {
	assert.False(t, schema.UIWidth("").IsValid())
	assert.False(t, schema.UIWidth("third").IsValid())
}

// TestFieldValidate_InvalidUIWidth ensures that a field with an
// unrecognized UI width is rejected during validation.
func TestFieldValidate_InvalidUIWidth(t *testing.T) {
	f := validStorageField("x", "x")
	f.UI = &schema.FieldUI{Component: schema.UIInput, Width: "third"}
	assert.ErrorContains(t, f.Validate(), "invalid ui width")
}

// TestFieldValidate_ValidUIWidth confirms that a known UI width
// passes validation.
func TestFieldValidate_ValidUIWidth(t *testing.T) {
	f := validStorageField("x", "x")
	f.UI = &schema.FieldUI{Component: schema.UIInput, Width: schema.HalfWidth}
	require.NoError(t, f.Validate())
}

// ============================================================
// RelationshipType.IsValid
// ============================================================

// TestRelationshipType_Valid verifies that pair and group are
// accepted as valid relationship types.
func TestRelationshipType_Valid(t *testing.T) {
	assert.True(t, schema.PairRelationship.IsValid())
	assert.True(t, schema.GroupRelationship.IsValid())
}

// TestRelationshipType_Invalid verifies that empty and unknown
// relationship types are rejected.
func TestRelationshipType_Invalid(t *testing.T) {
	assert.False(t, schema.RelationshipType("").IsValid())
	assert.False(t, schema.RelationshipType("dependency").IsValid())
}

// TestValidateRefs_InvalidRelationshipType ensures that a schema
// with an invalid relationship type is rejected.
func TestValidateRefs_InvalidRelationshipType(t *testing.T) {
	s := minimalSchema(validStorageField("a", "a"))
	s.Relationships = []schema.FieldRelationship{
		{Type: "dependency", Fields: []string{"a"}},
	}
	assert.ErrorContains(t, s.Validate(), "invalid type")
}

// ============================================================
// ValidateOptionValue — option type checking
// ============================================================

// TestValidateOptionValue_StringMatch confirms that string options
// are accepted for string fields.
func TestValidateOptionValue_StringMatch(t *testing.T) {
	assert.True(t, schema.ValidateOptionValue("hello", schema.StringType))
}

// TestValidateOptionValue_StringMismatch ensures that a numeric
// option is rejected for a string field.
func TestValidateOptionValue_StringMismatch(t *testing.T) {
	assert.False(t, schema.ValidateOptionValue(42, schema.StringType))
}

// TestValidateOptionValue_NumberInt confirms that int values are
// accepted for number fields.
func TestValidateOptionValue_NumberInt(t *testing.T) {
	assert.True(t, schema.ValidateOptionValue(42, schema.NumberType))
}

// TestValidateOptionValue_NumberFloat confirms that float64 values
// are accepted for number fields.
func TestValidateOptionValue_NumberFloat(t *testing.T) {
	assert.True(t, schema.ValidateOptionValue(3.14, schema.NumberType))
}

// TestValidateOptionValue_NumberMismatch ensures that a string
// value is rejected for a number field.
func TestValidateOptionValue_NumberMismatch(t *testing.T) {
	assert.False(t, schema.ValidateOptionValue("not-a-number", schema.NumberType))
}

// TestValidateOptionValue_BoolMatch confirms that bool values are
// accepted for boolean fields.
func TestValidateOptionValue_BoolMatch(t *testing.T) {
	assert.True(t, schema.ValidateOptionValue(true, schema.BooleanType))
}

// TestValidateOptionValue_BoolMismatch ensures that a string value
// is rejected for a boolean field.
func TestValidateOptionValue_BoolMismatch(t *testing.T) {
	assert.False(t, schema.ValidateOptionValue("true", schema.BooleanType))
}

// TestValidateOptionValue_NilRejected confirms that nil values are
// rejected for all field types, matching JSON Schema's "value is required".
func TestValidateOptionValue_NilRejected(t *testing.T) {
	assert.False(t, schema.ValidateOptionValue(nil, schema.StringType))
	assert.False(t, schema.ValidateOptionValue(nil, schema.NumberType))
	assert.False(t, schema.ValidateOptionValue(nil, schema.BooleanType))
	assert.False(t, schema.ValidateOptionValue(nil, schema.ArrayType))
	assert.False(t, schema.ValidateOptionValue(nil, schema.ObjectType))
}

// TestValidateOptionValue_ArrayObjectSkipped confirms that array
// and object fields skip type checking on option values.
func TestValidateOptionValue_ArrayObjectSkipped(t *testing.T) {
	assert.True(t, schema.ValidateOptionValue("anything", schema.ArrayType))
	assert.True(t, schema.ValidateOptionValue(42, schema.ObjectType))
}

// TestFieldValidate_OptionTypeMismatch ensures that a select field
// with an option value that doesn't match the field's valueType is
// rejected during validation.
func TestFieldValidate_OptionTypeMismatch(t *testing.T) {
	f := validStorageField("x", "x")
	f.ValueType = schema.StringType
	f.UI = &schema.FieldUI{
		Component: schema.UISelect,
		Options: []schema.FieldOption{
			{Label: "Good", Value: "good"},
			{Label: "Bad", Value: 42}, // mismatch: number in string field
		},
	}
	assert.ErrorContains(t, f.Validate(), "option[1] value type mismatch")
}

// TestFieldValidate_OptionTypeValid confirms that a select field
// with correctly-typed option values passes validation.
func TestFieldValidate_OptionTypeValid(t *testing.T) {
	f := validStorageField("x", "x")
	f.ValueType = schema.StringType
	f.UI = &schema.FieldUI{
		Component: schema.UISelect,
		Options: []schema.FieldOption{
			{Label: "GET", Value: "GET"},
			{Label: "POST", Value: "POST"},
		},
	}
	require.NoError(t, f.Validate())
}

// TestFieldValidate_NumberOptionTypeValid confirms that number
// field options with numeric values pass validation.
func TestFieldValidate_NumberOptionTypeValid(t *testing.T) {
	f := validStorageField("x", "x")
	f.ValueType = schema.NumberType
	f.UI = &schema.FieldUI{
		Component: schema.UISelect,
		Options: []schema.FieldOption{
			{Label: "Low", Value: 1},
			{Label: "High", Value: 100},
		},
	}
	require.NoError(t, f.Validate())
}

// ============================================================
// JSON round-trip compatibility
// ============================================================

// TestJSONRoundTrip_MinimalSchema verifies that a minimal schema
// survives JSON marshal/unmarshal and still validates.
func TestJSONRoundTrip_MinimalSchema(t *testing.T) {
	s := minimalSchema(validStorageField("url", "url"))
	require.NoError(t, s.Validate())

	data, err := json.Marshal(s)
	require.NoError(t, err)

	var decoded schema.DatasourceConfigSchema
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.NoError(t, decoded.Validate())

	assert.Equal(t, s.SchemaVersion, decoded.SchemaVersion)
	assert.Equal(t, s.PluginType, decoded.PluginType)
	assert.Len(t, decoded.Fields, 1)
	assert.Equal(t, "url", decoded.Fields[0].ID)
}

// TestJSONRoundTrip_FullSchema verifies that a complex schema with
// all feature areas (groups, relationships, validations, overrides,
// storage mappings, item fields) survives JSON round-trip.
func TestJSONRoundTrip_FullSchema(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1",
		PluginType:    "test",
		PluginName:    "Test Plugin",
		DocURL:        "https://example.com/docs",
		Fields: []schema.ConfigField{
			{
				ID: "url", Key: "url", ValueType: schema.StringType,
				Target: ptr(schema.RootTarget), Required: true,
				SemanticType: schema.URLType,
				Lifecycle:    schema.StableLifecycle,
				Validations: []schema.FieldValidationRule{
					{Type: schema.PatternValidation, Pattern: "^https?://", ID: "url-check", Message: "Must be URL"},
				},
				UI: &schema.FieldUI{Component: schema.UIInput, Width: schema.FullWidth, Placeholder: "https://..."},
			},
			{
				ID: "method", Key: "httpMethod", ValueType: schema.StringType,
				Target: ptr(schema.JSONDataTarget),
				Validations: []schema.FieldValidationRule{
					{Type: schema.AllowedValuesValidation, Values: []any{"GET", "POST"}},
				},
				UI: &schema.FieldUI{
					Component: schema.UISelect,
					Options: []schema.FieldOption{
						{Label: "GET", Value: "GET"},
						{Label: "POST", Value: "POST"},
					},
				},
				Overrides: []schema.FieldOverride{
					{When: "version == 'v2'", DefaultValue: "POST"},
				},
			},
			{
				ID: "headers", Key: "headers", ValueType: schema.ArrayType,
				Target: ptr(schema.JSONDataTarget),
				Item: &schema.FieldItemSchema{
					ValueType: schema.ObjectType,
					Fields: []schema.ConfigField{
						{ID: "headers.item.k", Key: "key", ValueType: schema.StringType, IsItemField: ptr(true)},
						{ID: "headers.item.v", Key: "value", ValueType: schema.StringType, IsItemField: ptr(true)},
					},
				},
				Storage: &schema.StorageMapping{
					Type:  schema.IndexedPairMapping,
					Key:   &schema.MappingField{Target: schema.JSONDataTarget, Pattern: "headerName{index}"},
					Value: &schema.MappingField{Target: schema.JSONDataTarget, Pattern: "headerValue{index}"},
				},
			},
		},
		Groups: []schema.ConfigGroup{
			{ID: "conn", Title: "Connection", FieldRefs: []string{"url", "method"}},
		},
		Relationships: []schema.FieldRelationship{
			{Type: schema.PairRelationship, Fields: []string{"headers.item.k", "headers.item.v"}},
		},
	}

	require.NoError(t, s.Validate())

	data, err := json.Marshal(s)
	require.NoError(t, err)

	var decoded schema.DatasourceConfigSchema
	require.NoError(t, json.Unmarshal(data, &decoded))
	require.NoError(t, decoded.Validate())

	assert.Equal(t, s.PluginType, decoded.PluginType)
	assert.Len(t, decoded.Fields, 3)
	assert.Len(t, decoded.Groups, 1)
	assert.Len(t, decoded.Relationships, 1)
	assert.Equal(t, schema.IndexedPairMapping, decoded.Fields[2].Storage.Type)
}

// TestJSONRoundTrip_ValidationRules verifies that all validation
// rule types survive JSON serialization with correct discriminators.
func TestJSONRoundTrip_ValidationRules(t *testing.T) {
	rules := []schema.FieldValidationRule{
		{Type: schema.PatternValidation, Pattern: "^[a-z]+$", Message: "lowercase only"},
		{Type: schema.RangeValidation, Min: ptr(0.0), Max: ptr(100.0)},
		{Type: schema.LengthValidation, Min: ptr(1.0)},
		{Type: schema.ItemCountValidation, Max: ptr(10.0)},
		{Type: schema.AllowedValuesValidation, Values: []any{"a", "b"}},
		{Type: schema.CustomValidation, Expression: "self.size() > 0"},
	}

	data, err := json.Marshal(rules)
	require.NoError(t, err)

	var decoded []schema.FieldValidationRule
	require.NoError(t, json.Unmarshal(data, &decoded))

	require.Len(t, decoded, 6)
	for i := range decoded {
		assert.Equal(t, rules[i].Type, decoded[i].Type)
		require.NoError(t, decoded[i].Validate())
	}
}

// TestJSONRoundTrip_StorageMappingTypes verifies that all three
// storage mapping types survive JSON serialization.
func TestJSONRoundTrip_StorageMappingTypes(t *testing.T) {
	mappings := []schema.StorageMapping{
		{Type: schema.DirectMapping},
		{
			Type:  schema.IndexedPairMapping,
			Key:   &schema.MappingField{Target: schema.JSONDataTarget, Pattern: "k{i}"},
			Value: &schema.MappingField{Target: schema.JSONDataTarget, Pattern: "v{i}"},
		},
		{Type: schema.ComputedMapping, Read: "expr"},
	}

	for _, m := range mappings {
		data, err := json.Marshal(m)
		require.NoError(t, err)

		var decoded schema.StorageMapping
		require.NoError(t, json.Unmarshal(data, &decoded))
		assert.Equal(t, m.Type, decoded.Type)
		require.NoError(t, decoded.Validate())
	}
}

// ============================================================
// Example schemas — Loki & Tempo
// ============================================================

// TestExampleSchema_Loki validates a Loki-like datasource schema
// with derived fields (array of objects), basic auth, and groups.
func TestExampleSchema_Loki(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1",
		PluginType:    "loki",
		PluginName:    "Loki",
		Fields: []schema.ConfigField{
			{
				ID: "url", Key: "url", ValueType: schema.StringType,
				Target: ptr(schema.RootTarget), Required: true,
				SemanticType: schema.URLType,
			},
			{
				ID: "jsonData.maxLines", Key: "maxLines", ValueType: schema.StringType,
				Target: ptr(schema.JSONDataTarget),
			},
			{
				ID: "jsonData.derivedFields", Key: "derivedFields",
				ValueType: schema.ArrayType, Target: ptr(schema.JSONDataTarget),
				Item: &schema.FieldItemSchema{
					ValueType: schema.ObjectType,
					Fields: []schema.ConfigField{
						{ID: "derivedFields.item.name", Key: "name", ValueType: schema.StringType, IsItemField: ptr(true)},
						{ID: "derivedFields.item.matcherRegex", Key: "matcherRegex", ValueType: schema.StringType, IsItemField: ptr(true)},
						{ID: "derivedFields.item.url", Key: "url", ValueType: schema.StringType, IsItemField: ptr(true),
							SemanticType: schema.URLType},
					},
				},
			},
			{
				ID: "jsonData.timeout", Key: "timeout", ValueType: schema.NumberType,
				Target: ptr(schema.JSONDataTarget),
				Validations: []schema.FieldValidationRule{
					{Type: schema.RangeValidation, Min: ptr(1.0), Max: ptr(600.0)},
				},
			},
		},
		Groups: []schema.ConfigGroup{
			{ID: "connection", Title: "Connection", FieldRefs: []string{"url", "jsonData.timeout"}},
			{ID: "derived", Title: "Derived Fields", FieldRefs: []string{"jsonData.derivedFields"}},
		},
	}
	require.NoError(t, s.Validate())

	ids, err := s.FieldIDs()
	require.NoError(t, err)
	assert.Len(t, ids, 7) // 4 top-level + 3 item fields
}

// TestExampleSchema_Tempo validates a Tempo-like datasource schema
// with nested config (service map), virtual fields, and custom
// validation rules.
func TestExampleSchema_Tempo(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1",
		PluginType:    "tempo",
		PluginName:    "Tempo",
		Fields: []schema.ConfigField{
			{
				ID: "url", Key: "url", ValueType: schema.StringType,
				Target: ptr(schema.RootTarget), Required: true,
				SemanticType: schema.URLType,
				Lifecycle:    schema.StableLifecycle,
			},
			{
				ID: "jsonData.serviceMap.datasourceUid", Key: "serviceMap.datasourceUid",
				ValueType: schema.StringType, Target: ptr(schema.JSONDataTarget),
			},
			{
				ID: "jsonData.search.hide", Key: "search.hide",
				ValueType: schema.BooleanType, Target: ptr(schema.JSONDataTarget),
			},
			{
				ID: "jsonData.nodeGraph.enabled", Key: "nodeGraph.enabled",
				ValueType: schema.BooleanType, Target: ptr(schema.JSONDataTarget),
			},
			{
				ID: "jsonData.streamingEnabled.search", Key: "streamingEnabled.search",
				ValueType: schema.BooleanType, Target: ptr(schema.JSONDataTarget),
			},
			{
				ID: "jsonData.streamingEnabled.metrics", Key: "streamingEnabled.metrics",
				ValueType: schema.BooleanType, Target: ptr(schema.JSONDataTarget),
			},
			{
				ID: "derived.hasServiceMap", Key: "hasServiceMap",
				ValueType: schema.BooleanType, Kind: schema.VirtualField,
				Lifecycle: schema.ExperimentalLifecycle,
				DependsOn: "jsonData.serviceMap.datasourceUid != ''",
			},
		},
		Groups: []schema.ConfigGroup{
			{ID: "connection", Title: "Connection", FieldRefs: []string{"url"}},
			{ID: "features", Title: "Features", FieldRefs: []string{
				"jsonData.nodeGraph.enabled",
				"jsonData.streamingEnabled.search",
				"jsonData.streamingEnabled.metrics",
			}},
		},
		Relationships: []schema.FieldRelationship{
			{
				Type:   schema.GroupRelationship,
				Fields: []string{"jsonData.streamingEnabled.search", "jsonData.streamingEnabled.metrics"},
			},
		},
	}
	require.NoError(t, s.Validate())
}

// TestExampleSchema_MySQL validates a MySQL-like datasource schema
// with secure fields, conditional requirements, and legacy patterns.
func TestExampleSchema_MySQL(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1",
		PluginType:    "mysql",
		PluginName:    "MySQL",
		Fields: []schema.ConfigField{
			{
				ID: "url", Key: "url", ValueType: schema.StringType,
				Target: ptr(schema.RootTarget), Required: true,
				SemanticType: schema.URLType,
				Validations: []schema.FieldValidationRule{
					{Type: schema.PatternValidation, Pattern: ".+:\\d+", Message: "Must include host:port"},
				},
			},
			{
				ID: "root.database", Key: "database", ValueType: schema.StringType,
				Target: ptr(schema.RootTarget),
			},
			{
				ID: "root.user", Key: "user", ValueType: schema.StringType,
				Target: ptr(schema.RootTarget),
			},
			{
				ID: "secureJsonData.password", Key: "password", ValueType: schema.StringType,
				Target: ptr(schema.SecureJSONTarget), SemanticType: schema.PasswordType,
				RequiredWhen: "root.user != ''",
			},
			{
				ID: "jsonData.maxOpenConns", Key: "maxOpenConns", ValueType: schema.NumberType,
				Target: ptr(schema.JSONDataTarget),
				Validations: []schema.FieldValidationRule{
					{Type: schema.RangeValidation, Min: ptr(0.0), Max: ptr(100.0)},
				},
			},
			{
				ID: "jsonData.connMaxLifetime", Key: "connMaxLifetime", ValueType: schema.NumberType,
				Target:       ptr(schema.JSONDataTarget),
				SemanticType: schema.DurationType,
			},
			{
				ID: "jsonData.tlsAuth", Key: "tlsAuth", ValueType: schema.BooleanType,
				Target: ptr(schema.JSONDataTarget),
			},
			{
				ID: "secureJsonData.tlsCACert", Key: "tlsCACert", ValueType: schema.StringType,
				Target:    ptr(schema.SecureJSONTarget),
				DependsOn: "jsonData.tlsAuth == true",
				UI:        &schema.FieldUI{Component: schema.UITextarea, Rows: 5},
			},
		},
		Groups: []schema.ConfigGroup{
			{ID: "connection", Title: "Connection", FieldRefs: []string{"url", "root.database"}},
			{ID: "auth", Title: "Authentication", FieldRefs: []string{"root.user", "secureJsonData.password"}},
			{ID: "tls", Title: "TLS / SSL", FieldRefs: []string{"jsonData.tlsAuth", "secureJsonData.tlsCACert"}},
		},
		Relationships: []schema.FieldRelationship{
			{Type: schema.PairRelationship, Fields: []string{"root.user", "secureJsonData.password"}},
		},
	}
	require.NoError(t, s.Validate())

	ids, err := s.FieldIDs()
	require.NoError(t, err)
	assert.Len(t, ids, 8)
}
// ============================================================
// FieldEffect.Validate
// ============================================================

// TestFieldEffect_Valid confirms that a well-formed effect passes.
func TestFieldEffect_Valid(t *testing.T) {
	e := schema.FieldEffect{When: "value == 'basic-auth'", Set: map[string]any{"auth.basicAuth": true}}
	require.NoError(t, e.Validate())
}

// TestFieldEffect_EmptyWhen ensures an effect without a when is rejected.
func TestFieldEffect_EmptyWhen(t *testing.T) {
	e := schema.FieldEffect{Set: map[string]any{"a": true}}
	assert.ErrorContains(t, e.Validate(), "effect when is required")
}

// TestFieldEffect_EmptySet ensures an effect with no set entries is rejected.
func TestFieldEffect_EmptySet(t *testing.T) {
	e := schema.FieldEffect{When: "value == 'x'", Set: map[string]any{}}
	assert.ErrorContains(t, e.Validate(), "effect set must not be empty")
}

// TestFieldEffect_NilSet ensures an effect with nil set is rejected.
func TestFieldEffect_NilSet(t *testing.T) {
	e := schema.FieldEffect{When: "value == 'x'"}
	assert.ErrorContains(t, e.Validate(), "effect set must not be empty")
}

// TestFieldValidate_PropagatesEffectError ensures that invalid effects
// on a field bubble up through field validation.
func TestFieldValidate_PropagatesEffectError(t *testing.T) {
	f := schema.ConfigField{
		ID: "x", Key: "x", ValueType: schema.StringType, Kind: schema.VirtualField,
		Effects: []schema.FieldEffect{{When: "", Set: map[string]any{"a": true}}},
	}
	assert.ErrorContains(t, f.Validate(), "invalid effect[0]")
}

// TestFieldValidate_ValidEffects confirms a field with well-formed effects passes.
func TestFieldValidate_ValidEffects(t *testing.T) {
	f := schema.ConfigField{
		ID: "x", Key: "x", ValueType: schema.StringType, Kind: schema.VirtualField,
		Effects: []schema.FieldEffect{
			{When: "value == 'a'", Set: map[string]any{"y": true}},
			{When: "value == 'b'", Set: map[string]any{"y": false}},
		},
	}
	require.NoError(t, f.Validate())
}

// TestValidateRefs_EffectSetRefsValid ensures effect set keys that
// reference existing field IDs pass validation.
func TestValidateRefs_EffectSetRefsValid(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "selector", Key: "selector", ValueType: schema.StringType,
				Kind: schema.VirtualField,
				Effects: []schema.FieldEffect{
					{When: "value == 'on'", Set: map[string]any{"target": true}},
				},
			},
			{ID: "target", Key: "target", ValueType: schema.BooleanType, Target: ptr(schema.JSONDataTarget)},
		},
	}
	require.NoError(t, s.Validate())
}

// TestValidateRefs_EffectSetRefsUnknown ensures effect set keys that
// reference non-existent field IDs are rejected.
func TestValidateRefs_EffectSetRefsUnknown(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test", PluginName: "Test",
		Fields: []schema.ConfigField{
			{
				ID: "selector", Key: "selector", ValueType: schema.StringType,
				Kind: schema.VirtualField,
				Effects: []schema.FieldEffect{
					{When: "value == 'on'", Set: map[string]any{"ghost": true}},
				},
			},
		},
	}
	assert.ErrorContains(t, s.Validate(), "effect[0].set references unknown field id: ghost")
}

// TestExampleSchema_AuthSelector validates the full auth-selector
// pattern: virtual dropdown + effects + dependent storage fields.
func TestExampleSchema_AuthSelector(t *testing.T) {
	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1", PluginType: "test-auth", PluginName: "Auth Test",
		Fields: []schema.ConfigField{
			{
				ID: "url", Key: "url", ValueType: schema.StringType,
				Target: ptr(schema.RootTarget), Required: true,
			},
			{
				ID: "auth.method", Key: "authMethod", Label: "Authentication method",
				ValueType: schema.StringType, Kind: schema.VirtualField,
				DefaultValue: "no-auth",
				Validations: []schema.FieldValidationRule{
					{Type: schema.AllowedValuesValidation, Values: []any{"no-auth", "basic-auth", "forward-oauth"}},
				},
				UI: &schema.FieldUI{
					Component: schema.UISelect,
					Options: []schema.FieldOption{
						{Label: "No Authentication", Value: "no-auth"},
						{Label: "Basic authentication", Value: "basic-auth"},
						{Label: "Forward OAuth Identity", Value: "forward-oauth"},
					},
				},
				Storage: &schema.StorageMapping{
					Type: schema.ComputedMapping,
					Read: "root.basicAuth == true ? 'basic-auth' : (jsonData.oauthPassThru == true ? 'forward-oauth' : 'no-auth')",
				},
				Effects: []schema.FieldEffect{
					{When: "value == 'no-auth'", Set: map[string]any{"auth.basicAuth": false, "auth.oauthPassThru": false}},
					{When: "value == 'basic-auth'", Set: map[string]any{"auth.basicAuth": true, "auth.oauthPassThru": false}},
					{When: "value == 'forward-oauth'", Set: map[string]any{"auth.basicAuth": false, "auth.oauthPassThru": true}},
				},
			},
			{
				ID: "auth.basicAuth", Key: "basicAuth", ValueType: schema.BooleanType,
				Target: ptr(schema.RootTarget), DefaultValue: false,
			},
			{
				ID: "auth.oauthPassThru", Key: "oauthPassThru", ValueType: schema.BooleanType,
				Target: ptr(schema.JSONDataTarget), DefaultValue: false,
			},
			{
				ID: "auth.basicAuthUser", Key: "basicAuthUser", ValueType: schema.StringType,
				Target: ptr(schema.RootTarget),
				DependsOn: "auth.method == 'basic-auth'", RequiredWhen: "auth.method == 'basic-auth'",
			},
			{
				ID: "auth.basicAuthPassword", Key: "basicAuthPassword", ValueType: schema.StringType,
				Target: ptr(schema.SecureJSONTarget), SemanticType: schema.PasswordType,
				DependsOn: "auth.method == 'basic-auth'",
			},
		},
		Groups: []schema.ConfigGroup{
			{ID: "connection", Title: "Connection", FieldRefs: []string{"url"}},
			{ID: "auth", Title: "Authentication", FieldRefs: []string{"auth.method", "auth.basicAuthUser", "auth.basicAuthPassword"}},
		},
		Relationships: []schema.FieldRelationship{
			{Type: schema.PairRelationship, Fields: []string{"auth.basicAuthUser", "auth.basicAuthPassword"}},
		},
	}
	require.NoError(t, s.Validate())

	ids, err := s.FieldIDs()
	require.NoError(t, err)
	assert.Len(t, ids, 6)
}