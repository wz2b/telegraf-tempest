package main

import (
	"flag"
)

type TempestAgentConfig struct {
	LogFile  string // may be file path, "stderr", or "stdout"
	LogLevel string
	Args     []string
}

func parseCommandLine() (TempestAgentConfig, error) {
	var config TempestAgentConfig

	flag.StringVar(&config.LogFile, "log", "stderr", "log destination, can be \"stdout\", \"stderr\", or file path")
	flag.StringVar(&config.LogLevel, "log-level", "error", "log destination, can be \"stdout\", \"stderr\", or file path")

	flag.Parse()

	// store remaining arguments in case somebody wants to use them
	config.Args = flag.Args()

	return config, nil
}
