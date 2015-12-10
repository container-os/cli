// +build daemon

package main

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/docker/docker/cli"
	"github.com/docker/docker/daemon"
	"github.com/docker/docker/opts"
	"github.com/docker/docker/pkg/mflag"
	"github.com/docker/go-connections/tlsconfig"
)

func TestLoadDaemonCliConfigWithoutOverriding(t *testing.T) {
	c := &daemon.Config{}
	common := &cli.CommonFlags{
		Debug: true,
	}

	flags := mflag.NewFlagSet("test", mflag.ContinueOnError)
	loadedConfig, err := loadDaemonCliConfig(c, flags, common, "/tmp/fooobarbaz")
	if err != nil {
		t.Fatal(err)
	}
	if loadedConfig == nil {
		t.Fatalf("expected configuration %v, got nil", c)
	}
	if !loadedConfig.Debug {
		t.Fatalf("expected debug to be copied from the common flags, got false")
	}
}

func TestLoadDaemonCliConfigWithTLS(t *testing.T) {
	c := &daemon.Config{}
	common := &cli.CommonFlags{
		TLS: true,
		TLSOptions: &tlsconfig.Options{
			CAFile: "/tmp/ca.pem",
		},
	}

	flags := mflag.NewFlagSet("test", mflag.ContinueOnError)
	loadedConfig, err := loadDaemonCliConfig(c, flags, common, "/tmp/fooobarbaz")
	if err != nil {
		t.Fatal(err)
	}
	if loadedConfig == nil {
		t.Fatalf("expected configuration %v, got nil", c)
	}
	if loadedConfig.TLSOptions.CAFile != "/tmp/ca.pem" {
		t.Fatalf("expected /tmp/ca.pem, got %s: %q", loadedConfig.TLSOptions.CAFile, loadedConfig)
	}
}

func TestLoadDaemonCliConfigWithConflicts(t *testing.T) {
	c := &daemon.Config{}
	common := &cli.CommonFlags{}
	f, err := ioutil.TempFile("", "docker-config-")
	if err != nil {
		t.Fatal(err)
	}

	configFile := f.Name()
	f.Write([]byte(`{"labels": ["l3=foo"]}`))
	f.Close()

	var labels []string

	flags := mflag.NewFlagSet("test", mflag.ContinueOnError)
	flags.String([]string{daemonConfigFileFlag}, "", "")
	flags.Var(opts.NewNamedListOptsRef("labels", &labels, opts.ValidateLabel), []string{"-label"}, "")

	flags.Set(daemonConfigFileFlag, configFile)
	if err := flags.Set("-label", "l1=bar"); err != nil {
		t.Fatal(err)
	}
	if err := flags.Set("-label", "l2=baz"); err != nil {
		t.Fatal(err)
	}

	_, err = loadDaemonCliConfig(c, flags, common, configFile)
	if err == nil {
		t.Fatalf("expected configuration error, got nil")
	}
	if !strings.Contains(err.Error(), "labels") {
		t.Fatalf("expected labels conflict, got %v", err)
	}
}
