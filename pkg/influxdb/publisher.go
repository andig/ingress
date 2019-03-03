package influxdb

import (
	"errors"
	"fmt"
	"net/url"

	influx "github.com/influxdata/influxdb1-client"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"
	"github.com/andig/ingress/pkg/registry"
)

func init() {
	registry.RegisterTarget("influxdb", NewFromTargetConfig)
}

type influxConfig = struct {
	config.Target `yaml:",squash"`
	URL           string
	Database      string
	Measurement   string
	Precision     string            `yaml:"precision"`
	Fields        map[string]string `yaml:"fields,omitempty"`
	Tags          map[string]string `yaml:"tags,omitempty"`
}

// Publisher is the influxdb data taerget
type Publisher struct {
	influxConfig
	conn *influx.Client
}

// NewFromTargetConfig creates influxdb data target
func NewFromTargetConfig(g config.Generic) (p api.Target, err error) {
	var c influxConfig
	err = config.Decode(g, &c)
	if err != nil {
		return nil, err
	}

	var uri *url.URL
	if uri, err = url.ParseRequestURI(c.URL); err != nil {
		return p, err
	}

	if c.Database == "" {
		return p, errors.New("missing database")
	}
	if c.Measurement == "" {
		return p, errors.New("missing measurement")
	}
	if c.Precision == "" {
		c.Precision = "ms" // match volkszaehler behaviour
	}

	conf := influx.Config{
		URL:      *uri,
		Username: c.User,
		Password: c.Password,
	}
	conn, err := influx.NewClient(conf)
	if err != nil {
		return p, err
	}

	p = &Publisher{
		influxConfig: c,
		conn:         conn,
	}
	p.(*Publisher).ping()

	return p, nil
}

// Publish implements api.Source
func (p *Publisher) Publish(d api.Data) {
	bps := influx.BatchPoints{
		Points:   []influx.Point{p.dataToPoint(d)},
		Database: p.Database,
	}

	response, err := p.conn.Write(bps)
	if err != nil {
		log.Context(
			log.TGT, p.Name,
		).Error(err)
		return
	}

	// log.Println(d)
	if response != nil {
		log.Context(
			log.TGT, p.Name,
		).Tracef("%v", response)
	}
}

func (p *Publisher) dataToPoint(d api.Data) influx.Point {
	measurement := d.MatchPattern(p.Measurement)

	fields := make(map[string]interface{})
	for k, v := range p.Fields {
		fields[k] = d.MatchPattern(v)
	}

	tags := make(map[string]string)
	for k, v := range p.Tags {
		tags[k] = d.MatchPattern(v)
	}

	point := influx.Point{
		Measurement: measurement,
		Fields:      fields,
		Tags:        tags,
		Time:        d.Timestamp(),
		Precision:   p.Precision,
	}
	fmt.Println(point)
	return point
}

func (p *Publisher) ping() {
	duration, version, err := p.conn.Ping()
	if err != nil {
		log.Error(err)
		return
	}

	if version == "" {
		version = "unknown"
	}

	log.Context(
		log.TGT, p.Name,
	).Debugf("version %s, ping %v", version, duration)
}
