package component

import (
	log "autossh/logger"
)

type Zap struct{}

func (*Zap) init() error {
	log.Init()
	return nil
}

func (*Zap) free() error {
	return log.Sync()
}
