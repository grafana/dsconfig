import { describe, it, expect } from "vitest"
import {
    isValueType,
    isSemanticType,
    isFieldKind,
    isLifecycle,
    isTargetLocation,
    isUIComponent,
    isUIWidth,
    isRelationshipType,
    isStorageMapping,
    isValidationRule,
    isValidOptionValue,
} from "./guards"

// ============================================================
// Enum guards
// ============================================================

describe("isValueType", () => {
    it.each(["string", "number", "boolean", "array", "object"])("accepts %s", (v) => {
        expect(isValueType(v)).toBe(true)
    })
    it.each(["", "blob", "int", null, undefined, 42])("rejects %s", (v) => {
        expect(isValueType(v)).toBe(false)
    })
})

describe("isSemanticType", () => {
    it.each(["url", "password", "token", "hostname", "duration"])("accepts %s", (v) => {
        expect(isSemanticType(v)).toBe(true)
    })
    it.each(["", "email", null, 42])("rejects %s", (v) => {
        expect(isSemanticType(v)).toBe(false)
    })
})

describe("isFieldKind", () => {
    it.each(["storage", "virtual"])("accepts %s", (v) => {
        expect(isFieldKind(v)).toBe(true)
    })
    it.each(["", "computed", null])("rejects %s", (v) => {
        expect(isFieldKind(v)).toBe(false)
    })
})

describe("isLifecycle", () => {
    it.each(["stable", "deprecated", "experimental"])("accepts %s", (v) => {
        expect(isLifecycle(v)).toBe(true)
    })
    it.each(["", "beta", null])("rejects %s", (v) => {
        expect(isLifecycle(v)).toBe(false)
    })
})

describe("isTargetLocation", () => {
    it.each(["root", "jsonData", "secureJsonData"])("accepts %s", (v) => {
        expect(isTargetLocation(v)).toBe(true)
    })
    it.each(["", "metadata", null])("rejects %s", (v) => {
        expect(isTargetLocation(v)).toBe(false)
    })
})

describe("isUIComponent", () => {
    it.each([
        "input", "textarea", "select", "multiselect", "radio",
        "checkbox", "switch", "code", "keyvalue", "list",
    ])("accepts %s", (v) => {
        expect(isUIComponent(v)).toBe(true)
    })
    it.each(["", "datepicker", null])("rejects %s", (v) => {
        expect(isUIComponent(v)).toBe(false)
    })
})

describe("isUIWidth", () => {
    it.each(["full", "half"])("accepts %s", (v) => {
        expect(isUIWidth(v)).toBe(true)
    })
    it.each(["", "third", null])("rejects %s", (v) => {
        expect(isUIWidth(v)).toBe(false)
    })
})

describe("isRelationshipType", () => {
    it.each(["pair", "group"])("accepts %s", (v) => {
        expect(isRelationshipType(v)).toBe(true)
    })
    it.each(["", "dependency", null])("rejects %s", (v) => {
        expect(isRelationshipType(v)).toBe(false)
    })
})

// ============================================================
// Discriminated union guards
// ============================================================

describe("isStorageMapping", () => {
    it("accepts direct", () => expect(isStorageMapping({ type: "direct" })).toBe(true))
    it("accepts indexedPair", () => expect(isStorageMapping({ type: "indexedPair" })).toBe(true))
    it("accepts computed", () => expect(isStorageMapping({ type: "computed" })).toBe(true))
    it("rejects unknown type", () => expect(isStorageMapping({ type: "magic" })).toBe(false))
    it("rejects null", () => expect(isStorageMapping(null)).toBe(false))
    it("rejects string", () => expect(isStorageMapping("direct")).toBe(false))
    it("rejects missing type", () => expect(isStorageMapping({})).toBe(false))
})

describe("isValidationRule", () => {
    it.each(["pattern", "range", "length", "itemCount", "allowedValues", "custom"])("accepts %s", (t) => {
        expect(isValidationRule({ type: t })).toBe(true)
    })
    it("rejects unknown type", () => expect(isValidationRule({ type: "banana" })).toBe(false))
    it("rejects null", () => expect(isValidationRule(null)).toBe(false))
    it("rejects string", () => expect(isValidationRule("pattern")).toBe(false))
})

// ============================================================
// Option value type checking
// ============================================================

describe("isValidOptionValue", () => {
    it("string match", () => expect(isValidOptionValue("hello", "string")).toBe(true))
    it("string mismatch", () => expect(isValidOptionValue(42, "string")).toBe(false))

    it("number int match", () => expect(isValidOptionValue(42, "number")).toBe(true))
    it("number float match", () => expect(isValidOptionValue(3.14, "number")).toBe(true))
    it("number mismatch", () => expect(isValidOptionValue("42", "number")).toBe(false))

    it("boolean match", () => expect(isValidOptionValue(true, "boolean")).toBe(true))
    it("boolean mismatch", () => expect(isValidOptionValue("true", "boolean")).toBe(false))

    it("null rejected for string", () => expect(isValidOptionValue(null, "string")).toBe(false))
    it("null rejected for number", () => expect(isValidOptionValue(null, "number")).toBe(false))
    it("null rejected for boolean", () => expect(isValidOptionValue(null, "boolean")).toBe(false))
    it("null rejected for array", () => expect(isValidOptionValue(null, "array")).toBe(false))
    it("undefined rejected", () => expect(isValidOptionValue(undefined, "string")).toBe(false))

    it("array/object values accepted (not type-checked)", () => {
        expect(isValidOptionValue("anything", "array")).toBe(true)
        expect(isValidOptionValue(42, "object")).toBe(true)
    })
})
