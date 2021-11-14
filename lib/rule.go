package lib

import (
	"math"
	"time"

	"github.com/creativeprojects/hardware-events/cfg"
)

// Rule conversion from temperature to fan speed
type Rule struct {
	slope           float64
	intercept       float64
	RunTimer        time.Duration
	TemperatureFrom int
	TemperatureTo   int
	FanFrom         int
	FanTo           int
	FanSet          int
}

// NewRule creates a new rule to convert a temperature into a fan speed
func NewRule(config cfg.SensorRule) (Rule, error) {
	var runTimer time.Duration
	var err error

	if config.RunEvery != "" {
		runTimer, err = time.ParseDuration(config.RunEvery)
		if err != nil {
			return Rule{}, err
		}
	}
	slope := float64(config.Fan.To-config.Fan.From) / float64(config.Temperature.To-config.Temperature.From)
	// works for either From or To (picked To):
	intercept := float64(config.Fan.To) - slope*float64(config.Temperature.To)
	return Rule{
		slope:           slope,
		intercept:       intercept,
		RunTimer:        runTimer,
		TemperatureFrom: config.Temperature.From,
		TemperatureTo:   config.Temperature.To,
		FanFrom:         config.Fan.From,
		FanTo:           config.Fan.To,
		FanSet:          config.Fan.Set,
	}, nil
}

// MatchTemperature is true when this rule matches the input temperature
func (r Rule) MatchTemperature(temperature int) bool {
	if r.TemperatureFrom > 0 && temperature < r.TemperatureFrom {
		return false
	}
	if r.TemperatureTo > 0 && temperature > r.TemperatureTo {
		return false
	}
	return true
}

// CalculateFanSpeed from the rule. If the temperature is out of range it will return a min, max or set speed
// The second value returned is a request to change the temperature reading timer
func (r Rule) CalculateFanSpeed(temperature int) (int, time.Duration) {
	if r.FanSet > 0 {
		return r.FanSet, r.RunTimer
	}
	if r.TemperatureFrom > 0 && temperature < r.TemperatureFrom {
		// bellow the range
		return r.min(), 0
	}
	if r.TemperatureTo > 0 && temperature > r.TemperatureTo {
		// above the range
		return r.max(), 0
	}
	if r.slope == 0 {
		// configuration error
		return 0, 0
	}
	return int(math.Round(float64(temperature)*r.slope + r.intercept)), r.RunTimer
}

func (r Rule) min() int {
	if r.FanSet > 0 {
		return r.FanSet
	}
	return r.FanFrom
}

func (r Rule) max() int {
	if r.FanSet > 0 {
		return r.FanSet
	}
	return r.FanTo
}
