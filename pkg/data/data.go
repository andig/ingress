package data

import "time"

type Data struct {
	ID        string
	Name      string
	Timestamp int64
	Value     float64
}

func Timestamp() int64 {
	return int64(time.Now().UnixNano() / 1e6)
}
