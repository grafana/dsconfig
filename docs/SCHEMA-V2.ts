/**
 * dsconfig — v2 additions.
 *
 * IMPLEMENTATION NOTE — READ THIS FIRST
 *
 * This file adds v2's new capabilities (scopes, role, roleConflicts,
 * pairRole) without modifying SCHEMA-V1.ts's ConfigField interface in any
 * way. This mirrors SCHEMA-V2.go's Go implementation exactly, including
 * its central caveat, which applies here for an analogous (though not
 * identical) reason: while TypeScript interfaces, unlike Go structs,
 * CAN be extended via declaration merging or intersection types without
 * touching the original file, doing so here would still mean every
 * existing function in SCHEMA-V1.ts (fieldById, valueById, validateSchema,
 * etc.) continues to operate on the plain ConfigField shape and would
 * not see the new properties unless every such function were also
 * duplicated against an extended type — which is the same practical
 * cascade the Go side avoids by using a side table. For consistency
 * between the two language implementations — a property this package
 * has maintained since v1 specifically so a schema author never gets
 * different answers depending on which language validated a document —
 * this file uses the identical side-table approach: SchemaV2Extensions,
 * keyed by the same ConfigField.id v1 already treats as the stable,
 * canonical reference.
 *
 * The intended upstream shape (what a real v2 release should ship) is
 * for scopes, role, roleConflicts, and pairRole to become four new,
 * optional properties declared directly on ConfigField in SCHEMA-V1.ts,
 * exactly as SCHEMA-V2.go's implementation note describes for the Go
 * side. This file is a faithful, behavior-preserving stand-in for that
 * shape, produced under the constraint of not modifying SCHEMA-V1.ts.
 *
 * WHAT v2 ADDS, AND WHY
 *
 *  1. Scopes — closes v1 Known Limitation 7 (no representation for a
 *     plugin with more than one independent connection within a single
 *     datasource instance).
 *  2. Role and roleConflicts — closes v1 Known Limitation 5 (no field
 *     carries a semantic role independent of its name).
 *  3. pairRole — closes v1 Known Limitations 13 and 14
 *     (resolveIndexedPairs/resolveIndexedPairsAsMap infer pair-role by
 *     declaration position).
 *
 * See SCHEMA-V2.go for the full design rationale; this file mirrors it
 * property-for-property and function-for-function.
 */

import type { Schema, ConfigField, TargetLocation } from './schema';

// ============================================================
// Schema Version
// ============================================================

/**
 * The value Schema.schemaVersion should carry for a document that uses
 * any v2-only capability (scopes, role, roleConflicts, or pairRole, via
 * the SchemaV2Extensions side table below). A document using none of
 * these may continue to declare 'v1' and is completely unaffected by
 * anything in this module.
 */
export const SCHEMA_VERSION_V2 = 'v2';

// ============================================================
// Side-Table Extension Model
// ============================================================

/**
 * Carries every piece of v2 metadata for one Schema document, keyed by
 * ConfigField.id. See the IMPLEMENTATION NOTE at the top of this file
 * for why this is a side table rather than properties directly on
 * ConfigField.
 */
export interface SchemaV2Extensions {
    /**
     * Declares every named scope this plugin's fields may belong to. A
     * document with no multi-connection needs leaves this empty; every
     * field is then implicitly shared across the document's single,
     * implicit scope, exactly as in v1.
     */
    scopeDefs?: ScopeDef[];

    /**
     * Maps a ConfigField.id to that field's v2 extension. A field with no
     * entry here has no scopes, no role, no roleConflicts, and (if it is
     * an item field) no pairRole — it behaves exactly as it did under v1.
     */
    fields?: Record<string, FieldExtensionV2>;
}

/**
 * One field's v2-only metadata. In the upstream shape this module
 * stands in for, every property here would instead be a property
 * directly on ConfigField.
 */
export interface FieldExtensionV2 {
    /**
     * This field's scope membership. Empty/absent means the field is
     * shared across every scope declared in the owning
     * SchemaV2Extensions.scopeDefs, including any scope declared later.
     */
    scopes?: string[];

    /** This field's semantic meaning, drawn from the Role vocabulary below. */
    role?: Role;

    /** Roles that must not be simultaneously active alongside this field's own role. */
    roleConflicts?: Role[];

    /**
     * Only meaningful for an item field nested inside an indexedPair
     * field's item.fields. Absent means positional inference applies,
     * exactly as under v1.
     */
    pairRole?: PairRole;
}

// ============================================================
// Scopes
// ============================================================

/**
 * Declares one named scope a multi-connection plugin's fields may
 * belong to — for example, a plugin with a "controller" API and a
 * separate "eum" API declares two ScopeDefs, one per connection.
 */
export interface ScopeDef {
    /** This scope's unique identifier within the schema document. */
    id: string;
    /** Human-readable name for this scope, for UI and documentation purposes. */
    label: string;
}

/**
 * Checks that every declared ScopeDef has a non-empty id, and that no
 * two ScopeDefs share the same id. Mirrors SCHEMA-V2.go's
 * validateScopeDefs exactly, collecting all errors per this module's
 * established all-errors-collected convention (see SCHEMA-V1.ts).
 */
export function validateScopeDefs(scopes: ScopeDef[]): string[] {
    const errors: string[] = [];
    const seen = new Set<string>();
    scopes.forEach((s, i) => {
        if (!s.id) {
            errors.push(`scopeDefs[${i}]: id is required`);
            return;
        }
        if (seen.has(s.id)) {
            errors.push(`scopeDefs[${i}]: duplicate scope id "${s.id}"`);
        }
        seen.add(s.id);
    });
    return errors;
}

/**
 * Checks one field's extension scopes against the schema's declared
 * ScopeDefs. Mirrors SCHEMA-V2.go's validateFieldScopes exactly,
 * including the no-full-list rule (see that function's documentation
 * for why listing every currently-declared scope id is rejected rather
 * than treated as equivalent to omission).
 */
export function validateFieldScopes(fieldId: string, scopes: string[] | undefined, scopeIds: Set<string>): string[] {
    const errors: string[] = [];
    if (!scopes || scopes.length === 0) {
        return errors;
    }
    if (scopeIds.size === 0) {
        errors.push(`field ${fieldId}: scopes is set but the schema declares no scopeDefs`);
        return errors;
    }

    const seen = new Set<string>();
    for (const ref of scopes) {
        if (!scopeIds.has(ref)) {
            errors.push(`field ${fieldId}: scopes references unknown scope id: ${ref}`);
        }
        if (seen.has(ref)) {
            errors.push(`field ${fieldId}: scopes contains duplicate scope id: ${ref}`);
        }
        seen.add(ref);
    }

    if (scopes.length === scopeIds.size) {
        errors.push(
            `field ${fieldId}: scopes lists every declared scope id; omit scopes entirely instead, which means "all scopes including any declared later" — listing them all explicitly does not have that property and is rejected to prevent the two spellings silently diverging`
        );
    }

    return errors;
}

/**
 * Returns the set of scope ids a field with the given extension scopes
 * belongs to, given the schema's full set of declared scope ids. Mirrors
 * SCHEMA-V2.go's EffectiveScopeIDs exactly.
 */
export function effectiveScopeIds(scopes: string[] | undefined, allScopeIds: string[]): string[] {
    if (!scopes || scopes.length === 0) {
        return allScopeIds;
    }
    return scopes;
}

/**
 * Returns every field in schema whose effective scope set includes
 * scopeId. Throws if scopeId does not reference a declared ScopeDef in
 * ext.scopeDefs. Mirrors SCHEMA-V2.go's FieldsForScope exactly.
 */
export function fieldsForScope(schema: Schema, ext: SchemaV2Extensions, scopeId: string): ConfigField[] {
    const known = (ext.scopeDefs ?? []).some((sc) => sc.id === scopeId);
    if (!known) {
        throw new Error(`no declared scope with id "${scopeId}"`);
    }

    const result: ConfigField[] = [];
    for (const f of schema.fields) {
        const fx = ext.fields?.[f.id] ?? {};
        if (!fx.scopes || fx.scopes.length === 0) {
            result.push(f);
            continue;
        }
        if (fx.scopes.includes(scopeId)) {
            result.push(f);
        }
    }
    return result;
}

// ============================================================
// Semantic Roles
// ============================================================

/**
 * A field's semantic meaning, independent of its key or id. Drawn from
 * the fixed vocabulary below so that a consumer — most concretely, a
 * future HTTP client builder, or Grafana Assistant reasoning about a
 * field it has never seen before — can recognize "this field is the TLS
 * client certificate" without hard-coding per-plugin field name lists.
 */
export type Role =
    | 'endpoint.baseUrl'
    | 'transport.timeoutSeconds'
    | 'transport.tlsSkipVerify'
    | 'tls.clientCert'
    | 'tls.clientKey'
    | 'tls.caCert'
    | 'tls.serverName'
    | 'auth.discriminator'
    | 'auth.basic.enabled'
    | 'auth.basic.username'
    | 'auth.basic.password'
    | 'auth.oauth2.clientId'
    | 'auth.oauth2.clientSecret'
    | 'auth.jwt.signingKey'
    | 'auth.awsSigV4.enabled'
    | 'auth.awsSigV4.accessKey'
    | 'auth.awsSigV4.secretKey'
    | 'identity.forwardOAuthToken'
    | 'http.header.name'
    | 'http.header.value';

/**
 * The complete, closed set of valid Role values for this version of the
 * module. Mirrors SCHEMA-V2.go's knownRoles exactly.
 */
const KNOWN_ROLES: ReadonlySet<string> = new Set<Role>([
    'endpoint.baseUrl',
    'transport.timeoutSeconds',
    'transport.tlsSkipVerify',
    'tls.clientCert',
    'tls.clientKey',
    'tls.caCert',
    'tls.serverName',
    'auth.discriminator',
    'auth.basic.enabled',
    'auth.basic.username',
    'auth.basic.password',
    'auth.oauth2.clientId',
    'auth.oauth2.clientSecret',
    'auth.jwt.signingKey',
    'auth.awsSigV4.enabled',
    'auth.awsSigV4.accessKey',
    'auth.awsSigV4.secretKey',
    'identity.forwardOAuthToken',
    'http.header.name',
    'http.header.value',
]);

/** Reports whether r is a member of this version's known role vocabulary. */
export function isValidRole(r: string): r is Role {
    return KNOWN_ROLES.has(r);
}

/**
 * Checks one field's extension role and roleConflicts. Mirrors
 * SCHEMA-V2.go's validateFieldRole exactly.
 */
export function validateFieldRole(fieldId: string, fx: FieldExtensionV2): string[] {
    const errors: string[] = [];
    if (fx.role && !isValidRole(fx.role)) {
        errors.push(`field ${fieldId}: unknown role "${fx.role}"`);
    }
    for (const rc of fx.roleConflicts ?? []) {
        if (!isValidRole(rc)) {
            errors.push(`field ${fieldId}: roleConflicts references unknown role "${rc}"`);
        }
        if (fx.role && rc === fx.role) {
            errors.push(`field ${fieldId}: roleConflicts lists its own role "${fx.role}", which is meaningless`);
        }
    }
    return errors;
}

/**
 * Checks every declared roleConflicts relationship for structural
 * consistency across one effective field set. Mirrors SCHEMA-V2.go's
 * ValidateRoleConflicts exactly.
 */
export function validateRoleConflicts(fields: ConfigField[], ext: SchemaV2Extensions): string[] {
    const errors: string[] = [];
    const roleOwner = new Map<Role, string>();

    for (const f of fields) {
        const fx = ext.fields?.[f.id] ?? {};
        if (!fx.role) continue;
        const existing = roleOwner.get(fx.role);
        if (existing) {
            errors.push(
                `role "${fx.role}" is carried by both field ${existing} and field ${f.id} within the same effective scope; a role must be unique within any one effective field set`
            );
            continue;
        }
        roleOwner.set(fx.role, f.id);
    }

    for (const f of fields) {
        const fx = ext.fields?.[f.id] ?? {};
        for (const rc of fx.roleConflicts ?? []) {
            const owner = roleOwner.get(rc);
            if (owner) {
                errors.push(
                    `field ${f.id} declares a conflict with role "${rc}", which field ${owner} carries within the same effective scope; both cannot be present together by this schema's own declaration`
                );
            }
        }
    }

    return errors;
}

// ============================================================
// Indexed-Pair Item Role
// ============================================================

/**
 * Declares, for an item field used inside an indexedPair field's
 * item.fields, whether that item field is the pair's "key" (name) half
 * or its "value" half.
 */
export type PairRole = 'key' | 'value';

/** Reports whether p is a valid PairRole value (or undefined). */
export function isValidPairRole(p: string | undefined): p is PairRole | undefined {
    return p === undefined || p === 'key' || p === 'value';
}

/**
 * Checks that, among a set of item field ids, at most one item field's
 * extension declares pairRole 'key' and at most one declares 'value'.
 * Mirrors SCHEMA-V2.go's validateItemPairRoles exactly.
 */
export function validateItemPairRoles(itemFieldIds: string[], ext: SchemaV2Extensions): string[] {
    const errors: string[] = [];
    let keyOwner: string | undefined;
    let valueOwner: string | undefined;

    for (const id of itemFieldIds) {
        const fx = ext.fields?.[id] ?? {};
        if (!fx.pairRole) continue;
        if (!isValidPairRole(fx.pairRole)) {
            errors.push(`item field ${id}: invalid pairRole "${fx.pairRole}"`);
            continue;
        }
        if (fx.pairRole === 'key') {
            if (keyOwner) {
                errors.push(`item field ${id}: pairRole "key" already claimed by item field ${keyOwner}`);
            } else {
                keyOwner = id;
            }
        } else if (fx.pairRole === 'value') {
            if (valueOwner) {
                errors.push(`item field ${id}: pairRole "value" already claimed by item field ${valueOwner}`);
            } else {
                valueOwner = id;
            }
        }
    }

    return errors;
}

/**
 * Returns the item-field key for the pair's name side and the item-field
 * key for its value side, preferring explicit pairRole extensions and
 * falling back to v1's positional inference when neither item field's
 * extension declares pairRole. Mirrors SCHEMA-V2.go's
 * resolvePairRoleKeys exactly.
 */
export function resolvePairRoleKeys(
    itemFields: ConfigField[],
    ext: SchemaV2Extensions
): { nameKey: string; valueKey: string } {
    if (itemFields.length < 2) {
        throw new Error('indexedPair requires an item schema with at least 2 fields');
    }

    let keyField: ConfigField | undefined;
    let valueField: ConfigField | undefined;
    for (const f of itemFields) {
        const fx = ext.fields?.[f.id] ?? {};
        if (fx.pairRole === 'key') keyField = f;
        if (fx.pairRole === 'value') valueField = f;
    }

    if (keyField && valueField) {
        return { nameKey: keyField.key, valueKey: valueField.key };
    }
    if (keyField || valueField) {
        throw new Error('indexedPair item schema declares pairRole on only one item field; declare it on both or on neither');
    }

    // Neither item field declares pairRole: fall back to v1's positional
    // inference, exactly as resolveIndexedPairs/resolveIndexedPairsAsMap
    // already do.
    return { nameKey: itemFields[0].key, valueKey: itemFields[1].key };
}

// ============================================================
// PairRole-Aware Indexed-Pair Resolution
// ============================================================

function bucketForTargetV2(
    t: TargetLocation,
    jsonData: Record<string, unknown>,
    secureJsonData: Record<string, unknown>
): Record<string, unknown> {
    if (t === 'jsonData') return jsonData;
    if (t === 'secureJsonData') return secureJsonData;
    throw new Error(`indexedPair target "${t}" must be jsonData or secureJsonData`);
}

function patternToIndexRegexV2(pattern: string): RegExp {
    const escaped = pattern.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    const withCapture = escaped.replace('\\{index\\}', '(\\d+)');
    return new RegExp(`^${withCapture}$`);
}

/**
 * Identical to v1's resolveIndexedPairs except that it resolves which
 * item field is the pair's name and which is its value via
 * resolvePairRoleKeys (pairRole-aware, falling back to positional
 * inference) rather than v1's purely positional logic. Mirrors
 * SCHEMA-V2.go's ResolveIndexedPairsV2 exactly.
 */
export function resolveIndexedPairsV2(
    f: ConfigField,
    ext: SchemaV2Extensions,
    jsonData: Record<string, unknown>,
    secureJsonData: Record<string, unknown>
): Array<Record<string, unknown>> {
    if (f.storage?.type !== 'indexedPair') {
        throw new Error(`field "${f.id}" is not an indexedPair field`);
    }
    const mapping = f.storage;

    const keyBucket = bucketForTargetV2(mapping.key!.target, jsonData, secureJsonData);
    const valueBucket = bucketForTargetV2(mapping.value!.target, jsonData, secureJsonData);

    const start = mapping.startIndex ?? 1;

    if (!f.item?.fields) {
        throw new Error(`field "${f.id}": indexedPair requires an item schema`);
    }
    const { nameKey: nameFieldKey, valueKey: valueFieldKey } = resolvePairRoleKeys(f.item.fields, ext);

    const results: Array<Record<string, unknown>> = [];
    for (let i = start; ; i++) {
        const nameKey = mapping.key!.pattern.replace('{index}', String(i));
        if (!(nameKey in keyBucket)) break;

        const item: Record<string, unknown> = { [nameFieldKey]: keyBucket[nameKey] };

        const valueKey = mapping.value!.pattern.replace('{index}', String(i));
        if (valueKey in valueBucket) {
            item[valueFieldKey] = valueBucket[valueKey];
        }

        results.push(item);
    }

    return results;
}

/**
 * Identical to v1's resolveIndexedPairsAsMap except that — like
 * resolveIndexedPairsV2 — it is pairRole-aware rather than purely
 * positional. Mirrors SCHEMA-V2.go's ResolveIndexedPairsAsMapV2 exactly.
 */
export function resolveIndexedPairsAsMapV2(
    f: ConfigField,
    ext: SchemaV2Extensions,
    jsonData: Record<string, unknown>,
    secureJsonData: Record<string, unknown>
): Record<string, string> {
    if (f.storage?.type !== 'indexedPair') {
        throw new Error(`field "${f.id}" is not an indexedPair field`);
    }
    const mapping = f.storage;

    const keyBucket = bucketForTargetV2(mapping.key!.target, jsonData, secureJsonData);
    const valueBucket = bucketForTargetV2(mapping.value!.target, jsonData, secureJsonData);

    const keyRe = patternToIndexRegexV2(mapping.key!.pattern);

    if (!f.item?.fields) {
        throw new Error(`field "${f.id}": indexedPair requires an item schema`);
    }
    // Validate pairRole declarations even though, for the map-shaped
    // output, only the stored *values* end up in the result — matching
    // SCHEMA-V2.go's ResolveIndexedPairsAsMapV2 exactly.
    resolvePairRoleKeys(f.item.fields, ext);

    const result: Record<string, string> = {};
    for (const [storedKey, storedVal] of Object.entries(keyBucket)) {
        const m = keyRe.exec(storedKey);
        if (!m) continue;
        const index = m[1];

        if (typeof storedVal !== 'string') continue;
        const name = storedVal;

        const valueKey = mapping.value!.pattern.replace('{index}', index);
        const rawValue = valueBucket[valueKey];
        const value = typeof rawValue === 'string' ? rawValue : '';

        result[name] = value;
    }

    return result;
}

// ============================================================
// v2 Structural Validation Entry Point
// ============================================================

/**
 * Performs every v2-specific structural check in addition to — not
 * instead of — v1's validateSchema. A v2 schema document should be
 * validated by calling both: first validateSchema(schema) (unchanged
 * from v1), then validateV2(schema, ext). Mirrors SCHEMA-V2.go's
 * ValidateV2 exactly, including its all-errors-collected behavior
 * (consistent with this module's existing convention, distinct from the
 * Go side's first-error-only behavior — see SCHEMA-V1.ts's own "KNOWN
 * LIMITATIONS" item on this difference, which applies identically here).
 */
export function validateV2(schema: Schema, ext: SchemaV2Extensions): string[] {
    const errors: string[] = [];

    errors.push(...validateScopeDefs(ext.scopeDefs ?? []));

    const scopeIds = new Set((ext.scopeDefs ?? []).map((sc) => sc.id));

    function visit(fields: ConfigField[], itemFieldIdsOfIndexedPairParent: string[]): void {
        for (const f of fields) {
            const fx = ext.fields?.[f.id] ?? {};

            errors.push(...validateFieldScopes(f.id, fx.scopes, scopeIds));
            errors.push(...validateFieldRole(f.id, fx));

            const isDeclaredIndexedPairItem = itemFieldIdsOfIndexedPairParent.includes(f.id);
            if (!isDeclaredIndexedPairItem && fx.pairRole) {
                errors.push(`field ${f.id}: pairRole is only meaningful on an item field of an indexedPair-mapped field`);
            }

            if (f.item?.fields) {
                const isIndexedPair = f.storage?.type === 'indexedPair';
                let childItemIds: string[] = [];
                if (isIndexedPair) {
                    childItemIds = f.item.fields.map((c) => c.id);
                    errors.push(...validateItemPairRoles(childItemIds, ext));
                }
                visit(f.item.fields, childItemIds);
            }
        }
    }
    visit(schema.fields, []);

    if (!ext.scopeDefs || ext.scopeDefs.length === 0) {
        errors.push(...validateRoleConflicts(schema.fields, ext));
    } else {
        const ids = [...scopeIds].sort();
        for (const id of ids) {
            try {
                const fields = fieldsForScope(schema, ext, id);
                errors.push(...validateRoleConflicts(fields, ext).map((e) => `scope "${id}": ${e}`));
            } catch (e) {
                errors.push((e as Error).message);
            }
        }
    }

    return errors;
}

// ============================================================
// KNOWN LIMITATIONS (v2)
// ============================================================
//
// This section is additional to, not a replacement for, v1's KNOWN
// LIMITATIONS (SCHEMA-V1.ts). Every v1 limitation that v2 does not close
// remains exactly as documented there.
//
//  1. This module represents scopes/role/roleConflicts/pairRole as a
//     side table (SchemaV2Extensions, FieldExtensionV2) keyed by
//     ConfigField.id, rather than as properties directly on
//     ConfigField, as a deliberate constraint of this deliverable (see
//     the IMPLEMENTATION NOTE at the top of this file). A schema author
//     using this exact reference implementation must keep a
//     SchemaV2Extensions document in sync with its corresponding Schema
//     document by id, by hand or by tooling.
//  2. The OpenAPI settings conversion on the Go side does not produce a
//     multi-connection-aware App Platform artifact. A v2 schema plus its
//     extensions can fully describe a multi-connection plugin, but the
//     conversion still emits exactly one Settings{spec, secureValues}
//     pair per document.
//  3. roleConflicts is structurally validated but not evaluated against
//     any real configuration payload.
//  4. A role being valid (isValidRole) only means it is a member of this
//     version's known vocabulary; it does not mean the field's valueType
//     or target is appropriate for that role.
//  5. The known-role vocabulary is fixed by this module's version and is
//     not extensible by a schema author.
//  6. pairRole resolution requires that if either item field declares
//     pairRole, the other must too; declaring it on only one of the two
//     is rejected as ambiguous rather than silently falling back to
//     positional inference for the other half.
//  7. v1's resolveIndexedPairs and resolveIndexedPairsAsMap (SCHEMA-V1.ts)
//     are completely unaffected by this module and continue to use
//     purely positional inference, with no awareness of pairRole.
