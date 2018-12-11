package config

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type Basics struct {
	Name string
	Type string
	URL  string
}

type Credentials struct {
	User     string
	Password string
}

type Input struct {
	Basics      `yaml:",inline"`
	Credentials `yaml:",inline"`
	Topic       string
}

type Output struct {
	Basics      `yaml:",inline"`
	Credentials `yaml:",inline"`
	Topic       string
}

type Wiring struct {
	Inputs  []string `yaml:"sources"`
	Outputs []string `yaml:"targets"`
	Mapping []string `yaml:"mappings"`
}

type MapEntry struct {
	Source string
	Target string
	Uuid   string
}

type Mapping struct {
	Name string
	Map  []MapEntry `yaml:"entries"`
}

type Config struct {
	Input   []Input   `yaml:"sources"`
	Output  []Output  `yaml:"targets"`
	Wiring  []Wiring  `yaml:"wires"`
	Mapping []Mapping `yaml:"mappings"`
}

// Load loads and parses configuration from file
func (c *Config) Load(file string) *Config {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	return c
}

// Dump dumps parsed config to console
func (c *Config) Dump() {
	d, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Println(string(d))
}
