package simulation

import (
	"os"
	"sync"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/creativeprojects/hardware-events/lib/enum"
)

type DiskStatus struct {
	status         map[string]enum.DiskStatus
	checkCommand   string
	standbyCommand string
	mutex          sync.Mutex
}

func NewDiskStatus(config cfg.DiskPowerStatus) (*DiskStatus, error) {
	return &DiskStatus{
		status:         make(map[string]enum.DiskStatus),
		checkCommand:   config.CheckCommand,
		standbyCommand: config.StandbyCommand,
		mutex:          sync.Mutex{},
	}, nil
}

func (s *DiskStatus) Get(expandEnv func(string) string) enum.DiskStatus {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	command := os.Expand(s.checkCommand, expandEnv)
	clog.Debugf("command: %s", command)

	disk := expandEnv("DEVICE")
	if status, ok := s.status[disk]; ok {
		return status
	}
	s.status[disk] = enum.DiskStatusActive
	return s.status[disk]
}

func (s *DiskStatus) Standby(expandEnv func(string) string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	command := os.Expand(s.standbyCommand, expandEnv)
	clog.Debugf("command: %s", command)

	disk := expandEnv("DEVICE")
	s.status[disk] = enum.DiskStatusStandby
	return nil
}
