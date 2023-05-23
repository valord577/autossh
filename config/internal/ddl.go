package internal

import (
	"encoding/json"
	"os"
)

var Conf configuration

func ReadInFile(file string) error {
	bs, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(ignoreComments(bs), &Conf)
}

type configuration struct {
	Tunnel    []*tunnel    `json:"tunnel"`
	SshConfig []*sshConfig `json:"sshConfig"`
}

type tunnel struct {
	Service   string `json:"service"`
	ListenOn  string `json:"listenOn"`
	ListenAt  string `json:"listenAt"`
	SshAlias  string `json:"sshAlias"`
	ForwardTo string `json:"forwardTo"`
}

type sshConfig struct {
	Alias   string         `json:"alias"`
	Address string         `json:"address"`
	User    string         `json:"user"`
	Auth    *sshConfigAuth `json:"auth"`
}

type sshConfigAuth struct {
	Pass string   `json:"pass"`
	Keys []string `json:"keys"`
}
