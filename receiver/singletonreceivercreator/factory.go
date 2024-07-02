package singletonreceivercreator

import (
	"context"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/singletonreceivercreator/internal/metadata"
)

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		metadata.Type,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, metadata.MetricsStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		APIConfig: k8sconfig.APIConfig{
			AuthType: k8sconfig.AuthTypeServiceAccount,
		},
		leaderElectionConfig: leaderElectionConfig{
			leaseName:      "singleton-receiver",
			leaseNamespace: "default",
			leaseDuration:  defaultLeaseDuration,
			renewDuration:  defaultRenewDeadline,
			retryPeriod:    defaultRetryPeriod,
		},
		subreceiverConfig: receiverConfig{},
	}
}

func createMetricsReceiver(_ context.Context, params receiver.Settings, cfg component.Config, consumer consumer.Metrics) (receiver.Metrics, error) {
	r, err := newSingletonReceiverCreator(params, cfg.(*Config))
	if err != nil {
		return nil, err
	}

	r.nextMetricsConsumer = consumer
	return r, nil
}
