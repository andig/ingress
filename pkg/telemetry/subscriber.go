package telemetry

import "github.com/andig/ingress/pkg/config"

func NewFromInputConfig(c config.Input) *Telemetry {
	t := NewTelemetry()
	return t
}
