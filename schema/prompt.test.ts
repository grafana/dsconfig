import { describe, it, expect } from "vitest"
import { toPromptSchema, toPromptString, toPromptText, extractLiteralFromWhen } from "./prompt"
import type { DatasourceConfigSchema, ConfigField } from "./schema"
import type { PromptSchema, PromptField } from "./prompt"

// ============================================================
// extractLiteralFromWhen
// ============================================================

describe("extractLiteralFromWhen", () => {
    it("extracts single-quoted string", () => {
        expect(extractLiteralFromWhen("value == 'basic-auth'")).toBe("basic-auth")
    })

    it("extracts double-quoted string", () => {
        expect(extractLiteralFromWhen('value == "forward-oauth"')).toBe("forward-oauth")
    })

    it("extracts true", () => {
        expect(extractLiteralFromWhen("value == true")).toBe(true)
    })

    it("extracts false", () => {
        expect(extractLiteralFromWhen("value == false")).toBe(false)
    })

    it("extracts integer", () => {
        expect(extractLiteralFromWhen("value == 42")).toBe(42)
    })

    it("extracts negative number", () => {
        expect(extractLiteralFromWhen("value == -3.14")).toBe(-3.14)
    })

    it("returns undefined for complex expressions", () => {
        expect(extractLiteralFromWhen("value.startsWith('http')")).toBeUndefined()
    })

    it("handles whitespace variations", () => {
        expect(extractLiteralFromWhen("value  ==  'spaced'")).toBe("spaced")
    })
})

// ============================================================
// toPromptSchema — simple fields
// ============================================================

describe("toPromptSchema", () => {
    const minimal: DatasourceConfigSchema = {
        schemaVersion: "v1",
        pluginType: "test",
        pluginName: "Test Plugin",
        fields: [
            {
                id: "url",
                key: "url",
                valueType: "string",
                semanticType: "url",
                target: "root",
                required: true,
                description: "Base URL",
            },
        ],
    }

    it("preserves identity and core properties", () => {
        const ps = toPromptSchema(minimal)
        expect(ps.pluginType).toBe("test")
        expect(ps.pluginName).toBe("Test Plugin")
        expect(ps.fields).toHaveLength(1)

        const f = ps.fields[0]
        expect(f.id).toBe("url")
        expect(f.path).toBe("root.url")
        expect(f.type).toBe("string")
        expect(f.semanticType).toBe("url")
        expect(f.required).toBe(true)
        expect(f.description).toBe("Base URL")
    })

    it("strips UI hints", () => {
        const schema: DatasourceConfigSchema = {
            ...minimal,
            fields: [
                {
                    ...minimal.fields[0],
                    ui: { component: "input", placeholder: "https://...", width: "full" },
                },
            ],
        }
        const ps = toPromptSchema(schema)
        const f = ps.fields[0] as unknown as Record<string, unknown>
        expect(f["ui"]).toBeUndefined()
        expect(f["placeholder"]).toBeUndefined()
    })

    it("strips groups", () => {
        const schema: DatasourceConfigSchema = {
            ...minimal,
            groups: [{ id: "conn", title: "Connection", fieldRefs: ["url"] }],
        }
        const ps = toPromptSchema(schema) as unknown as Record<string, unknown>
        expect(ps["groups"]).toBeUndefined()
    })

    it("flattens allowedValues from validations", () => {
        const schema: DatasourceConfigSchema = {
            ...minimal,
            fields: [
                {
                    id: "method",
                    key: "httpMethod",
                    valueType: "string",
                    target: "jsonData",
                    validations: [{ type: "allowedValues", values: ["GET", "POST"] }],
                },
            ],
        }
        const ps = toPromptSchema(schema)
        expect(ps.fields[0].allowedValues).toEqual(["GET", "POST"])
    })

    it("flattens pattern from validations", () => {
        const schema: DatasourceConfigSchema = {
            ...minimal,
            fields: [
                {
                    id: "url",
                    key: "url",
                    valueType: "string",
                    target: "root",
                    validations: [{ type: "pattern", pattern: "^https?://" }],
                },
            ],
        }
        const ps = toPromptSchema(schema)
        expect(ps.fields[0].pattern).toBe("^https?://")
    })

    it("flattens range from validations", () => {
        const schema: DatasourceConfigSchema = {
            ...minimal,
            fields: [
                {
                    id: "timeout",
                    key: "timeout",
                    valueType: "number",
                    target: "jsonData",
                    validations: [{ type: "range", min: 1, max: 600 }],
                },
            ],
        }
        const ps = toPromptSchema(schema)
        expect(ps.fields[0].range).toEqual({ min: 1, max: 600 })
    })

    it("includes defaultValue", () => {
        const schema: DatasourceConfigSchema = {
            ...minimal,
            fields: [
                {
                    id: "method",
                    key: "httpMethod",
                    valueType: "string",
                    target: "jsonData",
                    defaultValue: "POST",
                },
            ],
        }
        const ps = toPromptSchema(schema)
        expect(ps.fields[0].defaultValue).toBe("POST")
    })

    it("builds path with section", () => {
        const schema: DatasourceConfigSchema = {
            ...minimal,
            fields: [
                {
                    id: "nested",
                    key: "datasourceUid",
                    valueType: "string",
                    target: "jsonData",
                    section: "tracesToLogs",
                },
            ],
        }
        const ps = toPromptSchema(schema)
        expect(ps.fields[0].path).toBe("jsonData.tracesToLogs.datasourceUid")
    })
})

// ============================================================
// toPromptSchema — managed fields (hidden)
// ============================================================

describe("toPromptSchema — managed fields", () => {
    it("excludes fields tagged managed-by:*", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                { id: "visible", key: "url", valueType: "string", target: "root" },
                { id: "hidden", key: "basicAuth", valueType: "boolean", target: "root", tags: ["managed-by:auth.method"] },
            ],
        }
        const ps = toPromptSchema(schema)
        expect(ps.fields).toHaveLength(1)
        expect(ps.fields[0].id).toBe("visible")
    })
})

// ============================================================
// toPromptSchema — effects → options
// ============================================================

describe("toPromptSchema — effects", () => {
    const authSchema: DatasourceConfigSchema = {
        schemaVersion: "v1",
        pluginType: "test",
        pluginName: "Test",
        fields: [
            {
                id: "auth.method",
                key: "authMethod",
                label: "Authentication method",
                valueType: "string",
                kind: "virtual",
                defaultValue: "no-auth",
                ui: {
                    component: "select",
                    options: [
                        { label: "No Authentication", value: "no-auth" },
                        { label: "Basic authentication", value: "basic-auth" },
                        { label: "Forward OAuth Identity", value: "forward-oauth" },
                    ],
                },
                effects: [
                    { when: "value == 'no-auth'", set: { "auth.basicAuth": false, "auth.oauthPassThru": false } },
                    { when: "value == 'basic-auth'", set: { "auth.basicAuth": true, "auth.oauthPassThru": false } },
                    { when: "value == 'forward-oauth'", set: { "auth.basicAuth": false, "auth.oauthPassThru": true } },
                ],
            },
            { id: "auth.basicAuth", key: "basicAuth", valueType: "boolean", target: "root", tags: ["managed-by:auth.method"] },
            { id: "auth.oauthPassThru", key: "oauthPassThru", valueType: "boolean", target: "jsonData", tags: ["managed-by:auth.method"] },
        ],
    }

    it("flattens effects into options with labels from ui.options", () => {
        const ps = toPromptSchema(authSchema)
        // managed fields are excluded
        expect(ps.fields).toHaveLength(1)

        const f = ps.fields[0]
        expect(f.options).toHaveLength(3)
        expect(f.options![0]).toEqual({
            value: "no-auth",
            label: "No Authentication",
            sets: { "auth.basicAuth": false, "auth.oauthPassThru": false },
        })
        expect(f.options![1].label).toBe("Basic authentication")
        expect(f.options![2].value).toBe("forward-oauth")
    })

    it("uses value as label when no ui.options", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                {
                    id: "x",
                    key: "x",
                    valueType: "string",
                    kind: "virtual",
                    effects: [{ when: "value == 'a'", set: { y: true } }],
                },
                { id: "y", key: "y", valueType: "boolean", target: "jsonData" },
            ],
        }
        const ps = toPromptSchema(schema)
        expect(ps.fields[0].options![0].label).toBe("a")
    })
})

// ============================================================
// toPromptSchema — array items
// ============================================================

describe("toPromptSchema — items", () => {
    it("includes item fields for array of objects", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                {
                    id: "queries",
                    key: "queries",
                    valueType: "array",
                    target: "jsonData",
                    item: {
                        valueType: "object",
                        fields: [
                            { id: "queries.item.name", key: "name", valueType: "string", isItemField: true },
                            { id: "queries.item.query", key: "query", valueType: "string", isItemField: true, description: "PromQL query" },
                        ],
                    },
                },
            ],
        }
        const ps = toPromptSchema(schema)
        expect(ps.fields[0].items).toHaveLength(2)
        expect(ps.fields[0].items![0].id).toBe("queries.item.name")
        expect(ps.fields[0].items![1].description).toBe("PromQL query")
    })
})

// ============================================================
// toPromptString
// ============================================================

describe("toPromptString", () => {
    it("returns valid JSON", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                { id: "url", key: "url", valueType: "string", target: "root", required: true },
            ],
        }
        const str = toPromptString(schema)
        const parsed = JSON.parse(str) as PromptSchema
        expect(parsed.pluginType).toBe("test")
        expect(parsed.fields).toHaveLength(1)
    })

    it("is pretty-printed", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                { id: "url", key: "url", valueType: "string", target: "root" },
            ],
        }
        const str = toPromptString(schema)
        expect(str).toContain("\n")
        expect(str).toContain("  ")
    })
})
// ============================================================
// toPromptText — header
// ============================================================

describe("toPromptText", () => {
    it("renders header with plugin name and type", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "prometheus",
            pluginName: "Prometheus",
            fields: [
                { id: "url", key: "url", valueType: "string", target: "root" },
            ],
        }
        const text = toPromptText(schema)
        expect(text).toContain("Prometheus (pluginType: prometheus)")
        expect(text).toContain("Fields:")
    })

    it("renders a simple required field with description", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                {
                    id: "url",
                    key: "url",
                    label: "URL",
                    valueType: "string",
                    semanticType: "url",
                    target: "root",
                    required: true,
                    description: "Base URL of the server",
                },
            ],
        }
        const text = toPromptText(schema)
        expect(text).toContain("- URL (root.url) [string, url] REQUIRED — Base URL of the server")
    })

    it("renders field with default value", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                {
                    id: "method",
                    key: "httpMethod",
                    valueType: "string",
                    target: "jsonData",
                    defaultValue: "POST",
                },
            ],
        }
        const text = toPromptText(schema)
        expect(text).toContain('default: "POST"')
    })

    it("uses label when available, falls back to id", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                { id: "jsonData.timeout", key: "timeout", label: "Timeout", valueType: "number", target: "jsonData" },
                { id: "jsonData.debug", key: "debug", valueType: "boolean", target: "jsonData" },
            ],
        }
        const text = toPromptText(schema)
        expect(text).toContain("- Timeout (jsonData.timeout)")
        expect(text).toContain("- jsonData.debug (jsonData.debug)")
    })

    it("renders allowedValues as Allowed line", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                {
                    id: "method",
                    key: "httpMethod",
                    valueType: "string",
                    target: "jsonData",
                    validations: [{ type: "allowedValues", values: ["GET", "POST"] }],
                },
            ],
        }
        const text = toPromptText(schema)
        expect(text).toContain('Allowed: "GET", "POST"')
    })

    it("renders pattern constraint", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                {
                    id: "url",
                    key: "url",
                    valueType: "string",
                    target: "root",
                    validations: [{ type: "pattern", pattern: "^https?://" }],
                },
            ],
        }
        const text = toPromptText(schema)
        expect(text).toContain("Pattern: ^https?://")
    })

    it("renders range constraint", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                {
                    id: "timeout",
                    key: "timeout",
                    valueType: "number",
                    target: "jsonData",
                    validations: [{ type: "range", min: 1, max: 600 }],
                },
            ],
        }
        const text = toPromptText(schema)
        expect(text).toContain("Range: min: 1, max: 600")
    })

    it("renders dependsOn as Visible when", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                {
                    id: "user",
                    key: "basicAuthUser",
                    valueType: "string",
                    target: "root",
                    dependsOn: "auth.method == 'basic-auth'",
                },
            ],
        }
        const text = toPromptText(schema)
        expect(text).toContain("Visible when: auth.method == 'basic-auth'")
    })

    it("renders requiredWhen as Required when", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                {
                    id: "user",
                    key: "basicAuthUser",
                    valueType: "string",
                    target: "root",
                    requiredWhen: "auth.method == 'basic-auth'",
                },
            ],
        }
        const text = toPromptText(schema)
        expect(text).toContain("Required when: auth.method == 'basic-auth'")
    })

    it("renders virtual selector with effects as Options", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                {
                    id: "auth.method",
                    key: "authMethod",
                    label: "Authentication method",
                    valueType: "string",
                    kind: "virtual",
                    defaultValue: "no-auth",
                    ui: {
                        component: "select",
                        options: [
                            { label: "No Auth", value: "no-auth" },
                            { label: "Basic Auth", value: "basic-auth" },
                        ],
                    },
                    effects: [
                        { when: "value == 'no-auth'", set: { "ba": false } },
                        { when: "value == 'basic-auth'", set: { "ba": true } },
                    ],
                },
                { id: "ba", key: "basicAuth", valueType: "boolean", target: "root", tags: ["managed-by:auth.method"] },
            ],
        }
        const text = toPromptText(schema)
        expect(text).toContain("Options:")
        expect(text).toContain('"no-auth" (No Auth) → sets ba=false')
        expect(text).toContain('"basic-auth" (Basic Auth) → sets ba=true')
        // Should NOT show allowedValues when options are present
        expect(text).not.toContain("Allowed:")
    })

    it("excludes managed-by fields", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                { id: "visible", key: "url", valueType: "string", target: "root" },
                { id: "hidden", key: "basicAuth", valueType: "boolean", target: "root", tags: ["managed-by:auth.method"] },
            ],
        }
        const text = toPromptText(schema)
        expect(text).toContain("visible")
        expect(text).not.toContain("hidden")
        expect(text).not.toContain("basicAuth")
    })

    it("renders array item fields indented", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                {
                    id: "queries",
                    key: "queries",
                    label: "Queries",
                    valueType: "array",
                    target: "jsonData",
                    item: {
                        valueType: "object",
                        fields: [
                            { id: "queries.item.name", key: "name", label: "Name", valueType: "string", isItemField: true },
                            { id: "queries.item.query", key: "query", label: "Query", valueType: "string", isItemField: true, description: "PromQL" },
                        ],
                    },
                },
            ],
        }
        const text = toPromptText(schema)
        expect(text).toContain("Item fields:")
        expect(text).toContain("- Name")
        expect(text).toContain("- Query")
        expect(text).toContain("— PromQL")
    })

    it("renders section path correctly", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                {
                    id: "nested",
                    key: "datasourceUid",
                    valueType: "string",
                    semanticType: "datasourceUid",
                    target: "jsonData",
                    section: "tracesToLogs",
                },
            ],
        }
        const text = toPromptText(schema)
        expect(text).toContain("(jsonData.tracesToLogs.datasourceUid)")
        expect(text).toContain("[string, datasourceUid]")
    })

    it("renders full auth-selector example", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "prometheus",
            pluginName: "Prometheus",
            fields: [
                {
                    id: "url", key: "url", label: "URL",
                    valueType: "string", semanticType: "url", target: "root",
                    required: true, description: "Prometheus server URL",
                    validations: [{ type: "pattern", pattern: "^https?://" }],
                },
                {
                    id: "auth.method", key: "authMethod", label: "Authentication method",
                    description: "Choose an authentication method",
                    valueType: "string", kind: "virtual", defaultValue: "no-auth",
                    ui: {
                        component: "select",
                        options: [
                            { label: "No Authentication", value: "no-auth" },
                            { label: "Basic authentication", value: "basic-auth" },
                            { label: "Forward OAuth Identity", value: "forward-oauth" },
                        ],
                    },
                    effects: [
                        { when: "value == 'no-auth'", set: { "auth.basicAuth": false, "auth.oauthPassThru": false } },
                        { when: "value == 'basic-auth'", set: { "auth.basicAuth": true, "auth.oauthPassThru": false } },
                        { when: "value == 'forward-oauth'", set: { "auth.basicAuth": false, "auth.oauthPassThru": true } },
                    ],
                },
                { id: "auth.basicAuth", key: "basicAuth", valueType: "boolean", target: "root", tags: ["managed-by:auth.method"] },
                { id: "auth.oauthPassThru", key: "oauthPassThru", valueType: "boolean", target: "jsonData", tags: ["managed-by:auth.method"] },
                {
                    id: "auth.basicAuthUser", key: "basicAuthUser", label: "Username",
                    valueType: "string", target: "root",
                    dependsOn: "auth.method == 'basic-auth'",
                    requiredWhen: "auth.method == 'basic-auth'",
                },
                {
                    id: "auth.basicAuthPassword", key: "basicAuthPassword", label: "Password",
                    valueType: "string", semanticType: "password", target: "secureJsonData",
                    dependsOn: "auth.method == 'basic-auth'",
                },
                {
                    id: "jsonData.httpMethod", key: "httpMethod", label: "HTTP Method",
                    valueType: "string", target: "jsonData", defaultValue: "POST",
                    validations: [{ type: "allowedValues", values: ["GET", "POST"] }],
                },
                {
                    id: "jsonData.timeout", key: "timeout", label: "Timeout",
                    valueType: "number", target: "jsonData",
                    description: "HTTP timeout in seconds",
                    validations: [{ type: "range", min: 1, max: 600 }],
                },
            ],
        }
        const text = toPromptText(schema)

        // Header
        expect(text).toContain("Prometheus (pluginType: prometheus)")

        // Required URL field
        expect(text).toContain("- URL (root.url) [string, url] REQUIRED — Prometheus server URL")
        expect(text).toContain("Pattern: ^https?://")

        // Virtual auth selector with options
        expect(text).toContain('Authentication method [string] default: "no-auth" — Choose an authentication method')
        expect(text).toContain('"no-auth" (No Authentication)')
        expect(text).toContain('"basic-auth" (Basic authentication)')
        expect(text).toContain('"forward-oauth" (Forward OAuth Identity)')

        // Hidden managed fields NOT present as standalone field lines
        // (they may appear inside effects "sets" which is correct)
        expect(text).not.toMatch(/^- .*auth\.basicAuth/m)
        expect(text).not.toMatch(/^- .*auth\.oauthPassThru/m)

        // Conditional fields
        expect(text).toContain("- Username (root.basicAuthUser)")
        expect(text).toContain("Required when: auth.method == 'basic-auth'")
        expect(text).toContain("Visible when: auth.method == 'basic-auth'")

        // Password field
        expect(text).toContain("- Password (secureJsonData.basicAuthPassword) [string, password]")

        // Enum field
        expect(text).toContain('- HTTP Method (jsonData.httpMethod) [string] default: "POST"')
        expect(text).toContain('Allowed: "GET", "POST"')

        // Range field
        expect(text).toContain("- Timeout (jsonData.timeout) [number] — HTTP timeout in seconds")
        expect(text).toContain("Range: min: 1, max: 600")
    })

    it("renders boolean default values correctly", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                {
                    id: "flag",
                    key: "flag",
                    valueType: "boolean",
                    target: "jsonData",
                    defaultValue: false,
                },
            ],
        }
        const text = toPromptText(schema)
        expect(text).toContain("default: false")
    })

    it("renders semantic type in brackets", () => {
        const schema: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                {
                    id: "token",
                    key: "token",
                    label: "API Token",
                    valueType: "string",
                    semanticType: "token",
                    target: "secureJsonData",
                },
            ],
        }
        const text = toPromptText(schema)
        expect(text).toContain("[string, token]")
    })
})