package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/andig/ingress/pkg/config"
	"github.com/andig/ingress/pkg/log"
	mq "github.com/andig/ingress/pkg/mqtt"
	"github.com/andig/ingress/pkg/wiring"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/spf13/viper"
	latest "github.com/tcnksm/go-latest"
	cli "gopkg.in/urfave/cli.v1"
)

const DEFAULT_CONFIG = "ingress.yml"

func inject() {
	panic("foo")
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

	if res, err := latest.Check(githubTag, tag); err == nil {
		if res.Outdated {
			log.Warnf("updates available - please upgrade to %s", res.Current)
		}
	}
}

type MapUnmarshaler interface {
	UnmarshalMap(interface{}) (interface{}, error)
}

func parseConfig(c *cli.Context) (conf config.Config, err error) {
	viper.SetConfigType("yaml")

	if configFile := c.String("config"); configFile != DEFAULT_CONFIG {
		viper.SetConfigFile(configFile) // verbose config file
	} else {
		viper.SetConfigName("ingress") // name of config file (without extension)
		viper.AddConfigPath(".")
		viper.AddConfigPath("$HOME")
		viper.AddConfigPath("/etc")
	}

	if err = viper.ReadInConfig(); err != nil { // Handle errors reading the config file
		return conf, err
	}

	log.Printf("using %s", viper.ConfigFileUsed())

	decodeHook := func(from reflect.Type, to reflect.Type, v interface{}) (interface{}, error) {
		unmarshalerType := reflect.TypeOf((*MapUnmarshaler)(nil)).Elem()
		if to.Implements(unmarshalerType) {
			in := []reflect.Value{reflect.New(to).Elem(), reflect.ValueOf(v)}
			method, _ := to.MethodByName("UnmarshalMap")

			r := method.Func.Call(in)
			fmt.Printf("%v %v\n", from, to)
			fmt.Printf("%v\n", v)
			fmt.Printf("%v\n", r)
		}
		return v, nil
	}

	var meta mapstructure.Metadata
	msConf := func(conf *mapstructure.DecoderConfig) {
		conf.Metadata = &meta
		conf.WeaklyTypedInput = true
		conf.DecodeHook = decodeHook
	}

	if err = viper.Unmarshal(&conf, viper.DecoderConfigOption(msConf)); err != nil {
		return conf, errors.Wrap(err, "failed parsing config file")
	}

	if len(meta.Unused) > 0 {
		return conf, errors.New(fmt.Sprintf("invalid config entries: %v", meta.Unused))
	}

	if c.Bool("dump") {
		conf.Dump()
	}

	return conf, err
}

func waitForCtrlC() {
	var wg sync.WaitGroup
	wg.Add(1)
	channel := make(chan os.Signal, 1)
	signal.Notify(channel, os.Interrupt, os.Kill)
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
	app.Version = fmt.Sprintf("%s (https://github.com/andig/ingress/commit/%s)", tag, hash)
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: DEFAULT_CONFIG,
			Usage: "Config file (search path ., ~ and /etc)",
		},
		cli.BoolFlag{
			Name:  "dump, d",
			Usage: "Dump parsed config",
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

	app.OnUsageError = func(c *cli.Context, err error, isSubcommand bool) error {
		log.Configure(c.String("log"))
		log.Fatal(err)
		return err
	}

	app.Action = func(c *cli.Context) {
		log.Configure(c.String("log"))
		log.Printf("ingress v%s %s", tag, hash)

		conf, err := parseConfig(c)
		if err != nil {
			log.Fatal(err)
		}

		go checkVersion()

		connectors := wiring.NewConnectors(conf.Sources, conf.Targets)
		mappings := wiring.NewMappings(conf.Mappings, connectors)
		actions := wiring.NewActions(conf.Actions)
		wires := wiring.NewWiring(conf.Wires, connectors, mappings, actions)
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

	app.Run(os.Args)
}
