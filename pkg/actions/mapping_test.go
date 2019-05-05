package actions

import (
	"fmt"
	"testing"

	"github.com/andig/ingress/pkg/config"
	"gopkg.in/yaml.v2"
)

func TestKeys(t *testing.T) {
	yml := `
foo: bar
foo bar: bar baz`

	var c config.Generic
	err := yaml.Unmarshal([]byte(yml), &c)
	if err != nil {
		t.Error(err)
	}

	c1 := make(map[string]string)
	err = yaml.Unmarshal([]byte(yml), &c1)
	if err != nil {
		t.Error(err)
	}

	fmt.Println(c1)
}
