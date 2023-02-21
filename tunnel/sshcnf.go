package tunnel

import (
	"os"
	"path/filepath"
	"time"

	"autossh/config"
	log "autossh/logger"

	"golang.org/x/crypto/ssh"
)

func sshParseKeys(keys []string) (s []ssh.Signer) {
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
				log.Errorf("get user home dir, err: %s", err.Error())
				continue
			}
			path = filepath.Join(home, ".ssh", path)
		}

		bs, err := os.ReadFile(path)
		if err != nil {
			log.Errorf("read ssh key, path: %s, err: %s", path, err.Error())
			continue
		}
		signer, err := ssh.ParsePrivateKey(bs)
		if err != nil {
			log.Errorf("parse ssh key, err: %s", err.Error())
			continue
		}
		s = append(s, signer)
	}
	return
}

func sshBannerDisplay(banner string) (e error) {
	return
}

type sshConfig struct {
	alias   string
	address string
	config  *ssh.ClientConfig
}

func SshConfMap(conf []*config.SshConfig) map[string]*sshConfig {
	m := make(map[string]*sshConfig)

	for _, c := range conf {
		if c == nil || c.Auth == nil {
			continue
		}
		keys := sshParseKeys(c.Auth.Keys)
		if len(keys) < 1 && len(c.Auth.Pass) < 1 {
			continue
		}

		sshConfig := &sshConfig{
			alias:   c.Alias,
			address: c.Address,
			config: &ssh.ClientConfig{
				User: c.User,
				Auth: []ssh.AuthMethod{
					ssh.Password(c.Auth.Pass),
					ssh.PublicKeys(keys...),
				},
				Timeout:         10 * time.Second,
				BannerCallback:  sshBannerDisplay,
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			},
		}
		m[c.Alias] = sshConfig
	}
	return m
}
