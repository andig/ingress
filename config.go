package ingress

import (
	"fmt"
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type Input struct {
	Name     string
	Type     string
	URL      string
	User     string
	Password string
}

type Output struct {
	Name string
	Type string
	Url  string
}

type Config struct {
	Input  []Input
	Output []Output
	Mapper []struct {
		Input struct {
			Name     string
			Protocol string
		}
		Output struct {
			Name     string
			Protocol string
		}
		Map []struct {
			Source string
			Target string
			Uuid   string
		}
	}
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

	d, err := yaml.Marshal(c)
	if err != nil {
		log.Fatalf("error: %v", err)
	}
	fmt.Printf("--- dump:\n%s\n\n", string(d))

	return c
}
