package kymastatsreceiver

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.opentelemetry.io/collector/receiver/scraperhelper"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/kymastatsreceiver/internal/metadata"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
)

var (
	typeStr = component.MustNewType("kymastatsreceiver")
)

func createDefaultConfig() component.Config {
	return &Config{
		ControllerConfig: scraperhelper.NewDefaultControllerConfig(),
		APIConfig: k8sconfig.APIConfig{
			AuthType: k8sconfig.AuthTypeKubeConfig,
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
	config := baseCfg.(*Config)
	mbConfig := metadata.MetricsBuilderConfig{
		KymaTelemetryModuleStat: []metadata.MetricConfig{
			{
				ResourceGroup:   "operator.kyma-project.io",
				ResourceName:    "Telemetry",
				ResourceVersion: "v1alpha1",
				Name:            "kyma.telemetry.status.stat",
				Description:     "Kyma telemetry module status",
				Unit:            "1",
			},
			{
				ResourceGroup:   "operator.kyma-project.io",
				ResourceName:    "Telemetry",
				ResourceVersion: "v1alpha1",
				Name:            "Kyma.telemetry.status.condition",
				Description:     "Kyma telemetry module conditions",
				Unit:            "1",
			},
		},
	}

	client, err := config.getK8sDynamicClient()
	if err != nil {
		return nil, err
	}
	scrp, err := newKymaScraper(client, mbConfig)
	if err != nil {
		return nil, err
	}
	return scraperhelper.NewScraperControllerReceiver(&config.ControllerConfig, params, consumer, scraperhelper.AddScraper(scrp))
}
