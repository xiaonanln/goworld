package common

// EntityIDSet is the data structure for a set of entity IDs
type EntityIDSet map[EntityID]struct{}

// Add adds an entity ID to EntityIDSet
func (es EntityIDSet) Add(id EntityID) {
	es[id] = struct{}{}
}

// Del removes an entity ID from EntityIDSet
func (es EntityIDSet) Del(id EntityID) {
	delete(es, id)
}

// Contains checks if entity ID is in EntityIDSet
func (es EntityIDSet) Contains(id EntityID) bool {
	_, ok := es[id]
	return ok
}

// ToList convert EntityIDSet to a slice of entity IDs
func (es EntityIDSet) ToList() []EntityID {
	list := make([]EntityID, 0, len(es))
	for eid := range es {
		list = append(list, eid)
	}
	return list
}

func (es EntityIDSet) ForEach(cb func(eid EntityID) bool) {
	for eid := range es {
		if !cb(eid) {
			break
		}
	}
}
