package main

import (
	"context"

	"github.com/creativeprojects/hardware-events/cfg"
	"github.com/creativeprojects/hardware-events/lib"
	"go.opentelemetry.io/otel/exporters/prometheus"
)

func setupTelemetry(config cfg.Config, global *lib.Global) (func(context.Context) error, error) {
	if !config.Telemetry.Prometheus.Enabled {
		return func(_ context.Context) error { return nil }, nil
	}
	exporter, err := prometheus.New()
	if err != nil {
		return nil, err
	}
	telemetry, err := lib.NewTelemetry(global, exporter)
	if err != nil {
		return func(_ context.Context) error { return nil }, err
	}
	return telemetry.Shutdown, nil
}
