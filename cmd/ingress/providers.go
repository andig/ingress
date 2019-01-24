package main

// Register providers here. By importing they will be put into the provider registry.

import (
	_ "github.com/andig/ingress/pkg/homie"
	_ "github.com/andig/ingress/pkg/http"
	_ "github.com/andig/ingress/pkg/mqtt"
	_ "github.com/andig/ingress/pkg/volkszaehler"
)
