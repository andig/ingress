package data

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var patternRegex = regexp.MustCompile(`%\w+?%`)

type Data struct {
	ID        string
	Name      string
	Timestamp int64
	Value     float64
}

func Timestamp() int64 {
	return int64(time.Now().UnixNano() / 1e6)
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
		case "%id%":
			s = strings.Replace(s, match, d.ID, -1)
		case "%name%":
			s = strings.Replace(s, match, d.Name, -1)
		case "%value%":
			s = strings.Replace(s, match, d.ValStr(), -1)
		case "%timestamp%":
			s = strings.Replace(s, match, strconv.FormatInt(d.Timestamp, 10), -1)
		}
	}

	return s
}
