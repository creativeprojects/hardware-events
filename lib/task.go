package lib

import (
	"bytes"
	"io"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/creativeprojects/hardware-events/lib/simulation"
)

type Task struct {
	global        *Global
	Name          string
	Command       CommandRunner
	InputTemplate string
}

func NewTask(global *Global, name string, config cfg.Task, simulate bool) (*Task, error) {
	var command CommandRunner
	var timeout time.Duration
	var err error

	if config.Timeout != "" {
		timeout, err = time.ParseDuration(config.Timeout)
		if err != nil {
			return nil, err
		}
	}
	if simulate {
		command, err = simulation.NewCommand(config.Command, "")
		if err != nil {
			return nil, err
		}
	} else {
		command, err = NewCommand(config.Command, "", timeout) // an error can only be thrown by a wrong regexp
		if err != nil {
			return nil, err
		}
	}
	return &Task{
		global:        global,
		Name:          name,
		Command:       command,
		InputTemplate: config.Stdin.Template,
	}, nil
}

func (t *Task) Execute() error {
	var stdin io.Reader
	if t.InputTemplate != "" {
		input := &bytes.Buffer{}
		err := t.global.templ.ExecuteTemplate(input, t.global.Templates[t.InputTemplate].ID, t.global)
		if err != nil {
			return err
		}
		clog.Tracef("template %s:\n%s", t.InputTemplate, input.String())
		stdin = input
	}
	_, err := t.Command.Run(stdin, nil)
	return err
}

type Timer struct {
	task  *Task
	every time.Duration
}
