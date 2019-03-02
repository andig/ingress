package data

import (
	"testing"
	"time"
)

func TestString(t *testing.T) {
	d := New("NAME", 1.234234)
	d.SetTimestamp(time.Unix(1, 0))

	if s := d.String(); s != "NAME:1.234@1000" {
		t.Errorf("string not correct, got %s", s)
	}
}

func TestMatchPattern(t *testing.T) {
	d := New("NAME", 1.234234)
	d.SetTimestamp(time.Unix(1, 0))

	if s := d.MatchPattern("{name}"); s != "NAME" {
		t.Errorf("name not replaced, got %s", s)
	}
	if s := d.MatchPattern("{value}"); s != "1.234" {
		t.Errorf("value not replaced, got %s", s)
	}
	if s := d.MatchPattern("{timestamp}"); s != "1000" {
		t.Errorf("timestamp not replaced, got %s", s)
	}
}

func TestMatchPatternFormat(t *testing.T) {
	d := New("NAME", 1.234234)
	d.SetTimestamp(time.Unix(1, 0))

	if s := d.MatchPattern("{name:name:%8s}"); s != "name:    NAME" {
		t.Errorf("name not replaced, got %s", s)
	}
	if s := d.MatchPattern("{value:%.0f}"); s != "1" {
		t.Errorf("value not replaced, got %s", s)
	}
	if s := d.MatchPattern("{timestamp:ms}"); s != "1000" {
		t.Errorf("ms timestamp not replaced, got %s", s)
	}
	if s := d.MatchPattern("{timestamp:s}"); s != "1" {
		t.Errorf("s timestamp not replaced, got %s", s)
	}
	if s := d.MatchPattern("{timestamp:us}"); s != "1000000" {
		t.Errorf("us timestamp not replaced, got %s", s)
	}
	if s := d.MatchPattern("{timestamp:ns}"); s != "1000000000" {
		t.Errorf("ns timestamp not replaced, got %s", s)
	}
	if s := d.MatchPattern("{timestamp:02.01.2006 15:04:05.999 Z07}"); s != "01.01.1970 01:00:01 +01" {
		t.Errorf("formatted timestamp not replaced, got %s", s)
	}
}

func TestEventID(t *testing.T) {
	// reset counter
	eventID = 0

	d := New("NAME", 1.234234)
	if d.EventID() != 1 {
		t.Errorf("unexpected event id %d", d.EventID())
	}
	d = New("NAME", 1.234234)
	if d.EventID() != 2 {
		t.Errorf("unexpected event id %d", d.EventID())
	}
}
