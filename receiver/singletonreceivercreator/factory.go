package singletonreceivercreator

import (
	"context"
	"fmt"
	"os"

	"github.com/kyma-project/opentelemetry-collector-components/internal/k8sconfig"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"

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
	hostname, err := os.Hostname()
	if err != nil {
		return nil, fmt.Errorf("failed to get hostname: %w", err)
	}
	r := newSingletonReceiverCreator(params, cfg.(*Config), consumer, hostname)
	r.nextMetricsConsumer = consumer
	return r, nil
}
