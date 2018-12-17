package api

import "github.com/andig/ingress/pkg/data"

// Source is the interface data sources must implement
type Source interface {
	// NewFromSourceConfig(c config.Source)
	Run(receiver chan data.Data)
}

// Target is the interface data targets must implement
type Target interface {
	// NewFromTargetConfig(c config.Target)
	Publish(d data.Data)
}

// Action is the interface data targets must implement
type Action interface {
	// NewFromActionConfig(c config.Action)
	Process(d data.Data)
}
