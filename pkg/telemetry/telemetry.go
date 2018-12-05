package telemetry

import (
	"log"
	"runtime"
	"time"

	"github.com/andig/ingress/pkg/data"
)

type Telemetry struct {
}

func (h *Telemetry) Run(out chan data.Data) {
	for {
		time.Sleep(time.Duration(1000 * time.Millisecond))

		var memstats runtime.MemStats
		runtime.ReadMemStats(&memstats)

		data := []data.Data{
			data.Data{
				Name:  "NumGoroutine",
				Value: float64(runtime.NumGoroutine()),
			},
			data.Data{
				Name:  "Alloc",
				Value: float64(memstats.Alloc),
			},
		}

		for _, data := range data {
			log.Printf("telemetry: %v", data)
			out <- data
		}
	}
}
