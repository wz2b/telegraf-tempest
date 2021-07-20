package main

import (
	"flag"
)

type TempestAgentConfig struct {
	Args []string
}

func parseCommandLine() (TempestAgentConfig, error) {
	var config TempestAgentConfig

	flag.Parse()

	// store remaining arguments in case somebody wants to use them
	config.Args = flag.Args()

	return config, nil
}
