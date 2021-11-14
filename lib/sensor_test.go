package lib

import (
	"embed"
	"io/fs"
	"os"
	"strings"
	"testing"

	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//go:embed test_files
var testingFS embed.FS

func TestNotEnoughInformation(t *testing.T) {
	_, err := NewSensor("name", cfg.Task{}, nil)
	assert.Error(t, err)
}

func TestSimpleFileWithDivider(t *testing.T) {
	sensor, err := NewSensor("name", cfg.Task{File: "test_files/temp2_input", Divider: 1000}, testingFS)
	require.NoError(t, err)
	temperature, err := sensor.Get(nil)
	assert.NoError(t, err)
	assert.Equal(t, 20, temperature)
}

func TestAverageFiles(t *testing.T) {
	sensor, err := NewSensor("name", cfg.Task{Files: []string{"test_files/temp?_input"}, Divider: 1000}, testingFS)
	require.NoError(t, err)
	temperature, err := sensor.Get(nil)
	assert.NoError(t, err)
	assert.Equal(t, 30, temperature)
}

// This is not a unit test as such but a quick check on the behavior of fs.FS.Open
func TestReadDiskViaFS(t *testing.T) {
	wellKnowFile := "/proc/version"
	filesystem := os.DirFS("/")
	// you cannot use a rooted file with fs.FS interface
	_, err := fs.Stat(filesystem, wellKnowFile)
	assert.ErrorIs(t, err, fs.ErrInvalid)
	// but this should work
	fstat, err := fs.Stat(filesystem, strings.TrimPrefix(wellKnowFile, "/"))
	assert.NoError(t, err)
	assert.NotEmpty(t, fstat)
}
