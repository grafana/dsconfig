package schema

import "fmt"

// ============================================================
// Root Schema
// ============================================================

// DatasourceConfigSchema is the top-level schema definition.
// It acts as the single source of truth for datasource configuration.
type DatasourceConfigSchema struct {
	// SchemaVersion defines the version of the schema spec.
	SchemaVersion string `json:"schemaVersion"`

	// PluginType uniquely identifies the datasource plugin.
	PluginType string `json:"pluginType"`

	// PluginName is a human-readable name.
	PluginName string `json:"pluginName"`

	// Optional documentation URL.
	DocURL string `json:"docURL,omitempty"`

	// Fields defines all configuration fields.
	Fields []ConfigField `json:"fields"`

	// Optional UI grouping
	Groups []ConfigGroup `json:"groups,omitempty"`

	// Relationships between fields
	Relationships []FieldRelationship `json:"relationships,omitempty"`
}

func (s *DatasourceConfigSchema) Validate() error {
	if s.SchemaVersion == "" {
		return fmt.Errorf("schemaVersion is required")
	}
	if s.PluginType == "" {
		return fmt.Errorf("pluginType is required")
	}
	if s.PluginName == "" {
		return fmt.Errorf("pluginName is required")
	}
	if len(s.Fields) == 0 {
		return fmt.Errorf("fields is required")
	}

	for i := range s.Fields {
		if err := s.Fields[i].Validate(); err != nil {
			return err
		}
	}

	fieldIDs, err := s.FieldIDs()
	if err != nil {
		return err
	}

	if err := s.ValidateRefs(fieldIDs); err != nil {
		return err
	}

	return nil
}

// ValidateRefs checks that all group and relationship field references
// point to existing field IDs.
func (s *DatasourceConfigSchema) ValidateRefs(fieldIDs map[string]bool) error {
	for _, g := range s.Groups {
		for _, ref := range g.FieldRefs {
			if !fieldIDs[ref] {
				return fmt.Errorf("group %s references unknown field id: %s", g.ID, ref)
			}
		}
	}

	for _, r := range s.Relationships {
		if !r.Type.IsValid() {
			return fmt.Errorf("relationship has invalid type %q", r.Type)
		}
		for _, ref := range r.Fields {
			if !fieldIDs[ref] {
				return fmt.Errorf("relationship references unknown field id: %s", ref)
			}
		}
	}

	return nil
}

// ============================================================
// Field Definition
// ============================================================

// ConfigField represents a single configuration field.
type ConfigField struct {
	// ID is globally unique (used for references)
	ID string `json:"id"`

	// Key is the local key (used in storage or object structures)
	Key string `json:"key"`

	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`
	DocURL      string `json:"docURL,omitempty"`

	// Core typing
	ValueType    ValueType    `json:"valueType"`
	SemanticType SemanticType `json:"semanticType,omitempty"`

	// Storage location (required for storage fields)
	Target *TargetLocation `json:"target,omitempty"`

	// Field type: storage (default) or virtual
	Kind FieldKind `json:"kind,omitempty"`

	// True if part of array item schema
	IsItemField *bool `json:"isItemField,omitempty"`

	// Lifecycle: stable / deprecated / experimental
	Lifecycle Lifecycle `json:"lifecycle,omitempty"`

	// UI hints
	UI *FieldUI `json:"ui,omitempty"`

	// Validation rules
	Validations []FieldValidationRule `json:"validations,omitempty"`

	// Conditional behavior (CEL)
	DependsOn    string `json:"dependsOn,omitempty"`
	Required     bool   `json:"required,omitempty"`
	RequiredWhen string `json:"requiredWhen,omitempty"`
	DisabledWhen string `json:"disabledWhen,omitempty"`

	// Dynamic overrides
	Overrides []FieldOverride `json:"overrides,omitempty"`

	// Array schema (required when ValueType == array)
	Item *FieldItemSchema `json:"item,omitempty"`

	// Legacy indexed fields
	Repeatable bool   `json:"repeatable,omitempty"`
	Pattern    string `json:"pattern,omitempty"`

	// Storage mapping layer
	Storage *StorageMapping `json:"storage,omitempty"`

	// Metadata
	Tags         []string `json:"tags,omitempty"`
	Examples     []any    `json:"examples,omitempty"`
	DefaultValue any      `json:"defaultValue,omitempty"`
}

func (f *ConfigField) Validate() error {
	if f.ID == "" {
		return fmt.Errorf("field id is required")
	}
	if f.Key == "" {
		return fmt.Errorf("field %s: key is required", f.ID)
	}
	if !f.ValueType.IsValid() {
		return fmt.Errorf("field %s: invalid valueType %q", f.ID, f.ValueType)
	}

	isVirtual := f.Kind == VirtualField
	isItem := f.IsItemField != nil && *f.IsItemField

	if !isVirtual && !isItem && f.Target == nil {
		return fmt.Errorf("field %s: target is required for storage fields", f.ID)
	}

	if f.ValueType == ArrayType && f.Item == nil {
		return fmt.Errorf("field %s: item is required for array fields", f.ID)
	}

	if f.Storage != nil {
		if err := f.Storage.Validate(); err != nil {
			return fmt.Errorf("field %s: invalid storage mapping: %w", f.ID, err)
		}
	}

	if f.Kind != "" && !f.Kind.IsValid() {
		return fmt.Errorf("field %s: invalid kind %q", f.ID, f.Kind)
	}

	if f.SemanticType != "" && !f.SemanticType.IsValid() {
		return fmt.Errorf("field %s: invalid semanticType %q", f.ID, f.SemanticType)
	}

	if f.Lifecycle != "" && !f.Lifecycle.IsValid() {
		return fmt.Errorf("field %s: invalid lifecycle %q", f.ID, f.Lifecycle)
	}

	if f.UI != nil {
		if !f.UI.Component.IsValid() {
			return fmt.Errorf("field %s: invalid ui component %q", f.ID, f.UI.Component)
		}
		if f.UI.Width != "" && !f.UI.Width.IsValid() {
			return fmt.Errorf("field %s: invalid ui width %q", f.ID, f.UI.Width)
		}
		for i, opt := range f.UI.Options {
			if !ValidateOptionValue(opt.Value, f.ValueType) {
				return fmt.Errorf("field %s: ui option[%d] value type mismatch (expected %s)", f.ID, i, f.ValueType)
			}
		}
	}

	if f.Target != nil && !f.Target.IsValid() {
		return fmt.Errorf("field %s: invalid target: %s", f.ID, *f.Target)
	}

	if f.Item != nil {
		if !f.Item.ValueType.IsValid() {
			return fmt.Errorf("field %s: invalid item valueType %q", f.ID, f.Item.ValueType)
		}
		if f.Item.ValueType != ObjectType && len(f.Item.Fields) > 0 {
			return fmt.Errorf("field %s: item fields are only allowed when item valueType is object", f.ID)
		}
		for i := range f.Item.Fields {
			sub := &f.Item.Fields[i]
			if sub.IsItemField == nil || !*sub.IsItemField {
				return fmt.Errorf("field %s: item field %s must have isItemField=true", f.ID, sub.ID)
			}
			if err := sub.Validate(); err != nil {
				return fmt.Errorf("field %s: invalid item field %s: %w", f.ID, sub.ID, err)
			}
		}
	}

	for i := range f.Validations {
		if err := f.Validations[i].Validate(); err != nil {
			return fmt.Errorf("field %s: invalid validation rule: %w", f.ID, err)
		}
	}

	for i := range f.Overrides {
		for j := range f.Overrides[i].Validations {
			if err := f.Overrides[i].Validations[j].Validate(); err != nil {
				return fmt.Errorf("field %s: invalid override validation rule: %w", f.ID, err)
			}
		}
	}

	return nil
}

func (s *DatasourceConfigSchema) FieldIDs() (map[string]bool, error) {
	seen := map[string]bool{}

	var visit func(fields []ConfigField) error
	visit = func(fields []ConfigField) error {
		for i := range fields {
			f := fields[i]

			if f.ID == "" {
				return fmt.Errorf("field id is required")
			}

			if seen[f.ID] {
				return fmt.Errorf("duplicate field id: %s", f.ID)
			}
			seen[f.ID] = true

			if f.Item != nil {
				if err := visit(f.Item.Fields); err != nil {
					return err
				}
			}
		}

		return nil
	}

	if err := visit(s.Fields); err != nil {
		return nil, err
	}

	return seen, nil
}

func (f ConfigField) Path() string {
	if f.Target == nil {
		return f.Key
	}
	return string(*f.Target) + "." + f.Key
}

// ============================================================
// Array Item Schema
// ============================================================

// FieldItemSchema defines schema for array elements.
type FieldItemSchema struct {
	ValueType ValueType     `json:"valueType"`
	Fields    []ConfigField `json:"fields,omitempty"`
}

// ============================================================
// Value Types
// ============================================================

type ValueType string

const (
	StringType  ValueType = "string"
	NumberType  ValueType = "number"
	BooleanType ValueType = "boolean"
	ArrayType   ValueType = "array"
	ObjectType  ValueType = "object"
)

func (v ValueType) IsValid() bool {
	switch v {
	case StringType, NumberType, BooleanType, ArrayType, ObjectType:
		return true
	default:
		return false
	}
}

// ============================================================
// Semantic Types
// ============================================================

type SemanticType string

const (
	URLType      SemanticType = "url"
	PasswordType SemanticType = "password"
	TokenType    SemanticType = "token"
	HostnameType SemanticType = "hostname"
	DurationType SemanticType = "duration"
)

func (s SemanticType) IsValid() bool {
	switch s {
	case URLType, PasswordType, TokenType, HostnameType, DurationType:
		return true
	default:
		return false
	}
}

// ============================================================
// Field Kind
// ============================================================

type FieldKind string

const (
	StorageField FieldKind = "storage"
	VirtualField FieldKind = "virtual"
)

func (k FieldKind) IsValid() bool {
	switch k {
	case StorageField, VirtualField:
		return true
	default:
		return false
	}
}

// ============================================================
// Lifecycle
// ============================================================

type Lifecycle string

const (
	StableLifecycle       Lifecycle = "stable"
	DeprecatedLifecycle   Lifecycle = "deprecated"
	ExperimentalLifecycle Lifecycle = "experimental"
)

func (l Lifecycle) IsValid() bool {
	switch l {
	case StableLifecycle, DeprecatedLifecycle, ExperimentalLifecycle:
		return true
	default:
		return false
	}
}

// ============================================================
// Target Location
// ============================================================

type TargetLocation string

const (
	RootTarget       TargetLocation = "root"
	JSONDataTarget   TargetLocation = "jsonData"
	SecureJSONTarget TargetLocation = "secureJsonData"
)

func (t TargetLocation) IsValid() bool {
	switch t {
	case RootTarget, JSONDataTarget, SecureJSONTarget:
		return true
	default:
		return false
	}
}

// ============================================================
// UI Components
// ============================================================

// UIComponent defines supported UI elements.
type UIComponent string

const (
	UIInput       UIComponent = "input"
	UITextarea    UIComponent = "textarea"
	UISelect      UIComponent = "select"
	UIMultiselect UIComponent = "multiselect"
	UIRadio       UIComponent = "radio"
	UICheckbox    UIComponent = "checkbox"
	UISwitch      UIComponent = "switch"
	UICode        UIComponent = "code"
	UIKeyValue    UIComponent = "keyvalue"
	UIList        UIComponent = "list"
)

func (c UIComponent) IsValid() bool {
	switch c {
	case UIInput, UITextarea, UISelect, UIMultiselect, UIRadio,
		UICheckbox, UISwitch, UICode, UIKeyValue, UIList:
		return true
	default:
		return false
	}
}

// FieldUI defines UI rendering hints.
type FieldUI struct {
	Component UIComponent `json:"component"`

	Multiline bool          `json:"multiline,omitempty"`
	Rows      int           `json:"rows,omitempty"`
	Options   []FieldOption `json:"options,omitempty"`

	AllowCustom bool    `json:"allowCustom,omitempty"`
	Width       UIWidth `json:"width,omitempty"`

	Placeholder string `json:"placeholder,omitempty"`
}

// UIWidth defines layout width.
type UIWidth string

const (
	FullWidth UIWidth = "full"
	HalfWidth UIWidth = "half"
)

func (w UIWidth) IsValid() bool {
	switch w {
	case FullWidth, HalfWidth:
		return true
	default:
		return false
	}
}

// ============================================================
// Validations
// ============================================================

// ValidationRuleType defines the kind of validation rule.
type ValidationRuleType string

const (
	PatternValidation       ValidationRuleType = "pattern"
	RangeValidation         ValidationRuleType = "range"
	LengthValidation        ValidationRuleType = "length"
	ItemCountValidation     ValidationRuleType = "itemCount"
	AllowedValuesValidation ValidationRuleType = "allowedValues"
	CustomValidation        ValidationRuleType = "custom"
)

// FieldValidationRule is a discriminated union of validation rules.
type FieldValidationRule struct {
	Type    ValidationRuleType `json:"type"`
	ID      string             `json:"id,omitempty"`
	Message string             `json:"message,omitempty"`

	// PatternValidation
	Pattern string `json:"pattern,omitempty"`

	// RangeValidation / LengthValidation / ItemCountValidation
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`

	// AllowedValuesValidation
	Values []any `json:"values,omitempty"`

	// CustomValidation
	Expression string `json:"expression,omitempty"`
}

func (r *FieldValidationRule) Validate() error {
	switch r.Type {
	case PatternValidation:
		if r.Pattern == "" {
			return fmt.Errorf("pattern validation requires pattern")
		}
	case RangeValidation, LengthValidation, ItemCountValidation:
		if r.Min == nil && r.Max == nil {
			return fmt.Errorf("%s validation requires min or max", r.Type)
		}
	case AllowedValuesValidation:
		if len(r.Values) == 0 {
			return fmt.Errorf("allowedValues validation requires values")
		}
	case CustomValidation:
		if r.Expression == "" {
			return fmt.Errorf("custom validation requires expression")
		}
	default:
		return fmt.Errorf("unknown validation rule type: %s", r.Type)
	}
	return nil
}

// ============================================================
// Overrides
// ============================================================

// FieldOverride allows dynamic modifications.
type FieldOverride struct {
	When string `json:"when"`

	DefaultValue any    `json:"defaultValue,omitempty"`
	Description  string `json:"description,omitempty"`
	Placeholder  string `json:"placeholder,omitempty"`
	Tooltip      string `json:"tooltip,omitempty"`

	Validations []FieldValidationRule `json:"validations,omitempty"`
	Options     []FieldOption         `json:"options,omitempty"`
}

// ============================================================
// Storage Mapping
// ============================================================

// StorageMappingType defines mapping strategy.
type StorageMappingType string

const (
	DirectMapping      StorageMappingType = "direct"
	IndexedPairMapping StorageMappingType = "indexedPair"
	ComputedMapping    StorageMappingType = "computed"
)

// StorageMapping maps logical fields to Grafana storage.
type StorageMapping struct {
	Type StorageMappingType `json:"type"`

	// Indexed pair mapping
	Key        *MappingField `json:"key,omitempty"`
	Value      *MappingField `json:"value,omitempty"`
	StartIndex *int          `json:"startIndex,omitempty"`

	// Computed mapping
	Read  string `json:"read,omitempty"`
	Write string `json:"write,omitempty"`
}

func (m *StorageMapping) Validate() error {
	switch m.Type {
	case DirectMapping:
		if m.Key != nil || m.Value != nil || m.StartIndex != nil || m.Read != "" || m.Write != "" {
			return fmt.Errorf("direct mapping must not have key/value/startIndex/read/write")
		}

	case IndexedPairMapping:
		if m.Key == nil || m.Value == nil {
			return fmt.Errorf("indexedPair requires key and value")
		}
		if m.Read != "" || m.Write != "" {
			return fmt.Errorf("indexedPair must not have read/write")
		}
		if err := m.Key.Validate(); err != nil {
			return fmt.Errorf("indexedPair key: %w", err)
		}
		if err := m.Value.Validate(); err != nil {
			return fmt.Errorf("indexedPair value: %w", err)
		}

	case ComputedMapping:
		if m.Read == "" && m.Write == "" {
			return fmt.Errorf("computed mapping requires read or write")
		}
		if m.Key != nil || m.Value != nil || m.StartIndex != nil {
			return fmt.Errorf("computed mapping must not have key/value/startIndex")
		}

	default:
		return fmt.Errorf("unknown mapping type: %s", m.Type)
	}

	return nil
}

// MappingField describes a mapped field.
type MappingField struct {
	Target  TargetLocation `json:"target"`
	Pattern string         `json:"pattern"`
}

func (m MappingField) Validate() error {
	if !m.Target.IsValid() {
		return fmt.Errorf("invalid target %q", m.Target)
	}
	if m.Pattern == "" {
		return fmt.Errorf("pattern is required")
	}
	return nil
}

// ============================================================
// Options
// ============================================================

type FieldOption struct {
	Label string `json:"label"`
	Value any    `json:"value"`
}

// ValidateOptionValue checks that an option value is compatible with
// the given field valueType.
func ValidateOptionValue(v any, vt ValueType) bool {
	if v == nil {
		return true
	}
	switch vt {
	case StringType:
		_, ok := v.(string)
		return ok
	case NumberType:
		switch v.(type) {
		case int, int64, float64, float32:
			return true
		default:
			return false
		}
	case BooleanType:
		_, ok := v.(bool)
		return ok
	default:
		// array/object options are not type-checked
		return true
	}
}

// ============================================================
// Groups
// ============================================================

type ConfigGroup struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Order       *int     `json:"order,omitempty"`
	FieldRefs   []string `json:"fieldRefs"`
}

// ============================================================
// Relationships
// ============================================================

type RelationshipType string

const (
	PairRelationship  RelationshipType = "pair"
	GroupRelationship RelationshipType = "group"
)

func (r RelationshipType) IsValid() bool {
	switch r {
	case PairRelationship, GroupRelationship:
		return true
	default:
		return false
	}
}

type FieldRelationship struct {
	Type        RelationshipType `json:"type"`
	Fields      []string         `json:"fields"`
	Description string           `json:"description,omitempty"`
}
