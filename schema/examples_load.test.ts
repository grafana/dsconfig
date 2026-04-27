import { describe, it, expect } from "vitest"
import { readFileSync, readdirSync } from "fs"
import { join } from "path"
import type { DatasourceConfigSchema } from "./schema"
import { validateSchema } from "./validate"

/**
 * Cross-language contract test: loads example JSON files and validates
 * them with the TypeScript validateSchema(). The same files are validated
 * by Go tests in examples_load_test.go, ensuring both implementations
 * agree on what constitutes a valid schema.
 */

const examplesDir = join(__dirname, "..", "examples")

function loadExample(filename: string): DatasourceConfigSchema {
    const data = readFileSync(join(examplesDir, filename), "utf-8")
    return JSON.parse(data) as DatasourceConfigSchema
}

function exampleFiles(): string[] {
    return readdirSync(examplesDir).filter((f) => f.endsWith(".schema.json"))
}

describe("example files — TS validation", () => {
    const cases: Array<{
        file: string
        description: string
        fieldCount: number
    }> = [
            {
                file: "simple-url.schema.json",
                description: "Minimal schema: single URL field with pattern validation",
                fieldCount: 1,
            },
            {
                file: "bearer-token.schema.json",
                description: "Auth schema: URL + auth method select + secure bearer token",
                fieldCount: 3,
            },
            {
                file: "indexed-headers.schema.json",
                description: "Storage mapping: array with indexedPair mapping",
                fieldCount: 4,
            },
            {
                file: "virtual-auth.schema.json",
                description: "Virtual fields: basic auth with computed field",
                fieldCount: 5,
            },
            {
                file: "array-of-objects.schema.json",
                description: "Array of objects: trace-to-metrics queries",
                fieldCount: 4,
            },
        ]

    for (const tc of cases) {
        it(`${tc.file}: ${tc.description}`, () => {
            const schema = loadExample(tc.file)

            // Validate the full schema
            const errors = validateSchema(schema)
            expect(errors).toHaveLength(0)

            // Verify root fields
            expect(schema.schemaVersion).toBeTruthy()
            expect(schema.pluginType).toBeTruthy()
            expect(schema.pluginName).toBeTruthy()
        })
    }
})

describe("example files — TS round-trip", () => {
    const files = exampleFiles()

    for (const file of files) {
        it(`${file}: survives JSON.stringify/parse and validates`, () => {
            const original = loadExample(file)
            const errors = validateSchema(original)
            expect(errors).toHaveLength(0)

            // Round-trip
            const json = JSON.stringify(original)
            const decoded: DatasourceConfigSchema = JSON.parse(json)
            expect(validateSchema(decoded)).toHaveLength(0)

            // Key properties survive
            expect(decoded.schemaVersion).toBe(original.schemaVersion)
            expect(decoded.pluginType).toBe(original.pluginType)
            expect(decoded.fields).toHaveLength(original.fields.length)
        })
    }
})

describe("example files — all discovered files validate", () => {
    it("every .schema.json file in examples/ passes TS validation", () => {
        const files = exampleFiles()
        expect(files.length).toBeGreaterThanOrEqual(5)

        for (const file of files) {
            const schema = loadExample(file)
            const errors = validateSchema(schema)
            expect(errors, `${file} should validate cleanly`).toHaveLength(0)
        }
    })
})
