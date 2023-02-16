package config

import (
	"encoding/json"
	"os"
)

func Tunnels() []*Tunnel {
	return c.Tunnel
}

func SshConf() []*SshConfig {
	return c.SshConfig
}

var c configuration

func ReadInFile(file string) error {
	bs, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(ignoreComments(bs), &c)
}
