package schema

import (
	"fmt"

	"k8s.io/kube-openapi/pkg/validation/spec"
)

// PluginSettings mirrors pluginschema.Settings from
// grafana-plugin-sdk-go/experimental/pluginschema.
// Wire-compatible so values can be marshalled and consumed by the SDK.
type PluginSettings struct {
	Spec         *spec.Schema      `json:"spec"`
	SecureValues []SecureValueInfo `json:"secureValues,omitempty"`
}

// SecureValueInfo describes a secret required by the datasource.
type SecureValueInfo struct {
	Key          string `json:"key"`
	Description  string `json:"description,omitempty"`
	Required     bool   `json:"required,omitempty"`
	DependsOn    string `json:"x-depends-on,omitempty"`
	RequiredWhen string `json:"x-required-when,omitempty"`
}

// ToPluginSettings converts the schema to a PluginSettings object
// compatible with grafana-plugin-sdk-go's pluginschema.Settings.
//
// The Spec schema includes both root-level fields (url, basicAuth, etc.)
// and jsonData fields as properties. Fields targeting secureJsonData
// become SecureValues entries. Virtual fields are skipped.
func (s *DatasourceConfigSchema) ToPluginSettings() (*PluginSettings, error) {
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("invalid schema: %w", err)
	}

	rootProps := make(map[string]spec.Schema)
	var rootRequired []string

	jsonDataProps := make(map[string]spec.Schema)
	var jsonDataRequired []string

	var secureValues []SecureValueInfo

	for _, f := range s.Fields {
		if f.Kind == VirtualField {
			continue
		}

		if f.Target != nil && *f.Target == SecureJSONTarget {
			sv := SecureValueInfo{
				Key:         f.Key,
				Description: f.Description,
				Required:    f.Required,
			}
			if f.DependsOn != "" {
				sv.DependsOn = f.DependsOn
			}
			if f.RequiredWhen != "" {
				sv.RequiredWhen = f.RequiredWhen
			}
			secureValues = append(secureValues, sv)
			continue
		}

		if f.Target != nil && *f.Target == JSONDataTarget {
			placeInSection(jsonDataProps, f)
			if f.Required && f.Section == "" {
				jsonDataRequired = append(jsonDataRequired, f.Key)
			}
		} else {
			rootProps[f.Key] = fieldToSpecSchema(f)
			if f.Required {
				rootRequired = append(rootRequired, f.Key)
			}
		}
	}

	if len(jsonDataProps) > 0 {
		jd := spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:       spec.StringOrArray{"object"},
				Properties: jsonDataProps,
			},
		}
		if len(jsonDataRequired) > 0 {
			jd.Required = jsonDataRequired
		}
		rootProps["jsonData"] = jd
	}

	specSchema := &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:       spec.StringOrArray{"object"},
			Properties: rootProps,
		},
	}
	if len(rootRequired) > 0 {
		specSchema.Required = rootRequired
	}

	return &PluginSettings{
		Spec:         specSchema,
		SecureValues: secureValues,
	}, nil
}

// placeInSection places a field into the correct section sub-object within props.
// If the field has no Section, it is placed directly. If it has a Section,
// the field is nested under an object property with that section name.
func placeInSection(props map[string]spec.Schema, f ConfigField) {
	if f.Section == "" {
		props[f.Key] = fieldToSpecSchema(f)
		return
	}
	section, exists := props[f.Section]
	if !exists {
		section = spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:       spec.StringOrArray{"object"},
				Properties: make(map[string]spec.Schema),
			},
		}
	}
	if section.Properties == nil {
		section.Properties = make(map[string]spec.Schema)
	}
	section.Properties[f.Key] = fieldToSpecSchema(f)
	if f.Required {
		section.Required = append(section.Required, f.Key)
	}
	props[f.Section] = section
}

// fieldToSpecSchema converts a ConfigField to an OpenAPI spec.Schema.
func fieldToSpecSchema(f ConfigField) spec.Schema {
	s := spec.Schema{
		SchemaProps: spec.SchemaProps{
			Description: f.Description,
			Type:        spec.StringOrArray{valueTypeToJSONType(f.ValueType)},
		},
	}

	if f.DefaultValue != nil {
		s.Default = f.DefaultValue
	}

	if f.SemanticType != "" {
		if fmt := semanticTypeToFormat(f.SemanticType); fmt != "" {
			s.Format = fmt
		}
	}

	applyValidations(&s, f)
	applyUIEnum(&s, f)
	applyConditions(&s, f)

	if f.ValueType == ArrayType && f.Item != nil {
		itemSchema := itemSchemaToSpec(*f.Item)
		s.Items = &spec.SchemaOrArray{Schema: &itemSchema}
	}

	return s
}

// applyConditions maps conditional behavior (CEL expressions) to
// vendor extensions on the spec.Schema.
func applyConditions(s *spec.Schema, f ConfigField) {
	if f.DependsOn == "" && f.RequiredWhen == "" && f.DisabledWhen == "" {
		return
	}
	s.Extensions = make(spec.Extensions)
	if f.DependsOn != "" {
		s.Extensions["x-depends-on"] = f.DependsOn
	}
	if f.RequiredWhen != "" {
		s.Extensions["x-required-when"] = f.RequiredWhen
	}
	if f.DisabledWhen != "" {
		s.Extensions["x-disabled-when"] = f.DisabledWhen
	}
}

// applyValidations maps dsconfig validation rules to JSON Schema keywords.
func applyValidations(s *spec.Schema, f ConfigField) {
	for _, v := range f.Validations {
		switch v.Type {
		case PatternValidation:
			s.Pattern = v.Pattern
		case RangeValidation:
			s.Minimum = v.Min
			s.Maximum = v.Max
		case LengthValidation:
			if v.Min != nil {
				n := int64(*v.Min)
				s.MinLength = &n
			}
			if v.Max != nil {
				n := int64(*v.Max)
				s.MaxLength = &n
			}
		case ItemCountValidation:
			if v.Min != nil {
				n := int64(*v.Min)
				s.MinItems = &n
			}
			if v.Max != nil {
				n := int64(*v.Max)
				s.MaxItems = &n
			}
		case AllowedValuesValidation:
			s.Enum = make([]any, len(v.Values))
			copy(s.Enum, v.Values)
		}
	}
}

// applyUIEnum sets enum values from select/radio UI options when no
// explicit allowedValues validation is present.
func applyUIEnum(s *spec.Schema, f ConfigField) {
	if f.UI == nil || len(f.UI.Options) == 0 {
		return
	}
	switch f.UI.Component {
	case UISelect, UIRadio, UIMultiselect:
		if len(s.Enum) > 0 {
			return // explicit allowedValues takes precedence
		}
		for _, opt := range f.UI.Options {
			s.Enum = append(s.Enum, opt.Value)
		}
	}
}

func itemSchemaToSpec(item FieldItemSchema) spec.Schema {
	s := spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type: spec.StringOrArray{valueTypeToJSONType(item.ValueType)},
		},
	}

	if item.ValueType == ObjectType && len(item.Fields) > 0 {
		props := make(map[string]spec.Schema)
		var required []string
		for _, f := range item.Fields {
			props[f.Key] = fieldToSpecSchema(f)
			if f.Required {
				required = append(required, f.Key)
			}
		}
		s.Properties = props
		if len(required) > 0 {
			s.Required = required
		}
	}

	return s
}

func valueTypeToJSONType(vt ValueType) string {
	switch vt {
	case StringType:
		return "string"
	case NumberType:
		return "number"
	case BooleanType:
		return "boolean"
	case ArrayType:
		return "array"
	case ObjectType:
		return "object"
	default:
		return "string"
	}
}

func semanticTypeToFormat(st SemanticType) string {
	switch st {
	case URLType:
		return "uri"
	case PasswordType:
		return "password"
	case HostnameType:
		return "hostname"
	case DurationType:
		return "duration"
	default:
		return ""
	}
}
