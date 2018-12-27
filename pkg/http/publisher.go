package http

import (
	"bytes"
	transport "net/http"
	"strings"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	. "github.com/andig/ingress/pkg/log"
)

// Publisher is the HTTP data target
type Publisher struct {
	name    string
	url     string
	method  string
	headers map[string]string
	payload string
	client  *transport.Client
}

// NewFromTargetConfig creates HTTP data target
func NewFromTargetConfig(c config.Target) api.Target {
	method := strings.ToUpper(c.Method)
	if method == "" {
		method = "GET"
	}
	if method != "GET" && method != "POST" {
		Log(TGT, c.Name).Fatal("invalid method " + c.Method)
	}
	if method == "POST" && c.Payload == "" {
		Log(TGT, c.Name).Fatal("missing payload configuration for POST method")
	}
	if method == "GET" && c.Payload != "" {
		Log(TGT, c.Name).Fatal("invalid payload configuration for GET method")
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

// Discover implements api.Source
func (h *Publisher) Discover() {
}

// Publish implements api.Source
func (h *Publisher) Publish(d api.Data) {
	url := d.MatchPattern(h.url)
	Log(TGT, h.name).Debugf("%s %s", h.method, url)

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
		Log(TGT, h.name).Errorf("create request failed %s", err)
		return
	}

	// headers
	for key, value := range h.headers {
		req.Header.Set(key, value)
	}

	// requestDump, err := httputil.DumpRequest(req, true)
	// if err != nil {
	// 	Log(TGT, h.name).Error(err)
	// }

	// execute request
	resp, err = h.client.Do(req)
	if err != nil {
		Log(TGT, h.name).Errorf("send failed %s", err)
		return
	}

	if resp.StatusCode >= 300 {
		Log(TGT, h.name).Errorf("%s %s %d", h.method, url, resp.StatusCode)
	}
}
