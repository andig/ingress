package homie

import (
	"strings"
	"sync"
)

// PropertySet provides common operations on list of properties
type PropertySet struct {
	props map[string]bool
	mux   sync.RWMutex
}

// NewPropertySet creates PropertySet
func NewPropertySet(props ...string) *PropertySet {
	propMap := make(map[string]bool)
	for _, prop := range props {
		propMap[prop] = true
	}
	return &PropertySet{
		props: propMap,
	}
}

// Get returns property by name
func (p *PropertySet) Get(s string) interface{} {
	return p.props[s]
}

// All returns all properties of the set
func (p *PropertySet) All() []string {
	p.mux.RLock()
	defer p.mux.RUnlock()

	keys := make([]string, 0, len(p.props))
	for prop := range p.props {
		keys = append(keys, prop)
	}

	return keys
}

// Match returns all properties starting with match
func (p *PropertySet) Match(s string) []string {
	p.mux.RLock()
	defer p.mux.RUnlock()

	properties := make([]string, 0)
	for property := range p.props {
		if strings.HasPrefix(property, s) {
			properties = append(properties, property)
		}
	}

	return properties
}

// Contains checks if match is contained in the list
func (p *PropertySet) Contains(s string) bool {
	_, ok := p.props[s]
	return ok
}

// Add adds an entry to the set
func (p *PropertySet) Add(s string) bool {
	p.mux.Lock()
	defer p.mux.Unlock()

	if p.Contains(s) {
		return false
	}

	p.props[s] = true
	return true
}

// Remove removes an entry from the set
func (p *PropertySet) Remove(s string) bool {
	p.mux.Lock()
	defer p.mux.Unlock()

	if !p.Contains(s) {
		return false
	}

	delete(p.props, s)
	return true
}
