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

const (
	autosshConfigPath = "AUTOSSH_CONFIG_PATH"
)

func ReadInFile() error {
	fp := os.Getenv(autosshConfigPath)
	return readInFile(fp)
}

func readInFile(file string) error {
	bs, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	return json.Unmarshal(ignoreComments(bs), &c)
}
