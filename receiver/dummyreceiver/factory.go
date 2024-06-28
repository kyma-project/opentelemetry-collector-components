package dummyreceiver

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

var (
	typeStr = component.MustNewType("dummyreceiver")
)

const (
	defaultInterval = 1 * time.Minute
)

func createDefaultConfig() component.Config {
	return &Config{
		Interval: defaultInterval.String(),
	}
}

func createMetricsReceiver(_ context.Context, params receiver.Settings, baseCfg component.Config, consumer consumer.Metrics) (receiver.Metrics, error) {
	return &dummyreceiver{
		config:       baseCfg.(*Config),
		nextConsumer: consumer,
		settings:     &params,
	}, nil
}

// NewFactory creates a factory for dummyreceiver receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, component.StabilityLevelAlpha))
}
