import type { ConfigField, FieldItemSchema, FieldUI } from "./schema"

// Common field sets shared across many datasource schemas.
// These helpers reduce duplication when building config schemas.
// The generated JSON files remain self-contained.

/**
 * Standard basic-auth fields (toggle, username, password).
 */
export function basicAuthFields(): ConfigField[] {
    return [
        {
            id: "auth.basicAuth", key: "basicAuth",
            label: "Basic Auth", description: "Enable basic authentication",
            valueType: "boolean", target: "root",
            ui: { component: "switch" },
        },
        {
            id: "auth.basicAuthUser", key: "basicAuthUser",
            label: "Username", valueType: "string", target: "root",
            dependsOn: "auth.basicAuth == true",
            requiredWhen: "auth.basicAuth == true",
        },
        {
            id: "auth.basicAuthPassword", key: "basicAuthPassword",
            label: "Password", valueType: "string", target: "secureJsonData",
            semanticType: "password",
            dependsOn: "auth.basicAuth == true",
        },
    ]
}

/**
 * Standard TLS/SSL fields.
 */
export function tlsFields(): ConfigField[] {
    return [
        {
            id: "tls.tlsAuth", key: "tlsAuth",
            label: "TLS Client Authentication", valueType: "boolean", target: "jsonData",
            ui: { component: "switch" },
        },
        {
            id: "tls.tlsAuthWithCACert", key: "tlsAuthWithCACert",
            label: "With CA Cert", valueType: "boolean", target: "jsonData",
            ui: { component: "switch" },
        },
        {
            id: "tls.tlsSkipVerify", key: "tlsSkipVerify",
            label: "Skip TLS Verify", valueType: "boolean", target: "jsonData",
            ui: { component: "switch" },
        },
        {
            id: "tls.serverName", key: "serverName",
            label: "Server Name", valueType: "string", target: "jsonData",
            semanticType: "hostname",
        },
        {
            id: "tls.tlsCACert", key: "tlsCACert",
            label: "CA Cert", valueType: "string", target: "secureJsonData",
            dependsOn: "tls.tlsAuthWithCACert == true",
            ui: { component: "textarea", rows: 7 },
        },
        {
            id: "tls.tlsClientCert", key: "tlsClientCert",
            label: "Client Cert", valueType: "string", target: "secureJsonData",
            dependsOn: "tls.tlsAuth == true",
            ui: { component: "textarea", rows: 7 },
        },
        {
            id: "tls.tlsClientKey", key: "tlsClientKey",
            label: "Client Key", valueType: "string", target: "secureJsonData",
            dependsOn: "tls.tlsAuth == true",
            ui: { component: "textarea", rows: 7 },
        },
    ]
}

/**
 * Common network fields: timeout, keepCookies, oauthPassThru, pdcInjected.
 */
export function commonNetworkFields(): ConfigField[] {
    return [
        {
            id: "network.timeout", key: "timeout",
            label: "Timeout", description: "HTTP request timeout in seconds",
            valueType: "number", target: "jsonData",
            validations: [{ type: "range", min: 1, max: 600 }],
        },
        {
            id: "network.keepCookies", key: "keepCookies",
            label: "Allowed Cookies", description: "Cookies to forward to the datasource",
            valueType: "array", target: "jsonData",
            item: { valueType: "string" },
        },
        {
            id: "network.oauthPassThru", key: "oauthPassThru",
            label: "Forward OAuth Identity",
            valueType: "boolean", target: "jsonData",
            ui: { component: "switch" },
        },
        {
            id: "network.pdcInjected", key: "pdcInjected",
            label: "Private Data Source Connect",
            valueType: "boolean", target: "jsonData",
        },
    ]
}

/**
 * Standard custom HTTP headers field with indexedPair storage mapping.
 */
export function httpHeaderFields(): ConfigField[] {
    return [
        {
            id: "httpHeaders", key: "httpHeaders",
            label: "Custom HTTP Headers", description: "Additional headers sent with every request",
            valueType: "array", target: "jsonData",
            item: {
                valueType: "object",
                fields: [
                    { id: "httpHeaders.item.name", key: "name", label: "Header Name", valueType: "string", isItemField: true },
                    { id: "httpHeaders.item.value", key: "value", label: "Header Value", valueType: "string", isItemField: true },
                ],
            },
            storage: {
                type: "indexedPair",
                key: { target: "jsonData", pattern: "httpHeaderName{index}" },
                value: { target: "secureJsonData", pattern: "httpHeaderValue{index}" },
            },
            validations: [{ type: "itemCount", max: 10, message: "Maximum 10 custom headers" }],
        },
    ]
}
