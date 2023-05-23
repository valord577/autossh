package config

import (
	"os"

	"autossh/config/internal"
)

const (
	autosshConfigPath = "AUTOSSH_CONFIG_PATH"
)

func ReadInFile() error {
	fp := os.Getenv(autosshConfigPath)
	return internal.ReadInFile(fp)
}
