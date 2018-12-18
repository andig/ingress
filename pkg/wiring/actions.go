package wiring

import (
	"log"
	"strings"

	"github.com/andig/ingress/pkg/api"
	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/data"
)

type Actions struct {
	actions map[string]api.Action
}

func NewActions(actions []config.Action) *Actions {
	actionsMap := make(map[string]api.Action, 0)

	for _, action := range actions {
		var a api.Action

		switch strings.ToLower(action.Type) {
		case "aggregate":
			a = &AggregateAction{
				mode: action.Mode,
			}
		}

		if _, ok := actionsMap[action.Name]; ok {
			log.Fatal("configuration error: cannot redefine action " + action.Name)
		}

		actionsMap[action.Name] = a
	}

	a := &Actions{
		actions: actionsMap,
	}
	return a
}

type AggregateAction struct {
	mode string
}

func (a *AggregateAction) Process(d data.Data) {
	log.Println("AggregateAction")
}
