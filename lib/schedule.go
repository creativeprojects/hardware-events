package lib

import (
	"slices"
	"strings"
	"time"

	"github.com/creativeprojects/hardware-events/cfg"
)

type Schedule struct {
	global *Global
	Task   *Task
	When   []string
}

func NewSchedule(global *Global, config cfg.Schedule) *Schedule {
	return &Schedule{
		global: global,
		Task:   global.Tasks[config.Task],
		When:   config.When,
	}
}

func (s *Schedule) OnStartup() bool {
	return slices.Contains(s.When, "startup")
}

func (s *Schedule) OnTimer() time.Duration {
	for _, when := range s.When {
		if strings.HasPrefix(when, "every ") {
			duration, err := time.ParseDuration(when[6:])
			if err != nil {
				return 0
			}
			return duration
		}
	}
	return 0
}
