package volkszaehler

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	vz "github.com/andig/gravo/volkszaehler"
	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"
	"github.com/andig/ingress/pkg/registry"
)

func init() {
	registry.RegisterTarget("volkszaehler", NewFromTargetConfig)
}

// Publisher is the volkszaehler data taerget
type Publisher struct {
	vz.Client
	name string
}

type volkszaehlerConfig = struct {
	config.Target `yaml:",squash"`
	URL           string
	Timeout       time.Duration
}

// NewFromTargetConfig creates volkszaehler data target
func NewFromTargetConfig(g config.Generic) (p api.Target, err error) {
	var c volkszaehlerConfig
	err = config.Decode(g, &c)
	if err != nil {
		return nil, err
	}

	if _, err = url.ParseRequestURI(c.URL); err != nil {
		return p, err
	}

	httpClient := http.Client{}
	if c.Timeout != 0 {
		httpClient.Timeout = c.Timeout
	}

	client := vz.NewClient(c.URL, &httpClient, false)
	p = &Publisher{
		Client: client,
		name:   c.Name,
	}
	return p, nil
}

// Publish implements api.Source
func (p *Publisher) Publish(d api.Data) {
	log.Context(
		log.TGT, p.name,
		log.EV, d.Name(),
		log.VAL, d.ValStr(),
	).Debugf("send")

	// format url and payload
	url := fmt.Sprintf("/data/%s.json", d.Name())
	payload := fmt.Sprintf(`[
		[%d,%s]
	]`, d.Timestamp().UnixNano()/1e6, d.ValStr())

	resp, err := p.Client.Post(url, payload)
	if err != nil {
		log.Context(log.TGT, p.name).Errorf("send failed (%s)", err)
		return
	}
	defer resp.Close() // close body after checking for error

	body, err := ioutil.ReadAll(resp)
	if err != nil {
		log.Context(log.TGT, p.name).Errorf("reading response failed (%s)", err)
		return
	}

	var res vz.PostDataResponse
	if err := json.Unmarshal(body, &res); err != nil {
		log.Context(log.TGT, p.name).Errorf("decoding response failed (%s)", err)
		return
	} else if res.Rows != 1 {
		log.Context(log.TGT, p.name).Errorf("unexpected response (%s)", body)
	}
}
