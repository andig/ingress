package telemetry

import (
	"runtime"
	"sync"
	"time"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/data"
)

type MetricProvider interface {
	GetMetrics() []api.Data
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

func (h *Telemetry) Run(out chan api.Data) {
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

func (h *Telemetry) GetMetrics() []api.Data {
	var memstats runtime.MemStats

	runtime.ReadMemStats(&memstats)
	data := []api.Data{
		data.New("NumGoroutine", float64(runtime.NumGoroutine())),
		data.New("Alloc", float64(memstats.Alloc)),
	}

	return data
}
