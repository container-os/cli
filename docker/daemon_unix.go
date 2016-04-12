// +build daemon,!windows

package main

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"syscall"

	"github.com/Sirupsen/logrus"
	apiserver "github.com/docker/docker/api/server"
	"github.com/docker/docker/daemon"
	"github.com/docker/docker/libcontainerd"
	"github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/system"
	"github.com/docker/libnetwork/portallocator"
)

const defaultDaemonConfigFile = "/etc/docker/daemon.json"

func setPlatformServerConfig(serverConfig *apiserver.Config, daemonCfg *daemon.Config) *apiserver.Config {
	serverConfig.EnableCors = daemonCfg.EnableCors
	serverConfig.CorsHeaders = daemonCfg.CorsHeaders

	return serverConfig
}

// currentUserIsOwner checks whether the current user is the owner of the given
// file.
func currentUserIsOwner(f string) bool {
	if fileInfo, err := system.Stat(f); err == nil && fileInfo != nil {
		if int(fileInfo.UID()) == os.Getuid() {
			return true
		}
	}
	return false
}

// setDefaultUmask sets the umask to 0022 to avoid problems
// caused by custom umask
func setDefaultUmask() error {
	desiredUmask := 0022
	syscall.Umask(desiredUmask)
	if umask := syscall.Umask(desiredUmask); umask != desiredUmask {
		return fmt.Errorf("failed to set umask: expected %#o, got %#o", desiredUmask, umask)
	}

	return nil
}

func getDaemonConfDir() string {
	return "/etc/docker"
}

// setupConfigReloadTrap configures the USR2 signal to reload the configuration.
func setupConfigReloadTrap(configFile string, flags *mflag.FlagSet, reload func(*daemon.Config)) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGHUP)
	go func() {
		for range c {
			if err := daemon.ReloadConfiguration(configFile, flags, reload); err != nil {
				logrus.Error(err)
			}
		}
	}()
}

func (cli *DaemonCli) getPlatformRemoteOptions() []libcontainerd.RemoteOption {
	opts := []libcontainerd.RemoteOption{
		libcontainerd.WithDebugLog(cli.Config.Debug),
	}
	if cli.Config.ContainerdAddr != "" {
		opts = append(opts, libcontainerd.WithRemoteAddr(cli.Config.ContainerdAddr))
	} else {
		opts = append(opts, libcontainerd.WithStartDaemon(true))
	}
	if daemon.UsingSystemd(cli.Config) {
		args := []string{"--systemd-cgroup=true"}
		opts = append(opts, libcontainerd.WithRuntimeArgs(args))
	}
	return opts
}

// getLibcontainerdRoot gets the root directory for libcontainerd/containerd to
// store their state.
func (cli *DaemonCli) getLibcontainerdRoot() string {
	return filepath.Join(cli.Config.ExecRoot, "libcontainerd")
}

// allocateDaemonPort ensures that there are no containers
// that try to use any port allocated for the docker server.
func allocateDaemonPort(addr string) error {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return err
	}

	intPort, err := strconv.Atoi(port)
	if err != nil {
		return err
	}

	var hostIPs []net.IP
	if parsedIP := net.ParseIP(host); parsedIP != nil {
		hostIPs = append(hostIPs, parsedIP)
	} else if hostIPs, err = net.LookupIP(host); err != nil {
		return fmt.Errorf("failed to lookup %s address in host specification", host)
	}

	pa := portallocator.Get()
	for _, hostIP := range hostIPs {
		if _, err := pa.RequestPort(hostIP, "tcp", intPort); err != nil {
			return fmt.Errorf("failed to allocate daemon listening port %d (err: %v)", intPort, err)
		}
	}
	return nil
}
