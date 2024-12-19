package dummyreceiver

import (
	"context"
	"errors"
	"sync"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

var (
	typeStr = component.MustNewType("dummy")
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
	cfg, ok := baseCfg.(*Config)
	if !ok {
		return nil, errors.New("invalid configuration")
	}

	return &dummyReceiver{
		config:       cfg,
		nextConsumer: consumer,
		settings:     &params,
		wg:           &sync.WaitGroup{},
	}, nil
}

// NewFactory creates a factory for dummyReceiver receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, component.StabilityLevelAlpha))
}
