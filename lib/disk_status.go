package lib

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/creativeprojects/hardware-events/lib/enum"
)

type DiskStatuser interface {
	Get(expandEnv func(string) string) enum.DiskStatus
	Standby(expandEnv func(string) string) error
}

type DiskStatus struct {
	checkCommand   CommandRunner
	Name           string
	DiskActive     string
	DiskStandby    string
	DiskSleeping   string
	File           string
	standbyCommand CommandRunner
	mutex          sync.Mutex
}

func NewDiskStatus(name string, config cfg.DiskPowerStatus) (*DiskStatus, error) {
	var timeout time.Duration
	var checkCommand, standbyCommand *Command
	var err error

	if config.Timeout != "" {
		timeout, err = time.ParseDuration(config.Timeout)
		if err != nil {
			return nil, err
		}
	}

	if config.CheckCommand != "" {
		checkCommand, err = NewCommand(config.CheckCommand, "", timeout)
		if err != nil {
			return nil, err
		}
	}

	if config.StandbyCommand != "" {
		standbyCommand, err = NewCommand(config.StandbyCommand, "", timeout)
		if err != nil {
			return nil, err
		}
	}

	return &DiskStatus{
		checkCommand:   checkCommand,
		Name:           name,
		DiskActive:     config.Active,
		DiskStandby:    config.Standby,
		DiskSleeping:   config.Sleeping,
		File:           config.File,
		standbyCommand: standbyCommand,
		mutex:          sync.Mutex{},
	}, nil
}

func (s *DiskStatus) Get(expandEnv func(string) string) enum.DiskStatus {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var output string
	var err error

	if s.checkCommand == nil && len(s.File) == 0 {
		return enum.DiskStatusUnknown
	}

	// we check that the interface value is not nil, then that the implementation value is not nil
	if s.checkCommand != nil && s.checkCommand.(*Command) != nil {
		output, err = s.checkCommand.Run(nil, expandEnv)
		if err != nil {
			return enum.DiskStatusUnknown
		}
	} else if len(s.File) > 0 {
		// we want to reference and use a Sensor instead of copying this code (from Sensor)
		// resolve file name
		filename := os.Expand(s.File, expandEnv)
		matches, err := filepath.Glob(filename)
		if err != nil {
			return enum.DiskStatusUnknown
		}
		if len(matches) != 1 {
			return enum.DiskStatusUnknown
		}
		filename = matches[0]
		// that's a bit much :)
		clog.Tracef("reading file %q", filename)
		bytesOutput, err := os.ReadFile(filename)
		if err != nil {
			return enum.DiskStatusUnknown
		}
		output = strings.TrimSpace(string(bytesOutput))
	}
	if strings.Contains(output, s.DiskActive) {
		return enum.DiskStatusActive
	}
	if strings.Contains(output, s.DiskStandby) {
		return enum.DiskStatusStandby
	}
	if strings.Contains(output, s.DiskSleeping) {
		return enum.DiskStatusSleeping
	}
	return enum.DiskStatusUnknown
}

func (s *DiskStatus) Standby(expandEnv func(string) string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// we check that the interface value is nil, or that the implementation value is nil
	if s.standbyCommand == nil || s.standbyCommand.(*Command) == nil {
		return errors.New("no command defined to put the disk in standby mode")
	}
	_, err := s.standbyCommand.Run(nil, expandEnv)
	return err
}
