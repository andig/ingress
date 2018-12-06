package config

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type Credentials struct {
	User     string
	Password string
}

type Input struct {
	Name string
	Type string
	URL  string
	Credentials
	Topic string
}

type Output struct {
	Name string
	Type string
	URL  string
	Credentials
	Topic string
}

type Wiring struct {
	Inputs  []string `yaml:"input"`
	Outputs []string `yaml:"output"`
	Mapping []string `yaml:"mapping"`
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
	Input   []Input
	Output  []Output
	Wiring  []Wiring  `yaml:"wiring"`
	Mapping []Mapping `yaml:"mapping"`
}

func (c *Config) LoadConfig(file string) *Config {
	yamlFile, err := ioutil.ReadFile(file)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		log.Fatalf("Unmarshal: %v", err)
	}

	c.DumpConfig()
	return c
}

func (c *Config) DumpConfig() {
	d, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- dump:\n%s\n\n", string(d))
}
