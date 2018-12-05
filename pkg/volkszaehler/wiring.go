package volkszaehler

import (
	"log"
	"time"

	"github.com/andig/ingress/pkg/config"
)

func NewFromOutputConfig(c config.Output) *Publisher {
	api := NewAPI(c.Url, 1*time.Second, false)
	vz := &Publisher{
		Api: api,
	}
	return vz
}

type Publisher struct {
	*Api
}

func (vz *Publisher) Discover() {
	vz.discoverEntities(vz.GetPublicEntities())
}

func (vz *Publisher) discoverEntities(entities []Entity) {
	for _, e := range entities {
		log.Printf("%s %s: %s", e.UUID, e.Type, e.Title)
	}
	for _, e := range entities {
		if e.Type == TypeGroup {
			children := vz.GetEntity(e.UUID).Children
			vz.discoverEntities(children)
		}
	}
}
