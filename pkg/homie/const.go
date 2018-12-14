package homie

import "time"

const (
	propState      = "$state"
	propStateReady = "ready"
	propStateLost  = "lost"

	propName     = "$name"
	propUnit     = "$unit"
	propDatatype = "$datatype"

	timeout = 500 * time.Millisecond
)
