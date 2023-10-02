package simulation

import (
	"math"
	"math/rand"
	"strconv"
	"sync"

	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/creativeprojects/hardware-events/constants"
)

// Sensor simulates a hardware sensor
type Sensor struct {
	config        cfg.Task
	mutex         sync.Mutex
	values        map[string]float64
	randGenerator *rand.Rand
	Name          string
}

// NewSensor creates a new simulated sensor
func NewSensor(name string, config cfg.Task, randGenerator *rand.Rand) (*Sensor, error) {
	return &Sensor{
		config:        config,
		mutex:         sync.Mutex{},
		values:        make(map[string]float64),
		randGenerator: randGenerator,
		Name:          name,
	}, nil
}

// Get an integer value from the simulated sensor
func (s *Sensor) Get(expandEnv func(string) string) (int, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	device := ""
	if expandEnv != nil {
		device = expandEnv("DEVICE")
	}
	value := s.values[device]
	if value == 0 {
		value = constants.SimulationStartTemp
	}
	// when the temperature is bellow the middle range, make it raise faster
	adjustMiddle := 4.0
	if value > (constants.SimulationMaxTemp-constants.SimulationMinTemp)/2 {
		// but when it's above the middle, make it cool down faster
		adjustMiddle = 6
	}
	value += s.randGenerator.Float64()*10 - adjustMiddle
	if value < constants.SimulationMinTemp {
		value = constants.SimulationMinTemp
	}
	if value > constants.SimulationMaxTemp {
		value = constants.SimulationMaxTemp
	}
	s.values[device] = value
	return int(math.Round(value)), nil
}

// GetRaw value from the simulated sensor
func (s *Sensor) GetRaw(expandEnv func(string) string) (string, error) {
	value, err := s.Get(expandEnv)
	return strconv.Itoa(value), err
}
