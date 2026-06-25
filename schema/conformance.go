package schema

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/grafana/dsconfig/dsconfig"
	sdkSchema "github.com/grafana/grafana-plugin-sdk-go/experimental/pluginschema"
	"github.com/stretchr/testify/require"
)

// jsonMarshalerType is used to detect fields that customize their JSON encoding
// via a MarshalJSON method, so the kind-based type check can skip them.
var jsonMarshalerType = reflect.TypeOf((*json.Marshaler)(nil)).Elem()

// Params describes the plugin-specific inputs the conformance tests need. Every
// field is required unless documented otherwise.
type Params struct {
	// PluginID is the expected datasource plugin type (for example
	// "yesoreyeram-infinity-datasource").
	PluginID string

	// DSConfigSchema is the parsed and resolved dsconfig single source of truth.
	// If the schema uses baseFields, call ResolveBaseFields() or
	// ParseAndResolveSchemaJSON() before passing it here.
	// Passing an unresolved schema will fail the BaseFieldsResolved check.
	DSConfigSchema *dsconfig.Schema

	// PluginSchema is the full SDK PluginSchema assembled from ConfigSchema.
	PluginSchema *sdkSchema.PluginSchema

	// SettingsJSONModel is a zero value of the Go struct that backs jsonData
	// (for example models.InfinitySettingsJson{}). Its json tags are compared
	// against the schema's jsonData fields.
	SettingsJSONModel any

	// SecureKeys are the secureJsonData keys the plugin actually reads when
	// loading settings. They are compared against the schema's secure values.
	SecureKeys []string
}

// RunConformanceTests runs the full suite of plugin-agnostic schema guard rails
// as subtests. Call it from a single Test function in the plugin's package.
func RunConformanceTests(t *testing.T, p Params) {
	t.Helper()

	t.Run("BaseFieldsResolved", func(t *testing.T) { BaseFieldsResolved(t, p) })
	t.Run("SchemaRoundTrip", func(t *testing.T) { SchemaRoundTrip(t, p) })
	t.Run("SchemaArtifactInSync", func(t *testing.T) { SchemaArtifactInSync(t, p) })
	t.Run("SchemaSpecHasNoSecureJSON", func(t *testing.T) { SchemaSpecHasNoSecureJSON(t, p) })
	t.Run("ConfigSchemaValid", func(t *testing.T) { ConfigSchemaValid(t, p) })
	t.Run("JSONDataMatchesStruct", func(t *testing.T) { JSONDataMatchesStruct(t, p) })
	t.Run("JSONDataTypesMatchStruct", func(t *testing.T) { JSONDataTypesMatchStruct(t, p) })
	t.Run("SecureValuesMatchLoadSettings", func(t *testing.T) { SecureValuesMatchLoadSettings(t, p) })
}

// BaseFieldsResolved guards that DSConfigSchema was resolved before being
// passed to the conformance suite. An unresolved schema (BaseFields non-empty)
// means the plugin's production code path is also likely broken — it would
// fail at Validate() with an unhelpful error.
func BaseFieldsResolved(t *testing.T, p Params) {
	t.Helper()
	require.Empty(t, p.DSConfigSchema.BaseFields,
		"DSConfigSchema has unresolved baseFields; call ResolveBaseFields() or "+
			"ParseAndResolveSchemaJSON() before passing to RunConformanceTests")
}

// SchemaRoundTrip loads the committed artifact through the production provider
// that Grafana uses (NewSchemaProvider reads {apiVersion}.json), staging it
// under that name in a temp dir to load it exactly as Grafana would. The API
// version is taken from the PluginSchema itself.
func SchemaRoundTrip(t *testing.T, p Params) {
	t.Helper()

	apiVersion := p.PluginSchema.TargetAPIVersion
	data, err := os.ReadFile(ArtifactPathSchema) // #nosec G304 -- package-controlled path
	require.NoError(t, err, "schema artifact missing; run `%s`", "go generate ./...")

	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, apiVersion+".json"), data, 0o600))

	provider := sdkSchema.NewSchemaProvider(os.DirFS(dir))

	loaded, err := provider.Get(apiVersion)
	require.NoError(t, err)
	require.False(t, loaded.IsZero(), "loaded schema should not be empty")
	require.Equal(t, apiVersion, loaded.TargetAPIVersion)
	require.NotNil(t, loaded.SettingsSchema)
	require.NotEmpty(t, loaded.SettingsSchema.SecureValues)
}

// SchemaArtifactInSync fails if any committed JSON artifact has drifted from the
// in-memory PluginSchema. It checks the full schema, the settings schema, and the
// settings examples against their respective artifact files. Regenerate with the
// plugin's generate command.
func SchemaArtifactInSync(t *testing.T, p Params) {
	t.Helper()

	fullSchema, err := marshal(p.PluginSchema)
	require.NoError(t, err)
	assertArtifactInSync(t, p, ArtifactPathSchema, fullSchema)

	settings, err := marshal(p.PluginSchema.SettingsSchema)
	require.NoError(t, err)
	assertArtifactInSync(t, p, ArtifactPathSettings, settings)

	examples, err := marshal(p.PluginSchema.SettingsExamples)
	require.NoError(t, err)
	assertArtifactInSync(t, p, ArtifactPathSettingsExamples, examples)
}

// assertArtifactInSync compares the canonical encoding of an in-memory schema
// object against its committed artifact file.
func assertArtifactInSync(t *testing.T, p Params, path string, want []byte) {
	t.Helper()

	got, err := os.ReadFile(path) // #nosec G304 -- package-controlled path
	require.NoError(t, err, "schema artifact %s missing; run `%s`", path, "go generate ./...")
	require.JSONEq(t, string(want), string(got),
		"schema artifact %s is out of date; run `%s`", path, "go generate ./...")
}

// SchemaSpecHasNoSecureJSON guards the invariant the SDK enforces: secure values
// must be declared via SecureValues, never inside the settings spec.
func SchemaSpecHasNoSecureJSON(t *testing.T, p Params) {
	t.Helper()

	require.NotNil(t, p.PluginSchema.SettingsSchema)
	require.NotNil(t, p.PluginSchema.SettingsSchema.Spec)
	_, hasSecure := p.PluginSchema.SettingsSchema.Spec.Properties["secureJsonData"]
	require.False(t, hasSecure, "secureJsonData must not be defined on the spec; use SecureValues")
}

// ConfigSchemaValid validates the dsconfig single source of truth and that its
// plugin type matches the expected PluginID.
func ConfigSchemaValid(t *testing.T, p Params) {
	t.Helper()

	require.NoError(t, p.DSConfigSchema.Validate())
	require.Equal(t, p.PluginID, p.DSConfigSchema.PluginType)
}

// JSONDataMatchesStruct is the single-source-of-truth guard rail: the jsonData
// field keys declared in the dsconfig schema must exactly match the json tags on
// the settings model. Add/remove/rename a struct field without updating the
// schema (or vice versa) and this fails in both directions.
func JSONDataMatchesStruct(t *testing.T, p Params) {
	t.Helper()

	schemaKeys := []string{}
	for _, f := range p.DSConfigSchema.Fields {
		if f.Target != nil && *f.Target == dsconfig.JSONDataTarget {
			schemaKeys = append(schemaKeys, f.Key)
		}
	}
	structKeys := jsonTagKeys(reflect.TypeOf(p.SettingsJSONModel))
	sort.Strings(schemaKeys)
	sort.Strings(structKeys)
	require.ElementsMatch(t, structKeys, schemaKeys,
		"jsonData fields in the schema are out of sync with the settings model json tags")
}

// JSONDataTypesMatchStruct closes the type-drift gap left by
// JSONDataMatchesStruct: the declared JSON type of each jsonData field must also
// agree with the Go kind of the corresponding settings model field.
func JSONDataTypesMatchStruct(t *testing.T, p Params) {
	t.Helper()

	schemaTypes := map[string]dsconfig.ValueType{}
	for _, f := range p.DSConfigSchema.Fields {
		if f.Target != nil && *f.Target == dsconfig.JSONDataTarget {
			schemaTypes[f.Key] = f.ValueType
		}
	}

	structFields := jsonTagFields(reflect.TypeOf(p.SettingsJSONModel))

	for key, vt := range schemaTypes {
		info, ok := structFields[key]
		if !ok {
			continue // key-set drift is reported by JSONDataMatchesStruct
		}
		if info.customJSON {
			// The field defines its own MarshalJSON, so its Go kind does not
			// determine the JSON type (for example an int enum that marshals
			// to a string). The kind-based check cannot reason about it.
			continue
		}
		want := valueTypesForKind(info.kind)
		require.Contains(t, want, vt,
			"jsonData field %q is declared as %q in the schema but the struct field has Go kind %q",
			key, vt, info.kind)
	}
}

// SecureValuesMatchLoadSettings guards that the secureJsonData keys declared in
// the schema match the secret keys the plugin actually reads.
func SecureValuesMatchLoadSettings(t *testing.T, p Params) {
	t.Helper()

	schemaKeys := []string{}
	for _, f := range p.DSConfigSchema.Fields {
		if f.Target != nil && *f.Target == dsconfig.SecureJSONTarget {
			schemaKeys = append(schemaKeys, f.Key)
		}
	}
	secureKeys := append([]string(nil), p.SecureKeys...)
	sort.Strings(schemaKeys)
	sort.Strings(secureKeys)
	require.ElementsMatch(t, secureKeys, schemaKeys,
		"secureJsonData fields in the schema are out of sync with the secrets the plugin reads")
}

// fieldInfo captures the JSON-relevant facts about a settings struct field: the
// Go kind of the field and whether its type customizes JSON encoding via a
// MarshalJSON method (in which case the Go kind does not determine the JSON
// type).
type fieldInfo struct {
	kind       reflect.Kind
	customJSON bool
}

// jsonTagKeys returns the JSON field names produced by encoding/json for a
// struct, skipping fields without a json tag or tagged "-". Fields promoted from
// anonymous embedded structs (for example awsds.AWSDatasourceSettings) are
// included, mirroring how encoding/json marshals embedded fields.
func jsonTagKeys(t reflect.Type) []string {
	fields := jsonTagFields(t)
	keys := make([]string, 0, len(fields))
	for name := range fields {
		keys = append(keys, name)
	}
	return keys
}

// jsonTagFields maps each JSON field name to information about its struct field,
// skipping fields without a json tag or tagged "-". Fields promoted from
// anonymous embedded structs are flattened in, mirroring encoding/json. Fields
// declared on the outer struct take precedence over promoted ones of the same
// name, matching encoding/json's shallowest-wins rule.
func jsonTagFields(t reflect.Type) map[string]fieldInfo {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	fields := make(map[string]fieldInfo, t.NumField())
	promoted := make(map[string]fieldInfo)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		tag := field.Tag.Get("json")

		// An anonymous embedded struct without a json tag has its fields
		// promoted by encoding/json, so recurse into it. A type with a custom
		// MarshalJSON is treated as a single value, not flattened.
		if field.Anonymous && tag == "" && !implementsJSONMarshaler(field.Type) {
			ft := field.Type
			for ft.Kind() == reflect.Ptr {
				ft = ft.Elem()
			}
			if ft.Kind() == reflect.Struct {
				for name, info := range jsonTagFields(ft) {
					promoted[name] = info
				}
				continue
			}
		}

		if tag == "" || tag == "-" {
			continue
		}
		name := strings.Split(tag, ",")[0]
		if name == "" {
			continue
		}
		fields[name] = fieldInfo{
			kind:       field.Type.Kind(),
			customJSON: implementsJSONMarshaler(field.Type),
		}
	}

	// Outer fields win over promoted ones of the same name.
	for name, info := range promoted {
		if _, ok := fields[name]; !ok {
			fields[name] = info
		}
	}
	return fields
}

// implementsJSONMarshaler reports whether t (or a pointer to t) implements
// json.Marshaler, i.e. the field controls its own JSON representation.
func implementsJSONMarshaler(t reflect.Type) bool {
	if t.Implements(jsonMarshalerType) {
		return true
	}
	if t.Kind() != reflect.Ptr {
		return reflect.PtrTo(t).Implements(jsonMarshalerType)
	}
	return false
}

// valueTypesForKind returns the dsconfig ValueTypes compatible with a given Go
// reflect.Kind. A struct field may legitimately be declared as more than one
// JSON type, so the guard checks for membership rather than strict equality.
func valueTypesForKind(kind reflect.Kind) []dsconfig.ValueType {
	switch kind {
	case reflect.String:
		return []dsconfig.ValueType{dsconfig.StringType}
	case reflect.Bool:
		return []dsconfig.ValueType{dsconfig.BooleanType}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return []dsconfig.ValueType{dsconfig.NumberType}
	case reflect.Slice, reflect.Array:
		return []dsconfig.ValueType{dsconfig.ArrayType}
	case reflect.Struct, reflect.Map:
		return []dsconfig.ValueType{dsconfig.ObjectType, dsconfig.MapType}
	default:
		return nil
	}
}

// MustNewSDKSchema assembles this plugin's SDK PluginSchema from the embedded
// dsconfig.json, failing the test if construction fails. Test-only helper that
// wraps dsconfig.NewSDKSchema.
func MustNewSDKSchema(t *testing.T, data []byte, examples *sdkSchema.SettingsExamples) *sdkSchema.PluginSchema {
	t.Helper()
	s, err := dsconfig.NewSDKSchema(data, examples)
	require.NoError(t, err)
	return s
}
