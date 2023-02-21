package component

import (
	"errors"

	"autossh/config"
	"autossh/tunnel"
)

type Tun struct {
	service *tunnel.TunServ
}

func (t *Tun) init() error {
	sshConfMap := tunnel.SshConfMap(config.SshConf())
	if len(sshConfMap) < 1 {
		return errors.New("empty ssh config")
	}
	tunnels := tunnel.Tunnels(config.Tunnels(), sshConfMap)
	if len(tunnels) < 1 {
		return errors.New("empty ssh tunnels")
	}

	t.service = &tunnel.TunServ{}
	return t.service.Startup(tunnels)
}

func (t *Tun) free() error {
	t.service.Shutdown()
	return nil
}
