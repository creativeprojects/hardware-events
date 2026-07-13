package lib

import (
	"context"
	"fmt"

	api "go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/sdk/metric"
)

const meterName = "hardware-events"

type Telemetry struct {
	provider *metric.MeterProvider
}

func NewTelemetry(global *Global, exporter metric.Reader) (*Telemetry, error) {
	provider := metric.NewMeterProvider(metric.WithReader(exporter))

	meter := provider.Meter(meterName)

	for _, disk := range global.Disks {
		if disk == nil {
			continue
		}
		err := setupDisk(meter, disk)
		if err != nil {
			return nil, err
		}
	}
	if global.FanControl != nil {
		for _, zone := range global.FanControl.Zones {
			err := setupFanZone(meter, zone)
			if err != nil {
				return nil, err
			}
		}
	}

	return &Telemetry{
		provider: provider,
	}, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	return t.provider.Shutdown(ctx)
}

func setupDisk(meter api.Meter, disk *Disk) error {
	if disk.HasTemperature() {
		name := fmt.Sprintf("disk_temperature_%s", disk.Name)
		_, err := meter.Int64ObservableGauge(name,
			api.WithDescription("Internal temperature sensor from device "+disk.Device),
			api.WithUnit("degree Celsius"),
			api.WithInt64Callback(func(ctx context.Context, fo api.Int64Observer) error {
				if disk.TemperatureAvailable() {
					fo.Observe(int64(disk.Temperature()))
				}
				return nil
			}),
		)
		if err != nil {
			return err
		}
	}

	name := fmt.Sprintf("disk_active_%s", disk.Name)
	_, err := meter.Int64ObservableGauge(name,
		api.WithDescription("Active device "+disk.Device+": 0 when inactive, 1 when active"),
		api.WithInt64Callback(func(ctx context.Context, fo api.Int64Observer) error {
			var value int64 = 0
			if disk.IsActive() {
				value = 1
			}
			fo.Observe(value)
			return nil
		}),
	)
	if err != nil {
		return err
	}
	return nil
}

func setupFanZone(meter api.Meter, zone *Zone) error {
	name := fmt.Sprintf("fan_speed_%s", zone.Name)
	_, err := meter.Int64ObservableGauge(name,
		api.WithDescription(fmt.Sprintf("Fan speed on zone %d", zone.ID)),
		api.WithUnit("percent"),
		api.WithInt64Callback(func(ctx context.Context, fo api.Int64Observer) error {
			fo.Observe(int64(zone.CurrentSpeed()))
			return nil
		}),
	)
	if err != nil {
		return err
	}
	return nil
}
