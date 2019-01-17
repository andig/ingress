package volkszaehler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"time"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"
)

// Publisher is the volkszaehler data taerget
type Publisher struct {
	*Api
	name string
}

// NewFromTargetConfig creates volkszaehler data target
func NewFromTargetConfig(c config.Target) (p *Publisher, err error) {
	if _, err = url.ParseRequestURI(c.URL); err != nil {
		return p, err
	}

	if c.Timeout == 0 {
		c.Timeout = 1 * time.Second
	}

	api := NewAPI(c.URL, c.Timeout, false)
	p = &Publisher{
		Api:  api,
		name: c.Name,
	}
	return p, nil
}

func (p *Publisher) discoverEntities(entities []Entity) {
	for _, e := range entities {
		log.Context(log.TGT, p.name).Printf("s %s: %s", e.UUID, e.Type, e.Title)
	}
	for _, e := range entities {
		if e.Type == TypeGroup {
			children := p.GetEntity(e.UUID).Children
			p.discoverEntities(children)
		}
	}
}

// Publish implements api.Source
func (p *Publisher) Publish(d api.Data) {
	log.Context(
		log.TGT, p.name,
		log.EV, d.GetName(),
		log.VAL, d.ValStr(),
	).Debugf("send")

	// format url and payload
	url := fmt.Sprintf("/data/%s.json", d.GetName())
	payload := fmt.Sprintf(`[
		[%d,%s]
	]`, d.GetTimestamp(), d.ValStr())

	resp, err := p.Api.Post(url, payload)
	if err != nil {
		log.Context(log.TGT, p.name).Errorf("send failed (%s)", err)
		return
	}
	defer resp.Body.Close() // close body after checking for error

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Context(log.TGT, p.name).Errorf("reading response failed (%s)", err)
			return
		}

		var res ErrorResponse
		if err := json.Unmarshal(body, &res); err != nil {
			log.Context(log.TGT, p.name).Errorf("decoding response failed (%s)", err)
			return
		}

		log.Context(log.TGT, p.name).Errorf("send failed (%s)", res.Exception.Message)
	}
}
