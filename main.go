package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/creativeprojects/hardware-events/lib"
)

// These fields are populated by the goreleaser build
var (
	version = "0.1.0-dev"
	commit  = ""
	date    = ""
	builtBy = ""
)

func main() {
	var exitCode = 0
	var err error

	// run all defer functions before returning with an exit code
	defer func() {
		if exitCode != 0 {
			os.Exit(exitCode)
		}
	}()

	flag.Parse()

	cleanLogger := setupLogger(flags)
	if cleanLogger != nil {
		defer cleanLogger()
	}

	// keep this one last if possible (so it will be first at the end)
	defer showPanicData()

	clog.Debugf("hardware-events %s compiled with %s", version, runtime.Version())

	config, err := cfg.LoadFileConfig(flags.configFile)
	if err != nil {
		clog.Errorf("cannot load configuration: %v", err)
		exitCode = 1
		return
	}

	if flags.simulation {
		config.Simulation = true
	}
	if config.Simulation {
		clog.Warningf("running in simulation mode with seeds = %d and %d", flags.seed1, flags.seed2)
		config.Seed1 = flags.seed1
		config.Seed2 = flags.seed2
	}

	global, err := lib.NewGlobal(config)
	if err != nil {
		clog.Errorf("cannot load configuration: %v", err)
		exitCode = 1
		return
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	notifyReady()

	// systemd watchdog
	go setupWatchdog()

	// run all startup tasks
	for _, task := range global.GetStartupTasks() {
		clog.Debugf("running startup task %s", task.Name)
		err := task.Execute()
		if err != nil {
			clog.Error(err)
			continue
		}
	}

	// setup all the recurring tasks
	global.StartTimers()

	// start fan control
	err = global.FanControl.Init()
	if err != nil {
		clog.Errorf("cannot initialize fan control: %s", err)
	} else {
		global.FanControl.Start()
	}

	closeTelemetry, err := setupTelemetry(config, global)
	if err != nil {
		clog.Errorf("cannot start telemetry: %v", err)
		exitCode = 1
		return
	}

	closeMetricsServer, err := setupMetricsServer(config)
	if err != nil {
		clog.Errorf("cannot start http server: %v", err)
		exitCode = 1
		return
	}

	// wait until we're politely asked to leave
	<-stop
	ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
	closeMetricsServer(ctx)
	closeTelemetry(ctx)
	signal.Stop(stop)
	_ = global.FanControl.Exit()
	notifyLeaving()
	fmt.Println("Bye bye!")
}
