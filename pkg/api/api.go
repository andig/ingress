package api

// Data is the event interface
type Data interface {
	GetEventID() int64

	GetName() string
	SetName(name string)

	GetValue() float64
	SetValue(value float64)

	GetTimestamp() int64
	SetTimestamp(timestamp int64)

	ValStr() string
	Normalize()
	MatchPattern(s string) string
}

type Source interface {
	// NewFromSourceConfig(c config.Source)
	Run(receiver chan Data)
}

// Target is the interface data targets must implement
type Target interface {
	// NewFromTargetConfig(c config.Target)
	Publish(d Data)
}

// Action is the interface data targets must implement
type Action interface {
	// NewFromActionConfig(c config.Action)
	Process(d Data)
}
