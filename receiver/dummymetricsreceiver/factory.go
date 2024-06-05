package dummymetricsreceiver

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/consumer"
)

var (
	typeStr = component.MustNewType("dummymetricreceiver")
)

const (
	defaultInterval = 1 * time.Minute
)

func createDefaultConfig() component.Config {
	return &Config{
		Interval: defaultInterval.String(),
	}
}

func createMetricsReceiver(_ context.Context, params receiver.CreateSettings, baseCfg component.Config, consumer consumer.Metrics) (receiver.Metrics, error) {
	return &dummyMetricsReceiver{
		config:       baseCfg.(*Config),
		nextConsumer: consumer,
		settings:     &params,
	}, nil
}

// NewFactory creates a factory for dummymetricreceiver receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, component.StabilityLevelAlpha))
}
