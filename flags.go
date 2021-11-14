package main

import (
	"flag"
)

// Flags contains command line flags
type Flags struct {
	configFile string
	quiet      bool
	verbose    bool
	debug      bool
	simulation bool
	seed       int64
}

var (
	flags Flags
)

func init() {
	flag.StringVar(&flags.configFile, "c", "config.yaml", "configuration file")
	flag.BoolVar(&flags.quiet, "q", false, "quiet - do not send any output")
	flag.BoolVar(&flags.verbose, "v", false, "verbose - display debugging information")
	flag.BoolVar(&flags.debug, "d", false, "debug - display full debugging information")
	flag.BoolVar(&flags.simulation, "s", false, "simulation mode - test your rules with simulated sensors")
	flag.Int64Var(&flags.seed, "r", 42, "random number seed to use in simulation mode")
}
