package api

import "time"

// Data is the event interface
type Data interface {
	EventID() int64

	String() string

	Name() string
	SetName(name string)

	Value() float64
	SetValue(value float64)

	Timestamp() time.Time
	SetTimestamp(timestamp time.Time)
	TimestampForPrecision(precision string) int64

	ValStr() string
	Normalize()
	MatchPattern(s string) string
}

type Source interface {
	Run(receiver chan Data)
}

// Target is the interface data targets must implement
type Target interface {
	Publish(d Data)
}

// Action is the interface data targets must implement
type Action interface {
	Process(d Data) Data
}
