// +build linux freebsd

package main

import (
	"testing"

	"github.com/docker/docker/daemon/config"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
)

func TestDaemonParseShmSize(t *testing.T) {
	flags := pflag.NewFlagSet("test", pflag.ContinueOnError)

	conf := &config.Config{}
	installConfigFlags(conf, flags)
	// By default `--default-shm-size=64M`
	assert.Equal(t, int64(64*1024*1024), conf.ShmSize.Value())
	assert.NoError(t, flags.Set("default-shm-size", "128M"))
	assert.Equal(t, int64(128*1024*1024), conf.ShmSize.Value())
}
