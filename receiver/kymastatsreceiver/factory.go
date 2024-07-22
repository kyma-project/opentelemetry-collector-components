package kymastatsreceiver

import (
	"context"
	"errors"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scraperhelper"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/modulediscovery"
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

	discovery, err := config.getDiscoveryClient()
	if err != nil {
		return nil, err
	}

	dynamic, err := config.getDynamicClient()
	if err != nil {
		return nil, err
	}

	scrp, err := newKymaScraper(
		modulediscovery.New(discovery, params.Logger, config.ModuleGroups),
		dynamic,
		params,
		config.MetricsBuilderConfig,
	)
	if err != nil {
		return nil, err
	}
	return scraperhelper.NewScraperControllerReceiver(&config.ControllerConfig, params, consumer, scraperhelper.AddScraper(scrp))
}
