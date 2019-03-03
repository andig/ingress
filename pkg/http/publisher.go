package http

import (
	"bytes"
	"errors"
	"io/ioutil"
	transport "net/http"
	"net/url"
	"strings"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"
	"github.com/andig/ingress/pkg/registry"
)

func init() {
	registry.RegisterTarget("http", NewFromTargetConfig)
}

// Publisher is the HTTP data target
type Publisher struct {
	name    string
	url     string
	method  string
	headers map[string]string
	payload string
	client  *transport.Client
}

type httpConfig = struct {
	config.Target `yaml:",squash"`
	URL           string
	Method        string
	Headers       map[string]string `yaml:"headers,omitempty"`
	Payload       string
}

// NewFromTargetConfig creates HTTP data target
func NewFromTargetConfig(g config.Generic) (p api.Target, err error) {
	var c httpConfig
	err = config.Decode(g, &c)
	if err != nil {
		return nil, err
	}

	if _, err = url.ParseRequestURI(c.URL); err != nil {
		return p, err
	}
	method := strings.ToUpper(c.Method)
	if method == "" {
		method = "GET"
	}
	if method != "GET" && method != "POST" {
		return p, errors.New("invalid method " + c.Method)
	}
	if method == "POST" && c.Payload == "" {
		return p, errors.New("missing payload configuration for POST method")
	}
	if method == "GET" && c.Payload != "" {
		return p, errors.New("invalid payload configuration for GET method")
	}

	p = &Publisher{
		name:    c.Name,
		url:     c.URL,
		method:  method,
		payload: c.Payload,
		headers: c.Headers,
		client:  &transport.Client{},
	}
	return p, nil
}

// Discover implements api.Source
func (p *Publisher) Discover() {
}

// Publish implements api.Source
func (p *Publisher) Publish(d api.Data) {
	url := d.MatchPattern(p.url)
	log.Context(log.TGT, p.name).Debugf("%s %s", p.method, url)

	var resp *transport.Response
	var req *transport.Request
	var err error
	var payload string

	if p.method == "GET" {
		req, err = transport.NewRequest(p.method, url, nil)
	} else { // POST
		payload = d.MatchPattern(p.payload)
		req, err = transport.NewRequest(p.method, url, bytes.NewBuffer([]byte(payload)))
	}

	if err != nil {
		log.Context(log.TGT, p.name).Errorf("create request failed %s", err)
		return
	}

	// headers
	for key, value := range p.headers {
		req.Header.Set(key, value)
	}

	// requestDump, err := httputil.DumpRequest(req, true)
	// if err != nil {
	// 	log.Context(log.TGT, p.name).Error(err)
	// }

	// execute request
	resp, err = p.client.Do(req)
	if err != nil {
		log.Context(log.TGT, p.name).Errorf("send failed %s", err)
		return
	}
	defer resp.Body.Close() // close body after checking for error

	if resp.StatusCode != 200 {
		log.Context(log.TGT, p.name).Errorf("%s %s %d", p.method, url, resp.StatusCode)

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Context(log.
				TGT, p.name,
			).Errorf("reading response failed (%s)", err)
			return
		}

		log.Context(log.
			TGT, p.name,
		).Errorf("send failed (%s)", string(body))
	}
}
