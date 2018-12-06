package volkszaehler

import (
	"fmt"
	"log"
	"time"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"
)

func NewFromOutputConfig(c config.Output) *Publisher {
	api := NewAPI(c.URL, 1*time.Second, false)
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

func (vz *Publisher) Publish(d data.Data) {
	log.Printf("volkszaehler: send (%s=%f)", d.Name, d.Value)

	ts := int64(time.Now().UnixNano() / 1e3)
	val := fmt.Sprintf("%.3f", d.Value)
	payload := fmt.Sprintf(`[
		[%d,%s]
	]`, ts, val)

	id := d.ID
	if id == "" {
		id = d.Name
	}
	url := fmt.Sprintf("/data/%s.json", id)

	if _, err := vz.Api.Post(url, payload); err != nil {
		log.Printf("volkszaehler: send failed (%s)", err)
	}
}
