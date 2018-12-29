package actions

import (
	"strings"
	"sync"
	"time"

	"github.com/andig/ingress/pkg/api"
	. "github.com/andig/ingress/pkg/log"
	"github.com/andig/ingress/pkg/queue"
)

// NewAggregateAction creates an aggregatoin action of desired type
func NewAggregateAction(mode string, period time.Duration) (res api.Action) {
	a := &aggregateAction{
		events: make(map[string]*event),
		period: period,
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
		Log().Fatalf("Invalid aggregation mode %s", mode)
	}

	return res
}

type event struct {
	acc        float64
	lastUpdate int64
	queue      *queue.Queue
}

type aggregateAction struct {
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

	if ev, ok := a.events[d.GetName()]; ok {
		// update the value
		updatefunc(ev)

		// make sure time.After passes when timestamps identical
		now := time.Unix(0, d.GetTimestamp()*1e6)
		lastUpdate := time.Unix(0, ev.lastUpdate*1e6-1)
		if now.After(lastUpdate.Add(a.period)) {
			d := resultfunc(ev)
			ev.lastUpdate = d.GetTimestamp()
			return d
		}
	} else {
		// first event is swallowed
		ev := &event{
			lastUpdate: d.GetTimestamp(),
		}
		initfunc(ev)
		a.events[d.GetName()] = ev
	}

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
			ev.acc = d.GetValue()
		},
		func(ev *event) {
			val := d.GetValue()
			if val >= ev.acc {
				ev.acc = val
			} else {
				Log(EV, d.GetName()).Warn("unexpected non-monotonic value in aggregation")
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
			ev.acc = d.GetValue()
		},
		func(ev *event) {
			ev.acc += d.GetValue()
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
	var firstTimestamp, prevTimestamp int64

	// there will always be > 1 element in the queue
	for i := 0; i < q.Length(); i++ {
		v, err := q.Get(i)
		if err != nil {
			Log().Fatalf("invalid queue access %s", err)
		}

		d := v.(api.Data)
		if i == 0 {
			firstTimestamp = d.GetTimestamp()
		} else {
			sum += d.GetValue() * float64(d.GetTimestamp()-prevTimestamp)
		}
		prevTimestamp = d.GetTimestamp()
	}

	return sum / float64(prevTimestamp-firstTimestamp)
}

// Process implements the Action interface.
// The first data element of the queue represents the
// last timestamp from the previos aggregation cycle.
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
