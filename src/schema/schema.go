package schema

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
	Validation *FieldValidation `json:"validation,omitempty"`

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
	Tags         []string      `json:"tags,omitempty"`
	Examples     []interface{} `json:"examples,omitempty"`
	DefaultValue interface{}   `json:"defaultValue,omitempty"`
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

// ============================================================
// Field Kind
// ============================================================

type FieldKind string

const (
	StorageField FieldKind = "storage"
	VirtualField FieldKind = "virtual"
)

// ============================================================
// Lifecycle
// ============================================================

type Lifecycle string

const (
	StableLifecycle       Lifecycle = "stable"
	DeprecatedLifecycle   Lifecycle = "deprecated"
	ExperimentalLifecycle Lifecycle = "experimental"
)

// ============================================================
// Target Location
// ============================================================

type TargetLocation string

const (
	RootTarget       TargetLocation = "root"
	JSONDataTarget   TargetLocation = "jsonData"
	SecureJSONTarget TargetLocation = "secureJsonData"
)

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

// ============================================================
// Validation
// ============================================================

type FieldValidation struct {
	Pattern  string   `json:"pattern,omitempty"`
	Message  string   `json:"message,omitempty"`
	Min      *float64 `json:"min,omitempty"`
	Max      *float64 `json:"max,omitempty"`
	MinItems *int     `json:"minItems,omitempty"`
	MaxItems *int     `json:"maxItems,omitempty"`
}

// ============================================================
// Overrides
// ============================================================

// FieldOverride allows dynamic modifications.
type FieldOverride struct {
	When string `json:"when"`

	DefaultValue interface{} `json:"defaultValue,omitempty"`
	Description  string      `json:"description,omitempty"`
	Placeholder  string      `json:"placeholder,omitempty"`
	Tooltip      string      `json:"tooltip,omitempty"`

	Validation *FieldValidation `json:"validation,omitempty"`
	Options    []FieldOption    `json:"options,omitempty"`
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

// MappingField describes a mapped field.
type MappingField struct {
	Target  TargetLocation `json:"target"`
	Pattern string         `json:"pattern"`
}

// ============================================================
// Options
// ============================================================

type FieldOption struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
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

type FieldRelationship struct {
	Type        RelationshipType `json:"type"`
	Fields      []string         `json:"fields"`
	Description string           `json:"description,omitempty"`
}
