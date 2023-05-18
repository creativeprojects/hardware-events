package lib

import (
	"testing"

	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/creativeprojects/hardware-events/lib/enum"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCanReadStatusFromFile(t *testing.T) {
	diskStatus, err := NewDiskStatus("test", cfg.DiskPowerStatus{
		File: "fs_test_files/sys/block/nvme0n1/device/state",
	})
	require.NoError(t, err)

	assert.NotNil(t, diskStatus)
	assert.Equal(t, enum.DiskStatusActive, diskStatus.Get(func(s string) string {
		return s
	}))
}

func TestNoStandbyIsReturningAnError(t *testing.T) {
	diskStatus, err := NewDiskStatus("test", cfg.DiskPowerStatus{})
	require.NoError(t, err)

	assert.NotNil(t, diskStatus)
	err = diskStatus.Standby(func(s string) string {
		return s
	})
	assert.ErrorContains(t, err, "no command defined to put the disk in standby mode")
}
