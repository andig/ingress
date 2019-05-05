package homie

import "time"

type Property string

const (
	State      Property = "$state" // device
	StateReady Property = "ready"
	StateLost  Property = "lost"
	Properties Property = "$properties" // node
	Name       Property = "$name"       // property
	Unit       Property = "$unit"
	DataType   Property = "$datatype"

	timeout = 500 * time.Millisecond
)
