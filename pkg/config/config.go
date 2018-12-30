package config

import (
	"fmt"
	"io/ioutil"

	. "github.com/andig/ingress/pkg/log"

	"gopkg.in/yaml.v2"
)

type Source struct {
	Name     string
	Type     string
	URL      string
	User     string // Credentials
	Password string
	Topic    string
}

type Target struct {
	Name     string
	Type     string
	URL      string
	User     string // Credentials
	Password string
	Topic    string
	Method   string
	Headers  map[string]string `yaml:"headers,omitempty"`
	Payload  string
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
	Name   string `yaml:"name"`
	Type   string `yaml:"type"`
	Mode   string `yaml:"mode"`
	Period string `yaml:"period"`
}

type Config struct {
	Sources  []Source  `yaml:"sources"`
	Targets  []Target  `yaml:"targets"`
	Wires    []Wire    `yaml:"wires"`
	Mappings []Mapping `yaml:"mappings"`
	Actions  []Action  `yaml:"actions"`
}

// Load loads and parses configuration from file
func (c *Config) Load(file string) *Config {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		Log().Fatalf("cannot read config file %s (%v)", file, err)
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		Log().Fatalf("cannot parse config file (%v)", err)
	}

	return c
}

// Dump dumps parsed config to console
func (c *Config) Dump() {
	fmt.Println("Parsed configuration")
	fmt.Println("---")

	d, err := yaml.Marshal(c)
	if err != nil {
		Log().Fatalf("error (%v)", err)
	}
	fmt.Println(string(d))
}
