package main

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/creativeprojects/clog"
	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func setupMetricsServer(config cfg.Config) (func(context.Context) error, error) {
	if !config.Telemetry.Prometheus.Enabled {
		return func(_ context.Context) error { return nil }, nil
	}
	clog.Debugf("serving metrics at %s/metrics", config.Telemetry.Prometheus.Listen)
	http.Handle("/metrics", promhttp.Handler())
	server := http.Server{
		Addr: config.Telemetry.Prometheus.Listen,
	}
	wg := new(sync.WaitGroup)
	wg.Go(func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			clog.Errorf("metrics server: %s", err)
		}
		clog.Debug("metrics server closed")
	})

	return func(ctx context.Context) error {
		err := server.Shutdown(ctx)
		wg.Wait()
		return err
	}, nil
}
