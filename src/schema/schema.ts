// ============================================================
// Datasource Configuration Schema (TypeScript)
// ============================================================
//
// This schema defines datasource configuration in a declarative way.
//
// It is used for:
// - UI rendering
// - validation
// - storage mapping
// - documentation
// - LLM / automation tooling
//
// ============================================================


// ============================================================
// Root Schema
// ============================================================

export interface DatasourceConfigSchema {
    /**
     * Schema version (required)
     * Used for compatibility and migrations
     *
     * Example: "v2"
     */
    schemaVersion: string

    /**
     * Unique datasource plugin identifier
     * Example: "prometheus"
     */
    pluginType: string

    /**
     * Human-readable name
     */
    pluginName: string

    /**
     * Optional documentation URL
     */
    docURL?: string

    /**
     * Source of truth for configuration
     */
    fields: ConfigField[]

    /**
     * Optional UI grouping
     */
    groups?: ConfigGroup[]

    /**
     * Relationships between fields
     */
    relationships?: FieldRelationship[]
}


// ============================================================
// Field Definition
// ============================================================

export interface ConfigField {
    /**
     * Globally unique identifier
     * Recommended format: dot-separated path
     * Example: "auth.method", "httpHeaders.item.key"
     */
    id: string

    /**
     * Local key used in storage or object
     */
    key: string

    /**
     * UI metadata
     */
    label?: string
    description?: string
    docURL?: string

    /**
     * Core type of the field
     */
    valueType: ValueType

    /**
     * Semantic meaning (optional)
     */
    semanticType?: SemanticType

    /**
     * Storage location (required for storage fields)
     */
    target?: TargetLocation

    /**
     * Field kind
     * - storage (default)
     * - virtual (derived)
     */
    kind?: FieldKind

    /**
     * True if field belongs to array item schema
     */
    isItemField?: boolean

    /**
     * Lifecycle state
     */
    lifecycle?: Lifecycle

    /**
     * UI rendering hints
     */
    ui?: FieldUI

    /**
     * Validation rules
     */
    validation?: FieldValidation

    /**
     * Visibility condition (CEL expression)
     */
    dependsOn?: Expression

    /**
     * Always required
     */
    required?: boolean

    /**
     * Conditionally required
     */
    requiredWhen?: Expression

    /**
     * Disabled condition
     */
    disabledWhen?: Expression

    /**
     * Dynamic overrides
     */
    overrides?: FieldOverride[]

    /**
     * Array item schema
     * Required if valueType === "array"
     */
    item?: FieldItemSchema

    /**
     * Legacy indexed fields (deprecated)
     */
    repeatable?: boolean
    pattern?: string

    /**
     * Storage mapping
     */
    storage?: StorageMapping

    /**
     * Metadata
     */
    tags?: string[]
    examples?: unknown[]
    defaultValue?: unknown
}


// ============================================================
// Array Item Schema
// ============================================================

export interface FieldItemSchema {
    /**
     * Type of array items
     */
    valueType: ValueType

    /**
     * Required when valueType = "object"
     */
    fields?: ConfigField[]
}


// ============================================================
// Expressions
// ============================================================

/**
 * CEL expression string
 */
export type Expression = string


// ============================================================
// Value Types
// ============================================================

export type ValueType =
    | "string"
    | "number"
    | "boolean"
    | "array"
    | "object"


// ============================================================
// Semantic Types
// ============================================================

export type SemanticType =
    | "url"
    | "password"
    | "token"
    | "hostname"
    | "duration"


// ============================================================
// Field Kind
// ============================================================

export type FieldKind =
    | "storage"
    | "virtual"


// ============================================================
// Lifecycle
// ============================================================

export type Lifecycle =
    | "stable"
    | "deprecated"
    | "experimental"


// ============================================================
// Target Location
// ============================================================

export type TargetLocation =
    | "root"
    | "jsonData"
    | "secureJsonData"


// ============================================================
// UI Components
// ============================================================

export type UIComponent =
    | "input"
    | "textarea"
    | "select"
    | "multiselect"
    | "radio"
    | "checkbox"
    | "switch"
    | "code"
    | "keyvalue"
    | "list"

export type UIWidth =
    | "full"
    | "half"

/**
 * UI configuration
 */
export interface FieldUI {
    component: UIComponent

    multiline?: boolean
    rows?: number

    options?: FieldOption[]
    allowCustom?: boolean

    width?: UIWidth

    placeholder?: string
}


// ============================================================
// Validation
// ============================================================

export interface FieldValidation {
    pattern?: string
    message?: string

    min?: number
    max?: number

    minItems?: number
    maxItems?: number
}


// ============================================================
// Overrides
// ============================================================

export interface FieldOverride {
    /**
     * CEL condition
     */
    when: Expression

    defaultValue?: unknown
    description?: string
    placeholder?: string
    tooltip?: string

    validation?: FieldValidation
    options?: FieldOption[]
}


// ============================================================
// Storage Mapping
// ============================================================

export type StorageMapping =
    | DirectMapping
    | IndexedPairMapping
    | ComputedMapping

/**
 * Direct mapping
 */
export interface DirectMapping {
    type: "direct"
}

/**
 * Indexed pair mapping (headers)
 */
export interface IndexedPairMapping {
    type: "indexedPair"

    key: MappingField
    value: MappingField

    startIndex?: number
}

/**
 * Computed mapping (virtual fields)
 */
export interface ComputedMapping {
    type: "computed"

    read?: Expression
    write?: Expression
}

/**
 * Mapping field
 */
export interface MappingField {
    target: TargetLocation
    pattern: string
}


// ============================================================
// Options
// ============================================================

export interface FieldOption {
    label: string

    /**
     * Must match parent field valueType
     */
    value: unknown
}


// ============================================================
// Groups
// ============================================================

export interface ConfigGroup {
    id: string
    title: string
    description?: string
    order?: number

    /**
     * References field IDs (not keys)
     */
    fieldRefs: string[]
}


// ============================================================
// Relationships
// ============================================================

export type RelationshipType =
    | "pair"
    | "group"

export interface FieldRelationship {
    type: RelationshipType

    /**
     * References field IDs
     */
    fields: string[]

    description?: string
}


// ============================================================
// Runtime Types (UI Layer)
// ============================================================

export type SecureFieldState =
    | { type: "unset" }
    | { type: "configured" }
    | { type: "updated"; value: string }

export interface FormState {
    values: Record<string, unknown>
    secure: Record<string, SecureFieldState>
}
