package actions

import (
	"regexp"
	"sync"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"
	"github.com/andig/ingress/pkg/registry"
)

func init() {
	registry.RegisterAction("dropfilter", NewDropFilterFromActionConfig)
	registry.RegisterAction("passfilter", NewPassFilterFromActionConfig)
}

type filterConfig struct {
	config.Action `yaml:",squash"`
	Patterns      []string `yaml:"patterns,omitempty"`
	Matches       []string `yaml:"matches,omitempty"`
}

type patternStore struct {
	mux     sync.Mutex
	regexes map[string]*regexp.Regexp
}

func (p *patternStore) getRegEx(pattern string) *regexp.Regexp {
	p.mux.Lock()
	defer p.mux.Unlock()

	if re, ok := p.regexes[pattern]; ok {
		return re
	}

	re := regexp.MustCompile(pattern)
	p.regexes[pattern] = re
	return re
}

type DropFilterAction struct {
	filterConfig
	patternStore
}

func NewDropFilterFromActionConfig(g config.Generic) (a api.Action, err error) {
	var conf filterConfig
	err = config.Decode(g, &conf)
	if err != nil {
		return nil, err
	}

	a = &DropFilterAction{
		filterConfig: conf,
		patternStore: patternStore{regexes: make(map[string]*regexp.Regexp)},
	}

	return a, nil
}

// Process implements the Action interface.
func (a *DropFilterAction) Process(d api.Data) api.Data {
	dataName := d.Name()

	for _, name := range a.Matches {
		if name == dataName {
			log.Context(
				log.EV, d.Name(),
				log.ACT, a.Name,
			).Debugf("dropped")

			return nil
		}
	}

	for _, pattern := range a.Patterns {
		re := a.getRegEx(pattern)
		if re.MatchString(dataName) {
			log.Context(
				log.EV, d.Name(),
				log.ACT, a.Name,
			).Debugf("dropped")

			return nil
		}
	}

	return d
}

type PassFilterAction struct {
	filterConfig
	patternStore
}

func NewPassFilterFromActionConfig(g config.Generic) (a api.Action, err error) {
	var conf filterConfig
	err = config.Decode(g, &conf)
	if err != nil {
		return nil, err
	}

	a = &PassFilterAction{
		filterConfig: conf,
		patternStore: patternStore{regexes: make(map[string]*regexp.Regexp)},
	}

	return a, nil
}

// Process implements the Action interface.
func (a *PassFilterAction) Process(d api.Data) api.Data {
	dataName := d.Name()

	for _, name := range a.Matches {
		if name == dataName {
			return d
		}
	}

	for _, pattern := range a.Patterns {
		re := a.getRegEx(pattern)
		if re.MatchString(dataName) {
			return d
		}
	}

	log.Context(
		log.EV, d.Name(),
		log.ACT, a.Name,
	).Debugf("dropped")

	return nil
}
