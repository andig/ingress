package volkszaehler

import (
	"testing"

	"github.com/andig/ingress/pkg/config"
)

func TestCreate(t *testing.T) {
	c := config.Target{
		Name: "test",
		URL:  "https://demo.volkszaehler.org/",
	}

	_, err := NewFromTargetConfig(c)
	if err != nil {
		t.Error(err)
	}
}
