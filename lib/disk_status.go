package lib

import (
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/creativeprojects/hardware-events/lib/enum"
)

type DiskStatuser interface {
	Get(expandEnv func(string) string) enum.DiskStatus
	Standby(expandEnv func(string) string) error
}

type DiskStatus struct {
	checkCommand   CommandRunner
	DiskActive     string
	DiskStandby    string
	DiskSleeping   string
	standbyCommand CommandRunner
	mutex          sync.Mutex
}

func NewDiskStatus(config cfg.DiskPowerStatus) (*DiskStatus, error) {
	var timeout time.Duration
	var err error

	if config.Timeout != "" {
		timeout, err = time.ParseDuration(config.Timeout)
		if err != nil {
			return nil, err
		}
	}

	checkCommand, err := NewCommand(config.CheckCommand, "", timeout)
	if err != nil {
		return nil, err
	}

	standbyCommand, err := NewCommand(config.StandbyCommand, "", timeout)
	if err != nil {
		return nil, err
	}

	return &DiskStatus{
		checkCommand:   checkCommand,
		DiskActive:     config.Active,
		DiskStandby:    config.Standby,
		DiskSleeping:   config.Sleeping,
		standbyCommand: standbyCommand,
		mutex:          sync.Mutex{},
	}, nil
}

func (s *DiskStatus) Get(expandEnv func(string) string) enum.DiskStatus {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.checkCommand == nil {
		return enum.DiskStatusUnknown
	}
	output, err := s.checkCommand.Run(nil, expandEnv)
	if err != nil {
		return enum.DiskStatusUnknown
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

	if s.standbyCommand == nil {
		return errors.New("no command defined to put the disk in standby mode")
	}
	_, err := s.standbyCommand.Run(nil, expandEnv)
	return err
}
