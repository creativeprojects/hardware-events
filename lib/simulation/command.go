package simulation

import (
	"io"
	"os"

	"github.com/creativeprojects/clog"
)

type Command struct {
	CommandLine string
}

func NewCommand(commandLine, outputRegexp string) (*Command, error) {
	return &Command{
		CommandLine: commandLine,
	}, nil
}

func (c *Command) Run(stdin io.Reader, expand func(string) string) (string, error) {
	command := os.Expand(c.CommandLine, expand)
	clog.Debugf("command: %s", command)
	return "", nil
}
