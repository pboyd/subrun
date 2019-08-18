package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"

	"github.com/pboyd/subrun/shell"
	"github.com/pboyd/subrun/trigger"
	"github.com/pboyd/subrun/trigger/pubsub"
)

var pubSubEmulatorOpts *pubsub.TriggerOpts

func init() {
	host := os.Getenv("PUBSUB_EMULATOR_HOST")
	if host == "" {
		return
	}

	project := os.Getenv("PUBSUB_PROJECT_ID")
	if project == "" {
		project = "fake-project"

		log.Printf("PUBSUB_PROJECT_ID not set, defaulting to %q", project)
	}

	opts, err := pubsub.PubSubEmulatorOpts(context.Background(), host, project)
	if err != nil {
		log.Fatalf("error connecting to pubsub emulator: %v", err)
	}

	pubSubEmulatorOpts = &opts
}

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "/etc/subrun.yaml", "path to config file")
	flag.Parse()

	config, err := readConfigFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read %s: %v\n", configPath, err)
		os.Exit(1)
	}

	err = config.Check()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", configPath, err)
		os.Exit(2)
	}

	for i, sub := range config.Subscriptions {
		id := sub.Name
		if id == "" {
			id = strconv.Itoa(i)
		}

		startTrigger(id, sub)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}

func startTrigger(id string, cfg ConfigSubscription) error {
	triggerCfg := cfg.Trigger.PubSub

	opts := pubsub.TriggerOpts{
		CredentialsFile: triggerCfg.CredentialsFile,
	}

	if pubSubEmulatorOpts != nil {
		opts = *pubSubEmulatorOpts
	}

	trigger, err := pubsub.NewTrigger(id, triggerCfg.Project, triggerCfg.Topic, opts)
	if err != nil {
		return err
	}

	go func() {
		for msg := range trigger.C {
			go runTasks(id, msg, cfg)
		}
	}()

	return nil
}

func runTasks(id string, msg trigger.Message, cfg ConfigSubscription) {
	for _, taskConfig := range cfg.Tasks {
		task := shell.Task{
			Cmd: taskConfig.Cmd,
			Dir: cfg.Dir,
		}

		if len(msg.Payload) > 0 {
			task.Stdin = bytes.NewReader(msg.Payload)
		}

		ctx := context.Background()
		cancel := func() {}
		if taskConfig.Timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, taskConfig.Timeout)
		}

		err := shell.Run(ctx, task)
		if err != nil {
			msg.Callback(false)
			log.Printf("%s: error executing %s: %v", id, taskConfig.Cmd, err)
		} else {
			msg.Callback(true)
		}

		cancel()
	}
}
