package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"sort"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"
	mq "github.com/andig/ingress/pkg/mqtt"
	"github.com/andig/ingress/pkg/registry"
	"github.com/andig/ingress/pkg/wiring"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	latest "github.com/tcnksm/go-latest"
)

const (
	version = "unknown version"
	commit  = "unknown commit"
)

func inject() {
	mqttOptions := mq.NewMqttClientOptions("tcp://localhost:1883", "", "")
	mqttClient := mqtt.NewClient(mqttOptions)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("mqtt: error connecting: %s", token.Error())
	}

	time.Sleep(200 * time.Millisecond)
	token := mqttClient.Publish("input/inject", 0, false, "3.14")
	if token.WaitTimeout(100 * time.Millisecond) {
		log.Println("inject done")
	}

	mqttClient.Publish("homie/meter1/$nodes", 1, true, "zaehlwerk1")
	mqttClient.Publish("homie/meter1/zaehlwerk1/$properties", 1, true, "power")
	mqttClient.Publish("homie/meter1/zaehlwerk1/power/$name", 1, true, "Leistung")
	mqttClient.Publish("homie/meter1/zaehlwerk1/power/$unit", 1, true, "W")
	mqttClient.Publish("homie/meter1/zaehlwerk1/power/$datatype", 1, true, "float")
	mqttClient.Publish("homie/meter1/zaehlwerk1/power", 1, false, "3048")
}

func checkVersion() {
	githubTag := &latest.GithubTag{
		Owner:      "andig",
		Repository: "ingress",
	}

	if res, err := latest.Check(githubTag, version); err == nil {
		if res.Outdated {
			log.Warnf("updates available - please upgrade to %s", res.Current)
		}
	}
}

func displayCapabilities() {
	sources := make([]string, 0)
	for k := range registry.SourceProviders {
		sources = append(sources, k)
	}
	targets := make([]string, 0)
	for k := range registry.TargetProviders {
		targets = append(targets, k)
	}
	actions := make([]string, 0)
	for k := range registry.ActionProviders {
		actions = append(actions, k)
	}
	sort.Strings(sources)
	sort.Strings(targets)
	sort.Strings(actions)

	fmt.Printf(`
Available configuration options:

	sources:	%s
	targets:	%s
	actions:	%s
	`,
		strings.Join(sources, ", "),
		strings.Join(targets, ", "),
		strings.Join(actions, ", "),
	)
}

func waitForCtrlC() {
	var wg sync.WaitGroup
	wg.Add(1)
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-channel
		wg.Done()
	}()
	wg.Wait()
}

// main
func run(cmd *cobra.Command, args []string) {
	pflags := cmd.PersistentFlags()

	log.Configure(cfgLogLevel)
	log.Printf("ingress %s %s", version, commit)

	if caps, _ := pflags.GetBool("capabilities"); caps {
		displayCapabilities()
		os.Exit(0)
	}

	var conf config.Config
	log.Printf("using %s", viper.ConfigFileUsed())
	if err := viper.Unmarshal(&conf); err != nil {
		log.Fatalf("failed parsing config file: %v", err)
	}

	if dump, _ := pflags.GetBool("dump"); dump {
		conf.Dump()
		os.Exit(0)
	}

	go checkVersion()

	connectors := wiring.NewConnectors(conf.Sources, conf.Targets)
	actions := wiring.NewActions(conf.Actions)
	wires := wiring.NewWiring(conf.Wires, connectors, actions)
	mapper := wiring.NewMapper(wires, connectors)
	_ = actions
	_ = mapper

	ctx, cancel := context.WithCancel(context.Background())
	go connectors.Run(ctx, mapper)

	if test, _ := pflags.GetBool("test"); test {
		time.Sleep(time.Second)
		inject()
	}

	if diagnose, _ := pflags.GetBool("diagnose"); diagnose {
		go func() {
			for {
				time.Sleep(time.Second)
				var memstats runtime.MemStats
				runtime.ReadMemStats(&memstats)
				log.Debugf("%db\n", memstats.Alloc)
			}
		}()
	}

	waitForCtrlC()
	cancel() // cancel context
}
