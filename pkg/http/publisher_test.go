package http

import (
	"testing"

	"github.com/andig/ingress/pkg/config"
	"gopkg.in/yaml.v2"
)

func TestCreate(t *testing.T) {
	yml := `
name: test
url: https://demo.volkszaehler.org/middleware.php/data/{name}.json
method: POST
headers:
  Content-type: application/json
  Accept: application/json
payload: >-
  [[{timestamp:ms},{value}]]`

	var c config.Generic
	err := yaml.Unmarshal([]byte(yml), &c)
	if err != nil {
		t.Error(err)
	}

	_, err = NewFromTargetConfig(c)
	if err != nil {
		t.Error(err)
	}
}
