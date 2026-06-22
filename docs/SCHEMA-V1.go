// Package dsconfig defines the declarative configuration schema for Grafana
// datasource plugins.
//
// # PURPOSE
//
// dsconfig is a semantic description layer placed on top of Grafana's
// existing datasource configuration model (root fields, jsonData,
// secureJsonData). It does not change how Grafana stores datasource
// configuration today. Every field described by a dsconfig Schema still
// lives exactly where it lives now. dsconfig is additive: a plugin can
// adopt it without migrating, renaming, or moving a single stored value.
//
// dsconfig exists to serve four consumers that today have no shared,
// machine-readable contract to work from:
//
//  1. Config editors (frontend forms) — today hand-written per plugin,
//     with no guarantee they match what the backend actually parses.
//  2. Backend settings parsing (grafana-plugin-sdk-go) — today reads
//     untyped jsonData/secureJsonData maps with no schema-level contract.
//  3. Provisioning (datasources.yaml) and the Grafana App Platform /
//     Kubernetes-style datasource API — both need a description of a
//     valid datasource config that exists independently of any running
//     plugin instance, so a config can be validated before it is applied.
//  4. Automated and assisted configuration — including the Grafana
//     Assistant chat-driven datasource workflow, and any other tool that
//     needs to read, generate, or validate a datasource configuration
//     without parsing plugin source code.
//
// WHY THIS PACKAGE EXISTS: TWO PRIMARY DRIVERS
//
// Driver 1 — App Platform / Kubernetes-style API compatibility.
// Grafana's App Platform exposes resources through a Kubernetes-style
// API (CRD-shaped: apiVersion, kind, metadata, spec, status). A
// datasource's spec needs an OpenAPI-shaped schema describing what a
// valid instance of that resource looks like, the same way any
// Kubernetes Custom Resource Definition needs a structural schema for
// its spec. Datasource jsonData/secureJsonData today have no such
// schema; they are untyped maps. dsconfig is the semantic layer that
// produces that OpenAPI-shaped schema (see ToPluginSchemaSettings in
// convert.go) from one declarative source per plugin, so that App
// Platform's CRUD operations on the datasource resource — create,
// read, update, delete, including admission-time validation — have a
// structural schema to validate against, exactly as App Platform's own
// resource model expects. dsconfig does not replace or duplicate
// Grafana's existing datasource storage; it describes the same
// root/jsonData/secureJsonData shape App Platform must already
// represent, in the form App Platform's API machinery requires.
//
// Driver 2 — reliable, automatically derived HTTP clients.
// Today, the logic that turns a plugin's stored configuration into a
// working *http.Client (TLS setup, auth header/round-tripper wiring,
// timeout configuration) is hand-written, per plugin, and frequently
// duplicated with small inconsistencies across otherwise similar
// plugins. Because dsconfig fields carry typed, structured metadata
// (storage location, validation rules, and — in later schema versions —
// semantic role information) rather than being opaque map entries, the
// same schema that drives config-editor generation and provisioning
// validation is also the input a future SDK helper can use to build a
// transport-correct, auth-correct HTTP client without per-plugin code.
// This package's current version does not implement that derivation;
// see "KNOWN LIMITATIONS" below. The schema is shaped, from this
// version forward, so that derivation is possible without a breaking
// change to the fields already defined here.
//
// DESIGN POSTURE: ADDITIVE, NOT MIGRATORY
//
// Every design decision in this package follows one rule: adopting
// dsconfig must never require a plugin to change what it stores, how it
// stores it, or where. The schema describes root, jsonData, and
// secureJsonData exactly as Grafana persists them today. A plugin that
// adds a dsconfig schema file changes nothing about its existing
// behavior until something is built to consume that schema (a generated
// editor, a provisioning validator, an App Platform resource schema, a
// derived HTTP client). This is why TargetLocation enumerates exactly
// the three storage locations Grafana already has, why StorageMapping's
// "direct" type is a no-op over today's default behavior, and why
// adopting this package for an existing plugin is, by construction, a
// documentation exercise first and an automation opportunity second.
package dsconfig

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// ============================================================
// Root Schema
// ============================================================

// Schema is the top-level schema definition.
// It acts as the single source of truth for datasource configuration.
//
// One Schema value describes exactly one plugin type's configuration
// surface: every field stored in that plugin's root-level datasource
// properties, jsonData, and secureJsonData. A Schema is normally
// authored once per plugin and shipped as a JSON file alongside the
// plugin (see SCHEMA-V1.json for the wire format this struct serializes
// to/from), and is also the structure produced by code-level Go schema
// construction helpers for plugins that build their schema
// programmatically.
//
// Schema is consumed in at least three independent ways, and every
// field below should be read with all three consumers in mind:
//
//   - Structurally, by Validate() and ValidateRefs(), which check that
//     the schema document itself is internally well-formed (every
//     reference resolves, every field is shaped correctly for its
//     kind). This is schema-authoring-time validation; it says nothing
//     about whether a particular stored datasource config is valid.
//   - As an OpenAPI-shaped settings document, via ToPluginSchemaSettings
//     in convert.go, which is the shape grafana-plugin-sdk-go and the
//     Grafana App Platform / Kubernetes-style datasource API expect for
//     describing and validating an instance of this plugin's
//     configuration (its CRD-style spec).
//   - As a description for UI and automation consumers — config editor
//     generation, documentation generation, and chat-driven
//     configuration assistants (such as the Grafana Assistant) that
//     need to know what fields exist, what they mean, and what values
//     are valid, in order to create or repair a working datasource
//     configuration without hand-coded per-plugin logic.
//   - As a lookup target, via FieldByID, ValueByID,
//     ResolveIndexedPairs, and ResolveIndexedPairsAsMap, for any
//     consumer that needs to resolve an id (referenced by Groups,
//     Relationships, Effects, or supplied externally — for example, by
//     a person or an assistant referring to a field by name) to that
//     field's definition or its actual configured value in a real
//     instance settings payload, without re-implementing the
//     id-to-storage-location resolution that Target/Section/Key and
//     Storage already fully describe.
type Schema struct {
	// SchemaVersion defines the version of the schema spec.
	//
	// This versions the dsconfig schema *format* itself (the shape of
	// Schema/ConfigField/etc. as defined by this package), not the
	// plugin's own config version. A consumer reading a Schema document
	// uses SchemaVersion to know which version of this package's types
	// it must be parsed against. See "KNOWN LIMITATIONS" for the current
	// state of cross-version handling.
	SchemaVersion string `json:"schemaVersion"`

	// PluginType uniquely identifies the datasource plugin.
	//
	// This must match the plugin's own type identifier (the same
	// identifier Grafana's plugin system and the datasource's "type"
	// property already use). It is the join key between a dsconfig
	// Schema document and a real, running plugin instance, and the key
	// an App Platform resource's apiVersion/kind would reference when
	// locating the structural schema for a given datasource kind.
	PluginType string `json:"pluginType"`

	// PluginName is a human-readable name.
	//
	// Display-only. Not used as a reference key anywhere in this
	// package; PluginType is the identifier for that purpose.
	PluginName string `json:"pluginName"`

	// Optional documentation URL.
	//
	// Points to human-authored documentation for the plugin as a whole.
	// Individual fields may carry their own DocURL (see ConfigField) for
	// field-level documentation; this is the plugin-level equivalent.
	DocURL string `json:"docURL,omitempty"`

	// Fields defines all configuration fields.
	//
	// This is the source of truth referred to throughout this package's
	// documentation. Every other piece of schema-level metadata — groups,
	// relationships, instructions — is descriptive metadata layered on
	// top of Fields, and is validated against the field IDs declared
	// here (see ValidateRefs). Fields is required and must be non-empty;
	// a Schema with no fields describes nothing.
	Fields []ConfigField `json:"fields"`

	// Optional UI grouping
	//
	// Groups are presentation metadata only. They describe how a config
	// editor might lay fields out into sections (for example,
	// "Connection", "Authentication", "Advanced"); they have no effect
	// on storage, validation, or the OpenAPI settings produced by
	// ToPluginSchemaSettings. A Schema with no Groups is fully valid;
	// a consumer without a Groups-aware renderer can safely ignore this
	// field and render Fields in declaration order.
	Groups []ConfigGroup `json:"groups,omitempty"`

	// Optional Instruction
	//
	// Free-form, structured guidance intended for non-human or
	// semi-autonomous consumers of the schema — most directly, a
	// chat-driven configuration assistant that needs plugin-specific
	// guidance beyond what individual field descriptions convey (for
	// example, "ask the user for their Prometheus server's external URL,
	// not the internal one" or "TLS settings are only relevant when the
	// server enforces mTLS"). Instructions have no effect on validation
	// or storage; they exist purely to make the schema more useful to
	// a consumer trying to drive a configuration workflow to a working
	// result with as few back-and-forth turns as possible.
	Instructions []Instruction `json:"instructions,omitempty"`

	// Relationships between fields
	//
	// Semantic, not structural: Relationships record that two or more
	// fields are conceptually connected (for example, a username/
	// password pair, or a field that references another datasource by
	// UID) for the benefit of a renderer or assistant deciding how to
	// present or reason about related fields together. Relationships
	// do not affect validation beyond the reference-integrity check in
	// ValidateRefs (every referenced field ID must exist), and they do
	// not affect the OpenAPI settings produced by ToPluginSchemaSettings.
	Relationships []FieldRelationship `json:"relationships,omitempty"`
}

// Validate checks that a Schema document is internally well-formed.
//
// This is schema-authoring-time validation. It confirms the document
// itself is structurally sound — every required top-level property is
// present, every field is individually valid for its declared kind, and
// every cross-reference (group field refs, relationship field refs,
// effect target IDs) resolves to a real field ID. It does not validate
// any actual stored datasource configuration against this schema; that
// is a distinct, presently unimplemented capability — see "KNOWN
// LIMITATIONS" below.
func (s *Schema) Validate() error {
	if s.SchemaVersion == "" {
		return fmt.Errorf("schemaVersion is required")
	}
	if s.PluginType == "" {
		return fmt.Errorf("pluginType is required")
	}
	if s.PluginName == "" {
		return fmt.Errorf("pluginName is required")
	}
	if len(s.Fields) == 0 {
		return fmt.Errorf("fields is required")
	}

	for i := range s.Fields {
		if err := s.Fields[i].Validate(); err != nil {
			return err
		}
	}

	fieldIDs, err := s.FieldIDs()
	if err != nil {
		return err
	}

	if err := ValidateFieldIDFormat(fieldIDs); err != nil {
		return err
	}

	if err := s.ValidateRefs(fieldIDs); err != nil {
		return err
	}

	return nil
}

// idSegmentPattern matches a single dot-separated segment of a field ID.
// This is deliberately the same character class a CEL-style identifier
// accepts: ASCII letters, digits, and underscore, not starting with a
// digit. Hyphens, brackets, spaces, and other punctuation are excluded
// because, once a future expression evaluator parses DependsOn/
// RequiredWhen/DisabledWhen/FieldEffect.When strings that reference field
// IDs by name, those characters either are not valid identifier
// characters or are already meaningful CEL syntax (e.g. "[" is index/map
// access). Restricting the character set now, before any evaluator
// exists, converts a currently invisible failure mode (an ID that breaks
// expression parsing only once a future evaluator is run against it)
// into an immediate, attributable schema-authoring error. See "KNOWN
// LIMITATIONS" below for what this check does and does not guarantee.
var idSegmentPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// ValidateFieldIDFormat checks every field ID in the given set against two
// rules:
//
//  1. Each dot-separated segment of the ID must match idSegmentPattern.
//  2. No ID may be a strict dotted-path prefix of another ID in the same
//     set (for example, "tls.clientAuth" and "tls.clientAuth.enabled"
//     coexisting is rejected) — a future expression evaluator resolving
//     dotted paths against this same flat ID set cannot otherwise
//     distinguish "the field whose id is exactly tls.clientAuth.enabled"
//     from "the enabled member of whatever tls.clientAuth resolves to."
//
// fieldIDs is the set produced by FieldIDs. This check is run
// unconditionally as part of Validate; it is not optional and cannot be
// skipped by a caller validating a single field in isolation, because
// the prefix-collision rule is inherently a property of the whole set,
// not of any one field.
func ValidateFieldIDFormat(fieldIDs map[string]bool) error {
	all := make([]string, 0, len(fieldIDs))
	for id := range fieldIDs {
		all = append(all, id)
		for _, seg := range strings.Split(id, ".") {
			if !idSegmentPattern.MatchString(seg) {
				return fmt.Errorf("field id %q: segment %q is not a valid identifier (must match %s)", id, seg, idSegmentPattern.String())
			}
		}
	}

	// Prefix-collision check: no id may be a strict dotted-path prefix
	// of another. Sorting first makes this an O(n log n) adjacent-pair
	// scan rather than an O(n^2) all-pairs comparison.
	sort.Strings(all)
	for i := 0; i+1 < len(all); i++ {
		a, b := all[i], all[i+1]
		if strings.HasPrefix(b, a+".") {
			return fmt.Errorf("field id %q is a dotted-path prefix of field id %q; this is ambiguous for any future consumer that resolves ids as dotted paths (e.g. an expression evaluator)", a, b)
		}
	}

	return nil
}

// ValidateRefs checks that all group and relationship field references
// point to existing field IDs.
//
// fieldIDs is the complete set of field IDs declared anywhere in the
// schema (top-level fields and, recursively, item fields of array/map
// fields), as produced by FieldIDs. ValidateRefs checks three categories
// of reference, all keyed by field ID rather than by storage key,
// consistent with this package's id/key separation (see ConfigField):
// ConfigGroup.FieldRefs, FieldRelationship.Fields, and
// FieldEffect.Set keys.
func (s *Schema) ValidateRefs(fieldIDs map[string]bool) error {
	for _, g := range s.Groups {
		for _, ref := range g.FieldRefs {
			if !fieldIDs[ref] {
				return fmt.Errorf("group %s references unknown field id: %s", g.ID, ref)
			}
		}
	}

	for _, r := range s.Relationships {
		if !r.Type.IsValid() {
			return fmt.Errorf("relationship has invalid type %q", r.Type)
		}
		for _, ref := range r.Fields {
			if !fieldIDs[ref] {
				return fmt.Errorf("relationship references unknown field id: %s", ref)
			}
		}
	}

	// Validate effect set keys reference known field IDs
	var visitEffects func(fields []ConfigField) error
	visitEffects = func(fields []ConfigField) error {
		for _, f := range fields {
			for i, eff := range f.Effects {
				for ref := range eff.Set {
					if !fieldIDs[ref] {
						return fmt.Errorf("field %s: effect[%d].set references unknown field id: %s", f.ID, i, ref)
					}
				}
			}
			if f.Item != nil {
				if err := visitEffects(f.Item.Fields); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := visitEffects(s.Fields); err != nil {
		return err
	}

	return nil
}

// ============================================================
// Field Definition
// ============================================================

// ConfigField represents a single configuration field.
//
// ConfigField is the unit of description for everything this package
// models: a piece of data that is either stored somewhere in Grafana's
// existing datasource config model (root, jsonData, or secureJsonData),
// or computed/virtual and not stored at all. Every other type in this
// package exists to describe some aspect of a ConfigField in more
// detail (its UI presentation, its validation rules, its storage
// mapping, and so on).
//
// # ID VERSUS KEY
//
// ConfigField deliberately separates two identifiers that are easy to
// conflate but serve different purposes:
//
//   - ID is the field's globally unique name within the schema. It is
//     the identifier every cross-reference in this package uses:
//     ConfigGroup.FieldRefs, FieldRelationship.Fields, and
//     FieldEffect.Set keys all refer to fields by ID. ID is a schema-
//     authoring concern; it never appears in the stored datasource
//     configuration itself, and changing a field's ID does not change
//     anything about how or where its value is stored. A recommended,
//     but not currently enforced, convention is a dot-separated path
//     describing the field's logical position (for example,
//     "auth.basicAuthPassword"); see "KNOWN LIMITATIONS" regarding the
//     lack of enforcement.
//   - Key is the field's local name within whatever it is actually
//     stored in — a property name within root, within jsonData, within
//     secureJsonData, or within an item object for array/map fields.
//     Key is the identifier that matches what Grafana's existing
//     storage model, and any existing plugin backend code reading that
//     storage, already expects. Key is never required to be globally
//     unique; only unique within its immediate storage context.
//
// This separation exists specifically to keep the schema additive (see
// the package-level documentation): Key always matches what is already
// being stored, today, by the plugin, with zero changes, while ID gives
// every other part of this package (groups, relationships, effects, and
// any future role- or scope-based mechanism) a stable, storage-
// independent name to reference.
//
// # STORAGE TARGET
//
// Target (when set) declares which of Grafana's three existing storage
// locations holds this field's value: root-level datasource properties,
// the jsonData map, or the secureJsonData map. This package never
// introduces a fourth location and never changes the read/write
// semantics Grafana already applies to these three — most importantly,
// secureJsonData remains write-only from the schema's perspective: a
// field targeting SecureJSONTarget describes what may be written, not
// a value that can be read back. See TargetLocation and the discussion
// under "KNOWN LIMITATIONS" regarding secureJsonFields (the existing
// read-side indicator of "is a secret configured") and how it relates
// to fields described here.
//
// FIELD KIND: STORAGE VERSUS VIRTUAL
//
// Most fields are storage fields: they have a Target and describe a
// real, persisted value. A field may instead be declared Kind:
// VirtualField, meaning it has no Target and is not persisted at all —
// it exists only to describe computed or UI-only state. The canonical
// use of a virtual field is a selector control (for example, an
// "Authentication method" dropdown) whose own value is never stored,
// but whose selection drives the values of one or more real storage
// fields via Effects. See FieldEffect.
//
// APP PLATFORM / KUBERNETES-STYLE API RELEVANCE
//
// Each storage field, taken together with its ValueType, Validations,
// and Required/RequiredWhen state, supplies exactly the information an
// OpenAPI-style structural schema needs for one property: type,
// constraints, and required-ness. This is what ToPluginSchemaSettings
// (convert.go) walks to build the spec consumed by
// grafana-plugin-sdk-go and, in turn, usable as the structural schema
// for a Kubernetes-style Custom Resource Definition describing this
// plugin's datasource spec under Grafana's App Platform. A ConfigField
// is therefore always describing two things at once: a value Grafana
// already stores, and one property of the structural schema App
// Platform's CRUD and admission-validation machinery needs for that
// same value.
type ConfigField struct {
	// ID is globally unique (used for references)
	ID string `json:"id"`

	// Key is the local key (used in storage or object structures)
	Key string `json:"key"`

	Label       string `json:"label,omitempty"`
	Description string `json:"description,omitempty"`
	DocURL      string `json:"docURL,omitempty"`

	// Core typing
	ValueType ValueType `json:"valueType"`

	// Storage location (required for storage fields)
	Target *TargetLocation `json:"target,omitempty"`

	// Section is the dotted path prefix within the target for nested objects.
	// Example: for jsonData.tracesToLogs.datasourceUid, target="jsonData",
	// section="tracesToLogs", key="datasourceUid".
	Section string `json:"section,omitempty"`

	// Field type: storage (default) or virtual
	Kind FieldKind `json:"kind,omitempty"`

	// True if part of array item schema
	IsItemField *bool `json:"isItemField,omitempty"`

	// UI hints
	UI *FieldUI `json:"ui,omitempty"`

	// Validation rules
	Validations []FieldValidationRule `json:"validations,omitempty"`

	// Conditional behavior (CEL)
	//
	// DependsOn, RequiredWhen, and DisabledWhen are stored as CEL-like
	// expression strings describing a condition over other fields'
	// values. As of this schema version, these strings are validated
	// only for presence where required by a given rule shape (for
	// example, FieldValidationRule's CustomValidation requires a
	// non-empty Expression) — they are not parsed against a grammar and
	// are not evaluated by anything in this package. A schema document
	// can therefore declare a condition with a typo or with a reference
	// to a field that does not exist, and Validate will not detect it.
	// See "KNOWN LIMITATIONS" for the current scope of this gap and the
	// structured alternative (Effects) used where expressiveness allows.
	DependsOn    string `json:"dependsOn,omitempty"`
	Required     bool   `json:"required,omitempty"`
	RequiredWhen string `json:"requiredWhen,omitempty"`
	DisabledWhen string `json:"disabledWhen,omitempty"`

	// Dynamic overrides
	Overrides []FieldOverride `json:"overrides,omitempty"`

	// Effects: declarative multi-field write side-effects.
	// When this field's value matches a condition, the listed target
	// fields are set to the specified values. Typically used on virtual
	// selector fields (e.g. auth method dropdown) to drive multiple
	// storage fields without opaque CEL expressions.
	Effects []FieldEffect `json:"effects,omitempty"`

	// Array schema (required when ValueType == array)
	Item *FieldItemSchema `json:"item,omitempty"`

	// Legacy indexed fields
	//
	// Repeatable and Pattern are reserved for describing legacy,
	// hand-rolled indexed-field conventions (for example, a plugin that
	// stores a numbered series of similarly-named properties without
	// using the structured Storage.IndexedPairMapping representation
	// below). As of this schema version, neither is read by Validate or
	// by ToPluginSchemaSettings; see "KNOWN LIMITATIONS".
	Repeatable bool   `json:"repeatable,omitempty"`
	Pattern    string `json:"pattern,omitempty"`

	// Storage mapping layer
	Storage *StorageMapping `json:"storage,omitempty"`

	// Metadata
	//
	// Tags, Examples, and DefaultValue are descriptive metadata with no
	// effect on validation or on the settings produced by
	// ToPluginSchemaSettings, with one exception: DefaultValue is
	// propagated into the generated OpenAPI schema's default value (see
	// convert.go). Tags is free-text and is intended for documentation
	// and lightweight authoring conventions (for example, recording
	// that a field's value is driven by another field's Effects); it is
	// deliberately not validated against any fixed vocabulary and is
	// not intended to gate behavior — see "KNOWN LIMITATIONS". Examples
	// is intended for documentation and assisted-configuration use
	// (showing a chat-driven assistant or a generated doc page a
	// representative valid value) and is not currently consumed by any
	// code in this package.
	Tags         []string `json:"tags,omitempty"`
	Examples     []any    `json:"examples,omitempty"`
	DefaultValue any      `json:"defaultValue,omitempty"`
}

// Validate checks that a single ConfigField is internally well-formed
// for its declared Kind and ValueType.
//
// This includes: required identifying properties are present (ID, Key,
// a valid ValueType); a Target is present whenever required (storage
// fields that are neither virtual nor item fields); Section is not
// used in combination with item fields or virtual fields, since neither
// has a Target for Section to be a path within; array and map fields
// declare an Item schema; any Storage mapping, UI block, Validations,
// Overrides, Effects, and nested Item fields are themselves valid.
func (f *ConfigField) Validate() error {
	if f.ID == "" {
		return fmt.Errorf("field id is required")
	}
	if f.Key == "" {
		return fmt.Errorf("field %s: key is required", f.ID)
	}
	if !f.ValueType.IsValid() {
		return fmt.Errorf("field %s: invalid valueType %q", f.ID, f.ValueType)
	}

	isVirtual := f.Kind == VirtualField
	isItem := f.IsItemField != nil && *f.IsItemField

	if !isVirtual && !isItem && f.Target == nil {
		return fmt.Errorf("field %s: target is required for storage fields", f.ID)
	}

	if f.Section != "" && isItem {
		return fmt.Errorf("field %s: section is not allowed on item fields", f.ID)
	}
	if f.Section != "" && isVirtual {
		return fmt.Errorf("field %s: section is not allowed on virtual fields", f.ID)
	}

	if (f.ValueType == ArrayType || f.ValueType == MapType) && f.Item == nil {
		return fmt.Errorf("field %s: item is required for array and map fields", f.ID)
	}

	if f.Storage != nil {
		if err := f.Storage.Validate(); err != nil {
			return fmt.Errorf("field %s: invalid storage mapping: %w", f.ID, err)
		}
	}

	if f.Kind != "" && !f.Kind.IsValid() {
		return fmt.Errorf("field %s: invalid kind %q", f.ID, f.Kind)
	}

	if f.UI != nil {
		if !f.UI.Component.IsValid() {
			return fmt.Errorf("field %s: invalid ui component %q", f.ID, f.UI.Component)
		}
		if f.UI.Width != "" && !f.UI.Width.IsValid() {
			return fmt.Errorf("field %s: invalid ui width %q", f.ID, f.UI.Width)
		}
		for i, opt := range f.UI.Options {
			if !ValidateOptionValue(opt.Value, f.ValueType) {
				return fmt.Errorf("field %s: ui option[%d] value type mismatch (expected %s)", f.ID, i, f.ValueType)
			}
		}
	}

	if f.Target != nil && !f.Target.IsValid() {
		return fmt.Errorf("field %s: invalid target: %s", f.ID, *f.Target)
	}

	if f.Item != nil {
		if !f.Item.ValueType.IsValid() {
			return fmt.Errorf("field %s: invalid item valueType %q", f.ID, f.Item.ValueType)
		}
		if f.Item.ValueType != ObjectType && len(f.Item.Fields) > 0 {
			return fmt.Errorf("field %s: item fields are only allowed when item valueType is object", f.ID)
		}
		for i := range f.Item.Fields {
			sub := &f.Item.Fields[i]
			if sub.IsItemField == nil || !*sub.IsItemField {
				return fmt.Errorf("field %s: item field %s must have isItemField=true", f.ID, sub.ID)
			}
			if err := sub.Validate(); err != nil {
				return fmt.Errorf("field %s: invalid item field %s: %w", f.ID, sub.ID, err)
			}
		}
	}

	for i := range f.Validations {
		if err := f.Validations[i].Validate(); err != nil {
			return fmt.Errorf("field %s: invalid validation rule: %w", f.ID, err)
		}
	}

	for i := range f.Overrides {
		for j := range f.Overrides[i].Validations {
			if err := f.Overrides[i].Validations[j].Validate(); err != nil {
				return fmt.Errorf("field %s: invalid override validation rule: %w", f.ID, err)
			}
		}
	}

	for i := range f.Effects {
		if err := f.Effects[i].Validate(); err != nil {
			return fmt.Errorf("field %s: invalid effect[%d]: %w", f.ID, i, err)
		}
	}

	return nil
}

// FieldIDs walks the schema (including nested item fields of array and
// map fields) and returns the complete set of declared field IDs.
//
// This is the set ValidateRefs checks every group, relationship, and
// effect reference against. It also detects duplicate IDs: since ID is
// documented as globally unique (see ConfigField), a duplicate is a
// schema-authoring error and is reported as such rather than silently
// overwriting the first occurrence.
func (s *Schema) FieldIDs() (map[string]bool, error) {
	seen := map[string]bool{}

	var visit func(fields []ConfigField) error
	visit = func(fields []ConfigField) error {
		for i := range fields {
			f := fields[i]

			if f.ID == "" {
				return fmt.Errorf("field id is required")
			}

			if seen[f.ID] {
				return fmt.Errorf("duplicate field id: %s", f.ID)
			}
			seen[f.ID] = true

			if f.Item != nil {
				if err := visit(f.Item.Fields); err != nil {
					return err
				}
			}
		}

		return nil
	}

	if err := visit(s.Fields); err != nil {
		return nil, err
	}

	return seen, nil
}

// Path returns the dotted storage path for a field: its Target,
// optionally its Section, and its Key. A field with no Target (a
// virtual field, or an item field, neither of which is independently
// stored) returns just its Key.
//
// This is a convenience for diagnostics and documentation; it is not
// itself used as a reference key anywhere in this package (ID fills
// that role) and it is not the mechanism ToPluginSchemaSettings uses to
// place a field into the generated OpenAPI schema (see placeInSection
// and placeInSectionPath in convert.go, which perform the equivalent
// placement directly against the output schema's property tree).
func (f ConfigField) Path() string {
	if f.Target == nil {
		return f.Key
	}
	if f.Section != "" {
		return string(*f.Target) + "." + f.Section + "." + f.Key
	}
	return string(*f.Target) + "." + f.Key
}

// ============================================================
// Lookup and Value Resolution by ID
// ============================================================
//
// The functions in this section are read-side utilities, not part of
// schema authoring or structural validation. They exist because every
// consumer that needs to go from "a field id" to "that field's
// definition" or "that field's actual configured value" — a config
// editor resolving Schema.Groups' FieldRefs, Grafana Assistant resolving
// an id a person referred to in conversation, a future runtime validator
// reading a real instance settings payload — would otherwise have to
// hand-roll the same tree walk independently. See "KNOWN LIMITATIONS"
// for what these functions do not yet handle.

// FieldByID returns the ConfigField with the given id, searching Fields
// and recursing into the Item.Fields of any array/map field. Returns an
// error if no field with that id exists in the schema.
//
// This does not require Validate to have been called first, but its
// behavior — in particular, which field is returned if the schema
// contains a duplicate id, which Validate would otherwise reject — is
// only well-defined for a structurally valid schema.
func (s *Schema) FieldByID(id string) (*ConfigField, error) {
	var found *ConfigField

	var visit func(fields []ConfigField) bool
	visit = func(fields []ConfigField) bool {
		for i := range fields {
			f := &fields[i]
			if f.ID == id {
				found = f
				return true
			}
			if f.Item != nil && visit(f.Item.Fields) {
				return true
			}
		}
		return false
	}

	if !visit(s.Fields) {
		return nil, fmt.Errorf("no field with id %q", id)
	}
	return found, nil
}

// ValueByID returns the configured value for the field with the given
// id, read out of a real configuration payload. settings follows
// Grafana's existing storage shape: root-level keys at the top level of
// settings, jsonData and secureJsonData as nested maps under those same
// keys.
//
// ValueByID only resolves fields whose Storage is unset or DirectMapping
// (a field's Target/Section/Key correspond directly to one storage
// location). For an IndexedPairMapping field (for example, the legacy
// HTTP header convention), use ResolveIndexedPairs or
// ResolveIndexedPairsAsMap instead — ValueByID returns an error rather
// than guess at a single storage location that does not exist for that
// mapping type. For a ComputedMapping field, ValueByID returns an error
// because evaluating Storage.Read is out of scope for this package (see
// "KNOWN LIMITATIONS").
//
// ValueByID also returns an error for a virtual field (Kind ==
// VirtualField has no storage location to read) and for an item field
// (IsItemField fields describe the shape of each array/map element, not
// a single value at the document level).
func (s *Schema) ValueByID(id string, settings map[string]any) (any, error) {
	f, err := s.FieldByID(id)
	if err != nil {
		return nil, err
	}
	if f.Kind == VirtualField {
		return nil, fmt.Errorf("field %q is virtual and has no stored value", id)
	}
	if f.IsItemField != nil && *f.IsItemField {
		return nil, fmt.Errorf("field %q is an item field; it has no single value at the document level", id)
	}
	if f.Target == nil {
		return nil, fmt.Errorf("field %q has no target", id)
	}

	if f.Storage != nil {
		switch f.Storage.Type {
		case IndexedPairMapping:
			return nil, fmt.Errorf("field %q uses an indexedPair storage mapping; use ResolveIndexedPairs or ResolveIndexedPairsAsMap instead of ValueByID", id)
		case ComputedMapping:
			return nil, fmt.Errorf("field %q uses a computed storage mapping, which is not evaluated by this package", id)
		case DirectMapping:
			// fall through to direct resolution below
		}
	}

	bucket, err := resolveBucket(settings, *f.Target)
	if err != nil {
		return nil, fmt.Errorf("field %q: %w", id, err)
	}
	if f.Section != "" {
		bucket, err = navigateSection(bucket, f.Section)
		if err != nil {
			return nil, fmt.Errorf("field %q: %w", id, err)
		}
	}

	v, ok := bucket[f.Key]
	if !ok {
		return nil, fmt.Errorf("field %q (key %q) not present in configuration", id, f.Key)
	}
	return v, nil
}

// resolveBucket returns the map within settings corresponding to t.
// settings is expected to follow Grafana's existing storage shape:
// root-level keys live at the top level of settings itself; jsonData and
// secureJsonData are nested maps under settings["jsonData"] and
// settings["secureJsonData"] respectively.
func resolveBucket(settings map[string]any, t TargetLocation) (map[string]any, error) {
	switch t {
	case RootTarget:
		return settings, nil
	case JSONDataTarget:
		return nestedMap(settings, "jsonData")
	case SecureJSONTarget:
		return nestedMap(settings, "secureJsonData")
	default:
		return nil, fmt.Errorf("invalid target %q", t)
	}
}

func nestedMap(parent map[string]any, key string) (map[string]any, error) {
	v, ok := parent[key]
	if !ok {
		return map[string]any{}, nil // bucket absent entirely is not an error; it's an empty bucket
	}
	m, ok := v.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%q is present but is not an object", key)
	}
	return m, nil
}

// navigateSection walks a dotted Section path within bucket, returning
// the nested map at the end of that path. This supports exactly the
// nesting depth Section itself supports — see ConfigField.Section and
// "KNOWN LIMITATIONS" regarding the single-level constraint.
func navigateSection(bucket map[string]any, section string) (map[string]any, error) {
	cur := bucket
	for _, seg := range strings.Split(section, ".") {
		next, err := nestedMap(cur, seg)
		if err != nil {
			return nil, fmt.Errorf("section %q: %w", section, err)
		}
		cur = next
	}
	return cur, nil
}

// bucketForTarget is the IndexedPairMapping-specific counterpart of
// resolveBucket: an indexed-pair Key/Value MappingField's Target is only
// ever JSONDataTarget or SecureJSONTarget (RootTarget is not a valid
// target for either half of an indexed pair), so this returns an error
// for any other value rather than silently resolving it.
func bucketForTarget(t TargetLocation, jsonData, secureJSONData map[string]any) (map[string]any, error) {
	switch t {
	case JSONDataTarget:
		return jsonData, nil
	case SecureJSONTarget:
		return secureJSONData, nil
	default:
		return nil, fmt.Errorf("indexedPair target %q must be jsonData or secureJsonData", t)
	}
}

// ResolveIndexedPairs reads an IndexedPairMapping field's actual logical
// value out of a real configuration payload, by scanning for numbered
// key/value pairs starting at Storage.StartIndex and assembling them
// into an array of objects matching the field's Item schema (one object
// per pair, keyed by the item schema's field Keys).
//
// The scan stops at the first missing index — an index 2 that is absent
// while index 3 is present will not see index 3. This faithfully
// reflects what ToPluginSchemaSettings and Grafana's own legacy storage
// convention assume (sequential, gapless numbering); it does not attempt
// to recover from a gap left by a UI that deleted an entry without
// renumbering the rest. Use ResolveIndexedPairsAsMap if gap-tolerance
// matters more than preserving exact pair order and duplicate names; see
// "KNOWN LIMITATIONS" for the trade-off between the two functions.
//
// Returns an empty, non-nil slice (not an error) if no pairs are
// present — an indexedPair field with zero configured entries is a
// normal, valid state.
func ResolveIndexedPairs(f *ConfigField, jsonData, secureJSONData map[string]any) ([]map[string]any, error) {
	if f.Storage == nil || f.Storage.Type != IndexedPairMapping {
		return nil, fmt.Errorf("field %q is not an indexedPair field", f.ID)
	}
	mapping := f.Storage

	keyBucket, err := bucketForTarget(mapping.Key.Target, jsonData, secureJSONData)
	if err != nil {
		return nil, fmt.Errorf("field %q: resolving key bucket: %w", f.ID, err)
	}
	valueBucket, err := bucketForTarget(mapping.Value.Target, jsonData, secureJSONData)
	if err != nil {
		return nil, fmt.Errorf("field %q: resolving value bucket: %w", f.ID, err)
	}

	start := 1
	if mapping.StartIndex != nil {
		start = *mapping.StartIndex
	}

	// Which item field is the pair's "name" and which is its "value" is
	// inferred by position (first declared item field = name, second =
	// value), because ConfigField/FieldItemSchema have no explicit
	// pair-role tag today. See "KNOWN LIMITATIONS" — a schema that
	// declares its two item fields in the opposite order will silently
	// produce swapped results, with no validation error.
	if f.Item == nil || len(f.Item.Fields) < 2 {
		return nil, fmt.Errorf("field %q: indexedPair requires an item schema with at least 2 fields", f.ID)
	}
	nameFieldKey := f.Item.Fields[0].Key
	valueFieldKey := f.Item.Fields[1].Key

	var results []map[string]any
	for i := start; ; i++ {
		nameKey := strings.ReplaceAll(mapping.Key.Pattern, "{index}", strconv.Itoa(i))
		nameVal, ok := keyBucket[nameKey]
		if !ok {
			break
		}

		item := map[string]any{nameFieldKey: nameVal}

		valueKey := strings.ReplaceAll(mapping.Value.Pattern, "{index}", strconv.Itoa(i))
		if v, ok := valueBucket[valueKey]; ok {
			item[valueFieldKey] = v
		}
		// Value side absent (most commonly because it targets
		// secureJsonData and the caller's settings payload is a
		// live, already-saved datasource's settings rather than a
		// schema example — secureJsonData is write-only and is never
		// returned by Grafana's own API once saved) leaves item with
		// only its name key set, rather than being treated as an error.

		results = append(results, item)
	}

	if results == nil {
		results = []map[string]any{}
	}
	return results, nil
}

// ResolveIndexedPairsAsMap reads an IndexedPairMapping field's configured
// pairs by scanning every key actually present in the key bucket —
// rather than stopping at the first missing index — extracting each
// key's numeric index via the mapping's Key.Pattern, and returning a
// flat name -> value map. A name with no corresponding value present in
// the value bucket maps to an empty string rather than being omitted.
//
// This function is gap-tolerant where ResolveIndexedPairs is not: a
// stored configuration with index 2 deleted but index 3 still present is
// resolved correctly here. The trade-off, see "KNOWN LIMITATIONS", is
// that this shape cannot represent two distinct pairs that happen to
// share the same name (a later index silently overwrites an earlier one
// in the returned map, with map iteration order — and therefore which
// one wins — unspecified), and an empty-string value is indistinguishable
// from "the value side is genuinely unset or unreadable."
func ResolveIndexedPairsAsMap(f *ConfigField, jsonData, secureJSONData map[string]any) (map[string]string, error) {
	if f.Storage == nil || f.Storage.Type != IndexedPairMapping {
		return nil, fmt.Errorf("field %q is not an indexedPair field", f.ID)
	}
	mapping := f.Storage

	keyBucket, err := bucketForTarget(mapping.Key.Target, jsonData, secureJSONData)
	if err != nil {
		return nil, fmt.Errorf("field %q: resolving key bucket: %w", f.ID, err)
	}
	valueBucket, err := bucketForTarget(mapping.Value.Target, jsonData, secureJSONData)
	if err != nil {
		return nil, fmt.Errorf("field %q: resolving value bucket: %w", f.ID, err)
	}

	keyRe, err := patternToIndexRegex(mapping.Key.Pattern)
	if err != nil {
		return nil, fmt.Errorf("field %q: invalid key pattern: %w", f.ID, err)
	}

	result := map[string]string{}
	for storedKey, storedVal := range keyBucket {
		m := keyRe.FindStringSubmatch(storedKey)
		if m == nil {
			continue
		}
		index := m[1]

		name, ok := storedVal.(string)
		if !ok {
			continue
		}

		valueKey := strings.ReplaceAll(mapping.Value.Pattern, "{index}", index)
		value := ""
		if v, ok := valueBucket[valueKey]; ok {
			if s, ok := v.(string); ok {
				value = s
			}
		}

		result[name] = value
	}

	return result, nil
}

// patternToIndexRegex converts a storage pattern such as
// "httpHeaderName{index}" into a regular expression that matches real
// stored keys and captures the numeric index as its first group.
func patternToIndexRegex(pattern string) (*regexp.Regexp, error) {
	escaped := regexp.QuoteMeta(pattern)
	escaped = strings.ReplaceAll(escaped, regexp.QuoteMeta("{index}"), `(\d+)`)
	return regexp.Compile("^" + escaped + "$")
}

// ============================================================
// Array Item Schema
// ============================================================

// FieldItemSchema defines schema for array/map elements.
// For arrays, it describes each element.
// For maps, it describes each value (keys are always strings).
//
// An array or map field's own Target/Key/Section describe where the
// collection as a whole is stored; FieldItemSchema describes the shape
// of each element/value within that collection. When ValueType is
// ObjectType, Fields describes the element's own properties, each of
// which must be marked IsItemField (see ConfigField.Validate) since an
// item field's storage location is inherited from its parent collection
// rather than declared independently.
type FieldItemSchema struct {
	ValueType ValueType     `json:"valueType"`
	Fields    []ConfigField `json:"fields,omitempty"`
}

// ============================================================
// Value Types
// ============================================================

// ValueType enumerates the primitive and structural types a
// ConfigField's value may take.
//
// These map directly onto JSON's own type system (string, number,
// boolean, array, object) with two additions: MapType, for an object
// with dynamic string keys whose values share one schema (see
// FieldItemSchema), and AnyType, for fields whose value may legitimately
// take more than one shape and which therefore opt out of type-level
// validation (see ConfigField's "Any fields" usage; intended to be used
// sparingly, only where a single type genuinely cannot describe the
// data).
type ValueType string

const (
	StringType  ValueType = "string"
	NumberType  ValueType = "number"
	BooleanType ValueType = "boolean"
	ArrayType   ValueType = "array"
	ObjectType  ValueType = "object"
	MapType     ValueType = "map"
	AnyType     ValueType = "any"
)

func (v ValueType) IsValid() bool {
	switch v {
	case StringType, NumberType, BooleanType, ArrayType, ObjectType, MapType, AnyType:
		return true
	default:
		return false
	}
}

// ============================================================
// Field Kind
// ============================================================

// FieldKind distinguishes fields that are actually persisted
// (StorageField, the default) from fields that exist only to describe
// computed or UI-only state and are never written to root, jsonData,
// or secureJsonData (VirtualField). See ConfigField's discussion of
// "Field Kind: Storage Versus Virtual" for the canonical use of
// VirtualField alongside Effects.
type FieldKind string

const (
	StorageField FieldKind = "storage"
	VirtualField FieldKind = "virtual"
)

func (k FieldKind) IsValid() bool {
	switch k {
	case StorageField, VirtualField:
		return true
	default:
		return false
	}
}

// ============================================================
// Target Location
// ============================================================

// TargetLocation enumerates the storage locations Grafana's existing
// datasource configuration model already provides. This package
// introduces no storage location beyond these three, by design — see
// the package-level documentation's "Design Posture: Additive, Not
// Migratory" discussion.
type TargetLocation string

const (
	// RootTarget is a top-level property of the datasource resource
	// itself (for example, url, basicAuth, database) — the same
	// properties Grafana's datasource model, provisioning format, and
	// HTTP API have always exposed at the top level, outside jsonData
	// and secureJsonData.
	RootTarget TargetLocation = "root"

	// JSONDataTarget is a property within the datasource's jsonData
	// map: free-form, plugin-defined, non-secret configuration.
	JSONDataTarget TargetLocation = "jsonData"

	// SecureJSONTarget is a property within the datasource's
	// secureJsonData map: encrypted-at-rest, write-only configuration.
	// A field with this target describes what may be written; the
	// value cannot be read back through this schema or through
	// Grafana's existing API once saved. See the discussion of
	// secureJsonFields under "KNOWN LIMITATIONS" for how the read-side
	// "is this secret configured" signal relates to fields described
	// here.
	SecureJSONTarget TargetLocation = "secureJsonData"
)

func (t TargetLocation) IsValid() bool {
	switch t {
	case RootTarget, JSONDataTarget, SecureJSONTarget:
		return true
	default:
		return false
	}
}

// ============================================================
// UI Components
// ============================================================

// UIComponent enumerates the form control types a config editor
// renderer may use to present a field. This is a closed set by design:
// an unrecognized value is a schema-authoring error (see
// ConfigField.Validate), not a silently-ignored hint, since a renderer
// that does not recognize a component value has no defined fallback
// behavior to fall back to.
type UIComponent string

const (
	UIInput       UIComponent = "input"
	UITextarea    UIComponent = "textarea"
	UISelect      UIComponent = "select"
	UIMultiselect UIComponent = "multiselect"
	UIRadio       UIComponent = "radio"
	UICheckbox    UIComponent = "checkbox"
	UISwitch      UIComponent = "switch"
	UICode        UIComponent = "code"
	UIKeyValue    UIComponent = "keyvalue"
	UIList        UIComponent = "list"
)

func (c UIComponent) IsValid() bool {
	switch c {
	case UIInput, UITextarea, UISelect, UIMultiselect, UIRadio,
		UICheckbox, UISwitch, UICode, UIKeyValue, UIList:
		return true
	default:
		return false
	}
}

// FieldUI defines UI rendering hints.
//
// FieldUI is presentation metadata. It has no effect on validation or
// on the OpenAPI settings produced by ToPluginSchemaSettings, with one
// documented exception: applyUIEnum (convert.go) derives an OpenAPI
// enum from Options when the field has no explicit AllowedValuesValidation
// rule, as a convenience for authors who would otherwise have to state
// the same allowed-values list twice. See SCHEMA-V1.md's discussion of why
// Validations, not UI.Options, is the data contract and UI.Options is
// presentation only.
type FieldUI struct {
	Component UIComponent `json:"component"`

	Multiline bool          `json:"multiline,omitempty"`
	Rows      int           `json:"rows,omitempty"`
	Options   []FieldOption `json:"options,omitempty"`

	AllowCustom bool    `json:"allowCustom,omitempty"`
	Width       UIWidth `json:"width,omitempty"`

	Placeholder string `json:"placeholder,omitempty"`

	// Language hint for code editor components.
	// Example: "promql", "logql", "traceql", "sql", "json"
	Language string `json:"language,omitempty"`
}

// UIWidth defines layout width.
type UIWidth string

const (
	FullWidth UIWidth = "full"
	HalfWidth UIWidth = "half"
)

func (w UIWidth) IsValid() bool {
	switch w {
	case FullWidth, HalfWidth:
		return true
	default:
		return false
	}
}

// ============================================================
// Validations
// ============================================================

// ValidationRuleType enumerates the kinds of validation rule this
// package can express. This is the schema's data contract — see
// SCHEMA-V1.md — distinct from and authoritative over any allowed-values
// list that may also be present in a field's UI.Options for display
// purposes.
type ValidationRuleType string

const (
	PatternValidation       ValidationRuleType = "pattern"
	RangeValidation         ValidationRuleType = "range"
	LengthValidation        ValidationRuleType = "length"
	ItemCountValidation     ValidationRuleType = "itemCount"
	AllowedValuesValidation ValidationRuleType = "allowedValues"
	CustomValidation        ValidationRuleType = "custom"
)

// FieldValidationRule is a discriminated union of validation rules.
//
// Exactly one of the type-specific property groups below is meaningful
// for a given rule, selected by Type; see Validate for which properties
// each Type requires. As of this schema version, FieldValidationRule's
// structural shape is validated (the right properties are present for
// the declared Type), and PatternValidation, RangeValidation,
// LengthValidation, ItemCountValidation, and AllowedValuesValidation are
// further translated into real OpenAPI/JSON Schema constraints by
// ToPluginSchemaSettings (convert.go) — pattern, minimum/maximum,
// minLength/maxLength, minItems/maxItems, and enum, respectively.
// CustomValidation's Expression is stored but not evaluated by anything
// in this package; see "KNOWN LIMITATIONS".
type FieldValidationRule struct {
	Type    ValidationRuleType `json:"type"`
	ID      string             `json:"id,omitempty"`
	Message string             `json:"message,omitempty"`

	// PatternValidation
	Pattern string `json:"pattern,omitempty"`

	// RangeValidation / LengthValidation / ItemCountValidation
	Min *float64 `json:"min,omitempty"`
	Max *float64 `json:"max,omitempty"`

	// AllowedValuesValidation
	Values []any `json:"values,omitempty"`

	// CustomValidation
	Expression string `json:"expression,omitempty"`
}

func (r *FieldValidationRule) Validate() error {
	switch r.Type {
	case PatternValidation:
		if r.Pattern == "" {
			return fmt.Errorf("pattern validation requires pattern")
		}
	case RangeValidation, LengthValidation, ItemCountValidation:
		if r.Min == nil && r.Max == nil {
			return fmt.Errorf("%s validation requires min or max", r.Type)
		}
	case AllowedValuesValidation:
		if len(r.Values) == 0 {
			return fmt.Errorf("allowedValues validation requires values")
		}
	case CustomValidation:
		if r.Expression == "" {
			return fmt.Errorf("custom validation requires expression")
		}
	default:
		return fmt.Errorf("unknown validation rule type: %s", r.Type)
	}
	return nil
}

// ============================================================
// Overrides
// ============================================================

// FieldOverride allows dynamic modifications.
//
// An override describes how a field's presentation or validation
// should change under a stated condition (When), without duplicating
// the entire field definition. As with DependsOn/RequiredWhen/
// DisabledWhen, When is a CEL-like expression string that is not
// currently parsed or evaluated by this package; see "KNOWN
// LIMITATIONS". Overrides' nested Validations are independently
// structurally validated (see ConfigField.Validate), but — as a
// consequence of When not being evaluated — no override is currently
// applied to the OpenAPI settings produced by ToPluginSchemaSettings.
type FieldOverride struct {
	When string `json:"when"`

	DefaultValue any    `json:"defaultValue,omitempty"`
	Description  string `json:"description,omitempty"`
	Placeholder  string `json:"placeholder,omitempty"`
	Tooltip      string `json:"tooltip,omitempty"`

	Validations []FieldValidationRule `json:"validations,omitempty"`
	Options     []FieldOption         `json:"options,omitempty"`
}

// ============================================================
// Effects
// ============================================================

// FieldEffect declares that when a field's value matches a condition,
// the listed target fields should be set to the specified values.
//
// This provides a structured, machine-readable alternative to opaque
// computed write expressions for virtual selector fields.
//
// Example: an auth method dropdown that sets root.basicAuth and
// jsonData.oauthPassThru depending on which option is selected.
//
// Effects is deliberately structured rather than expressed as a CEL
// write expression: the set of "selector value picked -> these fields
// get these values" relationships that occur in practice is small and
// enumerable, and representing it as a validated when/set pair (rather
// than an opaque string naming a side-effecting function) lets
// ValidateRefs confirm every Set key resolves to a real field ID
// without needing to parse or evaluate the When condition itself to do
// so. When remains a CEL-like string today and is not evaluated by this
// package; only its presence is checked. See "KNOWN LIMITATIONS".
type FieldEffect struct {
	// When is a CEL expression evaluated against the field's value.
	// Convention: use "value" to refer to the field's current value.
	// Example: "value == 'basic-auth'"
	When string `json:"when"`

	// Set maps field IDs to the values they should be set to when
	// the condition matches.
	Set map[string]any `json:"set"`
}

func (e *FieldEffect) Validate() error {
	if e.When == "" {
		return fmt.Errorf("effect when is required")
	}
	if len(e.Set) == 0 {
		return fmt.Errorf("effect set must not be empty")
	}
	return nil
}

// ============================================================
// Storage Mapping
// ============================================================

// StorageMappingType enumerates how a logical field maps onto Grafana's
// existing storage representation when that mapping is not a simple
// one-to-one property (DirectMapping, the default and the common case).
type StorageMappingType string

const (
	// DirectMapping is the default: the field's Target and Key map
	// directly onto a single property in that target's storage
	// location. A field with no explicit Storage is implicitly
	// DirectMapping.
	DirectMapping StorageMappingType = "direct"

	// IndexedPairMapping describes Grafana's existing legacy convention
	// for representing a user-extensible list of name/value pairs as a
	// numbered series of individual properties (for example,
	// httpHeaderName1/httpHeaderValue1, httpHeaderName2/
	// httpHeaderValue2, and so on), optionally with the name and value
	// halves of each pair stored in different targets — the documented
	// convention for HTTP headers, where names are not secret
	// (jsonData) but values may be (secureJsonData). This mapping type
	// describes that existing convention; it does not change it. See
	// "KNOWN LIMITATIONS" for the current scope of what reads this
	// mapping today.
	IndexedPairMapping StorageMappingType = "indexedPair"

	// ComputedMapping describes a field whose stored representation is
	// derived from, or split across, other fields via a read and/or
	// write expression, rather than corresponding to a single stored
	// property. As with other CEL-like expression fields in this
	// package, Read and Write are stored as strings and are not
	// evaluated by anything in this package as of this schema version;
	// see "KNOWN LIMITATIONS".
	ComputedMapping StorageMappingType = "computed"
)

// StorageMapping maps logical fields to Grafana storage.
type StorageMapping struct {
	Type StorageMappingType `json:"type"`

	// Indexed pair mapping
	Key        *MappingField `json:"key,omitempty"`
	Value      *MappingField `json:"value,omitempty"`
	StartIndex *int          `json:"startIndex,omitempty"`

	// Computed mapping
	Read  string `json:"read,omitempty"`
	Write string `json:"write,omitempty"`
}

// Validate checks that a StorageMapping's populated properties are
// consistent with its declared Type — for example, that an
// IndexedPairMapping supplies both Key and Value mapping fields and
// does not also supply Read/Write (which belong only to
// ComputedMapping), and vice versa.
func (m *StorageMapping) Validate() error {
	switch m.Type {
	case DirectMapping:
		if m.Key != nil || m.Value != nil || m.StartIndex != nil || m.Read != "" || m.Write != "" {
			return fmt.Errorf("direct mapping must not have key/value/startIndex/read/write")
		}

	case IndexedPairMapping:
		if m.Key == nil || m.Value == nil {
			return fmt.Errorf("indexedPair requires key and value")
		}
		if m.Read != "" || m.Write != "" {
			return fmt.Errorf("indexedPair must not have read/write")
		}
		if err := m.Key.Validate(); err != nil {
			return fmt.Errorf("indexedPair key: %w", err)
		}
		if err := m.Value.Validate(); err != nil {
			return fmt.Errorf("indexedPair value: %w", err)
		}

	case ComputedMapping:
		if m.Read == "" && m.Write == "" {
			return fmt.Errorf("computed mapping requires read or write")
		}
		if m.Key != nil || m.Value != nil || m.StartIndex != nil {
			return fmt.Errorf("computed mapping must not have key/value/startIndex")
		}

	default:
		return fmt.Errorf("unknown mapping type: %s", m.Type)
	}

	return nil
}

// MappingField describes one half (key or value) of an IndexedPairMapping:
// which storage Target it lives in, and the naming Pattern used to
// generate each numbered property name (for example,
// "httpHeaderName{index}").
type MappingField struct {
	Target  TargetLocation `json:"target"`
	Pattern string         `json:"pattern"`
}

func (m MappingField) Validate() error {
	if !m.Target.IsValid() {
		return fmt.Errorf("invalid target %q", m.Target)
	}
	if m.Pattern == "" {
		return fmt.Errorf("pattern is required")
	}
	return nil
}

// ============================================================
// Options
// ============================================================

// FieldOption describes one selectable choice for a select/radio/
// multiselect UI component, or one entry in an AllowedValuesValidation
// rule's Values list.
type FieldOption struct {
	Label       string `json:"label"`
	Value       any    `json:"value"`
	Description string `json:"description,omitempty"`
}

// ValidateOptionValue checks that an option value is non-nil and
// compatible with the given field valueType.
func ValidateOptionValue(v any, vt ValueType) bool {
	if v == nil {
		return false
	}
	switch vt {
	case StringType:
		_, ok := v.(string)
		return ok
	case NumberType:
		switch v.(type) {
		case int, int64, float64, float32:
			return true
		default:
			return false
		}
	case BooleanType:
		_, ok := v.(bool)
		return ok
	default:
		// array/object/map/any options are not type-checked
		return true
	}
}

// ============================================================
// Groups
// ============================================================

// ConfigGroup describes a presentational grouping of fields — for
// example, a collapsible "Advanced" section in a generated config
// editor. Groups are pure UI layout metadata; see Schema.Groups for the
// full discussion of what depends on this (nothing structural) and what
// does not (storage, validation, the OpenAPI settings produced by
// ToPluginSchemaSettings).
type ConfigGroup struct {
	ID          string   `json:"id"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Order       *int     `json:"order,omitempty"`
	Optional    bool     `json:"optional,omitempty"`
	FieldRefs   []string `json:"fieldRefs"`
}

// ============================================================
// Relationships
// ============================================================

// RelationshipType enumerates the kinds of semantic connection a
// FieldRelationship may declare between fields.
type RelationshipType string

const (
	// PairRelationship connects two fields that together form one
	// logical unit of configuration — most commonly a username and a
	// password.
	PairRelationship RelationshipType = "pair"

	// GroupRelationship connects an arbitrary set of fields that are
	// semantically related but do not fit the narrower pair shape.
	GroupRelationship RelationshipType = "group"

	// DatasourceRefRelationship marks that one or more fields hold a
	// reference (typically a UID) to another Grafana datasource — for
	// example, a "derived fields" configuration that links log lines to
	// a tracing datasource. TargetPluginType, when set, constrains
	// which plugin type the referenced datasource UID is expected to
	// resolve to.
	DatasourceRefRelationship RelationshipType = "datasourceReference"
)

func (r RelationshipType) IsValid() bool {
	switch r {
	case PairRelationship, GroupRelationship, DatasourceRefRelationship:
		return true
	default:
		return false
	}
}

// FieldRelationship describes a semantic, non-structural connection
// between two or more fields. Like ConfigGroup, a relationship carries
// no effect on storage or on the OpenAPI settings produced by
// ToPluginSchemaSettings; it exists for renderers and automated/
// assisted configuration consumers that benefit from knowing fields are
// related (for example, presenting a username/password pair together,
// or warning that a datasource-reference field should be checked
// against the referenced datasource's existence).
type FieldRelationship struct {
	Type        RelationshipType `json:"type"`
	Fields      []string         `json:"fields"`
	Description string           `json:"description,omitempty"`

	// TargetPluginType constrains the datasource UID to a specific plugin.
	// Only applicable when Type is "datasourceReference".
	TargetPluginType string `json:"targetPluginType,omitempty"`
}

// Instruction is a structured, free-form guidance entry intended
// primarily for non-human or semi-autonomous consumers of the schema —
// most directly, a chat-driven configuration assistant (such as the
// Grafana Assistant) that needs plugin-specific guidance not otherwise
// captured by individual field definitions in order to help a person
// reach a working datasource connection in as few exchanges as
// possible. Tags allows an Instruction to be scoped or categorized at
// the author's discretion; neither Message nor Tags is validated
// against a fixed vocabulary, and neither has any effect on storage or
// on the OpenAPI settings produced by ToPluginSchemaSettings.
type Instruction struct {
	Message string   `json:"msg"`
	Tags    []string `json:"tags,omitempty"`
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
// schema document. See the accompanying RFC for the proposed sequencing
// of this work.
//
//  1. Expression strings are not parsed or evaluated. DependsOn,
//     RequiredWhen, DisabledWhen, FieldOverride.When, FieldEffect.When,
//     StorageMapping's Read/Write, and CustomValidation's Expression are
//     all stored as opaque, CEL-flavored strings. None of them is parsed
//     against a grammar at Validate time, and none is evaluated against
//     real configuration data by anything in this package. A typo or a
//     reference to a nonexistent field inside one of these strings is
//     not detected until — at the earliest — some future runtime
//     evaluator exists; today, it is never detected.
//
//  2. StorageMapping is descriptive metadata, not yet an executable
//     mapping. IndexedPairMapping and ComputedMapping are validated
//     structurally (the right sub-properties are present for the
//     declared Type), but no code in this package reads a
//     StorageMapping and applies it against real stored configuration —
//     for example, expanding an IndexedPairMapping's Pattern against an
//     actual jsonData/secureJsonData payload to produce a real list of
//     name/value pairs.
//
//  3. ToPluginSchemaSettings (convert.go) does not yet read Storage at
//     all; it places fields by Target and Section only. A consequence
//     specific to IndexedPairMapping: when a pair's Value.Target is
//     SecureJSONTarget, the generated OpenAPI settings give no
//     indication that the corresponding array's values are secret. This
//     is a correctness gap in the generated output, not merely a missing
//     feature, and is the highest-priority item in this list.
//
//  4. Section supports exactly one level of nesting. A field's Section
//     is a single dotted path, and placeInSectionPath (convert.go)
//     creates exactly the intermediate object levels that path implies
//     — but the modeling pattern this package documents for
//     self-referential structures (see SCHEMA-V1.md's discussion of
//     recursive types) relies on manually flattening each level of
//     recursion into its own Section value, which only remains
//     practical while the actual recursion depth in real plugin
//     configurations stays small.
//
//  5. No field carries a semantic role independent of its name. Two
//     fields that mean the same thing (for example, a base URL) but are
//     named differently across plugins (apiUrl, baseURL, endpoint)
//     cannot currently be recognized as equivalent by any automated
//     consumer without that consumer hard-coding plugin-specific name
//     lists. This directly limits the reliability of any automated HTTP
//     client derivation: without a name-independent way to identify
//     "this field is the TLS client certificate" or "this field is the
//     basic-auth password," a generic client-builder cannot be written
//     against ConfigField alone today, and must instead be told,
//     per plugin, which fields mean what. This package's Target, Key,
//     and ValueType remain stable, additive points of attachment for a
//     future role-style annotation; no field defined above needs to
//     change shape to support one.
//
//  6. Auth representation is whatever the plugin author chose, with no
//     schema-level distinction between an explicit discriminator field,
//     a set of independently-toggleable boolean flags, or a hybrid of
//     the two. This package can describe any of the three shapes (an
//     explicit-enum virtual field with Effects; independent boolean
//     storage fields; a boolean gating a Section of further fields) but
//     does not yet provide any shared mechanism for recognizing,
//     across plugins, which shape a given schema uses, or for detecting
//     that two boolean flags represent mutually incompatible auth
//     mechanisms when both are set true at once.
//
//  7. No mechanism exists for a plugin with more than one independent
//     connection (for example, a plugin that calls two unrelated
//     backend APIs, each with its own URL, auth, and TLS settings,
//     within a single datasource instance). Every field in a Schema is
//     implicitly part of one undifferentiated configuration surface;
//     there is no way to declare that a subset of fields belongs to one
//     logical connection and a different subset to another.
//
//  8. id format is now partially enforced. ValidateFieldIDFormat (called
//     from Validate) rejects an id containing a segment outside
//     [A-Za-z_][A-Za-z0-9_]* and rejects one id being a strict
//     dotted-path prefix of another. This closes the specific,
//     previously silent failure mode of an id that would break a future
//     CEL-style evaluator's dotted-path resolution. It does NOT enforce
//     the recommended hierarchical-by-meaning convention beyond that —
//     two unrelated ids, each individually well-formed and
//     non-conflicting, are still accepted regardless of whether their
//     segments reflect any consistent naming scheme across the document.
//     Any schema document written before this check existed that
//     happens to violate either rule will now fail Validate; this is a
//     deliberate, newly-enforced behavior change within the v1 schema
//     version, not a v1-to-v2 migration (see item 11).
//
//  9. tags and examples are accepted and stored but not read by any
//     code in this package. They are retained as forward-compatible,
//     purely descriptive metadata; see ConfigField's documentation of
//     Tags and Examples.
//
//  10. repeatable and pattern (on ConfigField, distinct from
//      MappingField.Pattern) are accepted and stored but not read by
//      Validate or by ToPluginSchemaSettings. Their relationship to
//      StorageMapping's IndexedPairMapping — which models the same
//      legacy indexed-field convention more explicitly — is not yet
//      resolved.
//
//  11. There is no schema-version migration mechanism. SchemaVersion is
//      a required string, but no code in this package interprets it
//      beyond requiring its presence; there is no defined behavior for
//      reading a Schema document written against a different version of
//      this package's types than the one doing the reading.
//
//  12. Validate returns on the first error encountered, rather than
//      collecting every validation failure in a document before
//      returning. A companion TypeScript validator for this same schema
//      shape (see SCHEMA-V1.ts) is documented as collecting all errors,
//      which means a single invalid schema document can currently
//      surface a different number of reported problems depending on
//      which language validated it.
//
//  13. ResolveIndexedPairs infers which of a field's two declared Item
//      fields is the pair's "name" and which is its "value" by their
//      position in Item.Fields (first = name, second = value), because
//      neither ConfigField nor FieldItemSchema carries an explicit
//      pair-role tag. A schema that declares these two item fields in
//      the opposite order produces silently swapped results, with no
//      validation error at Validate time or at resolution time. Adding
//      an explicit pair-role marker to item fields used inside an
//      IndexedPairMapping's Item schema would close this gap; no field
//      already defined needs to change shape to support one being added
//      additively.
//
//  14. ResolveIndexedPairsAsMap collapses two distinct stored pairs that
//      happen to share the same name into one map entry, with Go's
//      unspecified map iteration order determining which one survives.
//      It also returns an empty string for a name whose corresponding
//      value is absent, which is indistinguishable in the returned map
//      from a pair whose value was genuinely configured as an empty
//      string. ResolveIndexedPairs does not have either limitation (it
//      preserves every pair as a distinct array entry, and an absent
//      value is represented by the value key being absent from that
//      entry's map rather than as an empty string) but is not
//      gap-tolerant in the way ResolveIndexedPairsAsMap is — see each
//      function's own documentation for this trade-off.
//
//  15. Neither ResolveIndexedPairs nor ResolveIndexedPairsAsMap, nor
//      ValueByID, can read a value that targets SecureJSONTarget out of
//      a real, already-saved datasource's settings — secureJsonData is
//      write-only once saved (see SecureJSONTarget) and Grafana's own
//      API never returns it. Calling these functions against such a
//      payload for a secret-targeted field produces the same "absent"
//      result as a field that was simply never configured; the two
//      cases are not distinguishable from these functions' return values
//      alone. These functions work as expected against a schema's own
//      SettingsExamples or any other payload that genuinely embeds
//      secret values, such as a payload under direct, local test
//      construction.
