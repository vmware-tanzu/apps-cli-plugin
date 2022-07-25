//go:build integration
// +build integration

/*
Copyright 2022 VMware, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package suite_test

import (
	"os"
	"os/exec"
	"strings"
)

type CommandLine struct {
	cmd          string
	args         []string
	surveyAnswer string
	out          string
	err          string
	exec         bool
	env          []string
}

func NewCommandLine(cmd string, args ...string) *CommandLine {
	return &CommandLine{
		cmd:  cmd,
		args: args,
	}
}

func NewTanzuAppsCommandLine(args ...string) *CommandLine {
	a := append([]string{"apps"}, args...)
	return &CommandLine{
		cmd:  "tanzu",
		args: a,
	}
}

func (c CommandLine) GetOutput() string {
	return c.out
}

func (c CommandLine) GetError() string {
	return c.err
}

func (c CommandLine) String() string {
	cmdArgs := []string{c.cmd}
	cmdArgs = append(cmdArgs, c.args...)
	return strings.Join(cmdArgs, " ")
}

func (c CommandLine) IsExec() bool {
	return c.exec
}

func (c *CommandLine) SurveyAnswer(answer string) {
	c.surveyAnswer = answer
}

func (c *CommandLine) Exec() error {
	// TODO: Support surveyAnswer
	cmd := exec.Command(c.cmd, c.args...)
	cmd.Env = append(os.Environ(), c.env...)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		c.err = err.Error()
		return err
	}

	if c.surveyAnswer != "" {
		go func() {
			stdin.Write([]byte(c.surveyAnswer + "\n"))
		}()
	}
	out, err := cmd.CombinedOutput()
	c.out = string(out)
	if err != nil {
		c.err = err.Error()
		return err
	}

	return nil
}
