package client

import (
	"net/url"
	"time"

	Cli "github.com/docker/docker/cli"
	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/pkg/parsers/filters"
	"github.com/docker/docker/pkg/timeutils"
)

// CmdEvents prints a live stream of real time events from the server.
//
// Usage: docker events [OPTIONS]
func (cli *DockerCli) CmdEvents(args ...string) error {
	cmd := Cli.Subcmd("events", nil, Cli.DockerCommands["events"].Description, true)
	since := cmd.String([]string{"-since"}, "", "Show all events created since timestamp")
	until := cmd.String([]string{"-until"}, "", "Stream events until this timestamp")
	flFilter := opts.NewListOpts(nil)
	cmd.Var(&flFilter, []string{"f", "-filter"}, "Filter output based on conditions provided")
	cmd.Require(flag.Exact, 0)

	cmd.ParseFlags(args, true)

	var (
		v               = url.Values{}
		eventFilterArgs = filters.NewArgs()
	)

	// Consolidate all filter flags, and sanity check them early.
	// They'll get process in the daemon/server.
	for _, f := range flFilter.GetAll() {
		var err error
		eventFilterArgs, err = filters.ParseFlag(f, eventFilterArgs)
		if err != nil {
			return err
		}
	}
	ref := time.Now()
	if *since != "" {
		ts, err := timeutils.GetTimestamp(*since, ref)
		if err != nil {
			return err
		}
		v.Set("since", ts)
	}
	if *until != "" {
		ts, err := timeutils.GetTimestamp(*until, ref)
		if err != nil {
			return err
		}
		v.Set("until", ts)
	}
	if eventFilterArgs.Len() > 0 {
		filterJSON, err := filters.ToParam(eventFilterArgs)
		if err != nil {
			return err
		}
		v.Set("filters", filterJSON)
	}
	sopts := &streamOpts{
		rawTerminal: true,
		out:         cli.out,
	}
	if _, err := cli.stream("GET", "/events?"+v.Encode(), sopts); err != nil {
		return err
	}
	return nil
}
