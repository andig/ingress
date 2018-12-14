package homie

import "time"

const (
	// device
	propState      = "$state"
	propStateReady = "ready"
	propStateLost  = "lost"

	// node
	propProperties = "$properties"

	// property
	propName     = "$name"
	propUnit     = "$unit"
	propDatatype = "$datatype"

	timeout = 500 * time.Millisecond
)
