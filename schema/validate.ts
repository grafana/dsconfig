import type {
    DatasourceConfigSchema,
    ConfigField,
    FieldValidationRule,
    StorageMapping,
    FieldOverride,
    MappingField,
    FieldUI,
} from "./schema"

import {
    isValueType,
    isSemanticType,
    isFieldKind,
    isLifecycle,
    isTargetLocation,
    isUIComponent,
    isUIWidth,
    isRelationshipType,
    isValidOptionValue,
} from "./guards"

// ============================================================
// Validation Error
// ============================================================

export interface ValidationError {
    /** Dot-separated path to the invalid element */
    path: string
    /** Machine-readable error code */
    code: string
    /** Human-readable description */
    message: string
}

// ============================================================
// Schema Validation
// ============================================================

/**
 * Validates an entire DatasourceConfigSchema.
 *
 * Checks:
 * - root required fields (schemaVersion, pluginType, pluginName, fields)
 * - all field-level rules recursively
 * - global ID uniqueness (including item fields)
 * - group fieldRef existence
 * - relationship fieldRef existence and type validity
 */
export function validateSchema(schema: DatasourceConfigSchema): ValidationError[] {
    const errors: ValidationError[] = []

    if (!schema.schemaVersion) {
        errors.push({ path: "schemaVersion", code: "required", message: "schemaVersion is required" })
    }
    if (!schema.pluginType) {
        errors.push({ path: "pluginType", code: "required", message: "pluginType is required" })
    }
    if (!schema.pluginName) {
        errors.push({ path: "pluginName", code: "required", message: "pluginName is required" })
    }
    if (!schema.fields || schema.fields.length === 0) {
        errors.push({ path: "fields", code: "required", message: "fields is required and must not be empty" })
    }

    // Collect IDs and validate fields
    const fieldIDs = new Set<string>()
    if (schema.fields) {
        for (let i = 0; i < schema.fields.length; i++) {
            const fieldErrors = validateField(schema.fields[i], `fields[${i}]`)
            errors.push(...fieldErrors)
            collectFieldIDs(schema.fields[i], `fields[${i}]`, fieldIDs, errors)
        }
    }

    // Validate group refs
    if (schema.groups) {
        for (let i = 0; i < schema.groups.length; i++) {
            const g = schema.groups[i]
            if (g.fieldRefs) {
                for (let j = 0; j < g.fieldRefs.length; j++) {
                    if (!fieldIDs.has(g.fieldRefs[j])) {
                        errors.push({
                            path: `groups[${i}].fieldRefs[${j}]`,
                            code: "unknown_ref",
                            message: `group "${g.id}" references unknown field id: "${g.fieldRefs[j]}"`,
                        })
                    }
                }
            }
        }
    }

    // Validate relationship refs and type
    if (schema.relationships) {
        for (let i = 0; i < schema.relationships.length; i++) {
            const r = schema.relationships[i]
            if (!isRelationshipType(r.type)) {
                errors.push({
                    path: `relationships[${i}].type`,
                    code: "invalid_enum",
                    message: `invalid relationship type "${r.type}"`,
                })
            }
            if (r.fields) {
                for (let j = 0; j < r.fields.length; j++) {
                    if (!fieldIDs.has(r.fields[j])) {
                        errors.push({
                            path: `relationships[${i}].fields[${j}]`,
                            code: "unknown_ref",
                            message: `relationship references unknown field id: "${r.fields[j]}"`,
                        })
                    }
                }
            }
        }
    }

    return errors
}

/** Recursively collect field IDs, detecting duplicates */
function collectFieldIDs(
    field: ConfigField,
    path: string,
    seen: Set<string>,
    errors: ValidationError[],
): void {
    if (field.id) {
        if (seen.has(field.id)) {
            errors.push({
                path: `${path}.id`,
                code: "duplicate_id",
                message: `duplicate field id: "${field.id}"`,
            })
        }
        seen.add(field.id)
    }
    if (field.item?.fields) {
        for (let i = 0; i < field.item.fields.length; i++) {
            collectFieldIDs(field.item.fields[i], `${path}.item.fields[${i}]`, seen, errors)
        }
    }
}

// ============================================================
// Field Validation
// ============================================================

/**
 * Validates a single ConfigField.
 *
 * Checks identity, valueType, target requirement, kind, semanticType,
 * lifecycle, UI, validation rules, item schema, and storage mapping.
 */
export function validateField(field: ConfigField, path = "field"): ValidationError[] {
    const errors: ValidationError[] = []

    // Identity
    if (!field.id) {
        errors.push({ path: `${path}.id`, code: "required", message: "field id is required" })
    }
    if (!field.key) {
        errors.push({ path: `${path}.key`, code: "required", message: "field key is required" })
    }

    // Value type
    if (!isValueType(field.valueType)) {
        errors.push({
            path: `${path}.valueType`,
            code: "invalid_enum",
            message: `invalid valueType "${field.valueType}"`,
        })
    }

    // Target requirement
    const isVirtual = field.kind === "virtual"
    const isItem = field.isItemField === true

    if (!isVirtual && !isItem && !field.target) {
        errors.push({
            path: `${path}.target`,
            code: "missing_target",
            message: "target is required for storage fields",
        })
    }

    if (field.target && !isTargetLocation(field.target)) {
        errors.push({
            path: `${path}.target`,
            code: "invalid_enum",
            message: `invalid target "${field.target}"`,
        })
    }

    // Kind
    if (field.kind && !isFieldKind(field.kind)) {
        errors.push({
            path: `${path}.kind`,
            code: "invalid_enum",
            message: `invalid kind "${field.kind}"`,
        })
    }

    // Semantic type
    if (field.semanticType && !isSemanticType(field.semanticType)) {
        errors.push({
            path: `${path}.semanticType`,
            code: "invalid_enum",
            message: `invalid semanticType "${field.semanticType}"`,
        })
    }

    // Lifecycle
    if (field.lifecycle && !isLifecycle(field.lifecycle)) {
        errors.push({
            path: `${path}.lifecycle`,
            code: "invalid_enum",
            message: `invalid lifecycle "${field.lifecycle}"`,
        })
    }

    // Array requires item
    if (field.valueType === "array" && !field.item) {
        errors.push({
            path: `${path}.item`,
            code: "required",
            message: "item is required for array fields",
        })
    }

    // Item schema validation
    if (field.item) {
        if (!isValueType(field.item.valueType)) {
            errors.push({
                path: `${path}.item.valueType`,
                code: "invalid_enum",
                message: `invalid item valueType "${field.item.valueType}"`,
            })
        }
        if (field.item.valueType !== "object" && field.item.fields && field.item.fields.length > 0) {
            errors.push({
                path: `${path}.item.fields`,
                code: "invalid_item_fields",
                message: "item fields are only allowed when item valueType is object",
            })
        }
        if (field.item.fields) {
            for (let i = 0; i < field.item.fields.length; i++) {
                const sub = field.item.fields[i]
                if (sub.isItemField !== true) {
                    errors.push({
                        path: `${path}.item.fields[${i}].isItemField`,
                        code: "missing_item_flag",
                        message: `item field "${sub.id}" must have isItemField=true`,
                    })
                }
                errors.push(...validateField(sub, `${path}.item.fields[${i}]`))
            }
        }
    }

    // UI validation
    if (field.ui) {
        errors.push(...validateFieldUI(field.ui, field.valueType, `${path}.ui`))
    }

    // Validation rules
    if (field.validations) {
        for (let i = 0; i < field.validations.length; i++) {
            errors.push(...validateValidationRule(field.validations[i], `${path}.validations[${i}]`))
        }
    }

    // Override validation rules
    if (field.overrides) {
        for (let i = 0; i < field.overrides.length; i++) {
            const ov = field.overrides[i]
            if (ov.validations) {
                for (let j = 0; j < ov.validations.length; j++) {
                    errors.push(
                        ...validateValidationRule(ov.validations[j], `${path}.overrides[${i}].validations[${j}]`),
                    )
                }
            }
        }
    }

    // Storage mapping
    if (field.storage) {
        errors.push(...validateStorageMapping(field.storage, `${path}.storage`))
    }

    return errors
}

// ============================================================
// UI Validation
// ============================================================

function validateFieldUI(ui: FieldUI, valueType: string, path: string): ValidationError[] {
    const errors: ValidationError[] = []

    if (!isUIComponent(ui.component)) {
        errors.push({
            path: `${path}.component`,
            code: "invalid_enum",
            message: `invalid ui component "${ui.component}"`,
        })
    }

    if (ui.width && !isUIWidth(ui.width)) {
        errors.push({
            path: `${path}.width`,
            code: "invalid_enum",
            message: `invalid ui width "${ui.width}"`,
        })
    }

    if (ui.options) {
        for (let i = 0; i < ui.options.length; i++) {
            if (!isValidOptionValue(ui.options[i].value, valueType)) {
                errors.push({
                    path: `${path}.options[${i}].value`,
                    code: "option_type_mismatch",
                    message: `option value type mismatch (expected ${valueType})`,
                })
            }
        }
    }

    return errors
}

// ============================================================
// Validation Rule Validation
// ============================================================

/**
 * Validates a single FieldValidationRule.
 *
 * Checks the type discriminator and required fields for each variant.
 */
export function validateValidationRule(rule: FieldValidationRule, path = "rule"): ValidationError[] {
    const errors: ValidationError[] = []

    switch (rule.type) {
        case "pattern":
            if (!rule.pattern) {
                errors.push({ path: `${path}.pattern`, code: "required", message: "pattern validation requires pattern" })
            }
            break

        case "range":
        case "length":
        case "itemCount":
            if (rule.min == null && rule.max == null) {
                errors.push({
                    path,
                    code: "required",
                    message: `${rule.type} validation requires min or max`,
                })
            }
            break

        case "allowedValues":
            if (!rule.values || rule.values.length === 0) {
                errors.push({
                    path: `${path}.values`,
                    code: "required",
                    message: "allowedValues validation requires values",
                })
            }
            break

        case "custom":
            if (!rule.expression) {
                errors.push({
                    path: `${path}.expression`,
                    code: "required",
                    message: "custom validation requires expression",
                })
            }
            break

        default:
            errors.push({
                path: `${path}.type`,
                code: "invalid_enum",
                message: `unknown validation rule type "${(rule as { type: string }).type}"`,
            })
    }

    return errors
}

// ============================================================
// Storage Mapping Validation
// ============================================================

/**
 * Validates a StorageMapping.
 *
 * Enforces the discriminated union: each type may only use its own fields.
 */
export function validateStorageMapping(mapping: StorageMapping, path = "storage"): ValidationError[] {
    const errors: ValidationError[] = []

    switch (mapping.type) {
        case "direct":
            // Direct mapping must not have any extra fields
            break

        case "indexedPair":
            if (!mapping.key) {
                errors.push({ path: `${path}.key`, code: "required", message: "indexedPair requires key" })
            } else {
                errors.push(...validateMappingField(mapping.key, `${path}.key`))
            }
            if (!mapping.value) {
                errors.push({ path: `${path}.value`, code: "required", message: "indexedPair requires value" })
            } else {
                errors.push(...validateMappingField(mapping.value, `${path}.value`))
            }
            break

        case "computed":
            if (!mapping.read && !mapping.write) {
                errors.push({
                    path,
                    code: "required",
                    message: "computed mapping requires read or write",
                })
            }
            break

        default:
            errors.push({
                path: `${path}.type`,
                code: "invalid_enum",
                message: `unknown mapping type "${(mapping as { type: string }).type}"`,
            })
    }

    return errors
}

function validateMappingField(field: MappingField, path: string): ValidationError[] {
    const errors: ValidationError[] = []
    if (!isTargetLocation(field.target)) {
        errors.push({
            path: `${path}.target`,
            code: "invalid_enum",
            message: `invalid target "${field.target}"`,
        })
    }
    if (!field.pattern) {
        errors.push({ path: `${path}.pattern`, code: "required", message: "pattern is required" })
    }
    return errors
}
