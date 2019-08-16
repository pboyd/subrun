package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/davecgh/go-spew/spew"
)

func main() {
	var configPath string
	flag.StringVar(&configPath, "config", "/etc/subrunner.yaml", "path to config file")
	flag.Parse()

	config, err := readConfigFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to read %s: %v\n", configPath, err)
		os.Exit(1)
	}
	spew.Dump(config)
}
