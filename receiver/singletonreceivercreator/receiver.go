// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

package singletonreceivercreator

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
	"k8s.io/client-go/kubernetes"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/singletonreceivercreator/internal/k8sconfig"
)

var _ receiver.Metrics = (*leaderReceiverCreator)(nil)

// leaderReceiverCreator implements consumer.Metrics.
type leaderReceiverCreator struct {
	params              receiver.CreateSettings
	cfg                 *Config
	nextMetricsConsumer consumer.Metrics

	host              component.Host
	subReceiverRunner *receiverRunner
	cancel            context.CancelFunc
	getK8sClient      func(authType k8sconfig.AuthType) (kubernetes.Interface, error)
}

func newLeaderReceiverCreator(params receiver.CreateSettings, cfg *Config) component.Component {
	return &leaderReceiverCreator{
		params:       params,
		cfg:          cfg,
		getK8sClient: k8sconfig.GetK8sClient,
	}
}

// Start leader receiver creator.
func (c *leaderReceiverCreator) Start(_ context.Context, host component.Host) error {
	c.host = host
	// Create a new context as specified in the interface documentation
	ctx := context.Background()
	ctx, c.cancel = context.WithCancel(ctx)

	c.params.TelemetrySettings.Logger.Debug("Starting leader election receiver...")

	client, err := c.getK8sClient(c.cfg.authType)
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	c.params.TelemetrySettings.Logger.Debug("Creating leader elector...")
	c.subReceiverRunner = newReceiverRunner(c.params, c.host)

	leaderElector, err := newLeaderElector(
		client,
		func(ctx context.Context) {
			c.params.TelemetrySettings.Logger.Info("Elected as leader")
			if err := c.startSubReceiver(); err != nil {
				c.params.TelemetrySettings.Logger.Error("Failed to start subreceiver", zap.Error(err))
			}
		},
		func() {
			c.params.TelemetrySettings.Logger.Info("Lost leadership")
			if err := c.stopSubReceiver(); err != nil {
				c.params.TelemetrySettings.Logger.Error("Failed to stop subreceiver", zap.Error(err))
			}
		},
		c.cfg.leaderElectionConfig,
	)
	if err != nil {
		return fmt.Errorf("failed to create leader elector: %w", err)
	}

	go leaderElector.Run(ctx)
	return nil
}

func (c *leaderReceiverCreator) startSubReceiver() error {
	c.params.TelemetrySettings.Logger.Info("Starting sub-receiver",
		zap.String("name", c.cfg.subreceiverConfig.id.String()))
	if err := c.subReceiverRunner.start(
		receiverConfig{
			id:     c.cfg.subreceiverConfig.id,
			config: c.cfg.subreceiverConfig.config,
		},
		c.nextMetricsConsumer,
	); err != nil {
		return fmt.Errorf("failed to start subreceiver %s: %w", c.cfg.subreceiverConfig.id.String(), err)
	}
	return nil
}

func (c *leaderReceiverCreator) stopSubReceiver() error {
	c.params.TelemetrySettings.Logger.Info("Stopping subreceiver",
		zap.String("name", c.cfg.subreceiverConfig.id.String()))
	// if we dont get the lease then the subreceiver is not set
	if c.subReceiverRunner != nil {
		return c.subReceiverRunner.shutdown(context.Background())
	}
	return nil
}

// Shutdown stops the leader receiver creater and all its receivers started at runtime.
func (c *leaderReceiverCreator) Shutdown(context.Context) error {
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}
