package cmd

// Register providers here. By importing they will be put into the provider registry.

import (
	// import all actions and providers
	_ "github.com/andig/ingress/pkg/actions"
	_ "github.com/andig/ingress/pkg/homie"
	_ "github.com/andig/ingress/pkg/http"
	_ "github.com/andig/ingress/pkg/influxdb"
	_ "github.com/andig/ingress/pkg/mqtt"
	_ "github.com/andig/ingress/pkg/volkszaehler"
)
