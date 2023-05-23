package config

import (
	"os"
	"path/filepath"
	"time"

	"golang.org/x/crypto/ssh"

	"autossh/config/internal"
)

type tunListenType uint8

const (
	tunListenUnknown tunListenType = iota
	TunListenOnLocal
	TunListenOnRemote
)

func getListenType(s string) tunListenType {
	switch s {
	case "local":
		return TunListenOnLocal
	case "remote":
		return TunListenOnRemote

	default:
		return tunListenUnknown
	}
}

func TunListenTypeString(tp tunListenType) string {
	switch tp {
	case tunListenUnknown:
		return "unknown"
	case TunListenOnLocal:
		return "local"
	case TunListenOnRemote:
		return "remote"
	default:
		return ""
	}
}

type sshConfig struct {
	Alias   string
	Address string
	Config  *ssh.ClientConfig
}

type Tunnel struct {
	Service   string
	ListenOn  tunListenType
	ListenAt  string
	ForwardTo string
	SshConfig *sshConfig
}

func Tunnels() (tuns []*Tunnel) {
	tunConf := internal.Conf.Tunnel
	sshConf := internal.Conf.SshConfig

	tunConfSize := len(tunConf)
	sshConfSize := len(sshConf)
	if tunConfSize < 1 || sshConfSize < 1 {
		return
	}

	sshAuthKey := func(keys []string) (s []ssh.Signer) {
		l := len(keys)
		if l < 1 {
			return
		}

		s = make([]ssh.Signer, 0, l)
		for _, k := range keys {
			path := filepath.Clean(k)
			isAbs := filepath.IsAbs(path)
			if !isAbs {
				home, err := os.UserHomeDir()
				if err != nil {
					continue
				}
				path = filepath.Join(home, ".ssh", path)
			}

			bs, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			signer, err := ssh.ParsePrivateKey(bs)
			if err != nil {
				continue
			}
			s = append(s, signer)
		}
		return
	}
	sshConfMap := func() map[string]*sshConfig {
		m := make(map[string]*sshConfig, sshConfSize)

		for _, c := range sshConf {
			if c == nil || c.Auth == nil {
				continue
			}
			pass := c.Auth.Pass
			keys := sshAuthKey(c.Auth.Keys)
			if len(keys) < 1 && len(pass) < 1 {
				continue
			}

			sshConfig := &sshConfig{
				Alias:   c.Alias,
				Address: c.Address,
				Config: &ssh.ClientConfig{
					User: c.User,
					Auth: []ssh.AuthMethod{
						ssh.Password(pass),
						ssh.PublicKeys(keys...),
					},
					Timeout:         10 * time.Second,
					BannerCallback:  func(string) (e error) { return },
					HostKeyCallback: ssh.InsecureIgnoreHostKey(),
				},
			}
			m[c.Alias] = sshConfig
		}
		return m
	}

	sshConfigMap := sshConfMap()
	tuns = make([]*Tunnel, 0, tunConfSize)
	for _, c := range tunConf {
		listenOn := getListenType(c.ListenOn)
		if listenOn == tunListenUnknown {
			continue
		}
		sshConfig, ok := sshConfigMap[c.SshAlias]
		if !ok || sshConfig == nil {
			continue
		}

		tunnel := &Tunnel{
			Service:   c.Service,
			ListenOn:  listenOn,
			ListenAt:  c.ListenAt,
			ForwardTo: c.ForwardTo,
			SshConfig: sshConfig,
		}
		tuns = append(tuns, tunnel)
	}
	return
}
