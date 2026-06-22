// Package dsconfig — v2 additions.
//
// # IMPLEMENTATION NOTE — READ THIS FIRST
//
// This file adds v2's new capabilities (Scopes, Role, RoleConflicts,
// PairRole) without modifying SCHEMA-V1.go's ConfigField struct in any way.
// This is a deliberate constraint of this deliverable, not the shape a
// real v2 release should ship with — and the difference matters enough
// to state plainly before anything else in this file.
//
// Go does not allow a struct's field list to be declared across two
// files: ConfigField is defined once, in SCHEMA-V1.go, and this file
// cannot add fields directly to it without editing that definition. The
// idiomatic, intended shape of dsconfig v2 — the shape this file's
// companion RFC (GRF-RFC-0043) actually proposes for adoption — is for
// Scopes, Role, RoleConflicts, and PairRole to become four new, optional,
// `omitempty` fields declared directly on ConfigField in SCHEMA-V1.go,
// exactly the same way every v1 field was already designed to be
// extended (see SCHEMA-V1.go's own package documentation: "adopting
// dsconfig must never require a plugin to change what it stores").
// Adding four optional fields to an existing struct is itself a
// purely additive change by Go's own semantics — no existing field
// moves, no JSON tag changes, no existing caller breaks — and is
// the standard way this kind of evolution is done in this language.
//
// This file instead represents that same information in a side table —
// FieldExtensionV2, keyed by the same ConfigField.ID v1 already treats
// as the stable, canonical reference — specifically so that this
// deliverable can be reviewed and exercised without modifying
// SCHEMA-V1.go, per the constraint this round of work was produced under.
// Every function in this file that needs a field's v2 metadata looks it
// up by ID in that side table rather than reading it directly off a
// ConfigField value. Where the real upstream change (editing SCHEMA-V1.go)
// happens, every function below collapses to reading the four new
// fields directly off ConfigField, and the side table disappears
// entirely — nothing about this file's external behavior is meant to
// differ between the two shapes; the side table is a faithful,
// behavior-preserving stand-in for fields that belong on ConfigField
// itself.
//
// # WHAT v2 ADDS, AND WHY
//
// Three genuinely new pieces of schema surface, each closing a gap v1's
// own "Known Limitations" section (SCHEMA-V1.go) named explicitly:
//
//  1. Scopes — closes v1 Known Limitation 7 (no representation for a
//     plugin with more than one independent connection within a single
//     datasource instance). A field with no scope entry in its
//     extension is implicitly shared across every declared scope,
//     including any scope declared in a future revision of the same
//     plugin's schema — so a v1-shaped, single-connection schema needs
//     no extension entries at all to remain fully valid and fully
//     equivalent under v2.
//
//  2. Role and RoleConflicts — closes v1 Known Limitation 5 (no field
//     carries a semantic role independent of its name). A fixed,
//     versioned vocabulary lets a field declare what it means — "this
//     is the TLS client certificate" — independent of its Key.
//     RoleConflicts lets a field declare which other roles cannot
//     simultaneously be active alongside it.
//
//  3. PairRole — closes v1 Known Limitations 13 and 14
//     (ResolveIndexedPairs/ResolveIndexedPairsAsMap infer which item
//     field is a pair's name and which is its value by declaration
//     position, silently producing swapped results if that order is
//     wrong). PairRole, when declared, is read in preference to
//     positional inference.
//
// # WHAT v2 DELIBERATELY DOES NOT DO
//
// Consistent with this package's established discipline of describing a
// capability before executing it, v2 can DESCRIBE a multi-connection
// plugin, but ToPluginSchemaSettings (convert.go) does not produce a
// multi-connection-aware App Platform artifact — seeSCHEMA-V1.go— it
// continues to emit exactly one Settings{Spec, SecureValues} pair per
// schema document. v2 also does not implement expression evaluation;
// RoleConflicts is a structural, finite-vocabulary check, not a general
// evaluator. See "KNOWN LIMITATIONS (v2)" below for the complete
// accounting.
package dsconfig

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

// ============================================================
// Schema Version
// ============================================================

// SchemaVersionV2 is the value Schema.SchemaVersion should carry for a
// document that uses any v2-only capability (Scopes, Role,
// RoleConflicts, or PairRole, via the FieldExtensionV2 side table below).
// A document using none of these may continue to declare "v1" and is
// completely unaffected by anything in this file.
const SchemaVersionV2 = "v2"

// ============================================================
// Side-Table Extension Model
// ============================================================
//
// SchemaV2Extensions carries every piece of v2 metadata for one Schema
// document, keyed by ConfigField.ID. This is the side table described
// in the implementation note above: a stand-in for what would otherwise
// be four new fields directly on ConfigField. A SchemaV2Extensions value
// is meaningless on its own — it is always interpreted relative to a
// specific *Schema value, matched up by ID.
type SchemaV2Extensions struct {
	// ScopeDefs declares every named scope this plugin's fields may
	// belong to. A document with no multi-connection needs leaves this
	// empty; every field is then implicitly shared across the
	// document's single, implicit scope, exactly as in v1.
	ScopeDefs []ScopeDef `json:"scopeDefs,omitempty"`

	// Fields maps a ConfigField.ID to that field's v2 extension. A field
	// with no entry here has no Scopes, no Role, no RoleConflicts, and
	// (if it is an item field) no PairRole — it behaves exactly as it
	// did under v1.
	Fields map[string]FieldExtensionV2 `json:"fields,omitempty"`
}

// FieldExtensionV2 carries one field's v2-only metadata. In the upstream
// shape this deliverable stands in for, every property here would
// instead be a field directly on ConfigField.
type FieldExtensionV2 struct {
	// Scopes is this field's scope membership. Empty/absent means the
	// field is shared across every scope declared in the owning
	// SchemaV2Extensions.ScopeDefs, including any scope declared later.
	// See validateFieldScopes for the validation rules this implies.
	Scopes []string `json:"scopes,omitempty"`

	// Role is this field's semantic meaning, drawn from the Role
	// vocabulary below. Empty means the field carries no declared role.
	Role Role `json:"role,omitempty"`

	// RoleConflicts lists roles that must not be simultaneously active
	// alongside this field's own Role.
	RoleConflicts []Role `json:"roleConflicts,omitempty"`

	// PairRole is only meaningful for an item field nested inside an
	// IndexedPairMapping field's Item.Fields. Empty means positional
	// inference applies, exactly as under v1.
	PairRole PairRole `json:"pairRole,omitempty"`
}

// ============================================================
// Scopes
// ============================================================

// ScopeDef declares one named scope a multi-connection plugin's fields
// may belong to — for example, a plugin with a "controller" API and a
// separate "eum" API declares two ScopeDefs, one per connection.
type ScopeDef struct {
	// ID is this scope's unique identifier within the schema document
	// (for example, "controller", "eum"). Referenced by
	// FieldExtensionV2.Scopes.
	ID string `json:"id"`

	// Label is a human-readable name for this scope, for UI and
	// documentation purposes (for example, "Controller API").
	Label string `json:"label"`
}

// validateScopeDefs checks that every declared ScopeDef has a non-empty
// ID, and that no two ScopeDefs share the same ID.
func validateScopeDefs(scopes []ScopeDef) error {
	seen := map[string]bool{}
	for i, s := range scopes {
		if s.ID == "" {
			return fmt.Errorf("scopeDefs[%d]: id is required", i)
		}
		if seen[s.ID] {
			return fmt.Errorf("scopeDefs[%d]: duplicate scope id %q", i, s.ID)
		}
		seen[s.ID] = true
	}
	return nil
}

// validateFieldScopes checks one field's extension Scopes against the
// schema's declared ScopeDefs:
//
//   - every entry must reference a declared scope id;
//   - Scopes must not be used at all if the schema declares no
//     ScopeDefs (there is nothing to scope a field to);
//   - Scopes must not list every currently-declared scope id. Omitting
//     Scopes already means "applies to every scope, including ones
//     declared later" — listing every scope id that exists today is NOT
//     equivalent, because it would not extend to a scope added in a
//     future revision, and allowing both spellings of "all scopes" to
//     coexist invites silent divergence between them. A field meant to
//     be universally shared must omit Scopes; a field meant to be
//     scoped to a bounded subset must list fewer than all of them.
func validateFieldScopes(fieldID string, scopes []string, scopeIDs map[string]bool) error {
	if len(scopes) == 0 {
		return nil
	}
	if len(scopeIDs) == 0 {
		return fmt.Errorf("field %s: scopes is set but the schema declares no scopeDefs", fieldID)
	}

	seen := map[string]bool{}
	for _, ref := range scopes {
		if !scopeIDs[ref] {
			return fmt.Errorf("field %s: scopes references unknown scope id: %s", fieldID, ref)
		}
		if seen[ref] {
			return fmt.Errorf("field %s: scopes contains duplicate scope id: %s", fieldID, ref)
		}
		seen[ref] = true
	}

	if len(scopes) == len(scopeIDs) {
		return fmt.Errorf("field %s: scopes lists every declared scope id; omit scopes entirely instead, which means \"all scopes including any declared later\" — listing them all explicitly does not have that property and is rejected to prevent the two spellings silently diverging", fieldID)
	}

	return nil
}

// EffectiveScopeIDs returns the set of scope ids a field with the given
// extension Scopes belongs to, given the schema's full set of declared
// scope ids. Empty/absent Scopes returns allScopeIDs unchanged (shared
// across every scope, including ones not yet declared at the time this
// function is called with an updated allScopeIDs in a later schema
// revision). A non-empty Scopes returns exactly that set.
func EffectiveScopeIDs(scopes []string, allScopeIDs []string) []string {
	if len(scopes) == 0 {
		return allScopeIDs
	}
	return scopes
}

// FieldsForScope returns every field in schema whose effective scope set
// (per EffectiveScopeIDs, looking up each field's extension in ext)
// includes scopeID. Returns an error if scopeID does not reference a
// declared ScopeDef in ext.ScopeDefs.
//
// This is the function a multi-connection-aware consumer — for example,
// a future per-scope HTTP client builder — uses to get the field set
// relevant to one connection.
func FieldsForScope(schema *Schema, ext *SchemaV2Extensions, scopeID string) ([]ConfigField, error) {
	known := false
	for _, sc := range ext.ScopeDefs {
		if sc.ID == scopeID {
			known = true
			break
		}
	}
	if !known {
		return nil, fmt.Errorf("no declared scope with id %q", scopeID)
	}

	var result []ConfigField
	for i := range schema.Fields {
		f := &schema.Fields[i]
		fx := ext.Fields[f.ID] // zero value if absent: no Scopes, shared
		if len(fx.Scopes) == 0 {
			result = append(result, *f)
			continue
		}
		for _, sc := range fx.Scopes {
			if sc == scopeID {
				result = append(result, *f)
				break
			}
		}
	}
	return result, nil
}

// ============================================================
// Semantic Roles
// ============================================================

// Role is a field's semantic meaning, independent of its Key or ID. Role
// is drawn from a fixed, versioned vocabulary (see the Role* constants)
// so that a consumer — most concretely, a future HTTP client builder, or
// Grafana Assistant reasoning about a field it has never seen before —
// can recognize "this field is the TLS client certificate" without
// hard-coding per-plugin field name lists.
type Role string

const (
	RoleEndpointBaseURL Role = "endpoint.baseUrl"

	RoleTransportTimeoutSeconds Role = "transport.timeoutSeconds"
	RoleTransportTLSSkipVerify  Role = "transport.tlsSkipVerify"

	RoleTLSClientCert Role = "tls.clientCert"
	RoleTLSClientKey  Role = "tls.clientKey"
	RoleTLSCACert     Role = "tls.caCert"
	RoleTLSServerName Role = "tls.serverName"

	RoleAuthDiscriminator   Role = "auth.discriminator"
	RoleAuthBasicEnabled    Role = "auth.basic.enabled"
	RoleAuthBasicUsername   Role = "auth.basic.username"
	RoleAuthBasicPassword   Role = "auth.basic.password"
	RoleAuthOAuth2ClientID  Role = "auth.oauth2.clientId"
	RoleAuthOAuth2Secret    Role = "auth.oauth2.clientSecret"
	RoleAuthJWTSigningKey   Role = "auth.jwt.signingKey"
	RoleAuthAWSSigV4Enabled Role = "auth.awsSigV4.enabled"
	RoleAuthAWSSigV4Access  Role = "auth.awsSigV4.accessKey"
	RoleAuthAWSSigV4Secret  Role = "auth.awsSigV4.secretKey"

	RoleIdentityForwardOAuthToken Role = "identity.forwardOAuthToken"

	RoleHTTPHeaderName  Role = "http.header.name"
	RoleHTTPHeaderValue Role = "http.header.value"
)

// knownRoles is the complete, closed set of valid Role values for this
// version of the package. This set is expected to grow across future
// minor revisions of v2 — additively — but a Role string outside this
// set is rejected by ValidateV2 today rather than silently accepted,
// consistent with id and storage already being closed, validated
// vocabularies rather than free text. See "KNOWN LIMITATIONS (v2)" for
// what is and is not guaranteed by a Role being valid.
var knownRoles = map[Role]bool{
	RoleEndpointBaseURL:           true,
	RoleTransportTimeoutSeconds:   true,
	RoleTransportTLSSkipVerify:    true,
	RoleTLSClientCert:             true,
	RoleTLSClientKey:              true,
	RoleTLSCACert:                 true,
	RoleTLSServerName:             true,
	RoleAuthDiscriminator:         true,
	RoleAuthBasicEnabled:          true,
	RoleAuthBasicUsername:         true,
	RoleAuthBasicPassword:         true,
	RoleAuthOAuth2ClientID:        true,
	RoleAuthOAuth2Secret:          true,
	RoleAuthJWTSigningKey:         true,
	RoleAuthAWSSigV4Enabled:       true,
	RoleAuthAWSSigV4Access:        true,
	RoleAuthAWSSigV4Secret:        true,
	RoleIdentityForwardOAuthToken: true,
	RoleHTTPHeaderName:            true,
	RoleHTTPHeaderValue:           true,
}

// IsValid reports whether r is a member of this version's known role
// vocabulary.
func (r Role) IsValid() bool {
	return knownRoles[r]
}

// validateFieldRole checks one field's extension Role and RoleConflicts:
// that Role, if set, is a known vocabulary value, that every entry in
// RoleConflicts is also a known vocabulary value, and that a field does
// not list its own Role as a conflict. It does NOT check RoleConflicts
// against any other field — that is a property of the whole field set
// within an effective scope; see ValidateRoleConflicts.
func validateFieldRole(fieldID string, fx FieldExtensionV2) error {
	if fx.Role != "" && !fx.Role.IsValid() {
		return fmt.Errorf("field %s: unknown role %q", fieldID, fx.Role)
	}
	for _, rc := range fx.RoleConflicts {
		if !rc.IsValid() {
			return fmt.Errorf("field %s: roleConflicts references unknown role %q", fieldID, rc)
		}
		if fx.Role != "" && rc == fx.Role {
			return fmt.Errorf("field %s: roleConflicts lists its own role %q, which is meaningless", fieldID, fx.Role)
		}
	}
	return nil
}

// ValidateRoleConflicts checks every declared RoleConflicts relationship
// for structural consistency across one effective field set (typically
// the result of FieldsForScope for one scope, or every field in the
// schema for a document with no ScopeDefs at all): that no two fields in
// the same effective set carry the same Role, and that no field's
// RoleConflicts names a Role another field in the same set actually
// carries. It does NOT evaluate whether both fields are simultaneously
// "active" in any real configuration payload — see "KNOWN LIMITATIONS
// (v2)".
func ValidateRoleConflicts(fields []ConfigField, ext *SchemaV2Extensions) error {
	roleOwner := map[Role]string{}
	for i := range fields {
		f := &fields[i]
		fx := ext.Fields[f.ID]
		if fx.Role == "" {
			continue
		}
		if other, exists := roleOwner[fx.Role]; exists {
			return fmt.Errorf("role %q is carried by both field %s and field %s within the same effective scope; a role must be unique within any one effective field set", fx.Role, other, f.ID)
		}
		roleOwner[fx.Role] = f.ID
	}

	for i := range fields {
		f := &fields[i]
		fx := ext.Fields[f.ID]
		for _, rc := range fx.RoleConflicts {
			if owner, exists := roleOwner[rc]; exists {
				return fmt.Errorf("field %s declares a conflict with role %q, which field %s carries within the same effective scope; both cannot be present together by this schema's own declaration", f.ID, rc, owner)
			}
		}
	}

	return nil
}

// ============================================================
// Indexed-Pair Item Role
// ============================================================

// PairRole declares, for an item field used inside an IndexedPairMapping
// field's Item.Fields, whether that item field is the pair's "key"
// (name) half or its "value" half. This replaces ResolveIndexedPairs'
// and ResolveIndexedPairsAsMap's v1 reliance on declaration order (first
// item field = key, second = value) with an explicit declaration.
type PairRole string

const (
	PairRoleKey   PairRole = "key"
	PairRoleValue PairRole = "value"
)

// IsValid reports whether p is PairRoleKey, PairRoleValue, or empty
// (meaning "no explicit role declared; fall back to positional
// inference").
func (p PairRole) IsValid() bool {
	return p == "" || p == PairRoleKey || p == PairRoleValue
}

// validateItemPairRoles checks that, among a set of item field IDs (the
// Item.Fields of a field using IndexedPairMapping), at most one item
// field's extension declares PairRoleKey and at most one declares
// PairRoleValue — catching a schema-authoring contradiction at
// validation time rather than leaving it for a resolver function to
// discover.
func validateItemPairRoles(itemFieldIDs []string, ext *SchemaV2Extensions) error {
	var keyOwner, valueOwner string
	for _, id := range itemFieldIDs {
		fx := ext.Fields[id]
		if fx.PairRole == "" {
			continue
		}
		if !fx.PairRole.IsValid() {
			return fmt.Errorf("item field %s: invalid pairRole %q", id, fx.PairRole)
		}
		switch fx.PairRole {
		case PairRoleKey:
			if keyOwner != "" {
				return fmt.Errorf("item field %s: pairRole %q already claimed by item field %s", id, PairRoleKey, keyOwner)
			}
			keyOwner = id
		case PairRoleValue:
			if valueOwner != "" {
				return fmt.Errorf("item field %s: pairRole %q already claimed by item field %s", id, PairRoleValue, valueOwner)
			}
			valueOwner = id
		}
	}
	return nil
}

// resolvePairRoleKeys returns the item-field Key for the pair's name
// side and the item-field Key for its value side, preferring explicit
// PairRole extensions when present and falling back to v1's positional
// inference (first declared item field = name, second = value) when
// neither item field's extension declares PairRole.
func resolvePairRoleKeys(itemFields []ConfigField, ext *SchemaV2Extensions) (nameKey, valueKey string, err error) {
	if len(itemFields) < 2 {
		return "", "", fmt.Errorf("indexedPair requires an item schema with at least 2 fields")
	}

	var keyField, valueField *ConfigField
	for i := range itemFields {
		f := &itemFields[i]
		fx := ext.Fields[f.ID]
		switch fx.PairRole {
		case PairRoleKey:
			keyField = f
		case PairRoleValue:
			valueField = f
		}
	}

	if keyField != nil && valueField != nil {
		return keyField.Key, valueField.Key, nil
	}
	if keyField != nil || valueField != nil {
		return "", "", fmt.Errorf("indexedPair item schema declares pairRole on only one item field; declare it on both or on neither")
	}

	// Neither item field declares PairRole: fall back to v1's positional
	// inference, exactly as ResolveIndexedPairs/ResolveIndexedPairsAsMap
	// already do.
	return itemFields[0].Key, itemFields[1].Key, nil
}

// ============================================================
// PairRole-Aware Indexed-Pair Resolution
// ============================================================

// ResolveIndexedPairsV2 is identical to v1's ResolveIndexedPairs except
// that it resolves which item field is the pair's name and which is its
// value via resolvePairRoleKeys (PairRole-aware, falling back to
// positional inference) rather than v1's purely positional logic. Use
// this instead of ResolveIndexedPairs when ext declares PairRole for the
// field's item schema; the two functions behave identically when neither
// item field has a PairRole extension.
func ResolveIndexedPairsV2(f *ConfigField, ext *SchemaV2Extensions, jsonData, secureJSONData map[string]any) ([]map[string]any, error) {
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

	if f.Item == nil {
		return nil, fmt.Errorf("field %q: indexedPair requires an item schema", f.ID)
	}
	nameFieldKey, valueFieldKey, err := resolvePairRoleKeys(f.Item.Fields, ext)
	if err != nil {
		return nil, fmt.Errorf("field %q: %w", f.ID, err)
	}

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

		results = append(results, item)
	}

	if results == nil {
		results = []map[string]any{}
	}
	return results, nil
}

// ResolveIndexedPairsAsMapV2 is identical to v1's
// ResolveIndexedPairsAsMap except that — like ResolveIndexedPairsV2 — it
// is PairRole-aware via resolvePairRoleKeys rather than purely
// positional. See ResolveIndexedPairsV2's documentation for when to
// prefer this over the v1 function of the same conceptual purpose.
func ResolveIndexedPairsAsMapV2(f *ConfigField, ext *SchemaV2Extensions, jsonData, secureJSONData map[string]any) (map[string]string, error) {
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

	// nameFieldKey/valueFieldKey are resolved for parity with
	// ResolveIndexedPairsV2 and to validate the item schema's pairRole
	// declarations even though, for the map-shaped output, only the
	// stored *values* (not the item-field Keys) end up in the result.
	if f.Item == nil {
		return nil, fmt.Errorf("field %q: indexedPair requires an item schema", f.ID)
	}
	if _, _, err := resolvePairRoleKeys(f.Item.Fields, ext); err != nil {
		return nil, fmt.Errorf("field %q: %w", f.ID, err)
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

// ============================================================
// v2 Structural Validation Entry Point
// ============================================================

// ValidateV2 performs every v2-specific structural check (scopes, role
// vocabulary, role-conflict structure, pair-role contradiction
// detection) in addition to — not instead of — v1's Schema.Validate. A
// v2 schema document should be validated by calling both: first
// schema.Validate() (unchanged from v1), then ValidateV2(schema, ext).
//
// ValidateV2 is a separate entry point, rather than being folded into
// Schema.Validate, specifically because a v1 schema document — one with
// no SchemaV2Extensions at all — should never be subjected to v2-only
// checks it was never written against. Call this only for a document
// that declares schemaVersion "v2" and therefore has a corresponding
// SchemaV2Extensions value, even if that value's Fields map is empty.
func ValidateV2(schema *Schema, ext *SchemaV2Extensions) error {
	if err := validateScopeDefs(ext.ScopeDefs); err != nil {
		return err
	}

	scopeIDs := map[string]bool{}
	for _, sc := range ext.ScopeDefs {
		scopeIDs[sc.ID] = true
	}

	var visit func(fields []ConfigField, itemFieldIDsOfIndexedPairParent []string) error
	visit = func(fields []ConfigField, itemFieldIDsOfIndexedPairParent []string) error {
		for i := range fields {
			f := &fields[i]
			fx := ext.Fields[f.ID]

			if err := validateFieldScopes(f.ID, fx.Scopes, scopeIDs); err != nil {
				return err
			}
			if err := validateFieldRole(f.ID, fx); err != nil {
				return err
			}

			isDeclaredIndexedPairItem := contains(itemFieldIDsOfIndexedPairParent, f.ID)
			if !isDeclaredIndexedPairItem && fx.PairRole != "" {
				return fmt.Errorf("field %s: pairRole is only meaningful on an item field of an indexedPair-mapped field", f.ID)
			}

			if f.Item != nil {
				isIndexedPair := f.Storage != nil && f.Storage.Type == IndexedPairMapping
				var childItemIDs []string
				if isIndexedPair {
					for j := range f.Item.Fields {
						childItemIDs = append(childItemIDs, f.Item.Fields[j].ID)
					}
					if err := validateItemPairRoles(childItemIDs, ext); err != nil {
						return err
					}
				}
				if err := visit(f.Item.Fields, childItemIDs); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := visit(schema.Fields, nil); err != nil {
		return err
	}

	if len(ext.ScopeDefs) == 0 {
		if err := ValidateRoleConflicts(schema.Fields, ext); err != nil {
			return err
		}
	} else {
		ids := make([]string, 0, len(ext.ScopeDefs))
		for _, sc := range ext.ScopeDefs {
			ids = append(ids, sc.ID)
		}
		sort.Strings(ids)
		for _, id := range ids {
			fields, err := FieldsForScope(schema, ext, id)
			if err != nil {
				return err
			}
			if err := ValidateRoleConflicts(fields, ext); err != nil {
				return fmt.Errorf("scope %q: %w", id, err)
			}
		}
	}

	return nil
}

func contains(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

// ============================================================
// KNOWN LIMITATIONS (v2)
// ============================================================
//
// This section is additional to, not a replacement for, v1's KNOWN
// LIMITATIONS (SCHEMA-V1.go). Every v1 limitation that v2 does not close
// remains exactly as documented there.
//
//  1. This file represents Scopes/Role/RoleConflicts/PairRole as a side
//     table (SchemaV2Extensions, FieldExtensionV2) keyed by
//     ConfigField.ID, rather than as fields directly on ConfigField, as
//     a deliberate constraint of this deliverable (see the
//     IMPLEMENTATION NOTE at the top of this file). A schema author
//     using this exact reference implementation must keep a
//     SchemaV2Extensions document in sync with its corresponding Schema
//     document by ID, by hand or by tooling — nothing in this file
//     enforces that every ID in a Schema has a corresponding (even if
//     empty) entry, or that every ID in a SchemaV2Extensions.Fields map
//     actually exists in the Schema it is paired with beyond what
//     ValidateV2 itself checks at the point each ID is visited. The
//     proposed upstream shape (Scopes/Role/RoleConflicts/PairRole as
//     fields directly on ConfigField) does not have this two-document
//     synchronization burden at all, because there would be only one
//     document.
//
//  2. ToPluginSchemaSettings does not produce a multi-connection-aware
//     App Platform artifact. A v2 schema plus its extensions can fully
//     describe a multi-connection plugin, but the SDK conversion still
//     emits exactly one Settings{Spec, SecureValues} pair per document.
//
//  3. RoleConflicts is structurally validated but not evaluated against
//     any real configuration payload. ValidateRoleConflicts catches a
//     schema that could never be satisfied; it does not catch a real
//     configuration where two conflicting-by-declaration fields are both
//     actually set to a truthy/active value at the same time.
//
//  4. A Role being valid (IsValid) only means it is a member of this
//     version's known vocabulary. It does not mean the field's ValueType
//     or Target is appropriate for that role — nothing cross-checks
//     Role: RoleTLSClientCert against the field actually being a string
//     targeting SecureJSONTarget.
//
//  5. The known-role vocabulary (knownRoles) is fixed by this package
//     version and is not extensible by a schema author.
//
//  6. PairRole resolution requires that if either item field declares
//     PairRole, the other must too; a schema declaring it on only one of
//     the two is rejected as ambiguous rather than silently falling back
//     to positional inference for the other half.
//
//  7. v1's ResolveIndexedPairs and ResolveIndexedPairsAsMap (SCHEMA-V1.go)
//     are completely unaffected by this file and continue to use purely
//     positional inference, with no awareness of PairRole. A caller who
//     wants PairRole-aware resolution must call ResolveIndexedPairsV2 or
//     ResolveIndexedPairsAsMapV2 instead.
