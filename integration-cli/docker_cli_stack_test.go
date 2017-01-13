package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"sort"

	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/integration-cli/checker"
	"github.com/go-check/check"
)

func (s *DockerSwarmSuite) TestStackRemove(c *check.C) {
	d := s.AddDaemon(c, true, true)

	stackArgs := append([]string{"stack", "remove", "UNKNOWN_STACK"})

	out, err := d.Cmd(stackArgs...)
	c.Assert(err, checker.IsNil)
	c.Assert(out, check.Equals, "Nothing found in stack: UNKNOWN_STACK\n")
}

func (s *DockerSwarmSuite) TestStackTasks(c *check.C) {
	d := s.AddDaemon(c, true, true)

	stackArgs := append([]string{"stack", "ps", "UNKNOWN_STACK"})

	out, err := d.Cmd(stackArgs...)
	c.Assert(err, checker.IsNil)
	c.Assert(out, check.Equals, "Nothing found in stack: UNKNOWN_STACK\n")
}

func (s *DockerSwarmSuite) TestStackServices(c *check.C) {
	d := s.AddDaemon(c, true, true)

	stackArgs := append([]string{"stack", "services", "UNKNOWN_STACK"})

	out, err := d.Cmd(stackArgs...)
	c.Assert(err, checker.IsNil)
	c.Assert(out, check.Equals, "Nothing found in stack: UNKNOWN_STACK\n")
}

func (s *DockerSwarmSuite) TestStackDeployComposeFile(c *check.C) {
	d := s.AddDaemon(c, true, true)

	testStackName := "testdeploy"
	stackArgs := []string{
		"stack", "deploy",
		"--compose-file", "fixtures/deploy/default.yaml",
		testStackName,
	}
	out, err := d.Cmd(stackArgs...)
	c.Assert(err, checker.IsNil, check.Commentf(out))

	out, err = d.Cmd([]string{"stack", "ls"}...)
	c.Assert(err, checker.IsNil)
	c.Assert(out, check.Equals, "NAME        SERVICES\n"+"testdeploy  2\n")

	out, err = d.Cmd([]string{"stack", "rm", testStackName}...)
	c.Assert(err, checker.IsNil)
	out, err = d.Cmd([]string{"stack", "ls"}...)
	c.Assert(err, checker.IsNil)
	c.Assert(out, check.Equals, "NAME  SERVICES\n")
}

func (s *DockerSwarmSuite) TestStackDeployWithSecretsTwice(c *check.C) {
	d := s.AddDaemon(c, true, true)

	testStackName := "testdeploy"
	stackArgs := []string{
		"stack", "deploy",
		"--compose-file", "fixtures/deploy/secrets.yaml",
		testStackName,
	}
	out, err := d.Cmd(stackArgs...)
	c.Assert(err, checker.IsNil, check.Commentf(out))

	out, err = d.Cmd("service", "inspect", "--format", "{{ json .Spec.TaskTemplate.ContainerSpec.Secrets }}", "testdeploy_web")
	c.Assert(err, checker.IsNil)

	var refs []swarm.SecretReference
	c.Assert(json.Unmarshal([]byte(out), &refs), checker.IsNil)
	c.Assert(refs, checker.HasLen, 2)

	sort.Sort(sortSecrets(refs))
	c.Assert(refs[0].SecretName, checker.Equals, "testdeploy_special")
	c.Assert(refs[0].File.Name, checker.Equals, "special")
	c.Assert(refs[1].SecretName, checker.Equals, "testdeploy_super")
	c.Assert(refs[1].File.Name, checker.Equals, "foo.txt")
	c.Assert(refs[1].File.Mode, checker.Equals, os.FileMode(0400))

	// Deploy again to ensure there are no errors when secret hasn't changed
	out, err = d.Cmd(stackArgs...)
	c.Assert(err, checker.IsNil, check.Commentf(out))
}

type sortSecrets []swarm.SecretReference

func (s sortSecrets) Len() int           { return len(s) }
func (s sortSecrets) Swap(i, j int)      { s[i], s[j] = s[j], s[i] }
func (s sortSecrets) Less(i, j int) bool { return s[i].SecretName < s[j].SecretName }

// testDAB is the DAB JSON used for testing.
// TODO: Use template/text and substitute "Image" with the result of
// `docker inspect --format '{{index .RepoDigests 0}}' busybox:latest`
const testDAB = `{
    "Version": "0.1",
    "Services": {
	"srv1": {
	    "Image": "busybox@sha256:e4f93f6ed15a0cdd342f5aae387886fba0ab98af0a102da6276eaf24d6e6ade0",
	    "Command": ["top"]
	},
	"srv2": {
	    "Image": "busybox@sha256:e4f93f6ed15a0cdd342f5aae387886fba0ab98af0a102da6276eaf24d6e6ade0",
	    "Command": ["tail"],
	    "Args": ["-f", "/dev/null"]
	}
    }
}`

func (s *DockerSwarmSuite) TestStackDeployWithDAB(c *check.C) {
	testRequires(c, ExperimentalDaemon)
	// setup
	testStackName := "test"
	testDABFileName := testStackName + ".dab"
	defer os.RemoveAll(testDABFileName)
	err := ioutil.WriteFile(testDABFileName, []byte(testDAB), 0444)
	c.Assert(err, checker.IsNil)
	d := s.AddDaemon(c, true, true)
	// deploy
	stackArgs := []string{
		"stack", "deploy",
		"--bundle-file", testDABFileName,
		testStackName,
	}
	out, err := d.Cmd(stackArgs...)
	c.Assert(err, checker.IsNil)
	c.Assert(out, checker.Contains, "Loading bundle from test.dab\n")
	c.Assert(out, checker.Contains, "Creating service test_srv1\n")
	c.Assert(out, checker.Contains, "Creating service test_srv2\n")
	// ls
	stackArgs = []string{"stack", "ls"}
	out, err = d.Cmd(stackArgs...)
	c.Assert(err, checker.IsNil)
	c.Assert(out, check.Equals, "NAME  SERVICES\n"+"test  2\n")
	// rm
	stackArgs = []string{"stack", "rm", testStackName}
	out, err = d.Cmd(stackArgs...)
	c.Assert(err, checker.IsNil)
	c.Assert(out, checker.Contains, "Removing service test_srv1\n")
	c.Assert(out, checker.Contains, "Removing service test_srv2\n")
	// ls (empty)
	stackArgs = []string{"stack", "ls"}
	out, err = d.Cmd(stackArgs...)
	c.Assert(err, checker.IsNil)
	c.Assert(out, check.Equals, "NAME  SERVICES\n")
}
