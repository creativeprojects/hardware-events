package cfg

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEmptyConfiguration(t *testing.T) {
	config, err := loadConfig(bytes.NewReader([]byte("---")))
	assert.NoError(t, err)
	assert.Len(t, config.Disks, 0)
}

func TestSimpleConfiguration(t *testing.T) {
	content := `---
disks:
  # comment
  disk1:
    device: "/dev/sda"
    temperature_sensor: smartctl
    monitor_temperature: always
  disk2:
    device: "/dev/sdb"
    temperature_sensor: none
    monitor_temperature: never
`
	config, err := loadConfig(bytes.NewReader([]byte(content)))
	assert.NoError(t, err)
	assert.Len(t, config.Disks, 2)
}
