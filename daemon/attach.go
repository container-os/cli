package daemon

import (
	"io"

	"github.com/docker/docker/pkg/stdcopy"
)

type ContainerAttachWithLogsConfig struct {
	InStream                       io.ReadCloser
	OutStream                      io.Writer
	UseStdin, UseStdout, UseStderr bool
	Logs, Stream                   bool
	Multiplex                      bool
}

func (daemon *Daemon) ContainerAttachWithLogs(name string, c *ContainerAttachWithLogsConfig) error {
	container, err := daemon.Get(name)
	if err != nil {
		return err
	}

	var errStream io.Writer

	if !container.Config.Tty && c.Multiplex {
		errStream = stdcopy.NewStdWriter(c.OutStream, stdcopy.Stderr)
		c.OutStream = stdcopy.NewStdWriter(c.OutStream, stdcopy.Stdout)
	} else {
		errStream = c.OutStream
	}

	var stdin io.ReadCloser
	var stdout, stderr io.Writer

	if c.UseStdin {
		stdin = c.InStream
	}
	if c.UseStdout {
		stdout = c.OutStream
	}
	if c.UseStderr {
		stderr = errStream
	}

	return container.AttachWithLogs(stdin, stdout, stderr, c.Logs, c.Stream)
}

type ContainerWsAttachWithLogsConfig struct {
	InStream             io.ReadCloser
	OutStream, ErrStream io.Writer
	Logs, Stream         bool
}

func (daemon *Daemon) ContainerWsAttachWithLogs(name string, c *ContainerWsAttachWithLogsConfig) error {
	container, err := daemon.Get(name)
	if err != nil {
		return err
	}

	return container.AttachWithLogs(c.InStream, c.OutStream, c.ErrStream, c.Logs, c.Stream)
}
