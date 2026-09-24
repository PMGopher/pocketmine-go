package entity

// AttributeMap is a port of pocketmine\entity\AttributeMap. Insertion order is kept (PHP arrays
// are ordered, and it's the order attributes are sent to clients in).
type AttributeMap struct {
	attributes map[string]*Attribute
	order      []string
}

func NewAttributeMap() *AttributeMap {
	return &AttributeMap{attributes: map[string]*Attribute{}}
}

func (m *AttributeMap) Add(attribute *Attribute) {
	if _, exists := m.attributes[attribute.GetID()]; !exists {
		m.order = append(m.order, attribute.GetID())
	}
	m.attributes[attribute.GetID()] = attribute
}

// Get returns the attribute with the given ID, or nil.
func (m *AttributeMap) Get(id string) *Attribute { return m.attributes[id] }

// GetAll returns every attribute, in insertion order.
func (m *AttributeMap) GetAll() []*Attribute {
	result := make([]*Attribute, 0, len(m.order))
	for _, id := range m.order {
		result = append(result, m.attributes[id])
	}
	return result
}

// NeedSend returns the syncable attributes whose value changed since they were last synced.
func (m *AttributeMap) NeedSend() []*Attribute {
	var result []*Attribute
	for _, a := range m.GetAll() {
		if a.IsSyncable() && a.IsDesynchronized() {
			result = append(result, a)
		}
	}
	return result
}
