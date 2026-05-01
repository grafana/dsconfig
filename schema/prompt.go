package schema

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

// ============================================================
// Prompt Schema Types
// ============================================================

// PromptSchema is a compact, LLM-friendly projection of a
// DatasourceConfigSchema. It strips UI hints, groups, storage
// mappings, and other rendering/internal concerns.
type PromptSchema struct {
	PluginType string        `json:"pluginType"`
	PluginName string        `json:"pluginName"`
	Fields     []PromptField `json:"fields"`
}

// PromptField is a single field projected for LLM consumption.
type PromptField struct {
	ID            string         `json:"id"`
	Path          string         `json:"path"`
	Type          ValueType      `json:"type"`
	Label         string         `json:"label,omitempty"`
	SemanticType  SemanticType   `json:"semanticType,omitempty"`
	Description   string         `json:"description,omitempty"`
	Required      bool           `json:"required,omitempty"`
	RequiredWhen  string         `json:"requiredWhen,omitempty"`
	DependsOn     string         `json:"dependsOn,omitempty"`
	DefaultValue  any            `json:"defaultValue,omitempty"`
	AllowedValues []any          `json:"allowedValues,omitempty"`
	Pattern       string         `json:"pattern,omitempty"`
	Range         *PromptRange   `json:"range,omitempty"`
	Items         []PromptField  `json:"items,omitempty"`
	Options       []PromptOption `json:"options,omitempty"`
}

// PromptRange is a numeric bound constraint.
type PromptRange struct {
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`
}

// PromptOption is a virtual selector option with side-effects.
type PromptOption struct {
	Value any            `json:"value"`
	Label string         `json:"label"`
	Sets  map[string]any `json:"sets"`
}

// ============================================================
// Projection
// ============================================================

// ToPromptSchema projects a full DatasourceConfigSchema into a
// compact, LLM-friendly PromptSchema.
func ToPromptSchema(s *DatasourceConfigSchema) PromptSchema {
	var fields []PromptField
	for i := range s.Fields {
		if isManagedField(&s.Fields[i]) {
			continue
		}
		fields = append(fields, projectField(&s.Fields[i]))
	}
	return PromptSchema{
		PluginType: s.PluginType,
		PluginName: s.PluginName,
		Fields:     fields,
	}
}

// ToPromptString serializes a DatasourceConfigSchema into a compact
// JSON string suitable for embedding in an LLM prompt (tool/function calling).
func ToPromptString(s *DatasourceConfigSchema) (string, error) {
	ps := ToPromptSchema(s)
	data, err := json.MarshalIndent(ps, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToPromptText renders a DatasourceConfigSchema into human-readable
// text suitable for embedding in an LLM system/user prompt.
// More token-efficient and easier for LLMs to reason about than JSON.
func ToPromptText(s *DatasourceConfigSchema) string {
	ps := ToPromptSchema(s)
	var b strings.Builder

	b.WriteString(ps.PluginName)
	b.WriteString(" (pluginType: ")
	b.WriteString(ps.PluginType)
	b.WriteString(")\n\nFields:\n")

	for i := range ps.Fields {
		renderField(&b, &ps.Fields[i], "")
	}

	return b.String()
}

func renderField(b *strings.Builder, f *PromptField, indent string) {
	// Main line
	name := f.Label
	if name == "" {
		name = f.ID
	}

	b.WriteString(indent)
	b.WriteString("- ")
	b.WriteString(name)

	if strings.Contains(f.Path, ".") {
		b.WriteString(" (")
		b.WriteString(f.Path)
		b.WriteString(")")
	}

	b.WriteString(" [")
	b.WriteString(string(f.Type))
	if f.SemanticType != "" {
		b.WriteString(", ")
		b.WriteString(string(f.SemanticType))
	}
	b.WriteString("]")

	if f.Required {
		b.WriteString(" REQUIRED")
	}
	if f.DefaultValue != nil {
		b.WriteString(" default: ")
		b.WriteString(formatValue(f.DefaultValue))
	}
	if f.Description != "" {
		b.WriteString(" — ")
		b.WriteString(f.Description)
	}
	b.WriteString("\n")

	sub := indent + "  "

	if f.RequiredWhen != "" {
		b.WriteString(sub)
		b.WriteString("Required when: ")
		b.WriteString(f.RequiredWhen)
		b.WriteString("\n")
	}
	if f.DependsOn != "" {
		b.WriteString(sub)
		b.WriteString("Visible when: ")
		b.WriteString(f.DependsOn)
		b.WriteString("\n")
	}
	if len(f.AllowedValues) > 0 && len(f.Options) == 0 {
		b.WriteString(sub)
		b.WriteString("Allowed: ")
		for i, v := range f.AllowedValues {
			if i > 0 {
				b.WriteString(", ")
			}
			b.WriteString(formatValue(v))
		}
		b.WriteString("\n")
	}
	if f.Pattern != "" {
		b.WriteString(sub)
		b.WriteString("Pattern: ")
		b.WriteString(f.Pattern)
		b.WriteString("\n")
	}
	if f.Range != nil {
		b.WriteString(sub)
		b.WriteString("Range: ")
		parts := []string{}
		if f.Range.Min != nil {
			parts = append(parts, "min: "+strconv.FormatFloat(*f.Range.Min, 'f', -1, 64))
		}
		if f.Range.Max != nil {
			parts = append(parts, "max: "+strconv.FormatFloat(*f.Range.Max, 'f', -1, 64))
		}
		b.WriteString(strings.Join(parts, ", "))
		b.WriteString("\n")
	}

	if len(f.Options) > 0 {
		b.WriteString(sub)
		b.WriteString("Options:\n")
		for _, opt := range f.Options {
			b.WriteString(sub)
			b.WriteString("  ")
			b.WriteString(formatValue(opt.Value))
			b.WriteString(" (")
			b.WriteString(opt.Label)
			b.WriteString(") → sets ")
			first := true
			for k, v := range opt.Sets {
				if !first {
					b.WriteString(", ")
				}
				b.WriteString(k)
				b.WriteString("=")
				b.WriteString(formatValue(v))
				first = false
			}
			b.WriteString("\n")
		}
	}

	if len(f.Items) > 0 {
		b.WriteString(sub)
		b.WriteString("Item fields:\n")
		for i := range f.Items {
			renderField(b, &f.Items[i], sub+"  ")
		}
	}
}

func formatValue(v any) string {
	switch val := v.(type) {
	case string:
		return `"` + val + `"`
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	case nil:
		return "null"
	default:
		data, _ := json.Marshal(val)
		return string(data)
	}
}

// ============================================================
// Internals
// ============================================================

// isManagedField returns true if the field is tagged "managed-by:*",
// meaning it is driven by effects and should be hidden from the LLM.
func isManagedField(f *ConfigField) bool {
	for _, tag := range f.Tags {
		if strings.HasPrefix(tag, "managed-by:") {
			return true
		}
	}
	return false
}

func projectField(f *ConfigField) PromptField {
	pf := PromptField{
		ID:   f.ID,
		Path: f.Path(),
		Type: f.ValueType,
	}

	if f.Label != "" {
		pf.Label = f.Label
	}
	if f.SemanticType != "" {
		pf.SemanticType = f.SemanticType
	}
	if f.Description != "" {
		pf.Description = f.Description
	}
	if f.Required {
		pf.Required = true
	}
	if f.RequiredWhen != "" {
		pf.RequiredWhen = f.RequiredWhen
	}
	if f.DependsOn != "" {
		pf.DependsOn = f.DependsOn
	}
	if f.DefaultValue != nil {
		pf.DefaultValue = f.DefaultValue
	}

	// Flatten validations
	for _, v := range f.Validations {
		switch v.Type {
		case AllowedValuesValidation:
			pf.AllowedValues = v.Values
		case PatternValidation:
			pf.Pattern = v.Pattern
		case RangeValidation:
			pf.Range = &PromptRange{Min: v.Min, Max: v.Max}
		}
	}

	// Flatten effects into options
	if len(f.Effects) > 0 {
		labelMap := buildLabelMap(f)
		for _, eff := range f.Effects {
			val := ExtractLiteralFromWhen(eff.When)
			label := ""
			if val != nil {
				if l, ok := labelMap[toString(val)]; ok {
					label = l
				} else {
					label = toString(val)
				}
			} else {
				label = eff.When
			}
			optVal := val
			if optVal == nil {
				optVal = eff.When
			}
			pf.Options = append(pf.Options, PromptOption{
				Value: optVal,
				Label: label,
				Sets:  eff.Set,
			})
		}
	}

	// Recurse into array/map item fields
	if f.Item != nil && len(f.Item.Fields) > 0 {
		for i := range f.Item.Fields {
			pf.Items = append(pf.Items, projectField(&f.Item.Fields[i]))
		}
	}

	return pf
}

func buildLabelMap(f *ConfigField) map[string]string {
	m := make(map[string]string)
	if f.UI != nil {
		for _, opt := range f.UI.Options {
			m[toString(opt.Value)] = opt.Label
		}
	}
	return m
}

func toString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	case bool:
		if val {
			return "true"
		}
		return "false"
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case int:
		return strconv.Itoa(val)
	default:
		return ""
	}
}

// ExtractLiteralFromWhen extracts a literal value from a simple
// "value == '...'" expression. Returns nil for complex expressions.
var (
	reSingleQuote = regexp.MustCompile(`^value\s*==\s*'([^']*)'$`)
	reDoubleQuote = regexp.MustCompile(`^value\s*==\s*"([^"]*)"$`)
	reBool        = regexp.MustCompile(`^value\s*==\s*(true|false)$`)
	reNumber      = regexp.MustCompile(`^value\s*==\s*(-?\d+(?:\.\d+)?)$`)
)

func ExtractLiteralFromWhen(when string) any {
	if m := reSingleQuote.FindStringSubmatch(when); m != nil {
		return m[1]
	}
	if m := reDoubleQuote.FindStringSubmatch(when); m != nil {
		return m[1]
	}
	if m := reBool.FindStringSubmatch(when); m != nil {
		return m[1] == "true"
	}
	if m := reNumber.FindStringSubmatch(when); m != nil {
		n, err := strconv.ParseFloat(m[1], 64)
		if err == nil {
			return n
		}
	}
	return nil
}
