package dsconfig

import "sync"

var (
	packMu          sync.RWMutex
	registeredPacks = map[FieldPackID]*FieldPack{}
)

// RegisterFieldPack adds a FieldPack to the built-in registry.
// It is called by each pack's init() function via the packs sub-package.
// Registering a pack with a duplicate ID panics to surface misconfiguration early.
func RegisterFieldPack(p *FieldPack) {
	packMu.Lock()
	defer packMu.Unlock()
	if _, exists := registeredPacks[p.ID]; exists {
		panic("dsconfig: duplicate field pack registration for id " + string(p.ID))
	}
	registeredPacks[p.ID] = p
}

// lookupFieldPack returns the FieldPack for the given ID, or (nil, false) if unknown.
func lookupFieldPack(id FieldPackID) (*FieldPack, bool) {
	packMu.RLock()
	defer packMu.RUnlock()
	p, ok := registeredPacks[id]
	return p, ok
}

// RegisteredPackIDs returns all registered pack IDs.
// Exported for use by the schema.json generator.
func RegisteredPackIDs() []FieldPackID {
	packMu.RLock()
	defer packMu.RUnlock()
	ids := make([]FieldPackID, 0, len(registeredPacks))
	for id := range registeredPacks {
		ids = append(ids, id)
	}
	return ids
}
