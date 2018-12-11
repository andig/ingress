package http

import (
	"bytes"
	"log"
	transport "net/http"
	"strings"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"
)

type Publisher struct {
	name    string
	url     string
	method  string
	headers map[string]string
	payload string
	client  *transport.Client
}

func NewFromOutputConfig(c config.Output) *Publisher {
	method := strings.ToUpper(c.Method)
	if method != "GET" && method != "POST" {
		panic(c.Name + ": invalid method " + c.Method)
	}

	h := &Publisher{
		name:    c.Name,
		url:     c.URL,
		method:  method,
		payload: c.Payload,
		headers: c.Headers,
		client:  &transport.Client{},
	}
	return h
}

func (h *Publisher) Discover() {
}

func (h *Publisher) Publish(d data.Data) {
	url := d.MatchPattern(h.url)
	log.Printf(h.name+": send %s %s (%s=%f)", h.method, url, d.Name, d.Value)

	var resp *transport.Response
	var req *transport.Request
	var err error
	var payload string

	if h.method == "GET" {
		req, err = transport.NewRequest(h.method, url, nil)
	} else { // POST
		payload = d.MatchPattern(h.payload)
		req, err = transport.NewRequest(h.method, url, bytes.NewBuffer([]byte(payload)))
	}

	if err != nil {
		log.Printf(h.name+": create request failed %s", err)
		return
	}

	// headers
	for key, value := range h.headers {
		req.Header.Set(key, value)
	}

	// execute request
	resp, err = h.client.Do(req)

	if err != nil {
		log.Printf(h.name+": send failed (%s)", err)
		return
	}

	if resp.StatusCode > 300 {
		// body, err := ioutil.ReadAll(resp.Body)
		// if err != nil {
		// 	log.Printf(h.name+": reading response failed (%s)", err)
		// 	return
		// }

		log.Printf(h.name+": send failed %s %d %s", h.method, resp.StatusCode, url)
	}
}
