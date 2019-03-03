package data

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/log"
)

var mux sync.Mutex
var eventID int64
var patternRegex = regexp.MustCompile(`{(\w+?)(:([\w\d%:. ]+))?}`)

type Data struct {
	eventID   int64
	ID        string
	name      string
	timestamp time.Time
	value     float64
}

func GenerateEventID() int64 {
	mux.Lock()
	defer mux.Unlock()
	eventID++
	return eventID
}

// New creates data event with consecutive id and current timestamp
func New(name string, value float64, timestamp ...time.Time) api.Data {
	var ts time.Time
	if len(timestamp) > 0 {
		ts = timestamp[0]
	} else {
		ts = time.Now()
	}

	return &Data{
		eventID:   GenerateEventID(),
		timestamp: ts,
		name:      name,
		value:     value,
	}
}

func (d *Data) String() string {
	return fmt.Sprintf("%s:%s@%d", d.name, d.ValStr(), d.timestamp.UnixNano()/1e6)
}

func (d *Data) EventID() int64 {
	return d.eventID
}

func (d *Data) Name() string {
	return d.name
}

func (d *Data) SetName(name string) {
	d.name = name
}

func (d *Data) Value() float64 {
	return d.value
}

func (d *Data) SetValue(value float64) {
	d.value = value
}

// Timestamp returns ms timestamp
func (d *Data) Timestamp() time.Time {
	return d.timestamp
}

// SetTimestamp sets ms timestamp
func (d *Data) SetTimestamp(timestamp time.Time) {
	d.timestamp = timestamp
}

func (d *Data) Normalize() {
	// if d.timestamp == 0 {
	// 	d.timestamp = Timestamp()
	// }

	// if d.ID == "" {
	// 	d.ID = d.name
	// } else if d.name == "" {
	// 	d.name = d.ID
	// }
}

func (d *Data) ValStr() string {
	return fmt.Sprintf("%.3f", d.value)
}

func (d *Data) TimestampForPrecision(precision string) int64 {
	switch strings.ToLower(precision) {
	case "s":
		return d.timestamp.Unix()
	case "ms":
		return d.timestamp.UnixNano() / 1e6
	case "us":
		return d.timestamp.UnixNano() / 1e3
	case "ns":
		return d.timestamp.UnixNano()
	default:
		panic("invalid precision: " + precision)
	}
}

func (d *Data) FormatTimestamp(format string) (res string) {
	var ts int64
	if format == "" {
		format = "ms" // default milliseconds
	}

	switch format {
	case "s", "ms", "us", "ns":
		ts = d.TimestampForPrecision(format)
	default:
		// return timestamp formatted by golang pattern
		return d.timestamp.Format(format)
	}

	res = strconv.FormatInt(ts, 10)
	return res
}

func (d *Data) MatchPattern(s string) (res string) {
	matches := patternRegex.FindAllStringSubmatch(s, -1)

	for _, match := range matches {
		var val string
		literal := match[0]
		name := strings.ToLower(match[1])
		format := match[3]

		switch name {
		case "id":
			val = d.ID
			if format != "" {
				val = fmt.Sprintf(format, val)
			}
		case "name":
			val = d.name
			if format != "" {
				val = fmt.Sprintf(format, val)
			}
		case "value":
			if format == "" {
				format = "%.3f"
			}
			val = fmt.Sprintf(format, d.value)
		case "timestamp":
			val = d.FormatTimestamp(format)
		default:
			log.Fatalf("Invalid name pattern %s", s)
		}

		s = strings.Replace(s, literal, val, -1)
	}

	return s
}
