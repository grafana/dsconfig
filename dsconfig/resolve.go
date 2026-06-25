package dsconfig

import "fmt"

// ResolveBaseFields merges declared base field packs into a copy of s and
// returns the resolved Schema. It does not mutate the receiver.
//
// Resolution order:
//  1. Validate BaseFieldRef entries (unknown pack, duplicate pack, exclude/patch conflicts).
//  2. For each BaseFieldRef, copy pack fields, drop excluded IDs, apply patches.
//  3. Build merged field list: pack fields first (in declaration order), then s.Fields.
//  4. If a pack field ID collides with a plugin-declared field, the plugin field wins
//     and the pack field is silently dropped.
//  5. Return a new *Schema with Fields = merged list and BaseFields = nil.
//
// Call ResolveBaseFields before Validate() and ToPluginSchemaSettings().
func (s *Schema) ResolveBaseFields() (*Schema, error) {
	if len(s.BaseFields) == 0 {
		// Nothing to resolve; return a shallow copy with the same Fields.
		out := *s
		return &out, nil
	}

	// Check for duplicate pack declarations.
	seen := map[FieldPackID]bool{}
	for i, ref := range s.BaseFields {
		if seen[ref.From] {
			return nil, fmt.Errorf("baseFields[%d]: duplicate pack %q", i, ref.From)
		}
		seen[ref.From] = true
	}

	// Build the merged field list from packs, then deduplicate against s.Fields.
	pluginFieldIDs := map[string]bool{}
	for _, f := range s.Fields {
		pluginFieldIDs[f.ID] = true
	}

	var packFields []ConfigField

	for i, ref := range s.BaseFields {
		pack, ok := lookupFieldPack(ref.From)
		if !ok {
			return nil, fmt.Errorf("baseFields[%d]: unknown pack %q", i, ref.From)
		}

		// Build an index of pack field IDs for validation.
		packFieldIDs := map[string]bool{}
		for _, f := range pack.Fields {
			packFieldIDs[f.ID] = true
		}

		// Validate exclude IDs exist in the pack.
		for _, id := range ref.Exclude {
			if !packFieldIDs[id] {
				return nil, fmt.Errorf("baseFields[%d]: exclude references unknown field id %q in pack %q", i, id, ref.From)
			}
		}

		// Validate patch keys exist in the pack.
		for id := range ref.Patch {
			if !packFieldIDs[id] {
				return nil, fmt.Errorf("baseFields[%d]: patch references unknown field id %q in pack %q", i, id, ref.From)
			}
		}

		// Validate no ID is both excluded and patched.
		excludeSet := map[string]bool{}
		for _, id := range ref.Exclude {
			excludeSet[id] = true
		}
		for id := range ref.Patch {
			if excludeSet[id] {
				return nil, fmt.Errorf("baseFields[%d]: field %q is both excluded and patched in pack %q", i, id, ref.From)
			}
		}

		// Copy pack fields, applying excludes, patches, and collision checks.
		for _, f := range pack.Fields {
			if excludeSet[f.ID] {
				continue
			}
			if pluginFieldIDs[f.ID] {
				// Plugin-declared field wins; skip the pack field silently.
				continue
			}
			if patch, ok := ref.Patch[f.ID]; ok {
				f = applyFieldPatch(f, patch)
			}
			packFields = append(packFields, f)
		}
	}

	// Merge: pack fields first, then plugin fields.
	merged := make([]ConfigField, 0, len(packFields)+len(s.Fields))
	merged = append(merged, packFields...)
	merged = append(merged, s.Fields...)

	out := *s
	out.Fields = merged
	out.BaseFields = nil
	return &out, nil
}

// applyFieldPatch returns a copy of f with non-zero patch properties applied.
// Structural properties (ID, Key, ValueType, Target, Role) are never modified.
func applyFieldPatch(f ConfigField, p FieldPatch) ConfigField {
	if p.Label != "" {
		f.Label = p.Label
	}
	if p.Description != "" {
		f.Description = p.Description
	}
	if p.Placeholder != "" {
		if f.UI == nil {
			ui := FieldUI{}
			f.UI = &ui
		}
		f.UI.Placeholder = p.Placeholder
	}
	if p.DefaultValue != nil {
		f.DefaultValue = p.DefaultValue
	}
	if p.Required != nil {
		f.Required = *p.Required
	}
	// Hidden is not currently a ConfigField property; reserved for future use.
	return f
}

// ParseAndResolveSchemaJSON is a convenience function that parses dsconfig JSON,
// resolves baseFields, and validates the resulting schema.
// Use this instead of ParseSchemaJSON when the schema may declare baseFields.
// Requires the packs sub-package to be imported for its side effects:
//
//	import _ "github.com/grafana/dsconfig/dsconfig/packs"
func ParseAndResolveSchemaJSON(data []byte) (*Schema, error) {
	s, err := ParseSchemaJSON(data)
	if err != nil {
		return nil, err
	}
	resolved, err := s.ResolveBaseFields()
	if err != nil {
		return nil, fmt.Errorf("resolve baseFields: %w", err)
	}
	if err := resolved.Validate(); err != nil {
		return nil, err
	}
	return resolved, nil
}
