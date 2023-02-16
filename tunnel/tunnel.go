package tunnel

import (
	"errors"
	"os"

	"autossh/config"
)

const (
	autosshConfigPath = "AUTOSSH_CONFIG_PATH"
)

func Execute() (err error) {
	conf := os.Getenv(autosshConfigPath)
	if err = config.ReadInFile(conf); err != nil {
		return
	}
	sshConfMap := sshConfMap(config.SshConf())
	if len(sshConfMap) < 1 {
		return errors.New("empty ssh config")
	}
	tunnels := tunnels(config.Tunnels(), sshConfMap)
	if len(tunnels) < 1 {
		return errors.New("empty ssh tunnels")
	}
	return startup(tunnels)
}

type listenType int

const (
	listenUnknown listenType = iota
	listenOnLocal
	listenOnRemote
)

func getListenType(s string) listenType {
	switch s {
	case "local":
		return listenOnLocal
	case "remote":
		return listenOnRemote

	default:
		return listenUnknown
	}
}

type tunnel struct {
	listenOn  listenType
	listenAt  string
	forwardTo string
	sshConfig *sshConfig
}

func tunnels(conf []*config.Tunnel, sshConfMap map[string]*sshConfig) (tun []*tunnel) {
	l := len(conf)
	if l < 1 {
		return
	}

	tun = make([]*tunnel, 0, l)
	for _, c := range conf {
		listenOn := getListenType(c.ListenOn)
		if listenOn == listenUnknown {
			continue
		}
		sshConfig, ok := sshConfMap[c.SshAlias]
		if !ok || sshConfig == nil {
			continue
		}

		tunnel := &tunnel{
			listenOn:  listenOn,
			listenAt:  c.ListenAt,
			forwardTo: c.ForwardTo,
			sshConfig: sshConfig,
		}
		tun = append(tun, tunnel)
	}
	return tun
}
