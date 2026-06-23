package dsconfig

import (
	"encoding/json"
	"fmt"
)

// ============================================================
// Root Schema
// ============================================================

// Schema is the top-level schema definition.
// It acts as the single source of truth for datasource configuration.
type Schema struct {
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

	// Optional Instruction
	Instructions []Instruction `json:"instructions,omitempty"`

	// Relationships between fields
	Relationships []FieldRelationship `json:"relationships,omitempty"`
}

func (s *Schema) Validate() error {
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
func (s *Schema) ValidateRefs(fieldIDs map[string]bool) error {
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

	// Validate effect set keys reference known field IDs
	var visitEffects func(fields []ConfigField) error
	visitEffects = func(fields []ConfigField) error {
		for _, f := range fields {
			for i, eff := range f.Effects {
				for ref := range eff.Set {
					if !fieldIDs[ref] {
						return fmt.Errorf("field %s: effect[%d].set references unknown field id: %s", f.ID, i, ref)
					}
				}
			}
			if f.Item != nil {
				if err := visitEffects(f.Item.Fields); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := visitEffects(s.Fields); err != nil {
		return err
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

	Label       string     `json:"label,omitempty"`
	Description string     `json:"description,omitempty"`
	DocURL      string     `json:"docURL,omitempty"`
	Help        *FieldHelp `json:"help,omitempty"`

	// Core typing
	ValueType ValueType `json:"valueType"`

	// Storage location (required for storage fields)
	Target *TargetLocation `json:"target,omitempty"`

	// Section is the dotted path prefix within the target for nested objects.
	// Example: for jsonData.tracesToLogs.datasourceUid, target="jsonData",
	// section="tracesToLogs", key="datasourceUid".
	Section string `json:"section,omitempty"`

	// Field type: storage (default) or virtual
	Kind FieldKind `json:"kind,omitempty"`

	Role Role `json:"role,omitempty"`

	// True if part of array item schema
	IsItemField *bool `json:"isItemField,omitempty"`

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

	// Effects: declarative multi-field write side-effects.
	// When this field's value matches a condition, the listed target
	// fields are set to the specified values. Typically used on virtual
	// selector fields (e.g. auth method dropdown) to drive multiple
	// storage fields without opaque CEL expressions.
	Effects []FieldEffect `json:"effects,omitempty"`

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

	if f.Section != "" && isItem {
		return fmt.Errorf("field %s: section is not allowed on item fields", f.ID)
	}
	if f.Section != "" && isVirtual {
		return fmt.Errorf("field %s: section is not allowed on virtual fields", f.ID)
	}

	if (f.ValueType == ArrayType || f.ValueType == MapType) && f.Item == nil {
		return fmt.Errorf("field %s: item is required for array and map fields", f.ID)
	}

	if f.Storage != nil {
		if err := f.Storage.Validate(); err != nil {
			return fmt.Errorf("field %s: invalid storage mapping: %w", f.ID, err)
		}
	}

	if f.Kind != "" && !f.Kind.IsValid() {
		return fmt.Errorf("field %s: invalid kind %q", f.ID, f.Kind)
	}

	if f.Role != "" && !f.Role.IsValid() {
		return fmt.Errorf("field %s: invalid role %q", f.ID, f.Role)
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

	for i := range f.Effects {
		if err := f.Effects[i].Validate(); err != nil {
			return fmt.Errorf("field %s: invalid effect[%d]: %w", f.ID, i, err)
		}
	}

	if f.Help != nil {
		if err := f.Help.Validate(); err != nil {
			return fmt.Errorf("field %s: invalid help: %w", f.ID, err)
		}
	}

	return nil
}

func (s *Schema) FieldIDs() (map[string]bool, error) {
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
	if f.Section != "" {
		return string(*f.Target) + "." + f.Section + "." + f.Key
	}
	return string(*f.Target) + "." + f.Key
}

// ============================================================
// Array Item Schema
// ============================================================

// FieldItemSchema defines schema for array/map elements.
// For arrays, it describes each element.
// For maps, it describes each value (keys are always strings).
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
	MapType     ValueType = "map"
	AnyType     ValueType = "any"
)

func (v ValueType) IsValid() bool {
	switch v {
	case StringType, NumberType, BooleanType, ArrayType, ObjectType, MapType, AnyType:
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

	// Language hint for code editor components.
	// Example: "promql", "logql", "traceql", "sql", "json"
	Language string `json:"language,omitempty"`
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
// Help
// ============================================================

type FieldHelp struct {
	Title    string `json:"title,omitempty"`
	Subtitle string `json:"subtitle,omitempty"`
	Markdown string `json:"markdown"`
	DocURL   string `json:"docURL,omitempty"`
}

func (h *FieldHelp) Validate() error {
	if h.Markdown == "" {
		return fmt.Errorf("help requires markdown content")
	}
	return nil
}

// ============================================================
// Roles
// ============================================================

type Role string

const (
	RoleEndpointBaseURL Role = "endpoint.baseUrl"
	RoleEndpointScheme  Role = "endpoint.scheme"
	RoleEndpointDomain  Role = "endpoint.domain"
	RoleEndpointPort    Role = "endpoint.port"

	RoleTransportTimeoutSeconds Role = "transport.timeoutSeconds"
	RoleTransportTLSSkipVerify  Role = "transport.tlsSkipVerify"

	RoleTLSClientCert Role = "tls.clientCert"
	RoleTLSClientKey  Role = "tls.clientKey"
	RoleTLSCACert     Role = "tls.caCert"
	RoleTLSServerName Role = "tls.serverName"

	RoleAuthDiscriminator            Role = "auth.discriminator"
	RoleAuthBasicEnabled             Role = "auth.basic.enabled"
	RoleAuthBasicUsername            Role = "auth.basic.username"
	RoleAuthBasicPassword            Role = "auth.basic.password"
	RoleAuthOAuth2ClientID           Role = "auth.oauth2.clientId"
	RoleAuthOAuth2ClientSecret       Role = "auth.oauth2.clientSecret"
	RoleAuthOAuth2TokenURL           Role = "auth.oauth2.tokenUrl"
	RoleAuthJWTSigningKey            Role = "auth.jwt.signingKey"
	RoleAuthAWSSigV4Enabled          Role = "auth.awsSigV4.enabled"
	RoleAuthForwardOAuthTokenEnabled Role = "auth.forwardOAuthToken.enabled"

	RoleHTTPHeader      Role = "http.header"
	RoleHTTPHeaderName  Role = "http.header.name"
	RoleHTTPHeaderValue Role = "http.header.value"

	RoleHTTPQuery      Role = "http.query"
	RoleHTTPQueryName  Role = "http.query.name"
	RoleHTTPQueryValue Role = "http.query.value"
)

func (r Role) IsValid() bool {
	switch r {
	case RoleEndpointBaseURL, RoleEndpointScheme, RoleEndpointDomain, RoleEndpointPort,
		RoleTransportTimeoutSeconds, RoleTransportTLSSkipVerify,
		RoleTLSClientCert, RoleTLSClientKey, RoleTLSCACert, RoleTLSServerName,
		RoleAuthDiscriminator, RoleAuthBasicEnabled, RoleAuthBasicUsername, RoleAuthBasicPassword,
		RoleAuthOAuth2ClientID, RoleAuthOAuth2ClientSecret, RoleAuthOAuth2TokenURL, RoleAuthJWTSigningKey,
		RoleAuthAWSSigV4Enabled,
		RoleAuthForwardOAuthTokenEnabled,
		RoleHTTPHeader, RoleHTTPHeaderName, RoleHTTPHeaderValue,
		RoleHTTPQuery, RoleHTTPQueryName, RoleHTTPQueryValue:
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
// Effects
// ============================================================

// FieldEffect declares that when a field's value matches a condition,
// the listed target fields should be set to the specified values.
//
// This provides a structured, machine-readable alternative to opaque
// computed write expressions for virtual selector fields.
//
// Example: an auth method dropdown that sets root.basicAuth and
// jsonData.oauthPassThru depending on which option is selected.
type FieldEffect struct {
	// When is a CEL expression evaluated against the field's value.
	// Convention: use "value" to refer to the field's current value.
	// Example: "value == 'basic-auth'"
	When string `json:"when"`

	// Set maps field IDs to the values they should be set to when
	// the condition matches.
	Set map[string]any `json:"set"`
}

func (e *FieldEffect) Validate() error {
	if e.When == "" {
		return fmt.Errorf("effect when is required")
	}
	if len(e.Set) == 0 {
		return fmt.Errorf("effect set must not be empty")
	}
	return nil
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
	Label       string `json:"label"`
	Value       any    `json:"value"`
	Description string `json:"description,omitempty"`
}

// ValidateOptionValue checks that an option value is non-nil and
// compatible with the given field valueType.
func ValidateOptionValue(v any, vt ValueType) bool {
	if v == nil {
		return false
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
		// array/object/map/any options are not type-checked
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
	UI          *GroupUI `json:"ui,omitempty"`
	Order       *int     `json:"order,omitempty"`
	Optional    bool     `json:"optional,omitempty"`
	FieldRefs   []string `json:"fieldRefs"`
}

type GroupUI struct {
	Icon string `json:"icon,omitempty"`
}

// ============================================================
// Relationships
// ============================================================

type RelationshipType string

const (
	PairRelationship          RelationshipType = "pair"
	GroupRelationship         RelationshipType = "group"
	DatasourceRefRelationship RelationshipType = "datasourceReference"
)

func (r RelationshipType) IsValid() bool {
	switch r {
	case PairRelationship, GroupRelationship, DatasourceRefRelationship:
		return true
	default:
		return false
	}
}

type FieldRelationship struct {
	Type        RelationshipType `json:"type"`
	Fields      []string         `json:"fields"`
	Description string           `json:"description,omitempty"`

	// TargetPluginType constrains the datasource UID to a specific plugin.
	// Only applicable when Type is "datasourceReference".
	TargetPluginType string `json:"targetPluginType,omitempty"`
}

type Instruction struct {
	Message string   `json:"msg"`
	Tags    []string `json:"tags,omitempty"`
}

// ParseSchemaJSON unmarshals dsconfig JSON data into a Schema.
// This is a convenience function for plugins that embed their dsconfig.json
// and need to parse it at runtime.
func ParseSchemaJSON(data []byte) (*Schema, error) {
	var s Schema
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse dsconfig.json: %w", err)
	}
	return &s, nil
}
