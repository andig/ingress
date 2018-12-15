package homie

import (
	"strings"
	"sync"
)

type PropertySet struct {
	props []string
	mux   sync.RWMutex
}

func NewPropertySet(props ...string) *PropertySet {
	return &PropertySet{
		props: props,
	}
}

func (p *PropertySet) Get() []string {
	p.mux.RLock()
	defer p.mux.RUnlock()

	return p.props
}

func (p *PropertySet) Match(s string) []string {
	p.mux.RLock()
	defer p.mux.RUnlock()

	properties := make([]string, 0)
	for _, property := range p.props {
		if strings.Index(property, s) == 0 {
			properties = append(properties, property)
		}
	}

	return properties
}

func (p *PropertySet) Contains(s string) bool {
	return p.indexOf(s) >= 0
}

func (p *PropertySet) indexOf(s string) int {
	for i, prop := range p.props {
		if prop == s {
			return i
		}
	}

	return -1
}

func (p *PropertySet) Add(s string) bool {
	p.mux.Lock()
	defer p.mux.Unlock()

	if i := p.indexOf(s); i < 0 {
		p.props = append(p.props, s)
		return true
	}

	return false
}

func (p *PropertySet) Remove(s string) bool {
	p.mux.Lock()
	defer p.mux.Unlock()

	if i := p.indexOf(s); i >= 0 {
		// remove element i by moving last element to its position
		p.props[i] = p.props[len(p.props)-1]
		p.props = p.props[:len(p.props)-1]
		return true
	}

	return false
}
