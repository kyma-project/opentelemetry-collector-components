package kymastatsreceiver

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/scraper/scraperhelper"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
)

var (
	typeStr = component.MustNewType("kymastats")
)

func createDefaultConfig() component.Config {
	return &Config{
		APIConfig: k8sconfig.APIConfig{
			AuthType: k8sconfig.AuthTypeServiceAccount,
		},
		ControllerConfig:     scraperhelper.NewDefaultControllerConfig(),
		MetricsBuilderConfig: metadata.DefaultMetricsBuilderConfig(),
	}
}

// NewFactory creates a factory for receiver.
func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, component.StabilityLevelAlpha))
}

func createMetricsReceiver(_ context.Context, params receiver.Settings, baseCfg component.Config, consumer consumer.Metrics) (receiver.Metrics, error) {
	config, ok := baseCfg.(*Config)
	if !ok {
		return nil, errors.New("invalid configuration")
	}

	dynamic, err := config.getDynamicClient()
	if err != nil {
		return nil, err
	}

	scrp, err := newKymaScraper(
		*config,
		dynamic,
		params,
	)
	if err != nil {
		return nil, err
	}

	return scraperhelper.NewMetricsController(&config.ControllerConfig, params, consumer, scraperhelper.AddMetricsScraper(metadata.Type, scrp))
}
