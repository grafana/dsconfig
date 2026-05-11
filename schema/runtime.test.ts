import { describe, it, expect } from "vitest"
import { readFileSync } from "fs"
import { join } from "path"
import type { DatasourceConfigSchema, ConfigField } from "./schema"
import {
    loadAndValidate,
    newDatasourceConfig,
    type DatasourceConfig,
    type LoadMode,
} from "./runtime"

// ============================================================
// Helpers
// ============================================================

function makeSchema(...fields: ConfigField[]): DatasourceConfigSchema {
    return {
        schemaVersion: "v1",
        pluginType: "test",
        pluginName: "Test",
        fields,
    }
}

// ============================================================
// newDatasourceConfig
// ============================================================

describe("newDatasourceConfig", () => {
    it("parses all sections", () => {
        const dc = newDatasourceConfig({
            url: "https://example.com",
            basicAuth: true,
            jsonData: { timeout: 30 },
            secureJsonData: { password: "secret" },
            secureJsonFields: { password: true },
        })

        expect(dc.root.url).toBe("https://example.com")
        expect(dc.root.basicAuth).toBe(true)
        expect(dc.jsonData?.timeout).toBe(30)
        expect(dc.secureJsonData?.password).toBe("secret")
        expect(dc.secureJsonFields?.password).toBe(true)
    })
})

// ============================================================
// Simple field extraction
// ============================================================

describe("loadAndValidate — extraction", () => {
    it("extracts root field", () => {
        const s = makeSchema({
            id: "url", key: "url", valueType: "string",
            target: "root", required: true,
        })
        const config: DatasourceConfig = {
            root: { url: "https://example.com" },
        }

        const result = loadAndValidate(s, config, "read")
        expect(result.errors).toHaveLength(0)
        expect(result.values.url.value).toBe("https://example.com")
        expect(result.values.url.source).toBe("config")
    })

    it("extracts jsonData field", () => {
        const s = makeSchema({
            id: "timeout", key: "timeout", valueType: "number",
            target: "jsonData",
        })
        const config: DatasourceConfig = {
            root: {},
            jsonData: { timeout: 30 },
        }

        const result = loadAndValidate(s, config, "read")
        expect(result.errors).toHaveLength(0)
        expect(result.values.timeout.value).toBe(30)
    })

    it("extracts section field", () => {
        const s = makeSchema({
            id: "oauth.clientId", key: "clientId", valueType: "string",
            target: "jsonData", section: "oauth2",
        })
        const config: DatasourceConfig = {
            root: {},
            jsonData: { oauth2: { clientId: "my-client" } },
        }

        const result = loadAndValidate(s, config, "read")
        expect(result.values["oauth.clientId"].value).toBe("my-client")
    })

    it("extracts dotted section field", () => {
        const s = makeSchema({
            id: "oauth.tokenUrl", key: "tokenUrl", valueType: "string",
            target: "jsonData", section: "oauth2.endpoints",
        })
        const config: DatasourceConfig = {
            root: {},
            jsonData: {
                oauth2: { endpoints: { tokenUrl: "https://auth.example.com/token" } },
            },
        }

        const result = loadAndValidate(s, config, "read")
        expect(result.values["oauth.tokenUrl"].value).toBe("https://auth.example.com/token")
    })
})

// ============================================================
// Defaults
// ============================================================

describe("loadAndValidate — defaults", () => {
    it("applies default when value missing", () => {
        const s = makeSchema({
            id: "timeout", key: "timeout", valueType: "number",
            target: "jsonData", defaultValue: 30,
        })
        const config: DatasourceConfig = { root: {}, jsonData: {} }

        const result = loadAndValidate(s, config, "read")
        expect(result.values.timeout.value).toBe(30)
        expect(result.values.timeout.source).toBe("default")
    })

    it("config overrides default", () => {
        const s = makeSchema({
            id: "timeout", key: "timeout", valueType: "number",
            target: "jsonData", defaultValue: 30,
        })
        const config: DatasourceConfig = { root: {}, jsonData: { timeout: 60 } }

        const result = loadAndValidate(s, config, "read")
        expect(result.values.timeout.value).toBe(60)
        expect(result.values.timeout.source).toBe("config")
    })
})

// ============================================================
// Required validation
// ============================================================

describe("loadAndValidate — required", () => {
    it("errors on missing required field", () => {
        const s = makeSchema({
            id: "url", key: "url", valueType: "string",
            target: "root", required: true,
        })
        const config: DatasourceConfig = { root: {} }

        const result = loadAndValidate(s, config, "read")
        expect(result.errors).toHaveLength(1)
        expect(result.errors[0].code).toBe("required")
        expect(result.errors[0].fieldId).toBe("url")
    })
})

// ============================================================
// Type validation
// ============================================================

describe("loadAndValidate — type check", () => {
    it("errors on type mismatch", () => {
        const s = makeSchema({
            id: "timeout", key: "timeout", valueType: "number",
            target: "jsonData",
        })
        const config: DatasourceConfig = {
            root: {},
            jsonData: { timeout: "not-a-number" },
        }

        const result = loadAndValidate(s, config, "read")
        expect(result.errors).toHaveLength(1)
        expect(result.errors[0].code).toBe("type_mismatch")
    })
})

// ============================================================
// Validation rules
// ============================================================

describe("loadAndValidate — validation rules", () => {
    it("pattern — valid", () => {
        const s = makeSchema({
            id: "url", key: "url", valueType: "string", target: "root",
            validations: [{ type: "pattern", pattern: "^https?://" }],
        })
        const result = loadAndValidate(s, { root: { url: "https://x.com" } }, "read")
        expect(result.errors).toHaveLength(0)
    })

    it("pattern — invalid", () => {
        const s = makeSchema({
            id: "url", key: "url", valueType: "string", target: "root",
            validations: [{ type: "pattern", pattern: "^https?://" }],
        })
        const result = loadAndValidate(s, { root: { url: "ftp://x.com" } }, "read")
        expect(result.errors).toHaveLength(1)
        expect(result.errors[0].code).toBe("pattern")
    })

    it("range — below min", () => {
        const s = makeSchema({
            id: "t", key: "t", valueType: "number", target: "jsonData",
            validations: [{ type: "range", min: 1, max: 300 }],
        })
        const result = loadAndValidate(s, { root: {}, jsonData: { t: 0 } }, "read")
        expect(result.errors).toHaveLength(1)
        expect(result.errors[0].code).toBe("range")
    })

    it("allowedValues — valid", () => {
        const s = makeSchema({
            id: "m", key: "m", valueType: "string", target: "jsonData",
            validations: [{ type: "allowedValues", values: ["GET", "POST"] }],
        })
        const result = loadAndValidate(s, { root: {}, jsonData: { m: "GET" } }, "read")
        expect(result.errors).toHaveLength(0)
    })

    it("allowedValues — invalid", () => {
        const s = makeSchema({
            id: "m", key: "m", valueType: "string", target: "jsonData",
            validations: [{ type: "allowedValues", values: ["GET", "POST"] }],
        })
        const result = loadAndValidate(s, { root: {}, jsonData: { m: "DELETE" } }, "read")
        expect(result.errors).toHaveLength(1)
        expect(result.errors[0].code).toBe("allowedValues")
    })

    it("itemCount — exceeds max", () => {
        const s = makeSchema({
            id: "tags", key: "tags", valueType: "array", target: "jsonData",
            item: { valueType: "string" },
            validations: [{ type: "itemCount", max: 3 }],
        })
        const result = loadAndValidate(s, { root: {}, jsonData: { tags: ["a", "b", "c", "d"] } }, "read")
        expect(result.errors).toHaveLength(1)
        expect(result.errors[0].code).toBe("itemCount")
    })
})

// ============================================================
// Secure fields
// ============================================================

describe("loadAndValidate — secure fields", () => {
    it("read mode — configured", () => {
        const s = makeSchema({
            id: "password", key: "password", valueType: "string",
            target: "secureJsonData", required: true,
        })
        const config: DatasourceConfig = {
            root: {},
            secureJsonFields: { password: true },
        }

        const result = loadAndValidate(s, config, "read")
        expect(result.errors).toHaveLength(0)
        expect(result.secureFields.password).toBe("configured")
    })

    it("read mode — missing required", () => {
        const s = makeSchema({
            id: "password", key: "password", valueType: "string",
            target: "secureJsonData", required: true,
        })
        const config: DatasourceConfig = { root: {} }

        const result = loadAndValidate(s, config, "read")
        expect(result.errors).toHaveLength(1)
        expect(result.secureFields.password).toBe("unset")
    })

    it("write mode — provided", () => {
        const s = makeSchema({
            id: "password", key: "password", valueType: "string",
            target: "secureJsonData", required: true,
        })
        const config: DatasourceConfig = {
            root: {},
            secureJsonData: { password: "s3cret" },
        }

        const result = loadAndValidate(s, config, "write")
        expect(result.errors).toHaveLength(0)
        expect(result.secureFields.password).toBe("updated")
        expect(result.values.password.value).toBe("s3cret")
    })

    it("write mode — missing required", () => {
        const s = makeSchema({
            id: "password", key: "password", valueType: "string",
            target: "secureJsonData", required: true,
        })
        const config: DatasourceConfig = { root: {}, secureJsonData: {} }

        const result = loadAndValidate(s, config, "write")
        expect(result.errors).toHaveLength(1)
        expect(result.secureFields.password).toBe("unset")
    })
})

// ============================================================
// Virtual fields
// ============================================================

describe("loadAndValidate — virtual fields", () => {
    it("skips virtual fields", () => {
        const s = makeSchema(
            { id: "url", key: "url", valueType: "string", target: "root" },
            { id: "derived", key: "derived", valueType: "boolean", kind: "virtual" },
        )
        const config: DatasourceConfig = { root: { url: "https://x.com" } }

        const result = loadAndValidate(s, config, "read")
        expect(result.values.url).toBeDefined()
        expect(result.values.derived).toBeUndefined()
    })
})

// ============================================================
// IndexedPair expansion
// ============================================================

describe("loadAndValidate — indexedPair", () => {
    function headersSchema(): DatasourceConfigSchema {
        return makeSchema({
            id: "httpHeaders", key: "httpHeaders", valueType: "array",
            target: "jsonData",
            item: {
                valueType: "object",
                fields: [
                    { id: "httpHeaders.item.name", key: "name", valueType: "string", isItemField: true },
                    { id: "httpHeaders.item.value", key: "value", valueType: "string", isItemField: true },
                ],
            },
            storage: {
                type: "indexedPair",
                key: { target: "jsonData", pattern: "httpHeaderName{index}" },
                value: { target: "secureJsonData", pattern: "httpHeaderValue{index}" },
                startIndex: 1,
            },
        })
    }

    it("expands indexed pairs", () => {
        const config: DatasourceConfig = {
            root: {},
            jsonData: { httpHeaderName1: "X-Custom", httpHeaderName2: "X-Token" },
            secureJsonData: { httpHeaderValue1: "val1", httpHeaderValue2: "val2" },
        }

        const result = loadAndValidate(headersSchema(), config, "write")
        expect(result.errors).toHaveLength(0)

        const arr = result.values.httpHeaders.value as unknown[]
        expect(arr).toHaveLength(2)
        expect((arr[0] as Record<string, unknown>).name).toBe("X-Custom")
        expect((arr[0] as Record<string, unknown>).value).toBe("val1")
        expect((arr[1] as Record<string, unknown>).name).toBe("X-Token")
        expect((arr[1] as Record<string, unknown>).value).toBe("val2")
    })

    it("handles empty indexed pairs", () => {
        const config: DatasourceConfig = { root: {}, jsonData: {} }

        const result = loadAndValidate(headersSchema(), config, "read")
        expect(result.values.httpHeaders.value).toBeNull()
        expect(result.values.httpHeaders.source).toBe("none")
    })
})

// ============================================================
// File-based integration tests
// ============================================================

describe("loadAndValidate — testdata integration", () => {
    const cases: Array<{ dir: string; mode: LoadMode }> = [
        { dir: "simple-url", mode: "read" },
        { dir: "root-jsondata-secure-mix", mode: "write" },
        { dir: "bearer-token-auth", mode: "write" },
        { dir: "indexed-headers-storage", mode: "write" },
        { dir: "nested-object-jsondata", mode: "read" },
    ]

    for (const tc of cases) {
        it(`${tc.dir}`, () => {
            const dir = join(__dirname, "testdata", "convert", tc.dir)
            const inputSchema: DatasourceConfigSchema = JSON.parse(
                readFileSync(join(dir, "input.json"), "utf-8"),
            )

            const rawConfig = JSON.parse(
                readFileSync(join(dir, "config.json"), "utf-8"),
            )
            delete rawConfig._comment

            const config = newDatasourceConfig(rawConfig)
            const result = loadAndValidate(inputSchema, config, tc.mode)

            expect(result.errors, `unexpected errors in ${tc.dir}: ${JSON.stringify(result.errors)}`).toHaveLength(0)
            expect(Object.keys(result.values).length).toBeGreaterThan(0)
        })
    }
})
