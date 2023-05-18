package lib

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/creativeprojects/hardware-events/intmath"
)

type SensorGetter interface {
	Get(func(string) string) (int, error)
}

// Sensor reads value from the hardware (command line or sysfs file)
type Sensor struct {
	config       cfg.Task
	mutex        sync.Mutex
	fs           fs.FS
	aggregate    intmath.Aggregation
	timeout      time.Duration
	Name         string
	Command      CommandRunner
	File         string
	AverageFiles []string
}

// NewSensor creates a new access to the hardware
func NewSensor(name string, config cfg.Task, fileSystem fs.FS) (*Sensor, error) {
	var aggregate intmath.Aggregation
	var timeout time.Duration
	var err error

	if config.Timeout != "" {
		timeout, err = time.ParseDuration(config.Timeout)
		if err != nil {
			return nil, err
		}
	}

	if config.Command != "" {
		command, err := NewCommand(config.Command, config.Regexp, timeout)
		if err != nil {
			return nil, err
		}
		return &Sensor{
			config:    config,
			mutex:     sync.Mutex{},
			fs:        fileSystem,
			aggregate: aggregate,
			timeout:   timeout,
			Name:      name,
			Command:   command,
		}, nil
	}
	if config.File != "" {
		return &Sensor{
			config:    config,
			mutex:     sync.Mutex{},
			fs:        fileSystem,
			aggregate: aggregate,
			timeout:   timeout,
			Name:      name,
			File:      config.File,
		}, nil
	}
	if len(config.Files) > 0 {
		return &Sensor{
			config:       config,
			mutex:        sync.Mutex{},
			fs:           fileSystem,
			aggregate:    aggregate,
			timeout:      timeout,
			Name:         name,
			AverageFiles: config.Files,
		}, nil
	}
	return nil, errors.New("not enough information")
}

// Get an integer value from the sensor
func (s *Sensor) Get(expandEnv func(string) string) (int, error) {
	output, err := s.getRaw(expandEnv)
	if err != nil {
		return 0, err
	}
	if len(output) == 0 {
		return 0, errors.New("no value returned")
	}
	values := make([]int, len(output))
	for i, rawValue := range output {
		values[i], err = strconv.Atoi(rawValue)
		if err != nil {
			return 0, err
		}
	}
	temp := values[0]
	if len(values) > 1 {
		temp = intmath.Aggregate(values, s.aggregate)
	}
	if s.config.Divider > 0 {
		temp /= s.config.Divider
	}
	return temp, nil
}

// getRaw value(s) from the command line or a file (sysfs)
func (s *Sensor) getRaw(expandEnv func(string) string) ([]string, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.Command != nil {
		output, err := s.Command.Run(nil, expandEnv)
		if err != nil {
			return nil, err
		}
		return []string{strings.TrimSpace(output)}, nil
	}
	if s.File != "" {
		// resolve file name
		filename := os.Expand(s.File, expandEnv)
		matches, err := filepath.Glob(filename)
		if err != nil {
			return nil, fmt.Errorf("cannot find sensor file: %w", err)
		}
		if len(matches) != 1 {
			return nil, fmt.Errorf("expected 1 file but found %d: %q", len(matches), filename)
		}
		filename = matches[0]
		// that's a bit much :)
		clog.Tracef("reading file %q", filename)
		output, err := s.readFile(filename)
		if err != nil {
			return nil, err
		}
		return []string{output}, nil
	}
	if len(s.AverageFiles) > 0 {
		values := make([]string, 0, len(s.AverageFiles))
		for _, fileglob := range s.AverageFiles {
			matches, err := fs.Glob(s.fs, fileglob)
			if err != nil {
				return nil, err
			}
			for _, filename := range matches {
				// that's a bit much :)
				// clog.Tracef("reading file %q", filename)
				output, err := s.readFile(filename)
				if err != nil {
					return nil, err
				}
				values = append(values, output)
			}
		}
		return values, nil
	}
	return nil, nil
}

func (s *Sensor) readFile(filename string) (string, error) {
	// to use fs.Readfile we need to strip the first "/" from the path. let's leave it for now
	// file, err := fs.ReadFile(s.fs, filename)
	output, err := os.ReadFile(filename)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}
