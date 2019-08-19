package influxdb2

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	influx "github.com/influxdata/influxdb-client-go"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"
	"github.com/andig/ingress/pkg/registry"
)

const (
	writeTimeout = 30 * time.Second
)

func init() {
	registry.RegisterTarget("influxdb2", NewFromTargetConfig)
}

type influxConfig = struct {
	config.Target `yaml:",squash"`
	URL           string
	Token         string            `yaml:"token"`
	Bucket        string            `yaml:"bucket"`
	Org           string            `yaml:"org"`
	Measurement   string            `yaml:"measurement"`
	Fields        map[string]string `yaml:"fields,omitempty"`
	Tags          map[string]string `yaml:"tags,omitempty"`
}

// Publisher is the influxdb data taerget
type Publisher struct {
	influxConfig
	client *influx.Client
}

// NewFromTargetConfig creates influxdb data target
func NewFromTargetConfig(g config.Generic) (api.Target, error) {
	var c influxConfig
	if err := config.Decode(g, &c); err != nil {
		return nil, err
	}

	options := []influx.Option{influx.WithAddress(c.URL)}
	if c.Token != "" {
		options = append(options, influx.WithToken(c.Token))
	} else {
		options = append(options, influx.WithUserAndPass(c.User, c.Password))
	}

	http := &http.Client{Timeout: writeTimeout}
	client, err := influx.New(http, options...)
	if err != nil {
		return nil, fmt.Errorf("error creating client: %v", err)
	}

	if c.Bucket == "" {
		return nil, errors.New("missing bucket")
	}
	if c.Measurement == "" {
		return nil, errors.New("missing measurement")
	}

	p := &Publisher{
		influxConfig: c,
		client:       client,
	}

	go p.ping()

	return p, nil
}

// Publish implements api.Source
func (p *Publisher) Publish(d api.Data) {
	metrics := []influx.Metric{
		p.dataToPoint(d),
	}

	if err := p.client.Write(context.Background(), p.Bucket, p.Org, metrics...); err != nil {
		log.Context(
			log.TGT, p.Name,
		).Error(err)
	}
}

func (p *Publisher) dataToPoint(d api.Data) *influx.RowMetric {
	measurement := d.MatchPattern(p.Measurement)

	fields := make(map[string]interface{})
	for k, v := range p.Fields {
		fields[k] = d.MatchPattern(v)
	}

	tags := make(map[string]string)
	for k, v := range p.Tags {
		tags[k] = d.MatchPattern(v)
	}

	point := influx.NewRowMetric(
		fields,
		measurement,
		tags,
		d.Timestamp(),
	)
	return point
}

func (p *Publisher) ping() {
	if err := p.client.Ping(context.Background()); err != nil {
		log.Context(
			log.TGT, p.Name,
		).Error(err)
	}
}
