import type {
    DatasourceConfigSchema,
    ConfigField,
    ValueType,
    SemanticType,
    FieldEffect,
} from "./schema"

// ============================================================
// Prompt Schema Types
// ============================================================

export interface PromptSchema {
    pluginType: string
    pluginName: string
    fields: PromptField[]
}

export interface PromptField {
    id: string
    /** Storage path, e.g. "jsonData.httpMethod" */
    path: string
    type: ValueType
    /** Human-readable field name */
    label?: string
    semanticType?: SemanticType
    description?: string
    required?: boolean
    requiredWhen?: string
    dependsOn?: string
    defaultValue?: unknown
    allowedValues?: unknown[]
    pattern?: string
    range?: { min?: number; max?: number }
    /** Nested fields for array/map items */
    items?: PromptField[]
    /** Virtual selector options with side-effects */
    options?: PromptOption[]
}

export interface PromptOption {
    value: unknown
    label: string
    sets: Record<string, unknown>
}

// ============================================================
// Projection
// ============================================================

/**
 * Projects a full DatasourceConfigSchema into a compact, LLM-friendly
 * PromptSchema by stripping UI hints, groups, storage mappings, and
 * other rendering/internal concerns.
 *
 * Keeps: identity, types, constraints, defaults, effects, descriptions.
 * Strips: ui, groups, overrides, lifecycle, tags, storage, isItemField.
 */
export function toPromptSchema(schema: DatasourceConfigSchema): PromptSchema {
    return {
        pluginType: schema.pluginType,
        pluginName: schema.pluginName,
        fields: schema.fields
            .filter((f) => !isManagedField(f))
            .map(projectField),
    }
}

/**
 * Serializes a DatasourceConfigSchema into a compact JSON string
 * suitable for embedding in an LLM prompt (tool/function calling).
 */
export function toPromptString(schema: DatasourceConfigSchema): string {
    return JSON.stringify(toPromptSchema(schema), null, 2)
}

/**
 * Renders a DatasourceConfigSchema into human-readable text suitable
 * for embedding in an LLM system/user prompt. More token-efficient
 * and easier for LLMs to reason about than JSON.
 *
 * Output format:
 * ```
 * Prometheus (pluginType: prometheus)
 *
 * Fields:
 * - url (root.url) [string, url] REQUIRED — Base URL
 * - httpMethod (jsonData.httpMethod) [string] default: "POST"
 *   Allowed: "GET", "POST"
 * ```
 */
export function toPromptText(schema: DatasourceConfigSchema): string {
    const ps = toPromptSchema(schema)
    const lines: string[] = []

    lines.push(`${ps.pluginName} (pluginType: ${ps.pluginType})`)
    lines.push("")
    lines.push("Fields:")

    for (const f of ps.fields) {
        lines.push(renderField(f, ""))
    }

    return lines.join("\n")
}

function renderField(f: PromptField, indent: string): string {
    const lines: string[] = []

    // Main line: - id (path) [type, semanticType] REQUIRED default: X — description
    const parts: string[] = []

    const name = f.label ?? f.id
    parts.push(`${indent}- ${name}`)

    // path (skip for virtual fields that have key-only path)
    if (f.path.includes(".")) {
        parts.push(`(${f.path})`)
    }

    // type annotation
    const typeAnnotations: string[] = [f.type]
    if (f.semanticType) typeAnnotations.push(f.semanticType)
    parts.push(`[${typeAnnotations.join(", ")}]`)

    if (f.required) parts.push("REQUIRED")
    if (f.defaultValue !== undefined) parts.push(`default: ${formatValue(f.defaultValue)}`)
    if (f.description) parts.push(`— ${f.description}`)

    lines.push(parts.join(" "))

    // Constraint lines (indented)
    const sub = indent + "  "

    if (f.requiredWhen) {
        lines.push(`${sub}Required when: ${f.requiredWhen}`)
    }
    if (f.dependsOn) {
        lines.push(`${sub}Visible when: ${f.dependsOn}`)
    }
    if (f.allowedValues && !f.options) {
        lines.push(`${sub}Allowed: ${f.allowedValues.map(formatValue).join(", ")}`)
    }
    if (f.pattern) {
        lines.push(`${sub}Pattern: ${f.pattern}`)
    }
    if (f.range) {
        const bounds: string[] = []
        if (f.range.min != null) bounds.push(`min: ${f.range.min}`)
        if (f.range.max != null) bounds.push(`max: ${f.range.max}`)
        lines.push(`${sub}Range: ${bounds.join(", ")}`)
    }

    // Options (virtual selector with effects)
    if (f.options) {
        lines.push(`${sub}Options:`)
        for (const opt of f.options) {
            const sets = Object.entries(opt.sets)
                .map(([k, v]) => `${k}=${formatValue(v)}`)
                .join(", ")
            lines.push(`${sub}  ${formatValue(opt.value)} (${opt.label}) → sets ${sets}`)
        }
    }

    // Array/map item fields
    if (f.items && f.items.length > 0) {
        lines.push(`${sub}Item fields:`)
        for (const item of f.items) {
            lines.push(renderField(item, sub + "  "))
        }
    }

    return lines.join("\n")
}

function formatValue(v: unknown): string {
    if (typeof v === "string") return `"${v}"`
    if (typeof v === "boolean" || typeof v === "number") return String(v)
    if (v === null || v === undefined) return "null"
    return JSON.stringify(v)
}

// ============================================================
// Internals
// ============================================================

/** Fields tagged "managed-by:*" are driven by effects and hidden from the LLM. */
function isManagedField(f: ConfigField): boolean {
    return f.tags?.some((t) => t.startsWith("managed-by:")) === true
}

function projectField(f: ConfigField): PromptField {
    const pf: PromptField = {
        id: f.id,
        path: fieldPath(f),
        type: f.valueType,
    }

    if (f.label) pf.label = f.label
    if (f.semanticType) pf.semanticType = f.semanticType
    if (f.description) pf.description = f.description
    if (f.required) pf.required = true
    if (f.requiredWhen) pf.requiredWhen = f.requiredWhen
    if (f.dependsOn) pf.dependsOn = f.dependsOn
    if (f.defaultValue !== undefined) pf.defaultValue = f.defaultValue

    // Flatten validations into top-level constraint fields
    if (f.validations) {
        for (const v of f.validations) {
            switch (v.type) {
                case "allowedValues":
                    pf.allowedValues = v.values
                    break
                case "pattern":
                    pf.pattern = v.pattern
                    break
                case "range":
                    pf.range = {}
                    if (v.min != null) pf.range.min = v.min
                    if (v.max != null) pf.range.max = v.max
                    break
            }
        }
    }

    // Flatten effects + ui.options into prompt options for virtual selectors
    if (f.effects && f.effects.length > 0) {
        const labelMap = buildLabelMap(f)
        pf.options = f.effects.map((eff) => {
            const val = extractLiteralFromWhen(eff.when)
            return {
                value: val ?? eff.when,
                label: labelMap.get(String(val)) ?? String(val),
                sets: eff.set,
            }
        })
    }

    // Recurse into array/map item fields
    if (f.item) {
        if (f.item.fields && f.item.fields.length > 0) {
            pf.items = f.item.fields.map(projectField)
        }
    }

    return pf
}

function fieldPath(f: ConfigField): string {
    if (!f.target) return f.key
    if (f.section) return `${f.target}.${f.section}.${f.key}`
    return `${f.target}.${f.key}`
}

/**
 * Builds a map from option value → label using ui.options.
 * Falls back to empty map if no UI options defined.
 */
function buildLabelMap(f: ConfigField): Map<string, string> {
    const m = new Map<string, string>()
    if (f.ui?.options) {
        for (const opt of f.ui.options) {
            m.set(String(opt.value), opt.label)
        }
    }
    return m
}

/**
 * Extracts the literal value from a simple `value == '...'` expression.
 * Returns undefined for complex expressions.
 */
export function extractLiteralFromWhen(when: string): unknown {
    // Match: value == 'string-literal'
    const singleQuote = /^value\s*==\s*'([^']*)'$/.exec(when)
    if (singleQuote) return singleQuote[1]

    // Match: value == "string-literal"
    const doubleQuote = /^value\s*==\s*"([^"]*)"$/.exec(when)
    if (doubleQuote) return doubleQuote[1]

    // Match: value == true / value == false
    const boolMatch = /^value\s*==\s*(true|false)$/.exec(when)
    if (boolMatch) return boolMatch[1] === "true"

    // Match: value == 123 (integer)
    const numMatch = /^value\s*==\s*(-?\d+(?:\.\d+)?)$/.exec(when)
    if (numMatch) return Number(numMatch[1])

    return undefined
}
