import type {
    DatasourceConfigSchema,
    ConfigField,
    FieldValidationRule,
    TargetLocation,
    ValueType,
} from "./schema"

// ============================================================
// Load Mode
// ============================================================

export type LoadMode = "read" | "write"

// ============================================================
// Input: Grafana datasource config
// ============================================================

export interface DatasourceConfig {
    root: Record<string, unknown>
    jsonData?: Record<string, unknown>
    secureJsonData?: Record<string, unknown>
    secureJsonFields?: Record<string, boolean>
}

/** Parse a flat Grafana datasource resource into DatasourceConfig. */
export function newDatasourceConfig(raw: Record<string, unknown>): DatasourceConfig {
    const dc: DatasourceConfig = { root: {} }

    for (const [k, v] of Object.entries(raw)) {
        switch (k) {
            case "jsonData":
                if (v && typeof v === "object" && !Array.isArray(v)) {
                    dc.jsonData = v as Record<string, unknown>
                }
                break
            case "secureJsonData":
                if (v && typeof v === "object" && !Array.isArray(v)) {
                    dc.secureJsonData = v as Record<string, unknown>
                }
                break
            case "secureJsonFields":
                if (v && typeof v === "object" && !Array.isArray(v)) {
                    dc.secureJsonFields = v as Record<string, boolean>
                }
                break
            default:
                dc.root[k] = v
        }
    }

    return dc
}

// ============================================================
// Output
// ============================================================

export interface ConfigError {
    fieldId: string
    path: string
    code: string
    message: string
}

export type SecureState = "unset" | "configured" | "updated"

export type ValueSource = "config" | "default" | "none"

export interface FieldValue {
    value: unknown
    source: ValueSource
}

export interface LoadResult {
    errors: ConfigError[]
    values: Record<string, FieldValue>
    secureFields: Record<string, SecureState>
}

// ============================================================
// Pipeline
// ============================================================

export function loadAndValidate(
    schema: DatasourceConfigSchema,
    config: DatasourceConfig,
    mode: LoadMode,
): LoadResult {
    const result: LoadResult = {
        errors: [],
        values: {},
        secureFields: {},
    }

    for (const field of schema.fields) {
        loadField(field, config, mode, result)
    }

    return result
}

function loadField(
    f: ConfigField,
    config: DatasourceConfig,
    mode: LoadMode,
    result: LoadResult,
): void {
    if (f.kind === "virtual") {
        return
    }

    if (f.target === "secureJsonData") {
        loadSecureField(f, config, mode, result)
        return
    }

    let raw = extractValue(config, f)

    if (f.storage?.type === "indexedPair") {
        raw = expandIndexedPair(config, f)
    }

    let source: ValueSource = "config"
    if (raw == null) {
        if (f.defaultValue !== undefined) {
            raw = f.defaultValue
            source = "default"
        } else {
            source = "none"
        }
    }

    result.values[f.id] = { value: raw ?? null, source }

    validateFieldValue(f, raw ?? null, result)

    if (f.valueType === "array" && f.item && Array.isArray(raw)) {
        validateArrayItems(f, raw, result)
    }
}

function loadSecureField(
    f: ConfigField,
    config: DatasourceConfig,
    mode: LoadMode,
    result: LoadResult,
): void {
    const path = fieldPath(f)

    if (mode === "write") {
        const val = config.secureJsonData?.[f.key]
        if (val != null) {
            result.secureFields[f.id] = "updated"
            result.values[f.id] = { value: val, source: "config" }
        } else {
            result.secureFields[f.id] = "unset"
            result.values[f.id] = { value: null, source: "none" }
        }
        if (f.required && val == null) {
            result.errors.push({
                fieldId: f.id,
                path,
                code: "required",
                message: `field ${f.id} is required`,
            })
        }
    } else {
        if (config.secureJsonFields?.[f.key]) {
            result.secureFields[f.id] = "configured"
        } else {
            result.secureFields[f.id] = "unset"
            if (f.required) {
                result.errors.push({
                    fieldId: f.id,
                    path,
                    code: "required",
                    message: `field ${f.id} is required`,
                })
            }
        }
        result.values[f.id] = { value: null, source: "none" }
    }
}

// ============================================================
// Value extraction
// ============================================================

function extractValue(config: DatasourceConfig, f: ConfigField): unknown {
    if (!f.target) return undefined

    switch (f.target) {
        case "root":
            return config.root[f.key]
        case "jsonData": {
            if (!config.jsonData) return undefined
            if (!f.section) return config.jsonData[f.key]
            return extractFromSection(config.jsonData, f.section, f.key)
        }
        case "secureJsonData":
            return undefined // handled by loadSecureField
        default:
            return undefined
    }
}

function extractFromSection(
    data: Record<string, unknown>,
    section: string,
    key: string,
): unknown {
    const segments = section.split(".")
    let current: Record<string, unknown> = data

    for (const seg of segments) {
        const v = current[seg]
        if (v == null || typeof v !== "object" || Array.isArray(v)) {
            return undefined
        }
        current = v as Record<string, unknown>
    }

    return current[key]
}

// ============================================================
// IndexedPair expansion
// ============================================================

function expandIndexedPair(config: DatasourceConfig, f: ConfigField): unknown[] | undefined {
    const storage = f.storage
    if (!storage || storage.type !== "indexedPair") return undefined
    if (!("key" in storage) || !("value" in storage)) return undefined
    if (!f.item?.fields || f.item.fields.length < 2) return undefined

    const startIndex = ("startIndex" in storage && storage.startIndex != null)
        ? storage.startIndex
        : 1

    const keyPattern = storage.key.pattern
    const valuePattern = storage.value.pattern
    const keyFieldKey = f.item.fields[0].key
    const valueFieldKey = f.item.fields[1].key

    const items: unknown[] = []

    for (let i = startIndex; ; i++) {
        const idx = String(i)
        const kName = keyPattern.replace("{index}", idx)
        const vName = valuePattern.replace("{index}", idx)

        const kVal = getFromTarget(config, storage.key.target, kName)
        if (kVal == null) break

        const vVal = getFromTarget(config, storage.value.target, vName)

        items.push({
            [keyFieldKey]: kVal,
            [valueFieldKey]: vVal,
        })
    }

    return items.length > 0 ? items : undefined
}

function getFromTarget(
    config: DatasourceConfig,
    target: TargetLocation,
    key: string,
): unknown {
    switch (target) {
        case "root":
            return config.root[key]
        case "jsonData":
            return config.jsonData?.[key]
        case "secureJsonData":
            return config.secureJsonData?.[key]
        default:
            return undefined
    }
}

// ============================================================
// Value validation
// ============================================================

function validateFieldValue(
    f: ConfigField,
    value: unknown,
    result: LoadResult,
): void {
    const path = fieldPath(f)

    if (f.required && value == null) {
        result.errors.push({
            fieldId: f.id,
            path,
            code: "required",
            message: `field ${f.id} is required`,
        })
        return
    }

    if (value == null) return

    const typeErr = checkValueType(value, f.valueType)
    if (typeErr) {
        result.errors.push({
            fieldId: f.id,
            path,
            code: "type_mismatch",
            message: `field ${f.id}: ${typeErr}`,
        })
        return
    }

    if (f.validations) {
        for (const rule of f.validations) {
            const err = evaluateRule(value, rule)
            if (err) {
                result.errors.push({
                    fieldId: f.id,
                    path,
                    code: rule.type,
                    message: rule.message ?? err,
                })
            }
        }
    }
}

function checkValueType(value: unknown, expected: ValueType): string | null {
    switch (expected) {
        case "string":
            return typeof value === "string" ? null : `expected string, got ${typeof value}`
        case "number":
            return typeof value === "number" ? null : `expected number, got ${typeof value}`
        case "boolean":
            return typeof value === "boolean" ? null : `expected boolean, got ${typeof value}`
        case "array":
            return Array.isArray(value) ? null : `expected array, got ${typeof value}`
        case "object":
            return (value !== null && typeof value === "object" && !Array.isArray(value))
                ? null
                : `expected object, got ${typeof value}`
        default:
            return null
    }
}

function evaluateRule(value: unknown, rule: FieldValidationRule): string | null {
    switch (rule.type) {
        case "pattern": {
            if (typeof value !== "string") return null
            const re = new RegExp(rule.pattern)
            return re.test(value) ? null : `value does not match pattern "${rule.pattern}"`
        }

        case "range": {
            if (typeof value !== "number") return null
            if (rule.min != null && value < rule.min) {
                return `value ${value} is below minimum ${rule.min}`
            }
            if (rule.max != null && value > rule.max) {
                return `value ${value} exceeds maximum ${rule.max}`
            }
            return null
        }

        case "length": {
            if (typeof value !== "string") return null
            if (rule.min != null && value.length < rule.min) {
                return `length ${value.length} is below minimum ${rule.min}`
            }
            if (rule.max != null && value.length > rule.max) {
                return `length ${value.length} exceeds maximum ${rule.max}`
            }
            return null
        }

        case "itemCount": {
            if (!Array.isArray(value)) return null
            if (rule.min != null && value.length < rule.min) {
                return `item count ${value.length} is below minimum ${rule.min}`
            }
            if (rule.max != null && value.length > rule.max) {
                return `item count ${value.length} exceeds maximum ${rule.max}`
            }
            return null
        }

        case "allowedValues": {
            const vals = rule.values ?? []
            for (const allowed of vals) {
                if (String(value) === String(allowed)) return null
            }
            return `value ${value} is not in allowed values [${vals.join(", ")}]`
        }

        case "custom":
            return null // CEL not evaluated
    }

    return null
}

function validateArrayItems(
    f: ConfigField,
    arr: unknown[],
    result: LoadResult,
): void {
    if (!f.item) return

    for (let i = 0; i < arr.length; i++) {
        const elem = arr[i]
        if (f.item.valueType === "object" && f.item.fields && f.item.fields.length > 0) {
            if (typeof elem !== "object" || elem === null || Array.isArray(elem)) {
                result.errors.push({
                    fieldId: f.id,
                    path: `${fieldPath(f)}[${i}]`,
                    code: "type_mismatch",
                    message: `field ${f.id}[${i}]: expected object, got ${typeof elem}`,
                })
                continue
            }
            const obj = elem as Record<string, unknown>
            for (const sub of f.item.fields) {
                let val = obj[sub.key]
                if (val == null && sub.defaultValue !== undefined) {
                    val = sub.defaultValue
                }
                validateFieldValue(sub, val ?? null, result)
            }
        }
    }
}

// ============================================================
// Helpers
// ============================================================

function fieldPath(f: ConfigField): string {
    if (!f.target) return f.key
    if (f.section) return `${f.target}.${f.section}.${f.key}`
    return `${f.target}.${f.key}`
}
