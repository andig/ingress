package volkszaehler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"time"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"
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

func (vz *Publisher) discoverEntities(entities []Entity) {
	for _, e := range entities {
		log.Printf(vz.name+": %s %s: %s", e.UUID, e.Type, e.Title)
	}
	for _, e := range entities {
		if e.Type == TypeGroup {
			children := vz.GetEntity(e.UUID).Children
			vz.discoverEntities(children)
		}
	}
}

// Publish implements api.Source
func (vz *Publisher) Publish(d data.Data) {
	log.Printf(vz.name+": send (%s=%f)", d.Name, d.Value)

	// format payload
	payload := fmt.Sprintf(`[
		[%d,%s]
	]`, d.Timestamp, d.ValStr())

	id := d.ID
	if id == "" {
		id = d.Name
	}
	url := fmt.Sprintf("/data/%s.json", id)

	resp, err := vz.Api.Post(url, payload)
	if err != nil {
		log.Printf(vz.name+": send failed (%s)", err)
		return
	}

	if resp.StatusCode != 200 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf(vz.name+": reading response failed (%s)", err)
			return
		}

		var res ErrorResponse
		if err := json.Unmarshal(body, &res); err != nil {
			log.Printf(vz.name+": decoding response failed (%s)", err)
			return
		}

		log.Printf(vz.name+": send failed (%s)", res.Exception.Message)
	}
}
