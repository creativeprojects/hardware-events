package lib

import (
	"bytes"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSimpleEcho(t *testing.T) {
	command, err := NewCommand("echo TestSimpleEcho", "", 0)
	require.NoError(t, err)
	output, err := command.Run(nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "TestSimpleEcho\n", output)
}

func TestCancelledCommand(t *testing.T) {
	command, err := NewCommand("ls", "", 1*time.Microsecond)
	require.NoError(t, err)
	_, err = command.Run(nil, nil)
	require.Error(t, err)
}

func TestExpandEnv(t *testing.T) {
	command, err := NewCommand("echo ${TOTO}", "", 0)
	require.NoError(t, err)
	output, err := command.Run(nil, func(input string) string {
		if input == "TOTO" {
			return "Found!"
		}
		return "ERROR"
	})
	require.NoError(t, err)
	assert.Equal(t, "Found!\n", output)
}

func TestExpandEnvNotFound(t *testing.T) {
	command, err := NewCommand("echo ${OTHER}", "", 0)
	require.NoError(t, err)
	output, err := command.Run(nil, func(input string) string {
		if input == "TOTO" {
			return "Found!"
		}
		return "NOT Found!"
	})
	require.NoError(t, err)
	assert.Equal(t, "NOT Found!\n", output)
}

func TestPatternNotFound(t *testing.T) {
	command, err := NewCommand("echo TestPatternNotFound", `Current Temperature:\s+(\d+) Celsius`, 0)
	require.NoError(t, err)
	output, err := command.Run(nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "TestPatternNotFound\n", output)
}

func TestPatternFound(t *testing.T) {
	command, err := NewCommand("echo Current Temperature:                    26 Celsius", `Current Temperature:\s+(\d+) Celsius`, 0)
	require.NoError(t, err)
	output, err := command.Run(nil, nil)
	require.NoError(t, err)
	assert.Equal(t, "26", output)
}

func TestFullPatternFomStdin(t *testing.T) {
	stdin := `smartctl 6.6 2017-11-05 r4594 [x86_64-linux-4.19.0-14-amd64] (local build)
Copyright (C) 2002-17, Bruce Allen, Christian Franke, www.smartmontools.org

=== START OF READ SMART DATA SECTION ===
SCT Status Version:                  3
SCT Version (vendor specific):       256 (0x0100)
SCT Support Level:                   1
Device State:                        Active (0)
Current Temperature:                    27 Celsius
Power Cycle Min/Max Temperature:     22/40 Celsius
Lifetime    Min/Max Temperature:      0/70 Celsius
Under/Over Temperature Limit Count:   0/0

`
	command, err := NewCommand("cat", `Current Temperature:\s+(\d+) Celsius`, 0)
	require.NoError(t, err)
	output, err := command.Run(bytes.NewBufferString(stdin), nil)
	require.NoError(t, err)
	assert.Equal(t, "27", output)
}
