package lib

import (
	"sync"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/creativeprojects/hardware-events/constants"
)

// SensorReader is a function that takes a name and returns a function to read the named sensor
type SensorReader func(sensorName string) func() (int, error)

// Zone fan control
type Zone struct {
	currentSpeed   int
	requestedSpeed map[string]int // fan speed requested by all the different sensor rules
	defaultSpeed   int
	minSpeed       int
	maxSpeed       int
	setSpeed       func(zoneID, speed int) error // function to send the fan speed to the hardware
	mutex          sync.Mutex
	ID             int
	Name           string
	Sensors        map[string]*TemperatureSensor
}

// NewZone creates a new fan control zone
func NewZone(sensorReader SensorReader, config cfg.FanZone, name string, setSpeed func(zoneID, speed int) error) (*Zone, error) {
	var timer time.Duration
	var err error

	if config.RunEvery != "" {
		timer, err = time.ParseDuration(config.RunEvery)
		if err != nil {
			return nil, err
		}
	}
	defaultSpeed := constants.DefaultSpeed
	minSpeed := constants.DefaultMinFanSpeed
	maxSpeed := constants.DefaultMaxFanSpeed
	if config.DefaultSpeed > 0 {
		defaultSpeed = config.DefaultSpeed
	}
	if config.MinSpeed > 0 {
		minSpeed = config.MinSpeed
	}
	if config.MaxSpeed > 0 {
		maxSpeed = config.MaxSpeed
	}

	zone := &Zone{
		requestedSpeed: make(map[string]int, len(config.Sensors)),
		setSpeed:       setSpeed,
		defaultSpeed:   defaultSpeed,
		minSpeed:       minSpeed,
		maxSpeed:       maxSpeed,
		mutex:          sync.Mutex{},
		ID:             config.ID,
		Name:           name,
	}

	sensors := make(map[string]*TemperatureSensor, len(config.Sensors))
	for sensorName, sensorCfg := range config.Sensors {
		sensor, err := NewTemperatureSensor(sensorCfg, sensorName, timer, sensorReader(sensorName), zone.RequestSpeed)
		if err != nil {
			return nil, err
		}
		sensors[sensorName] = sensor
	}
	zone.Sensors = sensors

	return zone, nil
}

func (z *Zone) Start() {
	for _, zoneSensor := range z.Sensors {
		go zoneSensor.Run()
	}
}

func (z *Zone) SetSpeed(speed int) {
	if speed < z.minSpeed {
		speed = z.minSpeed
	}
	if speed > z.maxSpeed {
		speed = z.maxSpeed
	}
	if z.currentSpeed == speed {
		return
	}
	clog.Tracef("%s: set fan speed %d%%", z.Name, speed)
	z.currentSpeed = speed
	if z.setSpeed != nil {
		z.setSpeed(z.ID, speed)
	}
}

func (z *Zone) RequestSpeed(name string, speed int, min, max bool) {
	if min {
		speed = z.minSpeed
	} else if max {
		speed = z.maxSpeed
	}
	z.mutex.Lock()
	defer z.mutex.Unlock()

	z.requestedSpeed[name] = speed
	clog.Tracef("%s: request fan speed %d%%", name, speed)

	// now pick the highest bid
	for _, bid := range z.requestedSpeed {
		if bid > speed {
			speed = bid
		}
	}
	z.SetSpeed(speed)
}
