package actions

import (
	"strings"
	"sync"
	"time"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"
	"github.com/andig/ingress/pkg/queue"
	"github.com/andig/ingress/pkg/registry"
)

func init() {
	registry.RegisterAction("aggsum", NewAggSumFromActionConfig)
	registry.RegisterAction("aggmax", NewAggMaxFromActionConfig)
	registry.RegisterAction("aggavg", NewAggAvgFromActionConfig)
}

type aggregateConfig struct {
	config.Action `yaml:",squash"`
	Period        time.Duration `yaml:"period"`
}

func NewAggSumFromActionConfig(g config.Generic) (a api.Action, err error) {
	return newAggregateActionOfType("sum", g)
}

func NewAggMaxFromActionConfig(g config.Generic) (a api.Action, err error) {
	return newAggregateActionOfType("max", g)
}

func NewAggAvgFromActionConfig(g config.Generic) (a api.Action, err error) {
	return newAggregateActionOfType("avg", g)
}

// newAggregateActionOfType creates an aggregatoin action of desired type
func newAggregateActionOfType(mode string, g config.Generic) (res api.Action, err error) {
	var conf aggregateConfig
	err = config.Decode(g, &conf)
	if err != nil {
		return nil, err
	}

	a := &aggregateAction{
		aggregateConfig: conf, // embed config for Action.String()
		events:          make(map[string]*event),
		period:          conf.Period,
	}

	switch strings.ToLower(mode) {
	case "max":
		res = &AggregateMaxAction{
			aggregateAction: a,
		}
	case "sum":
		res = &AggregateSumAction{
			aggregateAction: a,
		}
	case "avg":
		res = &AggregateAvgAction{
			aggregateAction: a,
		}
	default:
		log.Fatalf("Invalid aggregation mode %s", mode)
	}

	return res, nil
}

type event struct {
	acc        float64
	lastUpdate time.Time
	queue      *queue.Queue
}

type aggregateAction struct {
	aggregateConfig
	mux    sync.Mutex
	events map[string]*event
	period time.Duration
}

func (a *aggregateAction) process(d api.Data,
	initfunc func(*event),
	updatefunc func(*event),
	resultfunc func(*event) api.Data,
) api.Data {
	a.mux.Lock()
	defer a.mux.Unlock()

	if ev, ok := a.events[d.Name()]; ok {
		// update the value
		updatefunc(ev)

		// make sure time.After passes when timestamps identical
		periodEnd := ev.lastUpdate.Add(a.period)
		if d.Timestamp().After(periodEnd) || d.Timestamp().Equal(periodEnd) {
			d := resultfunc(ev)
			ev.lastUpdate = d.Timestamp()

			log.Context(
				log.EV, d.Name(),
				log.ACT, a.Name,
			).Debug("aggregate result")

			return d
		}
	} else {
		// first event is swallowed
		ev := &event{
			lastUpdate: d.Timestamp(),
		}
		initfunc(ev)
		a.events[d.Name()] = ev
	}

	log.Context(
		log.EV, d.Name(),
		log.ACT, a.Name,
	).Debug("aggregate")

	// either first value or time not elapsed
	return nil
}

// AggregateMaxAction is used for aggregating monotonic data
type AggregateMaxAction struct {
	*aggregateAction
}

// Process implements the Action interface
func (a *AggregateMaxAction) Process(d api.Data) api.Data {
	return a.process(d,
		func(ev *event) {
			ev.acc = d.Value()
		},
		func(ev *event) {
			val := d.Value()
			if val >= ev.acc {
				ev.acc = val
			} else {
				log.Context(
					log.EV, d.Name(),
					log.ACT, a.Name,
				).Warn("unexpected non-monotonic value in aggregation")
			}
		},
		func(ev *event) api.Data {
			d.SetValue(ev.acc)
			return d
		},
	)
}

// AggregateSumAction is used for aggregating monotonic data
type AggregateSumAction struct {
	*aggregateAction
}

// Process implements the Action interface
func (a *AggregateSumAction) Process(d api.Data) api.Data {
	return a.process(d,
		func(ev *event) {
			ev.acc = d.Value()
		},
		func(ev *event) {
			ev.acc += d.Value()
		},
		func(ev *event) api.Data {
			d.SetValue(ev.acc)
			ev.acc = 0
			return d
		},
	)
}

// AggregateAvgAction is used for aggregating monotonic data
type AggregateAvgAction struct {
	*aggregateAction
}

func (a *AggregateAvgAction) queueToAverage(q *queue.Queue) float64 {
	var sum float64
	var firstTimestamp, prevTimestamp time.Time

	// there will always be > 1 element in the queue
	for i := 0; i < q.Length(); i++ {
		v, err := q.Get(i)
		if err != nil {
			log.Fatalf("invalid queue access %s", err)
		}

		d := v.(api.Data)
		if i == 0 {
			firstTimestamp = d.Timestamp()
		} else {
			sum += d.Value() * float64(d.Timestamp().Sub(prevTimestamp).Nanoseconds()/1e6)
		}
		prevTimestamp = d.Timestamp()
	}

	return sum / float64(prevTimestamp.Sub(firstTimestamp).Nanoseconds()/1e6)
}

// Process implements the Action interface.
// The first data element of the queue represents the
// last timestamp from the previous aggregation cycle.
func (a *AggregateAvgAction) Process(d api.Data) api.Data {
	return a.process(d,
		func(ev *event) {
			ev.queue = queue.New()
			ev.queue.Add(d)
		},
		func(ev *event) {
			ev.queue.Add(d)
		},
		func(ev *event) api.Data {
			d.SetValue(a.queueToAverage(ev.queue))
			ev.queue = queue.New() // clear queue
			ev.queue.Add(d)        // re-add last timestamp
			return d
		},
	)
}
