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
	seed1      uint64
	seed2      uint64
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
	flag.Uint64Var(&flags.seed1, "r1", 42, "random number seed to use in simulation mode")
	flag.Uint64Var(&flags.seed2, "r2", 42, "random number seed to use in simulation mode")
}
