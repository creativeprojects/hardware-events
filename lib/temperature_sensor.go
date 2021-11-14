package lib

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/creativeprojects/hardware-events/intmath"
)

type TemperatureSensor struct {
	valuesCount     int
	values          []int
	mutex           sync.Mutex
	readTemperature func() (int, error)
	requestSpeed    func(name string, speed int, min, max bool) // function to send the calculated speed to - if no speed value we can ask for min or max speed
	DefaultTimer    time.Duration
	RunTimer        time.Duration
	Name            string
	Rules           []Rule
	minTemp         int // minimum temp from the rules
	maxTemp         int // maximum temp from the rules
}

func NewTemperatureSensor(config cfg.Sensor, name string, timer time.Duration, readTemperature func() (int, error), requestSpeed func(string, int, bool, bool)) (*TemperatureSensor, error) {
	var err error

	average, err := time.ParseDuration(config.Average)
	if err != nil {
		return nil, fmt.Errorf("invalid average: %w", err)
	}
	// take the timer from the sensor configuration if defined
	if config.RunEvery != "" {
		timer, err = time.ParseDuration(config.RunEvery)
		if err != nil {
			return nil, err
		}
	}
	count := average / timer
	if count <= 0 {
		return nil, fmt.Errorf("cannot keep an average of %s of data when taking values every %s", average, timer)
	}
	rules := make([]Rule, len(config.Rules))
	for id, ruleCfg := range config.Rules {
		rule, err := NewRule(ruleCfg)
		if err != nil {
			return nil, err
		}
		rules[id] = rule
	}
	// sort rules by temperature descending
	sort.Slice(rules, func(i, j int) bool {
		if rules[i].TemperatureFrom > 0 && rules[j].TemperatureFrom > 0 {
			return rules[i].TemperatureFrom > rules[j].TemperatureFrom
		}
		if rules[i].TemperatureTo > 0 && rules[j].TemperatureTo > 0 {
			return rules[i].TemperatureTo > rules[j].TemperatureTo
		}
		// @todo: finish sorting out rules with no from or no to
		return false
	})
	// find minimum and maximum rule temperatures
	minTemp, maxTemp := 0, 0
	for _, rule := range rules {
		if minTemp == 0 && rule.TemperatureFrom > 0 {
			minTemp = rule.TemperatureFrom
		}
		if maxTemp == 0 && rule.TemperatureTo > 0 {
			maxTemp = rule.TemperatureTo
		}
		if rule.TemperatureFrom > 0 && minTemp > rule.TemperatureFrom {
			minTemp = rule.TemperatureFrom
		}
		if rule.TemperatureTo > 0 && maxTemp < rule.TemperatureTo {
			maxTemp = rule.TemperatureTo
		}
	}
	return &TemperatureSensor{
		valuesCount:     int(count),
		values:          make([]int, 0, int(count)),
		mutex:           sync.Mutex{},
		readTemperature: readTemperature,
		requestSpeed:    requestSpeed,
		DefaultTimer:    timer,
		RunTimer:        timer,
		Name:            name,
		Rules:           rules,
		minTemp:         minTemp,
		maxTemp:         maxTemp,
	}, nil
}

// Run reads temperature and requests a change in fan speed. This method runs in a infinite loop and should be called inside a goroutine.
func (s *TemperatureSensor) Run() {
	for {
		time.Sleep(s.RunTimer)
		err := s.run()
		if err != nil {
			clog.Error(err)
			return
		}
	}
}

func (s *TemperatureSensor) run() error {
	if s.readTemperature == nil {
		// nothing for me to do here
		return fmt.Errorf("%s: no temperature sensor attached, cancelling Run() now", s.Name)
	}
	if s.requestSpeed == nil {
		// nothing for me to do here
		return fmt.Errorf("%s: no fan speed attached, cancelling Run() now", s.Name)
	}
	temperature, err := s.readTemperature()
	if err != nil {
		return fmt.Errorf("%s: %v", s.Name, err)
	}
	temperature = s.average(temperature)
	clog.Tracef("%s: %d°C (average %d * %v)", s.Name, temperature, s.valuesCount, s.RunTimer)
	for _, rule := range s.Rules {
		if rule.MatchTemperature(temperature) {
			speed, timer := rule.CalculateFanSpeed(temperature)
			if timer > 0 {
				s.RunTimer = timer
			} else {
				s.RunTimer = s.DefaultTimer
			}
			s.requestSpeed(s.Name, speed, false, false)
			return nil
		}
	}

	// No temperature found, let's guess if we go for the min or the max
	if temperature > s.maxTemp {
		s.requestSpeed(s.Name, 0, false, true)
		// don't change the timer at this stage (keep the latest set)
		return nil
	}
	// set minimum then
	s.requestSpeed(s.Name, 0, true, false)
	// also put default timer back
	s.RunTimer = s.DefaultTimer
	return nil
}

func (s *TemperatureSensor) average(temperature int) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if len(s.values) < s.valuesCount {
		s.values = append(s.values, temperature)
	} else {
		s.values = append(s.values[1:s.valuesCount], temperature)
	}
	return intmath.Avg(s.values)
}
