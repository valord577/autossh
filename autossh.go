package main

import (
	"os"

	"autossh/component"
	log "autossh/logger"
	"autossh/tunnel"
	"autossh/version"
)

const (
	EXIT_SUCCESS = 0
	EXIT_FAILURE = 1
)

var logger = &component.Zap{}

func main() {
	_ = component.Use(logger)
	defer func() {
		_ = component.Free(logger)
	}()
	log.Infof("%s", version.String())

	exitCode := EXIT_SUCCESS
	err := tunnel.Execute()
	if err != nil {
		log.Errorf("%s", err.Error())
		exitCode = EXIT_FAILURE
	}
	os.Exit(exitCode)
}
