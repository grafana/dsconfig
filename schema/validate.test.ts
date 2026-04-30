import { describe, it, expect } from "vitest"
import type {
    DatasourceConfigSchema,
    ConfigField,
    FieldValidationRule,
    StorageMapping,
    TargetLocation,
} from "./schema"
import {
    validateSchema,
    validateField,
    validateValidationRule,
    validateStorageMapping,
} from "./validate"

// ============================================================
// Helpers
// ============================================================

function validStorageField(id: string, key: string): ConfigField {
    return { id, key, valueType: "string", target: "jsonData" }
}

function minimalSchema(...fields: ConfigField[]): DatasourceConfigSchema {
    return {
        schemaVersion: "v1",
        pluginType: "test",
        pluginName: "Test",
        fields,
    }
}

function errorCodes(errors: { code: string }[]): string[] {
    return errors.map((e) => e.code)
}

function errorAt(errors: { path: string; code: string }[], path: string): { path: string; code: string } | undefined {
    return errors.find((e) => e.path === path)
}

// ============================================================
// Root-level validation
// ============================================================

describe("validateSchema — root required fields", () => {
    it("passes for a minimal valid schema", () => {
        const errors = validateSchema(minimalSchema(validStorageField("url", "url")))
        expect(errors).toHaveLength(0)
    })

    it("rejects missing schemaVersion", () => {
        const s = minimalSchema(validStorageField("x", "x"))
        s.schemaVersion = ""
        const errors = validateSchema(s)
        expect(errorAt(errors, "schemaVersion")?.code).toBe("required")
    })

    it("rejects missing pluginType", () => {
        const s = minimalSchema(validStorageField("x", "x"))
        s.pluginType = ""
        const errors = validateSchema(s)
        expect(errorAt(errors, "pluginType")?.code).toBe("required")
    })

    it("rejects missing pluginName", () => {
        const s = minimalSchema(validStorageField("x", "x"))
        s.pluginName = ""
        const errors = validateSchema(s)
        expect(errorAt(errors, "pluginName")?.code).toBe("required")
    })

    it("rejects empty fields array", () => {
        const s = minimalSchema()
        const errors = validateSchema(s)
        expect(errorAt(errors, "fields")?.code).toBe("required")
    })
})

// ============================================================
// Field ID validation
// ============================================================

describe("validateSchema — field IDs", () => {
    it("detects duplicate top-level field IDs", () => {
        const s = minimalSchema(
            validStorageField("dup", "a"),
            validStorageField("dup", "b"),
        )
        const errors = validateSchema(s)
        expect(errors.some((e) => e.code === "duplicate_id")).toBe(true)
    })

    it("detects duplicate IDs between top-level and item fields", () => {
        const s = minimalSchema(
            validStorageField("conflict", "x"),
            {
                id: "arr", key: "arr", valueType: "array", target: "jsonData",
                item: {
                    valueType: "object",
                    fields: [{ id: "conflict", key: "k", valueType: "string", isItemField: true }],
                },
            },
        )
        const errors = validateSchema(s)
        expect(errors.some((e) => e.code === "duplicate_id")).toBe(true)
    })

    it("collects item field IDs for group ref resolution", () => {
        const s: DatasourceConfigSchema = {
            schemaVersion: "v1", pluginType: "test", pluginName: "Test",
            fields: [{
                id: "arr", key: "arr", valueType: "array", target: "jsonData",
                item: {
                    valueType: "object",
                    fields: [{ id: "arr.item.name", key: "name", valueType: "string", isItemField: true }],
                },
            }],
            groups: [{ id: "g1", title: "G", fieldRefs: ["arr.item.name"] }],
        }
        expect(validateSchema(s)).toHaveLength(0)
    })
})

// ============================================================
// Group and relationship ref validation
// ============================================================

describe("validateSchema — refs", () => {
    it("accepts valid group refs", () => {
        const s = minimalSchema(validStorageField("a", "a"), validStorageField("b", "b"))
        s.groups = [{ id: "g1", title: "G", fieldRefs: ["a", "b"] }]
        expect(validateSchema(s)).toHaveLength(0)
    })

    it("rejects unknown group ref", () => {
        const s = minimalSchema(validStorageField("a", "a"))
        s.groups = [{ id: "g1", title: "G", fieldRefs: ["missing"] }]
        const errors = validateSchema(s)
        expect(errorAt(errors, "groups[0].fieldRefs[0]")?.code).toBe("unknown_ref")
    })

    it("accepts valid relationship refs", () => {
        const s = minimalSchema(validStorageField("u", "u"), validStorageField("p", "p"))
        s.relationships = [{ type: "pair", fields: ["u", "p"] }]
        expect(validateSchema(s)).toHaveLength(0)
    })

    it("rejects unknown relationship ref", () => {
        const s = minimalSchema(validStorageField("a", "a"))
        s.relationships = [{ type: "pair", fields: ["a", "ghost"] }]
        const errors = validateSchema(s)
        expect(errors.some((e) => e.code === "unknown_ref")).toBe(true)
    })

    it("rejects invalid relationship type", () => {
        const s = minimalSchema(validStorageField("a", "a"))
        s.relationships = [{ type: "dependency" as any, fields: ["a"] }]
        const errors = validateSchema(s)
        expect(errors.some((e) => e.code === "invalid_enum")).toBe(true)
    })
})

// ============================================================
// ConfigField.Validate — identity
// ============================================================

describe("validateField — identity", () => {
    it("rejects empty id", () => {
        const errors = validateField({ id: "", key: "x", valueType: "string", target: "jsonData" })
        expect(errors.some((e) => e.code === "required" && e.path.endsWith(".id"))).toBe(true)
    })

    it("rejects empty key", () => {
        const errors = validateField({ id: "x", key: "", valueType: "string", target: "jsonData" })
        expect(errors.some((e) => e.code === "required" && e.path.endsWith(".key"))).toBe(true)
    })
})

// ============================================================
// ConfigField.Validate — valueType
// ============================================================

describe("validateField — valueType", () => {
    it("rejects invalid valueType", () => {
        const errors = validateField({ id: "x", key: "x", valueType: "blob" as any, target: "jsonData" })
        expect(errors.some((e) => e.code === "invalid_enum")).toBe(true)
    })

    it("accepts all valid valueTypes", () => {
        for (const vt of ["string", "number", "boolean", "array", "object"] as const) {
            const f: ConfigField = { id: "x", key: "x", valueType: vt, target: "jsonData" }
            if (vt === "array") f.item = { valueType: "string" }
            const errors = validateField(f)
            expect(errors.filter((e) => e.path.endsWith(".valueType"))).toHaveLength(0)
        }
    })
})

// ============================================================
// ConfigField.Validate — target requirement
// ============================================================

describe("validateField — target", () => {
    it("requires target for storage fields", () => {
        const errors = validateField({ id: "x", key: "x", valueType: "string" })
        expect(errors.some((e) => e.code === "missing_target")).toBe(true)
    })

    it("allows virtual fields without target", () => {
        const errors = validateField({ id: "x", key: "x", valueType: "string", kind: "virtual" })
        expect(errors.filter((e) => e.code === "missing_target")).toHaveLength(0)
    })

    it("allows item fields without target", () => {
        const errors = validateField({ id: "x", key: "x", valueType: "string", isItemField: true })
        expect(errors.filter((e) => e.code === "missing_target")).toHaveLength(0)
    })

    it("rejects section on item fields", () => {
        const errors = validateField({ id: "x", key: "x", valueType: "string", isItemField: true, section: "nested" })
        expect(errors.some((e) => e.code === "invalid_section")).toBe(true)
    })

    it("rejects section on virtual fields", () => {
        const errors = validateField({ id: "x", key: "x", valueType: "string", kind: "virtual", section: "nested" })
        expect(errors.some((e) => e.code === "invalid_section")).toBe(true)
    })

    it("allows section on storage fields", () => {
        const errors = validateField({ id: "x", key: "x", valueType: "string", target: "jsonData", section: "oauth2.endpoints" })
        expect(errors.filter((e) => e.code === "invalid_section")).toHaveLength(0)
    })

    it("rejects invalid target", () => {
        const errors = validateField({ id: "x", key: "x", valueType: "string", target: "bad" as any })
        expect(errors.some((e) => e.code === "invalid_enum" && e.path.endsWith(".target"))).toBe(true)
    })

    it("accepts all valid targets", () => {
        for (const t of ["root", "jsonData", "secureJsonData"] as const) {
            const errors = validateField({ id: "x", key: "x", valueType: "string", target: t })
            expect(errors.filter((e) => e.path.endsWith(".target"))).toHaveLength(0)
        }
    })
})

// ============================================================
// ConfigField.Validate — kind
// ============================================================

describe("validateField — kind", () => {
    it("allows empty kind (defaults to storage)", () => {
        const errors = validateField(validStorageField("x", "x"))
        expect(errors.filter((e) => e.path.endsWith(".kind"))).toHaveLength(0)
    })

    it("allows storage kind", () => {
        const f = validStorageField("x", "x")
        f.kind = "storage"
        expect(validateField(f).filter((e) => e.path.endsWith(".kind"))).toHaveLength(0)
    })

    it("allows virtual kind", () => {
        const errors = validateField({ id: "x", key: "x", valueType: "string", kind: "virtual" })
        expect(errors.filter((e) => e.path.endsWith(".kind"))).toHaveLength(0)
    })

    it("rejects invalid kind", () => {
        const errors = validateField({ id: "x", key: "x", valueType: "string", kind: "unknown" as any, target: "jsonData" })
        expect(errors.some((e) => e.code === "invalid_enum" && e.path.endsWith(".kind"))).toBe(true)
    })
})

// ============================================================
// ConfigField.Validate — enum fields
// ============================================================

describe("validateField — semanticType", () => {
    it("accepts valid semanticType", () => {
        const f = validStorageField("x", "x")
        f.semanticType = "password"
        expect(validateField(f).filter((e) => e.path.endsWith(".semanticType"))).toHaveLength(0)
    })

    it("rejects invalid semanticType", () => {
        const f = validStorageField("x", "x")
        f.semanticType = "email" as any
        expect(validateField(f).some((e) => e.code === "invalid_enum")).toBe(true)
    })
})

describe("validateField — lifecycle", () => {
    it("accepts valid lifecycle", () => {
        const f = validStorageField("x", "x")
        f.lifecycle = "deprecated"
        expect(validateField(f).filter((e) => e.path.endsWith(".lifecycle"))).toHaveLength(0)
    })

    it("rejects invalid lifecycle", () => {
        const f = validStorageField("x", "x")
        f.lifecycle = "beta" as any
        expect(validateField(f).some((e) => e.code === "invalid_enum")).toBe(true)
    })
})

// ============================================================
// ConfigField.Validate — array / item
// ============================================================

describe("validateField — item schema", () => {
    it("requires item for array fields", () => {
        const errors = validateField({ id: "x", key: "x", valueType: "array", target: "jsonData" })
        expect(errors.some((e) => e.code === "required" && e.path.endsWith(".item"))).toBe(true)
    })

    it("rejects invalid item valueType", () => {
        const errors = validateField({
            id: "x", key: "x", valueType: "array", target: "jsonData",
            item: { valueType: "invalid" as any },
        })
        expect(errors.some((e) => e.path.includes("item.valueType"))).toBe(true)
    })

    it("rejects non-object item with fields", () => {
        const errors = validateField({
            id: "x", key: "x", valueType: "array", target: "jsonData",
            item: {
                valueType: "string",
                fields: [{ id: "sub", key: "sub", valueType: "string", isItemField: true }],
            },
        })
        expect(errors.some((e) => e.code === "invalid_item_fields")).toBe(true)
    })

    it("rejects item field without isItemField=true", () => {
        const errors = validateField({
            id: "x", key: "x", valueType: "array", target: "jsonData",
            item: {
                valueType: "object",
                fields: [{ id: "sub", key: "sub", valueType: "string" }],
            },
        })
        expect(errors.some((e) => e.code === "missing_item_flag")).toBe(true)
    })

    it("propagates item field validation errors", () => {
        const errors = validateField({
            id: "x", key: "x", valueType: "array", target: "jsonData",
            item: {
                valueType: "object",
                fields: [{ id: "sub", key: "", valueType: "string", isItemField: true }],
            },
        })
        expect(errors.some((e) => e.code === "required" && e.path.includes("item.fields[0].key"))).toBe(true)
    })

    it("accepts valid object item with fields", () => {
        const errors = validateField({
            id: "headers", key: "headers", valueType: "array", target: "jsonData",
            item: {
                valueType: "object",
                fields: [
                    { id: "headers.k", key: "key", valueType: "string", isItemField: true },
                    { id: "headers.v", key: "val", valueType: "string", isItemField: true },
                ],
            },
        })
        expect(errors).toHaveLength(0)
    })
})

// ============================================================
// ConfigField.Validate — UI
// ============================================================

describe("validateField — UI", () => {
    it("rejects invalid ui component", () => {
        const f = validStorageField("x", "x")
        f.ui = { component: "datepicker" as any }
        expect(validateField(f).some((e) => e.path.endsWith(".component"))).toBe(true)
    })

    it("rejects invalid ui width", () => {
        const f = validStorageField("x", "x")
        f.ui = { component: "input", width: "third" as any }
        expect(validateField(f).some((e) => e.path.endsWith(".width"))).toBe(true)
    })

    it("rejects option value type mismatch", () => {
        const f = validStorageField("x", "x")
        f.ui = {
            component: "select",
            options: [{ label: "OK", value: "ok" }, { label: "Bad", value: 42 }],
        }
        const errors = validateField(f)
        expect(errors.some((e) => e.code === "option_type_mismatch")).toBe(true)
    })

    it("accepts correctly-typed options", () => {
        const f = validStorageField("x", "x")
        f.ui = {
            component: "select",
            options: [{ label: "GET", value: "GET" }, { label: "POST", value: "POST" }],
        }
        expect(validateField(f)).toHaveLength(0)
    })
})

// ============================================================
// ConfigField.Validate — validation rules
// ============================================================

describe("validateField — validation rules", () => {
    it("accepts valid rules", () => {
        const f = validStorageField("x", "x")
        f.validations = [
            { type: "pattern", pattern: "^https?://" },
            { type: "range", min: 0, max: 100 },
        ]
        expect(validateField(f)).toHaveLength(0)
    })

    it("propagates invalid rule errors", () => {
        const f = validStorageField("x", "x")
        f.validations = [{ type: "pattern" } as any]
        expect(validateField(f).some((e) => e.path.includes("validations[0]"))).toBe(true)
    })

    it("propagates override validation rule errors", () => {
        const f = validStorageField("x", "x")
        f.overrides = [{
            when: "x == true",
            validations: [{ type: "custom" } as any],
        }]
        expect(validateField(f).some((e) => e.path.includes("overrides[0].validations[0]"))).toBe(true)
    })
})

// ============================================================
// ConfigField.Validate — storage mapping
// ============================================================

describe("validateField — storage mapping", () => {
    it("accepts valid direct mapping", () => {
        const f = validStorageField("x", "x")
        f.storage = { type: "direct" }
        expect(validateField(f)).toHaveLength(0)
    })

    it("propagates invalid storage mapping errors", () => {
        const f = validStorageField("x", "x")
        f.storage = { type: "computed" } as any
        expect(validateField(f).some((e) => e.path.includes("storage"))).toBe(true)
    })
})

// ============================================================
// FieldValidationRule.Validate
// ============================================================

describe("validateValidationRule", () => {
    it("pattern — valid", () => {
        expect(validateValidationRule({ type: "pattern", pattern: "^[a-z]+$" })).toHaveLength(0)
    })

    it("pattern — missing pattern", () => {
        expect(validateValidationRule({ type: "pattern" } as any).some((e) => e.code === "required")).toBe(true)
    })

    it("range — min only", () => {
        expect(validateValidationRule({ type: "range", min: 1 })).toHaveLength(0)
    })

    it("range — max only", () => {
        expect(validateValidationRule({ type: "range", max: 100 })).toHaveLength(0)
    })

    it("range — neither", () => {
        expect(validateValidationRule({ type: "range" } as any).some((e) => e.code === "required")).toBe(true)
    })

    it("length — valid", () => {
        expect(validateValidationRule({ type: "length", min: 1, max: 255 })).toHaveLength(0)
    })

    it("length — neither", () => {
        expect(validateValidationRule({ type: "length" } as any)).not.toHaveLength(0)
    })

    it("itemCount — valid", () => {
        expect(validateValidationRule({ type: "itemCount", max: 10 })).toHaveLength(0)
    })

    it("itemCount — neither", () => {
        expect(validateValidationRule({ type: "itemCount" } as any)).not.toHaveLength(0)
    })

    it("allowedValues — valid", () => {
        expect(validateValidationRule({ type: "allowedValues", values: ["a", "b"] })).toHaveLength(0)
    })

    it("allowedValues — empty", () => {
        expect(validateValidationRule({ type: "allowedValues", values: [] })).not.toHaveLength(0)
    })

    it("allowedValues — missing values", () => {
        expect(validateValidationRule({ type: "allowedValues" } as any)).not.toHaveLength(0)
    })

    it("custom — valid", () => {
        expect(validateValidationRule({ type: "custom", expression: "self > 0" })).toHaveLength(0)
    })

    it("custom — missing expression", () => {
        expect(validateValidationRule({ type: "custom" } as any)).not.toHaveLength(0)
    })

    it("unknown type", () => {
        expect(validateValidationRule({ type: "banana" } as any).some((e) => e.code === "invalid_enum")).toBe(true)
    })
})

// ============================================================
// StorageMapping.Validate
// ============================================================

describe("validateStorageMapping", () => {
    it("direct — valid", () => {
        expect(validateStorageMapping({ type: "direct" })).toHaveLength(0)
    })

    it("indexedPair — valid", () => {
        expect(validateStorageMapping({
            type: "indexedPair",
            key: { target: "jsonData", pattern: "k{i}" },
            value: { target: "jsonData", pattern: "v{i}" },
        })).toHaveLength(0)
    })

    it("indexedPair — missing key", () => {
        expect(validateStorageMapping({
            type: "indexedPair",
            value: { target: "jsonData", pattern: "v{i}" },
        } as any)).not.toHaveLength(0)
    })

    it("indexedPair — missing value", () => {
        expect(validateStorageMapping({
            type: "indexedPair",
            key: { target: "jsonData", pattern: "k{i}" },
        } as any)).not.toHaveLength(0)
    })

    it("indexedPair — invalid key target", () => {
        expect(validateStorageMapping({
            type: "indexedPair",
            key: { target: "bad" as any, pattern: "k{i}" },
            value: { target: "jsonData", pattern: "v{i}" },
        }).some((e) => e.path.includes("key.target"))).toBe(true)
    })

    it("indexedPair — empty value pattern", () => {
        expect(validateStorageMapping({
            type: "indexedPair",
            key: { target: "jsonData", pattern: "k{i}" },
            value: { target: "jsonData", pattern: "" },
        }).some((e) => e.path.includes("value.pattern"))).toBe(true)
    })

    it("computed — read only", () => {
        expect(validateStorageMapping({ type: "computed", read: "expr" })).toHaveLength(0)
    })

    it("computed — write only", () => {
        expect(validateStorageMapping({ type: "computed", write: "expr" })).toHaveLength(0)
    })

    it("computed — both", () => {
        expect(validateStorageMapping({ type: "computed", read: "r", write: "w" })).toHaveLength(0)
    })

    it("computed — neither", () => {
        expect(validateStorageMapping({ type: "computed" } as any)).not.toHaveLength(0)
    })

    it("unknown type", () => {
        expect(validateStorageMapping({ type: "magic" } as any).some((e) => e.code === "invalid_enum")).toBe(true)
    })
})
