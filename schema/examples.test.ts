import { describe, it, expect } from "vitest"
import type { DatasourceConfigSchema, ConfigField } from "./schema"
import { validateSchema } from "./validate"

// ============================================================
// Realistic example schemas that exercise the full contract.
// Each test proves the schema validates cleanly AND demonstrates
// a real Grafana datasource pattern.
// ============================================================

describe("example schemas", () => {
    // --------------------------------------------------------
    // Prometheus — the canonical example
    // --------------------------------------------------------
    it("Prometheus: URL + auth + HTTP method + headers + virtual", () => {
        const s: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "prometheus",
            pluginName: "Prometheus",
            fields: [
                {
                    id: "url", key: "url", valueType: "string",
                    target: "root", required: true,
                    semanticType: "url",
                    validations: [{ type: "pattern", pattern: "^https?://", message: "Must be HTTP(S)" }],
                    ui: { component: "input", placeholder: "https://prometheus.example.com" },
                },
                {
                    id: "auth.basicAuth", key: "basicAuth",
                    valueType: "boolean", target: "root",
                },
                {
                    id: "auth.basicAuthUser", key: "basicAuthUser",
                    valueType: "string", target: "root",
                    requiredWhen: "auth.basicAuth == true",
                },
                {
                    id: "auth.basicAuthPassword", key: "basicAuthPassword",
                    valueType: "string", target: "secureJsonData",
                    semanticType: "password",
                },
                {
                    id: "jsonData.httpMethod", key: "httpMethod",
                    valueType: "string", target: "jsonData",
                    validations: [{ type: "allowedValues", values: ["GET", "POST"] }],
                    ui: {
                        component: "select",
                        options: [
                            { label: "GET", value: "GET" },
                            { label: "POST", value: "POST" },
                        ],
                    },
                },
                {
                    id: "jsonData.timeout", key: "timeout",
                    valueType: "number", target: "jsonData",
                    validations: [{ type: "range", min: 1, max: 300 }],
                },
                {
                    id: "httpHeaders", key: "httpHeaders",
                    valueType: "array", target: "jsonData",
                    item: {
                        valueType: "object",
                        fields: [
                            { id: "httpHeaders.item.key", key: "key", valueType: "string", isItemField: true },
                            { id: "httpHeaders.item.value", key: "value", valueType: "string", isItemField: true },
                        ],
                    },
                    storage: {
                        type: "indexedPair",
                        key: { target: "jsonData", pattern: "httpHeaderName{index}" },
                        value: { target: "jsonData", pattern: "httpHeaderValue{index}" },
                    },
                },
                {
                    id: "derived.hasAuth", key: "hasAuth",
                    valueType: "boolean", kind: "virtual",
                    dependsOn: "auth.basicAuth == true",
                },
            ],
            groups: [
                { id: "connection", title: "Connection", fieldRefs: ["url", "jsonData.httpMethod", "jsonData.timeout"] },
                { id: "auth", title: "Auth", fieldRefs: ["auth.basicAuth", "auth.basicAuthUser", "auth.basicAuthPassword"] },
            ],
            relationships: [
                { type: "pair", fields: ["auth.basicAuthUser", "auth.basicAuthPassword"], description: "Basic auth" },
            ],
        }
        expect(validateSchema(s)).toHaveLength(0)
    })

    // --------------------------------------------------------
    // Loki — derived fields array, simple config
    // --------------------------------------------------------
    it("Loki: URL + derived fields array + timeout", () => {
        const s: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "loki",
            pluginName: "Loki",
            fields: [
                {
                    id: "url", key: "url", valueType: "string",
                    target: "root", required: true, semanticType: "url",
                },
                {
                    id: "jsonData.maxLines", key: "maxLines",
                    valueType: "string", target: "jsonData",
                },
                {
                    id: "jsonData.derivedFields", key: "derivedFields",
                    valueType: "array", target: "jsonData",
                    item: {
                        valueType: "object",
                        fields: [
                            { id: "df.item.name", key: "name", valueType: "string", isItemField: true },
                            { id: "df.item.matcherRegex", key: "matcherRegex", valueType: "string", isItemField: true },
                            { id: "df.item.url", key: "url", valueType: "string", isItemField: true, semanticType: "url" },
                        ],
                    },
                },
                {
                    id: "jsonData.timeout", key: "timeout",
                    valueType: "number", target: "jsonData",
                    validations: [{ type: "range", min: 1, max: 600 }],
                },
            ],
            groups: [
                { id: "connection", title: "Connection", fieldRefs: ["url", "jsonData.timeout"] },
                { id: "derived", title: "Derived Fields", fieldRefs: ["jsonData.derivedFields"] },
            ],
        }
        expect(validateSchema(s)).toHaveLength(0)
    })

    // --------------------------------------------------------
    // Tempo — streaming booleans, service map, virtual field
    // --------------------------------------------------------
    it("Tempo: nested booleans + virtual field + group relationship", () => {
        const s: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "tempo",
            pluginName: "Tempo",
            fields: [
                { id: "url", key: "url", valueType: "string", target: "root", required: true, semanticType: "url" },
                { id: "jd.serviceMap.uid", key: "serviceMap.datasourceUid", valueType: "string", target: "jsonData" },
                { id: "jd.nodeGraph.enabled", key: "nodeGraph.enabled", valueType: "boolean", target: "jsonData" },
                { id: "jd.streaming.search", key: "streamingEnabled.search", valueType: "boolean", target: "jsonData" },
                { id: "jd.streaming.metrics", key: "streamingEnabled.metrics", valueType: "boolean", target: "jsonData" },
                {
                    id: "derived.hasServiceMap", key: "hasServiceMap",
                    valueType: "boolean", kind: "virtual",
                    lifecycle: "experimental",
                    dependsOn: "jd.serviceMap.uid != ''",
                },
            ],
            groups: [
                { id: "conn", title: "Connection", fieldRefs: ["url"] },
                { id: "features", title: "Features", fieldRefs: ["jd.nodeGraph.enabled", "jd.streaming.search", "jd.streaming.metrics"] },
            ],
            relationships: [
                { type: "group", fields: ["jd.streaming.search", "jd.streaming.metrics"] },
            ],
        }
        expect(validateSchema(s)).toHaveLength(0)
    })

    // --------------------------------------------------------
    // MySQL — secure fields, conditional auth, TLS
    // --------------------------------------------------------
    it("MySQL: database + secure password + TLS + conditional requirement", () => {
        const s: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "mysql",
            pluginName: "MySQL",
            fields: [
                {
                    id: "url", key: "url", valueType: "string", target: "root",
                    required: true, semanticType: "url",
                    validations: [{ type: "pattern", pattern: ".+:\\d+", message: "host:port required" }],
                },
                { id: "root.database", key: "database", valueType: "string", target: "root" },
                { id: "root.user", key: "user", valueType: "string", target: "root" },
                {
                    id: "sjd.password", key: "password", valueType: "string",
                    target: "secureJsonData", semanticType: "password",
                    requiredWhen: "root.user != ''",
                },
                {
                    id: "jd.maxConns", key: "maxOpenConns", valueType: "number", target: "jsonData",
                    validations: [{ type: "range", min: 0, max: 100 }],
                },
                {
                    id: "jd.tlsAuth", key: "tlsAuth", valueType: "boolean", target: "jsonData",
                },
                {
                    id: "sjd.tlsCACert", key: "tlsCACert", valueType: "string",
                    target: "secureJsonData",
                    dependsOn: "jd.tlsAuth == true",
                    ui: { component: "textarea", rows: 5 },
                },
            ],
            groups: [
                { id: "conn", title: "Connection", fieldRefs: ["url", "root.database"] },
                { id: "auth", title: "Auth", fieldRefs: ["root.user", "sjd.password"] },
                { id: "tls", title: "TLS", fieldRefs: ["jd.tlsAuth", "sjd.tlsCACert"] },
            ],
            relationships: [
                { type: "pair", fields: ["root.user", "sjd.password"] },
            ],
        }
        expect(validateSchema(s)).toHaveLength(0)
    })

    // --------------------------------------------------------
    // Simple bearer token — minimal practical config
    // --------------------------------------------------------
    it("Bearer token: URL + token + direct mapping", () => {
        const s: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "graphite",
            pluginName: "Graphite",
            fields: [
                { id: "url", key: "url", valueType: "string", target: "root", required: true, semanticType: "url" },
                {
                    id: "sjd.token", key: "token", valueType: "string",
                    target: "secureJsonData", semanticType: "token",
                    storage: { type: "direct" },
                },
            ],
        }
        expect(validateSchema(s)).toHaveLength(0)
    })

    // --------------------------------------------------------
    // Computed storage mapping — virtual field reads
    // --------------------------------------------------------
    it("Computed storage: virtual field with read expression", () => {
        const s: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test",
            fields: [
                { id: "jd.a", key: "a", valueType: "string", target: "jsonData" },
                { id: "jd.b", key: "b", valueType: "string", target: "jsonData" },
                {
                    id: "derived.ab", key: "ab", valueType: "string", kind: "virtual",
                    storage: { type: "computed", read: "jsonData.a + '-' + jsonData.b" },
                },
            ],
        }
        expect(validateSchema(s)).toHaveLength(0)
    })
})

// ============================================================
// Invalid example schemas — prove rejection works
// ============================================================

describe("invalid example schemas", () => {
    it("rejects schema with broken group ref", () => {
        const s: DatasourceConfigSchema = {
            schemaVersion: "v1", pluginType: "test", pluginName: "Test",
            fields: [{ id: "url", key: "url", valueType: "string", target: "root" }],
            groups: [{ id: "g1", title: "G", fieldRefs: ["missing"] }],
        }
        const errors = validateSchema(s)
        expect(errors.length).toBeGreaterThan(0)
        expect(errors.some((e) => e.code === "unknown_ref")).toBe(true)
    })

    it("rejects schema with option type mismatch", () => {
        const s: DatasourceConfigSchema = {
            schemaVersion: "v1", pluginType: "test", pluginName: "Test",
            fields: [{
                id: "x", key: "x", valueType: "string", target: "jsonData",
                ui: {
                    component: "select",
                    options: [{ label: "Bad", value: 42 }],
                },
            }],
        }
        const errors = validateSchema(s)
        expect(errors.some((e) => e.code === "option_type_mismatch")).toBe(true)
    })

    it("rejects schema with invalid validation rule in override", () => {
        const s: DatasourceConfigSchema = {
            schemaVersion: "v1", pluginType: "test", pluginName: "Test",
            fields: [{
                id: "x", key: "x", valueType: "string", target: "jsonData",
                overrides: [{
                    when: "true",
                    validations: [{ type: "custom" } as any],
                }],
            }],
        }
        const errors = validateSchema(s)
        expect(errors.some((e) => e.path.includes("overrides[0].validations[0]"))).toBe(true)
    })

    it("rejects schema with duplicate field IDs", () => {
        const s: DatasourceConfigSchema = {
            schemaVersion: "v1", pluginType: "test", pluginName: "Test",
            fields: [
                { id: "dup", key: "a", valueType: "string", target: "jsonData" },
                { id: "dup", key: "b", valueType: "string", target: "jsonData" },
            ],
        }
        const errors = validateSchema(s)
        expect(errors.some((e) => e.code === "duplicate_id")).toBe(true)
    })

    it("rejects schema with invalid storage mapping", () => {
        const s: DatasourceConfigSchema = {
            schemaVersion: "v1", pluginType: "test", pluginName: "Test",
            fields: [{
                id: "x", key: "x", valueType: "string", target: "jsonData",
                storage: { type: "computed" } as any,
            }],
        }
        const errors = validateSchema(s)
        expect(errors.some((e) => e.path.includes("storage"))).toBe(true)
    })
})

// ============================================================
// JSON round-trip: parse JSON, validate in TS
// ============================================================

describe("JSON round-trip", () => {
    it("schema survives JSON.stringify/parse and validates", () => {
        const s: DatasourceConfigSchema = {
            schemaVersion: "v1",
            pluginType: "test",
            pluginName: "Test Plugin",
            docURL: "https://example.com/docs",
            fields: [
                {
                    id: "url", key: "url", valueType: "string", target: "root",
                    required: true, semanticType: "url", lifecycle: "stable",
                    validations: [{ type: "pattern", pattern: "^https?://", id: "url-check", message: "Must be URL" }],
                    ui: { component: "input", width: "full", placeholder: "https://..." },
                },
                {
                    id: "method", key: "httpMethod", valueType: "string", target: "jsonData",
                    validations: [{ type: "allowedValues", values: ["GET", "POST"] }],
                    ui: {
                        component: "select",
                        options: [{ label: "GET", value: "GET" }, { label: "POST", value: "POST" }],
                    },
                },
                {
                    id: "headers", key: "headers", valueType: "array", target: "jsonData",
                    item: {
                        valueType: "object",
                        fields: [
                            { id: "headers.k", key: "key", valueType: "string", isItemField: true },
                            { id: "headers.v", key: "value", valueType: "string", isItemField: true },
                        ],
                    },
                    storage: {
                        type: "indexedPair",
                        key: { target: "jsonData", pattern: "headerName{i}" },
                        value: { target: "jsonData", pattern: "headerValue{i}" },
                    },
                },
            ],
            groups: [{ id: "conn", title: "Connection", fieldRefs: ["url", "method"] }],
            relationships: [{ type: "pair", fields: ["headers.k", "headers.v"] }],
        }

        const json = JSON.stringify(s)
        const parsed: DatasourceConfigSchema = JSON.parse(json)

        expect(validateSchema(parsed)).toHaveLength(0)
        expect(parsed.pluginType).toBe("test")
        expect(parsed.fields).toHaveLength(3)
    })
})
