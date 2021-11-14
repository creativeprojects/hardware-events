package lib

import (
	"fmt"
	"sync"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/creativeprojects/hardware-events/lib/simulation"
)

// Control temperature
type Control struct {
	config      cfg.FanControl
	mutex       sync.Mutex
	InitCommand CommandRunner
	SetCommand  CommandRunner
	ExitCommand CommandRunner
	Zones       map[string]*Zone
}

// NewControl creates a new fan controller from configuration
func NewControl(sensorReader SensorReader, config cfg.FanControl, simulate bool) (*Control, error) {
	var initCmd, setCmd, exitCmd CommandRunner
	var timeout time.Duration
	var err error

	if config.Timeout != "" {
		timeout, err = time.ParseDuration(config.Timeout)
		if err != nil {
			return nil, err
		}
	}
	// init command
	if config.InitCommand != "" {
		if simulate {
			initCmd, err = simulation.NewCommand(config.InitCommand, "")
		} else {
			initCmd, err = NewCommand(config.InitCommand, "", timeout)
		}
		if err != nil {
			return nil, err
		}
	}

	// set command
	if simulate {
		setCmd, err = simulation.NewCommand(config.SetCommand, "")
	} else {
		setCmd, err = NewCommand(config.SetCommand, "", timeout)
	}
	if err != nil {
		return nil, err
	}

	// exit command
	if config.ExitCommand != "" {
		if simulate {
			exitCmd, err = simulation.NewCommand(config.ExitCommand, "")
		} else {
			exitCmd, err = NewCommand(config.ExitCommand, "", timeout)
		}
		if err != nil {
			return nil, err
		}
	}

	control := &Control{
		config:      config,
		mutex:       sync.Mutex{},
		InitCommand: initCmd,
		SetCommand:  setCmd,
		ExitCommand: exitCmd,
	}
	zones := make(map[string]*Zone, len(config.Zones))
	for zoneName, zoneCfg := range config.Zones {
		zone, err := NewZone(sensorReader, zoneCfg, zoneName, control.setSpeedCommand)
		if err != nil {
			return nil, fmt.Errorf("cannot create zone %s: %v", zoneName, err)
		}
		zones[zoneName] = zone
	}
	control.Zones = zones
	return control, nil
}

// Init runs the fan initialization
func (c *Control) Init() error {
	if c.InitCommand == nil {
		return nil
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, err := c.InitCommand.Run(nil, nil)
	return err
}

// Exit resets the fan to be controlled back by the motherboard (or set a specific speed)
func (c *Control) Exit() error {
	if c.ExitCommand == nil {
		return nil
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, err := c.ExitCommand.Run(nil, nil)
	return err
}

// Start the controller. The method will return after it has kicked off the necessary goroutines.
func (c *Control) Start() {
	for name, zone := range c.Zones {
		clog.Debugf("starting %s", name)
		zone.Start()
	}
}

func (c *Control) setSpeedCommand(zoneID, speed int) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	_, err := c.SetCommand.Run(nil, func(input string) string {
		switch input {
		case "FAN_ZONE":
			return c.formatValue(input, zoneID)
		case "FAN_SPEED":
			return c.formatValue(input, speed)
		}
		return "$" + input
	})
	return err
}

func (c *Control) formatValue(name string, value int) string {
	format := "%d"
	if param, ok := c.config.Parameters[name]; ok {
		if param.Format != "" {
			format = param.Format
		}
	}
	return fmt.Sprintf(format, value)
}
