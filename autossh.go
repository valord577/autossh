package main

import (
	"os"
	"os/signal"
	"syscall"

	"autossh/component"
	"autossh/config"
	log "autossh/logger"
	"autossh/version"
)

const (
	EXIT_SUCCESS = 0
	EXIT_FAILURE = 1
)

var logger = &component.Zap{}

func main() {
	exitCode := EXIT_SUCCESS
	defer os.Exit(exitCode)

	_ = component.Use(logger)
	defer func() {
		_ = component.Free(logger)
	}()
	log.Infof("%s", version.String())

	if err := config.ReadInFile(); err != nil {
		exitCode = EXIT_FAILURE
		log.Errorf("%s", err.Error())
		return
	}

	tun := &component.Tun{}
	if err := component.Use(tun); err != nil {
		exitCode = EXIT_FAILURE
		log.Errorf("%s", err.Error())
		return
	}
	defer func() {
		_ = component.Free(tun)
	}()

	// block and listen for signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	s := <-sig
	log.Infof("recv signal: %d", s)
}
