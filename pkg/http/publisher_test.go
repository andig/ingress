package http

import (
	"testing"

	"github.com/andig/ingress/pkg/config"
)

func TestCreate(t *testing.T) {
	c := config.Target{
		Name:   "test",
		URL:    "https://demo.volkszaehler.org/middleware.php/data/@name@.json",
		Method: "POST",
		Headers: map[string]string{
			"Content-type": "application/json",
			"Accept":       "application/json",
		},
		Payload: "[[%timestamp%,%value%]]",
	}

	_, err := NewFromTargetConfig(c)
	if err != nil {
		t.Error(err)
	}
}
