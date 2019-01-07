package data

import "testing"

func TestMatchPattern(t *testing.T) {
	d := NewData("NAME", 1.234234)
	d.SetTimestamp(1)

	if s := d.MatchPattern("%name%"); s != "NAME" {
		t.Errorf("name not replaced, got %s", s)
	}
	if s := d.MatchPattern("%value%"); s != "1.234" {
		t.Errorf("value not replaced, got %s", s)
	}
	if s := d.MatchPattern("%timestamp%"); s != "1" {
		t.Errorf("timestamp not replaced, got %s", s)
	}
}

func TestEventID(t *testing.T) {
	// reset counter
	eventID = 0

	d := NewData("NAME", 1.234234)
	if d.GetEventID() != 1 {
		t.Errorf("unexpected event id %d", d.GetEventID())
	}
	d = NewData("NAME", 1.234234)
	if d.GetEventID() != 2 {
		t.Errorf("unexpected event id %d", d.GetEventID())
	}
}