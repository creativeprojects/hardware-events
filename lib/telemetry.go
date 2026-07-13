package lib

import (
	"context"

	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
)

const meterName = "github.com/creativeprojects/hardware-events"

type Telemetry struct{}

func NewTelemetry(global *Global, exporter metric.Reader) (*Telemetry, error) {
	provider := metric.NewMeterProvider(metric.WithReader(exporter))
	meter := provider.Meter(meterName)
	for _, disk := range global.Disks {
		if disk.HasTemperature() {
			meter.Float64ObservableGauge(disk.Name,
				api.WithDescription("device "+disk.Device),
				api.WithFloat64Callback(func(ctx context.Context, fo api.Float64Observer) error {
					if disk.TemperatureAvailable() {
						fo.Observe(float64(disk.Temperature()))
					}
					return nil
				}),
			)
		}
	}
	return &Telemetry{}, nil
}
