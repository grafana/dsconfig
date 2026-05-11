import { describe, it, expect } from "vitest"
import { readFileSync, readdirSync, existsSync } from "fs"
import { join } from "path"
import type { DatasourceConfigSchema, ConfigField } from "./schema"
import { validateSchema } from "./validate"

/**
 * Cross-language contract test: loads example JSON files and validates
 * them with the TypeScript validateSchema(). The same files are validated
 * by Go tests in examples_load_test.go, ensuring both implementations
 * agree on what constitutes a valid schema.
 */

const examplesDir = join(__dirname, "examples")

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

// ============================================================
// Datasource config.schema.json files (src/*/)
// ============================================================

const srcDir = join(__dirname, "..", "src")

function datasourceSchemaFiles(): string[] {
    if (!existsSync(srcDir)) {
        return []
    }
    return readdirSync(srcDir, { withFileTypes: true })
        .filter((d) => d.isDirectory() && existsSync(join(srcDir, d.name, "config.schema.json")))
        .map((d) => d.name)
}

function loadDatasourceSchema(dir: string): DatasourceConfigSchema {
    const data = readFileSync(join(srcDir, dir, "config.schema.json"), "utf-8")
    return JSON.parse(data) as DatasourceConfigSchema
}

/** Collect all fields including nested item fields */
function collectAllFields(fields: ConfigField[]): Array<{ path: string; field: ConfigField }> {
    const result: Array<{ path: string; field: ConfigField }> = []
    for (const f of fields) {
        result.push({ path: f.id, field: f })
        if (f.item?.fields) {
            for (const sub of f.item.fields) {
                result.push({ path: `${f.id}.item.${sub.id}`, field: sub })
            }
        }
    }
    return result
}

describe("datasource schemas — validation", () => {
    const dirs = datasourceSchemaFiles()

    it("should find at least 60 datasource schemas", () => {
        expect(dirs.length).toBeGreaterThanOrEqual(60)
    })

    for (const dir of dirs) {
        it(`${dir}: validates cleanly`, () => {
            const schema = loadDatasourceSchema(dir)
            const errors = validateSchema(schema)
            expect(errors, `${dir}/config.schema.json validation errors`).toHaveLength(0)
        })
    }
})

describe("datasource schemas — every field must have label and description", () => {
    const dirs = datasourceSchemaFiles()

    for (const dir of dirs) {
        it(`${dir}: all fields have label and description`, () => {
            const schema = loadDatasourceSchema(dir)
            const allFields = collectAllFields(schema.fields)
            const missing: string[] = []

            for (const { path, field } of allFields) {
                if (!field.label) {
                    missing.push(`${path}: missing label`)
                }
                if (!field.description) {
                    missing.push(`${path}: missing description`)
                }
            }

            expect(missing, `${dir}/config.schema.json fields missing label/description`).toHaveLength(0)
        })
    }
})

describe("datasource schemas — pluginType matches directory name", () => {
    const dirs = datasourceSchemaFiles()

    for (const dir of dirs) {
        it(`${dir}: pluginType matches directory`, () => {
            const schema = loadDatasourceSchema(dir)
            expect(schema.pluginType).toBe(dir)
        })
    }
})
