package lib

import (
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/creativeprojects/clog"
)

type CommandRunner interface {
	Run(stdin io.Reader, expand func(string) string) (string, error)
}

type Command struct {
	CommandLine  string
	OutputRegexp *regexp.Regexp
	timeout      time.Duration
}

func NewCommand(commandLine, outputRegexp string, timeout time.Duration) (*Command, error) {
	var outputPattern *regexp.Regexp
	var err error
	if outputRegexp != "" {
		outputPattern, err = regexp.Compile(outputRegexp)
		if err != nil {
			return nil, err
		}
	}
	return &Command{
		CommandLine:  commandLine,
		OutputRegexp: outputPattern,
		timeout:      timeout,
	}, nil
}

func (c *Command) Run(stdin io.Reader, expand func(string) string) (string, error) {
	command := os.Expand(c.CommandLine, expand)
	clog.Debugf("command: %s", command)
	output, err := c.runCommand(command, stdin)
	if err != nil {
		clog.Errorf("error running command `%s`: %v", command, err)
	}
	if c.OutputRegexp != nil {
		found := c.OutputRegexp.FindStringSubmatch(output)
		if len(found) > 1 {
			output = found[1]
		}
	}
	clog.Trace(output)
	return output, err
}

// runCommand cancels the context as quick as possible
func (c *Command) runCommand(command string, stdin io.Reader) (string, error) {
	buffer := &bytes.Buffer{}
	var cmd *exec.Cmd
	if c.timeout == 0 {
		cmd = exec.Command("/bin/sh", "-c", command)
	} else {
		ctx, cancel := context.WithTimeout(context.Background(), c.timeout)
		cmd = exec.CommandContext(ctx, "/bin/sh", "-c", command)
		defer cancel()
	}
	cmd.Stdin = stdin
	cmd.Stdout = buffer
	err := cmd.Run()
	return buffer.String(), err
}
