package version

import (
	"os"
	"path/filepath"
	"runtime"
)

var (
	version  = "dev"
	datetime = "-"
	flavor   = "default"
)

func String() string {
	return filepath.Base(os.Args[0]) + " " + version + "." + flavor + " " + datetime +
		" " + runtime.Version() + " " + runtime.GOOS + "/" + runtime.GOARCH
}
