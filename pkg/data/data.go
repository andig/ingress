package data

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/andig/ingress/pkg/api"
)

var mux sync.Mutex
var eventID int64
var patternRegex = regexp.MustCompile(`@\w+?@`)

type Data struct {
	EventID   int64
	ID        string
	Name      string
	Timestamp int64
	Value     float64
}

func Timestamp() int64 {
	return int64(time.Now().UnixNano() / 1e6)
}

func GenerateEventID() int64 {
	mux.Lock()
	defer mux.Unlock()
	eventID++
	return eventID
}

// NewData creates data event with consecutive id and current timestamp
func NewData(name string, value float64) api.Data {
	return &Data{
		EventID:   GenerateEventID(),
		Timestamp: Timestamp(),
		Name:      name,
		Value:     value,
	}
}

func (d *Data) String() string {
	return fmt.Sprintf("%s:%s@%d", d.Name, d.ValStr(), d.Timestamp)
}

func (d *Data) GetEventID() int64 {
	return d.EventID
}

func (d *Data) GetName() string {
	return d.Name
}

func (d *Data) SetName(name string) {
	d.Name = name
}

func (d *Data) GetValue() float64 {
	return d.Value
}

func (d *Data) SetValue(value float64) {
	d.Value = value
}

func (d *Data) GetTimestamp() int64 {
	return d.Timestamp
}

func (d *Data) SetTimestamp(timestamp int64) {
	d.Timestamp = timestamp
}

func (d *Data) Normalize() {
	if d.Timestamp == 0 {
		d.Timestamp = Timestamp()
	}

	// if d.ID == "" {
	// 	d.ID = d.Name
	// } else if d.Name == "" {
	// 	d.Name = d.ID
	// }
}

func (d *Data) ValStr() string {
	return fmt.Sprintf("%.3f", d.Value)
}

func (d *Data) MatchPattern(s string) string {
	matches := patternRegex.FindAllString(s, -1)
	for _, match := range matches {
		switch match {
		case "@id@":
			s = strings.Replace(s, match, d.ID, -1)
		case "@name@":
			s = strings.Replace(s, match, d.Name, -1)
		case "@value@":
			s = strings.Replace(s, match, d.ValStr(), -1)
		case "@timestamp@":
			s = strings.Replace(s, match, strconv.FormatInt(d.Timestamp, 10), -1)
			// default:
			// 	log.log.Fatalf("Invalid match pattern %s", s)
		}
	}

	return s
}
