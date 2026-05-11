package schema_test

import (
	"github.com/grafana/dsconfig/schema"
)

// ptr returns a pointer to the given value. Useful for constructing
// test fixtures with optional pointer fields inline.
func ptr[T any](v T) *T { return &v }

// validStorageField returns a minimal valid storage field with the given
// id and key, targeting jsonData. Use as a baseline for tests that need
// a valid ConfigField without boilerplate.
func validStorageField(id, key string) schema.ConfigField {
	return schema.ConfigField{
		ID:        id,
		Key:       key,
		ValueType: schema.StringType,
		Target:    ptr(schema.JSONDataTarget),
	}
}

// minimalSchema returns a DatasourceConfigSchema with sensible defaults
// and the provided fields. Useful for testing schema-level validation
// without repeating root-level boilerplate.
func minimalSchema(fields ...schema.ConfigField) *schema.DatasourceConfigSchema {
	if len(fields) == 0 {
		fields = append(fields, schema.ConfigField{
			ID:        "url",
			Key:       "url",
			ValueType: schema.StringType,
			Target:    ptr(schema.RootTarget),
		})
	}
	return &schema.DatasourceConfigSchema{
		SchemaVersion: "v1",
		PluginType:    "test",
		PluginName:    "Test",
		Fields:        fields,
	}
}
