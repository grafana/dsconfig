/**
 * dsconfig — declarative configuration schema for Grafana datasource
 * plugins (TypeScript mirror).
 *
 * PURPOSE
 *
 * This module mirrors, field for field, the Go types defined in
 * SCHEMA-V1.go. It exists so that the same Schema document — one JSON file
 * per plugin, see SCHEMA-V1.json for the wire format — can be consumed by
 * TypeScript code (a config editor renderer, a frontend-side validator,
 * or any other browser/Node consumer) with the identical shape and the
 * identical semantics as the Go side. Any divergence between this file
 * and SCHEMA-V1.go is a defect: both describe one schema, not two.
 *
 * dsconfig is a semantic description layer placed on top of Grafana's
 * existing datasource configuration model (root fields, jsonData,
 * secureJsonData). It does not change how Grafana stores datasource
 * configuration today. Every field described by a Schema still lives
 * exactly where it lives now. dsconfig is additive: a plugin can adopt
 * it without migrating, renaming, or moving a single stored value.
 *
 * dsconfig exists to serve four consumers that today have no shared,
 * machine-readable contract to work from:
 *
 *   1. Config editors (frontend forms) — today hand-written per plugin,
 *      with no guarantee they match what the backend actually parses.
 *   2. Backend settings parsing (grafana-plugin-sdk-go) — today reads
 *      untyped jsonData/secureJsonData maps with no schema-level
 *      contract.
 *   3. Provisioning (datasources.yaml) and the Grafana App Platform /
 *      Kubernetes-style datasource API — both need a description of a
 *      valid datasource config that exists independently of any
 *      running plugin instance, so a config can be validated before it
 *      is applied.
 *   4. Automated and assisted configuration — including the Grafana
 *      Assistant chat-driven datasource workflow, and any other tool
 *      that needs to read, generate, or validate a datasource
 *      configuration without parsing plugin source code.
 *
 * WHY THIS MODULE EXISTS: TWO PRIMARY DRIVERS
 *
 * Driver 1 — App Platform / Kubernetes-style API compatibility.
 * Grafana's App Platform exposes resources through a Kubernetes-style
 * API (CRD-shaped: apiVersion, kind, metadata, spec, status). A
 * datasource's spec needs an OpenAPI-shaped schema describing what a
 * valid instance of that resource looks like, the same way any
 * Kubernetes Custom Resource Definition needs a structural schema for
 * its spec. dsconfig is the semantic layer that produces that
 * OpenAPI-shaped schema (see the Go side's ToPluginSchemaSettings) from
 * one declarative source per plugin. This TypeScript mirror lets
 * frontend tooling — including a generated or assisted config editor —
 * reason about the same structural schema App Platform's CRUD and
 * admission-validation machinery uses, without re-deriving it.
 *
 * Driver 2 — reliable, automatically derived HTTP clients.
 * Today, the logic that turns a plugin's stored configuration into a
 * working HTTP client (TLS setup, auth header/round-tripper wiring,
 * timeout configuration) is hand-written, per plugin, and frequently
 * duplicated with small inconsistencies across otherwise similar
 * plugins. Because Schema fields carry typed, structured metadata
 * rather than being opaque map entries, the same schema that drives
 * config-editor generation and provisioning validation is also the
 * input a future helper can use to build a transport-correct,
 * auth-correct HTTP client without per-plugin code. This module's
 * current version does not implement that derivation; see "KNOWN
 * LIMITATIONS" at the bottom of this file. The schema is shaped, from
 * this version forward, so that derivation is possible without a
 * breaking change to the fields already defined here.
 *
 * DESIGN POSTURE: ADDITIVE, NOT MIGRATORY
 *
 * Every design decision in this module follows one rule: adopting
 * dsconfig must never require a plugin to change what it stores, how
 * it stores it, or where. The schema describes root, jsonData, and
 * secureJsonData exactly as Grafana persists them today. See SCHEMA-V1.go
 * for the full discussion; this file makes the identical guarantee.
 */

// ============================================================
// Root Schema
// ============================================================

/**
 * Schema is the top-level schema definition.
 * It acts as the single source of truth for datasource configuration.
 *
 * One Schema value describes exactly one plugin type's configuration
 * surface: every field stored in that plugin's root-level datasource
 * properties, jsonData, and secureJsonData.
 *
 * Schema is consumed in at least three independent ways, and every
 * field below should be read with all three consumers in mind:
 *
 *   - Structurally, by validateSchema(), which checks that the schema
 *     document itself is internally well-formed (every reference
 *     resolves, every field is shaped correctly for its kind). This is
 *     schema-authoring-time validation; it says nothing about whether a
 *     particular stored datasource config is valid.
 *   - As an OpenAPI-shaped settings document, via the Go side's
 *     ToPluginSchemaSettings, which is the shape grafana-plugin-sdk-go
 *     and the Grafana App Platform / Kubernetes-style datasource API
 *     expect for describing and validating an instance of this
 *     plugin's configuration (its CRD-style spec). This TypeScript
 *     mirror does not itself produce that OpenAPI document; see "KNOWN
 *     LIMITATIONS".
 *   - As a description for UI and automation consumers — config editor
 *     generation, documentation generation, and chat-driven
 *     configuration assistants (such as the Grafana Assistant) that
 *     need to know what fields exist, what they mean, and what values
 *     are valid, in order to create or repair a working datasource
 *     configuration without hand-coded per-plugin logic.
 */
export interface Schema {
    /**
     * SchemaVersion defines the version of the schema spec.
     *
     * This versions the dsconfig schema *format* itself (the shape of
     * Schema/ConfigField/etc. as defined by this module), not the
     * plugin's own config version. See "KNOWN LIMITATIONS" for the
     * current state of cross-version handling.
     */
    schemaVersion: string;

    /**
     * PluginType uniquely identifies the datasource plugin.
     *
     * Must match the plugin's own type identifier (the same identifier
     * Grafana's plugin system and the datasource's "type" property
     * already use). This is the join key between a dsconfig Schema
     * document and a real, running plugin instance, and the key an App
     * Platform resource's apiVersion/kind would reference when locating
     * the structural schema for a given datasource kind.
     */
    pluginType: string;

    /**
     * PluginName is a human-readable name.
     *
     * Display-only. Not used as a reference key anywhere in this module;
     * pluginType is the identifier for that purpose.
     */
    pluginName: string;

    /** Optional documentation URL for the plugin as a whole. */
    docURL?: string;

    /**
     * Fields defines all configuration fields.
     *
     * This is the source of truth. Every other piece of schema-level
     * metadata — groups, relationships, instructions — is descriptive
     * metadata layered on top of fields, and is validated against the
     * field IDs declared here (see validateRefs). fields is required and
     * must be non-empty.
     */
    fields: ConfigField[];

    /**
     * Optional UI grouping.
     *
     * Groups are presentation metadata only. They describe how a config
     * editor might lay fields out into sections (for example,
     * "Connection", "Authentication", "Advanced"); they have no effect on
     * storage, validation, or the OpenAPI settings produced on the Go
     * side. A Schema with no groups is fully valid.
     */
    groups?: ConfigGroup[];

    /**
     * Optional Instruction list.
     *
     * Free-form, structured guidance intended for non-human or
     * semi-autonomous consumers of the schema — most directly, a
     * chat-driven configuration assistant that needs plugin-specific
     * guidance beyond what individual field descriptions convey.
     * Instructions have no effect on validation or storage.
     */
    instructions?: Instruction[];

    /**
     * Relationships between fields.
     *
     * Semantic, not structural: records that two or more fields are
     * conceptually connected, for the benefit of a renderer or assistant
     * deciding how to present or reason about related fields together.
     */
    relationships?: FieldRelationship[];
}

/**
 * validateSchema checks that a Schema document is internally
 * well-formed.
 *
 * This is schema-authoring-time validation. It confirms the document
 * itself is structurally sound — every required top-level property is
 * present, every field is individually valid for its declared kind,
 * and every cross-reference (group field refs, relationship field
 * refs, effect target IDs) resolves to a real field ID. It does not
 * validate any actual stored datasource configuration against this
 * schema; that is a distinct, presently unimplemented capability — see
 * "KNOWN LIMITATIONS" below.
 *
 * Unlike the Go side's Validate (which returns on the first error
 * encountered), validateSchema collects every validation failure found
 * in the document before returning, and returns the full list. This
 * difference in error-collection behavior between the two languages is
 * itself recorded under "KNOWN LIMITATIONS" as something a single
 * shared validation contract should eventually resolve.
 */
export function validateSchema(schema: Schema): string[] {
    const errors: string[] = [];

    if (!schema.schemaVersion) {
        errors.push('schemaVersion is required');
    }
    if (!schema.pluginType) {
        errors.push('pluginType is required');
    }
    if (!schema.pluginName) {
        errors.push('pluginName is required');
    }
    if (!schema.fields || schema.fields.length === 0) {
        errors.push('fields is required');
    }

    for (const field of schema.fields ?? []) {
        errors.push(...validateConfigField(field));
    }

    let fieldIds: Set<string>;
    try {
        fieldIds = collectFieldIds(schema);
    } catch (e) {
        errors.push((e as Error).message);
        return errors;
    }

    errors.push(...validateFieldIdFormat(fieldIds));
    errors.push(...validateRefs(schema, fieldIds));

    return errors;
}

/**
 * idSegmentPattern matches a single dot-separated segment of a field id.
 * This is deliberately the same character class a CEL-style identifier
 * accepts: ASCII letters, digits, and underscore, not starting with a
 * digit. Mirrors SCHEMA-V1.go's idSegmentPattern exactly.
 */
const idSegmentPattern = /^[A-Za-z_][A-Za-z0-9_]*$/;

/**
 * validateFieldIdFormat checks every field id in the given set against
 * two rules: each dot-separated segment must match idSegmentPattern, and
 * no id may be a strict dotted-path prefix of another id in the same
 * set. Mirrors SCHEMA-V1.go's ValidateFieldIDFormat exactly, except that —
 * consistent with this module's all-errors-collected convention — it
 * returns every violation found rather than stopping at the first.
 */
export function validateFieldIdFormat(fieldIds: Set<string>): string[] {
    const errors: string[] = [];
    const all = Array.from(fieldIds);

    for (const id of all) {
        for (const seg of id.split('.')) {
            if (!idSegmentPattern.test(seg)) {
                errors.push(`field id "${id}": segment "${seg}" is not a valid identifier (must match ${idSegmentPattern})`);
            }
        }
    }

    const sorted = [...all].sort();
    for (let i = 0; i + 1 < sorted.length; i++) {
        const a = sorted[i];
        const b = sorted[i + 1];
        if (b.startsWith(a + '.')) {
            errors.push(
                `field id "${a}" is a dotted-path prefix of field id "${b}"; this is ambiguous for any future consumer that resolves ids as dotted paths (e.g. an expression evaluator)`
            );
        }
    }

    return errors;
}

/**
 * validateRefs checks that all group and relationship field references
 * point to existing field IDs.
 *
 * fieldIds is the complete set of field IDs declared anywhere in the
 * schema (top-level fields and, recursively, item fields of array/map
 * fields), as produced by collectFieldIds. validateRefs checks three
 * categories of reference, all keyed by field ID rather than by storage
 * key, consistent with this module's id/key separation (see
 * ConfigField): ConfigGroup.fieldRefs, FieldRelationship.fields, and
 * FieldEffect.set keys.
 */
export function validateRefs(schema: Schema, fieldIds: Set<string>): string[] {
    const errors: string[] = [];

    for (const g of schema.groups ?? []) {
        for (const ref of g.fieldRefs) {
            if (!fieldIds.has(ref)) {
                errors.push(`group ${g.id} references unknown field id: ${ref}`);
            }
        }
    }

    for (const r of schema.relationships ?? []) {
        if (!isValidRelationshipType(r.type)) {
            errors.push(`relationship has invalid type "${r.type}"`);
        }
        for (const ref of r.fields) {
            if (!fieldIds.has(ref)) {
                errors.push(`relationship references unknown field id: ${ref}`);
            }
        }
    }

    // Validate effect set keys reference known field IDs
    const visitEffects = (fields: ConfigField[]): void => {
        for (const f of fields) {
            (f.effects ?? []).forEach((eff, i) => {
                for (const ref of Object.keys(eff.set ?? {})) {
                    if (!fieldIds.has(ref)) {
                        errors.push(`field ${f.id}: effect[${i}].set references unknown field id: ${ref}`);
                    }
                }
            });
            if (f.item) {
                visitEffects(f.item.fields ?? []);
            }
        }
    };
    visitEffects(schema.fields);

    return errors;
}

// ============================================================
// Field Definition
// ============================================================

/**
 * ConfigField represents a single configuration field.
 *
 * ConfigField is the unit of description for everything this module
 * models: a piece of data that is either stored somewhere in Grafana's
 * existing datasource config model (root, jsonData, or secureJsonData),
 * or computed/virtual and not stored at all. Every other type in this
 * module exists to describe some aspect of a ConfigField in more
 * detail (its UI presentation, its validation rules, its storage
 * mapping, and so on).
 *
 * ID VERSUS KEY
 *
 * ConfigField deliberately separates two identifiers that are easy to
 * conflate but serve different purposes:
 *
 *   - id is the field's globally unique name within the schema. It is
 *     the identifier every cross-reference in this module uses:
 *     ConfigGroup.fieldRefs, FieldRelationship.fields, and
 *     FieldEffect.set keys all refer to fields by id. id is a
 *     schema-authoring concern; it never appears in the stored
 *     datasource configuration itself, and changing a field's id does
 *     not change anything about how or where its value is stored. A
 *     recommended, but not currently enforced, convention is a
 *     dot-separated path describing the field's logical position (for
 *     example, "auth.basicAuthPassword"); see "KNOWN LIMITATIONS"
 *     regarding the lack of enforcement.
 *   - key is the field's local name within whatever it is actually
 *     stored in — a property name within root, within jsonData, within
 *     secureJsonData, or within an item object for array/map fields.
 *     key is the identifier that matches what Grafana's existing
 *     storage model, and any existing plugin backend code reading that
 *     storage, already expects. key is never required to be globally
 *     unique; only unique within its immediate storage context.
 *
 * This separation exists specifically to keep the schema additive (see
 * the module-level documentation): key always matches what is already
 * being stored, today, by the plugin, with zero changes, while id gives
 * every other part of this module a stable, storage-independent name to
 * reference.
 *
 * STORAGE TARGET
 *
 * target (when set) declares which of Grafana's three existing storage
 * locations holds this field's value: root-level datasource
 * properties, the jsonData map, or the secureJsonData map. See
 * TargetLocation. secureJsonData remains write-only from the schema's
 * perspective: a field targeting "secureJsonData" describes what may
 * be written, not a value that can be read back. See "KNOWN
 * LIMITATIONS" regarding secureJsonFields (the existing read-side
 * indicator of "is a secret configured") and how it relates to fields
 * described here.
 *
 * FIELD KIND: STORAGE VERSUS VIRTUAL
 *
 * Most fields are storage fields: they have a target and describe a
 * real, persisted value. A field may instead be declared kind:
 * "virtual", meaning it has no target and is not persisted at all — it
 * exists only to describe computed or UI-only state. The canonical use
 * of a virtual field is a selector control (for example, an
 * "Authentication method" dropdown) whose own value is never stored,
 * but whose selection drives the values of one or more real storage
 * fields via effects. See FieldEffect.
 *
 * APP PLATFORM / KUBERNETES-STYLE API RELEVANCE
 *
 * Each storage field, taken together with its valueType, validations,
 * and required/requiredWhen state, supplies exactly the information an
 * OpenAPI-style structural schema needs for one property: type,
 * constraints, and required-ness. This is what the Go side's
 * ToPluginSchemaSettings walks to build the spec consumed by
 * grafana-plugin-sdk-go and, in turn, usable as the structural schema
 * for a Kubernetes-style Custom Resource Definition describing this
 * plugin's datasource spec under Grafana's App Platform.
 */
export interface ConfigField {
    /** ID is globally unique (used for references). */
    id: string;

    /** Key is the local key (used in storage or object structures). */
    key: string;

    label?: string;
    description?: string;
    docURL?: string;

    /** Core typing. */
    valueType: ValueType;

    /** Storage location (required for storage fields). */
    target?: TargetLocation;

    /**
     * Section is the dotted path prefix within the target for nested
     * objects. Example: for jsonData.tracesToLogs.datasourceUid,
     * target="jsonData", section="tracesToLogs", key="datasourceUid".
     */
    section?: string;

    /** Field type: storage (default) or virtual. */
    kind?: FieldKind;

    /** True if part of array item schema. */
    isItemField?: boolean;

    /** UI hints. */
    ui?: FieldUI;

    /** Validation rules. */
    validations?: FieldValidationRule[];

    /**
     * Conditional behavior (CEL).
     *
     * dependsOn, requiredWhen, and disabledWhen are stored as CEL-like
     * expression strings describing a condition over other fields'
     * values. As of this schema version, these strings are validated
     * only for presence where required by a given rule shape — they are
     * not parsed against a grammar and are not evaluated by anything in
     * this module. A schema document can therefore declare a condition
     * with a typo or with a reference to a field that does not exist, and
     * validateSchema will not detect it. See "KNOWN LIMITATIONS" for the
     * current scope of this gap and the structured alternative (effects)
     * used where expressiveness allows.
     */
    dependsOn?: string;
    required?: boolean;
    requiredWhen?: string;
    disabledWhen?: string;

    /** Dynamic overrides. */
    overrides?: FieldOverride[];

    /**
     * effects: declarative multi-field write side-effects.
     * When this field's value matches a condition, the listed target
     * fields are set to the specified values. Typically used on virtual
     * selector fields (e.g. auth method dropdown) to drive multiple
     * storage fields without opaque CEL expressions.
     */
    effects?: FieldEffect[];

    /** Array schema (required when valueType == array). */
    item?: FieldItemSchema;

    /**
     * Legacy indexed fields.
     *
     * repeatable and pattern are reserved for describing legacy,
     * hand-rolled indexed-field conventions (for example, a plugin that
     * stores a numbered series of similarly-named properties without
     * using the structured storage.indexedPair representation below). As
     * of this schema version, neither is read by validateSchema or by
     * anything that produces OpenAPI-shaped settings; see "KNOWN
     * LIMITATIONS".
     */
    repeatable?: boolean;
    pattern?: string;

    /** Storage mapping layer. */
    storage?: StorageMapping;

    /**
     * Metadata.
     *
     * tags, examples, and defaultValue are descriptive metadata with no
     * effect on validation, with one exception: defaultValue is
     * propagated into the generated OpenAPI schema's default value on the
     * Go side. tags is free-text and is intended for documentation and
     * lightweight authoring conventions (for example, recording that a
     * field's value is driven by another field's effects); it is
     * deliberately not validated against any fixed vocabulary and is not
     * intended to gate behavior — see "KNOWN LIMITATIONS". examples is
     * intended for documentation and assisted-configuration use (showing
     * a chat-driven assistant or a generated doc page a representative
     * valid value) and is not currently consumed by any code in this
     * module.
     */
    tags?: string[];
    examples?: unknown[];
    defaultValue?: unknown;
}

/**
 * validateConfigField checks that a single ConfigField is internally
 * well-formed for its declared kind and valueType.
 *
 * This includes: required identifying properties are present (id, key,
 * a valid valueType); a target is present whenever required (storage
 * fields that are neither virtual nor item fields); section is not used
 * in combination with item fields or virtual fields, since neither has
 * a target for section to be a path within; array and map fields
 * declare an item schema; any storage mapping, ui block, validations,
 * overrides, effects, and nested item fields are themselves valid.
 *
 * Unlike the Go side's per-field Validate (which returns on the first
 * error for that field), validateConfigField collects every error found
 * for the field — and, recursively, for its item fields — before
 * returning. See "KNOWN LIMITATIONS" regarding this cross-language
 * difference.
 */
export function validateConfigField(field: ConfigField): string[] {
    const errors: string[] = [];

    if (!field.id) {
        errors.push('field id is required');
        return errors; // no id to prefix subsequent errors with
    }
    if (!field.key) {
        errors.push(`field ${field.id}: key is required`);
    }
    if (!isValidValueType(field.valueType)) {
        errors.push(`field ${field.id}: invalid valueType "${field.valueType}"`);
    }

    const isVirtual = field.kind === 'virtual';
    const isItem = field.isItemField === true;

    if (!isVirtual && !isItem && field.target === undefined) {
        errors.push(`field ${field.id}: target is required for storage fields`);
    }

    if (field.section && isItem) {
        errors.push(`field ${field.id}: section is not allowed on item fields`);
    }
    if (field.section && isVirtual) {
        errors.push(`field ${field.id}: section is not allowed on virtual fields`);
    }

    if ((field.valueType === 'array' || field.valueType === 'map') && !field.item) {
        errors.push(`field ${field.id}: item is required for array and map fields`);
    }

    if (field.storage) {
        errors.push(...validateStorageMapping(field.storage).map((e) => `field ${field.id}: invalid storage mapping: ${e}`));
    }

    if (field.kind && !isValidFieldKind(field.kind)) {
        errors.push(`field ${field.id}: invalid kind "${field.kind}"`);
    }

    if (field.ui) {
        if (!isValidUIComponent(field.ui.component)) {
            errors.push(`field ${field.id}: invalid ui component "${field.ui.component}"`);
        }
        if (field.ui.width && !isValidUIWidth(field.ui.width)) {
            errors.push(`field ${field.id}: invalid ui width "${field.ui.width}"`);
        }
        (field.ui.options ?? []).forEach((opt, i) => {
            if (!validateOptionValue(opt.value, field.valueType)) {
                errors.push(`field ${field.id}: ui option[${i}] value type mismatch (expected ${field.valueType})`);
            }
        });
    }

    if (field.target !== undefined && !isValidTargetLocation(field.target)) {
        errors.push(`field ${field.id}: invalid target: ${field.target}`);
    }

    if (field.item) {
        if (!isValidValueType(field.item.valueType)) {
            errors.push(`field ${field.id}: invalid item valueType "${field.item.valueType}"`);
        }
        if (field.item.valueType !== 'object' && (field.item.fields ?? []).length > 0) {
            errors.push(`field ${field.id}: item fields are only allowed when item valueType is object`);
        }
        for (const sub of field.item.fields ?? []) {
            if (sub.isItemField !== true) {
                errors.push(`field ${field.id}: item field ${sub.id} must have isItemField=true`);
            }
            errors.push(...validateConfigField(sub).map((e) => `field ${field.id}: invalid item field ${sub.id}: ${e}`));
        }
    }

    for (const rule of field.validations ?? []) {
        errors.push(...validateFieldValidationRule(rule).map((e) => `field ${field.id}: invalid validation rule: ${e}`));
    }

    for (const ov of field.overrides ?? []) {
        for (const rule of ov.validations ?? []) {
            errors.push(...validateFieldValidationRule(rule).map((e) => `field ${field.id}: invalid override validation rule: ${e}`));
        }
    }

    (field.effects ?? []).forEach((eff, i) => {
        errors.push(...validateFieldEffect(eff).map((e) => `field ${field.id}: invalid effect[${i}]: ${e}`));
    });

    return errors;
}

/**
 * collectFieldIds walks the schema (including nested item fields of
 * array and map fields) and returns the complete set of declared field
 * IDs.
 *
 * This is the set validateRefs checks every group, relationship, and
 * effect reference against. It also detects duplicate IDs: since id is
 * documented as globally unique (see ConfigField), a duplicate is a
 * schema-authoring error and is reported as such (by throwing) rather
 * than silently overwriting the first occurrence.
 */
export function collectFieldIds(schema: Schema): Set<string> {
    const seen = new Set<string>();

    const visit = (fields: ConfigField[]): void => {
        for (const f of fields) {
            if (!f.id) {
                throw new Error('field id is required');
            }
            if (seen.has(f.id)) {
                throw new Error(`duplicate field id: ${f.id}`);
            }
            seen.add(f.id);

            if (f.item) {
                visit(f.item.fields ?? []);
            }
        }
    };

    visit(schema.fields);

    return seen;
}

/**
 * fieldPath returns the dotted storage path for a field: its target,
 * optionally its section, and its key. A field with no target (a
 * virtual field, or an item field, neither of which is independently
 * stored) returns just its key.
 *
 * This is a convenience for diagnostics and documentation; it is not
 * itself used as a reference key anywhere in this module (id fills that
 * role) and it is not the mechanism that places a field into a
 * generated OpenAPI schema on the Go side.
 */
export function fieldPath(field: ConfigField): string {
    if (field.target === undefined) {
        return field.key;
    }
    if (field.section) {
        return `${field.target}.${field.section}.${field.key}`;
    }
    return `${field.target}.${field.key}`;
}

// ============================================================
// Lookup and Value Resolution by ID
// ============================================================
//
// The functions in this section are read-side utilities, not part of
// schema authoring or structural validation. They exist because every
// consumer that needs to go from "a field id" to "that field's
// definition" or "that field's actual configured value" — a config
// editor resolving Schema.groups' fieldRefs, Grafana Assistant resolving
// an id a person referred to in conversation, a future runtime validator
// reading a real instance settings payload — would otherwise have to
// hand-roll the same tree walk independently. Mirrors SCHEMA-V1.go's
// equivalent section exactly. See "KNOWN LIMITATIONS" for what these
// functions do not yet handle.

/**
 * fieldById returns the ConfigField with the given id, searching
 * schema.fields and recursing into the item.fields of any array/map
 * field. Throws if no field with that id exists in the schema.
 */
export function fieldById(schema: Schema, id: string): ConfigField {
    function visit(fields: ConfigField[]): ConfigField | undefined {
        for (const f of fields) {
            if (f.id === id) return f;
            if (f.item?.fields) {
                const found = visit(f.item.fields);
                if (found) return found;
            }
        }
        return undefined;
    }

    const found = visit(schema.fields);
    if (!found) {
        throw new Error(`no field with id "${id}"`);
    }
    return found;
}

/**
 * valueById returns the configured value for the field with the given
 * id, read out of a real configuration payload. settings follows
 * Grafana's existing storage shape: root-level keys at the top level of
 * settings, jsonData and secureJsonData as nested objects under those
 * same keys.
 *
 * valueById only resolves fields whose storage is unset or "direct" (a
 * field's target/section/key correspond directly to one storage
 * location). For an "indexedPair" field (for example, the legacy HTTP
 * header convention), use resolveIndexedPairs or
 * resolveIndexedPairsAsMap instead — valueById throws rather than guess
 * at a single storage location that does not exist for that mapping
 * type. For a "computed" field, valueById throws because evaluating
 * storage.read is out of scope for this module (see "KNOWN
 * LIMITATIONS").
 *
 * valueById also throws for a virtual field (kind === 'virtual' has no
 * storage location to read) and for an item field (isItemField fields
 * describe the shape of each array/map element, not a single value at
 * the document level).
 */
export function valueById(schema: Schema, id: string, settings: Record<string, unknown>): unknown {
    const f = fieldById(schema, id);

    if (f.kind === 'virtual') {
        throw new Error(`field "${id}" is virtual and has no stored value`);
    }
    if (f.isItemField) {
        throw new Error(`field "${id}" is an item field; it has no single value at the document level`);
    }
    if (f.target === undefined) {
        throw new Error(`field "${id}" has no target`);
    }

    if (f.storage) {
        if (f.storage.type === 'indexedPair') {
            throw new Error(`field "${id}" uses an indexedPair storage mapping; use resolveIndexedPairs or resolveIndexedPairsAsMap instead of valueById`);
        }
        if (f.storage.type === 'computed') {
            throw new Error(`field "${id}" uses a computed storage mapping, which is not evaluated by this module`);
        }
    }

    let bucket = resolveBucket(settings, f.target);
    if (f.section) {
        bucket = navigateSection(bucket, f.section);
    }

    if (!(f.key in bucket)) {
        throw new Error(`field "${id}" (key "${f.key}") not present in configuration`);
    }
    return bucket[f.key];
}

/**
 * resolveBucket returns the object within settings corresponding to t.
 * settings is expected to follow Grafana's existing storage shape:
 * root-level keys live at the top level of settings itself; jsonData and
 * secureJsonData are nested objects under settings.jsonData and
 * settings.secureJsonData respectively.
 */
function resolveBucket(settings: Record<string, unknown>, t: TargetLocation): Record<string, unknown> {
    switch (t) {
        case 'root':
            return settings;
        case 'jsonData':
            return nestedObject(settings, 'jsonData');
        case 'secureJsonData':
            return nestedObject(settings, 'secureJsonData');
        default:
            throw new Error(`invalid target "${t}"`);
    }
}

function nestedObject(parent: Record<string, unknown>, key: string): Record<string, unknown> {
    if (!(key in parent)) {
        return {}; // bucket absent entirely is not an error; it's an empty bucket
    }
    const v = parent[key];
    if (typeof v !== 'object' || v === null || Array.isArray(v)) {
        throw new Error(`"${key}" is present but is not an object`);
    }
    return v as Record<string, unknown>;
}

/**
 * navigateSection walks a dotted section path within bucket, returning
 * the nested object at the end of that path. This supports exactly the
 * nesting depth section itself supports — see ConfigField.section and
 * "KNOWN LIMITATIONS" regarding the single-level constraint.
 */
function navigateSection(bucket: Record<string, unknown>, section: string): Record<string, unknown> {
    let cur = bucket;
    for (const seg of section.split('.')) {
        cur = nestedObject(cur, seg);
    }
    return cur;
}

/**
 * bucketForTarget is the indexedPair-specific counterpart of
 * resolveBucket: an indexed-pair key/value MappingField's target is only
 * ever "jsonData" or "secureJsonData" ("root" is not a valid target for
 * either half of an indexed pair), so this throws for any other value
 * rather than silently resolving it.
 */
function bucketForTarget(
    t: TargetLocation,
    jsonData: Record<string, unknown>,
    secureJsonData: Record<string, unknown>
): Record<string, unknown> {
    if (t === 'jsonData') return jsonData;
    if (t === 'secureJsonData') return secureJsonData;
    throw new Error(`indexedPair target "${t}" must be jsonData or secureJsonData`);
}

/**
 * resolveIndexedPairs reads an "indexedPair"-mapped field's actual
 * logical value out of a real configuration payload, by scanning for
 * numbered key/value pairs starting at storage.startIndex and assembling
 * them into an array of objects matching the field's item schema (one
 * object per pair, keyed by the item schema's field keys).
 *
 * The scan stops at the first missing index — an index 2 that is absent
 * while index 3 is present will not be seen. Use resolveIndexedPairsAsMap
 * if gap-tolerance matters more than preserving exact pair order and
 * duplicate names; see "KNOWN LIMITATIONS" for the trade-off.
 *
 * Returns an empty array (not a thrown error) if no pairs are present.
 */
export function resolveIndexedPairs(
    f: ConfigField,
    jsonData: Record<string, unknown>,
    secureJsonData: Record<string, unknown>
): Array<Record<string, unknown>> {
    if (f.storage?.type !== 'indexedPair') {
        throw new Error(`field "${f.id}" is not an indexedPair field`);
    }
    const mapping = f.storage;

    const keyBucket = bucketForTarget(mapping.key!.target, jsonData, secureJsonData);
    const valueBucket = bucketForTarget(mapping.value!.target, jsonData, secureJsonData);

    const start = mapping.startIndex ?? 1;

    if (!f.item?.fields || f.item.fields.length < 2) {
        throw new Error(`field "${f.id}": indexedPair requires an item schema with at least 2 fields`);
    }
    const nameFieldKey = f.item.fields[0].key;
    const valueFieldKey = f.item.fields[1].key;

    const results: Array<Record<string, unknown>> = [];
    for (let i = start; ; i++) {
        const nameKey = mapping.key!.pattern.replace('{index}', String(i));
        if (!(nameKey in keyBucket)) break;

        const item: Record<string, unknown> = { [nameFieldKey]: keyBucket[nameKey] };

        const valueKey = mapping.value!.pattern.replace('{index}', String(i));
        if (valueKey in valueBucket) {
            item[valueFieldKey] = valueBucket[valueKey];
        }
        // Value side absent (most commonly because it targets
        // secureJsonData and the caller's settings payload is a live,
        // already-saved datasource's settings rather than a schema example)
        // leaves item with only its name key set, rather than being an error.

        results.push(item);
    }

    return results;
}

/**
 * resolveIndexedPairsAsMap reads an "indexedPair"-mapped field's
 * configured pairs by scanning every key present in the key bucket —
 * rather than stopping at the first missing index — extracting each
 * key's numeric index via the mapping's key.pattern, and returning a
 * flat name -> value map. A name with no corresponding value present
 * maps to an empty string rather than being omitted.
 *
 * This function is gap-tolerant where resolveIndexedPairs is not, at the
 * cost documented under "KNOWN LIMITATIONS": it cannot represent two
 * distinct pairs sharing the same name, and an empty-string value is
 * indistinguishable from "the value side is genuinely unset or
 * unreadable."
 */
export function resolveIndexedPairsAsMap(
    f: ConfigField,
    jsonData: Record<string, unknown>,
    secureJsonData: Record<string, unknown>
): Record<string, string> {
    if (f.storage?.type !== 'indexedPair') {
        throw new Error(`field "${f.id}" is not an indexedPair field`);
    }
    const mapping = f.storage;

    const keyBucket = bucketForTarget(mapping.key!.target, jsonData, secureJsonData);
    const valueBucket = bucketForTarget(mapping.value!.target, jsonData, secureJsonData);

    const keyRe = patternToIndexRegex(mapping.key!.pattern);

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

/**
 * patternToIndexRegex converts a storage pattern such as
 * "httpHeaderName{index}" into a regular expression that matches real
 * stored keys and captures the numeric index as its first group.
 */
function patternToIndexRegex(pattern: string): RegExp {
    const escaped = pattern.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
    const withCapture = escaped.replace('\\{index\\}', '(\\d+)');
    return new RegExp(`^${withCapture}$`);
}

// ============================================================
// Array Item Schema
// ============================================================

/**
 * FieldItemSchema defines schema for array/map elements.
 * For arrays, it describes each element.
 * For maps, it describes each value (keys are always strings).
 *
 * An array or map field's own target/key/section describe where the
 * collection as a whole is stored; FieldItemSchema describes the shape
 * of each element/value within that collection. When valueType is
 * "object", fields describes the element's own properties, each of
 * which must be marked isItemField (see validateConfigField) since an
 * item field's storage location is inherited from its parent collection
 * rather than declared independently.
 */
export interface FieldItemSchema {
    valueType: ValueType;
    fields?: ConfigField[];
}

// ============================================================
// Value Types
// ============================================================

/**
 * ValueType enumerates the primitive and structural types a
 * ConfigField's value may take.
 *
 * These map directly onto JSON's own type system (string, number,
 * boolean, array, object) with two additions: "map", for an object
 * with dynamic string keys whose values share one schema (see
 * FieldItemSchema), and "any", for fields whose value may legitimately
 * take more than one shape and which therefore opt out of type-level
 * validation (intended to be used sparingly, only where a single type
 * genuinely cannot describe the data).
 */
export type ValueType = 'string' | 'number' | 'boolean' | 'array' | 'object' | 'map' | 'any';

const VALUE_TYPES: readonly ValueType[] = ['string', 'number', 'boolean', 'array', 'object', 'map', 'any'];

export function isValidValueType(v: string): v is ValueType {
    return (VALUE_TYPES as readonly string[]).includes(v);
}

// ============================================================
// Field Kind
// ============================================================

/**
 * FieldKind distinguishes fields that are actually persisted
 * ("storage", the default) from fields that exist only to describe
 * computed or UI-only state and are never written to root, jsonData,
 * or secureJsonData ("virtual"). See ConfigField's discussion of "Field
 * Kind: Storage Versus Virtual" for the canonical use of "virtual"
 * alongside effects.
 */
export type FieldKind = 'storage' | 'virtual';

const FIELD_KINDS: readonly FieldKind[] = ['storage', 'virtual'];

export function isValidFieldKind(k: string): k is FieldKind {
    return (FIELD_KINDS as readonly string[]).includes(k);
}

// ============================================================
// Target Location
// ============================================================

/**
 * TargetLocation enumerates the storage locations Grafana's existing
 * datasource configuration model already provides. This module
 * introduces no storage location beyond these three, by design — see
 * the module-level documentation's "Design Posture: Additive, Not
 * Migratory" discussion.
 *
 *   - "root": a top-level property of the datasource resource itself
 *     (for example, url, basicAuth, database) — the same properties
 *     Grafana's datasource model, provisioning format, and HTTP API
 *     have always exposed at the top level, outside jsonData and
 *     secureJsonData.
 *   - "jsonData": a property within the datasource's jsonData map:
 *     free-form, plugin-defined, non-secret configuration.
 *   - "secureJsonData": a property within the datasource's
 *     secureJsonData map: encrypted-at-rest, write-only configuration.
 *     A field with this target describes what may be written; the
 *     value cannot be read back through this schema or through
 *     Grafana's existing API once saved. See "KNOWN LIMITATIONS" for
 *     how the read-side "is this secret configured" signal
 *     (secureJsonFields) relates to fields described here.
 */
export type TargetLocation = 'root' | 'jsonData' | 'secureJsonData';

const TARGET_LOCATIONS: readonly TargetLocation[] = ['root', 'jsonData', 'secureJsonData'];

export function isValidTargetLocation(t: string): t is TargetLocation {
    return (TARGET_LOCATIONS as readonly string[]).includes(t);
}

// ============================================================
// UI Components
// ============================================================

/**
 * UIComponent enumerates the form control types a config editor
 * renderer may use to present a field. This is a closed set by design:
 * an unrecognized value is a schema-authoring error (see
 * validateConfigField), not a silently-ignored hint, since a renderer
 * that does not recognize a component value has no defined fallback
 * behavior to fall back to.
 */
export type UIComponent =
    | 'input'
    | 'textarea'
    | 'select'
    | 'multiselect'
    | 'radio'
    | 'checkbox'
    | 'switch'
    | 'code'
    | 'keyvalue'
    | 'list';

const UI_COMPONENTS: readonly UIComponent[] = [
    'input',
    'textarea',
    'select',
    'multiselect',
    'radio',
    'checkbox',
    'switch',
    'code',
    'keyvalue',
    'list',
];

export function isValidUIComponent(c: string): c is UIComponent {
    return (UI_COMPONENTS as readonly string[]).includes(c);
}

/**
 * FieldUI defines UI rendering hints.
 *
 * FieldUI is presentation metadata. It has no effect on validation or
 * on the OpenAPI settings produced on the Go side, with one documented
 * exception: an enum is derived from options when the field has no
 * explicit "allowedValues" validation rule, as a convenience for
 * authors who would otherwise have to state the same allowed-values
 * list twice. See SCHEMA-V1.md's discussion of why validations, not
 * ui.options, is the data contract and ui.options is presentation only.
 */
export interface FieldUI {
    component: UIComponent;

    multiline?: boolean;
    rows?: number;
    options?: FieldOption[];

    allowCustom?: boolean;
    width?: UIWidth;

    placeholder?: string;

    /**
     * Language hint for code editor components.
     * Example: "promql", "logql", "traceql", "sql", "json"
     */
    language?: string;
}

/** UIWidth defines layout width. */
export type UIWidth = 'full' | 'half';

const UI_WIDTHS: readonly UIWidth[] = ['full', 'half'];

export function isValidUIWidth(w: string): w is UIWidth {
    return (UI_WIDTHS as readonly string[]).includes(w);
}

// ============================================================
// Validations
// ============================================================

/**
 * ValidationRuleType enumerates the kinds of validation rule this
 * module can express. This is the schema's data contract — see
 * SCHEMA-V1.md — distinct from and authoritative over any allowed-values
 * list that may also be present in a field's ui.options for display
 * purposes.
 */
export type ValidationRuleType = 'pattern' | 'range' | 'length' | 'itemCount' | 'allowedValues' | 'custom';

/**
 * FieldValidationRule is a discriminated union of validation rules.
 *
 * Exactly one of the type-specific property groups below is meaningful
 * for a given rule, selected by type; see validateFieldValidationRule
 * for which properties each type requires. As of this schema version,
 * FieldValidationRule's structural shape is validated (the right
 * properties are present for the declared type), and "pattern",
 * "range", "length", "itemCount", and "allowedValues" are further
 * translated into real OpenAPI/JSON Schema constraints on the Go side —
 * pattern, minimum/maximum, minLength/maxLength, minItems/maxItems, and
 * enum, respectively. "custom"'s expression is stored but not evaluated
 * by anything in this module; see "KNOWN LIMITATIONS".
 */
export interface FieldValidationRule {
    type: ValidationRuleType;
    id?: string;
    message?: string;

    /** pattern validation */
    pattern?: string;

    /** range / length / itemCount validation */
    min?: number;
    max?: number;

    /** allowedValues validation */
    values?: unknown[];

    /** custom validation */
    expression?: string;
}

export function validateFieldValidationRule(rule: FieldValidationRule): string[] {
    const errors: string[] = [];
    switch (rule.type) {
        case 'pattern':
            if (!rule.pattern) {
                errors.push('pattern validation requires pattern');
            }
            break;
        case 'range':
        case 'length':
        case 'itemCount':
            if (rule.min === undefined && rule.max === undefined) {
                errors.push(`${rule.type} validation requires min or max`);
            }
            break;
        case 'allowedValues':
            if (!rule.values || rule.values.length === 0) {
                errors.push('allowedValues validation requires values');
            }
            break;
        case 'custom':
            if (!rule.expression) {
                errors.push('custom validation requires expression');
            }
            break;
        default:
            errors.push(`unknown validation rule type: ${rule.type}`);
    }
    return errors;
}

// ============================================================
// Overrides
// ============================================================

/**
 * FieldOverride allows dynamic modifications.
 *
 * An override describes how a field's presentation or validation
 * should change under a stated condition (when), without duplicating
 * the entire field definition. As with dependsOn/requiredWhen/
 * disabledWhen, when is a CEL-like expression string that is not
 * currently parsed or evaluated by this module; see "KNOWN
 * LIMITATIONS". Overrides' nested validations are independently
 * structurally validated, but — as a consequence of when not being
 * evaluated — no override is currently applied to the OpenAPI settings
 * produced on the Go side.
 */
export interface FieldOverride {
    when: string;

    defaultValue?: unknown;
    description?: string;
    placeholder?: string;
    tooltip?: string;

    validations?: FieldValidationRule[];
    options?: FieldOption[];
}

// ============================================================
// Effects
// ============================================================

/**
 * FieldEffect declares that when a field's value matches a condition,
 * the listed target fields should be set to the specified values.
 *
 * This provides a structured, machine-readable alternative to opaque
 * computed write expressions for virtual selector fields.
 *
 * Example: an auth method dropdown that sets root.basicAuth and
 * jsonData.oauthPassThru depending on which option is selected.
 *
 * effects is deliberately structured rather than expressed as a CEL
 * write expression: the set of "selector value picked -> these fields
 * get these values" relationships that occur in practice is small and
 * enumerable, and representing it as a validated when/set pair (rather
 * than an opaque string naming a side-effecting function) lets
 * validateRefs confirm every set key resolves to a real field id
 * without needing to parse or evaluate the when condition itself to do
 * so. when remains a CEL-like string today and is not evaluated by
 * this module; only its presence is checked. See "KNOWN LIMITATIONS".
 */
export interface FieldEffect {
    /**
     * when is a CEL expression evaluated against the field's value.
     * Convention: use "value" to refer to the field's current value.
     * Example: "value == 'basic-auth'"
     */
    when: string;

    /** set maps field IDs to the values they should be set to when the condition matches. */
    set: Record<string, unknown>;
}

export function validateFieldEffect(effect: FieldEffect): string[] {
    const errors: string[] = [];
    if (!effect.when) {
        errors.push('effect when is required');
    }
    if (!effect.set || Object.keys(effect.set).length === 0) {
        errors.push('effect set must not be empty');
    }
    return errors;
}

// ============================================================
// Storage Mapping
// ============================================================

/**
 * StorageMappingType enumerates how a logical field maps onto
 * Grafana's existing storage representation when that mapping is not
 * a simple one-to-one property ("direct", the default and the common
 * case).
 *
 *   - "direct": the field's target and key map directly onto a single
 *     property in that target's storage location. A field with no
 *     explicit storage is implicitly "direct".
 *   - "indexedPair": describes Grafana's existing legacy convention for
 *     representing a user-extensible list of name/value pairs as a
 *     numbered series of individual properties (for example,
 *     httpHeaderName1/httpHeaderValue1, httpHeaderName2/
 *     httpHeaderValue2, and so on), optionally with the name and value
 *     halves of each pair stored in different targets — the documented
 *     convention for HTTP headers, where names are not secret
 *     (jsonData) but values may be (secureJsonData). This mapping type
 *     describes that existing convention; it does not change it. See
 *     "KNOWN LIMITATIONS" for the current scope of what reads this
 *     mapping today.
 *   - "computed": describes a field whose stored representation is
 *     derived from, or split across, other fields via a read and/or
 *     write expression, rather than corresponding to a single stored
 *     property. As with other CEL-like expression fields in this
 *     module, read and write are stored as strings and are not
 *     evaluated by anything in this module as of this schema version;
 *     see "KNOWN LIMITATIONS".
 */
export type StorageMappingType = 'direct' | 'indexedPair' | 'computed';

/** StorageMapping maps logical fields to Grafana storage. */
export interface StorageMapping {
    type: StorageMappingType;

    /** Indexed pair mapping. */
    key?: MappingField;
    value?: MappingField;
    startIndex?: number;

    /** Computed mapping. */
    read?: string;
    write?: string;
}

/**
 * validateStorageMapping checks that a StorageMapping's populated
 * properties are consistent with its declared type — for example, that
 * an "indexedPair" mapping supplies both key and value mapping fields
 * and does not also supply read/write (which belong only to
 * "computed"), and vice versa.
 */
export function validateStorageMapping(mapping: StorageMapping): string[] {
    const errors: string[] = [];
    switch (mapping.type) {
        case 'direct':
            if (mapping.key || mapping.value || mapping.startIndex !== undefined || mapping.read || mapping.write) {
                errors.push('direct mapping must not have key/value/startIndex/read/write');
            }
            break;

        case 'indexedPair':
            if (!mapping.key || !mapping.value) {
                errors.push('indexedPair requires key and value');
            }
            if (mapping.read || mapping.write) {
                errors.push('indexedPair must not have read/write');
            }
            if (mapping.key) {
                errors.push(...validateMappingField(mapping.key).map((e) => `indexedPair key: ${e}`));
            }
            if (mapping.value) {
                errors.push(...validateMappingField(mapping.value).map((e) => `indexedPair value: ${e}`));
            }
            break;

        case 'computed':
            if (!mapping.read && !mapping.write) {
                errors.push('computed mapping requires read or write');
            }
            if (mapping.key || mapping.value || mapping.startIndex !== undefined) {
                errors.push('computed mapping must not have key/value/startIndex');
            }
            break;

        default:
            errors.push(`unknown mapping type: ${mapping.type}`);
    }
    return errors;
}

/**
 * MappingField describes one half (key or value) of an indexedPair
 * mapping: which storage target it lives in, and the naming pattern
 * used to generate each numbered property name (for example,
 * "httpHeaderName{index}").
 */
export interface MappingField {
    target: TargetLocation;
    pattern: string;
}

export function validateMappingField(field: MappingField): string[] {
    const errors: string[] = [];
    if (!isValidTargetLocation(field.target)) {
        errors.push(`invalid target "${field.target}"`);
    }
    if (!field.pattern) {
        errors.push('pattern is required');
    }
    return errors;
}

// ============================================================
// Options
// ============================================================

/**
 * FieldOption describes one selectable choice for a select/radio/
 * multiselect UI component, or one entry in an allowedValues
 * validation rule's values list.
 */
export interface FieldOption {
    label: string;
    value: unknown;
    description?: string;
}

/**
 * validateOptionValue checks that an option value is non-null/
 * non-undefined and compatible with the given field valueType.
 */
export function validateOptionValue(v: unknown, vt: ValueType): boolean {
    if (v === null || v === undefined) {
        return false;
    }
    switch (vt) {
        case 'string':
            return typeof v === 'string';
        case 'number':
            return typeof v === 'number';
        case 'boolean':
            return typeof v === 'boolean';
        default:
            // array/object/map/any options are not type-checked
            return true;
    }
}

// ============================================================
// Groups
// ============================================================

/**
 * ConfigGroup describes a presentational grouping of fields — for
 * example, a collapsible "Advanced" section in a generated config
 * editor. Groups are pure UI layout metadata; see Schema.groups for the
 * full discussion of what depends on this (nothing structural) and what
 * does not (storage, validation, the OpenAPI settings produced on the
 * Go side).
 */
export interface ConfigGroup {
    id: string;
    title: string;
    description?: string;
    order?: number;
    optional?: boolean;
    fieldRefs: string[];
}

// ============================================================
// Relationships
// ============================================================

/**
 * RelationshipType enumerates the kinds of semantic connection a
 * FieldRelationship may declare between fields.
 *
 *   - "pair": connects two fields that together form one logical unit
 *     of configuration — most commonly a username and a password.
 *   - "group": connects an arbitrary set of fields that are
 *     semantically related but do not fit the narrower pair shape.
 *   - "datasourceReference": marks that one or more fields hold a
 *     reference (typically a UID) to another Grafana datasource — for
 *     example, a "derived fields" configuration that links log lines to
 *     a tracing datasource. targetPluginType, when set, constrains
 *     which plugin type the referenced datasource UID is expected to
 *     resolve to.
 */
export type RelationshipType = 'pair' | 'group' | 'datasourceReference';

const RELATIONSHIP_TYPES: readonly RelationshipType[] = ['pair', 'group', 'datasourceReference'];

export function isValidRelationshipType(r: string): r is RelationshipType {
    return (RELATIONSHIP_TYPES as readonly string[]).includes(r);
}

/**
 * FieldRelationship describes a semantic, non-structural connection
 * between two or more fields. Like ConfigGroup, a relationship carries
 * no effect on storage or on the OpenAPI settings produced on the Go
 * side; it exists for renderers and automated/assisted configuration
 * consumers that benefit from knowing fields are related (for example,
 * presenting a username/password pair together, or warning that a
 * datasource-reference field should be checked against the referenced
 * datasource's existence).
 */
export interface FieldRelationship {
    type: RelationshipType;
    fields: string[];
    description?: string;

    /** targetPluginType constrains the datasource UID to a specific plugin. Only applicable when type is "datasourceReference". */
    targetPluginType?: string;
}

/**
 * Instruction is a structured, free-form guidance entry intended
 * primarily for non-human or semi-autonomous consumers of the schema —
 * most directly, a chat-driven configuration assistant (such as the
 * Grafana Assistant) that needs plugin-specific guidance not otherwise
 * captured by individual field definitions in order to help a person
 * reach a working datasource connection in as few exchanges as
 * possible. tags allows an Instruction to be scoped or categorized at
 * the author's discretion; neither msg nor tags is validated against a
 * fixed vocabulary, and neither has any effect on storage or on the
 * OpenAPI settings produced on the Go side.
 */
export interface Instruction {
    msg: string;
    tags?: string[];
}

// ============================================================
// KNOWN LIMITATIONS
// ============================================================
//
// This section records, deliberately and without euphemism, what this
// version of the schema does not yet do. Each item is scoped as a
// future, additive enhancement — none requires changing the shape of
// any type already defined above, and none requires migrating any
// already-stored datasource configuration or any already-published
// schema document. This list mirrors SCHEMA-V1.go's "KNOWN LIMITATIONS"
// exactly; see that file for the canonical, identically-numbered
// version, and the accompanying RFC for the proposed sequencing of
// this work.
//
//  1. Expression strings are not parsed or evaluated.
//  2. StorageMapping is descriptive metadata, not yet an executable
//     mapping.
//  3. The OpenAPI settings conversion on the Go side does not yet read
//     storage at all; a consequence specific to indexedPair mappings
//     whose value target is secureJsonData: the generated settings give
//     no indication that the corresponding array's values are secret.
//     This is a correctness gap in the generated output, not merely a
//     missing feature, and is the highest-priority item in this list.
//  4. section supports exactly one level of nesting.
//  5. No field carries a semantic role independent of its name, which
//     directly limits the reliability of any automated HTTP client
//     derivation.
//  6. Auth representation is whatever the plugin author chose, with no
//     schema-level distinction between an explicit discriminator field,
//     a set of independently-toggleable boolean flags, or a hybrid of
//     the two, and no detection of mutually incompatible combinations.
//  7. No mechanism exists for a plugin with more than one independent
//     connection within a single datasource instance.
//  8. id format is now partially enforced, by validateFieldIdFormat
//     (called from validateSchema). It rejects an id segment outside
//     [A-Za-z_][A-Za-z0-9_]* and rejects one id being a strict
//     dotted-path prefix of another, mirroring SCHEMA-V1.go's
//     ValidateFieldIDFormat exactly. It does not enforce the recommended
//     hierarchical-by-meaning convention beyond that. A schema document
//     written before this check existed that violates either rule will
//     now fail validateSchema; this is a deliberate, newly-enforced
//     behavior change within the v1 schema version, not a v1-to-v2
//     migration (see item 11).
//  9. tags and examples are accepted and stored but not read by any
//     code in this module.
//  10. repeatable and pattern (on ConfigField) are accepted and stored
//      but not read by validateConfigField or by anything that produces
//      OpenAPI-shaped settings.
//  11. There is no schema-version migration mechanism.
//  12. This module's validation functions collect every error found
//      before returning, while the Go side's Validate returns on the
//      first error encountered. A single invalid schema document can
//      therefore currently surface a different number of reported
//      problems depending on which language validated it. This
//      TypeScript module is, today, the more permissive of the two in
//      terms of how much it reports per call; the Go side is
//      authoritative for go/no-go validity, but not for completeness
//      of the error list.
//  13. resolveIndexedPairs infers which of a field's two declared item
//      fields is the pair's "name" and which is its "value" by their
//      position in item.fields (first = name, second = value), because
//      neither ConfigField nor FieldItemSchema carries an explicit
//      pair-role tag. A schema that declares these two item fields in
//      the opposite order produces silently swapped results, with no
//      validation error. Mirrors SCHEMA-V1.go's identical limitation
//      exactly.
//  14. resolveIndexedPairsAsMap collapses two distinct stored pairs that
//      happen to share the same name into one map entry, with object
//      key insertion order determining which one survives in practice.
//      It also returns an empty string for a name whose corresponding
//      value is absent, indistinguishable from a pair whose value was
//      genuinely configured as an empty string. resolveIndexedPairs has
//      neither limitation but is not gap-tolerant the way
//      resolveIndexedPairsAsMap is — see each function's own
//      documentation for this trade-off.
//  15. Neither resolveIndexedPairs nor resolveIndexedPairsAsMap, nor
//      valueById, can read a value that targets secureJsonData out of a
//      real, already-saved datasource's settings — secureJsonData is
//      write-only once saved and Grafana's own API never returns it.
//      Calling these functions against such a payload for a
//      secret-targeted field produces the same "absent" result as a
//      field that was simply never configured. These functions work as
//      expected against a schema's own settingsExamples or any other
//      payload that genuinely embeds secret values.
