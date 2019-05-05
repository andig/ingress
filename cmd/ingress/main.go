package main

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
	"github.com/spf13/viper"
	latest "github.com/tcnksm/go-latest"
	cli "gopkg.in/urfave/cli.v1"
)

const DefaultConfig = "ingress.yml"

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

func main() {
	app := cli.NewApp()
	app.Name = "ingress"
	app.Usage = "ingress data mapper daemon"
	app.Version = fmt.Sprintf("%s (https://github.com/andig/ingress/commit/%s)", version, commit)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: DefaultConfig,
			Usage: "Config file (search path ., ~ and /etc)",
		},
		cli.BoolFlag{
			Name:  "dump, d",
			Usage: "Dump parsed config",
		},
		cli.BoolFlag{
			Name:  "capabilities, i",
			Usage: "Display available capabilities",
		},
		cli.BoolFlag{
			Name:  "diagnose",
			Usage: "Memory diagnostics",
		},
		cli.StringFlag{
			Name:  "log, l",
			Value: "debug",
			Usage: "Log level (error, info, debug, trace)",
		},
		cli.BoolFlag{
			Name:  "test, t",
			Usage: "Inject test data",
		},
	}

	app.Action = func(c *cli.Context) {
		log.Configure(c.String("log"))
		log.Printf("ingress %s %s", version, commit)

		if c.NArg() > 0 {
			log.Fatalf("unexpected arguments: %v", c.Args())
		}

		if c.Bool("capabilities") {
			displayCapabilities()
			os.Exit(0)
		}

		var conf config.Config
		viper.SetConfigType("yaml") // or viper.SetConfigType("YAML")

		if configFile := c.String("config"); configFile != DefaultConfig {
			viper.SetConfigFile(configFile) // verbose config file
		} else {
			viper.SetConfigName("ingress") // name of config file (without extension)
			viper.AddConfigPath(".")
			viper.AddConfigPath("$HOME")
			viper.AddConfigPath("/etc")
		}

		if err := viper.ReadInConfig(); err != nil { // Handle errors reading the config file
			log.Fatal(err)
		}

		log.Printf("using %s", viper.ConfigFileUsed())
		if err := viper.Unmarshal(&conf); err != nil {
			log.Fatalf("failed parsing config file: %v", err)
		}

		if c.Bool("dump") {
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

		if c.Bool("test") {
			time.Sleep(time.Second)
			inject()
		}

		if c.Bool("diagnose") {
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

	_ = app.Run(os.Args)
}
