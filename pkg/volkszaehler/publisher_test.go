package volkszaehler

import (
	"testing"

	"github.com/andig/ingress/pkg/config"
)

func TestCreate(t *testing.T) {
	c := config.Generic{
		"name": "test",
		"url":  "https://demo.volkszaehler.org/",
	}

	_, err := NewFromTargetConfig(c)
	if err != nil {
		t.Error(err)
	}
}
