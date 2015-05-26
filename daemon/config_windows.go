package daemon

import (
	"os"
)

var (
	defaultPidFile = os.Getenv("programdata") + string(os.PathSeparator) + "docker.pid"
	defaultGraph   = os.Getenv("programdata") + string(os.PathSeparator) + "docker"
	defaultExec    = "windows"
)

// Config defines the configuration of a docker daemon.
// These are the configuration settings that you pass
// to the docker daemon when you launch it with say: `docker -d -e windows`
type Config struct {
	CommonConfig

	// Fields below here are platform specific. (There are none presently
	// for the Windows daemon.)
}

// InstallFlags adds command-line options to the top-level flag parser for
// the current process.
// Subsequent calls to `flag.Parse` will populate config with values parsed
// from the command-line.
func (config *Config) InstallFlags() {
	// First handle install flags which are consistent cross-platform
	config.InstallCommonFlags()

	// Then platform-specific install flags. There are none presently on Windows

}
