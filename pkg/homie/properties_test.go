package homie

import "testing"

func TestEmpty(t *testing.T) {
	d := NewPropertySet()

	if s := d.All(); len(s) != 0 {
		t.Errorf("unexpected %s", s)
	}
}
func TestAddContains(t *testing.T) {
	d := NewPropertySet("foo")
	if s := d.All(); len(s) != 1 {
		t.Errorf("unexpected %s", s)
	}
	if !d.Contains("foo") {
		t.Errorf("contains didn't find expected element")
	}

	d.Add("bar")
	if s := d.All(); len(s) != 2 {
		t.Errorf("unexpected %s", s)
	}
	if !(d.Contains("foo") && d.Contains("bar")) {
		t.Errorf("contains didn't find expected element")
	}
}

func TestRenoveContains(t *testing.T) {
	d := NewPropertySet("foo", "bar", "baz")
	if s := d.All(); len(s) != 3 {
		t.Errorf("unexpected %s", s)
	}
	if !(d.Contains("foo") && d.Contains("bar") && d.Contains("baz")) {
		t.Errorf("contains didn't find expected element")
	}

	if !d.Remove("bar") {
		t.Errorf("couldn't remove element")
	}
	if s := d.All(); len(s) != 2 {
		t.Errorf("unexpected %s", s)
	}
	if !(d.Contains("foo") && d.Contains("baz")) {
		t.Errorf("contains didn't find expected element")
	}
}
