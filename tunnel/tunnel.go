package tunnel

import (
	"autossh/config"
)

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

func mapListenType(tp listenType) string {
	switch tp {
	case listenUnknown:
		return "unknown"
	case listenOnLocal:
		return "local"
	case listenOnRemote:
		return "remote"
	default:
		return ""
	}
}

type tunnel struct {
	service   string
	listenOn  listenType
	listenAt  string
	forwardTo string
	sshConfig *sshConfig
}

func Tunnels(conf []*config.Tunnel, sshConfMap map[string]*sshConfig) (tun []*tunnel) {
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
			service:   c.Service,
			listenOn:  listenOn,
			listenAt:  c.ListenAt,
			forwardTo: c.ForwardTo,
			sshConfig: sshConfig,
		}
		tun = append(tun, tunnel)
	}
	return tun
}
