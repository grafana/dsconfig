import { describe, it, expect } from "vitest"
import Ajv from "ajv"
import addFormats from "ajv-formats"
import { readFileSync, readdirSync } from "node:fs"
import { join } from "node:path"

/**
 * JSON Schema contract tests.
 *
 * These tests prove that every example schema file validates against
 * the official schema.json. This is a different guarantee than the
 * TypeScript validateSchema() tests — it proves the public JSON Schema
 * contract (consumed by external tools, docs generators, and LLM
 * integrations) accepts the examples we ship.
 */

const schemaDir = join(__dirname)
const examplesDir = join(__dirname, "examples")

function loadJsonSchema() {
    const raw = readFileSync(join(schemaDir, "schema.json"), "utf-8")
    return JSON.parse(raw)
}

function exampleFiles(): string[] {
    return readdirSync(examplesDir).filter((f) => f.endsWith(".schema.json"))
}

describe("JSON Schema (AJV) validation", () => {
    const ajv = new Ajv({ allErrors: true })
    addFormats(ajv)
    const schemaSpec = loadJsonSchema()
    const validate = ajv.compile(schemaSpec)

    it("schema.json is a valid JSON Schema", () => {
        expect(validate).toBeDefined()
    })

    const files = exampleFiles()

    it("found at least 5 example files", () => {
        expect(files.length).toBeGreaterThanOrEqual(5)
    })

    for (const file of files) {
        it(`${file} validates against schema.json`, () => {
            const raw = readFileSync(join(examplesDir, file), "utf-8")
            const example = JSON.parse(raw)

            const ok = validate(example)

            expect(ok, `${file}: ${ajv.errorsText(validate.errors)}`).toBe(true)
        })
    }
})
