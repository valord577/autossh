package main

import (
	"os"
	"os/signal"
	"syscall"

	"autossh/config"
	"autossh/logs"
	"autossh/tunnel2"
	"autossh/version"
)

const (
	EXIT_SUCCESS = 0
	EXIT_FAILURE = 1
)

func main() {
	exitCode := EXIT_SUCCESS
	defer func() {
		os.Exit(exitCode)
	}()
	logs.Infof("%s", version.String())

	if err := config.ReadInFile(); err != nil {
		exitCode = EXIT_FAILURE
		logs.Errorf("%s", err.Error())
		return
	}

	if err := tunnel2.Startup(); err != nil {
		exitCode = EXIT_FAILURE
		logs.Errorf("%s", err.Error())
		return
	}
	defer tunnel2.Shutdown()

	// block and listen for signals
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
	s := <-sig
	logs.Infof("recv signal: %s", s.String())
}
