package influxdb2

import (
	"errors"
	"net/url"

	influx "github.com/influxdata/influxdb-client-go"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"
	"github.com/andig/ingress/pkg/registry"
)

func init() {
	registry.RegisterTarget("influxdb2", NewFromTargetConfig)
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
	client *influx.Client
	points      []*influxdb.RowMetric
	bucket      string
	org         string
	measurement string
}

// NewFromTargetConfig creates influxdb data target
func NewFromTargetConfig(g config.Generic) (p api.Target, err error) {
	var c influxConfig
	err = config.Decode(g, &c)
	if err != nil {
		return nil, err
	}

	options := []influxdb.Option{influxdb.WithAddress(c.URL)}
	if c.Token != "" {
		options = append(options, influxdb.WithToken(c.Token))
	} else {
		options = append(options, influxdb.WithUserAndPass(c.User, c.Password))
	}

	http := &http.Client{Timeout: writeTimeout}
	client, err := influxdb.New(http, options...)
	if err != nil {
		log.Fatalf("error creating client: %v", err)
	}

	if bucket == "" {
		log.Fatal("missing bucket")
	}
	if measurement == "" {
		log.Fatal("missing measurement")
	}

	return &Influx2{
		client:      client,
		interval:    interval,
		measurement: measurement,
		bucket:      bucket,
		org:         org,
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
