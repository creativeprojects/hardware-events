package simulation

import (
	"testing"

	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/creativeprojects/hardware-events/constants"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultValues(t *testing.T) {
	sensor, err := NewSensor("name", cfg.Task{})
	require.NoError(t, err)

	testData := []func(string) string{
		nil,
		func(string) string { return "/dev/sda" },
		func(string) string { return "/dev/sdb" },
		func(string) string { return "/dev/sdc" },
	}

	for _, testItem := range testData {
		value, err := sensor.Get(testItem)
		require.NoError(t, err)
		assert.InEpsilon(t, float64(constants.SimulationStartTemp), value, 5)
	}
}
