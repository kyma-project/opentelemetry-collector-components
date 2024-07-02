package singletonreceivercreator

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/singletonreceivercreator/internal/metadata"
)

// singletonreceivercreator implements consumer.Metrics.
type singletonReceiverCreator struct {
	params              receiver.Settings
	cfg                 *Config
	nextMetricsConsumer consumer.Metrics
	telemetryBuilder    *metadata.TelemetryBuilder

	host              component.Host
	subReceiverRunner *receiverRunner
	cancel            context.CancelFunc
}

func newSingletonReceiverCreator(params receiver.Settings, cfg *Config) (*singletonReceiverCreator, error) {
	telemetry, err := metadata.NewTelemetryBuilder(params.TelemetrySettings)
	return &singletonReceiverCreator{
		params:           params,
		cfg:              cfg,
		telemetryBuilder: telemetry,
	}, err
}

// Start leader receiver creator.
func (c *singletonReceiverCreator) Start(_ context.Context, host component.Host) error {
	c.host = host
	// Create a new context as specified in the interface documentation
	ctx := context.Background()
	ctx, c.cancel = context.WithCancel(ctx)

	c.params.TelemetrySettings.Logger.Info("Starting singleton election receiver...")

	client, err := c.cfg.getK8sClient()
	if err != nil {
		return fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	c.params.TelemetrySettings.Logger.Debug("Creating leader elector...")
	c.subReceiverRunner = newReceiverRunner(c.params, c.host)

	leaderElector, err := newLeaderElector(
		client,
		func(ctx context.Context) {
			c.params.TelemetrySettings.Logger.Info("Elected as leader")
			//nolint:contextcheck // no context passed, as this follows the same pattern as the upstream implementation
			if err := c.startSubReceiver(); err != nil {
				c.params.TelemetrySettings.Logger.Error("Failed to start subreceiver", zap.Error(err))
			}
		},
		//nolint:contextcheck // no context passed, as this follows the same pattern as the upstream implementation
		func() {
			c.params.TelemetrySettings.Logger.Info("Lost leadership")
			if err := c.stopSubReceiver(); err != nil {
				c.params.TelemetrySettings.Logger.Error("Failed to stop subreceiver", zap.Error(err))
			}
		},
		c.cfg.leaderElectionConfig,
		c.telemetryBuilder,
	)
	if err != nil {
		return fmt.Errorf("failed to create leader elector: %w", err)
	}

	//nolint:contextcheck // Create a new context as specified in the interface documentation
	go leaderElector.Run(ctx)
	return nil
}

func (c *singletonReceiverCreator) startSubReceiver() error {
	c.params.TelemetrySettings.Logger.Info("Starting wrapped receiver", zap.String("name", c.cfg.subreceiverConfig.id.String()))
	if err := c.subReceiverRunner.start(
		receiverConfig{
			id:     c.cfg.subreceiverConfig.id,
			config: c.cfg.subreceiverConfig.config,
		},
		c.nextMetricsConsumer,
	); err != nil {
		return fmt.Errorf("failed to start wrapped receiver %s: %w", c.cfg.subreceiverConfig.id.String(), err)
	}
	return nil
}

func (c *singletonReceiverCreator) stopSubReceiver() error {
	c.params.TelemetrySettings.Logger.Info("Stopping wrapped receiver", zap.String("name", c.cfg.subreceiverConfig.id.String()))
	// if we don't get the lease then the wrapped receiver is not set
	if c.subReceiverRunner != nil {
		return c.subReceiverRunner.shutdown(context.Background())
	}
	return nil
}

// Shutdown stops the leader receiver creature and all its receivers started at runtime.
func (c *singletonReceiverCreator) Shutdown(context.Context) error {
	if c.cancel != nil {
		c.cancel()
	}
	return nil
}
