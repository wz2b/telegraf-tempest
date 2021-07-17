package main

import (
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"os"
	"strings"
	"telegraf-tempest/internal/tclogger"
)

func setupLogging(config TempestAgentConfig) {
	logout := strings.ToLower(config.LogFile)
	if logout == "stderr" || logout == "" {
		log.SetOutput(os.Stderr)
	} else if logout == "stdout" {
		telegrafLogger := tclogger.Create()
		telegrafLogger.Start(os.Stdout)
	} else {
		jack := &lumberjack.Logger{
			Filename:   config.LogFile,
			MaxSize:    10, // megabytes
			MaxBackups: 10,
			MaxAge:     30,   //days
			Compress:   true, // disabled by default
		}
		log.SetOutput(jack)

	}
	log.Println("Logging started")
}
