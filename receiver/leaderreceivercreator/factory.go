// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package leaderreceivercreator

import (
	"context"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/leaderreceivercreator/internal/k8sconfig"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/leaderreceivercreator/internal/metadata"
	"github.com/kyma-project/opentelemetry-collector-components/receiver/leaderreceivercreator/internal/sharedcomponent"
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
		leaderElectionConfig: leaderElectionConfig{
			authType:             k8sconfig.AuthTypeServiceAccount,
			leaseName:            "my-lease",
			leaseNamespace:       "default",
			leaseDurationSeconds: defaultLeaseDuration,
			renewDeadlineSeconds: defaultRenewDeadline,
			retryPeriodSeconds:   defaultRetryPeriod,
		},
		subreceiverConfig: receiverConfig{},
	}
}

func createMetricsReceiver(_ context.Context, params receiver.CreateSettings, cfg component.Config, consumer consumer.Metrics) (receiver.Metrics, error) {
	r := receivers.GetOrAdd(cfg, func() component.Component {
		return newLeaderReceiverCreator(params, cfg.(*Config))
	})
	r.Component.(*leaderReceiverCreator).nextMetricsConsumer = consumer
	return r, nil
}
