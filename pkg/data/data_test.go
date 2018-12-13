package data

import "testing"

func TestMatchPattern(t *testing.T) {
	d := &Data{
		ID:        "ID",
		Name:      "NAME",
		Value:     1.234234,
		Timestamp: 1,
	}

	if s := d.MatchPattern("%id%"); s != "ID" {
		t.Errorf("id not replaced, got %s", s)
	}
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
