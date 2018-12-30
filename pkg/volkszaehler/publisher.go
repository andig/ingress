package volkszaehler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"time"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	. "github.com/andig/ingress/pkg/log"
)

// Publisher is the volkszaehler data taerget
type Publisher struct {
	*Api
	name string
}

// NewFromTargetConfig creates volkszaehler data target
func NewFromTargetConfig(c config.Target) *Publisher {
	api := NewAPI(c.URL, 1*time.Second, false)
	vz := &Publisher{
		Api:  api,
		name: c.Name,
	}
	return vz
}

func (p *Publisher) discoverEntities(entities []Entity) {
	for _, e := range entities {
		Log(
			TGT, p.name,
		).Printf("s %s: %s", e.UUID, e.Type, e.Title)
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
	Log(
		TGT, p.name,
		EV, d.GetName(),
		VAL, d.ValStr(),
	).Debugf("send")

	// format url and payload
	url := fmt.Sprintf("/data/%s.json", d.GetName())
	payload := fmt.Sprintf(`[
		[%d,%s]
	]`, d.GetTimestamp(), d.ValStr())

	resp, err := p.Api.Post(url, payload)
	if err != nil {
		Log(
			TGT, p.name,
		).Errorf("send failed (%s)", err)
		return
	}
	defer resp.Body.Close() // close body after checking for error

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			Log(
				TGT, p.name,
			).Errorf("reading response failed (%s)", err)
			return
		}

		var res ErrorResponse
		if err := json.Unmarshal(body, &res); err != nil {
			Log(
				TGT, p.name,
			).Errorf("decoding response failed (%s)", err)
			return
		}

		Log(
			TGT, p.name,
		).Errorf("send failed (%s)", res.Exception.Message)
	}
}
