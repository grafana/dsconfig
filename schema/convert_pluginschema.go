package schema

import (
	"fmt"
	"strings"

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
	Key         string `json:"key"`
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`
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

	fieldByID := make(map[string]*ConfigField)
	for i := range s.Fields {
		fieldByID[s.Fields[i].ID] = &s.Fields[i]
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
			secureValues = append(secureValues, SecureValueInfo{
				Key:         f.Key,
				Description: f.Description,
				Required:    f.Required,
			})
			continue
		}

		if f.Target != nil && *f.Target == JSONDataTarget {
			jsonDataProps[f.Key] = fieldToSpecSchema(f)
			if f.Required {
				jsonDataRequired = append(jsonDataRequired, f.Key)
			}
		} else {
			rootProps[f.Key] = fieldToSpecSchema(f)
			if f.Required {
				rootRequired = append(rootRequired, f.Key)
			}
		}
	}

	// Build conditional constraints using anyOf/allOf.
	jsonDataConstraints := buildAnyOfConstraints(s.Fields, fieldByID, JSONDataTarget)
	rootConstraints := buildAnyOfConstraints(s.Fields, fieldByID, RootTarget)

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
		applyAnyOfConstraints(&jd, jsonDataConstraints)
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
	applyAnyOfConstraints(specSchema, rootConstraints)

	return &PluginSettings{
		Spec:         specSchema,
		SecureValues: secureValues,
	}, nil
}

// applyAnyOfConstraints adds conditional constraints to a schema.
// For a single discriminator group, anyOf is applied directly.
// For multiple groups, each is wrapped in allOf.
func applyAnyOfConstraints(s *spec.Schema, constraints []spec.Schema) {
	switch len(constraints) {
	case 0:
		return
	case 1:
		s.AnyOf = constraints[0].AnyOf
	default:
		s.AllOf = constraints
	}
}

// ============================================================
// Conditional constraint builder
// ============================================================

// conditionExpr represents a parsed "fieldID == value" condition.
type conditionExpr struct {
	fieldID string
	value   string
}

// parseEqualityCondition parses "fieldID == 'value'" or "fieldID == value".
func parseEqualityCondition(expr string) (conditionExpr, bool) {
	idx := strings.Index(expr, " == ")
	if idx < 0 {
		return conditionExpr{}, false
	}
	fieldID := strings.TrimSpace(expr[:idx])
	valueStr := strings.TrimSpace(expr[idx+4:])

	if len(valueStr) >= 2 {
		if (valueStr[0] == '\'' && valueStr[len(valueStr)-1] == '\'') ||
			(valueStr[0] == '"' && valueStr[len(valueStr)-1] == '"') {
			valueStr = valueStr[1 : len(valueStr)-1]
		}
	}

	if fieldID == "" || valueStr == "" {
		return conditionExpr{}, false
	}
	return conditionExpr{fieldID: fieldID, value: valueStr}, true
}

// discriminatorGroup tracks conditional requirements for one discriminator.
type discriminatorGroup struct {
	key        string
	valueType  ValueType
	enumValues []any
	// condValue (string repr) → required field keys
	requiredByValue map[string][]string
}

// buildAnyOfConstraints analyses requiredWhen conditions on fields in the
// given target and builds anyOf constraint schemas grouped by discriminator.
//
// Each discriminator produces an anyOf with:
//  1. A "not present" branch (discriminator field absent).
//  2. One branch per known enum value with const + required fields.
func buildAnyOfConstraints(
	fields []ConfigField,
	fieldByID map[string]*ConfigField,
	target TargetLocation,
) []spec.Schema {
	var order []string
	groups := make(map[string]*discriminatorGroup)

	for i := range fields {
		f := &fields[i]
		if f.Kind == VirtualField || f.Target == nil || *f.Target != target {
			continue
		}
		if f.RequiredWhen == "" {
			continue
		}

		cond, ok := parseEqualityCondition(f.RequiredWhen)
		if !ok {
			continue
		}

		discField, exists := fieldByID[cond.fieldID]
		if !exists || discField.Target == nil || *discField.Target != target {
			continue // discriminator not found or in a different target
		}

		g, exists := groups[cond.fieldID]
		if !exists {
			enumVals := collectDiscriminatorValues(discField)
			if len(enumVals) == 0 {
				continue
			}
			g = &discriminatorGroup{
				key:             discField.Key,
				valueType:       discField.ValueType,
				enumValues:      enumVals,
				requiredByValue: make(map[string][]string),
			}
			groups[cond.fieldID] = g
			order = append(order, cond.fieldID)
		}

		g.requiredByValue[cond.value] = append(g.requiredByValue[cond.value], f.Key)
	}

	var result []spec.Schema
	for _, id := range order {
		g := groups[id]
		result = append(result, spec.Schema{
			SchemaProps: spec.SchemaProps{
				AnyOf: buildAnyOfBranches(g),
			},
		})
	}
	return result
}

// buildAnyOfBranches generates anyOf branches for a single discriminator.
func buildAnyOfBranches(g *discriminatorGroup) []spec.Schema {
	branches := make([]spec.Schema, 0, 1+len(g.enumValues))

	// Branch: discriminator not present (field omitted entirely).
	branches = append(branches, spec.Schema{
		SchemaProps: spec.SchemaProps{
			Not: &spec.Schema{
				SchemaProps: spec.SchemaProps{
					Required: []string{g.key},
				},
			},
		},
	})

	// One branch per known enum value.
	for _, ev := range g.enumValues {
		evStr := fmt.Sprintf("%v", ev)
		required := []string{g.key}
		if extra, ok := g.requiredByValue[evStr]; ok {
			required = append(required, extra...)
		}
		branches = append(branches, spec.Schema{
			SchemaProps: spec.SchemaProps{
				Properties: map[string]spec.Schema{
					g.key: {
						SchemaProps: spec.SchemaProps{
							Enum: []any{ev},
						},
					},
				},
				Required: required,
			},
		})
	}

	return branches
}

// collectDiscriminatorValues returns the known values for a discriminator field.
func collectDiscriminatorValues(f *ConfigField) []any {
	for _, v := range f.Validations {
		if v.Type == AllowedValuesValidation {
			return v.Values
		}
	}
	if f.UI != nil {
		switch f.UI.Component {
		case UISelect, UIRadio:
			var vals []any
			for _, opt := range f.UI.Options {
				vals = append(vals, opt.Value)
			}
			if len(vals) > 0 {
				return vals
			}
		}
	}
	if f.ValueType == BooleanType {
		return []any{true, false}
	}
	return nil
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

	if f.ValueType == ArrayType && f.Item != nil {
		itemSchema := itemSchemaToSpec(*f.Item)
		s.Items = &spec.SchemaOrArray{Schema: &itemSchema}
	}

	return s
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
