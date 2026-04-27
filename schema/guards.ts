import type {
    ValueType,
    SemanticType,
    FieldKind,
    Lifecycle,
    TargetLocation,
    UIComponent,
    UIWidth,
    RelationshipType,
    StorageMapping,
    FieldValidationRule,
} from "./schema"

// ============================================================
// Enum Guards
// ============================================================

const VALUE_TYPES: ReadonlySet<string> = new Set(["string", "number", "boolean", "array", "object"])

/** Checks if a value is a valid ValueType */
export function isValueType(value: unknown): value is ValueType {
    return typeof value === "string" && VALUE_TYPES.has(value)
}

const SEMANTIC_TYPES: ReadonlySet<string> = new Set(["url", "password", "token", "hostname", "duration"])

/** Checks if a value is a valid SemanticType */
export function isSemanticType(value: unknown): value is SemanticType {
    return typeof value === "string" && SEMANTIC_TYPES.has(value)
}

const FIELD_KINDS: ReadonlySet<string> = new Set(["storage", "virtual"])

/** Checks if a value is a valid FieldKind */
export function isFieldKind(value: unknown): value is FieldKind {
    return typeof value === "string" && FIELD_KINDS.has(value)
}

const LIFECYCLES: ReadonlySet<string> = new Set(["stable", "deprecated", "experimental"])

/** Checks if a value is a valid Lifecycle */
export function isLifecycle(value: unknown): value is Lifecycle {
    return typeof value === "string" && LIFECYCLES.has(value)
}

const TARGET_LOCATIONS: ReadonlySet<string> = new Set(["root", "jsonData", "secureJsonData"])

/** Checks if a value is a valid TargetLocation */
export function isTargetLocation(value: unknown): value is TargetLocation {
    return typeof value === "string" && TARGET_LOCATIONS.has(value)
}

const UI_COMPONENTS: ReadonlySet<string> = new Set([
    "input", "textarea", "select", "multiselect", "radio",
    "checkbox", "switch", "code", "keyvalue", "list",
])

/** Checks if a value is a valid UIComponent */
export function isUIComponent(value: unknown): value is UIComponent {
    return typeof value === "string" && UI_COMPONENTS.has(value)
}

const UI_WIDTHS: ReadonlySet<string> = new Set(["full", "half"])

/** Checks if a value is a valid UIWidth */
export function isUIWidth(value: unknown): value is UIWidth {
    return typeof value === "string" && UI_WIDTHS.has(value)
}

const RELATIONSHIP_TYPES: ReadonlySet<string> = new Set(["pair", "group"])

/** Checks if a value is a valid RelationshipType */
export function isRelationshipType(value: unknown): value is RelationshipType {
    return typeof value === "string" && RELATIONSHIP_TYPES.has(value)
}

// ============================================================
// Discriminated Union Guards
// ============================================================

const STORAGE_MAPPING_TYPES: ReadonlySet<string> = new Set(["direct", "indexedPair", "computed"])

/** Checks if a value is a valid StorageMapping (has a recognized type discriminator) */
export function isStorageMapping(value: unknown): value is StorageMapping {
    return (
        typeof value === "object" &&
        value !== null &&
        "type" in value &&
        typeof (value as { type: unknown }).type === "string" &&
        STORAGE_MAPPING_TYPES.has((value as { type: string }).type)
    )
}

const VALIDATION_RULE_TYPES: ReadonlySet<string> = new Set([
    "pattern", "range", "length", "itemCount", "allowedValues", "custom",
])

/** Checks if a value is a valid FieldValidationRule (has a recognized type discriminator) */
export function isValidationRule(value: unknown): value is FieldValidationRule {
    return (
        typeof value === "object" &&
        value !== null &&
        "type" in value &&
        typeof (value as { type: unknown }).type === "string" &&
        VALIDATION_RULE_TYPES.has((value as { type: string }).type)
    )
}

// ============================================================
// Option Value Type Checking
// ============================================================

/**
 * Checks that an option value is non-null and compatible with the
 * given valueType string.
 *
 * Matches Go's ValidateOptionValue semantics:
 * - nil/null → false (JSON Schema requires value)
 * - string fields → must be string
 * - number fields → must be number
 * - boolean fields → must be boolean
 * - array/object → always true (not type-checked)
 */
export function isValidOptionValue(value: unknown, valueType: string): boolean {
    if (value == null) {
        return false
    }
    switch (valueType) {
        case "string":
            return typeof value === "string"
        case "number":
            return typeof value === "number"
        case "boolean":
            return typeof value === "boolean"
        default:
            return true
    }
}
