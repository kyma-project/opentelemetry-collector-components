package singletonreceivercreator

import (
	"context"
	"errors"
	"os"

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

func createMetricsReceiver(
	_ context.Context,
	params receiver.Settings,
	baseCfg component.Config,
	consumer consumer.Metrics,
) (receiver.Metrics, error) {
	cfg, ok := baseCfg.(*Config)
	if !ok {
		return nil, errors.New("invalid configuration")
	}
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	telemetryBuilder, err := metadata.NewTelemetryBuilder(params.TelemetrySettings)
	if err != nil {
		return nil, err
	}

	return newSingletonReceiverCreator(
		params,
		cfg,
		consumer,
		telemetryBuilder,
		hostname,
	), nil
}
