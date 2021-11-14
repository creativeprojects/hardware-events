package main

import (
	"log"

	"github.com/creativeprojects/clog"
)

// setupLogger returns a cleaning function, if any
func setupLogger(flags Flags) func() {
	level := clog.LevelInfo
	if flags.debug {
		level = clog.LevelTrace
	} else if flags.verbose {
		level = clog.LevelDebug
	} else if flags.quiet {
		level = clog.LevelWarning
	}
	clog.SetDefaultLogger(
		clog.NewLogger(
			clog.NewLevelFilter(level,
				clog.NewConsoleHandler("", log.LstdFlags))))

	return nil
}
