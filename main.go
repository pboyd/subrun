package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
)

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
	spew.Dump(config)
}
