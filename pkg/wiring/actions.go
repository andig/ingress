package wiring

import (
	"errors"
	"strings"
	"time"

	"github.com/andig/ingress/pkg/actions"
	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	. "github.com/andig/ingress/pkg/log"
)

type Actions struct {
	actions map[string]api.Action
}

func NewActions(c []config.Action) *Actions {
	actionsMap := make(map[string]api.Action, 0)

	for _, action := range c {
		var a api.Action

		switch strings.ToLower(action.Type) {
		case "aggregate":
			period, err := time.ParseDuration(action.Period)
			if err != nil {
				Log().Fatalf("configuration error: %s", err)
			}
			a = actions.NewAggregateAction(action.Mode, period)
		}

		if _, ok := actionsMap[action.Name]; ok {
			Log().Fatal("configuration error: cannot redefine action " + action.Name)
		}

		actionsMap[action.Name] = a
	}

	a := &Actions{
		actions: actionsMap,
	}
	return a
}

func (a *Actions) ActionForName(name string) (api.Action, error) {
	action, ok := a.actions[name]
	if !ok {
		return nil, errors.New("configuration error: undefined action " + name)
	}
	return action, nil
}
