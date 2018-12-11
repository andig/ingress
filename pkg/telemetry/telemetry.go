package telemetry

import (
	"runtime"
	"sync"
	"time"

	"github.com/andig/ingress/pkg/data"
)

type MetricProvider interface {
	GetMetrics() []data.Data
}

type Telemetry struct {
	providers []MetricProvider
	mux       sync.Mutex
}

func NewTelemetry() *Telemetry {
	telemetry := &Telemetry{
		providers: make([]MetricProvider, 1),
	}

	telemetry.providers[0] = telemetry
	return telemetry
}

func (h *Telemetry) AddProvider(provider MetricProvider) {
	h.mux.Lock()
	defer h.mux.Unlock()

	// don't add providers twice
	for _, p := range h.providers {
		if p == provider {
			return
		}
	}

	h.providers = append(h.providers, provider)
}

func (h *Telemetry) Run(out chan data.Data) {
	for {
		time.Sleep(time.Duration(1000 * time.Millisecond))

		for _, provider := range h.providers {
			data := provider.GetMetrics()
			for _, data := range data {
				out <- data
			}
		}
	}
}

func (h *Telemetry) GetMetrics() []data.Data {
	var memstats runtime.MemStats

	ts := data.Timestamp()
	runtime.ReadMemStats(&memstats)

	data := []data.Data{
		data.Data{
			Timestamp: ts,
			Name:      "NumGoroutine",
			Value:     float64(runtime.NumGoroutine()),
		},
		data.Data{
			Timestamp: ts,
			Name:      "Alloc",
			Value:     float64(memstats.Alloc),
		},
	}

	return data
}