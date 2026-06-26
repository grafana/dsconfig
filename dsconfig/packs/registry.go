// Package packs provides the built-in field pack definitions.
// Each pack registers itself into the dsconfig package registry via init().
// Consumers import this package for its side effects:
//
//	import _ "github.com/grafana/dsconfig/dsconfig/packs"
//
// After import, all built-in packs are available to ResolveBaseFields().
//
// Each pack's fields are defined in a companion JSON file (e.g.
// plugin_sdk_settings.json) and embedded at compile time. The JSON contains a
// []ConfigField array using the same format as the dsconfig.json fields array.
// Field IDs must be namespaced with the pack ID prefix, e.g.
// "plugin_sdk_settings.url" for a field in plugin_sdk_settings.json.
package packs

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/dsconfig/dsconfig"
)

// packJSON is the on-disk shape of a field pack JSON file.
// It mirrors the top-level shape of dsconfig.json so both formats
// are immediately recognisable to contributors.
type packJSON struct {
	// ID is the canonical pack identifier. When present it must match the
	// FieldPackID constant used to register the pack — a belt-and-suspenders
	// check that catches copy-paste errors at startup.
	ID          dsconfig.FieldPackID   `json:"id,omitempty"`
	Description string                 `json:"description,omitempty"`
	Fields      []dsconfig.ConfigField `json:"fields"`
}

// mustLoadPack parses a { "id": "...", "description": "...", "fields": [...] }
// pack JSON file and registers the resulting FieldPack. It panics on malformed
// JSON or an ID mismatch to surface authoring errors at startup.
func mustLoadPack(id dsconfig.FieldPackID, data []byte) {
	var p packJSON
	if err := json.Unmarshal(data, &p); err != nil {
		panic(fmt.Sprintf("dsconfig/packs: failed to parse %s.json: %v", id, err))
	}
	if p.ID != "" && p.ID != id {
		panic(fmt.Sprintf("dsconfig/packs: %s.json declares id %q but was registered as %q", id, p.ID, id))
	}
	dsconfig.RegisterFieldPack(&dsconfig.FieldPack{
		ID:     id,
		Fields: p.Fields,
	})
}
