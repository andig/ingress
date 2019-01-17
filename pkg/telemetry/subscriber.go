package telemetry

import "github.com/andig/ingress/pkg/config"

func NewFromSourceConfig(c config.Source) (*Telemetry, error) {
	t := NewTelemetry()
	return t, nil
}
