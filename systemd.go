package main

import (
	"time"

	"github.com/coreos/go-systemd/v22/daemon"
	"github.com/creativeprojects/clog"
)

func notifyReady() {
	_, err := daemon.SdNotify(false, daemon.SdNotifyReady)
	if err != nil {
		clog.Errorf("cannot notify systemd: %s", err)
	}
}

func notifyLeaving() {
	_, _ = daemon.SdNotify(false, daemon.SdNotifyStopping)
}

func setupWatchdog() {
	interval, err := daemon.SdWatchdogEnabled(false)
	if err != nil {
		clog.Errorf("cannot verify if systemd watchdog is enabled: %s", err)
		return
	}
	if interval == 0 {
		// watchdog not enabled
		return
	}
	for {
		// TODO: Check that the service is healthy?
		_, err := daemon.SdNotify(false, daemon.SdNotifyWatchdog)
		if err != nil {
			clog.Errorf("cannot notify systemd watchdog: %s", err)
		}
		time.Sleep(interval / 3)
	}
}
