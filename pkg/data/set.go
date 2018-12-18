package data

import (
	"sync"
)

// Set is a syncronized collection
type Set struct {
	props map[string]interface{}
	mux   sync.RWMutex
}

// NewSet creates Set
func NewSet() *Set {
	return &Set{
		props: make(map[string]interface{}),
	}
}

// Size returns length of the map
func (p *Set) Size() int {
	p.mux.RLock()
	defer p.mux.RUnlock()
	return len(p.props)
}

// Get returns property by name
func (p *Set) Get(key string) interface{} {
	p.mux.RLock()
	defer p.mux.RUnlock()

	val, ok := p.props[key]
	if !ok {
		panic("invalid set index: " + key)
	}
	return val
}

// Keys returns all keys of the set
func (p *Set) Keys() []string {
	p.mux.RLock()
	defer p.mux.RUnlock()

	keys := make([]string, len(p.props))
	for key := range p.props {
		keys = append(keys, key)
	}

	return keys
}

// Values returns all values of the set
func (p *Set) Values() []interface{} {
	p.mux.RLock()
	defer p.mux.RUnlock()

	values := make([]interface{}, len(p.props))
	for _, value := range p.props {
		values = append(values, value)
	}

	return values
}

// Filter returns matching subset of the set
func (p *Set) Filter(f func(key string, value interface{}) bool) *Set {
	p.mux.RLock()
	defer p.mux.RUnlock()

	subset := NewSet()
	for key, value := range p.props {
		if f(key, value) {
			subset.Add(key, value)
		}
	}

	return subset
}

// All returns all properties of the set
func (p *Set) All() map[string]interface{} {
	p.mux.RLock()
	defer p.mux.RUnlock()

	propMap := make(map[string]interface{})
	for key, val := range p.props {
		propMap[key] = val
	}

	return propMap
}

// Contains checks if match is contained in the list
func (p *Set) Contains(s string) bool {
	p.mux.RLock()
	defer p.mux.RUnlock()

	_, ok := p.props[s]
	return ok
}

// Add adds an entry to the set
func (p *Set) Add(key string, val interface{}) bool {
	if p.Contains(key) {
		return false
	}

	p.mux.Lock()
	defer p.mux.Unlock()
	p.props[key] = val
	return true
}

// Remove removes an entry from the set
func (p *Set) Remove(key string) bool {
	if !p.Contains(key) {
		return false
	}

	p.mux.Lock()
	defer p.mux.Unlock()
	delete(p.props, key)
	return true
}
