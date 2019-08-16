package main

import (
	"io"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Subscriptions []ConfigSubscription `yaml:"subscriptions"`
}

type ConfigSubscription struct {
	Topic string       `yaml:"topic"`
	Path  string       `yaml:"path"`
	Tasks []ConfigTask `yaml:"tasks"`
}

type ConfigTask struct {
	Cmd     string        `yaml:"cmd"`
	Timeout time.Duration `yaml:"timeout"`
}

func readConfigFile(path string) (*Config, error) {
	fh, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fh.Close()

	return readConfig(fh)
}

func readConfig(r io.Reader) (*Config, error) {
	decoder := yaml.NewDecoder(r)

	var cfg Config
	err := decoder.Decode(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
