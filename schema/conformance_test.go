package schema

import (
	"reflect"
	"testing"

	"github.com/grafana/dsconfig/dsconfig"
	sdkSchema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
	"github.com/stretchr/testify/require"
	"k8s.io/kube-openapi/pkg/validation/spec"
)

// ----------------------------------------------------------------------------
// Test helpers
// ----------------------------------------------------------------------------

// runCapture runs fn with a throwaway *testing.T and reports whether fn caused a
// test failure. It lets us exercise the failure branches of functions that take
// a concrete *testing.T (require.* calls FailNow -> runtime.Goexit, which still
// runs the deferred close) without failing the parent test. The throwaway T has
// no parent, so its failure does not propagate.
func runCapture(fn func(t *testing.T)) bool {
	tt := &testing.T{}
	done := make(chan struct{})
	go func() {
		defer close(done)
		fn(tt)
	}()
	<-done
	return tt.Failed()
}

func targetPtr(t dsconfig.TargetLocation) *dsconfig.TargetLocation { return &t }

// validJSONModel backs jsonData in the happy-path params.
type validJSONModel struct {
	Path string `json:"path"`
}

const (
	testAPIVersion = "v0alpha1"
	testPluginID   = "test-datasource"
)

// validPluginSchema returns a PluginSchema with a jsonData property, no
// secureJsonData in the spec, and a non-empty SecureValues list.
func validPluginSchema() *sdkSchema.PluginSchema {
	return &sdkSchema.PluginSchema{
		TargetAPIVersion: testAPIVersion,
		SettingsSchema: &sdkSchema.Settings{
			Spec: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Type: spec.StringOrArray{"object"},
					Properties: map[string]spec.Schema{
						"jsonData": {
							SchemaProps: spec.SchemaProps{
								Type: spec.StringOrArray{"object"},
							},
						},
					},
				},
			},
			SecureValues: []sdkSchema.SecureValueInfo{
				{Key: "apiKey", Description: "API key", Required: true},
			},
		},
	}
}

// validDSConfigSchema returns a valid dsconfig schema with one jsonData field
// (path) and one secureJsonData field (apiKey).
func validDSConfigSchema() *dsconfig.Schema {
	return &dsconfig.Schema{
		SchemaVersion: "1.0.0",
		PluginType:    testPluginID,
		PluginName:    "Test Datasource",
		Fields: []dsconfig.ConfigField{
			{
				ID:        "path",
				Key:       "path",
				ValueType: dsconfig.StringType,
				Target:    targetPtr(dsconfig.JSONDataTarget),
			},
			{
				ID:        "apiKey",
				Key:       "apiKey",
				ValueType: dsconfig.StringType,
				Target:    targetPtr(dsconfig.SecureJSONTarget),
			},
		},
	}
}

func validParams() Params {
	return Params{
		PluginID:          testPluginID,
		DSConfigSchema:    validDSConfigSchema(),
		PluginSchema:      validPluginSchema(),
		SettingsJSONModel: validJSONModel{},
		SecureKeys:        []string{"apiKey"},
	}
}

// writeArtifactsInTempCWD changes the working directory to a temp dir and writes
// the canonical artifacts for the supplied schema there, so the file-based
// conformance checks read the freshly generated files.
func writeArtifactsInTempCWD(t *testing.T, schema *sdkSchema.PluginSchema) {
	t.Helper()
	t.Chdir(t.TempDir())
	require.NoError(t, WriteArtifacts(schema))
}

// ----------------------------------------------------------------------------
// Full suite + happy path
// ----------------------------------------------------------------------------

func TestRunConformanceTests(t *testing.T) {
	p := validParams()
	writeArtifactsInTempCWD(t, p.PluginSchema)
	RunConformanceTests(t, p)
}

// ----------------------------------------------------------------------------
// SchemaRoundTrip
// ----------------------------------------------------------------------------

func TestSchemaRoundTrip(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := validParams()
		writeArtifactsInTempCWD(t, p.PluginSchema)
		require.False(t, runCapture(func(tt *testing.T) { SchemaRoundTrip(tt, p) }))
	})

	t.Run("missing artifact fails", func(t *testing.T) {
		p := validParams()
		// Empty CWD with no artifacts -> ReadFile error.
		t.Chdir(t.TempDir())
		require.True(t, runCapture(func(tt *testing.T) { SchemaRoundTrip(tt, p) }))
	})
}

// ----------------------------------------------------------------------------
// SchemaArtifactInSync / assertArtifactInSync
// ----------------------------------------------------------------------------

func TestSchemaArtifactInSync(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := validParams()
		writeArtifactsInTempCWD(t, p.PluginSchema)
		require.False(t, runCapture(func(tt *testing.T) { SchemaArtifactInSync(tt, p) }))
	})

	t.Run("drift fails", func(t *testing.T) {
		p := validParams()
		writeArtifactsInTempCWD(t, p.PluginSchema)
		// Mutate the in-memory schema so it no longer matches the artifacts.
		p.PluginSchema.SettingsSchema.SecureValues[0].Key = "different"
		require.True(t, runCapture(func(tt *testing.T) { SchemaArtifactInSync(tt, p) }))
	})

	t.Run("missing artifact fails", func(t *testing.T) {
		p := validParams()
		t.Chdir(t.TempDir())
		require.True(t, runCapture(func(tt *testing.T) { SchemaArtifactInSync(tt, p) }))
	})
}

// ----------------------------------------------------------------------------
// SchemaSpecHasNoSecureJSON
// ----------------------------------------------------------------------------

func TestSchemaSpecHasNoSecureJSON(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := validParams()
		require.False(t, runCapture(func(tt *testing.T) { SchemaSpecHasNoSecureJSON(tt, p) }))
	})

	t.Run("secureJsonData in spec fails", func(t *testing.T) {
		p := validParams()
		p.PluginSchema.SettingsSchema.Spec.Properties["secureJsonData"] = spec.Schema{}
		require.True(t, runCapture(func(tt *testing.T) { SchemaSpecHasNoSecureJSON(tt, p) }))
	})
}

// ----------------------------------------------------------------------------
// ConfigSchemaValid
// ----------------------------------------------------------------------------

func TestConfigSchemaValid(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := validParams()
		require.False(t, runCapture(func(tt *testing.T) { ConfigSchemaValid(tt, p) }))
	})

	t.Run("invalid schema fails", func(t *testing.T) {
		p := validParams()
		p.DSConfigSchema.PluginName = "" // breaks Validate()
		require.True(t, runCapture(func(tt *testing.T) { ConfigSchemaValid(tt, p) }))
	})

	t.Run("plugin id mismatch fails", func(t *testing.T) {
		p := validParams()
		p.PluginID = "some-other-datasource"
		require.True(t, runCapture(func(tt *testing.T) { ConfigSchemaValid(tt, p) }))
	})
}

// ----------------------------------------------------------------------------
// JSONDataMatchesStruct
// ----------------------------------------------------------------------------

func TestJSONDataMatchesStruct(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := validParams()
		require.False(t, runCapture(func(tt *testing.T) { JSONDataMatchesStruct(tt, p) }))
	})

	t.Run("extra schema key fails", func(t *testing.T) {
		p := validParams()
		p.DSConfigSchema.Fields = append(p.DSConfigSchema.Fields, dsconfig.ConfigField{
			ID:        "extra",
			Key:       "extra",
			ValueType: dsconfig.StringType,
			Target:    targetPtr(dsconfig.JSONDataTarget),
		})
		require.True(t, runCapture(func(tt *testing.T) { JSONDataMatchesStruct(tt, p) }))
	})

	t.Run("extra struct field fails", func(t *testing.T) {
		p := validParams()
		p.SettingsJSONModel = struct {
			Path  string `json:"path"`
			Extra string `json:"extra"`
		}{}
		require.True(t, runCapture(func(tt *testing.T) { JSONDataMatchesStruct(tt, p) }))
	})
}

// ----------------------------------------------------------------------------
// JSONDataTypesMatchStruct
// ----------------------------------------------------------------------------

type customJSONField int

func (customJSONField) MarshalJSON() ([]byte, error) { return []byte(`"custom"`), nil }

func TestJSONDataTypesMatchStruct(t *testing.T) {
	t.Run("matching types succeed", func(t *testing.T) {
		p := validParams()
		require.False(t, runCapture(func(tt *testing.T) { JSONDataTypesMatchStruct(tt, p) }))
	})

	t.Run("key missing from struct is skipped", func(t *testing.T) {
		p := validParams()
		// Schema declares a jsonData key with no matching struct field. The
		// type check skips it (key-set drift is reported elsewhere).
		p.DSConfigSchema.Fields = append(p.DSConfigSchema.Fields, dsconfig.ConfigField{
			ID:        "missing",
			Key:       "missing",
			ValueType: dsconfig.NumberType,
			Target:    targetPtr(dsconfig.JSONDataTarget),
		})
		require.False(t, runCapture(func(tt *testing.T) { JSONDataTypesMatchStruct(tt, p) }))
	})

	t.Run("custom json marshaler is skipped", func(t *testing.T) {
		p := validParams()
		// The struct field controls its own JSON encoding, so its Go kind
		// (int) does not have to agree with the declared string type.
		p.SettingsJSONModel = struct {
			Code customJSONField `json:"code"`
		}{}
		p.DSConfigSchema.Fields = []dsconfig.ConfigField{
			{
				ID:        "code",
				Key:       "code",
				ValueType: dsconfig.StringType,
				Target:    targetPtr(dsconfig.JSONDataTarget),
			},
		}
		require.False(t, runCapture(func(tt *testing.T) { JSONDataTypesMatchStruct(tt, p) }))
	})

	t.Run("type mismatch fails", func(t *testing.T) {
		p := validParams()
		// Declare path as number while the struct field is a string.
		p.DSConfigSchema.Fields[0].ValueType = dsconfig.NumberType
		require.True(t, runCapture(func(tt *testing.T) { JSONDataTypesMatchStruct(tt, p) }))
	})
}

// ----------------------------------------------------------------------------
// SecureValuesMatchLoadSettings
// ----------------------------------------------------------------------------

func TestSecureValuesMatchLoadSettings(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		p := validParams()
		require.False(t, runCapture(func(tt *testing.T) { SecureValuesMatchLoadSettings(tt, p) }))
	})

	t.Run("mismatch fails", func(t *testing.T) {
		p := validParams()
		p.SecureKeys = []string{"apiKey", "extraSecret"}
		require.True(t, runCapture(func(tt *testing.T) { SecureValuesMatchLoadSettings(tt, p) }))
	})
}

// ----------------------------------------------------------------------------
// jsonTagKeys / jsonTagFields
// ----------------------------------------------------------------------------

type embedded struct {
	Promoted string `json:"promoted"`
	Path     string `json:"path"` // shadowed by outer Path
}

type embeddedPtr struct {
	FromPtr string `json:"fromPtr"`
}

type ptrMarshaler struct{}

func (*ptrMarshaler) MarshalJSON() ([]byte, error) { return []byte(`"p"`), nil }

type taggedStruct struct {
	embedded                     // anonymous, no tag -> promoted
	*embeddedPtr                 // anonymous pointer, no tag -> promoted
	Path         string          `json:"path"`              // outer wins over promoted
	Renamed      string          `json:"renamed,omitempty"` // name from tag prefix
	Skipped      string          `json:"-"`                 // explicitly skipped
	NoTag        string          // no json tag -> skipped
	EmptyName    string          `json:",omitempty"` // empty name -> skipped
	Custom       customJSONField `json:"custom"`     // custom marshaler value type
}

func TestJSONTagFields(t *testing.T) {
	fields := jsonTagFields(reflect.TypeOf(&taggedStruct{})) // pointer deref path

	require.Contains(t, fields, "promoted")
	require.Contains(t, fields, "fromPtr")
	require.Contains(t, fields, "path")
	require.Contains(t, fields, "renamed")
	require.Contains(t, fields, "custom")

	require.NotContains(t, fields, "")        // empty name skipped
	require.NotContains(t, fields, "-")       // skipped tag never added
	require.NotContains(t, fields, "NoTag")   // untagged field skipped
	require.NotContains(t, fields, "Skipped") // raw field name never present

	// Outer Path (string) wins over the promoted embedded.Path.
	require.Equal(t, reflect.String, fields["path"].kind)
	require.False(t, fields["path"].customJSON)

	// Custom marshaler is flagged.
	require.True(t, fields["custom"].customJSON)
}

func TestJSONTagKeys(t *testing.T) {
	keys := jsonTagKeys(reflect.TypeOf(taggedStruct{}))
	require.ElementsMatch(t,
		[]string{"promoted", "fromPtr", "path", "renamed", "custom"},
		keys,
	)
}

// ----------------------------------------------------------------------------
// implementsJSONMarshaler
// ----------------------------------------------------------------------------

type valueMarshaler struct{}

func (valueMarshaler) MarshalJSON() ([]byte, error) { return []byte(`"v"`), nil }

type plainStruct struct{}

func TestImplementsJSONMarshaler(t *testing.T) {
	// Value receiver: both value and pointer types implement it.
	require.True(t, implementsJSONMarshaler(reflect.TypeOf(valueMarshaler{})))
	require.True(t, implementsJSONMarshaler(reflect.TypeOf(&valueMarshaler{})))

	// Pointer receiver: value type does not implement directly, but PtrTo does.
	require.True(t, implementsJSONMarshaler(reflect.TypeOf(ptrMarshaler{})))
	require.True(t, implementsJSONMarshaler(reflect.TypeOf(&ptrMarshaler{})))

	// Plain type: neither value nor pointer implements it.
	require.False(t, implementsJSONMarshaler(reflect.TypeOf(plainStruct{})))
	require.False(t, implementsJSONMarshaler(reflect.TypeOf(&plainStruct{})))
}

// ----------------------------------------------------------------------------
// valueTypesForKind
// ----------------------------------------------------------------------------

func TestValueTypesForKind(t *testing.T) {
	cases := []struct {
		kind reflect.Kind
		want []dsconfig.ValueType
	}{
		{reflect.String, []dsconfig.ValueType{dsconfig.StringType}},
		{reflect.Bool, []dsconfig.ValueType{dsconfig.BooleanType}},
		{reflect.Int, []dsconfig.ValueType{dsconfig.NumberType}},
		{reflect.Int64, []dsconfig.ValueType{dsconfig.NumberType}},
		{reflect.Uint, []dsconfig.ValueType{dsconfig.NumberType}},
		{reflect.Uint64, []dsconfig.ValueType{dsconfig.NumberType}},
		{reflect.Float32, []dsconfig.ValueType{dsconfig.NumberType}},
		{reflect.Float64, []dsconfig.ValueType{dsconfig.NumberType}},
		{reflect.Slice, []dsconfig.ValueType{dsconfig.ArrayType}},
		{reflect.Array, []dsconfig.ValueType{dsconfig.ArrayType}},
		{reflect.Struct, []dsconfig.ValueType{dsconfig.ObjectType, dsconfig.MapType}},
		{reflect.Map, []dsconfig.ValueType{dsconfig.ObjectType, dsconfig.MapType}},
		{reflect.Chan, nil},
		{reflect.Func, nil},
		{reflect.Interface, nil},
	}

	for _, c := range cases {
		require.Equal(t, c.want, valueTypesForKind(c.kind), "kind %s", c.kind)
	}
}
