package lib

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDiskDevice(t *testing.T) {
	// test that a full path can be resolved into a 3 letters device (sda)
	const diskByIdPath = "drivetemp_test_files/dev/disk/by-id/"
	entries, err := os.ReadDir(diskByIdPath)
	require.NoError(t, err)

	for id, entry := range entries {
		devicePath := filepath.Join(diskByIdPath, entry.Name())
		config := cfg.Disk{
			Device: devicePath,
		}
		disk, err := NewDisk(&Global{}, strconv.Itoa(id), config, nil)
		require.NoError(t, err)

		assert.Len(t, filepath.Base(disk.Device), 3)
	}
}
