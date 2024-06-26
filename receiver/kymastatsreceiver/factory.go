package kymastatsreceiver

import (
	"context"
	"time"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
)

var (
	typeStr = component.MustNewType("kymastatsreceiver")
)

const defaultInterval = 30 * time.Second

func createDefaultConfig() component.Config {
	return &Config{
		CollectionInterval: defaultInterval,
		APIConfig: k8sconfig.APIConfig{
			AuthType: k8sconfig.AuthTypeServiceAccount,
		},
	}
}

// NewFactory creates a factory for receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, component.StabilityLevelAlpha))
}

func createMetricsReceiver(_ context.Context, params receiver.CreateSettings, baseCfg component.Config, consumer consumer.Metrics) (receiver.Metrics, error) {
	return &kymaStatsReceiver{
		config:       baseCfg.(*Config),
		nextConsumer: consumer,
		settings:     &params,
	}, nil
}
