package main

import (
	"log"
	"time"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/homie"
	"github.com/andig/ingress/pkg/wiring"
)

func inject() {
	dev := &homie.Device{
		Name: "meter1",
		Nodes: []*homie.Node{
			&homie.Node{
				Name: "zaehlwerk1",
				Properties: []*homie.Property{
					&homie.Property{
						Name: "power",
					},
					&homie.Property{
						Name: "zaehlerstand",
					},
				},
			},
		},
	}
	_ = dev
}

func main() {
	var c config.Config
	c.LoadConfig("config.yml")
	log.Println(c.Wiring)
	log.Println(c.Mapping)

	connectors := wiring.NewConnectors(c.Input, c.Output)
	mapper := wiring.NewMapper(c.Wiring, connectors.Output)
	go connectors.Run(mapper)

	// test data
	inject()

	time.Sleep(3 * time.Second)
}
