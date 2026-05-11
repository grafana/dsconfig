package schema_test

import (
	"testing"

	"github.com/grafana/dsconfig/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestBasicAuthFields verifies that BasicAuthFields returns valid
// fields and each field passes individual validation.
func TestBasicAuthFields(t *testing.T) {
	fields := schema.BasicAuthFields()
	require.Len(t, fields, 3)
	for _, f := range fields {
		assert.NoError(t, f.Validate(), "field %s should be valid", f.ID)
	}
}

// TestTLSFields verifies that TLSFields returns valid fields.
func TestTLSFields(t *testing.T) {
	fields := schema.TLSFields()
	require.Len(t, fields, 7)
	for _, f := range fields {
		assert.NoError(t, f.Validate(), "field %s should be valid", f.ID)
	}
}

// TestCommonNetworkFields verifies that CommonNetworkFields returns
// valid fields.
func TestCommonNetworkFields(t *testing.T) {
	fields := schema.CommonNetworkFields()
	require.Len(t, fields, 4)
	for _, f := range fields {
		assert.NoError(t, f.Validate(), "field %s should be valid", f.ID)
	}
}

// TestHTTPHeaderFields verifies that HTTPHeaderFields returns valid
// fields with correct storage mapping.
func TestHTTPHeaderFields(t *testing.T) {
	fields := schema.HTTPHeaderFields()
	require.Len(t, fields, 1)
	h := fields[0]
	require.NoError(t, h.Validate())
	assert.Equal(t, schema.ArrayType, h.ValueType)
	require.NotNil(t, h.Storage)
	assert.Equal(t, schema.IndexedPairMapping, h.Storage.Type)
}

// TestCommonFieldsInSchema verifies that a schema using common field
// helpers validates end-to-end with no duplicate IDs.
func TestCommonFieldsInSchema(t *testing.T) {
	fields := []schema.ConfigField{
		{
			ID: "url", Key: "url", ValueType: schema.StringType,
			Target: ptr(schema.RootTarget), Required: true,
		},
	}
	fields = append(fields, schema.BasicAuthFields()...)
	fields = append(fields, schema.TLSFields()...)
	fields = append(fields, schema.CommonNetworkFields()...)
	fields = append(fields, schema.HTTPHeaderFields()...)

	s := &schema.DatasourceConfigSchema{
		SchemaVersion: "v1",
		PluginType:    "test-common",
		PluginName:    "Test Common",
		Fields:        fields,
	}
	require.NoError(t, s.Validate())

	ids, err := s.FieldIDs()
	require.NoError(t, err)
	// 1 url + 3 basic auth + 7 TLS + 4 network + 1 headers + 2 header items = 18
	assert.Len(t, ids, 18)
}
