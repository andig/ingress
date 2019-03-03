package telemetry

import (
	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/registry"
)

func init() {
	registry.RegisterSource("telemetry", NewFromSourceConfig)
}

func NewFromSourceConfig(c config.Generic) (api.Source, error) {
	t := NewTelemetry()
	return t, nil
}
