package lib

import (
	"context"

	"go.opentelemetry.io/otel/attribute"
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

	err := setupDisks(meter, global.Disks)
	if err != nil {
		return nil, err
	}

	if global.FanControl != nil {
		err := setupFanZones(meter, global.FanControl.Zones)
		if err != nil {
			return nil, err
		}
	}

	return &Telemetry{
		provider: provider,
	}, nil
}

func (t *Telemetry) Shutdown(ctx context.Context) error {
	return t.provider.Shutdown(ctx)
}

func setupDisks(meter api.Meter, disks map[string]*Disk) error {
	_, err := meter.Int64ObservableGauge("disk_temperature",
		api.WithDescription("Disk internal temperature sensor"),
		api.WithUnit("degree Celsius"),
		api.WithInt64Callback(func(ctx context.Context, fo api.Int64Observer) error {
			for _, disk := range disks {
				if disk == nil {
					return nil
				}
				if disk.TemperatureAvailable() {
					fo.Observe(int64(disk.Temperature()), api.WithAttributeSet(attribute.NewSet(diskAttributes(disk)...)))
				}
			}
			return nil
		}),
	)
	if err != nil {
		return err
	}

	_, err = meter.Int64ObservableGauge("disk_active",
		api.WithDescription("Active device: 0 when inactive, 1 when active"),
		api.WithInt64Callback(func(ctx context.Context, fo api.Int64Observer) error {
			for _, disk := range disks {
				if disk == nil {
					return nil
				}
				var value int64 = 0
				if disk.IsActive() {
					value = 1
				}
				fo.Observe(value, api.WithAttributeSet(attribute.NewSet(diskAttributes(disk)...)))
			}
			return nil
		}),
	)
	if err != nil {
		return err
	}
	return nil
}

func setupFanZones(meter api.Meter, zones map[string]*Zone) error {
	_, err := meter.Int64ObservableGauge("fan_speed",
		api.WithDescription("Fan speed from 0 to 100%"),
		api.WithUnit("percent"),
		api.WithInt64Callback(func(ctx context.Context, fo api.Int64Observer) error {
			for _, zone := range zones {
				attributes := []attribute.KeyValue{
					{Key: "name", Value: attribute.StringValue(zone.Name)},
					{Key: "id", Value: attribute.IntValue(zone.ID)},
				}
				fo.Observe(int64(zone.CurrentFanSpeed()), api.WithAttributeSet(attribute.NewSet(attributes...)))
			}
			return nil
		}),
	)
	if err != nil {
		return err
	}
	return nil
}

func diskAttributes(disk *Disk) []attribute.KeyValue {
	attributes := []attribute.KeyValue{
		{Key: "name", Value: attribute.StringValue(disk.Name)},
		{Key: "device", Value: attribute.StringValue(disk.Device)},
	}
	if disk.Pool != "" {
		attributes = append(attributes, attribute.KeyValue{
			Key:   "pool",
			Value: attribute.StringValue(disk.Pool),
		})
	}
	return attributes
}
