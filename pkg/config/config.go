package config

import (
	"fmt"

	"github.com/andig/ingress/pkg/log"
	"github.com/mitchellh/mapstructure"

	yaml "gopkg.in/yaml.v2"
)

type Generic map[string]interface{}

type Entity struct {
	Name string `yaml:"name"`
	Type string `yaml:"type"`
}

type Credentials struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
}

type Source struct {
	Entity      `yaml:",squash"`
	Credentials `yaml:",squash"`
}

type Target struct {
	Entity      `yaml:",squash"`
	Credentials `yaml:",squash"`
}

type Wire struct {
	Sources  []string `yaml:"sources"`
	Targets  []string `yaml:"targets"`
	Mappings []string `yaml:"mappings"`
	Actions  []string `yaml:"actions"`
}

type MapEntry struct {
	From string
	To   string
}

type Mapping struct {
	Name    string     `yaml:"name"`
	Entries []MapEntry `yaml:"entries"`
}

type Action struct {
	Entity `yaml:",squash"`
}

type Config struct {
	Sources  []Generic `yaml:"sources"`
	Targets  []Generic `yaml:"targets"`
	Wires    []Wire    `yaml:"wires"`
	Mappings []Mapping `yaml:"mappings"`
	Actions  []Generic `yaml:"actions"`
}

// Dump dumps parsed config to console
func (c *Config) Dump() {
	fmt.Println("Parsed configuration")
	fmt.Println("---")

	d, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalf("error (%v)", err)
	}
	fmt.Println(string(d))
}

// default mapstructure decoder configuration for yaml config
func defaultDecoderConfig(target interface{}) *mapstructure.DecoderConfig {
	return &mapstructure.DecoderConfig{
		TagName:          "yaml",
		DecodeHook:       mapstructure.StringToTimeDurationHookFunc(),
		WeaklyTypedInput: true,
		Result:           &target,
	}
}

// decode creates decoder from config and invokes it
func decode(conf map[string]interface{}, dc *mapstructure.DecoderConfig) error {
	decoder, err := mapstructure.NewDecoder(dc)
	if err != nil {
		log.Fatal(err)
	}
	return decoder.Decode(conf)
}

// PartialDecode converts a configuration map into the desired target type.
// PartialDecode does not error on unused keys
func PartialDecode(conf map[string]interface{}, target interface{}) error {
	dc := defaultDecoderConfig(target)
	return decode(conf, dc)
}

// Decode converts a configuration map into the desired target type
// Decode errors on unused keys
func Decode(conf map[string]interface{}, target interface{}) error {
	dc := defaultDecoderConfig(target)
	dc.ErrorUnused = true
	return decode(conf, dc)
}
