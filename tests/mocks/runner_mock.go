package mocks

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type FakeRunner struct{}

func (r *FakeRunner) Run(command string, args ...string) ([]byte, error) {
	var cs []string
	// If the command is trying to get the cmdline call the TestHelperBootedFrom test
	// Maybe a switch statement would be better here??
	if command == "cat" && len(args) > 0 && args[0] == "/proc/cmdline" {
		cs = []string{"-test.run=TestHelperBootedFrom", "--"}
		cs = append(cs, args...)
	} else if command == "blkid" && len(args) == 2 && args[1] == "EXISTS" {
		cs = []string{"-test.run=TestHelperFindLabel", "--"}
		cs = append(cs, args...)
	} else {
		return make([]byte, 0), nil
	}
	cmd := exec.Command(os.Args[0], cs...)
	cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
	out, err := cmd.CombinedOutput()
	return out, err
}

type TestRunnerV2 struct {
	cmds        [][]string
	ReturnValue []byte
	SideEffect  func(command string, args ...string) ([]byte, error)
}

func NewTestRunnerV2() *TestRunnerV2 {
	return &TestRunnerV2{cmds: [][]string{}, ReturnValue: []byte{}, SideEffect: nil}
}

func (r *TestRunnerV2) Run(command string, args ...string) ([]byte, error) {
	r.cmds = append(r.cmds, append([]string{command}, args...))
	if len(r.ReturnValue) > 0 {
		return r.ReturnValue, nil
	} else if r.SideEffect != nil {
		return r.SideEffect(command, args...)
	}
	return []byte{}, nil
}

func (r *TestRunnerV2) ClearCmds() {
	r.cmds = [][]string{}
}

// It matches the commands list. Note HasPrefix is being used to evaluate the
// match, so expecting inital part of the command is enough to get a match.
// It facilitates testing commands with dynamic arguments (aka temporary files)
func (r TestRunnerV2) CmdsMatch(cmdList [][]string) error {
	if len(cmdList) != len(r.cmds) {
		return errors.New(fmt.Sprintf("Number of calls mismatch, expected %d calls but got %d", len(cmdList), len(r.cmds)))
	}
	for i, cmd := range cmdList {
		expect := strings.Join(cmd[:], " ")
		got := strings.Join(r.cmds[i][:], " ")
		if !strings.HasPrefix(got, expect) {
			return errors.New(fmt.Sprintf("Expected command: '%s.*' got: '%s'", expect, got))
		}
	}
	return nil
}