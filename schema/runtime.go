package schema

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ============================================================
// Load Mode
// ============================================================

// LoadMode controls how secure fields are handled.
type LoadMode string

const (
	// ReadMode uses secureJsonFields (boolean map) to check secret presence.
	ReadMode LoadMode = "read"

	// WriteMode uses secureJsonData (actual values) for full validation.
	WriteMode LoadMode = "write"
)

// ============================================================
// Input: Grafana datasource config
// ============================================================

// DatasourceConfig represents a Grafana datasource storage payload.
// This is an internal runtime representation — use NewDatasourceConfig
// to parse from a raw Grafana datasource resource.
//
// In Grafana's storage model, root-level fields (url, basicAuth, etc.)
// sit at the top level alongside jsonData and secureJsonData. The Root
// map groups those top-level fields for programmatic access.
type DatasourceConfig struct {
	// Root-level fields (url, basicAuth, basicAuthUser, etc.)
	// These are top-level in Grafana's storage, not nested under "root".
	Root map[string]any

	// Plugin-specific JSON config.
	JSONData map[string]any

	// Secret values (write path only).
	SecureJSONData map[string]any

	// Boolean map indicating which secrets are configured (read path).
	SecureJSONFields map[string]bool
}

// NewDatasourceConfig creates a DatasourceConfig from a flat Grafana
// datasource resource (as stored in the database or provisioning YAML).
// It extracts root fields, jsonData, secureJsonData, and secureJsonFields
// from the raw map.
func NewDatasourceConfig(raw map[string]any) DatasourceConfig {
	dc := DatasourceConfig{
		Root: make(map[string]any),
	}

	for k, v := range raw {
		switch k {
		case "jsonData":
			if m, ok := v.(map[string]any); ok {
				dc.JSONData = m
			}
		case "secureJsonData":
			if m, ok := v.(map[string]any); ok {
				dc.SecureJSONData = m
			}
		case "secureJsonFields":
			if m, ok := v.(map[string]any); ok {
				dc.SecureJSONFields = make(map[string]bool, len(m))
				for sk, sv := range m {
					if b, ok := sv.(bool); ok {
						dc.SecureJSONFields[sk] = b
					}
				}
			}
		default:
			dc.Root[k] = v
		}
	}

	return dc
}

// ============================================================
// Output: Load result
// ============================================================

// ConfigError represents a validation error for a specific field.
type ConfigError struct {
	// FieldID is the schema field ID (e.g. "auth.password").
	FieldID string `json:"fieldId"`

	// Path is the storage path (e.g. "secureJsonData.password").
	Path string `json:"path"`

	// Code is a machine-readable error code.
	Code string `json:"code"`

	// Message is a human-readable description.
	Message string `json:"message"`
}

// SecureState represents the state of a secure field.
type SecureState string

const (
	SecureUnset      SecureState = "unset"
	SecureConfigured SecureState = "configured"
	SecureUpdated    SecureState = "updated"
)

// FieldValue holds the extracted and resolved value for a single field.
type FieldValue struct {
	// Value is the extracted (and possibly defaulted) value.
	Value any

	// Source indicates where the value came from.
	Source ValueSource
}

// ValueSource indicates the origin of a field value.
type ValueSource string

const (
	SourceConfig  ValueSource = "config"  // extracted from config
	SourceDefault ValueSource = "default" // applied from schema default
	SourceNone    ValueSource = "none"    // not present, no default
)

// LoadResult is the output of LoadAndValidate.
type LoadResult struct {
	// Errors contains all validation errors found.
	Errors []ConfigError `json:"errors,omitempty"`

	// Values contains extracted field values keyed by field ID.
	Values map[string]FieldValue `json:"values,omitempty"`

	// SecureFields contains secret field states keyed by field ID.
	SecureFields map[string]SecureState `json:"secureFields,omitempty"`
}

// HasErrors returns true if any validation errors were found.
func (r *LoadResult) HasErrors() bool {
	return len(r.Errors) > 0
}

// ============================================================
// Pipeline: LoadAndValidate
// ============================================================

// LoadAndValidate extracts field values from a Grafana datasource config
// using the schema, applies defaults, validates values against field
// rules, and returns the result.
func LoadAndValidate(s *DatasourceConfigSchema, config DatasourceConfig, mode LoadMode) (*LoadResult, error) {
	if err := s.Validate(); err != nil {
		return nil, fmt.Errorf("invalid schema: %w", err)
	}

	result := &LoadResult{
		Values:       make(map[string]FieldValue),
		SecureFields: make(map[string]SecureState),
	}

	for i := range s.Fields {
		loadField(&s.Fields[i], config, mode, result)
	}

	return result, nil
}

// loadField processes a single top-level field.
func loadField(f *ConfigField, config DatasourceConfig, mode LoadMode, result *LoadResult) {
	if f.Kind == VirtualField {
		return
	}

	// Secure field handling
	if f.Target != nil && *f.Target == SecureJSONTarget {
		loadSecureField(f, config, mode, result)
		return
	}

	// Extract value
	raw := extractValue(config, f)

	// Expand indexedPair storage mapping
	if f.Storage != nil && f.Storage.Type == IndexedPairMapping {
		raw = expandIndexedPair(config, f)
	}

	// Apply default
	source := SourceConfig
	if raw == nil {
		if f.DefaultValue != nil {
			raw = f.DefaultValue
			source = SourceDefault
		} else {
			source = SourceNone
		}
	}

	result.Values[f.ID] = FieldValue{Value: raw, Source: source}

	// Validate
	validateFieldValue(f, raw, result)

	// Recurse into array item fields for validation
	if f.ValueType == ArrayType && f.Item != nil && raw != nil {
		if arr, ok := raw.([]any); ok {
			validateArrayItems(f, arr, result)
		}
	}
}

// loadSecureField handles fields targeting secureJsonData.
func loadSecureField(f *ConfigField, config DatasourceConfig, mode LoadMode, result *LoadResult) {
	switch mode {
	case WriteMode:
		val := config.SecureJSONData[f.Key]
		if val != nil {
			result.SecureFields[f.ID] = SecureUpdated
			result.Values[f.ID] = FieldValue{Value: val, Source: SourceConfig}
		} else {
			result.SecureFields[f.ID] = SecureUnset
			result.Values[f.ID] = FieldValue{Value: nil, Source: SourceNone}
		}
		// Validate required in write mode
		if f.Required && val == nil {
			result.Errors = append(result.Errors, ConfigError{
				FieldID: f.ID,
				Path:    f.Path(),
				Code:    "required",
				Message: fmt.Sprintf("field %s is required", f.ID),
			})
		}

	case ReadMode:
		if config.SecureJSONFields[f.Key] {
			result.SecureFields[f.ID] = SecureConfigured
		} else {
			result.SecureFields[f.ID] = SecureUnset
			if f.Required {
				result.Errors = append(result.Errors, ConfigError{
					FieldID: f.ID,
					Path:    f.Path(),
					Code:    "required",
					Message: fmt.Sprintf("field %s is required", f.ID),
				})
			}
		}
		// No value extraction in read mode for secrets
		result.Values[f.ID] = FieldValue{Value: nil, Source: SourceNone}
	}
}

// ============================================================
// Value extraction
// ============================================================

// extractValue reads a field's value from the correct config location.
func extractValue(config DatasourceConfig, f *ConfigField) any {
	if f.Target == nil {
		return nil
	}

	switch *f.Target {
	case RootTarget:
		return config.Root[f.Key]

	case JSONDataTarget:
		if config.JSONData == nil {
			return nil
		}
		if f.Section == "" {
			return config.JSONData[f.Key]
		}
		return extractFromSection(config.JSONData, f.Section, f.Key)

	case SecureJSONTarget:
		// Handled separately by loadSecureField
		return nil

	default:
		return nil
	}
}

// extractFromSection walks a dotted section path through a nested map
// and returns the value at the final key.
func extractFromSection(data map[string]any, section, key string) any {
	segments := strings.Split(section, ".")
	current := data

	for _, seg := range segments {
		v, ok := current[seg]
		if !ok {
			return nil
		}
		m, ok := v.(map[string]any)
		if !ok {
			return nil
		}
		current = m
	}

	return current[key]
}

// ============================================================
// IndexedPair expansion
// ============================================================

// expandIndexedPair reads indexed key/value pairs from the config
// and assembles them into an array of objects.
func expandIndexedPair(config DatasourceConfig, f *ConfigField) any {
	if f.Storage == nil || f.Storage.Type != IndexedPairMapping {
		return nil
	}
	if f.Storage.Key == nil || f.Storage.Value == nil {
		return nil
	}
	if f.Item == nil || len(f.Item.Fields) < 2 {
		return nil
	}

	startIndex := 1
	if f.Storage.StartIndex != nil {
		startIndex = *f.Storage.StartIndex
	}

	keyPattern := f.Storage.Key.Pattern
	valuePattern := f.Storage.Value.Pattern

	// Determine item field keys (first field = key name, second = value name)
	keyFieldKey := f.Item.Fields[0].Key
	valueFieldKey := f.Item.Fields[1].Key

	var items []any

	for i := startIndex; ; i++ {
		idx := strconv.Itoa(i)
		kName := strings.ReplaceAll(keyPattern, "{index}", idx)
		vName := strings.ReplaceAll(valuePattern, "{index}", idx)

		kVal := getFromTarget(config, f.Storage.Key.Target, kName)
		if kVal == nil {
			break // Stop when we hit a gap
		}

		vVal := getFromTarget(config, f.Storage.Value.Target, vName)

		item := map[string]any{
			keyFieldKey:   kVal,
			valueFieldKey: vVal,
		}
		items = append(items, item)
	}

	if items == nil {
		return nil
	}
	return items
}

// getFromTarget reads a key from the appropriate config section.
func getFromTarget(config DatasourceConfig, target TargetLocation, key string) any {
	switch target {
	case RootTarget:
		return config.Root[key]
	case JSONDataTarget:
		if config.JSONData == nil {
			return nil
		}
		return config.JSONData[key]
	case SecureJSONTarget:
		if config.SecureJSONData == nil {
			return nil
		}
		return config.SecureJSONData[key]
	default:
		return nil
	}
}

// ============================================================
// Value validation
// ============================================================

// validateFieldValue validates an extracted value against field rules.
func validateFieldValue(f *ConfigField, value any, result *LoadResult) {
	// Required check
	if f.Required && value == nil {
		result.Errors = append(result.Errors, ConfigError{
			FieldID: f.ID,
			Path:    f.Path(),
			Code:    "required",
			Message: fmt.Sprintf("field %s is required", f.ID),
		})
		return // No further validation on nil
	}

	if value == nil {
		return
	}

	// Type check
	if err := checkValueType(value, f.ValueType); err != nil {
		result.Errors = append(result.Errors, ConfigError{
			FieldID: f.ID,
			Path:    f.Path(),
			Code:    "type_mismatch",
			Message: fmt.Sprintf("field %s: %s", f.ID, err),
		})
		return // No further validation on wrong type
	}

	// Validation rules
	for _, rule := range f.Validations {
		if err := evaluateRule(value, rule, f); err != nil {
			msg := err.Error()
			if rule.Message != "" {
				msg = rule.Message
			}
			result.Errors = append(result.Errors, ConfigError{
				FieldID: f.ID,
				Path:    f.Path(),
				Code:    string(rule.Type),
				Message: msg,
			})
		}
	}
}

// checkValueType verifies that a value matches the expected valueType.
func checkValueType(value any, expected ValueType) error {
	switch expected {
	case StringType:
		if _, ok := value.(string); !ok {
			return fmt.Errorf("expected string, got %T", value)
		}
	case NumberType:
		switch value.(type) {
		case float64, float32, int, int64, int32:
			// ok — JSON numbers unmarshal as float64
		default:
			return fmt.Errorf("expected number, got %T", value)
		}
	case BooleanType:
		if _, ok := value.(bool); !ok {
			return fmt.Errorf("expected boolean, got %T", value)
		}
	case ArrayType:
		if _, ok := value.([]any); !ok {
			return fmt.Errorf("expected array, got %T", value)
		}
	case ObjectType:
		if _, ok := value.(map[string]any); !ok {
			return fmt.Errorf("expected object, got %T", value)
		}
	}
	return nil
}

// evaluateRule validates a value against a single validation rule.
func evaluateRule(value any, rule FieldValidationRule, f *ConfigField) error {
	switch rule.Type {
	case PatternValidation:
		s, ok := value.(string)
		if !ok {
			return nil // Type mismatch caught separately
		}
		matched, err := regexp.MatchString(rule.Pattern, s)
		if err != nil {
			return fmt.Errorf("invalid regex pattern %q: %w", rule.Pattern, err)
		}
		if !matched {
			return fmt.Errorf("value does not match pattern %q", rule.Pattern)
		}

	case RangeValidation:
		n, ok := toFloat64(value)
		if !ok {
			return nil
		}
		if rule.Min != nil && n < *rule.Min {
			return fmt.Errorf("value %v is below minimum %v", n, *rule.Min)
		}
		if rule.Max != nil && n > *rule.Max {
			return fmt.Errorf("value %v exceeds maximum %v", n, *rule.Max)
		}

	case LengthValidation:
		s, ok := value.(string)
		if !ok {
			return nil
		}
		l := float64(len(s))
		if rule.Min != nil && l < *rule.Min {
			return fmt.Errorf("length %d is below minimum %v", len(s), *rule.Min)
		}
		if rule.Max != nil && l > *rule.Max {
			return fmt.Errorf("length %d exceeds maximum %v", len(s), *rule.Max)
		}

	case ItemCountValidation:
		arr, ok := value.([]any)
		if !ok {
			return nil
		}
		l := float64(len(arr))
		if rule.Min != nil && l < *rule.Min {
			return fmt.Errorf("item count %d is below minimum %v", len(arr), *rule.Min)
		}
		if rule.Max != nil && l > *rule.Max {
			return fmt.Errorf("item count %d exceeds maximum %v", len(arr), *rule.Max)
		}

	case AllowedValuesValidation:
		for _, allowed := range rule.Values {
			if fmt.Sprintf("%v", value) == fmt.Sprintf("%v", allowed) {
				return nil
			}
		}
		return fmt.Errorf("value %v is not in allowed values %v", value, rule.Values)

	case CustomValidation:
		// CEL not evaluated — skip silently
		return nil
	}

	return nil
}

// validateArrayItems validates each element of an array against item field rules.
func validateArrayItems(f *ConfigField, arr []any, result *LoadResult) {
	if f.Item == nil {
		return
	}

	for i, elem := range arr {
		if f.Item.ValueType == ObjectType && len(f.Item.Fields) > 0 {
			obj, ok := elem.(map[string]any)
			if !ok {
				result.Errors = append(result.Errors, ConfigError{
					FieldID: f.ID,
					Path:    fmt.Sprintf("%s[%d]", f.Path(), i),
					Code:    "type_mismatch",
					Message: fmt.Sprintf("field %s[%d]: expected object, got %T", f.ID, i, elem),
				})
				continue
			}
			for j := range f.Item.Fields {
				sub := &f.Item.Fields[j]
				val := obj[sub.Key]
				if val == nil && sub.DefaultValue != nil {
					val = sub.DefaultValue
				}
				validateFieldValue(sub, val, result)
			}
		}
	}
}

// toFloat64 converts numeric types to float64 for comparison.
func toFloat64(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case int32:
		return float64(n), true
	default:
		return 0, false
	}
}
