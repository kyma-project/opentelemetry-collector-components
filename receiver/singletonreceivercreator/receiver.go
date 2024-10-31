package singletonreceivercreator

import (
	"context"
	"fmt"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
	"go.uber.org/zap"
	"k8s.io/client-go/tools/leaderelection"

	"github.com/kyma-project/opentelemetry-collector-components/receiver/singletonreceivercreator/internal/metadata"
)

// singletonreceivercreator implements consumer.Metrics.
type singletonReceiverCreator struct {
	params              receiver.Settings
	cfg                 *Config
	nextMetricsConsumer consumer.Metrics
	telemetryBuilder    *metadata.TelemetryBuilder

	leaseHolderID     string
	subReceiverRunner *receiverRunner
	cancel            context.CancelFunc
}

func newSingletonReceiverCreator(
	params receiver.Settings,
	cfg *Config,
	consumer consumer.Metrics,
	telemetryBuilder *metadata.TelemetryBuilder,
	leaseHolderID string,
) *singletonReceiverCreator {
	return &singletonReceiverCreator{
		params:              params,
		cfg:                 cfg,
		nextMetricsConsumer: consumer,
		telemetryBuilder:    telemetryBuilder,
		leaseHolderID:       leaseHolderID,
	}
}

// host is an interface that the component.Host passed to singletonreceivercreator's Start function must implement
type host interface {
	component.Host
	GetFactory(kind component.Kind, typ component.Type) component.Factory
}

// Start leader receiver creator.
func (c *singletonReceiverCreator) Start(_ context.Context, h component.Host) error {
	rcHost, ok := h.(host)
	if !ok {
		return fmt.Errorf("the receivercreator is not compatible with the provided component.host")
	} // Create a new context as specified in the interface documentation
	ctx := context.Background()
	ctx, c.cancel = context.WithCancel(ctx)

	c.params.TelemetrySettings.Logger.Info("Starting singleton election receiver...")

	c.params.TelemetrySettings.Logger.Debug("Creating leader elector...")
	c.subReceiverRunner = newReceiverRunner(c.params, rcHost)

	leaderElector, err := c.initLeaderElector()
	if err != nil {
		return fmt.Errorf("failed to create leader elector: %w", err)
	}

	go c.runLeaderElector(ctx, leaderElector)

	return nil
}

func (c *singletonReceiverCreator) initLeaderElector() (*leaderelection.LeaderElector, error) {
	client, err := c.cfg.getK8sClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return newLeaderElector(
		c.cfg.leaderElectionConfig,
		client,
		c.telemetryBuilder,
		func(ctx context.Context) {
			c.params.TelemetrySettings.Logger.Info("Leader lease acquired")
			//nolint:contextcheck // no context passed, as this follows the same pattern as the upstream implementation
			if err := c.startSubReceiver(); err != nil {
				c.params.TelemetrySettings.Logger.Error("Failed to start subreceiver", zap.Error(err))
			}
		},
		//nolint:contextcheck // no context passed, as this follows the same pattern as the upstream implementation
		func() {
			c.params.TelemetrySettings.Logger.Info("Leader lease lost")
			if err := c.stopSubReceiver(); err != nil {
				c.params.TelemetrySettings.Logger.Error("Failed to stop subreceiver", zap.Error(err))
			}
		},
		c.leaseHolderID,
	)
}

func (c *singletonReceiverCreator) runLeaderElector(ctx context.Context, leaderElector *leaderelection.LeaderElector) {
	// Leader election loop stops if context is canceled or the leader elector loses the lease.
	// The loop allows continued participation in leader election, even if the lease is lost.
	for {
		leaderElector.Run(ctx)

		if ctx.Err() != nil {
			break
		}

		c.params.TelemetrySettings.Logger.Info("Leader lease lost. Returning to standby mode...")
	}
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
