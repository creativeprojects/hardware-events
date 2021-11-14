package lib

import (
	"strconv"
	"testing"
	"time"

	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInvalidAverage(t *testing.T) {
	timer := 15 * time.Second
	_, err := NewTemperatureSensor(cfg.Sensor{Average: "10s"}, "name", timer, nil, nil)
	t.Log(err)
	assert.Error(t, err)
}

func TestInvalidAverageFromSensorConfig(t *testing.T) {
	timer := 15 * time.Second
	_, err := NewTemperatureSensor(cfg.Sensor{Average: "30s", RunEvery: "1m"}, "name", timer, nil, nil)
	t.Log(err)
	assert.Error(t, err)
}

func TestValidAverage(t *testing.T) {
	testData := []struct {
		every       string
		average     string
		valuesCount int
	}{
		{"15s", "15s", 1},
		{"14s", "15s", 1},
		{"8s", "15s", 1},
		{"7s", "15s", 2},
		{"1s", "15s", 15},
	}
	for _, testItem := range testData {
		t.Run("", func(t *testing.T) {
			timer, _ := time.ParseDuration(testItem.every)
			sensor, err := NewTemperatureSensor(cfg.Sensor{Average: testItem.average}, "name", timer, nil, nil)
			require.NoError(t, err)
			assert.Equal(t, testItem.valuesCount, sensor.valuesCount)
		})
	}
}

func TestAverage(t *testing.T) {
	testData := []struct {
		input   int
		average int
	}{
		{10, 10},
		{20, 15},
		{30, 20},
		{40, 30},
		{50, 40},
		{60, 50},
	}

	sensor, err := NewTemperatureSensor(cfg.Sensor{Average: "30s"}, "name", 10*time.Second, nil, nil)
	require.NoError(t, err)

	for _, testItem := range testData {
		t.Run(strconv.Itoa(testItem.input), func(t *testing.T) {
			avg := sensor.average(testItem.input)
			assert.Equal(t, testItem.average, avg)
		})
	}
}

func TestRunValidCalculation(t *testing.T) {
	config := cfg.Sensor{
		Average: "10s",
		Rules: []cfg.SensorRule{
			{Temperature: cfg.FromTo{From: 20, To: 40}, Fan: cfg.SetFromTo{Set: 40}},
			{Temperature: cfg.FromTo{From: 40, To: 60}, Fan: cfg.SetFromTo{From: 50, To: 90}},
			{Temperature: cfg.FromTo{From: 60, To: 80}, Fan: cfg.SetFromTo{Set: 100}},
		},
	}

	testData := []struct {
		temperature int
		speed       int
		min         bool
		max         bool
	}{
		{10, 0, true, false},
		{20, 40, false, false},
		{40, 50, false, false},
		{50, 70, false, false},
		{60, 100, false, false},
		{80, 100, false, false},
		{100, 0, false, true},
	}

	for _, testItem := range testData {
		t.Run(strconv.Itoa(testItem.temperature), func(t *testing.T) {
			asserted := false
			sensor, err := NewTemperatureSensor(config, "name", 10*time.Second, func() (int, error) {
				// reader function
				return testItem.temperature, nil
			}, func(name string, speed int, min, max bool) {
				// setter function
				assert.Equal(t, testItem.speed, speed)
				assert.Equal(t, testItem.min, min)
				assert.Equal(t, testItem.max, max)
				asserted = true
			})
			require.NoError(t, err)

			err = sensor.run()
			assert.NoError(t, err)
			assert.True(t, asserted)
		})
	}
}
