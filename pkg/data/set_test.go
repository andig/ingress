package data

import (
	"strings"
	"testing"
)

func TestEmpty(t *testing.T) {
	d := NewSet()
	if d.Size() != 0 {
		t.Errorf("size mismatch")
	}
}

func TestAll(t *testing.T) {
	d := NewSet()
	if s := d.All(); len(s) != 0 {
		t.Errorf("size mismatch")
	}
}

func TestAdd(t *testing.T) {
	d := NewSet()
	d.Add("foo", "bar")
	if d.Size() != 1 {
		t.Errorf("size mismatch")
	}
	if s := d.All(); len(s) != 1 {
		t.Errorf("size mismatch")
	}
}

func TestContains(t *testing.T) {
	d := NewSet()
	d.Add("foo", "bar")

	if !d.Contains("foo") {
		t.Errorf("contains didn't find expected element")
	}

	d.Add("bar", "baz")
	if !(d.Contains("foo") && d.Contains("bar")) {
		t.Errorf("contains didn't find expected element")
	}
}

func TestGet(t *testing.T) {
	d := NewSet()
	d.Add("foo", "bar")

	if d.Get("foo") != "bar" {
		t.Errorf("get didn't find expected element")
	}
}

func TestRemove(t *testing.T) {
	d := NewSet()
	d.Add("foo", "bar")
	d.Add("bar", "baz")

	if d.Remove("baz") {
		t.Errorf("invalid remove")
	}

	if !d.Remove("bar") {
		t.Errorf("couldn't remove element")
	}
	if d.Size() != 1 {
		t.Errorf("unexpected size")
	}
}

func TestFilter(t *testing.T) {
	d := NewSet()
	d.Add("foo", "bar")
	d.Add("bar", "baz")

	s := d.Filter(func(key string, val interface{}) bool {
		return strings.HasPrefix(key, "b")
	})

	if s.Size() != 1 {
		t.Errorf("wrong subset size")
	}
	if !s.Contains("bar") {
		t.Errorf("wrong subset key")
	}
}
