package lib

import (
	"strconv"
	"testing"
	"time"

	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetFan(t *testing.T) {
	rule, err := NewRule(cfg.SensorRule{
		Temperature: cfg.FromTo{From: 10, To: 90},
		Fan:         cfg.SetFromTo{Set: 50},
	})
	require.NoError(t, err)

	testData := []struct {
		temperature int
		fan         int
	}{
		{0, 50},
		{10, 50},
		{50, 50},
		{90, 50},
		{100, 50},
	}

	for _, testItem := range testData {
		t.Run(strconv.Itoa(testItem.temperature), func(t *testing.T) {
			speed, timer := rule.CalculateFanSpeed(testItem.temperature)
			assert.Equal(t, testItem.fan, speed)
			assert.Equal(t, time.Duration(0), timer)
		})
	}
}

func TestLinearFan(t *testing.T) {
	rule, err := NewRule(cfg.SensorRule{
		Temperature: cfg.FromTo{From: 20, To: 80},
		Fan:         cfg.SetFromTo{From: 50, To: 100},
		RunEvery:    "1m",
	})
	require.NoError(t, err)

	testData := []struct {
		temperature int
		fan         int
		timer       time.Duration
	}{
		{0, 50, 0},
		{20, 50, time.Minute},
		{50, 75, time.Minute},
		{80, 100, time.Minute},
		{90, 100, 0},
	}

	for _, testItem := range testData {
		t.Run(strconv.Itoa(testItem.temperature), func(t *testing.T) {
			speed, timer := rule.CalculateFanSpeed(testItem.temperature)
			assert.Equal(t, testItem.fan, speed)
			assert.Equal(t, testItem.timer, timer)
		})
	}
}

func TestTemperatureMatchRule(t *testing.T) {
	testData := []struct {
		config      cfg.FromTo
		temperature int
		match       bool
	}{
		{cfg.FromTo{From: 10}, 7, false},
		{cfg.FromTo{To: 20}, 8, true},
		{cfg.FromTo{From: 10, To: 20}, 9, false},
		{cfg.FromTo{From: 10, To: 20}, 10, true},
		{cfg.FromTo{From: 10, To: 20}, 20, true},
		{cfg.FromTo{From: 10, To: 20}, 21, false},
		{cfg.FromTo{From: 10}, 22, true},
		{cfg.FromTo{To: 20}, 23, false},
	}

	for _, testItem := range testData {
		t.Run(strconv.Itoa(testItem.temperature), func(t *testing.T) {
			rule, err := NewRule(cfg.SensorRule{Temperature: testItem.config})
			require.NoError(t, err)
			assert.Equal(t, testItem.match, rule.MatchTemperature(testItem.temperature))
		})
	}
}
