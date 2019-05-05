package actions

import (
	"testing"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"
	"gopkg.in/yaml.v2"
)

func TestPassFilterMatches(t *testing.T) {
	yml := `
matches:
- foo
- bar`

	var c config.Generic
	err := yaml.Unmarshal([]byte(yml), &c)
	if err != nil {
		t.Error(err)
	}

	a, err := NewPassFilterFromActionConfig(c)
	if err != nil {
		t.Error(err)
	}

	d := data.New("foo", 1)
	expectNil(t, a.Process(d))
}
