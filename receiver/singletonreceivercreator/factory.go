package singletonreceivercreator

import (
	"context"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/singletonreceivercreator/internal/k8sconfig"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/singletonreceivercreator/internal/metadata"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/singletonreceivercreator/internal/sharedcomponent"
)

var receivers = sharedcomponent.NewSharedComponents()

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		metadata.Type,
		createDefaultConfig,
		receiver.WithMetrics(createMetricsReceiver, metadata.MetricsStability),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		authType: k8sconfig.AuthTypeServiceAccount,
		leaderElectionConfig: leaderElectionConfig{
			leaseName:            "singleton-receiver",
			leaseNamespace:       "default",
			leaseDurationSeconds: defaultLeaseDuration,
			renewDeadlineSeconds: defaultRenewDeadline,
			retryPeriodSeconds:   defaultRetryPeriod,
		},
		subreceiverConfig: receiverConfig{},
	}
}

func createMetricsReceiver(_ context.Context, params receiver.Settings, cfg component.Config, consumer consumer.Metrics) (receiver.Metrics, error) {
	r := receivers.GetOrAdd(cfg, func() component.Component {
		return newSingletonReceiverCreator(params, cfg.(*Config))
	})
	r.Component.(*singletonReceiverCreator).nextMetricsConsumer = consumer
	return r, nil
}
