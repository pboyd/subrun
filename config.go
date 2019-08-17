package main

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Subscriptions []ConfigSubscription `yaml:"subscriptions"`
}

func (c *Config) Check() error {
	if len(c.Subscriptions) == 0 {
		return errors.New("no subscriptions")
	}

	for i, sub := range c.Subscriptions {
		name := sub.Name
		if name == "" {
			name = strconv.Itoa(i)
		}

		if sub.Trigger == nil {
			return fmt.Errorf("subscription %q: no trigger", name)
		}

		if sub.Trigger.PubSub == nil {
			return fmt.Errorf("subscription %q: empty trigger", name)
		}

		if sub.Trigger.PubSub.Topic == "" {
			return fmt.Errorf("subscription %q: pubsub trigger is missing topic", name)
		}

		if sub.Trigger.PubSub.Project == "" {
			return fmt.Errorf("subscription %q: pubsub trigger is missing project", name)
		}

		if len(sub.Tasks) == 0 {
			return fmt.Errorf("subscription %q: no tasks", name)
		}

		for j, task := range sub.Tasks {
			if task.Cmd == "" {
				return fmt.Errorf("subscription %q: task %d: no command", name, j)
			}
		}
	}

	return nil
}

type ConfigSubscription struct {
	Name    string         `yaml:"name"`
	Dir     string         `yaml:"dir"`
	Trigger *ConfigTrigger `yaml:"trigger"`
	Tasks   []ConfigTask   `yaml:"tasks"`
}

type ConfigTrigger struct {
	PubSub *PubSubTrigger `yaml:"pubsub"`
}

type PubSubTrigger struct {
	Project         string `yaml:"project"`
	Topic           string `yaml:"topic"`
	Endpoint        string `yaml:"endpoint"`
	CredentialsFile string `yaml:"credentialsFile"`
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
